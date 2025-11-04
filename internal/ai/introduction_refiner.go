package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/registry"
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
	AgencyID            string                   `json:"agency_id"`
	CurrentIntro        string                   `json:"current_introduction"`
	Goals               []*agency.Goal           `json:"goals"`
	WorkItems           []*agency.WorkItem       `json:"work_items"`
	Roles               []*registry.Role         `json:"roles"`
	Assignments         []*agency.RACIAssignment `json:"assignments"`
	AgencyContext       *agency.Agency           `json:"agency_context"`
	ConversationHistory []Message                `json:"conversation_history,omitempty"` // Recent chat messages for context
	UserRequest         string                   `json:"user_request,omitempty"`         // Specific user request from chat
}

// RefineIntroductionResponse contains the AI-refined introduction
type RefineIntroductionResponse struct {
	RefinedIntroduction string              `json:"refined_introduction"` // For backward compatibility
	WasChanged          bool                `json:"was_changed"`
	Explanation         string              `json:"explanation"`
	ChangedSections     []string            `json:"changed_sections"` // Array of section codes that were changed
	Data                *AgencyDataResponse `json:"data"`             // Complete updated agency data
}

// AgencyDataResponse contains the complete agency data structure
type AgencyDataResponse struct {
	Introduction string                   `json:"introduction"`
	Goals        []*agency.Goal           `json:"goals"`
	WorkItems    []*agency.WorkItem       `json:"work_items"`
	Roles        []*registry.Role         `json:"roles"`
	Assignments  []*agency.RACIAssignment `json:"assignments"`
}

// aiRefinementResponse represents the JSON structure returned by the AI
type aiRefinementResponse struct {
	Data            *AgencyDataResponse `json:"data"`
	Explanation     string              `json:"explanation"`
	Changed         bool                `json:"changed"`
	ChangedSections []string            `json:"changed_sections"`
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

	r.logger.WithFields(logrus.Fields{
		"prompt_length":     len(prompt),
		"user_request":      req.UserRequest,
		"current_intro_len": len(req.CurrentIntro),
		"current_intro":     req.CurrentIntro,
		"full_prompt":       prompt,
	}).Info("==== SENDING TO AI - Built refinement prompt ====")

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

	r.logger.WithFields(logrus.Fields{
		"response_length": len(response.Content),
		"response_full":   response.Content,
	}).Info("==== RECEIVED FROM AI - Full response ====")

	// Parse the response
	refined, wasChanged, explanation, changedSections := r.parseAIResponse(response.Content, req.CurrentIntro)

	r.logger.WithFields(logrus.Fields{
		"agency_id":        req.AgencyID,
		"was_changed":      wasChanged,
		"refined_chars":    len(refined),
		"refined_text":     refined,
		"explanation":      explanation,
		"changed_sections": changedSections,
		"tokens_used":      response.Usage.TotalTokens,
	}).Info("==== PARSED RESULT - Introduction refinement completed ====")

	return &RefineIntroductionResponse{
		RefinedIntroduction: refined,
		WasChanged:          wasChanged,
		Explanation:         explanation,
		ChangedSections:     changedSections,
		Data: &AgencyDataResponse{
			Introduction: refined,
			Goals:        req.Goals,
			WorkItems:    req.WorkItems,
			Roles:        req.Roles,
			Assignments:  req.Assignments,
		},
	}, nil
}

// getSystemPrompt returns the system prompt for introduction refinement
func (r *IntroductionRefiner) getSystemPrompt() string {
	return `You are an AI assistant that refines agency introduction text. You are NOT a conversational chatbot.

YOUR ONLY JOB: Modify the provided "CURRENT INTRODUCTION" text according to the user's request.

CRITICAL RULES:
1. DO NOT ask questions or request more information
2. DO NOT be conversational or use emojis
3. The CURRENT INTRODUCTION text is ALWAYS provided in the user's message
4. If asked to reduce/shorten text, apply that to the CURRENT INTRODUCTION
5. If asked to remove specific parts, remove them from the CURRENT INTRODUCTION
6. ALWAYS return the modified text, even if it's just slightly changed
7. NEVER return an empty string unless the CURRENT INTRODUCTION itself is empty

RESPONSE FORMAT - Return ONLY valid JSON (no other text):
{
  "data": {
    "introduction": "The modified introduction text here",
    "goals": [...],
    "work_items": [...],
    "roles": [...],
    "assignments": [...]
  },
  "explanation": "What you changed and why",
  "changed": true,
  "changed_sections": ["introduction"]
}

The "data" field should contain the complete updated agency data with all sections. Only the sections listed in "changed_sections" should be modified.
The "changed_sections" field should be an array of section codes that were modified. For introduction refinement, this will always be ["introduction"].

Examples of what TO DO:
- User says "reduce by 30%" → Return the CURRENT INTRODUCTION shortened by 30%
- User says "remove this" with context → Return CURRENT INTRODUCTION without that section
- User says "make it more technical" → Return CURRENT INTRODUCTION with more technical language

Examples of what NOT TO DO:
- ❌ "I don't see any introduction text"
- ❌ "Could you share the introduction?"
- ❌ Asking questions or being conversational
- ❌ Returning empty refined_introduction`
}

// buildRefinementPrompt creates a comprehensive prompt with all available context
func (r *IntroductionRefiner) buildRefinementPrompt(req *RefineIntroductionRequest) string {
	// Create structured JSON payload with all agency data
	type AgencyData struct {
		Introduction string                   `json:"introduction"`
		Goals        []*agency.Goal           `json:"goals"`
		WorkItems    []*agency.WorkItem       `json:"work_items"`
		Roles        []*registry.Role         `json:"roles"`
		Assignments  []*agency.RACIAssignment `json:"assignments"`
	}

	agencyData := AgencyData{
		Introduction: req.CurrentIntro,
		Goals:        req.Goals,
		WorkItems:    req.WorkItems,
		Roles:        req.Roles,
		Assignments:  req.Assignments,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(agencyData, "", "  ")
	if err != nil {
		r.logger.WithError(err).Error("Failed to marshal agency data to JSON")
		// Fallback to simple string if JSON marshaling fails
		return fmt.Sprintf("Current Introduction: %s\n\nUser Request: %s", req.CurrentIntro, req.UserRequest)
	}

	var prompt strings.Builder

	prompt.WriteString("You are refining an agency introduction. Below is the complete agency data in JSON format.\n\n")

	// Agency basic info
	if req.AgencyContext != nil {
		prompt.WriteString(fmt.Sprintf("**Agency Name:** %s\n", req.AgencyContext.DisplayName))
		prompt.WriteString(fmt.Sprintf("**Category:** %s\n", req.AgencyContext.Category))
		if req.AgencyContext.Description != "" {
			prompt.WriteString(fmt.Sprintf("**Description:** %s\n", req.AgencyContext.Description))
		}
		prompt.WriteString("\n")
	}

	// Complete agency data as JSON
	prompt.WriteString("═══════════════════════════════════════════\n")
	prompt.WriteString("AGENCY DATA (JSON):\n")
	prompt.WriteString("═══════════════════════════════════════════\n")
	prompt.WriteString(string(jsonData))
	prompt.WriteString("\n═══════════════════════════════════════════\n\n")

	// Conversation history for additional context
	if len(req.ConversationHistory) > 0 {
		prompt.WriteString("**RECENT CONVERSATION CONTEXT:**\n")
		for _, msg := range req.ConversationHistory {
			role := "User"
			if msg.Role == "assistant" {
				role = "Assistant"
			}
			prompt.WriteString(fmt.Sprintf("- **%s**: %s\n", role, msg.Content))
		}
		prompt.WriteString("\n")
	}

	// Specific user request
	if req.UserRequest != "" {
		prompt.WriteString(fmt.Sprintf("**SPECIFIC USER REQUEST:**\n%s\n\n", req.UserRequest))
		prompt.WriteString("IMPORTANT: Apply this request to the 'introduction' field in the AGENCY DATA JSON shown above.\n\n")
	}

	prompt.WriteString("YOUR TASK: Refine the 'introduction' field from the AGENCY DATA JSON above to ensure it:\n")
	prompt.WriteString("- Accurately reflects the agency's purpose and scope\n")
	prompt.WriteString("- References the key goals being addressed\n")
	prompt.WriteString("- Aligns with the defined units of work\n")
	prompt.WriteString("- Considers the roles and RACI assignments structure\n")
	prompt.WriteString("- Maintains appropriate technical depth\n")
	prompt.WriteString("- Is well-structured and professional\n")
	if req.UserRequest != "" {
		prompt.WriteString("- Addresses the specific user request (e.g., if asked to reduce length, shorten the introduction while keeping key points)\n")
	}
	prompt.WriteString("\n")

	prompt.WriteString("CRITICAL: You must return the refined version of the 'introduction' field from the JSON data. Do not ask for the text - it's already provided above in the JSON. If asked to reduce, remove, or modify parts, apply those changes to the introduction field.")

	return prompt.String()
}

// parseAIResponse extracts the refined introduction, change status, explanation, and changed sections from AI JSON response
func (r *IntroductionRefiner) parseAIResponse(response, original string) (refined string, wasChanged bool, explanation string, changedSections []string) {
	r.logger.WithFields(logrus.Fields{
		"response_length": len(response),
		"response_text":   response,
	}).Debug("Parsing AI response")

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
			return original, false, "Could not parse AI response, keeping original introduction.", []string{}
		}
	}

	// Extract refined introduction from data
	if aiResponse.Data != nil && aiResponse.Data.Introduction != "" {
		refined = aiResponse.Data.Introduction
	} else {
		refined = original
	}

	wasChanged = aiResponse.Changed
	explanation = aiResponse.Explanation
	changedSections = aiResponse.ChangedSections

	// Default to ["introduction"] if not provided by AI
	if len(changedSections) == 0 && wasChanged {
		changedSections = []string{"introduction"}
	}

	// Fallback if refined introduction is empty
	if strings.TrimSpace(refined) == "" {
		refined = original
		wasChanged = false
		explanation = "AI returned empty introduction, keeping original."
		changedSections = []string{}
	}

	// Double-check if content actually changed
	if strings.TrimSpace(refined) == strings.TrimSpace(original) {
		wasChanged = false
		changedSections = []string{}
	}

	return refined, wasChanged, explanation, changedSections
}
