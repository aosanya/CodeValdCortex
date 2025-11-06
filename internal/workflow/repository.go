package workflow

import "context"

// Repository defines the interface for workflow persistence
type Repository interface {
	// Workflow CRUD operations
	Create(ctx context.Context, workflow *Workflow) error
	GetByID(ctx context.Context, id string) (*Workflow, error)
	GetByAgencyID(ctx context.Context, agencyID string) ([]*Workflow, error)
	Update(ctx context.Context, workflow *Workflow) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Workflow, error)
	
	// Workflow execution operations
	CreateExecution(ctx context.Context, execution *WorkflowExecution) error
	GetExecution(ctx context.Context, id string) (*WorkflowExecution, error)
	GetExecutionsByWorkflowID(ctx context.Context, workflowID string) ([]*WorkflowExecution, error)
	UpdateExecution(ctx context.Context, execution *WorkflowExecution) error
	
	// Node execution operations
	UpdateNodeExecution(ctx context.Context, executionID string, nodeExecution *NodeExecution) error
}
