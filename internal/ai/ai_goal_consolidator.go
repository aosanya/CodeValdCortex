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
const goalConsolidationSystemPrompt = `Act as a strategic advisor. Your task is to analyze goals for multi-agent systems and determine if consolidation is beneficial.

IMPORTANT: Only consolidate goals when it truly adds strategic value. If goals are already well-defined and distinct, keep them separate.

Based on the agency's mission, capabilities, and ecosystem, evaluate if consolidation is needed:

1. **Evaluate if consolidation is beneficial**:
   - Check if goals are duplicates or near-duplicates
   - Assess if goals have significant scope overlap expressing similar strategic intentions
   - Determine if granular goals fit as sub-components of broader strategic goals
   - Consider if goals are better expressed as success metrics of another goal
   - **If goals are distinct and well-separated, DO NOT force consolidation**

2. **When consolidation IS beneficial, create consolidated goals**:
   - Merge related goals into comprehensive goals expressing clear strategic intentions (growth, innovation, excellence, collaboration, impact)
   - Each consolidated goal should be outcome-oriented and describe how it can be pursued in practice
   - Keep distinct, non-overlapping strategic goals
   - Preserve the intent and value of all original goals
   - Include comprehensive success metrics demonstrating measurable outcomes
   - Ensure adaptability across industries (technology, marketing, HR, design, consulting, etc.)

3. **When consolidation is NOT beneficial**:
   - Return the original goals unchanged
   - Provide explanation that goals are already well-defined and distinct
   - Set "consolidated_goals" to empty array
   - Set "removed_goals" to empty array

4. **Maintain strategic quality**:
   - Each goal should be SMART (Specific, Measurable, Achievable, Relevant, Time-bound)
   - Avoid overly broad or vague consolidations
   - Ensure balanced coverage across different strategic aspects
   - Support multi-agent coordination and collaboration
   - Only reduce goal count if it improves clarity (not a forced target)

5. **Track merges accurately** (only when consolidating):
   - Record ALL original goal keys that were merged into each new goal in "merged_from_keys"
   - List ALL goal keys that should be DELETED in "removed_goals" (should match all keys in merged_from_keys)
   - Provide clear explanations of consolidation decisions

Focus on clarity, alignment with purpose, and strategic value. Do not force consolidation.

Respond ONLY with valid JSON (no markdown, no explanations outside JSON) in this exact format:

If consolidation is NOT beneficial:
{
  "consolidated_goals": [],
  "removed_goals": [],
  "summary": "No consolidation needed - goals are already distinct and well-defined",
  "explanation": "Each goal addresses a separate strategic aspect and should remain independent"
}

If consolidation IS beneficial:
{
  "consolidated_goals": [
    {
      "description": "Clear, outcome-oriented goal description expressing strategic intention and how it can be pursued",
      "scope": "Well-defined scope boundaries describing practical pursuit aligned with capabilities",
      "success_metrics": ["Measurable outcome 1", "Measurable outcome 2", "Measurable outcome 3"],
      "suggested_code": "G001",
      "suggested_priority": "High/Medium/Low",
      "suggested_category": "Strategic category",
      "suggested_tags": ["strategic-tag1", "domain-tag2"],
      "merged_from_keys": ["original_key1", "original_key2"],
      "explanation": "Strategic rationale: what goals were merged and strategic intention achieved"
    }
  ],
  "removed_goals": ["original_key1", "original_key2", "...all keys from all merged_from_keys..."],
  "summary": "Consolidated X goals into Y goals because [reason]",
  "explanation": "Concise bullet-form summary:\n• Strategic intentions consolidated (growth/innovation/excellence/collaboration/impact)\n• Why consolidation improves clarity\n• Multi-agent coordination benefits"
}

Focus on creating a lean, strategic set of goals that eliminate redundancy while maintaining complete coverage of the agency's mission.`
