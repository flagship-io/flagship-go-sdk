package flagship

// Start creates and returns a Client with the given environment ID and functional options
func Start(envID string, APIKey string, clientOptions ...OptionBuilder) (*Client, error) {
	options := &FlagshipOptions{
		EnvID:  envID,
		APIKey: APIKey,
	}

	options.BuildOptions(clientOptions...)

	return NewInstance(options)
}
