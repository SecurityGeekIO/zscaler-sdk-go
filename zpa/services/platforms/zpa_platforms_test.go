package platforms

import (
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/tests"
)

func TestGetAllPlatforms(t *testing.T) {
	client, err := tests.NewZpaClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
		return
	}

	service := New(client)
	platforms, _, err := service.GetAllPlatforms()

	if err != nil {
		t.Fatalf("Error getting all platforms: %v", err)
		return
	}

	if platforms == nil {
		t.Fatal("Received nil platforms")
		return
	}

	// Verifying some of the platforms (you can add more based on your use-case)
	if platforms.Linux == "" {
		t.Error("Expected Linux platform version, but got empty string.")
	}
	if platforms.Android == "" {
		t.Error("Expected Android platform version, but got empty string.")
	}
	if platforms.Windows == "" {
		t.Error("Expected Windows platform version, but got empty string.")
	}
	if platforms.IOS == "" {
		t.Error("Expected IOS platform version, but got empty string.")
	}
	if platforms.MacOS == "" {
		t.Error("Expected MacOS platform version, but got empty string.")
	}

	// Any additional checks or logics based on real-world scenarios can be added here.
}
