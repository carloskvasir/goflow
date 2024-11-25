// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/carloskvasir/goflow/internal/core"
	"github.com/carloskvasir/goflow/internal/models"
	"github.com/gin-gonic/gin"
)

func loadEnv(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("error setting environment variable %s: %w", key, err)
		}
	}

	return nil
}

func main() {
	// Carregar variáveis de ambiente do arquivo .env
	envFile := filepath.Join(".", ".env")
	if err := loadEnv(envFile); err != nil {
		log.Printf("Aviso: não foi possível carregar o arquivo .env: %v", err)
	}

	// Criar engine de workflows
	engine := core.NewWorkflowEngine()

	// Configurar router
	router := setupRouter(engine)

	// Obter porta do ambiente ou usar padrão
	port := os.Getenv("GOFLOW_PORT")
	if port == "" {
		port = "3000"
	}

	// Iniciar servidor
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func setupRouter(engine *core.WorkflowEngine) *gin.Engine {
	router := gin.Default()

	// Endpoints da API
	api := router.Group("/api/v1")
	{
		// Workflows
		api.GET("/workflows/:id", func(c *gin.Context) {
			workflowID := c.Param("id")

			workflow, exists := engine.GetWorkflow(workflowID)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
				return
			}

			c.JSON(http.StatusOK, workflow)
		})

		api.POST("/workflows", func(c *gin.Context) {
			var workflow models.Workflow
			if err := c.ShouldBindJSON(&workflow); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := engine.RegisterWorkflow(&workflow); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, workflow)
		})

		api.DELETE("/workflows/:id", func(c *gin.Context) {
			workflowID := c.Param("id")
			if err := engine.DeleteWorkflow(workflowID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		api.POST("/workflows/:id/execute", func(c *gin.Context) {
			workflowID := c.Param("id")

			result, err := engine.ExecuteWorkflow(c.Request.Context(), workflowID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, result)
		})
	}

	return router
}
