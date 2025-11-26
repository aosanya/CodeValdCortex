package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
	"github.com/google/uuid"
)

const (
	instanceCollectionName      = "agency_instances"
	instanceAgentCollectionName = "instance_agents"
)

// InstanceRepository implements the instance repository using agency-specific databases
type InstanceRepository struct {
	client driver.Client // ArangoDB client to access different databases
}

// NewInstanceRepository creates a new instance repository
func NewInstanceRepository(client driver.Client) (*InstanceRepository, error) {
	return &InstanceRepository{
		client: client,
	}, nil
}

// getInstanceCollection gets or creates the instances collection in the agency's database
func (r *InstanceRepository) getInstanceCollection(ctx context.Context, agencyDB string) (driver.Collection, error) {
	// Get the agency's database
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to access agency database %s: %w", agencyDB, err)
	}

	// Check if collection exists
	exists, err := db.CollectionExists(ctx, instanceCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if exists {
		collection, err = db.Collection(ctx, instanceCollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	} else {
		// Create the collection
		collection, err = db.CreateCollection(ctx, instanceCollectionName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}

		// Create indexes
		if err := r.createInstanceIndexes(ctx, collection); err != nil {
			return nil, fmt.Errorf("failed to create indexes: %w", err)
		}
	}

	return collection, nil
}

// getInstanceAgentCollection gets or creates the instance_agents collection
func (r *InstanceRepository) getInstanceAgentCollection(ctx context.Context, agencyDB string) (driver.Collection, error) {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to access agency database %s: %w", agencyDB, err)
	}

	exists, err := db.CollectionExists(ctx, instanceAgentCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if exists {
		collection, err = db.Collection(ctx, instanceAgentCollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	} else {
		collection, err = db.CreateCollection(ctx, instanceAgentCollectionName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
	}

	return collection, nil
}

// createInstanceIndexes creates necessary indexes for the instances collection
func (r *InstanceRepository) createInstanceIndexes(ctx context.Context, collection driver.Collection) error {
	// Index on instance_id for quick lookups
	_, _, err := collection.EnsurePersistentIndex(ctx, []string{"instance_id"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
		Name:   "idx_instance_id",
	})
	if err != nil {
		return fmt.Errorf("failed to create instance_id index: %w", err)
	}

	// Index on agency_id for listing instances by agency
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"agency_id"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_agency_id",
	})
	if err != nil {
		return fmt.Errorf("failed to create agency_id index: %w", err)
	}

	// Unique index on agency_id + name (enforce unique names per agency)
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"agency_id", "name"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
		Name:   "idx_agency_name_unique",
	})
	if err != nil {
		return fmt.Errorf("failed to create agency_id+name unique index: %w", err)
	}

	// Index on tag_id for finding instances by tag
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"tag_id"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_tag_id",
	})
	if err != nil {
		return fmt.Errorf("failed to create tag_id index: %w", err)
	}

	// Index on state for filtering by state
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"state"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_state",
	})
	if err != nil {
		return fmt.Errorf("failed to create state index: %w", err)
	}

	// Index on deployed_at for chronological sorting
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"deployed_at"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_deployed_at",
	})
	if err != nil {
		return fmt.Errorf("failed to create deployed_at index: %w", err)
	}

	return nil
}

// Create creates a new instance in the agency's database
func (r *InstanceRepository) Create(ctx context.Context, instance *models.AgencyInstance, agencyDB string) error {
	collection, err := r.getInstanceCollection(ctx, agencyDB)
	if err != nil {
		return err
	}

	// Generate instance ID if not set
	if instance.InstanceID == "" {
		instance.InstanceID = uuid.New().String()
	}

	// Set document key to instance_id
	instance.Key = instance.InstanceID

	// Set timestamps
	if instance.StartedAt == nil {
		now := time.Now()
		instance.StartedAt = &now
	}
	if instance.DeployedAt.IsZero() {
		instance.DeployedAt = time.Now()
	}
	if instance.LastHeartbeat.IsZero() {
		instance.LastHeartbeat = time.Now()
	}

	meta, err := collection.CreateDocument(ctx, instance)
	if err != nil {
		if driver.IsConflict(err) {
			return fmt.Errorf("instance with ID '%s' already exists", instance.InstanceID)
		}
		return fmt.Errorf("failed to create instance: %w", err)
	}

	instance.ID = meta.ID.String()
	instance.Rev = meta.Rev

	return nil
}

// GetByID retrieves an instance by its ID
func (r *InstanceRepository) GetByID(ctx context.Context, instanceID string, agencyDB string) (*models.AgencyInstance, error) {
	collection, err := r.getInstanceCollection(ctx, agencyDB)
	if err != nil {
		return nil, err
	}

	var instance models.AgencyInstance
	_, err = collection.ReadDocument(ctx, instanceID, &instance)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("instance not found: %s", instanceID)
		}
		return nil, fmt.Errorf("failed to read instance: %w", err)
	}

	return &instance, nil
}

// Update updates an existing instance
func (r *InstanceRepository) Update(ctx context.Context, instance *models.AgencyInstance, agencyDB string) error {
	collection, err := r.getInstanceCollection(ctx, agencyDB)
	if err != nil {
		return err
	}

	meta, err := collection.UpdateDocument(ctx, instance.Key, instance)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("instance not found: %s", instance.InstanceID)
		}
		return fmt.Errorf("failed to update instance: %w", err)
	}

	instance.Rev = meta.Rev
	return nil
}

// Delete deletes an instance
func (r *InstanceRepository) Delete(ctx context.Context, instanceID string, agencyDB string) error {
	collection, err := r.getInstanceCollection(ctx, agencyDB)
	if err != nil {
		return err
	}

	// Soft delete: mark as deleted, don't remove from database
	now := time.Now()
	updateDoc := map[string]interface{}{
		"is_deleted": true,
		"deleted_at": now,
	}

	_, err = collection.UpdateDocument(ctx, instanceID, updateDoc)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("instance not found: %s", instanceID)
		}
		return fmt.Errorf("failed to soft delete instance: %w", err)
	}

	return nil
}

// ListByAgency lists all instances for an agency
func (r *InstanceRepository) ListByAgency(ctx context.Context, agencyID string, agencyDB string) ([]*models.AgencyInstance, error) {
	_, err := r.getInstanceCollection(ctx, agencyDB)
	if err != nil {
		return nil, err
	}

	query := `
		FOR instance IN @@collection
			FILTER instance.agency_id == @agencyID
			FILTER instance.is_deleted != true
			SORT instance.started_at DESC
			RETURN instance
	`

	bindVars := map[string]interface{}{
		"@collection": instanceCollectionName,
		"agencyID":    agencyID,
	}

	cursor, err := r.executeQuery(ctx, agencyDB, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var instances []*models.AgencyInstance
	for cursor.HasMore() {
		var instance models.AgencyInstance
		_, err := cursor.ReadDocument(ctx, &instance)
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		instances = append(instances, &instance)
	}

	return instances, nil
}

// ListByTag lists all instances created from a specific tag
func (r *InstanceRepository) ListByTag(ctx context.Context, agencyID string, tagName string, agencyDB string) ([]*models.AgencyInstance, error) {
	_, err := r.getInstanceCollection(ctx, agencyDB)
	if err != nil {
		return nil, err
	}

	query := `
		FOR instance IN @@collection
			FILTER instance.agency_id == @agencyID
			FILTER instance.tag_name == @tagName
			FILTER instance.is_deleted != true
			SORT instance.started_at DESC
			RETURN instance
	`

	bindVars := map[string]interface{}{
		"@collection": instanceCollectionName,
		"agencyID":    agencyID,
		"tagName":     tagName,
	}

	cursor, err := r.executeQuery(ctx, agencyDB, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var instances []*models.AgencyInstance
	for cursor.HasMore() {
		var instance models.AgencyInstance
		_, err := cursor.ReadDocument(ctx, &instance)
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		instances = append(instances, &instance)
	}

	return instances, nil
}

// ListByState lists instances in a specific state
func (r *InstanceRepository) ListByState(ctx context.Context, agencyID string, state models.InstanceState, agencyDB string) ([]*models.AgencyInstance, error) {
	_, err := r.getInstanceCollection(ctx, agencyDB)
	if err != nil {
		return nil, err
	}

	query := `
		FOR instance IN @@collection
			FILTER instance.agency_id == @agencyID
			FILTER instance.state == @state
			FILTER instance.is_deleted != true
			SORT instance.started_at DESC
			RETURN instance
	`

	bindVars := map[string]interface{}{
		"@collection": instanceCollectionName,
		"agencyID":    agencyID,
		"state":       string(state),
	}

	cursor, err := r.executeQuery(ctx, agencyDB, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var instances []*models.AgencyInstance
	for cursor.HasMore() {
		var instance models.AgencyInstance
		_, err := cursor.ReadDocument(ctx, &instance)
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		instances = append(instances, &instance)
	}

	return instances, nil
}

// CreateAgent creates a new instance agent record
func (r *InstanceRepository) CreateAgent(ctx context.Context, agent *models.InstanceAgent, agencyDB string) error {
	collection, err := r.getInstanceAgentCollection(ctx, agencyDB)
	if err != nil {
		return err
	}

	// Generate key from agent_id
	agent.Key = agent.AgentID

	if agent.SpawnedAt.IsZero() {
		agent.SpawnedAt = time.Now()
	}

	meta, err := collection.CreateDocument(ctx, agent)
	if err != nil {
		if driver.IsConflict(err) {
			return fmt.Errorf("agent with ID '%s' already exists", agent.AgentID)
		}
		return fmt.Errorf("failed to create agent: %w", err)
	}

	agent.ID = meta.ID.String()
	agent.Rev = meta.Rev

	return nil
}

// ListAgentsByInstance lists all agents for an instance
func (r *InstanceRepository) ListAgentsByInstance(ctx context.Context, instanceID string, agencyDB string) ([]*models.InstanceAgent, error) {
	_, err := r.getInstanceAgentCollection(ctx, agencyDB)
	if err != nil {
		return nil, err
	}

	query := `
		FOR agent IN @@collection
			FILTER agent.instance_id == @instanceID
			SORT agent.spawned_at ASC
			RETURN agent
	`

	bindVars := map[string]interface{}{
		"@collection": instanceAgentCollectionName,
		"instanceID":  instanceID,
	}

	cursor, err := r.executeQuery(ctx, agencyDB, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var agents []*models.InstanceAgent
	for cursor.HasMore() {
		var agent models.InstanceAgent
		_, err := cursor.ReadDocument(ctx, &agent)
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		agents = append(agents, &agent)
	}

	return agents, nil
}

// DeleteAgentsByInstance deletes all agents for an instance
func (r *InstanceRepository) DeleteAgentsByInstance(ctx context.Context, instanceID string, agencyDB string) error {
	_, err := r.getInstanceAgentCollection(ctx, agencyDB)
	if err != nil {
		return err
	}

	query := `
		FOR agent IN @@collection
			FILTER agent.instance_id == @instanceID
			REMOVE agent IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": instanceAgentCollectionName,
		"instanceID":  instanceID,
	}

	cursor, err := r.executeQuery(ctx, agencyDB, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to delete agents: %w", err)
	}
	defer cursor.Close()

	return nil
}

// ExistsByName checks if an instance with the given name exists for an agency
func (r *InstanceRepository) ExistsByName(ctx context.Context, agencyID string, name string, agencyDB string) (bool, error) {
	_, err := r.getInstanceCollection(ctx, agencyDB)
	if err != nil {
		return false, err
	}

	query := `
		FOR instance IN @@collection
			FILTER instance.agency_id == @agencyID
			FILTER instance.name == @name
			FILTER instance.is_deleted == false
			LIMIT 1
			RETURN true
	`

	bindVars := map[string]interface{}{
		"@collection": instanceCollectionName,
		"agencyID":    agencyID,
		"name":        name,
	}

	cursor, err := r.executeQuery(ctx, agencyDB, query, bindVars)
	if err != nil {
		return false, err
	}
	defer cursor.Close()

	return cursor.HasMore(), nil
}

// executeQuery executes an AQL query
func (r *InstanceRepository) executeQuery(ctx context.Context, agencyDB string, query string, bindVars map[string]interface{}) (driver.Cursor, error) {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to access database: %w", err)
	}

	cursor, err := db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return cursor, nil
}
