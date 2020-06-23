package cache

import (
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestInitManager(t *testing.T) {
	cacheCampaignsVisitors := map[string]map[string]*CampaignCache{}
	get := func(visitorID string) (map[string]*CampaignCache, error) {
		cacheCampaigns := cacheCampaignsVisitors[visitorID]
		return cacheCampaigns, nil
	}

	set := func(visitorID string, cache map[string]*CampaignCache) error {
		cacheCampaignsVisitors[visitorID] = cache
		return nil
	}

	// Test custom
	cache, err := InitManager(WithCustomOptions(CustomOptions{
		Getter: get,
		Setter: set,
	}))

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, cache.(*CustomManager))

	// Test local
	testFolder := "test"
	cache, err = InitManager(WithLocalOptions(LocalOptions{
		DbPath: testFolder,
	}))

	lCache := cache.(*LocalDBManager)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, lCache)
	lCache.Dispose()
	os.RemoveAll(testFolder)

	// Test Redis
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	cache, err = InitManager(WithRedisOptions(RedisOptions{
		Host: s.Addr(),
	}))
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, cache.(*RedisManager))
}
