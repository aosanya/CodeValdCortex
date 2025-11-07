package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

// CreateGoal creates a new goal in the agency's goals collection
func (r *Repository) CreateGoal(ctx context.Context, goal *models.Goal) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, goal.AgencyID)
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

	// Ensure goals collection exists
	goalsColl, err := ensureGoalsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure goals collection: %w", err)
	}

	// Set timestamps
	now := time.Now()
	goal.CreatedAt = now
	goal.UpdatedAt = now

	// If number is not set, get the next number
	if goal.Number == 0 {
		// Count existing goals
		query := "FOR p IN goals FILTER p.agency_id == @agencyId COLLECT WITH COUNT INTO length RETURN length"
		cursor, err := agencyDB.Query(ctx, query, map[string]interface{}{"agencyId": goal.AgencyID})
		if err != nil {
			return fmt.Errorf("failed to count goals: %w", err)
		}
		defer cursor.Close()

		var count int
		if cursor.HasMore() {
			_, err := cursor.ReadDocument(ctx, &count)
			if err != nil {
				return fmt.Errorf("failed to read count: %w", err)
			}
		}
		goal.Number = count + 1
	}

	// Create the document
	meta, err := goalsColl.CreateDocument(ctx, goal)
	if err != nil {
		return fmt.Errorf("failed to create goal: %w", err)
	}

	goal.Key = meta.Key
	return nil
}

// GetGoals retrieves all goals for an agency
func (r *Repository) GetGoals(ctx context.Context, agencyID string) ([]*models.Goal, error) {
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

	// Ensure goals collection exists
	goalsColl, err := ensureGoalsCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure goals collection: %w", err)
	}

	// Query all goals for this agency, ordered by code
	query := "FOR p IN @@collection FILTER p.agency_id == @agencyId SORT p.code ASC RETURN p"
	bindVars := map[string]interface{}{
		"@collection": goalsColl.Name(),
		"agencyId":    agencyID,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query goals: %w", err)
	}
	defer cursor.Close()

	var goals []*models.Goal
	for cursor.HasMore() {
		var goal models.Goal
		_, err := cursor.ReadDocument(ctx, &goal)
		if err != nil {
			return nil, fmt.Errorf("failed to read goal: %w", err)
		}
		goals = append(goals, &goal)
	}

	return goals, nil
}

// GetGoal retrieves a single goal by key
func (r *Repository) GetGoal(ctx context.Context, agencyID string, key string) (*models.Goal, error) {
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

	// Ensure goals collection exists
	goalsColl, err := ensureGoalsCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure goals collection: %w", err)
	}

	// Read the document
	var goal models.Goal
	_, err = goalsColl.ReadDocument(ctx, key, &goal)
	if err != nil {
		return nil, fmt.Errorf("failed to read goal: %w", err)
	}

	return &goal, nil
}

// UpdateGoal updates an existing goal
func (r *Repository) UpdateGoal(ctx context.Context, goal *models.Goal) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, goal.AgencyID)
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

	// Ensure goals collection exists
	goalsColl, err := ensureGoalsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure goals collection: %w", err)
	}

	// Update timestamp
	goal.UpdatedAt = time.Now()

	// Update the document
	_, err = goalsColl.UpdateDocument(ctx, goal.Key, goal)
	if err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}

	return nil
}

// DeleteGoal deletes a goal and renumbers remaining goals
func (r *Repository) DeleteGoal(ctx context.Context, agencyID string, key string) error {
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

	// Ensure goals collection exists
	goalsColl, err := ensureGoalsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure goals collection: %w", err)
	}

	// Get the goal to find its number
	var goalToDelete models.Goal
	_, err = goalsColl.ReadDocument(ctx, key, &goalToDelete)
	if err != nil {
		return fmt.Errorf("failed to read goal: %w", err)
	}

	// Delete the document
	_, err = goalsColl.RemoveDocument(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete goal: %w", err)
	}

	// Renumber goals with higher numbers
	query := `
		FOR p IN @@collection 
		FILTER p.agency_id == @agencyId AND p.number > @deletedNumber
		UPDATE p WITH { number: p.number - 1 } IN @@collection
	`
	bindVars := map[string]interface{}{
		"@collection":   goalsColl.Name(),
		"agencyId":      agencyID,
		"deletedNumber": goalToDelete.Number,
	}

	_, err = agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to renumber goals: %w", err)
	}

	return nil
}

// ensureGoalsCollection ensures the goals collection exists in an agency database
func ensureGoalsCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	const collectionName = "goals"

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
