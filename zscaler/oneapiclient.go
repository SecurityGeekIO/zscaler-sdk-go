package zscaler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/cache"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/logger"
	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type contextKey string

func (c contextKey) String() string {
	return "zscaler " + string(c)
}

var (
	// ContextAccessToken takes a string OAuth2 access token as authentication for the request.
	ContextAccessToken = contextKey("access_token")
)

const (
	VERSION               = "3.0.0"
	ZSCALER_CLIENT_ID     = "ZSCALER_CLIENT_ID"
	ZSCALER_CLIENT_SECRET = "ZSCALER_CLIENT_SECRET"
	ZSCALER_VANITY_DOMAIN = "ZSCALER_VANITY_DOMAIN"
	ZSCALER_PRIVATE_KEY   = "ZSCALER_PRIVATE_KEY"
	ZSCALER_CLOUD         = "ZSCALER_CLOUD"
)

// AuthToken represents the OAuth2 authentication token and its expiration time.
type AuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Expiry      time.Time
}

// Configuration struct holds the config for ZIA, ZPA, and common fields like HTTPClient and AuthToken.
type Configuration struct {
	UserAgent      string `json:"userAgent,omitempty"`
	Debug          bool   `json:"debug,omitempty"`
	HTTPClient     *http.Client
	UserAgentExtra string
	Context        context.Context
	Zscaler        struct {
		Client struct {
			ClientID      string     `yaml:"clientId" envconfig:"ZSCALER_CLIENT_ID"`
			ClientSecret  string     `yaml:"clientSecret" envconfig:"ZSCALER_CLIENT_SECRET"`
			VanityDomain  string     `yaml:"vanityDomain" envconfig:"ZSCALER_VANITY_DOMAIN"`
			Cloud         string     `yaml:"cloud" envconfig:"ZSCALER_CLOUD"`
			CustomerID    string     `yaml:"customerId" envconfig:"ZPA_CUSTOMER_ID"`
			MicrotenantID string     `yaml:"microtenantId" envconfig:"ZPA_MICROTENANT_ID"`
			PrivateKey    string     `yaml:"privateKey" envconfig:"ZSCALER_PRIVATE_KEY"`
			AuthToken     *AuthToken `yaml:"authToken"`
			AccessToken   *AuthToken `yaml:"accessToken"`
			SandboxToken  string     `yaml:"sandboxToken" envconfig:"ZSCALER_SANDBOX_TOKEN"`
			SandboxCloud  string     `yaml:"sandboxCloud" envconfig:"ZSCALER_SANDBOX_CLOUD"`
			Cache         struct {
				Enabled               bool          `yaml:"enabled" envconfig:"ZSCALER_CLIENT_CACHE_ENABLED"`
				DefaultTtl            time.Duration `yaml:"defaultTtl" envconfig:"ZSCALER_CLIENT_CACHE_DEFAULT_TTL"`
				DefaultTti            time.Duration `yaml:"defaultTti" envconfig:"ZSCALER_CLIENT_CACHE_DEFAULT_TTI"`
				DefaultCacheMaxSizeMB int64         `yaml:"defaultTti" envconfig:"ZSCALER_CLIENT_CACHE_DEFAULT_SIZE"`
			}
			Proxy struct {
				Port     int32  `yaml:"port" envconfig:"ZSCALER_CLIENT_PROXY_PORT"`
				Host     string `yaml:"host" envconfig:"ZSCALER_CLIENT_PROXY_HOST"`
				Username string `yaml:"username" envconfig:"ZSCALER_CLIENT_PROXY_USERNAME"`
				Password string `yaml:"password" envconfig:"ZSCALER_CLIENT_PROXY_PASSWORD"`
			}
			ConnectionTimeout int64 `yaml:"connectionTimeout" envconfig:"ZSCALER_CLIENT_CONNECTION_TIMEOUT"`
			RequestTimeout    int64 `yaml:"requestTimeout" envconfig:"ZSCALER_CLIENT_REQUEST_TIMEOUT"`
			RateLimit         struct {
				MaxRetries int32 `yaml:"maxRetries" envconfig:"ZSCALER_CLIENT_RATE_LIMIT_MAX_RETRIES"`
				MaxBackoff int64 `yaml:"maxBackoff" envconfig:"ZSCALER_CLIENT_RATE_LIMIT_MAX_BACKOFF"`
			}
			HttpClient *http.Client
		}
		Testing struct {
			DisableHttpsCheck bool `yaml:"disableHttpsCheck" envconfig:"ZSCALER_TESTING_DISABLE_HTTPS_CHECK"`
		}
	}
	PrivateKeySigner jose.Signer
	CacheManager     cache.Cache
}

// NewConfiguration is the main configuration function, implementing the ConfigSetter pattern.
func NewConfiguration(conf ...ConfigSetter) (*Configuration, error) {
	cfg := &Configuration{
		UserAgent: fmt.Sprintf("zscaler-sdk-go/%s golang/%s %s/%s", "3.0.0", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		Debug:     false,
		Context:   context.Background(), // Set default context
	}

	// Initialize cache
	if cfg.Zscaler.Client.Cache.DefaultTtl == 0 {
		cfg.Zscaler.Client.Cache.DefaultTtl = time.Minute * 10
	}

	if cfg.Zscaler.Client.Cache.DefaultTti == 0 {
		cfg.Zscaler.Client.Cache.DefaultTti = time.Minute * 8
	}

	cche, err := cache.NewCache(time.Duration(cfg.Zscaler.Client.Cache.DefaultTtl), time.Duration(cfg.Zscaler.Client.Cache.DefaultTti), int(cfg.Zscaler.Client.Cache.DefaultCacheMaxSizeMB))
	if err != nil {
		cche = cache.NewNopCache()
	}
	cfg.CacheManager = cche

	cfg.Zscaler.Testing.DisableHttpsCheck = false

	cfg = readConfigFromSystem(*cfg)
	cfg = readConfigFromEnvironment(*cfg)

	// Apply each ConfigSetter function.
	for _, confSetter := range conf {
		confSetter(cfg)
	}

	// UserAgentExtra gets added if provided.
	if cfg.UserAgentExtra != "" {
		cfg.UserAgent = fmt.Sprintf("%s %s", cfg.UserAgent, cfg.UserAgentExtra)
	}

	ctx := context.WithValue(
		context.Background(),
		ContextAccessToken,
		cfg.Zscaler.Client.AuthToken.AccessToken,
	)
	cfg.Context = ctx

	return cfg, nil
}

// Authenticate performs OAuth2 authentication and retrieves an AuthToken.
func Authenticate(ctx context.Context, cfg *Configuration, l logger.Logger) (*AuthToken, error) {
	creds := cfg.Zscaler.Client

	if creds.ClientID == "" || (creds.ClientSecret == "" && creds.PrivateKey == "") {
		return nil, errors.New("no client credentials were provided")
	}

	// If private key is provided, use JWT-based authentication.
	if creds.PrivateKey != "" {
		return authenticateWithCert(cfg)
	}

	// Determine the OAuth2 provider URL based on the cloud parameter.
	var authUrl string
	if creds.Cloud == "" || strings.EqualFold(creds.Cloud, "PRODUCTION") {
		authUrl = fmt.Sprintf("https://%s.zslogin.net/oauth2/v1/token", creds.VanityDomain)
	} else {
		authUrl = fmt.Sprintf("https://%s.zslogin%s.net/oauth2/v1/token", creds.VanityDomain, strings.ToLower(creds.Cloud))
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_secret", creds.ClientSecret)
	data.Set("client_id", creds.ClientID)
	data.Set("audience", "https://api.zscaler.com")

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", authUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to sign in the user %s: %v", creds.ClientID, err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", cfg.UserAgent)
	// start := time.Now()
	reqID := uuid.NewString()
	logger.LogRequest(l, req, reqID, nil, false)
	resp, err := cfg.Zscaler.Client.HttpClient.Do(req)
	// logger.LogResponse(l, resp, start, reqID)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to sign in the user %s, err: %v", creds.ClientID, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to sign in the user %s, err: %v", creds.ClientID, err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("[ERROR] Failed to sign in the user %s, got http status: %d, response body: %s", creds.ClientID, resp.StatusCode, respBody)
	}

	var token AuthToken
	err = json.Unmarshal(respBody, &token)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to sign in the user %s, err: %v", creds.ClientID, err)
	}

	// Calculate token expiration time
	token.Expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return &token, nil
}

// authenticateWithCert performs JWT-based authentication using a private key.
func authenticateWithCert(cfg *Configuration) (*AuthToken, error) {
	creds := cfg.Zscaler.Client

	if creds.ClientID == "" || creds.PrivateKey == "" {
		return nil, errors.New("client ID or private key is missing")
	}

	// Create the JWT payload.
	privateKeyData, err := os.ReadFile(creds.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error reading private key: %v", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": creds.ClientID,
		"sub": creds.ClientID,
		"aud": "https://api.zscaler.com",
		"exp": time.Now().Add(10 * time.Minute).Unix(),
	})

	assertion, err := token.SignedString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error signing JWT: %v", err)
	}

	formData := url.Values{
		"grant_type":            {"client_credentials"},
		"client_assertion":      {assertion},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"audience":              {"https://api.zscaler.com"},
	}

	// Determine the OAuth2 provider URL based on the cloud parameter.
	var authUrl string
	if creds.Cloud == "" || strings.EqualFold(creds.Cloud, "PRODUCTION") {
		authUrl = fmt.Sprintf("https://%s.zslogin.net/oauth2/v1/token", creds.VanityDomain)
	} else {
		authUrl = fmt.Sprintf("https://%s.zslogin%s.net/oauth2/v1/token", creds.VanityDomain, strings.ToLower(creds.Cloud))
	}

	// Make the POST request.
	resp, err := cfg.Zscaler.Client.HttpClient.PostForm(authUrl, formData)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Parse the response.
	var tokenResponse AuthToken
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &tokenResponse, nil
}

// getServiceHTTPClient returns the appropriate http client for the current service
func (client *Client) getServiceHTTPClient(endpoint string) *http.Client {
	service := detectServiceType(endpoint)
	switch service {
	case "zpa":
		return client.ZPAHTTPClient
	case "zia":
		return client.ZIAHTTPClient
	case "zcc":
		return client.ZCCHTTPClient
	default:
		return client.HTTPClient
	}
}

func detectServiceType(endpoint string) string {
	path := strings.TrimPrefix(endpoint, "/")
	// Detect the service type based on the endpoint prefix
	if strings.HasPrefix(path, "zia") {
		return "zia"
	} else if strings.HasPrefix(path, "zpa") {
		return "zpa"
	} else if strings.HasPrefix(endpoint, "/zcc") {
		return "zcc"
	}

	panic("unsupported service")
}

// GetAPIBaseURL gets the appropriate base url based on the cloud and sandbox mode.
func GetAPIBaseURL(cloud string, sandboxEnabled bool) string {
	baseURL := "https://api.zsapi.net"
	if cloud != "" && !strings.EqualFold(cloud, "PRODUCTION") {
		baseURL = fmt.Sprintf("https://api.%s.zsapi.net", strings.ToLower(cloud))
	}
	if sandboxEnabled {
		// Return the Sandbox base URL
		return fmt.Sprintf("https://csbapi.%s.net/zscsb/submit", strings.ToLower(cloud))
	}

	return baseURL
}

func readConfigFromFile(location string, c Configuration) (*Configuration, error) {
	yamlConfig, err := os.ReadFile(location)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlConfig, &c)
	if err != nil {
		return nil, err
	}
	return &c, err
}

func readConfigFromSystem(c Configuration) *Configuration {
	currUser, err := user.Current()
	if err != nil {
		return &c
	}
	if currUser.HomeDir == "" {
		return &c
	}
	conf, err := readConfigFromFile(currUser.HomeDir+"/.zscaler/zscaler.yaml", c)
	if err != nil {
		return &c
	}
	return conf
}

func readConfigFromEnvironment(c Configuration) *Configuration {
	err := envconfig.Process("zscaler", &c)
	if err != nil {
		fmt.Println("error parsing")
		return &c
	}
	return &c
}

type ConfigSetter func(*Configuration)

// ConfigSetter type defines a function that modifies a Config struct.
// WithClientID sets the ClientID in the Config.
func WithClientID(clientID string) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.ClientID = clientID
	}
}

// WithClientSecret sets the ClientSecret in the Config.
func WithClientSecret(clientSecret string) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.ClientSecret = clientSecret
	}
}

// WithOauth2ProviderUrl sets the Oauth2ProviderUrl in the Config.
func WithVanityDomain(domain string) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.VanityDomain = domain
	}
}

func WithZscalerCloud(cloud string) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.Cloud = cloud
	}
}

// WithSandboxToken is a ConfigSetter that sets the Sandbox token for the Zscaler Client.
func WithSandboxToken(token string) ConfigSetter {
	return func(cfg *Configuration) {
		cfg.Zscaler.Client.SandboxToken = token
	}
}

func WithSandboxCloud(sandboxCloud string) ConfigSetter {
	return func(cfg *Configuration) {
		cfg.Zscaler.Client.SandboxCloud = sandboxCloud
	}
}

func WithZPACustomerID(customerID string) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.CustomerID = customerID
	}
}

func WithZPAMicrotenantID(microtenantID string) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.MicrotenantID = microtenantID
	}
}

func WithCache(cache bool) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.Cache.Enabled = cache
	}
}

func WithCacheManager(cacheManager cache.Cache) ConfigSetter {
	return func(c *Configuration) {
		c.CacheManager = cacheManager
	}
}

func WithCacheTtl(i time.Duration) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.Cache.DefaultTtl = i
	}
}

func WithCacheTti(i time.Duration) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.Cache.DefaultTti = i
	}
}

// WithHttpClient sets the HttpClient in the Config.
func WithHttpClient(httpClient *http.Client) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.HttpClient = httpClient
	}
}

func WithProxyPort(i int32) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.Proxy.Port = i
	}
}

func WithProxyHost(host string) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.Proxy.Host = host
	}
}

func WithProxyUsername(username string) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.Proxy.Username = username
	}
}

func WithProxyPassword(pass string) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.Proxy.Password = pass
	}
}

func WithTestingDisableHttpsCheck(httpsCheck bool) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Testing.DisableHttpsCheck = httpsCheck
	}
}

func WithRequestTimeout(requestTimeout int64) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.RequestTimeout = requestTimeout
	}
}

func WithConnectionTimeout(i int64) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.ConnectionTimeout = i
	}
}

func WithRateLimitMaxRetries(maxRetries int32) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.RateLimit.MaxRetries = maxRetries
	}
}

func WithRateLimitMaxBackOff(maxBackoff int64) ConfigSetter {
	return func(c *Configuration) {
		c.Zscaler.Client.RateLimit.MaxBackoff = maxBackoff
	}
}

// WithUserAgent sets the UserAgent in the Config.
func WithUserAgentExtra(userAgent string) ConfigSetter {
	return func(c *Configuration) {
		c.UserAgentExtra = userAgent
	}
}

func WithDebug(debug bool) ConfigSetter {
	return func(c *Configuration) {
		c.Debug = debug
		// Automatically set the environment variables if debug is enabled
		if debug {
			_ = os.Setenv("ZSCALER_SDK_LOG", "true")
			_ = os.Setenv("ZSCALER_SDK_VERBOSE", "true")
		}
	}
}

// WithPrivateKey sets private key key. Can be either a path to a private key or private key itself.
func WithPrivateKey(privateKey string) ConfigSetter {
	return func(c *Configuration) {
		if fileExists(privateKey) {
			content, err := os.ReadFile(privateKey)
			if err != nil {
				fmt.Printf("failed to read from provided private key file path: %v", err)
			}
			c.Zscaler.Client.PrivateKey = string(content)
		} else {
			c.Zscaler.Client.PrivateKey = privateKey
		}
	}
}

func WithPrivateKeySigner(signer jose.Signer) ConfigSetter {
	return func(c *Configuration) {
		c.PrivateKeySigner = signer
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) || errors.Is(err, syscall.ENAMETOOLONG) {
			return false
		}
		fmt.Println("can not get information about the file containing private key, using provided value as the key itself")
		return false
	}
	return !info.IsDir()
}
