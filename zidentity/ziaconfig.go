package zidentity

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/cache"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/logger"
	rl "github.com/SecurityGeekIO/zscaler-sdk-go/v2/ratelimiter"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	maxIdleConnections  int = 40
	requestTimeout      int = 60
	contentTypeJSON         = "application/json"
	MaxNumOfRetries         = 100
	RetryWaitMaxSeconds     = 20
	RetryWaitMinSeconds     = 5
	loggerPrefix            = "zia-logger: "
)

// Client defines the ZIA client structure.
type Client struct {
	sync.Mutex
	cloud             string
	URL               string
	HTTPClient        *http.Client
	Logger            logger.Logger
	UserAgent         string
	freshCache        bool
	cacheEnabled      bool
	cache             cache.Cache
	cacheTtl          time.Duration
	cacheCleanwindow  time.Duration
	cacheMaxSizeMB    int
	rateLimiter       *rl.RateLimiter
	useOneAPI         bool
	oauth2Credentials *Configuration
	stopTicker        chan bool
	Config            *Config
}

// NewOneAPIClient creates a new ZIA Client using OAuth2 authentication.
func NewOneAPIClient(config *Configuration, service string) (*Service, error) {
	logger := logger.GetDefaultLogger(loggerPrefix)
	rateLimiter := rl.NewRateLimiter(2, 1, 1, 1)
	// Pass the config to getHTTPClient so it can access proxy settings
	httpClient := getHTTPClient(logger, rateLimiter, config)

	// Build the API endpoint based on the service and cloud parameters
	url := GetAPIEndpoint(service, config.Zscaler.Client.Cloud, false) // Pass the service explicitly

	// Adjust rate limits if ZCC is selected
	if service == "zcc" {
		rateLimiter = rl.NewRateLimiter(100, 3, 3, 86400) // ZCC has distinct limits
	}

	// Create the client configuration
	cli := &Client{
		cloud:             config.Zscaler.Client.Cloud,
		HTTPClient:        httpClient,
		URL:               url,
		Logger:            logger,
		UserAgent:         config.UserAgent,
		cacheEnabled:      config.Zscaler.Client.Cache.Enabled,
		cacheTtl:          time.Minute * 10,
		cacheCleanwindow:  time.Minute * 8,
		cacheMaxSizeMB:    0,
		rateLimiter:       rateLimiter,
		useOneAPI:         true,
		oauth2Credentials: config,
		stopTicker:        make(chan bool),
	}

	// Initialize cache
	cche, err := cache.NewCache(cli.cacheTtl, cli.cacheCleanwindow, cli.cacheMaxSizeMB)
	if err != nil {
		cche = cache.NewNopCache()
	}
	cli.cache = cche

	// Start token renewal ticker
	cli.startTokenRenewalTicker()

	// Return the service directly
	return NewService(cli), nil
}

// GetSandboxURL retrieves the sandbox URL for the ZIA service.
func (c *Client) GetSandboxURL() string {
	return "https://csbapi." + c.cloud + ".net"
}

// GetSandboxToken retrieves the sandbox token from the environment.
func (c *Client) GetSandboxToken() string {
	return os.Getenv("ZIA_SANDBOX_TOKEN")
}

// startTokenRenewalTicker starts a ticker to renew the token before it expires.
// startTokenRenewalTicker starts a ticker to renew the token before it expires.
func (c *Client) startTokenRenewalTicker() {
	if c.useOneAPI {
		tokenExpiry := c.oauth2Credentials.Zscaler.Client.AuthToken.Expiry
		renewalInterval := time.Until(tokenExpiry) - (time.Minute * 1) // Renew 1 minute before expiration

		if renewalInterval > 0 {
			ticker := time.NewTicker(renewalInterval)
			go func() {
				for {
					select {
					case <-ticker.C:
						// Refresh the token
						authToken, err := Authenticate(c.oauth2Credentials)
						if err != nil {
							c.Logger.Printf("[ERROR] Failed to renew OAuth2 token: %v", err)
						} else {
							c.oauth2Credentials.Zscaler.Client.AuthToken = authToken
							c.Logger.Printf("[INFO] OAuth2 token successfully renewed")
							// Reset the ticker for the next renewal
							renewalInterval = time.Until(authToken.Expiry) - (time.Minute * 1)
							ticker.Reset(renewalInterval)
						}
					case <-c.stopTicker:
						ticker.Stop()
						return
					}
				}
			}()
		}
	}
}

// getHTTPClient sets up the retryable HTTP client with backoff and retry policies.
func getHTTPClient(l logger.Logger, rateLimiter *rl.RateLimiter, cfg *Configuration) *http.Client {
	retryableClient := retryablehttp.NewClient()
	retryableClient.RetryWaitMin = time.Second * time.Duration(RetryWaitMinSeconds)
	retryableClient.RetryWaitMax = time.Second * time.Duration(RetryWaitMaxSeconds)
	retryableClient.RetryMax = MaxNumOfRetries

	// Configure backoff and retry policies
	retryableClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		if resp != nil {
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
				retryAfter := getRetryAfter(resp, l)
				if retryAfter > 0 {
					return retryAfter
				}
			}
			if resp.Request != nil {
				wait, d := rateLimiter.Wait(resp.Request.Method)
				if wait {
					return d
				}
				return 0
			}
		}
		mult := math.Pow(2, float64(attemptNum)) * float64(min)
		sleep := time.Duration(mult)
		if float64(sleep) != mult || sleep > max {
			sleep = max
		}
		return sleep
	}
	retryableClient.CheckRetry = checkRetry
	retryableClient.Logger = l
	retryableClient.HTTPClient.Timeout = time.Duration(requestTimeout) * time.Second

	// Configure proxy settings from configuration
	proxyFunc := http.ProxyFromEnvironment // Default behavior (uses system/env variables)
	if cfg.Zscaler.Client.Proxy.Host != "" {
		// Include username and password if provided
		proxyURLString := fmt.Sprintf("http://%s:%d", cfg.Zscaler.Client.Proxy.Host, cfg.Zscaler.Client.Proxy.Port)
		if cfg.Zscaler.Client.Proxy.Username != "" && cfg.Zscaler.Client.Proxy.Password != "" {
			// URL-encode the username and password
			proxyAuth := url.UserPassword(cfg.Zscaler.Client.Proxy.Username, cfg.Zscaler.Client.Proxy.Password)
			proxyURLString = fmt.Sprintf("http://%s@%s:%d", proxyAuth.String(), cfg.Zscaler.Client.Proxy.Host, cfg.Zscaler.Client.Proxy.Port)
		}

		proxyURL, err := url.Parse(proxyURLString)
		if err == nil {
			proxyFunc = http.ProxyURL(proxyURL) // Use custom proxy from configuration
		} else {
			l.Printf("[ERROR] Invalid proxy URL: %v", err)
		}
	}

	// Setup transport with custom proxy, if applicable, and check for HTTPS certificate check disabling
	transport := &http.Transport{
		Proxy:               proxyFunc,
		MaxIdleConnsPerHost: maxIdleConnections,
	}

	// Disable HTTPS check if the configuration requests it
	if cfg.Zscaler.Testing.DisableHttpsCheck {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // This disables HTTPS certificate validation
		}
		l.Printf("[INFO] HTTPS certificate validation is disabled (testing mode).")
	}

	retryableClient.HTTPClient.Transport = transport
	return retryableClient.StandardClient()
}

func (c *Client) GetContentType() string {
	return contentTypeJSON
}

// getRetryAfter checks for the Retry-After header or response body to determine retry wait time.
func getRetryAfter(resp *http.Response, l logger.Logger) time.Duration {
	if s := resp.Header.Get("Retry-After"); s != "" {
		if sleep, err := strconv.ParseInt(s, 10, 64); err == nil {
			l.Printf("[INFO] got Retry-After from header: %s\n", s)
			return time.Second * time.Duration(sleep)
		} else {
			dur, err := time.ParseDuration(s)
			if err == nil {
				return dur
			}
			l.Printf("[INFO] error parsing Retry-After header: %s\n", err)
		}
	}
	return 0
}

// getRetryOnStatusCodes return a list of http status codes we want to apply retry on.
// Return empty slice to enable retry on all connection & server errors.
// Or return []int{429}  to retry on only TooManyRequests error.
func getRetryOnStatusCodes() []int {
	return []int{http.StatusTooManyRequests}
}

// checkRetry defines the retry logic based on status codes or response body errors.
func checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if resp != nil && containsInt(getRetryOnStatusCodes(), resp.StatusCode) {
		return true, nil
	}
	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}

func containsInt(codes []int, code int) bool {
	for _, a := range codes {
		if a == code {
			return true
		}
	}
	return false
}

// WithCache enables or disables caching.
func (c *Client) WithCache(cache bool) {
	c.cacheEnabled = cache
}

// WithCacheTtl sets the time-to-live (TTL) for cache.
func (c *Client) WithCacheTtl(i time.Duration) {
	c.cacheTtl = i
	c.Lock()
	c.cache.Close()
	cche, err := cache.NewCache(i, c.cacheCleanwindow, c.cacheMaxSizeMB)
	if err != nil {
		cche = cache.NewNopCache()
	}
	c.cache = cche
	c.Unlock()
}

func (c *Client) WithCacheCleanWindow(i time.Duration) {
	c.cacheCleanwindow = i
	c.Lock()
	c.cache.Close()
	cche, err := cache.NewCache(c.cacheTtl, i, c.cacheMaxSizeMB)
	if err != nil {
		cche = cache.NewNopCache()
	}
	c.cache = cche
	c.Unlock()
}
