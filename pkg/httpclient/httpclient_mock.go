package httpclient

import (
	"net/http"
)

// HTTPClientMock represents the HTTPClientMock infos
type HTTPClientMock struct {
	responseCode    int
	responseBody    []byte
	responseHeaders http.Header
}

// NewHTTPClientMock creates an HTTP mock object
func NewHTTPClientMock(responseCode int, responseBody []byte, responseHeaders http.Header) HTTPClientInterface {
	return &HTTPClientMock{
		responseCode,
		responseBody,
		responseHeaders,
	}
}

// Call executes request with retries and returns response body, headers, status code and error
func (r *HTTPClientMock) Call(path, method string, body []byte, headers map[string]string) (*HTTPResponse, error) {
	return &HTTPResponse{r.responseBody, r.responseHeaders, r.responseCode}, nil
}
