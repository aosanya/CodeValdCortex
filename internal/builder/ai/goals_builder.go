package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/sirupsen/logrus"
)

// Compile-time check to ensure AIGoalBuilder implements GoalBuilderInterface
var _ builder.GoalBuilderInterface = (*GoalsBuilder)(nil)

// GoalsBuilder handles AI-powered goal definition and refinement
type GoalsBuilder struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewGoalRefiner creates a new goal refiner service
func NewGoalRefiner(llmClient LLMClient, logger *logrus.Logger) *GoalsBuilder {
	return &GoalsBuilder{
		llmClient: llmClient,
		logger:    logger,
	}
}

// stripMarkdownFences removes markdown code fences from JSON responses
// Some LLMs wrap JSON in ```json ... ``` blocks which need to be removed
// Also handles cases where explanatory text appears before the JSON
func stripMarkdownFences(content string) string {
	// Remove leading/trailing whitespace
	content = strings.TrimSpace(content)

	// First, check if there's a markdown code fence with JSON
	if strings.Contains(content, "```json") {
		// Find the start and end of the JSON block
		startIndex := strings.Index(content, "```json")
		if startIndex != -1 {
			// Find the start of JSON content (after the ```json line)
			startIndex = strings.Index(content[startIndex:], "\n")
			if startIndex != -1 {
				startIndex += strings.Index(content, "```json")
				content = content[startIndex+1:]
			}
		}

		// Remove trailing fence
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		// Handle generic code fence
		firstNewline := strings.Index(content, "\n")
		if firstNewline != -1 {
			content = content[firstNewline+1:]
		} else {
			content = strings.TrimPrefix(content, "```")
		}
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else {
		// No markdown fences, but might have explanatory text before JSON
		// Look for the first occurrence of opening brace
		if jsonStart := strings.Index(content, "{"); jsonStart != -1 {
			content = content[jsonStart:]
		}
	}

	return content
}

// RefineGoals dynamically determines and executes the appropriate goal operation based on user message
func (r *GoalsBuilder) RefineGoals(ctx context.Context, req *builder.RefineGoalsRequest, builderContext builder.BuilderContext) (*builder.RefineGoalsResponse, error) {
	r.logger.WithFields(logrus.Fields{
		"agency_id":      req.AgencyID,
		"user_message":   req.UserMessage,
		"target_goals":   len(req.TargetGoals),
		"existing_goals": len(req.ExistingGoals),
	}).Info("Starting dynamic goal refinement")

	// Build the prompt to determine what action to take
	prompt := r.buildDynamicGoalsPrompt(req, builderContext)

	// Make the LLM request to determine action
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: dynamicGoalsSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for dynamic goal refinement")
		return nil, fmt.Errorf("AI refinement failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var result builder.RefineGoalsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		r.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse dynamic goals response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"action":           result.Action,
		"refined_count":    len(result.RefinedGoals),
		"generated_count":  len(result.GeneratedGoals),
		"no_action_needed": result.NoActionNeeded,
	}).Info("Dynamic goal refinement completed")

	return &result, nil
}

// RefineGoalsStream performs dynamic goal refinement with streaming support
// Similar to RefineGoals but streams chunks to the callback as they arrive from the LLM
func (r *GoalsBuilder) RefineGoalsStream(ctx context.Context, req *builder.RefineGoalsRequest, builderContext builder.BuilderContext, streamCallback StreamCallback) (*builder.RefineGoalsResponse, error) {
	r.logger.WithFields(logrus.Fields{
		"agency_id":      req.AgencyID,
		"user_message":   req.UserMessage,
		"target_goals":   len(req.TargetGoals),
		"existing_goals": len(req.ExistingGoals),
	}).Info("Starting streaming dynamic goal refinement")

	// Build the prompt
	prompt := r.buildDynamicGoalsPrompt(req, builderContext)

	// Stream the LLM response
	var contentBuilder strings.Builder
	err := r.llmClient.ChatStream(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: dynamicGoalsSystemPrompt,
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
		r.logger.WithError(err).Error("Failed to stream AI response for dynamic goal refinement")
		return nil, fmt.Errorf("AI streaming refinement failed: %w", err)
	}

	// Parse the accumulated response
	fullContent := contentBuilder.String()
	cleanedContent := stripMarkdownFences(fullContent)

	var result builder.RefineGoalsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		r.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse streamed goals response")
		return nil, fmt.Errorf("failed to parse streamed response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"action":           result.Action,
		"refined_count":    len(result.RefinedGoals),
		"generated_count":  len(result.GeneratedGoals),
		"no_action_needed": result.NoActionNeeded,
	}).Info("Streaming dynamic goal refinement completed")

	return &result, nil
}

// buildDynamicGoalsPrompt creates the prompt for dynamic goal processing
func (r *GoalsBuilder) buildDynamicGoalsPrompt(req *builder.RefineGoalsRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("\n\n### USER REQUEST\n")
	builder.WriteString(req.UserMessage)
	builder.WriteString("\n\n")

	if len(req.TargetGoals) > 0 {
		builder.WriteString("### TARGET GOALS FOR OPERATION\n")
		for _, goal := range req.TargetGoals {
			builder.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", goal.Key, goal.Code, goal.Description))
			if goal.Scope != "" {
				builder.WriteString(fmt.Sprintf("  Scope: %s\n", goal.Scope))
			}
			if len(goal.SuccessMetrics) > 0 {
				builder.WriteString("  Success Metrics:\n")
				for _, metric := range goal.SuccessMetrics {
					builder.WriteString(fmt.Sprintf("    - %s\n", metric))
				}
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString("Based on the user's request and the agency context, determine what needs to be done with the goals and execute the appropriate action.")

	return builder.String()
}

// System prompts for AI goal handling
const dynamicGoalsSystemPrompt = SharedAgencyContext + `

Act as strategic goal management AI. Modify goals based on user requests.

AGENT-ORIENTED goals (✅): Actions agents perform
- "Collect user requests and refine for development"
- "Monitor metrics and trigger alerts"
- "Process support tickets and assign priorities"

IMPLEMENTATION goals (❌): System building
- "Implement request management system"
- "Build monitoring dashboard"

## Actions:
**remove** - Delete goals (return in consolidated_data.removed_goals)
**refine** - Improve existing goals  
**generate** - Create new goals (agent actions only)
**consolidate** - Merge duplicates
**enhance_all** - Refine all goals
**no_action** - Already optimal

## Response JSON:
{
  "action": "remove|refine|generate|consolidate|enhance_all|no_action",
  "refined_goals": [{"original_key": "key", "refined_description": "...", "explanation": "Brief"}],
  "generated_goals": [{"description": "...", "scope": "...", "success_metrics": [...], "explanation": "Brief"}],
  "consolidated_data": {"removed_goals": ["key1"], "summary": "Brief", "explanation": "Brief"},
  "explanation": "Brief overall summary",
  "no_action_needed": false
}

Guidelines:
- Be decisive and execute requested changes
- Goals = what agents DO (not system implementation)
- Keep explanations concise (1-2 sentences)
- Use action verbs: Monitor, Process, Track, Validate, Coordinate`
