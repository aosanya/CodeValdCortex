package arangodb

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	driver "github.com/arangodb/go-driver"
)

const publicationsCollection = "agency_publications"

// PublicationRepository defines the interface for publication data access
type PublicationRepository interface {
	Create(ctx context.Context, pub *models.AgencyPublication) error
	GetByID(ctx context.Context, pubID string) (*models.AgencyPublication, error)
	ListByAgency(ctx context.Context, agencyID string) ([]*models.AgencyPublication, error)
	GetLatest(ctx context.Context, agencyID string) (*models.AgencyPublication, error)
	Update(ctx context.Context, pub *models.AgencyPublication) error
	GetByVersion(ctx context.Context, agencyID string, version string) (*models.AgencyPublication, error)
}

// publicationRepository is the ArangoDB implementation of PublicationRepository
type publicationRepository struct {
	db     driver.Database
	logger *slog.Logger
}

// NewPublicationRepository creates a new publication repository instance
func NewPublicationRepository(db driver.Database, logger *slog.Logger) PublicationRepository {
	return &publicationRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new publication into the database
func (r *publicationRepository) Create(ctx context.Context, pub *models.AgencyPublication) error {
	if pub == nil {
		return fmt.Errorf("publication is nil")
	}

	coll, err := r.db.Collection(ctx, publicationsCollection)
	if err != nil {
		r.logger.Error("failed to get publications collection", "error", err)
		return fmt.Errorf("failed to get publications collection: %w", err)
	}

	// Insert publication
	meta, err := coll.CreateDocument(ctx, pub)
	if err != nil {
		r.logger.Error("failed to create publication", "error", err, "agency_id", pub.AgencyID)
		return fmt.Errorf("failed to create publication: %w", err)
	}

	// Update publication with generated key
	pub.Key = meta.Key
	pub.ID = meta.ID.String()
	pub.Rev = meta.Rev

	r.logger.Info("publication created",
		"publication_id", pub.Key,
		"agency_id", pub.AgencyID,
		"version", pub.Version)

	return nil
}

// GetByID retrieves a publication by its document ID
func (r *publicationRepository) GetByID(ctx context.Context, pubID string) (*models.AgencyPublication, error) {
	if pubID == "" {
		return nil, fmt.Errorf("publication ID is empty")
	}

	coll, err := r.db.Collection(ctx, publicationsCollection)
	if err != nil {
		r.logger.Error("failed to get publications collection", "error", err)
		return nil, fmt.Errorf("failed to get publications collection: %w", err)
	}

	var pub models.AgencyPublication
	meta, err := coll.ReadDocument(ctx, pubID, &pub)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("publication not found: %s", pubID)
		}
		r.logger.Error("failed to read publication", "error", err, "publication_id", pubID)
		return nil, fmt.Errorf("failed to read publication: %w", err)
	}

	// Set metadata
	pub.Key = meta.Key
	pub.ID = meta.ID.String()
	pub.Rev = meta.Rev

	return &pub, nil
}

// ListByAgency retrieves all publications for a specific agency
func (r *publicationRepository) ListByAgency(ctx context.Context, agencyID string) ([]*models.AgencyPublication, error) {
	if agencyID == "" {
		return nil, fmt.Errorf("agency ID is empty")
	}

	query := `
		FOR pub IN @@collection
		FILTER pub.agency_id == @agencyID
		SORT pub.published_at DESC
		RETURN pub
	`

	bindVars := map[string]interface{}{
		"@collection": publicationsCollection,
		"agencyID":    agencyID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		r.logger.Error("failed to execute query", "error", err, "agency_id", agencyID)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	var publications []*models.AgencyPublication
	for cursor.HasMore() {
		var pub models.AgencyPublication
		meta, err := cursor.ReadDocument(ctx, &pub)
		if err != nil {
			r.logger.Error("failed to read publication document", "error", err)
			return nil, fmt.Errorf("failed to read publication document: %w", err)
		}

		// Set metadata
		pub.Key = meta.Key
		pub.ID = meta.ID.String()
		pub.Rev = meta.Rev

		publications = append(publications, &pub)
	}

	r.logger.Info("publications retrieved",
		"agency_id", agencyID,
		"count", len(publications))

	return publications, nil
}

// GetLatest retrieves the most recent publication for an agency
func (r *publicationRepository) GetLatest(ctx context.Context, agencyID string) (*models.AgencyPublication, error) {
	if agencyID == "" {
		return nil, fmt.Errorf("agency ID is empty")
	}

	query := `
		FOR pub IN @@collection
		FILTER pub.agency_id == @agencyID
		SORT pub.published_at DESC
		LIMIT 1
		RETURN pub
	`

	bindVars := map[string]interface{}{
		"@collection": publicationsCollection,
		"agencyID":    agencyID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		r.logger.Error("failed to execute query", "error", err, "agency_id", agencyID)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("no publications found for agency: %s", agencyID)
	}

	var pub models.AgencyPublication
	meta, err := cursor.ReadDocument(ctx, &pub)
	if err != nil {
		r.logger.Error("failed to read publication document", "error", err)
		return nil, fmt.Errorf("failed to read publication document: %w", err)
	}

	// Set metadata
	pub.Key = meta.Key
	pub.ID = meta.ID.String()
	pub.Rev = meta.Rev

	return &pub, nil
}

// Update modifies an existing publication
func (r *publicationRepository) Update(ctx context.Context, pub *models.AgencyPublication) error {
	if pub == nil {
		return fmt.Errorf("publication is nil")
	}

	if pub.Key == "" {
		return fmt.Errorf("publication key is empty")
	}

	coll, err := r.db.Collection(ctx, publicationsCollection)
	if err != nil {
		r.logger.Error("failed to get publications collection", "error", err)
		return fmt.Errorf("failed to get publications collection: %w", err)
	}

	// Marshal publication to map for update
	data, err := json.Marshal(pub)
	if err != nil {
		return fmt.Errorf("failed to marshal publication: %w", err)
	}

	var updateDoc map[string]interface{}
	if err := json.Unmarshal(data, &updateDoc); err != nil {
		return fmt.Errorf("failed to unmarshal publication to map: %w", err)
	}

	// Remove ArangoDB metadata fields (they shouldn't be updated)
	delete(updateDoc, "_key")
	delete(updateDoc, "_id")
	delete(updateDoc, "_rev")

	// Update document
	meta, err := coll.UpdateDocument(ctx, pub.Key, updateDoc)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("publication not found: %s", pub.Key)
		}
		r.logger.Error("failed to update publication", "error", err, "publication_id", pub.Key)
		return fmt.Errorf("failed to update publication: %w", err)
	}

	// Update revision
	pub.Rev = meta.Rev

	r.logger.Info("publication updated",
		"publication_id", pub.Key,
		"agency_id", pub.AgencyID)

	return nil
}

// GetByVersion retrieves a publication by agency ID and version string
func (r *publicationRepository) GetByVersion(ctx context.Context, agencyID string, version string) (*models.AgencyPublication, error) {
	if agencyID == "" {
		return nil, fmt.Errorf("agency ID is empty")
	}

	if version == "" {
		return nil, fmt.Errorf("version is empty")
	}

	query := `
		FOR pub IN @@collection
		FILTER pub.agency_id == @agencyID AND pub.version == @version
		LIMIT 1
		RETURN pub
	`

	bindVars := map[string]interface{}{
		"@collection": publicationsCollection,
		"agencyID":    agencyID,
		"version":     version,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		r.logger.Error("failed to execute query", "error", err, "agency_id", agencyID, "version", version)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("publication not found: agency=%s version=%s", agencyID, version)
	}

	var pub models.AgencyPublication
	meta, err := cursor.ReadDocument(ctx, &pub)
	if err != nil {
		r.logger.Error("failed to read publication document", "error", err)
		return nil, fmt.Errorf("failed to read publication document: %w", err)
	}

	// Set metadata
	pub.Key = meta.Key
	pub.ID = meta.ID.String()
	pub.Rev = meta.Rev

	return &pub, nil
}
