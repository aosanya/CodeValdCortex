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

	// UnitOfWork methods
	CreateUnitOfWork(ctx context.Context, agencyID string, code string, description string) (*UnitOfWork, error)
	GetUnitsOfWork(ctx context.Context, agencyID string) ([]*UnitOfWork, error)
	UpdateUnitOfWork(ctx context.Context, agencyID string, key string, code string, description string) error
	DeleteUnitOfWork(ctx context.Context, agencyID string, key string) error
}

// Use services.New() or services.NewWithDBInit() to create a service instance.
