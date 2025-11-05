package builder

import (
	"context"
)

// This file defines the standard interface signatures for all AI builder services.
// All AI operations follow the same signature pattern to ensure consistency:
//
// Standard Signature Pattern:
//   func (r *ServiceType) MethodName(ctx context.Context, req *SpecificRequest, aiContext BuilderContext) (*SpecificResponse, error)
//
// Where:
//   - ctx: Go context for cancellation and deadlines
//   - req: Service-specific request struct containing only request metadata (AgencyID, conversation history, etc.)
//   - aiContext: Shared BuilderContext containing all agency data (introduction, goals, work items, roles, assignments)
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
	RefineIntroduction(ctx context.Context, req *RefineIntroductionRequest, aiContext BuilderContext) (*RefineIntroductionResponse, error)
}

// GoalBuilderInterface defines the contract for all goal-related AI operations (refinement, generation, consolidation)
type GoalBuilderInterface interface {
	RefineGoal(ctx context.Context, req *RefineGoalRequest, aiContext BuilderContext) (*RefineGoalResponse, error)
	GenerateGoal(ctx context.Context, req *GenerateGoalRequest, aiContext BuilderContext) (*GenerateGoalResponse, error)
	GenerateGoals(ctx context.Context, req *GenerateGoalRequest, aiContext BuilderContext) (*GenerateGoalsResponse, error)
	ConsolidateGoals(ctx context.Context, req *ConsolidateGoalsRequest, aiContext BuilderContext) (*ConsolidateGoalsResponse, error)
}

// WorkItemBuilderInterface defines the contract for all work item-related AI operations (refinement, generation, consolidation)
type WorkItemBuilderInterface interface {
	RefineWorkItem(ctx context.Context, req *RefineWorkItemRequest, aiContext BuilderContext) (*RefineWorkItemResponse, error)
	GenerateWorkItem(ctx context.Context, req *GenerateWorkItemRequest, aiContext BuilderContext) (*GenerateWorkItemResponse, error)
	GenerateWorkItems(ctx context.Context, req *GenerateWorkItemRequest, aiContext BuilderContext) (*GenerateWorkItemsResponse, error)
	ConsolidateWorkItems(ctx context.Context, req *ConsolidateWorkItemsRequest, aiContext BuilderContext) (*ConsolidateWorkItemsResponse, error)
}

// RoleBuilderInterface defines the contract for all role-related AI operations (refinement, generation, consolidation)
type RoleBuilderInterface interface {
	RefineRole(ctx context.Context, req *RefineRoleRequest, aiContext BuilderContext) (*RefineRoleResponse, error)
	GenerateRole(ctx context.Context, req *GenerateRoleRequest, aiContext BuilderContext) (*GenerateRoleResponse, error)
	GenerateRoles(ctx context.Context, req *GenerateRolesRequest, aiContext BuilderContext) (*GenerateRolesResponse, error)
	ConsolidateRoles(ctx context.Context, req *ConsolidateRolesRequest, aiContext BuilderContext) (*ConsolidateRolesResponse, error)
}

// RACIBuilderInterface defines the contract for all RACI-related AI operations (refinement, generation, creation, consolidation)
type RACIBuilderInterface interface {
	RefineRACIMapping(ctx context.Context, req *RefineRACIMappingRequest, aiContext BuilderContext) (*RefineRACIMappingResponse, error)
	GenerateRACIMapping(ctx context.Context, req *GenerateRACIMappingRequest, aiContext BuilderContext) (*GenerateRACIMappingResponse, error)
	CreateRACIMappings(ctx context.Context, req *CreateRACIMappingsRequest, aiContext BuilderContext) (*CreateRACIMappingsResponse, error)
	ConsolidateRACIMappings(ctx context.Context, req *ConsolidateRACIMappingsRequest, aiContext BuilderContext) (*ConsolidateRACIMappingsResponse, error)
}
