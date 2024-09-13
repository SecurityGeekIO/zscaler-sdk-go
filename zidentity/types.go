package zidentity

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/cache"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/logger"
	rl "github.com/SecurityGeekIO/zscaler-sdk-go/v2/ratelimiter"
)

const (
	defaultTimeout  = 240 * time.Second
	contentTypeJSON = "application/json"
	ZPA_CUSTOMER_ID = "ZPA_CUSTOMER_ID"
	loggerPrefix    = "zpa-logger: "
	ZSCALER_CLOUD   = "ZSCALER_CLOUD"
	configPath      = ".zscaler/credentials.json"
)

type backoffConfig struct {
	Enabled             bool // Set to true to enable backoff and retry mechanism
	RetryWaitMinSeconds int  // Minimum time to wait
	RetryWaitMaxSeconds int  // Maximum time to wait
	MaxNumOfRetries     int  // Maximum number of retries
}

var defaultBackoffConf = &backoffConfig{
	Enabled:             true,
	MaxNumOfRetries:     100,
	RetryWaitMaxSeconds: 10,
	RetryWaitMinSeconds: 2,
}

// config contains all the configuration data for the API client
type config struct {
	sync.Mutex
	baseURL           *url.URL
	customerID        string
	backoffConf       *backoffConfig
	userAgent         string
	cacheEnabled      bool
	freshCache        bool
	cacheTtl          time.Duration
	cacheCleanwindow  time.Duration
	cacheMaxSizeMB    int
	oauth2Credentials *Credentials
}

type zidentityClient struct {
	config      *config
	cache       cache.Cache
	rateLimiter *rl.RateLimiter
	httpClient  *http.Client
	logger      logger.Logger
}

type CredentialsConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CustomerID   string `json:"zpa_customer_id"`
	Cloud        string `json:"cloud"`
}

type ErrorResponse struct {
	Response *http.Response
	Message  string
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("FAILED: %v, %v, %d, %v, %v", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Response.Status, r.Message)
}
