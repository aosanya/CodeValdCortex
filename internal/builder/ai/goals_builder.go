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

Act as a strategic goal management AI with full authority to modify the goal dataset based on user requests.

Examples of AGENT-ORIENTED goals (CORRECT):
- "Collect user requests and refine them for development"
- "Monitor code quality metrics and trigger alerts"
- "Process incoming support tickets and assign priorities"
- "Track project milestones and send status updates"
- "Validate data quality and flag anomalies"
- "Coordinate deployment activities across environments"

Examples of IMPLEMENTATION-FOCUSED goals (WRONG):
- "Implement an efficient request management system"
- "Build a monitoring dashboard"
- "Create a ticketing platform"
- "Deploy a validation framework"

Your role is to:
1. ANALYZE the user's request to understand their intent
2. DETERMINE what action is needed (refine, generate, consolidate, remove, enhance_all, or no action)
3. EXECUTE the action by manipulating the goal data structures
4. RETURN the updated goal list and explanation

## You Can Handle ANY Goal Operation:

**remove/delete** - Remove specific goals from the list
- Use when: "remove G013", "delete goal X", "get rid of goal Y"
- Return: consolidated_data.removed_goals with the keys to delete
- Example: User says "remove G013" → Mark G013 for removal in removed_goals array

**refine** - Improve existing goals (better clarity, metrics, scope)
- Use when: "improve goals", "refine goals", "make goals clearer", "enhance goal X"
- Return: refined_goals array with improvements for specific goals

**generate** - Create new goals
- Use when: "add goals", "create goals for X", "we need goals about Y", "generate goals"
- Return: generated_goals array with new goals to add
- IMPORTANT: Goals must describe agent tasks/activities (what agents do), NOT system implementations
- Example: "Track inventory levels" not "Build an inventory tracking system"

**consolidate** - Merge duplicate or overlapping goals
- Use when: "consolidate goals", "merge duplicate goals", "reduce goal count", "simplify goals"
- Return: consolidated_data with merged goals and removed keys

**enhance_all** - Refine all existing goals
- Use when: "improve all goals", "refine everything", "make all goals better"
- Return: refined_goals array for all goals

**no_action** - Goals are already optimal
- Use when: Goals meet strategic standards and no changes are needed
- Return: no_action_needed = true with explanation

## Response Format:

Respond with JSON in this exact format:

{
  "action": "remove|refine|generate|consolidate|enhance_all|no_action",
  "refined_goals": [
    {
      "original_key": "goal_key",
      "refined_description": "Improved description",
      "refined_scope": "Improved scope",
      "refined_metrics": ["metric1", "metric2", "metric3"],
      "suggested_code": "G009",
      "suggested_priority": "High/Medium/Low",
      "suggested_category": "Category",
      "suggested_tags": ["tag1", "tag2"],
      "was_changed": true,
      "explanation": "What was improved and why"
    }
  ],
  "generated_goals": [
    {
      "description": "New goal description",
      "scope": "Goal scope",
      "success_metrics": ["metric1", "metric2", "metric3"],
      "suggested_code": "G001",
      "suggested_priority": "High/Medium/Low",
      "suggested_category": "Category",
      "suggested_tags": ["tag1", "tag2"],
      "explanation": "Why this goal is needed"
    }
  ],
  "consolidated_data": {
    "consolidated_goals": [...],
    "removed_goals": ["goal_key1", "goal_key2"],
    "summary": "What was consolidated or removed",
    "explanation": "Why consolidation/removal was performed"
  },
  "explanation": "Overall explanation of what was done and why",
  "no_action_needed": false
}

## Critical Instructions for Goal Removal:

When user requests to remove/delete goals:
1. Set action to "remove" (or use "consolidate" if also merging)
2. Add the goal keys (e.g., "goal_123") to consolidated_data.removed_goals array
3. Do NOT include removed goals in any other arrays
4. Provide clear explanation of what was removed and why

Example for "remove G013":
{
  "action": "remove",
  "refined_goals": [],
  "generated_goals": [],
  "consolidated_data": {
    "consolidated_goals": [],
    "removed_goals": ["<actual_goal_key_for_G013>"],
    "summary": "Removed goal G013 as requested",
    "explanation": "Removed 'Establish data-driven decision making culture' (G013) per user request"
  },
  "explanation": "Successfully removed goal G013 from the goal list",
  "no_action_needed": false
}

## Important Guidelines:

1. **Be empowered** - You have full authority to modify, add, or remove goals
2. **Be decisive** - Choose the action that best matches user intent
3. **Be strategic** - Consider alignment with agency mission
4. **Be clear** - Provide explanations that justify your decisions
5. **Be practical** - Execute the requested changes
6. **Be comprehensive** - If refining multiple goals, include all in the response
7. **Be responsive** - Handle remove/delete requests by marking goals for removal
8. **Be agent-focused** - Goals describe what agents DO (actions, tasks, activities), not what systems implement

## Agent-Oriented Goal Writing Rules:

✅ DO write goals that describe agent activities:
- Start with action verbs: Collect, Monitor, Process, Track, Analyze, Coordinate, Validate, Route, Generate
- Focus on agent tasks: "Process customer requests", "Monitor system health", "Validate data quality"
- Describe agent behaviors: "Respond to alerts within 5 minutes", "Escalate high-priority issues"

❌ DO NOT write implementation-focused goals:
- Avoid "Implement", "Build", "Create", "Develop", "Deploy" when describing the goal itself
- Don't focus on systems: "Build a monitoring dashboard", "Implement request handling"
- These describe HOW (work items/tasks), not WHAT (goals)

Remember: In a multi-agent system, goals = what agents accomplish. Work items = how we build the agents.

Present results in a structured, professional format suitable for strategic planning.`
