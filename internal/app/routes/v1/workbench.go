package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	webhandlers "github.com/aosanya/CodeValdCortex/internal/web/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterWorkbenchRoutes registers workbench and issue management endpoints
func RegisterWorkbenchRoutes(rg *gin.RouterGroup, issueService *services.IssueService, workbenchService *services.WorkbenchService, instanceService services.InstanceService, agencyService agency.Service, logger *logrus.Logger) {
	workbenchHandler := webhandlers.NewWorkbenchHandler(issueService, workbenchService, instanceService, agencyService, logger)

	rg.GET("/agencies/:id/instances/:instance_id/workbench", workbenchHandler.ShowWorkbench)
	rg.POST("/agencies/:id/instances/:instance_id/issues", workbenchHandler.CreateIssue)
	rg.GET("/agencies/:id/instances/:instance_id/issues", workbenchHandler.ListIssues)
	rg.GET("/agencies/:id/instances/:instance_id/issues/:issue_id", workbenchHandler.GetIssue)
	rg.PUT("/agencies/:id/instances/:instance_id/issues/:issue_id", workbenchHandler.UpdateIssue)
	rg.PATCH("/agencies/:id/instances/:instance_id/issues/:issue_id", workbenchHandler.UpdateIssue)
	rg.DELETE("/agencies/:id/instances/:instance_id/issues/:issue_id", workbenchHandler.DeleteIssue)
	rg.POST("/agencies/:id/instances/:instance_id/issues/:issue_id/assign", workbenchHandler.AssignIssue)
	rg.POST("/agencies/:id/instances/:instance_id/issues/:issue_id/claim", workbenchHandler.ClaimIssue)
	rg.POST("/agencies/:id/instances/:instance_id/issues/:issue_id/progress", workbenchHandler.ProgressIssue)

	logger.Info("Workbench and issue management endpoints registered")
}
