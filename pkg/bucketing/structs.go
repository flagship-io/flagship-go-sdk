package bucketing

import (
	"time"

	commonDecision "github.com/flagship-io/flagship-common/decision"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
	targetingTypes "github.com/flagship-io/flagship-proto/targeting"
)

// EngineOptions represents the options for the Bucketing decision mode
type EngineOptions struct {
	// PollingInterval is the number of milliseconds between each poll. If -1, then no polling will be done
	PollingInterval time.Duration
}

// Configuration represents a bucketing configuration
type Configuration struct {
	Panic     bool        `json:"panic"`
	Campaigns []*Campaign `json:"campaigns"`
}

// Campaign represents a bucketing campaign
type Campaign struct {
	ID              string            `json:"id"`
	CustomID        string            `json:"custom_id"`
	Type            string            `json:"type"`
	VariationGroups []*VariationGroup `json:"variationGroups"`
}

// VariationGroup represents a bucketing variation group
type VariationGroup struct {
	ID         string                    `json:"id"`
	Targeting  *targetingTypes.Targeting `json:"targeting"`
	Variations []*Variation              `json:"variations"`
}

func (vgi *VariationGroup) ToCommonStruct() *commonDecision.VariationsGroup {
	var variationsArray []*commonDecision.Variation
	for _, v := range vgi.Variations {
		variationsArray = append(variationsArray, v.ToCommonStruct())
	}
	return &commonDecision.VariationsGroup{
		ID:         vgi.ID,
		Targetings: vgi.Targeting,
		Variations: variationsArray,
	}
}

// Variation represents a bucketing variation
type Variation struct {
	ID            string             `json:"id"`
	Modifications model.Modification `json:"modifications"`
	Allocation    float32            `json:"allocation"`
	Reference     bool               `json:"reference"`
}

func (vi *Variation) ToCommonStruct() *commonDecision.Variation {
	return &commonDecision.Variation{
		ID:         vi.ID,
		Allocation: vi.Allocation,
	}
}
