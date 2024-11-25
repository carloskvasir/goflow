// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/carloskvasir/goflow/internal/models"
	"github.com/carloskvasir/goflow/internal/steps"
)

// WorkflowEngine is responsible for executing workflows and managing their lifecycle.
type WorkflowEngine struct {
	workflows map[string]*models.Workflow
	results   map[string]*models.WorkflowResult
	mu        sync.RWMutex
}

// NewWorkflowEngine creates a new instance of the workflow engine.
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		workflows: make(map[string]*models.Workflow),
		results:   make(map[string]*models.WorkflowResult),
	}
}

// RegisterWorkflow registers a new workflow in the engine.
func (w *WorkflowEngine) RegisterWorkflow(workflow *models.Workflow) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if workflow.ID == "" {
		return fmt.Errorf("workflow ID cannot be empty")
	}

	if _, exists := w.workflows[workflow.ID]; exists {
		return fmt.Errorf("workflow with ID %s already exists", workflow.ID)
	}

	workflow.Status = models.StatusPending
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = workflow.CreatedAt

	w.workflows[workflow.ID] = workflow
	return nil
}

// ExecuteWorkflow executes a specific workflow.
func (w *WorkflowEngine) ExecuteWorkflow(ctx context.Context, workflowID string) (*models.WorkflowResult, error) {
	w.mu.Lock()
	workflow, exists := w.workflows[workflowID]
	if !exists {
		w.mu.Unlock()
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	result := &models.WorkflowResult{
		WorkflowID:  workflowID,
		Status:      models.StatusRunning,
		StepResults: make(map[string]models.StepResult),
		StartTime:   time.Now(),
	}
	w.results[workflowID] = result
	w.mu.Unlock()

	err := w.executeSteps(ctx, workflow, result)

	w.mu.Lock()
	result.EndTime = time.Now()
	if err != nil {
		result.Status = models.StatusFailed
		result.Error = err.Error()
	} else {
		result.Status = models.StatusCompleted
	}
	w.mu.Unlock()

	return result, err
}

// GetWorkflow retorna um workflow pelo seu ID
func (w *WorkflowEngine) GetWorkflow(id string) (*models.Workflow, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	workflow, exists := w.workflows[id]
	return workflow, exists
}

// DeleteWorkflow removes a workflow from the engine
func (w *WorkflowEngine) DeleteWorkflow(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.workflows[id]; !exists {
		return fmt.Errorf("workflow %s not found", id)
	}

	delete(w.workflows, id)
	delete(w.results, id)
	return nil
}

// executeSteps executes the workflow steps.
func (w *WorkflowEngine) executeSteps(ctx context.Context, workflow *models.Workflow, result *models.WorkflowResult) error {
	completed := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(workflow.Steps))

	// Função para executar um step e seus dependentes
	var executeStepAndDependents func(step models.Step)
	executeStepAndDependents = func(step models.Step) {
		defer wg.Done()

		// Verifica se o step já foi executado
		mu.Lock()
		if completed[step.ID] {
			mu.Unlock()
			return
		}
		mu.Unlock()

		// Verifica se todas as dependências foram completadas
		if !w.canExecuteStep(&step, completed, workflow) {
			wg.Add(1)
			go executeStepAndDependents(step)
			return
		}

		// Executa o step
		if err := w.executeStep(ctx, step, workflow, result, completed); err != nil {
			errChan <- fmt.Errorf("error in step %s: %w", step.ID, err)
			return
		}

		// Executa os próximos steps
		for _, nextStepID := range step.Next {
			nextStep := w.findStep(workflow, nextStepID)
			if nextStep != nil {
				wg.Add(1)
				go executeStepAndDependents(*nextStep)
			}
		}
	}

	// Inicia a execução pelos steps iniciais
	initialSteps := w.findInitialSteps(workflow)
	for _, step := range initialSteps {
		wg.Add(1)
		go executeStepAndDependents(step)
	}

	wg.Wait()
	close(errChan)

	// Retorna o primeiro erro encontrado
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// executeStep executes a single step.
func (w *WorkflowEngine) executeStep(ctx context.Context, step models.Step, workflow *models.Workflow, result *models.WorkflowResult, completed map[string]bool) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	stepResult := models.StepResult{
		StepID:    step.ID,
		Status:    models.StatusRunning,
		StartTime: time.Now(),
		Attempts:  0,
	}

	var err error
	if step.Retry != nil {
		err = w.executeWithRetry(ctx, step, &stepResult)
	} else {
		err = w.executeSingleStep(step, &stepResult)
	}

	stepResult.EndTime = time.Now()
	if err != nil {
		stepResult.Status = models.StatusFailed
		stepResult.Error = err.Error()
		if step.Required {
			return err
		}
	} else {
		stepResult.Status = models.StatusCompleted
	}

	w.mu.Lock()
	result.StepResults[step.ID] = stepResult
	completed[step.ID] = true
	w.mu.Unlock()

	if err == nil {
		for _, nextStepID := range step.Next {
			nextStep := w.findStep(workflow, nextStepID)
			if nextStep != nil && w.canExecuteStep(nextStep, completed, workflow) {
				if err := w.executeStep(ctx, *nextStep, workflow, result, completed); err != nil {
					return err
				}
			}
		}
	} else if len(step.OnError) > 0 {
		for _, errorStepID := range step.OnError {
			errorStep := w.findStep(workflow, errorStepID)
			if errorStep != nil {
				if err := w.executeStep(ctx, *errorStep, workflow, result, completed); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// canExecuteStep checks if a step can be executed based on completed dependencies.
func (w *WorkflowEngine) canExecuteStep(step *models.Step, completed map[string]bool, workflow *models.Workflow) bool {
	for _, s := range workflow.Steps {
		for _, next := range s.Next {
			if next == step.ID && !completed[s.ID] {
				return false
			}
		}
	}
	return true
}

// executeWithRetry executes a step with retry logic.
func (w *WorkflowEngine) executeWithRetry(ctx context.Context, step models.Step, result *models.StepResult) error {
	var lastErr error
	delay := step.Retry.Delay

	maxAttempts := step.Retry.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		result.Attempts = attempt + 1
		
		if err := w.executeSingleStep(step, result); err != nil {
			lastErr = err
			if attempt < maxAttempts-1 { // Só espera se houver mais tentativas
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
					if step.Retry.Multiplier > 0 {
						delay = time.Duration(float64(delay) * step.Retry.Multiplier)
						if step.Retry.MaxDelay > 0 && delay > step.Retry.MaxDelay {
							delay = step.Retry.MaxDelay
						}
					}
					continue
				}
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("max retry attempts reached: %v", lastErr)
}

// executeSingleStep executes a single step without retry.
func (w *WorkflowEngine) executeSingleStep(step models.Step, result *models.StepResult) error {
	var stepExecutor interface{ Execute(map[string]interface{}) (*models.StepResult, error) }

	switch step.Type {
	case "rest":
		stepExecutor = steps.NewRestStep(step.Config)
	case "transform":
		stepExecutor = steps.NewTransformStep(step.Config)
	case "echo":
		stepExecutor = steps.NewEchoStep(step.Config)
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}

	stepResult, err := stepExecutor.Execute(nil) // TODO: Passar contexto adequado
	if err != nil {
		return err
	}

	result.Data = stepResult.Data
	return nil
}

// findInitialSteps finds steps that are not referenced as "next" in any other step.
func (w *WorkflowEngine) findInitialSteps(workflow *models.Workflow) []models.Step {
	nextSteps := make(map[string]bool)
	for _, step := range workflow.Steps {
		for _, next := range step.Next {
			nextSteps[next] = true
		}
	}

	var initialSteps []models.Step
	for _, step := range workflow.Steps {
		if !nextSteps[step.ID] {
			initialSteps = append(initialSteps, step)
		}
	}
	return initialSteps
}

// findStep finds a step by its ID.
func (w *WorkflowEngine) findStep(workflow *models.Workflow, stepID string) *models.Step {
	for i := range workflow.Steps {
		if workflow.Steps[i].ID == stepID {
			return &workflow.Steps[i]
		}
	}
	return nil
}
