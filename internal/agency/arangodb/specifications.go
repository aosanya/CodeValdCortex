package arangodb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	if cursor.HasMore() {
		// First read as raw map to handle type conversions
		var rawDoc map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &rawDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to read raw document: %w", err)
		}

		log.Printf("[MVP-054] GetSpecification: Read raw document from DB")
		if workItems, ok := rawDoc["work_items"].([]interface{}); ok {
			log.Printf("[MVP-054] GetSpecification: Raw document has %d work items", len(workItems))
			if len(workItems) > 0 {
				if wi, ok := workItems[0].(map[string]interface{}); ok {
					if deliv, ok := wi["deliverables_structured"]; ok {
						if delivArr, ok := deliv.([]interface{}); ok {
							log.Printf("[MVP-054] GetSpecification: First work item has %d deliverables_structured", len(delivArr))
						} else {
							log.Printf("[MVP-054] GetSpecification: First work item deliverables_structured is not an array: %T", deliv)
						}
					} else {
						log.Printf("[MVP-054] GetSpecification: First work item has NO deliverables_structured field")
					}
				}
			}
		}

		// Convert raw document to JSON and then unmarshal properly
		jsonData, err := json.Marshal(rawDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document to JSON: %w", err)
		}

		if err := json.Unmarshal(jsonData, &spec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal specification document: %w", err)
		}

		log.Printf("[MVP-054] GetSpecification: After unmarshal, spec has %d work items", len(spec.WorkItems))
		if len(spec.WorkItems) > 0 {
			log.Printf("[MVP-054] GetSpecification: First work item has %d deliverables_structured", len(spec.WorkItems[0].DeliverablesStructured))
		}

		return &spec, nil
	} else {
		// No specification exists, create a default one
		return r.CreateSpecification(ctx, agencyID, &models.CreateSpecificationRequest{
			Introduction: "",
			Goals:        []models.Goal{},
			WorkItems:    []models.WorkItem{},
			Roles:        []models.Role{},
			RACIMatrix:   nil,
		})
	}
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
	if req.Workflows != nil {
		existing.SetWorkflows(*req.Workflows, req.UpdatedBy)
	}
	if req.Roles != nil {
		existing.SetRoles(*req.Roles, req.UpdatedBy)
	}
	if req.RACIMatrix != nil {
		existing.SetRACIMatrix(req.RACIMatrix, req.UpdatedBy)
	}
	if req.AIPolicy != nil {
		existing.AIPolicy = req.AIPolicy
		existing.IncrementVersion(req.UpdatedBy)
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
			log.Printf("[MVP-054] Repository: Setting %d work items", len(items))
			for i, wi := range items {
				log.Printf("[MVP-054] Repository - WorkItem[%d] before SetWorkItems: code=%s, deliverables_structured=%d",
					i, wi.Code, len(wi.DeliverablesStructured))
			}
			existing.SetWorkItems(items, updatedBy)
			log.Printf("[MVP-054] Repository: After SetWorkItems, existing.WorkItems=%d", len(existing.WorkItems))
			for i, wi := range existing.WorkItems {
				log.Printf("[MVP-054] Repository - existing.WorkItems[%d]: code=%s, deliverables_structured=%d",
					i, wi.Code, len(wi.DeliverablesStructured))
			}
		} else {
			return nil, fmt.Errorf("invalid data type for work_items section")
		}
	case "workflows":
		if workflows, ok := data.([]models.Workflow); ok {
			existing.SetWorkflows(workflows, updatedBy)
		} else {
			return nil, fmt.Errorf("invalid data type for workflows section")
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

	// Log before database update
	log.Printf("[MVP-054] Repository: About to update document in database, key=%s", existing.Key)
	log.Printf("[MVP-054] Repository: Document has %d work items before DB update", len(existing.WorkItems))
	if len(existing.WorkItems) > 0 {
		log.Printf("[MVP-054] Repository: First work item deliverables_structured=%d", len(existing.WorkItems[0].DeliverablesStructured))
	}

	// Update in database
	meta, err := collection.UpdateDocument(ctx, existing.Key, existing)
	if err != nil {
		log.Printf("[MVP-054] Repository: UpdateDocument FAILED: %v", err)
		return nil, fmt.Errorf("failed to update specification document: %w", err)
	}

	log.Printf("[MVP-054] Repository: UpdateDocument SUCCESS, new rev=%s", meta.Rev)
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
