package common

import (
	"net/http"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/logger"
)

type OptionName string

const (
	ZscalerInfraOption OptionName = "infra"
)

type Option struct {
	Name  OptionName
	Value string
}

type Client interface {
	NewRequestDoGeneric(baseUrl, path, method string, options, body, v interface{}, contentType string, infraOptions ...Option) (*http.Response, error)
	NewRequestDo(method, path string, options, body, v interface{}, infraOptions ...Option) (*http.Response, error)
	GetCustomerID() string
	SetCustomerID(string)
	GetLogger() logger.Logger
	GetBaseURL() string
	GetCloud() string
}
