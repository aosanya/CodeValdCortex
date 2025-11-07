package agency

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// Repository defines the interface for agency data persistence
type Repository interface {
	Create(ctx context.Context, agency *models.Agency) error
	GetByID(ctx context.Context, id string) (*models.Agency, error)
	List(ctx context.Context, filters models.AgencyFilters) ([]*models.Agency, error)
	Update(ctx context.Context, agency *models.Agency) error
	Delete(ctx context.Context, id string) error
	GetStatistics(ctx context.Context, id string) (*models.AgencyStatistics, error)
	Exists(ctx context.Context, id string) (bool, error)

	// Overview methods
	GetOverview(ctx context.Context, agencyID string) (*models.Overview, error)
	UpdateOverview(ctx context.Context, overview *models.Overview) error

	// Goal methods
	CreateGoal(ctx context.Context, goal *models.Goal) error
	GetGoals(ctx context.Context, agencyID string) ([]*models.Goal, error)
	GetGoal(ctx context.Context, agencyID string, key string) (*models.Goal, error)
	UpdateGoal(ctx context.Context, goal *models.Goal) error
	DeleteGoal(ctx context.Context, agencyID string, key string) error

	// WorkItem methods
	CreateWorkItem(ctx context.Context, workItem *models.WorkItem) error
	GetWorkItems(ctx context.Context, agencyID string) ([]*models.WorkItem, error)
	GetWorkItem(ctx context.Context, agencyID string, key string) (*models.WorkItem, error)
	GetWorkItemByCode(ctx context.Context, agencyID string, code string) (*models.WorkItem, error)
	UpdateWorkItem(ctx context.Context, workItem *models.WorkItem) error
	DeleteWorkItem(ctx context.Context, agencyID string, key string) error
	ValidateDependencies(ctx context.Context, agencyID string, workItemCode string, dependencies []string) error

	// RACI Matrix methods
	SaveRACIMatrix(ctx context.Context, agencyID string, matrix *models.RACIMatrix) error
	GetRACIMatrix(ctx context.Context, agencyID string, key string) (*models.RACIMatrix, error)
	ListRACIMatrices(ctx context.Context, agencyID string) ([]*models.RACIMatrix, error)
	UpdateRACIMatrix(ctx context.Context, agencyID string, matrix *models.RACIMatrix) error
	DeleteRACIMatrix(ctx context.Context, agencyID string, key string) error

	// RACI Assignment edge methods (graph-based)
	CreateRACIAssignment(ctx context.Context, agencyID string, assignment *models.RACIAssignment) error
	GetRACIAssignmentsForWorkItem(ctx context.Context, agencyID string, workItemKey string) ([]*models.RACIAssignment, error)
	GetRACIAssignmentsForRole(ctx context.Context, agencyID string, roleID string) ([]*models.RACIAssignment, error)
	GetAllRACIAssignments(ctx context.Context, agencyID string) ([]*models.RACIAssignment, error)
	UpdateRACIAssignment(ctx context.Context, agencyID string, key string, assignment *models.RACIAssignment) error
	DeleteRACIAssignment(ctx context.Context, agencyID string, key string) error
	DeleteRACIAssignmentsForWorkItem(ctx context.Context, agencyID string, workItemKey string) error

	// WorkItem-Goal Link edge methods (graph-based)
	CreateWorkItemGoalLink(ctx context.Context, agencyID string, link *models.WorkItemGoalLink) error
	GetWorkItemGoalLinks(ctx context.Context, agencyID, workItemKey string) ([]*models.WorkItemGoalLink, error)
	GetGoalWorkItems(ctx context.Context, agencyID, goalKey string) ([]*models.WorkItemGoalLink, error)
	DeleteWorkItemGoalLink(ctx context.Context, agencyID, linkKey string) error
	DeleteWorkItemGoalLinks(ctx context.Context, agencyID, workItemKey string) error
}
