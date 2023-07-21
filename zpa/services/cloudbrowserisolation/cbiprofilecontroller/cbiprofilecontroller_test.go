package cbiprofilecontroller

/*
import (
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/zpa/services/cloudbrowserisolation/cbibannercontroller"
	"github.com/SecurityGeekIO/zscaler-sdk-go/zpa/services/cloudbrowserisolation/cbicertificatecontroller"
	"github.com/SecurityGeekIO/zscaler-sdk-go/zpa/services/cloudbrowserisolation/cbiregions"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestCBIProfileController(t *testing.T) {
	name := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updateName := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	client, err := tests.NewZpaClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}
	bannerIDService := cbibannercontroller.New(client)
	bannersList, _, err := bannerIDService.GetAll()
	if err != nil {
		t.Errorf("Error getting banners: %v", err)
		return
	}
	if len(bannersList) == 0 {
		t.Error("Expected retrieved idps to be non-empty, but got empty slice")
	}

	regionIDService := cbiregions.New(client)
	regionsList, _, err := regionIDService.GetAll()
	if err != nil {
		t.Errorf("Error getting regions: %v", err)
		return
	}
	if len(regionsList) == 0 {
		t.Error("Expected retrieved idps to be non-empty, but got empty slice")
	}

	certificateIDService := cbicertificatecontroller.New(client)
	certificatesList, _, err := certificateIDService.GetAll()
	if err != nil {
		t.Errorf("Error getting certificates: %v", err)
		return
	}
	if len(certificatesList) == 0 {
		t.Error("Expected retrieved idps to be non-empty, but got empty slice")
	}

	service := New(client)

	cbiProfile := IsolationProfile{
		Name:           name,
		Description:    name,
		BannerID:       bannersList[0].ID,
		RegionIDs:      []string{regionsList[0].ID},
		CertificateIDs: []string{certificatesList[0].ID},
		UserExperience: &UserExperience{
			SessionPersistence: true,
			BrowserInBrowser:   true,
		},
		SecurityControls: &SecurityControls{
			CopyPaste:          "all",
			UploadDownload:     "all",
			DocumentViewer:     true,
			LocalRender:        true,
			AllowPrinting:      true,
			RestrictKeystrokes: false,
		},
	}

	// Test resource creation
	createdResource, _, err := service.Create(&cbiProfile)

	// Check if the request was successful
	if err != nil {
		t.Errorf("Error making POST request: %v", err)
	}

	if createdResource.ID == "" {
		t.Error("Expected created resource ID to be non-empty, but got ''")
	}
	if createdResource.Name != name {
		t.Errorf("Expected created resource name '%s', but got '%s'", name, createdResource.Name)
	}
	// Test resource retrieval
	retrievedResource, _, err := service.Get(createdResource.ID)
	if err != nil {
		t.Errorf("Error retrieving resource: %v", err)
	}
	if retrievedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%s', but got '%s'", createdResource.ID, retrievedResource.ID)
	}
	if retrievedResource.Name != name {
		t.Errorf("Expected retrieved resource name '%s', but got '%s'", name, createdResource.Name)
	}
	// Test resource update
	retrievedResource.Name = updateName
	_, err = service.Update(createdResource.ID, retrievedResource)
	if err != nil {
		t.Errorf("Error updating resource: %v", err)
	}
	updatedResource, _, err := service.Get(createdResource.ID)
	if err != nil {
		t.Errorf("Error retrieving resource: %v", err)
	}
	if updatedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved updated resource ID '%s', but got '%s'", createdResource.ID, updatedResource.ID)
	}
	if updatedResource.Name != updateName {
		t.Errorf("Expected retrieved updated resource name '%s', but got '%s'", updateName, updatedResource.Name)
	}

	// Test resource retrieval by name
	retrievedResource, _, err = service.GetByName(updateName)
	if err != nil {
		t.Errorf("Error retrieving resource by name: %v", err)
	}
	if retrievedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%s', but got '%s'", createdResource.ID, retrievedResource.ID)
	}
	if retrievedResource.Name != updateName {
		t.Errorf("Expected retrieved resource name '%s', but got '%s'", updateName, createdResource.Name)
	}
	// Test resources retrieval
	resources, _, err := service.GetAll()
	if err != nil {
		t.Errorf("Error retrieving resources: %v", err)
	}
	if len(resources) == 0 {
		t.Error("Expected retrieved resources to be non-empty, but got empty slice")
	}
	// check if the created resource is in the list
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
	// Test resource removal
	_, err = service.Delete(createdResource.ID)
	if err != nil {
		t.Errorf("Error deleting resource: %v", err)
		return
	}

	// Test resource retrieval after deletion
	_, _, err = service.Get(createdResource.ID)
	if err == nil {
		t.Errorf("Expected error retrieving deleted resource, but got nil")
	}

}
*/