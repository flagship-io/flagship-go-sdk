package flagship

import (
	"github.com/abtasty/flagship-go-sdk/pkg/client"
)

// Start creates and returns a Client with the given environment ID and functional options
func Start(envID string, APIKey string, clientOptions ...client.OptionBuilder) (*client.Client, error) {
	flagshipOptions := &client.Options{
		EnvID:  envID,
		APIKey: APIKey,
	}

	flagshipOptions.BuildOptions(clientOptions...)

	return client.Create(flagshipOptions)
}
