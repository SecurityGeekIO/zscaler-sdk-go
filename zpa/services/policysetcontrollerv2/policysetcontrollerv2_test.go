package policysetcontrollerv2

import (
	"fmt"
	"testing"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa/services/idpcontroller"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa/services/samlattribute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestPolicyAccessRule(t *testing.T) {
	policyType := "ACCESS_POLICY"
	client, err := tests.NewZpaClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}
	idpService := idpcontroller.New(client)
	idpList, _, err := idpService.GetAll()
	if err != nil {
		t.Errorf("Error getting idps: %v", err)
		return
	}
	if len(idpList) == 0 {
		t.Error("Expected retrieved idps to be non-empty, but got empty slice")
	}
	samlService := samlattribute.New(client)
	samlsList, _, err := samlService.GetAll()
	if err != nil {
		t.Errorf("Error getting saml attributes: %v", err)
		return
	}
	if len(samlsList) == 0 {
		t.Error("Expected retrieved saml attributes to be non-empty, but got empty slice")
	}
	service := New(client)
	accessPolicySet, _, err := service.GetByPolicyType(policyType)
	if err != nil {
		t.Errorf("Error getting access policy set: %v", err)
		return
	}

	var ruleIDs []string // Store the IDs of the created rules

	for i := 0; i < 6; i++ {
		// Generate a unique name for each iteration
		name := fmt.Sprintf("tests-%s-%d", acctest.RandStringFromCharSet(10, acctest.CharSetAlpha), i)

		accessPolicyRule := PolicyRule{
			Name:        name,
			Description: name,
			PolicySetID: accessPolicySet.ID,
			Action:      "ALLOW",
			Conditions:  []PolicyRuleResourceConditions{
				// Your specific conditions here
			},
		}

		// Test resource creation
		createdResource, _, err := service.CreateRule(&accessPolicyRule)
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
		_, updateErr := service.UpdateRule(accessPolicySet.ID, createdResource.ID, &accessPolicyRule)
		if updateErr != nil {
			t.Errorf("Error updating rule: %v", updateErr)
			continue
		}

		// Verify the update was successful
		updatedResource, _, getErr := service.GetPolicyRule(accessPolicySet.ID, createdResource.ID)
		if getErr != nil {
			t.Errorf("Error retrieving updated resource: %v", getErr)
			continue
		}
		if updatedResource.Name != updatedName {
			t.Errorf("Expected updated resource name '%s', but got '%s'", updatedName, updatedResource.Name)
		}

		// Introduce a delay to prevent rate limit issues
		time.Sleep(10 * time.Second)
	}

	// Reorder the rules after all have been created and updated
	ruleIdToOrder := make(map[string]int)
	for i, id := range ruleIDs {
		ruleIdToOrder[id] = len(ruleIDs) - i // Reverse the order
	}

	_, err = service.BulkReorder(policyType, ruleIdToOrder)
	if err != nil {
		t.Errorf("Error reordering rules: %v", err)
	}

	// Optionally verify the new order of rules here

	// Clean up: Delete the rules
	for _, ruleID := range ruleIDs {
		_, err = service.Delete(accessPolicySet.ID, ruleID)
		if err != nil {
			t.Errorf("Error deleting resource: %v", err)
		}
	}
}
