package ai_refine

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RefineIntroduction handles POST /api/v1/agencies/:id/overview/refine
// Refines the agency introduction using AI with full context
func (h *Handler) RefineIntroduction(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithField("agency_id", agencyID).Info("Processing AI introduction refinement request")

	// Get agency context
	ag, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch agency")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusNotFound, `
			<div class="notification is-warning">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-warning mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>Agency Not Found</strong>
						<p class="mb-0">The requested agency could not be found.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Get current specification/introduction for reference
	spec, err := h.agencyService.GetSpecification(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch specification")
		// Create empty spec if not found
		spec = &models.AgencySpecification{
			Introduction: "",
		}
	}

	// ALWAYS use current introduction text from form (the textarea value)
	// This ensures we refine what the user is currently editing, not what's in the database
	currentIntroduction := c.PostForm("introduction-editor")

	h.logger.WithFields(logrus.Fields{
		"agency_id":              agencyID,
		"form_intro_length":      len(currentIntroduction),
		"database_intro_length":  len(spec.Introduction),
		"form_intro_preview":     truncateString(currentIntroduction, 100),
		"database_intro_preview": truncateString(spec.Introduction, 100),
	}).Info("Refine introduction - using textarea value")

	if currentIntroduction == "" {
		h.logger.Warn("⚠️ Textarea is empty - using database value as fallback")
		currentIntroduction = spec.Introduction
	}

	// Check if there's a specific user request from the form
	userRequest := c.PostForm("user-request")
	if userRequest == "" {
		// Check if there's a pending request from chat (passed via header or session)
		userRequest = c.GetHeader("X-User-Request")
	}

	if userRequest != "" {
		h.logger.WithFields(logrus.Fields{
			"agency_id":    agencyID,
			"user_request": userRequest,
		}).Info("User provided specific refinement request")
	}

	// Build AI context data using shared context builder (pass the full agency object)
	builderContextData, err := h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, currentIntroduction, userRequest)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build AI context data")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>Context Build Failed</strong>
						<p class="mb-0">Failed to gather necessary context data.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Get conversation context for recent chat messages
	var conversationHistory []ai.Message
	conv, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err == nil && conv != nil {
		// Include recent conversation messages (last 5) for context
		messageCount := len(conv.Messages)
		startIdx := 0
		if messageCount > 5 {
			startIdx = messageCount - 5
		}
		conversationHistory = conv.Messages[startIdx:]

		h.logger.WithFields(logrus.Fields{
			"agency_id":     agencyID,
			"message_count": len(conversationHistory),
		}).Info("Including conversation context in introduction refinement")
	}

	// Build refinement request using the structured AI context data
	refineReq := &builder.RefineIntroductionRequest{
		AgencyID:            agencyID,
		ConversationHistory: conversationHistory,
	}

	// Call AI refiner service with builderContextData passed separately
	refinedResult, err := h.introductionRefiner.RefineIntroduction(c.Request.Context(), refineReq, builderContextData)
	if err != nil {
		h.logger.WithError(err).Error("AI refinement failed")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>AI Refinement Failed</strong>
						<p class="mb-0">Please check your AI configuration and try again.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":        agencyID,
		"was_changed":      refinedResult.WasChanged,
		"explanation":      refinedResult.Explanation,
		"changed_sections": refinedResult.ChangedSections,
		"data_nil":         refinedResult.Data == nil,
		"data_intro_len": func() int {
			if refinedResult.Data != nil {
				return len(refinedResult.Data.Introduction)
			} else {
				return 0
			}
		}(),
	}).Info("AI refinement completed")

	// Extract introduction from the refined data
	var introToSave string
	if refinedResult.Data != nil && refinedResult.Data.Introduction != "" {
		introToSave = refinedResult.Data.Introduction
		h.logger.Info("🔵 Using refined introduction from AI",
			"length", len(introToSave),
			"preview", truncateString(introToSave, 80))
	} else {
		// Fallback to current introduction if data is missing
		introToSave = currentIntroduction
		h.logger.Warn("⚠️ refinedResult.Data is nil or empty, using current introduction as fallback",
			"data_nil", refinedResult.Data == nil,
			"current_length", len(currentIntroduction))
	}

	// Check if the introduction is different from what's in the database
	needsSave := (introToSave != spec.Introduction)

	h.logger.Info("🔵 Checking if save is needed",
		"needs_save", needsSave,
		"intro_to_save_length", len(introToSave),
		"spec_intro_length", len(spec.Introduction),
		"are_equal", introToSave == spec.Introduction)

	if needsSave {
		h.logger.WithFields(logrus.Fields{
			"agency_id":           agencyID,
			"ai_changed":          refinedResult.WasChanged,
			"intro_length":        len(introToSave),
			"stored_intro_length": len(spec.Introduction),
		}).Info("Introduction differs from database, saving")

		_, err = h.agencyService.UpdateIntroduction(c.Request.Context(), agencyID, introToSave, "ai-refine")
		if err != nil {
			h.logger.WithError(err).Error("❌ Failed to save introduction")
			// Show error notification
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
				<div class="notification is-danger">
					<div class="is-flex is-align-items-center">
						<span class="icon has-text-danger mr-2">
							<i class="fas fa-exclamation-triangle"></i>
						</span>
						<div>
							<strong>Save Failed</strong>
							<p class="mb-0">The introduction could not be saved. Please try again.</p>
						</div>
					</div>
				</div>
			`)
			return
		}

		h.logger.Info("✅ Successfully saved introduction to database",
			"agency_id", agencyID,
			"saved_length", len(introToSave),
			"saved_preview", truncateString(introToSave, 80))
	} else {
		h.logger.Info("ℹ️ No save needed - introduction unchanged",
			"agency_id", agencyID,
			"intro_length", len(introToSave))
	}

	// Add the AI refinement explanation to the chat conversation
	h.logger.Info("Attempting to add introduction refinement to chat",
		"agencyID", agencyID,
		"wasChanged", refinedResult.WasChanged,
		"explanationLength", len(refinedResult.Explanation))

	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		h.logger.Warn("No conversation exists for introduction refine, creating new one",
			"agencyID", agencyID,
			"error", err)
		// No conversation exists, create one
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation for AI refinement message")
		}
	}

	if conversation != nil {
		chatMessage := refinedResult.Explanation
		if refinedResult.WasChanged {
			chatMessage = "✨ **Introduction Refined & Saved**\n\n" + chatMessage
		} else {
			chatMessage = "✅ **Introduction Review Complete**\n\n" + chatMessage
		}

		h.logger.Info("Adding introduction refinement message to chat",
			"agencyID", agencyID,
			"conversationID", conversation.ID,
			"messageLength", len(chatMessage),
			"wasChanged", refinedResult.WasChanged)

		if addErr := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); addErr != nil {
			h.logger.WithError(addErr).Error("Failed to add refinement explanation to chat")
		} else {
			h.logger.Info("Successfully added introduction refinement to chat",
				"agencyID", agencyID,
				"conversationID", conversation.ID)
		}
	} else {
		h.logger.Error("Conversation is nil after creation attempt for introduction refine",
			"agencyID", agencyID)
	}

	// CRITICAL: Don't use the old overview object - use refinedResult.Data.Introduction directly
	// This ensures we render the LATEST AI-refined content, not stale data
	h.logger.Info("🔵 Using AI refined introduction for rendering",
		"refined_intro_length", len(refinedResult.Data.Introduction),
		"refined_intro_preview", truncateString(refinedResult.Data.Introduction, 100))

	// CRITICAL: Verify refinedResult.Data before rendering
	if refinedResult.Data == nil {
		h.logger.Error("❌ CRITICAL: refinedResult.Data is nil - cannot render response")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>Data Error</strong>
						<p class="mb-0">AI response data is missing. Please try again.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Check if this request came from chat (has conversationId param or user-request)
	conversationID := c.Param("conversationId")
	isFromChat := conversationID != "" || userRequest != ""

	if isFromChat {
		h.logger.Info("🔵 Request is from chat - rendering AI message + OOB introduction editor",
			"conversationID", conversationID,
			"hasUserRequest", userRequest != "")

		// Render only the AI message that was just added (not all messages)
		// The user message was already added by JavaScript
		var chatHTML string
		if conversation != nil && len(conversation.Messages) > 0 {
			// Get the last message (should be the AI refinement message we just added)
			lastMessage := conversation.Messages[len(conversation.Messages)-1]
			if lastMessage.Role == "assistant" {
				component := agency_designer.AIMessage(lastMessage)
				c.Header("Content-Type", "text/html")

				// Render just the AI message to a buffer
				var chatBuf strings.Builder
				err = component.Render(c.Request.Context(), &chatBuf)
				if err != nil {
					h.logger.WithError(err).Error("Failed to render AI message")
					c.String(http.StatusInternalServerError, "Failed to render AI message")
					return
				}
				chatHTML = chatBuf.String()
			}
		}

		if chatHTML == "" {
			// Fallback: no message to render
			h.logger.Warn("No AI message found in conversation after introduction refine")
		}

		// Render introduction editor for out-of-band swap to #introduction-content
		var introBuf strings.Builder
		introComponent := agency_designer.AIRefineResponse(refinedResult, ag)
		err = introComponent.Render(c.Request.Context(), &introBuf)
		if err != nil {
			h.logger.WithError(err).Error("Failed to render introduction editor OOB")
			c.String(http.StatusInternalServerError, "Failed to render introduction")
			return
		}

		// Wrap introduction content with HTMX OOB swap attribute
		introOOB := fmt.Sprintf(`<div id="introduction-content" hx-swap-oob="true">%s</div>`, introBuf.String())

		// Send both: chat messages + OOB introduction editor
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, chatHTML+introOOB)
		return
	}

	// Simple: Just log and render introduction editor (for direct button clicks)
	h.logger.WithFields(logrus.Fields{
		"agency_id":     agencyID,
		"intro_length":  len(refinedResult.Data.Introduction),
		"intro_preview": truncateString(refinedResult.Data.Introduction, 100),
		"was_changed":   refinedResult.WasChanged,
	}).Info("🚀 Rendering AI refined introduction")

	// Render the refined introduction response - template uses refinedResult.Data.Introduction directly
	component := agency_designer.AIRefineResponse(refinedResult, ag)
	c.Header("Content-Type", "text/html")
	err = component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render AI refine response")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>Render Error</strong>
						<p class="mb-0">Failed to render the response. Please try again.</p>
					</div>
				</div>
			</div>
		`)
		return
	}
}

// truncateString safely truncates a string to a max length for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
