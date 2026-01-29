package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/auth"
	webhandlers "github.com/aosanya/CodeValdCortex/internal/web/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterWorkbenchRoutes registers workbench and issue management endpoints
func RegisterWorkbenchRoutes(rg *gin.RouterGroup, issueService *services.IssueService, workbenchService *services.WorkbenchService, instanceService services.InstanceService, agencyService agency.Service, authMiddleware *auth.Middleware, logger *logrus.Logger) {
	workbenchHandler := webhandlers.NewWorkbenchHandler(issueService, workbenchService, instanceService, agencyService, logger)

	// Protected workbench routes - require authentication
	protected := rg.Group("/agencies/:id/instances/:instance_id")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.GET("/workbench", workbenchHandler.ShowWorkbench)
		protected.POST("/issues", workbenchHandler.CreateIssue)
		protected.GET("/issues", workbenchHandler.ListIssues)
		protected.GET("/issues/:issue_id", workbenchHandler.GetIssue)
		protected.PUT("/issues/:issue_id", workbenchHandler.UpdateIssue)
		protected.PATCH("/issues/:issue_id", workbenchHandler.UpdateIssue)
		protected.DELETE("/issues/:issue_id", workbenchHandler.DeleteIssue)
		protected.POST("/issues/:issue_id/assign", workbenchHandler.AssignIssue)
		protected.POST("/issues/:issue_id/claim", workbenchHandler.ClaimIssue)
		protected.POST("/issues/:issue_id/progress", workbenchHandler.ProgressIssue)
	}

	logger.Info("Workbench and issue management endpoints registered (protected)")
}
