package bucketing

import (
	"time"

	"github.com/abtasty/flagship-go-sdk/pkg/model"
)

// TargetingOperator express a targeting operator
type TargetingOperator string

// The different targeting operators
const (
	NULL                   TargetingOperator = "NULL"
	LOWER_THAN             TargetingOperator = "LOWER_THAN"
	GREATER_THAN_OR_EQUALS TargetingOperator = "GREATER_THAN_OR_EQUALS"
	LOWER_THAN_OR_EQUALS   TargetingOperator = "LOWER_THAN_OR_EQUALS"
	EQUALS                 TargetingOperator = "EQUALS"
	NOT_EQUALS             TargetingOperator = "NOT_EQUALS"
	STARTS_WITH            TargetingOperator = "STARTS_WITH"
	ENDS_WITH              TargetingOperator = "ENDS_WITH"
	CONTAINS               TargetingOperator = "CONTAINS"
	NOT_CONTAINS           TargetingOperator = "NOT_CONTAINS"
	GREATER_THAN           TargetingOperator = "GREATER_THAN"
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
	Type            string            `json:"type"`
	VariationGroups []*VariationGroup `json:"variationGroups"`
}

// VariationGroup represents a bucketing variation group
type VariationGroup struct {
	ID         string           `json:"id"`
	Targeting  TargetingWrapper `json:"targeting"`
	Variations []*Variation     `json:"variations"`
}

// Variation represents a bucketing variation
type Variation struct {
	ID            string             `json:"id"`
	Modifications model.Modification `json:"modifications"`
	Allocation    int                `json:"allocation"`
	Reference     bool               `json:"reference"`
}

// TargetingWrapper represents a bucketing targeting wrapper
type TargetingWrapper struct {
	TargetingGroups []*TargetingGroup `json:"targetingGroups"`
}

// TargetingGroup represents a bucketing targeting group ('or' linked targetings)
type TargetingGroup struct {
	Targetings []*Targeting `json:"targetings"`
}

// Targeting represents a bucketing targeting group ('or' linked targetings)
type Targeting struct {
	Operator TargetingOperator `json:"operator"`
	Key      string            `json:"key"`
	Value    interface{}       `json:"value"`
}
