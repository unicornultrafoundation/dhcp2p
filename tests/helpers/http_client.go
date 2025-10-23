package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPClient provides utilities for HTTP testing
type HTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Headers    map[string]string
}

// NewHTTPClient creates a new HTTP client for testing
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Headers: make(map[string]string),
	}
}

// SetHeader sets a header for all requests
func (c *HTTPClient) SetHeader(key, value string) {
	c.Headers[key] = value
}

// Get performs a GET request
func (c *HTTPClient) Get(path string) (*http.Response, error) {
	return c.doRequest("GET", path, nil)
}

// Post performs a POST request with JSON body
func (c *HTTPClient) Post(path string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	return c.doRequest("POST", path, bodyBytes)
}

// Put performs a PUT request with JSON body
func (c *HTTPClient) Put(path string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	return c.doRequest("PUT", path, bodyBytes)
}

// Delete performs a DELETE request
func (c *HTTPClient) Delete(path string) (*http.Response, error) {
	return c.doRequest("DELETE", path, nil)
}

// doRequest performs the actual HTTP request
func (c *HTTPClient) doRequest(method, path string, body []byte) (*http.Response, error) {
	url := c.BaseURL + path

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}

	// Set custom headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	return resp, nil
}

// ParseJSONResponse parses a JSON response into the target struct
func ParseJSONResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// AssertStatusCode checks if the response has the expected status code
func AssertStatusCode(resp *http.Response, expected int) error {
	if resp.StatusCode != expected {
		return fmt.Errorf("expected status code %d, got %d", expected, resp.StatusCode)
	}
	return nil
}
