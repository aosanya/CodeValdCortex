package workflow

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// Repository defines the interface for workflow persistence
type Repository interface {
	// Workflow CRUD operations
	Create(ctx context.Context, workflow *models.Workflow) error
	GetByID(ctx context.Context, id string) (*models.Workflow, error)
	GetByAgencyID(ctx context.Context, agencyID string) ([]*models.Workflow, error)
	Update(ctx context.Context, workflow *models.Workflow) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Workflow, error)
}
