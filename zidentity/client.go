package zidentity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	rl "github.com/SecurityGeekIO/zscaler-sdk-go/v2/ratelimiter"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/utils"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"

	"github.com/google/go-querystring/query"
	"github.com/google/uuid"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/cache"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/common"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/logger"
)

func (client *zidentityClient) WithFreshCache() {
	client.config.freshCache = true
}

func (client *zidentityClient) GetLogger() logger.Logger {
	return client.logger
}

func (client *zidentityClient) GetCustomerID() string {
	return client.config.customerID
}

func (client *zidentityClient) GetCloud() string {
	return client.config.oauth2Credentials.Cloud
}

func (client *zidentityClient) SetCustomerID(customerID string) {
	client.config.customerID = customerID
}

func (client *zidentityClient) GetBaseURL() string {
	return client.config.baseURL.String()
}

func (client *zidentityClient) NewRequestDoGeneric(baseUrl, path, method string, options, body, v interface{}, contentType string, infraOptions ...common.Option) (*http.Response, error) {
	infra := "zpa"
	if len(infraOptions) > 0 {
		for _, o := range infraOptions {
			if o.Name == common.ZscalerInfraOption {
				infra = o.Value
			}
		}
	}
	if baseUrl == "" {
		baseUrl = client.config.baseURL.String()
	}

	baseUrlParsed, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	if baseUrlParsed.Host == client.config.baseURL.Host {
		if infra == "zia" {
			path = "api/v1/" + strings.TrimPrefix(path, "/")
		}
		path = fmt.Sprintf("%s/%s", infra, strings.TrimPrefix(path, "/"))
	}

	baseUrlParsed, err = url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	req, err := client.getRequest(baseUrlParsed, method, path, options, body)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	key := cache.CreateCacheKey(req)
	if resp, err := client.handleCache(key, req, v); resp != nil && err == nil {
		return resp, nil
	}

	resp, err := client.newRequestDoCustom(baseUrlParsed, method, path, options, body, v)
	if err != nil {
		return resp, err
	}
	if client.config.cacheEnabled && resp.StatusCode >= 200 && resp.StatusCode <= 299 && req.Method == http.MethodGet && v != nil && reflect.TypeOf(v).Kind() != reflect.Slice {
		d, err := json.Marshal(v)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewReader(d))
			client.logger.Printf("[INFO] saving to cache, key:%s\n", key)
			client.cache.Set(key, cache.CopyResponse(resp))
		} else {
			client.logger.Printf("[ERROR] saving to cache error:%s, key:%s\n", err, key)
		}
	}
	return resp, nil
}

func (client *zidentityClient) NewRequestDo(method, path string, options, body, v interface{}, infraOptions ...common.Option) (*http.Response, error) {
	return client.NewRequestDoGeneric(client.config.baseURL.String(), path, method, options, body, v, contentTypeJSON, infraOptions...)
}

func (client *zidentityClient) handleCache(key string, req *http.Request, v interface{}) (*http.Response, error) {
	if client.config.cacheEnabled {
		if req.Method != http.MethodGet {
			// this will allow to remove resource from cache when PUT/DELETE/PATCH requests are called, which modifies the resource
			client.cache.Delete(key)
			// to avoid resources that GET url is not the same as DELETE/PUT/PATCH url, because of different query params.
			// example delete app segment has key url/<id>?forceDelete=true but GET has url/<id>, in this case we clean the whole cache entries with key prefix url/<id>
			client.cache.ClearAllKeysWithPrefix(strings.Split(key, "?")[0])
		}
		resp := client.cache.Get(key)
		inCache := resp != nil
		if client.config.freshCache {
			client.cache.Delete(key)
			inCache = false
			client.config.freshCache = false
		}
		if inCache {
			if v != nil {
				respData, err := io.ReadAll(resp.Body)
				if err == nil {
					resp.Body = io.NopCloser(bytes.NewBuffer(respData))
				}
				if err := decodeJSON(respData, v); err != nil {
					return resp, err
				}
			}
			unescapeHTML(v)
			client.logger.Printf("[INFO] served from cache, key:%s\n", key)
			return resp, nil
		}
	}
	return nil, nil
}

func (client *zidentityClient) newRequestDoCustom(baseURL *url.URL, method, urlStr string, options, body, v interface{}) (*http.Response, error) {
	err := client.authenticate()
	if err != nil {
		return nil, err
	}
	req, err := client.newRequest(baseURL, method, urlStr, options, body)
	if err != nil {
		return nil, err
	}
	reqID := uuid.NewString()
	start := time.Now()
	logger.LogRequest(client.logger, req, reqID, nil, true)
	resp, err := client.do(req, v, start, reqID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		err := client.authenticate()
		if err != nil {
			return nil, err
		}

		resp, err := client.do(req, v, start, reqID)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		return resp, nil
	}
	return resp, err
}

func (client *zidentityClient) authenticate() error {
	client.config.Lock()
	defer client.config.Unlock()

	if client.config.oauth2Credentials.AuthToken == nil || client.config.oauth2Credentials.AuthToken.AccessToken == "" || utils.IsTokenExpired(client.config.oauth2Credentials.AuthToken.AccessToken) {
		a, err := Authenticate(
			client.config.oauth2Credentials,
			client.GetHTTPClient(),
			client.config.userAgent,
		)
		if err != nil {
			return err
		}
		client.config.oauth2Credentials.AuthToken = &AuthToken{
			TokenType:   a.TokenType,
			AccessToken: a.AccessToken,
		}
		return nil
	}
	return nil
}

func getMicrotenantIDFromBody(body interface{}) string {
	if body == nil {
		return ""
	}

	d, err := json.Marshal(body)
	if err != nil {
		return ""
	}
	dataMap := map[string]interface{}{}
	err = json.Unmarshal(d, &dataMap)
	if err != nil {
		return ""
	}
	if microTenantID, ok := dataMap["microtenantId"]; ok && microTenantID != nil && microTenantID != "" {
		return fmt.Sprintf("%v", microTenantID)
	}
	return ""
}

func getMicrotenantIDFromEnvVar(body interface{}) string {
	return os.Getenv("ZPA_MICROTENANT_ID")
}

func (client *zidentityClient) injectMicrotentantID(body interface{}, q url.Values) url.Values {
	if q.Has("microtenantId") && q.Get("microtenantId") != "" {
		return q
	}

	microTenantID := getMicrotenantIDFromBody(body)
	if microTenantID != "" {
		q.Add("microtenantId", microTenantID)
		return q
	}

	microTenantID = getMicrotenantIDFromEnvVar(body)
	if microTenantID != "" {
		q.Add("microtenantId", microTenantID)
		return q
	}
	return q
}

func (client *zidentityClient) getRequest(baseURL *url.URL, method, urlPath string, options, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	// Parse the URL path to separate the path from any query string
	parsedPath, err := url.Parse(urlPath)
	if err != nil {
		return nil, err
	}

	parsedPath.Path = baseURL.Path + parsedPath.Path
	u := baseURL.ResolveReference(parsedPath)

	// Handle query parameters from options and any additional logic
	if options == nil {
		options = struct{}{}
	}
	var q url.Values

	if options == nil {
		options = struct{}{}
	}

	switch opt := options.(type) {
	case url.Values:
		q = opt
	default:
		q, err = query.Values(options)
		if err != nil {
			return nil, err
		}
	}

	// Here, injectMicrotenantID or any similar function should ensure
	// it's not duplicating query parameters that may already be present in urlPath
	q = client.injectMicrotentantID(body, q)

	// Merge query params from urlPath and options. Avoid overwriting any existing params.
	for key, values := range parsedPath.Query() {
		for _, value := range values {
			q.Add(key, value)
		}
	}

	// Set the final query string
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Generating the Http request
func (client *zidentityClient) newRequest(baseURL *url.URL, method, urlPath string, options, body interface{}) (*http.Request, error) {
	if client.config.oauth2Credentials.AuthToken == nil || client.config.oauth2Credentials.AuthToken.AccessToken == "" {
		client.logger.Printf("[ERROR] Failed to signin the user\n")
		return nil, fmt.Errorf("failed to signin the user\n")
	}
	req, err := client.getRequest(baseURL, method, urlPath, options, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.config.oauth2Credentials.AuthToken.AccessToken))
	req.Header.Add("Content-Type", "application/json")

	if client.config.userAgent != "" {
		req.Header.Add("User-Agent", client.config.userAgent)
	}

	return req, nil
}

func checkErrorInResponse(res *http.Response, respData []byte) error {
	if c := res.StatusCode; c >= 200 && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: res}
	if len(respData) > 0 {
		errorResponse.Message = string(respData)
	}
	return errorResponse
}

func (client *zidentityClient) do(req *http.Request, v interface{}, start time.Time, reqID string) (*http.Response, error) {
	resp, err := client.GetHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err == nil {
		resp.Body = io.NopCloser(bytes.NewBuffer(respData))
	}
	if err := checkErrorInResponse(resp, respData); err != nil {
		return resp, err
	}

	if v != nil {
		if err := decodeJSON(respData, v); err != nil {
			return resp, err
		}
	}
	logger.LogResponse(client.logger, resp, start, reqID)
	unescapeHTML(v)
	return resp, nil
}

func decodeJSON(respData []byte, v interface{}) error {
	return json.NewDecoder(bytes.NewBuffer(respData)).Decode(&v)
}

func unescapeHTML(entity interface{}) {
	if entity == nil {
		return
	}
	data, err := json.Marshal(entity)
	if err != nil {
		return
	}
	var mapData map[string]interface{}
	err = json.Unmarshal(data, &mapData)
	if err != nil {
		return
	}
	for _, field := range []string{"name", "description"} {
		if v, ok := mapData[field]; ok && v != nil {
			str, ok := v.(string)
			if ok {
				mapData[field] = html.UnescapeString(html.UnescapeString(str))
			}
		}
	}
	data, err = json.Marshal(mapData)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, entity)
}

func (c *zidentityClient) GetHTTPClient() *http.Client {
	if c.httpClient == nil {
		if c.config.backoffConf != nil && c.config.backoffConf.Enabled {
			retryableClient := retryablehttp.NewClient()
			retryableClient.Logger = c.logger
			retryableClient.RetryWaitMin = time.Second * time.Duration(c.config.backoffConf.RetryWaitMinSeconds)
			retryableClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
				if resp != nil {
					if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
						if s := resp.Header.Get("Retry-After"); s != "" {
							if sleep, err := strconv.ParseInt(s, 10, 64); err == nil {
								return time.Second * time.Duration(sleep)
							} else {
								dur, err := time.ParseDuration(s)
								if err == nil {
									return dur
								}
							}
						}
					}
					if resp.Request != nil {
						wait, duration := c.rateLimiter.Wait(resp.Request.Method)
						if wait {
							c.logger.Printf("[INFO] rate limiter wait duration:%s\n", duration.String())
						} else {
							return 0
						}
					}
				}
				// default to exp backoff
				mult := math.Pow(2, float64(attemptNum)) * float64(min)
				sleep := time.Duration(mult)
				if float64(sleep) != mult || sleep > max {
					sleep = max
				}
				return sleep
			}
			retryableClient.RetryWaitMax = time.Second * time.Duration(c.config.backoffConf.RetryWaitMaxSeconds)
			retryableClient.RetryMax = c.config.backoffConf.MaxNumOfRetries
			retryableClient.HTTPClient.Transport = logging.NewSubsystemLoggingHTTPTransport("gozscaler", retryableClient.HTTPClient.Transport)
			retryableClient.CheckRetry = checkRetry
			retryableClient.HTTPClient.Timeout = defaultTimeout
			c.httpClient = retryableClient.StandardClient()
		} else {
			c.httpClient = &http.Client{
				Timeout: defaultTimeout,
			}
		}
	}
	return c.httpClient
}

func containsInt(codes []int, code int) bool {
	for _, a := range codes {
		if a == code {
			return true
		}
	}
	return false
}

// getRetryOnStatusCodes return a list of http status codes we want to apply retry on.
// return empty slice to enable retry on all connection & server errors.
// or return []int{429}  to retry on only TooManyRequests error
func getRetryOnStatusCodes() []int {
	return []int{http.StatusTooManyRequests, http.StatusConflict}
}

// Used to make http client retry on provided list of response status codes
func checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if resp != nil && containsInt(getRetryOnStatusCodes(), resp.StatusCode) {
		if resp.StatusCode == http.StatusConflict {
			respMap := map[string]string{}
			data, err := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewBuffer(data))
			if err == nil {
				_ = json.Unmarshal(data, &respMap)
				if errorID, ok := respMap["id"]; ok && (errorID == "api.concurrent.access.error") {
					return true, nil
				}
			}
		}
		return true, nil
	}
	if resp != nil && resp.StatusCode == http.StatusBadRequest {
		respMap := map[string]string{}
		data, err := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(data))
		if err == nil {
			_ = json.Unmarshal(data, &respMap)
			if errorID, ok := respMap["id"]; ok && (errorID == "non.restricted.entity.authorization.failed" || errorID == "bad.request") {
				return true, nil
			}
		}
		// Implemented to handle upstream restrictions on simultaneous requests when dealing with CRUD operations, related to ZPA Access policy rule order
		// ET-53585: https://jira.corp.zscaler.com/browse/ET-53585
		// ET-48860: https://confluence.corp.zscaler.com/display/ET/ET-48860+incorrect+rules+order
		if err == nil {
			_ = json.Unmarshal(data, &respMap)
			if errorID, ok := respMap["id"]; ok && (errorID == "db.simultaneous.request" || errorID == "bad.request") {
				return true, nil
			}
		}

		// ET-66174: https://jira.corp.zscaler.com/browse/ET-66174
		// DOC-51102: https://jira.corp.zscaler.com/browse/DOC-51102
		if err == nil {
			_ = json.Unmarshal(data, &respMap)
			if errorID, ok := respMap["id"]; ok && (errorID == "api.concurrent.access.error" || errorID == "bad.request") {
				return true, nil
			}
		}
	}
	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}

// The cloud parameter is optional and is handled in the Config object.
func NewClient(clientID, clientSecret, privateKeyPath, customerID, vanityDomain, userAgent string, optionalCloud ...string) (common.Client, error) {
	var logger logger.Logger = logger.GetDefaultLogger(loggerPrefix)

	config, err := newConfig(clientID, clientSecret, privateKeyPath, customerID, vanityDomain, userAgent, logger, optionalCloud...)
	if err != nil {
		return nil, err
	}

	// Setup the cache
	cche, err := cache.NewCache(config.cacheTtl, config.cacheCleanwindow, config.cacheMaxSizeMB)
	if err != nil {
		cche = cache.NewNopCache()
	}

	// Create and return the Client with the provided Config
	client := &zidentityClient{
		config:      config,
		cache:       cche,
		rateLimiter: rl.NewRateLimiter(20, 10, 10, 10),
		logger:      logger,
	}
	return client, nil
}

func newConfig(clientID, clientSecret, privateKeyPath, customerID, vanityDomain, userAgent string, logger logger.Logger, optionalCloud ...string) (*config, error) {

	// Fallback to environment variables if clientID, clientSecret, customerID, or userAgent are not provided
	if clientID == "" || (clientSecret == "" && privateKeyPath == "") || customerID == "" || userAgent == "" {
		clientID = os.Getenv(ZIDENTITY_CLIENT_ID)
		clientSecret = os.Getenv(ZIDENTITY_CLIENT_SECRET)
		privateKeyPath = os.Getenv(ZIDENTITY_PRIVATE_KEY)
		customerID = os.Getenv(ZPA_CUSTOMER_ID)
	}

	var cloud string
	if len(optionalCloud) > 0 && optionalCloud[0] != "" {
		cloud = optionalCloud[0] // Use provided cloud value
	} else {
		cloud = os.Getenv(ZSCALER_CLOUD) // Fallback to environment variable
	}

	// Default to production if no cloud is specified
	if cloud == "" {
		cloud = "PRODUCTION"
	}

	// Check for vanity domain and ensure proper formatting of the OAuth2 provider URL based on the cloud
	if vanityDomain == "" {
		vanityDomain = os.Getenv(ZIDENTITY_VANITY_DOMAIN)
	}

	// Fallback to configuration file if credentials are not provided
	if clientID == "" || clientSecret == "" {
		creds, err := loadCredentialsFromConfig(logger)
		if err != nil || creds == nil {
			return nil, err
		}
		clientID = creds.ClientID
		clientSecret = creds.ClientSecret
		customerID = creds.CustomerID
		cloud = creds.Cloud
	}

	// Default to production if no cloud is specified
	var rawUrl string
	if strings.EqualFold(cloud, "PRODUCTION") {
		rawUrl = "https://api.zsapi.net/"
	} else {
		rawUrl = fmt.Sprintf("https://api.%s.zsapi.net/", strings.ToLower(cloud))
	}

	// Parse the base URL
	baseURL, err := url.Parse(rawUrl)
	if err != nil {
		logger.Printf("[ERROR] error occurred while configuring the client: %v", err)
	}

	// Check if cache is disabled via environment variable
	cacheDisabled, _ := strconv.ParseBool(os.Getenv("ZSCALER_SDK_CACHE_DISABLED"))

	// Return the Config object
	return &config{
		baseURL:          baseURL,
		customerID:       customerID,
		backoffConf:      defaultBackoffConf,
		userAgent:        userAgent,
		cacheEnabled:     !cacheDisabled,
		cacheTtl:         time.Minute * 10,
		cacheCleanwindow: time.Minute * 8,
		cacheMaxSizeMB:   0,
		oauth2Credentials: &Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			VanityDomain: vanityDomain,
			PrivateKey:   privateKeyPath,
			Cloud:        cloud,
		},
	}, err
}

// loadCredentialsFromConfig Returns the credentials found in a config file
func loadCredentialsFromConfig(logger logger.Logger) (*CredentialsConfig, error) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	path := filepath.Join(dir, configPath)
	logger.Printf("[INFO]Loading configuration file at:%s", path)
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("Could not open credentials file, needs to contain one json object with keys: zpa_client_id, zpa_client_secret, zpa_customer_id, and zpa_cloud. " + err.Error())
	}
	configBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var config CredentialsConfig
	err = json.Unmarshal(configBytes, &config)
	if err != nil || config.ClientID == "" || config.ClientSecret == "" || config.CustomerID == "" || config.Cloud == "" {
		return nil, fmt.Errorf("could not parse credentials file, needs to contain one json object with keys: client_id, client_secret, customer_id, and cloud. error: %v", err)
	}
	return &config, nil
}
