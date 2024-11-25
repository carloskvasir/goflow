// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/carloskvasir/goflow/internal/models"
)

// RestStep executes HTTP requests
type RestStep struct {
	config models.StepConfig
}

// NewRestStep creates a new REST step
func NewRestStep(config models.StepConfig) *RestStep {
	return &RestStep{
		config: config,
	}
}

// Execute performs the HTTP request
func (s *RestStep) Execute(context map[string]interface{}) (*models.StepResult, error) {
	// Get configuration
	method := s.config["method"].(string)
	urlStr := s.config["url"].(string)

	// Process environment variables
	urlStr = processEnvVars(urlStr)

	// Add query parameters if present
	if params, ok := s.config["params"].(map[string]interface{}); ok {
		urlStr = addQueryParams(urlStr, params)
	}

	// Prepare request body if present
	var body io.Reader
	if data, ok := s.config["body"]; ok {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	// Create request
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers if present
	if headers, ok := s.config["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var responseData interface{}
	if err := json.Unmarshal(respBody, &responseData); err != nil {
		// If not JSON, use raw response
		responseData = string(respBody)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, respBody)
	}

	return &models.StepResult{
		Status: models.StatusCompleted,
		Data:   responseData,
	}, nil
}

// Helper functions

func processEnvVars(input string) string {
	result := input
	// Find all ${VAR} patterns
	for _, match := range strings.Split(input, "${") {
		if !strings.Contains(match, "}") {
			continue
		}
		varName := strings.Split(match, "}")[0]
		if value := os.Getenv(varName); value != "" {
			result = strings.ReplaceAll(result, "${"+varName+"}", value)
		}
	}
	return result
}

func addQueryParams(baseURL string, params map[string]interface{}) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, fmt.Sprintf("%v", value))
	}
	u.RawQuery = q.Encode()

	return u.String()
}
