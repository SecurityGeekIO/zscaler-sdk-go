package zscaler

// Service defines the structure that contains the common client
type Service struct {
	Client *Client // use the common zidentity Client here
}

// NewService is a generic function to instantiate a Service with the zidentity client
func NewService(client *Client) *Service {
	return &Service{Client: client}
}
