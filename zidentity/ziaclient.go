package zidentity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/cache"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/logger"
	"github.com/google/uuid"
)

// Performs the HTTP request and manages caching and token refresh logic.
// Performs the HTTP request and manages caching and token refresh logic.
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

	// Perform the HTTP request
	resp, err := c.HTTPClient.Do(req)
	logger.LogResponse(c.Logger, resp, start, reqID)
	if err != nil {
		return resp, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	// Handle 401 Unauthorized or token expiration cases for OAuth2
	if resp.StatusCode == http.StatusUnauthorized {
		// Retry with refreshed token by calling Authenticate directly
		authToken, err := Authenticate(c.oauth2Credentials)
		if err != nil {
			return nil, err
		}
		c.oauth2Credentials.Zscaler.Client.AuthToken = authToken

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.oauth2Credentials.Zscaler.Client.AuthToken.AccessToken))
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

// GenericRequest handles a generic HTTP request.
// GenericRequest handles a generic HTTP request.
func (c *Client) GenericRequest(baseUrl, endpoint, method string, body io.Reader, urlParams url.Values, contentType string) ([]byte, error) {
	if contentType == "" {
		contentType = contentTypeJSON
	}

	var req *http.Request
	var resp *http.Response
	var err error
	params := ""
	if urlParams != nil {
		params = urlParams.Encode()
	}
	if strings.Contains(endpoint, "?") && params != "" {
		endpoint += "&" + params
	} else if params != "" {
		endpoint += "?" + params
	}
	fullURL := fmt.Sprintf("%s%s", baseUrl, endpoint)
	isSandboxRequest := baseUrl == c.GetSandboxURL()
	req, err = http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}

	if !isSandboxRequest {
		// Set the OAuth2 token in the request
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.oauth2Credentials.Zscaler.Client.AuthToken.AccessToken))
	}

	reqID := uuid.New().String()
	start := time.Now()
	logger.LogRequest(c.Logger, req, reqID, nil, !isSandboxRequest)

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

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyResp, nil
}

// Request sends a request using the client's URL.
func (c *Client) Request(endpoint, method string, data []byte, contentType string) ([]byte, error) {
	return c.GenericRequest(c.URL, endpoint, method, bytes.NewReader(data), nil, contentType)
}

// Allows refreshing the cache.
func (client *Client) WithFreshCache() {
	client.freshCache = true
}

// Create sends a POST request to create an object.
func (c *Client) Create(endpoint string, o interface{}) (interface{}, error) {
	if o == nil {
		return nil, errors.New("tried to create with a nil payload not a Struct")
	}
	t := reflect.TypeOf(o)
	if t.Kind() != reflect.Struct {
		return nil, errors.New("tried to create with a " + t.Kind().String() + " not a Struct")
	}
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	resp, err := c.Request(endpoint, "POST", data, "application/json")
	if err != nil {
		return nil, err
	}
	if len(resp) > 0 {
		responseObject := reflect.New(t).Interface()
		err = json.Unmarshal(resp, &responseObject)
		if err != nil {
			return nil, err
		}
		id := reflect.Indirect(reflect.ValueOf(responseObject)).FieldByName("ID")

		c.Logger.Printf("Created Object with ID %v", id)
		return responseObject, nil
	} else {
		// in case of 204 no content
		return nil, nil
	}
}

// Read ...
func (c *Client) Read(endpoint string, o interface{}) error {
	contentType := c.GetContentType()
	resp, err := c.Request(endpoint, "GET", nil, contentType)
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, o)
	if err != nil {
		return err
	}

	return nil
}

// UpdateWithPut sends an update (PUT request) with the given object.
func (c *Client) UpdateWithPut(endpoint string, o interface{}) (interface{}, error) {
	return c.updateGeneric(endpoint, o, "PUT", "application/json")
}

// Update sends an update (PATCH request) with the given object.
func (c *Client) Update(endpoint string, o interface{}) (interface{}, error) {
	return c.updateGeneric(endpoint, o, "PATCH", "application/merge-patch+json")
}

// General method to update an object using the specified HTTP method.
func (c *Client) updateGeneric(endpoint string, o interface{}, method, contentType string) (interface{}, error) {
	if o == nil {
		return nil, errors.New("tried to update with a nil payload not a Struct")
	}
	t := reflect.TypeOf(o)
	if t.Kind() != reflect.Struct {
		return nil, errors.New("tried to update with a " + t.Kind().String() + " not a Struct")
	}
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	resp, err := c.Request(endpoint, method, data, contentType)
	if err != nil {
		return nil, err
	}

	responseObject := reflect.New(t).Interface()
	err = json.Unmarshal(resp, &responseObject)
	return responseObject, err
}

// Delete sends a DELETE request to the specified endpoint.
func (c *Client) Delete(endpoint string) error {
	_, err := c.Request(endpoint, "DELETE", nil, "application/json")
	if err != nil {
		return err
	}
	return nil
}

// BulkDelete sends a POST request for bulk deletion.
func (c *Client) BulkDelete(endpoint string, payload interface{}) (*http.Response, error) {
	if payload == nil {
		return nil, errors.New("tried to delete with a nil payload, expected a struct")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.Request(endpoint, "POST", data, "application/json")
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		c.Logger.Printf("[DEBUG] Bulk delete successful with 204 No Content")
		return &http.Response{StatusCode: 204}, nil
	}

	return &http.Response{StatusCode: 200}, fmt.Errorf("unexpected response: %s", string(resp))
}
