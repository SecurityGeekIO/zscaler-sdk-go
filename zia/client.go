package zia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/cache"
	cmmon "github.com/SecurityGeekIO/zscaler-sdk-go/v2/common"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/logger"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia/services/common"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
)

func (c *Client) do(req *http.Request, start time.Time, reqID string) (*http.Response, error) {
	key := cache.CreateCacheKey(req)
	if c.cacheEnabled {
		if req.Method != http.MethodGet {
			c.cache.Delete(key)
			c.cache.ClearAllKeysWithPrefix(strings.Split(key, "?")[0])
		}
		resp := c.cache.Get(key)
		inCache := resp != nil
		if c.freshCache {
			c.cache.Delete(key)
			inCache = false
			c.freshCache = false
		}
		if inCache {
			c.Logger.Printf("[INFO] served from cache, key:%s\n", key)
			return resp, nil
		}
	}

	// Ensure the session is valid before making the request
	err := c.checkSession()
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	logger.LogResponse(c.Logger, resp, start, reqID)
	if err != nil {
		return resp, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	// Fallback check for SESSION_NOT_VALID
	if resp.StatusCode == http.StatusUnauthorized || strings.Contains(string(body), "SESSION_NOT_VALID") {
		// Refresh session and retry
		err := c.refreshSession()
		if err != nil {
			return nil, err
		}
		req.Header.Set("JSessionID", c.session.JSessionID)
		resp, err = c.HTTPClient.Do(req)
		logger.LogResponse(c.Logger, resp, start, reqID)
		if err != nil {
			return resp, err
		}
	}

	if c.cacheEnabled && resp.StatusCode >= 200 && resp.StatusCode <= 299 && req.Method == http.MethodGet {
		c.Logger.Printf("[INFO] saving to cache, key:%s\n", key)
		c.cache.Set(key, cache.CopyResponse(resp))
	}

	return resp, nil
}

func (c *Client) NewRequestDoGeneric(baseUrl, endpoint, method string, options, body, v interface{}, contentType string, infraOptions ...cmmon.Option) (*http.Response, error) {
	if contentType == "" {
		contentType = contentTypeJSON
	}

	var req *http.Request
	var resp *http.Response
	var err error
	// Handle query parameters from options and any additional logic
	if options == nil {
		options = struct{}{}
	}
	var params string
	if options != nil {
		switch opt := options.(type) {
		case url.Values:
			params = opt.Encode()
		default:
			q, err := query.Values(options)
			if err != nil {
				return nil, err
			}
			params = q.Encode()
		}
	}

	if strings.Contains(endpoint, "?") && params != "" {
		endpoint += "&" + params
	} else if params != "" {
		endpoint += "?" + params
	}
	fullURL := fmt.Sprintf("%s%s", baseUrl, endpoint)
	isSandboxRequest := baseUrl == GetSandboxURL(c.cloud)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err = http.NewRequest(method, fullURL, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}
	var otherHeaders map[string]string
	if !isSandboxRequest {
		err = c.checkSession()
		if err != nil {
			return nil, err
		}
		otherHeaders = map[string]string{"JSessionID": c.session.JSessionID}
	}
	reqID := uuid.New().String()
	start := time.Now()
	logger.LogRequest(c.Logger, req, reqID, otherHeaders, !isSandboxRequest)
	for retry := 1; retry <= 5; retry++ {
		resp, err = c.do(req, start, reqID)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode <= 299 {
			defer resp.Body.Close()
			break
		}

		resp.Body.Close()
		if resp.StatusCode > 299 && resp.StatusCode != http.StatusUnauthorized {
			return nil, checkErrorInResponse(resp, fmt.Errorf("api responded with code: %d", resp.StatusCode))
		}
	}

	respData, err := io.ReadAll(resp.Body)
	if err == nil {
		resp.Body = io.NopCloser(bytes.NewBuffer(respData))
	}

	if v != nil {
		if err := json.NewDecoder(bytes.NewBuffer(respData)).Decode(&v); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func (client *Client) NewRequestDo(method, path string, options, body, v interface{}, infraOptions ...cmmon.Option) (*http.Response, error) {
	return client.NewRequestDoGeneric(client.URL, path, method, options, body, v, contentTypeJSON, infraOptions...)
}

// Request ... // Needs to review this function.
func (c *Client) GenericRequest(baseUrl, endpoint, method string, body io.Reader, urlParams url.Values, contentType string) ([]byte, error) {
	return common.GenericRequest(c, baseUrl, endpoint, method, body, urlParams, contentType)
}

// Request ... // Needs to review this function.
func (c *Client) Request(endpoint, method string, data []byte, contentType string) ([]byte, error) {
	return c.GenericRequest(c.URL, endpoint, method, bytes.NewReader(data), nil, contentType)
}

func (client *Client) WithFreshCache() {
	client.freshCache = true
}

func (client *Client) GetBaseURL() string {
	return client.URL
}

func (client *Client) GetLogger() logger.Logger {
	return client.Logger
}

func (client *Client) GetCustomerID() string {
	return ""
}

func (client *Client) SetCustomerID(_ string) {
}

func (client *Client) GetCloud() string {
	return client.cloud
}
