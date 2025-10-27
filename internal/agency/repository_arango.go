package agency

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
)

const (
	// CollectionName is the name of the agencies collection
	CollectionName = "agencies"
)

// arangoRepository implements Repository using ArangoDB
type arangoRepository struct {
	client     driver.Client
	db         driver.Database
	collection driver.Collection
}

// NewArangoRepository creates a new ArangoDB repository for agencies
func NewArangoRepository(client driver.Client, db driver.Database) (Repository, error) {
	// Ensure collection exists
	collection, err := ensureCollection(db)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	return &arangoRepository{
		client:     client,
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

// GetOverview retrieves the overview document for an agency
func (r *arangoRepository) GetOverview(ctx context.Context, agencyID string) (*Overview, error) {
	// Get the agency-specific database
	agency, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agency.Database
	if dbName == "" {
		dbName = agency.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure overview collection exists
	overviewColl, err := ensureOverviewCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure overview collection: %w", err)
	}

	// Try to read the overview document (using "main" as the key)
	var overview Overview
	_, err = overviewColl.ReadDocument(ctx, "main", &overview)
	if err != nil {
		if driver.IsNotFound(err) {
			// Create a default overview if it doesn't exist
			overview = Overview{
				Key:          "main",
				AgencyID:     agencyID,
				Introduction: "",
				UpdatedAt:    time.Now(),
			}
			_, err = overviewColl.CreateDocument(ctx, &overview)
			if err != nil {
				return nil, fmt.Errorf("failed to create default overview: %w", err)
			}
			return &overview, nil
		}
		return nil, fmt.Errorf("failed to read overview: %w", err)
	}

	return &overview, nil
}

// UpdateOverview updates the overview document for an agency
func (r *arangoRepository) UpdateOverview(ctx context.Context, overview *Overview) error {
	// Get the agency-specific database
	agency, err := r.GetByID(ctx, overview.AgencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agency.Database
	if dbName == "" {
		dbName = agency.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure overview collection exists
	overviewColl, err := ensureOverviewCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure overview collection: %w", err)
	}

	// Set the key and updated timestamp
	overview.Key = "main"
	overview.UpdatedAt = time.Now()

	// Update or create the document
	_, err = overviewColl.UpdateDocument(ctx, overview.Key, overview)
	if err != nil {
		if driver.IsNotFound(err) {
			// Create if it doesn't exist
			_, err = overviewColl.CreateDocument(ctx, overview)
			if err != nil {
				return fmt.Errorf("failed to create overview: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to update overview: %w", err)
	}

	return nil
}

// ensureOverviewCollection ensures the overview collection exists in an agency database
func ensureOverviewCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	const collectionName = "overview"

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

// CreateProblem creates a new problem in the agency's problems collection
func (r *arangoRepository) CreateProblem(ctx context.Context, problem *Problem) error {
	// Get the agency-specific database
	agency, err := r.GetByID(ctx, problem.AgencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agency.Database
	if dbName == "" {
		dbName = agency.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Set timestamps
	now := time.Now()
	problem.CreatedAt = now
	problem.UpdatedAt = now

	// If number is not set, get the next number
	if problem.Number == 0 {
		// Count existing problems
		query := "FOR p IN problems FILTER p.agency_id == @agencyId COLLECT WITH COUNT INTO length RETURN length"
		cursor, err := agencyDB.Query(ctx, query, map[string]interface{}{"agencyId": problem.AgencyID})
		if err != nil {
			return fmt.Errorf("failed to count problems: %w", err)
		}
		defer cursor.Close()

		var count int
		if cursor.HasMore() {
			_, err := cursor.ReadDocument(ctx, &count)
			if err != nil {
				return fmt.Errorf("failed to read count: %w", err)
			}
		}
		problem.Number = count + 1
	}

	// Create the document
	meta, err := problemsColl.CreateDocument(ctx, problem)
	if err != nil {
		return fmt.Errorf("failed to create problem: %w", err)
	}

	problem.Key = meta.Key
	return nil
}

// GetProblems retrieves all problems for an agency
func (r *arangoRepository) GetProblems(ctx context.Context, agencyID string) ([]*Problem, error) {
	// Get the agency-specific database
	agency, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agency.Database
	if dbName == "" {
		dbName = agency.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Query all problems for this agency, ordered by number
	query := "FOR p IN @@collection FILTER p.agency_id == @agencyId SORT p.number ASC RETURN p"
	bindVars := map[string]interface{}{
		"@collection": problemsColl.Name(),
		"agencyId":    agencyID,
	}

	cursor, err := agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query problems: %w", err)
	}
	defer cursor.Close()

	var problems []*Problem
	for cursor.HasMore() {
		var problem Problem
		_, err := cursor.ReadDocument(ctx, &problem)
		if err != nil {
			return nil, fmt.Errorf("failed to read problem: %w", err)
		}
		problems = append(problems, &problem)
	}

	return problems, nil
}

// GetProblem retrieves a single problem by key
func (r *arangoRepository) GetProblem(ctx context.Context, agencyID string, key string) (*Problem, error) {
	// Get the agency-specific database
	agency, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agency.Database
	if dbName == "" {
		dbName = agency.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Read the document
	var problem Problem
	_, err = problemsColl.ReadDocument(ctx, key, &problem)
	if err != nil {
		return nil, fmt.Errorf("failed to read problem: %w", err)
	}

	return &problem, nil
}

// UpdateProblem updates an existing problem
func (r *arangoRepository) UpdateProblem(ctx context.Context, problem *Problem) error {
	// Get the agency-specific database
	agency, err := r.GetByID(ctx, problem.AgencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agency.Database
	if dbName == "" {
		dbName = agency.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Update timestamp
	problem.UpdatedAt = time.Now()

	// Update the document
	_, err = problemsColl.UpdateDocument(ctx, problem.Key, problem)
	if err != nil {
		return fmt.Errorf("failed to update problem: %w", err)
	}

	return nil
}

// DeleteProblem deletes a problem and renumbers remaining problems
func (r *arangoRepository) DeleteProblem(ctx context.Context, agencyID string, key string) error {
	// Get the agency-specific database
	agency, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency ID as database name if not set
	dbName := agency.Database
	if dbName == "" {
		dbName = agency.ID
	}

	// Get connection to agency's database
	agencyDB, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to agency database: %w", err)
	}

	// Ensure problems collection exists
	problemsColl, err := ensureProblemsCollection(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to ensure problems collection: %w", err)
	}

	// Get the problem to find its number
	var problemToDelete Problem
	_, err = problemsColl.ReadDocument(ctx, key, &problemToDelete)
	if err != nil {
		return fmt.Errorf("failed to read problem: %w", err)
	}

	// Delete the document
	_, err = problemsColl.RemoveDocument(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete problem: %w", err)
	}

	// Renumber problems with higher numbers
	query := `
		FOR p IN @@collection 
		FILTER p.agency_id == @agencyId AND p.number > @deletedNumber
		UPDATE p WITH { number: p.number - 1 } IN @@collection
	`
	bindVars := map[string]interface{}{
		"@collection":   problemsColl.Name(),
		"agencyId":      agencyID,
		"deletedNumber": problemToDelete.Number,
	}

	_, err = agencyDB.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to renumber problems: %w", err)
	}

	return nil
}

// ensureProblemsCollection ensures the problems collection exists in an agency database
func ensureProblemsCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	const collectionName = "problems"

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
