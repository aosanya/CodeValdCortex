package agency

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Validator defines the interface for agency validation
type Validator interface {
	ValidateAgency(agency *Agency) error
	ValidateConfiguration(configPath string) error
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

	// Validate ID format (should start with UC-)
	if !strings.HasPrefix(agency.ID, "UC-") {
		return fmt.Errorf("agency ID must start with 'UC-'")
	}

	// Validate status
	if agency.Status != "" {
		if !isValidStatus(agency.Status) {
			return fmt.Errorf("invalid agency status: %s", agency.Status)
		}
	}

	// Validate config path if provided
	if agency.ConfigPath != "" {
		if err := v.ValidateConfiguration(agency.ConfigPath); err != nil {
			return fmt.Errorf("invalid config path: %w", err)
		}
	}

	return nil
}

// ValidateConfiguration validates a configuration path
func (v *validator) ValidateConfiguration(configPath string) error {
	if configPath == "" {
		return fmt.Errorf("config path cannot be empty")
	}

	// Check if path is absolute or starts with /usecases/
	if !filepath.IsAbs(configPath) && !strings.HasPrefix(configPath, "/usecases/") {
		return fmt.Errorf("config path must be absolute or start with /usecases/")
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
