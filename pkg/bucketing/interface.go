package bucketing

import "github.com/flagship-io/flagship-proto/bucketing"

// ConfigAPIInterface manage the bucketing configuration
type ConfigAPIInterface interface {
	GetConfiguration() (*bucketing.Bucketing_BucketingResponse, error)
}
