package admins

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zia/services/common"
)

const (
	adminUsersEndpoint     = "/zia/api/v1/adminUsers"
	passwordExpiryEndpoint = "/zia/api/v1/passwordExpiry/settings"
)

type AdminUsers struct {
	// Admin or auditor's user ID
	ID int `json:"id,omitempty"`

	// Admin or auditor's login name. loginName is in email format and uses the domain name associated to the Zscaler account
	LoginName string `json:"loginName,omitempty"`

	// Admin or auditor's username
	UserName string `json:"userName,omitempty"`

	// Admin or auditor's email address
	Email string `json:"email,omitempty"`

	// Additional information about the admin or auditor
	Comments string `json:"comments,omitempty"`

	// Indicates whether or not the admin account is disabled
	Disabled bool `json:"disabled,omitempty"`

	// The admin's password. If admin single sign-on (SSO) is disabled, then this field is mandatory for POST requests. This information is not provided in a GET response."
	Password string `json:"password,omitempty"`

	PasswordLastModifiedTime int `json:"pwdLastModifiedTime,omitempty"`

	// Indicates whether or not the admin can be edited or deleted
	IsNonEditable bool `json:"isNonEditable,omitempty"`

	// The default is true when SAML Authentication is disabled. When SAML Authentication is enabled, this can be set to false in order to force the admin to login via SSO only.
	IsPasswordLoginAllowed bool `json:"isPasswordLoginAllowed,omitempty"`

	// Indicates whether or not an admin's password has expired
	IsPasswordExpired bool `json:"isPasswordExpired,omitempty"`

	// Indicates whether the user is an auditor. This attribute is subject to change.
	IsAuditor bool `json:"isAuditor,omitempty"`

	// Communication for Security Report is enabled.
	IsSecurityReportCommEnabled bool `json:"isSecurityReportCommEnabled,omitempty"`

	// Communication setting for Service Update
	IsServiceUpdateCommEnabled bool `json:"isServiceUpdateCommEnabled,omitempty"`

	// Communication setting for Product Update
	IsProductUpdateCommEnabled bool `json:"isProductUpdateCommEnabled,omitempty"`

	// Indicates whether or not Executive Insights App access is enabled for the admin
	IsExecMobileAppEnabled bool `json:"isExecMobileAppEnabled,omitempty"`

	// Only applicable for the LOCATION_GROUP admin scope type, in which case this attribute gives the list of ID/name pairs of locations within the location group. The attribute name is subject to change
	AdminScopeGroupMemberEntities []common.IDNameExtensions `json:"adminScopescopeGroupMemberEntities,omitempty"`

	// Based on the admin scope type, the entities can be the ID/name pair of departments, locations, or location groups. The attribute name is subject to change
	AdminScopeEntities []common.IDNameExtensions `json:"adminScopeScopeEntities,omitempty"`

	// The admin's scope. A scope is required for admins, but not applicable to auditors. This attribute is subject to change
	AdminScopeType string `json:"adminScopeType,omitempty"`

	// Role of the admin. This is not required for an auditor
	Role *Role `json:"role,omitempty"`

	// Read-only information about a Executive Insights App token, if it exists
	ExecMobileAppTokens []ExecMobileAppTokens `json:"execMobileAppTokens,omitempty"`
}

type Role struct {
	// Identifier that uniquely identifies an entity
	ID int `json:"id,omitempty"`

	// The configured name of the entity
	Name         string                 `json:"name,omitempty"`
	IsNameL10Tag bool                   `json:"isNameL10nTag,omitempty"`
	Extensions   map[string]interface{} `json:"extensions,omitempty"`
}

type ExecMobileAppTokens struct {
	Cloud       string `json:"cloud,omitempty"`
	OrgId       int    `json:"orgId,omitempty"`
	Name        string `json:"name,omitempty"`
	TokenId     string `json:"tokenId,omitempty"`
	Token       string `json:"token,omitempty"`
	TokenExpiry int    `json:"tokenExpiry,omitempty"`
	CreateTime  int    `json:"createTime,omitempty"`
	DeviceId    string `json:"deviceId,omitempty"`
	DeviceName  string `json:"deviceName,omitempty"`
}

type PasswordExpiry struct {
	PasswordExpirationEnabled bool `json:"passwordExpirationEnabled,omitempty"`
	PasswordExpiryDays        int  `json:"passwordExpiryDays,omitempty"`
}

func GetAdminUsers(ctx context.Context, service *zscaler.Service, adminUserId int) (*AdminUsers, error) {
	v := new(AdminUsers)
	relativeURL := fmt.Sprintf("%s/%d", adminUsersEndpoint, adminUserId)
	err := service.Client.Read(ctx, relativeURL, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func GetAdminUsersByLoginName(ctx context.Context, service *zscaler.Service, adminUsersLoginName string) (*AdminUsers, error) {
	adminUsers, err := GetAllAdminUsers(ctx, service)
	if err != nil {
		return nil, err
	}
	for _, adminUser := range adminUsers {
		if strings.EqualFold(adminUser.LoginName, adminUsersLoginName) {
			return &adminUser, nil
		}
	}
	return nil, fmt.Errorf("no admin login found with name: %s", adminUsersLoginName)
}

func GetAdminByUsername(ctx context.Context, service *zscaler.Service, adminUsername string) (*AdminUsers, error) {
	adminUsers, err := GetAllAdminUsers(ctx, service)
	if err != nil {
		return nil, err
	}
	for _, adminUser := range adminUsers {
		if strings.EqualFold(adminUser.UserName, adminUsername) {
			return &adminUser, nil
		}
	}
	return nil, fmt.Errorf("no admin found with username: %s", adminUsername)
}

func CreateAdminUser(ctx context.Context, service *zscaler.Service, adminUser AdminUsers) (*AdminUsers, error) {
	resp, err := service.Client.Create(ctx, adminUsersEndpoint, adminUser)
	if err != nil {
		return nil, err
	}
	res, ok := resp.(*AdminUsers)
	if !ok {
		return nil, fmt.Errorf("couldn't marshal response to a valid objectm: %#v", resp)
	}
	return res, nil
}

func UpdateAdminUser(ctx context.Context, service *zscaler.Service, adminUserID int, adminUser AdminUsers) (*AdminUsers, error) {
	path := fmt.Sprintf("%s/%d", adminUsersEndpoint, adminUserID)
	resp, err := service.Client.UpdateWithPut(ctx, path, adminUser)
	if err != nil {
		return nil, err
	}
	res, _ := resp.(AdminUsers)
	return &res, err
}

func DeleteAdminUser(ctx context.Context, service *zscaler.Service, adminUserID int) (*http.Response, error) {
	err := service.Client.Delete(ctx, fmt.Sprintf("%s/%d", adminUsersEndpoint, adminUserID))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func GetAllAdminUsers(ctx context.Context, service *zscaler.Service) ([]AdminUsers, error) {
	var adminUsers []AdminUsers
	err := common.ReadAllPages(ctx, service.Client, adminUsersEndpoint+"?includeAuditorUsers=true&includeAdminUsers=true", &adminUsers)
	return adminUsers, err
}

func GetPasswordExpirySettings(ctx context.Context, service *zscaler.Service) ([]PasswordExpiry, error) {
	var expiry []PasswordExpiry
	err := common.ReadAllPages(ctx, service.Client, passwordExpiryEndpoint, &expiry)
	return expiry, err
}

func UpdatePasswordExpirySettings(ctx context.Context, service *zscaler.Service, advancedSettings *PasswordExpiry) (*PasswordExpiry, *http.Response, error) {
	resp, err := service.Client.UpdateWithPut(ctx, (passwordExpiryEndpoint), *advancedSettings)
	if err != nil {
		return nil, nil, err
	}
	updatedSettings, _ := resp.(*PasswordExpiry)

	service.Client.GetLogger().Printf("[DEBUG]returning updates password expiry from update: %d", updatedSettings)
	return updatedSettings, nil, nil
}
