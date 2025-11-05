package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/sirupsen/logrus"
)

// Verify AIRolesBuilder implements RoleBuilderInterface
var _ builder.RoleBuilderInterface = (*AIRolesBuilder)(nil)

// AIRolesBuilder handles AI-powered role operations (generation, refinement, consolidation)
type AIRolesBuilder struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewAIRolesBuilder creates a new AI roles builder
func NewAIRolesBuilder(llmClient LLMClient, logger *logrus.Logger) *AIRolesBuilder {
	return &AIRolesBuilder{
		llmClient: llmClient,
		logger:    logger,
	}
}

// Stub methods - to be implemented following the goals/work items pattern

// RefineRole refines a role definition (to be implemented)
func (r *AIRolesBuilder) RefineRole(ctx context.Context, req *builder.RefineRoleRequest, builderContext builder.BuilderContext) (*builder.RefineRoleResponse, error) {
	// TODO: Implement role refinement
	return nil, fmt.Errorf("RefineRole not yet implemented")
}

// GenerateRole generates a single role (to be implemented)
func (r *AIRolesBuilder) GenerateRole(ctx context.Context, req *builder.GenerateRoleRequest, builderContext builder.BuilderContext) (*builder.GenerateRoleResponse, error) {
	// TODO: Implement single role generation
	return nil, fmt.Errorf("GenerateRole not yet implemented")
}

// ConsolidateRoles consolidates roles into a lean list (to be implemented)
func (r *AIRolesBuilder) ConsolidateRoles(ctx context.Context, req *builder.ConsolidateRolesRequest, builderContext builder.BuilderContext) (*builder.ConsolidateRolesResponse, error) {
	// TODO: Implement role consolidation
	return nil, fmt.Errorf("ConsolidateRoles not yet implemented")
}

// GenerateRoles uses AI to generate roles from work items
func (r *AIRolesBuilder) GenerateRoles(ctx context.Context, req *builder.GenerateRolesRequest, builderContext builder.BuilderContext) (*builder.GenerateRolesResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI role generation from work items")

	// Build the prompt for role generation
	prompt := r.buildRoleGenerationPrompt(req, builderContext)

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
	var aiResponse builder.GenerateRolesResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		r.logger.WithError(err).Error("Failed to parse AI role generation response")
		r.logger.WithField("response", cleanedContent).Debug("Raw AI response for debugging")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	r.logger.WithField("role_count", len(aiResponse.Roles)).Info("AI role generation completed")
	return &aiResponse, nil
}

func (r *AIRolesBuilder) getRoleGenerationSystemPrompt() string {
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

func (r *AIRolesBuilder) buildRoleGenerationPrompt(_ *builder.GenerateRolesRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	// Request
	builder.WriteString("\n## Task\n\n")
	builder.WriteString("Based on the work items above, generate agent roles that would be needed to execute this work.\n")
	builder.WriteString("Consider:\n")
	builder.WriteString("- What types of specialized agents are needed?\n")
	builder.WriteString("- What capabilities and skills should each role have?\n")
	builder.WriteString("- What level of autonomy is appropriate for each role?\n")
	builder.WriteString("- How should roles coordinate with each other?\n\n")
	builder.WriteString("Generate roles as a JSON response.")

	return builder.String()
}
