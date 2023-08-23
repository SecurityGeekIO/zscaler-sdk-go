package policysetcontroller

import (
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/zpa/services/idpcontroller"
	"github.com/SecurityGeekIO/zscaler-sdk-go/zpa/services/machinegroup"
	"github.com/SecurityGeekIO/zscaler-sdk-go/zpa/services/samlattribute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAccessForwardingPolicy(t *testing.T) {
	policyType := "CLIENT_FORWARDING_POLICY"

	client, err := tests.NewZpaClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	idpService := idpcontroller.New(client)
	idpList, _, err := idpService.GetAll()
	if err != nil {
		t.Fatalf("Error getting idps: %v", err)
	}
	if len(idpList) == 0 {
		t.Fatal("Expected retrieved idps to be non-empty, but got empty slice")
	}

	samlService := samlattribute.New(client)
	samlsList, _, err := samlService.GetAll()
	if err != nil {
		t.Fatalf("Error getting saml attributes: %v", err)
	}
	if len(samlsList) == 0 {
		t.Fatal("Expected retrieved saml attributes to be non-empty, but got empty slice")
	}

	machineGroupService := machinegroup.New(client)
	machineGroupList, _, err := machineGroupService.GetAll()
	if err != nil {
		t.Fatalf("Error getting posture profiles: %v", err)
	}
	if len(machineGroupList) == 0 {
		t.Fatal("Expected retrieved posture profiles to be non-empty, but got empty slice")
	}

	service := New(client)
	accessPolicySet, _, err := service.GetByPolicyType(policyType)
	if err != nil {
		t.Fatalf("Error getting access forwarding policy set: %v", err)
	}

	for i := 0; i < 3; i++ {
		name := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
		updateName := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

		accessPolicyRule := PolicyRule{
			Name:        name,
			Description: "New application segment",
			PolicySetID: accessPolicySet.ID,
			Action:      "INTERCEPT",
			Conditions: []Conditions{
				{
					Operator: "OR",
					Operands: []Operands{
						{
							ObjectType: "SAML",
							LHS:        samlsList[0].ID,
							RHS:        "user1@acme.com",
							IdpID:      idpList[0].ID,
						},
					},
				},
			},
		}

		// Creation
		createdResource, _, err := service.Create(&accessPolicyRule)
		if err != nil {
			t.Fatalf("Error making POST request: %v", err)
		}

		// Retrieval
		retrievedResource, _, err := service.GetPolicyRule(accessPolicySet.ID, createdResource.ID)
		if err != nil {
			t.Fatalf("Error retrieving resource: %v", err)
		}

		// Update
		retrievedResource.Name = updateName
		_, err = service.Update(accessPolicySet.ID, createdResource.ID, retrievedResource)
		if err != nil {
			t.Fatalf("Error updating resource: %v", err)
		}

		updatedResource, _, err := service.GetPolicyRule(accessPolicySet.ID, createdResource.ID)
		if err != nil {
			t.Fatalf("Error retrieving resource: %v", err)
		}

		if updatedResource.Name != updateName {
			t.Errorf("Expected updated resource name to be '%s', but got '%s'", updateName, updatedResource.Name)
		}

		// Retrieval by Name
		retrievedResource, _, err = service.GetByNameAndType(policyType, updateName)
		if err != nil {
			t.Fatalf("Error retrieving resource by name: %v", err)
		}

		if retrievedResource.ID != createdResource.ID {
			t.Errorf("Expected retrieved by name resource ID to be '%s', but got '%s'", createdResource.ID, retrievedResource.ID)
		}

		// Retrieval of All Resources
		resources, _, err := service.GetAllByType(policyType)
		if err != nil {
			t.Fatalf("Error retrieving resources: %v", err)
		}

		// Check if the created resource is in the list
		found := false
		for _, resource := range resources {
			if resource.ID == createdResource.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected retrieved resources to contain created resource '%s', but it didn't", createdResource.ID)
		}

		// Deletion
		_, err = service.Delete(accessPolicySet.ID, createdResource.ID)
		if err != nil {
			t.Fatalf("Error deleting resource: %v", err)
		}
	}
}
