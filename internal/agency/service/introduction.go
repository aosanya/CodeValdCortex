package service

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/repository"
)

// IntroductionService provides business logic for agency introductions
type IntroductionService struct {
	repo *repository.IntroductionRepository
}

// NewIntroductionService creates a new IntroductionService
func NewIntroductionService(repo *repository.IntroductionRepository) *IntroductionService {
	return &IntroductionService{
		repo: repo,
	}
}

// GetByAgencyID retrieves the introduction for a specific agency
func (s *IntroductionService) GetByAgencyID(ctx context.Context, agencyID string) (*models.AgencyIntroduction, error) {
	if agencyID == "" {
		return nil, fmt.Errorf("agency_id is required")
	}

	return s.repo.GetByAgencyID(ctx, agencyID)
}

// Create creates a new agency introduction
func (s *IntroductionService) Create(ctx context.Context, intro *models.AgencyIntroduction) error {
	if intro == nil {
		return fmt.Errorf("introduction cannot be nil")
	}

	// Business logic: Ensure sections are ordered sequentially
	if err := s.validateSectionOrdering(intro.Sections); err != nil {
		return fmt.Errorf("section ordering invalid: %w", err)
	}

	// Business logic: Check for at least one required section
	if !s.hasRequiredSection(intro.Sections) {
		return fmt.Errorf("introduction must have at least one required section")
	}

	return s.repo.Create(ctx, intro)
}

// Update updates an existing agency introduction
func (s *IntroductionService) Update(ctx context.Context, intro *models.AgencyIntroduction) error {
	if intro == nil {
		return fmt.Errorf("introduction cannot be nil")
	}

	// Verify the introduction exists
	existing, err := s.repo.GetByAgencyID(ctx, intro.AgencyID)
	if err != nil {
		return fmt.Errorf("introduction not found: %w", err)
	}

	// Business logic: Cannot change from a template to custom or vice versa without explicit migration
	if existing.Template != intro.Template && existing.Template != "custom" {
		return fmt.Errorf("cannot change template type from %s to %s", existing.Template, intro.Template)
	}

	// Validate section ordering
	if err := s.validateSectionOrdering(intro.Sections); err != nil {
		return fmt.Errorf("section ordering invalid: %w", err)
	}

	// Check for at least one required section
	if !s.hasRequiredSection(intro.Sections) {
		return fmt.Errorf("introduction must have at least one required section")
	}

	return s.repo.Update(ctx, intro)
}

// UpdateSection updates a specific section within an introduction
func (s *IntroductionService) UpdateSection(ctx context.Context, agencyID, sectionID string, section *models.IntroductionSection) error {
	if agencyID == "" {
		return fmt.Errorf("agency_id is required")
	}
	if sectionID == "" {
		return fmt.Errorf("section_id is required")
	}
	if section == nil {
		return fmt.Errorf("section cannot be nil")
	}

	// Verify the introduction exists
	intro, err := s.repo.GetByAgencyID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("introduction not found: %w", err)
	}

	// Business logic: Verify the section exists
	sectionExists := false
	for _, s := range intro.Sections {
		if s.ID == sectionID {
			sectionExists = true
			break
		}
	}
	if !sectionExists {
		return fmt.Errorf("section with id %s not found", sectionID)
	}

	// Ensure section ID matches
	section.ID = sectionID

	return s.repo.UpdateSection(ctx, agencyID, sectionID, section)
}

// AddSection adds a new section to an introduction
func (s *IntroductionService) AddSection(ctx context.Context, agencyID string, section *models.IntroductionSection) error {
	if agencyID == "" {
		return fmt.Errorf("agency_id is required")
	}
	if section == nil {
		return fmt.Errorf("section cannot be nil")
	}

	// Get current introduction to determine next order
	intro, err := s.repo.GetByAgencyID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("introduction not found: %w", err)
	}

	// Business logic: Ensure unique section ID
	for _, s := range intro.Sections {
		if s.ID == section.ID {
			return fmt.Errorf("section with id %s already exists", section.ID)
		}
	}

	// Auto-assign order if not provided
	if section.Order == 0 {
		section.Order = len(intro.Sections) + 1
	}

	return s.repo.AddSection(ctx, agencyID, section)
}

// DeleteSection removes a section from an introduction
func (s *IntroductionService) DeleteSection(ctx context.Context, agencyID, sectionID string) error {
	if agencyID == "" {
		return fmt.Errorf("agency_id is required")
	}
	if sectionID == "" {
		return fmt.Errorf("section_id is required")
	}

	// Get current introduction
	intro, err := s.repo.GetByAgencyID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("introduction not found: %w", err)
	}

	// Business logic: Cannot delete a required section if it's the last required one
	var targetSection *models.IntroductionSection
	requiredSectionCount := 0
	for i := range intro.Sections {
		if intro.Sections[i].ID == sectionID {
			targetSection = &intro.Sections[i]
		}
		if intro.Sections[i].Required {
			requiredSectionCount++
		}
	}

	if targetSection == nil {
		return fmt.Errorf("section with id %s not found", sectionID)
	}

	if targetSection.Required && requiredSectionCount == 1 {
		return fmt.Errorf("cannot delete the last required section")
	}

	return s.repo.DeleteSection(ctx, agencyID, sectionID)
}

// ReorderSections updates the order of sections
func (s *IntroductionService) ReorderSections(ctx context.Context, agencyID string, sectionIDs []string) error {
	if agencyID == "" {
		return fmt.Errorf("agency_id is required")
	}
	if len(sectionIDs) == 0 {
		return fmt.Errorf("section_ids cannot be empty")
	}

	// Get current introduction to validate section IDs
	intro, err := s.repo.GetByAgencyID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("introduction not found: %w", err)
	}

	// Business logic: Ensure all section IDs are provided (no missing sections)
	if len(sectionIDs) != len(intro.Sections) {
		return fmt.Errorf("must provide all section IDs for reordering, expected %d, got %d", len(intro.Sections), len(sectionIDs))
	}

	// Validate no duplicate IDs
	seenIDs := make(map[string]bool)
	for _, id := range sectionIDs {
		if seenIDs[id] {
			return fmt.Errorf("duplicate section ID in reorder: %s", id)
		}
		seenIDs[id] = true
	}

	return s.repo.ReorderSections(ctx, agencyID, sectionIDs)
}

// Delete removes an introduction (used when deleting an agency)
func (s *IntroductionService) Delete(ctx context.Context, agencyID string) error {
	if agencyID == "" {
		return fmt.Errorf("agency_id is required")
	}

	return s.repo.Delete(ctx, agencyID)
}

// GetByTemplate retrieves all introductions using a specific template
func (s *IntroductionService) GetByTemplate(ctx context.Context, template string) ([]*models.AgencyIntroduction, error) {
	if template == "" {
		return nil, fmt.Errorf("template is required")
	}

	validTemplates := map[string]bool{
		"genesis": true,
		"minimal": true,
		"custom":  true,
	}

	if !validTemplates[template] {
		return nil, fmt.Errorf("invalid template: %s", template)
	}

	return s.repo.GetByTemplate(ctx, template)
}

// validateSectionOrdering ensures sections are ordered sequentially starting from 1
func (s *IntroductionService) validateSectionOrdering(sections []models.IntroductionSection) error {
	if len(sections) == 0 {
		return nil
	}

	// Check that orders are sequential and start from 1
	expectedOrder := 1
	for _, section := range sections {
		if section.Order != expectedOrder {
			return fmt.Errorf("sections must be ordered sequentially starting from 1, expected order %d, got %d", expectedOrder, section.Order)
		}
		expectedOrder++
	}

	return nil
}

// hasRequiredSection checks if there's at least one required section
func (s *IntroductionService) hasRequiredSection(sections []models.IntroductionSection) bool {
	for _, section := range sections {
		if section.Required {
			return true
		}
	}
	return false
}

// ValidateSection performs comprehensive validation on a section
func (s *IntroductionService) ValidateSection(section *models.IntroductionSection) error {
	if section == nil {
		return fmt.Errorf("section cannot be nil")
	}

	// Validate section ID format
	if section.ID == "" {
		return fmt.Errorf("section ID is required")
	}

	// Validate title
	if section.Title == "" {
		return fmt.Errorf("section title is required")
	}

	// Validate order is positive
	if section.Order < 1 {
		return fmt.Errorf("section order must be >= 1")
	}

	// Validate section type
	validTypes := map[models.SectionType]bool{
		models.SectionTypeText:   true,
		models.SectionTypeList:   true,
		models.SectionTypeNested: true,
		models.SectionTypeTable:  true,
	}

	if !validTypes[section.Type] {
		return fmt.Errorf("invalid section type: %s", section.Type)
	}

	// Use the section's own validation method
	return section.Validate()
}

// GetSectionByID retrieves a specific section from an introduction
func (s *IntroductionService) GetSectionByID(ctx context.Context, agencyID, sectionID string) (*models.IntroductionSection, error) {
	intro, err := s.repo.GetByAgencyID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("introduction not found: %w", err)
	}

	for _, section := range intro.Sections {
		if section.ID == sectionID {
			return &section, nil
		}
	}

	return nil, fmt.Errorf("section with id %s not found", sectionID)
}

// CountSections returns the number of sections in an introduction
func (s *IntroductionService) CountSections(ctx context.Context, agencyID string) (int, error) {
	intro, err := s.repo.GetByAgencyID(ctx, agencyID)
	if err != nil {
		return 0, fmt.Errorf("introduction not found: %w", err)
	}

	return len(intro.Sections), nil
}

// GetRequiredSections returns all required sections from an introduction
func (s *IntroductionService) GetRequiredSections(ctx context.Context, agencyID string) ([]models.IntroductionSection, error) {
	intro, err := s.repo.GetByAgencyID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("introduction not found: %w", err)
	}

	var requiredSections []models.IntroductionSection
	for _, section := range intro.Sections {
		if section.Required {
			requiredSections = append(requiredSections, section)
		}
	}

	return requiredSections, nil
}
