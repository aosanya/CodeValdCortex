package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

// GetOverview retrieves the overview document for an agency
func (r *Repository) GetOverview(ctx context.Context, agencyID string) (*models.Overview, error) {
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

	// Ensure overview collection exists
	overviewColl, err := ensureOverviewCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure overview collection: %w", err)
	}

	// Try to read the overview document (using "main" as the key)
	var overview models.Overview
	_, err = overviewColl.ReadDocument(ctx, "main", &overview)
	if err != nil {
		if driver.IsNotFound(err) {
			// Create a default overview if it doesn't exist
			overview = models.Overview{
				Key:          "main",
				AgencyID:     agencyID,
				Introduction: "",
				UpdatedAt:    time.Now(),
			}
			_, err = overviewColl.CreateDocument(ctx, &overview)
			if err != nil {
				return nil, fmt.Errorf("failed to create default overview: %w", err)
			}
			return &overview, nil
		}
		return nil, fmt.Errorf("failed to read overview: %w", err)
	}

	return &overview, nil
}

// UpdateOverview updates the overview document for an agency
func (r *Repository) UpdateOverview(ctx context.Context, overview *models.Overview) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, overview.AgencyID)
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

	// Ensure overview collection exists
	overviewColl, err := ensureOverviewCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure overview collection: %w", err)
	}

	// Set the key and updated timestamp
	overview.Key = "main"
	overview.UpdatedAt = time.Now()

	// Update or create the document
	_, err = overviewColl.UpdateDocument(ctx, overview.Key, overview)
	if err != nil {
		if driver.IsNotFound(err) {
			// Create if it doesn't exist
			_, err = overviewColl.CreateDocument(ctx, overview)
			if err != nil {
				return fmt.Errorf("failed to create overview: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to update overview: %w", err)
	}

	return nil
}

// ensureOverviewCollection ensures the overview collection exists in an agency database
func ensureOverviewCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	const collectionName = "overview"

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
