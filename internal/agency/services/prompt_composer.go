package services

import (
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// PromptComposer handles composition of prompt instructions for deliverables
// with hierarchical inheritance from parent folders
type PromptComposer struct{}

// NewPromptComposer creates a new PromptComposer instance
func NewPromptComposer() *PromptComposer {
	return &PromptComposer{}
}

// CompositionFormat defines the format for composing prompts
type CompositionFormat string

const (
	// FormatHierarchicalSections uses distinct sections for project/folder/task
	FormatHierarchicalSections CompositionFormat = "hierarchical_sections"
	
	// FormatConcatenated simply concatenates all prompts
	FormatConcatenated CompositionFormat = "concatenated"
)

// ComposePrompt generates the full prompt for a deliverable node
// by combining its own instructions with inherited parent context
func (pc *PromptComposer) ComposePrompt(node *models.DeliverableNode, parents []models.DeliverableNode, format CompositionFormat) string {
	switch format {
	case FormatHierarchicalSections:
		return pc.composeHierarchicalSections(node, parents)
	case FormatConcatenated:
		return pc.composeConcatenated(node, parents)
	default:
		return pc.composeHierarchicalSections(node, parents)
	}
}

// composeHierarchicalSections creates a structured prompt with clear sections
func (pc *PromptComposer) composeHierarchicalSections(node *models.DeliverableNode, parents []models.DeliverableNode) string {
	var builder strings.Builder

	// Section 1: Project-level guidelines (root level instructions)
	if len(parents) > 0 && parents[0].PromptInstructions != "" {
		builder.WriteString("# PROJECT GUIDELINES\n\n")
		builder.WriteString(parents[0].PromptInstructions)
		builder.WriteString("\n\n")
	}

	// Section 2: Folder-level guidelines (intermediate parent instructions)
	if len(parents) > 1 {
		builder.WriteString("# FOLDER GUIDELINES\n\n")
		for i := 1; i < len(parents); i++ {
			if parents[i].PromptInstructions != "" {
				builder.WriteString(fmt.Sprintf("## %s\n\n", parents[i].Name))
				builder.WriteString(parents[i].PromptInstructions)
				builder.WriteString("\n\n")
			}
		}
	}

	// Section 3: Specific task instructions (node's own instructions)
	if node.PromptInstructions != "" {
		builder.WriteString("# YOUR TASK\n\n")
		builder.WriteString(fmt.Sprintf("**File**: %s\n\n", node.Path))
		builder.WriteString(node.PromptInstructions)
		builder.WriteString("\n")
	}

	result := builder.String()
	if result == "" {
		return "No specific instructions provided."
	}

	return result
}

// composeConcatenated simply concatenates all instructions from parents to node
func (pc *PromptComposer) composeConcatenated(node *models.DeliverableNode, parents []models.DeliverableNode) string {
	var builder strings.Builder

	// Add parent instructions in order
	for _, parent := range parents {
		if parent.PromptInstructions != "" {
			builder.WriteString(parent.PromptInstructions)
			builder.WriteString("\n\n")
		}
	}

	// Add node's own instructions
	if node.PromptInstructions != "" {
		builder.WriteString(node.PromptInstructions)
		builder.WriteString("\n")
	}

	result := builder.String()
	if result == "" {
		return "No specific instructions provided."
	}

	return strings.TrimSpace(result)
}

// BuildParentChain traverses up the tree to build the parent chain for a node
func (pc *PromptComposer) BuildParentChain(nodeID string, rootNodes []models.DeliverableNode) []models.DeliverableNode {
	var chain []models.DeliverableNode
	
	// Find the node and build its parent chain
	for i := range rootNodes {
		if found, parents := pc.findNodeWithParents(&rootNodes[i], nodeID, []models.DeliverableNode{}); found != nil {
			// Build the chain: root → intermediate parents → current node
			return append(parents, *found)
		}
	}
	
	return chain
}

// findNodeWithParents recursively searches for a node and builds its parent chain
func (pc *PromptComposer) findNodeWithParents(current *models.DeliverableNode, targetID string, parents []models.DeliverableNode) (*models.DeliverableNode, []models.DeliverableNode) {
	// If this is the target node, return it with its parents
	if current.ID == targetID {
		return current, parents
	}
	
	// If this is a folder, search in children
	if current.Type == models.DeliverableTypeFolder {
		// Add current to parent chain
		newParents := append(parents, *current)
		
		for i := range current.Children {
			if found, chain := pc.findNodeWithParents(&current.Children[i], targetID, newParents); found != nil {
				return found, chain
			}
		}
	}
	
	return nil, nil
}

// ComposePromptForNode is a convenience method that handles the parent chain building
func (pc *PromptComposer) ComposePromptForNode(nodeID string, rootNodes []models.DeliverableNode, format CompositionFormat) string {
	// Build parent chain
	chain := pc.BuildParentChain(nodeID, rootNodes)
	
	if len(chain) == 0 {
		return "Node not found in tree."
	}
	
	// Last element in chain is the target node
	node := chain[len(chain)-1]
	parents := chain[:len(chain)-1]
	
	return pc.ComposePrompt(&node, parents, format)
}

// GeneratePromptPreview creates a summary preview of how prompts will be inherited
func (pc *PromptComposer) GeneratePromptPreview(rootNodes []models.DeliverableNode) map[string]PromptPreview {
	previews := make(map[string]PromptPreview)
	
	for i := range rootNodes {
		pc.generatePreviewRecursive(&rootNodes[i], []models.DeliverableNode{}, previews)
	}
	
	return previews
}

// PromptPreview contains information about prompt composition for a node
type PromptPreview struct {
	NodeID             string   `json:"node_id"`
	NodeName           string   `json:"node_name"`
	NodePath           string   `json:"node_path"`
	HasOwnInstructions bool     `json:"has_own_instructions"`
	InheritedFrom      []string `json:"inherited_from"`      // Names of parent nodes with instructions
	TotalPromptLength  int      `json:"total_prompt_length"` // Total length of composed prompt
}

// generatePreviewRecursive builds prompt previews for all nodes in the tree
func (pc *PromptComposer) generatePreviewRecursive(node *models.DeliverableNode, parents []models.DeliverableNode, previews map[string]PromptPreview) {
	// Build inherited from list
	inheritedFrom := []string{}
	for _, parent := range parents {
		if parent.PromptInstructions != "" {
			inheritedFrom = append(inheritedFrom, parent.Name)
		}
	}
	
	// Compose the full prompt to get its length
	fullPrompt := pc.ComposePrompt(node, parents, FormatHierarchicalSections)
	
	// Create preview
	previews[node.ID] = PromptPreview{
		NodeID:             node.ID,
		NodeName:           node.Name,
		NodePath:           node.Path,
		HasOwnInstructions: node.PromptInstructions != "",
		InheritedFrom:      inheritedFrom,
		TotalPromptLength:  len(fullPrompt),
	}
	
	// Recurse into children
	if node.Type == models.DeliverableTypeFolder {
		newParents := append(parents, *node)
		for i := range node.Children {
			pc.generatePreviewRecursive(&node.Children[i], newParents, previews)
		}
	}
}
