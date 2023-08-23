package clienttypes

import (
	"testing"

	"github.com/SecurityGeekIO/zscaler-sdk-go/tests"
)

func TestGetAllClientTypes(t *testing.T) {
	client, err := tests.NewZpaClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
		return
	}

	service := New(client)
	clientTypes, _, err := service.GetAllClientTypes()

	if err != nil {
		t.Fatalf("Error getting all client types: %v", err)
		return
	}

	if clientTypes == nil {
		t.Fatal("Received nil client types")
		return
	}

	// Verifying some of the client types (you can add more based on your use-case)
	if clientTypes.ZPNClientTypeExplorer == "" {
		t.Error("Expected ZPNClientTypeExplorer, but got empty string.")
	}
	if clientTypes.ZPNClientTypeNoAuth == "" {
		t.Error("Expected ZPNClientTypeNoAuth, but got empty string.")
	}
	if clientTypes.ZPNClientTypeBrowserIsolation == "" {
		t.Error("Expected ZPNClientTypeBrowserIsolation, but got empty string.")
	}
	if clientTypes.ZPNClientTypeMachineTunnel == "" {
		t.Error("Expected ZPNClientTypeMachineTunnel, but got empty string.")
	}
	if clientTypes.ZPNClientTypeIPAnchoring == "" {
		t.Error("Expected ZPNClientTypeIPAnchoring, but got empty string.")
	}
	if clientTypes.ZPNClientTypeEdgeConnector == "" {
		t.Error("Expected ZPNClientTypeEdgeConnector, but got empty string.")
	}
	if clientTypes.ZPNClientTypeZAPP == "" {
		t.Error("Expected ZPNClientTypeZAPP, but got empty string.")
	}
	if clientTypes.ZPNClientTypeSlogger == "" {
		t.Error("Expected ZPNClientTypeSlogger, but got empty string.")
	}
	if clientTypes.ZPNClientTypeBranchConnector == "" {
		t.Error("Expected ZPNClientTypeBranchConnector, but got empty string.")
	}
	if clientTypes.ZPNClientTypePartner == "" {
		t.Error("Expected ZPNClientTypePartner, but got empty string.")
	}
}
