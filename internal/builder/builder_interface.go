package builder

import (
	"context"
)

// This file defines the standard interface signatures for all AI builder services.
// All AI operations follow the same signature pattern to ensure consistency:
//
// Standard Signature Pattern:
//   func (r *ServiceType) MethodName(ctx context.Context, req *SpecificRequest, aiContext AIContext) (*SpecificResponse, error)
//
// Where:
//   - ctx: Go context for cancellation and deadlines
//   - req: Service-specific request struct containing only request metadata (AgencyID, conversation history, etc.)
//   - aiContext: Shared AIContext containing all agency data (introduction, goals, work items, roles, assignments)
//
// Benefits:
//   - Consistent parameter passing across all AI services
//   - Clear separation between request metadata and agency data
//   - Easier testing and mocking (can pass different AIContext instances)
//   - Type-safe context handling
//   - Predictable method signatures for all AI operations
//
// Note: The concrete request/response types are defined in their respective *_types.go files in this package
// (introduction_types.go, goal_types.go, work_item_types.go, role_types.go, raci_types.go)

// IntroductionBuilderInterface defines the contract for introduction refinement
type IntroductionBuilderInterface interface {
	RefineIntroduction(ctx context.Context, req *RefineIntroductionRequest, aiContext AIContext) (*RefineIntroductionResponse, error)
}

// GoalBuilderInterface defines the contract for all goal-related AI operations (refinement, generation, consolidation)
type GoalBuilderInterface interface {
	RefineGoal(ctx context.Context, req *RefineGoalRequest, aiContext AIContext) (*RefineGoalResponse, error)
	GenerateGoal(ctx context.Context, req *GenerateGoalRequest, aiContext AIContext) (*GenerateGoalResponse, error)
	GenerateGoals(ctx context.Context, req *GenerateGoalRequest, aiContext AIContext) (*GenerateGoalsResponse, error)
	ConsolidateGoals(ctx context.Context, req *ConsolidateGoalsRequest, aiContext AIContext) (*ConsolidateGoalsResponse, error)
}

// WorkItemBuilderInterface defines the contract for all work item-related AI operations (refinement, generation, consolidation)
type WorkItemBuilderInterface interface {
	RefineWorkItem(ctx context.Context, req *RefineWorkItemRequest, aiContext AIContext) (*RefineWorkItemResponse, error)
	GenerateWorkItem(ctx context.Context, req *GenerateWorkItemRequest, aiContext AIContext) (*GenerateWorkItemResponse, error)
	GenerateWorkItems(ctx context.Context, req *GenerateWorkItemRequest, aiContext AIContext) (*GenerateWorkItemsResponse, error)
	ConsolidateWorkItems(ctx context.Context, req *ConsolidateWorkItemsRequest, aiContext AIContext) (*ConsolidateWorkItemsResponse, error)
}

// RoleBuilderInterface defines the contract for all role-related AI operations (refinement, generation, consolidation)
type RoleBuilderInterface interface {
	RefineRole(ctx context.Context, req *RefineRoleRequest, aiContext AIContext) (*RefineRoleResponse, error)
	GenerateRole(ctx context.Context, req *GenerateRoleRequest, aiContext AIContext) (*GenerateRoleResponse, error)
	GenerateRoles(ctx context.Context, req *GenerateRolesRequest, aiContext AIContext) (*GenerateRolesResponse, error)
	ConsolidateRoles(ctx context.Context, req *ConsolidateRolesRequest, aiContext AIContext) (*ConsolidateRolesResponse, error)
}

// RACIBuilderInterface defines the contract for all RACI-related AI operations (refinement, generation, creation, consolidation)
type RACIBuilderInterface interface {
	RefineRACIMapping(ctx context.Context, req *RefineRACIMappingRequest, aiContext AIContext) (*RefineRACIMappingResponse, error)
	GenerateRACIMapping(ctx context.Context, req *GenerateRACIMappingRequest, aiContext AIContext) (*GenerateRACIMappingResponse, error)
	CreateRACIMappings(ctx context.Context, req *CreateRACIMappingsRequest, aiContext AIContext) (*CreateRACIMappingsResponse, error)
	ConsolidateRACIMappings(ctx context.Context, req *ConsolidateRACIMappingsRequest, aiContext AIContext) (*ConsolidateRACIMappingsResponse, error)
}
