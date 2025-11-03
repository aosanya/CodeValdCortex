package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/sirupsen/logrus"
)

// RoleCreator handles AI-powered role generation from work items
type RoleCreator struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewRoleCreator creates a new role creator service
func NewRoleCreator(llmClient LLMClient, logger *logrus.Logger) *RoleCreator {
	return &RoleCreator{
		llmClient: llmClient,
		logger:    logger,
	}
}

// GenerateRolesRequest contains the context for generating roles
type GenerateRolesRequest struct {
	AgencyID      string             `json:"agency_id"`
	AgencyContext *agency.Agency     `json:"agency_context"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
	ExistingRoles []*registry.Role   `json:"existing_roles"`
}

// GenerateRolesResponse contains the AI-generated roles
type GenerateRolesResponse struct {
	Roles       []GeneratedRole `json:"roles"`
	Explanation string          `json:"explanation"`
}

// GeneratedRole represents a single AI-generated role
type GeneratedRole struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Tags           []string `json:"tags"`
	AutonomyLevel  string   `json:"autonomy_level"`
	Capabilities   []string `json:"capabilities"`
	RequiredSkills []string `json:"required_skills"`
	TokenBudget    int64    `json:"token_budget"`
	Icon           string   `json:"icon"`
	Color          string   `json:"color"`
	SuggestedCode  string   `json:"suggested_code"`
}

// GenerateRoles uses AI to generate roles from work items
func (r *RoleCreator) GenerateRoles(ctx context.Context, req *GenerateRolesRequest) (*GenerateRolesResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI role generation from work items")

	// Build the prompt for role generation
	prompt := r.buildRoleGenerationPrompt(req)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: r.getRoleGenerationSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for role generation")
		return nil, fmt.Errorf("AI role generation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var aiResponse GenerateRolesResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		r.logger.WithError(err).Error("Failed to parse AI role generation response")
		r.logger.WithField("response", cleanedContent).Debug("Raw AI response for debugging")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	r.logger.WithField("role_count", len(aiResponse.Roles)).Info("AI role generation completed")
	return &aiResponse, nil
}

func (r *RoleCreator) getRoleGenerationSystemPrompt() string {
	return `You are an expert AI agent system architect helping to design multi-agent systems.

Your task is to analyze work items and generate appropriate agent role definitions.

For each role, provide:
- name: Clear, descriptive role name
- description: Detailed description of responsibilities
- tags: Relevant categorization tags (e.g., "development", "coordination", "analysis")
- autonomy_level: One of "L0" (fully automated), "L1" (human approval), "L2" (human in loop), "L3" (human monitored), "L4" (human initiated)
- capabilities: List of specific capabilities/actions the role can perform
- required_skills: Technical skills needed (e.g., "Python", "Data Analysis", "API Integration")
- token_budget: Estimated token budget for AI operations (e.g., 10000 for simple tasks, 50000 for complex)
- icon: Single emoji that represents the role
- color: Hex color code for UI display
- suggested_code: Short code identifier (e.g., "DEV", "COORD", "QA")

Generate 3-7 roles that cover the major functional areas needed based on the work items.
Avoid duplication with existing roles.

Response must be valid JSON matching this structure:
{
  "roles": [
    {
      "name": "string",
      "description": "string",
      "tags": ["string"],
      "autonomy_level": "L0|L1|L2|L3|L4",
      "capabilities": ["string"],
      "required_skills": ["string"],
      "token_budget": number,
      "icon": "emoji",
      "color": "#hexcode",
      "suggested_code": "string"
    }
  ],
  "explanation": "Brief explanation of the generated roles"
}`
}

func (r *RoleCreator) buildRoleGenerationPrompt(req *GenerateRolesRequest) string {
	var prompt strings.Builder

	// Agency context
	prompt.WriteString(fmt.Sprintf("## Agency: %s\n\n", req.AgencyContext.Name))
	if req.AgencyContext.Description != "" {
		prompt.WriteString(fmt.Sprintf("**Description:** %s\n\n", req.AgencyContext.Description))
	}

	// Work items context
	prompt.WriteString(fmt.Sprintf("## Work Items (%d total)\n\n", len(req.WorkItems)))

	// Group work items by tags to identify functional areas
	tagCounts := make(map[string]int)
	for _, wi := range req.WorkItems {
		for _, tag := range wi.Tags {
			tagCounts[tag]++
		}
	}

	// Show work item summary
	prompt.WriteString("**Work Item Categories:**\n")
	for tag, count := range tagCounts {
		prompt.WriteString(fmt.Sprintf("- %s: %d items\n", tag, count))
	}
	prompt.WriteString("\n")

	// Show first 10 work items as examples
	prompt.WriteString("**Sample Work Items:**\n")
	maxDisplay := 10
	if len(req.WorkItems) < maxDisplay {
		maxDisplay = len(req.WorkItems)
	}
	for i := 0; i < maxDisplay; i++ {
		wi := req.WorkItems[i]
		prompt.WriteString(fmt.Sprintf("- **%s**: %s\n", wi.Code, wi.Title))
		if len(wi.Tags) > 0 {
			prompt.WriteString(fmt.Sprintf("  Tags: %s\n", strings.Join(wi.Tags, ", ")))
		}
	}
	if len(req.WorkItems) > maxDisplay {
		prompt.WriteString(fmt.Sprintf("... and %d more work items\n", len(req.WorkItems)-maxDisplay))
	}
	prompt.WriteString("\n")

	// Existing roles context
	if len(req.ExistingRoles) > 0 {
		prompt.WriteString(fmt.Sprintf("## Existing Roles (%d)\n\n", len(req.ExistingRoles)))
		prompt.WriteString("**Avoid creating roles similar to these:**\n")
		for _, role := range req.ExistingRoles {
			prompt.WriteString(fmt.Sprintf("- %s: %s\n", role.Name, role.Description))
		}
		prompt.WriteString("\n")
	}

	// Request
	prompt.WriteString("## Task\n\n")
	prompt.WriteString("Based on the work items above, generate agent roles that would be needed to execute this work.\n")
	prompt.WriteString("Consider:\n")
	prompt.WriteString("- What types of specialized agents are needed?\n")
	prompt.WriteString("- What capabilities and skills should each role have?\n")
	prompt.WriteString("- What level of autonomy is appropriate for each role?\n")
	prompt.WriteString("- How should roles coordinate with each other?\n\n")
	prompt.WriteString("Generate roles as a JSON response.")

	return prompt.String()
}
