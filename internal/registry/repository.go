package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/database"
	driver "github.com/arangodb/go-driver"
	log "github.com/sirupsen/logrus"
)

const (
	// CollectionAgents is the name of the agents collection
	CollectionAgents = "agents"
)

// Repository handles agent persistence in ArangoDB
type Repository struct {
	db         *database.ArangoClient
	collection driver.Collection
}

// AgentDocument represents an agent document in ArangoDB
type AgentDocument struct {
	Key       string            `json:"_key,omitempty"`
	Rev       string            `json:"_rev,omitempty"`
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	State     string            `json:"state"`
	Metadata  map[string]string `json:"metadata"`
	Config    agent.Config      `json:"config"`
	IsHealthy bool              `json:"is_healthy"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// NewRepository creates a new agent registry repository
func NewRepository(dbClient *database.ArangoClient) (*Repository, error) {
	ctx := dbClient.Context()
	db := dbClient.Database()

	// Ensure collection exists
	col, err := ensureCollection(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	// Create indexes
	if err := ensureIndexes(ctx, col); err != nil {
		return nil, fmt.Errorf("failed to ensure indexes: %w", err)
	}

	log.WithField("collection", CollectionAgents).Info("Agent registry repository initialized")

	return &Repository{
		db:         dbClient,
		collection: col,
	}, nil
}

// ensureCollection creates the collection if it doesn't exist
func ensureCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	exists, err := db.CollectionExists(ctx, CollectionAgents)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	if exists {
		col, err := db.Collection(ctx, CollectionAgents)
		if err != nil {
			return nil, fmt.Errorf("failed to open collection: %w", err)
		}
		log.WithField("collection", CollectionAgents).Debug("Using existing collection")
		return col, nil
	}

	// Create collection
	col, err := db.CreateCollection(ctx, CollectionAgents, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	log.WithField("collection", CollectionAgents).Info("Created new collection")
	return col, nil
}

// ensureIndexes creates necessary indexes for efficient queries
func ensureIndexes(ctx context.Context, col driver.Collection) error {
	// Index on agent type
	_, _, err := col.EnsurePersistentIndex(ctx, []string{"type"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_type",
		Unique: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create type index: %w", err)
	}

	// Index on agent state
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"state"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_state",
		Unique: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create state index: %w", err)
	}

	// Index on health status
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"is_healthy"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_health",
		Unique: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create health index: %w", err)
	}

	// Composite index on type and state for common queries
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"type", "state"}, &driver.EnsurePersistentIndexOptions{
		Name:   "idx_type_state",
		Unique: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create type_state index: %w", err)
	}

	log.Debug("Indexes created successfully")
	return nil
}

// Create stores a new agent in the registry
func (r *Repository) Create(ctx context.Context, ag *agent.Agent) error {
	doc := toDocument(ag)
	doc.UpdatedAt = time.Now()

	meta, err := r.collection.CreateDocument(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to create agent document: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": ag.ID,
		"key":      meta.Key,
	}).Debug("Agent created in registry")

	return nil
}

// Get retrieves an agent by ID
func (r *Repository) Get(ctx context.Context, id string) (*agent.Agent, error) {
	var doc AgentDocument

	// Use ID as the document key
	_, err := r.collection.ReadDocument(ctx, id, &doc)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("agent not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read agent document: %w", err)
	}

	return fromDocument(&doc), nil
}

// Update updates an existing agent in the registry
func (r *Repository) Update(ctx context.Context, ag *agent.Agent) error {
	doc := toDocument(ag)
	doc.UpdatedAt = time.Now()

	_, err := r.collection.UpdateDocument(ctx, ag.ID, doc)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("agent not found: %s", ag.ID)
		}
		return fmt.Errorf("failed to update agent document: %w", err)
	}

	log.WithField("agent_id", ag.ID).Debug("Agent updated in registry")
	return nil
}

// Delete removes an agent from the registry
func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.RemoveDocument(ctx, id)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("agent not found: %s", id)
		}
		return fmt.Errorf("failed to delete agent document: %w", err)
	}

	log.WithField("agent_id", id).Debug("Agent deleted from registry")
	return nil
}

// List retrieves all agents from the registry
func (r *Repository) List(ctx context.Context) ([]*agent.Agent, error) {
	query := "FOR doc IN @@collection RETURN doc"
	bindVars := map[string]interface{}{
		"@collection": CollectionAgents,
	}

	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var agents []*agent.Agent
	for {
		var doc AgentDocument
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		agents = append(agents, fromDocument(&doc))
	}

	return agents, nil
}

// FindByType retrieves agents by type
func (r *Repository) FindByType(ctx context.Context, agentType string) ([]*agent.Agent, error) {
	query := "FOR doc IN @@collection FILTER doc.type == @type RETURN doc"
	bindVars := map[string]interface{}{
		"@collection": CollectionAgents,
		"type":        agentType,
	}

	return r.executeQuery(ctx, query, bindVars)
}

// FindByState retrieves agents by state
func (r *Repository) FindByState(ctx context.Context, state agent.State) ([]*agent.Agent, error) {
	query := "FOR doc IN @@collection FILTER doc.state == @state RETURN doc"
	bindVars := map[string]interface{}{
		"@collection": CollectionAgents,
		"state":       string(state),
	}

	return r.executeQuery(ctx, query, bindVars)
}

// FindHealthy retrieves all healthy agents
func (r *Repository) FindHealthy(ctx context.Context) ([]*agent.Agent, error) {
	query := "FOR doc IN @@collection FILTER doc.is_healthy == true RETURN doc"
	bindVars := map[string]interface{}{
		"@collection": CollectionAgents,
	}

	return r.executeQuery(ctx, query, bindVars)
}

// FindByTypeAndState retrieves agents by type and state
func (r *Repository) FindByTypeAndState(ctx context.Context, agentType string, state agent.State) ([]*agent.Agent, error) {
	query := "FOR doc IN @@collection FILTER doc.type == @type AND doc.state == @state RETURN doc"
	bindVars := map[string]interface{}{
		"@collection": CollectionAgents,
		"type":        agentType,
		"state":       string(state),
	}

	return r.executeQuery(ctx, query, bindVars)
}

// executeQuery executes an AQL query and returns agents
func (r *Repository) executeQuery(ctx context.Context, query string, bindVars map[string]interface{}) ([]*agent.Agent, error) {
	cursor, err := r.db.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var agents []*agent.Agent
	for {
		var doc AgentDocument
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		agents = append(agents, fromDocument(&doc))
	}

	return agents, nil
}

// Count returns the total number of agents in the registry
func (r *Repository) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	return count, nil
}

// toDocument converts an Agent to an AgentDocument
func toDocument(ag *agent.Agent) *AgentDocument {
	return &AgentDocument{
		Key:       ag.ID,
		ID:        ag.ID,
		Name:      ag.Name,
		Type:      ag.Type,
		State:     string(ag.State),
		Metadata:  ag.Metadata,
		Config:    ag.Config,
		IsHealthy: ag.IsHealthy(),
		CreatedAt: ag.CreatedAt,
	}
}

// fromDocument converts an AgentDocument to an Agent
func fromDocument(doc *AgentDocument) *agent.Agent {
	ag := &agent.Agent{
		ID:        doc.ID,
		Name:      doc.Name,
		Type:      doc.Type,
		State:     agent.State(doc.State),
		Metadata:  doc.Metadata,
		Config:    doc.Config,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
	return ag
}
