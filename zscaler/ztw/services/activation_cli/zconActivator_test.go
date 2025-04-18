package main

/*
import (
	"context"
	"os"
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/tests"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/ztw/services"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/ztw/services/activation"
)

func TestActivationCLI(t *testing.T) {
	// Check that necessary environment variables are set
	checkEnvVarForTest(t, "ZCON_USERNAME")
	checkEnvVarForTest(t, "ZCON_PASSWORD")
	checkEnvVarForTest(t, "ZCON_API_KEY")
	checkEnvVarForTest(t, "ZCON_CLOUD")

	// Construct the client
	client, err := tests.NewZConClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	service := services.New(client)

	_, err = activation.ForceActivationStatus(context.Background(), service, activation.ECAdminActivation{
		AdminActivateStatus: "ADM_ACTV_DONE",
	})
	if err != nil {
		t.Fatalf("[ERROR] Activation Failed: %v", err)
	}

	// Destroy the session
	if err := client.Logout(context.Background()); err != nil {
		t.Fatalf("[ERROR] Failed destroying session: %v", err)
	}
}

func checkEnvVarForTest(t *testing.T, k string) {
	if v := os.Getenv(k); v == "" {
		t.Fatalf("[ERROR] Couldn't find environment variable %s", k)
	}
}
*/
