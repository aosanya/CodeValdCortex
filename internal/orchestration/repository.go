package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
	log "github.com/sirupsen/logrus"
)

// Repository provides persistence for workflows and executions
type Repository struct {
	// ArangoDB connection
	db driver.Database

	// Collections
	workflowsCollection  driver.Collection
	executionsCollection driver.Collection

	// Logger
	logger *log.Logger
}

// RepositoryConfig configures the workflow repository
type RepositoryConfig struct {
	// DatabaseName for ArangoDB
	DatabaseName string

	// WorkflowsCollection name
	WorkflowsCollection string

	// ExecutionsCollection name
	ExecutionsCollection string

	// EnableIndexes creates performance indexes
	EnableIndexes bool
}

// DefaultRepositoryConfig returns default repository configuration
func DefaultRepositoryConfig() RepositoryConfig {
	return RepositoryConfig{
		DatabaseName:         "codevaldcortex",
		WorkflowsCollection:  "workflows",
		ExecutionsCollection: "workflow_executions",
		EnableIndexes:        true,
	}
}

// NewRepository creates a new workflow repository
func NewRepository(db driver.Database, config RepositoryConfig, logger *log.Logger) (*Repository, error) {
	repo := &Repository{
		db:     db,
		logger: logger,
	}

	// Initialize collections
	if err := repo.initializeCollections(config); err != nil {
		return nil, fmt.Errorf("failed to initialize collections: %w", err)
	}

	// Create indexes if enabled
	if config.EnableIndexes {
		if err := repo.createIndexes(); err != nil {
			logger.WithError(err).Warn("Failed to create indexes")
		}
	}

	return repo, nil
}

// Workflow CRUD operations

// CreateWorkflow stores a new workflow definition
func (r *Repository) CreateWorkflow(ctx context.Context, workflow *Workflow) error {

	// Set creation timestamp
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = workflow.CreatedAt

	// Insert into database
	_, err := r.workflowsCollection.CreateDocument(ctx, workflow)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	r.logger.WithField("workflow_id", workflow.ID).Info("Workflow created successfully")
	return nil
}

// GetWorkflow retrieves a workflow by ID
func (r *Repository) GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error) {

	var workflow Workflow
	_, err := r.workflowsCollection.ReadDocument(ctx, workflowID, &workflow)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("workflow %s not found", workflowID)
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	return &workflow, nil
}

// GetWorkflowByName retrieves a workflow by name and version
func (r *Repository) GetWorkflowByName(ctx context.Context, name, version string) (*Workflow, error) {

	query := `
		FOR w IN @@collection
		FILTER w.name == @name AND w.version == @version
		RETURN w
	`

	bindVars := map[string]interface{}{
		"@collection": r.workflowsCollection.Name(),
		"name":        name,
		"version":     version,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("workflow %s:%s not found", name, version)
	}

	var workflow Workflow
	_, err = cursor.ReadDocument(ctx, &workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow document: %w", err)
	}

	return &workflow, nil
}

// UpdateWorkflow updates an existing workflow
func (r *Repository) UpdateWorkflow(ctx context.Context, workflow *Workflow) error {

	// Update timestamp
	workflow.UpdatedAt = time.Now()

	// Update document
	_, err := r.workflowsCollection.UpdateDocument(ctx, workflow.ID, workflow)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("workflow %s not found", workflow.ID)
		}
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	r.logger.WithField("workflow_id", workflow.ID).Info("Workflow updated successfully")
	return nil
}

// DeleteWorkflow removes a workflow
func (r *Repository) DeleteWorkflow(ctx context.Context, workflowID string) error {

	_, err := r.workflowsCollection.RemoveDocument(ctx, workflowID)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("workflow %s not found", workflowID)
		}
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	r.logger.WithField("workflow_id", workflowID).Info("Workflow deleted successfully")
	return nil
}

// ListWorkflows retrieves workflows with pagination
func (r *Repository) ListWorkflows(ctx context.Context, limit, offset int) ([]*Workflow, error) {

	query := `
		FOR w IN @@collection
		SORT w.created_at DESC
		LIMIT @offset, @limit
		RETURN w
	`

	bindVars := map[string]interface{}{
		"@collection": r.workflowsCollection.Name(),
		"limit":       limit,
		"offset":      offset,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflows: %w", err)
	}
	defer cursor.Close()

	var workflows []*Workflow
	for cursor.HasMore() {
		var workflow Workflow
		_, err := cursor.ReadDocument(ctx, &workflow)
		if err != nil {
			r.logger.WithError(err).Warn("Failed to read workflow document")
			continue
		}
		workflows = append(workflows, &workflow)
	}

	return workflows, nil
}

// SearchWorkflows searches workflows by name, description, or tags
func (r *Repository) SearchWorkflows(ctx context.Context, searchTerm string, limit int) ([]*Workflow, error) {

	// Prepare search term for AQL
	searchPattern := fmt.Sprintf("%%%s%%", strings.ToLower(searchTerm))

	query := `
		FOR w IN @@collection
		FILTER LIKE(LOWER(w.name), @pattern, true) OR 
		       LIKE(LOWER(w.description), @pattern, true) OR
		       LENGTH(
		           FOR tag IN w.tags
		           FILTER LIKE(LOWER(tag), @pattern, true)
		           RETURN tag
		       ) > 0
		SORT w.created_at DESC
		LIMIT @limit
		RETURN w
	`

	bindVars := map[string]interface{}{
		"@collection": r.workflowsCollection.Name(),
		"pattern":     searchPattern,
		"limit":       limit,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to search workflows: %w", err)
	}
	defer cursor.Close()

	var workflows []*Workflow
	for cursor.HasMore() {
		var workflow Workflow
		_, err := cursor.ReadDocument(ctx, &workflow)
		if err != nil {
			r.logger.WithError(err).Warn("Failed to read workflow document")
			continue
		}
		workflows = append(workflows, &workflow)
	}

	return workflows, nil
}

// Execution CRUD operations

// CreateExecution stores a new workflow execution
func (r *Repository) CreateExecution(ctx context.Context, execution *WorkflowExecution) error {

	// Insert into database
	_, err := r.executionsCollection.CreateDocument(ctx, execution)
	if err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	r.logger.WithField("execution_id", execution.ID).Info("Execution created successfully")
	return nil
}

// GetExecution retrieves an execution by ID
func (r *Repository) GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error) {

	var execution WorkflowExecution
	_, err := r.executionsCollection.ReadDocument(ctx, executionID, &execution)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("execution %s not found", executionID)
		}
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	return &execution, nil
}

// UpdateExecution updates an existing execution
func (r *Repository) UpdateExecution(ctx context.Context, execution *WorkflowExecution) error {

	// Update document
	_, err := r.executionsCollection.UpdateDocument(ctx, execution.ID, execution)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("execution %s not found", execution.ID)
		}
		return fmt.Errorf("failed to update execution: %w", err)
	}

	return nil
}

// DeleteExecution removes an execution
func (r *Repository) DeleteExecution(ctx context.Context, executionID string) error {

	_, err := r.executionsCollection.RemoveDocument(ctx, executionID)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("execution %s not found", executionID)
		}
		return fmt.Errorf("failed to delete execution: %w", err)
	}

	r.logger.WithField("execution_id", executionID).Info("Execution deleted successfully")
	return nil
}

// ListExecutions retrieves executions with pagination and filtering
func (r *Repository) ListExecutions(ctx context.Context, workflowID string, status WorkflowStatus, limit, offset int) ([]*WorkflowExecution, error) {

	// Build query with optional filters
	var filters []string
	bindVars := map[string]interface{}{
		"@collection": r.executionsCollection.Name(),
		"limit":       limit,
		"offset":      offset,
	}

	if workflowID != "" {
		filters = append(filters, "e.workflow_id == @workflow_id")
		bindVars["workflow_id"] = workflowID
	}

	if status != "" {
		filters = append(filters, "e.status == @status")
		bindVars["status"] = string(status)
	}

	filterClause := ""
	if len(filters) > 0 {
		filterClause = "FILTER " + strings.Join(filters, " AND ")
	}

	query := fmt.Sprintf(`
		FOR e IN @@collection
		%s
		SORT e.start_time DESC
		LIMIT @offset, @limit
		RETURN e
	`, filterClause)

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}
	defer cursor.Close()

	var executions []*WorkflowExecution
	for cursor.HasMore() {
		var execution WorkflowExecution
		_, err := cursor.ReadDocument(ctx, &execution)
		if err != nil {
			r.logger.WithError(err).Warn("Failed to read execution document")
			continue
		}
		executions = append(executions, &execution)
	}

	return executions, nil
}

// GetExecutionsByStatus retrieves executions by status
func (r *Repository) GetExecutionsByStatus(ctx context.Context, status WorkflowStatus) ([]*WorkflowExecution, error) {

	query := `
		FOR e IN @@collection
		FILTER e.status == @status
		SORT e.start_time DESC
		RETURN e
	`

	bindVars := map[string]interface{}{
		"@collection": r.executionsCollection.Name(),
		"status":      string(status),
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query executions by status: %w", err)
	}
	defer cursor.Close()

	var executions []*WorkflowExecution
	for cursor.HasMore() {
		var execution WorkflowExecution
		_, err := cursor.ReadDocument(ctx, &execution)
		if err != nil {
			r.logger.WithError(err).Warn("Failed to read execution document")
			continue
		}
		executions = append(executions, &execution)
	}

	return executions, nil
}

// GetActiveExecutions retrieves all running executions
func (r *Repository) GetActiveExecutions(ctx context.Context) ([]*WorkflowExecution, error) {
	return r.GetExecutionsByStatus(ctx, WorkflowStatusRunning)
}

// Repository statistics and analytics

// GetWorkflowStats returns statistics for a workflow
func (r *Repository) GetWorkflowStats(ctx context.Context, workflowID string) (*WorkflowStats, error) {

	query := `
		LET executions = (
			FOR e IN @@executions_collection
			FILTER e.workflow_id == @workflow_id
			RETURN e
		)
		
		RETURN {
			total_executions: LENGTH(executions),
			completed_executions: LENGTH(
				FOR e IN executions
				FILTER e.status == "completed"
				RETURN e
			),
			failed_executions: LENGTH(
				FOR e IN executions
				FILTER e.status == "failed"
				RETURN e
			),
			running_executions: LENGTH(
				FOR e IN executions
				FILTER e.status == "running"
				RETURN e
			),
			average_duration: AVERAGE(
				FOR e IN executions
				FILTER e.status == "completed" AND e.end_time != null
				RETURN DATE_DIFF(e.start_time, e.end_time, "s")
			),
			last_execution: FIRST(
				FOR e IN executions
				SORT e.start_time DESC
				LIMIT 1
				RETURN e.start_time
			)
		}
	`

	bindVars := map[string]interface{}{
		"@executions_collection": r.executionsCollection.Name(),
		"workflow_id":            workflowID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow stats: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return &WorkflowStats{}, nil
	}

	var result map[string]interface{}
	_, err = cursor.ReadDocument(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to read stats result: %w", err)
	}

	// Convert result to WorkflowStats
	stats := &WorkflowStats{
		WorkflowID:          workflowID,
		TotalExecutions:     getIntFromInterface(result["total_executions"]),
		CompletedExecutions: getIntFromInterface(result["completed_executions"]),
		FailedExecutions:    getIntFromInterface(result["failed_executions"]),
		RunningExecutions:   getIntFromInterface(result["running_executions"]),
		AverageDuration:     time.Duration(getFloat64FromInterface(result["average_duration"])) * time.Second,
	}

	if lastExec := result["last_execution"]; lastExec != nil {
		if lastExecTime, ok := lastExec.(time.Time); ok {
			stats.LastExecution = &lastExecTime
		}
	}

	return stats, nil
}

// WorkflowStats represents statistics for a workflow
type WorkflowStats struct {
	WorkflowID          string        `json:"workflow_id"`
	TotalExecutions     int           `json:"total_executions"`
	CompletedExecutions int           `json:"completed_executions"`
	FailedExecutions    int           `json:"failed_executions"`
	RunningExecutions   int           `json:"running_executions"`
	AverageDuration     time.Duration `json:"average_duration"`
	LastExecution       *time.Time    `json:"last_execution,omitempty"`
}

// Helper methods

func (r *Repository) initializeCollections(config RepositoryConfig) error {
	ctx := context.Background()
	var err error

	// Initialize workflows collection
	r.workflowsCollection, err = r.db.Collection(ctx, config.WorkflowsCollection)
	if err != nil {
		// Collection doesn't exist, create it
		r.workflowsCollection, err = r.db.CreateCollection(ctx, config.WorkflowsCollection, nil)
		if err != nil {
			return fmt.Errorf("failed to create workflows collection: %w", err)
		}
		r.logger.WithField("collection", config.WorkflowsCollection).Info("Created workflows collection")
	}

	// Initialize executions collection
	r.executionsCollection, err = r.db.Collection(ctx, config.ExecutionsCollection)
	if err != nil {
		// Collection doesn't exist, create it
		r.executionsCollection, err = r.db.CreateCollection(ctx, config.ExecutionsCollection, nil)
		if err != nil {
			return fmt.Errorf("failed to create executions collection: %w", err)
		}
		r.logger.WithField("collection", config.ExecutionsCollection).Info("Created executions collection")
	}

	return nil
}

func (r *Repository) createIndexes() error {
	ctx := context.Background()

	// Workflow indexes
	workflowIndexes := []map[string]interface{}{
		{
			"type":   "persistent",
			"fields": []string{"name", "version"},
			"unique": true,
		},
		{
			"type":   "persistent",
			"fields": []string{"created_at"},
		},
		{
			"type":   "persistent",
			"fields": []string{"tags[*]"},
		},
	}

	for _, indexDef := range workflowIndexes {
		options := &driver.EnsurePersistentIndexOptions{
			Unique: indexDef["unique"] != nil && indexDef["unique"].(bool),
		}
		_, _, err := r.workflowsCollection.EnsurePersistentIndex(ctx, indexDef["fields"].([]string), options)
		if err != nil {
			r.logger.WithError(err).WithField("fields", indexDef["fields"]).Warn("Failed to create workflow index")
		}
	}

	// Execution indexes
	executionIndexes := []map[string]interface{}{
		{
			"type":   "persistent",
			"fields": []string{"workflow_id"},
		},
		{
			"type":   "persistent",
			"fields": []string{"status"},
		},
		{
			"type":   "persistent",
			"fields": []string{"start_time"},
		},
		{
			"type":   "persistent",
			"fields": []string{"workflow_id", "status"},
		},
	}

	for _, indexDef := range executionIndexes {
		options := &driver.EnsurePersistentIndexOptions{}
		_, _, err := r.executionsCollection.EnsurePersistentIndex(ctx, indexDef["fields"].([]string), options)
		if err != nil {
			r.logger.WithError(err).WithField("fields", indexDef["fields"]).Warn("Failed to create execution index")
		}
	}

	r.logger.Info("Database indexes created successfully")
	return nil
}

// Utility functions for type conversion
func getIntFromInterface(val interface{}) int {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
	}
	return 0
}

func getFloat64FromInterface(val interface{}) float64 {
	if val == nil {
		return 0.0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f
		}
	}
	return 0.0
}
