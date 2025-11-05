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
var _ builder.GoalBuilderInterface = (*AIGoalsBuilder)(nil)

// AIGoalsBuilder handles AI-powered goal definition and refinement
type AIGoalsBuilder struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewGoalRefiner creates a new goal refiner service
func NewGoalRefiner(llmClient LLMClient, logger *logrus.Logger) *AIGoalsBuilder {
	return &AIGoalsBuilder{
		llmClient: llmClient,
		logger:    logger,
	}
}

// aiGoalRefinementResponse represents the JSON structure returned by the AI for goal refinement
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

// RefineGoal uses AI to refine a goal definition based on agency context
func (r *AIGoalsBuilder) RefineGoal(ctx context.Context, req *builder.RefineGoalRequest, builderContext builder.BuilderContext) (*builder.RefineGoalResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI goal refinement")

	// Build the refinement prompt using the provided context
	prompt := r.buildGoalRefinementPrompt(req, builderContext)

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
	result := &builder.RefineGoalResponse{
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
func (r *AIGoalsBuilder) GenerateGoal(ctx context.Context, req *builder.GenerateGoalRequest, builderContext builder.BuilderContext) (*builder.GenerateGoalResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI goal generation")

	// Build the prompt for goal generation
	prompt := r.buildGoalGenerationPrompt(req, builderContext)

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
	var aiResponse builder.GenerateGoalResponse
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
func (r *AIGoalsBuilder) GenerateGoals(ctx context.Context, req *builder.GenerateGoalRequest, builderContext builder.BuilderContext) (*builder.GenerateGoalsResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI multiple goals generation")

	// Build the prompt for multiple goals generation
	prompt := r.buildGoalsGenerationPrompt(req, builderContext)

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
	var aiResponse builder.GenerateGoalsResponse
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

// ConsolidateGoals analyzes and consolidates goals into a lean, concise list
func (r *AIGoalsBuilder) ConsolidateGoals(ctx context.Context, req *builder.ConsolidateGoalsRequest, builderContext builder.BuilderContext) (*builder.ConsolidateGoalsResponse, error) {
	r.logger.WithFields(logrus.Fields{
		"agency_id":   req.AgencyID,
		"total_goals": len(req.CurrentGoals),
	}).Info("Starting goal consolidation")

	// Build the prompt for goal consolidation
	prompt := r.buildGoalConsolidationPrompt(req, builderContext)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
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
		r.logger.WithError(err).Error("Failed to get AI response for goal consolidation")
		return nil, fmt.Errorf("AI consolidation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)

	var consolidationResp builder.ConsolidateGoalsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &consolidationResp); err != nil {
		r.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse goal consolidation response")
		return nil, fmt.Errorf("failed to parse consolidation response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"original_count":     len(req.CurrentGoals),
		"consolidated_count": len(consolidationResp.ConsolidatedGoals),
		"removed_count":      len(consolidationResp.RemovedGoals),
	}).Info("Goal consolidation completed")

	return &consolidationResp, nil
}

// buildGoalRefinementPrompt creates the prompt for goal refinement
func (r *AIGoalsBuilder) buildGoalRefinementPrompt(_ *builder.RefineGoalRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("Please refine this goal definition to be clearer, more specific, and better aligned with the agency's purpose. Provide specific, measurable success metrics and suggest appropriate priority, category, and tags.")

	return builder.String()
}

// buildGoalGenerationPrompt creates the prompt for goal generation
func (r *AIGoalsBuilder) buildGoalGenerationPrompt(_ *builder.GenerateGoalRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("Please generate a well-defined goal based on the user's request. Make it specific to this agency type and avoid duplicating existing goals. Include specific, measurable success metrics and suggest an appropriate goal code, priority, category, and tags.")

	return builder.String()
}

// buildGoalsGenerationPrompt creates the prompt for multiple goals generation
func (r *AIGoalsBuilder) buildGoalsGenerationPrompt(_ *builder.GenerateGoalRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("Based on the agency context and input, generate 3-5 well-defined goals. Each goal should be specific to this agency type, avoid duplicating existing goals, and include measurable success metrics. Suggest appropriate goal codes (P001, P002, etc.), priorities, categories, and tags for each.")

	return builder.String()
}

// buildGoalConsolidationPrompt creates the prompt for goal consolidation
func (r *AIGoalsBuilder) buildGoalConsolidationPrompt(_ *builder.ConsolidateGoalsRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("\n\nPlease analyze these goals and consolidate them into a lean, strategic list. Aim to reduce the count by 30-50% while maintaining complete coverage.")

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

const goalConsolidationSystemPrompt = `Act as a strategic advisor. Your task is to analyze goals for multi-agent systems and determine if consolidation is beneficial.

IMPORTANT: Only consolidate goals when it truly adds value. If goals are already well-defined and strategically distinct, keep them separate.

Evaluate if consolidation is needed:

1. **Assess consolidation value**:
   - Check for duplicate or near-duplicate goals
   - Look for goals with significant strategic overlap
   - Identify goals that are really subgoals of a larger strategic objective
   - Determine if goals can be combined without losing strategic clarity
   - **If goals are distinct and well-scoped, DO NOT force consolidation**

2. **When consolidation IS beneficial**:
   - Merge duplicate or overlapping goals
   - Combine related objectives when it clarifies strategy
   - Ensure each consolidated goal remains measurable
   - Preserve all success metrics and requirements
   - Maintain clear strategic direction
   - Keep goals at appropriate strategic level

3. **When consolidation is NOT beneficial**:
   - Return empty arrays for consolidated_goals and removed_goals
   - Provide explanation that goals are already well-defined
   - Don't force consolidation just to reduce count

4. **Maintain strategic quality**:
   - Each goal should express clear strategic intention
   - Avoid creating overly broad or vague goals
   - Ensure proper categorization and prioritization
   - Balance strategic scope and achievability
   - Support effective agency planning

5. **Track merges accurately** (only when consolidating):
   - Record ALL original goal keys that were merged in "merged_from_keys"
   - List ALL goal keys to DELETE in "removed_goals"
   - Provide clear explanations of consolidation decisions

Focus on practical strategic management. Do not force consolidation.

Respond ONLY with valid JSON (no markdown, no explanations outside JSON) in this exact format:

If consolidation is NOT beneficial:
{
  "consolidated_goals": [],
  "removed_goals": [],
  "summary": "No consolidation needed - goals are already distinct and well-scoped",
  "explanation": "Each goal addresses a specific strategic objective and should remain independent"
}

If consolidation IS beneficial:
{
  "consolidated_goals": [
    {
      "description": "Clear, outcome-oriented goal description",
      "scope": "Well-defined scope describing strategic approach",
      "success_metrics": ["Measurable outcome 1", "Measurable outcome 2", "Measurable outcome 3"],
      "suggested_code": "G001",
      "suggested_priority": "High/Medium/Low",
      "suggested_category": "Strategic category",
      "suggested_tags": ["tag1", "tag2"],
      "merged_from_keys": ["original_key1", "original_key2"],
      "explanation": "Brief explanation of what was consolidated and why"
    }
  ],
  "removed_goals": ["original_key1", "original_key2"],
  "summary": "Consolidated X goals into Y more focused strategic objectives",
  "explanation": "Overall consolidation strategy and benefits"
}
`
