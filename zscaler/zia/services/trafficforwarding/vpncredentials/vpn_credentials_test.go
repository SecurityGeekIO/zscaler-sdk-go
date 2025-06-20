package vpncredentials

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zia/services/trafficforwarding/staticips"
)

const (
	maxRetries    = 3
	retryInterval = 2 * time.Second
)

// Constants for conflict retries
const (
	maxConflictRetries    = 5
	conflictRetryInterval = 1 * time.Second
)

func retryOnConflict(operation func() error) error {
	var lastErr error
	for i := 0; i < maxConflictRetries; i++ {
		lastErr = operation()
		if lastErr == nil {
			return nil
		}

		if strings.Contains(lastErr.Error(), `"code":"EDIT_LOCK_NOT_AVAILABLE"`) {
			log.Printf("Conflict error detected, retrying in %v... (Attempt %d/%d)", conflictRetryInterval, i+1, maxConflictRetries)
			time.Sleep(conflictRetryInterval)
			continue
		}

		return lastErr
	}
	return lastErr
}

func TestTrafficForwardingVPNCreds(t *testing.T) {
	ipAddress, _ := acctest.RandIpAddress("104.239.239.0/24")
	comment := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updateComment := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	rPassword := tests.TestPassword(20)

	service, err := tests.NewOneAPIClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}

	staticIP, _, err := staticips.Create(context.Background(), service, &staticips.StaticIP{
		IpAddress: ipAddress,
		Comment:   comment,
	})
	if err != nil {
		t.Fatalf("Creating static ip failed: %v", err)
	}

	defer func() {
		_, err := staticips.Delete(context.Background(), service, staticIP.ID)
		if err != nil {
			t.Errorf("Deleting static ip failed: %v", err)
		}
	}()

	cred := VPNCredentials{
		Type:         "IP",
		IPAddress:    ipAddress,
		Comments:     comment,
		PreSharedKey: rPassword,
	}

	var createdResource *VPNCredentials

	err = retryOnConflict(func() error {
		createdResource, _, err = Create(context.Background(), service, &cred)
		return err
	})
	if err != nil {
		t.Fatalf("Error creating VPN credential: %v", err)
	}

	if createdResource.ID == 0 {
		t.Fatal("Expected created resource ID to be non-empty, but got ''")
	}

	t.Logf("Created VPN credential: ID=%d, Comments=%s", createdResource.ID, createdResource.Comments)

	if createdResource.Comments != comment {
		t.Errorf("Expected created resource comment '%s', but got '%s'", comment, createdResource.Comments)
	}

	// Test resource retrieval
	retrievedResource, err := tryRetrieveResource(service, createdResource.ID)
	if err != nil {
		t.Fatalf("Error retrieving VPN Credential resource: %v", err)
	}

	t.Logf("Retrieved VPN credential: ID=%d, Comments=%s", retrievedResource.ID, retrievedResource.Comments)

	if retrievedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%d', but got '%d'", createdResource.ID, retrievedResource.ID)
	}

	if retrievedResource.Comments != comment {
		t.Errorf("Expected retrieved resource comment '%s', but got '%s'", comment, retrievedResource.Comments)
	}

	retrievedResource.Comments = updateComment
	err = retryOnConflict(func() error {
		_, _, err = Update(context.Background(), service, createdResource.ID, retrievedResource)
		return err
	})
	if err != nil {
		t.Fatalf("Error updating VPN credential: %v", err)
	}

	// Wait for propagation
	time.Sleep(2 * time.Second)

	updatedResource, err := Get(context.Background(), service, createdResource.ID)
	if err != nil {
		t.Fatalf("Error retrieving updated VPN credential: %v", err)
	}

	t.Logf("Updated VPN credential: ID=%d, Comments=%s", updatedResource.ID, updatedResource.Comments)

	if updatedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved updated resource ID '%d', but got '%d'", createdResource.ID, updatedResource.ID)
	}

	if updatedResource.Comments != updateComment {
		t.Errorf("Expected updated VPN credential comment '%s', but got '%s'", updateComment, updatedResource.Comments)
	}

	retrievedResources, err := GetVPNByType(context.Background(), service, "IP", nil, nil, nil)
	if err != nil {
		t.Fatalf("Error retrieving VPN credentials by type: %v", err)
	}

	if len(retrievedResources) == 0 {
		t.Fatal("Expected at least one VPN credential, but got none")
	}

	// Continue with your assertions
	if retrievedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%d', but got '%d'", createdResource.ID, retrievedResource.ID)
	}

	if retrievedResource.Comments != updateComment {
		t.Errorf("Expected retrieved resource comment '%s', but got '%s'", updateComment, retrievedResource.Comments)
	}

	resources, err := GetAll(context.Background(), service)
	if err != nil {
		t.Fatalf("Error retrieving resources: %v", err)
	}

	if len(resources) == 0 {
		t.Fatal("Expected retrieved resources to be non-empty, but got empty slice")
	}

	found := false
	for _, resource := range resources {
		if resource.ID == createdResource.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected retrieved resources to contain created resource '%d', but it didn't", createdResource.ID)
	}

	err = retryOnConflict(func() error {
		return Delete(context.Background(), service, createdResource.ID)
	})
	if err != nil {
		t.Fatalf("Error deleting resource: %v", err)
	}

	_, err = Get(context.Background(), service, createdResource.ID)
	if err == nil {
		t.Fatalf("Expected error retrieving deleted resource, but got nil")
	}
}

// tryRetrieveResource attempts to retrieve a resource with retry mechanism.
func tryRetrieveResource(s *zscaler.Service, id int) (*VPNCredentials, error) {
	var resource *VPNCredentials
	var err error

	for i := 0; i < maxRetries; i++ {
		resource, err = Get(context.Background(), s, id)
		if err == nil && resource != nil && resource.ID == id {
			return resource, nil
		}
		log.Printf("Attempt %d: Error retrieving resource, retrying in %v...", i+1, retryInterval)
		time.Sleep(retryInterval)
	}

	return nil, err
}

func TestRetrieveNonExistentResource(t *testing.T) {
	service, err := tests.NewOneAPIClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}

	_, err = Get(context.Background(), service, 0)
	if err == nil {
		t.Error("Expected error retrieving non-existent resource, but got nil")
	}
}

func TestDeleteNonExistentResource(t *testing.T) {
	service, err := tests.NewOneAPIClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}

	err = Delete(context.Background(), service, 0)
	if err == nil {
		t.Error("Expected error deleting non-existent resource, but got nil")
	}
}

func TestUpdateNonExistentResource(t *testing.T) {
	service, err := tests.NewOneAPIClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}

	_, _, err = Update(context.Background(), service, 0, &VPNCredentials{})
	if err == nil {
		t.Error("Expected error updating non-existent resource, but got nil")
	}
}

func TestGetByNameNonExistentResource(t *testing.T) {
	service, err := tests.NewOneAPIClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}

	_, err = GetByFQDN(context.Background(), service, "non-existent-fqdn")
	if err == nil {
		t.Error("Expected error retrieving resource by non-existent fqdn, but got nil")
	}
}
