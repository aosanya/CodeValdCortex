package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AgencyDesignerHandler handles HTTP requests for the AI agency designer
type AgencyDesignerHandler struct {
	service *AgencyDesignerService
	logger  *logrus.Logger
}

// NewAgencyDesignerHandler creates a new handler
func NewAgencyDesignerHandler(service *AgencyDesignerService, logger *logrus.Logger) *AgencyDesignerHandler {
	return &AgencyDesignerHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers the agency designer routes
func (h *AgencyDesignerHandler) RegisterRoutes(router *gin.RouterGroup) {
	designer := router.Group("/agencies/:id/designer")
	{
		designer.POST("/conversations", h.StartConversation)
		designer.POST("/conversations/:conversationId/messages", h.SendMessage)
		designer.GET("/conversations/:conversationId", h.GetConversation)
		designer.POST("/conversations/:conversationId/generate", h.GenerateDesign)
	}
}

// StartConversation handles POST /agencies/:id/designer/conversations
func (h *AgencyDesignerHandler) StartConversation(c *gin.Context) {
	agencyID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	conversation, err := h.service.StartConversation(ctx, agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to start conversation")
		c.JSON(500, gin.H{"error": "Failed to start conversation"})
		return
	}

	c.JSON(200, gin.H{
		"conversation_id": conversation.ID,
		"agency_id":       conversation.AgencyID,
		"phase":           conversation.Phase,
		"messages":        conversation.Messages,
	})
}

// SendMessage handles POST /agencies/:agencyId/designer/conversations/:conversationId/messages
func (h *AgencyDesignerHandler) SendMessage(c *gin.Context) {
	conversationID := c.Param("conversationId")

	var req struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	response, err := h.service.SendMessage(ctx, conversationID, req.Message)
	if err != nil {
		h.logger.WithError(err).Error("Failed to send message")
		c.JSON(500, gin.H{"error": "Failed to process message"})
		return
	}

	// Get updated conversation for phase info
	conversation, _ := h.service.GetConversation(conversationID)

	c.JSON(200, gin.H{
		"message": response,
		"phase":   conversation.Phase,
	})
}

// GetConversation handles GET /agencies/:agencyId/designer/conversations/:conversationId
func (h *AgencyDesignerHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("conversationId")

	conversation, err := h.service.GetConversation(conversationID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(200, conversation)
}

// GenerateDesign handles POST /agencies/:agencyId/designer/conversations/:conversationId/generate
func (h *AgencyDesignerHandler) GenerateDesign(c *gin.Context) {
	conversationID := c.Param("conversationId")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	design, err := h.service.GenerateAgencyDesign(ctx, conversationID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate design")
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to generate design: %v", err)})
		return
	}

	c.JSON(200, design)
}
