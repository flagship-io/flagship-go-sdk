package client

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/cache"
	"github.com/abtasty/flagship-go-sdk/pkg/decisionapi"
	"github.com/abtasty/flagship-go-sdk/pkg/model"

	"github.com/stretchr/testify/assert"

	"github.com/abtasty/flagship-go-sdk/pkg/bucketing"
	"github.com/abtasty/flagship-go-sdk/pkg/decision"
)

var testEnvID = "test_env_id"
var vID = "test_visitor_id"
var realEnvID = "blvo2kijq6pg023l8edg"
var testAPIKey = "test_api_key"

func createClient() *Client {
	client, _ := Create(&Options{
		EnvID:  testEnvID,
		APIKey: testAPIKey,
	})
	return client
}

func TestCreate(t *testing.T) {
	options := &Options{
		EnvID: testEnvID,
	}

	client, err := Create(options)

	assert.NotEqual(t, nil, err)

	options.APIKey = testAPIKey

	client, err = Create(options)

	assert.Equal(t, nil, err)
	assert.Equal(t, testEnvID, client.envID)
	assert.NotEqual(t, nil, client.decisionClient)
	assert.NotEqual(t, nil, client.trackingAPIClient)
	assert.Equal(t, nil, client.cacheManager)
	assert.Equal(t, testEnvID, client.GetEnvID())
}

func TestCreateBucketing(t *testing.T) {
	options := &Options{
		EnvID:  testEnvID,
		APIKey: testAPIKey,
	}
	options.BuildOptions(WithBucketing())
	client, err := Create(options)

	assert.NotNil(t, err)

	if len(options.bucketingOptions) != 0 {
		t.Errorf(
			"Bucketing Client default options wrong. Expected default %v, got %v",
			0,
			len(options.bucketingOptions))
	}

	options.EnvID = realEnvID
	options.BuildOptions(WithBucketing(bucketing.PollingInterval(2 * time.Second)))

	client, err = Create(options)
	assert.Nil(t, err)
	assert.Equal(t, realEnvID, client.envID)

	if client.decisionClient == nil {
		t.Error("decision Bucketing Client has not been initialized")
	}

	bucketing, castOK := client.decisionClient.(*bucketing.Engine)

	if !castOK {
		t.Errorf("decision Bucketing Client has not been initialized correctly")
	}

	pollingInterval := reflect.ValueOf(bucketing).Elem().FieldByName("pollingInterval")
	if pollingInterval.Int() != (2 * time.Second).Nanoseconds() {
		t.Errorf(
			"decision Bucketing Client polling interval wrong. Expected %v, got %v",
			(2 * time.Second).Nanoseconds(),
			pollingInterval.Int())
	}

	if client.trackingAPIClient == nil {
		t.Error("tracking API Client has not been initialized")
	}
}

func TestCreateAPIVersion(t *testing.T) {
	options := &Options{
		EnvID:  testEnvID,
		APIKey: testAPIKey,
	}

	options.BuildOptions(WithDecisionAPI(decisionapi.APIVersion(2), decisionapi.APIKey("testapi")))

	client, err := Create(options)

	if err != nil {
		t.Errorf("Error when creating flagship client : %v", err)
	}

	_, castOK := client.decisionClient.(*decision.APIClient)
	if !castOK {
		t.Errorf("decision API Client has not been initialized correctly")
	}
}

func TestCreateCache(t *testing.T) {
	options := &Options{
		EnvID:  testEnvID,
		APIKey: testAPIKey,
	}

	get := func(visitorID string) (map[string]*cache.CampaignCache, error) {
		return nil, nil
	}

	set := func(visitorID string, cache map[string]*cache.CampaignCache) error {
		return nil
	}

	options.BuildOptions(WithVisitorCache(cache.WithCustomOptions(cache.CustomOptions{
		Getter: get,
		Setter: set,
	})))

	client, err := Create(options)

	if err != nil {
		t.Errorf("Error when creating flagship client : %v", err)
	}

	assert.Nil(t, err)
	assert.NotNil(t, client.cacheManager)

	testGet, err := client.cacheManager.Get(testVID)
	assert.NotNil(t, err)
	assert.Nil(t, testGet)
}

func TestInit(t *testing.T) {
	client := createClient()

	assert.Equal(t, testEnvID, client.envID)
	assert.Equal(t, testAPIKey, client.apiKey)
	assert.NotEqual(t, nil, client.decisionClient)
	assert.NotEqual(t, nil, client.trackingAPIClient)
}

func TestCreateVisitor(t *testing.T) {
	client := createClient()

	context := map[string]interface{}{}
	context["test_string"] = "123"
	context["test_number"] = 36.5
	context["test_bool"] = true
	context["test_int"] = 4
	context["test_wrong"] = errors.New("wrong type")

	_, err := client.NewVisitor("", nil)

	if err != nil {
		t.Error("Empty visitor ID should raise an error")
	}

	_, err = client.NewVisitor(vID, context)

	if err == nil {
		t.Error("Visitor with wrong context variable should raise an error")
	}

	_, conv64Ok := context["test_int"].(float64)
	if !conv64Ok {
		t.Errorf("Integer context key has not been converted. Got %v", context["test_int"])
	}

	delete(context, "test_wrong")

	visitor, err := client.NewVisitor(vID, context)
	if err != nil {
		t.Errorf("Visitor creation failed. error : %v", err)
	}

	if visitor == nil {
		t.Error("Visitor creation failed. Visitor is null")
	}

	if visitor.ID != vID {
		t.Error("Visitor creation failed. Visitor id is not set")
	}

	for key, val := range context {
		valV, exists := visitor.Context[key]
		if !exists {
			t.Errorf("Visitor creation failed. Visitor context key %s is not set", key)
		}
		if val != valV {
			t.Errorf("Visitor creation failed. Visitor context key %s value %v is wrong. Should be %v", key, valV, val)
		}
	}
}

func TestSendHitClient(t *testing.T) {
	client := createClient()

	err := client.SendHit(vID, &model.EventHit{})

	if err == nil {
		t.Errorf("Expected error as hit is malformed.")
	}

	err = client.SendHit(vID, &model.EventHit{
		Action: "test_action",
	})
	if err != nil {
		t.Errorf("Did not expect error as hit is correct. Got %v", err)
	}
}
