package app

import (
	"net/http"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	webhandlers "github.com/aosanya/CodeValdCortex/internal/web/handlers"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers/ai_refine"
	"github.com/gin-gonic/gin"
)

// registerAPIRoutes sets up all API routes
func (a *App) registerAPIRoutes(router *gin.Engine, aiRefineHandler interface{}) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "dev",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Agency endpoints
		agencyHandler := handlers.NewAgencyHandler(a.agencyService, a.logger)
		v1.GET("/agencies", agencyHandler.ListAgencies)
		v1.GET("/agencies/:id", agencyHandler.GetAgency)
		v1.POST("/agencies", agencyHandler.CreateAgency)
		v1.PUT("/agencies/:id", agencyHandler.UpdateAgency)
		v1.DELETE("/agencies/:id", agencyHandler.DeleteAgency)
		v1.GET("/agencies/active", agencyHandler.GetActiveAgency)
		v1.GET("/agencies/:id/statistics", agencyHandler.GetAgencyStatistics)

		// Unified Specification endpoints (replaces separate overview/goals/work-items)
		v1.GET("/agencies/:id/specification", agencyHandler.GetSpecification)
		v1.PUT("/agencies/:id/specification", agencyHandler.UpdateSpecification)
		v1.PUT("/agencies/:id/specification/introduction", agencyHandler.UpdateIntroduction)
		v1.PUT("/agencies/:id/specification/goals", agencyHandler.UpdateGoals)
		v1.PUT("/agencies/:id/specification/work-items", agencyHandler.UpdateWorkItems)
		v1.PUT("/agencies/:id/specification/workflows", agencyHandler.UpdateWorkflows)
		v1.PUT("/agencies/:id/specification/roles", agencyHandler.UpdateRoles)
		v1.PUT("/agencies/:id/specification/raci-matrix", agencyHandler.UpdateRACIMatrixSection)

		// RACI Matrix CRUD endpoints
		v1.GET("/agencies/:id/raci-matrix", agencyHandler.GetRACIMatrix)
		v1.POST("/agencies/:id/raci-matrix", agencyHandler.SaveRACIMatrix)

		// Roles endpoints
		v1.GET("/agencies/:id/roles", agencyHandler.GetAgencyRoles)
		v1.GET("/agencies/:id/roles/html", agencyHandler.GetAgencyRolesHTML)
		v1.POST("/agencies/:id/roles", agencyHandler.CreateAgencyRole)
		v1.GET("/agencies/:id/roles/:key", agencyHandler.GetAgencyRole)
		v1.PUT("/agencies/:id/roles/:key", agencyHandler.UpdateAgencyRole)
		v1.DELETE("/agencies/:id/roles/:key", agencyHandler.DeleteAgencyRole)

		// Tag endpoints (if tag service is available)
		if a.tagService != nil {
			tagHandler := handlers.NewTagHandler(*a.tagService, a.logger)
			v1.POST("/agencies/:id/tags", tagHandler.CreateTag)
			v1.GET("/agencies/:id/tags", tagHandler.ListTags)
			v1.GET("/agencies/:id/tags/:name", tagHandler.GetTag)
			v1.DELETE("/agencies/:id/tags/:name", tagHandler.DeleteTag)
			v1.POST("/agencies/:id/tags/:name/restore", tagHandler.RestoreFromTag)
			v1.GET("/tags/:tag1/compare/:tag2", tagHandler.CompareTags)
			a.logger.Info("Tag endpoints registered")
		}

		// Instance endpoints (MVP-PUB-007)
		if a.instanceService != nil {
			instanceHandler := handlers.NewInstanceHandler(a.instanceService, a.logger)
			v1.POST("/agencies/:id/tags/:name/instances", instanceHandler.StartInstance)
			v1.GET("/agencies/:id/instances", instanceHandler.ListInstances)
			v1.GET("/agencies/:id/instances/:instance_id", instanceHandler.GetInstance)
			v1.DELETE("/agencies/:id/instances/:instance_id", instanceHandler.DeleteInstance)
			v1.POST("/agencies/:id/instances/:instance_id/stop", instanceHandler.StopInstance)
			v1.POST("/agencies/:id/instances/:instance_id/restart", instanceHandler.RestartInstance)
			v1.GET("/agencies/:id/instances/:instance_id/health", instanceHandler.GetInstanceHealth)
			v1.GET("/agencies/:id/instances/:instance_id/agents", instanceHandler.GetInstanceAgents)
			v1.POST("/agencies/:id/instances/:instance_id/accept-job", instanceHandler.AcceptJob)
			v1.GET("/agencies/:id/tags/:name/instances", instanceHandler.ListInstancesByTag)
			a.logger.Info("Instance endpoints registered")
		}

		// Workbench and issue management endpoints (MVP-WI-008)
		if a.workbenchService != nil && a.issueService != nil && a.instanceService != nil {
			workbenchHandler := webhandlers.NewWorkbenchHandler(a.issueService, a.workbenchService, a.instanceService, a.agencyService, a.logger)
			v1.GET("/agencies/:id/instances/:instance_id/workbench", workbenchHandler.ShowWorkbench)
			v1.POST("/agencies/:id/instances/:instance_id/issues", workbenchHandler.CreateIssue)
			v1.GET("/agencies/:id/instances/:instance_id/issues", workbenchHandler.ListIssues)
			v1.GET("/agencies/:id/instances/:instance_id/issues/:issue_id", workbenchHandler.GetIssue)
			v1.PUT("/agencies/:id/instances/:instance_id/issues/:issue_id", workbenchHandler.UpdateIssue)
			v1.PATCH("/agencies/:id/instances/:instance_id/issues/:issue_id", workbenchHandler.UpdateIssue)
			v1.DELETE("/agencies/:id/instances/:instance_id/issues/:issue_id", workbenchHandler.DeleteIssue)
			v1.POST("/agencies/:id/instances/:instance_id/issues/:issue_id/assign", workbenchHandler.AssignIssue)
			v1.POST("/agencies/:id/instances/:instance_id/issues/:issue_id/claim", workbenchHandler.ClaimIssue)
			v1.POST("/agencies/:id/instances/:instance_id/issues/:issue_id/progress", workbenchHandler.ProgressIssue)
			a.logger.Info("Workbench and issue management endpoints registered")
		}

		// Publication endpoints (MVP-PUB-003)
		if a.publicationService != nil {
			pubHandler := handlers.NewPublicationHandler(a.publicationService, a.logger)
			v1.POST("/agencies/:id/validate", pubHandler.ValidateForPublish)
			v1.POST("/agencies/:id/publish", pubHandler.Publish)
			v1.POST("/agencies/:id/activate", pubHandler.Activate)
			v1.POST("/agencies/:id/deactivate", pubHandler.Deactivate)
			v1.GET("/agencies/:id/publications", pubHandler.GetPublicationHistory)
			v1.POST("/publications/:id/activate", pubHandler.ActivatePublication)
			a.logger.Info("Publication endpoints registered")
		}

		// Activation/Lifecycle endpoints (MVP-PUB-004)
		if a.activationService != nil {
			activationHandler := handlers.NewActivationHandler(a.activationService, a.logger)
			v1.POST("/agencies/:id/lifecycle/pause", activationHandler.PauseAgency)
			v1.POST("/agencies/:id/lifecycle/resume", activationHandler.ResumeAgency)
			v1.POST("/agencies/:id/lifecycle/drain", activationHandler.DrainAgency)
			v1.POST("/agencies/:id/lifecycle/stop", activationHandler.StopAgency)
			a.logger.Info("Activation lifecycle endpoints registered")
		}

		// Workflow endpoints
		if a.workflowService != nil {
			workflowHandler := handlers.NewWorkflowHandler(a.workflowService, a.agencyService, a.logger)
			v1.POST("/agencies/:id/workflows", workflowHandler.CreateWorkflow)
			v1.GET("/agencies/:id/workflows", workflowHandler.GetWorkflows)
			v1.GET("/agencies/:id/workflows/html", workflowHandler.GetWorkflowsHTML)
			v1.GET("/workflows/:id", workflowHandler.GetWorkflow)
			v1.PUT("/workflows/:id", workflowHandler.UpdateWorkflow)
			v1.DELETE("/workflows/:id", workflowHandler.DeleteWorkflow)
			v1.POST("/workflows/:id/duplicate", workflowHandler.DuplicateWorkflow)
			v1.POST("/workflows/validate", workflowHandler.ValidateWorkflow)
			a.logger.Info("Workflow endpoints registered")
		}

		// AI Refine endpoints (if AI services are available)
		if aiRefineHandler != nil {
			// Type assertion to get the concrete handler
			refineHandler := aiRefineHandler.(*ai_refine.Handler)

			v1.POST("/agencies/:id/overview/refine", refineHandler.RefineIntroduction)
			if a.goalRefiner != nil {
				// Main dynamic router - handles all goal operations through natural language prompts
				v1.POST("/agencies/:id/goals/refine-dynamic", refineHandler.RefineGoals)
				// Convenience routes that use RefineGoals with preset prompts
				v1.POST("/agencies/:id/goals/:goalKey/refine", refineHandler.RefineSpecificGoal)
				v1.POST("/agencies/:id/goals/generate", refineHandler.GenerateGoalWithPrompt)
				v1.POST("/agencies/:id/goals/consolidate", refineHandler.ConsolidateGoalsWithPrompt)
			}
			if a.workItemBuilder != nil {
				// DISABLED: Work item AI handlers need refactoring for unified specification model
				// Main dynamic router - handles all work item operations through natural language prompts
				// v1.POST("/agencies/:id/work-items/refine-dynamic", refineHandler.RefineWorkItems)
				// v1.POST("/agencies/:id/work-items/refine-specific", refineHandler.RefineSpecificWorkItem)
				// v1.POST("/agencies/:id/work-items/generate", refineHandler.GenerateWorkItemWithPrompt)
				// v1.POST("/agencies/:id/work-items/consolidate", refineHandler.ConsolidateWorkItemsWithPrompt)
				// v1.POST("/agencies/:id/work-items/enhance-all", refineHandler.EnhanceAllWorkItems)
				a.logger.Warn("Work item AI refine endpoints disabled - need refactoring for unified specification")
			}
			if a.roleBuilder != nil {
				// Main dynamic router - handles all role operations through natural language prompts
				v1.POST("/agencies/:id/roles/refine-dynamic", refineHandler.RefineRoles)
				// Convenience routes that use RefineRoles with preset prompts
				v1.POST("/agencies/:id/roles/refine-specific", refineHandler.RefineSpecificRole)
				v1.POST("/agencies/:id/roles/generate", refineHandler.GenerateRoleWithPrompt)
				v1.POST("/agencies/:id/roles/consolidate", refineHandler.ConsolidateRolesWithPrompt)
				v1.POST("/agencies/:id/roles/enhance-all", refineHandler.EnhanceAllRolesWithPrompt)
			}
			if a.raciBuilder != nil {
				// Main dynamic router - handles all RACI operations through natural language prompts
				v1.POST("/agencies/:id/raci-matrix/refine-dynamic", refineHandler.RefineRACIMappings)
				// Convenience routes that use RefineRACIMappings with preset prompts
				v1.POST("/agencies/:id/raci-matrix/refine-specific", refineHandler.RefineSpecificRACIMapping)
				v1.POST("/agencies/:id/raci-matrix/generate", refineHandler.GenerateRACIMappingWithPrompt)
				v1.POST("/agencies/:id/raci-matrix/consolidate", refineHandler.ConsolidateRACIMappingsWithPrompt)
				v1.POST("/agencies/:id/raci-matrix/create-complete", refineHandler.CreateCompleteRACIMatrixWithPrompt)
			}
			if a.workflowBuilder != nil {
				// DISABLED: Workflow handler needs refactoring for unified specification model
				// v1.POST("/agencies/:id/workflows/refine-dynamic", refineHandler.RefineWorkflows)
				a.logger.Warn("Workflow AI refine endpoint disabled - needs refactoring for unified specification")
			}
			a.logger.Info("AI Refine endpoints registered")
		}

		// AI Agency Designer endpoints (if available)
		if a.aiDesignerService != nil {
			aiDesignerHandler := ai.NewAgencyDesignerHandler(a.aiDesignerService, a.logger)
			aiDesignerHandler.RegisterRoutes(v1)
			a.logger.Info("AI Agency Designer endpoints registered")
		}

		// Webhook endpoints for work item integration (if available)
		if a.webhookHandler != nil {
			work := v1.Group("/work")
			{
				work.POST("/issues", a.webhookHandler.HandleIssueWebhook)
				work.POST("/pull-requests", a.webhookHandler.HandlePullRequestWebhook)
			}
			a.logger.Info("Work item webhook endpoints registered")
		}

		v1.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"app_name": a.config.AppName,
				"status":   "running",
				"version":  "dev",
			})
		})
	}
}
