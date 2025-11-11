package ai_refine

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
)

// RefineIntroduction handles POST /api/v1/agencies/:id/overview/refine
// Supports both streaming (via ?stream=true) and standard modes
func (h *Handler) RefineIntroduction(c *gin.Context) {
	agencyID := c.Param("id")
	streamMode := c.Query("stream") == "true"

	if streamMode {
		h.refineIntroductionStreaming(c, agencyID)
	} else {
		h.refineIntroductionStandard(c, agencyID)
	}
}

// refineIntroductionStandard handles non-streaming introduction refinement
func (h *Handler) refineIntroductionStandard(c *gin.Context, agencyID string) {
	h.logger.WithField("agency_id", agencyID).Info("Processing AI introduction refinement")

	// Fetch agency and specification
	ag, spec, err := h.fetchAgencyAndSpec(c, agencyID)
	if err != nil {
		return // Error already handled
	}

	// Get current introduction and user request
	currentIntroduction := h.getCurrentIntroduction(c, spec)
	userRequest := h.getUserRequest(c)

	// Build AI context
	builderContextData, err := h.buildAIContext(c, ag, currentIntroduction, userRequest)
	if err != nil {
		return // Error already handled
	}

	// Get conversation history
	conversationHistory := h.getConversationHistory(agencyID)

	// Perform AI refinement
	refinedResult, err := h.introductionRefiner.RefineIntroduction(
		c.Request.Context(),
		&builder.RefineIntroductionRequest{
			AgencyID:            agencyID,
			ConversationHistory: conversationHistory,
		},
		builderContextData,
	)
	if err != nil {
		h.sendError(c, false, "AI Refinement Failed", "Please check your AI configuration and try again.")
		return
	}

	// Save refined introduction
	if refinedResult.Data != nil && refinedResult.Data.Introduction != "" {
		_, err = h.agencyService.UpdateIntroduction(c.Request.Context(), agencyID, refinedResult.Data.Introduction, "ai-refine")
		if err != nil {
			h.logger.WithError(err).Error("Failed to save refined introduction")
		}
	}

	// Add to conversation
	h.addToConversation(agencyID, refinedResult)

	// Render response
	h.renderStandardResponse(c, agencyID, refinedResult, ag, userRequest)
}

// refineIntroductionStreaming handles streaming introduction refinement via SSE
func (h *Handler) refineIntroductionStreaming(c *gin.Context, agencyID string) {
	h.logger.WithField("agency_id", agencyID).Info("🌊 Processing streaming AI introduction refinement")

	// Fetch agency and specification
	ag, spec, err := h.fetchAgencyAndSpec(c, agencyID)
	if err != nil {
		c.SSEvent("error", `{"error": "Agency not found"}`)
		return
	}

	// Get current introduction and user request
	currentIntroduction := h.getCurrentIntroduction(c, spec)
	userRequest := h.getUserRequest(c)

	// Build AI context
	builderContextData, err := h.buildAIContext(c, ag, currentIntroduction, userRequest)
	if err != nil {
		c.SSEvent("error", `{"error": "Failed to build context"}`)
		return
	}

	// Setup SSE
	h.setupSSE(c)

	// Stream refinement
	chunkCount := 0
	result, err := h.introductionRefiner.RefineIntroductionStream(
		c.Request.Context(),
		&builder.RefineIntroductionRequest{AgencyID: agencyID},
		builderContextData,
		func(chunk string) error {
			chunkCount++
			h.logger.WithField("chunk_num", chunkCount).Debug("📦 Sending chunk")
			c.SSEvent("chunk", chunk)
			c.Writer.Flush()
			return nil
		},
	)

	if err != nil {
		h.logger.WithError(err).Error("❌ Streaming refinement failed")
		c.SSEvent("error", fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}

	h.logger.WithField("total_chunks", chunkCount).Info("✅ Streaming completed")

	// Save if changed
	if result.Data != nil && result.Data.Introduction != "" && result.Data.Introduction != spec.Introduction {
		_, err = h.agencyService.UpdateIntroduction(c.Request.Context(), agencyID, result.Data.Introduction, "ai-refine-stream")
		if err != nil {
			h.logger.WithError(err).Error("Failed to save")
			c.SSEvent("error", `{"error": "Failed to save changes"}`)
			return
		}
	}

	// Send completion
	completionData := map[string]interface{}{
		"was_changed":      result.WasChanged,
		"explanation":      result.Explanation,
		"changed_sections": result.ChangedSections,
	}
	if result.Data != nil {
		completionData["introduction"] = result.Data.Introduction
	}

	c.SSEvent("complete", completionData)
	c.Writer.Flush()
}

// Helper functions

func (h *Handler) fetchAgencyAndSpec(c *gin.Context, agencyID string) (*models.Agency, *models.AgencySpecification, error) {
	ag, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch agency")
		h.sendError(c, false, "Agency Not Found", "The requested agency could not be found.")
		return nil, nil, err
	}

	spec, err := h.agencyService.GetSpecification(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch specification")
		spec = &models.AgencySpecification{Introduction: ""}
	}

	return ag, spec, nil
}

func (h *Handler) getCurrentIntroduction(c *gin.Context, spec *models.AgencySpecification) string {
	currentIntroduction := c.PostForm("introduction-editor")
	if currentIntroduction == "" {
		h.logger.Warn("⚠️ Form empty, using database value")
		currentIntroduction = spec.Introduction
	}
	return currentIntroduction
}

func (h *Handler) getUserRequest(c *gin.Context) string {
	userRequest := c.PostForm("user-request")
	if userRequest == "" {
		userRequest = c.GetHeader("X-User-Request")
	}
	return userRequest
}

func (h *Handler) buildAIContext(c *gin.Context, ag *models.Agency, currentIntro, userRequest string) (builder.BuilderContext, error) {
	builderContextData, err := h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, currentIntro, userRequest)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build AI context")
		h.sendError(c, false, "Context Build Failed", "Failed to gather necessary context data.")
		return builder.BuilderContext{}, err
	}
	return builderContextData, nil
}

func (h *Handler) getConversationHistory(agencyID string) []ai.Message {
	conv, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err == nil && conv != nil {
		messageCount := len(conv.Messages)
		startIdx := 0
		if messageCount > 5 {
			startIdx = messageCount - 5
		}
		return conv.Messages[startIdx:]
	}
	return nil
}

func (h *Handler) addToConversation(agencyID string, result *builder.RefineIntroductionResponse) {
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		ctx := context.Background()
		conversation, err = h.designerService.StartConversation(ctx, agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation")
			return
		}
	}

	if conversation != nil {
		chatMessage := result.Explanation
		if result.WasChanged {
			chatMessage = "✨ **Introduction Refined & Saved**\n\n" + chatMessage
		} else {
			chatMessage = "✅ **Introduction Review Complete**\n\n" + chatMessage
		}
		if err := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); err != nil {
			h.logger.WithError(err).Error("Failed to add message to conversation")
		}
	}
}

func (h *Handler) renderStandardResponse(c *gin.Context, agencyID string, result *builder.RefineIntroductionResponse, ag *models.Agency, userRequest string) {
	conversationID := c.Param("conversationId")
	isFromChat := conversationID != "" || userRequest != ""

	if isFromChat {
		h.renderChatResponse(c, agencyID, result, ag)
	} else {
		h.renderDirectResponse(c, result, ag)
	}
}

func (h *Handler) renderChatResponse(c *gin.Context, agencyID string, result *builder.RefineIntroductionResponse, ag *models.Agency) {
	conversation, _ := h.designerService.GetConversationByAgencyID(agencyID)
	var chatHTML string

	if conversation != nil && len(conversation.Messages) > 0 {
		lastMessage := conversation.Messages[len(conversation.Messages)-1]
		if lastMessage.Role == "assistant" {
			var chatBuf strings.Builder
			component := agency_designer.AIMessage(lastMessage)
			if err := component.Render(c.Request.Context(), &chatBuf); err == nil {
				chatHTML = chatBuf.String()
			}
		}
	}

	// Render introduction editor for OOB swap
	var introBuf strings.Builder
	introComponent := agency_designer.AIRefineResponse(result, ag)
	if err := introComponent.Render(c.Request.Context(), &introBuf); err != nil {
		h.logger.WithError(err).Error("Failed to render introduction editor")
		c.String(http.StatusInternalServerError, "Failed to render introduction")
		return
	}

	introOOB := fmt.Sprintf(`<div id="introduction-content" hx-swap-oob="true">%s</div>`, introBuf.String())
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, chatHTML+introOOB)
}

func (h *Handler) renderDirectResponse(c *gin.Context, result *builder.RefineIntroductionResponse, ag *models.Agency) {
	component := agency_designer.AIRefineResponse(result, ag)
	c.Header("Content-Type", "text/html")
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render AI refine response")
		c.String(http.StatusInternalServerError, "Render error")
	}
}

func (h *Handler) setupSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.Flush()
	c.SSEvent("start", `{"status": "streaming"}`)
	c.Writer.Flush()
}

func (h *Handler) sendError(c *gin.Context, isStreaming bool, title, message string) {
	if isStreaming {
		c.SSEvent("error", fmt.Sprintf(`{"error": "%s"}`, message))
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusInternalServerError, fmt.Sprintf(`
		<div class="notification is-danger">
			<div class="is-flex is-align-items-center">
				<span class="icon has-text-danger mr-2">
					<i class="fas fa-exclamation-triangle"></i>
				</span>
				<div>
					<strong>%s</strong>
					<p class="mb-0">%s</p>
				</div>
			</div>
		</div>
	`, title, message))
}
