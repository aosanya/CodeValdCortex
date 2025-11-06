package ai_refine

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// formatExplanationAsBullets formats explanation text as bullet points
func (h *Handler) formatExplanationAsBullets(explanation string) string {
	// Split by common sentence delimiters or line breaks
	sentences := strings.Split(explanation, ". ")

	var bullets []string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Add period back if it was removed by split
		if !strings.HasSuffix(sentence, ".") && !strings.HasSuffix(sentence, "!") && !strings.HasSuffix(sentence, "?") {
			sentence += "."
		}

		// Format as bullet point
		bullets = append(bullets, "• "+sentence)
	}

	return strings.Join(bullets, "\n")
}

// ProcessGoalChatRequest handles chat-based goal interactions
// This is similar to RefineIntroduction but for goals in chat context
func (h *Handler) ProcessGoalChatRequest(c *gin.Context) {
	h.logger.Info("🔵 HANDLER CALLED: ProcessGoalChatRequest")

	agencyID := c.Param("id")

	// Get user's chat message/request
	userRequest := c.PostForm("user-request")
	if userRequest == "" {
		userRequest = c.PostForm("message")
	}

	if userRequest == "" {
		h.logger.Error("No user request provided for goal chat")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, `
			<div class="notification is-warning">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-warning mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>No Request Provided</strong>
						<p class="mb-0">Please provide a message or request.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":    agencyID,
		"user_request": userRequest,
	}).Info("Processing chat-based goal request")

	// Get agency context
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
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

	// Get or create conversation
	conv, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		conv, err = h.designerService.StartConversation(ctx, agencyID)
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

	// Add user message to conversation
	if err := h.designerService.AddMessage(conv.ID, "user", userRequest); err != nil {
		h.logger.WithError(err).Error("Failed to add user message to conversation")
	}

	// Build AI context data using shared context builder
	builderContext, err := h.contextBuilder.BuildBuilderContext(ctx, ag, "", userRequest)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build context")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>Context Error</strong>
						<p class="mb-0">Failed to gather agency context for AI processing.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Get existing goals for context
	existingGoals := builderContext.Goals

	// Use the new RefineGoals method to dynamically determine and execute the appropriate action
	refineReq := &builder.RefineGoalsRequest{
		AgencyID:      agencyID,
		UserMessage:   userRequest,
		TargetGoals:   nil, // Will analyze all goals
		ExistingGoals: existingGoals,
		WorkItems:     builderContext.WorkItems,
		AgencyContext: ag,
	}

	result, err := h.goalRefiner.RefineGoals(ctx, refineReq, builderContext)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process goal request dynamically")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>AI Processing Failed</strong>
						<p class="mb-0">Failed to process your goal request.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Format the response based on the action taken
	var responseMessage string
	switch result.Action {
	case "refine", "enhance_all":
		if result.NoActionNeeded {
			responseMessage = "✅ **Goals Review Complete**\n\n" + result.Explanation
		} else {
			// Apply refinements to goals
			updatedCount := 0
			for _, rg := range result.RefinedGoals {
				if rg.WasChanged {
					// Find and update the goal
					for _, goal := range existingGoals {
						if goal.Key == rg.OriginalKey {
							// Use the suggested code if provided, otherwise keep the original
							goalCode := goal.Code
							if rg.SuggestedCode != "" {
								goalCode = rg.SuggestedCode
							}

							// Update the goal in the database
							updateErr := h.agencyService.UpdateGoal(ctx, agencyID, goal.Key, goalCode, rg.RefinedDescription)
							if updateErr != nil {
								h.logger.WithError(updateErr).Error("Failed to update refined goal", "goalKey", goal.Key)
							} else {
								updatedCount++
								h.logger.Info("Successfully updated refined goal", "goalKey", goal.Key, "newCode", goalCode)
							}
							break
						}
					}
				}
			}

			if updatedCount > 0 {
				responseMessage = fmt.Sprintf("✅ **Refined %d Goal(s)**\n\n%s", updatedCount, result.Explanation)
			} else {
				responseMessage = fmt.Sprintf("ℹ️ **Goals Analysis**\n\n%s", result.Explanation)
			}
		}

	case "generate":
		if len(result.GeneratedGoals) > 0 {
			createdCount := 0
			goalsList := []string{}

			// Create each generated goal
			for _, gGoal := range result.GeneratedGoals {
				createdGoal, createErr := h.agencyService.CreateGoal(ctx, agencyID, gGoal.SuggestedCode, gGoal.Description)
				if createErr != nil {
					h.logger.WithError(createErr).Error("Failed to create generated goal", "goalCode", gGoal.SuggestedCode)
				} else {
					createdCount++
					goalsList = append(goalsList, fmt.Sprintf("**%s**: %s", createdGoal.Code, createdGoal.Description))
					h.logger.Info("Successfully created generated goal", "goalKey", createdGoal.Key, "goalCode", createdGoal.Code)
				}
			}

			if createdCount > 0 {
				responseMessage = fmt.Sprintf("✨ **Generated %d New Goals**\n\n%s\n\n%s",
					createdCount,
					strings.Join(goalsList, "\n"),
					result.Explanation)
			} else {
				responseMessage = fmt.Sprintf("❌ **Failed to Create Goals**\n\n%s", result.Explanation)
			}
		} else {
			responseMessage = "ℹ️ " + result.Explanation
		}

	case "remove":
		// Handle goal removal - actually delete the goals from the database
		if result.ConsolidatedData != nil && len(result.ConsolidatedData.RemovedGoals) > 0 {
			deletedCount := 0
			deletedCodes := []string{}

			// Delete each goal
			for _, goalKey := range result.ConsolidatedData.RemovedGoals {
				// Find the goal to get its code for the response message
				for _, goal := range existingGoals {
					if goal.Key == goalKey {
						deleteErr := h.agencyService.DeleteGoal(ctx, agencyID, goalKey)
						if deleteErr != nil {
							h.logger.WithError(deleteErr).Error("Failed to delete goal", "goalKey", goalKey, "goalCode", goal.Code)
						} else {
							deletedCount++
							deletedCodes = append(deletedCodes, goal.Code)
							h.logger.Info("Successfully deleted goal", "goalKey", goalKey, "goalCode", goal.Code)
						}
						break
					}
				}
			}

			if deletedCount > 0 {
				responseMessage = fmt.Sprintf("🗑️ **Removed %d Goal(s)**\n\n**Removed**: %s\n\n%s",
					deletedCount,
					strings.Join(deletedCodes, ", "),
					result.Explanation)
			} else {
				responseMessage = fmt.Sprintf("❌ **Failed to Remove Goals**\n\n%s", result.Explanation)
			}
		} else {
			responseMessage = "ℹ️ " + result.Explanation
		}

	case "consolidate":
		if result.ConsolidatedData != nil {
			var parts []string
			if len(result.ConsolidatedData.ConsolidatedGoals) > 0 {
				suggestions := make([]string, len(result.ConsolidatedData.ConsolidatedGoals))
				for i, cGoal := range result.ConsolidatedData.ConsolidatedGoals {
					suggestions[i] = fmt.Sprintf("**%s**: %s", cGoal.SuggestedCode, cGoal.Description)
				}
				parts = append(parts, fmt.Sprintf("💡 **Consolidation Suggestions**\n\n%s", strings.Join(suggestions, "\n")))
			}
			if len(result.ConsolidatedData.RemovedGoals) > 0 {
				parts = append(parts, fmt.Sprintf("**Removed Goals**: %s", strings.Join(result.ConsolidatedData.RemovedGoals, ", ")))
			}
			parts = append(parts, result.ConsolidatedData.Summary)
			responseMessage = strings.Join(parts, "\n\n")
		} else {
			responseMessage = "ℹ️ " + result.Explanation
		}

	case "no_action":
		responseMessage = "✅ " + result.Explanation

	default:
		responseMessage = result.Explanation
	}

	h.logger.Info("Dynamic goal processing completed",
		"action", result.Action,
		"refined_count", len(result.RefinedGoals),
		"generated_count", len(result.GeneratedGoals),
		"no_action", result.NoActionNeeded)

	// Add AI response to conversation
	if addErr := h.designerService.AddMessage(conv.ID, "assistant", responseMessage); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add AI response to conversation")
	}

	// Render chat messages (user + assistant)
	c.Header("Content-Type", "text/html")

	// Get updated conversation to retrieve messages
	updatedConv, err := h.designerService.GetConversation(conv.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get updated conversation")
		c.String(http.StatusInternalServerError, "Failed to render messages")
		return
	}

	// Render last 2 messages (user + assistant)
	messageCount := len(updatedConv.Messages)
	if messageCount >= 2 {
		// Render user message
		userMsg := &updatedConv.Messages[messageCount-2]
		if renderErr := agency_designer.UserMessage(*userMsg).Render(ctx, c.Writer); renderErr != nil {
			h.logger.WithError(renderErr).Error("Failed to render user message")
		}
		// Render assistant message
		aiMsg := &updatedConv.Messages[messageCount-1]
		if renderErr := agency_designer.AIMessage(*aiMsg).Render(ctx, c.Writer); renderErr != nil {
			h.logger.WithError(renderErr).Error("Failed to render AI message")
		}
	}

	h.logger.Info("Successfully processed chat-based goal request",
		"agencyID", agencyID,
		"action", result.Action)
}
