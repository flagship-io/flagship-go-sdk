package model

import (
	"time"

	"github.com/flagship-io/flagship-common/decision"
	"github.com/flagship-io/flagship-proto/bucketing"
	"github.com/flagship-io/flagship-proto/decision_response"
)

// APIOptions represents the options for the Decision API Client
type APIOptions struct {
	APIUrl  string
	APIKey  string
	Timeout time.Duration
	Retries int
}

// APIClientRequest represents the API client informations
type APIClientRequest struct {
	VisitorID   string  `json:"visitor_id"`
	AnonymousID *string `json:"anonymous_id"`
	Context     Context `json:"context"`
	TriggerHit  bool    `json:"trigger_hit"`
}

// APIClientResponse represents a decision response
type APIClientResponse struct {
	VisitorID string     `json:"visitorId"`
	Panic     bool       `json:"panic"`
	Campaigns []Campaign `json:"campaigns"`
}

// Campaign represents a decision campaign
type Campaign struct {
	ID               string          `json:"id"`
	CustomID         string          `json:"-"`
	VariationGroupID string          `json:"variationGroupId"`
	Variation        ClientVariation `json:"variation"`
}

// ClientVariation represents a decision campaign variation
type ClientVariation struct {
	ID            string       `json:"id"`
	Modifications Modification `json:"modifications"`
	Reference     bool         `json:"reference"`
}

// Modification represents a decision campaign variation modification
type Modification struct {
	Type  string                 `json:"type"`
	Value map[string]interface{} `json:"value"`
}

// FlagInfos represents a decision campaign variation modification
type FlagInfos struct {
	Value    interface{}
	Campaign Campaign
}

func VariationToCommonStruct(v *decision_response.FullVariation) *decision.Variation {
	return &decision.Variation{
		ID:            v.Id.Value,
		Reference:     v.Reference,
		Allocation:    float32(v.Allocation),
		Modifications: v.Modifications,
	}
}

func VariationGroupToCommonStruct(vg *bucketing.Bucketing_BucketingVariationGroups, campaign *bucketing.Bucketing_BucketingCampaign) *decision.VariationsGroup {
	variations := []*decision.Variation{}
	for _, v := range vg.Variations {
		variations = append(variations, VariationToCommonStruct(v))
	}
	bucketRange := [][]float64{}
	for _, r := range campaign.BucketRanges {
		bucketRange = append(bucketRange, r.R)
	}
	return &decision.VariationsGroup{
		ID: vg.Id,
		Campaign: &decision.CampaignInfo{
			ID:           campaign.Id,
			Type:         campaign.Type,
			BucketRanges: bucketRange,
		},
		Targetings: vg.Targeting,
		Variations: variations,
	}
}

func CampaignToCommonStruct(c *bucketing.Bucketing_BucketingCampaign) *decision.CampaignInfo {
	variationGroups := map[string]*decision.VariationsGroup{}
	for _, vg := range c.VariationGroups {
		variationGroups[vg.Id] = VariationGroupToCommonStruct(vg, c)
	}
	bucketRange := [][]float64{}
	for _, r := range c.BucketRanges {
		bucketRange = append(bucketRange, r.R)
	}
	return &decision.CampaignInfo{
		ID:               c.Id,
		CustomID:         &c.CustomId,
		Type:             c.Type,
		VariationsGroups: variationGroups,
		BucketRanges:     bucketRange,
	}
}
