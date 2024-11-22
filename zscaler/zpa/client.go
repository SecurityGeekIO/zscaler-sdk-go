package zpa

// type Client struct {
// 	Config *Config
// 	cache  cache.Cache
// }

// NewClient returns a new client for the specified apiKey.
// func NewClient(config *Config) (c *Client) {
// 	if config == nil {
// 		config, _ = NewConfig("", "", "", "", "")
// 	}
// 	cche, err := cache.NewCache(config.cacheTtl, config.cacheCleanwindow, config.cacheMaxSizeMB)
// 	if err != nil {
// 		cche = cache.NewNopCache()
// 	}
// 	c = &Client{Config: config, cache: cche}
// 	return
// }

// func (client *Client) WithFreshCache() {
// 	client.Config.freshCache = true
// }

// func (client *Client) authenticate() error {
// 	client.Config.Lock()
// 	defer client.Config.Unlock()
// 	if client.Config.AuthToken == nil || client.Config.AuthToken.AccessToken == "" || utils.IsTokenExpired(client.Config.AuthToken.AccessToken) {
// 		if client.Config.ClientID == "" || client.Config.ClientSecret == "" {
// 			client.Config.Logger.Printf("[ERROR] No client credentials were provided. Please set %s, %s and %s environment variables.\n", ZPA_CLIENT_ID, ZPA_CLIENT_SECRET, ZPA_CUSTOMER_ID)
// 			return errors.New("no client credentials were provided")
// 		}
// 		client.Config.Logger.Printf("[TRACE] Getting access token for %s=%s\n", ZPA_CLIENT_ID, client.Config.ClientID)
// 		data := url.Values{}
// 		data.Set("client_id", client.Config.ClientID)
// 		data.Set("client_secret", client.Config.ClientSecret)
// 		authUrl := client.Config.BaseURL.String() + "/signin"
// 		if client.Config.Cloud == "DEV" {
// 			authUrl = devAuthUrl
// 		}
// 		req, err := http.NewRequest("POST", authUrl, strings.NewReader(data.Encode()))
// 		if err != nil {
// 			client.Config.Logger.Printf("[ERROR] Failed to signin the user %s=%s, err: %v\n", ZPA_CLIENT_ID, client.Config.ClientID, err)
// 			return fmt.Errorf("[ERROR] Failed to signin the user %s=%s, err: %v", ZPA_CLIENT_ID, client.Config.ClientID, err)
// 		}

// 		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 		if client.Config.UserAgent != "" {
// 			req.Header.Add("User-Agent", client.Config.UserAgent)
// 		}
// 		resp, err := client.Config.GetHTTPClient().Do(req)
// 		if err != nil {
// 			client.Config.Logger.Printf("[ERROR] Failed to signin the user %s=%s, err: %v\n", ZPA_CLIENT_ID, client.Config.ClientID, err)
// 			return fmt.Errorf("[ERROR] Failed to signin the user %s=%s, err: %v", ZPA_CLIENT_ID, client.Config.ClientID, err)
// 		}
// 		defer resp.Body.Close()
// 		respBody, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			client.Config.Logger.Printf("[ERROR] Failed to signin the user %s=%s, err: %v\n", ZPA_CLIENT_ID, client.Config.ClientID, err)
// 			return fmt.Errorf("[ERROR] Failed to signin the user %s=%s, err: %v", ZPA_CLIENT_ID, client.Config.ClientID, err)
// 		}
// 		if resp.StatusCode >= 300 {
// 			client.Config.Logger.Printf("[ERROR] Failed to signin the user %s=%s, got http status:%dn response body:%s\n", ZPA_CLIENT_ID, client.Config.ClientID, resp.StatusCode, respBody)
// 			return fmt.Errorf("[ERROR] Failed to signin the user %s=%s, got http status:%d, response body:%s", ZPA_CLIENT_ID, client.Config.ClientID, resp.StatusCode, respBody)
// 		}
// 		var a AuthToken
// 		err = json.Unmarshal(respBody, &a)
// 		if err != nil {
// 			client.Config.Logger.Printf("[ERROR] Failed to signin the user %s=%s, err: %v\n", ZPA_CLIENT_ID, client.Config.ClientID, err)
// 			return fmt.Errorf("[ERROR] Failed to signin the user %s=%s, err: %v", ZPA_CLIENT_ID, client.Config.ClientID, err)
// 		}
// 		// we need keep auth token for future http request
// 		client.Config.AuthToken = &a
// 	}
// 	return nil
// }

// func (client *Client) newRequestDoCustom(method, urlStr string, options, body, v interface{}) (*http.Response, error) {
// 	err := client.authenticate()
// 	if err != nil {
// 		return nil, err
// 	}
// 	req, err := client.newRequest(method, urlStr, options, body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	reqID := uuid.NewString()
// 	start := time.Now()
// 	logger.LogRequest(client.Config.Logger, req, reqID, nil, true)
// 	resp, err := client.do(req, v, start, reqID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
// 		err := client.authenticate()
// 		if err != nil {
// 			return nil, err
// 		}

// 		resp, err := client.do(req, v, start, reqID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		resp.Body.Close()
// 		return resp, nil
// 	}
// 	return resp, err
// }
