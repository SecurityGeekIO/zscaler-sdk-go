package cbizpaprofile

import (
	"strings"
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/tests"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestCBIZPAProfile(t *testing.T) {
	client, err := tests.NewZpaClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}

	service := New(client)

	// Test to retrieve all profiles
	profiles, _, err := service.GetAll()
	if err != nil {
		t.Errorf("Error getting isolation profiles: %v", err)
		return
	}
	if len(profiles) == 0 {
		t.Errorf("No isolation profile found")
		return
	}

	// Test to retrieve a profile by its name
	name := profiles[0].Name
	t.Log("Getting isolation profile by name:" + name)
	profile, _, err := service.GetByName(name)
	if err != nil {
		t.Errorf("Error getting isolation profile by name: %v", err)
		return
	}
	if profile.Name != name {
		t.Errorf("isolation profile name does not match: expected %s, got %s", name, profile.Name)
		return
	}

	// Negative Test: Try to retrieve a profile with a non-existent name
	nonExistentName := "ThisProfileNameDoesNotExist"
	_, _, err = service.GetByName(nonExistentName)
	if err == nil {
		t.Errorf("Expected error when getting by non-existent name, got nil")
		return
	}
}

func TestResponseFormatValidation(t *testing.T) {
	client, err := tests.NewZpaClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}

	service := New(client)

	profiles, _, err := service.GetAll()
	if err != nil {
		t.Errorf("Error getting isolation profiles: %v", err)
		return
	}
	if len(profiles) == 0 {
		t.Errorf("No isolation profile found")
		return
	}

	// Validate each profile
	for _, profile := range profiles {
		// Checking if essential fields are not empty
		if profile.ID == "" {
			t.Errorf("IsolationProfile ID is empty")
		}
		if profile.Name == "" {
			t.Errorf("IsolationProfile Name is empty")
		}
		if profile.CBIURL == "" {
			t.Errorf("IsolationProfile CBI URL is empty")
		}
	}
}

func TestCaseSensitivityOfGetByName(t *testing.T) {
	client, err := tests.NewZpaClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}

	service := New(client)

	// Assuming a profile with the name "BD_SA_Profile1" exists
	knownName := "BD_SA_Profile1"

	// Case variations to test
	variations := []string{
		strings.ToUpper(knownName),
		strings.ToLower(knownName),
		cases.Title(language.English).String(knownName),
	}

	for _, variation := range variations {
		t.Logf("Attempting to retrieve profile with name variation: %s", variation)
		profile, _, err := service.GetByName(variation)
		if err != nil {
			t.Errorf("Error getting isolation profile with name variation '%s': %v", variation, err)
			continue
		}

		// Check if the profile's actual name matches the known name
		if profile.Name != knownName {
			t.Errorf("Expected profile name to be '%s' for variation '%s', but got '%s'", knownName, variation, profile.Name)
		}
	}
}

func TestProfileNamesWithSpaces(t *testing.T) {
	client, err := tests.NewZpaClient()
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}

	service := New(client)

	// Assuming that there are profiles with the following name variations
	variations := []string{
		"BD SA Profile",     // Single space
		"BD  SA Profile",    // Double space
		"BD   SA   Profile", // Multiple spaces
	}

	for _, variation := range variations {
		t.Logf("Attempting to retrieve profile with name: %s", variation)
		profile, _, err := service.GetByName(variation)
		if err != nil {
			t.Errorf("Error getting isolation profile with name '%s': %v", variation, err)
			continue
		}

		// Verify if the profile's actual name matches the expected variation
		if profile.Name != variation {
			t.Errorf("Expected profile name to be '%s' but got '%s'", variation, profile.Name)
		}
	}
}
