package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/sirupsen/logrus"
)

// GoalRefiner handles AI-powered goal definition and refinement
type GoalRefiner struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewGoalRefiner creates a new goal refiner service
func NewGoalRefiner(llmClient LLMClient, logger *logrus.Logger) *GoalRefiner {
	return &GoalRefiner{
		llmClient: llmClient,
		logger:    logger,
	}
}

// RefineGoalRequest contains the context for refining a goal
type RefineGoalRequest struct {
	AgencyID         string               `json:"agency_id"`
	CurrentGoal   *agency.Goal      `json:"current_goal"`
	Description      string               `json:"description"`
	Scope            string               `json:"scope"`
	SuccessMetrics   []string             `json:"success_metrics"`
	ExistingGoals []*agency.Goal    `json:"existing_goals"`
	UnitsOfWork      []*agency.UnitOfWork `json:"units_of_work"`
	AgencyContext    *agency.Agency       `json:"agency_context"`
}

// RefineGoalResponse contains the AI-refined goal
type RefineGoalResponse struct {
	RefinedDescription string   `json:"refined_description"`
	RefinedScope       string   `json:"refined_scope"`
	RefinedMetrics     []string `json:"refined_metrics"`
	SuggestedPriority  string   `json:"suggested_priority"`
	SuggestedCategory  string   `json:"suggested_category"`
	SuggestedTags      []string `json:"suggested_tags"`
	WasChanged         bool     `json:"was_changed"`
	Explanation        string   `json:"explanation"`
}

// aiGoalRefinementResponse represents the JSON structure returned by the AI
type aiGoalRefinementResponse struct {
	RefinedDescription string   `json:"refined_description"`
	RefinedScope       string   `json:"refined_scope"`
	RefinedMetrics     []string `json:"refined_metrics"`
	SuggestedPriority  string   `json:"suggested_priority"`
	SuggestedCategory  string   `json:"suggested_category"`
	SuggestedTags      []string `json:"suggested_tags"`
	Explanation        string   `json:"explanation"`
	Changed            bool     `json:"changed"`
}

// GenerateGoalRequest contains the context for generating a new goal
type GenerateGoalRequest struct {
	AgencyID         string               `json:"agency_id"`
	AgencyContext    *agency.Agency       `json:"agency_context"`
	ExistingGoals []*agency.Goal    `json:"existing_goals"`
	UnitsOfWork      []*agency.UnitOfWork `json:"units_of_work"`
	UserInput        string               `json:"user_input"`
}

// GenerateGoalResponse contains the AI-generated goal
type GenerateGoalResponse struct {
	Description       string   `json:"description"`
	Scope             string   `json:"scope"`
	SuccessMetrics    []string `json:"success_metrics"`
	SuggestedCode     string   `json:"suggested_code"`
	SuggestedPriority string   `json:"suggested_priority"`
	SuggestedCategory string   `json:"suggested_category"`
	SuggestedTags     []string `json:"suggested_tags"`
	Explanation       string   `json:"explanation"`
}

// RefineGoal uses AI to refine a goal definition based on all available context
func (r *GoalRefiner) RefineGoal(ctx context.Context, req *RefineGoalRequest) (*RefineGoalResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI goal refinement")

	// Build the prompt for goal refinement
	prompt := r.buildGoalRefinementPrompt(req)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: goalRefinementSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for goal refinement")
		return nil, fmt.Errorf("AI refinement failed: %w", err)
	}

	// Parse the AI response
	var aiResponse aiGoalRefinementResponse
	if err := json.Unmarshal([]byte(response.Content), &aiResponse); err != nil {
		r.logger.WithError(err).WithField("response", response.Content).Error("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Convert to our response format
	result := &RefineGoalResponse{
		RefinedDescription: aiResponse.RefinedDescription,
		RefinedScope:       aiResponse.RefinedScope,
		RefinedMetrics:     aiResponse.RefinedMetrics,
		SuggestedPriority:  aiResponse.SuggestedPriority,
		SuggestedCategory:  aiResponse.SuggestedCategory,
		SuggestedTags:      aiResponse.SuggestedTags,
		WasChanged:         aiResponse.Changed,
		Explanation:        aiResponse.Explanation,
	}

	r.logger.WithFields(logrus.Fields{
		"agency_id":   req.AgencyID,
		"was_changed": result.WasChanged,
		"description": len(result.RefinedDescription),
		"scope":       len(result.RefinedScope),
		"metrics":     len(result.RefinedMetrics),
	}).Info("AI goal refinement completed")

	return result, nil
}

// GenerateGoal uses AI to generate a new goal from user input
func (r *GoalRefiner) GenerateGoal(ctx context.Context, req *GenerateGoalRequest) (*GenerateGoalResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI goal generation")

	// Build the prompt for goal generation
	prompt := r.buildGoalGenerationPrompt(req)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: goalGenerationSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for goal generation")
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse the AI response
	var aiResponse GenerateGoalResponse
	if err := json.Unmarshal([]byte(response.Content), &aiResponse); err != nil {
		r.logger.WithError(err).WithField("response", response.Content).Error("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"agency_id":   req.AgencyID,
		"description": len(aiResponse.Description),
		"scope":       len(aiResponse.Scope),
		"metrics":     len(aiResponse.SuccessMetrics),
	}).Info("AI goal generation completed")

	return &aiResponse, nil
}

// buildGoalRefinementPrompt creates the prompt for goal refinement
func (r *GoalRefiner) buildGoalRefinementPrompt(req *RefineGoalRequest) string {
	var builder strings.Builder

	// Agency context
	builder.WriteString(fmt.Sprintf("Agency: %s (%s)\n", req.AgencyContext.DisplayName, req.AgencyContext.Category))
	builder.WriteString(fmt.Sprintf("Description: %s\n\n", req.AgencyContext.Description))

	// Current goal to refine
	builder.WriteString("Current Goal to Refine:\n")
	builder.WriteString(fmt.Sprintf("Description: %s\n", req.Description))
	if req.Scope != "" {
		builder.WriteString(fmt.Sprintf("Scope: %s\n", req.Scope))
	}
	if len(req.SuccessMetrics) > 0 {
		builder.WriteString(fmt.Sprintf("Success Metrics: %s\n", strings.Join(req.SuccessMetrics, ", ")))
	}
	builder.WriteString("\n")

	// Existing goals for context
	if len(req.ExistingGoals) > 0 {
		builder.WriteString("Existing Goals in Agency:\n")
		for i, goal := range req.ExistingGoals {
			if i < 5 { // Limit to avoid token overflow
				builder.WriteString(fmt.Sprintf("- %s: %s\n", goal.Code, goal.Description))
			}
		}
		builder.WriteString("\n")
	}

	// Units of work for context
	if len(req.UnitsOfWork) > 0 {
		builder.WriteString("Existing Work Items in Agency:\n")
		for i, unit := range req.UnitsOfWork {
			if i < 5 { // Limit to avoid token overflow
				builder.WriteString(fmt.Sprintf("- %s: %s\n", unit.Code, unit.Description))
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString("Please refine this goal definition to be clearer, more specific, and better aligned with the agency's purpose. Provide specific, measurable success metrics and suggest appropriate priority, category, and tags.")

	return builder.String()
}

// buildGoalGenerationPrompt creates the prompt for goal generation
func (r *GoalRefiner) buildGoalGenerationPrompt(req *GenerateGoalRequest) string {
	var builder strings.Builder

	// Agency context
	builder.WriteString(fmt.Sprintf("Agency: %s (%s)\n", req.AgencyContext.DisplayName, req.AgencyContext.Category))
	builder.WriteString(fmt.Sprintf("Description: %s\n\n", req.AgencyContext.Description))

	// User input
	builder.WriteString(fmt.Sprintf("User Request: %s\n\n", req.UserInput))

	// Existing goals for context and to avoid duplicates
	if len(req.ExistingGoals) > 0 {
		builder.WriteString("Existing Goals (to avoid duplication):\n")
		for i, goal := range req.ExistingGoals {
			if i < 10 { // Show more for duplication checking
				builder.WriteString(fmt.Sprintf("- %s: %s\n", goal.Code, goal.Description))
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString("Please generate a well-defined goal based on the user's request. Make it specific to this agency type and avoid duplicating existing goals. Include specific, measurable success metrics and suggest an appropriate goal code, priority, category, and tags.")

	return builder.String()
}

// System prompts for AI goal handling
const goalRefinementSystemPrompt = `You are an expert business analyst and goal definition specialist. Your role is to help refine and improve goal definitions for multi-agent systems and organizations.

When refining a goal, you should:
1. Make the description clearer and more specific
2. Define appropriate scope boundaries
3. Suggest concrete, measurable success metrics
4. Recommend priority level (High, Medium, Low)
5. Suggest category (Operational, Strategic, Technical, Financial, etc.)
6. Recommend relevant tags

Respond with JSON in this exact format:
{
  "refined_description": "Clear, specific goal description",
  "refined_scope": "Well-defined scope boundaries",
  "refined_metrics": ["Metric 1", "Metric 2", "Metric 3"],
  "suggested_priority": "High/Medium/Low",
  "suggested_category": "Category name",
  "suggested_tags": ["tag1", "tag2", "tag3"],
  "explanation": "Brief explanation of changes made",
  "changed": true/false
}

Focus on making goals actionable, measurable, and aligned with the agency's mission.`

const goalGenerationSystemPrompt = `You are an expert business analyst and goal definition specialist. Your role is to help generate well-defined goals for multi-agent systems and organizations based on user input.

When generating a goal, you should:
1. Create a clear, specific goal description
2. Define appropriate scope boundaries
3. Suggest concrete, measurable success metrics
4. Generate a unique goal code (follow pattern like P001, PROB-001, etc.)
5. Recommend priority level (High, Medium, Low)
6. Suggest category (Operational, Strategic, Technical, Financial, etc.)
7. Recommend relevant tags
8. Avoid duplicating existing goals

Respond with JSON in this exact format:
{
  "description": "Clear, specific goal description",
  "scope": "Well-defined scope boundaries",
  "success_metrics": ["Metric 1", "Metric 2", "Metric 3"],
  "suggested_code": "P001",
  "suggested_priority": "High/Medium/Low",
  "suggested_category": "Category name",
  "suggested_tags": ["tag1", "tag2", "tag3"],
  "explanation": "Brief explanation of the goal and how it fits the agency"
}

Focus on creating goals that are actionable, measurable, and aligned with the agency's mission.`
