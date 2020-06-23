package cache

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalCache(t *testing.T) {
	testFolder := "test"

	notInitialized := &LocalDBManager{}
	_, err := notInitialized.Get("test")
	assert.Equal(t, "Cache db manager not initialized", err.Error())

	err = notInitialized.Set("test", nil)
	assert.Equal(t, "Cache db manager not initialized", err.Error())

	cacheOptions := Options{}
	optionsFunc := WithLocalOptions(LocalOptions{
		DbPath: testFolder,
	})
	optionsFunc(&cacheOptions)
	assert.Equal(t, testFolder, cacheOptions.LocalOptions.DbPath)

	m, err := initLocalDBManager(cacheOptions.LocalOptions)

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

	err = m.Dispose()
	assert.Equal(t, nil, err)

	err = os.RemoveAll(testFolder)
	assert.Equal(t, nil, err)
}
