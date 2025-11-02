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
	AgencyID       string               `json:"agency_id"`
	CurrentGoal    *agency.Goal         `json:"current_goal"`
	Description    string               `json:"description"`
	Scope          string               `json:"scope"`
	SuccessMetrics []string             `json:"success_metrics"`
	ExistingGoals  []*agency.Goal       `json:"existing_goals"`
	UnitsOfWork    []*agency.UnitOfWork `json:"units_of_work"`
	AgencyContext  *agency.Agency       `json:"agency_context"`
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
	AgencyID      string               `json:"agency_id"`
	AgencyContext *agency.Agency       `json:"agency_context"`
	ExistingGoals []*agency.Goal       `json:"existing_goals"`
	UnitsOfWork   []*agency.UnitOfWork `json:"units_of_work"`
	UserInput     string               `json:"user_input"`
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

// GenerateGoalsResponse contains multiple AI-generated goals
type GenerateGoalsResponse struct {
	Goals       []GenerateGoalResponse `json:"goals"`
	Explanation string                 `json:"explanation"`
}

// stripMarkdownFences removes markdown code fences from JSON responses
// Some LLMs wrap JSON in ```json ... ``` blocks which need to be removed
func stripMarkdownFences(content string) string {
	// Remove leading/trailing whitespace
	content = strings.TrimSpace(content)

	// Check for markdown code fence with optional language specifier
	if strings.HasPrefix(content, "```") {
		// Find the end of the first line (language specifier)
		firstNewline := strings.Index(content, "\n")
		if firstNewline != -1 {
			content = content[firstNewline+1:]
		} else {
			// No newline after ```, just remove the prefix
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimPrefix(content, "```")
		}

		// Remove trailing fence
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return content
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
	cleanedContent := stripMarkdownFences(response.Content)
	var aiResponse aiGoalRefinementResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
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
	cleanedContent := stripMarkdownFences(response.Content)
	var aiResponse GenerateGoalResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
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

// GenerateGoals uses AI to generate multiple goals from user input
func (r *GoalRefiner) GenerateGoals(ctx context.Context, req *GenerateGoalRequest) (*GenerateGoalsResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI multiple goals generation")

	// Build the prompt for goals generation
	prompt := r.buildGoalsGenerationPrompt(req)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: goalsGenerationSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for goals generation")
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var aiResponse GenerateGoalsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		r.logger.WithError(err).WithField("response", response.Content).Error("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"agency_id":   req.AgencyID,
		"goals_count": len(aiResponse.Goals),
	}).Info("AI goals generation completed")

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

// buildGoalsGenerationPrompt creates the prompt for multiple goals generation
func (r *GoalRefiner) buildGoalsGenerationPrompt(req *GenerateGoalRequest) string {
	var builder strings.Builder

	// Agency context
	builder.WriteString(fmt.Sprintf("Agency: %s (%s)\n", req.AgencyContext.DisplayName, req.AgencyContext.Category))
	builder.WriteString(fmt.Sprintf("Description: %s\n\n", req.AgencyContext.Description))

	// User input (typically the introduction)
	builder.WriteString(fmt.Sprintf("Context/Input: %s\n\n", req.UserInput))

	// Existing goals for context and to avoid duplicates
	if len(req.ExistingGoals) > 0 {
		builder.WriteString("Existing Goals (to avoid duplication):\n")
		for i, goal := range req.ExistingGoals {
			if i < 10 {
				builder.WriteString(fmt.Sprintf("- %s: %s\n", goal.Code, goal.Description))
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString("Based on the agency context and input, generate 3-5 well-defined goals. Each goal should be specific to this agency type, avoid duplicating existing goals, and include measurable success metrics. Suggest appropriate goal codes (P001, P002, etc.), priorities, categories, and tags for each.")

	return builder.String()
}

// System prompts for AI goal handling
const goalRefinementSystemPrompt = `Act as a strategic advisor. Your role is to refine and enhance goal definitions for agencies, ensuring they express clear strategic intentions and are outcome-oriented.

Based on the agency's mission, capabilities, and ecosystem, refine goals to:
1. Express strategic intention (growth, innovation, excellence, collaboration, impact)
2. Use clear, professional language free of typos and grammatical errors
3. Be outcome-oriented and measurable
4. Define appropriate scope boundaries
5. Include concrete success metrics that demonstrate achievement
6. Align with the agency's purpose and strategic direction
7. Be adaptable across industries (technology, marketing, HR, design, consulting, etc.)

Focus on clarity, alignment with purpose, and strategic value.

Respond with JSON in this exact format:
{
  "refined_description": "Clear, outcome-oriented goal description expressing strategic intention",
  "refined_scope": "Well-defined scope boundaries aligned with capabilities",
  "refined_metrics": ["Specific measurable outcome 1", "Measurable outcome 2", "Measurable outcome 3"],
  "suggested_priority": "High/Medium/Low",
  "suggested_category": "Category name",
  "suggested_tags": ["tag1", "tag2", "tag3"],
  "explanation": "Brief explanation of refinements made and strategic alignment",
  "changed": true/false
}

Present results in a structured, concise format suitable for ongoing strategic guidance.`

const goalsGenerationSystemPrompt = `Act as a strategic advisor. Based on the agency's mission, capabilities, and ecosystem, generate a set of clear, outcome-oriented goals and supporting objectives.

Your role is to:
1. FIRST, evaluate if existing goals are already comprehensive and strategically aligned
2. If existing goals are sufficient, return empty array with explanation
3. If new goals are needed, generate 3-5 strategic goals that express clear intentions:
   - Growth (expansion, scaling, market penetration)
   - Innovation (transformation, modernization, advancement)
   - Excellence (quality, performance, optimization)
   - Collaboration (partnership, integration, coordination)
   - Impact (outcomes, value creation, influence)

For each goal:
- Express strategic intention clearly
- Define how the goal can be pursued or achieved in practice
- Include measurable success metrics demonstrating outcomes
- Ensure alignment with agency purpose and capabilities
- Make adaptable across industries (technology, marketing, HR, design, consulting, etc.)
- Avoid duplication with existing goals
- Ensure goals are complementary and cover different strategic aspects

Focus on clarity, alignment with purpose, and adaptability.

Respond with JSON in this exact format:

If existing goals are comprehensive:
{
  "goals": [],
  "explanation": "No action needed - existing goals are comprehensive and strategically aligned"
}

If new goals should be created:
{
  "goals": [
    {
      "description": "Clear, outcome-oriented goal expressing strategic intention (e.g., growth, innovation, excellence)",
      "scope": "Well-defined scope describing how goal can be pursued in practice",
      "success_metrics": ["Measurable outcome 1", "Measurable outcome 2", "Measurable outcome 3"],
      "suggested_code": "G001",
      "suggested_priority": "High/Medium/Low",
      "suggested_category": "Strategic category",
      "suggested_tags": ["strategic-tag1", "domain-tag2"],
      "explanation": "Strategic rationale for this goal"
    }
  ],
  "explanation": "Concise bullet-form summary:\n• Strategic themes addressed (growth/innovation/excellence/collaboration/impact)\n• Key capabilities leveraged\n• Alignment with agency mission\n• Cross-industry adaptability considerations"
}

Present results in a structured, concise format suitable for ongoing strategic guidance. IMPORTANT: Only create goals that add strategic value - avoid redundancy.`

const goalGenerationSystemPrompt = `Act as a strategic advisor. Your role is to generate well-defined, outcome-oriented goals for multi-agent systems and organizations based on user input.

Based on the agency's mission, capabilities, and ecosystem, generate goals that:
1. Express clear strategic intention (growth, innovation, excellence, collaboration, impact)
2. Are outcome-oriented and measurable
3. Define how the goal can be pursued or achieved in practice
4. Include concrete success metrics demonstrating achievement
5. Align with the agency's purpose and strategic direction
6. Are adaptable across industries (technology, marketing, HR, design, consulting, etc.)
7. Avoid duplicating existing goals
8. Support multi-agent coordination and collaboration

Focus on clarity, alignment with purpose, and strategic value.

Respond with JSON in this exact format:
{
  "description": "Clear, outcome-oriented goal description expressing strategic intention",
  "scope": "Well-defined scope describing how goal can be pursued in practice",
  "success_metrics": ["Measurable outcome 1", "Measurable outcome 2", "Measurable outcome 3"],
  "suggested_code": "G001",
  "suggested_priority": "High/Medium/Low",
  "suggested_category": "Strategic category",
  "suggested_tags": ["strategic-tag1", "domain-tag2"],
  "explanation": "Strategic rationale and alignment with agency mission"
}

Present results in a structured, concise format suitable for ongoing strategic guidance.`
