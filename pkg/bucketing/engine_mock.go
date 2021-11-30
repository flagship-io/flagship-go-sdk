package bucketing

import (
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/cache"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
	targetingTypes "github.com/flagship-io/flagship-proto/targeting"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var engineMockConfig = &Configuration{
	Campaigns: []*Campaign{{
		ID: "test_cid",
		VariationGroups: []*VariationGroup{{
			ID: "test_vgid",
			Targeting: &targetingTypes.Targeting{
				TargetingGroups: []*targetingTypes.Targeting_TargetingGroup{{
					Targetings: []*targetingTypes.Targeting_InnerTargeting{{
						Operator: targetingTypes.Targeting_EQUALS,
						Key:      wrapperspb.String("test"),
						Value:    structpb.NewBoolValue(true),
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
