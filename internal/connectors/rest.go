// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package connectors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RestConnector implements the Connector interface for REST APIs
type RestConnector struct {
	client  *http.Client
	config  Config
	headers map[string]string
}

// NewRestConnector creates a new instance of RestConnector
func NewRestConnector(config Config) *RestConnector {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &RestConnector{
		client:  client,
		config:  config,
		headers: make(map[string]string),
	}
}

// Connect implements the Connect method of the Connector interface
func (r *RestConnector) Connect(ctx context.Context) error {
	// Set default headers
	r.headers["Content-Type"] = "application/json"
	
	// Configure authentication
	if err := r.setupAuth(); err != nil {
		return fmt.Errorf("failed to setup authentication: %w", err)
	}
	
	return nil
}

// Execute implements the Execute method of the Connector interface
func (r *RestConnector) Execute(ctx context.Context, req Request) (*Response, error) {
	url := r.buildURL(req.URL)
	
	// Prepare request body
	var bodyReader io.Reader
	if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("error serializing body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Add headers
	for k, v := range r.headers {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	
	// Execute request with retry if configured
	var resp *http.Response
	if req.RetryConfig != nil {
		resp, err = r.executeWithRetry(httpReq, *req.RetryConfig)
	} else {
		resp, err = r.client.Do(httpReq)
	}
	
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	
	// Convert headers
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	
	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:      body,
	}, nil
}

// Close implements the Close method of the Connector interface
func (r *RestConnector) Close() error {
	r.client.CloseIdleConnections()
	return nil
}

// Helper methods

// buildURL constructs the full URL for the request
func (r *RestConnector) buildURL(path string) string {
	return fmt.Sprintf("%s%s", r.config.BaseURL, path)
}

// setupAuth configures authentication based on the config
func (r *RestConnector) setupAuth() error {
	switch r.config.Auth.Type {
	case "basic":
		username := r.config.Auth.Credentials["username"]
		password := r.config.Auth.Credentials["password"]
		if username != "" && password != "" {
			r.headers["Authorization"] = fmt.Sprintf("Basic %s:%s", username, password)
		}
	case "bearer":
		token := r.config.Auth.Credentials["token"]
		if token != "" {
			r.headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
		}
	case "apikey":
		key := r.config.Auth.Credentials["key"]
		keyName := r.config.Auth.Credentials["key_name"]
		if key != "" {
			if keyName == "" {
				keyName = "X-API-Key"
			}
			r.headers[keyName] = key
		}
	}
	return nil
}

// executeWithRetry executes a request with retry logic
func (r *RestConnector) executeWithRetry(req *http.Request, retryConfig RetryConfig) (*http.Response, error) {
	var lastErr error
	delay := retryConfig.RetryDelay
	
	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		resp, err := r.client.Do(req)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		
		if attempt == retryConfig.MaxRetries {
			break
		}
		
		// Calculate next delay
		if retryConfig.Multiplier > 0 {
			delay = time.Duration(float64(delay) * retryConfig.Multiplier)
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}
		}
		
		// Wait before next attempt
		time.Sleep(delay)
	}
	
	return nil, fmt.Errorf("maximum retry attempts reached: %w", lastErr)
}
