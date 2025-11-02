package ai_refine

import (
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RefineGoal handles POST /api/v1/agencies/:id/goals/:goalKey/refine
// Refines a goal definition using AI with full context
func (h *Handler) RefineGoal(c *gin.Context) {
	agencyID := c.Param("id")
	goalKey := c.Param("goalKey")

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"goal_key":  goalKey,
	}).Info("Processing AI goal refinement request")

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

	// Get current goal
	currentGoal, err := h.agencyService.GetGoal(c.Request.Context(), agencyID, goalKey)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch goal")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusNotFound, `
			<div class="notification is-warning">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-warning mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>Goal Not Found</strong>
						<p class="mb-0">The requested goal could not be found.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Get current form data (user might have edited it)
	currentDescription := c.PostForm("goal-editor")
	if currentDescription == "" {
		currentDescription = currentGoal.Description
	}

	currentScope := c.PostForm("scope-editor")
	if currentScope == "" {
		currentScope = currentGoal.Scope
	}

	// Parse success metrics (could be multiline)
	metricsText := c.PostForm("metrics-editor")
	var currentMetrics []string
	if metricsText != "" {
		currentMetrics = strings.Split(strings.TrimSpace(metricsText), "\n")
		// Clean up empty lines
		var cleanMetrics []string
		for _, metric := range currentMetrics {
			if strings.TrimSpace(metric) != "" {
				cleanMetrics = append(cleanMetrics, strings.TrimSpace(metric))
			}
		}
		currentMetrics = cleanMetrics
	} else {
		currentMetrics = currentGoal.SuccessMetrics
	}

	// Get all existing goals for context
	existingGoals, err := h.agencyService.GetGoals(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch existing goals, continuing without them")
		existingGoals = []*agency.Goal{}
	}

	// Get all units of work for context
	unitsOfWork, err := h.agencyService.GetUnitsOfWork(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch units of work, continuing without them")
		unitsOfWork = []*agency.UnitOfWork{}
	}

	// Build refinement request
	refineReq := &ai.RefineGoalRequest{
		AgencyID:       agencyID,
		CurrentGoal:    currentGoal,
		Description:    currentDescription,
		Scope:          currentScope,
		SuccessMetrics: currentMetrics,
		ExistingGoals:  existingGoals,
		UnitsOfWork:    unitsOfWork,
		AgencyContext:  ag,
	}

	// Call AI refinement service
	result, err := h.goalRefiner.RefineGoal(c.Request.Context(), refineReq)
	if err != nil {
		h.logger.WithError(err).Error("AI goal refinement failed")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>AI Refinement Failed</strong>
						<p class="mb-0">The AI service encountered an error. Please try again later.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":   agencyID,
		"goal_key":    goalKey,
		"was_changed": result.WasChanged,
		"description": len(result.RefinedDescription),
		"scope":       len(result.RefinedScope),
		"metrics":     len(result.RefinedMetrics),
	}).Info("AI goal refinement completed successfully")

	// Update the goal in database if it was changed
	if result.WasChanged {
		// Update the current goal with refined values
		currentGoal.Description = result.RefinedDescription
		currentGoal.Scope = result.RefinedScope
		currentGoal.SuccessMetrics = result.RefinedMetrics
		currentGoal.Priority = result.SuggestedPriority
		currentGoal.Category = result.SuggestedCategory
		currentGoal.Tags = result.SuggestedTags

		// Save to database
		err = h.agencyService.UpdateGoal(c.Request.Context(), agencyID, goalKey, currentGoal.Code, currentGoal.Description)
		if err != nil {
			h.logger.WithError(err).Error("Failed to save refined goal")
			// Continue to show the result even if saving failed
		}
	}

	// If no changes were made, return current form
	if !result.WasChanged {
		h.renderGoalForm(c, currentDescription, currentScope, currentMetrics, agencyID, goalKey,
			"No changes were needed. The goal definition is already well-structured!", false)
		return
	}

	// Render the refined goal form
	h.renderGoalForm(c, result.RefinedDescription, result.RefinedScope, result.RefinedMetrics, agencyID, goalKey,
		result.Explanation, true)
}

func (h *Handler) renderGoalForm(c *gin.Context, description, scope string, metrics []string,
	agencyID, goalKey, message string, isSuccess bool) {

	notificationClass := "is-info"
	notificationIcon := "fa-info-circle"
	notificationIconClass := "has-text-info"
	notificationTitle := ""

	if isSuccess {
		notificationClass = "is-success"
		notificationIcon = "fa-check-circle"
		notificationIconClass = "has-text-success"
		notificationTitle = "<strong>AI Refinement Complete!</strong><br>"
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `
		<div class="content">
			<div class="field">
				<label class="label">Goal Description</label>
				<div class="control">
					<textarea
						class="textarea"
						id="goal-editor"
						placeholder="Describe the goal this agency aims to solve..."
						rows="15"
						style="font-family: monospace; font-size: 14px;">%s</textarea>
				</div>
			</div>
			
			<div class="field">
				<label class="label">Scope</label>
				<div class="control">
					<textarea
						class="textarea"
						id="scope-editor"
						placeholder="Define the scope and boundaries of this goal..."
						rows="8">%s</textarea>
				</div>
			</div>
			
			<div class="field">
				<label class="label">Success Metrics</label>
				<div class="control">
					<textarea
						class="textarea"
						id="metrics-editor"
						placeholder="Define how success will be measured..."
						rows="8">%s</textarea>
				</div>
			</div>
		</div>
		
		<div class="buttons is-right">
			<button
				class="button is-primary"
				onclick="saveGoalDefinition()"
				id="save-goal-btn">
				<span class="icon"><i class="fas fa-save"></i></span>
				<span>Save</span>
			</button>
			
			<button
				class="button is-info"
				hx-post="/api/v1/agencies/%s/goals/%s/refine"
				hx-include="#goal-editor, #scope-editor, #metrics-editor"
				hx-target="#goal-content"
				hx-indicator="#ai-process-status"
				hx-on::after-request="
					console.log('🏁 HTMX request completed, hiding status...');
					if (window.hideAIProcessStatus) {
						window.hideAIProcessStatus();
					} else {
						console.log('❌ hideAIProcessStatus not available');
						const status = document.getElementById('ai-process-status');
						if (status) {
							status.style.display = 'none';
							console.log('✅ Status hidden manually');
						}
					}
				"
				id="ai-sparkle-btn"
				onclick="window.handleRefineClick && window.handleRefineClick()"
				title="Refine with AI">
				<span class="icon"><i class="fas fa-magic"></i></span>
				<span>Refine</span>
			</button>
			
			<button
				class="button"
				onclick="undoGoalDefinition()"
				id="undo-goal-btn">
				<span class="icon"><i class="fas fa-undo"></i></span>
				<span>Undo</span>
			</button>
		</div>
		
		<div class="notification %s is-light mt-3">
			<div class="is-flex is-align-items-center">
				<span class="icon %s mr-2">
					<i class="fas %s"></i>
				</span>
				<div>
					%s<p class="mb-0">%s</p>
				</div>
			</div>
		</div>
	`,
		description,
		scope,
		strings.Join(metrics, "\n"),
		agencyID,
		goalKey,
		notificationClass,
		notificationIconClass,
		notificationIcon,
		notificationTitle,
		message)
}
