package vpncredentials

import (
	"log"
	"testing"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/zia/services/trafficforwarding/staticips"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

const maxRetries = 3
const retryInterval = 2 * time.Second

func TestTrafficForwardingVPNCreds(t *testing.T) {
	ipAddress, _ := acctest.RandIpAddress("104.239.238.0/24")
	comment := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updateComment := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

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
		PreSharedKey: "newPassword123!",
	}

	createdResource, _, err := service.Create(&cred)
	if err != nil {
		t.Fatalf("Error making POST request: %v", err)
	}

	if createdResource.ID == 0 {
		t.Fatal("Expected created resource ID to be non-empty, but got ''")
	}

	if createdResource.Comments != comment {
		t.Errorf("Expected created resource comment '%s', but got '%s'", comment, createdResource.Comments)
	}

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
	_, _, err = service.Update(createdResource.ID, retrievedResource)
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

	err = service.Delete(createdResource.ID)
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
