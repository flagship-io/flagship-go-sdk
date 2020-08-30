package bucketing

// APIClientMock represents the API client mock informations
type APIClientMock struct {
	envID        string
	responseMock *Configuration
	statusCode   int
}

// NewAPIClientMock creates a fake api client that returns a specific response
func NewAPIClientMock(envID string, responseMock *Configuration, statusCode int) *APIClientMock {
	res := APIClientMock{
		envID:        envID,
		responseMock: responseMock,
		statusCode:   statusCode,
	}

	return &res
}

// GetConfiguration mocks a configuration
func (r *APIClientMock) GetConfiguration() (*Configuration, error) {
	return r.responseMock, nil
}
