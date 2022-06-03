package client

import (
	"errors"
	"testing"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/bucketing"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/cache"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/tracking"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/decision"
	"github.com/stretchr/testify/assert"
)

var caID = "cid"
var vgID = "vgid"
var testVID = "vid"

func createVisitor(vID string, context model.Context, options ...VisitorOptionBuilder) *Visitor {
	client := createClient()
	client.decisionClient = createMockClient()
	client.trackingAPIClient = &FakeTrackingAPIClient{}

	visitor, _ := client.NewVisitor(vID, context, options...)
	return visitor
}

func createMockClient() decision.ClientInterface {
	modification := model.Modification{
		Type: "FLAG",
		Value: map[string]interface{}{
			"test_string": "string",
			"test_bool":   true,
			"test_number": 35.6,
			"test_nil":    nil,
			"test_object": map[string]interface{}{
				"test_key": true,
			},
			"test_array": []interface{}{true},
		},
	}
	variation := model.ClientVariation{
		ID:            testVID,
		Reference:     true,
		Modifications: modification,
	}
	return decision.NewAPIClientMock(testEnvID, &model.APIClientResponse{
		VisitorID: "test_vid",
		Campaigns: []model.Campaign{
			{
				ID:               caID,
				VariationGroupID: vgID,
				Variation:        variation,
			},
		},
	}, 200)
}

func TestGenerateID(t *testing.T) {
	visitor := createVisitor("", nil)
	assert.NotEqual(t, "", visitor.ID)
}

func TestUpdateContext(t *testing.T) {
	visitor := createVisitor("test", nil)

	context := model.Context{}
	context["test_string"] = "123"
	context["test_number"] = 36.5
	context["test_bool"] = true
	context["test_wrong"] = errors.New("wrong type")

	err := visitor.UpdateContext(context)

	if err == nil {
		t.Error("Visitor with wrong context variable should raise an error")
	}

	delete(context, "test_wrong")

	err = visitor.UpdateContext(context)

	if err != nil {
		t.Errorf("Visitor update context raised an error : %v", err)
		return
	}

	if visitor.Context["test_string"] != "123" {
		t.Errorf("Visitor update context string failed. Expected %s, got %s", "123", visitor.Context["test_string"])
	}
	if visitor.Context["test_number"] != 36.5 {
		t.Errorf("Visitor update context string failed. Expected %f, got %v", 36.5, visitor.Context["test_number"])
	}
	if visitor.Context["test_bool"] != true {
		t.Errorf("Visitor update context string failed. Expected %v, got %v", true, visitor.Context["test_bool"])
	}
}

func TestUpdateContextKey(t *testing.T) {
	context := model.Context{}
	context["test_string"] = "123"
	context["test_number"] = 36.5
	context["test_bool"] = true

	visitor := createVisitor("test", context)

	err := visitor.UpdateContextKey("test_error", errors.New("wrong type"))

	if err == nil {
		t.Error("Visitor with wrong context variable should raise an error")
	}

	delete(context, "test_wrong")

	err = visitor.UpdateContextKey("test_ok", true)

	if err != nil {
		t.Errorf("Visitor update context raised an error : %v", err)
	}

	if visitor.Context["test_ok"] != true {
		t.Errorf("Visitor update context string failed. Expected %v, got %v", true, visitor.Context["test_ok"])
	}
}

func TestAuthenticate(t *testing.T) {
	context := map[string]interface{}{}
	visitor := createVisitor("firstID", context)
	err := visitor.Authenticate("newID", nil, false)
	assert.Nil(t, err)
	assert.Equal(t, "newID", visitor.ID)
	assert.Equal(t, "firstID", *visitor.AnonymousID)

	newContext := model.Context{
		"test": "string",
	}
	visitor.Authenticate("newerID", newContext, false)
	assert.Equal(t, "newerID", visitor.ID)
	assert.Equal(t, newContext, visitor.Context)
	assert.Equal(t, "firstID", *visitor.AnonymousID)

	visitor.decisionMode = API
	newContext = model.Context{
		"test2": "string",
	}
	err = visitor.Unauthenticate(newContext, false)
	assert.Nil(t, err)
	assert.Equal(t, "firstID", visitor.ID)
	assert.Equal(t, newContext, visitor.Context)
	assert.Nil(t, visitor.AnonymousID)

	visitor = createVisitor("firstID", context, WithAuthenticated(false))
	assert.Nil(t, visitor.AnonymousID)

	visitor = createVisitor("firstID", context, WithAuthenticated(true))
	assert.NotNil(t, visitor.AnonymousID)
}

func TestSynchronizeModifications(t *testing.T) {
	visitor := &Visitor{}
	err := visitor.SynchronizeModifications()
	if err == nil {
		t.Error("Flag synchronization without visitorID should raise an error")
	}

	visitor = createVisitor("test", nil)

	errorMock := decision.NewAPIClientMock(testEnvID, nil, 400)
	visitor.decisionClient = errorMock

	err = visitor.SynchronizeModifications()
	if err == nil {
		t.Error("Flag synchronization should have raised the http error")
	}

	visitor = createVisitor("test", nil)

	flag, ok := visitor.GetAllModifications()["test_string"]
	if ok {
		t.Errorf("Flag should be nil before synchronization. Got %v", flag)
	}

	err = visitor.SynchronizeModifications()
	if err != nil {
		t.Errorf("Flag synchronization should not raise error. Got %v", err)
	}

	_, ok = visitor.GetAllModifications()["test_string"]

	if !ok {
		t.Errorf("Flag should exist after synchronization")
	}
}

func TestGetModification(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	_, err := visitor.getModification("not_exists", true)
	assert.NotEqual(t, nil, err, "Should raise an error as modifications are not synced")

	// Test infos before sync
	_, err = visitor.GetModificationInfo("not_exists")
	assert.NotEqual(t, nil, err, "Should raise an error as modifications are not synced")

	visitor.SynchronizeModifications()

	// Test default value
	val, err := visitor.getModification("not_exists", true)
	assert.Nil(t, err, "Should not have an error when flag does not exists")
	assert.Equal(t, nil, val, "Expected nil value")

	// Test infos of missing key
	_, err = visitor.GetModificationInfo("not_exists")
	assert.Nil(t, err, "Should not have an error when flag does not exists")

	// Test response value
	val, err = visitor.getModification("test_string", true)
	assert.Equal(t, nil, err, "Should not have an error as flag exists")
	assert.Equal(t, "string", val, "Expected string value")

	// Test modification info response value
	infos, err := visitor.GetModificationInfo("test_string")
	assert.Equal(t, nil, err, "Should not have an error as flag exists")

	assert.Equal(t, caID, infos.CampaignID)
	assert.Equal(t, vgID, infos.VariationGroupID)
	assert.Equal(t, testVID, infos.VariationID)
	assert.Equal(t, true, infos.IsReference)
	assert.Equal(t, "string", infos.Value)
}

func TestGetModificationBool(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	_, err := visitor.GetModificationBool("not_exists", false, true)
	assert.NotEqual(t, nil, err, "Should have an error as modifications are not synced")

	visitor.SynchronizeModifications()

	// Test default value
	val, err := visitor.GetModificationBool("not_exists", false, true)
	assert.Nil(t, err, "Should not have an error when flag does not exists")
	assert.Equal(t, false, val, "Expected default value getting nil flag")

	// Test wrong type value
	val, err = visitor.GetModificationBool("test_string", false, true)
	assert.NotEqual(t, nil, err, "Should have an error as flag test_string is not of type bool")
	assert.Equal(t, false, val, "Expected default value getting nil flag")

	// Test nil value
	val, err = visitor.GetModificationBool("test_nil", false, true)
	assert.Equal(t, nil, err, "Did not expect error when getting nil flag")
	assert.Equal(t, false, val, "Expected default value getting nil flag")

	// Test response value
	val, err = visitor.GetModificationBool("test_bool", false, true)
	assert.Equal(t, nil, err, "Should not have an error as flag does exists")
	assert.Equal(t, true, val, "Expected value true")
}

func TestGetModificationNumber(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	_, err := visitor.GetModificationNumber("not_exists", 10, true)
	assert.NotEqual(t, nil, err, "Should have an error as modifications are not synced")

	visitor.SynchronizeModifications()

	// Test default value
	val, err := visitor.GetModificationNumber("not_exists", 10, true)
	assert.Nil(t, err, "Should not have an error when flag does not exists")
	assert.Equal(t, 10., val, "Expected default value getting nil flag")

	// Test wrong type value
	val, err = visitor.GetModificationNumber("test_string", 10, true)
	assert.NotEqual(t, nil, err, "Should have an error as flag test_string is not of type float")
	assert.Equal(t, 10., val, "Expected default value getting nil flag")

	// Test nil value
	val, err = visitor.GetModificationNumber("test_nil", 10, true)
	assert.Equal(t, nil, err, "Did not expect error when getting nil flag")
	assert.Equal(t, 10., val, "Expected default value getting nil flag")

	// Test response value
	val, err = visitor.GetModificationNumber("test_number", 10, true)
	assert.Equal(t, nil, err, "Should not have an error as flag does exists")
	assert.Equal(t, 35.6, val, "Expected value 36.5")
}

func TestGetModificationString(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	_, err := visitor.GetModificationString("not_exists", "default", true)
	assert.NotEqual(t, nil, err, "Should have an error as modifications are not synced")

	visitor.SynchronizeModifications()

	// Test default value
	val, err := visitor.GetModificationString("not_exists", "default", true)
	assert.Nil(t, err, "Should not have an error when flag does not exists")
	assert.Equal(t, "default", val, "Expected default value getting nil flag")

	// Test wrong type value
	val, err = visitor.GetModificationString("test_bool", "default", true)
	assert.NotEqual(t, nil, err, "Should have an error as flag test_string is not of type float")
	assert.Equal(t, "default", val, "Expected default value getting nil flag")

	// Test nil value
	val, err = visitor.GetModificationString("test_nil", "default", true)
	assert.Equal(t, nil, err, "Did not expect error when getting nil flag")
	assert.Equal(t, "default", val, "Expected default value getting nil flag")

	// Test response value
	val, err = visitor.GetModificationString("test_string", "default", true)
	assert.Equal(t, nil, err, "Did not expect error when getting nil flag")
	assert.Equal(t, "string", val, "Expected value string")
}

func TestGetModificationObject(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	_, err := visitor.GetModificationObject("not_exists", nil, true)
	assert.NotEqual(t, nil, err, "Should have an error as modifications are not synced")

	visitor.SynchronizeModifications()

	defaultValue := map[string]interface{}{
		"default_key": false,
	}
	// Test default value
	val, err := visitor.GetModificationObject("not_exists", defaultValue, true)
	assert.Nil(t, err, "Should not have an error when flag does not exists")
	assert.Equal(t, defaultValue["default_key"], val["default_key"])

	// Test wrong type value
	val, err = visitor.GetModificationObject("test_bool", defaultValue, true)
	assert.NotEqual(t, nil, err, "Should have an error as flag does not exists")
	assert.Equal(t, defaultValue["default_key"], val["default_key"])

	// Test nil value
	val, err = visitor.GetModificationObject("test_nil", defaultValue, true)
	assert.Equal(t, nil, err, "Did not expect error when getting nil flag")
	assert.Equal(t, defaultValue["default_key"], val["default_key"])

	// Test response value
	val, err = visitor.GetModificationObject("test_object", defaultValue, true)
	assert.Equal(t, nil, err, "Should not have an error as flag exists")
	assert.Equal(t, true, val["test_key"])
}

func TestGetModificationArray(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	_, err := visitor.GetModificationArray("not_exists", nil, true)
	assert.NotEqual(t, nil, err, "Should have an error as modifications are not synced")

	visitor.SynchronizeModifications()

	defaultValue := []interface{}{true}
	// Test default value
	val, err := visitor.GetModificationArray("not_exists", defaultValue, true)
	assert.Nil(t, err, "Should not have an error when flag does not exists")
	assert.Equal(t, defaultValue[0], val[0])

	// Test wrong type value
	val, err = visitor.GetModificationArray("test_bool", defaultValue, true)
	assert.NotEqual(t, nil, err, "Should have an error as flag does not exists")
	assert.Equal(t, defaultValue[0], val[0])

	// Test nil value
	val, err = visitor.GetModificationArray("test_nil", defaultValue, true)
	assert.Equal(t, nil, err, "Did not expect error when getting nil flag")
	assert.Equal(t, defaultValue[0], val[0])

	// Test response value
	val, err = visitor.GetModificationArray("test_array", defaultValue, true)
	assert.Equal(t, nil, err, "Should not have an error as flag exists")
	assert.Equal(t, true, val[0])
}

func TestActivateModification(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	err := visitor.ActivateModification("not_exists")
	assert.NotEqual(t, nil, err, "Should raise an error as modifications are not synced")

	visitor.SynchronizeModifications()

	// Test default value
	err = visitor.ActivateModification("not_exists")
	assert.Nil(t, err, "Should not have an error when flag does not exists")

	// Test response value
	err = visitor.ActivateModification("test_string")
	assert.Equal(t, nil, err, "Should not have an error as flag exists")
}

func TestActivateModificationCache(t *testing.T) {
	// Test engine with cache
	cacheCampaignsVisitors := map[string]map[string]*cache.CampaignCache{}
	get := func(visitorID string) (map[string]*cache.CampaignCache, error) {
		cacheCampaigns := cacheCampaignsVisitors[visitorID]
		return cacheCampaigns, nil
	}

	set := func(visitorID string, cache map[string]*cache.CampaignCache) error {
		cacheCampaignsVisitors[visitorID] = cache
		return nil
	}

	cache, _ := cache.InitManager(cache.WithCustomOptions(cache.CustomOptions{
		Getter: get,
		Setter: set,
	}))

	client, _ := Create(&Options{
		EnvID:  testEnvID,
		APIKey: testAPIKey,
	})
	client.cacheManager = cache

	engine := bucketing.GetBucketingEngineMock(testEnvID, cache)
	client.decisionClient = engine
	client.decisionMode = Bucketing
	client.trackingAPIClient = &FakeTrackingAPIClient{}

	visitor, _ := client.NewVisitor("test", map[string]interface{}{
		"test": true,
	})

	// Test before sync
	err := visitor.ActivateCacheModification("not_exists")

	if err == nil {
		t.Errorf("Should have an error as modifications are not synced")
	}

	visitor.SynchronizeModifications()

	// Test default value
	err = visitor.ActivateCacheModification("not_exists")

	if err == nil {
		t.Errorf("Should have an error as flag does not exists")
	}

	// Test response value
	err = visitor.ActivateCacheModification("test")

	if err != nil {
		t.Errorf("Should not have an error as flag does exists. Got %v", err)
	}
}

func TestSendHitVisitor(t *testing.T) {
	visitor := createVisitor("test", nil)
	trackingAPIClient, _ := tracking.NewAPIClient(testEnvID, testAPIKey)
	visitor.trackingAPIClient = trackingAPIClient
	err := visitor.SendHit(&model.EventHit{})

	if err == nil {
		t.Errorf("Expected error as hit is malformed.")
	}

	err = visitor.SendHit(&model.EventHit{
		Action: "test_action",
	})
	if err != nil {
		t.Errorf("Did not expect error as hit is correct. Got %v", err)
	}
}
