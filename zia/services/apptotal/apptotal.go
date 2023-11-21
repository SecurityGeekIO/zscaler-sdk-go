package apptotal

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	appTotalEndpoint = "/apps/app"
)

type AppTotal struct {
	AppID             int               `json:"appId,omitempty"`
	Name              string            `json:"name,omitempty"`
	Description       string            `json:"description,omitempty"`
	RedirectURLs      []string          `json:"redirectUrls,omitempty"`
	WebSiteURLs       []string          `json:"websiteUrls,omitempty"`
	Categories        []string          `json:"categories,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
	Compliance        []string          `json:"compliance,omitempty"`
	PermissionLevel   float32           `json:"permissionLevel,omitempty"`
	Risk              string            `json:"risk,omitempty"`
	ClientID          string            `json:"clientId,omitempty"`
	ClientType        string            `json:"clientType,omitempty"`
	LogoURL           string            `json:"logoUrl,omitempty"`
	PrivatePolicyURL  string            `json:"privacyPolicyUrl,omitempty"`
	TermsOfServiceUrl string            `json:"termsOfServiceUrl,omitempty"`
	MarketplaceUrl    string            `json:"marketplaceUrl,omitempty"`
	ExtractedURLs     []string          `json:"extractedUrls,omitempty"`
	ExtractedApiCalls []string          `json:"extractedApiCalls,omitempty"`
	Platform          string            `json:"platform,omitempty"`
	DataRetention     string            `json:"dataRetention,omitempty"`
	PlatformVerified  bool              `json:"platformVerified"`
	DeveloperEmail    string            `json:"developerEmail,omitempty"`
	ConsentScreenshot string            `json:"consentScreenshot,omitempty"`
	ExternalIDs       []ExternalIDs     `json:"externalIds,omitempty"`
	Publisher         []Publisher       `json:"publisher,omitempty"`
	Permissions       []Permissions     `json:"permissions,omitempty"`
	MarketplaceData   []MarketplaceData `json:"marketplaceData,omitempty"`
	IPAddresses       []IPAddresses     `json:"ipAddresses,omitempty"`
	Vulnerabilities   []Vulnerabilities `json:"vulnerabilities,omitempty"`
	APIActivities     []APIActivities   `json:"apiActivities,omitempty"`
	Risks             []Risks           `json:"risks,omitempty"`
	Insights          []Insights        `json:"insights,omitempty"`
}

type ExternalIDs struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type Publisher struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	SiteURL     string `json:"siteUrl,omitempty"`
	LogoURL     string `json:"logoUrl,omitempty"`
}

type Permissions struct {
	Scope       string `json:"scope,omitempty"`
	Service     string `json:"service,omitempty"`
	Description string `json:"description,omitempty"`
	AccessType  string `json:"accessType,omitempty"`
	Level       string `json:"level,omitempty"`
}

type MarketplaceData struct {
	Stars     int `json:"stars,omitempty"`
	Downloads int `json:"downloads,omitempty"`
	Reviews   int `json:"reviews,omitempty"`
}

type IPAddresses struct {
	ISPName     string `json:"ispName,omitempty"`
	IPAddress   string `json:"ipAddress,omitempty"`
	ProxyType   string `json:"proxyType,omitempty"`
	UsageType   string `json:"usageType,omitempty"`
	DomainName  string `json:"domainName,omitempty"`
	CountryCode string `json:"countryCode,omitempty"`
}

type Vulnerabilities struct {
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
	CVEID    string `json:"cveId,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Severity string `json:"severity,omitempty"`
}

type APIActivities struct {
	OperationType string `json:"operationType,omitempty"`
	Percentage    string `json:"percentage,omitempty"`
}

type Risks struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Severity    string `json:"severity,omitempty"`
}

type Insights struct {
	Description string                 `json:"description,omitempty"`
	Timestamp   int                    `json:"timestamp,omitempty"`
	URLs        map[string]interface{} `json:"urls,omitempty"`
}

type AppTotalCreation struct {
	AppID int `json:"appId"`
}

func (service *Service) Get(appID int, verbose bool) (*AppTotal, error) {
	var app AppTotal

	// Construct the endpoint URL with the appID and the verbose parameter
	endpoint := fmt.Sprintf("%s/%d", appTotalEndpoint, appID)
	endpoint = addVerboseQueryParam(endpoint, verbose)

	err := service.Client.Read(endpoint, &app)
	if err != nil {
		return nil, err
	}

	service.Client.Logger.Printf("[DEBUG] Returning app details from Get: %d", appID)
	return &app, nil
}

func addVerboseQueryParam(endpoint string, verbose bool) string {
	// Convert the endpoint string to URL type
	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}

	// Add the verbose query parameter
	q := u.Query()
	q.Set("verbose", fmt.Sprintf("%t", verbose))
	u.RawQuery = q.Encode()

	return u.String()
}

func (service *Service) Create(app *AppTotalCreation) (*AppTotal, *http.Response, error) {
	resp, err := service.Client.Create(appTotalEndpoint, *app)
	if err != nil {
		return nil, nil, err
	}

	createdApp, ok := resp.(*AppTotal)
	if !ok {
		return nil, nil, errors.New("object returned from api was not an AppTotal pointer")
	}

	service.Client.Logger.Printf("[DEBUG] Returning new app details from Create: %d", createdApp.AppID)
	return createdApp, nil, nil
}
