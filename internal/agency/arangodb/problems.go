package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/arangodb/go-driver"
)

// CreateProblem creates a new problem in the agency's problems collection
func (r *Repository) CreateProblem(ctx context.Context, problem *agency.Problem) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, problem.AgencyID)
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

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Set timestamps
	now := time.Now()
	problem.CreatedAt = now
	problem.UpdatedAt = now

	// If number is not set, get the next number
	if problem.Number == 0 {
		// Count existing problems
		query := "FOR p IN problems FILTER p.agency_id == @agencyId COLLECT WITH COUNT INTO length RETURN length"
		cursor, err := agencyDB.Query(ctx, query, map[string]interface{}{"agencyId": problem.AgencyID})
		if err != nil {
			return fmt.Errorf("failed to count problems: %w", err)
		}
		defer cursor.Close()

		var count int
		if cursor.HasMore() {
			_, err := cursor.ReadDocument(ctx, &count)
			if err != nil {
				return fmt.Errorf("failed to read count: %w", err)
			}
		}
		problem.Number = count + 1
	}

	// Create the document
	meta, err := problemsColl.CreateDocument(ctx, problem)
	if err != nil {
		return fmt.Errorf("failed to create problem: %w", err)
	}

	problem.Key = meta.Key
	return nil
}

// GetProblems retrieves all problems for an agency
func (r *Repository) GetProblems(ctx context.Context, agencyID string) ([]*agency.Problem, error) {
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

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Query all problems for this agency, ordered by number
	query := "FOR p IN @@collection FILTER p.agency_id == @agencyId SORT p.number ASC RETURN p"
	bindVars := map[string]interface{}{
		"@collection": problemsColl.Name(),
		"agencyId":    agencyID,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query problems: %w", err)
	}
	defer cursor.Close()

	var problems []*agency.Problem
	for cursor.HasMore() {
		var problem agency.Problem
		_, err := cursor.ReadDocument(ctx, &problem)
		if err != nil {
			return nil, fmt.Errorf("failed to read problem: %w", err)
		}
		problems = append(problems, &problem)
	}

	return problems, nil
}

// GetProblem retrieves a single problem by key
func (r *Repository) GetProblem(ctx context.Context, agencyID string, key string) (*agency.Problem, error) {
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

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Read the document
	var problem agency.Problem
	_, err = problemsColl.ReadDocument(ctx, key, &problem)
	if err != nil {
		return nil, fmt.Errorf("failed to read problem: %w", err)
	}

	return &problem, nil
}

// UpdateProblem updates an existing problem
func (r *Repository) UpdateProblem(ctx context.Context, problem *agency.Problem) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, problem.AgencyID)
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

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Update timestamp
	problem.UpdatedAt = time.Now()

	// Update the document
	_, err = problemsColl.UpdateDocument(ctx, problem.Key, problem)
	if err != nil {
		return fmt.Errorf("failed to update problem: %w", err)
	}

	return nil
}

// DeleteProblem deletes a problem and renumbers remaining problems
func (r *Repository) DeleteProblem(ctx context.Context, agencyID string, key string) error {
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

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Get the problem to find its number
	var problemToDelete agency.Problem
	_, err = problemsColl.ReadDocument(ctx, key, &problemToDelete)
	if err != nil {
		return fmt.Errorf("failed to read problem: %w", err)
	}

	// Delete the document
	_, err = problemsColl.RemoveDocument(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete problem: %w", err)
	}

	// Renumber problems with higher numbers
	query := `
		FOR p IN @@collection 
		FILTER p.agency_id == @agencyId AND p.number > @deletedNumber
		UPDATE p WITH { number: p.number - 1 } IN @@collection
	`
	bindVars := map[string]interface{}{
		"@collection":   problemsColl.Name(),
		"agencyId":      agencyID,
		"deletedNumber": problemToDelete.Number,
	}

	_, err = agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to renumber problems: %w", err)
	}

	return nil
}

// ensureProblemsCollection ensures the problems collection exists in an agency database
func ensureProblemsCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	const collectionName = "problems"

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
