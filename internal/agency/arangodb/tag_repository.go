package arangodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/arangodb/go-driver"
)

const (
	tagCollectionName = "agency_tags"
)

// TagRepository implements the tag repository using agency-specific databases
type TagRepository struct {
	client driver.Client // ArangoDB client to access different databases
}

// NewTagRepository creates a new tag repository
func NewTagRepository(client driver.Client) (*TagRepository, error) {
	return &TagRepository{
		client: client,
	}, nil
}

// getTagCollection gets or creates the tags collection in the agency's database
func (r *TagRepository) getTagCollection(ctx context.Context, agencyDB string) (driver.Collection, error) {
	// Get the agency's database
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to access agency database %s: %w", agencyDB, err)
	}

	// Check if collection exists
	exists, err := db.CollectionExists(ctx, tagCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if exists {
		collection, err = db.Collection(ctx, tagCollectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	} else {
		// Create the collection
		collection, err = db.CreateCollection(ctx, tagCollectionName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
	}

	return collection, nil
}

// Create creates a new tag in the agency's database
func (r *TagRepository) Create(ctx context.Context, tag *models.AgencyTag, agencyID string, agencyDB string) error {
	collection, err := r.getTagCollection(ctx, agencyDB)
	if err != nil {
		return err
	}

	// Generate document key from agency_id and name
	tag.Key = generateTagKey(agencyID, tag.Name)
	tag.AgencyID = agencyID

	if tag.CreatedAt.IsZero() {
		tag.CreatedAt = time.Now()
	}

	meta, err := collection.CreateDocument(ctx, tag)
	if err != nil {
		if driver.IsConflict(err) {
			return fmt.Errorf("tag with name '%s' already exists", tag.Name)
		}
		return fmt.Errorf("failed to create tag: %w", err)
	}

	tag.ID = meta.ID.String()
	tag.Rev = meta.Rev

	return nil
}

// GetByAgencyAndName retrieves a tag by agency ID and tag name
func (r *TagRepository) GetByAgencyAndName(ctx context.Context, agencyID, name string, agencyDB string) (*models.AgencyTag, error) {
	collection, err := r.getTagCollection(ctx, agencyDB)
	if err != nil {
		return nil, err
	}

	key := generateTagKey(agencyID, name)

	var tag models.AgencyTag
	_, err = collection.ReadDocument(ctx, key, &tag)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read tag: %w", err)
	}

	return &tag, nil
}

// List retrieves tags for an agency with optional filtering
func (r *TagRepository) List(ctx context.Context, agencyID string, agencyDB string, filters *services.TagFilters) ([]*models.AgencyTag, error) {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to access agency database: %w", err)
	}

	query, bindVars := r.buildListQuery(agencyID, filters)

	cursor, err := db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	// Initialize with empty slice instead of nil to ensure JSON marshals to [] not null
	tags := make([]*models.AgencyTag, 0)
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

// Delete removes a tag from the agency's database
func (r *TagRepository) Delete(ctx context.Context, agencyID, name string, agencyDB string) error {
	collection, err := r.getTagCollection(ctx, agencyDB)
	if err != nil {
		return err
	}

	key := generateTagKey(agencyID, name)

	_, err = collection.RemoveDocument(ctx, key)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return fmt.Errorf("tag not found: %s", name)
		}
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// buildListQuery constructs the AQL query for listing tags with filters
func (r *TagRepository) buildListQuery(agencyID string, filters *services.TagFilters) (string, map[string]interface{}) {
	bindVars := map[string]interface{}{
		"agency_id": agencyID,
	}

	var queryParts []string
	queryParts = append(queryParts, fmt.Sprintf("FOR tag IN %s", tagCollectionName))
	queryParts = append(queryParts, "FILTER tag.agency_id == @agency_id")

	if filters != nil {
		if filters.Type != "" {
			queryParts = append(queryParts, "FILTER tag.type == @type")
			bindVars["type"] = filters.Type
		}

		if filters.NameLike != "" {
			queryParts = append(queryParts, "FILTER LIKE(tag.name, @name_pattern, true)")
			bindVars["name_pattern"] = fmt.Sprintf("%%%s%%", filters.NameLike)
		}

		if filters.FromDate != nil {
			queryParts = append(queryParts, "FILTER tag.created_at >= @from_date")
			bindVars["from_date"] = filters.FromDate
		}

		if filters.ToDate != nil {
			queryParts = append(queryParts, "FILTER tag.created_at <= @to_date")
			bindVars["to_date"] = filters.ToDate
		}
	}

	queryParts = append(queryParts, "SORT tag.created_at DESC")

	if filters != nil {
		if filters.Limit > 0 {
			queryParts = append(queryParts, "LIMIT @offset, @limit")
			bindVars["offset"] = filters.Offset
			bindVars["limit"] = filters.Limit
		}
	}

	queryParts = append(queryParts, "RETURN tag")

	query := strings.Join(queryParts, "\n")
	return query, bindVars
}

// generateTagKey creates a unique key for a tag
func generateTagKey(agencyID, tagName string) string {
	cleanID := strings.TrimPrefix(agencyID, "agency_")
	cleanID = strings.TrimPrefix(cleanID, "agencies/")

	replacer := strings.NewReplacer(
		"/", "_",
		" ", "_",
		"-", "_",
	)
	cleanName := replacer.Replace(tagName)

	return fmt.Sprintf("%s_%s", cleanID, cleanName)
}
