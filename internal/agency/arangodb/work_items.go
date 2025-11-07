package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

// CreateWorkItem creates a new work item in the agency's work_items collection
func (r *Repository) CreateWorkItem(ctx context.Context, workItem *models.WorkItem) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, workItem.AgencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure work_items collection exists
	workItemsColl, err := ensureWorkItemsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure work_items collection: %w", err)
	}

	// Set timestamps
	now := time.Now()
	workItem.CreatedAt = now
	workItem.UpdatedAt = now

	// If number is not set, get the next number
	if workItem.Number == 0 {
		// Count existing work items
		query := "FOR wi IN work_items FILTER wi.agency_id == @agencyId COLLECT WITH COUNT INTO length RETURN length"
		cursor, err := agencyDB.Query(ctx, query, map[string]interface{}{"agencyId": workItem.AgencyID})
		if err != nil {
			return fmt.Errorf("failed to count work items: %w", err)
		}
		defer cursor.Close()

		var count int
		if cursor.HasMore() {
			_, err := cursor.ReadDocument(ctx, &count)
			if err != nil {
				return fmt.Errorf("failed to read count: %w", err)
			}
		}
		workItem.Number = count + 1
	}

	// Auto-generate code if not provided
	if workItem.Code == "" {
		workItem.Code = fmt.Sprintf("WI-%03d", workItem.Number)
	}

	// Initialize empty slices if nil
	if workItem.Deliverables == nil {
		workItem.Deliverables = []string{}
	}
	if workItem.Tags == nil {
		workItem.Tags = []string{}
	}

	// Create the document
	meta, err := workItemsColl.CreateDocument(ctx, workItem)
	if err != nil {
		return fmt.Errorf("failed to create work item: %w", err)
	}

	workItem.Key = meta.Key
	return nil
}

// GetWorkItems retrieves all work items for an agency
func (r *Repository) GetWorkItems(ctx context.Context, agencyID string) ([]*models.WorkItem, error) {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure work_items collection exists
	workItemsColl, err := ensureWorkItemsCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure work_items collection: %w", err)
	}

	// Query all work items for this agency, ordered by number
	query := "FOR wi IN @@collection FILTER wi.agency_id == @agencyId SORT wi.number ASC RETURN wi"
	bindVars := map[string]interface{}{
		"@collection": workItemsColl.Name(),
		"agencyId":    agencyID,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query work items: %w", err)
	}
	defer cursor.Close()

	var workItems []*models.WorkItem
	for cursor.HasMore() {
		var workItem models.WorkItem
		_, err := cursor.ReadDocument(ctx, &workItem)
		if err != nil {
			return nil, fmt.Errorf("failed to read work item: %w", err)
		}
		workItems = append(workItems, &workItem)
	}

	return workItems, nil
}

// GetWorkItem retrieves a single work item by key
func (r *Repository) GetWorkItem(ctx context.Context, agencyID string, key string) (*models.WorkItem, error) {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure work_items collection exists
	workItemsColl, err := ensureWorkItemsCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure work_items collection: %w", err)
	}

	// Read the document
	var workItem models.WorkItem
	_, err = workItemsColl.ReadDocument(ctx, key, &workItem)
	if err != nil {
		return nil, fmt.Errorf("failed to read work item: %w", err)
	}

	return &workItem, nil
}

// GetWorkItemByCode retrieves a single work item by code
func (r *Repository) GetWorkItemByCode(ctx context.Context, agencyID string, code string) (*models.WorkItem, error) {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure work_items collection exists
	workItemsColl, err := ensureWorkItemsCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure work_items collection: %w", err)
	}

	// Query for the work item by code
	query := "FOR wi IN @@collection FILTER wi.agency_id == @agencyId AND wi.code == @code LIMIT 1 RETURN wi"
	bindVars := map[string]interface{}{
		"@collection": workItemsColl.Name(),
		"agencyId":    agencyID,
		"code":        code,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query work item: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("work item not found: %s", code)
	}

	var workItem models.WorkItem
	_, err = cursor.ReadDocument(ctx, &workItem)
	if err != nil {
		return nil, fmt.Errorf("failed to read work item: %w", err)
	}

	return &workItem, nil
}

// UpdateWorkItem updates an existing work item
func (r *Repository) UpdateWorkItem(ctx context.Context, workItem *models.WorkItem) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, workItem.AgencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure work_items collection exists
	workItemsColl, err := ensureWorkItemsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure work_items collection: %w", err)
	}

	// Update timestamp
	workItem.UpdatedAt = time.Now()

	// Update the document
	_, err = workItemsColl.UpdateDocument(ctx, workItem.Key, workItem)
	if err != nil {
		return fmt.Errorf("failed to update work item: %w", err)
	}

	return nil
}

// DeleteWorkItem deletes a work item and renumbers remaining work items
func (r *Repository) DeleteWorkItem(ctx context.Context, agencyID string, key string) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure work_items collection exists
	workItemsColl, err := ensureWorkItemsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure work_items collection: %w", err)
	}

	// Get the work item to find its number and code
	var workItemToDelete models.WorkItem
	_, err = workItemsColl.ReadDocument(ctx, key, &workItemToDelete)
	if err != nil {
		return fmt.Errorf("failed to read work item: %w", err)
	}

	// Check if any other work items depend on this one
	dependencyCheckQuery := `
		FOR wi IN @@collection 
		FILTER wi.agency_id == @agencyId AND @code IN wi.dependencies
		RETURN wi.code
	`
	dependencyBindVars := map[string]interface{}{
		"@collection": workItemsColl.Name(),
		"agencyId":    agencyID,
		"code":        workItemToDelete.Code,
	}

	dependencyCursor, err := agencyDB.Query(ctx, dependencyCheckQuery, dependencyBindVars)
	if err != nil {
		return fmt.Errorf("failed to check dependencies: %w", err)
	}
	defer dependencyCursor.Close()

	if dependencyCursor.HasMore() {
		var dependentCodes []string
		for dependencyCursor.HasMore() {
			var code string
			_, err := dependencyCursor.ReadDocument(ctx, &code)
			if err != nil {
				return fmt.Errorf("failed to read dependent code: %w", err)
			}
			dependentCodes = append(dependentCodes, code)
		}
		return fmt.Errorf("cannot delete work item %s: it is a dependency for %v", workItemToDelete.Code, dependentCodes)
	}

	// Delete the document
	_, err = workItemsColl.RemoveDocument(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete work item: %w", err)
	}

	// Renumber work items with higher numbers
	renumberQuery := `
		FOR wi IN @@collection 
		FILTER wi.agency_id == @agencyId AND wi.number > @deletedNumber
		UPDATE wi WITH { number: wi.number - 1 } IN @@collection
	`
	renumberBindVars := map[string]interface{}{
		"@collection":   workItemsColl.Name(),
		"agencyId":      agencyID,
		"deletedNumber": workItemToDelete.Number,
	}

	_, err = agencyDB.Query(ctx, renumberQuery, renumberBindVars)
	if err != nil {
		return fmt.Errorf("failed to renumber work items: %w", err)
	}

	return nil
}

// ensureWorkItemsCollection ensures the work_items collection exists in an agency database
func ensureWorkItemsCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	const collectionName = "work_items"

	// Check if collection exists
	exists, err := db.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if !exists {
		// Create collection
		collection, err = db.CreateCollection(ctx, collectionName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
	} else {
		collection, err = db.Collection(ctx, collectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	}

	return collection, nil
}
