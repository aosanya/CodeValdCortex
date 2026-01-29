package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/auth"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/aosanya/CodeValdCortex/internal/workflow"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterWorkflowRoutes registers workflow management endpoints
func RegisterWorkflowRoutes(rg *gin.RouterGroup, workflowService *workflow.Service, agencyService agency.Service, authMiddleware *auth.Middleware, logger *logrus.Logger) {
	workflowHandler := handlers.NewWorkflowHandler(workflowService, agencyService, logger)

	// Protected workflow routes - require authentication
	protected := rg.Group("/agencies/:id/workflows")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.POST("", workflowHandler.CreateWorkflow)
		protected.GET("", workflowHandler.GetWorkflows)
		protected.GET("/html", workflowHandler.GetWorkflowsHTML)
	}

	// Protected workflow operations - require authentication
	workflowsProtected := rg.Group("/workflows")
	workflowsProtected.Use(authMiddleware.RequireAuth())
	{
		workflowsProtected.GET("/:id", workflowHandler.GetWorkflow)
		workflowsProtected.PUT("/:id", workflowHandler.UpdateWorkflow)
		workflowsProtected.DELETE("/:id", workflowHandler.DeleteWorkflow)
		workflowsProtected.POST("/:id/duplicate", workflowHandler.DuplicateWorkflow)
		workflowsProtected.POST("/validate", workflowHandler.ValidateWorkflow)
	}

	logger.Info("Workflow endpoints registered (protected)")
}
