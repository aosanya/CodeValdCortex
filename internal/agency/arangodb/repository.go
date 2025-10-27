package arangodb

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/arangodb/go-driver"
)

const (
	// CollectionName is the name of the agencies collection
	CollectionName = "agencies"
)

// Repository implements agency.Repository using ArangoDB
type Repository struct {
	client     driver.Client
	db         driver.Database
	collection driver.Collection
}

// New creates a new ArangoDB repository for agencies
func New(client driver.Client, db driver.Database) (agency.Repository, error) {
	// Ensure collection exists
	collection, err := ensureCollection(db)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	return &Repository{
		client:     client,
		db:         db,
		collection: collection,
	}, nil
}

// ensureCollection ensures the agencies collection exists with proper indexes
func ensureCollection(db driver.Database) (driver.Collection, error) {
	ctx := context.Background()

	// Check if collection exists
	exists, err := db.CollectionExists(ctx, CollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if !exists {
		// Create collection
		collection, err = db.CreateCollection(ctx, CollectionName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
	} else {
		collection, err = db.Collection(ctx, CollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	}

	// Ensure indexes
	if err := ensureIndexes(ctx, collection); err != nil {
		return nil, fmt.Errorf("failed to ensure indexes: %w", err)
	}

	return collection, nil
}

// ensureIndexes creates necessary indexes on the collection
func ensureIndexes(ctx context.Context, collection driver.Collection) error {
	// Index on ID field (for unique constraint)
	_, _, err := collection.EnsurePersistentIndex(ctx, []string{"id"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create id index: %w", err)
	}

	// Index on category field
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"category"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create category index: %w", err)
	}

	// Index on status field
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"status"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	// Compound index on category and status
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"category", "status"}, nil)
	if err != nil {
		return fmt.Errorf("failed to create category-status index: %w", err)
	}

	return nil
}
