package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/aosanya/CodeValdCortex/internal/web/pages"
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
	// For now, we'll start fresh each time. Later we can add conversation persistence
	var conversation *ai.ConversationContext

	// Render the designer page
	component := pages.AgencyDesignerPage(ag, conversation)
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

	// Render the designer page with the conversation
	component := pages.AgencyDesignerPage(ag, conversation)
	err = component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render agency designer page")
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
}
