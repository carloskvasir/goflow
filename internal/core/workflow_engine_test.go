package core

import (
	"context"
	"testing"
	"time"

	"github.com/carloskvasir/goflow/internal/models"
)

func TestWorkflowEngine(t *testing.T) {
	engine := NewWorkflowEngine()

	// Criar um workflow de teste
	workflow := &models.Workflow{
		ID:          "test-workflow",
		Name:        "Test Workflow",
		Description: "A test workflow",
		Steps: []models.Step{
			{
				ID:   "step1",
				Name: "Echo Step 1",
				Type: "echo",
				Config: map[string]interface{}{
					"message": "Hello from Step 1",
				},
			},
			{
				ID:   "step2",
				Name: "Echo Step 2",
				Type: "echo",
				Config: map[string]interface{}{
					"message": "Hello from Step 2",
				},
				Next: []string{"step3"},
			},
			{
				ID:   "step3",
				Name: "Echo Step 3",
				Type: "echo",
				Config: map[string]interface{}{
					"message": "Hello from Step 3",
				},
			},
		},
	}

	// Testar registro do workflow
	if err := engine.RegisterWorkflow(workflow); err != nil {
		t.Errorf("Erro ao registrar workflow: %v", err)
	}

	// Testar registro duplicado
	if err := engine.RegisterWorkflow(workflow); err == nil {
		t.Error("Esperava erro ao registrar workflow duplicado")
	}

	// Testar execução do workflow
	ctx := context.Background()
	result, err := engine.ExecuteWorkflow(ctx, workflow.ID)
	if err != nil {
		t.Errorf("Erro ao executar workflow: %v", err)
	}

	// Verificar resultado
	if result.Status != models.StatusCompleted {
		t.Errorf("Status esperado %s, mas obteve %s", models.StatusCompleted, result.Status)
	}

	if len(result.StepResults) != 3 {
		t.Errorf("Esperava 3 resultados de steps, mas obteve %d", len(result.StepResults))
	}

	// Testar workflow não existente
	_, err = engine.ExecuteWorkflow(ctx, "non-existent")
	if err == nil {
		t.Error("Esperava erro ao executar workflow não existente")
	}
}

func TestWorkflowEngineWithRetry(t *testing.T) {
	engine := NewWorkflowEngine()

	// Criar um workflow com retry
	workflow := &models.Workflow{
		ID:   "retry-workflow",
		Name: "Retry Workflow",
		Steps: []models.Step{
			{
				ID:   "retry-step",
				Name: "Retry Step",
				Type: "echo",
				Config: map[string]interface{}{
					"message": "Hello with retry",
				},
				Retry: &models.RetryConfig{
					MaxAttempts: 3,
					Delay:      time.Millisecond * 100,
					MaxDelay:   time.Second,
					Multiplier: 2.0,
				},
			},
		},
	}

	// Registrar e executar
	if err := engine.RegisterWorkflow(workflow); err != nil {
		t.Errorf("Erro ao registrar workflow: %v", err)
	}

	ctx := context.Background()
	result, err := engine.ExecuteWorkflow(ctx, workflow.ID)
	if err != nil {
		t.Errorf("Erro ao executar workflow: %v", err)
	}

	// Verificar resultado
	if result.Status != models.StatusCompleted {
		t.Errorf("Status esperado %s, mas obteve %s", models.StatusCompleted, result.Status)
	}

	stepResult := result.StepResults["retry-step"]
	if stepResult.Attempts != 1 {
		t.Errorf("Esperava 1 tentativa, mas obteve %d", stepResult.Attempts)
	}
}
