package zidentity

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ZIDENTITY_CLIENT_ID     = "ZIDENTITY_CLIENT_ID"
	ZIDENTITY_CLIENT_SECRET = "ZIDENTITY_CLIENT_SECRET"
	ZIDENTITY_VANITY_DOMAIN = "ZIDENTITY_VANITY_DOMAIN"
	ZIDENTITY_PRIVATE_KEY   = "ZIDENTITY_PRIVATE_KEY"
)

type AuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

type Credentials struct {
	AuthToken    *AuthToken
	ClientID     string
	ClientSecret string
	VanityDomain string
	PrivateKey   string
	Cloud        string
	UserAgent    string
}

func Authenticate(creds *Credentials, httpClient *http.Client) (*AuthToken, error) {
	if creds.ClientID == "" || (creds.ClientSecret == "" && creds.PrivateKey == "") {
		return nil, errors.New("no client credentials were provided")
	}

	if creds.PrivateKey != "" {
		return authenticatWithCert(creds, httpClient)
	}

	var authUrl string
	// Ensure the vanity domain is provided and does not include protocol
	if creds.Cloud == "PRODUCTION" || creds.Cloud == "" {
		authUrl = fmt.Sprintf("https://%s.zslogin.net/oauth2/v1/token", creds.VanityDomain)
	} else {
		authUrl = fmt.Sprintf("https://%s.zslogin%s.net/oauth2/v1/token", creds.VanityDomain, strings.ToLower(creds.Cloud))
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_secret", creds.ClientSecret)
	data.Set("client_id", creds.ClientID)
	data.Set("audience", "https://api.zscaler.com")

	req, err := http.NewRequest("POST", authUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to signin the user %s, err: %v", creds.ClientID, err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if creds.UserAgent != "" {
		req.Header.Add("User-Agent", creds.UserAgent)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to signin the user %s, err: %v", creds.ClientID, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to signin the user %s, err: %v", creds.ClientID, err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("[ERROR] Failed to signin the user %s, got http status:%d, response body:%s", creds.ClientID, resp.StatusCode, respBody)
	}
	var a AuthToken
	err = json.Unmarshal(respBody, &a)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to signin the user %s, err: %v", creds.ClientID, err)
	}

	return &a, nil
}

func authenticatWithCert(creds *Credentials, httpClient *http.Client) (*AuthToken, error) {
	if creds.ClientID == "" {
		return nil, errors.New("no client ID was provided")
	}
	if creds.PrivateKey == "" {
		creds.PrivateKey = os.Getenv(ZIDENTITY_PRIVATE_KEY)
	}

	if creds.PrivateKey == "" {
		return nil, errors.New("no private key path was provided")
	}

	// Load the private key
	privateKeyData, err := os.ReadFile(creds.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error reading private key: %v", err)
	}

	// Parse the private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}

	// Create the JWT payload
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": creds.ClientID,
		"sub": creds.ClientID,
		"aud": "https://api.zscaler.com",
		"exp": time.Now().Add(10 * time.Minute).Unix(),
	})

	assertionString, err := token.SignedString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error signing JWT: %v", err)
	}

	// Prepare the form data for the OAuth token request
	formData := url.Values{
		"grant_type":            {"client_credentials"},
		"client_id":             {creds.ClientID},
		"client_assertion":      {assertionString},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"audience":              {"https://api.zscaler.com"},
	}

	var authUrl string
	// Ensure the vanity domain is provided and does not include protocol
	if creds.Cloud == "PRODUCTION" || creds.Cloud == "" {
		authUrl = fmt.Sprintf("https://%s.zslogin.net/oauth2/v1/token", creds.VanityDomain)
	} else {
		authUrl = fmt.Sprintf("https://%s.zslogin%s.net/oauth2/v1/token", creds.VanityDomain, strings.ToLower(creds.Cloud))
	}

	// Make the POST request
	resp, err := httpClient.PostForm(authUrl, formData)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Parse the response
	var authToken AuthToken
	err = json.Unmarshal(body, &authToken)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &authToken, nil
}
