package scimgroup

import (
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa"
)

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

type Service struct {
	Client    *zpa.Client // Legacy client
	ClientI   zpa.ClientI // New interface-based client
	sortOrder SortOrder
	sortBy    SortField
}

// New instantiates a Service with both the legacy Client and the new ClientI
func New(c *zpa.Client, ci zpa.ClientI) *Service {
	return &Service{
		Client:    c,  // The legacy client (can be nil)
		ClientI:   ci, // The OneAPIClient, which implements ClientI
		sortOrder: ASCSortOrder,
		sortBy:    NameSortField,
	}
}

// WithSort returns a copy of the Service, modifying the sort options
func (service *Service) WithSort(sortBy SortField, sortOrder SortOrder) *Service {
	// Create a copy of the current service
	c := &Service{
		Client:    service.Client,
		ClientI:   service.ClientI, // Ensure ClientI is copied over
		sortOrder: service.sortOrder,
		sortBy:    service.sortBy,
	}

	// Validate and set sort field
	if sortBy == IDSortField || sortBy == NameSortField || sortBy == CreationTimeSortField || sortBy == ModifiedTimeSortField {
		c.sortBy = sortBy
	}

	// Validate and set sort order
	if sortOrder == ASCSortOrder || sortOrder == DESCSortOrder {
		c.sortOrder = sortOrder
	}
	return c
}
