package database

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/config"
	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	log "github.com/sirupsen/logrus"
)

// ArangoClient wraps the ArangoDB client and database connection
type ArangoClient struct {
	client   driver.Client
	db       driver.Database
	config   *config.DatabaseConfig
	ctx      context.Context
	cancelFn context.CancelFunc
}

// NewArangoClient creates a new ArangoDB client with connection pooling
func NewArangoClient(cfg *config.DatabaseConfig) (*ArangoClient, error) {
	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	// Create HTTP connection
	connConfig := http.ConnectionConfig{
		Endpoints: []string{fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)},
	}

	conn, err := http.NewConnection(connConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	// Create client configuration
	clientConfig := driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(cfg.Username, cfg.Password),
	}

	// Create client
	client, err := driver.NewClient(clientConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Get or create database
	db, err := ensureDatabase(ctx, client, cfg.Database)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to ensure database: %w", err)
	}

	log.WithFields(log.Fields{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"database": cfg.Database,
	}).Info("Connected to ArangoDB")

	return &ArangoClient{
		client:   client,
		db:       db,
		config:   cfg,
		ctx:      ctx,
		cancelFn: cancel,
	}, nil
}

// ensureDatabase creates the database if it doesn't exist
func ensureDatabase(ctx context.Context, client driver.Client, dbName string) (driver.Database, error) {
	// Check if database exists
	exists, err := client.DatabaseExists(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	if exists {
		db, err := client.Database(ctx, dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		log.WithField("database", dbName).Info("Using existing database")
		return db, nil
	}

	// Create database
	db, err := client.CreateDatabase(ctx, dbName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	log.WithField("database", dbName).Info("Created new database")
	return db, nil
}

// Database returns the database instance
func (ac *ArangoClient) Database() driver.Database {
	return ac.db
}

// Client returns the client instance
func (ac *ArangoClient) Client() driver.Client {
	return ac.client
}

// Context returns the client context
func (ac *ArangoClient) Context() context.Context {
	return ac.ctx
}

// Close closes the client connection
func (ac *ArangoClient) Close() error {
	if ac.cancelFn != nil {
		ac.cancelFn()
	}
	log.Info("Closed ArangoDB connection")
	return nil
}

// Ping verifies the connection to ArangoDB
func (ac *ArangoClient) Ping() error {
	version, err := ac.client.Version(ac.ctx)
	if err != nil {
		return fmt.Errorf("failed to ping ArangoDB: %w", err)
	}

	log.WithField("version", version.Version).Debug("ArangoDB ping successful")
	return nil
}
