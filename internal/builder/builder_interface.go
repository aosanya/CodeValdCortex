package builder

import (
	"context"
)

// StreamCallback is a function type for handling streaming chunks
type StreamCallback func(chunk string) error

// This file defines the standard interface signatures for all AI builder services.
// All AI operations follow the same signature pattern to ensure consistency:
//
// Standard Signature Pattern:
//   func (r *ServiceType) MethodName(ctx context.Context, req *SpecificRequest, builderContext BuilderContext) (*SpecificResponse, error)
//
// Where:
//   - ctx: Go context for cancellation and deadlines
//   - req: Service-specific request struct containing only request metadata (AgencyID, conversation history, etc.)
//   - builderContext: Shared BuilderContext containing all agency data (introduction, goals, work items, roles, assignments)
//
// Benefits:
//   - Consistent parameter passing across all AI services
//   - Clear separation between request metadata and agency data
//   - Easier testing and mocking (can pass different BuilderContext instances)
//   - Type-safe context handling
//   - Predictable method signatures for all AI operations
//
// Note: The concrete request/response types are defined in their respective *_types.go files in this package
// (introduction_types.go, goal_types.go, work_item_types.go, role_types.go, raci_types.go)

// IntroductionBuilderInterface defines the contract for introduction refinement
type IntroductionBuilderInterface interface {
	RefineIntroduction(ctx context.Context, req *RefineIntroductionRequest, builderContext BuilderContext) (*RefineIntroductionResponse, error)
}

// GoalBuilderInterface defines the contract for all goal-related AI operations (refinement, generation, consolidation)
type GoalBuilderInterface interface {
	RefineGoals(ctx context.Context, req *RefineGoalsRequest, builderContext BuilderContext) (*RefineGoalsResponse, error)
}

// WorkItemBuilderInterface defines the contract for all work item-related AI operations (refinement, generation, consolidation)
type WorkItemBuilderInterface interface {
	RefineWorkItems(ctx context.Context, req *RefineWorkItemsRequest, builderContext BuilderContext) (*RefineWorkItemsResponse, error)
	RefineWorkItemsStream(ctx context.Context, req *RefineWorkItemsRequest, builderContext BuilderContext, streamCallback StreamCallback) (*RefineWorkItemsResponse, error)
}

// RoleBuilderInterface for AI-powered role operations
type RoleBuilderInterface interface {
	RefineRoles(ctx context.Context, req *RefineRolesRequest, builderContext BuilderContext) (*RefineRolesResponse, error)
	RefineRolesStream(ctx context.Context, req *RefineRolesRequest, builderContext BuilderContext, streamCallback StreamCallback) (*RefineRolesResponse, error)
}

// RACIBuilderInterface defines the contract for all RACI-related AI operations (refinement, generation, creation, consolidation)
type RACIBuilderInterface interface {
	RefineRACIMappings(ctx context.Context, req *RefineRACIMappingsRequest, builderContext BuilderContext) (*RefineRACIMappingsResponse, error)
}
