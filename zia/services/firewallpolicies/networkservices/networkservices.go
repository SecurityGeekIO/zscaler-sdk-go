package networkservices

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia/services/common"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zidentity"
)

const (
	networkServicesEndpoint = "/networkServices"
)

type NetworkServices struct {
	ID            int            `json:"id"`
	Name          string         `json:"name,omitempty"`
	Tag           string         `json:"tag,omitempty"`
	SrcTCPPorts   []NetworkPorts `json:"srcTcpPorts,omitempty"`
	DestTCPPorts  []NetworkPorts `json:"destTcpPorts,omitempty"`
	SrcUDPPorts   []NetworkPorts `json:"srcUdpPorts,omitempty"`
	DestUDPPorts  []NetworkPorts `json:"destUdpPorts,omitempty"`
	Type          string         `json:"type,omitempty"`
	Description   string         `json:"description,omitempty"`
	Protocol      string         `json:"protocol,omitempty"`
	IsNameL10nTag bool           `json:"isNameL10nTag,omitempty"`
}

type NetworkPorts struct {
	Start int `json:"start,omitempty"`
	End   int `json:"end,omitempty"`
}

func Get(service *zidentity.Service, serviceID int) (*NetworkServices, error) {
	var networkServices NetworkServices
	err := service.Client.Read(fmt.Sprintf("%s/%d", networkServicesEndpoint, serviceID), &networkServices)
	if err != nil {
		return nil, err
	}

	service.Client.Logger.Printf("[DEBUG]Returning network services from Get: %d", networkServices.ID)
	return &networkServices, nil
}

func GetByName(service *zidentity.Service, networkServiceName string) (*NetworkServices, error) {
	var networkServices []NetworkServices
	err := common.ReadAllPages(service.Client, networkServicesEndpoint, &networkServices)
	if err != nil {
		return nil, err
	}
	for _, networkService := range networkServices {
		if strings.EqualFold(networkService.Name, networkServiceName) {
			return &networkService, nil
		}
	}
	return nil, fmt.Errorf("no network services found with name: %s", networkServiceName)
}

func Create(service *zidentity.Service, networkService *NetworkServices) (*NetworkServices, error) {
	resp, err := service.Client.Create(networkServicesEndpoint, *networkService)
	if err != nil {
		return nil, err
	}

	createdNetworkServices, ok := resp.(*NetworkServices)
	if !ok {
		return nil, errors.New("object returned from api was not a network service pointer")
	}

	service.Client.Logger.Printf("[DEBUG]returning network service from create: %d", createdNetworkServices.ID)
	return createdNetworkServices, nil
}

func Update(service *zidentity.Service, serviceID int, networkService *NetworkServices) (*NetworkServices, *http.Response, error) {
	resp, err := service.Client.UpdateWithPut(fmt.Sprintf("%s/%d", networkServicesEndpoint, serviceID), *networkService)
	if err != nil {
		return nil, nil, err
	}
	updatedNetworkServices, _ := resp.(*NetworkServices)

	service.Client.Logger.Printf("[DEBUG]returning network service from Update: %d", updatedNetworkServices.ID)
	return updatedNetworkServices, nil, nil
}

func Delete(service *zidentity.Service, serviceID int) (*http.Response, error) {
	err := service.Client.Delete(fmt.Sprintf("%s/%d", networkServicesEndpoint, serviceID))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func GetAllNetworkServices(service *zidentity.Service) ([]NetworkServices, error) {
	var networkServices []NetworkServices
	err := common.ReadAllPages(service.Client, networkServicesEndpoint, &networkServices)
	return networkServices, err
}
