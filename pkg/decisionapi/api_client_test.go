package decisionapi

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/utils"
	"github.com/stretchr/testify/assert"
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
	assert.NotNil(t, client)
	assert.Equal(t, defaultV2APIURL, client.url)
}

func TestNewAPIClientParams(t *testing.T) {
	customHeaders := map[string]string{
		"custom": "value",
	}
	client, _ := NewAPIClient(
		testEnvID,
		testAPIKey,
		APIVersion(1),
		APIKey(testAPIKey),
		Timeout(10),
		Retries(12),
		AdditionalHeaders(customHeaders))

	assert.NotNil(t, client)
	assert.Equal(t, defaultV1APIURL, client.url)
	assert.Equal(t, testAPIKey, client.apiKey)
	assert.Equal(t, 12, client.retries)
	assert.Equal(t, client.additionalHeaders, customHeaders)

	httpClient := client.httpClient.(*utils.HTTPClient)
	headers := reflect.ValueOf(httpClient).Elem().FieldByName("baseHeaders")
	customValue := headers.MapIndex(reflect.ValueOf("custom"))
	assert.Equal(t, customValue.String(), "value")

	client, _ = NewAPIClient(
		testEnvID,
		testAPIKey,
		APIVersion(2),
		APIKey(testAPIKey),
		Timeout(10),
		Retries(12))

	assert.Equal(t, defaultV2APIURL, client.url)
}

func TestGetModifications(t *testing.T) {
	client, _ := NewAPIClient(testEnvID, testAPIKey)
	_, err := client.GetModifications("test_vid", nil, nil)

	assert.NotNil(t, err, "Expected error for unknown env id")

	client, _ = NewAPIClient(realEnvID, testAPIKey)
	_, err = client.GetModifications("test_vid", nil, nil)

	assert.NotNil(t, err, "Expected error for wrong api key")

	response := &model.APIClientResponse{
		VisitorID: "vis_id",
		Panic:     false,
		Campaigns: []model.Campaign{},
	}
	responseJSON, _ := json.Marshal(response)
	client.httpClient = utils.NewHTTPClientMock(200, responseJSON, nil)

	_, err = client.GetModifications("test_vid", nil, nil)
	assert.Nil(t, err)
}

func TestActivate(t *testing.T) {
	client, _ := NewAPIClient(testEnvID, testAPIKey)
	err := client.ActivateCampaign(model.ActivationHit{})

	assert.NotNil(t, err, "Expected error for empty request")

	err = client.ActivateCampaign(model.ActivationHit{
		EnvironmentID:    testEnvID,
		VisitorID:        "test_vid",
		VariationGroupID: "vgid",
		VariationID:      "vid",
	})

	assert.Nil(t, err, "Did not expect error for correct activation request")
}

func TestSendEvent(t *testing.T) {
	client, _ := NewAPIClient(testEnvID, testAPIKey)
	err := client.SendEvent(model.Event{})

	assert.NotNil(t, err, "Expected error for empty request")

	err = client.SendEvent(model.Event{
		VisitorID: "test_vid",
		Type:      "CONTEXT",
		Data: model.Context{
			"hello": "world",
		},
	})

	assert.NotNil(t, err, "Expected error for not existing envID")

	client, _ = NewAPIClient(realEnvID, testAPIKey)
	err = client.SendEvent(model.Event{
		VisitorID: "test_vid",
		Type:      "CONTEXT",
		Data: model.Context{
			"hello": "world",
		},
	})

	assert.Nil(t, err, "Did not expect error for correct activation request")
}
