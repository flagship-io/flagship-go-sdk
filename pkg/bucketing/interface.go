package bucketing

// ConfigAPIInterface manage the bucketing configuration
type ConfigAPIInterface interface {
	GetConfiguration() (*Configuration, error)
}
