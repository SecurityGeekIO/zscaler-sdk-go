package publicapi

import (
	"github.com/willguibr/zscaler-sdk-go/zcc"
)

type Service struct {
	Client *zcc.Client
}

func New(c *zcc.Client) *Service {
	return &Service{Client: c}
}
