package tracking

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/flagship-io/flagship-go-sdk/v2/pkg/model"
)

// MockAPIClient represents a fake API client informations
type MockAPIClient struct {
	envID         string
	shouldFail    bool
	requestString string
}

// NewMockAPIClient creates a mock API client that returns success or fail status
func NewMockAPIClient(envID string, shouldFail bool) *MockAPIClient {
	res := MockAPIClient{
		shouldFail: shouldFail,
		envID:      envID,
	}

	return &res
}

// SendHit sends a tracking hit to the Data Collect API
func (r *MockAPIClient) SendHit(visitorID string, anonymousID *string, hit model.HitInterface) error {
	if hit == nil {
		err := errors.New("Hit should not be empty")
		apiLogger.Error(err.Error(), err)
		return err
	}

	hit.SetBaseInfos(r.envID, visitorID, anonymousID)

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

	r.requestString = string(b)
	log.Printf("Sending hit : %v", string(b))

	if r.shouldFail {
		return errors.New("Mock fail send hit error")
	}

	return err
}
