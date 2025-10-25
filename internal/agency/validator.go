package agency

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Validator defines the interface for agency validation
type Validator interface {
	ValidateAgency(agency *Agency) error
}

// validator implements the Validator interface
type validator struct{}

// NewValidator creates a new validator
func NewValidator() Validator {
	return &validator{}
}

// ValidateAgency validates an agency's fields
func (v *validator) ValidateAgency(agency *Agency) error {
	if agency == nil {
		return fmt.Errorf("agency cannot be nil")
	}

	// Validate required fields
	if agency.ID == "" {
		return fmt.Errorf("agency ID is required")
	}
	if agency.Name == "" {
		return fmt.Errorf("agency name is required")
	}
	if agency.DisplayName == "" {
		return fmt.Errorf("agency display name is required")
	}
	if agency.Category == "" {
		return fmt.Errorf("agency category is required")
	}

	// Validate ID format (should be "agency_" + 32 hex characters without hyphens)
	if !strings.HasPrefix(agency.ID, "agency_") {
		return fmt.Errorf("agency ID must start with 'agency_' prefix")
	}
	
	// Extract the UUID part after the prefix
	uuidPart := strings.TrimPrefix(agency.ID, "agency_")
	
	// Validate the UUID part (32 hex characters without hyphens)
	if len(uuidPart) == 32 {
		// Without hyphens - validate it's all hex
		for _, c := range uuidPart {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return fmt.Errorf("agency ID must be 'agency_' followed by 32 hex characters without hyphens")
			}
		}
	} else if len(uuidPart) == 36 {
		// Try parsing with hyphens (for backwards compatibility)
		if _, err := uuid.Parse(uuidPart); err != nil {
			return fmt.Errorf("agency ID UUID part must be valid UUID format: %w", err)
		}
	} else {
		return fmt.Errorf("agency ID must be 'agency_' followed by a valid UUID (32 or 36 characters)")
	}

	// Validate category
	if !isValidCategory(agency.Category) {
		return fmt.Errorf("invalid category: %s", agency.Category)
	}

	// Validate status
	if agency.Status != "" {
		if !isValidStatus(agency.Status) {
			return fmt.Errorf("invalid agency status: %s", agency.Status)
		}
	}

	return nil
}

// isValidStatus checks if the status is valid
func isValidStatus(status AgencyStatus) bool {
	switch status {
	case AgencyStatusActive, AgencyStatusInactive, AgencyStatusPaused, AgencyStatusArchived:
		return true
	default:
		return false
	}
}

// isValidCategory checks if the category is valid
func isValidCategory(category string) bool {
	validCategories := map[string]bool{
		"infrastructure": true,
		"agriculture":    true,
		"logistics":      true,
		"transportation": true,
		"healthcare":     true,
		"education":      true,
		"finance":        true,
		"retail":         true,
		"energy":         true,
		"other":          true,
	}
	return validCategories[strings.ToLower(category)]
}

// GenerateAgencyID generates a new UUID for an agency with "agency_" prefix and without hyphens
func GenerateAgencyID() string {
	return "agency_" + strings.ReplaceAll(uuid.New().String(), "-", "")
}
