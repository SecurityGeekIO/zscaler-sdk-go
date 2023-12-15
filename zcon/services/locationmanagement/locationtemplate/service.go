package locationtemplate

import "github.com/SecurityGeekIO/zscaler-sdk-go/v2/zcon"

type Service struct {
	Client *zcon.Client
}

func New(c *zcon.Client) *Service {
	return &Service{Client: c}
}
