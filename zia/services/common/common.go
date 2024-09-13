package common

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

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/common"
)

const pageSize = 1000

type IDNameExtensions struct {
	ID         int                    `json:"id,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type IDExtensions struct {
	ID         int                    `json:"id,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type IDName struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type IDCustom struct {
	ID int `json:"id,omitempty"`
}

type ZPAAppSegments struct {
	// A unique identifier assigned to the Application Segment
	ID int `json:"id"`

	// The name of the Application Segment
	Name string `json:"name,omitempty"`

	// Indicates the external ID. Applicable only when this reference is of an external entity.
	ExternalID string `json:"externalId"`
}

type UserGroups struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	IdpID    int    `json:"idp_id,omitempty"`
	Comments string `json:"comments,omitempty"`
}

type UserDepartment struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	IdpID    int    `json:"idp_id,omitempty"`
	Comments string `json:"comments,omitempty"`
	Deleted  bool   `json:"deleted,omitempty"`
}

type DeviceGroups struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Devices struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type IDNameWorkloadGroup struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type DatacenterSearchParameters struct {
	RoutableIP                bool
	WithinCountryOnly         bool
	IncludePrivateServiceEdge bool
	IncludeCurrentVips        bool
	SourceIp                  string
	Latitude                  float64
	Longitude                 float64
	Subcloud                  string
}

type SandboxRSS struct {
	Risk             string `json:"Risk,omitempty"`
	Signature        string `json:"Signature,omitempty"`
	SignatureSources string `json:"SignatureSources,omitempty"`
}

// GetPageSize returns the page size.
func GetPageSize() int {
	return pageSize
}

func ReadAllPages[T any](client common.Client, endpoint string, list *[]T) error {
	if list == nil {
		return nil
	}
	page := 1
	if !strings.Contains(endpoint, "?") {
		endpoint += "?"
	}

	for {
		pageItems := []T{}
		err := Read(client, fmt.Sprintf("%s&pageSize=%d&page=%d", endpoint, pageSize, page), &pageItems)
		if err != nil {
			return err
		}
		*list = append(*list, pageItems...)
		if len(pageItems) < pageSize {
			break
		}
		page++
	}
	return nil
}

func ReadPage[T any](client common.Client, endpoint string, page int, list *[]T) error {
	if list == nil {
		return nil
	}

	// Parse the endpoint into a URL.
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("could not parse endpoint URL: %w", err)
	}

	// Get the existing query parameters and add new ones.
	q := u.Query()
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))
	q.Set("page", fmt.Sprintf("%d", page))

	// Set the URL's RawQuery to the encoded query parameters.
	u.RawQuery = q.Encode()

	// Convert the URL back to a string and read the page.
	pageItems := []T{}
	err = Read(client, u.String(), &pageItems)
	if err != nil {
		return err
	}
	*list = pageItems
	return nil
}

type (
	SortOrder string
	SortField string
)

const (
	ASCSortOrder          SortOrder = "ASC"
	DESCSortOrder                   = "DESC"
	IDSortField           SortField = "id"
	NameSortField                   = "name"
	CreationTimeSortField           = "creationTime"
	ModifiedTimeSortField           = "modifiedTime"
)

func GetSortParams(sortBy SortField, sortOrder SortOrder) string {
	params := ""
	if sortBy != "" {
		params = "sortBy=" + string(sortBy)
	}
	if sortOrder != "" {
		if params != "" {
			params += "&"
		}
		params += "sortOrder=" + string(sortOrder)
	}
	return params
}

func Read(client common.Client, endpoint string, o interface{}) error {
	contentType := "application/json"
	resp, err := Request(client, endpoint, "GET", nil, contentType)
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, o)
	if err != nil {
		return err
	}

	return nil
}

// Create sends an HTTP POST request.
func Create(client common.Client, endpoint string, o interface{}) (interface{}, error) {
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

	resp, err := Request(client, endpoint, "POST", data, "application/json")
	if err != nil {
		return nil, err
	}
	if len(resp) > 0 {
		// Check if the response is an array of strings
		var stringArrayResponse []string
		if json.Unmarshal(resp, &stringArrayResponse) == nil {
			return stringArrayResponse, nil
		}

		// Otherwise, handle as usual
		responseObject := reflect.New(t).Interface()
		err = json.Unmarshal(resp, &responseObject)
		if err != nil {
			return nil, err
		}
		id := reflect.Indirect(reflect.ValueOf(responseObject)).FieldByName("ID")

		client.GetLogger().Printf("Created Object with ID %v", id)
		return responseObject, nil
	} else {
		// in case of 204 no content
		return nil, nil
	}
}

func CreateWithSlicePayload(client common.Client, endpoint string, slice interface{}) ([]byte, error) {
	if slice == nil {
		return nil, errors.New("tried to create with a nil payload not a Slice")
	}

	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil, errors.New("tried to create with a " + v.Kind().String() + " not a Slice")
	}

	data, err := json.Marshal(slice)
	if err != nil {
		return nil, err
	}

	resp, err := Request(client, endpoint, "POST", data, "application/json")
	if err != nil {
		return nil, err
	}
	if len(resp) > 0 {
		return resp, nil
	} else {
		// in case of 204 no content
		return nil, nil
	}
}

func UpdateWithSlicePayload(client common.Client, endpoint string, slice interface{}) ([]byte, error) {
	if slice == nil {
		return nil, errors.New("tried to update with a nil payload not a Slice")
	}

	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil, errors.New("tried to update with a " + v.Kind().String() + " not a Slice")
	}

	data, err := json.Marshal(slice)
	if err != nil {
		return nil, err
	}

	resp, err := Request(client, endpoint, "PUT", data, "application/json")
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Update ...
func UpdateWithPut(client common.Client, endpoint string, o interface{}) (interface{}, error) {
	return updateGeneric(client, endpoint, o, "PUT", "application/json")
}

// Update ...
func Update(client common.Client, endpoint string, o interface{}) (interface{}, error) {
	return updateGeneric(client, endpoint, o, "PATCH", "application/merge-patch+json")
}

// Update ...
func updateGeneric(client common.Client, endpoint string, o interface{}, method, contentType string) (interface{}, error) {
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

	resp, err := Request(client, endpoint, method, data, contentType)
	if err != nil {
		return nil, err
	}

	responseObject := reflect.New(t).Interface()
	err = json.Unmarshal(resp, &responseObject)
	return responseObject, err
}

// Delete ...
func Delete(client common.Client, endpoint string) error {
	_, err := Request(client, endpoint, "DELETE", nil, "application/json")
	if err != nil {
		return err
	}
	return nil
}

// BulkDelete sends an HTTP POST request for bulk deletion and expects a 204 No Content response.
func BulkDelete(client common.Client, endpoint string, payload interface{}) (*http.Response, error) {
	if payload == nil {
		return nil, errors.New("tried to delete with a nil payload, expected a struct")
	}

	// Marshal the payload into JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Send the POST request
	resp, err := Request(client, endpoint, "POST", data, "application/json")
	if err != nil {
		return nil, err
	}

	// Check the status code (204 No Content expected)
	if len(resp) == 0 {
		client.GetLogger().Printf("[DEBUG] Bulk delete successful with 204 No Content")
		return &http.Response{StatusCode: 204}, nil
	}

	// If the response is not empty, this might indicate an error or unexpected behavior
	return &http.Response{StatusCode: 200}, fmt.Errorf("unexpected response: %s", string(resp))
}

// Request ... // Needs to review this function.
func GenericRequest(client common.Client, baseUrl, endpoint, method string, body io.Reader, urlParams url.Values, contentType string) ([]byte, error) {
	bodyData, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	var respData interface{}
	_, err = client.NewRequestDoGeneric(baseUrl, endpoint, method, urlParams, bodyData, &respData, contentType, common.Option{Name: common.ZscalerInfraOption, Value: "zia"})
	if err != nil {
		return nil, err
	}
	return json.Marshal(respData)
}

func Request(client common.Client, endpoint, method string, data []byte, contentType string) ([]byte, error) {
	return GenericRequest(client, client.GetBaseURL(), endpoint, method, bytes.NewReader(data), nil, contentType)
}
