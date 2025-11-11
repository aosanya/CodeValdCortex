package ai_refine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GoalsResponse captures the JSON response from RefineGoals
type GoalsResponse struct {
	Action           string      `json:"action"`
	RefinedGoals     interface{} `json:"refined_goals"`
	GeneratedGoals   interface{} `json:"generated_goals"`
	ConsolidatedData interface{} `json:"consolidated_data"`
	Explanation      string      `json:"explanation"`
	NoActionNeeded   bool        `json:"no_action_needed"`
	Summary          string      `json:"summary"`
}

// responseCapture is a custom response writer that captures the response body
type responseCapture struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write implements gin.ResponseWriter
func (w *responseCapture) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

// ProcessGoalsChatRequest handles chat-based goal interactions
// This wraps the RefineGoals method and handles chat-specific response formatting
func (h *Handler) ProcessGoalsChatRequest(c *gin.Context) {
	h.logger.Info("🔵 HANDLER CALLED: ProcessGoalsChatRequest")

	agencyID := c.Param("id")

	// Get user message from dynamic_request (set by chat_context_processor)
	dynamicReq, exists := c.Get("dynamic_request")
	if !exists {
		h.logger.Error("No dynamic_request found in context")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Invalid Request</strong>
						<p class="mb-0">Missing request data.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	req, ok := dynamicReq.(struct {
		UserMessage string   `json:"user_message"`
		GoalKeys    []string `json:"goal_keys"`
	})
	if !ok {
		h.logger.Error("Failed to cast dynamic_request to expected type")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Invalid Request Format</strong>
						<p class="mb-0">Unable to parse request data.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	userMessage := req.UserMessage
	if userMessage == "" {
		h.logger.Error("No user message provided")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>No Message Provided</strong>
						<p class="mb-0">Please provide a message.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":    agencyID,
		"user_message": userMessage,
		"goal_keys":    req.GoalKeys,
	}).Info("Processing chat-based goal request")

	// Get or create conversation
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		h.logger.Warn("No conversation exists, creating new one", "agencyID", agencyID)
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation")
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
				<div class="notification is-danger">
					<div class="is-flex is-align-items-center">
						<span class="icon has-text-danger mr-2">
							<i class="fas fa-exclamation-triangle"></i>
						</span>
						<div>
							<strong>Conversation Error</strong>
							<p class="mb-0">Failed to initialize conversation.</p>
						</div>
					</div>
				</div>
			`)
			return
		}
	}

	// Capture the response from RefineGoals
	// We'll use a custom response writer to intercept the JSON response

	// Store original writer
	originalWriter := c.Writer

	// Create capture buffer
	captureBuffer := &bytes.Buffer{}
	c.Writer = &responseCapture{
		ResponseWriter: c.Writer,
		body:           captureBuffer,
	}

	// Call RefineGoals - it will write JSON to our capture buffer
	h.RefineGoals(c)

	// Restore original writer
	c.Writer = originalWriter

	// Check if RefineGoals returned an error (non-200 status)
	if c.Writer.Status() >= 400 {
		// RefineGoals already set error HTML, just return
		h.logger.Error("RefineGoals returned an error")
		return
	}

	// Parse the captured JSON response
	var goalsResp GoalsResponse
	if err := json.Unmarshal(captureBuffer.Bytes(), &goalsResp); err != nil {
		h.logger.WithError(err).Error("Failed to parse RefineGoals response")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>Processing Error</strong>
						<p class="mb-0">Failed to process AI response.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Format chat message based on the action taken
	chatMessage := formatGoalsChatMessage(goalsResp)

	// Add the message to the conversation
	h.logger.Info("Adding goals AI response to chat",
		"agencyID", agencyID,
		"conversationID", conversation.ID,
		"messageLength", len(chatMessage))

	if err := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); err != nil {
		h.logger.WithError(err).Error("Failed to add message to conversation")
	}

	// Reload conversation to get the updated messages
	conversation, err = h.designerService.GetConversation(conversation.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to reload conversation")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, "Failed to load conversation")
		return
	}

	// Render only the AI message (user message was already added by JavaScript)
	if len(conversation.Messages) > 0 {
		lastMessage := conversation.Messages[len(conversation.Messages)-1]
		if lastMessage.Role == "assistant" {
			component := agency_designer.AIMessage(lastMessage)
			c.Header("Content-Type", "text/html")
			err = component.Render(c.Request.Context(), c.Writer)
			if err != nil {
				h.logger.WithError(err).Error("Failed to render AI message")
				c.String(http.StatusInternalServerError, "Failed to render message")
				return
			}
		}
	}
}

// formatGoalsChatMessage formats the goals AI response for chat display
func formatGoalsChatMessage(resp GoalsResponse) string {
	var message strings.Builder

	// Add emoji and title based on action
	if resp.NoActionNeeded {
		message.WriteString("✅ **Goals Review Complete**\n\n")
	} else {
		switch resp.Action {
		case "refine":
			message.WriteString("✨ **Goals Refined**\n\n")
		case "generate":
			message.WriteString("🎯 **Goals Generated**\n\n")
		case "consolidate":
			message.WriteString("📊 **Goals Consolidated**\n\n")
		case "enhance_all":
			message.WriteString("⚡ **Goals Enhanced**\n\n")
		default:
			message.WriteString("✨ **Goals Updated**\n\n")
		}
	}

	// Add summary if available
	if resp.Summary != "" {
		message.WriteString(resp.Summary)
		message.WriteString("\n\n")
	}

	// Add explanation
	if resp.Explanation != "" {
		message.WriteString(resp.Explanation)
	}

	return message.String()
}

// ProcessGoalsChatRequestStreaming handles chat-based goal interactions with streaming
// Uses the shared ExecuteChatStreaming helper for consistency with introduction
func (h *Handler) ProcessGoalsChatRequestStreaming(c *gin.Context) {
	h.logger.Info("🌊 HANDLER CALLED: ProcessGoalsChatRequestStreaming")

	agencyID := c.Param("id")

	// Get user message from dynamic_request (set by chat_context_processor)
	dynamicReq, exists := c.Get("dynamic_request")
	if !exists {
		h.logger.Error("No dynamic_request found in context")
		c.SSEvent("error", `{"error": "Missing request data"}`)
		return
	}

	req, ok := dynamicReq.(struct {
		UserMessage string   `json:"user_message"`
		GoalKeys    []string `json:"goal_keys"`
	})
	if !ok {
		h.logger.Error("Failed to cast dynamic_request to expected type")
		c.SSEvent("error", `{"error": "Invalid request format"}`)
		return
	}

	userMessage := req.UserMessage
	if userMessage == "" {
		h.logger.Error("No user message provided")
		c.SSEvent("error", `{"error": "No message provided"}`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":    agencyID,
		"user_message": userMessage,
		"goal_keys":    req.GoalKeys,
	}).Info("Processing streaming chat-based goal request")

	// Use shared streaming helper
	h.ExecuteChatStreaming(c, ChatStreamingOptions{
		AgencyID:    agencyID,
		UserMessage: userMessage,

		// Process function: calls RefineGoals and captures response
		ProcessFunc: func(c *gin.Context) (interface{}, error) {
			// Store original writer to capture response
			originalWriter := c.Writer
			captureBuffer := &bytes.Buffer{}
			c.Writer = &responseCapture{
				ResponseWriter: c.Writer,
				body:           captureBuffer,
			}

			// Call RefineGoals (it will write JSON response)
			h.RefineGoals(c)

			// Restore original writer
			c.Writer = originalWriter

			// Check for errors
			if c.Writer.Status() >= 400 {
				return nil, fmt.Errorf("goals processing failed")
			}

			// Parse the captured JSON response
			var goalsResp GoalsResponse
			if err := json.Unmarshal(captureBuffer.Bytes(), &goalsResp); err != nil {
				return nil, fmt.Errorf("failed to parse goals response: %w", err)
			}

			return &goalsResp, nil
		},

		// Format function: formats GoalsResponse for chat display
		FormatFunc: func(response interface{}) string {
			goalsResp := response.(*GoalsResponse)
			return formatGoalsChatMessage(*goalsResp)
		},

		// WasChanged function: determines if goals were modified
		WasChangedFunc: func(response interface{}) bool {
			goalsResp := response.(*GoalsResponse)
			return !goalsResp.NoActionNeeded
		},
	})
}
