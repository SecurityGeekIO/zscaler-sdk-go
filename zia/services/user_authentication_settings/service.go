package user_authentication_settings

import "github.com/SecurityGeekIO/zscaler-sdk-go/zia"

type Service struct {
	Client *zia.Client
}

func New(c *zia.Client) *Service {
	return &Service{Client: c}
}
