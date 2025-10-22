package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/database"
	driver "github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
)

const agentTypesCollection = "agent_types"

// ArangoAgentTypeRepository implements AgentTypeRepository using ArangoDB
type ArangoAgentTypeRepository struct {
	db         driver.Database
	collection driver.Collection
}

// NewArangoAgentTypeRepository creates a new ArangoDB-backed repository
func NewArangoAgentTypeRepository(dbClient *database.ArangoClient) (*ArangoAgentTypeRepository, error) {
	db := dbClient.Database()

	// Ensure collection exists
	collection, err := ensureAgentTypesCollection(db)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure agent types collection: %w", err)
	}

	logrus.WithField("collection", agentTypesCollection).Info("Agent type repository initialized")

	return &ArangoAgentTypeRepository{
		db:         db,
		collection: collection,
	}, nil
}

// ensureAgentTypesCollection creates the agent_types collection if it doesn't exist
func ensureAgentTypesCollection(db driver.Database) (driver.Collection, error) {
	ctx := context.Background()

	// Check if collection exists
	exists, err := db.CollectionExists(ctx, agentTypesCollection)
	if err != nil {
		return nil, err
	}

	if exists {
		logrus.WithField("collection", agentTypesCollection).Info("Using existing collection")
		return db.Collection(ctx, agentTypesCollection)
	}

	// Create collection
	logrus.WithField("collection", agentTypesCollection).Info("Creating new collection")
	collection, err := db.CreateCollection(ctx, agentTypesCollection, nil)
	if err != nil {
		return nil, err
	}

	logrus.WithField("collection", agentTypesCollection).Info("Created new collection")
	return collection, nil
}

// Create stores a new agent type in the database
func (r *ArangoAgentTypeRepository) Create(ctx context.Context, agentType *AgentType) error {
	// Check if already exists
	exists, err := r.collection.DocumentExists(ctx, agentType.ID)
	if err != nil {
		return fmt.Errorf("failed to check existence: %w", err)
	}
	if exists {
		return fmt.Errorf("agent type %s already exists", agentType.ID)
	}

	// Set timestamps
	now := time.Now()
	agentType.CreatedAt = now
	agentType.UpdatedAt = now

	// Set ArangoDB key to match ID
	agentType.Key = agentType.ID

	// Store in database with explicit key
	_, err = r.collection.CreateDocument(ctx, agentType)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// Get retrieves an agent type by ID
func (r *ArangoAgentTypeRepository) Get(ctx context.Context, id string) (*AgentType, error) {
	var agentType AgentType
	_, err := r.collection.ReadDocument(ctx, id, &agentType)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("agent type %s not found", id)
		}
		return nil, fmt.Errorf("failed to read document: %w", err)
	}

	return &agentType, nil
}

// Update modifies an existing agent type
func (r *ArangoAgentTypeRepository) Update(ctx context.Context, agentType *AgentType) error {
	// Get existing document to preserve CreatedAt
	existing, err := r.Get(ctx, agentType.ID)
	if err != nil {
		return err
	}

	// Preserve creation timestamp and update modification timestamp
	agentType.CreatedAt = existing.CreatedAt
	agentType.UpdatedAt = time.Now()
	agentType.Key = agentType.ID

	// Replace document in database
	_, err = r.collection.ReplaceDocument(ctx, agentType.ID, agentType)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("agent type %s not found", agentType.ID)
		}
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

// Delete removes an agent type from the database
func (r *ArangoAgentTypeRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.RemoveDocument(ctx, id)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("agent type %s not found", id)
		}
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// List returns all agent types
func (r *ArangoAgentTypeRepository) List(ctx context.Context) ([]*AgentType, error) {
	query := "FOR doc IN @@collection RETURN doc"
	bindVars := map[string]interface{}{
		"@collection": agentTypesCollection,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var types []*AgentType
	for cursor.HasMore() {
		var agentType AgentType
		_, err := cursor.ReadDocument(ctx, &agentType)
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		types = append(types, &agentType)
	}

	return types, nil
}

// ListByCategory returns agent types filtered by category
func (r *ArangoAgentTypeRepository) ListByCategory(ctx context.Context, category string) ([]*AgentType, error) {
	query := "FOR doc IN @@collection FILTER doc.category == @category RETURN doc"
	bindVars := map[string]interface{}{
		"@collection": agentTypesCollection,
		"category":    category,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var types []*AgentType
	for cursor.HasMore() {
		var agentType AgentType
		_, err := cursor.ReadDocument(ctx, &agentType)
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		types = append(types, &agentType)
	}

	return types, nil
}

// ListEnabled returns only enabled agent types
func (r *ArangoAgentTypeRepository) ListEnabled(ctx context.Context) ([]*AgentType, error) {
	query := "FOR doc IN @@collection FILTER doc.is_enabled == true RETURN doc"
	bindVars := map[string]interface{}{
		"@collection": agentTypesCollection,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var types []*AgentType
	for cursor.HasMore() {
		var agentType AgentType
		_, err := cursor.ReadDocument(ctx, &agentType)
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		types = append(types, &agentType)
	}

	return types, nil
}

// Exists checks if an agent type exists
func (r *ArangoAgentTypeRepository) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := r.collection.DocumentExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return exists, nil
}

// Count returns the total number of agent types
func (r *ArangoAgentTypeRepository) Count(ctx context.Context) (int64, error) {
	query := "RETURN COUNT(@@collection)"
	bindVars := map[string]interface{}{
		"@collection": agentTypesCollection,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var count int64
	if cursor.HasMore() {
		_, err := cursor.ReadDocument(ctx, &count)
		if err != nil {
			return 0, fmt.Errorf("failed to read count: %w", err)
		}
	}

	return count, nil
}
