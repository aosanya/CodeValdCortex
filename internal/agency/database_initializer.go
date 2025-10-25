package agency

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
)

// DatabaseInitializer handles creation and initialization of agency databases
type DatabaseInitializer interface {
	InitializeAgencyDatabase(ctx context.Context, agencyID string) error
}

// databaseInitializer implements DatabaseInitializer
type databaseInitializer struct {
	client driver.Client
	logger *logrus.Logger
}

// NewDatabaseInitializer creates a new database initializer
func NewDatabaseInitializer(client driver.Client, logger *logrus.Logger) DatabaseInitializer {
	return &databaseInitializer{
		client: client,
		logger: logger,
	}
}

// InitializeAgencyDatabase creates a new database for the agency and initializes standard collections
func (d *databaseInitializer) InitializeAgencyDatabase(ctx context.Context, agencyID string) error {
	// Use agency ID directly as database name (already has "agency_" prefix)
	dbName := agencyID
	
	d.logger.WithFields(logrus.Fields{
		"agencyID": agencyID,
		"dbName":   dbName,
	}).Debug("Initializing agency database")
	
	// Check if database already exists
	exists, err := d.client.DatabaseExists(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if exists {
		d.logger.WithField("database", dbName).Info("Database already exists, skipping creation")
		return nil
	}

	// Create database
	db, err := d.client.CreateDatabase(ctx, dbName, nil)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	d.logger.WithField("database", dbName).Info("Created agency database")	// Initialize standard collections
	collections := []string{
		"agents",
		"agent_types",
		"agent_messages",
		"agent_publications",
		"agent_subscriptions",
	}

	for _, collName := range collections {
		// Check if collection exists
		exists, err := db.CollectionExists(ctx, collName)
		if err != nil {
			return fmt.Errorf("failed to check collection %s: %w", collName, err)
		}

		if exists {
			continue
		}

		// Create collection
		options := &driver.CreateCollectionOptions{
			WaitForSync: false,
		}
		_, err = db.CreateCollection(ctx, collName, options)
		if err != nil {
			return fmt.Errorf("failed to create collection %s: %w", collName, err)
		}

		d.logger.WithFields(logrus.Fields{
			"database":   agencyID,
			"collection": collName,
		}).Debug("Created collection")
	}

	d.logger.WithField("database", agencyID).Info("Initialized agency collections")

	return nil
}
