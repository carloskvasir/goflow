// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package steps

import (
	"github.com/carloskvasir/goflow/internal/models"
)

// EchoStep is a simple step that returns a message
type EchoStep struct {
	config models.StepConfig
}

// NewEchoStep creates a new echo step
func NewEchoStep(config models.StepConfig) *EchoStep {
	return &EchoStep{
		config: config,
	}
}

// Execute returns the message from the config
func (s *EchoStep) Execute(context map[string]interface{}) (*models.StepResult, error) {
	message := s.config["message"].(string)

	return &models.StepResult{
		Status: models.StatusCompleted,
		Data:   message,
	}, nil
}
