package servergroup

import (
	"github.com/SecurityGeekIO/zscaler-sdk-go/zpa"
)

type Service struct {
	Client *zpa.Client
}

func New(c *zpa.Client) *Service {
	return &Service{Client: c}
}
