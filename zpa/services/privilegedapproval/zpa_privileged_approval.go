package privilegedapproval

import (
	"fmt"
	"net/http"

	"github.com/SecurityGeekIO/zscaler-sdk-go/zpa/services/common"
)

const (
	mgmtConfig                 = "/mgmtconfig/v1/admin/customers/"
	privilegedApprovalEndpoint = "/privilegedApproval"
)

type PrivilegedApproval struct {
	ID           string         `json:"id,omitempty"`
	EmailIDs     []string       `json:"emailIds,omitempty"`
	EndTime      string         `json:"endTime,omitempty"`
	StartTime    string         `json:"startTime,omitempty"`
	Status       string         `json:"status,omitempty"`
	CreationTime string         `json:"creationTime,omitempty"`
	ModifiedBy   string         `json:"modifiedBy,omitempty"`
	ModifiedTime string         `json:"modifiedTime,omitempty"`
	WorkingHours WorkingHours   `json:"workingHours"`
	Applications []Applications `json:"applications"`
}

type Applications struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type WorkingHours struct {
	Days          string `json:"days,omitempty"`
	EndTime       string `json:"endTime,omitempty"`
	EndTimeCron   string `json:"endTimeCron,omitempty"`
	StartTime     string `json:"startTime,omitempty"`
	StartTimeCron string `json:"startTimeCron,omitempty"`
	TimeZone      string `json:"timeZone,omitempty"`
}

func (service *Service) Get(approvalID string) (*PrivilegedApproval, *http.Response, error) {
	v := new(PrivilegedApproval)
	relativeURL := fmt.Sprintf("%s/%s", mgmtConfig+service.Client.Config.CustomerID+privilegedApprovalEndpoint, approvalID)
	resp, err := service.Client.NewRequestDo("GET", relativeURL, nil, nil, v)
	if err != nil {
		return nil, nil, err
	}
	return v, resp, nil
}

/*
// Need to implement search by Email ID
func (service *Service) GetByEmailID(emailID string) (*PrivilegedApproval, *http.Response, error) {
	relativeURL := mgmtConfig + service.Client.Config.CustomerID + privilegedApprovalEndpoint
	list, resp, err := common.GetAllPagesGeneric[PrivilegedApproval](service.Client, relativeURL, emailID)
	if err != nil {
		return nil, nil, err
	}
	for _, app := range list {
		if strings.EqualFold(app.EmailIDs[], emailID) {
			return &app, resp, nil
		}
	}
	return nil, resp, fmt.Errorf("no application named '%s' was found", emailID)
}
*/

func (service *Service) Create(privilegedApproval *PrivilegedApproval) (*PrivilegedApproval, *http.Response, error) {
	v := new(PrivilegedApproval)
	resp, err := service.Client.NewRequestDo("POST", mgmtConfig+service.Client.Config.CustomerID+privilegedApprovalEndpoint, nil, privilegedApproval, &v)
	if err != nil {
		return nil, nil, err
	}
	return v, resp, nil
}

func (service *Service) Update(approvalID string, privilegedApproval *PrivilegedApproval) (*http.Response, error) {
	path := fmt.Sprintf("%v/%v", mgmtConfig+service.Client.Config.CustomerID+privilegedApprovalEndpoint, approvalID)
	resp, err := service.Client.NewRequestDo("PUT", path, nil, privilegedApproval, nil)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (service *Service) Delete(approvalID string) (*http.Response, error) {
	path := fmt.Sprintf("%v/%v", mgmtConfig+service.Client.Config.CustomerID+privilegedApprovalEndpoint, approvalID)
	resp, err := service.Client.NewRequestDo("DELETE", path, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (service *Service) GetAll() ([]PrivilegedApproval, *http.Response, error) {
	relativeURL := mgmtConfig + service.Client.Config.CustomerID + privilegedApprovalEndpoint
	list, resp, err := common.GetAllPagesGeneric[PrivilegedApproval](service.Client, relativeURL, "")
	if err != nil {
		return nil, nil, err
	}
	return list, resp, nil
}
