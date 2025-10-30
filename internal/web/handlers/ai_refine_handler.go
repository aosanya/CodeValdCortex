package handlers

import (
	"fmt"
	"net/http"
	"strings"

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
	goalRefiner         *ai.GoalRefiner
	designerService     *ai.AgencyDesignerService
	logger              *logrus.Logger
}

// NewAIRefineHandler creates a new AI refine handler
func NewAIRefineHandler(
	agencyService agency.Service,
	introductionRefiner *ai.IntroductionRefiner,
	goalRefiner *ai.GoalRefiner,
	designerService *ai.AgencyDesignerService,
	logger *logrus.Logger,
) *AIRefineHandler {
	return &AIRefineHandler{
		agencyService:       agencyService,
		introductionRefiner: introductionRefiner,
		goalRefiner:         goalRefiner,
		designerService:     designerService,
		logger:              logger,
	}
}

// getMapKeys returns the keys of a map as a slice for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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

	// Get all goals for context
	goals, err := h.agencyService.GetGoals(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch goals, continuing without them")
		goals = []*agency.Goal{}
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
		Goals:         goals,
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
			chatMessage = "✨ **Introduction Refined & Saved**\n\n" + chatMessage
		} else {
			chatMessage = "✅ **Introduction Review Complete**\n\n" + chatMessage
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

// RefineGoal handles POST /api/v1/agencies/:id/goals/:goalKey/refine
// Refines a goal definition using AI with full context
func (h *AIRefineHandler) RefineGoal(c *gin.Context) {
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

		// Save to database (we need to add an UpdateGoalFull method or use the existing update)
		// For now, we'll just use the basic update which handles code and description
		// TODO: Extend UpdateGoal to handle all fields or add UpdateGoalFull method
		err = h.agencyService.UpdateGoal(c.Request.Context(), agencyID, goalKey, currentGoal.Code, currentGoal.Description)
		if err != nil {
			h.logger.WithError(err).Error("Failed to save refined goal")
			// Continue to show the result even if saving failed
		}
	}

	// If no changes were made, return current form
	if !result.WasChanged {
		h.logger.WithField("agency_id", agencyID).Info("No changes needed, returning current form")
		// Return the current form with a "no changes" message
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
			
			<div class="notification is-info is-light mt-3">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-info mr-2">
						<i class="fas fa-info-circle"></i>
					</span>
					<span>No changes were needed. The goal definition is already well-structured!</span>
				</div>
			</div>
		`,
			currentDescription,
			currentScope,
			strings.Join(currentMetrics, "\n"),
			agencyID,
			goalKey)
		return
	}

	// Render the refined goal form
	h.logger.WithField("agency_id", agencyID).Info("Rendering refined goal form")

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
		
		<div class="notification is-success is-light mt-3">
			<div class="is-flex is-align-items-center">
				<span class="icon has-text-success mr-2">
					<i class="fas fa-check-circle"></i>
				</span>
				<div>
					<strong>AI Refinement Complete!</strong>
					<p class="mb-0">%s</p>
				</div>
			</div>
		</div>
	`,
		result.RefinedDescription,
		result.RefinedScope,
		strings.Join(result.RefinedMetrics, "\n"),
		agencyID,
		goalKey,
		result.Explanation)
}

// GenerateGoal handles POST /api/v1/agencies/:id/goals/generate
// Generates a new goal using AI based on user input
func (h *AIRefineHandler) GenerateGoal(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse request body to get user input
	var req struct {
		UserInput string `json:"userInput" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse generate goal request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate agency exists and get context
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		h.logger.Error("Agency not found", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	// Get existing goals for context
	existingGoals, err := h.agencyService.GetGoals(ctx, agencyID)
	if err != nil {
		h.logger.Error("Failed to get existing goals", "agencyID", agencyID, "error", err)
		existingGoals = []*agency.Goal{} // Continue with empty list
	}

	// Get units of work for context
	unitsOfWork, err := h.agencyService.GetUnitsOfWork(ctx, agencyID)
	if err != nil {
		h.logger.Error("Failed to get units of work", "agencyID", agencyID, "error", err)
		unitsOfWork = []*agency.UnitOfWork{} // Continue with empty list
	}

	// Build generation request
	genReq := &ai.GenerateGoalRequest{
		AgencyID:      agencyID,
		AgencyContext: ag,
		ExistingGoals: existingGoals,
		UnitsOfWork:   unitsOfWork,
		UserInput:     req.UserInput,
	}

	// Generate goal using AI
	result, err := h.goalRefiner.GenerateGoal(ctx, genReq)
	if err != nil {
		h.logger.Error("Failed to generate goal", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate goal"})
		return
	}

	// Save the generated goal to database using the service method signature
	// CreateGoal(ctx context.Context, agencyID string, code string, description string)
	createdGoal, err := h.agencyService.CreateGoal(ctx, agencyID, result.SuggestedCode, result.Description)
	if err != nil {
		h.logger.Error("Failed to save generated goal", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save generated goal"})
		return
	}

	// Update the created goal with additional fields from AI response
	// Note: We may need to add an UpdateGoal call here if we want to set Scope, Metrics, etc.

	h.logger.Info("Goal generated successfully",
		"agencyID", agencyID,
		"goalKey", createdGoal.Key,
		"code", createdGoal.Code)

	// Return the generated goal data as JSON for HTMX to handle
	c.JSON(http.StatusOK, gin.H{
		"goal":               createdGoal,
		"scope":              result.Scope,
		"success_metrics":    result.SuccessMetrics,
		"suggested_priority": result.SuggestedPriority,
		"suggested_category": result.SuggestedCategory,
		"suggested_tags":     result.SuggestedTags,
		"explanation":        result.Explanation,
		"success":            true,
	})
}

// ProcessAIGoalRequest handles POST /api/v1/agencies/:id/goals/ai-process
// Processes multiple AI operations on goals (create, enhance, consolidate)
func (h *AIRefineHandler) ProcessAIGoalRequest(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse request body
	var req struct {
		Operations []string `json:"operations" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse AI process request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":  agencyID,
		"operations": req.Operations,
	}).Info("Processing AI goal operations")

	// Validate agency exists and get context
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		h.logger.Error("Agency not found", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	// Get introduction for context
	overview, err := h.agencyService.GetAgencyOverview(ctx, agencyID)
	if err != nil {
		h.logger.Warn("Failed to get overview", "agencyID", agencyID, "error", err)
		overview = &agency.Overview{AgencyID: agencyID}
	}

	// Get existing goals
	existingGoals, err := h.agencyService.GetGoals(ctx, agencyID)
	if err != nil {
		h.logger.Error("Failed to get existing goals", "agencyID", agencyID, "error", err)
		existingGoals = []*agency.Goal{}
	}

	// Get units of work for context
	unitsOfWork, err := h.agencyService.GetUnitsOfWork(ctx, agencyID)
	if err != nil {
		h.logger.Warn("Failed to get units of work", "agencyID", agencyID, "error", err)
		unitsOfWork = []*agency.UnitOfWork{}
	}

	results := make(map[string]interface{})
	var createdGoals []*agency.Goal
	var enhancedGoals []*agency.Goal
	var consolidationSuggestions []string

	// Process each operation
	for _, operation := range req.Operations {
		h.logger.Info("Processing operation", "operation", operation, "agencyID", agencyID)

		switch operation {
		case "create":
			// Generate new goals based on introduction
			if overview.Introduction == "" {
				h.logger.Warn("No introduction found for goal generation", "agencyID", agencyID)
				results["create_error"] = "No introduction found. Please add an introduction first."
				continue
			}

			h.logger.Info("Starting multiple goal generation from introduction",
				"agencyID", agencyID,
				"introductionLength", len(overview.Introduction),
				"existingGoalsCount", len(existingGoals))

			// Generate multiple goals in one AI call
			genReq := &ai.GenerateGoalRequest{
				AgencyID:      agencyID,
				AgencyContext: ag,
				ExistingGoals: existingGoals,
				UnitsOfWork:   unitsOfWork,
				UserInput:     "Based on the agency introduction: " + overview.Introduction,
			}

			h.logger.Info("Calling AI to generate multiple goals", "agencyID", agencyID)

			result, err := h.goalRefiner.GenerateGoals(ctx, genReq)
			if err != nil {
				h.logger.Error("Failed to generate goals from AI", "agencyID", agencyID, "error", err)
				results["create_error"] = err.Error()
				continue
			}

			h.logger.Info("AI generated goals successfully",
				"agencyID", agencyID,
				"goalsCount", len(result.Goals),
				"explanation", result.Explanation)

			// Save each generated goal to database
			for i, goalData := range result.Goals {
				h.logger.Info("Saving generated goal to database",
					"agencyID", agencyID,
					"goalIndex", i+1,
					"goalCode", goalData.SuggestedCode,
					"descriptionLength", len(goalData.Description))

				goal, err := h.agencyService.CreateGoal(ctx, agencyID, goalData.SuggestedCode, goalData.Description)
				if err != nil {
					h.logger.Error("Failed to save generated goal",
						"agencyID", agencyID,
						"goalIndex", i+1,
						"goalCode", goalData.SuggestedCode,
						"error", err)
					// Continue with other goals even if one fails
					continue
				}

				h.logger.Info("Goal saved successfully",
					"agencyID", agencyID,
					"goalKey", goal.Key,
					"goalCode", goal.Code,
					"goalNumber", goal.Number)

				createdGoals = append(createdGoals, goal)
			}

			h.logger.Info("Completed creating multiple goals",
				"agencyID", agencyID,
				"totalCreated", len(createdGoals),
				"requested", len(result.Goals))

			results["create_success"] = fmt.Sprintf("Created %d goals", len(createdGoals))
			results["ai_explanation"] = result.Explanation

		case "enhance":
			// TODO: Implement goal enhancement logic
			// For now, return placeholder
			results["enhance_message"] = "Enhancement feature coming soon"

		case "consolidate":
			// TODO: Implement goal consolidation logic
			// For now, return placeholder
			results["consolidate_message"] = "Consolidation feature coming soon"
		}
	}

	// Add AI explanation to chat conversation if goals were created
	if len(createdGoals) > 0 {
		h.logger.Info("Attempting to add AI explanation to chat",
			"agencyID", agencyID,
			"createdGoalsCount", len(createdGoals),
			"resultsKeys", fmt.Sprintf("%v", getMapKeys(results)))

		explanation, hasExplanation := results["ai_explanation"].(string)
		h.logger.Info("Explanation check",
			"agencyID", agencyID,
			"hasExplanation", hasExplanation,
			"explanationLength", len(explanation))

		if hasExplanation && explanation != "" {
			// Get or create conversation
			conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
			if err != nil {
				h.logger.Warn("No conversation exists, creating new one",
					"agencyID", agencyID,
					"error", err)
				// No conversation exists, create one
				conversation, err = h.designerService.StartConversation(ctx, agencyID)
				if err != nil {
					h.logger.WithError(err).Error("Failed to create conversation for AI goal generation message")
				}
			}

			if conversation != nil {
				chatMessage := fmt.Sprintf("✨ **Created %d Goals**\n\n%s", len(createdGoals), explanation)
				h.logger.Info("Adding message to chat",
					"agencyID", agencyID,
					"conversationID", conversation.ID,
					"messageLength", len(chatMessage))

				if addErr := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); addErr != nil {
					h.logger.WithError(addErr).Error("Failed to add goal generation explanation to chat")
				} else {
					h.logger.Info("Successfully added AI explanation to chat",
						"agencyID", agencyID,
						"conversationID", conversation.ID)
				}
			} else {
				h.logger.Error("Conversation is nil after creation attempt", "agencyID", agencyID)
			}
		} else {
			h.logger.Warn("No explanation to add to chat",
				"agencyID", agencyID,
				"hasExplanation", hasExplanation,
				"explanation", explanation)
		}
	} else {
		h.logger.Info("No goals were created, skipping chat message",
			"agencyID", agencyID)
	}

	// Build response
	response := gin.H{
		"success": true,
		"results": results,
	}

	if len(createdGoals) > 0 {
		response["created_goals"] = createdGoals
		response["created_count"] = len(createdGoals)
	}

	if len(enhancedGoals) > 0 {
		response["enhanced_goals"] = enhancedGoals
		response["enhanced_count"] = len(enhancedGoals)
	}

	if len(consolidationSuggestions) > 0 {
		response["consolidation_suggestions"] = consolidationSuggestions
	}

	h.logger.Info("AI goal operations completed",
		"agencyID", agencyID,
		"created", len(createdGoals),
		"enhanced", len(enhancedGoals))

	c.JSON(http.StatusOK, response)
}
