package migrations

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
)

// Migration006AgencyPublishing adds publishing/tagging collections
func Migration006AgencyPublishing(ctx context.Context, db driver.Database) error {
	// Create agency_publications collection
	if err := createPublicationsCollection(ctx, db); err != nil {
		return fmt.Errorf("failed to create publications collection: %w", err)
	}

	// Create agency_tags collection
	if err := createTagsCollection(ctx, db); err != nil {
		return fmt.Errorf("failed to create tags collection: %w", err)
	}

	return nil
}

func createPublicationsCollection(ctx context.Context, db driver.Database) error {
	// Create collection
	_, err := db.CreateCollection(ctx, "agency_publications", nil)
	if err != nil && !driver.IsConflict(err) {
		return err
	}

	col, err := db.Collection(ctx, "agency_publications")
	if err != nil {
		return err
	}

	// Create indexes
	_, _, err = col.EnsureHashIndex(ctx, []string{"agency_id"}, nil)
	if err != nil {
		return err
	}

	_, _, err = col.EnsureSkipListIndex(ctx, []string{"published_at"}, nil)
	if err != nil {
		return err
	}

	_, _, err = col.EnsureHashIndex(ctx, []string{"version"}, nil)
	if err != nil {
		return err
	}

	return nil
}

func createTagsCollection(ctx context.Context, db driver.Database) error {
	// Create collection
	_, err := db.CreateCollection(ctx, "agency_tags", nil)
	if err != nil && !driver.IsConflict(err) {
		return err
	}

	col, err := db.Collection(ctx, "agency_tags")
	if err != nil {
		return err
	}

	// Create indexes
	_, _, err = col.EnsureHashIndex(ctx, []string{"agency_id"}, nil)
	if err != nil {
		return err
	}

	// Unique index on agency_id + name
	_, _, err = col.EnsureHashIndex(ctx, []string{"agency_id", "name"}, &driver.EnsureHashIndexOptions{
		Unique: true,
	})
	if err != nil {
		return err
	}

	_, _, err = col.EnsureSkipListIndex(ctx, []string{"version"}, nil)
	if err != nil {
		return err
	}

	_, _, err = col.EnsureSkipListIndex(ctx, []string{"created_at"}, nil)
	if err != nil {
		return err
	}

	_, _, err = col.EnsureHashIndex(ctx, []string{"type"}, nil)
	if err != nil {
		return err
	}

	return nil
}
