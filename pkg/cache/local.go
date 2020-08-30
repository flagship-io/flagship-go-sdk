package cache

import (
	"encoding/json"
	"errors"

	"github.com/prologic/bitcask"
)

// LocalDBManager represents the local db manager object
type LocalDBManager struct {
	db *bitcask.Bitcask
}

// LocalOptions are the options necessary to make the local cache manager work
type LocalOptions struct {
	DbPath string
}

// WithLocalOptions configures local options for manager
func WithLocalOptions(localOptions LocalOptions) func(options *Options) {
	return func(options *Options) {
		options.cacheType = Local
		options.LocalOptions = localOptions
	}
}

func initLocalDBManager(localOptions LocalOptions) (m *LocalDBManager, err error) {
	db, err := bitcask.Open(localOptions.DbPath)
	if err != nil {
		return nil, err
	}

	m = &LocalDBManager{
		db: db,
	}

	return m, nil
}

// Set saves the campaigns in cache for this visitor
func (m *LocalDBManager) Set(visitorID string, campaignCache map[string]*CampaignCache) error {
	if m.db == nil {
		return errors.New("Cache db manager not initialized")
	}

	cache, err := json.Marshal(campaignCache)

	if err == nil {
		err = m.db.Put([]byte(visitorID), cache)
	}

	return err
}

// Get returns the campaigns in cache for this visitor
func (m *LocalDBManager) Get(visitorID string) (map[string]*CampaignCache, error) {
	if m.db == nil {
		return nil, errors.New("Cache db manager not initialized")
	}

	data, err := m.db.Get([]byte(visitorID))

	if err != nil {
		return nil, err
	}

	campaignCache := map[string]*CampaignCache{}
	err = json.Unmarshal(data, &campaignCache)

	if err != nil {
		return nil, err
	}

	return campaignCache, nil
}

// Dispose frees IO resources
func (m *LocalDBManager) Dispose() error {
	if m.db == nil {
		return nil
	}
	return m.db.Close()
}
