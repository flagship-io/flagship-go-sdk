package cache

import (
	"errors"
)

// CustomManager represents the local db manager object
type CustomManager struct {
	getter func(visitorID string) (map[string]*CampaignCache, error)
	setter func(visitorID string, campaignCache map[string]*CampaignCache) error
}

// CustomOptions are the options necessary to make the local cache manager work
type CustomOptions struct {
	Getter func(visitorID string) (map[string]*CampaignCache, error)
	Setter func(visitorID string, campaignCache map[string]*CampaignCache) error
}

// WithCustomOptions configures custom manager options
func WithCustomOptions(customOptions CustomOptions) func(options *Options) {
	return func(options *Options) {
		options.cacheType = Custom
		options.CustomOptions = customOptions
	}
}

func initCustomManager(customOptions CustomOptions) (m *CustomManager, err error) {
	m = &CustomManager{}
	if customOptions.Getter == nil {
		err = errors.New("Missing getter function")
	}

	if customOptions.Getter == nil {
		err = errors.New("Missing setter function")
	}

	m.getter = customOptions.Getter
	m.setter = customOptions.Setter

	return m, err
}

// Set saves the campaigns in cache for this visitor
func (m *CustomManager) Set(visitorID string, campaignCache map[string]*CampaignCache) (err error) {
	if m.setter == nil {
		return errors.New("Custom cache manager not initialized")
	}

	err = m.setter(visitorID, campaignCache)

	return err
}

// Get returns the campaigns in cache for this visitor
func (m *CustomManager) Get(visitorID string) (cache map[string]*CampaignCache, err error) {
	if m.getter == nil {
		return nil, errors.New("Custom cache manager not initialized")
	}

	cache, err = m.getter(visitorID)
	if err == nil && cache == nil {
		err = errors.New("Key does not exist")
	}
	return cache, err
}
