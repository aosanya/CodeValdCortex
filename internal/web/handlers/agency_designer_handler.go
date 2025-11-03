package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AgencyDesignerWebHandler handles web requests for the AI agency designer
type AgencyDesignerWebHandler struct {
	designerService *ai.AgencyDesignerService
	agencyRepo      agency.Repository
	logger          *logrus.Logger
}

// NewAgencyDesignerWebHandler creates a new web handler
func NewAgencyDesignerWebHandler(
	designerService *ai.AgencyDesignerService,
	agencyRepo agency.Repository,
	logger *logrus.Logger,
) *AgencyDesignerWebHandler {
	return &AgencyDesignerWebHandler{
		designerService: designerService,
		agencyRepo:      agencyRepo,
		logger:          logger,
	}
}

// ShowDesigner renders the AI agency designer page
func (h *AgencyDesignerWebHandler) ShowDesigner(c *gin.Context) {
	agencyID := c.Param("id")

	// Get the agency
	ag, err := h.agencyRepo.GetByID(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch agency")
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Agency not found",
		})
		return
	}

	// Check if there's an active conversation for this agency
	var conversation *ai.ConversationContext
	if existingConv, err := h.designerService.GetConversationByAgencyID(agencyID); err == nil {
		conversation = existingConv
	}

	// Try to load the overview so we can pre-fill the introduction editor server-side
	var overview *agency.Overview
	if ov, err := h.agencyRepo.GetOverview(c.Request.Context(), agencyID); err == nil {
		overview = ov
	}

	// Render the designer page (pass overview so introduction is pre-filled)
	component := agency_designer.AgencyDesignerPage(ag, conversation, overview)
	err = component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render agency designer page")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to render page",
		})
		return
	}
}

// ShowConversation retrieves and displays a specific conversation
func (h *AgencyDesignerWebHandler) ShowConversation(c *gin.Context) {
	agencyID := c.Param("id")
	conversationID := c.Param("conversationId")

	// Get the agency
	ag, err := h.agencyRepo.GetByID(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch agency")
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Agency not found",
		})
		return
	}

	// Get the conversation
	conversation, err := h.designerService.GetConversation(conversationID)
	if err != nil || conversation == nil {
		h.logger.WithField("conversation_id", conversationID).Warn("Conversation not found")
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Conversation not found",
		})
		return
	}

	// Try to load the overview so we can pre-fill the introduction editor server-side
	var overview *agency.Overview
	if ov, err := h.agencyRepo.GetOverview(c.Request.Context(), agencyID); err == nil {
		overview = ov
	}

	// Render the designer page with the conversation (pass overview)
	component := agency_designer.AgencyDesignerPage(ag, conversation, overview)
	err = component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render agency designer page")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to render page",
		})
		return
	}
}

// GetRoleDetails returns the details for a specific agent type
func (h *AgencyDesignerWebHandler) GetRoleDetails(c *gin.Context) {
	conversationID := c.Param("conversationId")
	agentTypeID := c.Param("agentId")

	// Get the conversation
	conversation, err := h.designerService.GetConversation(conversationID)
	if err != nil || conversation == nil {
		h.logger.WithField("conversation_id", conversationID).Warn("Conversation not found")
		c.String(http.StatusNotFound, "Conversation not found")
		return
	}

	// Find the agent type
	var agentType *ai.RoleSpec
	for i := range conversation.CurrentDesign.Roles {
		if conversation.CurrentDesign.Roles[i].ID == agentTypeID {
			agentType = &conversation.CurrentDesign.Roles[i]
			break
		}
	}

	if agentType == nil {
		h.logger.WithField("agent_id", agentTypeID).Warn("Agent type not found")
		c.String(http.StatusNotFound, "Agent type not found")
		return
	}

	// Find relationships involving this agent type
	var relationships []ai.AgentRelationship
	for _, rel := range conversation.CurrentDesign.Relationships {
		if rel.From == agentType.ID || rel.To == agentType.ID {
			relationships = append(relationships, rel)
		}
	}

	// Render the agent type details
	component := agency_designer.RoleDetails(*agentType, relationships)
	err = component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render agent type details")
		c.String(http.StatusInternalServerError, "Failed to render details")
		return
	}
}

// GetChatMessages renders just the chat messages for an agency (HTMX endpoint)
func (h *AgencyDesignerWebHandler) GetChatMessages(c *gin.Context) {
	agencyID := c.Param("id")

	// Try to get the most recent conversation for this agency
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		// No conversation exists, render welcome message
		agencyName := c.Query("agencyName")
		if agencyName == "" {
			agencyName = "Agency"
		}

		component := agency_designer.WelcomeMessage(agencyName)
		c.Header("Content-Type", "text/html")
		err = component.Render(c.Request.Context(), c.Writer)
		if err != nil {
			h.logger.WithError(err).Error("Failed to render welcome message")
			c.String(http.StatusInternalServerError, "Failed to render chat")
		}
		return
	}

	// Render the chat messages
	component := agency_designer.ChatMessages(conversation.Messages)
	c.Header("Content-Type", "text/html")
	err = component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render chat messages")
		c.String(http.StatusInternalServerError, "Failed to render chat")
		return
	}
}

// ShowRACIMatrix renders the RACI matrix editor page
func (h *AgencyDesignerWebHandler) ShowRACIMatrix(c *gin.Context) {
	agencyID := c.Param("id")

	// Get the agency
	ag, err := h.agencyRepo.GetByID(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch agency")
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Agency not found",
		})
		return
	}

	// Render the RACI matrix page
	component := agency_designer.AgencyDesignerRACIPage(ag)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render RACI matrix page")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to render page",
		})
		return
	}
}

// RegisterRoutes registers the web routes
func (h *AgencyDesignerWebHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Main designer page (starts new conversation)
	router.GET("/agencies/:id/designer", h.ShowDesigner)

	// View specific conversation
	router.GET("/agencies/:id/designer/conversations/:conversationId", h.ShowConversation)

	// RACI matrix editor
	router.GET("/agencies/:id/raci", h.ShowRACIMatrix)

	// Get agent type details (HTMX endpoint)
	router.GET("/api/v1/conversations/:conversationId/agents/:agentId", h.GetRoleDetails)

	// Get chat messages for an agency (HTMX endpoint)
	router.GET("/agencies/:id/chat-messages", h.GetChatMessages)
}
