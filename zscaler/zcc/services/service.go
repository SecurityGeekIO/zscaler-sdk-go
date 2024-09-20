package services

import (
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zcc"
)

type Service struct {
	Client *zcc.Client
}

func New(c *zcc.Client) *Service {
	return &Service{Client: c}
}
