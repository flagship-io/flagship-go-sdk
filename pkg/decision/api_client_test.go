package decision

import (
	"testing"
)

var testEnvID = "env_id_test"
var realEnvID = "blvo2kijq6pg023l8edg"
var testAPIKey = "test_api_key"

func TestNewAPIClient(t *testing.T) {
	client, _ := NewAPIClient(testEnvID, testAPIKey)

	if client == nil {
		t.Error("Api client V2 with API Key should not fail")
	}
}

func TestGetModifications(t *testing.T) {
	client, _ := NewAPIClient(testEnvID, testAPIKey)

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	_, err := client.GetModifications("testID", nil, nil)

	if err == nil {
		t.Error("Error should be raised for empty context")
	}
}
