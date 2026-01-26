package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

const (
	// WorkItemsCollectionName is the collection name for work items
	WorkItemsCollectionName = "work_items"
)

// WorkItemRepository handles work item data access
type WorkItemRepository struct {
	db         driver.Database
	collection driver.Collection
}

// NewWorkItemRepository creates a new work item repository
func NewWorkItemRepository(db driver.Database) (*WorkItemRepository, error) {
	collection, err := ensureCollection(db, WorkItemsCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure work_items collection: %w", err)
	}

	return &WorkItemRepository{
		db:         db,
		collection: collection,
	}, nil
}

// ListWorkItems retrieves all work items for an agency
func (r *WorkItemRepository) ListWorkItems(ctx context.Context, agencyID string, filters map[string]interface{}) ([]*models.WorkItem, error) {
	query := `
		FOR wi IN work_items
			FILTER wi.agency_id == @agency_id
	`
	bindVars := map[string]interface{}{
		"agency_id": agencyID,
	}

	// Add optional filters
	if typeFilter, ok := filters["type"].(string); ok && typeFilter != "" {
		query += ` FILTER wi.type == @type`
		bindVars["type"] = typeFilter
	}

	if statusFilter, ok := filters["status"].(string); ok && statusFilter != "" {
		query += ` FILTER wi.status == @status`
		bindVars["status"] = statusFilter
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		query += ` FILTER CONTAINS(LOWER(wi.title), LOWER(@search)) OR CONTAINS(LOWER(wi.description), LOWER(@search))`
		bindVars["search"] = search
	}

	// Add sorting and return
	query += `
			SORT wi.created_at DESC
			RETURN wi
	`

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query work items: %w", err)
	}
	defer cursor.Close()

	var workItems []*models.WorkItem
	for {
		var wi models.WorkItem
		_, err := cursor.ReadDocument(ctx, &wi)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read work item: %w", err)
		}
		workItems = append(workItems, &wi)
	}

	return workItems, nil
}

// GetWorkItem retrieves a single work item by ID
func (r *WorkItemRepository) GetWorkItem(ctx context.Context, agencyID, workItemID string) (*models.WorkItem, error) {
	query := `
		FOR wi IN work_items
			FILTER wi._key == @key AND wi.agency_id == @agency_id
			RETURN wi
	`
	bindVars := map[string]interface{}{
		"key":       workItemID,
		"agency_id": agencyID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query work item: %w", err)
	}
	defer cursor.Close()

	var workItem models.WorkItem
	_, err = cursor.ReadDocument(ctx, &workItem)
	if driver.IsNoMoreDocuments(err) {
		return nil, fmt.Errorf("work item not found: %s", workItemID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read work item: %w", err)
	}

	return &workItem, nil
}

// CreateWorkItem creates a new work item
func (r *WorkItemRepository) CreateWorkItem(ctx context.Context, workItem *models.WorkItem) (*models.WorkItem, error) {
	now := time.Now()
	workItem.CreatedAt = now
	workItem.UpdatedAt = now

	// Get next work item number for the agency
	number, err := r.getNextWorkItemNumber(ctx, workItem.AgencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next work item number: %w", err)
	}
	workItem.Number = number

	meta, err := r.collection.CreateDocument(ctx, workItem)
	if err != nil {
		return nil, fmt.Errorf("failed to create work item: %w", err)
	}

	workItem.Key = meta.Key
	workItem.ID = meta.ID.String()

	return workItem, nil
}

// UpdateWorkItem updates an existing work item
func (r *WorkItemRepository) UpdateWorkItem(ctx context.Context, agencyID, workItemID string, updates *models.WorkItem) (*models.WorkItem, error) {
	// First, verify the work item exists and belongs to the agency
	existing, err := r.GetWorkItem(ctx, agencyID, workItemID)
	if err != nil {
		return nil, err
	}

	// Update timestamp
	updates.UpdatedAt = time.Now()
	updates.CreatedAt = existing.CreatedAt // Preserve original creation time
	updates.Number = existing.Number       // Preserve number
	updates.AgencyID = agencyID            // Ensure agency ID doesn't change
	updates.Key = workItemID               // Ensure key doesn't change

	_, err = r.collection.UpdateDocument(ctx, workItemID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update work item: %w", err)
	}

	return updates, nil
}

// DeleteWorkItem deletes a work item
func (r *WorkItemRepository) DeleteWorkItem(ctx context.Context, agencyID, workItemID string) error {
	// Verify the work item exists and belongs to the agency
	_, err := r.GetWorkItem(ctx, agencyID, workItemID)
	if err != nil {
		return err
	}

	_, err = r.collection.RemoveDocument(ctx, workItemID)
	if err != nil {
		return fmt.Errorf("failed to delete work item: %w", err)
	}

	return nil
}

// getNextWorkItemNumber gets the next sequential number for work items in an agency
func (r *WorkItemRepository) getNextWorkItemNumber(ctx context.Context, agencyID string) (int, error) {
	query := `
		FOR wi IN work_items
			FILTER wi.agency_id == @agency_id
			SORT wi.number DESC
			LIMIT 1
			RETURN wi.number
	`
	bindVars := map[string]interface{}{
		"agency_id": agencyID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return 0, fmt.Errorf("failed to query max work item number: %w", err)
	}
	defer cursor.Close()

	var maxNumber int
	_, err = cursor.ReadDocument(ctx, &maxNumber)
	if driver.IsNoMoreDocuments(err) {
		// No work items yet, start at 1
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to read max number: %w", err)
	}

	return maxNumber + 1, nil
}

// CountWorkItems counts work items for an agency with optional filters
func (r *WorkItemRepository) CountWorkItems(ctx context.Context, agencyID string, filters map[string]interface{}) (int, error) {
	query := `
		FOR wi IN work_items
			FILTER wi.agency_id == @agency_id
	`
	bindVars := map[string]interface{}{
		"agency_id": agencyID,
	}

	// Add optional filters (same as ListWorkItems)
	if typeFilter, ok := filters["type"].(string); ok && typeFilter != "" {
		query += ` FILTER wi.type == @type`
		bindVars["type"] = typeFilter
	}

	if statusFilter, ok := filters["status"].(string); ok && statusFilter != "" {
		query += ` FILTER wi.status == @status`
		bindVars["status"] = statusFilter
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		query += ` FILTER CONTAINS(LOWER(wi.title), LOWER(@search)) OR CONTAINS(LOWER(wi.description), LOWER(@search))`
		bindVars["search"] = search
	}

	query += ` COLLECT WITH COUNT INTO length RETURN length`

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return 0, fmt.Errorf("failed to count work items: %w", err)
	}
	defer cursor.Close()

	var count int
	_, err = cursor.ReadDocument(ctx, &count)
	if err != nil {
		return 0, fmt.Errorf("failed to read count: %w", err)
	}

	return count, nil
}
