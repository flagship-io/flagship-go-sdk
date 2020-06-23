package bucketing

import (
	"testing"
)

var testEnvID = "test_env_id"
var realEnvID = "blvo2kijq6pg023l8edg"

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(testEnvID)

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.url != defaultAPIURL {
		t.Error("Api url should be set to default")
	}
}

func TestNewAPIClientParams(t *testing.T) {
	client := NewAPIClient(testEnvID, Timeout(10), Retries(12), APIKey("api_key"), APIUrl("http://test.com"))

	if client == nil {
		t.Error("Api client tracking should not be nil")
	}

	if client.apiKey != "api_key" {
		t.Errorf("Wrong api key. Expected %v, got %v", "api_key", client.apiKey)
	}

	if client.url != "http://test.com" {
		t.Errorf("Wrong api key. Expected %v, got %v", "http://test.com", client.url)
	}

	if client.retries != 12 {
		t.Errorf("Wrong retries. Expected %v, got %v", 12, client.retries)
	}
}

func TestGetConfiguration(t *testing.T) {
	client := NewAPIClient(testEnvID)
	_, err := client.GetConfiguration()

	if err == nil {
		t.Error("Wrong env id should return an err")
	}

	client = NewAPIClient(realEnvID)
	conf, err := client.GetConfiguration()

	if err != nil {
		t.Errorf("Correct env id should not return an err. Got %v", err)
	}

	if conf == nil {
		t.Error("Correct env id should return a conf. Got nil")
	}
}
