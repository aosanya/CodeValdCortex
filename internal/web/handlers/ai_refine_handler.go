package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AIRefineHandler handles AI refinement requests for agency components
type AIRefineHandler struct {
	agencyService       agency.Service
	introductionRefiner *ai.IntroductionRefiner
	designerService     *ai.AgencyDesignerService
	logger              *logrus.Logger
}

// NewAIRefineHandler creates a new AI refine handler
func NewAIRefineHandler(
	agencyService agency.Service,
	introductionRefiner *ai.IntroductionRefiner,
	designerService *ai.AgencyDesignerService,
	logger *logrus.Logger,
) *AIRefineHandler {
	return &AIRefineHandler{
		agencyService:       agencyService,
		introductionRefiner: introductionRefiner,
		designerService:     designerService,
		logger:              logger,
	}
}

// RefineIntroduction handles POST /api/v1/agencies/:id/overview/refine
// Refines the agency introduction using AI with full context
func (h *AIRefineHandler) RefineIntroduction(c *gin.Context) {
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

	// Get all problems for context
	problems, err := h.agencyService.GetProblems(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch problems, continuing without them")
		problems = []*agency.Problem{}
	}

	// Get all units of work for context
	unitsOfWork, err := h.agencyService.GetUnitsOfWork(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch units of work, continuing without them")
		unitsOfWork = []*agency.UnitOfWork{}
	}

	// Build refinement request
	refineReq := &ai.RefineIntroductionRequest{
		AgencyID:      agencyID,
		CurrentIntro:  currentIntroduction,
		Problems:      problems,
		UnitsOfWork:   unitsOfWork,
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

	// Update the overview with refined introduction if it was changed
	if refinedResult.WasChanged {
		err = h.agencyService.UpdateAgencyOverview(c.Request.Context(), agencyID, refinedResult.RefinedIntroduction)
		if err != nil {
			h.logger.WithError(err).Error("Failed to save refined introduction")
			// Continue to show the result even if saving failed
		}
	}

	// Add the AI refinement explanation to the chat conversation
	// Create conversation if it doesn't exist
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		// No conversation exists, create one
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to create conversation for AI refinement message")
		}
	}

	if conversation != nil {
		chatMessage := refinedResult.Explanation
		if refinedResult.WasChanged {
			chatMessage = "✨ **Introduction Refined & Saved**\n\n" + refinedResult.Explanation
		} else {
			chatMessage = "✅ **Introduction Review Complete**\n\n" + refinedResult.Explanation
		}

		if addErr := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); addErr != nil {
			h.logger.WithError(addErr).Warn("Failed to add refinement explanation to chat")
		}
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
