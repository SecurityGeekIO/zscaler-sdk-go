package emergencyaccess

import (
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/tests"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/stretchr/testify/assert"
)

func TestEmergencyAccessIntegration(t *testing.T) {
	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	client, err := tests.NewZpaClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}
	service := New(client)

	// Create new resource
	createdResource, _, err := service.Create(&EmergencyAccess{
		ActivatedOn:       "1",
		AllowedActivate:   true,
		AllowedDeactivate: true,
		EmailID:           randomName + "@bd-hashicorp.com",
		FirstName:         "John",
		LastName:          "Smith",
		UserID:            "jsmith",
	})
	if err != nil {
		t.Fatalf("Failed to create emergency user: %v", err)
	}

	// Test Get
	gotResource, _, err := service.Get(createdResource.UserID)
	if err != nil {
		t.Errorf("Failed to get emergency user by UserID: %v", err)
	}
	assert.Equal(t, createdResource.UserID, gotResource.UserID, "UserID does not match")

	//Test Update
	updatedResource := *createdResource
	updatedResource.FirstName = randomName
	_, err = service.Update(createdResource.UserID, &updatedResource)
	if err != nil {
		t.Errorf("Failed to update emergency user: %v", err)
	}

	// Verify Update
	updated, _, err := service.Get(createdResource.UserID)
	if err != nil {
		t.Errorf("Failed to get updated emergency user: %v", err)
	}
	assert.Equal(t, randomName, updated.FirstName, "FirstName was not updated")

	// Test Emergency Access User Deactivation
	_, err = service.Deactivate(createdResource.UserID)
	if err != nil {
		t.Errorf("Failed to deactivate emergency user: %v", err)
	}

	// Test Emergency Access User Activate
	_, err = service.Activate(createdResource.UserID)
	if err != nil {
		t.Errorf("Failed to activate emergency user: %v", err)
	}
}
