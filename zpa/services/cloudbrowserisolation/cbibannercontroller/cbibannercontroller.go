package cbibannercontroller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa/services"
)

const (
	cbiConfig          = "/cbiconfig/cbi/api/customers/"
	cbiBannerEndpoint  = "/banner"
	cbiBannersEndpoint = "/banners"
)

type CBIBannerController struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	PrimaryColor      string `json:"primaryColor,omitempty"`
	TextColor         string `json:"textColor,omitempty"`
	NotificationTitle string `json:"notificationTitle,omitempty"`
	NotificationText  string `json:"notificationText,omitempty"`
	Logo              string `json:"logo,omitempty"`
	Banner            bool   `json:"banner,omitempty"`
	IsDefault         bool   `json:"isDefault,omitempty"`
	Persist           bool   `json:"persist,omitempty"`
}

func Get(service *services.Service, bannerID string) (*CBIBannerController, *http.Response, error) {
	v := new(CBIBannerController)
	relativeURL := fmt.Sprintf("%s/%s", cbiConfig+service.Client.GetCustomerID()+cbiBannersEndpoint, bannerID)
	resp, err := service.NewRequestDo("GET", relativeURL, nil, nil, &v)
	if err != nil {
		return nil, nil, err
	}

	return v, resp, nil
}

func GetByNameOrID(service *services.Service, identifier string) (*CBIBannerController, *http.Response, error) {
	// Retrieve all banners
	list, resp, err := GetAll(service)
	if err != nil {
		return nil, nil, err
	}

	// Try to find by ID
	for _, banner := range list {
		if banner.ID == identifier {
			return Get(service, banner.ID)
		}
	}

	// Try to find by name
	for _, banner := range list {
		if strings.EqualFold(banner.Name, identifier) {
			return Get(service, banner.ID)
		}
	}

	return nil, resp, fmt.Errorf("no isolation banner named or with ID '%s' was found", identifier)
}

func Create(service *services.Service, cbiBanner *CBIBannerController) (*CBIBannerController, *http.Response, error) {
	v := new(CBIBannerController)
	resp, err := service.NewRequestDo("POST", cbiConfig+service.Client.GetCustomerID()+cbiBannerEndpoint, nil, cbiBanner, &v)
	if err != nil {
		return nil, nil, err
	}
	return v, resp, nil
}

func Update(service *services.Service, cbiBannerID string, cbiBanner *CBIBannerController) (*http.Response, error) {
	path := fmt.Sprintf("%v/%v", cbiConfig+service.Client.GetCustomerID()+cbiBannersEndpoint, cbiBannerID)
	resp, err := service.NewRequestDo("PUT", path, nil, cbiBanner, nil)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func Delete(service *services.Service, cbiBannerID string) (*http.Response, error) {
	path := fmt.Sprintf("%v/%v", cbiConfig+service.Client.GetCustomerID()+cbiBannersEndpoint, cbiBannerID)
	resp, err := service.NewRequestDo("DELETE", path, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func GetAll(service *services.Service) ([]CBIBannerController, *http.Response, error) {
	relativeURL := cbiConfig + service.Client.GetCustomerID() + cbiBannersEndpoint
	var list []CBIBannerController
	resp, err := service.NewRequestDo("GET", relativeURL, nil, nil, &list)
	if err != nil {
		return nil, nil, err
	}
	return list, resp, nil
}
