package handlers

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers/ai_refine"
	"github.com/gin-gonic/gin"
)

// This file defines the standard interface signatures for AI refine handler services.
// All interfaces follow consistent signature patterns for HTTP handlers and context builders.
//
// Standard HTTP Handler Signature Pattern:
//   func (h *Handler) MethodName(c *gin.Context)
//
// Where:
//   - c: Gin context containing HTTP request/response, parameters, and request data
//   - Methods handle: request parsing, validation, calling AI services, returning responses
//
// Standard Context Builder Signature Pattern:
//   func (b *BuilderContextBuilder) BuildBuilderContext(ctx context.Context, agencyObj *models.Agency, currentIntroduction string, userRequest string) (builder.BuilderContext, error)
//
// Where:
//   - ctx: Go context for cancellation and deadlines
//   - agencyObj: The agency entity containing core metadata
//   - Returns: BuilderContext with all agency data (goals, work items, roles, RACI assignments)
//
// Benefits:
//   - Clear interface contracts for dependency injection
//   - Easier testing and mocking
//   - Type-safe handler and context builder definitions
//   - Predictable method signatures across all AI operations
//   - Consistent with builder package interface patterns

// IntroductionBuilderInterface defines the contract for introduction refinement HTTP handlers.
type IntroductionBuilderInterface interface {
	// RefineIntroduction handles POST /api/v1/agencies/:id/introduction/refine
	// Refines an agency introduction using AI with full agency context.
	RefineIntroduction(c *gin.Context)
}

// GoalBuilderInterface defines the contract for all goal-related AI HTTP handlers.
type GoalBuilderInterface interface {
	// RefineGoal handles POST /api/v1/agencies/:id/goals/:goalKey/refine
	// Refines a specific goal definition using AI with full agency context.
	RefineGoal(c *gin.Context)

	// RefineGoals handles POST /api/v1/agencies/:id/goals/refine-dynamic
	// Dynamically determines and executes the appropriate goal operation based on user message.
	// Can refine, generate, consolidate, or enhance goals based on natural language input.
	RefineGoals(c *gin.Context)

	// GenerateGoal handles POST /api/v1/agencies/:id/goals/generate
	// Generates a single goal using AI based on user input.
	GenerateGoal(c *gin.Context)

	// GenerateGoals handles POST /api/v1/agencies/:id/goals/generate-multiple
	// Generates multiple goals using AI based on user input.
	GenerateGoals(c *gin.Context)

	// ConsolidateGoals handles POST /api/v1/agencies/:id/goals/consolidate
	// Consolidates multiple goals into a lean, strategic list using AI.
	ConsolidateGoals(c *gin.Context)

	// ProcessAIGoalRequest handles POST /api/v1/agencies/:id/ai/goals/process
	// Processes batch AI operations on goals (create, enhance, consolidate).
	ProcessAIGoalRequest(c *gin.Context)

	// ProcessGoalsChatRequest handles chat-based goal interactions (non-streaming)
	ProcessGoalsChatRequest(c *gin.Context)

	// ProcessGoalsChatRequestStreaming handles chat-based goal interactions with streaming
	ProcessGoalsChatRequestStreaming(c *gin.Context)
}

// WorkItemBuilderInterface defines the contract for all work item-related AI HTTP handlers.
type WorkItemBuilderInterface interface {
	// RefineWorkItem handles POST /api/v1/agencies/:id/work-items/:workItemKey/refine
	// Refines a specific work item using AI with full agency context.
	RefineWorkItem(c *gin.Context)

	// GenerateWorkItem handles POST /api/v1/agencies/:id/work-items/generate
	// Generates a single work item using AI based on user input.
	GenerateWorkItem(c *gin.Context)

	// GenerateWorkItems handles POST /api/v1/agencies/:id/work-items/generate-multiple
	// Generates multiple work items using AI based on user input.
	GenerateWorkItems(c *gin.Context)

	// ConsolidateWorkItems handles POST /api/v1/agencies/:id/work-items/consolidate
	// Consolidates multiple work items into a lean, strategic list using AI.
	ConsolidateWorkItems(c *gin.Context)

	// ProcessAIWorkItemRequest handles POST /api/v1/agencies/:id/ai/work-items/process
	// Processes batch AI operations on work items (create, enhance, consolidate).
	ProcessAIWorkItemRequest(c *gin.Context)

	// ProcessWorkItemsChatRequestStreaming handles chat-based work item interactions with streaming
	ProcessWorkItemsChatRequestStreaming(c *gin.Context)
}

// RoleBuilderInterface defines the contract for all role-related AI HTTP handlers.
type RoleBuilderInterface interface {
	// RefineRole handles POST /api/v1/agencies/:id/roles/:roleKey/refine
	// Refines a specific role using AI with full agency context.
	RefineRole(c *gin.Context)

	// GenerateRole handles POST /api/v1/agencies/:id/roles/generate
	// Generates a single role using AI based on user input.
	GenerateRole(c *gin.Context)

	// GenerateRoles handles POST /api/v1/agencies/:id/roles/generate-multiple
	// Generates multiple roles using AI based on agency needs.
	GenerateRoles(c *gin.Context)

	// ConsolidateRoles handles POST /api/v1/agencies/:id/roles/consolidate
	// Consolidates multiple roles into a lean, strategic list using AI.
	ConsolidateRoles(c *gin.Context)

	// ProcessAIRoleRequest handles POST /api/v1/agencies/:id/ai/roles/process
	// Processes batch AI operations on roles (create roles based on agency needs).
	ProcessAIRoleRequest(c *gin.Context)

	// ProcessRolesChatRequestStreaming handles chat-based role interactions with streaming
	ProcessRolesChatRequestStreaming(c *gin.Context)
}

// RACIBuilderInterface defines the contract for all RACI-related AI HTTP handlers.
type RACIBuilderInterface interface {
	// RefineRACIMapping handles POST /api/v1/agencies/:id/raci/:raciKey/refine
	// Refines a specific RACI mapping using AI with full agency context.
	RefineRACIMapping(c *gin.Context)

	// GenerateRACIMapping handles POST /api/v1/agencies/:id/raci/generate
	// Generates a single RACI mapping using AI based on user input.
	GenerateRACIMapping(c *gin.Context)

	// CreateRACIMappings handles POST /api/v1/agencies/:id/raci/create-multiple
	// Creates multiple RACI mappings using AI for all roles and work items.
	CreateRACIMappings(c *gin.Context)

	// ConsolidateRACIMappings handles POST /api/v1/agencies/:id/raci/consolidate
	// Consolidates RACI mappings to ensure proper responsibility distribution.
	ConsolidateRACIMappings(c *gin.Context)

	// ProcessAIRACIRequest handles POST /api/v1/agencies/:id/ai/raci/process
	// Processes batch AI operations on RACI mappings (create RACI assignments).
	ProcessAIRACIRequest(c *gin.Context)
}

// ContextBuilderInterface defines the contract for building AI context from agency data.
// This interface abstracts the process of gathering all necessary agency data (goals, work items,
// roles, RACI assignments) and packaging it into a BuilderContext for AI operations.
type ContextBuilderInterface interface {
	// BuildBuilderContext gathers all agency context data and returns it as a structured BuilderContext.
	// This is the centralized function used by all AI HTTP handlers to ensure consistent context.
	//
	// The method fetches:
	//   - All goals for the agency
	//   - All work items for the agency
	//   - All role types from the registry
	//   - All RACI assignments for the agency
	//
	// Parameters:
	//   - ctx: Go context for cancellation and timeouts
	//   - agencyObj: The agency entity containing metadata (ID, name, category, description)
	//   - currentIntroduction: Optional current introduction text to include in context
	//   - userRequest: Optional user input/request text to include in context
	//
	// Returns:
	//   - builder.BuilderContext: Structured context containing all agency data for AI operations
	//   - error: Any critical error encountered during data gathering (non-critical errors are logged and ignored)
	BuildBuilderContext(
		ctx context.Context,
		agencyObj *models.Agency,
		currentIntroduction string,
		userRequest string,
	) (builder.BuilderContext, error)
}

// Compile-time interface compliance checks
// These ensure that our concrete types implement the interfaces correctly
// TODO: Uncomment these as methods are implemented in ai_refine.Handler
var (
	_ IntroductionBuilderInterface = (*ai_refine.Handler)(nil)
	// _ GoalBuilderInterface         = (*ai_refine.Handler)(nil) // TODO: Implement GenerateGoals
	// _ WorkItemBuilderInterface     = (*ai_refine.Handler)(nil) // TODO: Implement RefineWorkItem, GenerateWorkItem, GenerateWorkItems, ConsolidateWorkItems
	// _ RoleBuilderInterface         = (*ai_refine.Handler)(nil) // TODO: Implement RefineRole, GenerateRole, GenerateRoles, ConsolidateRoles
	// _ RACIBuilderInterface         = (*ai_refine.Handler)(nil) // TODO: Implement RefineRACIMapping, GenerateRACIMapping, CreateRACIMappings, ConsolidateRACIMappings
	_ ContextBuilderInterface = (*ai_refine.BuilderContextBuilder)(nil)
)
