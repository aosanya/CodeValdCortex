package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/yourusername/codevaldcortex/internal/agency/models"
)

// IntroductionRepository handles database operations for agency introductions
type IntroductionRepository struct {
	db         driver.Database
	collection driver.Collection
}

// NewIntroductionRepository creates a new IntroductionRepository
func NewIntroductionRepository(db driver.Database) (*IntroductionRepository, error) {
	ctx := context.Background()

	// Ensure collection exists
	var collection driver.Collection
	exists, err := db.CollectionExists(ctx, "agency_introductions")
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !exists {
		collection, err = db.CreateCollection(ctx, "agency_introductions", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}

		// Create indexes
		if err := createIndexes(ctx, collection); err != nil {
			return nil, fmt.Errorf("failed to create indexes: %w", err)
		}
	} else {
		collection, err = db.Collection(ctx, "agency_introductions")
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	}

	return &IntroductionRepository{
		db:         db,
		collection: collection,
	}, nil
}

func createIndexes(ctx context.Context, collection driver.Collection) error {
	// Index on agency_id for fast lookups
	_, _, err := collection.EnsurePersistentIndex(ctx, []string{"agency_id"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
		Name:   "idx_agency_id",
	})
	if err != nil {
		return fmt.Errorf("failed to create agency_id index: %w", err)
	}

	// Index on template for filtering
	_, _, err = collection.EnsurePersistentIndex(ctx, []string{"template"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_template",
	})
	if err != nil {
		return fmt.Errorf("failed to create template index: %w", err)
	}

	return nil
}

// GetByAgencyID retrieves the introduction for a specific agency
func (r *IntroductionRepository) GetByAgencyID(ctx context.Context, agencyID string) (*models.AgencyIntroduction, error) {
	query := `
		FOR intro IN agency_introductions
		FILTER intro.agency_id == @agency_id
		RETURN intro
	`

	cursor, err := r.db.Query(ctx, query, map[string]interface{}{
		"agency_id": agencyID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query introduction: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, driver.ArangoError{ErrorNum: driver.ErrArangoDocumentNotFound}
	}

	var intro models.AgencyIntroduction
	_, err = cursor.ReadDocument(ctx, &intro)
	if err != nil {
		return nil, fmt.Errorf("failed to read introduction: %w", err)
	}

	return &intro, nil
}

// Create creates a new agency introduction
func (r *IntroductionRepository) Create(ctx context.Context, intro *models.AgencyIntroduction) error {
	// Validate before creating
	if err := intro.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	intro.CreatedAt = time.Now()
	intro.UpdatedAt = intro.CreatedAt

	meta, err := r.collection.CreateDocument(ctx, intro)
	if err != nil {
		return fmt.Errorf("failed to create introduction: %w", err)
	}

	intro.Key = meta.Key
	intro.ID = meta.ID.String()
	intro.Rev = meta.Rev

	return nil
}

// Update updates an existing agency introduction
func (r *IntroductionRepository) Update(ctx context.Context, intro *models.AgencyIntroduction) error {
	// Validate before updating
	if err := intro.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	intro.UpdatedAt = time.Now()

	meta, err := r.collection.UpdateDocument(ctx, intro.Key, intro)
	if err != nil {
		return fmt.Errorf("failed to update introduction: %w", err)
	}

	intro.Rev = meta.Rev

	return nil
}

// UpdateSection updates a specific section within an introduction
func (r *IntroductionRepository) UpdateSection(ctx context.Context, agencyID, sectionID string, section *models.IntroductionSection) error {
	// Validate section
	if err := section.Validate(); err != nil {
		return fmt.Errorf("section validation failed: %w", err)
	}

	section.Metadata.UpdatedAt = time.Now()

	query := `
		FOR intro IN agency_introductions
		FILTER intro.agency_id == @agency_id
		LET updated_sections = (
			FOR s IN intro.sections
			RETURN s.id == @section_id ? @new_section : s
		)
		UPDATE intro WITH {
			sections: updated_sections,
			updated_at: DATE_ISO8601(DATE_NOW())
		} IN agency_introductions
		RETURN NEW
	`

	cursor, err := r.db.Query(ctx, query, map[string]interface{}{
		"agency_id":   agencyID,
		"section_id":  sectionID,
		"new_section": section,
	})
	if err != nil {
		return fmt.Errorf("failed to update section: %w", err)
	}
	defer cursor.Close()

	return nil
}

// AddSection adds a new section to an introduction
func (r *IntroductionRepository) AddSection(ctx context.Context, agencyID string, section *models.IntroductionSection) error {
	// Validate section
	if err := section.Validate(); err != nil {
		return fmt.Errorf("section validation failed: %w", err)
	}

	now := time.Now()
	section.Metadata.CreatedAt = now
	section.Metadata.UpdatedAt = now

	query := `
		FOR intro IN agency_introductions
		FILTER intro.agency_id == @agency_id
		UPDATE intro WITH {
			sections: APPEND(intro.sections, [@new_section]),
			updated_at: DATE_ISO8601(DATE_NOW())
		} IN agency_introductions
		RETURN NEW
	`

	cursor, err := r.db.Query(ctx, query, map[string]interface{}{
		"agency_id":   agencyID,
		"new_section": section,
	})
	if err != nil {
		return fmt.Errorf("failed to add section: %w", err)
	}
	defer cursor.Close()

	return nil
}

// DeleteSection removes a section from an introduction
func (r *IntroductionRepository) DeleteSection(ctx context.Context, agencyID, sectionID string) error {
	query := `
		FOR intro IN agency_introductions
		FILTER intro.agency_id == @agency_id
		LET filtered_sections = (
			FOR s IN intro.sections
			FILTER s.id != @section_id
			RETURN s
		)
		UPDATE intro WITH {
			sections: filtered_sections,
			updated_at: DATE_ISO8601(DATE_NOW())
		} IN agency_introductions
		RETURN NEW
	`

	cursor, err := r.db.Query(ctx, query, map[string]interface{}{
		"agency_id":  agencyID,
		"section_id": sectionID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}
	defer cursor.Close()

	return nil
}

// ReorderSections updates the order of sections based on provided section IDs
func (r *IntroductionRepository) ReorderSections(ctx context.Context, agencyID string, sectionIDs []string) error {
	// First, get the current introduction
	intro, err := r.GetByAgencyID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get introduction: %w", err)
	}

	// Create a map of section ID to section for quick lookup
	sectionMap := make(map[string]*models.IntroductionSection)
	for i := range intro.Sections {
		sectionMap[intro.Sections[i].ID] = &intro.Sections[i]
	}

	// Reorder sections based on provided IDs
	reorderedSections := make([]models.IntroductionSection, 0, len(sectionIDs))
	for i, sectionID := range sectionIDs {
		section, exists := sectionMap[sectionID]
		if !exists {
			return fmt.Errorf("section ID not found: %s", sectionID)
		}
		section.Order = i + 1 // Orders are 1-indexed
		section.Metadata.UpdatedAt = time.Now()
		reorderedSections = append(reorderedSections, *section)
	}

	// Update the introduction with reordered sections
	intro.Sections = reorderedSections
	return r.Update(ctx, intro)
}

// Delete removes an introduction (used when deleting an agency)
func (r *IntroductionRepository) Delete(ctx context.Context, agencyID string) error {
	query := `
		FOR intro IN agency_introductions
		FILTER intro.agency_id == @agency_id
		REMOVE intro IN agency_introductions
	`

	_, err := r.db.Query(ctx, query, map[string]interface{}{
		"agency_id": agencyID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete introduction: %w", err)
	}

	return nil
}

// GetByTemplate retrieves all introductions using a specific template
func (r *IntroductionRepository) GetByTemplate(ctx context.Context, template string) ([]*models.AgencyIntroduction, error) {
	query := `
		FOR intro IN agency_introductions
		FILTER intro.template == @template
		RETURN intro
	`

	cursor, err := r.db.Query(ctx, query, map[string]interface{}{
		"template": template,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query by template: %w", err)
	}
	defer cursor.Close()

	var introductions []*models.AgencyIntroduction
	for cursor.HasMore() {
		var intro models.AgencyIntroduction
		_, err := cursor.ReadDocument(ctx, &intro)
		if err != nil {
			return nil, fmt.Errorf("failed to read introduction: %w", err)
		}
		introductions = append(introductions, &intro)
	}

	return introductions, nil
}
