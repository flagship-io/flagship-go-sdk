package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/abtasty/flagship-go-sdk/v2/pkg/logging"
)

const defaultTimeout = 10 * time.Second

var httpLogger = logging.CreateLogger("HTTP Request")

// HTTPClient represents the HTTPClient infos
type HTTPClient struct {
	baseURL     string
	baseHeaders map[string]string
	client      *http.Client
	retries     int
}

// HTTPResponse represents the HTTPResponse infos
type HTTPResponse struct {
	Body       []byte
	Headers    http.Header
	StatusCode int
}

// HTTPOptions represents the options for the HTTPRequest object
type HTTPOptions struct {
	Retries int
	Timeout time.Duration
	Headers map[string]string
}

// NewHTTPClient creates an HTTP requester object
func NewHTTPClient(baseURL string, options HTTPOptions) *HTTPClient {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	retries := 1
	timeout := defaultTimeout

	if options.Retries > 1 {
		retries = options.Retries
	}

	if options.Timeout != 0 {
		timeout = options.Timeout
	}

	if options.Headers != nil {
		for k, v := range options.Headers {
			headers[k] = v
		}
	}

	client := &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		retries:     retries,
		baseURL:     baseURL,
		baseHeaders: headers,
	}

	return client
}

// Call executes request with retries and returns response body, headers, status code and error
func (r *HTTPClient) Call(path, method string, body []byte, headers map[string]string) (*HTTPResponse, error) {
	url := fmt.Sprintf("%s%s", r.baseURL, path)
	httpLogger.Debugf("Requesting %s", url)

	var resp *http.Response
	var req *http.Request
	var err error

	for i := r.retries; i >= 0; i-- {
		reader := bytes.NewBuffer(body)
		req, err = http.NewRequest(method, url, reader)
		if err != nil {
			httpLogger.Error(fmt.Sprintf("failed to create new http request %s", url), err)
			return nil, err
		}

		for k, v := range r.baseHeaders {
			req.Header.Add(k, v)
		}

		for k, v := range headers {
			req.Header.Add(k, v)
		}

		resp, err = r.client.Do(req)

		if resp != nil && resp.StatusCode < http.StatusBadRequest {
			break
		}
	}

	if err != nil {
		httpLogger.Error("Error on HTTP request: ", err)
		return nil, err
	}

	defer func() {
		if e := resp.Body.Close(); e != nil {
			httpLogger.Warning("Error when closing response body: ", e)
		}
	}()

	code := resp.StatusCode
	var response []byte

	if resp.StatusCode >= http.StatusBadRequest {
		httpLogger.Warning("HTTP Error status code: ", resp.StatusCode)
		return &HTTPResponse{response, resp.Header, resp.StatusCode}, fmt.Errorf("Request error message: %v", resp.Status)
	}

	if response, err = ioutil.ReadAll(resp.Body); err != nil {
		httpLogger.Error("Error when reading body: ", err)
		return &HTTPResponse{nil, resp.Header, resp.StatusCode}, err
	}

	return &HTTPResponse{response, resp.Header, code}, nil
}
