package main

import (
	"context"
	"log"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

// MigrateAgencyStatusToState migrates existing Status values to State field
// This script should be run once after deploying the new agency models
//
// Usage: go run scripts/migrate-agency-status-to-state.go
//
// Note: This is a standalone script that requires manual database connection setup
func main() {
	log.Println("Starting agency status to state migration...")
	log.Println("Note: This is a template script. Update with your database connection details.")

	// Suppress unused warning - this is a template script
	_ = migrateAgencies

	// TODO: Initialize database connection (example below)
	// db, err := connectToDatabase()
	// if err != nil {
	//     log.Fatalf("Failed to connect to database: %v", err)
	// }
	//
	// ctx := context.Background()
	// if err := migrateAgencies(ctx, db); err != nil {
	//     log.Fatalf("Migration failed: %v", err)
	// }

	log.Println("Migration script template created. Please configure database connection.")
}

// migrateAgencies performs the actual migration
//
//nolint:all // Template function for manual migration usage
func migrateAgencies(ctx context.Context, db driver.Database) error {
	// Get agencies collection
	col, err := db.Collection(ctx, "agencies")
	if err != nil {
		return err
	}

	// Query all agencies
	query := "FOR agency IN agencies RETURN agency"
	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()

	migratedCount := 0
	errorCount := 0

	// Process each agency
	for {
		var agency models.Agency
		meta, err := cursor.ReadDocument(ctx, &agency)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Printf("Error reading agency: %v", err)
			errorCount++
			continue
		}

		// Determine new state based on agency condition
		newState := determineState(agency)

		// Update agency
		update := map[string]interface{}{
			"state":      newState,
			"updated_at": time.Now(),
		}

		_, err = col.UpdateDocument(ctx, meta.Key, update)
		if err != nil {
			log.Printf("Failed to update agency %s: %v", agency.ID, err)
			errorCount++
			continue
		}

		log.Printf("Migrated agency %s: state -> %s", agency.ID, newState)
		migratedCount++
	}

	log.Printf("\nMigration complete:")
	log.Printf("  Migrated: %d agencies", migratedCount)
	log.Printf("  Errors: %d", errorCount)

	return nil
}

// determineState maps old Status to new State based on agency condition
//
//nolint:all // Template function called by migrateAgencies
func determineState(agency models.Agency) models.AgencyState {
	// If State is already set and valid, keep it
	if models.IsValidAgencyState(agency.State) {
		return agency.State
	}

	// For agencies with no explicit state, infer from context
	// Note: Since we removed Status, we need to handle agencies that may not have State set yet

	// Check if agency has publication metadata
	if agency.PublishedAt != nil {
		return models.AgencyStatePublished
	}

	// Default to draft for new/incomplete agencies
	return models.AgencyStateDraft
}
