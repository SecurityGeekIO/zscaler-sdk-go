package policysetcontroller

import (
	"fmt"
	"testing"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zpa/services/idpcontroller"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zpa/services/postureprofile"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zpa/services/samlattribute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestPolicyAccessRule(t *testing.T) {
	policyType := "ACCESS_POLICY"
	service, err := tests.NewOneAPIClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	idpList, _, err := idpcontroller.GetAll(service)
	if err != nil {
		t.Errorf("Error getting idps: %v", err)
		return
	}
	if len(idpList) == 0 {
		t.Error("Expected retrieved idps to be non-empty, but got empty slice")
	}
	samlsList, _, err := samlattribute.GetAll(service)
	if err != nil {
		t.Errorf("Error getting saml attributes: %v", err)
		return
	}
	if len(samlsList) == 0 {
		t.Error("Expected retrieved saml attributes to be non-empty, but got empty slice")
	}

	postureList, _, err := postureprofile.GetAll(service)
	if err != nil {
		t.Errorf("Error getting posture profiles: %v", err)
		return
	}
	if len(postureList) == 0 {
		t.Error("Expected retrieved posture profiles to be non-empty, but got empty slice")
	}
	accessPolicySet, _, err := GetByPolicyType(service, policyType)
	if err != nil {
		t.Errorf("Error getting access policy set: %v", err)
		return
	}

	var ruleIDs []string // Store the IDs of the created rules

	for i := 0; i < 5; i++ {
		// Generate a unique name for each iteration
		name := fmt.Sprintf("tests-%s-%d", acctest.RandStringFromCharSet(10, acctest.CharSetAlpha), i)

		accessPolicyRule := PolicyRule{
			Name:        name,
			Description: name,
			PolicySetID: accessPolicySet.ID,
			Action:      "ALLOW",
			Conditions: []Conditions{
				{
					Operator: "OR",
					Operands: []Operands{
						{
							ObjectType: "POSTURE",
							LHS:        postureList[0].PostureudID,
							RHS:        "false",
						},
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

		// Test resource creation
		createdResource, _, err := CreateRule(service, &accessPolicyRule)

		if err != nil {
			t.Errorf("Error making POST request: %v", err)
			continue
		}
		if createdResource.ID == "" {
			t.Error("Expected created resource ID to be non-empty, but got ''")
			continue
		}
		ruleIDs = append(ruleIDs, createdResource.ID) // Collect rule ID for reordering

		// Update the rule name
		updatedName := name + "-updated"
		accessPolicyRule.Name = updatedName
		_, updateErr := UpdateRule(service, accessPolicySet.ID, createdResource.ID, &accessPolicyRule)

		if updateErr != nil {
			t.Errorf("Error updating rule: %v", updateErr)
			continue
		}

		// Retrieve and verify the updated resource
		updatedResource, _, getErr := GetPolicyRule(service, accessPolicySet.ID, createdResource.ID)
		if getErr != nil {
			t.Errorf("Error retrieving updated resource: %v", getErr)
			continue
		}
		if updatedResource.Name != updatedName {
			t.Errorf("Expected updated resource name '%s', but got '%s'", updatedName, updatedResource.Name)
		}

		// Test resource retrieval by name
		updatedResource, _, err = GetByNameAndType(service, policyType, updatedName)
		if err != nil {
			t.Errorf("Error retrieving resource by name: %v", err)
		}
		if updatedResource.ID != createdResource.ID {
			t.Errorf("Expected retrieved resource ID '%s', but got '%s'", createdResource.ID, updatedResource.ID)
		}
		if updatedResource.Name != updatedName {
			t.Errorf("Expected retrieved resource name '%s', but got '%s'", updatedName, updatedResource.Name)
		}
		time.Sleep(5 * time.Second)
	}

	// Reorder the rules after all have been created and updated
	ruleIdToOrder := make(map[string]int)
	for i, id := range ruleIDs {
		ruleIdToOrder[id] = len(ruleIDs) - i // Reverse the order
	}

	_, err = BulkReorder(service, policyType, ruleIdToOrder)
	if err != nil {
		t.Errorf("Error reordering rules: %v", err)
	}

	// Clean up: Delete the rules
	for _, ruleID := range ruleIDs {
		_, err = Delete(service, accessPolicySet.ID, ruleID)
		if err != nil {
			t.Errorf("Error deleting resource: %v", err)
		}
	}
}