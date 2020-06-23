package bucketing

import (
	"github.com/abtasty/flagship-go-sdk/pkg/cache"
	"github.com/abtasty/flagship-go-sdk/pkg/model"
)

var engineMockConfig = &Configuration{
	Campaigns: []*Campaign{{
		ID: "test_cid",
		VariationGroups: []*VariationGroup{{
			ID: "test_vgid",
			Targeting: TargetingWrapper{
				TargetingGroups: []*TargetingGroup{{
					Targetings: []*Targeting{{
						Operator: EQUALS,
						Key:      "test",
						Value:    true,
					}},
				}},
			},
			Variations: []*Variation{{
				ID:         "1",
				Allocation: 50,
				Modifications: model.Modification{
					Type:  "FLAG",
					Value: map[string]interface{}{"test": true},
				},
			}, {
				ID:         "2",
				Allocation: 50,
				Modifications: model.Modification{
					Type:  "FLAG",
					Value: map[string]interface{}{"test": false},
				},
			}},
		}},
	}},
}

// GetBucketingEngineMock returns a bucketing engine with mock config
func GetBucketingEngineMock(testEnvID string, cache cache.Manager) *Engine {
	engine, _ := NewEngine(testEnvID, nil)

	engine.apiClient = NewAPIClientMock(testEnvID, engineMockConfig, 200)
	engine.cacheManager = cache
	return engine
}
