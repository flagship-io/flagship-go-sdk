package bucketing

import (
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/cache"
	"github.com/flagship-io/flagship-proto/bucketing"
	"github.com/flagship-io/flagship-proto/decision_response"
	targetingTypes "github.com/flagship-io/flagship-proto/targeting"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var engineMockConfig = &bucketing.Bucketing_BucketingResponse{
	Campaigns: []*bucketing.Bucketing_BucketingCampaign{{
		Id: "test_cid",
		VariationGroups: []*bucketing.Bucketing_BucketingVariationGroups{{
			Id: "test_vgid",
			Targeting: &targetingTypes.Targeting{
				TargetingGroups: []*targetingTypes.Targeting_TargetingGroup{{
					Targetings: []*targetingTypes.Targeting_InnerTargeting{{
						Operator: targetingTypes.Targeting_EQUALS,
						Key:      wrapperspb.String("test"),
						Value:    structpb.NewBoolValue(true),
					}},
				}},
			},
			Variations: []*decision_response.FullVariation{{
				Id:         wrapperspb.String("1"),
				Allocation: 50,
				Modifications: &decision_response.Modifications{
					Type: decision_response.ModificationsType_FLAG,
					Value: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"test": structpb.NewBoolValue(true),
						},
					},
				},
			}, {
				Id:         wrapperspb.String("1"),
				Allocation: 50,
				Modifications: &decision_response.Modifications{
					Type: decision_response.ModificationsType_FLAG,
					Value: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"test": structpb.NewBoolValue(false),
						},
					},
				},
			}},
		}},
		BucketRanges: []*bucketing.Bucketing_BucketingCampaign_BucketRange{
			{
				R: []float64{0, 100},
			},
		},
	}},
}

// GetBucketingEngineMock returns a bucketing engine with mock config
func GetBucketingEngineMock(testEnvID string, cache cache.Manager) *Engine {
	engine, _ := NewEngine(testEnvID, nil)

	engine.apiClient = NewAPIClientMock(testEnvID, engineMockConfig, 200)
	engine.cacheManager = cache
	return engine
}
