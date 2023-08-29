package dlp_web_rules

import (
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/zia/services/common"
	"github.com/SecurityGeekIO/zscaler-sdk-go/zia/services/rule_labels"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

const maxRetries = 3
const retryInterval = 2 * time.Second

// Constants for conflict retries
const maxConflictRetries = 5
const conflictRetryInterval = 1 * time.Second

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

// clean all resources
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
	if !present {
		return true
	}
	shouldClean, err := strconv.ParseBool(val)
	if err != nil {
		return true
	}
	log.Printf("ZSCALER_SDK_TEST_SWEEP value: %v", shouldClean)
	return shouldClean
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
	resources, _ := service.GetAll()
	for _, r := range resources {
		if !strings.HasPrefix(r.Name, "tests-") {
			continue
		}
		_, err := service.Delete(r.ID)
		if err != nil {
			log.Printf("Error deleting resource with ID %d: %v", r.ID, err)
		} else {
			log.Printf("Successfully deleted resource with ID %d", r.ID)
		}
	}
}

func TestDLPWebRule(t *testing.T) {
	cleanResources()                // At the start of the test
	defer t.Cleanup(cleanResources) // Will be called at the end

	name := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updateName := "tests-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	client, err := tests.NewZiaClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	// create rule label for testing
	ruleLabelService := rule_labels.New(client)
	ruleLabel, _, err := ruleLabelService.Create(&rule_labels.RuleLabels{
		Name:        name,
		Description: name,
	})
	if err != nil {
		t.Fatalf("Error creating rule label for testing: %v", err)
	}

	// Ensure the rule label is cleaned up at the end of this test
	defer func() {
		_, err := ruleLabelService.Delete(ruleLabel.ID)
		if err != nil {
			t.Errorf("Error deleting rule label: %v", err)
		}
	}()

	service := New(client)
	rule := WebDLPRules{
		Name:                     name,
		Description:              name,
		Order:                    1,
		Rank:                     7,
		State:                    "ENABLED",
		Action:                   "BLOCK",
		OcrEnabled:               true,
		ZscalerIncidentReceiver:  true,
		WithoutContentInspection: false,
		Protocols:                []string{"FTP_RULE", "HTTPS_RULE", "HTTP_RULE"},
		CloudApplications:        []string{"WINDOWS_LIVE_HOTMAIL"},
		FileTypes:                []string{"WINDOWS_META_FORMAT", "BITMAP", "JPEG", "PNG", "TIFF"},
		Labels: []common.IDNameExtensions{
			{
				ID: ruleLabel.ID,
			},
		},
	}

	var createdResource *WebDLPRules

	// Test resource creation
	err = retryOnConflict(func() error {
		createdResource, err = service.Create(&rule)
		return err
	})
	if err != nil {
		t.Fatalf("Error making POST request: %v", err)
	}

	// Other assertions based on the creation result
	if createdResource.ID == 0 {
		t.Fatal("Expected created resource ID to be non-empty, but got ''")
	}
	if createdResource.Name != name {
		t.Errorf("Expected created resource name '%s', but got '%s'", name, createdResource.Name)
	}

	// Test resource retrieval
	retrievedResource, err := tryRetrieveResource(service, createdResource.ID)
	if err != nil {
		t.Fatalf("Error retrieving resource: %v", err)
	}
	if retrievedResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%d', but got '%d'", createdResource.ID, retrievedResource.ID)
	}
	if retrievedResource.Name != name {
		t.Errorf("Expected retrieved dlp engine '%s', but got '%s'", name, retrievedResource.Name)
	}

	// Test resource update
	retrievedResource.Name = updateName
	err = retryOnConflict(func() error {
		_, err = service.Update(createdResource.ID, retrievedResource)
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
	if updatedResource.Name != updateName {
		t.Errorf("Expected retrieved updated resource name '%s', but got '%s'", updateName, updatedResource.Name)
	}

	// Test resource retrieval by name
	retrievedByNameResource, err := service.GetByName(updateName)
	if err != nil {
		t.Fatalf("Error retrieving resource by name: %v", err)
	}
	if retrievedByNameResource.ID != createdResource.ID {
		t.Errorf("Expected retrieved resource ID '%d', but got '%d'", createdResource.ID, retrievedByNameResource.ID)
	}
	if retrievedByNameResource.Name != updateName {
		t.Errorf("Expected retrieved resource name '%s', but got '%s'", updateName, retrievedByNameResource.Name)
	}

	// Test resources retrieval
	allResources, err := service.GetAll()
	if err != nil {
		t.Fatalf("Error retrieving resources: %v", err)
	}
	if len(allResources) == 0 {
		t.Fatal("Expected retrieved resources to be non-empty, but got empty slice")
	}

	// check if the created resource is in the list
	found := false
	for _, resource := range allResources {
		if resource.ID == createdResource.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected retrieved resources to contain created resource '%d', but it didn't", createdResource.ID)
	}

	// Introduce a delay before deleting
	time.Sleep(5 * time.Second) // sleep for 5 seconds

	// Test resource removal
	err = retryOnConflict(func() error {
		_, delErr := service.Delete(createdResource.ID)
		return delErr
	})
	_, err = service.Get(createdResource.ID)
	if err == nil {
		t.Fatalf("Expected error retrieving deleted resource, but got nil")
	}
}

// tryRetrieveResource attempts to retrieve a resource with retry mechanism.
func tryRetrieveResource(s *Service, id int) (*WebDLPRules, error) {
	var resource *WebDLPRules
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
