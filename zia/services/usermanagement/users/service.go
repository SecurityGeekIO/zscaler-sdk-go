package users

import (
	"net/http"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia/services/common"
)

type Service struct {
	Client    *zia.Client
	sortOrder common.SortOrder
	sortBy    common.SortField
}

func New(c *zia.Client) *Service {
	return &Service{
		Client:    c,
		sortOrder: common.ASCSortOrder,
		sortBy:    common.NameSortField,
	}
}

func (service *Service) WithSort(sortBy common.SortField, sortOrder common.SortOrder) *Service {
	c := Service{
		Client:    service.Client,
		sortOrder: service.sortOrder,
		sortBy:    service.sortBy,
	}
	if sortBy == common.IDSortField || sortBy == common.NameSortField || sortBy == common.CreationTimeSortField || sortBy == common.ModifiedTimeSortField {
		c.sortBy = sortBy
	}

	if sortOrder == common.ASCSortOrder || sortOrder == common.DESCSortOrder {
		c.sortOrder = sortOrder
	}
	return &c
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
