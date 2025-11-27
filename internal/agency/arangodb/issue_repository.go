package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

const (
	// IssueCollectionName is the name of the work_issues collection
	IssueCollectionName = "work_issues"
)

// IssueRepository implements agency.IssueRepository using ArangoDB
type IssueRepository struct {
	client     driver.Client
	db         driver.Database
	collection driver.Collection
}

// NewIssueRepository creates a new ArangoDB repository for work issues
func NewIssueRepository(client driver.Client, db driver.Database) (agency.IssueRepository, error) {
	// Ensure work_issues collection exists
	collection, err := ensureIssueCollection(db)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure work_issues collection: %w", err)
	}

	return &IssueRepository{
		client:     client,
		db:         db,
		collection: collection,
	}, nil
}

// ensureIssueCollection ensures the work_issues collection exists with proper indexes
func ensureIssueCollection(db driver.Database) (driver.Collection, error) {
	ctx := context.Background()

	// Check if collection exists
	exists, err := db.CollectionExists(ctx, IssueCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if !exists {
		// Create collection
		collection, err = db.CreateCollection(ctx, IssueCollectionName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
	} else {
		collection, err = db.Collection(ctx, IssueCollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	}

	// Ensure indexes
	if err := ensureIssueIndexes(ctx, collection); err != nil {
		return nil, fmt.Errorf("failed to ensure indexes: %w", err)
	}

	return collection, nil
}

// ensureIssueIndexes creates necessary indexes on the work_issues collection
func ensureIssueIndexes(ctx context.Context, collection driver.Collection) error {
	// Composite index on agency_id + instance_id
	_, _, err := collection.EnsurePersistentIndex(ctx, []string{"agency_id", "instance_id"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create agency-instance index: %w", err)
	}

	// Index on workflow_id
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"workflow_id"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create workflow index: %w", err)
	}

	// Composite index on current_step + status (for board queries)
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"current_step", "status"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create step-status index: %w", err)
	}

	// Index on assigned_to
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"assigned_to"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create assigned_to index: %w", err)
	}

	// Index on created_at for sorting
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"created_at"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create created_at index: %w", err)
	}

	// Composite index on instance_id + number (for unique issue numbers per instance)
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"instance_id", "number"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create instance-number unique index: %w", err)
	}

	return nil
}

// Create creates a new work issue
func (r *IssueRepository) Create(ctx context.Context, issue *models.WorkIssue) error {
	// Set timestamps
	now := time.Now()
	issue.CreatedAt = now
	issue.UpdatedAt = now

	// Create document
	meta, err := r.collection.CreateDocument(ctx, issue)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	issue.Key = meta.Key
	issue.ID = meta.ID.String()
	issue.Rev = meta.Rev

	return nil
}

// GetByID retrieves an issue by its ID
func (r *IssueRepository) GetByID(ctx context.Context, agencyID, instanceID, issueID string) (*models.WorkIssue, error) {
	var issue models.WorkIssue

	query := `
		FOR issue IN @@collection
			FILTER issue.agency_id == @agency_id
			AND issue.instance_id == @instance_id
			AND issue._key == @issue_id
			RETURN issue
	`

	bindVars := map[string]interface{}{
		"@collection": IssueCollectionName,
		"agency_id":   agencyID,
		"instance_id": instanceID,
		"issue_id":    issueID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query issue: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("issue not found")
	}

	_, err = cursor.ReadDocument(ctx, &issue)
	if err != nil {
		return nil, fmt.Errorf("failed to read issue: %w", err)
	}

	return &issue, nil
}

// GetByKey retrieves an issue by its key
func (r *IssueRepository) GetByKey(ctx context.Context, agencyID, key string) (*models.WorkIssue, error) {
	var issue models.WorkIssue

	_, err := r.collection.ReadDocument(ctx, key, &issue)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("issue not found")
		}
		return nil, fmt.Errorf("failed to read issue: %w", err)
	}

	// Verify agency_id matches
	if issue.AgencyID != agencyID {
		return nil, fmt.Errorf("issue not found")
	}

	return &issue, nil
}

// Update updates an existing issue
func (r *IssueRepository) Update(ctx context.Context, issue *models.WorkIssue) error {
	// Update timestamp
	issue.UpdatedAt = time.Now()

	meta, err := r.collection.UpdateDocument(ctx, issue.Key, issue)
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}

	issue.Rev = meta.Rev
	return nil
}

// Delete deletes an issue
func (r *IssueRepository) Delete(ctx context.Context, agencyID, instanceID, issueID string) error {
	// First verify the issue exists and belongs to the agency/instance
	_, err := r.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		return err
	}

	_, err = r.collection.RemoveDocument(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}

	return nil
}

// List retrieves issues with filters
func (r *IssueRepository) List(ctx context.Context, agencyID, instanceID string, filters agency.IssueFilters) ([]*models.WorkIssue, error) {
	query := `
		FOR issue IN @@collection
			FILTER issue.agency_id == @agency_id
			AND issue.instance_id == @instance_id
	`

	bindVars := map[string]interface{}{
		"@collection": IssueCollectionName,
		"agency_id":   agencyID,
		"instance_id": instanceID,
	}

	// Add optional filters
	if filters.WorkflowID != "" {
		query += " AND issue.workflow_id == @workflow_id"
		bindVars["workflow_id"] = filters.WorkflowID
	}

	if filters.Status != "" {
		query += " AND issue.status == @status"
		bindVars["status"] = filters.Status
	}

	if filters.AssignedTo != "" {
		query += " AND issue.assigned_to == @assigned_to"
		bindVars["assigned_to"] = filters.AssignedTo
	}

	// Sort by creation date (newest first)
	query += " SORT issue.created_at DESC"

	// Pagination
	if filters.Limit > 0 {
		query += " LIMIT @offset, @limit"
		bindVars["offset"] = filters.Offset
		bindVars["limit"] = filters.Limit
	}

	query += " RETURN issue"

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer cursor.Close()

	var issues []*models.WorkIssue
	for cursor.HasMore() {
		var issue models.WorkIssue
		_, err := cursor.ReadDocument(ctx, &issue)
		if err != nil {
			return nil, fmt.Errorf("failed to read issue: %w", err)
		}
		issues = append(issues, &issue)
	}

	return issues, nil
}

// ListByWorkflowStep retrieves all issues at a specific workflow step
func (r *IssueRepository) ListByWorkflowStep(ctx context.Context, agencyID, instanceID, workflowID, step string) ([]*models.WorkIssue, error) {
	query := `
		FOR issue IN @@collection
			FILTER issue.agency_id == @agency_id
			AND issue.instance_id == @instance_id
			AND issue.workflow_id == @workflow_id
			AND issue.current_step == @step
			SORT issue.created_at ASC
			RETURN issue
	`

	bindVars := map[string]interface{}{
		"@collection": IssueCollectionName,
		"agency_id":   agencyID,
		"instance_id": instanceID,
		"workflow_id": workflowID,
		"step":        step,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues by step: %w", err)
	}
	defer cursor.Close()

	var issues []*models.WorkIssue
	for cursor.HasMore() {
		var issue models.WorkIssue
		_, err := cursor.ReadDocument(ctx, &issue)
		if err != nil {
			return nil, fmt.Errorf("failed to read issue: %w", err)
		}
		issues = append(issues, &issue)
	}

	return issues, nil
}

// ListAvailable retrieves all available (unassigned) issues at a specific step
func (r *IssueRepository) ListAvailable(ctx context.Context, agencyID, instanceID string, step string) ([]*models.WorkIssue, error) {
	query := `
		FOR issue IN @@collection
			FILTER issue.agency_id == @agency_id
			AND issue.instance_id == @instance_id
			AND issue.status == @status
	`

	bindVars := map[string]interface{}{
		"@collection": IssueCollectionName,
		"agency_id":   agencyID,
		"instance_id": instanceID,
		"status":      models.IssueStatusOpen,
	}

	// Add step filter if specified
	if step != "" {
		query += " AND issue.current_step == @step"
		bindVars["step"] = step
	}

	query += " SORT issue.created_at ASC RETURN issue"

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query available issues: %w", err)
	}
	defer cursor.Close()

	var issues []*models.WorkIssue
	for cursor.HasMore() {
		var issue models.WorkIssue
		_, err := cursor.ReadDocument(ctx, &issue)
		if err != nil {
			return nil, fmt.Errorf("failed to read issue: %w", err)
		}
		issues = append(issues, &issue)
	}

	return issues, nil
}

// CountByStep counts issues at a specific workflow step
func (r *IssueRepository) CountByStep(ctx context.Context, agencyID, instanceID, workflowID, step string) (int, error) {
	query := `
		FOR issue IN @@collection
			FILTER issue.agency_id == @agency_id
			AND issue.instance_id == @instance_id
			AND issue.workflow_id == @workflow_id
			AND issue.current_step == @step
			COLLECT WITH COUNT INTO count
			RETURN count
	`

	bindVars := map[string]interface{}{
		"@collection": IssueCollectionName,
		"agency_id":   agencyID,
		"instance_id": instanceID,
		"workflow_id": workflowID,
		"step":        step,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return 0, fmt.Errorf("failed to count issues: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return 0, nil
	}

	var count int
	_, err = cursor.ReadDocument(ctx, &count)
	if err != nil {
		return 0, fmt.Errorf("failed to read count: %w", err)
	}

	return count, nil
}

// GetNextIssueNumber gets the next available issue number for an instance
func (r *IssueRepository) GetNextIssueNumber(ctx context.Context, agencyID, instanceID string) (int, error) {
	query := `
		FOR issue IN @@collection
			FILTER issue.agency_id == @agency_id
			AND issue.instance_id == @instance_id
			SORT issue.number DESC
			LIMIT 1
			RETURN issue.number
	`

	bindVars := map[string]interface{}{
		"@collection": IssueCollectionName,
		"agency_id":   agencyID,
		"instance_id": instanceID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return 0, fmt.Errorf("failed to query max issue number: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		// No issues yet, start at 1
		return 1, nil
	}

	var maxNumber int
	_, err = cursor.ReadDocument(ctx, &maxNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to read max number: %w", err)
	}

	return maxNumber + 1, nil
}

// Exists checks if an issue exists
func (r *IssueRepository) Exists(ctx context.Context, agencyID, instanceID, issueID string) (bool, error) {
	_, err := r.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		if err.Error() == "issue not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
