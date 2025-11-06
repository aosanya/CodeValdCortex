package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
)

const (
	workflowsCollection          = "workflows"
	workflowExecutionsCollection = "workflow_executions"
)

// ArangoRepository implements Repository interface using ArangoDB
type ArangoRepository struct {
	db     driver.Database
	logger *logrus.Logger
}

// NewArangoRepository creates a new ArangoDB repository for workflows
func NewArangoRepository(db driver.Database, logger *logrus.Logger) (*ArangoRepository, error) {
	repo := &ArangoRepository{
		db:     db,
		logger: logger,
	}

	// Ensure collections and indexes exist
	if err := repo.ensureCollections(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure collections: %w", err)
	}

	return repo, nil
}

// ensureCollections creates collections and indexes if they don't exist
func (r *ArangoRepository) ensureCollections(ctx context.Context) error {
	// Create workflows collection
	if err := r.ensureCollection(ctx, workflowsCollection); err != nil {
		return err
	}

	// Create workflow_executions collection
	if err := r.ensureCollection(ctx, workflowExecutionsCollection); err != nil {
		return err
	}

	// Create indexes
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}

	return nil
}

// ensureCollection creates a collection if it doesn't exist
func (r *ArangoRepository) ensureCollection(ctx context.Context, name string) error {
	exists, err := r.db.CollectionExists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !exists {
		_, err := r.db.CreateCollection(ctx, name, nil)
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
		r.logger.Infof("Created collection: %s", name)
	} else {
		r.logger.Debugf("Using existing collection: %s", name)
	}

	return nil
}

// ensureIndexes creates necessary indexes
func (r *ArangoRepository) ensureIndexes(ctx context.Context) error {
	workflowsCol, err := r.db.Collection(ctx, workflowsCollection)
	if err != nil {
		return fmt.Errorf("failed to get workflows collection: %w", err)
	}

	// Index on agency_id for fast agency-specific queries
	_, _, err = workflowsCol.EnsurePersistentIndex(ctx, []string{"agency_id"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_workflows_agency_id",
		Unique: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create agency_id index: %w", err)
	}

	// Index on status for filtering
	_, _, err = workflowsCol.EnsurePersistentIndex(ctx, []string{"status"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_workflows_status",
		Unique: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	// Index on created_at for sorting
	_, _, err = workflowsCol.EnsurePersistentIndex(ctx, []string{"created_at"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_workflows_created_at",
		Unique: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create created_at index: %w", err)
	}

	// Indexes for workflow_executions
	executionsCol, err := r.db.Collection(ctx, workflowExecutionsCollection)
	if err != nil {
		return fmt.Errorf("failed to get executions collection: %w", err)
	}

	// Index on workflow_id
	_, _, err = executionsCol.EnsurePersistentIndex(ctx, []string{"workflow_id"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_executions_workflow_id",
		Unique: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create workflow_id index: %w", err)
	}

	// Index on status
	_, _, err = executionsCol.EnsurePersistentIndex(ctx, []string{"status"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_executions_status",
		Unique: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create execution status index: %w", err)
	}

	r.logger.Info("Workflow indexes created successfully")
	return nil
}

// Create creates a new workflow
func (r *ArangoRepository) Create(ctx context.Context, workflow *Workflow) error {
	col, err := r.db.Collection(ctx, workflowsCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Set timestamps
	now := time.Now()
	workflow.CreatedAt = now
	workflow.UpdatedAt = now

	// Set default status if not provided
	if workflow.Status == "" {
		workflow.Status = WorkflowStatusDraft
	}

	// Initialize empty arrays if nil
	if workflow.Nodes == nil {
		workflow.Nodes = []Node{}
	}
	if workflow.Edges == nil {
		workflow.Edges = []Edge{}
	}
	if workflow.Variables == nil {
		workflow.Variables = make(map[string]interface{})
	}

	meta, err := col.CreateDocument(ctx, workflow)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	workflow.ID = meta.Key
	r.logger.WithFields(logrus.Fields{
		"workflow_id": workflow.ID,
		"name":        workflow.Name,
		"agency_id":   workflow.AgencyID,
	}).Info("Created workflow")

	return nil
}

// GetByID retrieves a workflow by its ID
func (r *ArangoRepository) GetByID(ctx context.Context, id string) (*Workflow, error) {
	col, err := r.db.Collection(ctx, workflowsCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	var workflow Workflow
	_, err = col.ReadDocument(ctx, id, &workflow)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("workflow not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read workflow: %w", err)
	}

	workflow.ID = id
	return &workflow, nil
}

// GetByAgencyID retrieves all workflows for a specific agency
func (r *ArangoRepository) GetByAgencyID(ctx context.Context, agencyID string) ([]*Workflow, error) {
	query := `
		FOR w IN @@collection
		FILTER w.agency_id == @agency_id
		SORT w.created_at DESC
		RETURN w
	`

	bindVars := map[string]interface{}{
		"@collection": workflowsCollection,
		"agency_id":   agencyID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflows: %w", err)
	}
	defer cursor.Close()

	var workflows []*Workflow
	for {
		var workflow Workflow
		meta, err := cursor.ReadDocument(ctx, &workflow)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read workflow document: %w", err)
		}

		workflow.ID = meta.Key
		workflows = append(workflows, &workflow)
	}

	return workflows, nil
}

// Update updates an existing workflow
func (r *ArangoRepository) Update(ctx context.Context, workflow *Workflow) error {
	col, err := r.db.Collection(ctx, workflowsCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Update timestamp
	workflow.UpdatedAt = time.Now()

	_, err = col.UpdateDocument(ctx, workflow.ID, workflow)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("workflow not found: %s", workflow.ID)
		}
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"workflow_id": workflow.ID,
		"name":        workflow.Name,
	}).Info("Updated workflow")

	return nil
}

// Delete deletes a workflow
func (r *ArangoRepository) Delete(ctx context.Context, id string) error {
	col, err := r.db.Collection(ctx, workflowsCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	_, err = col.RemoveDocument(ctx, id)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("workflow not found: %s", id)
		}
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	r.logger.WithField("workflow_id", id).Info("Deleted workflow")
	return nil
}

// List retrieves workflows with pagination
func (r *ArangoRepository) List(ctx context.Context, limit, offset int) ([]*Workflow, error) {
	query := `
		FOR w IN @@collection
		SORT w.created_at DESC
		LIMIT @offset, @limit
		RETURN w
	`

	bindVars := map[string]interface{}{
		"@collection": workflowsCollection,
		"limit":       limit,
		"offset":      offset,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflows: %w", err)
	}
	defer cursor.Close()

	var workflows []*Workflow
	for {
		var workflow Workflow
		meta, err := cursor.ReadDocument(ctx, &workflow)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read workflow document: %w", err)
		}

		workflow.ID = meta.Key
		workflows = append(workflows, &workflow)
	}

	return workflows, nil
}

// CreateExecution creates a new workflow execution
func (r *ArangoRepository) CreateExecution(ctx context.Context, execution *WorkflowExecution) error {
	col, err := r.db.Collection(ctx, workflowExecutionsCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Set start time
	execution.StartedAt = time.Now()

	// Initialize arrays if nil
	if execution.NodeExecutions == nil {
		execution.NodeExecutions = []NodeExecution{}
	}
	if execution.Errors == nil {
		execution.Errors = []string{}
	}
	if execution.Context == nil {
		execution.Context = make(map[string]interface{})
	}

	meta, err := col.CreateDocument(ctx, execution)
	if err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	execution.ID = meta.Key
	r.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"workflow_id":  execution.WorkflowID,
	}).Info("Created workflow execution")

	return nil
}

// GetExecution retrieves an execution by its ID
func (r *ArangoRepository) GetExecution(ctx context.Context, id string) (*WorkflowExecution, error) {
	col, err := r.db.Collection(ctx, workflowExecutionsCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	var execution WorkflowExecution
	_, err = col.ReadDocument(ctx, id, &execution)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("execution not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read execution: %w", err)
	}

	execution.ID = id
	return &execution, nil
}

// GetExecutionsByWorkflowID retrieves all executions for a specific workflow
func (r *ArangoRepository) GetExecutionsByWorkflowID(ctx context.Context, workflowID string) ([]*WorkflowExecution, error) {
	query := `
		FOR e IN @@collection
		FILTER e.workflow_id == @workflow_id
		SORT e.started_at DESC
		RETURN e
	`

	bindVars := map[string]interface{}{
		"@collection": workflowExecutionsCollection,
		"workflow_id": workflowID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}
	defer cursor.Close()

	var executions []*WorkflowExecution
	for {
		var execution WorkflowExecution
		meta, err := cursor.ReadDocument(ctx, &execution)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read execution document: %w", err)
		}

		execution.ID = meta.Key
		executions = append(executions, &execution)
	}

	return executions, nil
}

// UpdateExecution updates an existing execution
func (r *ArangoRepository) UpdateExecution(ctx context.Context, execution *WorkflowExecution) error {
	col, err := r.db.Collection(ctx, workflowExecutionsCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	_, err = col.UpdateDocument(ctx, execution.ID, execution)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("execution not found: %s", execution.ID)
		}
		return fmt.Errorf("failed to update execution: %w", err)
	}

	return nil
}

// UpdateNodeExecution updates a specific node execution within a workflow execution
func (r *ArangoRepository) UpdateNodeExecution(ctx context.Context, executionID string, nodeExecution *NodeExecution) error {
	// Get the execution
	execution, err := r.GetExecution(ctx, executionID)
	if err != nil {
		return err
	}

	// Find and update the node execution
	found := false
	for i, ne := range execution.NodeExecutions {
		if ne.NodeID == nodeExecution.NodeID {
			execution.NodeExecutions[i] = *nodeExecution
			found = true
			break
		}
	}

	// If not found, append
	if !found {
		execution.NodeExecutions = append(execution.NodeExecutions, *nodeExecution)
	}

	// Update the execution
	return r.UpdateExecution(ctx, execution)
}
