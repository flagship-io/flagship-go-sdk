package model

import (
	"time"
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
	VisitorID   string                 `json:"visitor_id"`
	AnonymousID *string                `json:"anonymous_id"`
	Context     map[string]interface{} `json:"context"`
	TriggerHit  bool                   `json:"trigger_hit"`
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
