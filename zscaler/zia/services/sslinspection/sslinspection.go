package sslinspection

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zia/services/common"
)

const (
	sslInspectionEndpoint = "/zia/api/v1/sslInspectionRules"
)

type SSLInspectionRules struct {
	// System generated identifier for the SSL inspection rule
	ID int `json:"id,omitempty"`

	// The name of the SSL Inspection rule
	Name string `json:"name,omitempty"`

	// Additional information about the SSL Inspection rule
	Description string `json:"description,omitempty"`

	// The action taken when traffic matches the SSL Inspection rule criteria.
	Action Action `json:"action,omitempty"`

	// Enables or disables the SSL Inspection rule.
	State string `json:"state,omitempty"`

	// Access privilege to this rule based on the admin's RBA
	AccessControl string `json:"accessControl,omitempty"`

	// Order of rule execution with respect to other SSL inspection rules.
	Order int `json:"order,omitempty"`

	// Admin rank of the admin who creates this rule
	Rank int `json:"rank,omitempty"`

	// Name-ID pairs of locations for which rule must be applied
	Locations []common.IDNameExtensions `json:"locations,omitempty"`

	// Name-ID pairs of the location groups to which the rule must be applied.
	LocationGroups []common.IDNameExtensions `json:"locationGroups,omitempty"`

	// Name-ID pairs of groups for which rule must be applied
	Groups []common.IDNameExtensions `json:"groups,omitempty"`

	// Name-ID pairs of departments for which rule must be applied
	Departments []common.IDNameExtensions `json:"departments,omitempty"`

	// Name-ID pairs of users for which rule must be applied
	Users []common.IDNameExtensions `json:"users,omitempty"`

	// Zscaler Client Connector device platforms for which the rule must be applied.
	Platforms []string `json:"platforms,omitempty"`

	// When set to true, the rule is applied to remote users that use PAC with Kerberos authentication.
	RoadWarriorForKerberos bool `json:"roadWarriorForKerberos"`

	// List of URL categories for which rule must be applied
	URLCategories []string `json:"urlCategories,omitempty"`

	// The list of cloud applications to which the DLP policy rule must be applied.
	CloudApplications []string `json:"cloudApplications,omitempty"`

	// User agent type list
	UserAgentTypes []string `json:"userAgentTypes,omitempty"`

	// List of device trust levels for which the rule must be applied. This field is applicable for devices that are managed using Zscaler Client Connector. The trust levels are assigned to the devices based on your posture configurations in the Zscaler Client Connector Portal. If no value is set, this field is ignored during the policy evaluation.
	DeviceTrustLevels []string `json:"deviceTrustLevels,omitempty"`

	// This field is applicable for devices that are managed using Zscaler Client Connector. If no value is set, this field is ignored during the policy evaluation.
	DeviceGroups []common.IDNameExtensions `json:"deviceGroups,omitempty"`

	// Name-ID pairs of devices for which rule must be applied. Specifies devices that are managed using Zscaler Client Connector. If no value is set, this field is ignored during the policy evaluation.
	Devices []common.IDNameExtensions `json:"devices,omitempty"`

	// Timestamp when the rule was last modified. Ignore if the request is POST or PUT
	LastModifiedTime int `json:"lastModifiedTime,omitempty"`

	// Admin user that last modified the rule. Ignore if the request is POST or PUT.
	LastModifiedBy *common.IDNameExtensions `json:"lastModifiedBy,omitempty"`

	// User-defined destination IP address groups on which the rule is applied. If not set, the rule is not restricted to a specific destination IP address group.
	// Note: For organizations that have enabled IPv6, the destIpv6Groups field lists the IPv6 source address groups for which the rule is applicable.
	DestIpGroups []common.IDNameExtensions `json:"destIpGroups,omitempty"`

	// Source IP address groups for which the rule is applicable.
	// If not set, the rule is not restricted to a specific source IP address group.
	SourceIPGroups []common.IDNameExtensions `json:"sourceIpGroups,omitempty"`

	// The proxy chaining gateway for which this rule is applicable.
	// Ignore if the forwarding method is not Proxy Chaining.
	ProxyGateways []common.IDNameExtensions `json:"proxyGateways,omitempty"`

	// Name-ID pairs of rule labels associated with the rule
	Labels []common.IDNameExtensions `json:"labels,omitempty"`

	// The time intervals during which the rule applies
	TimeWindows []common.IDNameExtensions `json:"timeWindows,omitempty"`

	// The list of ZPA Application Segments for which this rule is applicable. This field is applicable only for the ZPA Gateway forwarding method.
	ZPAAppSegments []common.ZPAAppSegments `json:"zpaAppSegments,omitempty"`

	// The list of preconfigured workload groups to which the policy must be applied.
	WorkloadGroups []common.IDName `json:"workloadGroups,omitempty"`

	// If set to true, the default rule is applied
	DefaultRule bool `json:"defaultRule,omitempty"`

	// If set to true, a predefined rule is applied
	Predefined bool `json:"predefined,omitempty"`
}

type Action struct {
	// Supported values: "BLOCK", "DECRYPT", "DO_NOT_DECRYPT",
	Type                       string                  `json:"type,omitempty"`
	ShowEUN                    bool                    `json:"showEUN,omitempty"`
	ShowEUNATP                 bool                    `json:"showEUNATP,omitempty"`
	OverrideDefaultCertificate bool                    `json:"overrideDefaultCertificate,omitempty"`
	SSLInterceptionCert        *SSLInterceptionCert    `json:"sslInterceptionCert,omitempty"`
	DecryptSubActions          *DecryptSubActions      `json:"decryptSubActions,omitempty"`
	DoNotDecryptSubActions     *DoNotDecryptSubActions `json:"doNotDecryptSubActions,omitempty"`
}

type SSLInterceptionCert struct {
	ID                 int    `json:"id,omitempty"`
	Name               string `json:"name,omitempty"`
	DefaultCertificate bool   `json:"defaultCertificate,omitempty"`
}

type DoNotDecryptSubActions struct {
	BypassOtherPolicies bool `json:"bypassOtherPolicies,omitempty"`
	// ALLOW, BLOCK, PASS_THRU
	ServerCertificates              string `json:"serverCertificates,omitempty"`
	OcspCheck                       bool   `json:"ocspCheck,omitempty"`
	BlockSslTrafficWithNoSniEnabled bool   `json:"blockSslTrafficWithNoSniEnabled,omitempty"`
	MinTLSVersion                   string `json:"minTLSVersion,omitempty"`
}

type DecryptSubActions struct {
	ServerCertificates              string `json:"serverCertificates,omitempty"`
	OcspCheck                       bool   `json:"ocspCheck,omitempty"`
	BlockSslTrafficWithNoSniEnabled bool   `json:"blockSslTrafficWithNoSniEnabled,omitempty"`
	MinClientTLSVersion             string `json:"minClientTLSVersion,omitempty"`
	MinServerTLSVersion             string `json:"minServerTLSVersion,omitempty"`
	BlockUndecrypt                  bool   `json:"blockUndecrypt,omitempty"`
	HTTP2Enabled                    bool   `json:"http2Enabled,omitempty"`
}

func Get(ctx context.Context, service *zscaler.Service, ruleID int) (*SSLInspectionRules, error) {
	var rule SSLInspectionRules
	err := service.Client.Read(ctx, fmt.Sprintf("%s/%d", sslInspectionEndpoint, ruleID), &rule)
	if err != nil {
		return nil, err
	}

	service.Client.GetLogger().Printf("[DEBUG]Returning ssl inpection from Get: %d", rule.ID)
	return &rule, nil
}

func GetByName(ctx context.Context, service *zscaler.Service, ruleName string) (*SSLInspectionRules, error) {
	var rules []SSLInspectionRules
	err := common.ReadAllPages(ctx, service.Client, sslInspectionEndpoint, &rules)
	if err != nil {
		return nil, err
	}
	for _, rule := range rules {
		if strings.EqualFold(rule.Name, ruleName) {
			return &rule, nil
		}
	}
	return nil, fmt.Errorf("no ssl inpection rule found with name: %s", ruleName)
}

func Create(ctx context.Context, service *zscaler.Service, rule *SSLInspectionRules) (*SSLInspectionRules, error) {
	// Apply the validation
	if err := validateSSLInspectionRule(rule); err != nil {
		return nil, err
	}

	resp, err := service.Client.Create(ctx, sslInspectionEndpoint, *rule)
	if err != nil {
		return nil, err
	}

	createdRules, ok := resp.(*SSLInspectionRules)
	if !ok {
		return nil, errors.New("object returned from api was not a rule Pointer")
	}

	service.Client.GetLogger().Printf("[DEBUG]returning rule from create: %d", createdRules.ID)
	return createdRules, nil
}

func Update(ctx context.Context, service *zscaler.Service, ruleID int, rule *SSLInspectionRules) (*SSLInspectionRules, error) {
	// Apply the validation
	if err := validateSSLInspectionRule(rule); err != nil {
		return nil, err
	}

	resp, err := service.Client.UpdateWithPut(ctx, fmt.Sprintf("%s/%d", sslInspectionEndpoint, ruleID), *rule)
	if err != nil {
		return nil, err
	}
	updatedRules, _ := resp.(*SSLInspectionRules)
	service.Client.GetLogger().Printf("[DEBUG]returning ssl inpection from update: %d", updatedRules.ID)
	return updatedRules, nil
}

func Delete(ctx context.Context, service *zscaler.Service, ruleID int) (*http.Response, error) {
	err := service.Client.Delete(ctx, fmt.Sprintf("%s/%d", sslInspectionEndpoint, ruleID))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func GetAll(ctx context.Context, service *zscaler.Service) ([]SSLInspectionRules, error) {
	var rules []SSLInspectionRules
	err := common.ReadAllPages(ctx, service.Client, sslInspectionEndpoint, &rules)
	return rules, err
}

// validateSSLInspectionRule checks that the provided rule adheres to the business logic.
func validateSSLInspectionRule(rule *SSLInspectionRules) error {
	// Quick nil-check
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}

	// Validate based on action type
	switch rule.Action.Type {
	case "DECRYPT":
		// Ensure decryptSubActions is set and not empty
		if rule.Action.DecryptSubActions == nil {
			return fmt.Errorf("when action.type is 'DECRYPT', decryptSubActions block must be set")
		}

		// Ensure all required fields in decryptSubActions are set
		if rule.Action.DecryptSubActions.ServerCertificates == "" ||
			rule.Action.DecryptSubActions.MinClientTLSVersion == "" ||
			rule.Action.DecryptSubActions.MinServerTLSVersion == "" {
			return fmt.Errorf("when action.type is 'DECRYPT', all required fields in decryptSubActions must be set")
		}

		if !rule.Action.OverrideDefaultCertificate && rule.Action.SSLInterceptionCert != nil {
			return fmt.Errorf("when action.type is 'DECRYPT' and overrideDefaultCertificate is false, sslInterceptionCert cannot be set")
		}

		// Ensure showEUN and showEUNATP are not set
		if rule.Action.ShowEUN || rule.Action.ShowEUNATP {
			return fmt.Errorf("when action.type is 'DECRYPT', neither showEUN nor showEUNATP can be set")
		}

	case "DO_NOT_DECRYPT":
		// Ensure doNotDecryptSubActions is set and not empty
		if rule.Action.DoNotDecryptSubActions == nil {
			return fmt.Errorf("when action.type is 'DO_NOT_DECRYPT', doNotDecryptSubActions block must be set")
		}

		// If bypassOtherPolicies is true, serverCertificates and minTLSVersion cannot be set
		if rule.Action.DoNotDecryptSubActions.BypassOtherPolicies {
			if rule.Action.DoNotDecryptSubActions.ServerCertificates != "" ||
				rule.Action.DoNotDecryptSubActions.MinTLSVersion != "" {
				return fmt.Errorf("when action.type is 'DO_NOT_DECRYPT' and bypassOtherPolicies is true, serverCertificates and minTLSVersion cannot be set")
			}
		} else {
			// If bypassOtherPolicies is false, ensure serverCertificates and minTLSVersion are set
			if rule.Action.DoNotDecryptSubActions.ServerCertificates == "" ||
				rule.Action.DoNotDecryptSubActions.MinTLSVersion == "" {
				return fmt.Errorf("when action.type is 'DO_NOT_DECRYPT' and bypassOtherPolicies is false, serverCertificates and minTLSVersion must be set")
			}
		}

	case "BLOCK":
		// Ensure decryptSubActions and doNotDecryptSubActions are not set
		if rule.Action.DecryptSubActions != nil || rule.Action.DoNotDecryptSubActions != nil {
			return fmt.Errorf("when action.type is 'BLOCK', neither decryptSubActions nor doNotDecryptSubActions can be set")
		}

		// When action.type is BLOCK and overrideDefaultCertificate is false,
		// sslInterceptionCert cannot be set
		if !rule.Action.OverrideDefaultCertificate && rule.Action.SSLInterceptionCert != nil {
			return fmt.Errorf("when action.type is 'BLOCK' and overrideDefaultCertificate is false, sslInterceptionCert cannot be set")
		}

		// Ensure showEUNATP is not set
		if rule.Action.ShowEUNATP {
			return fmt.Errorf("when action.type is 'BLOCK', showEUNATP cannot be set")
		}

	default:
		return fmt.Errorf("invalid action type: %s", rule.Action.Type)
	}

	return nil
}
