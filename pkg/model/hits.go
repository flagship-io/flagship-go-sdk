package model

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// HitType express the type of the event
type HitType string

// The different event types
const (
	ACTIVATION  HitType = "ACTIVATION"
	CAMPAIGN    HitType = "CAMPAIGN"
	SCREEN      HitType = "SCREENVIEW"
	PAGE        HitType = "PAGEVIEW"
	EVENT       HitType = "EVENT"
	ITEM        HitType = "ITEM"
	TRANSACTION HitType = "TRANSACTION"
	BATCH       HitType = "BATCH"
)

// EventType express the type of the event
type EventType string

// The different event types
const (
	CONTEXT EventType = "CONTEXT"
)

// BaseHit represents the API client informations
type BaseHit struct {
	VisitorID               string    `json:"vid,omitempty"`
	EnvironmentID           string    `json:"cid,omitempty"`
	Type                    HitType   `json:"t,omitempty"`
	DataSource              string    `json:"ds,omitempty"`
	ProtocolVersion         string    `json:"v,omitempty"`
	UserIP                  string    `json:"uip,omitempty"`
	DocumentReferrer        string    `json:"dr,omitempty"`
	ViewportSize            string    `json:"vp,omitempty"`
	ScreenResolution        string    `json:"sr,omitempty"`
	Title                   string    `json:"pt,omitempty"`
	DocumentEncoding        string    `json:"de,omitempty"`
	ScreenColorDepth        string    `json:"sd,omitempty"`
	UserLanguage            string    `json:"ul,omitempty"`
	JavaEnabled             *bool     `json:"je,omitempty"`
	FlashVersion            string    `json:"fl,omitempty"`
	QueueTime               int64     `json:"qt,omitempty"`
	DocumentLocation        string    `json:"dl,omitempty"`
	CurrentSessionTimestamp int64     `json:"cst,omitempty"`
	SessionNumber           int64     `json:"sn,omitempty"`
	CreatedAt               time.Time `json:"-"`
}

// SetBaseInfos sets the mandatory information for the hit
func (b *BaseHit) SetBaseInfos(envID string, visitorID string) {
	b.EnvironmentID = envID
	b.VisitorID = visitorID
	b.DataSource = "APP"
	b.CreatedAt = time.Now()
}

func (b *BaseHit) validateBase() []error {
	errorsList := []error{}
	if b.VisitorID == "" {
		errorsList = append(errorsList, errors.New("Visitor ID should not by empty"))
	}
	if b.EnvironmentID == "" {
		errorsList = append(errorsList, errors.New("Environment ID should not by empty"))
	}
	if b.DataSource != "APP" {
		errorsList = append(errorsList, errors.New("DataSource should be APP"))
	}

	switch b.Type {
	case
		TRANSACTION,
		EVENT,
		PAGE,
		SCREEN,
		ITEM,
		CAMPAIGN,
		BATCH:
		break
	default:
		errorsList = append(errorsList, errors.New("Type is not handled"))
	}

	isScreenOrPageHit := b.Type == PAGE || b.Type == SCREEN
	if isScreenOrPageHit && b.DocumentLocation == "" {
		errorsList = append(errorsList, errors.New("Document location must not by empty for this hit PAGE or SCREEN"))
	} else if !isScreenOrPageHit && b.DocumentLocation != "" {
		errorsList = append(errorsList, errors.New("Document location must be empty for this type of hit"))
	}

	return errorsList
}

// ComputeQueueTime computes hit queue time
func (b *BaseHit) ComputeQueueTime() {
	b.QueueTime = int64((time.Since(b.CreatedAt)).Milliseconds())
}

// PageHit represents a pageview hit for the datacollect
type PageHit struct {
	BaseHit
}

// SetBaseInfos sets the mandatory information for the hit
func (b *PageHit) SetBaseInfos(envID string, visitorID string) {
	b.BaseHit.SetBaseInfos(envID, visitorID)
	b.Type = PAGE
}

// Validate checks that the hit is well formed
func (b *PageHit) Validate() []error {
	errorsList := b.validateBase()

	// Check url format
	_, err := url.ParseRequestURI(b.DocumentLocation)
	if err != nil {
		errorsList = append(errorsList, errors.New("Document location should be a real url for hit page"))
	}

	return errorsList
}

// ScreenHit represents a screenview hit for the datacollect
type ScreenHit struct {
	BaseHit
}

// SetBaseInfos sets the mandatory information for the hit
func (b *ScreenHit) SetBaseInfos(envID string, visitorID string) {
	b.BaseHit.SetBaseInfos(envID, visitorID)
	b.Type = SCREEN
}

// Validate checks that the hit is well formed
func (b *ScreenHit) Validate() []error {
	return b.validateBase()
}

// EventHit represents an event hit for the datacollect
type EventHit struct {
	BaseHit
	Action   string `json:"ea"`
	Category string `json:"ec,omitempty"`
	Label    string `json:"el,omitempty"`
	Value    int64  `json:"ev,omitempty"`
}

// SetBaseInfos sets the mandatory information for the hit
func (b *EventHit) SetBaseInfos(envID string, visitorID string) {
	b.BaseHit.SetBaseInfos(envID, visitorID)
	b.Type = EVENT
}

// Validate checks that the hit is well formed
func (b *EventHit) Validate() []error {
	errorsList := b.validateBase()
	if b.Action == "" {
		errorsList = append(errorsList, errors.New("Event Action should not by empty"))
	}
	return errorsList
}

// TransactionHit represents a transaction hit for the datacollect
type TransactionHit struct {
	BaseHit
	TransactionID  string  `json:"tid"`
	Affiliation    string  `json:"ta"`
	Revenue        float64 `json:"tr,omitempty"`
	Shipping       float64 `json:"ts,omitempty"`
	Tax            float64 `json:"tt,omitempty"`
	Currency       string  `json:"tc,omitempty"`
	CouponCode     string  `json:"tcc,omitempty"`
	PaymentMethod  string  `json:"pm,omitempty"`
	ShippingMethod string  `json:"sm,omitempty"`
	ItemCount      int     `json:"icn,omitempty"`
}

// SetBaseInfos sets the mandatory information for the hit
func (b *TransactionHit) SetBaseInfos(envID string, visitorID string) {
	b.BaseHit.SetBaseInfos(envID, visitorID)
	b.Type = TRANSACTION
}

// Validate checks that the hit is well formed
func (b *TransactionHit) Validate() []error {
	errorsList := b.validateBase()
	if b.TransactionID == "" {
		errorsList = append(errorsList, errors.New("Transaction ID should not by empty"))
	}
	if b.Affiliation == "" {
		errorsList = append(errorsList, errors.New("Transaction affiliation should not by empty"))
	}
	return errorsList
}

// ItemHit represents an item hit for the datacollect
type ItemHit struct {
	BaseHit
	TransactionID string  `json:"tid"`
	Name          string  `json:"in"`
	Price         float64 `json:"ip,omitempty"`
	Quantity      int     `json:"iq,omitempty"`
	Code          string  `json:"ic,omitempty"`
	Category      string  `json:"iv,omitempty"`
}

// SetBaseInfos sets the mandatory information for the hit
func (b *ItemHit) SetBaseInfos(envID string, visitorID string) {
	b.BaseHit.SetBaseInfos(envID, visitorID)
	b.Type = ITEM
}

// Validate checks that the hit is well formed
func (b *ItemHit) Validate() []error {
	errorsList := b.validateBase()
	if b.TransactionID == "" {
		errorsList = append(errorsList, errors.New("Item Transaction ID should not by empty"))
	}
	if b.Name == "" {
		errorsList = append(errorsList, errors.New("Item name should not by empty"))
	}
	if b.Code == "" {
		errorsList = append(errorsList, errors.New("Item code should not by empty"))
	}
	return errorsList
}

// ActivationHit represents an item hit for the datacollect
type ActivationHit struct {
	VisitorID        string    `json:"vid"`
	EnvironmentID    string    `json:"cid"`
	VariationGroupID string    `json:"caid"`
	VariationID      string    `json:"vaid"`
	CreatedAt        time.Time `json:"-"`
	QueueTime        int64     `json:"-"`
}

// SetBaseInfos sets the mandatory information for the hit
func (b *ActivationHit) SetBaseInfos(envID string, visitorID string) {
	b.EnvironmentID = envID
	b.VisitorID = visitorID
}

// Validate checks that the hit is well formed
func (b *ActivationHit) Validate() []error {
	errorsList := []error{}
	if b.VisitorID == "" {
		errorsList = append(errorsList, errors.New("Visitor ID should not by empty"))
	}
	if b.EnvironmentID == "" {
		errorsList = append(errorsList, errors.New("Environment ID should not by empty"))
	}
	if b.VariationGroupID == "" {
		errorsList = append(errorsList, errors.New("Campaign ID should not by empty"))
	}
	if b.VariationID == "" {
		errorsList = append(errorsList, errors.New("Variation should not by empty"))
	}
	return errorsList
}

// ComputeQueueTime computes hit queue time
func (b *ActivationHit) ComputeQueueTime() {
	b.QueueTime = int64((time.Since(b.CreatedAt)).Seconds())
}

// Event represents a context event hit for Flagship
type Event struct {
	VisitorID string    `json:"visitorId"`
	Type      EventType `json:"type"`
	Data      Context   `json:"data"`
	CreatedAt time.Time `json:"-"`
	QueueTime int64     `json:"-"`
}

// SetBaseInfos sets the mandatory information for the hit
func (b *Event) SetBaseInfos(envID string, visitorID string) {
	b.VisitorID = visitorID
}

// Validate checks that the hit is well formed
func (b *Event) Validate() []error {
	errorsList := []error{}
	if b.VisitorID == "" {
		errorsList = append(errorsList, errors.New("Visitor ID should not by empty"))
	}
	if b.Type != "CONTEXT" {
		errorsList = append(errorsList, fmt.Errorf("Type %s, is not handled", b.Type))
	}

	contextErrs := b.Data.Validate()
	errorsList = append(errorsList, contextErrs...)
	return errorsList
}

// ComputeQueueTime computes hit queue time
func (b *Event) ComputeQueueTime() {
	b.QueueTime = int64((time.Since(b.CreatedAt)).Seconds())
}

// BatchHit represents a batch of hits for the datacollect
type BatchHit struct {
	BaseHit
	Hits []HitInterface `json:"h"`
}

// SetBaseInfos sets the mandatory information for the hit
func (b *BatchHit) SetBaseInfos(envID string, visitorID string) {
	b.BaseHit.SetBaseInfos(envID, visitorID)
	b.Type = BATCH
}

// Validate checks that the hit is well formed
func (b *BatchHit) Validate() []error {
	return b.validateBase()
}

// AddHit adds a hit to the batch
func (b *BatchHit) AddHit(hit HitInterface) {
	b.Hits = append(b.Hits, hit)
}

func createBatchHit(baseHit BaseHit) BatchHit {
	bHit := BatchHit{
		BaseHit: baseHit,
		Hits:    []HitInterface{},
	}
	bHit.SetBaseInfos(bHit.EnvironmentID, bHit.VisitorID)
	return bHit
}
