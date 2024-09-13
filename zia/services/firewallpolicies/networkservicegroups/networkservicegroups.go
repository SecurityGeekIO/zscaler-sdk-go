package networkservicegroups

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia/services"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia/services/common"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia/services/firewallpolicies/networkservices"
)

const (
	networkServiceGroupsEndpoint = "/networkServiceGroups"
)

type NetworkServiceGroups struct {
	ID          int        `json:"id"`
	Name        string     `json:"name,omitempty"`
	Services    []Services `json:"services,omitempty"`
	Description string     `json:"description,omitempty"`
}

type Services struct {
	ID            int                            `json:"id"`
	Name          string                         `json:"name,omitempty"`
	Tag           string                         `json:"tag,omitempty"`
	SrcTCPPorts   []networkservices.NetworkPorts `json:"srcTcpPorts,omitempty"`
	DestTCPPorts  []networkservices.NetworkPorts `json:"destTcpPorts,omitempty"`
	SrcUDPPorts   []networkservices.NetworkPorts `json:"srcUdpPorts,omitempty"`
	DestUDPPorts  []networkservices.NetworkPorts `json:"destUdpPorts,omitempty"`
	Type          string                         `json:"type,omitempty"`
	Description   string                         `json:"description,omitempty"`
	IsNameL10nTag bool                           `json:"isNameL10nTag,omitempty"`
}

func GetNetworkServiceGroups(service *services.Service, serviceGroupID int) (*NetworkServiceGroups, error) {
	var networkServiceGroups NetworkServiceGroups
	err := service.Read(fmt.Sprintf("%s/%d", networkServiceGroupsEndpoint, serviceGroupID), &networkServiceGroups)
	if err != nil {
		return nil, err
	}

	service.Client.GetLogger().Printf("[DEBUG]Returning network service groups from Get: %d", networkServiceGroups.ID)
	return &networkServiceGroups, nil
}

func GetNetworkServiceGroupsByName(service *services.Service, serviceGroupsName string) (*NetworkServiceGroups, error) {
	var networkServiceGroups []NetworkServiceGroups
	err := common.ReadAllPages(service.Client, networkServiceGroupsEndpoint, &networkServiceGroups)
	if err != nil {
		return nil, err
	}
	for _, networkServiceGroup := range networkServiceGroups {
		if strings.EqualFold(networkServiceGroup.Name, serviceGroupsName) {
			return &networkServiceGroup, nil
		}
	}
	return nil, fmt.Errorf("no network service groups found with name: %s", serviceGroupsName)
}

func CreateNetworkServiceGroups(service *services.Service, networkServiceGroups *NetworkServiceGroups) (*NetworkServiceGroups, error) {
	resp, err := service.Create(networkServiceGroupsEndpoint, *networkServiceGroups)
	if err != nil {
		return nil, err
	}

	createdNetworkServiceGroups, ok := resp.(*NetworkServiceGroups)
	if !ok {
		return nil, errors.New("object returned from api was not a network service groups pointer")
	}

	service.Client.GetLogger().Printf("[DEBUG]returning network service groups from create: %d", createdNetworkServiceGroups.ID)
	return createdNetworkServiceGroups, nil
}

func UpdateNetworkServiceGroups(service *services.Service, serviceGroupID int, networkServiceGroups *NetworkServiceGroups) (*NetworkServiceGroups, *http.Response, error) {
	resp, err := service.UpdateWithPut(fmt.Sprintf("%s/%d", networkServiceGroupsEndpoint, serviceGroupID), *networkServiceGroups)
	if err != nil {
		return nil, nil, err
	}
	updatedNetworkServiceGroups, _ := resp.(*NetworkServiceGroups)

	service.Client.GetLogger().Printf("[DEBUG]returning network service groups from Update: %d", updatedNetworkServiceGroups.ID)
	return updatedNetworkServiceGroups, nil, nil
}

func DeleteNetworkServiceGroups(service *services.Service, serviceGroupID int) (*http.Response, error) {
	err := service.Delete(fmt.Sprintf("%s/%d", networkServiceGroupsEndpoint, serviceGroupID))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func GetAllNetworkServiceGroups(service *services.Service) ([]NetworkServiceGroups, error) {
	var networkServiceGroups []NetworkServiceGroups
	err := common.ReadAllPages(service.Client, networkServiceGroupsEndpoint, &networkServiceGroups)
	return networkServiceGroups, err
}
