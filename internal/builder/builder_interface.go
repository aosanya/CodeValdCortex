package builder

import (
	"context"
)

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
	RefineGoal(ctx context.Context, req *RefineGoalRequest, builderContext BuilderContext) (*RefineGoalResponse, error)
	GenerateGoal(ctx context.Context, req *GenerateGoalRequest, builderContext BuilderContext) (*GenerateGoalResponse, error)
	GenerateGoals(ctx context.Context, req *GenerateGoalRequest, builderContext BuilderContext) (*GenerateGoalsResponse, error)
	ConsolidateGoals(ctx context.Context, req *ConsolidateGoalsRequest, builderContext BuilderContext) (*ConsolidateGoalsResponse, error)
}

// WorkItemBuilderInterface defines the contract for all work item-related AI operations (refinement, generation, consolidation)
type WorkItemBuilderInterface interface {
	RefineWorkItem(ctx context.Context, req *RefineWorkItemRequest, builderContext BuilderContext) (*RefineWorkItemResponse, error)
	GenerateWorkItem(ctx context.Context, req *GenerateWorkItemRequest, builderContext BuilderContext) (*GenerateWorkItemResponse, error)
	GenerateWorkItems(ctx context.Context, req *GenerateWorkItemRequest, builderContext BuilderContext) (*GenerateWorkItemsResponse, error)
	ConsolidateWorkItems(ctx context.Context, req *ConsolidateWorkItemsRequest, builderContext BuilderContext) (*ConsolidateWorkItemsResponse, error)
}

// RoleBuilderInterface defines the contract for all role-related AI operations (refinement, generation, consolidation)
type RoleBuilderInterface interface {
	RefineRole(ctx context.Context, req *RefineRoleRequest, builderContext BuilderContext) (*RefineRoleResponse, error)
	GenerateRole(ctx context.Context, req *GenerateRoleRequest, builderContext BuilderContext) (*GenerateRoleResponse, error)
	GenerateRoles(ctx context.Context, req *GenerateRolesRequest, builderContext BuilderContext) (*GenerateRolesResponse, error)
	ConsolidateRoles(ctx context.Context, req *ConsolidateRolesRequest, builderContext BuilderContext) (*ConsolidateRolesResponse, error)
}

// RACIBuilderInterface defines the contract for all RACI-related AI operations (refinement, generation, creation, consolidation)
type RACIBuilderInterface interface {
	RefineRACIMapping(ctx context.Context, req *RefineRACIMappingRequest, builderContext BuilderContext) (*RefineRACIMappingResponse, error)
	GenerateRACIMapping(ctx context.Context, req *GenerateRACIMappingRequest, builderContext BuilderContext) (*GenerateRACIMappingResponse, error)
	CreateRACIMappings(ctx context.Context, req *CreateRACIMappingsRequest, builderContext BuilderContext) (*CreateRACIMappingsResponse, error)
	ConsolidateRACIMappings(ctx context.Context, req *ConsolidateRACIMappingsRequest, builderContext BuilderContext) (*ConsolidateRACIMappingsResponse, error)
}
