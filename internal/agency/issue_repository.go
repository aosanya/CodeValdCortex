package agency

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// IssueRepository defines the interface for work issue data persistence
type IssueRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, issue *models.WorkIssue) error
	GetByID(ctx context.Context, agencyID, instanceID, issueID string) (*models.WorkIssue, error)
	GetByKey(ctx context.Context, agencyID, key string) (*models.WorkIssue, error)
	Update(ctx context.Context, issue *models.WorkIssue) error
	Delete(ctx context.Context, agencyID, instanceID, issueID string) error

	// Query operations
	List(ctx context.Context, agencyID, instanceID string, filters IssueFilters) ([]*models.WorkIssue, error)
	ListByWorkflowStep(ctx context.Context, agencyID, instanceID, workflowID, step string) ([]*models.WorkIssue, error)
	ListAvailable(ctx context.Context, agencyID, instanceID string, step string) ([]*models.WorkIssue, error)
	CountByStep(ctx context.Context, agencyID, instanceID, workflowID, step string) (int, error)

	// Issue number management
	GetNextIssueNumber(ctx context.Context, agencyID, instanceID string) (int, error)

	// Existence checks
	Exists(ctx context.Context, agencyID, instanceID, issueID string) (bool, error)
}

// IssueFilters represents query filters for listing issues
type IssueFilters struct {
	WorkflowID string
	Status     string
	AssignedTo string
	Offset     int
	Limit      int
}
