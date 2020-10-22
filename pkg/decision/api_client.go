package decision

import (
	"github.com/abtasty/flagship-go-sdk/v2/pkg/decisionapi"
	"github.com/abtasty/flagship-go-sdk/v2/pkg/logging"
	"github.com/abtasty/flagship-go-sdk/v2/pkg/model"
)

var apiLogger = logging.CreateLogger("API Client")

// APIClient represents the API client informations
type APIClient struct {
	decisionAPIClient *decisionapi.APIClient
}

// NewAPIClient creates a decision API client with API options
func NewAPIClient(envID string, params ...func(*decisionapi.APIClient)) (*APIClient, error) {
	dAPIClient, err := decisionapi.NewAPIClient(envID, params...)
	if err != nil {
		return nil, err
	}
	res := APIClient{
		decisionAPIClient: dAPIClient,
	}

	return &res, nil
}

// GetModifications gets modifications from Decision API
func (r APIClient) GetModifications(visitorID string, context map[string]interface{}) (*model.APIClientResponse, error) {
	return r.decisionAPIClient.GetModifications(visitorID, context)
}
