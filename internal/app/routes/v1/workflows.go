package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/aosanya/CodeValdCortex/internal/workflow"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterWorkflowRoutes registers workflow management endpoints
func RegisterWorkflowRoutes(rg *gin.RouterGroup, workflowService *workflow.Service, agencyService agency.Service, logger *logrus.Logger) {
	workflowHandler := handlers.NewWorkflowHandler(workflowService, agencyService, logger)

	rg.POST("/agencies/:id/workflows", workflowHandler.CreateWorkflow)
	rg.GET("/agencies/:id/workflows", workflowHandler.GetWorkflows)
	rg.GET("/agencies/:id/workflows/html", workflowHandler.GetWorkflowsHTML)
	rg.GET("/workflows/:id", workflowHandler.GetWorkflow)
	rg.PUT("/workflows/:id", workflowHandler.UpdateWorkflow)
	rg.DELETE("/workflows/:id", workflowHandler.DeleteWorkflow)
	rg.POST("/workflows/:id/duplicate", workflowHandler.DuplicateWorkflow)
	rg.POST("/workflows/validate", workflowHandler.ValidateWorkflow)

	logger.Info("Workflow endpoints registered")
}
