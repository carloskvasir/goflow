// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package models

import (
	"time"
)

// WorkflowStatus represents the current state of a workflow
type WorkflowStatus string

const (
	StatusPending   WorkflowStatus = "pending"
	StatusRunning   WorkflowStatus = "running"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
)

// Workflow represents a complete integration flow
type Workflow struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Steps       []Step        `json:"steps"`
	Status      WorkflowStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Step represents an individual step in the workflow
type Step struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"` // "rest", "soap", "graphql", "transform"
	Config   StepConfig             `json:"config"`
	Next     []string              `json:"next,omitempty"`    // IDs of next steps
	OnError  []string              `json:"on_error,omitempty"` // IDs of steps to execute on error
	Retry    *RetryConfig          `json:"retry,omitempty"`
	Timeout  time.Duration         `json:"timeout,omitempty"`
	Required bool                  `json:"required"` // If true, step failure fails the entire workflow
}

// StepConfig represents the configuration for a step
type StepConfig map[string]interface{}

// RetryConfig configures retry attempts for a step
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts"`
	Delay       time.Duration `json:"delay"`
	MaxDelay    time.Duration `json:"max_delay"`
	Multiplier  float64       `json:"multiplier"`
}

// WorkflowResult represents the result of a workflow execution
type WorkflowResult struct {
	WorkflowID  string                 `json:"workflow_id"`
	Status      WorkflowStatus         `json:"status"`
	StepResults map[string]StepResult  `json:"step_results"`
	StartTime   time.Time             `json:"start_time"`
	EndTime     time.Time             `json:"end_time"`
	Error       string                `json:"error,omitempty"`
}

// StepResult represents the result of a step execution
type StepResult struct {
	StepID      string                 `json:"step_id"`
	Status      WorkflowStatus         `json:"status"`
	StartTime   time.Time             `json:"start_time"`
	EndTime     time.Time             `json:"end_time"`
	Data        interface{}           `json:"data,omitempty"`
	Error       string                `json:"error,omitempty"`
	Attempts    int                   `json:"attempts"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
