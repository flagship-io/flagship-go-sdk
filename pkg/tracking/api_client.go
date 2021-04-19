package tracking

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/decisionapi"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/logging"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/utils"
)

const defaultAPIURLTracking = "https://ariane.abtasty.com"

var apiLogger = logging.CreateLogger("DataCollect API")

// APIClient represents the API client informations
type APIClient struct {
	urlTracking        string
	envID              string
	httpClientTracking *utils.HTTPClient
	decisionAPIClient  *decisionapi.APIClient
}

// NewAPIClient creates a tracking API Client with environment ID and option builders
func NewAPIClient(envID string, apiKey string, params ...func(r *decisionapi.APIClient)) (*APIClient, error) {
	res := APIClient{
		envID: envID,
	}

	decisionAPIClient, err := decisionapi.NewAPIClient(envID, apiKey, params...)

	if err != nil {
		return nil, err
	}

	res.decisionAPIClient = decisionAPIClient

	if res.urlTracking == "" {
		res.urlTracking = defaultAPIURLTracking
	}

	httpClientTracking := utils.NewHTTPClient(res.urlTracking, utils.HTTPOptions{})
	res.httpClientTracking = httpClientTracking

	return &res, nil
}

// SendHit sends a tracking hit to the Data Collect API
func (r *APIClient) SendHit(visitorID string, hit model.HitInterface) error {
	if hit == nil {
		err := errors.New("Hit should not be empty")
		apiLogger.Error(err.Error(), err)
		return err
	}

	hit.SetBaseInfos(r.envID, visitorID)

	errs := hit.Validate()
	if len(errs) > 0 {
		for _, e := range errs {
			apiLogger.Errorf("Hit validation error : %v", e)
		}
		return errors.New("Hit validation failed")
	}
	hit.ComputeQueueTime()

	b, err := json.Marshal(hit)

	if err != nil {
		return err
	}

	apiLogger.Info(fmt.Sprintf("Sending hit : %v", string(b)))
	resp, err := r.httpClientTracking.Call("", "POST", b, nil)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error when calling activation API : %v", err)
	}

	return nil
}

// ActivateCampaign activate a campaign / variation id to the Decision API
func (r *APIClient) ActivateCampaign(request model.ActivationHit) error {
	return r.decisionAPIClient.ActivateCampaign(request)
}

// SendEvent sends an event to the Flagship event collection
func (r *APIClient) SendEvent(request model.Event) error {
	return r.decisionAPIClient.SendEvent(request)
}
