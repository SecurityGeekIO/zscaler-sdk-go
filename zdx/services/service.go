package services

import (
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zdx"
)

type Service struct {
	Client *zdx.Client
}

func New(c *zdx.Client) *Service {
	return &Service{Client: c}
}
