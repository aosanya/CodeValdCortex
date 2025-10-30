package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/sirupsen/logrus"
)

// GoalConsolidator handles AI-powered goal consolidation
type GoalConsolidator struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewGoalConsolidator creates a new goal consolidator service
func NewGoalConsolidator(llmClient LLMClient, logger *logrus.Logger) *GoalConsolidator {
	return &GoalConsolidator{
		llmClient: llmClient,
		logger:    logger,
	}
}

// ConsolidateGoalsRequest contains the context for consolidating goals
type ConsolidateGoalsRequest struct {
	AgencyID      string               `json:"agency_id"`
	AgencyContext *agency.Agency       `json:"agency_context"`
	CurrentGoals  []*agency.Goal       `json:"current_goals"`
	UnitsOfWork   []*agency.UnitOfWork `json:"units_of_work"`
}

// ConsolidateGoalsResponse contains the consolidated goals
type ConsolidateGoalsResponse struct {
	ConsolidatedGoals []ConsolidatedGoal `json:"consolidated_goals"`
	RemovedGoals      []string           `json:"removed_goals"` // Keys of goals that were consolidated/removed
	Summary           string             `json:"summary"`
	Explanation       string             `json:"explanation"`
}

// ConsolidatedGoal represents a goal after consolidation
type ConsolidatedGoal struct {
	Description       string   `json:"description"`
	Scope             string   `json:"scope"`
	SuccessMetrics    []string `json:"success_metrics"`
	SuggestedCode     string   `json:"suggested_code"`
	SuggestedPriority string   `json:"suggested_priority"`
	SuggestedCategory string   `json:"suggested_category"`
	SuggestedTags     []string `json:"suggested_tags"`
	MergedFromKeys    []string `json:"merged_from_keys"` // Keys of original goals that were merged
	Explanation       string   `json:"explanation"`
}

// ConsolidateGoals analyzes and consolidates goals into a lean, concise list
func (c *GoalConsolidator) ConsolidateGoals(ctx context.Context, req *ConsolidateGoalsRequest) (*ConsolidateGoalsResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"agency_id":   req.AgencyID,
		"total_goals": len(req.CurrentGoals),
	}).Info("Starting goal consolidation")

	// Build the prompt for goal consolidation
	prompt := c.buildConsolidationPrompt(req)

	// Make the LLM request
	response, err := c.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: goalConsolidationSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		c.logger.WithError(err).Error("Failed to get AI response for goal consolidation")
		return nil, fmt.Errorf("AI consolidation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)

	var consolidationResp ConsolidateGoalsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &consolidationResp); err != nil {
		c.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse goal consolidation response")
		return nil, fmt.Errorf("failed to parse consolidation response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"original_count":     len(req.CurrentGoals),
		"consolidated_count": len(consolidationResp.ConsolidatedGoals),
		"removed_count":      len(consolidationResp.RemovedGoals),
	}).Info("Goal consolidation completed")

	return &consolidationResp, nil
}

// buildConsolidationPrompt creates the prompt for goal consolidation
func (c *GoalConsolidator) buildConsolidationPrompt(req *ConsolidateGoalsRequest) string {
	var builder strings.Builder

	// Agency context
	builder.WriteString(fmt.Sprintf("Agency: %s\n", req.AgencyContext.Name))
	builder.WriteString(fmt.Sprintf("Description: %s\n", req.AgencyContext.Description))
	builder.WriteString(fmt.Sprintf("Total Goals: %d\n\n", len(req.CurrentGoals)))

	// Current goals list
	builder.WriteString("Current Goals:\n")
	for i, goal := range req.CurrentGoals {
		builder.WriteString(fmt.Sprintf("\n%d. [%s] %s\n", i+1, goal.Key, goal.Description))
		if goal.Scope != "" {
			builder.WriteString(fmt.Sprintf("   Scope: %s\n", goal.Scope))
		}
		if len(goal.SuccessMetrics) > 0 {
			builder.WriteString(fmt.Sprintf("   Metrics: %s\n", strings.Join(goal.SuccessMetrics, ", ")))
		}
		if goal.Priority != "" {
			builder.WriteString(fmt.Sprintf("   Priority: %s\n", goal.Priority))
		}
	}

	// Units of work for context
	if len(req.UnitsOfWork) > 0 {
		builder.WriteString("\n\nExisting Work Items:\n")
		for i, unit := range req.UnitsOfWork {
			if i < 10 { // Limit to avoid token overflow
				builder.WriteString(fmt.Sprintf("- %s: %s\n", unit.Code, unit.Description))
			}
		}
	}

	builder.WriteString("\n\nPlease analyze these goals and consolidate them into a lean, strategic list. Aim to reduce the count by 30-50% while maintaining complete coverage.")

	return builder.String()
}

// goalConsolidationSystemPrompt defines the AI's role for goal consolidation
const goalConsolidationSystemPrompt = `You are an expert at agency design and goal consolidation. Your task is to analyze goals and consolidate them into a lean, strategic list.

Your consolidation strategy should:

1. **Identify goals to merge**:
   - Duplicates or near-duplicates
   - Goals with significant scope overlap
   - Granular goals that fit as sub-components of broader goals
   - Goals better expressed as success metrics of another goal

2. **Create consolidated goals**:
   - Merge similar/related goals into comprehensive parent goals
   - Keep distinct, non-overlapping strategic goals
   - Ensure each consolidated goal is clear, actionable, and measurable
   - Preserve the intent and value of all original goals
   - Include comprehensive success metrics from all merged goals

3. **Maintain quality**:
   - Each consolidated goal should be SMART (Specific, Measurable, Achievable, Relevant, Time-bound)
   - Avoid overly broad or vague consolidations
   - Ensure balanced coverage across different agency areas
   - Aim for 30-50% reduction in goal count

4. **Track merges accurately**:
   - Record all original goal keys that were merged into each new goal
   - Provide clear explanations of consolidation decisions

Respond ONLY with valid JSON (no markdown, no explanations outside JSON) in this exact format:
{
  "consolidated_goals": [
    {
      "description": "Clear, comprehensive goal description that captures all merged goals",
      "scope": "Well-defined scope boundaries that cover all merged goals",
      "success_metrics": ["Specific metric 1", "Measurable metric 2", "Actionable metric 3"],
      "suggested_code": "G001",
      "suggested_priority": "High/Medium/Low",
      "suggested_category": "Category name",
      "suggested_tags": ["tag1", "tag2"],
      "merged_from_keys": ["original_key1", "original_key2", "original_key3"],
      "explanation": "Brief explanation of what goals were merged and why"
    }
  ],
  "removed_goals": ["key1", "key2", "key3"],
  "summary": "Consolidated X goals into Y goals, achieving Z% reduction",
  "explanation": "Overall consolidation strategy: focused on merging [specific areas], maintained distinct [other areas], ensured comprehensive coverage of [key aspects]"
}

Focus on creating a lean, strategic set of goals that eliminate redundancy while maintaining complete coverage of the agency's mission.`
