package agency

import (
	"context"
)

// Repository defines the interface for agency data persistence
type Repository interface {
	Create(ctx context.Context, agency *Agency) error
	GetByID(ctx context.Context, id string) (*Agency, error)
	List(ctx context.Context, filters AgencyFilters) ([]*Agency, error)
	Update(ctx context.Context, agency *Agency) error
	Delete(ctx context.Context, id string) error
	GetStatistics(ctx context.Context, id string) (*AgencyStatistics, error)
	Exists(ctx context.Context, id string) (bool, error)

	// Overview methods
	GetOverview(ctx context.Context, agencyID string) (*Overview, error)
	UpdateOverview(ctx context.Context, overview *Overview) error

	// Goal methods
	CreateGoal(ctx context.Context, goal *Goal) error
	GetGoals(ctx context.Context, agencyID string) ([]*Goal, error)
	GetGoal(ctx context.Context, agencyID string, key string) (*Goal, error)
	UpdateGoal(ctx context.Context, goal *Goal) error
	DeleteGoal(ctx context.Context, agencyID string, key string) error

	// WorkItem methods
	CreateWorkItem(ctx context.Context, workItem *WorkItem) error
	GetWorkItems(ctx context.Context, agencyID string) ([]*WorkItem, error)
	GetWorkItem(ctx context.Context, agencyID string, key string) (*WorkItem, error)
	GetWorkItemByCode(ctx context.Context, agencyID string, code string) (*WorkItem, error)
	UpdateWorkItem(ctx context.Context, workItem *WorkItem) error
	DeleteWorkItem(ctx context.Context, agencyID string, key string) error
	ValidateDependencies(ctx context.Context, agencyID string, workItemCode string, dependencies []string) error

	// UnitOfWork methods
	CreateUnitOfWork(ctx context.Context, unit *UnitOfWork) error
	GetUnitsOfWork(ctx context.Context, agencyID string) ([]*UnitOfWork, error)
	GetUnitOfWork(ctx context.Context, agencyID string, key string) (*UnitOfWork, error)
	UpdateUnitOfWork(ctx context.Context, unit *UnitOfWork) error
	DeleteUnitOfWork(ctx context.Context, agencyID string, key string) error
}
