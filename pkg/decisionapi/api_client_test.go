package decisionapi

import (
	"testing"

	"github.com/abtasty/flagship-go-sdk/pkg/model"
)

var testEnvID = "env_id_test"
var realEnvID = "blvo2kijq6pg023l8edg"

func TestNewAPIClient(t *testing.T) {
	client, _ := NewAPIClient(testEnvID)

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.url != defaultV1APIURL {
		t.Error("Api url should be set to default")
	}
}

func TestNewAPIClientParams(t *testing.T) {
	client, _ := NewAPIClient(
		testEnvID,
		APIVersion(1),
		APIKey("api_key"),
		Timeout(10),
		Retries(12))

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.url != defaultV1APIURL {
		t.Error("Api url should be set to default")
	}

	if client.apiKey != "api_key" {
		t.Errorf("Wrong api key. Expected %v, got %v", "api_key", client.apiKey)
	}

	if client.retries != 12 {
		t.Errorf("Wrong retries. Expected %v, got %v", 12, client.retries)
	}

	client, err := NewAPIClient(
		testEnvID,
		APIVersion(2),
		Timeout(10),
		Retries(12))

	if err == nil {
		t.Error("Client should return an error because V2 API required API Key")
	}

	client, _ = NewAPIClient(
		testEnvID,
		APIVersion(2),
		APIKey("api_key"),
		Timeout(10),
		Retries(12))

	if client.url != defaultV2APIURL {
		t.Error("Api url should be set to V2")
	}
}

func TestGetModifications(t *testing.T) {
	client, _ := NewAPIClient(testEnvID)
	_, err := client.GetModifications("test_vid", nil)

	if err == nil {
		t.Error("Expected error for unknown env id")
	}

	client, _ = NewAPIClient(realEnvID)
	resp, err := client.GetModifications("test_vid", nil)

	if err != nil {
		t.Errorf("Unexpected error for correct env id : %v", err)
	}

	if resp == nil {
		t.Errorf("Expected not nil response for correct env id")
	}
}

func TestActivate(t *testing.T) {
	client, _ := NewAPIClient(testEnvID)
	err := client.ActivateCampaign(model.ActivationHit{})

	if err == nil {
		t.Errorf("Expected error for empty request")
	}

	err = client.ActivateCampaign(model.ActivationHit{
		EnvironmentID:    testEnvID,
		VisitorID:        "test_vid",
		VariationGroupID: "vgid",
		VariationID:      "vid",
	})

	if err != nil {
		t.Errorf("Did not expect error for correct activation request. Got %v", err)
	}
}

func TestSendEvent(t *testing.T) {
	client, _ := NewAPIClient(testEnvID)
	err := client.SendEvent(model.Event{})

	if err == nil {
		t.Errorf("Expected error for empty request")
	}

	err = client.SendEvent(model.Event{
		VisitorID: "test_vid",
		Type:      "CONTEXT",
		Data: model.Context{
			"hello": "world",
		},
	})

	if err == nil {
		t.Errorf("Expected error for not existing envID. Got nil")
	}

	client, _ = NewAPIClient(realEnvID)
	err = client.SendEvent(model.Event{
		VisitorID: "test_vid",
		Type:      "CONTEXT",
		Data: model.Context{
			"hello": "world",
		},
	})

	if err != nil {
		t.Errorf("Did not expect error for correct activation request. Got %v", err)
	}
}
