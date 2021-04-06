package decisionapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abtasty/flagship-go-sdk/v2/pkg/logging"
	"github.com/abtasty/flagship-go-sdk/v2/pkg/model"
	"github.com/abtasty/flagship-go-sdk/v2/pkg/utils"
)

const defaultTimeout = 2 * time.Second
const defaultV1APIURL = "https://decision-api.flagship.io/v1"
const defaultV2APIURL = "https://decision.flagship.io/v2"

var apiLogger = logging.CreateLogger("Decision API")

// APIClient represents the Decision API client informations
type APIClient struct {
	url        string
	envID      string
	apiKey     string
	timeout    time.Duration
	retries    int
	httpClient utils.HTTPClientInterface
}

// APIVersionNumber specifies the version of the Decision API to use
type APIVersionNumber int

const (
	// V1 is Decision API V1
	V1 = iota + 1
	// V2 is Decision API V2
	V2
)

// APIVersion sets http client base URL
func APIVersion(version APIVersionNumber) func(r *APIClient) {
	return func(r *APIClient) {
		switch version {
		case V1:
			r.url = defaultV1APIURL
		case V2:
			r.url = defaultV2APIURL
		}
	}
}

// APIKey sets http client api key
func APIKey(apiKey string) func(r *APIClient) {
	return func(r *APIClient) {
		r.apiKey = apiKey
	}
}

// Timeout sets http client timeout
func Timeout(timeout time.Duration) func(r *APIClient) {
	return func(r *APIClient) {
		r.timeout = timeout
	}
}

// Retries sets max number of retries for failed calls
func Retries(retries int) func(r *APIClient) {
	return func(r *APIClient) {
		r.retries = retries
	}
}

// NewAPIClient creates a Decision API client from the environment ID and option builders
func NewAPIClient(envID string, apiKey string, params ...func(*APIClient)) (*APIClient, error) {
	res := APIClient{
		envID:   envID,
		apiKey:  apiKey,
		retries: 1,
	}

	headers := map[string]string{}

	for _, param := range params {
		param(&res)
	}

	if res.url == "" {
		res.url = defaultV2APIURL
	}

	if res.url == defaultV2APIURL && res.apiKey == "" {
		return nil, errors.New("API Key missing for Decision API V2")
	}

	if res.timeout == 0 {
		res.timeout = defaultTimeout
	}

	res.httpClient = utils.NewHTTPClient(res.url, utils.HTTPOptions{
		Timeout: res.timeout,
		Headers: headers,
	})

	return &res, nil
}

// GetModifications gets modifications from Decision API
func (r *APIClient) GetModifications(visitorID string, anonymousID *string, context map[string]interface{}) (*model.APIClientResponse, error) {
	b, err := json.Marshal(model.APIClientRequest{
		VisitorID:   visitorID,
		AnonymousID: anonymousID,
		Context:     context,
		TriggerHit:  false,
	})

	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/%s/campaigns?exposeAllKeys=true", r.envID)
	apiLogger.Infof("Sending call decision API: %s", string(b))
	response, err := r.httpClient.Call(path, "POST", b, map[string]string{
		"x-api-key": r.apiKey,
	})

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Error when calling decision API : %v", err)
	}

	resp := &model.APIClientResponse{}
	err = json.Unmarshal(response.Body, &resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ActivateCampaign activate a campaign / variation id to the Decision API
func (r *APIClient) ActivateCampaign(request model.ActivationHit) error {
	request.EnvironmentID = r.envID

	errs := request.Validate()

	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			apiLogger.Error("Activate hit validation error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return fmt.Errorf("Invalid activation hit : %s", strings.Join(errorStrings, ", "))
	}

	b, err := json.Marshal(request)

	if err != nil {
		return err
	}
	apiLogger.Debugf("Sending activate to API: %s", string(b))
	resp, err := r.httpClient.Call("/activate", "POST", b, nil)

	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("Error when calling activation API : %v", err)
	}

	return nil
}

// SendEvent sends an event to flagship Event endpoint
func (r *APIClient) SendEvent(request model.Event) error {
	errs := request.Validate()

	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			apiLogger.Error("Send event validation error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return fmt.Errorf("Invalid send hit : %s", strings.Join(errorStrings, ", "))
	}

	b, err := json.Marshal(request)

	if err != nil {
		return err
	}

	apiLogger.Debugf("Sending event to API: %s", string(b))
	resp, err := r.httpClient.Call(fmt.Sprintf("/%s/events", r.envID), "POST", b, nil)

	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("Error when calling activation API : %v", err)
	}

	return nil
}
