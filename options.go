package flagship

import (
	"github.com/flagship-io/flagship-go-sdk/v3/pkg/bucketing"
	"github.com/flagship-io/flagship-go-sdk/v3/pkg/cache"
	"github.com/flagship-io/flagship-go-sdk/v3/pkg/decisionapi"
	"github.com/flagship-io/flagship-go-sdk/v3/pkg/tracking"
)

// Options represent the options passed to the Flagship SDK client
type FlagshipOptions struct {
	EnvID               string
	APIKey              string
	decisionMode        DecisionMode
	bucketingOptions    []func(*bucketing.Engine)
	decisionAPIOptions  []func(*decisionapi.APIClient)
	cacheManagerOptions []cache.OptionBuilder
	trackingAPIClient   tracking.APIClientInterface
}

// OptionBuilder is a func type to set options to the FlagshipOption.
type OptionBuilder func(*FlagshipOptions)

// BuildOptions fill out the FlagshipOption struct from option builders
func (f *FlagshipOptions) BuildOptions(clientOptions ...OptionBuilder) {
	f.decisionMode = MODE_API

	// extract options
	for _, opt := range clientOptions {
		opt(f)
	}
}

// WithBucketing enables the bucketing decision mode for the SDK
func WithBucketing(options ...func(*bucketing.Engine)) OptionBuilder {
	return func(f *FlagshipOptions) {
		f.decisionMode = MODE_BUCKETING
		f.bucketingOptions = options
	}
}

// WithDecisionAPI changes the decision API options
func WithDecisionAPI(options ...func(*decisionapi.APIClient)) OptionBuilder {
	return func(f *FlagshipOptions) {
		f.decisionAPIOptions = options
	}
}

// WithVisitorCache enables visitor assignment caching with options
func WithVisitorCache(options ...cache.OptionBuilder) OptionBuilder {
	return func(f *FlagshipOptions) {
		f.cacheManagerOptions = options
	}
}

// WithTrackingAPIClient changes the tracking api client used by the SDK
func WithTrackingAPIClient(trackingAPIClient tracking.APIClientInterface) OptionBuilder {
	return func(f *FlagshipOptions) {
		f.trackingAPIClient = trackingAPIClient
	}
}
