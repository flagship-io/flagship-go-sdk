package tracking

import (
	"testing"

	"github.com/flagship-io/flagship-go-sdk/v3/pkg/model"
	"github.com/stretchr/testify/assert"
)

var testVisitorID = "test_visitor_id"
var testEnvID = "test_env_id"
var realEnvID = "blvo2kijq6pg023l8edg"
var testAPIKey = "api_key_test"

func TestNewAPIClient(t *testing.T) {
	_, err := NewAPIClient(testEnvID, testAPIKey)

	if err != nil {
		t.Error("Api client V2 with API Key should not fail")
	}

	client, _ := NewAPIClient(testEnvID, testAPIKey)

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.urlTracking != defaultAPIURLTracking {
		t.Error("Api url should be set to default")
	}
}

func TestSendInternalHit(t *testing.T) {
	client, _ := NewAPIClient(testEnvID, testAPIKey)
	err := client.SendHit(testVisitorID, nil, nil)

	if err == nil {
		t.Error("Empty hit should return and err")
	}

	event := &model.EventHit{}
	event.SetBaseInfos(testEnvID, testVisitorID, nil)

	err = client.SendHit(testVisitorID, nil, event)

	if err == nil {
		t.Error("Invalid event hit should return error")
	}

	event.Action = "test_action"
	err = client.SendHit(testVisitorID, nil, event)

	if err != nil {
		t.Errorf("Right hit should not return and err : %v", err)
	}
}

func TestActivate(t *testing.T) {
	client, _ := NewAPIClient(realEnvID, testAPIKey)
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
	client, _ := NewAPIClient(realEnvID, testAPIKey)
	err := client.SendEvent(model.Event{})

	if err == nil {
		t.Errorf("Expected error for empty event request")
	}

	err = client.SendEvent(model.Event{
		VisitorID: "test_vid",
		Type:      model.CONTEXT,
		Data:      model.Context{},
	})

	if err != nil {
		t.Errorf("Did not expect error for correct event request. Got %v", err)
	}
}

func TestAnonymousID(t *testing.T) {
	client := NewMockAPIClient(realEnvID, false)

	_ = client.SendHit("vis_id", nil, &model.EventHit{
		Action: "action",
		Value:  1,
	})

	assert.Equal(t, `{"vid":"vis_id","cuid":"vis_id","cid":"blvo2kijq6pg023l8edg","t":"EVENT","ds":"APP","ea":"action","ev":1}`, client.requestString)

	anonymousID := "anon_id"
	_ = client.SendHit("vis_id", &anonymousID, &model.EventHit{
		Action: "action",
		Value:  1,
	})
	assert.Equal(t, `{"vid":"anon_id","cuid":"vis_id","cid":"blvo2kijq6pg023l8edg","t":"EVENT","ds":"APP","ea":"action","ev":1}`, client.requestString)
}
