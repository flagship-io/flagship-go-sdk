package bucketing

import (
	"fmt"
	"sync"
	"time"

	"github.com/abtasty/flagship-go-sdk/v2/pkg/cache"
	"github.com/abtasty/flagship-go-sdk/v2/pkg/logging"
	"github.com/abtasty/flagship-go-sdk/v2/pkg/model"
)

var logger = logging.CreateLogger("Bucketing Engine")

// Engine represents a bucketing engine
type Engine struct {
	pollingInterval  time.Duration
	config           *Configuration
	apiClient        ConfigAPIInterface
	apiClientOptions []func(*APIClient)
	cacheManager     cache.Manager
	envID            string
	configMux        sync.Mutex
	ticker           *time.Ticker
}

// PollingInterval sets the polling interval for the bucketing engine
func PollingInterval(interval time.Duration) func(r *Engine) {
	return func(r *Engine) {
		r.pollingInterval = interval
	}
}

// APIOptions sets the func option for the engine client API
func APIOptions(apiOptions ...func(*APIClient)) func(r *Engine) {
	return func(r *Engine) {
		r.apiClientOptions = apiOptions
	}
}

// NewEngine creates a new engine for bucketing
func NewEngine(envID string, cacheManager cache.Manager, params ...func(*Engine)) (*Engine, error) {
	engine := &Engine{
		pollingInterval:  1 * time.Minute,
		envID:            envID,
		apiClientOptions: []func(*APIClient){},
		cacheManager:     cacheManager,
	}

	for _, param := range params {
		param(engine)
	}

	engine.apiClient = NewAPIClient(envID, engine.apiClientOptions...)

	err := engine.Load()

	if engine.pollingInterval != -1 {
		go engine.startTicker()
	}

	return engine, err
}

// startTicker starts new ticker for polling bucketing infos
func (b *Engine) startTicker() {
	if b.ticker != nil {
		return
	}
	b.ticker = time.NewTicker(b.pollingInterval)

	for {
		select {
		case <-b.ticker.C:
			logger.Info("Bucketing engine ticked, loading configuration")
			b.Load()
			return
		}
	}
}

// Load loads the env configuration in cache
func (b *Engine) Load() error {
	newConfig, err := b.apiClient.GetConfiguration()

	if err != nil {
		logger.Error("Error when loading environment configuration", err)
		return err
	}

	b.configMux.Lock()
	b.config = newConfig
	b.configMux.Unlock()

	return nil
}

// GetModifications gets modifications from Decision API
func (b *Engine) GetModifications(visitorID string, context map[string]interface{}) (*model.APIClientResponse, error) {
	if b.config == nil {
		logger.Info("Configuration not loaded. Loading it now")
		err := b.Load()
		if err != nil {
			logger.Warning("Configuration could not be loaded.")
			return nil, err
		}
	}

	resp := &model.APIClientResponse{
		VisitorID: visitorID,
		Campaigns: []model.Campaign{},
	}

	if b.config.Panic {
		logger.Info("Environment is in panic mode. Skipping all campaigns")
		return resp, nil
	}

	var campaignsCache = make(map[string]*cache.CampaignCache)
	if b.cacheManager != nil {
		campaignsCache, _ = b.cacheManager.Get(visitorID)
		if campaignsCache == nil {
			campaignsCache = make(map[string]*cache.CampaignCache)
		}
	}

	for _, c := range b.config.Campaigns {
		var matchedVg *VariationGroup
		for _, vg := range c.VariationGroups {
			matched, err := TargetingMatch(vg, visitorID, context)
			if err != nil {
				logger.Warning(fmt.Sprintf("Error occurred when checking targeting : %v", err))
				continue
			}

			if matched {
				matchedVg = vg
				break
			}
		}

		if matchedVg != nil {
			var variation *Variation
			var err error

			// Handle cache campaigns
			cacheCampaign, ok := campaignsCache[c.ID]
			if ok && cacheCampaign.VariationGroupID == matchedVg.ID {
				for _, v := range matchedVg.Variations {
					if v.ID == cacheCampaign.VariationID {
						variation = v
					}
				}
			}

			if variation == nil {
				variation, err = GetRandomAllocation(visitorID, matchedVg)
			}

			if err != nil {
				logger.Warning(fmt.Sprintf("Error occurred when allocating variation : %v", err))
				continue
			}
			campaign := model.Campaign{
				ID:               c.ID,
				VariationGroupID: matchedVg.ID,
				Variation: model.ClientVariation{
					ID:        variation.ID,
					Reference: variation.Reference,
					Modifications: model.Modification{
						Type:  variation.Modifications.Type,
						Value: variation.Modifications.Value,
					},
				},
			}
			resp.Campaigns = append(resp.Campaigns, campaign)

			keys := make([]string, 0, len(variation.Modifications.Value))

			for k := range variation.Modifications.Value {
				keys = append(keys, k)
			}

			campaignsCache[c.ID] = &cache.CampaignCache{
				VariationGroupID: matchedVg.ID,
				VariationID:      variation.ID,
				FlagKeys:         keys,
			}
		}
	}

	if b.cacheManager != nil {
		err := b.cacheManager.Set(visitorID, campaignsCache)
		if err != nil {
			logger.Warnf("Cache saving failed: %v", err)
		}
	}

	return resp, nil
}
