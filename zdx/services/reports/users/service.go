package users

import (
	"github.com/willguibr/zscaler-sdk-go/zdx"
)

type Service struct {
	Client *zdx.Client
}

func New(c *zdx.Client) *Service {
	return &Service{Client: c}
}
