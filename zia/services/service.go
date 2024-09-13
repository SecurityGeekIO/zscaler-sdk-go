package services

import (
	"net/http"

	cmmon "github.com/SecurityGeekIO/zscaler-sdk-go/v2/common"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia/services/common"
)

type Service struct {
	Client cmmon.Client
}

func New(c cmmon.Client) *Service {
	return &Service{Client: c}
}

// Create sends an HTTP POST request.
func (service *Service) Create(endpoint string, o interface{}) (interface{}, error) {
	return common.Create(service.Client, endpoint, o)
}

func (service *Service) CreateWithSlicePayload(endpoint string, slice interface{}) ([]byte, error) {
	return common.CreateWithSlicePayload(service.Client, endpoint, slice)
}

func (service *Service) UpdateWithSlicePayload(endpoint string, slice interface{}) ([]byte, error) {
	return common.UpdateWithSlicePayload(service.Client, endpoint, slice)
}

// Read ...
func (service *Service) Read(endpoint string, o interface{}) error {
	return common.Read(service.Client, endpoint, o)
}

// Update ...
func (service *Service) UpdateWithPut(endpoint string, o interface{}) (interface{}, error) {
	return common.UpdateWithPut(service.Client, endpoint, o)
}

// Update ...
func (service *Service) Update(endpoint string, o interface{}) (interface{}, error) {
	return common.Update(service.Client, endpoint, o)
}

// Delete ...
func (service *Service) Delete(endpoint string) error {
	return common.Delete(service.Client, endpoint)
}

// BulkDelete sends an HTTP POST request for bulk deletion and expects a 204 No Content response.
func (service *Service) BulkDelete(endpoint string, payload interface{}) (*http.Response, error) {
	return common.BulkDelete(service.Client, endpoint, payload)
}
