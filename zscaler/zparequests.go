package zscaler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/cache"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/logger"
	"github.com/google/uuid"
)

const (
	mgmtConfig = "/mgmtconfig/v1/admin/customers/"
)

func (client *Client) NewRequestDo(method, url string, options, body, v interface{}) (*http.Response, error) {
	// Adjusting to match the ExecuteRequest signature
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	// Execute the request using ExecuteRequest and capture the request object (req)
	_, req, err := client.ExecuteRequest(method, url, bodyReader, nil, contentTypeJSON)
	if err != nil {
		return nil, err
	}

	// Create cache key using the actual request
	key := cache.CreateCacheKey(req)

	if client.cacheEnabled {
		if method != http.MethodGet {
			client.cache.Delete(key)
			client.cache.ClearAllKeysWithPrefix(strings.Split(key, "?")[0])
		}
		resp := client.cache.Get(key)
		inCache := resp != nil
		if client.freshCache {
			client.cache.Delete(key)
			inCache = false
			client.freshCache = false
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
			client.Logger.Printf("[INFO] served from cache, key:%s\n", key)
			return resp, nil
		}
	}

	// Call the custom request handler
	resp, err := client.newRequestDoCustom(method, url, options, body)
	if err != nil {
		return resp, err
	}

	// Cache logic for successful GET requests
	if client.cacheEnabled && resp.StatusCode >= 200 && resp.StatusCode <= 299 && method == http.MethodGet && v != nil && reflect.TypeOf(v).Kind() != reflect.Slice {
		d, err := json.Marshal(v)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewReader(d))
			client.Logger.Printf("[INFO] saving to cache, key:%s\n", key)
			client.cache.Set(key, cache.CopyResponse(resp))
		} else {
			client.Logger.Printf("[ERROR] saving to cache error:%s, key:%s\n", err, key)
		}
	}
	return resp, nil
}

func (client *Client) newRequestDoCustom(method, urlStr string, body, v interface{}) (*http.Response, error) {
	// Authenticate before executing the request
	err := client.authenticate()
	if err != nil {
		return nil, err
	}

	// Use ExecuteRequest to handle the request
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	// Capture the three return values from ExecuteRequest
	reqBody, req, err := client.ExecuteRequest(method, urlStr, bodyReader, nil, contentTypeJSON)
	if err != nil {
		return nil, err
	}

	// Log the request
	reqID := uuid.NewString()
	start := time.Now()
	logger.LogRequest(client.Logger, req, reqID, nil, true)

	// Create a dummy HTTP response using the request body for response body
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(reqBody)),
	}

	logger.LogResponse(client.Logger, resp, start, reqID)

	// Check for specific HTTP status codes (401/403) and re-authenticate if necessary
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		err = client.authenticate()
		if err != nil {
			return nil, err
		}

		// Retry the request after re-authentication
		resp, err = client.do(req, v, start, reqID) // Remove reqBody from the arguments
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (client *Client) do(req *http.Request, v interface{}, start time.Time, reqID string) (*http.Response, error) {
	// Dynamically get the configured HTTP client to respect custom configurations
	httpClient := getHTTPClient(client.Logger, client.rateLimiter, client.oauth2Credentials)

	// Execute the HTTP request using the dynamically configured HTTP client
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Read the response body
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(respData)) // Reset body for future reads

	// Check for errors in the response
	if err := checkErrorInResponse(resp, nil); err != nil {
		return resp, err
	}

	// Decode the response if 'v' is provided
	if v != nil {
		if err := decodeJSON(respData, v); err != nil {
			return resp, err
		}
	}

	// Log the response
	logger.LogResponse(client.Logger, resp, start, reqID)

	// Unescape any HTML characters in the response
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

func (c *Client) GetCustomerID() string {
	return c.oauth2Credentials.Zscaler.Client.CustomerID
}

func (client *Client) GetFullPath(endpoint string) (string, error) {
	customerID := client.GetCustomerID()
	if customerID == "" {
		return "", fmt.Errorf("CustomerID is not set")
	}
	// Construct the full path with mgmtConfig and CustomerID
	return fmt.Sprintf("%s%s%s", mgmtConfig, customerID, endpoint), nil
}

/*
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
*/
