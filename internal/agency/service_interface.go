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

	// Problem methods
	CreateProblem(ctx context.Context, agencyID string, code string, description string) (*Problem, error)
	GetProblems(ctx context.Context, agencyID string) ([]*Problem, error)
	UpdateProblem(ctx context.Context, agencyID string, key string, code string, description string) error
	DeleteProblem(ctx context.Context, agencyID string, key string) error

	// UnitOfWork methods
	CreateUnitOfWork(ctx context.Context, agencyID string, code string, description string) (*UnitOfWork, error)
	GetUnitsOfWork(ctx context.Context, agencyID string) ([]*UnitOfWork, error)
	UpdateUnitOfWork(ctx context.Context, agencyID string, key string, code string, description string) error
	DeleteUnitOfWork(ctx context.Context, agencyID string, key string) error
}

// Use services.New() or services.NewWithDBInit() to create a service instance.
