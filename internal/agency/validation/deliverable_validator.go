package validation

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

var (
	// ErrInvalidDeliverableName indicates the deliverable name is invalid
	ErrInvalidDeliverableName = errors.New("deliverable name is invalid")

	// ErrExcessiveNesting indicates the tree depth exceeds the maximum allowed
	ErrExcessiveNesting = errors.New("deliverable tree nesting exceeds maximum depth")

	// ErrTooManyChildren indicates a node has too many children
	ErrTooManyChildren = errors.New("deliverable node has too many children")

	// ErrInvalidFileExtension indicates the file extension is not supported
	ErrInvalidFileExtension = errors.New("file extension is not supported")

	// ErrMissingFileExtension indicates a file node is missing its extension
	ErrMissingFileExtension = errors.New("file deliverable is missing extension")

	// ErrFolderWithExtension indicates a folder incorrectly has a file extension
	ErrFolderWithExtension = errors.New("folder deliverable cannot have file extension")

	// ErrFileWithChildren indicates a file node has children
	ErrFileWithChildren = errors.New("file deliverable cannot have children")

	// ErrPromptTooLong indicates the prompt instructions exceed the maximum length
	ErrPromptTooLong = errors.New("prompt instructions exceed maximum length")

	// ErrDuplicateID indicates duplicate IDs exist in the tree
	ErrDuplicateID = errors.New("duplicate deliverable ID found")

	// ErrEmptyName indicates the deliverable name is empty
	ErrEmptyName = errors.New("deliverable name cannot be empty")

	// ErrEmptyID indicates the deliverable ID is missing
	ErrEmptyID = errors.New("deliverable ID cannot be empty")

	// ErrDuplicateName indicates duplicate names exist in the same parent
	ErrDuplicateName = errors.New("duplicate deliverable name in same parent")
)

// Valid file name pattern (alphanumeric, hyphens, underscores, spaces, periods)
var validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9\s._-]+$`)

// DeliverableValidator provides validation for deliverable tree structures
type DeliverableValidator struct {
	allowedExtensions map[string]bool
}

// NewDeliverableValidator creates a new validator with default settings
func NewDeliverableValidator() *DeliverableValidator {
	return &DeliverableValidator{
		allowedExtensions: map[string]bool{
			".md": true, // Initially only markdown files
		},
	}
}

// ValidateTree validates an entire deliverable tree
func (v *DeliverableValidator) ValidateTree(nodes []models.DeliverableNode) error {
	if len(nodes) == 0 {
		return nil // Empty tree is valid
	}

	// Check for duplicate IDs across the entire tree
	ids := make(map[string]bool)
	for _, node := range nodes {
		if err := v.checkDuplicateIDs(&node, ids); err != nil {
			return err
		}
	}

	// Check for duplicate names at root level
	if err := v.checkDuplicateNames(nodes); err != nil {
		return err
	}

	// Validate each root node
	for i := range nodes {
		if err := v.ValidateNode(&nodes[i], 1); err != nil {
			return fmt.Errorf("validation failed for node '%s': %w", nodes[i].Name, err)
		}
	}

	return nil
}

// ValidateNode validates a single deliverable node and its children
func (v *DeliverableValidator) ValidateNode(node *models.DeliverableNode, depth int) error {
	// Check depth limit
	if depth > models.MaxNestingDepth {
		return fmt.Errorf("%w: depth %d exceeds maximum %d", ErrExcessiveNesting, depth, models.MaxNestingDepth)
	}

	// Validate name
	if err := v.validateName(node.Name); err != nil {
		return err
	}

	// Validate type-specific constraints
	if err := v.validateType(node); err != nil {
		return err
	}

	// Validate prompt instructions length
	if len(node.PromptInstructions) > models.MaxPromptLength {
		return fmt.Errorf("%w: %d characters exceeds maximum %d", ErrPromptTooLong, len(node.PromptInstructions), models.MaxPromptLength)
	}

	// Validate children count
	if len(node.Children) > models.MaxChildrenPerNode {
		return fmt.Errorf("%w: %d children exceeds maximum %d", ErrTooManyChildren, len(node.Children), models.MaxChildrenPerNode)
	}

	// Check for duplicate names among children
	if err := v.checkDuplicateNames(node.Children); err != nil {
		return err
	}

	// Recursively validate children
	for i := range node.Children {
		if err := v.ValidateNode(&node.Children[i], depth+1); err != nil {
			return fmt.Errorf("child '%s': %w", node.Children[i].Name, err)
		}
	}

	return nil
}

// validateName checks if the deliverable name is valid
func (v *DeliverableValidator) validateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}

	if len(name) > models.MaxNameLength {
		return fmt.Errorf("%w: %d characters exceeds maximum %d", ErrInvalidDeliverableName, len(name), models.MaxNameLength)
	}

	// Check for invalid characters
	if !validNamePattern.MatchString(name) {
		return fmt.Errorf("%w: name contains invalid characters", ErrInvalidDeliverableName)
	}

	// Prevent path traversal attempts
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("%w: name cannot contain path separators or parent directory references", ErrInvalidDeliverableName)
	}

	return nil
}

// validateType validates type-specific constraints
func (v *DeliverableValidator) validateType(node *models.DeliverableNode) error {
	switch node.Type {
	case models.DeliverableTypeFile:
		// Files must have an extension
		if node.FileExtension == "" {
			return ErrMissingFileExtension
		}

		// Validate extension is allowed
		if !v.allowedExtensions[node.FileExtension] {
			return fmt.Errorf("%w: '%s' is not supported (only .md currently allowed)", ErrInvalidFileExtension, node.FileExtension)
		}

		// Files cannot have children
		if len(node.Children) > 0 {
			return ErrFileWithChildren
		}

	case models.DeliverableTypeFolder:
		// Folders cannot have file extensions
		if node.FileExtension != "" {
			return ErrFolderWithExtension
		}

	default:
		return fmt.Errorf("invalid deliverable type: %s", node.Type)
	}

	return nil
}

// checkDuplicateIDs recursively checks for duplicate IDs in the tree
func (v *DeliverableValidator) checkDuplicateIDs(node *models.DeliverableNode, ids map[string]bool) error {
	// Check for empty ID first
	if node.ID == "" {
		return fmt.Errorf("%w: node '%s' (this should not happen - IDs should be generated before validation)", ErrEmptyID, node.Name)
	}

	if ids[node.ID] {
		return fmt.Errorf("%w: ID '%s'", ErrDuplicateID, node.ID)
	}
	ids[node.ID] = true

	for i := range node.Children {
		if err := v.checkDuplicateIDs(&node.Children[i], ids); err != nil {
			return err
		}
	}

	return nil
}

// checkDuplicateNames checks for duplicate names within the same parent node
func (v *DeliverableValidator) checkDuplicateNames(nodes []models.DeliverableNode) error {
	names := make(map[string]bool)

	for _, node := range nodes {
		// Normalize name for comparison (case-insensitive)
		normalizedName := strings.ToLower(node.Name)

		if names[normalizedName] {
			return fmt.Errorf("%w: '%s' appears multiple times in the same parent", ErrDuplicateName, node.Name)
		}
		names[normalizedName] = true
	}

	return nil
}

// SanitizePath cleans and validates a deliverable path
func (v *DeliverableValidator) SanitizePath(path string) (string, error) {
	// Clean the path
	cleaned := filepath.Clean(path)

	// Prevent path traversal
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "../") {
		return "", fmt.Errorf("invalid path: contains parent directory references")
	}

	return cleaned, nil
}

// AddAllowedExtension adds a new allowed file extension
func (v *DeliverableValidator) AddAllowedExtension(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	v.allowedExtensions[ext] = true
}

// RemoveAllowedExtension removes an allowed file extension
func (v *DeliverableValidator) RemoveAllowedExtension(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	delete(v.allowedExtensions, ext)
}

// IsExtensionAllowed checks if a file extension is allowed
func (v *DeliverableValidator) IsExtensionAllowed(ext string) bool {
	return v.allowedExtensions[ext]
}
