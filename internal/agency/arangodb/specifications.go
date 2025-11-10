package arangodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/arangodb/go-driver"
)

const (
	// SpecificationsCollection is the collection name for agency specifications
	SpecificationsCollection = "specifications"
)

// ensureSpecificationsCollection ensures the specifications collection exists in the agency database
func ensureSpecificationsCollection(ctx context.Context, db driver.Database) (driver.Collection, error) {
	exists, err := db.CollectionExists(ctx, SpecificationsCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	var collection driver.Collection
	if !exists {
		collection, err = db.CreateCollection(ctx, SpecificationsCollection, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}

		// Create index on agency_id for faster lookups
		_, _, err = collection.EnsurePersistentIndex(ctx, []string{"agency_id"}, &driver.EnsurePersistentIndexOptions{
			Unique: true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create agency_id index: %w", err)
		}
	} else {
		collection, err = db.Collection(ctx, SpecificationsCollection)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
	}

	return collection, nil
}

// GetSpecification retrieves the complete specification for an agency
func (r *Repository) GetSpecification(ctx context.Context, agencyID string) (*models.AgencySpecification, error) {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	db, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database for agency %s: %w", agencyID, err)
	}

	// Ensure collection exists
	_, err = ensureSpecificationsCollection(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure specifications collection: %w", err)
	}

	// Query for the specification document for this agency
	// There should be only one specification per agency
	query := `
		FOR spec IN @@collection
			FILTER spec.agency_id == @agencyID
			LIMIT 1
			RETURN spec
	`

	bindVars := map[string]interface{}{
		"@collection": SpecificationsCollection,
		"agencyID":    agencyID,
	}

	cursor, err := db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query specification: %w", err)
	}
	defer cursor.Close()

	var spec models.AgencySpecification
	if _, err := cursor.ReadDocument(ctx, &spec); err != nil {
		if driver.IsNoMoreDocuments(err) {
			// No specification exists, create a default one
			return r.CreateSpecification(ctx, agencyID, &models.CreateSpecificationRequest{
				Introduction: "",
				Goals:        []models.Goal{},
				WorkItems:    []models.WorkItem{},
				Roles:        []models.Role{},
				RACIMatrix:   nil,
			})
		}
		return nil, fmt.Errorf("failed to read specification document: %w", err)
	}

	return &spec, nil
}

// CreateSpecification creates a new specification document for an agency
func (r *Repository) CreateSpecification(ctx context.Context, agencyID string, req *models.CreateSpecificationRequest) (*models.AgencySpecification, error) {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	db, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database for agency %s: %w", agencyID, err)
	}

	collection, err := ensureSpecificationsCollection(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure specifications collection: %w", err)
	}

	now := time.Now()
	spec := &models.AgencySpecification{
		AgencyID:     agencyID,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
		UpdatedBy:    "system",
		Introduction: req.Introduction,
		Goals:        req.Goals,
		WorkItems:    req.WorkItems,
		Roles:        req.Roles,
		RACIMatrix:   req.RACIMatrix,
	}

	// Ensure arrays are initialized (not nil)
	if spec.Goals == nil {
		spec.Goals = []models.Goal{}
	}
	if spec.WorkItems == nil {
		spec.WorkItems = []models.WorkItem{}
	}
	if spec.Roles == nil {
		spec.Roles = []models.Role{}
	}

	meta, err := collection.CreateDocument(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to create specification document: %w", err)
	}

	spec.Key = meta.Key
	spec.ID = meta.ID.String()
	spec.Rev = meta.Rev

	return spec, nil
}

// UpdateSpecification updates the entire specification document
func (r *Repository) UpdateSpecification(ctx context.Context, agencyID string, req *models.SpecificationUpdateRequest) (*models.AgencySpecification, error) {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	db, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database for agency %s: %w", agencyID, err)
	}

	// Get existing specification
	existing, err := r.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing specification: %w", err)
	}

	collection, err := db.Collection(ctx, SpecificationsCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get specifications collection: %w", err)
	}

	// Apply updates
	if req.Introduction != nil {
		existing.UpdateIntroduction(*req.Introduction, req.UpdatedBy)
	}
	if req.Goals != nil {
		existing.SetGoals(*req.Goals, req.UpdatedBy)
	}
	if req.WorkItems != nil {
		existing.SetWorkItems(*req.WorkItems, req.UpdatedBy)
	}
	if req.Roles != nil {
		existing.SetRoles(*req.Roles, req.UpdatedBy)
	}
	if req.RACIMatrix != nil {
		existing.SetRACIMatrix(req.RACIMatrix, req.UpdatedBy)
	}

	// Update in database
	meta, err := collection.UpdateDocument(ctx, existing.Key, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update specification document: %w", err)
	}

	existing.Rev = meta.Rev

	return existing, nil
}

// PatchSpecificationSection updates a specific section of the specification
func (r *Repository) PatchSpecificationSection(ctx context.Context, agencyID, section string, data interface{}, updatedBy string) (*models.AgencySpecification, error) {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	db, err := r.client.Database(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database for agency %s: %w", agencyID, err)
	}

	// Get existing specification
	existing, err := r.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing specification: %w", err)
	}

	collection, err := db.Collection(ctx, SpecificationsCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get specifications collection: %w", err)
	}

	// Update the specific section
	switch section {
	case "introduction":
		if intro, ok := data.(string); ok {
			existing.UpdateIntroduction(intro, updatedBy)
		} else {
			return nil, fmt.Errorf("invalid data type for introduction section")
		}
	case "goals":
		if goals, ok := data.([]models.Goal); ok {
			existing.SetGoals(goals, updatedBy)
		} else {
			return nil, fmt.Errorf("invalid data type for goals section")
		}
	case "work_items":
		if items, ok := data.([]models.WorkItem); ok {
			existing.SetWorkItems(items, updatedBy)
		} else {
			return nil, fmt.Errorf("invalid data type for work_items section")
		}
	case "roles":
		if roles, ok := data.([]models.Role); ok {
			existing.SetRoles(roles, updatedBy)
		} else {
			return nil, fmt.Errorf("invalid data type for roles section")
		}
	case "raci_matrix":
		if matrix, ok := data.(*models.RACIMatrix); ok {
			existing.SetRACIMatrix(matrix, updatedBy)
		} else {
			return nil, fmt.Errorf("invalid data type for raci_matrix section")
		}
	default:
		return nil, fmt.Errorf("unknown section: %s", section)
	}

	// Update in database
	meta, err := collection.UpdateDocument(ctx, existing.Key, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update specification document: %w", err)
	}

	existing.Rev = meta.Rev

	return existing, nil
}

// DeleteSpecification deletes the specification for an agency
func (r *Repository) DeleteSpecification(ctx context.Context, agencyID string) error {
	// Get the agency-specific database
	agencyDoc, err := r.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	dbName := agencyDoc.Database
	if dbName == "" {
		dbName = agencyDoc.ID
	}

	db, err := r.client.Database(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to get database for agency %s: %w", agencyID, err)
	}

	spec, err := r.GetSpecification(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get specification: %w", err)
	}

	collection, err := db.Collection(ctx, SpecificationsCollection)
	if err != nil {
		return fmt.Errorf("failed to get specifications collection: %w", err)
	}

	_, err = collection.RemoveDocument(ctx, spec.Key)
	if err != nil {
		return fmt.Errorf("failed to delete specification: %w", err)
	}

	return nil
}
