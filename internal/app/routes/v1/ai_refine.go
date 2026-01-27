package v1

import (
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
	logger *logrus.Logger,
) {
	if refineHandler == nil {
		return
	}

	// Introduction refinement
	rg.POST("/agencies/:id/overview/refine", refineHandler.RefineIntroduction)

	// Deliverable management endpoints
	rg.POST("/agencies/:id/deliverables/move", refineHandler.MoveDeliverable)

	// Goal refinement endpoints
	if goalRefiner != nil {
		// Main dynamic router - handles all goal operations through natural language prompts
		rg.POST("/agencies/:id/goals/refine-dynamic", refineHandler.RefineGoals)
		// Convenience routes that use RefineGoals with preset prompts
		rg.POST("/agencies/:id/goals/:goalKey/refine", refineHandler.RefineSpecificGoal)
		rg.POST("/agencies/:id/goals/generate", refineHandler.GenerateGoalWithPrompt)
		rg.POST("/agencies/:id/goals/consolidate", refineHandler.ConsolidateGoalsWithPrompt)
	}

	// Work Item refinement endpoints
	if workItemBuilder != nil {
		// DISABLED: Work item AI handlers need refactoring for unified specification model
		// Main dynamic router - handles all work item operations through natural language prompts
		// rg.POST("/agencies/:id/work-items/refine-dynamic", refineHandler.RefineWorkItems)
		// rg.POST("/agencies/:id/work-items/refine-specific", refineHandler.RefineSpecificWorkItem)
		// rg.POST("/agencies/:id/work-items/generate", refineHandler.GenerateWorkItemWithPrompt)
		// rg.POST("/agencies/:id/work-items/consolidate", refineHandler.ConsolidateWorkItemsWithPrompt)
		// rg.POST("/agencies/:id/work-items/enhance-all", refineHandler.EnhanceAllWorkItems)
		logger.Warn("Work item AI refine endpoints disabled - need refactoring for unified specification")
	}

	// Role refinement endpoints
	if roleBuilder != nil {
		// Main dynamic router - handles all role operations through natural language prompts
		rg.POST("/agencies/:id/roles/refine-dynamic", refineHandler.RefineRoles)
		// Convenience routes that use RefineRoles with preset prompts
		rg.POST("/agencies/:id/roles/refine-specific", refineHandler.RefineSpecificRole)
		rg.POST("/agencies/:id/roles/generate", refineHandler.GenerateRoleWithPrompt)
		rg.POST("/agencies/:id/roles/consolidate", refineHandler.ConsolidateRolesWithPrompt)
		rg.POST("/agencies/:id/roles/enhance-all", refineHandler.EnhanceAllRolesWithPrompt)
	}

	// RACI Matrix refinement endpoints
	if raciBuilder != nil {
		// Main dynamic router - handles all RACI operations through natural language prompts
		rg.POST("/agencies/:id/raci-matrix/refine-dynamic", refineHandler.RefineRACIMappings)
		// Convenience routes that use RefineRACIMappings with preset prompts
		rg.POST("/agencies/:id/raci-matrix/refine-specific", refineHandler.RefineSpecificRACIMapping)
		rg.POST("/agencies/:id/raci-matrix/generate", refineHandler.GenerateRACIMappingWithPrompt)
		rg.POST("/agencies/:id/raci-matrix/consolidate", refineHandler.ConsolidateRACIMappingsWithPrompt)
		rg.POST("/agencies/:id/raci-matrix/create-complete", refineHandler.CreateCompleteRACIMatrixWithPrompt)
	}

	// Workflow refinement endpoints
	if workflowBuilder != nil {
		// DISABLED: Workflow handler needs refactoring for unified specification model
		// rg.POST("/agencies/:id/workflows/refine-dynamic", refineHandler.RefineWorkflows)
		logger.Warn("Workflow AI refine endpoint disabled - needs refactoring for unified specification")
	}

	logger.Info("AI Refine endpoints registered")
}
