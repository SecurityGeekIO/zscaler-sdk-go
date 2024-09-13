package zpa_gateways

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
func (service *Service) create(endpoint string, o interface{}) (interface{}, error) {
	return common.Create(service.Client, endpoint, o)
}

func (service *Service) createWithSlicePayload(endpoint string, slice interface{}) ([]byte, error) {
	return common.CreateWithSlicePayload(service.Client, endpoint, slice)
}

func (service *Service) updateWithSlicePayload(endpoint string, slice interface{}) ([]byte, error) {
	return common.UpdateWithSlicePayload(service.Client, endpoint, slice)
}

// Read ...
func (service *Service) read(endpoint string, o interface{}) error {
	return common.Read(service.Client, endpoint, o)
}

// Update ...
func (service *Service) updateWithPut(endpoint string, o interface{}) (interface{}, error) {
	return common.UpdateWithPut(service.Client, endpoint, o)
}

// Update ...
func (service *Service) update(endpoint string, o interface{}) (interface{}, error) {
	return common.Update(service.Client, endpoint, o)
}

// Delete ...
func (service *Service) delete(endpoint string) error {
	return common.Delete(service.Client, endpoint)
}

// BulkDelete sends an HTTP POST request for bulk deletion and expects a 204 No Content response.
func (service *Service) bulkDelete(endpoint string, payload interface{}) (*http.Response, error) {
	return common.BulkDelete(service.Client, endpoint, payload)
}
