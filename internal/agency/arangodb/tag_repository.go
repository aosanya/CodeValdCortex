package arangodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/arangodb/go-driver"
)

const (
	TagCollectionName = "agency_tags"
)

// TagRepository implements the tag repository using ArangoDB
type TagRepository struct {
	db         driver.Database
	collection driver.Collection
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db driver.Database) (*TagRepository, error) {
	// Get or create tags collection
	var collection driver.Collection
	exists, err := db.CollectionExists(context.Background(), TagCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	if exists {
		collection, err = db.Collection(context.Background(), TagCollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	} else {
		collection, err = db.CreateCollection(context.Background(), TagCollectionName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
	}

	return &TagRepository{
		db:         db,
		collection: collection,
	}, nil
}

// Create creates a new tag
func (r *TagRepository) Create(ctx context.Context, tag *models.AgencyTag) error {
	// Generate document key from agency_id and name
	tag.Key = generateTagKey(tag.AgencyID, tag.Name)

	meta, err := r.collection.CreateDocument(ctx, tag)
	if err != nil {
		if driver.IsConflict(err) {
			return fmt.Errorf("tag with name '%s' already exists for agency", tag.Name)
		}
		return fmt.Errorf("failed to create tag: %w", err)
	}

	// Set the full ID
	tag.ID = meta.ID.String()
	tag.Rev = meta.Rev

	return nil
}

// GetByID retrieves a tag by its document ID
func (r *TagRepository) GetByID(ctx context.Context, tagID string) (*models.AgencyTag, error) {
	// Extract key from full ID if needed
	key := tagID
	if strings.Contains(tagID, "/") {
		parts := strings.Split(tagID, "/")
		if len(parts) == 2 {
			key = parts[1]
		}
	}

	var tag models.AgencyTag
	_, err := r.collection.ReadDocument(ctx, key, &tag)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, nil // Return nil instead of error for not found
		}
		return nil, fmt.Errorf("failed to read tag: %w", err)
	}

	return &tag, nil
}

// GetByAgencyAndName retrieves a tag by agency ID and tag name
func (r *TagRepository) GetByAgencyAndName(ctx context.Context, agencyID, name string) (*models.AgencyTag, error) {
	key := generateTagKey(agencyID, name)

	var tag models.AgencyTag
	_, err := r.collection.ReadDocument(ctx, key, &tag)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, nil // Return nil instead of error for not found
		}
		return nil, fmt.Errorf("failed to read tag: %w", err)
	}

	return &tag, nil
}

// List retrieves tags for an agency with optional filtering
func (r *TagRepository) List(ctx context.Context, agencyID string, filters *services.TagFilters) ([]*models.AgencyTag, error) {
	query, bindVars := r.buildListQuery(agencyID, filters)

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var tags []*models.AgencyTag
	for {
		var tag models.AgencyTag
		_, err := cursor.ReadDocument(ctx, &tag)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read document: %w", err)
		}
		tags = append(tags, &tag)
	}

	return tags, nil
}

// Delete removes a tag
func (r *TagRepository) Delete(ctx context.Context, agencyID, name string) error {
	key := generateTagKey(agencyID, name)

	_, err := r.collection.RemoveDocument(ctx, key)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("tag not found: %s", name)
		}
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// buildListQuery constructs the AQL query for listing tags with filters
func (r *TagRepository) buildListQuery(agencyID string, filters *services.TagFilters) (string, map[string]interface{}) {
	bindVars := map[string]interface{}{
		"collection": TagCollectionName,
		"agency_id":  agencyID,
	}

	// Start building the query
	var queryParts []string
	queryParts = append(queryParts, fmt.Sprintf("FOR tag IN %s", TagCollectionName))
	queryParts = append(queryParts, "FILTER tag.agency_id == @agency_id")

	// Add type filter
	if filters != nil && filters.Type != "" {
		queryParts = append(queryParts, "FILTER tag.type == @type")
		bindVars["type"] = filters.Type
	}

	// Add name filter (LIKE)
	if filters != nil && filters.NameLike != "" {
		queryParts = append(queryParts, "FILTER LIKE(tag.name, @name_pattern, true)")
		bindVars["name_pattern"] = fmt.Sprintf("%%%s%%", filters.NameLike)
	}

	// Add date range filters
	if filters != nil && filters.FromDate != nil {
		queryParts = append(queryParts, "FILTER tag.created_at >= @from_date")
		bindVars["from_date"] = filters.FromDate
	}

	if filters != nil && filters.ToDate != nil {
		queryParts = append(queryParts, "FILTER tag.created_at <= @to_date")
		bindVars["to_date"] = filters.ToDate
	}

	// Sort by created_at descending (newest first)
	queryParts = append(queryParts, "SORT tag.created_at DESC")

	// Add pagination
	if filters != nil {
		if filters.Limit > 0 {
			queryParts = append(queryParts, "LIMIT @offset, @limit")
			bindVars["limit"] = filters.Limit
			bindVars["offset"] = filters.Offset
		}
	}

	queryParts = append(queryParts, "RETURN tag")

	query := strings.Join(queryParts, "\n")
	return query, bindVars
}

// generateTagKey creates a unique key for a tag (agency_id + name)
func generateTagKey(agencyID, name string) string {
	// Remove "agency_" prefix if present
	cleanID := strings.TrimPrefix(agencyID, "agency_")
	// Create key: agencyid_tagname (sanitize to be URL-safe)
	key := fmt.Sprintf("%s_%s", cleanID, sanitizeName(name))
	return key
}

// sanitizeName removes special characters from tag name for key generation
func sanitizeName(name string) string {
	// Replace spaces and special characters with underscores
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		".", "_",
	)
	return replacer.Replace(name)
}
