package templates

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// DefaultValidator implements the Validator interface
type DefaultValidator struct {
}

// NewDefaultValidator creates a new default template validator
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{}
}

// NewDefaultValidatorWithTypeService creates a validator (kept for compatibility, no longer uses type service)
// Deprecated: Use NewDefaultValidator instead
func NewDefaultValidatorWithTypeService() *DefaultValidator {
	return &DefaultValidator{}
}

// SetRoleService is deprecated - no longer uses role service
// Deprecated: Kept for compatibility only
func (v *DefaultValidator) SetRoleService(service interface{}) {
	// No-op - registry removed
}

// ValidateTemplate validates a template
func (v *DefaultValidator) ValidateTemplate(tmpl *Template) error {
	var errors []string

	// Validate basic fields
	if tmpl.Name == "" {
		errors = append(errors, "template name cannot be empty")
	}

	if tmpl.AgentType == "" {
		errors = append(errors, "agent type cannot be empty")
	}

	if tmpl.Content == "" {
		errors = append(errors, "template content cannot be empty")
	}

	// Validate agent type
	if !v.isValidAgentType(tmpl.AgentType) {
		errors = append(errors, fmt.Sprintf("invalid agent type: %s", tmpl.AgentType))
	}

	// Validate template syntax
	if err := v.validateTemplateSyntax(tmpl.Content); err != nil {
		errors = append(errors, fmt.Sprintf("template syntax error: %v", err))
	}

	// Validate variables
	if err := v.validateTemplateVariables(tmpl.Variables); err != nil {
		errors = append(errors, fmt.Sprintf("template variables error: %v", err))
	}

	// Validate that template content uses declared variables
	if err := v.validateVariableUsage(tmpl.Content, tmpl.Variables); err != nil {
		errors = append(errors, fmt.Sprintf("variable usage error: %v", err))
	}

	// Validate labels
	if err := v.validateLabels(tmpl.Labels); err != nil {
		errors = append(errors, fmt.Sprintf("labels error: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("template validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ValidateVariables validates template variables against their definitions
func (v *DefaultValidator) ValidateVariables(variables map[string]interface{}, templateVars []TemplateVariable) error {
	var errors []string

	// Check required variables
	for _, tv := range templateVars {
		value, exists := variables[tv.Name]

		if tv.Required && !exists {
			errors = append(errors, fmt.Sprintf("required variable '%s' is missing", tv.Name))
			continue
		}

		if !exists {
			continue // Optional variable not provided
		}

		// Validate variable type and constraints
		if err := v.validateVariableValue(tv, value); err != nil {
			errors = append(errors, fmt.Sprintf("variable '%s': %v", tv.Name, err))
		}
	}

	// Check for unexpected variables
	expectedVars := make(map[string]bool)
	for _, tv := range templateVars {
		expectedVars[tv.Name] = true
	}

	for name := range variables {
		if !expectedVars[name] {
			errors = append(errors, fmt.Sprintf("unexpected variable '%s'", name))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("variable validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateTemplateSyntax validates Go template syntax
func (v *DefaultValidator) validateTemplateSyntax(content string) error {
	_, err := template.New("test").Parse(content)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}

	// Try to validate it produces valid JSON structure
	// We'll use a dummy set of variables for this test
	testVars := make(map[string]interface{})

	// Extract variable references from template content
	varRefs := v.extractVariableReferences(content)
	for _, varRef := range varRefs {
		testVars[varRef] = "test_value"
	}

	tmpl, err := template.New("test").Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, testVars); err != nil {
		return fmt.Errorf("template execution failed with test variables: %w", err)
	}

	// Try to parse the result as JSON to ensure it's structurally valid
	result := buf.String()
	if strings.TrimSpace(result) != "" {
		var testObj interface{}
		if err := json.Unmarshal([]byte(result), &testObj); err != nil {
			return fmt.Errorf("template output is not valid JSON: %w", err)
		}
	}

	return nil
}

// validateTemplateVariables validates the variable definitions
func (v *DefaultValidator) validateTemplateVariables(variables []TemplateVariable) error {
	var errors []string
	names := make(map[string]bool)

	for i, tv := range variables {
		// Check for duplicate names
		if names[tv.Name] {
			errors = append(errors, fmt.Sprintf("duplicate variable name '%s'", tv.Name))
		}
		names[tv.Name] = true

		// Validate variable name format
		if !v.isValidVariableName(tv.Name) {
			errors = append(errors, fmt.Sprintf("invalid variable name '%s' at index %d", tv.Name, i))
		}

		// Validate variable type
		if !v.isValidVariableType(tv.Type) {
			errors = append(errors, fmt.Sprintf("invalid variable type '%s' for variable '%s'", tv.Type, tv.Name))
		}

		// Validate default value type matches declared type
		if tv.DefaultValue != nil {
			if err := v.validateValueType(tv.DefaultValue, tv.Type); err != nil {
				errors = append(errors, fmt.Sprintf("default value for variable '%s' doesn't match type '%s': %v", tv.Name, tv.Type, err))
			}
		}

		// Validate constraints
		if err := v.validateVariableConstraints(tv); err != nil {
			errors = append(errors, fmt.Sprintf("invalid constraints for variable '%s': %v", tv.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("variable definitions validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateVariableUsage validates that template uses declared variables correctly
func (v *DefaultValidator) validateVariableUsage(content string, variables []TemplateVariable) error {
	// Extract variable references from template
	usedVars := v.extractVariableReferences(content)

	// Create map of declared variables
	declaredVars := make(map[string]bool)
	for _, tv := range variables {
		declaredVars[tv.Name] = true
	}

	var errors []string

	// Check if all used variables are declared
	for _, usedVar := range usedVars {
		if !declaredVars[usedVar] {
			errors = append(errors, fmt.Sprintf("undeclared variable used in template: %s", usedVar))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("variable usage validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateVariableValue validates a variable value against its definition
func (v *DefaultValidator) validateVariableValue(tv TemplateVariable, value interface{}) error {
	// Validate type
	if err := v.validateValueType(value, tv.Type); err != nil {
		return err
	}

	// Validate constraints based on type
	switch tv.Type {
	case "string":
		if str, ok := value.(string); ok {
			if tv.Pattern != "" {
				if matched, err := regexp.MatchString(tv.Pattern, str); err != nil {
					return fmt.Errorf("pattern validation failed: %w", err)
				} else if !matched {
					return fmt.Errorf("value '%s' doesn't match pattern '%s'", str, tv.Pattern)
				}
			}
		}
	case "int", "float":
		if num, ok := v.convertToFloat64(value); ok {
			if tv.MinValue != nil && num < *tv.MinValue {
				return fmt.Errorf("value %v is less than minimum %v", num, *tv.MinValue)
			}
			if tv.MaxValue != nil && num > *tv.MaxValue {
				return fmt.Errorf("value %v is greater than maximum %v", num, *tv.MaxValue)
			}
		}
	}

	// Validate against valid values (enum)
	if len(tv.ValidValues) > 0 {
		found := false
		for _, validValue := range tv.ValidValues {
			if v.valuesEqual(value, validValue) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value %v is not in the list of valid values %v", value, tv.ValidValues)
		}
	}

	return nil
}

// validateLabels validates label format and content
func (v *DefaultValidator) validateLabels(labels map[string]string) error {
	labelRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-_\.]*[a-zA-Z0-9])?$`)

	for key, value := range labels {
		if len(key) == 0 || len(key) > 63 {
			return fmt.Errorf("label key '%s' must be 1-63 characters", key)
		}

		if !labelRegex.MatchString(key) {
			return fmt.Errorf("label key '%s' contains invalid characters", key)
		}

		if len(value) > 63 {
			return fmt.Errorf("label value '%s' must be at most 63 characters", value)
		}

		if value != "" && !labelRegex.MatchString(value) {
			return fmt.Errorf("label value '%s' contains invalid characters", value)
		}
	}

	return nil
}

// validateVariableConstraints validates variable constraints
func (v *DefaultValidator) validateVariableConstraints(tv TemplateVariable) error {
	if tv.MinValue != nil && tv.MaxValue != nil && *tv.MinValue > *tv.MaxValue {
		return fmt.Errorf("min_value cannot be greater than max_value")
	}

	if tv.Pattern != "" && tv.Type != "string" {
		return fmt.Errorf("pattern constraint can only be used with string type")
	}

	if (tv.MinValue != nil || tv.MaxValue != nil) && tv.Type != "int" && tv.Type != "float" {
		return fmt.Errorf("min_value/max_value constraints can only be used with numeric types")
	}

	return nil
}

// Helper methods

func (v *DefaultValidator) isValidAgentType(agentType string) bool {
	// Use hardcoded list of valid types
	validTypes := []string{"worker", "coordinator", "monitor", "proxy", "gateway"}
	for _, valid := range validTypes {
		if agentType == valid {
			return true
		}
	}
	return false
}

func (v *DefaultValidator) isValidVariableName(name string) bool {
	// Variable names should follow identifier rules
	regex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return regex.MatchString(name)
}

func (v *DefaultValidator) isValidVariableType(varType string) bool {
	validTypes := []string{"string", "int", "float", "bool", "array", "object"}
	for _, valid := range validTypes {
		if varType == valid {
			return true
		}
	}
	return false
}

func (v *DefaultValidator) validateValueType(value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "int":
		switch value.(type) {
		case int, int32, int64, float64:
			if f, ok := value.(float64); ok && f != float64(int(f)) {
				return fmt.Errorf("expected integer, got float %v", value)
			}
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case "float":
		switch value.(type) {
		case float32, float64, int, int32, int64:
			// All numeric types are acceptable for float
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	default:
		return fmt.Errorf("unknown type: %s", expectedType)
	}
	return nil
}

func (v *DefaultValidator) convertToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func (v *DefaultValidator) valuesEqual(a, b interface{}) bool {
	// Simple equality check - could be made more sophisticated
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func (v *DefaultValidator) extractVariableReferences(content string) []string {
	// Extract variable references from Go template syntax
	// This is a simple regex-based approach - could be more sophisticated
	regex := regexp.MustCompile(`\{\{\s*\.([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)
	matches := regex.FindAllStringSubmatch(content, -1)

	var variables []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if !seen[varName] {
				variables = append(variables, varName)
				seen[varName] = true
			}
		}
	}

	return variables
}
