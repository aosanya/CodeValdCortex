package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

const (
	// WorkItemGoalLinksCollectionName is the name of the work item-goal links edge collection
	WorkItemGoalLinksCollectionName = "work_item_goals"
)

// ensureWorkItemGoalLinksCollection ensures the work item-goal links edge collection exists
func ensureWorkItemGoalLinksCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	// Check if collection exists
	exists, err := db.CollectionExists(ctx, WorkItemGoalLinksCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if !exists {
		// Create edge collection
		options := &driver.CreateCollectionOptions{
			Type: driver.CollectionTypeEdge,
		}
		collection, err = db.CreateCollection(ctx, WorkItemGoalLinksCollectionName, options)
		if err != nil {
			return nil, fmt.Errorf("failed to create edge collection: %w", err)
		}
	} else {
		collection, err = db.Collection(ctx, WorkItemGoalLinksCollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	}

	// Ensure indexes for efficient queries
	if err := ensureWorkItemGoalLinksIndexes(ctx, collection); err != nil {
		return nil, fmt.Errorf("failed to ensure indexes: %w", err)
	}

	return collection, nil
}

// ensureWorkItemGoalLinksIndexes creates necessary indexes on the edge collection
func ensureWorkItemGoalLinksIndexes(ctx context.Context, collection driver.Collection) error {
	// Index on work_item_key for finding all goals for a work item
	_, _, err := collection.EnsurePersistentIndex(ctx, []string{"work_item_key"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create work_item_key index: %w", err)
	}

	// Index on goal_key for finding all work items for a goal
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"goal_key"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create goal_key index: %w", err)
	}

	// Index on relationship type
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"relationship"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create relationship index: %w", err)
	}

	return nil
}

// CreateWorkItemGoalLink creates a new edge linking a work item to a goal
func (r *Repository) CreateWorkItemGoalLink(ctx context.Context, agencyID string, link *models.WorkItemGoalLink) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure work item-goal links edge collection exists
	linksColl, err := ensureWorkItemGoalLinksCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure work item-goal links collection: %w", err)
	}

	// Set timestamps
	now := time.Now()
	link.CreatedAt = now
	link.UpdatedAt = now

	// Set default relationship if not provided
	if link.Relationship == "" {
		link.Relationship = "addresses"
	}

	// Build full _from and _to IDs if not provided
	if link.From == "" {
		link.From = fmt.Sprintf("work_items/%s", link.WorkItemKey)
	}
	if link.To == "" {
		link.To = fmt.Sprintf("goals/%s", link.GoalKey)
	}

	// Create the edge document
	meta, err := linksColl.CreateDocument(ctx, link)
	if err != nil {
		return fmt.Errorf("failed to create work item-goal link: %w", err)
	}

	link.Key = meta.Key
	link.ID = meta.ID.String()

	return nil
}

// GetWorkItemGoalLinks retrieves all goal links for a work item
func (r *Repository) GetWorkItemGoalLinks(ctx context.Context, agencyID, workItemKey string) ([]*models.WorkItemGoalLink, error) {
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure collection exists
	_, err = ensureWorkItemGoalLinksCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	// Query for links by work item key
	query := `
		FOR link IN work_item_goals
		FILTER link.work_item_key == @workItemKey
		RETURN link
	`
	bindVars := map[string]interface{}{
		"workItemKey": workItemKey,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query work item-goal links: %w", err)
	}
	defer cursor.Close()

	var links []*models.WorkItemGoalLink
	for cursor.HasMore() {
		var link models.WorkItemGoalLink
		_, err := cursor.ReadDocument(ctx, &link)
		if err != nil {
			return nil, fmt.Errorf("failed to read link document: %w", err)
		}
		links = append(links, &link)
	}

	return links, nil
}

// GetGoalWorkItems retrieves all work items linked to a goal
func (r *Repository) GetGoalWorkItems(ctx context.Context, agencyID, goalKey string) ([]*models.WorkItemGoalLink, error) {
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure collection exists
	_, err = ensureWorkItemGoalLinksCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	// Query for links by goal key
	query := `
		FOR link IN work_item_goals
		FILTER link.goal_key == @goalKey
		RETURN link
	`
	bindVars := map[string]interface{}{
		"goalKey": goalKey,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query goal work items: %w", err)
	}
	defer cursor.Close()

	var links []*models.WorkItemGoalLink
	for cursor.HasMore() {
		var link models.WorkItemGoalLink
		_, err := cursor.ReadDocument(ctx, &link)
		if err != nil {
			return nil, fmt.Errorf("failed to read link document: %w", err)
		}
		links = append(links, &link)
	}

	return links, nil
}

// DeleteWorkItemGoalLink deletes a specific work item-goal link
func (r *Repository) DeleteWorkItemGoalLink(ctx context.Context, agencyID, linkKey string) error {
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	linksColl, err := ensureWorkItemGoalLinksCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	_, err = linksColl.RemoveDocument(ctx, linkKey)
	if err != nil {
		return fmt.Errorf("failed to delete work item-goal link: %w", err)
	}

	return nil
}

// DeleteWorkItemGoalLinks deletes all goal links for a work item
func (r *Repository) DeleteWorkItemGoalLinks(ctx context.Context, agencyID, workItemKey string) error {
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure collection exists
	_, err = ensureWorkItemGoalLinksCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	// Delete all links for this work item
	query := `
		FOR link IN work_item_goals
		FILTER link.work_item_key == @workItemKey
		REMOVE link IN work_item_goals
	`
	bindVars := map[string]interface{}{
		"workItemKey": workItemKey,
	}

	_, err = agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to delete work item-goal links: %w", err)
	}

	return nil
}
