package decisionapi

import (
	"testing"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
)

var testEnvID = "env_id_test"
var realEnvID = "blvo2kijq6pg023l8edg"
var testAPIKey = "api_key_test"

func TestNewAPIClient(t *testing.T) {
	_, err := NewAPIClient(testEnvID, "")

	if err == nil {
		t.Error("Api client with empty api key should return an error")
	}

	client, _ := NewAPIClient(testEnvID, testAPIKey)

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.url != defaultV2APIURL {
		t.Error("Api url should be set to default")
	}
}

func TestNewAPIClientParams(t *testing.T) {
	client, _ := NewAPIClient(
		testEnvID,
		testAPIKey,
		APIVersion(1),
		APIKey(testAPIKey),
		Timeout(10),
		Retries(12))

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.url != defaultV1APIURL {
		t.Error("Api url should be set to default")
	}

	if client.apiKey != testAPIKey {
		t.Errorf("Wrong api key. Expected %v, got %v", testAPIKey, client.apiKey)
	}

	if client.retries != 12 {
		t.Errorf("Wrong retries. Expected %v, got %v", 12, client.retries)
	}

	client, _ = NewAPIClient(
		testEnvID,
		testAPIKey,
		APIVersion(2),
		APIKey(testAPIKey),
		Timeout(10),
		Retries(12))

	if client.url != defaultV2APIURL {
		t.Error("Api url should be set to V2")
	}
}

func TestGetModifications(t *testing.T) {
	client, _ := NewAPIClient(testEnvID, testAPIKey)
	_, err := client.GetModifications("test_vid", nil)

	if err == nil {
		t.Error("Expected error for unknown env id")
	}

	client, _ = NewAPIClient(realEnvID, testAPIKey)
	_, err = client.GetModifications("test_vid", nil)

	if err == nil {
		t.Errorf("Expected error for wrong api key : %v", err)
	}
}

func TestActivate(t *testing.T) {
	client, _ := NewAPIClient(testEnvID, testAPIKey)
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
	client, _ := NewAPIClient(testEnvID, testAPIKey)
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

	client, _ = NewAPIClient(realEnvID, testAPIKey)
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
