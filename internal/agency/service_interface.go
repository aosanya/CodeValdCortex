package agency

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// Service defines the interface for agency business logic operations
type Service interface {
	// Agency operations
	CreateAgency(ctx context.Context, agency *models.Agency) error
	GetAgency(ctx context.Context, id string) (*models.Agency, error)
	ListAgencies(ctx context.Context, filters models.AgencyFilters) ([]*models.Agency, error)
	UpdateAgency(ctx context.Context, id string, updates models.AgencyUpdates) error
	DeleteAgency(ctx context.Context, id string) error
	SetActiveAgency(ctx context.Context, id string) error
	GetActiveAgency(ctx context.Context) (*models.Agency, error)
	GetAgencyStatistics(ctx context.Context, id string) (*models.AgencyStatistics, error)

	// Overview methods
	GetAgencyOverview(ctx context.Context, agencyID string) (*models.Overview, error)
	UpdateAgencyOverview(ctx context.Context, agencyID string, introduction string) error

	// Goal methods
	CreateGoal(ctx context.Context, agencyID string, code string, description string) (*models.Goal, error)
	GetGoals(ctx context.Context, agencyID string) ([]*models.Goal, error)
	GetGoal(ctx context.Context, agencyID string, key string) (*models.Goal, error)
	UpdateGoal(ctx context.Context, agencyID string, key string, code string, description string) error
	DeleteGoal(ctx context.Context, agencyID string, key string) error

	// WorkItem methods
	CreateWorkItem(ctx context.Context, agencyID string, req models.CreateWorkItemRequest) (*models.WorkItem, error)
	GetWorkItems(ctx context.Context, agencyID string) ([]*models.WorkItem, error)
	GetWorkItem(ctx context.Context, agencyID string, key string) (*models.WorkItem, error)
	GetWorkItemByCode(ctx context.Context, agencyID string, code string) (*models.WorkItem, error)
	UpdateWorkItem(ctx context.Context, agencyID string, key string, req models.UpdateWorkItemRequest) error
	DeleteWorkItem(ctx context.Context, agencyID string, key string) error
	ValidateWorkItemDependencies(ctx context.Context, agencyID string, workItemCode string, dependencies []string) error

	// RACI Assignment methods (graph-based)
	CreateRACIAssignment(ctx context.Context, agencyID string, assignment *models.RACIAssignment) error
	GetRACIAssignmentsForWorkItem(ctx context.Context, agencyID string, workItemKey string) ([]*models.RACIAssignment, error)
	GetRACIAssignmentsForRole(ctx context.Context, agencyID string, roleID string) ([]*models.RACIAssignment, error)
	GetAllRACIAssignments(ctx context.Context, agencyID string) ([]*models.RACIAssignment, error)
	UpdateRACIAssignment(ctx context.Context, agencyID string, key string, assignment *models.RACIAssignment) error
	DeleteRACIAssignment(ctx context.Context, agencyID string, key string) error
	DeleteRACIAssignmentsForWorkItem(ctx context.Context, agencyID string, workItemKey string) error

	// WorkItem-Goal Link methods (graph-based)
	CreateWorkItemGoalLink(ctx context.Context, agencyID string, link *models.WorkItemGoalLink) error
	GetWorkItemGoalLinks(ctx context.Context, agencyID, workItemKey string) ([]*models.WorkItemGoalLink, error)
	GetGoalWorkItems(ctx context.Context, agencyID, goalKey string) ([]*models.WorkItemGoalLink, error)
	DeleteWorkItemGoalLink(ctx context.Context, agencyID, linkKey string) error
	DeleteWorkItemGoalLinks(ctx context.Context, agencyID, workItemKey string) error
}

// Use services.New() or services.NewWithDBInit() to create a service instance.
