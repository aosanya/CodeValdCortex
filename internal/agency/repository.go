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

	// Specification methods (unified document approach)
	GetSpecification(ctx context.Context, agencyID string) (*models.AgencySpecification, error)
	CreateSpecification(ctx context.Context, agencyID string, req *models.CreateSpecificationRequest) (*models.AgencySpecification, error)
	UpdateSpecification(ctx context.Context, agencyID string, req *models.SpecificationUpdateRequest) (*models.AgencySpecification, error)
	PatchSpecificationSection(ctx context.Context, agencyID, section string, data interface{}, updatedBy string) (*models.AgencySpecification, error)
	DeleteSpecification(ctx context.Context, agencyID string) error
}
