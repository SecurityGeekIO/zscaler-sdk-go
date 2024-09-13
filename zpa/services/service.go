package services

import (
	"net/http"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/common"
)

type Service struct {
	Client        common.Client
	microTenantID *string
}

func New(c common.Client) *Service {
	return &Service{Client: c}
}

func (service *Service) NewRequestDo(method, url string, options, body, v interface{}) (*http.Response, error) {
	return service.Client.NewRequestDo(method, url, options, body, v, common.Option{Name: common.ZscalerInfraOption, Value: "zpa"})
}

func (service *Service) WithMicroTenant(microTenantID string) *Service {
	var mid *string
	if microTenantID != "" {
		mid_ := microTenantID
		mid = &mid_
	}
	return &Service{
		Client:        service.Client,
		microTenantID: mid,
	}
}

func (service *Service) MicroTenantID() *string {
	return service.microTenantID
}
