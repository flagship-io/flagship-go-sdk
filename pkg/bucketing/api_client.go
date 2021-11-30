package bucketing

import (
	"fmt"
	"time"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/logging"
	"github.com/flagship-io/flagship-go-sdk/v2/pkg/utils"
	bucketingProto "github.com/flagship-io/flagship-proto/bucketing"
	"github.com/golang/protobuf/jsonpb"
)

const defaultTimeout = 10 * time.Second
const defaultAPIURL = "https://cdn.flagship.io"

var apiLogger = logging.CreateLogger("Bucketing API")

// APIClient represents the API client informations
type APIClient struct {
	url         string
	envID       string
	apiKey      string
	timeout     time.Duration
	retries     int
	httpRequest utils.HTTPClientInterface
}

// APIUrl sets http client base URL
func APIUrl(url string) func(r *APIClient) {
	return func(r *APIClient) {
		r.url = url
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

// NewAPIClient creates a bucketing API Client to poll bucketing infos
func NewAPIClient(envID string, params ...func(*APIClient)) *APIClient {
	res := APIClient{
		envID:   envID,
		retries: 1,
	}

	headers := map[string]string{}

	for _, param := range params {
		param(&res)
	}

	if res.apiKey != "" {
		headers["x-api-key"] = res.apiKey
	}

	if res.url == "" {
		res.url = defaultAPIURL
	}

	if res.timeout == 0 {
		res.timeout = defaultTimeout
	}

	res.httpRequest = utils.NewHTTPClient(res.url, utils.HTTPOptions{
		Timeout: res.timeout,
		Headers: headers,
		Retries: res.retries,
	})

	return &res
}

// GetConfiguration gets an environment configuration from bucketing file
func (r *APIClient) GetConfiguration() (*bucketingProto.Bucketing_BucketingResponse, error) {
	path := fmt.Sprintf("/%s/bucketing.json", r.envID)

	apiLogger.Info("Calling bucketing file to get configuration")
	resp, err := r.httpRequest.Call(path, "GET", nil, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 304 {
		return nil, fmt.Errorf("Error when calling Bucketing API : %v", err)
	}

	conf := &bucketingProto.Bucketing_BucketingResponse{}
	err = jsonpb.UnmarshalString(string(resp.Body), conf)

	if err != nil {
		return nil, err
	}

	return conf, nil
}
