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
var _ builder.RoleBuilderInterface = (*RolesBuilder)(nil)

// RolesBuilder handles AI-powered role operations (generation, refinement, consolidation)
type RolesBuilder struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewAIRolesBuilder creates a new AI roles builder
func NewAIRolesBuilder(llmClient LLMClient, logger *logrus.Logger) *RolesBuilder {
	return &RolesBuilder{
		llmClient: llmClient,
		logger:    logger,
	}
}

// RefineRoles dynamically determines and executes the appropriate role operation based on user message
func (r *RolesBuilder) RefineRoles(ctx context.Context, req *builder.RefineRolesRequest, builderContext builder.BuilderContext) (*builder.RefineRolesResponse, error) {
	r.logger.WithFields(logrus.Fields{
		"agency_id":      req.AgencyID,
		"user_message":   req.UserMessage,
		"target_roles":   len(req.TargetRoles),
		"existing_roles": len(req.ExistingRoles),
	}).Info("Starting dynamic role refinement")

	// Build the prompt to determine what action to take
	prompt := r.buildDynamicRolesPrompt(req, builderContext)

	// Make the LLM request to determine action
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: dynamicRolesSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for dynamic role refinement")
		return nil, fmt.Errorf("AI refinement failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var result builder.RefineRolesResponse
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		r.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse dynamic roles response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"action":           result.Action,
		"refined_count":    len(result.RefinedRoles),
		"generated_count":  len(result.GeneratedRoles),
		"no_action_needed": result.NoActionNeeded,
	}).Info("Dynamic role refinement completed")

	return &result, nil
}

// RefineRolesStream performs dynamic role refinement with streaming support
// Similar to RefineRoles but streams chunks to the callback as they arrive from the LLM
func (r *RolesBuilder) RefineRolesStream(ctx context.Context, req *builder.RefineRolesRequest, builderContext builder.BuilderContext, streamCallback builder.StreamCallback) (*builder.RefineRolesResponse, error) {
	r.logger.WithFields(logrus.Fields{
		"agency_id":      req.AgencyID,
		"user_message":   req.UserMessage,
		"target_roles":   len(req.TargetRoles),
		"existing_roles": len(req.ExistingRoles),
	}).Info("Starting streaming dynamic role refinement")

	// Build the prompt
	prompt := r.buildDynamicRolesPrompt(req, builderContext)

	// Stream the LLM response
	var contentBuilder strings.Builder
	err := r.llmClient.ChatStream(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: dynamicRolesSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}, func(chunk string) error {
		// Accumulate content for final parsing
		contentBuilder.WriteString(chunk)

		// Forward chunk to the callback (for SSE streaming)
		if streamCallback != nil {
			return streamCallback(chunk)
		}
		return nil
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to stream AI response for dynamic role refinement")
		return nil, fmt.Errorf("AI streaming refinement failed: %w", err)
	}

	// Parse the accumulated response
	fullContent := contentBuilder.String()
	cleanedContent := stripMarkdownFences(fullContent)

	var result builder.RefineRolesResponse
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		r.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse streamed roles response")
		return nil, fmt.Errorf("failed to parse streamed response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"action":           result.Action,
		"refined_count":    len(result.RefinedRoles),
		"generated_count":  len(result.GeneratedRoles),
		"no_action_needed": result.NoActionNeeded,
	}).Info("Streaming dynamic role refinement completed")

	return &result, nil
}

// buildDynamicRolesPrompt creates the prompt for dynamic role processing
func (r *RolesBuilder) buildDynamicRolesPrompt(req *builder.RefineRolesRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("\n\n### USER REQUEST\n")
	builder.WriteString(req.UserMessage)
	builder.WriteString("\n\n")

	if len(req.TargetRoles) > 0 {
		builder.WriteString("### TARGET ROLES FOR OPERATION\n")
		for _, role := range req.TargetRoles {
			builder.WriteString(fmt.Sprintf("- **%s** (%s - %s): %s\n", role.Key, role.Code, role.Name, role.Description))
			if role.AutonomyLevel != "" {
				builder.WriteString(fmt.Sprintf("  Autonomy: %s\n", role.AutonomyLevel))
			}
			if role.TokenBudget > 0 {
				builder.WriteString(fmt.Sprintf("  Token Budget: %d\n", role.TokenBudget))
			}
			if len(role.Tags) > 0 {
				builder.WriteString(fmt.Sprintf("  Tags: %v\n", role.Tags))
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString("Based on the user's request and the agency context, determine what needs to be done with the roles and execute the appropriate action.")

	return builder.String()
}

// System prompts for role operations
const dynamicRolesSystemPrompt = SharedAgencyContext + `

Act as a strategic role management AI. Modify roles based on user requests.

CRITICAL: Roles are AGENT ROLES that define capabilities and responsibilities within the agency.
They are NOT job descriptions or organizational positions.

AGENT ROLE characteristics (✅): What agents can do and their scope
- Clear capabilities definition
- Well-defined autonomy level (autonomous, semi-autonomous, supervised)
- Specific required skills
- Appropriate token budget for operations
- Aligned with agency work items and goals

Examples of GOOD roles:
- "Technical Reviewer" with capabilities: ["code_review", "architecture_analysis"]
- "Test Executor" with capabilities: ["run_tests", "analyze_results", "report_issues"]
- "Deployment Coordinator" with capabilities: ["deploy_services", "monitor_deployments", "rollback"]

Examples of BAD roles (too generic or not agent-focused):
- "Software Engineer" (too broad, not agent-specific)
- "Manager" (not an agent capability)

## Actions:
**remove** - Delete roles (return in consolidated_data.removed_roles)
**refine** - Improve existing roles to be more capability-focused
**generate** - Create new agent roles aligned with work items
**consolidate** - Merge duplicate or overlapping roles
**enhance_all** - Refine all roles
**no_action** - Already optimal

## Response JSON:
{
  "action": "remove|refine|generate|consolidate|enhance_all|no_action",
  "refined_roles": [{"original_key": "key", "refined_name": "...", "refined_description": "...", "suggested_autonomy_level": "autonomous|semi-autonomous|supervised", "suggested_capabilities": [...], "suggested_skills": [...], "suggested_token_budget": 1000000, "suggested_tags": [...], "was_changed": true, "explanation": "Brief"}],
  "generated_roles": [{"name": "...", "description": "...", "autonomy_level": "autonomous|semi-autonomous|supervised", "capabilities": [...], "required_skills": [...], "token_budget": 1000000, "tags": [...], "explanation": "Brief"}],
  "consolidated_data": {"consolidated_roles": [...], "removed_roles": ["key1"], "summary": "Brief", "explanation": "Brief"},
  "explanation": "Brief overall summary",
  "no_action_needed": false
}

Guidelines:
- Roles = AGENT CAPABILITIES (what agents CAN DO), not job titles
- Define clear, specific capabilities (action-oriented)
- Set appropriate autonomy levels based on task complexity
- Align capabilities with agency work items
- Keep explanations concise (1-2 sentences)
- Token budgets should match role complexity and scope
- Autonomy levels: autonomous (no human intervention), semi-autonomous (occasional human input), supervised (requires human oversight)
- Required skills should be specific technical or domain skills needed`
