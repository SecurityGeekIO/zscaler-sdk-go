package services

import (
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia"
)

type Service struct {
	Client *zia.Client
}

func New(c *zia.Client) *Service {
	return &Service{Client: c}
}
