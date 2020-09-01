package client

import (
	"github.com/abtasty/flagship-go-sdk/pkg/bucketing"
	"github.com/abtasty/flagship-go-sdk/pkg/cache"
	"github.com/abtasty/flagship-go-sdk/pkg/decisionapi"
)

// Options represent the options passed to the Flagship SDK client
type Options struct {
	EnvID               string
	APIKey              string
	decisionMode        DecisionMode
	bucketingOptions    []func(*bucketing.Engine)
	decisionAPIOptions  []func(*decisionapi.APIClient)
	cacheManagerOptions []cache.OptionBuilder
}

// OptionBuilder is a func type to set options to the FlagshipOption.
type OptionBuilder func(*Options)

// BuildOptions fill out the FlagshipOption struct from option builders
func (f *Options) BuildOptions(clientOptions ...OptionBuilder) {
	f.decisionMode = API

	// extract options
	for _, opt := range clientOptions {
		opt(f)
	}
}

// WithBucketing enables the bucketing decision mode for the SDK
func WithBucketing(options ...func(*bucketing.Engine)) OptionBuilder {
	return func(f *Options) {
		f.decisionMode = Bucketing
		f.bucketingOptions = options
	}
}

// WithDecisionAPI changes the decision API options
func WithDecisionAPI(options ...func(*decisionapi.APIClient)) OptionBuilder {
	return func(f *Options) {
		f.decisionAPIOptions = options
	}
}

// WithVisitorCache enables visitor assignment caching with options
func WithVisitorCache(options ...cache.OptionBuilder) OptionBuilder {
	return func(f *Options) {
		f.cacheManagerOptions = options
	}
}
