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
	"sync"
	"time"

	rl "github.com/SecurityGeekIO/zscaler-sdk-go/v2/ratelimiter"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"

	"github.com/google/go-querystring/query"
	"github.com/google/uuid"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/cache"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/logger"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/utils"
)

const (
	defaultTimeout  = 240 * time.Second
	contentTypeJSON = "application/json"
	ZPA_CUSTOMER_ID = "ZPA_CUSTOMER_ID"
	loggerPrefix    = "zpa-logger: "
	ZSCALER_CLOUD   = "ZSCALER_CLOUD"
	configPath      = ".zscaler/credentials.json"
)

var defaultBackoffConf = &BackoffConfig{
	Enabled:             true,
	MaxNumOfRetries:     100,
	RetryWaitMaxSeconds: 10,
	RetryWaitMinSeconds: 2,
}

type Client struct {
	Config *Config
	cache  cache.Cache
}

type BackoffConfig struct {
	Enabled             bool // Set to true to enable backoff and retry mechanism
	RetryWaitMinSeconds int  // Minimum time to wait
	RetryWaitMaxSeconds int  // Maximum time to wait
	MaxNumOfRetries     int  // Maximum number of retries
}

// Config contains all the configuration data for the API client
type Config struct {
	sync.Mutex
	BaseURL     *url.URL
	httpClient  *http.Client
	rateLimiter *rl.RateLimiter
	// The logger writer interface to write logging messages to. Defaults to standard out.
	Logger     logger.Logger
	CustomerID string
	// Backoff config
	BackoffConf       *BackoffConfig
	AuthToken         *AuthToken
	UserAgent         string
	cacheEnabled      bool
	freshCache        bool
	cacheTtl          time.Duration
	cacheCleanwindow  time.Duration
	cacheMaxSizeMB    int
	oauth2Credentials *Credentials
}

func (client *Client) WithFreshCache() {
	client.Config.freshCache = true
}

func (client *Client) NewRequestDo(method, url string, options, body, v interface{}) (*http.Response, error) {
	req, err := client.getRequest(method, url, options, body)
	if err != nil {
		return nil, err
	}
	key := cache.CreateCacheKey(req)
	if client.Config.cacheEnabled {
		if req.Method != http.MethodGet {
			// this will allow to remove resource from cache when PUT/DELETE/PATCH requests are called, which modifies the resource
			client.cache.Delete(key)
			// to avoid resources that GET url is not the same as DELETE/PUT/PATCH url, because of different query params.
			// example delete app segment has key url/<id>?forceDelete=true but GET has url/<id>, in this case we clean the whole cache entries with key prefix url/<id>
			client.cache.ClearAllKeysWithPrefix(strings.Split(key, "?")[0])
		}
		resp := client.cache.Get(key)
		inCache := resp != nil
		if client.Config.freshCache {
			client.cache.Delete(key)
			inCache = false
			client.Config.freshCache = false
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
			client.Config.Logger.Printf("[INFO] served from cache, key:%s\n", key)
			return resp, nil
		}
	}
	resp, err := client.newRequestDoCustom(method, url, options, body, v)
	if err != nil {
		return resp, err
	}
	if client.Config.cacheEnabled && resp.StatusCode >= 200 && resp.StatusCode <= 299 && req.Method == http.MethodGet && v != nil && reflect.TypeOf(v).Kind() != reflect.Slice {
		d, err := json.Marshal(v)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewReader(d))
			client.Config.Logger.Printf("[INFO] saving to cache, key:%s\n", key)
			client.cache.Set(key, cache.CopyResponse(resp))
		} else {
			client.Config.Logger.Printf("[ERROR] saving to cache error:%s, key:%s\n", err, key)
		}
	}
	return resp, nil
}

func (client *Client) authenticate() error {
	client.Config.Lock()
	defer client.Config.Unlock()

	if client.Config.AuthToken == nil || client.Config.AuthToken.AccessToken == "" || utils.IsTokenExpired(client.Config.AuthToken.AccessToken) {
		a, err := Authenticate(
			client.Config.oauth2Credentials,
			client.Config.GetHTTPClient(),
		)
		if err != nil {
			return err
		}
		client.Config.AuthToken = &AuthToken{
			TokenType:   a.TokenType,
			AccessToken: a.AccessToken,
		}
		return nil
	}
	return nil
}

func (client *Client) newRequestDoCustom(method, urlStr string, options, body, v interface{}) (*http.Response, error) {
	err := client.authenticate()
	if err != nil {
		return nil, err
	}
	req, err := client.newRequest(method, urlStr, options, body)
	if err != nil {
		return nil, err
	}
	reqID := uuid.NewString()
	start := time.Now()
	logger.LogRequest(client.Config.Logger, req, reqID, nil, true)
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

func (client *Client) injectMicrotentantID(body interface{}, q url.Values) url.Values {
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

func (client *Client) GetCustomerID() string {
	return client.Config.CustomerID
}

func (client *Client) getRequest(method, urlPath string, options, body interface{}) (*http.Request, error) {
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

	parsedPath.Path = client.Config.BaseURL.Path + parsedPath.Path
	u := client.Config.BaseURL.ResolveReference(parsedPath)

	// Handle query parameters from options and any additional logic
	if options == nil {
		options = struct{}{}
	}
	q, err := query.Values(options)
	if err != nil {
		return nil, err
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
func (client *Client) newRequest(method, urlPath string, options, body interface{}) (*http.Request, error) {
	if client.Config.AuthToken == nil || client.Config.AuthToken.AccessToken == "" {
		client.Config.Logger.Printf("[ERROR] Failed to signin the user\n")
		return nil, fmt.Errorf("failed to signin the user\n")
	}
	req, err := client.getRequest(method, urlPath, options, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.Config.AuthToken.AccessToken))
	req.Header.Add("Content-Type", "application/json")

	if client.Config.UserAgent != "" {
		req.Header.Add("User-Agent", client.Config.UserAgent)
	}

	return req, nil
}

type ErrorResponse struct {
	Response *http.Response
	Message  string
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("FAILED: %v, %v, %d, %v, %v", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Response.Status, r.Message)
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

func (client *Client) do(req *http.Request, v interface{}, start time.Time, reqID string) (*http.Response, error) {
	resp, err := client.Config.GetHTTPClient().Do(req)
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
	logger.LogResponse(client.Config.Logger, resp, start, reqID)
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

func (c *Config) GetHTTPClient() *http.Client {
	if c.httpClient == nil {
		if c.BackoffConf != nil && c.BackoffConf.Enabled {
			retryableClient := retryablehttp.NewClient()
			retryableClient.Logger = c.Logger
			retryableClient.RetryWaitMin = time.Second * time.Duration(c.BackoffConf.RetryWaitMinSeconds)
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
							c.Logger.Printf("[INFO] rate limiter wait duration:%s\n", duration.String())
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
			retryableClient.RetryWaitMax = time.Second * time.Duration(c.BackoffConf.RetryWaitMaxSeconds)
			retryableClient.RetryMax = c.BackoffConf.MaxNumOfRetries
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
func NewClient(clientID, clientSecret, privateKeyPath, customerID, vanityDomain, userAgent string, optionalCloud ...string) (c *Client, err error) {
	config, err := newConfig(clientID, clientSecret, privateKeyPath, customerID, vanityDomain, userAgent, optionalCloud...)
	if err != nil {
		return nil, err
	}

	// Setup the cache
	cche, err := cache.NewCache(config.cacheTtl, config.cacheCleanwindow, config.cacheMaxSizeMB)
	if err != nil {
		cche = cache.NewNopCache()
	}

	// Create and return the Client with the provided Config
	c = &Client{Config: config, cache: cche}
	return
}

func newConfig(clientID, clientSecret, privateKeyPath, customerID, vanityDomain, userAgent string, optionalCloud ...string) (*Config, error) {
	var logger logger.Logger = logger.GetDefaultLogger(loggerPrefix)

	// Fallback to environment variables if clientID, clientSecret, customerID, or userAgent are not provided
	if clientID == "" || (clientSecret == "" && privateKeyPath == "") || customerID == "" || userAgent == "" {
		clientID = os.Getenv(ZIDENTITY_CLIENT_ID)
		clientSecret = os.Getenv(ZIDENTITY_CLIENT_SECRET)
		privateKeyPath = os.Getenv(ZIDENTITY_PRIVATE_KEY)
		customerID = os.Getenv(ZPA_CUSTOMER_ID)
	}

	// Handle the optional cloud parameter
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
	if clientID == "" || clientSecret == "" || customerID == "" {
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
		rawUrl = "https://api.zsapi.net/zpa"
	} else {
		rawUrl = fmt.Sprintf("https://api.%s.zsapi.net/zpa", strings.ToLower(cloud))
	}

	// Parse the base URL
	baseURL, err := url.Parse(rawUrl)
	if err != nil {
		logger.Printf("[ERROR] error occurred while configuring the client: %v", err)
	}

	// Check if cache is disabled via environment variable
	cacheDisabled, _ := strconv.ParseBool(os.Getenv("ZSCALER_SDK_CACHE_DISABLED"))

	// Return the Config object
	return &Config{
		BaseURL:          baseURL,
		Logger:           logger,
		httpClient:       nil,
		CustomerID:       customerID,
		BackoffConf:      defaultBackoffConf,
		UserAgent:        userAgent,
		rateLimiter:      rl.NewRateLimiter(20, 10, 10, 10),
		cacheEnabled:     !cacheDisabled,
		cacheTtl:         time.Minute * 10,
		cacheCleanwindow: time.Minute * 8,
		cacheMaxSizeMB:   0,
		oauth2Credentials: &Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			VanityDomain: vanityDomain,
			PrivateKey:   privateKeyPath,
			UserAgent:    userAgent,
			Cloud:        cloud,
		},
	}, err
}

type CredentialsConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CustomerID   string `json:"customer_id"`
	Cloud        string `json:"cloud"`
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
