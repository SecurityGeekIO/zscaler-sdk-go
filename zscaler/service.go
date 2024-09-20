package zscaler

// Service defines the structure that contains the common client
type Service struct {
	Client        *Client // use the common Zscaler OneAPI Client here
	microTenantID *string
}

// NewService is a generic function to instantiate a Service with the Zscaler OneAPI Client
func NewService(client *Client) *Service {
	return &Service{Client: client}
}

func (service *Service) WithMicroTenant(microTenantID string) *Service {
	var mid *string
	if microTenantID != "" {
		mid_ := microTenantID
		mid = &mid_
	}
	return &Service{
		Client:        service.Client,
		microTenantID: mid,
	}
}

func (service *Service) MicroTenantID() *string {
	return service.microTenantID
}
