package decision

import (
	"github.com/flagship-io/flagship-go-sdk/v3/pkg/decisionapi"
	"github.com/flagship-io/flagship-go-sdk/v3/pkg/logging"
	"github.com/flagship-io/flagship-go-sdk/v3/pkg/model"
)

var apiLogger = logging.CreateLogger("API Client")

// APIClient represents the API client informations
type APIClient struct {
	decisionAPIClient *decisionapi.APIClient
}

// NewAPIClient creates a decision API client with API options
func NewAPIClient(envID string, apiKey string, params ...func(*decisionapi.APIClient)) (*APIClient, error) {
	dAPIClient, err := decisionapi.NewAPIClient(envID, apiKey, params...)
	if err != nil {
		return nil, err
	}
	res := APIClient{
		decisionAPIClient: dAPIClient,
	}

	return &res, nil
}

// GetModifications gets modifications from Decision API
func (r *APIClient) GetModifications(visitorID string, anonymousID *string, context model.Context) (*model.APIClientResponse, error) {
	apiLogger.Info("Getting modifications from API")
	return r.decisionAPIClient.GetModifications(visitorID, anonymousID, context)
}
