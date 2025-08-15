package activation

import (
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/ztw"
)

type Service struct {
	Client *ztw.Client
}

func New(c *ztw.Client) *Service {
	return &Service{Client: c}
}
