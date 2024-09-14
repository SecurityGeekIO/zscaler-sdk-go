package appservicegroups

import (
	"fmt"
	"strings"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia/services/common"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zidentity"
)

const (
	appServicesGroupLiteEndpoint = "/appServiceGroups/lite"
)

type ApplicationServicesGroupLite struct {
	ID          int    `json:"id"`
	Name        string `json:"name,omitempty"`
	NameL10nTag bool   `json:"nameL10nTag"`
}

func GetByName(service *zidentity.Service, serviceGroupName string) (*ApplicationServicesGroupLite, error) {
	var appServicesGroupLite []ApplicationServicesGroupLite
	err := common.ReadAllPages(service.Client, appServicesGroupLiteEndpoint, &appServicesGroupLite)
	if err != nil {
		return nil, err
	}
	for _, appServicesGroupLite := range appServicesGroupLite {
		if strings.EqualFold(appServicesGroupLite.Name, serviceGroupName) {
			return &appServicesGroupLite, nil
		}
	}
	return nil, fmt.Errorf("no app services group found with name: %s", serviceGroupName)
}

func GetAll(service *zidentity.Service) ([]ApplicationServicesGroupLite, error) {
	var appServiceGroups []ApplicationServicesGroupLite
	err := common.ReadAllPages(service.Client, appServicesGroupLiteEndpoint, &appServiceGroups)
	return appServiceGroups, err
}
