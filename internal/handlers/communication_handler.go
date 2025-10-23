package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/communication"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CommunicationHandler handles HTTP requests for communication operations
type CommunicationHandler struct {
	messageService *communication.MessageService
	pubSubService  *communication.PubSubService
	logger         *logrus.Logger
}

// NewCommunicationHandler creates a new communication handler
func NewCommunicationHandler(messageService *communication.MessageService, pubSubService *communication.PubSubService, logger *logrus.Logger) *CommunicationHandler {
	return &CommunicationHandler{
		messageService: messageService,
		pubSubService:  pubSubService,
		logger:         logger,
	}
}

// SendMessageRequest represents the request body for sending a message
type SendMessageRequest struct {
	FromAgentID   string                 `json:"from_agent_id" binding:"required"`
	ToAgentID     string                 `json:"to_agent_id" binding:"required"`
	MessageType   string                 `json:"message_type" binding:"required"`
	Payload       map[string]interface{} `json:"payload" binding:"required"`
	Priority      int                    `json:"priority"`
	CorrelationID string                 `json:"correlation_id"`
	ReplyTo       string                 `json:"reply_to"`
	TTL           int                    `json:"ttl"`
	Metadata      map[string]string      `json:"metadata"`
}

// PublishMessageRequest represents the request body for publishing a message
type PublishMessageRequest struct {
	PublisherAgentID   string                 `json:"publisher_agent_id" binding:"required"`
	PublisherAgentType string                 `json:"publisher_agent_type"`
	EventName          string                 `json:"event_name" binding:"required"`
	Payload            map[string]interface{} `json:"payload" binding:"required"`
	PublicationType    string                 `json:"publication_type"`
	TTLSeconds         int                    `json:"ttl_seconds"`
	Metadata           map[string]string      `json:"metadata"`
}

// SendMessage godoc
// @Summary Send a direct message between agents
// @Description Sends a direct message from one agent to another
// @Tags communication
// @Accept json
// @Produce json
// @Param message body SendMessageRequest true "Message details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/communications/messages [post]
func (h *CommunicationHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prepare message options
	opts := &communication.MessageOptions{
		Priority:      req.Priority,
		CorrelationID: req.CorrelationID,
		ReplyTo:       req.ReplyTo,
		TTL:           req.TTL,
		Metadata:      req.Metadata,
	}

	// Send message
	ctx := c.Request.Context()
	msgType := communication.MessageType(req.MessageType)
	messageID, err := h.messageService.SendMessage(ctx, req.FromAgentID, req.ToAgentID, msgType, req.Payload, opts)
	if err != nil {
		h.logger.WithError(err).Error("Failed to send message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message_id": messageID,
		"status":     "sent",
	})
}

// PublishMessage godoc
// @Summary Publish a message to a topic
// @Description Publishes an event or status update that subscribers can receive
// @Tags communication
// @Accept json
// @Produce json
// @Param publication body PublishMessageRequest true "Publication details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/communications/publish [post]
func (h *CommunicationHandler) PublishMessage(c *gin.Context) {
	var req PublishMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prepare publication options
	opts := &communication.PublicationOptions{
		TTLSeconds: req.TTLSeconds,
		Metadata:   req.Metadata,
	}

	if req.PublicationType != "" {
		opts.Type = communication.PublicationType(req.PublicationType)
	}

	// Default agent type if not provided
	agentType := req.PublisherAgentType
	if agentType == "" {
		agentType = "unknown"
	}

	// Publish event
	ctx := c.Request.Context()
	pubID, err := h.pubSubService.Publish(ctx, req.PublisherAgentID, agentType, req.EventName, req.Payload, opts)
	if err != nil {
		h.logger.WithError(err).Error("Failed to publish message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"publication_id": pubID,
		"status":         "published",
	})
}

// RegisterRoutes registers the communication routes
func (h *CommunicationHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/communications")
	{
		// Direct messaging
		v1.POST("/messages", h.SendMessage)

		// Pub/sub messaging
		v1.POST("/publish", h.PublishMessage)
	}
}
