package agency

import "context"

// Service defines the interface for agency business logic operations
type Service interface {
	// Agency operations
	CreateAgency(ctx context.Context, agency *Agency) error
	GetAgency(ctx context.Context, id string) (*Agency, error)
	ListAgencies(ctx context.Context, filters AgencyFilters) ([]*Agency, error)
	UpdateAgency(ctx context.Context, id string, updates AgencyUpdates) error
	DeleteAgency(ctx context.Context, id string) error
	SetActiveAgency(ctx context.Context, id string) error
	GetActiveAgency(ctx context.Context) (*Agency, error)
	GetAgencyStatistics(ctx context.Context, id string) (*AgencyStatistics, error)

	// Overview methods
	GetAgencyOverview(ctx context.Context, agencyID string) (*Overview, error)
	UpdateAgencyOverview(ctx context.Context, agencyID string, introduction string) error

	// Goal methods
	CreateGoal(ctx context.Context, agencyID string, code string, description string) (*Goal, error)
	GetGoals(ctx context.Context, agencyID string) ([]*Goal, error)
	GetGoal(ctx context.Context, agencyID string, key string) (*Goal, error)
	UpdateGoal(ctx context.Context, agencyID string, key string, code string, description string) error
	DeleteGoal(ctx context.Context, agencyID string, key string) error

	// WorkItem methods
	CreateWorkItem(ctx context.Context, agencyID string, req CreateWorkItemRequest) (*WorkItem, error)
	GetWorkItems(ctx context.Context, agencyID string) ([]*WorkItem, error)
	GetWorkItem(ctx context.Context, agencyID string, key string) (*WorkItem, error)
	GetWorkItemByCode(ctx context.Context, agencyID string, code string) (*WorkItem, error)
	UpdateWorkItem(ctx context.Context, agencyID string, key string, req UpdateWorkItemRequest) error
	DeleteWorkItem(ctx context.Context, agencyID string, key string) error
	ValidateWorkItemDependencies(ctx context.Context, agencyID string, workItemCode string, dependencies []string) error
}

// Use services.New() or services.NewWithDBInit() to create a service instance.
