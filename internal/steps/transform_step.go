// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package steps

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/carloskvasir/goflow/internal/models"
	"github.com/tidwall/gjson"
)

// TransformStep handles data transformation using templates and JSON path mapping
type TransformStep struct {
	config models.StepConfig
}

// NewTransformStep creates a new transform step
func NewTransformStep(config models.StepConfig) *TransformStep {
	return &TransformStep{
		config: config,
	}
}

// Execute processes the transformation
func (s *TransformStep) Execute(context map[string]interface{}) (*models.StepResult, error) {
	// Get template and mapping from config
	templateStr := s.config["template"].(string)
	mapping := s.config["mapping"].(map[string]interface{})

	// Create data map for template
	data := make(map[string]interface{})
	
	// Process each mapping
	for key, path := range mapping {
		jsonPath := path.(string)
		// Convert context to JSON to use gjson
		contextJSON, err := json.Marshal(context)
		if err != nil {
			return nil, err
		}
		value := gjson.Get(string(contextJSON), jsonPath)
		data[key] = value.Value()
	}

	// Process template
	tmpl, err := template.New("message").Parse(templateStr)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return &models.StepResult{
		Status: models.StatusCompleted,
		Data:   buf.String(),
	}, nil
}
