package policysetcontrollerv2

import (
	"fmt"
	"net/http"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa/services/common"
)

const (
	mgmtConfigV2 = "/mgmtconfig/v2/admin/customers/"
)

type PolicySet struct {
	CreationTime       string               `json:"creationTime,omitempty"`
	Description        string               `json:"description,omitempty"`
	Enabled            bool                 `json:"enabled"`
	ID                 string               `json:"id,omitempty"`
	ModifiedBy         string               `json:"modifiedBy,omitempty"`
	ModifiedTime       string               `json:"modifiedTime,omitempty"`
	Name               string               `json:"name,omitempty"`
	Sorted             bool                 `json:"sorted"`
	PolicyType         string               `json:"policyType,omitempty"`
	MicroTenantID      string               `json:"microtenantId,omitempty"`
	MicroTenantName    string               `json:"microtenantName,omitempty"`
	Rules              []PolicyRule         `json:"rules"`
	AppServerGroups    []AppServerGroups    `json:"appServerGroups"`
	AppConnectorGroups []AppConnectorGroups `json:"appConnectorGroups"`
}

type PolicyRule struct {
	ID                       string       `json:"id,omitempty"`
	Name                     string       `json:"name,omitempty"`
	Action                   string       `json:"action,omitempty"`
	ActionID                 string       `json:"actionId,omitempty"`
	BypassDefaultRule        bool         `json:"bypassDefaultRule,omitempty"`
	CustomMsg                string       `json:"customMsg,omitempty"`
	DefaultRule              bool         `json:"defaultRule,omitempty"`
	Description              string       `json:"description,omitempty"`
	IsolationDefaultRule     bool         `json:"isolationDefaultRule,omitempty"`
	CreationTime             string       `json:"creationTime,omitempty"`
	ModifiedBy               string       `json:"modifiedBy,omitempty"`
	ModifiedTime             string       `json:"modifiedTime,omitempty"`
	Operator                 string       `json:"operator,omitempty"`
	PolicySetID              string       `json:"policySetId,omitempty"`
	PolicyType               string       `json:"policyType,omitempty"`
	Priority                 string       `json:"priority,omitempty"`
	ReauthDefaultRule        bool         `json:"reauthDefaultRule,omitempty"`
	ReauthIdleTimeout        string       `json:"reauthIdleTimeout,omitempty"`
	ReauthTimeout            string       `json:"reauthTimeout,omitempty"`
	RuleOrder                string       `json:"ruleOrder,omitempty"`
	LssDefaultRule           bool         `json:"lssDefaultRule,omitempty"`
	ZpnCbiProfileID          string       `json:"zpnCbiProfileId,omitempty"`
	ZpnInspectionProfileID   string       `json:"zpnInspectionProfileId,omitempty"`
	ZpnInspectionProfileName string       `json:"zpnInspectionProfileName,omitempty"`
	MicroTenantID            string       `json:"microtenantId,omitempty"`
	MicroTenantName          string       `json:"microtenantName,omitempty"`
	Conditions               []Conditions `json:"conditions,omitempty"`
}

type Conditions struct {
	ID           string     `json:"id,omitempty"`
	CreationTime string     `json:"creationTime,omitempty"`
	ModifiedBy   string     `json:"modifiedBy,omitempty"`
	ModifiedTime string     `json:"modifiedTime,omitempty"`
	Negated      bool       `json:"negated"`
	Operator     string     `json:"operator,omitempty"`
	Operands     []Operands `json:"operands,omitempty"`
}

type Operands struct {
	ID                string        `json:"id,omitempty"`
	CreationTime      string        `json:"creationTime,omitempty"`
	ModifiedBy        string        `json:"modifiedBy,omitempty"`
	ModifiedTime      string        `json:"modifiedTime,omitempty"`
	ObjectType        string        `json:"objectType,omitempty"`
	Values            []string      `json:"values,omitempty"`
	IDPID             string        `json:"idpId,omitempty"`
	EntryValuesLHSRHS []LHSRHSValue `json:"entryValues,omitempty"`
}

type LHSRHSValue struct {
	RHS string `json:"rhs,omitempty"`
	LHS string `json:"lhs,omitempty"`
}

type AppServerGroups struct {
	ID string `json:"id,omitempty"`
}

type AppConnectorGroups struct {
	ID string `json:"id,omitempty"`
}

// POST --> mgmtconfig​/v2​/admin​/customers​/{customerId}​/policySet​/{policySetId}​/rule
func (service *Service) CreateRule(rule *PolicyRule) (*PolicyRule, *http.Response, error) {
	v := new(PolicyRule)
	path := fmt.Sprintf(mgmtConfigV2+service.Client.Config.CustomerID+"/policySet/%s/rule", rule.PolicySetID)
	resp, err := service.Client.NewRequestDo("POST", path, common.Filter{MicroTenantID: service.microTenantID}, &rule, v)
	if err != nil {
		return nil, nil, err
	}
	return v, resp, nil
}

// PUT --> mgmtconfig​/v1​/admin​/customers​/{customerId}​/policySet​/{policySetId}​/rule​/{ruleId}
func (service *Service) UpdateRule(policySetID, ruleId string, policySetRule *PolicyRule) (*http.Response, error) {
	// Correct the initialization of Conditions slice with the correct type
	if policySetRule != nil && len(policySetRule.Conditions) == 0 {
		policySetRule.Conditions = []Conditions{}
	} else {
		for i, condition := range policySetRule.Conditions {
			if len(condition.Operands) == 0 {
				policySetRule.Conditions[i].Operands = []Operands{}
			} else {
				for j, operand := range condition.Operands {
					// Clearing the ID if present, assuming you want to ensure IDs are not sent in updates
					if operand.ID != "" {
						condition.Operands[j].ID = ""
					}
					// If there's more logic to be added for handling Operands, do so here
				}
			}
		}
	}

	path := fmt.Sprintf(mgmtConfigV2+service.Client.Config.CustomerID+"/policySet/%s/rule/%s", policySetID, ruleId)
	resp, err := service.Client.NewRequestDo("PUT", path, common.Filter{MicroTenantID: service.microTenantID}, policySetRule, nil)
	if err != nil {
		return nil, err
	}
	return resp, err
}
