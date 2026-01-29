package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/auth"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers/ai_refine"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterAIRefineRoutes registers AI refinement endpoints
func RegisterAIRefineRoutes(
	rg *gin.RouterGroup,
	refineHandler *ai_refine.Handler,
	goalRefiner *ai.GoalsBuilder,
	workItemBuilder *ai.WorkItemsBuilder,
	roleBuilder *ai.RolesBuilder,
	raciBuilder *ai.RACIBuilder,
	workflowBuilder *ai.WorkflowsBuilder,
	authMiddleware *auth.Middleware,
	logger *logrus.Logger,
) {
	if refineHandler == nil {
		return
	}

	// Protected AI refinement routes - require authentication
	protected := rg.Group("/agencies/:id")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Introduction refinement
		protected.POST("/overview/refine", refineHandler.RefineIntroduction)

		// Deliverable management endpoints
		protected.POST("/deliverables/move", refineHandler.MoveDeliverable)

		// Goal refinement endpoints
		if goalRefiner != nil {
			// Main dynamic router - handles all goal operations through natural language prompts
			protected.POST("/goals/refine-dynamic", refineHandler.RefineGoals)
			// Convenience routes that use RefineGoals with preset prompts
			protected.POST("/goals/:goalKey/refine", refineHandler.RefineSpecificGoal)
			protected.POST("/goals/generate", refineHandler.GenerateGoalWithPrompt)
			protected.POST("/goals/consolidate", refineHandler.ConsolidateGoalsWithPrompt)
		}

		// Work Item refinement endpoints
		if workItemBuilder != nil {
			// DISABLED: Work item AI handlers need refactoring for unified specification model
			logger.Warn("Work item AI refine endpoints disabled - need refactoring for unified specification")
		}

		// Role refinement endpoints
		if roleBuilder != nil {
			// Main dynamic router - handles all role operations through natural language prompts
			protected.POST("/roles/refine-dynamic", refineHandler.RefineRoles)
			// Convenience routes that use RefineRoles with preset prompts
			protected.POST("/roles/refine-specific", refineHandler.RefineSpecificRole)
			protected.POST("/roles/generate", refineHandler.GenerateRoleWithPrompt)
			protected.POST("/roles/consolidate", refineHandler.ConsolidateRolesWithPrompt)
			protected.POST("/roles/enhance-all", refineHandler.EnhanceAllRolesWithPrompt)
		}

		// RACI Matrix refinement endpoints
		if raciBuilder != nil {
			// Main dynamic router - handles all RACI operations through natural language prompts
			protected.POST("/raci-matrix/refine-dynamic", refineHandler.RefineRACIMappings)
			// Convenience routes that use RefineRACIMappings with preset prompts
			protected.POST("/raci-matrix/refine-specific", refineHandler.RefineSpecificRACIMapping)
			protected.POST("/raci-matrix/generate", refineHandler.GenerateRACIMappingWithPrompt)
			protected.POST("/raci-matrix/consolidate", refineHandler.ConsolidateRACIMappingsWithPrompt)
			protected.POST("/raci-matrix/create-complete", refineHandler.CreateCompleteRACIMatrixWithPrompt)
		}

		// Workflow refinement endpoints
		if workflowBuilder != nil {
			// DISABLED: Workflow handler needs refactoring for unified specification model
			logger.Warn("Workflow AI refine endpoint disabled - needs refactoring for unified specification")
		}
	}

	logger.Info("AI Refine endpoints registered (protected)")
}
