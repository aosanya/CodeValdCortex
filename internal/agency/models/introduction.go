package models

import "time"

// SectionType defines the type of introduction section
type SectionType string

const (
	SectionTypeText   SectionType = "text"
	SectionTypeList   SectionType = "list"
	SectionTypeNested SectionType = "nested"
	SectionTypeTable  SectionType = "table"
)

// SectionValidation defines validation rules for a section
type SectionValidation struct {
	MinLength *int    `json:"min_length,omitempty"`
	MaxLength *int    `json:"max_length,omitempty"`
	Pattern   *string `json:"pattern,omitempty"`
	MinItems  *int    `json:"min_items,omitempty"`
	MaxItems  *int    `json:"max_items,omitempty"`
}

// SectionMetadata tracks section version history and authorship
type SectionMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	Version   string    `json:"version"`
}

// IntroductionSection represents a single section in the agency introduction
type IntroductionSection struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	Order      int                `json:"order"`
	Type       SectionType        `json:"type"`
	Required   bool               `json:"required"`
	Content    interface{}        `json:"content"` // Can be string, ListContent, NestedContent, or TableContent
	Validation *SectionValidation `json:"validation,omitempty"`
	Metadata   SectionMetadata    `json:"metadata"`
}

// AgencyIntroduction represents the complete introduction structure for an agency
type AgencyIntroduction struct {
	Key       string                `json:"_key" arangodb:"_key"`
	ID        string                `json:"_id,omitempty" arangodb:"_id,omitempty"`
	Rev       string                `json:"_rev,omitempty" arangodb:"_rev,omitempty"`
	AgencyID  string                `json:"agency_id"`
	Sections  []IntroductionSection `json:"sections"`
	Template  string                `json:"template"` // "genesis", "minimal", "custom"
	Version   string                `json:"version"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	CreatedBy string                `json:"created_by"`
}

// ListItem represents an item in a list-type section
type ListItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ListContent represents the content structure for list-type sections
type ListContent []ListItem

// NestedContent represents nested section structure (e.g., in/out of scope)
type NestedContent struct {
	InScope  []ListItem `json:"in_scope,omitempty"`
	OutScope []ListItem `json:"out_of_scope,omitempty"`
}

// TableRow represents a single row in a table
type TableRow struct {
	ID   string            `json:"id"`
	Data map[string]string `json:"data"`
}

// TableContent represents tabular data
type TableContent struct {
	Columns []string   `json:"columns"`
	Rows    []TableRow `json:"rows"`
}

// Validate performs validation on an IntroductionSection based on its type and validation rules
func (s *IntroductionSection) Validate() error {
	if s.Title == "" {
		return NewValidationError("section title is required")
	}

	if s.Validation != nil {
		switch s.Type {
		case SectionTypeText:
			return s.validateTextSection()
		case SectionTypeList:
			return s.validateListSection()
		case SectionTypeNested:
			return s.validateNestedSection()
		case SectionTypeTable:
			return s.validateTableSection()
		}
	}

	return nil
}

func (s *IntroductionSection) validateTextSection() error {
	content, ok := s.Content.(string)
	if !ok {
		return NewValidationError("text section content must be a string")
	}

	if s.Validation.MinLength != nil && len(content) < *s.Validation.MinLength {
		return NewValidationError("content length below minimum")
	}

	if s.Validation.MaxLength != nil && len(content) > *s.Validation.MaxLength {
		return NewValidationError("content length exceeds maximum")
	}

	return nil
}

func (s *IntroductionSection) validateListSection() error {
	// Content validation for list sections
	if s.Validation.MinItems != nil || s.Validation.MaxItems != nil {
		// This would need type assertion to []ListItem to check count
		// Implementation depends on how content is structured
	}
	return nil
}

func (s *IntroductionSection) validateNestedSection() error {
	// Content validation for nested sections
	return nil
}

func (s *IntroductionSection) validateTableSection() error {
	// Content validation for table sections
	return nil
}

// Validate performs validation on the entire AgencyIntroduction
func (ai *AgencyIntroduction) Validate() error {
	if ai.AgencyID == "" {
		return NewValidationError("agency_id is required")
	}

	if len(ai.Sections) == 0 {
		return NewValidationError("at least one section is required")
	}

	// Check for duplicate section IDs
	seenIDs := make(map[string]bool)
	seenOrders := make(map[int]bool)

	for _, section := range ai.Sections {
		if seenIDs[section.ID] {
			return NewValidationError("duplicate section ID: " + section.ID)
		}
		seenIDs[section.ID] = true

		if seenOrders[section.Order] {
			return NewValidationError("duplicate section order")
		}
		seenOrders[section.Order] = true

		if err := section.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// NewAgencyIntroduction creates a new AgencyIntroduction with defaults
func NewAgencyIntroduction(agencyID, createdBy string) *AgencyIntroduction {
	now := time.Now()
	return &AgencyIntroduction{
		Key:       "agency_" + agencyID + "_intro",
		AgencyID:  agencyID,
		Sections:  []IntroductionSection{},
		Template:  "custom",
		Version:   "1.0.0",
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: createdBy,
	}
}

// NewIntroductionSection creates a new IntroductionSection with metadata
func NewIntroductionSection(id, title string, order int, sectionType SectionType, required bool, createdBy string) *IntroductionSection {
	now := time.Now()
	return &IntroductionSection{
		ID:       id,
		Title:    title,
		Order:    order,
		Type:     sectionType,
		Required: required,
		Metadata: SectionMetadata{
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: createdBy,
			Version:   "1.0.0",
		},
	}
}

// IntroValidationError represents a validation error for introductions
type IntroValidationError struct {
	Message string
}

func (e *IntroValidationError) Error() string {
	return e.Message
}

// NewIntroValidationError creates a new IntroValidationError
func NewIntroValidationError(message string) error {
	return &IntroValidationError{Message: message}
}
