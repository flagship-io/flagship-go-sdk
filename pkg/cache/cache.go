package cache

import (
	"github.com/abtasty/flagship-go-sdk/pkg/logging"
)

// ManagerType represents infrastructure types of cache manager
type ManagerType string

const (
	// Local is a local database based cache manager
	Local ManagerType = "local"
	// Redis is a redis based cache manager
	Redis ManagerType = "redis"
	// Custom is a custom based cache manager that requires set / get implementation
	Custom ManagerType = "custom"
)

// CampaignCache expresses the campaign cache object to be saved for a visitor
type CampaignCache struct {
	VariationGroupID string
	VariationID      string
	FlagKeys         []string
}

// Options expresses all the possible options for cache manager
type Options struct {
	cacheType ManagerType
	RedisOptions
	LocalOptions
	CustomOptions
}

// OptionBuilder is a func type to set options to the FlagshipOption.
type OptionBuilder func(*Options)

// Manager is the interface that exposes cache manager functions
type Manager interface {
	Set(visitorID string, campaignInfos map[string]*CampaignCache) error
	Get(visitorID string) (map[string]*CampaignCache, error)
}

var cacheLogger = logging.CreateLogger("cache")

// InitManager initialize the manager with a type and options
func InitManager(optionsFunc ...OptionBuilder) (manager Manager, err error) {
	options := &Options{}

	for _, o := range optionsFunc {
		o(options)
	}

	switch options.cacheType {
	case Local:
		manager, err = initLocalDBManager(options.LocalOptions)
	case Custom:
		manager, err = initCustomManager(options.CustomOptions)
	case Redis:
		manager, err = initRedisManager(options.RedisOptions)
	}

	return manager, err
}
