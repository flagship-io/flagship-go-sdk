package client

import (
	"errors"
	"testing"

	"github.com/abtasty/flagship-go-sdk/pkg/bucketing"
	"github.com/abtasty/flagship-go-sdk/pkg/cache"
	"github.com/abtasty/flagship-go-sdk/pkg/model"

	"github.com/abtasty/flagship-go-sdk/pkg/decision"
	"github.com/stretchr/testify/assert"
)

var caID = "cid"
var vgID = "vgid"
var testVID = "vid"

func createVisitor(vID string, context map[string]interface{}) *Visitor {
	client := createClient()
	client.decisionClient = createMockClient()

	visitor, _ := client.NewVisitor(vID, context)
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
		},
	}
	variation := model.ClientVariation{
		ID:            testVID,
		Reference:     false,
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

func TestUpdateContext(t *testing.T) {
	visitor := createVisitor("test", nil)

	context := map[string]interface{}{}
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
	context := map[string]interface{}{}
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

	if err == nil {
		t.Errorf("getModification should raise an error as modifications are not synced")
	}

	// Test infos before sync
	_, err = visitor.GetModificationInfo("not_exists")

	if err == nil {
		t.Errorf("GetModificationInfo Should raise an error as modifications are not synced")
	}

	visitor.SynchronizeModifications()

	// Test default value
	val, err := visitor.getModification("not_exists", true)

	if err == nil {
		t.Errorf("Should have an error as modifications are not synced")
	}

	if val != nil {
		t.Errorf("Expected default value, got %v", val)
	}

	// Test infos of missing key
	_, err = visitor.GetModificationInfo("not_exists")

	if err == nil {
		t.Errorf("GetModificationInfo should raise an error as modification key does not exist")
	}

	// Test response value
	val, err = visitor.getModification("test_string", true)

	if err != nil {
		t.Errorf("Should have an error as flag exists. Got %v", err)
	}

	if val != "string" {
		t.Errorf("Expected string value, got %v", val)
	}

	// Test modification info response value
	infos, err := visitor.GetModificationInfo("test_string")

	if err != nil {
		t.Errorf("Should have an error as flag exists. Got %v", err)
	}

	assert.Equal(t, caID, infos.CampaignID)
	assert.Equal(t, vgID, infos.VariationGroupID)
	assert.Equal(t, testVID, infos.VariationID)
	assert.Equal(t, "string", infos.Value)
}

func TestGetModificationBool(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	_, err := visitor.GetModificationBool("not_exists", false, true)

	if err == nil {
		t.Errorf("Should have an error as modifications are not synced")
	}

	visitor.SynchronizeModifications()

	// Test default value
	val, err := visitor.GetModificationBool("not_exists", false, true)

	if err == nil {
		t.Errorf("Should have an error as flag does not exists")
	}

	if val != false {
		t.Errorf("Expected default value, got %v", val)
	}

	// Test wrong type value
	val, err = visitor.GetModificationBool("test_string", true, true)

	if err == nil {
		t.Errorf("Should have an error as flag test_string is not of type bool")
	}

	if val != true {
		t.Errorf("Expected default value true, got %v", val)
	}

	// Test nil value
	val, err = visitor.GetModificationBool("test_nil", true, true)

	assert.Equal(t, nil, err, "Did not expect error when getting nil flag")
	assert.Equal(t, true, val, "Expected default value getting nil flag")

	// Test response value
	val, err = visitor.GetModificationBool("test_bool", false, true)

	if err != nil {
		t.Errorf("Should have an error as flag does exists. Got %v", err)
	}

	if val != true {
		t.Errorf("Expected value true, got %v", val)
	}
}

func TestGetModificationNumber(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	_, err := visitor.GetModificationNumber("not_exists", 10, true)

	if err == nil {
		t.Errorf("Should have an error as modifications are not synced")
	}

	visitor.SynchronizeModifications()

	// Test default value
	val, err := visitor.GetModificationNumber("not_exists", 10, true)

	if err == nil {
		t.Errorf("Should have an error as flag does not exists")
	}

	if val != 10 {
		t.Errorf("Expected value 10, got %v", val)
	}

	// Test wrong type value
	val, err = visitor.GetModificationNumber("test_string", 10, true)

	if err == nil {
		t.Errorf("Should have an error as flag test_string is not of type float")
	}

	if val != 10 {
		t.Errorf("Expected default value 10, got %v", val)
	}

	// Test nil value
	val, err = visitor.GetModificationNumber("test_nil", 10, true)

	assert.Equal(t, nil, err, "Did not expect error when getting nil flag")
	assert.Equal(t, 10., val, "Expected default value getting nil flag")

	// Test response value
	val, err = visitor.GetModificationNumber("test_number", 10, true)

	if err != nil {
		t.Errorf("Should not have an error as flag does exists. Got %v", err)
	}

	if val != 35.6 {
		t.Errorf("Expected value 36.5, got %v", val)
	}
}

func TestGetModificationString(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	_, err := visitor.GetModificationString("not_exists", "default", true)

	if err == nil {
		t.Errorf("Should have an error as modifications are not synced")
	}

	visitor.SynchronizeModifications()

	// Test default value
	val, err := visitor.GetModificationString("not_exists", "default", true)

	if err == nil {
		t.Errorf("Should have an error as flag does not exists")
	}

	if val != "default" {
		t.Errorf("Expected value default, got %v", val)
	}

	// Test wrong type value
	val, err = visitor.GetModificationString("test_bool", "default", true)

	if err == nil {
		t.Errorf("Should have an error as flag test_string is not of type float")
	}

	if val != "default" {
		t.Errorf("Expected default value default, got %v", val)
	}

	// Test nil value
	val, err = visitor.GetModificationString("test_nil", "default", true)

	assert.Equal(t, nil, err, "Did not expect error when getting nil flag")
	assert.Equal(t, "default", val, "Expected default value getting nil flag")

	// Test response value
	val, err = visitor.GetModificationString("test_string", "default", true)

	if err != nil {
		t.Errorf("Should have an error as flag does exists. Got %v", err)
	}

	if val != "string" {
		t.Errorf("Expected value string, got %v", val)
	}
}

func TestActivateModification(t *testing.T) {
	visitor := createVisitor("test", nil)

	// Test before sync
	err := visitor.ActivateModification("not_exists")

	if err == nil {
		t.Errorf("Should have an error as modifications are not synced")
	}

	visitor.SynchronizeModifications()

	// Test default value
	err = visitor.ActivateModification("not_exists")

	if err == nil {
		t.Errorf("Should have an error as flag does not exists")
	}

	// Test response value
	err = visitor.ActivateModification("test_string")

	if err != nil {
		t.Errorf("Should have an error as flag does exists. Got %v", err)
	}
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
		EnvID: testEnvID,
	})
	client.cacheManager = cache

	engine := bucketing.GetBucketingEngineMock(testEnvID, cache)
	client.decisionClient = engine
	client.decisionMode = Bucketing

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
