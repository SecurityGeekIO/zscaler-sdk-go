package apptotal

import (
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/tests"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAppTotalIntegration(t *testing.T) {
	name := acctest.RandStringFromCharSet(30, acctest.CharSetAlpha)
	description := acctest.RandStringFromCharSet(30, acctest.CharSetAlpha)

	client, err := tests.NewZiaClient() // Assuming you use the same client structure for apptotal.
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}
	service := New(client)

	appData := AppTotal{
		Name:        name,
		Description: description,
		// Fill out any other necessary fields here.
	}

	// Test app creation
	createdApp, _, err := service.Create(&AppTotalCreation{
		AppID: appData.AppID,
	})
	if err != nil {
		t.Errorf("Error making POST request: %v", err)
	}

	if createdApp.AppID == 0 {
		t.Error("Expected created app ID to be non-empty, but got ''")
	}
	if createdApp.Name != name {
		t.Errorf("Expected created app name '%s', but got '%s'", name, createdApp.Name)
	}

	// Test app retrieval
	retrievedApp, err := service.Get(createdApp.AppID, true)
	if err != nil {
		t.Errorf("Error retrieving app: %v", err)
	}
	if retrievedApp.AppID != createdApp.AppID {
		t.Errorf("Expected retrieved app ID '%d', but got '%d'", createdApp.AppID, retrievedApp.AppID)
	}
	if retrievedApp.Name != name {
		t.Errorf("Expected retrieved app name '%s', but got '%s'", name, retrievedApp.Name)
	}

	// More tests can be added, similar to the provided example, if you have methods like GetAll or GetByName for AppTotal.
}
