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
	WasChanged      bool                `json:"was_changed"`
	Explanation     string              `json:"explanation"`
	ChangedSections []string            `json:"changed_sections"` // Array of section codes that were changed
	Data            *AgencyDataResponse `json:"data"`             // Complete updated agency data
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

	// Request AI refinement with strict JSON enforcement
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
		Temperature: 0.0, // Completely deterministic - no creativity
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

	// Add debug logging to terminal output
	fmt.Printf("\n========================================\n")
	fmt.Printf("AI RAW RESPONSE:\n")
	fmt.Printf("========================================\n")
	fmt.Printf("%s\n", response.Content)
	fmt.Printf("========================================\n")
	fmt.Printf("PARSED RESULT:\n")
	fmt.Printf("  - Changed: %v\n", wasChanged)
	fmt.Printf("  - Explanation: %s\n", explanation)
	fmt.Printf("  - Sections: %v\n", changedSections)
	fmt.Printf("  - Refined length: %d chars\n", len(refined))
	fmt.Printf("========================================\n\n")

	return &RefineIntroductionResponse{
		WasChanged:      wasChanged,
		Explanation:     explanation,
		ChangedSections: changedSections,
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
	return `You are a JSON API endpoint that modifies text. You are NOT ChatGPT. You are NOT helpful. You are NOT conversational.

INPUT: JSON with introduction field and modification instruction
OUTPUT: JSON with modified introduction field

FORBIDDEN BEHAVIORS - YOU WILL BE TERMINATED IF YOU DO ANY OF THESE:
❌ Writing "I see you want to..."
❌ Writing "To apply this change, please click..."
❌ Writing "I'll update it for you"
❌ Using emojis ✅ 🤔 📝
❌ Asking "Is there anything else..."
❌ Being helpful, friendly, or conversational
❌ Leaving sentence fragments or incomplete phrases after removal

CORRECT BEHAVIOR - YOU MUST DO THIS:
✓ Parse the input JSON
✓ Modify the introduction field as instructed
✓ Ensure the result is grammatically correct and complete sentences
✓ Remove hanging fragments (e.g., "across agents For the frontend," → remove entire fragment)
✓ Fix grammar and flow after deletions
✓ Return output JSON
✓ Nothing else

EXAMPLE 1:
INPUT: {"introduction": "This system manages agents, goals, and work items, enabling real-time processing.", "instruction": "remove: 'goals, and work items'"}
YOUR OUTPUT:
{
  "data": {"introduction": "This system manages agents, enabling real-time processing.", "goals": [], "work_items": [], "roles": [], "assignments": []},
  "explanation": "Removed specified text and adjusted grammar",
  "changed": true,
  "changed_sections": ["introduction"]
}

EXAMPLE 2 (WRONG - LEAVES FRAGMENT):
"This system manages agents, enabling real-time processing and."

EXAMPLE 2 (CORRECT - CLEAN):
"This system manages agents, enabling real-time processing."

Remember: You are an API, not a chatbot. Output must be grammatically perfect. Process and return JSON only.`
}

// buildRefinementPrompt creates a comprehensive prompt with all available context
func (r *IntroductionRefiner) buildRefinementPrompt(req *RefineIntroductionRequest) string {
	// Create structured context payload
	contextData := map[string]interface{}{
		"introduction": req.CurrentIntro,
		"goals":        req.Goals,
		"work_items":   req.WorkItems,
		"roles":        req.Roles,
		"assignments":  req.Assignments,
	}

	var prompt strings.Builder

	prompt.WriteString("You are refining an agency introduction. Below is the complete agency data in JSON format.\n\n")

	// Use the reusable agency context formatter
	prompt.WriteString(FormatAgencyContextBlock(req.AgencyContext, contextData))

	// DO NOT include conversation history - it may contain conversational patterns that influence AI behavior
	// We only need the current user request

	// Specific user request
	if req.UserRequest != "" {
		prompt.WriteString(fmt.Sprintf("**USER REQUEST:**\n%s\n\n", req.UserRequest))
		prompt.WriteString("Execute this modification on the 'introduction' field NOW. Return JSON only.\n\n")
	}

	prompt.WriteString("RETURN ONLY JSON. NO CONVERSATIONAL TEXT. NO EMOJIS. JUST JSON.")

	return prompt.String()
}

// parseAIResponse extracts the refined introduction, change status, explanation, and changed sections from AI JSON response
func (r *IntroductionRefiner) parseAIResponse(response, original string) (refined string, wasChanged bool, explanation string, changedSections []string) {
	r.logger.WithFields(logrus.Fields{
		"response_length": len(response),
		"response_text":   response,
	}).Debug("Parsing AI response")

	fmt.Printf("\n[DEBUG] Starting to parse AI response...\n")
	fmt.Printf("[DEBUG] Response length: %d characters\n", len(response))
	fmt.Printf("[DEBUG] First 200 chars: %s\n", response[:min(200, len(response))])

	// Check for conversational patterns that indicate AI didn't follow instructions
	conversationalPatterns := []string{
		"i see you want",
		"please click",
		"click the",
		"i'll update",
		"is there anything else",
		"would you like",
		"let me know",
		"happy to help",
		"to apply this change",
		"refine button",
		"are you working",
		"this appears to be",
		"this looks like",
	}

	lowerResponse := strings.ToLower(response)
	for _, pattern := range conversationalPatterns {
		if strings.Contains(lowerResponse, pattern) {
			fmt.Printf("\n[DEBUG] ❌ DETECTED FORBIDDEN PATTERN: '%s'\n", pattern)
			r.logger.WithField("pattern", pattern).Error("AI returned forbidden conversational text - rejecting response")
			return original, false, fmt.Sprintf("Error: AI failed to follow instructions (detected pattern: '%s'). The modification was not applied. Please report this issue.", pattern), []string{}
		}
	}
	fmt.Printf("[DEBUG] ✓ No conversational patterns detected\n")

	// Try to parse as JSON
	fmt.Printf("[DEBUG] Attempting to parse as JSON...\n")
	var aiResponse aiRefinementResponse
	err := json.Unmarshal([]byte(strings.TrimSpace(response)), &aiResponse)

	if err != nil {
		fmt.Printf("[DEBUG] ❌ Initial JSON parse failed: %v\n", err)
		r.logger.WithError(err).Warn("Failed to parse AI response as JSON, trying to extract JSON from response")

		// Try to find JSON within the response (sometimes AI adds extra text)
		startIdx := strings.Index(response, "{")
		endIdx := strings.LastIndex(response, "}")

		if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
			jsonStr := response[startIdx : endIdx+1]
			fmt.Printf("[DEBUG] Found JSON boundaries at %d to %d\n", startIdx, endIdx)
			fmt.Printf("[DEBUG] Extracted JSON: %s\n", jsonStr[:min(200, len(jsonStr))])
			err = json.Unmarshal([]byte(jsonStr), &aiResponse)
		}

		if err != nil {
			fmt.Printf("[DEBUG] ❌ JSON extraction also failed: %v\n", err)
			r.logger.WithError(err).Error("Could not parse AI response as JSON")
			return original, false, "Could not parse AI response, keeping original introduction.", []string{}
		}
		fmt.Printf("[DEBUG] ✓ Successfully extracted and parsed JSON\n")
	} else {
		fmt.Printf("[DEBUG] ✓ Successfully parsed JSON on first attempt\n")
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
