package intermediatecacertificates

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
	intermediateCaCertificatesEndpoint = "/zia/api/v1/intermediateCaCertificate"
	intCADownloadAttestationEndpoint   = "/zia/api/v1/intermediateCaCertificate/downloadAttestation"
	intCADownloadCSREndpoint           = "/zia/api/v1/intermediateCaCertificate/downloadCsr"
	intCADownloadPublicKeyEndpoint     = "/zia/api/v1/intermediateCaCertificate/downloadPublicKey"
	intCAGenerateCSREndpoint           = "/zia/api/v1/intermediateCaCertificate/generateCsr"
	intCAFinalizeCSREndpoint           = "/zia/api/v1/intermediateCaCertificate/finalizeCert"
	intCAKeyPairEndpoint               = "/zia/api/v1/intermediateCaCertificate/keyPair"
	intCACertMakeDefaultEndpoint       = "/zia/api/v1/intermediateCaCertificate/makeDefault"
	intCAReadyToUseEndpoint            = "/zia/api/v1/intermediateCaCertificate/readyToUse"
	intCAShowCertEndpoint              = "/zia/api/v1/intermediateCaCertificate/showCert"
	intCAShowCSREndpoint               = "/zia/api/v1/intermediateCaCertificate/showCsr"
	intCAUploadCert                    = "/zia/api/v1/intermediateCaCertificate/uploadCert"
	intCAUploadCertChain               = "/zia/api/v1/intermediateCaCertificate/uploadCertChain"
	intCAVerifyKeyAttestation          = "/zia/api/v1/intermediateCaCertificate/verifyKeyAttestation"
)

type IntermediateCACertificate struct {
	// Unique identifier for the intermediate CA certificat
	ID int `json:"id"`

	// Name of the intermediate CA certificate
	Name string `json:"name,omitempty"`

	// Description for the intermediate CA certificate
	Description string `json:"description,omitempty"`

	// Type of the intermediate CA certificate. Available types: Zscaler’s intermediate CA certificate (provided by Zscaler), custom intermediate certificate with software protection, and custom intermediate certificate with cloud HSM protection.
	Type string `json:"type,omitempty"`

	// Location of the HSM resources. Required for custom intermediate CA certificates with cloud HSM protection
	Region string `json:"region,omitempty"`

	// Determines whether the intermediate CA certificate is enabled or disabled for SSL inspection. Subscription to cloud HSM protection allows a maximum of four active certificates for SSL inspection at a time, whereas software protection subscription allows only one active certificate
	Status string `json:"status,omitempty"`

	// If set to true, the intermediate CA certificate is the default intermediate certificate. Only one certificate can be marked as the default intermediate certificate at a time
	DefaultCertificate bool `json:"defaultCertificate,omitempty"`

	// Start date of the intermediate CA certificate’s validity period
	CertStartDate int `json:"certStartDate,omitempty"`

	// Expiration date of the intermediate CA certificate’s validity period
	CertExpDate int `json:"certExpDate,omitempty"`

	// Tracks the progress of the intermediate CA certificate in the configuration workflow
	CurrentState string `json:"currentState,omitempty"`

	// Public key in the HSM key pair generated for the intermediate CA certificate
	PublicKey string `json:"publicKey,omitempty"`

	// Timestamp when the HSM key was generated
	KeyGenerationTime int `json:"keyGenerationTime,omitempty"`

	// Timestamp when the attestation for the HSM key was verified
	HSMAttestationVerifiedTime int `json:"hsmAttestationVerifiedTime,omitempty"`

	// Certificate Signing Request (CSR) file name
	CSRFileName string `json:"csrFileName,omitempty"`

	// Timestamp when the Certificate Signing Request (CSR) was generated
	CSRGenerationTime int `json:"csrGenerationTime,omitempty"`
}

type CertSigningRequest struct {
	// Unique identifier for the intermediate CA certificate
	CertID int `json:"certId"`

	// Name of the CSR file
	CSRFileName string `json:"csrFileName,omitempty"`

	// Common Name (CN) of your organization’s domain, such as zscaler.com
	CommName string `json:"commName,omitempty"`

	// Name of your organization or company
	ORGName string `json:"orgName,omitempty"`

	// Name of your department or division
	DeptName string `json:"deptName,omitempty"`

	// Name of the city or town where your organization is located
	City string `json:"city,omitempty"`

	// State, province, region, or county where your organization is located
	State string `json:"state,omitempty"`

	// Country where your organization is located
	Country string `json:"country,omitempty"`

	// Key size to be used in the encryption algorithm in bits. Default size: 2048 bits
	KeySize int `json:"keySize,omitempty"`

	// Signature algorithm to be used for generating intermediate CA certificate. Default value: SHA256
	SignatureAlgorithm string `json:"signatureAlgorithm,omitempty"`

	// The path length constraint for the intermediate CA certificate. Default values: 0 for cloud HSM, 1 for software protection
	PathLengthConstraint int `json:"pathLengthConstraint,omitempty"`
}

func GetCertificate(ctx context.Context, service *zscaler.Service, certID int) (*IntermediateCACertificate, error) {
	var intermediateCACertificate IntermediateCACertificate
	err := service.Client.Read(ctx, fmt.Sprintf("%s/%d", intermediateCaCertificatesEndpoint, certID), &intermediateCACertificate)
	if err != nil {
		return nil, err
	}

	service.Client.GetLogger().Printf("[DEBUG]Returning intermediate ca certificate from Get: %d", intermediateCACertificate.ID)
	return &intermediateCACertificate, nil
}

func GetByName(ctx context.Context, service *zscaler.Service, certName string) (*IntermediateCACertificate, error) {
	var intermediateCACertificate []IntermediateCACertificate
	err := common.ReadAllPages(ctx, service.Client, intermediateCaCertificatesEndpoint, &intermediateCACertificate)
	if err != nil {
		return nil, err
	}
	for _, certificate := range intermediateCACertificate {
		if strings.EqualFold(certificate.Name, certName) {
			return &certificate, nil
		}
	}
	return nil, fmt.Errorf("no intermediate ca certificate found with name: %s", certName)
}

func GetDownloadAttestation(ctx context.Context, service *zscaler.Service, certID int) (*IntermediateCACertificate, error) {
	var downloadAttestation IntermediateCACertificate
	err := service.Client.Read(ctx, fmt.Sprintf("%s/%d", intCADownloadAttestationEndpoint, certID), &downloadAttestation)
	if err != nil {
		return nil, err
	}

	service.Client.GetLogger().Printf("[DEBUG]Returning downloaded attestation from Get: %d", downloadAttestation.ID)
	return &downloadAttestation, nil
}

func GetDownloadCSR(ctx context.Context, service *zscaler.Service, certID int) (*IntermediateCACertificate, error) {
	var downloadCSR IntermediateCACertificate
	err := service.Client.Read(ctx, fmt.Sprintf("%s/%d", intCADownloadCSREndpoint, certID), &downloadCSR)
	if err != nil {
		return nil, err
	}

	service.Client.GetLogger().Printf("[DEBUG]Returning downloaded csr from Get: %d", downloadCSR.ID)
	return &downloadCSR, nil
}

func GetDownloadPublicKey(ctx context.Context, service *zscaler.Service, certID int) (*IntermediateCACertificate, error) {
	var downloadPublicKey IntermediateCACertificate
	err := service.Client.Read(ctx, fmt.Sprintf("%s/%d", intCADownloadPublicKeyEndpoint, certID), &downloadPublicKey)
	if err != nil {
		return nil, err
	}

	service.Client.GetLogger().Printf("[DEBUG]Returning downloaded public key from Get: %d", downloadPublicKey.ID)
	return &downloadPublicKey, nil
}

func GetIntCAReadyToUse(ctx context.Context, service *zscaler.Service) ([]IntermediateCACertificate, error) {
	var readyToUse []IntermediateCACertificate
	err := service.Client.Read(ctx, intCAReadyToUseEndpoint, &readyToUse)
	if err != nil {
		return nil, err
	}

	service.Client.GetLogger().Printf("[DEBUG]Returning downloaded public key from Get: %v", readyToUse)
	return readyToUse, nil
}

func GetShowCert(ctx context.Context, service *zscaler.Service, certID int) (*CertSigningRequest, error) {
	var showCert CertSigningRequest
	err := service.Client.Read(ctx, fmt.Sprintf("%s/%d", intCAShowCertEndpoint, certID), &showCert)
	if err != nil {
		return nil, err
	}

	service.Client.GetLogger().Printf("[DEBUG]Returning info about signed intrermediate CA certificates from Get: %d", showCert.CertID)
	return &showCert, nil
}

func GetShowCSR(ctx context.Context, service *zscaler.Service, certID int) (*CertSigningRequest, error) {
	var showCSR CertSigningRequest
	err := service.Client.Read(ctx, fmt.Sprintf("%s/%d", intCAShowCSREndpoint, certID), &showCSR)
	if err != nil {
		return nil, err
	}

	service.Client.GetLogger().Printf("[DEBUG]Returning info about signed intermediate CA certificates from Get: %d", showCSR.CertID)
	return &showCSR, nil
}

func GetAll(ctx context.Context, service *zscaler.Service) ([]IntermediateCACertificate, error) {
	var intermediateCACertificates []IntermediateCACertificate
	err := common.ReadAllPages(ctx, service.Client, intermediateCaCertificatesEndpoint, &intermediateCACertificates)
	if err != nil {
		return nil, err
	}
	return intermediateCACertificates, nil
}

func CreateIntCACertificate(ctx context.Context, service *zscaler.Service, cert *IntermediateCACertificate) (*IntermediateCACertificate, error) {
	resp, err := service.Client.Create(ctx, intermediateCaCertificatesEndpoint, *cert)
	if err != nil {
		return nil, err
	}

	createdIntermediateCACert, ok := resp.(*IntermediateCACertificate)
	if !ok {
		return nil, errors.New("object returned from api was not an intermediate ca certificate Pointer")
	}

	service.Client.GetLogger().Printf("[DEBUG]returning intermediate ca certificate from create: %d", createdIntermediateCACert.ID)
	return createdIntermediateCACert, nil
}

func CreateIntCAGenerateCSR(ctx context.Context, service *zscaler.Service, cert *IntermediateCACertificate) (*IntermediateCACertificate, error) {
	resp, err := service.Client.Create(ctx, intCAGenerateCSREndpoint, *cert)
	if err != nil {
		return nil, err
	}

	createdIntCAGenerateCSR, ok := resp.(*IntermediateCACertificate)
	if !ok {
		return nil, errors.New("object returned from api was not an intermediate ca certificate Pointer")
	}

	service.Client.GetLogger().Printf("[DEBUG]returning intermediate ca certificate from create: %d", createdIntCAGenerateCSR.ID)
	return createdIntCAGenerateCSR, nil
}

func CreateIntCAFinalizeCert(ctx context.Context, service *zscaler.Service, cert *IntermediateCACertificate) (*IntermediateCACertificate, error) {
	resp, err := service.Client.Create(ctx, intCAFinalizeCSREndpoint, *cert)
	if err != nil {
		return nil, err
	}

	createdIntCAFinalizeCSR, ok := resp.(*IntermediateCACertificate)
	if !ok {
		return nil, errors.New("object returned from api was not an intermediate ca certificate Pointer")
	}

	service.Client.GetLogger().Printf("[DEBUG]returning intermediate ca certificate from create: %d", createdIntCAFinalizeCSR.ID)
	return createdIntCAFinalizeCSR, nil
}

func CreateIntCAKeyPair(ctx context.Context, service *zscaler.Service, keyPair *IntermediateCACertificate) (*IntermediateCACertificate, error) {
	resp, err := service.Client.Create(ctx, intCAKeyPairEndpoint, *keyPair)
	if err != nil {
		return nil, err
	}

	createdIntCAKeyPair, ok := resp.(*IntermediateCACertificate)
	if !ok {
		return nil, errors.New("object returned from api was not an intermediate ca certificate Pointer")
	}

	service.Client.GetLogger().Printf("[DEBUG]returning intermediate ca certificate from create: %d", createdIntCAKeyPair.ID)
	return createdIntCAKeyPair, nil
}

func CreateUploadCert(ctx context.Context, service *zscaler.Service, certID *IntermediateCACertificate) (*IntermediateCACertificate, error) {
	resp, err := service.Client.Create(ctx, intCAUploadCert, *certID)
	if err != nil {
		return nil, err
	}

	createdIntCAUploadCert, ok := resp.(*IntermediateCACertificate)
	if !ok {
		return nil, errors.New("object returned from api was not an intermediate ca certificate Pointer")
	}

	service.Client.GetLogger().Printf("[DEBUG]returning uploaded customer intermediate ca certificate from create: %d", createdIntCAUploadCert.ID)
	return createdIntCAUploadCert, nil
}

func CreateUploadCertChain(ctx context.Context, service *zscaler.Service, certID *IntermediateCACertificate) (*IntermediateCACertificate, error) {
	resp, err := service.Client.Create(ctx, intCAUploadCertChain, *certID)
	if err != nil {
		return nil, err
	}

	createdIntCAUploadChain, ok := resp.(*IntermediateCACertificate)
	if !ok {
		return nil, errors.New("object returned from api was not an intermediate ca certificate Pointer")
	}

	service.Client.GetLogger().Printf("[DEBUG]returning uploaded certificate chain from create: %d", createdIntCAUploadChain.ID)
	return createdIntCAUploadChain, nil
}

func CreateVerifyKeyAttestation(ctx context.Context, service *zscaler.Service, certID *IntermediateCACertificate) (*IntermediateCACertificate, error) {
	resp, err := service.Client.Create(ctx, intCAVerifyKeyAttestation, *certID)
	if err != nil {
		return nil, err
	}

	createdVerifyKeyAttestation, ok := resp.(*IntermediateCACertificate)
	if !ok {
		return nil, errors.New("object returned from api was not an intermediate ca certificate Pointer")
	}

	service.Client.GetLogger().Printf("[DEBUG]returning key attestation from create: %d", createdVerifyKeyAttestation.ID)
	return createdVerifyKeyAttestation, nil
}

func Update(ctx context.Context, service *zscaler.Service, certID int, certificates *IntermediateCACertificate) (*IntermediateCACertificate, error) {
	resp, err := service.Client.UpdateWithPut(ctx, fmt.Sprintf("%s/%d", intermediateCaCertificatesEndpoint, certID), *certificates)
	if err != nil {
		return nil, err
	}
	updatedIntermediateCaCert, _ := resp.(*IntermediateCACertificate)
	service.Client.GetLogger().Printf("[DEBUG]returning intermediate ca certificate from update: %d", updatedIntermediateCaCert.ID)
	return updatedIntermediateCaCert, nil
}

func UpdateMakeDefault(ctx context.Context, service *zscaler.Service, certID int, certificates *IntermediateCACertificate) (*IntermediateCACertificate, error) {
	resp, err := service.Client.UpdateWithPut(ctx, fmt.Sprintf("%s/%d", intCACertMakeDefaultEndpoint, certID), *certificates)
	if err != nil {
		return nil, err
	}
	updatedIntermediateCaCert, _ := resp.(*IntermediateCACertificate)
	service.Client.GetLogger().Printf("[DEBUG]returning default certificate from update: %d", updatedIntermediateCaCert.ID)
	return updatedIntermediateCaCert, nil
}

func Delete(ctx context.Context, service *zscaler.Service, certID int) (*http.Response, error) {
	err := service.Client.Delete(ctx, fmt.Sprintf("%s/%d", intermediateCaCertificatesEndpoint, certID))
	if err != nil {
		return nil, err
	}

	return nil, nil
}
