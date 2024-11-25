package main

import (
	"context"
	"fmt"
	"log"

	"github.com/carloskvasir/goflow/internal/core"
	"github.com/carloskvasir/goflow/internal/models"
)

func main() {
	// Criar uma nova instância do motor de workflow
	engine := core.NewWorkflowEngine()

	// Criar um novo workflow
	workflow := &models.Workflow{
		ID:          "hello-world",
		Name:        "Hello World Workflow",
		Description: "Um workflow simples de exemplo",
		Steps: []models.Step{
			{
				ID:   "step1",
				Name: "Cumprimentar",
				Type: "echo",
				Config: map[string]interface{}{
					"message": "Hello, World!",
				},
			},
			{
				ID:   "step2",
				Name: "Processar",
				Type: "echo",
				Config: map[string]interface{}{
					"message": "Processando...",
				},
				Next: []string{"step3"},
			},
			{
				ID:   "step3",
				Name: "Despedir",
				Type: "echo",
				Config: map[string]interface{}{
					"message": "Goodbye, World!",
				},
			},
		},
	}

	// Registrar o workflow
	if err := engine.RegisterWorkflow(workflow); err != nil {
		log.Fatalf("Erro ao registrar workflow: %v", err)
	}

	// Executar o workflow
	ctx := context.Background()
	result, err := engine.ExecuteWorkflow(ctx, workflow.ID)
	if err != nil {
		log.Fatalf("Erro ao executar workflow: %v", err)
	}

	fmt.Printf("\nWorkflow concluído com status: %s\n", result.Status)
}
