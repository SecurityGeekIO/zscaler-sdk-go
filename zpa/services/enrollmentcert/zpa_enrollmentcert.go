package enrollmentcert

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa/services"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa/services/common"
)

const (
	mgmtConfigV1           = "/mgmtconfig/v1/admin/customers/"
	mgmtConfigV2           = "/mgmtconfig/v2/admin/customers/"
	enrollmentCertEndpoint = "/enrollmentCert"
)

type EnrollmentCert struct {
	AllowSigning            bool   `json:"allowSigning,omitempty"`
	Cname                   string `json:"cName,omitempty"`
	Certificate             string `json:"certificate,omitempty"`
	ClientCertType          string `json:"clientCertType,omitempty"`
	CreationTime            string `json:"creationTime,omitempty"`
	CSR                     string `json:"csr,omitempty"`
	Description             string `json:"description,omitempty"`
	ID                      string `json:"id,omitempty"`
	IssuedBy                string `json:"issuedBy,omitempty"`
	IssuedTo                string `json:"issuedTo,omitempty"`
	ModifiedBy              string `json:"modifiedBy,omitempty"`
	ModifiedTime            string `json:"modifiedTime,omitempty"`
	Name                    string `json:"name,omitempty"`
	ParentCertID            string `json:"parentCertId,omitempty"`
	ParentCertName          string `json:"parentCertName,omitempty"`
	PrivateKey              string `json:"privateKey,omitempty"`
	PrivateKeyPresent       bool   `json:"privateKeyPresent,omitempty"`
	SerialNo                string `json:"serialNo,omitempty"`
	ValidFromInEpochSec     string `json:"validFromInEpochSec,omitempty"`
	ValidToInEpochSec       string `json:"validToInEpochSec,omitempty"`
	ZrsaEncryptedPrivateKey string `json:"zrsaencryptedprivatekey,omitempty"`
	ZrsaEncryptedSessionKey string `json:"zrsaencryptedsessionkey,omitempty"`
	MicrotenantID           string `json:"microtenantId,omitempty"`
}

func Get(service *services.Service, id string) (*EnrollmentCert, *http.Response, error) {
	v := new(EnrollmentCert)
	relativeURL := fmt.Sprintf("%s/%s", mgmtConfigV1+service.Client.Config.CustomerID+enrollmentCertEndpoint, id)
	resp, err := service.Client.NewRequestDo("GET", relativeURL, common.Filter{MicroTenantID: service.MicroTenantID()}, nil, v)
	if err != nil {
		return nil, nil, err
	}

	return v, resp, nil
}

func GetByName(service *services.Service, certName string) (*EnrollmentCert, *http.Response, error) {
	relativeURL := mgmtConfigV2 + service.Client.Config.CustomerID + enrollmentCertEndpoint
	list, resp, err := common.GetAllPagesGenericWithCustomFilters[EnrollmentCert](service.Client, relativeURL, common.Filter{Search: certName, MicroTenantID: service.MicroTenantID()})
	if err != nil {
		return nil, nil, err
	}
	for _, cert := range list {
		if strings.EqualFold(cert.Name, certName) {
			return &cert, resp, nil
		}
	}
	return nil, resp, fmt.Errorf("no signing certificate named '%s' was found", certName)
}

func GetAll(service *services.Service) ([]EnrollmentCert, *http.Response, error) {
	relativeURL := mgmtConfigV2 + service.Client.Config.CustomerID + enrollmentCertEndpoint
	list, resp, err := common.GetAllPagesGenericWithCustomFilters[EnrollmentCert](service.Client, relativeURL, common.Filter{MicroTenantID: service.MicroTenantID()})
	if err != nil {
		return nil, nil, err
	}
	return list, resp, nil
}
