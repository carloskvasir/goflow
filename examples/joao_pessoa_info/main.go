package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/carloskvasir/goflow/internal/core"
	"github.com/carloskvasir/goflow/internal/models"
)

func main() {
	// Verificar API key
	if os.Getenv("OPENWEATHER_API_KEY") == "" {
		log.Fatal("OPENWEATHER_API_KEY environment variable is required")
	}

	// Criar engine
	engine := core.NewWorkflowEngine()

	// Carregar workflow do arquivo
	workflowData, err := os.ReadFile("workflow.json")
	if err != nil {
		log.Fatal(err)
	}

	var workflow models.Workflow
	if err := json.Unmarshal(workflowData, &workflow); err != nil {
		log.Fatal(err)
	}

	// Registrar workflow
	if err := engine.RegisterWorkflow(&workflow); err != nil {
		log.Fatal(err)
	}

	// Executar workflow
	result, err := engine.ExecuteWorkflow(context.Background(), workflow.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Obter resultado formatado
	if message, ok := result.StepResults["format-message"].Data.(string); ok {
		fmt.Println(message)
	} else {
		log.Fatal("Failed to get formatted message from workflow result")
	}
}
