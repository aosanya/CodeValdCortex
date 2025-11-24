package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// TagService defines the interface for agency tag management
type TagService interface {
	// CreateTag creates a tag from current agency state
	CreateTag(ctx context.Context, agencyID string, req *CreateTagRequest) (*models.AgencyTag, error)

	// ListTags lists all tags for an agency
	ListTags(ctx context.Context, agencyID string, filters *TagFilters) ([]*models.AgencyTag, error)

	// GetTag retrieves a specific tag
	GetTag(ctx context.Context, agencyID string, tagName string) (*models.AgencyTag, error)

	// CompareTags generates a diff between two tags (now requires agency ID and tag names)
	CompareTags(ctx context.Context, agencyID string, tagName1, tagName2 string) (*TagComparison, error)

	// RestoreFromTag overwrites agency draft with tag snapshot
	RestoreFromTag(ctx context.Context, agencyID string, tagName string) error

	// DeleteTag removes a tag
	DeleteTag(ctx context.Context, agencyID string, tagName string) error
}

// CreateTagRequest represents a request to create a tag
type CreateTagRequest struct {
	Name        string            `json:"name" binding:"required"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description" binding:"required"`
	Type        models.TagType    `json:"type" binding:"required"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedBy   string            `json:"created_by,omitempty"`
}

// TagFilters defines criteria for filtering tags
type TagFilters struct {
	Type     models.TagType `json:"type,omitempty"`
	NameLike string         `json:"name_like,omitempty"`
	FromDate *time.Time     `json:"from_date,omitempty"`
	ToDate   *time.Time     `json:"to_date,omitempty"`
	Limit    int            `json:"limit,omitempty"`
	Offset   int            `json:"offset,omitempty"`
}

// TagComparison represents a comparison between two tags
type TagComparison struct {
	Tag1        *models.AgencyTag `json:"tag1"`
	Tag2        *models.AgencyTag `json:"tag2"`
	Differences []TagDifference   `json:"differences"`
	Summary     string            `json:"summary"`
}

// TagDifference represents a single difference between tags
type TagDifference struct {
	Path     string      `json:"path"`
	Type     string      `json:"type"` // added, removed, modified
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}

// TagRepository defines the interface for tag data persistence
type TagRepository interface {
	Create(ctx context.Context, tag *models.AgencyTag, agencyID string, agencyDB string) error
	GetByAgencyAndName(ctx context.Context, agencyID, name string, agencyDB string) (*models.AgencyTag, error)
	List(ctx context.Context, agencyID string, agencyDB string, filters *TagFilters) ([]*models.AgencyTag, error)
	Delete(ctx context.Context, agencyID, name string, agencyDB string) error
}

// tagService implements the TagService interface
type tagService struct {
	tagRepo    TagRepository
	agencyRepo agency.Repository
	logger     *slog.Logger
}

// NewTagService creates a new tag service
func NewTagService(tagRepo TagRepository, agencyRepo agency.Repository, logger *slog.Logger) TagService {
	return &tagService{
		tagRepo:    tagRepo,
		agencyRepo: agencyRepo,
		logger:     logger,
	}
}

// CreateTag creates a tag from the current agency state
func (s *tagService) CreateTag(ctx context.Context, agencyID string, req *CreateTagRequest) (*models.AgencyTag, error) {
	// Validate tag type
	if !models.IsValidTagType(req.Type) {
		return nil, fmt.Errorf("invalid tag type: %s", req.Type)
	}

	// Retrieve current agency state
	agencyDoc, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve agency: %w", err)
	}

	// Check if tag name already exists
	existing, _ := s.tagRepo.GetByAgencyAndName(ctx, agencyID, req.Name, agencyDoc.Database)
	if existing != nil {
		return nil, fmt.Errorf("tag with name '%s' already exists", req.Name)
	}

	// Retrieve agency specification
	spec, err := s.agencyRepo.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve specification: %w", err)
	}

	// Generate snapshot
	snapshot, err := s.generateSnapshot(agencyDoc, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to generate snapshot: %w", err)
	}

	// Generate content hash (SHA-256)
	contentHash, err := s.generateContentHash(snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content hash: %w", err)
	}

	// Create tag
	customFields := make(map[string]interface{})
	for k, v := range req.Metadata {
		customFields[k] = v
	}

	tag := &models.AgencyTag{
		Name:        req.Name,
		Type:        req.Type,
		Version:     req.Version,
		Description: req.Description,
		Snapshot:    *snapshot,
		SHA:         contentHash,
		Metadata: models.TagMetadata{
			CustomFields: customFields,
		},
		CreatedAt: time.Now(),
		CreatedBy: req.CreatedBy,
	}

	// Save tag to repository (in agency-specific database)
	if err := s.tagRepo.Create(ctx, tag, agencyID, agencyDoc.Database); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	s.logger.Info("created tag",
		"agency_id", agencyID,
		"agency_db", agencyDoc.Database,
		"tag_name", req.Name,
		"tag_type", req.Type,
		"content_hash", contentHash,
	)

	return tag, nil
}

// ListTags retrieves tags for an agency with optional filtering
func (s *tagService) ListTags(ctx context.Context, agencyID string, filters *TagFilters) ([]*models.AgencyTag, error) {
	// Get agency to retrieve database name
	agencyDoc, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve agency: %w", err)
	}

	tags, err := s.tagRepo.List(ctx, agencyID, agencyDoc.Database, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return tags, nil
}

// GetTag retrieves a specific tag by name
func (s *tagService) GetTag(ctx context.Context, agencyID string, tagName string) (*models.AgencyTag, error) {
	// Get agency to retrieve database name
	agencyDoc, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve agency: %w", err)
	}

	tag, err := s.tagRepo.GetByAgencyAndName(ctx, agencyID, tagName, agencyDoc.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	if tag == nil {
		return nil, fmt.Errorf("tag '%s' not found", tagName)
	}
	return tag, nil
}

// CompareTags generates a diff between two tags
func (s *tagService) CompareTags(ctx context.Context, agencyID string, tagName1, tagName2 string) (*TagComparison, error) {
	// Get agency to retrieve database name
	agencyDoc, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve agency: %w", err)
	}

	// Retrieve both tags by name
	tag1, err := s.tagRepo.GetByAgencyAndName(ctx, agencyID, tagName1, agencyDoc.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tag1: %w", err)
	}
	if tag1 == nil {
		return nil, fmt.Errorf("tag1 not found: %s", tagName1)
	}

	tag2, err := s.tagRepo.GetByAgencyAndName(ctx, agencyID, tagName2, agencyDoc.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tag2: %w", err)
	}
	if tag2 == nil {
		return nil, fmt.Errorf("tag2 not found: %s", tagName2)
	}

	// Generate differences
	differences := s.compareSnapshots(&tag1.Snapshot, &tag2.Snapshot)

	// Generate summary
	added := 0
	removed := 0
	modified := 0
	for _, diff := range differences {
		switch diff.Type {
		case "added":
			added++
		case "removed":
			removed++
		case "modified":
			modified++
		}
	}

	summary := fmt.Sprintf("%d changes: %d added, %d removed, %d modified",
		len(differences), added, removed, modified)

	return &TagComparison{
		Tag1:        tag1,
		Tag2:        tag2,
		Differences: differences,
		Summary:     summary,
	}, nil
}

// RestoreFromTag overwrites the agency draft with a tag's snapshot
func (s *tagService) RestoreFromTag(ctx context.Context, agencyID string, tagName string) error {
	// Retrieve current agency
	agencyDoc, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to retrieve agency: %w", err)
	}

	// Retrieve tag
	tag, err := s.tagRepo.GetByAgencyAndName(ctx, agencyID, tagName, agencyDoc.Database)
	if err != nil {
		return fmt.Errorf("failed to retrieve tag: %w", err)
	}
	if tag == nil {
		return fmt.Errorf("tag '%s' not found", tagName)
	}

	// Verify agency is in draft state (can only restore to draft)
	if agencyDoc.State != models.AgencyStateDraft {
		return fmt.Errorf("can only restore to draft agency, current state: %s", agencyDoc.State)
	}

	// Restore agency configuration from snapshot
	agencyDoc.Settings = tag.Snapshot.Settings
	agencyDoc.Metadata = tag.Snapshot.Metadata
	agencyDoc.UpdatedAt = time.Now()
	agencyDoc.UpdatedBy = fmt.Sprintf("restored_from_tag:%s", tagName)

	// Update agency
	if err := s.agencyRepo.Update(ctx, agencyDoc); err != nil {
		return fmt.Errorf("failed to update agency: %w", err)
	}

	// Restore specification
	updateReq := &models.SpecificationUpdateRequest{
		Introduction: &tag.Snapshot.Specification.Introduction,
		Goals:        &tag.Snapshot.Specification.Goals,
		WorkItems:    &tag.Snapshot.Specification.WorkItems,
		Roles:        &tag.Snapshot.Specification.Roles,
		RACIMatrix:   tag.Snapshot.Specification.RACIMatrix,
		Workflows:    &tag.Snapshot.Specification.Workflows,
		UpdatedBy:    fmt.Sprintf("restored_from_tag:%s", tagName),
	}

	if _, err := s.agencyRepo.UpdateSpecification(ctx, agencyID, updateReq); err != nil {
		return fmt.Errorf("failed to update specification: %w", err)
	}

	s.logger.Info("restored agency from tag",
		"agency_id", agencyID,
		"tag_name", tagName,
		"tag_version", tag.Version,
	)

	return nil
}

// DeleteTag removes a tag
func (s *tagService) DeleteTag(ctx context.Context, agencyID string, tagName string) error {
	// Get agency to retrieve database name
	agencyDoc, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to retrieve agency: %w", err)
	}

	// Check if tag exists
	tag, err := s.tagRepo.GetByAgencyAndName(ctx, agencyID, tagName, agencyDoc.Database)
	if err != nil {
		return fmt.Errorf("failed to check tag: %w", err)
	}
	if tag == nil {
		return fmt.Errorf("tag '%s' not found", tagName)
	}

	// Delete tag
	if err := s.tagRepo.Delete(ctx, agencyID, tagName, agencyDoc.Database); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	s.logger.Info("deleted tag",
		"agency_id", agencyID,
		"tag_name", tagName,
	)

	return nil
}

// generateSnapshot creates a deep copy of the agency state
func (s *tagService) generateSnapshot(agencyDoc *models.Agency, spec *models.AgencySpecification) (*models.AgencySnapshot, error) {
	// Create snapshot with deep copies
	snapshot := &models.AgencySnapshot{
		Specification: *s.deepCopySpecification(spec),
		Settings:      agencyDoc.Settings,
		Metadata:      agencyDoc.Metadata,
	}

	// Note: AIPolicy conversion from models.Policy to models.AIPolicy
	// is handled separately if needed. For now, we'll leave AIPolicy empty
	// as the Policy struct in specification is different from AIPolicy in publication

	return snapshot, nil
}

// deepCopySpecification creates a deep copy of the agency specification
func (s *tagService) deepCopySpecification(spec *models.AgencySpecification) *models.AgencySpecification {
	if spec == nil {
		return nil
	}

	// Create new specification
	copied := &models.AgencySpecification{
		AgencyID:     spec.AgencyID,
		Version:      spec.Version,
		CreatedAt:    spec.CreatedAt,
		UpdatedAt:    spec.UpdatedAt,
		UpdatedBy:    spec.UpdatedBy,
		Introduction: spec.Introduction,
	}

	// Deep copy slices
	if spec.Goals != nil {
		copied.Goals = make([]models.Goal, len(spec.Goals))
		copy(copied.Goals, spec.Goals)
	}

	if spec.WorkItems != nil {
		copied.WorkItems = make([]models.WorkItem, len(spec.WorkItems))
		copy(copied.WorkItems, spec.WorkItems)
	}

	if spec.Roles != nil {
		copied.Roles = make([]models.Role, len(spec.Roles))
		copy(copied.Roles, spec.Roles)
	}

	if spec.Workflows != nil {
		copied.Workflows = make([]models.Workflow, len(spec.Workflows))
		copy(copied.Workflows, spec.Workflows)
	}

	// Deep copy RACI matrix
	if spec.RACIMatrix != nil {
		copiedMatrix := &models.RACIMatrix{}
		if spec.RACIMatrix.Assignments != nil {
			copiedMatrix.Assignments = make([]models.RACIRoleAssignment, len(spec.RACIMatrix.Assignments))
			copy(copiedMatrix.Assignments, spec.RACIMatrix.Assignments)
		}
		copied.RACIMatrix = copiedMatrix
	}

	// Deep copy AI policy (using models.Policy from specification.go)
	if spec.AIPolicy != nil {
		copied.AIPolicy = &models.Policy{
			Version:    spec.AIPolicy.Version,
			Owner:      spec.AIPolicy.Owner,
			Stance:     s.deepCopyMap(spec.AIPolicy.Stance),
			Models:     s.deepCopyMap(spec.AIPolicy.Models),
			Autonomy:   s.deepCopyMap(spec.AIPolicy.Autonomy),
			DataAccess: s.deepCopyMap(spec.AIPolicy.DataAccess),
			Actions:    s.deepCopyMap(spec.AIPolicy.Actions),
			Risk:       s.deepCopyMap(spec.AIPolicy.Risk),
			Compliance: s.deepCopyMap(spec.AIPolicy.Compliance),
			Monitoring: s.deepCopyMap(spec.AIPolicy.Monitoring),
		}
	}

	return copied
}

// deepCopyMap creates a deep copy of a map
func (s *tagService) deepCopyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}

	// Use JSON marshaling for deep copy (simple but effective)
	data, err := json.Marshal(m)
	if err != nil {
		return make(map[string]interface{})
	}

	var copied map[string]interface{}
	if err := json.Unmarshal(data, &copied); err != nil {
		return make(map[string]interface{})
	}

	return copied
}

// generateContentHash generates a SHA-256 hash of the snapshot
func (s *tagService) generateContentHash(snapshot *models.AgencySnapshot) (string, error) {
	// Serialize to JSON with sorted keys (canonical form)
	data, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(data)

	// Return hex-encoded hash (git-style)
	return hex.EncodeToString(hash[:]), nil
}

// compareSnapshots generates differences between two snapshots
func (s *tagService) compareSnapshots(snap1, snap2 *models.AgencySnapshot) []TagDifference {
	differences := []TagDifference{}

	// Convert snapshots to JSON for comparison
	json1, _ := json.Marshal(snap1)
	json2, _ := json.Marshal(snap2)

	var map1, map2 map[string]interface{}
	json.Unmarshal(json1, &map1)
	json.Unmarshal(json2, &map2)

	// Compare recursively
	s.compareObjects("", map1, map2, &differences)

	return differences
}

// compareObjects recursively compares two objects and generates diffs
func (s *tagService) compareObjects(path string, obj1, obj2 map[string]interface{}, diffs *[]TagDifference) {
	// Check for added/modified keys in obj2
	for key, val2 := range obj2 {
		newPath := s.buildPath(path, key)
		val1, exists := obj1[key]

		if !exists {
			// Key added in obj2
			*diffs = append(*diffs, TagDifference{
				Path:     newPath,
				Type:     "added",
				NewValue: val2,
			})
		} else if !reflect.DeepEqual(val1, val2) {
			// Key modified
			// Recursively compare if both are maps
			map1, isMap1 := val1.(map[string]interface{})
			map2, isMap2 := val2.(map[string]interface{})

			if isMap1 && isMap2 {
				s.compareObjects(newPath, map1, map2, diffs)
			} else {
				*diffs = append(*diffs, TagDifference{
					Path:     newPath,
					Type:     "modified",
					OldValue: val1,
					NewValue: val2,
				})
			}
		}
	}

	// Check for removed keys (in obj1 but not in obj2)
	for key, val1 := range obj1 {
		if _, exists := obj2[key]; !exists {
			newPath := s.buildPath(path, key)
			*diffs = append(*diffs, TagDifference{
				Path:     newPath,
				Type:     "removed",
				OldValue: val1,
			})
		}
	}
}

// buildPath builds a JSON path notation
func (s *tagService) buildPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

// Validation helpers

var (
	ErrTagNotFound      = errors.New("tag not found")
	ErrTagAlreadyExists = errors.New("tag already exists")
	ErrInvalidTagType   = errors.New("invalid tag type")
	ErrInvalidState     = errors.New("invalid agency state for operation")
)
