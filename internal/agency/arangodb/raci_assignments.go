package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

const (
	// RACIAssignmentCollectionName is the name of the RACI assignments edge collection
	RACIAssignmentCollectionName = "raci_assignments"
)

// ensureRACIAssignmentCollection ensures the RACI assignments edge collection exists
func ensureRACIAssignmentCollection(db driver.Database) (driver.Collection, error) {
	ctx := context.Background()

	// Check if collection exists
	exists, err := db.CollectionExists(ctx, RACIAssignmentCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if !exists {
		// Create edge collection
		options := &driver.CreateCollectionOptions{
			Type: driver.CollectionTypeEdge,
		}
		collection, err = db.CreateCollection(ctx, RACIAssignmentCollectionName, options)
		if err != nil {
			return nil, fmt.Errorf("failed to create edge collection: %w", err)
		}
	} else {
		collection, err = db.Collection(ctx, RACIAssignmentCollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	}

	// Ensure indexes
	if err := ensureRACIAssignmentIndexes(ctx, collection); err != nil {
		return nil, fmt.Errorf("failed to ensure indexes: %w", err)
	}

	return collection, nil
}

// ensureRACIAssignmentIndexes creates necessary indexes on the RACI assignments collection
func ensureRACIAssignmentIndexes(ctx context.Context, collection driver.Collection) error {
	// Index on work_item_key for faster lookups
	_, _, err := collection.EnsurePersistentIndex(ctx, []string{"work_item_key"}, &driver.EnsurePersistentIndexOptions{
		Unique: false,
		Name:   "idx_work_item_key",
	})
	if err != nil {
		return fmt.Errorf("failed to create work_item_key index: %w", err)
	}

	// Index on role_key for faster lookups
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"role_key"}, &driver.EnsurePersistentIndexOptions{
		Unique: false,
		Name:   "idx_role_key",
	})
	if err != nil {
		return fmt.Errorf("failed to create role_key index: %w", err)
	}

	// Unique index on work_item_key + role_key to prevent duplicates
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"work_item_key", "role_key"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
		Name:   "idx_work_item_role_unique",
	})
	if err != nil {
		return fmt.Errorf("failed to create unique work_item_role index: %w", err)
	}

	return nil
}

// CreateRACIAssignment creates a new RACI assignment edge
func (r *Repository) CreateRACIAssignment(ctx context.Context, agencyID string, assignment *models.RACIAssignment) error {
	// Get the agency document
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

	// Ensure RACI assignments edge collection exists
	collection, err := ensureRACIAssignmentCollection(agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure RACI assignments collection: %w", err)
	}

	// Set timestamps
	now := time.Now()
	assignment.CreatedAt = now
	assignment.UpdatedAt = now

	// Build _from and _to references
	// _from: work_items/{work_item_key}
	// _to: roles/{role_key}
	assignment.From = fmt.Sprintf("work_items/%s", assignment.WorkItemKey)
	assignment.To = fmt.Sprintf("roles/%s", assignment.RoleKey)

	// Create the edge document
	meta, err := collection.CreateDocument(ctx, assignment)
	if err != nil {
		return fmt.Errorf("failed to create RACI assignment edge: %w", err)
	}

	assignment.Key = meta.Key
	assignment.ID = meta.ID.String()

	return nil
}

// GetRACIAssignmentsForWorkItem retrieves all RACI assignments for a work item
func (r *Repository) GetRACIAssignmentsForWorkItem(ctx context.Context, agencyID string, workItemKey string) ([]*models.RACIAssignment, error) {
	// Get the agency document
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

	// Check if collection exists - if not, return empty array
	exists, err := agencyDB.CollectionExists(ctx, RACIAssignmentCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !exists {
		// No assignments yet - return empty array
		return []*models.RACIAssignment{}, nil
	}

	// Get collection
	collection, err := agencyDB.Collection(ctx, RACIAssignmentCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Query assignments for this work item
	query := `
		FOR assignment IN @@collection
		FILTER assignment.work_item_key == @workItemKey
		SORT assignment.created_at DESC
		RETURN assignment
	`
	bindVars := map[string]interface{}{
		"@collection": collection.Name(),
		"workItemKey": workItemKey,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query RACI assignments: %w", err)
	}
	defer cursor.Close()

	var assignments []*models.RACIAssignment
	for cursor.HasMore() {
		var assignment models.RACIAssignment
		_, err := cursor.ReadDocument(ctx, &assignment)
		if err != nil {
			return nil, fmt.Errorf("failed to read assignment: %w", err)
		}
		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

// GetRACIAssignmentsForRole retrieves all RACI assignments for a role
func (r *Repository) GetRACIAssignmentsForRole(ctx context.Context, agencyID string, roleKey string) ([]*models.RACIAssignment, error) {
	// Get the agency document
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

	// Check if collection exists - if not, return empty array
	exists, err := agencyDB.CollectionExists(ctx, RACIAssignmentCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !exists {
		// No assignments yet - return empty array
		return []*models.RACIAssignment{}, nil
	}

	// Get collection
	collection, err := agencyDB.Collection(ctx, RACIAssignmentCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Query assignments for this role
	query := `
		FOR assignment IN @@collection
		FILTER assignment.role_key == @roleKey
		SORT assignment.work_item_key ASC
		RETURN assignment
	`
	bindVars := map[string]interface{}{
		"@collection": collection.Name(),
		"roleKey":     roleKey,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query RACI assignments: %w", err)
	}
	defer cursor.Close()

	var assignments []*models.RACIAssignment
	for cursor.HasMore() {
		var assignment models.RACIAssignment
		_, err := cursor.ReadDocument(ctx, &assignment)
		if err != nil {
			return nil, fmt.Errorf("failed to read assignment: %w", err)
		}
		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

// GetAllRACIAssignments retrieves all RACI assignments for an agency
func (r *Repository) GetAllRACIAssignments(ctx context.Context, agencyID string) ([]*models.RACIAssignment, error) {
	// Get the agency document
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

	// Check if collection exists - if not, return empty array
	exists, err := agencyDB.CollectionExists(ctx, RACIAssignmentCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !exists {
		// No assignments yet - return empty array
		return []*models.RACIAssignment{}, nil
	}

	// Get collection
	collection, err := agencyDB.Collection(ctx, RACIAssignmentCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Query all assignments for this agency
	query := `
		FOR assignment IN @@collection
		SORT assignment.work_item_key ASC, assignment.role_key ASC
		RETURN assignment
	`
	bindVars := map[string]interface{}{
		"@collection": collection.Name(),
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query RACI assignments: %w", err)
	}
	defer cursor.Close()

	var assignments []*models.RACIAssignment
	for cursor.HasMore() {
		var assignment models.RACIAssignment
		_, err := cursor.ReadDocument(ctx, &assignment)
		if err != nil {
			return nil, fmt.Errorf("failed to read assignment: %w", err)
		}
		assignments = append(assignments, &assignment)
	}

	return assignments, nil
}

// UpdateRACIAssignment updates an existing RACI assignment
func (r *Repository) UpdateRACIAssignment(ctx context.Context, agencyID string, key string, assignment *models.RACIAssignment) error {
	// Get agency-specific database
	agencyDB, err := r.getAgencyDatabase(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency database: %w", err)
	}

	// Ensure collection exists
	collection, err := ensureRACIAssignmentCollection(agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure RACI assignments collection: %w", err)
	}

	// Update timestamp
	assignment.UpdatedAt = time.Now()

	// Update the document
	_, err = collection.UpdateDocument(ctx, key, assignment)
	if err != nil {
		return fmt.Errorf("failed to update RACI assignment: %w", err)
	}

	return nil
}

// DeleteRACIAssignment deletes a RACI assignment by key
func (r *Repository) DeleteRACIAssignment(ctx context.Context, agencyID string, key string) error {
	// Get agency-specific database
	agencyDB, err := r.getAgencyDatabase(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency database: %w", err)
	}

	// Ensure collection exists
	collection, err := ensureRACIAssignmentCollection(agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure RACI assignments collection: %w", err)
	}

	// Delete the document
	_, err = collection.RemoveDocument(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete RACI assignment: %w", err)
	}

	return nil
}

// DeleteRACIAssignmentsForWorkItem deletes all RACI assignments for a work item
func (r *Repository) DeleteRACIAssignmentsForWorkItem(ctx context.Context, agencyID string, workItemKey string) error {
	// Get agency-specific database
	agencyDB, err := r.getAgencyDatabase(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency database: %w", err)
	}

	// Ensure collection exists
	collection, err := ensureRACIAssignmentCollection(agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure RACI assignments collection: %w", err)
	}

	// Delete all assignments for this work item
	query := `
		FOR assignment IN @@collection
		FILTER assignment.agency_id == @agencyId
		AND assignment.work_item_key == @workItemKey
		REMOVE assignment IN @@collection
	`
	bindVars := map[string]interface{}{
		"@collection": collection.Name(),
		"agencyId":    agencyID,
		"workItemKey": workItemKey,
	}

	_, err = agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to delete RACI assignments: %w", err)
	}

	return nil
}
