package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestRedisCache(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	// Wrong host test
	_, err = initRedisManager(RedisOptions{
		Host: "localhost:4567",
	})
	assert.NotEqual(t, nil, err)

	notInitialized := &RedisManager{}
	_, err = notInitialized.Get("test")
	assert.Equal(t, "Redis cache manager not initialized", err.Error())

	err = notInitialized.Set("test", nil)
	assert.Equal(t, "Redis cache manager not initialized", err.Error())

	cacheOptions := Options{}
	optionsFunc := WithRedisOptions(RedisOptions{
		Host: s.Addr(),
	})
	optionsFunc(&cacheOptions)
	assert.Equal(t, s.Addr(), cacheOptions.RedisOptions.Host)

	m, err := initRedisManager(cacheOptions.RedisOptions)

	assert.Equal(t, nil, err)

	r, err := m.Get("test")

	var nullResp map[string]*CampaignCache
	assert.NotEqual(t, nil, err)
	assert.Equal(t, nullResp, r)

	cache := map[string]*CampaignCache{}
	cache["testC"] = &CampaignCache{VariationGroupID: "vgID"}
	err = m.Set("test", cache)

	assert.Equal(t, nil, err)

	r, err = m.Get("test")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, r["testC"])
}
