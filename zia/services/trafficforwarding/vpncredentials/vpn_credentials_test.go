package vpncredentials

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zia/services/trafficforwarding/staticips"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
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

	client, err := tests.NewZiaClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	staticipsService := staticips.New(client)
	staticIP, _, err := staticipsService.Create(&staticips.StaticIP{
		IpAddress: ipAddress,
		Comment:   comment,
	})
	if err != nil {
		t.Fatalf("Creating static ip failed: %v", err)
	}

	defer func() {
		_, err := staticipsService.Delete(staticIP.ID)
		if err != nil {
			t.Errorf("Deleting static ip failed: %v", err)
		}
	}()

	service := New(client)
	cred := VPNCredentials{
		Type:         "IP",
		IPAddress:    ipAddress,
		Comments:     comment,
		PreSharedKey: rPassword,
	}

	var createdResource *VPNCredentials

	err = retryOnConflict(func() error {
		createdResource, _, err = service.Create(&cred)
		return err
	})
	if err != nil {
		t.Fatalf("Error making POST request: %v", err)
	}

	if createdResource.ID == 0 {
		t.Fatal("Expected created resource ID to be non-empty, but got ''")
	}

	if createdResource.Comments != comment {
		t.Errorf("Expected created resource comment '%s', but got '%s'", comment, createdResource.Comments)
	}

	// Test resource retrieval
	retrievedResource, err := tryRetrieveResource(service, createdResource.ID)
	if err != nil {
		t.Fatalf("Error retrieving resource: %v", err)
	}

	if retrievedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%d', but got '%d'", createdResource.ID, retrievedResource.ID)
	}

	if retrievedResource.Comments != comment {
		t.Errorf("Expected retrieved resource comment '%s', but got '%s'", comment, retrievedResource.Comments)
	}

	retrievedResource.Comments = updateComment
	err = retryOnConflict(func() error {
		_, _, err = service.Update(createdResource.ID, retrievedResource)
		return err
	})
	if err != nil {
		t.Fatalf("Error updating resource: %v", err)
	}

	updatedResource, err := service.Get(createdResource.ID)
	if err != nil {
		t.Fatalf("Error retrieving resource: %v", err)
	}

	if updatedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved updated resource ID '%d', but got '%d'", createdResource.ID, updatedResource.ID)
	}

	if updatedResource.Comments != updateComment {
		t.Errorf("Expected retrieved updated resource comment '%s', but got '%s'", updateComment, updatedResource.Comments)
	}

	retrievedResource, err = service.GetVPNByType("IP")
	if err != nil {
		t.Fatalf("Error retrieving resource by name: %v", err)
	}

	if retrievedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%d', but got '%d'", createdResource.ID, retrievedResource.ID)
	}

	if retrievedResource.Comments != updateComment {
		t.Errorf("Expected retrieved resource comment '%s', but got '%s'", updateComment, retrievedResource.Comments)
	}

	resources, err := service.GetAll()
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
		return service.Delete(createdResource.ID)
	})
	if err != nil {
		t.Fatalf("Error deleting resource: %v", err)
	}

	_, err = service.Get(createdResource.ID)
	if err == nil {
		t.Fatalf("Expected error retrieving deleted resource, but got nil")
	}
}

// tryRetrieveResource attempts to retrieve a resource with retry mechanism.
func tryRetrieveResource(s *Service, id int) (*VPNCredentials, error) {
	var resource *VPNCredentials
	var err error

	for i := 0; i < maxRetries; i++ {
		resource, err = s.Get(id)
		if err == nil && resource != nil && resource.ID == id {
			return resource, nil
		}
		log.Printf("Attempt %d: Error retrieving resource, retrying in %v...", i+1, retryInterval)
		time.Sleep(retryInterval)
	}

	return nil, err
}

func TestRetrieveNonExistentResource(t *testing.T) {
	client, err := tests.NewZiaClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}
	service := New(client)

	_, err = service.Get(0)
	if err == nil {
		t.Error("Expected error retrieving non-existent resource, but got nil")
	}
}

func TestDeleteNonExistentResource(t *testing.T) {
	client, err := tests.NewZiaClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}
	service := New(client)

	err = service.Delete(0)
	if err == nil {
		t.Error("Expected error deleting non-existent resource, but got nil")
	}
}

func TestUpdateNonExistentResource(t *testing.T) {
	client, err := tests.NewZiaClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}
	service := New(client)

	_, _, err = service.Update(0, &VPNCredentials{})
	if err == nil {
		t.Error("Expected error updating non-existent resource, but got nil")
	}
}

func TestGetByNameNonExistentResource(t *testing.T) {
	client, err := tests.NewZiaClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}
	service := New(client)

	_, err = service.GetByFQDN("non-existent-fqdn")
	if err == nil {
		t.Error("Expected error retrieving resource by non-existent fqdn, but got nil")
	}
}
