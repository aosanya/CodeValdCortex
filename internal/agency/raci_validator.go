package agency

import (
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// RACIValidator provides validation logic for RACI matrices
type RACIValidator struct{}

// NewRACIValidator creates a new RACI validator
func NewRACIValidator() *RACIValidator {
	return &RACIValidator{}
}

// ValidateMatrix validates a complete RACI matrix
func (v *RACIValidator) ValidateMatrix(matrix *models.RACIMatrix) *models.RACIValidationResult {
	result := &models.RACIValidationResult{
		IsValid:  true,
		Errors:   []models.RACIValidationError{},
		Warnings: []models.RACIValidationWarning{},
		Summary: models.RACIValidationSummary{
			TotalActivities: len(matrix.Activities),
		},
	}

	for _, activity := range matrix.Activities {
		v.validateActivity(&activity, result)
	}

	// Calculate summary
	result.Summary.ActivitiesWithErrors = len(result.Errors)
	result.Summary.ActivitiesWithWarnings = len(result.Warnings)
	result.Summary.ValidActivities = result.Summary.TotalActivities - result.Summary.ActivitiesWithErrors

	// Set overall validity
	result.IsValid = len(result.Errors) == 0

	return result
}

// validateActivity validates a single activity
func (v *RACIValidator) validateActivity(activity *models.RACIActivity, result *models.RACIValidationResult) {
	// Count RACI roles
	accountableCount := 0
	responsibleCount := 0
	consultedCount := 0
	informedCount := 0

	for _, role := range activity.Assignments {
		switch role {
		case models.RACIAccountable:
			accountableCount++
		case models.RACIResponsible:
			responsibleCount++
		case models.RACIConsulted:
			consultedCount++
		case models.RACIInformed:
			informedCount++
		}
	}

	// Validation Rule 1: Must have exactly ONE Accountable
	if accountableCount == 0 {
		result.Errors = append(result.Errors, models.RACIValidationError{
			ActivityID: activity.ID,
			Activity:   activity.Name,
			ErrorType:  "missing_accountable",
			Message:    fmt.Sprintf("Activity '%s' has no Accountable (A) role assigned. Each activity must have exactly one Accountable.", activity.Name),
		})
	} else if accountableCount > 1 {
		result.Errors = append(result.Errors, models.RACIValidationError{
			ActivityID: activity.ID,
			Activity:   activity.Name,
			ErrorType:  "multiple_accountable",
			Message:    fmt.Sprintf("Activity '%s' has %d Accountable (A) roles assigned. Each activity must have exactly one Accountable.", activity.Name, accountableCount),
		})
	}

	// Validation Rule 2: Must have at least ONE Responsible
	if responsibleCount == 0 {
		result.Errors = append(result.Errors, models.RACIValidationError{
			ActivityID: activity.ID,
			Activity:   activity.Name,
			ErrorType:  "missing_responsible",
			Message:    fmt.Sprintf("Activity '%s' has no Responsible (R) role assigned. Each activity must have at least one Responsible.", activity.Name),
		})
	}

	// Warning: No Consulted roles (not an error, but worth noting)
	if consultedCount == 0 {
		result.Warnings = append(result.Warnings, models.RACIValidationWarning{
			ActivityID:  activity.ID,
			Activity:    activity.Name,
			WarningType: "no_consulted",
			Message:     fmt.Sprintf("Activity '%s' has no Consulted (C) roles. Consider if anyone should be consulted.", activity.Name),
		})
	}

	// Warning: No Informed roles (not an error, but worth noting)
	if informedCount == 0 {
		result.Warnings = append(result.Warnings, models.RACIValidationWarning{
			ActivityID:  activity.ID,
			Activity:    activity.Name,
			WarningType: "no_informed",
			Message:     fmt.Sprintf("Activity '%s' has no Informed (I) roles. Consider if anyone should be kept informed.", activity.Name),
		})
	}
}

// Validatemodels.RACIRole checks if a RACI role value is valid
func (v *RACIValidator) ValidateRACIRole(role models.RACIRole) bool {
	switch role {
	case models.RACIResponsible, models.RACIAccountable, models.RACIConsulted, models.RACIInformed:
		return true
	default:
		return false
	}
}

// ValidateRoles checks if all role names are valid (non-empty)
func (v *RACIValidator) ValidateRoles(roles []string) []string {
	errors := []string{}

	if len(roles) == 0 {
		errors = append(errors, "RACI matrix must have at least one role")
		return errors
	}

	roleMap := make(map[string]bool)
	for i, role := range roles {
		if role == "" {
			errors = append(errors, fmt.Sprintf("Role at position %d is empty", i))
		}

		if roleMap[role] {
			errors = append(errors, fmt.Sprintf("Duplicate role name: %s", role))
		}
		roleMap[role] = true
	}

	return errors
}

// ValidateActivities checks if activities are valid
func (v *RACIValidator) ValidateActivities(activities []models.RACIActivity, roles []string) []string {
	errors := []string{}

	if len(activities) == 0 {
		errors = append(errors, "RACI matrix must have at least one activity")
		return errors
	}

	roleSet := make(map[string]bool)
	for _, role := range roles {
		roleSet[role] = true
	}

	activityIDs := make(map[string]bool)
	for i, activity := range activities {
		// Check for duplicate IDs
		if activity.ID == "" {
			errors = append(errors, fmt.Sprintf("Activity at position %d has no ID", i))
		} else if activityIDs[activity.ID] {
			errors = append(errors, fmt.Sprintf("Duplicate activity ID: %s", activity.ID))
		}
		activityIDs[activity.ID] = true

		// Check activity name
		if activity.Name == "" {
			errors = append(errors, fmt.Sprintf("Activity '%s' has no name", activity.ID))
		}

		// Check assignments reference valid roles
		for roleName, raciRole := range activity.Assignments {
			if !roleSet[roleName] {
				errors = append(errors, fmt.Sprintf("Activity '%s' references unknown role: %s", activity.Name, roleName))
			}

			if !v.ValidateRACIRole(raciRole) {
				errors = append(errors, fmt.Sprintf("Activity '%s' has invalid RACI role '%s' for role '%s'", activity.Name, raciRole, roleName))
			}
		}
	}

	return errors
}
