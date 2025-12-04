package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// DeliverableNode represents a hierarchical deliverable structure (folder or file)
// with associated prompt instructions and nested children.
type DeliverableNode struct {
	ID                 string            `json:"id"`                  // UUID for UI tracking and references
	Name               string            `json:"name"`                // "stakeholders.md" or "introduction"
	Description        string            `json:"description"`         // Brief description of the deliverable
	Path               string            `json:"path"`                // Computed full path (e.g., "docs/requirements/stakeholders.md")
	Type               DeliverableType   `json:"type"`                // "folder" or "file"
	PromptInstructions string            `json:"prompt_instructions"` // AI instructions (what to achieve)
	Children           []DeliverableNode `json:"children,omitempty"`  // Nested deliverables (for folders)
	FileExtension      string            `json:"file_extension"`      // ".md" (initially only markdown)
	ParentID           string            `json:"parent_id,omitempty"` // Reference to parent node (for validation)
	Order              int               `json:"order"`               // Display order within parent
}

// DeliverableType defines whether a deliverable is a folder or file
type DeliverableType string

const (
	DeliverableTypeFolder DeliverableType = "folder"
	DeliverableTypeFile   DeliverableType = "file"
)

// Validation constants for deliverable structure
const (
	MaxNestingDepth    = 10   // Prevent excessive nesting
	MaxNameLength      = 255  // File/folder name limit
	MaxPromptLength    = 5000 // Prompt instruction limit
	MaxChildrenPerNode = 100  // Prevent overly complex trees
)

// ComputePath calculates the full path for the deliverable node
// based on its position in the tree hierarchy
func (d *DeliverableNode) ComputePath(parentPath string) string {
	if parentPath == "" {
		if d.Type == DeliverableTypeFile {
			return d.Name + d.FileExtension
		}
		return d.Name
	}

	if d.Type == DeliverableTypeFile {
		return parentPath + "/" + d.Name + d.FileExtension
	}
	return parentPath + "/" + d.Name
}

// IsValid performs basic validation on the deliverable node
func (d *DeliverableNode) IsValid() bool {
	// Name must not be empty
	if d.Name == "" {
		return false
	}

	// Name length check
	if len(d.Name) > MaxNameLength {
		return false
	}

	// Prompt instructions length check
	if len(d.PromptInstructions) > MaxPromptLength {
		return false
	}

	// Type must be valid
	if d.Type != DeliverableTypeFolder && d.Type != DeliverableTypeFile {
		return false
	}

	// Files must have extension (initially .md only)
	if d.Type == DeliverableTypeFile && d.FileExtension == "" {
		return false
	}

	// Folders cannot have file extensions
	if d.Type == DeliverableTypeFolder && d.FileExtension != "" {
		return false
	}

	// Files cannot have children
	if d.Type == DeliverableTypeFile && len(d.Children) > 0 {
		return false
	}

	// Children count check
	if len(d.Children) > MaxChildrenPerNode {
		return false
	}

	return true
}

// GetDepth calculates the depth of the deliverable tree from this node
func (d *DeliverableNode) GetDepth() int {
	if len(d.Children) == 0 {
		return 1
	}

	maxChildDepth := 0
	for _, child := range d.Children {
		childDepth := child.GetDepth()
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return 1 + maxChildDepth
}

// ComputeAllPaths recursively updates paths for the entire tree
func (d *DeliverableNode) ComputeAllPaths(parentPath string) {
	d.Path = d.ComputePath(parentPath)

	for i := range d.Children {
		d.Children[i].ComputeAllPaths(d.Path)
	}
}

// GetAllFiles returns a flat list of all file nodes in the tree
func (d *DeliverableNode) GetAllFiles() []DeliverableNode {
	files := []DeliverableNode{}

	if d.Type == DeliverableTypeFile {
		files = append(files, *d)
	}

	for _, child := range d.Children {
		files = append(files, child.GetAllFiles()...)
	}

	return files
}

// FindNodeByID recursively searches for a node by its ID
func (d *DeliverableNode) FindNodeByID(id string) *DeliverableNode {
	if d.ID == id {
		return d
	}

	for i := range d.Children {
		if found := d.Children[i].FindNodeByID(id); found != nil {
			return found
		}
	}

	return nil
}

// GenerateID creates a unique ID for a deliverable node
// Format: node-{timestamp}-{random}
func GenerateID() string {
	// Get current timestamp in milliseconds
	timestamp := time.Now().UnixMilli()

	// Generate 4 random bytes
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("node-%d", timestamp)
	}

	// Convert to hex string
	randomStr := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("node-%d-%s", timestamp, randomStr)
}

// EnsureID ensures this node has a valid ID, generating one if needed
func (d *DeliverableNode) EnsureID() {
	if d.ID == "" {
		d.ID = GenerateID()
	}
}

// EnsureAllIDs recursively ensures all nodes in the tree have valid IDs
func (d *DeliverableNode) EnsureAllIDs() {
	d.EnsureID()

	for i := range d.Children {
		d.Children[i].EnsureAllIDs()
	}
}

// EnsureAllIDsInTree ensures all nodes in a slice have valid IDs
func EnsureAllIDsInTree(nodes []DeliverableNode) {
	for i := range nodes {
		nodes[i].EnsureAllIDs()
	}
}
