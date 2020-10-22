package tracking

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/abtasty/flagship-go-sdk/v2/pkg/model"
)

// MockAPIClient represents a fake API client informations
type MockAPIClient struct {
	envID      string
	shouldFail bool
}

// NewMockAPIClient creates a mock API client that returns success or fail status
func NewMockAPIClient(envID string, shouldFail bool) *MockAPIClient {
	res := MockAPIClient{
		shouldFail: shouldFail,
	}

	return &res
}

// SendHit sends a tracking hit to the Data Collect API
func (r MockAPIClient) SendHit(hit model.HitInterface) error {
	errs := hit.Validate()
	if len(errs) > 0 {
		errorStrings := []string{}
		for _, e := range errs {
			apiLogger.Error("Hit validation error", e)
			errorStrings = append(errorStrings, e.Error())
		}
		return fmt.Errorf("Invalid hit : %s", strings.Join(errorStrings, ", "))
	}
	hit.ComputeQueueTime()

	json, err := json.Marshal(hit)

	log.Printf("Sending hit : %v", string(json))

	if r.shouldFail {
		return errors.New("Mock fail send hit error")
	}

	return err
}
