package agency

import (
	"context"
	"fmt"
	"strings"

	"github.com/arangodb/go-driver"
)

const (
	// CollectionName is the name of the agencies collection
	CollectionName = "agencies"
)

// arangoRepository implements Repository using ArangoDB
type arangoRepository struct {
	db         driver.Database
	collection driver.Collection
}

// NewArangoRepository creates a new ArangoDB repository for agencies
func NewArangoRepository(db driver.Database) (Repository, error) {
	// Ensure collection exists
	collection, err := ensureCollection(db)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	return &arangoRepository{
		db:         db,
		collection: collection,
	}, nil
}

// Create creates a new agency in the database
func (r *arangoRepository) Create(ctx context.Context, agency *Agency) error {
	// Use ID as the document key
	agency.Key = agency.ID

	_, err := r.collection.CreateDocument(ctx, agency)
	if err != nil {
		return fmt.Errorf("failed to create agency document: %w", err)
	}

	return nil
}

// GetByID retrieves an agency by its ID
func (r *arangoRepository) GetByID(ctx context.Context, id string) (*Agency, error) {
	var agency Agency
	_, err := r.collection.ReadDocument(ctx, id, &agency)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("agency not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read agency: %w", err)
	}

	return &agency, nil
}

// List retrieves agencies with optional filtering
func (r *arangoRepository) List(ctx context.Context, filters AgencyFilters) ([]*Agency, error) {
	query, bindVars := buildListQuery(filters)

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var agencies []*Agency
	for {
		var agency Agency
		_, err := cursor.ReadDocument(ctx, &agency)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		agencies = append(agencies, &agency)
	}

	return agencies, nil
}

// Update updates an existing agency
func (r *arangoRepository) Update(ctx context.Context, agency *Agency) error {
	_, err := r.collection.UpdateDocument(ctx, agency.Key, agency)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("agency not found: %s", agency.ID)
		}
		return fmt.Errorf("failed to update agency: %w", err)
	}

	return nil
}

// Delete deletes an agency
func (r *arangoRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.RemoveDocument(ctx, id)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("agency not found: %s", id)
		}
		return fmt.Errorf("failed to delete agency: %w", err)
	}

	return nil
}

// GetStatistics retrieves operational statistics for an agency
func (r *arangoRepository) GetStatistics(ctx context.Context, id string) (*AgencyStatistics, error) {
	// Query to get statistics from related collections
	query := `
		LET agency = DOCUMENT(CONCAT(@collection, '/', @id))
		LET agents = (
			FOR agent IN agents
			FILTER agent.agency_id == @id
			RETURN agent
		)
		LET tasks = (
			FOR task IN tasks
			FILTER task.agent_id IN agents[*]._key
			RETURN task
		)
		RETURN {
			agency_id: @id,
			active_agents: LENGTH(agents[* FILTER CURRENT.status == 'running']),
			inactive_agents: LENGTH(agents[* FILTER CURRENT.status != 'running']),
			total_tasks: LENGTH(tasks),
			completed_tasks: LENGTH(tasks[* FILTER CURRENT.status == 'completed']),
			failed_tasks: LENGTH(tasks[* FILTER CURRENT.status == 'failed']),
			last_activity: MAX(tasks[*].updated_at),
			uptime: 99.9
		}
	`

	bindVars := map[string]interface{}{
		"collection": CollectionName,
		"id":         id,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute statistics query: %w", err)
	}
	defer cursor.Close()

	var stats AgencyStatistics
	if cursor.HasMore() {
		_, err := cursor.ReadDocument(ctx, &stats)
		if err != nil {
			return nil, fmt.Errorf("failed to read statistics: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no statistics found for agency: %s", id)
	}

	return &stats, nil
}

// Exists checks if an agency exists
func (r *arangoRepository) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := r.collection.DocumentExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return exists, nil
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

// buildListQuery constructs an AQL query for listing agencies with filters
func buildListQuery(filters AgencyFilters) (string, map[string]interface{}) {
	var conditions []string
	bindVars := make(map[string]interface{})

	// Base query
	query := "FOR agency IN " + CollectionName

	// Apply filters
	if filters.Category != "" {
		conditions = append(conditions, "agency.category == @category")
		bindVars["category"] = filters.Category
	}

	if filters.Status != "" {
		conditions = append(conditions, "agency.status == @status")
		bindVars["status"] = filters.Status
	}

	if filters.Search != "" {
		conditions = append(conditions, "(CONTAINS(LOWER(agency.name), LOWER(@search)) OR CONTAINS(LOWER(agency.description), LOWER(@search)))")
		bindVars["search"] = filters.Search
	}

	if len(filters.Tags) > 0 {
		conditions = append(conditions, "LENGTH(INTERSECTION(agency.metadata.tags, @tags)) > 0")
		bindVars["tags"] = filters.Tags
	}

	// Add filter conditions
	if len(conditions) > 0 {
		query += " FILTER " + strings.Join(conditions, " AND ")
	}

	// Sort by name
	query += " SORT agency.name ASC"

	// Add pagination
	if filters.Limit > 0 {
		query += " LIMIT @offset, @limit"
		bindVars["offset"] = filters.Offset
		bindVars["limit"] = filters.Limit
	}

	query += " RETURN agency"

	return query, bindVars
}
