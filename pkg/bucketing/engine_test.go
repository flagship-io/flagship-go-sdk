package bucketing

import (
	"reflect"
	"testing"
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/cache"
	"github.com/stretchr/testify/assert"
)

var testVID = "test_vid"
var testContext = map[string]interface{}{}

func TestNewEngine(t *testing.T) {
	engine, err := NewEngine(testEnvID, nil)

	if err == nil {
		t.Error("Bucketing engine creation should return an error for incorrect envID")
	}

	if engine == nil {
		t.Error("Bucketing engine should not be nil")
	}

	if engine.envID != testEnvID {
		t.Errorf("Bucketing engine env ID incorrect. Expected %v, got %v", testEnvID, engine.envID)
	}

	url := "http://google.fr"
	engine, err = NewEngine(testEnvID, nil, APIOptions(APIUrl(url)))

	if err == nil {
		t.Error("Bucketing engine creation should return an error for incorrect url")
	}

	if engine == nil {
		t.Error("Bucketing engine should not be nil")
	}

	apiClient, castOK := engine.apiClient.(*APIClient)
	if !castOK {
		t.Errorf("bucketing API Client has not been initialized correctly")
	}

	urlClient := reflect.ValueOf(apiClient).Elem().FieldByName("url")
	assert.Equal(t, url, urlClient.String())
}

func TestLoad(t *testing.T) {
	engine, _ := NewEngine(testEnvID, nil)
	err := engine.Load()

	if err == nil {
		t.Error("Expected error for incorrect env ID")
	}

	engine, _ = NewEngine(realEnvID, nil)
	err = engine.Load()

	if err != nil {
		t.Errorf("Unexpected error for correct env ID: %v", err)
	}
}

func TestGetModifications(t *testing.T) {
	engine, _ := NewEngine(testEnvID, nil)

	modifs, err := engine.GetModifications(testVID, testContext)

	if err == nil {
		t.Errorf("Expected error for test env ID")
	}

	if modifs != nil {
		t.Errorf("Unexpected modifs for test env ID. Got %v", modifs)
	}

	engine, _ = NewEngine(realEnvID, nil)

	_, err = engine.GetModifications(testVID, testContext)

	if err != nil {
		t.Errorf("Unexpected error for correct env ID: %v", err)
	}
}

func TestGetModificationsMock(t *testing.T) {
	engine := GetBucketingEngineMock(testEnvID, nil)

	modifs, err := engine.GetModifications(testVID, map[string]interface{}{"test": true})

	if err != nil {
		t.Errorf("Unexpected error for correct env ID: %v", err)
	}
	assert.Equal(t, 1, len(modifs.Campaigns))

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
	engine, _ = NewEngine(testEnvID, cache)
	engine.apiClient = NewAPIClientMock(testEnvID, engineMockConfig, 200)

	modifs, err = engine.GetModifications(testVID, map[string]interface{}{"test": true})

	if err != nil {
		t.Errorf("Unexpected error for correct env ID: %v", err)
	}
	assert.Equal(t, 1, len(modifs.Campaigns))

	// Check cache is set
	cacheCheck, _ := cache.Get(testVID)
	assert.Equal(t, 1, len(cacheCheck))

	campaignCacheCheck := cacheCheck["test_cid"]
	assert.Equal(t, "1", campaignCacheCheck.VariationID)
	assert.Equal(t, 1, len(campaignCacheCheck.FlagKeys))
	assert.Equal(t, "test", campaignCacheCheck.FlagKeys[0])

	// Check new GetModifications return cache
	modifs, err = engine.GetModifications(testVID, map[string]interface{}{"test": true})
	assert.Equal(t, 1, len(modifs.Campaigns))
	assert.Equal(t, campaignCacheCheck.VariationID, modifs.Campaigns[0].Variation.ID)

	// Setting panic
	engineMockConfig.Panic = true
	engine.apiClient = NewAPIClientMock(testEnvID, engineMockConfig, 200)

	modifs, err = engine.GetModifications(testVID, map[string]interface{}{"test": true})

	if err != nil {
		t.Errorf("Unexpected error for correct env ID: %v", err)
	}
	assert.Equal(t, 0, len(modifs.Campaigns))
	engineMockConfig.Panic = false
}

func TestPollingPanic(t *testing.T) {
	engine, _ := NewEngine(testEnvID, nil, PollingInterval(1*time.Second))

	config := &Configuration{
		Campaigns: []*Campaign{{
			ID: "test_cid",
		}},
	}

	engine.apiClient = NewAPIClientMock(testEnvID, config, 200)
	time.Sleep(1100 * time.Millisecond)

	assert.Equal(t, 1, len(engine.config.Campaigns))
	assert.Equal(t, false, engine.config.Panic)

	// Setting panic
	config.Panic = true

	time.Sleep(1100 * time.Millisecond)

	assert.Equal(t, 1, len(engine.config.Campaigns))
	assert.Equal(t, true, engine.config.Panic)
}
