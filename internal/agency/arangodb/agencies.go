package arangodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/arangodb/go-driver"
)

// Create creates a new agency in the database
func (r *Repository) Create(ctx context.Context, agencyDoc *agency.Agency) error {
	// Use ID as the document key
	agencyDoc.Key = agencyDoc.ID

	_, err := r.collection.CreateDocument(ctx, agencyDoc)
	if err != nil {
		return fmt.Errorf("failed to create agency document: %w", err)
	}

	return nil
}

// GetByID retrieves an agency by its ID
func (r *Repository) GetByID(ctx context.Context, id string) (*agency.Agency, error) {
	var agencyDoc agency.Agency
	_, err := r.collection.ReadDocument(ctx, id, &agencyDoc)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("agency not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read agency: %w", err)
	}

	return &agencyDoc, nil
}

// List retrieves agencies with optional filtering
func (r *Repository) List(ctx context.Context, filters agency.AgencyFilters) ([]*agency.Agency, error) {
	query, bindVars := buildListQuery(filters)

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var agencies []*agency.Agency
	for {
		var agencyDoc agency.Agency
		_, err := cursor.ReadDocument(ctx, &agencyDoc)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		agencies = append(agencies, &agencyDoc)
	}

	return agencies, nil
}

// Update updates an existing agency
func (r *Repository) Update(ctx context.Context, agencyDoc *agency.Agency) error {
	_, err := r.collection.UpdateDocument(ctx, agencyDoc.Key, agencyDoc)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("agency not found: %s", agencyDoc.ID)
		}
		return fmt.Errorf("failed to update agency: %w", err)
	}

	return nil
}

// Delete deletes an agency
func (r *Repository) Delete(ctx context.Context, id string) error {
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
func (r *Repository) GetStatistics(ctx context.Context, id string) (*agency.AgencyStatistics, error) {
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

	var stats agency.AgencyStatistics
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
func (r *Repository) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := r.collection.DocumentExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return exists, nil
}

// buildListQuery constructs an AQL query for listing agencies with filters
func buildListQuery(filters agency.AgencyFilters) (string, map[string]interface{}) {
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
