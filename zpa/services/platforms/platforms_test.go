package platforms

import (
	"reflect"
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa/services"
)

func TestGetAllPlatforms(t *testing.T) {
	client, err := tests.NewZpaClient()
	if err != nil {
		t.Fatalf("Failed to create ZPA client: %v", err)
	}
	service := services.New(client)

	// Test case: Normal scenario
	t.Run("TestGetAllPlatformsNormal", func(t *testing.T) {
		platforms, resp, err := GetAllPlatforms(service)
		if err != nil {
			t.Fatalf("Failed to fetch platforms: %v", err)
		}

		if resp.StatusCode >= 400 {
			t.Errorf("Received an HTTP error %d when fetching platforms", resp.StatusCode)
		}

		if platforms == nil {
			t.Fatal("Platforms nil, expected a valid response")
		}

		tests := map[string]string{
			"linux":   "Linux",
			"android": "Android",
			"windows": "Windows",
			"ios":     "iOS",
			"mac":     "Mac", // adjusted this line
		}

		platformValues := getValuesByTags(platforms)
		for jsonTag, expectedValue := range tests {
			actualValue, found := platformValues[jsonTag]
			if !found || actualValue != expectedValue {
				t.Errorf("Expected %s but got %s for json tag %s", expectedValue, actualValue, jsonTag)
			}
		}
	})
}

func getValuesByTags(types *Platforms) map[string]string {
	values := make(map[string]string)
	r := reflect.ValueOf(types).Elem()
	for i := 0; i < r.NumField(); i++ {
		fieldTag := r.Type().Field(i).Tag.Get("json")
		values[fieldTag] = r.Field(i).String()
	}
	return values
}
