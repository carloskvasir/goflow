// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package connectors

import (
	"context"
	"time"
)

// Request represents a generic request for any type of API
type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	Body        interface{}
	Timeout     time.Duration
	RetryConfig *RetryConfig
}

// Response represents a generic response from any type of API
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Error      error
}

// RetryConfig configures the retry policy for requests
type RetryConfig struct {
	MaxRetries  int
	RetryDelay  time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// Config represents the base configuration for any connector
type Config struct {
	BaseURL     string
	Timeout     time.Duration
	RetryConfig RetryConfig
	Auth        AuthConfig
}

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	Type        string // "basic", "oauth2", "apikey"
	Credentials map[string]string
}

// Connector is the base interface for all API connector types
type Connector interface {
	// Connect establishes the initial connection with the API
	Connect(ctx context.Context) error
	
	// Execute performs a request to the API
	Execute(ctx context.Context, req Request) (*Response, error)
	
	// Close closes the connection and releases resources
	Close() error
}
