package bucketing

import (
	"sync"
	"time"

	common "github.com/flagship-io/flagship-common"
	"github.com/flagship-io/flagship-common/targeting"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/cache"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/logging"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
	bucketingProto "github.com/flagship-io/flagship-proto/bucketing"
)

var logger = logging.CreateLogger("Bucketing Engine")

// EngineOptions represents the options for the Bucketing decision mode
type EngineOptions struct {
	// PollingInterval is the number of milliseconds between each poll. If -1, then no polling will be done
	PollingInterval time.Duration
}

// Engine represents a bucketing engine
type Engine struct {
	pollingInterval  time.Duration
	config           *bucketingProto.Bucketing_BucketingResponse
	apiClient        ConfigAPIInterface
	apiClientOptions []func(*APIClient)
	cacheManager     cache.Manager
	envID            string
	configMux        sync.RWMutex
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

	for range b.ticker.C {
		logger.Info("Bucketing engine ticked, loading configuration")
		err := b.Load()
		if err != nil {
			logger.Warnf("Bucketing engine load failed: %v", err)
		}
	}
}

func (b *Engine) getConfig() *bucketingProto.Bucketing_BucketingResponse {
	b.configMux.RLock()
	defer b.configMux.RUnlock()
	return b.config
}

// Load loads the env configuration in cache
func (b *Engine) Load() error {
	b.configMux.Lock()
	defer b.configMux.Unlock()
	newConfig, err := b.apiClient.GetConfiguration()

	if err != nil {
		logger.Error("Error when loading environment configuration", err)
		return err
	}

	b.config = newConfig

	return nil
}

func (b *Engine) getCampaignCache(visitorID string) cache.CampaignCacheMap {
	var campaignsCache = make(map[string]*cache.CampaignCache)
	if b.cacheManager != nil {
		campaignsCache, _ = b.cacheManager.Get(visitorID)
		if campaignsCache == nil {
			campaignsCache = make(map[string]*cache.CampaignCache)
		}
	}
	return campaignsCache
}

// GetModifications gets modifications from Decision API
func (b *Engine) GetModifications(visitorID string, anonymousID *string, context model.Context) (*model.APIClientResponse, error) {

	if b.getConfig() == nil {
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

	if b.getConfig().Panic {
		logger.Info("Environment is in panic mode. Skipping all campaigns")
		return resp, nil
	}

	campaignsCache := b.getCampaignCache(visitorID)

	config := b.getConfig()

	commonCampaigns := []*common.Campaign{}
	for _, v := range config.Campaigns {
		commonCampaigns = append(commonCampaigns, model.CampaignToCommonStruct(v))
	}
	anonymousIDString := ""
	if anonymousID != nil {
		anonymousIDString = *anonymousID
	}

	contextProto, err := context.ToProtoMap()
	if err != nil {
		logger.Errorf("error converting context to proto map: %v", err)
		return resp, nil
	}

	enableBucketAllocation := false
	decisionResponse, err := common.GetDecision(common.Visitor{
		ID:          visitorID,
		AnonymousID: anonymousIDString,
		Context: &targeting.Context{
			Standard: contextProto,
		},
	}, common.Environment{
		ID:                b.envID,
		Campaigns:         commonCampaigns,
		IsPanic:           config.Panic,
		SingleAssignment:  config.GetAccountSettings().GetEnabled1V1T(),
		UseReconciliation: config.GetAccountSettings().GetEnabledXPC(),
		CacheEnabled:      b.cacheManager != nil,
	}, common.DecisionOptions{
		EnableBucketAllocation: &enableBucketAllocation,
	}, common.DecisionHandlers{
		GetCache: func(environmentID, id string) (*common.VisitorAssignments, error) {
			return campaignsCache.ToCommonStruct(), nil
		},
	})

	if err != nil {
		logger.Errorf("error computing decision response: %v", err)
		return nil, err
	}

	for _, c := range decisionResponse.Campaigns {
		campaign := model.Campaign{
			ID:               c.Id.Value,
			VariationGroupID: c.VariationGroupId.Value,
			Variation: model.ClientVariation{
				ID:        c.Variation.Id.Value,
				Reference: c.Variation.Reference,
				Modifications: model.Modification{
					Type:  c.Variation.Modifications.Type.String(),
					Value: c.Variation.Modifications.Value.AsMap(),
				},
			},
		}
		resp.Campaigns = append(resp.Campaigns, campaign)

		keys := make([]string, 0, len(c.Variation.Modifications.Value.AsMap()))
		for k := range c.Variation.Modifications.Value.AsMap() {
			keys = append(keys, k)
		}

		alreadyActivated := false
		if campaignsCache[c.Id.Value] != nil {
			alreadyActivated = campaignsCache[c.Id.Value].Activated
		}
		campaignsCache[c.Id.Value] = &cache.CampaignCache{
			VariationGroupID: c.VariationGroupId.Value,
			VariationID:      c.Variation.Id.Value,
			Activated:        alreadyActivated,
			FlagKeys:         keys,
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
