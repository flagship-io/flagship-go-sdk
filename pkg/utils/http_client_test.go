package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestServer struct {
	nbCalls *int
	server  *httptest.Server
}

func createTestServer(callsBeforeOK int) TestServer {
	nbCalls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/ok-endpoint" {
			w.Write([]byte("ok"))
		}
		if r.URL.String() == "/ok-headers" {
			data, _ := json.Marshal(r.Header)
			w.Write([]byte(data))
		}
		if r.URL.String() == "/retry" {
			nbCalls++
			if nbCalls >= callsBeforeOK {
				w.Write([]byte("ok"))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
		if r.URL.String() == "/error-endpoint" {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	return TestServer{
		nbCalls: &nbCalls,
		server:  ts,
	}
}

func TestNewHTTPClient(t *testing.T) {
	url := "http://google.fr"
	r := NewHTTPClient(url, HTTPOptions{})

	if r == nil {
		t.Error("HTTP Request should not be empty")
	}

	if r.baseURL != url {
		t.Errorf("HTTP Request url incorrect. Should be default %v, got %v", url, r.baseURL)
	}

	if r.retries != 1 {
		t.Errorf("HTTP Request retries incorrect. Should be default 1, got %v", r.retries)
	}

	if r.client.Timeout != defaultTimeout {
		t.Errorf("HTTP Request timeout incorrect. Should be default %v, got %v", defaultTimeout, r.client.Timeout)
	}

	if len(r.baseHeaders) != 2 {
		t.Errorf("HTTP Request headers incorrect. Should be default %v, got %v", 2, len(r.baseHeaders))
	}
}

func TestNewHTTPClientOptions(t *testing.T) {
	url := "http://google.fr"
	r := NewHTTPClient(url, HTTPOptions{
		Retries: 2,
		Timeout: 10,
		Headers: map[string]string{
			"test": "value",
		},
	})

	if r.retries != 2 {
		t.Errorf("HTTP Request retries incorrect. Should be 2, got %v", r.retries)
	}

	if r.client.Timeout != 10 {
		t.Errorf("HTTP Request timeout incorrect. Should be default %v, got %v", 10, r.client.Timeout)
	}

	if len(r.baseHeaders) != 3 {
		t.Errorf("HTTP Request headers incorrect. Should be default %v, got %v", 3, len(r.baseHeaders))
	}
}

func TestCall(t *testing.T) {
	ts := createTestServer(0)
	defer ts.server.Close()

	httpreq := NewHTTPClient(ts.server.URL, HTTPOptions{})
	resp, _ := httpreq.Call("/ok-endpoint", "GET", nil, nil)
	if string(resp.Body) != "ok" {
		t.Errorf("Expected response %v, got %v", "ok\n", resp)
	}

	resp, err := httpreq.Call("/ok-headers", "GET", nil, map[string]string{
		"Test": "value",
	})

	var testJSON map[string][]string = make(map[string][]string)
	err = json.Unmarshal(resp.Body, &testJSON)

	if err != nil {
		t.Errorf("Expected response json error nil, got %v", err)
	}

	if testJSON["Test"][0] != "value" {
		t.Errorf("Expected response %v, got %v", "value", testJSON["Test"])
	}

	resp, err = httpreq.Call("/error-endpoint", "GET", nil, nil)
	if err == nil {
		t.Error("Expected error for internal error, got nil")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestRetry(t *testing.T) {
	callsBeforeOK := 3
	ts := createTestServer(callsBeforeOK)
	defer ts.server.Close()

	httpreq := NewHTTPClient(ts.server.URL, HTTPOptions{
		Retries: 10,
	})

	resp, _ := httpreq.Call("/retry", "GET", nil, nil)
	if "ok" != string(resp.Body) {
		t.Errorf("Expected response %v, got %v", "ok", string(resp.Body))
	}
	if *ts.nbCalls != callsBeforeOK {
		t.Errorf("Expected %v http calls, got %v", 5, *ts.nbCalls)
	}

	httpreq = NewHTTPClient(ts.server.URL, HTTPOptions{
		Retries: callsBeforeOK - 2,
	})

	*ts.nbCalls = 0
	resp, err := httpreq.Call("/retry", "GET", nil, nil)
	if err == nil {
		t.Error("Expected error for error request, got nil")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, resp.StatusCode)
	}
	if *ts.nbCalls != callsBeforeOK-1 {
		t.Errorf("Expected %v http calls, got %v", callsBeforeOK-1, *ts.nbCalls)
	}
}

func TestFailCall(t *testing.T) {
	httpreq := NewHTTPClient("url_not_exists", HTTPOptions{})

	_, err := httpreq.Call("/", "GET", nil, nil)
	assert.NotNil(t, err)
}
