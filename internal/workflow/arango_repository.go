package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
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

	r.logger.Info("models.Workflow indexes created successfully")
	return nil
}

// Create creates a new workflow
func (r *ArangoRepository) Create(ctx context.Context, workflow *models.Workflow) error {
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
		workflow.Status = models.WorkflowStatusDraft
	}

	// Initialize empty arrays if nil
	if workflow.Nodes == nil {
		workflow.Nodes = []models.WorkflowNode{}
	}
	if workflow.Edges == nil {
		workflow.Edges = []models.WorkflowEdge{}
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
func (r *ArangoRepository) GetByID(ctx context.Context, id string) (*models.Workflow, error) {
	col, err := r.db.Collection(ctx, workflowsCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	var workflow models.Workflow
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
func (r *ArangoRepository) GetByAgencyID(ctx context.Context, agencyID string) ([]*models.Workflow, error) {
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

	var workflows []*models.Workflow
	for {
		var workflow models.Workflow
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
func (r *ArangoRepository) Update(ctx context.Context, workflow *models.Workflow) error {
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
func (r *ArangoRepository) List(ctx context.Context, limit, offset int) ([]*models.Workflow, error) {
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

	var workflows []*models.Workflow
	for {
		var workflow models.Workflow
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
