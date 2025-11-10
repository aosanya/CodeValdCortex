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

	// Workflow execution operations
	CreateExecution(ctx context.Context, execution *models.WorkflowExecution) error
	GetExecution(ctx context.Context, id string) (*models.WorkflowExecution, error)
	GetExecutionsByWorkflowID(ctx context.Context, workflowID string) ([]*models.WorkflowExecution, error)
	UpdateExecution(ctx context.Context, execution *models.WorkflowExecution) error

	// Node execution operations
	UpdateNodeExecution(ctx context.Context, executionID string, nodeExecution *models.NodeExecution) error
}
