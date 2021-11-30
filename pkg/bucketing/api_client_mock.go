package bucketing

import "github.com/flagship-io/flagship-proto/bucketing"

// APIClientMock represents the API client mock informations
type APIClientMock struct {
	envID        string
	responseMock *bucketing.Bucketing_BucketingResponse
	statusCode   int
}

// NewAPIClientMock creates a fake api client that returns a specific response
func NewAPIClientMock(envID string, responseMock *bucketing.Bucketing_BucketingResponse, statusCode int) *APIClientMock {
	res := APIClientMock{
		envID:        envID,
		responseMock: responseMock,
		statusCode:   statusCode,
	}

	return &res
}

// GetConfiguration mocks a configuration
func (r *APIClientMock) GetConfiguration() (*bucketing.Bucketing_BucketingResponse, error) {
	return r.responseMock, nil
}
