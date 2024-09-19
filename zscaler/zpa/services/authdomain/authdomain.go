package authdomain

import (
	"net/http"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zscaler/zpa/services"
)

const (
	mgmtConfig         = "/mgmtconfig/v1/admin/customers/"
	authDomainEndpoint = "/authDomains"
)

type AuthDomain struct {
	AuthDomains []string `json:"authDomains"`
}

func GetAllAuthDomains(service *services.Service) (*AuthDomain, *http.Response, error) {
	v := new(AuthDomain)
	relativeURL := mgmtConfig + service.Client.Config.CustomerID + authDomainEndpoint
	resp, err := service.Client.NewRequestDo("GET", relativeURL, nil, nil, &v)
	if err != nil {
		return nil, nil, err
	}

	return v, resp, nil
}
