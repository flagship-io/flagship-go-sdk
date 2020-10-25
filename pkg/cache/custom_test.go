package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomCache(t *testing.T) {
	notInitialized := &CustomManager{}
	_, err := notInitialized.Get("test")
	assert.Equal(t, "Custom cache manager not initialized", err.Error())

	err = notInitialized.Set("test", nil)
	assert.Equal(t, "Custom cache manager not initialized", err.Error())

	cacheCampaignsVisitors := map[string]map[string]*CampaignCache{}
	get := func(visitorID string) (map[string]*CampaignCache, error) {
		cacheCampaigns := cacheCampaignsVisitors[visitorID]
		return cacheCampaigns, nil
	}

	set := func(visitorID string, cache map[string]*CampaignCache) error {
		cacheCampaignsVisitors[visitorID] = cache
		return nil
	}

	cacheOptions := Options{}
	optionsFunc := WithCustomOptions(CustomOptions{
		Getter: get,
		Setter: set,
	})
	optionsFunc(&cacheOptions)

	_, err = cacheOptions.CustomOptions.Getter("test1")
	assert.Equal(t, nil, err)

	err = cacheOptions.CustomOptions.Setter("test1", nil)
	assert.Equal(t, nil, err)

	m, err := initCustomManager(cacheOptions.CustomOptions)

	assert.Equal(t, nil, err)

	_, err = m.Get("test")

	assert.NotEqual(t, nil, err)

	cache := map[string]*CampaignCache{}
	cache["testC"] = &CampaignCache{VariationGroupID: "vgID"}
	err = m.Set("test", cache)

	assert.Equal(t, nil, err)

	r, err := m.Get("test")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, r["testC"])
}
