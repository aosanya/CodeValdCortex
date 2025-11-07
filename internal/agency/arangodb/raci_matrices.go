package arangodb

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

const (
	// RACIMatrixCollectionName is the name of the RACI matrices collection
	RACIMatrixCollectionName = "raci_matrices"
)

// SaveRACIMatrix saves a RACI matrix to the agency database
func (r *Repository) SaveRACIMatrix(ctx context.Context, agencyID string, matrix *models.RACIMatrix) error {
	// Get agency-specific database
	agencyDB, err := r.getAgencyDatabase(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency database: %w", err)
	}

	// Ensure RACI matrices collection exists
	collection, err := ensureRACIMatrixCollection(agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure RACI matrices collection: %w", err)
	}

	// Insert the matrix
	meta, err := collection.CreateDocument(ctx, matrix)
	if err != nil {
		return fmt.Errorf("failed to create RACI matrix document: %w", err)
	}

	matrix.Key = meta.Key
	matrix.ID = meta.ID.String()

	return nil
}

// GetRACIMatrix retrieves a RACI matrix by key
func (r *Repository) GetRACIMatrix(ctx context.Context, agencyID string, key string) (*models.RACIMatrix, error) {
	// Get agency-specific database
	agencyDB, err := r.getAgencyDatabase(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency database: %w", err)
	}

	// Ensure RACI matrices collection exists
	collection, err := ensureRACIMatrixCollection(agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure RACI matrices collection: %w", err)
	}

	// Read the document
	var matrix models.RACIMatrix
	_, err = collection.ReadDocument(ctx, key, &matrix)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("RACI matrix not found: %s", key)
		}
		return nil, fmt.Errorf("failed to read RACI matrix: %w", err)
	}

	return &matrix, nil
}

// ListRACIMatrices lists all RACI matrices for an agency
func (r *Repository) ListRACIMatrices(ctx context.Context, agencyID string) ([]*models.RACIMatrix, error) {
	// Get agency-specific database
	agencyDB, err := r.getAgencyDatabase(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency database: %w", err)
	}

	// Ensure RACI matrices collection exists
	collection, err := ensureRACIMatrixCollection(agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure RACI matrices collection: %w", err)
	}

	// Query all matrices
	query := `FOR matrix IN @@collection SORT matrix.created_at DESC RETURN matrix`
	bindVars := map[string]interface{}{
		"@collection": collection.Name(),
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query RACI matrices: %w", err)
	}
	defer cursor.Close()

	var matrices []*models.RACIMatrix
	for {
		var matrix models.RACIMatrix
		_, err := cursor.ReadDocument(ctx, &matrix)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read RACI matrix document: %w", err)
		}
		matrices = append(matrices, &matrix)
	}

	return matrices, nil
}

// UpdateRACIMatrix updates an existing RACI matrix
func (r *Repository) UpdateRACIMatrix(ctx context.Context, agencyID string, matrix *models.RACIMatrix) error {
	// Get agency-specific database
	agencyDB, err := r.getAgencyDatabase(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency database: %w", err)
	}

	// Ensure RACI matrices collection exists
	collection, err := ensureRACIMatrixCollection(agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure RACI matrices collection: %w", err)
	}

	// Update the document
	_, err = collection.UpdateDocument(ctx, matrix.Key, matrix)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("RACI matrix not found: %s", matrix.Key)
		}
		return fmt.Errorf("failed to update RACI matrix: %w", err)
	}

	return nil
}

// DeleteRACIMatrix deletes a RACI matrix
func (r *Repository) DeleteRACIMatrix(ctx context.Context, agencyID string, key string) error {
	// Get agency-specific database
	agencyDB, err := r.getAgencyDatabase(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency database: %w", err)
	}

	// Ensure RACI matrices collection exists
	collection, err := ensureRACIMatrixCollection(agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure RACI matrices collection: %w", err)
	}

	// Delete the document
	_, err = collection.RemoveDocument(ctx, key)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("RACI matrix not found: %s", key)
		}
		return fmt.Errorf("failed to delete RACI matrix: %w", err)
	}

	return nil
}

// ensureRACIMatrixCollection ensures the RACI matrices collection exists
func ensureRACIMatrixCollection(db driver.Database) (driver.Collection, error) {
	ctx := context.Background()

	// Check if collection exists
	exists, err := db.CollectionExists(ctx, RACIMatrixCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if !exists {
		// Create collection
		collection, err = db.CreateCollection(ctx, RACIMatrixCollectionName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}

		// Create index on agency_id
		_, _, err = collection.EnsurePersistentIndex(ctx, []string{"agency_id"}, &driver.EnsurePersistentIndexOptions{
			Name:   "idx_agency_id",
			Unique: false,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create agency_id index: %w", err)
		}

		// Create index on work_item_key for linking RACI to work items
		_, _, err = collection.EnsurePersistentIndex(ctx, []string{"work_item_key"}, &driver.EnsurePersistentIndexOptions{
			Name:   "idx_work_item_key",
			Unique: false,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create work_item_key index: %w", err)
		}
	} else {
		collection, err = db.Collection(ctx, RACIMatrixCollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to open collection: %w", err)
		}
	}

	return collection, nil
}

// getAgencyDatabase gets the database for a specific agency
func (r *Repository) getAgencyDatabase(ctx context.Context, agencyID string) (driver.Database, error) {
	// Get agency to find database name
	agency, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if Database field is not set
	dbName := agency.Database
	if dbName == "" {
		dbName = agency.ID
	}

	// Check if database exists
	exists, err := r.client.DatabaseExists(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("agency database not found: %s", dbName)
	}

	// Open database
	db, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to open agency database: %w", err)
	}

	return db, nil
}
