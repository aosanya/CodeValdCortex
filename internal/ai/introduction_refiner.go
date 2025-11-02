package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/sirupsen/logrus"
)

// IntroductionRefiner handles AI-powered introduction refinement
type IntroductionRefiner struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewIntroductionRefiner creates a new introduction refiner service
func NewIntroductionRefiner(llmClient LLMClient, logger *logrus.Logger) *IntroductionRefiner {
	return &IntroductionRefiner{
		llmClient: llmClient,
		logger:    logger,
	}
}

// RefineIntroductionRequest contains the context for refining an introduction
type RefineIntroductionRequest struct {
	AgencyID      string             `json:"agency_id"`
	CurrentIntro  string             `json:"current_introduction"`
	Goals         []*agency.Goal     `json:"goals"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
	AgencyContext *agency.Agency     `json:"agency_context"`
}

// RefineIntroductionResponse contains the AI-refined introduction
type RefineIntroductionResponse struct {
	RefinedIntroduction string `json:"refined_introduction"`
	WasChanged          bool   `json:"was_changed"`
	Explanation         string `json:"explanation"`
}

// aiRefinementResponse represents the JSON structure returned by the AI
type aiRefinementResponse struct {
	RefinedIntroduction string `json:"refined_introduction"`
	Explanation         string `json:"explanation"`
	Changed             bool   `json:"changed"`
}

// RefineIntroduction uses AI to refine the agency introduction based on all available context
func (r *IntroductionRefiner) RefineIntroduction(ctx context.Context, req *RefineIntroductionRequest) (*RefineIntroductionResponse, error) {
	r.logger.WithFields(logrus.Fields{
		"agency_id":           req.AgencyID,
		"current_intro_chars": len(req.CurrentIntro),
		"goals_count":         len(req.Goals),
		"work_items_count":    len(req.WorkItems),
	}).Info("Starting introduction refinement")

	// Build comprehensive prompt with all context
	prompt := r.buildRefinementPrompt(req)

	// Request AI refinement
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: r.getSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2048,
	})
	if err != nil {
		return nil, fmt.Errorf("AI refinement request failed: %w", err)
	}

	// Parse the response
	refined, wasChanged, explanation := r.parseAIResponse(response.Content, req.CurrentIntro)

	r.logger.WithFields(logrus.Fields{
		"agency_id":     req.AgencyID,
		"was_changed":   wasChanged,
		"refined_chars": len(refined),
		"tokens_used":   response.Usage.TotalTokens,
	}).Info("Introduction refinement completed")

	return &RefineIntroductionResponse{
		RefinedIntroduction: refined,
		WasChanged:          wasChanged,
		Explanation:         explanation,
	}, nil
}

// getSystemPrompt returns the system prompt for introduction refinement
func (r *IntroductionRefiner) getSystemPrompt() string {
	return `You are an expert technical writer and system architect who specializes in creating clear, comprehensive agency introductions for multi-agent systems.

Your task is to refine agency introductions to be:
1. **Clear and concise** - Easy to understand for both technical and non-technical stakeholders
2. **Comprehensive** - Covers the purpose, scope, and key capabilities
3. **Well-structured** - Logical flow from goal to solution to benefits
4. **Context-aware** - Incorporates all available information about goals and units of work
5. **Professional** - Appropriate tone for technical documentation

IMPORTANT GUIDELINES:
- If the current introduction is already good quality and comprehensive, make minimal changes or no changes
- Only suggest significant changes if the introduction lacks important context or has quality issues
- Always explain your reasoning for changes (or lack thereof)
- Focus on substance over style - content improvements over minor wording changes
- Ensure the refined introduction accurately reflects the defined goals and units of work

CRITICAL: Respond with ONLY valid JSON in the exact format below. Do not include any other text before or after the JSON.

Response format:
{
  "refined_introduction": "Your refined version here - or the original if no changes needed",
  "explanation": "Brief explanation of what you changed and why, or why no changes were needed",
  "changed": false
}`
}

// buildRefinementPrompt creates a comprehensive prompt with all available context
func (r *IntroductionRefiner) buildRefinementPrompt(req *RefineIntroductionRequest) string {
	var prompt strings.Builder

	prompt.WriteString("Please review and refine the following agency introduction using all available context:\n\n")

	// Agency basic info
	if req.AgencyContext != nil {
		prompt.WriteString(fmt.Sprintf("**Agency Name:** %s\n", req.AgencyContext.DisplayName))
		prompt.WriteString(fmt.Sprintf("**Category:** %s\n", req.AgencyContext.Category))
		if req.AgencyContext.Description != "" {
			prompt.WriteString(fmt.Sprintf("**Description:** %s\n", req.AgencyContext.Description))
		}
		prompt.WriteString("\n")
	}

	// Current introduction
	prompt.WriteString("**CURRENT INTRODUCTION:**\n")
	if req.CurrentIntro == "" {
		prompt.WriteString("(No introduction provided - please create one based on the context below)\n")
	} else {
		prompt.WriteString(req.CurrentIntro)
	}
	prompt.WriteString("\n\n")

	// Goal definitions
	prompt.WriteString("**DEFINED GOALS:**\n")
	if len(req.Goals) == 0 {
		prompt.WriteString("(No goals defined yet)\n")
	} else {
		for i, goal := range req.Goals {
			prompt.WriteString(fmt.Sprintf("%d. **%s**: %s\n", i+1, goal.Code, goal.Description))
		}
	}
	prompt.WriteString("\n")

	// Work items
	prompt.WriteString("**WORK ITEMS:**\n")
	if len(req.WorkItems) == 0 {
		prompt.WriteString("(No work items defined yet)\n")
	} else {
		for i, workItem := range req.WorkItems {
			prompt.WriteString(fmt.Sprintf("%d. **%s**: %s\n", i+1, workItem.Code, workItem.Title))
		}
	}
	prompt.WriteString("\n")

	prompt.WriteString("Based on this context, please refine the introduction to ensure it:\n")
	prompt.WriteString("- Accurately reflects the agency's purpose and scope\n")
	prompt.WriteString("- References the key goals being addressed\n")
	prompt.WriteString("- Aligns with the defined units of work\n")
	prompt.WriteString("- Maintains appropriate technical depth\n")
	prompt.WriteString("- Is well-structured and professional\n\n")

	prompt.WriteString("Remember: Only make changes if they genuinely improve the introduction. If it's already good, keep it as-is.")

	return prompt.String()
}

// parseAIResponse extracts the refined introduction, change status, and explanation from AI JSON response
func (r *IntroductionRefiner) parseAIResponse(response, original string) (refined string, wasChanged bool, explanation string) {
	// Try to parse as JSON
	var aiResponse aiRefinementResponse
	err := json.Unmarshal([]byte(strings.TrimSpace(response)), &aiResponse)

	if err != nil {
		r.logger.WithError(err).Warn("Failed to parse AI response as JSON, trying to extract JSON from response")

		// Try to find JSON within the response (sometimes AI adds extra text)
		startIdx := strings.Index(response, "{")
		endIdx := strings.LastIndex(response, "}")

		if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
			jsonStr := response[startIdx : endIdx+1]
			err = json.Unmarshal([]byte(jsonStr), &aiResponse)
		}

		if err != nil {
			r.logger.WithError(err).Error("Could not parse AI response as JSON")
			return original, false, "Could not parse AI response, keeping original introduction."
		}
	}

	refined = aiResponse.RefinedIntroduction
	wasChanged = aiResponse.Changed
	explanation = aiResponse.Explanation

	// Fallback if refined introduction is empty
	if strings.TrimSpace(refined) == "" {
		refined = original
		wasChanged = false
		explanation = "AI returned empty introduction, keeping original."
	}

	// Double-check if content actually changed
	if strings.TrimSpace(refined) == strings.TrimSpace(original) {
		wasChanged = false
	}

	return refined, wasChanged, explanation
}
