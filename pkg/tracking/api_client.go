package tracking

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/abtasty/flagship-go-sdk/v2/pkg/decisionapi"

	"github.com/abtasty/flagship-go-sdk/v2/pkg/model"

	"github.com/abtasty/flagship-go-sdk/v2/pkg/logging"
	"github.com/abtasty/flagship-go-sdk/v2/pkg/utils"
)

const defaultTimeout = 2 * time.Second
const defaultAPIURLTracking = "https://ariane.abtasty.com"

var apiLogger = logging.CreateLogger("DataCollect API")

// APIClient represents the API client informations
type APIClient struct {
	urlTracking        string
	urlDecision        string
	envID              string
	decisionTimeout    time.Duration
	apiKey             string
	httpClientTracking *utils.HTTPClient
	decisionAPIClient  *decisionapi.APIClient
}

// NewAPIClient creates a tracking API Client with environment ID and option builders
func NewAPIClient(envID string, params ...func(r *decisionapi.APIClient)) (*APIClient, error) {
	res := APIClient{
		envID: envID,
	}

	decisionAPIClient, err := decisionapi.NewAPIClient(envID, params...)

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
func (r APIClient) SendHit(visitorID string, hit model.HitInterface) error {
	if hit == nil {
		err := errors.New("Hit should not be empty")
		apiLogger.Error(err.Error(), err)
		return err
	}

	hit.SetBaseInfos(r.envID, visitorID)

	errs := hit.Validate()
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			apiLogger.Errorf("Hit validation error : %v", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return errors.New("Hit validation failed")
	}
	hit.ComputeQueueTime()

	b, err := json.Marshal(hit)

	if err != nil {
		return err
	}

	apiLogger.Info(fmt.Sprintf("Sending hit : %v", string(b)))
	resp, err := r.httpClientTracking.Call("", "POST", bytes.NewBuffer(b), nil)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error when calling activation API : %v", err)
	}

	return nil
}

// ActivateCampaign activate a campaign / variation id to the Decision API
func (r APIClient) ActivateCampaign(request model.ActivationHit) error {
	return r.decisionAPIClient.ActivateCampaign(request)
}

// SendEvent sends an event to the Flagship event collection
func (r APIClient) SendEvent(request model.Event) error {
	return r.decisionAPIClient.SendEvent(request)
}
