package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/arangodb/go-driver"
)

// CreateUnitOfWork creates a new unit of work in an agency's database
func (r *Repository) CreateUnitOfWork(ctx context.Context, unit *agency.UnitOfWork) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, unit.AgencyID)
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

	// Ensure units_of_work collection exists
	unitsColl, err := ensureUnitsOfWorkCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure units_of_work collection: %w", err)
	}

	// Set timestamps
	now := time.Now()
	unit.CreatedAt = now
	unit.UpdatedAt = now

	// If number is not set, get the next number
	if unit.Number == 0 {
		// Count existing units
		query := "FOR u IN units_of_work FILTER u.agency_id == @agencyId COLLECT WITH COUNT INTO length RETURN length"
		cursor, err := agencyDB.Query(ctx, query, map[string]interface{}{"agencyId": unit.AgencyID})
		if err != nil {
			return fmt.Errorf("failed to count units: %w", err)
		}
		defer cursor.Close()

		var count int
		if cursor.HasMore() {
			_, err := cursor.ReadDocument(ctx, &count)
			if err != nil {
				return fmt.Errorf("failed to read count: %w", err)
			}
		}
		unit.Number = count + 1
	}

	// Create the document
	meta, err := unitsColl.CreateDocument(ctx, unit)
	if err != nil {
		return fmt.Errorf("failed to create unit of work: %w", err)
	}

	unit.Key = meta.Key
	return nil
}

// GetUnitsOfWork retrieves all units of work for an agency
func (r *Repository) GetUnitsOfWork(ctx context.Context, agencyID string) ([]*agency.UnitOfWork, error) {
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

	// Ensure units_of_work collection exists
	unitsColl, err := ensureUnitsOfWorkCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure units_of_work collection: %w", err)
	}

	// Query all units for this agency, ordered by number
	query := "FOR u IN @@collection FILTER u.agency_id == @agencyId SORT u.number ASC RETURN u"
	bindVars := map[string]interface{}{
		"@collection": unitsColl.Name(),
		"agencyId":    agencyID,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query units of work: %w", err)
	}
	defer cursor.Close()

	var units []*agency.UnitOfWork
	for cursor.HasMore() {
		var unit agency.UnitOfWork
		_, err := cursor.ReadDocument(ctx, &unit)
		if err != nil {
			return nil, fmt.Errorf("failed to read unit of work: %w", err)
		}
		units = append(units, &unit)
	}

	return units, nil
}

// GetUnitOfWork retrieves a single unit of work by key
func (r *Repository) GetUnitOfWork(ctx context.Context, agencyID string, key string) (*agency.UnitOfWork, error) {
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

	// Ensure units_of_work collection exists
	unitsColl, err := ensureUnitsOfWorkCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure units_of_work collection: %w", err)
	}

	// Read the document
	var unit agency.UnitOfWork
	_, err = unitsColl.ReadDocument(ctx, key, &unit)
	if err != nil {
		return nil, fmt.Errorf("failed to read unit of work: %w", err)
	}

	return &unit, nil
}

// UpdateUnitOfWork updates an existing unit of work
func (r *Repository) UpdateUnitOfWork(ctx context.Context, unit *agency.UnitOfWork) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, unit.AgencyID)
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

	// Ensure units_of_work collection exists
	unitsColl, err := ensureUnitsOfWorkCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure units_of_work collection: %w", err)
	}

	// Update timestamp
	unit.UpdatedAt = time.Now()

	// Update the document
	_, err = unitsColl.UpdateDocument(ctx, unit.Key, unit)
	if err != nil {
		return fmt.Errorf("failed to update unit of work: %w", err)
	}

	return nil
}

// DeleteUnitOfWork deletes a unit of work and renumbers remaining units
func (r *Repository) DeleteUnitOfWork(ctx context.Context, agencyID string, key string) error {
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

	// Ensure units_of_work collection exists
	unitsColl, err := ensureUnitsOfWorkCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure units_of_work collection: %w", err)
	}

	// Get the unit to find its number
	var unitToDelete agency.UnitOfWork
	_, err = unitsColl.ReadDocument(ctx, key, &unitToDelete)
	if err != nil {
		return fmt.Errorf("failed to read unit of work: %w", err)
	}

	// Delete the document
	_, err = unitsColl.RemoveDocument(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete unit of work: %w", err)
	}

	// Renumber units with higher numbers
	query := `
		FOR u IN @@collection 
		FILTER u.agency_id == @agencyId AND u.number > @deletedNumber
		UPDATE u WITH { number: u.number - 1 } IN @@collection
	`
	bindVars := map[string]interface{}{
		"@collection":   unitsColl.Name(),
		"agencyId":      agencyID,
		"deletedNumber": unitToDelete.Number,
	}

	_, err = agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to renumber units of work: %w", err)
	}

	return nil
}

// ensureUnitsOfWorkCollection ensures the units_of_work collection exists in an agency database
func ensureUnitsOfWorkCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	const collectionName = "units_of_work"

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
