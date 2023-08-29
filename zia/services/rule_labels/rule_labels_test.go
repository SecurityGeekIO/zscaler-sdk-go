package rule_labels

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/tests"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	cleanResources()
}

func teardown() {
	cleanResources()
}

func shouldClean() bool {
	val, present := os.LookupEnv("ZSCALER_SDK_TEST_SWEEP")
	return !present || (present && (val == "" || val == "true")) // simplified for clarity
}

func cleanResources() {
	if !shouldClean() {
		return
	}

	client, err := tests.NewZiaClient()
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	service := New(client)
	resources, err := service.GetAll()
	if err != nil {
		log.Printf("Error retrieving resources during cleanup: %v", err)
		return
	}

	for _, r := range resources {
		if strings.HasPrefix(r.Name, "tests-") {
			_, err := service.Delete(r.ID)
			if err != nil {
				log.Printf("Error deleting resource %d: %v", r.ID, err)
			}
		}
	}
}

func TestRuleLabels(t *testing.T) {
	cleanResources()                // At the start of the test
	defer t.Cleanup(cleanResources) // Will be called at the end

	name := acctest.RandStringFromCharSet(30, acctest.CharSetAlpha)
	description := acctest.RandStringFromCharSet(30, acctest.CharSetAlpha)
	updateDescription := acctest.RandStringFromCharSet(30, acctest.CharSetAlpha)
	client, err := tests.NewZiaClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}
	service := New(client)

	ip := RuleLabels{
		Name:        name,
		Description: description,
	}

	// Test resource creation
	createdResource, _, err := service.Create(&ip)
	// Check if the request was successful
	if err != nil {
		t.Errorf("Error making POST request: %v", err)
	}

	if createdResource.ID == 0 {
		t.Error("Expected created resource ID to be non-empty, but got ''")
	}
	if createdResource.Name != name {
		t.Errorf("Expected created rule label '%s', but got '%s'", name, createdResource.Name)
	}
	// Test resource retrieval
	retrievedResource, err := service.Get(createdResource.ID)
	if err != nil {
		t.Errorf("Error retrieving resource: %v", err)
	}
	if retrievedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%d', but got '%d'", createdResource.ID, retrievedResource.ID)
	}
	if retrievedResource.Name != name {
		t.Errorf("Expected retrieved rule label '%s', but got '%s'", name, retrievedResource.Name)
	}
	// Test resource update
	retrievedResource.Description = updateDescription
	_, _, err = service.Update(createdResource.ID, retrievedResource)
	if err != nil {
		t.Errorf("Error updating resource: %v", err)
	}
	updatedResource, err := service.Get(createdResource.ID)
	if err != nil {
		t.Errorf("Error retrieving resource: %v", err)
	}
	if updatedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved updated resource ID '%d', but got '%d'", createdResource.ID, updatedResource.ID)
	}
	if updatedResource.Description != updateDescription {
		t.Errorf("Expected retrieved updated resource comment '%s', but got '%s'", updateDescription, updatedResource.Description)
	}

	// Test resource retrieval by name
	retrievedResource, err = service.GetRuleLabelByName(name)
	if err != nil {
		t.Errorf("Error retrieving resource by name: %v", err)
	}
	if retrievedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%d', but got '%d'", createdResource.ID, retrievedResource.ID)
	}
	if retrievedResource.Description != updateDescription {
		t.Errorf("Expected retrieved resource comment '%s', but got '%s'", updateDescription, createdResource.Description)
	}
	// Test resources retrieval
	resources, err := service.GetAll()
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
		t.Errorf("Expected retrieved resources to contain created resource '%d', but it didn't", createdResource.ID)
	}
	// Test resource removal
	_, err = service.Delete(createdResource.ID)
	if err != nil {
		t.Errorf("Error deleting resource: %v", err)
		return
	}

	// Test resource retrieval after deletion
	_, err = service.Get(createdResource.ID)
	if err == nil {
		t.Errorf("Expected error retrieving deleted resource, but got nil")
	}
}
