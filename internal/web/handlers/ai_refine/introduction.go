package ai_refine

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
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

	// Get current overview/introduction
	overview, err := h.agencyService.GetAgencyOverview(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch overview")
		// Create empty overview if not found
		overview = &agency.Overview{
			AgencyID:     agencyID,
			Introduction: "",
		}
	}

	// Get current introduction text from form (user might have edited it)
	currentIntroduction := c.PostForm("introduction-editor")
	if currentIntroduction == "" {
		// Fallback to stored introduction if form is empty
		currentIntroduction = overview.Introduction
	}

	// Get all goals for context
	goals, err := h.agencyService.GetGoals(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch goals, continuing without them")
		goals = []*agency.Goal{}
	}

	// Get all units of work for context
	workItems, err := h.agencyService.GetWorkItems(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch units of work, continuing without them")
		workItems = []*agency.WorkItem{}
	}

	// Build refinement request
	refineReq := &ai.RefineIntroductionRequest{
		AgencyID:      agencyID,
		CurrentIntro:  currentIntroduction,
		Goals:         goals,
		WorkItems:     workItems,
		AgencyContext: ag,
	}

	// Call AI refiner service
	refinedResult, err := h.introductionRefiner.RefineIntroduction(c.Request.Context(), refineReq)
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
		"agency_id":   agencyID,
		"was_changed": refinedResult.WasChanged,
		"explanation": refinedResult.Explanation,
	}).Info("AI refinement completed")

	// Determine what text to save
	// If AI refined it, use the refined version
	// If AI didn't change it, still use the refined version (which is the current intro)
	// This ensures user edits are saved even if AI says "no changes needed"
	introToSave := refinedResult.RefinedIntroduction

	// Check if the introduction is different from what's in the database
	needsSave := (introToSave != overview.Introduction)

	if needsSave {
		h.logger.WithFields(logrus.Fields{
			"agency_id":           agencyID,
			"ai_changed":          refinedResult.WasChanged,
			"intro_length":        len(introToSave),
			"stored_intro_length": len(overview.Introduction),
		}).Info("Introduction differs from database, saving")

		err = h.agencyService.UpdateAgencyOverview(c.Request.Context(), agencyID, introToSave)
		if err != nil {
			h.logger.WithError(err).Error("Failed to save introduction")
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

		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
		}).Info("Successfully saved introduction to database")
	} else {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
		}).Info("Introduction matches database, no save needed")
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

	// Update overview object for template rendering
	overview.Introduction = refinedResult.RefinedIntroduction

	// Render the refined introduction response
	component := agency_designer.AIRefineResponse(refinedResult, ag, overview)
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
