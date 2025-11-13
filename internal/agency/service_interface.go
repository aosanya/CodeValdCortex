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

	// Specification methods (unified document approach)
	GetSpecification(ctx context.Context, agencyID string) (*models.AgencySpecification, error)
	UpdateSpecification(ctx context.Context, agencyID string, req *models.SpecificationUpdateRequest) (*models.AgencySpecification, error)
	UpdateIntroduction(ctx context.Context, agencyID, introduction, updatedBy string) (*models.AgencySpecification, error)
	UpdateSpecificationGoals(ctx context.Context, agencyID string, goals []models.Goal, updatedBy string) (*models.AgencySpecification, error)
	UpdateSpecificationWorkItems(ctx context.Context, agencyID string, workItems []models.WorkItem, updatedBy string) (*models.AgencySpecification, error)
	UpdateSpecificationWorkflows(ctx context.Context, agencyID string, workflows []models.Workflow, updatedBy string) (*models.AgencySpecification, error)
	UpdateSpecificationRoles(ctx context.Context, agencyID string, roles []models.Role, updatedBy string) (*models.AgencySpecification, error)
	UpdateSpecificationRACIMatrix(ctx context.Context, agencyID string, matrix *models.RACIMatrix, updatedBy string) (*models.AgencySpecification, error)
}

// Use services.New() or services.NewWithDBInit() to create a service instance.
