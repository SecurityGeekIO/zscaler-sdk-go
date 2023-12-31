package inspection_custom_controls

import (
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa"
)

type Service struct {
	Client *zpa.Client
}

func New(c *zpa.Client) *Service {
	return &Service{Client: c}
}
