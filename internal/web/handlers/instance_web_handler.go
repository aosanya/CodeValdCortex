package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/web/pages"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// InstanceWebHandler handles web UI routes for instance management
type InstanceWebHandler struct {
	instanceService services.InstanceService
	agencyService   agency.Service
	tagService      services.TagService
	logger          *logrus.Logger
}

// NewInstanceWebHandler creates a new instance web handler
func NewInstanceWebHandler(
	instanceService services.InstanceService,
	agencyService agency.Service,
	tagService services.TagService,
	logger *logrus.Logger,
) *InstanceWebHandler {
	return &InstanceWebHandler{
		instanceService: instanceService,
		agencyService:   agencyService,
		tagService:      tagService,
		logger:          logger,
	}
}

// ShowInstancesList shows the instances list page (hybrid view)
func (h *InstanceWebHandler) ShowInstancesList(c *gin.Context) {
	agencyID := c.Param("id")

	// Get agency
	agency, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load agency",
		})
		return
	}

	// Get all tags for this agency (no filters)
	tags, err := h.tagService.ListTags(c.Request.Context(), agencyID, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list tags")
		tags = []*models.AgencyTag{} // Continue with empty list
	}

	// Get all instances for this agency
	instances, err := h.instanceService.ListInstances(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list instances")
		instances = []*models.AgencyInstance{} // Continue with empty list
	}

	// Render the instances list page
	component := pages.InstancesList(agency, tags, instances)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render instances list")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to render page",
		})
		return
	}
}

// ShowInstanceDashboard shows the detailed dashboard for a single instance
func (h *InstanceWebHandler) ShowInstanceDashboard(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")

	// Get agency
	agency, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load agency",
		})
		return
	}

	// Get instance
	instance, err := h.instanceService.GetInstance(c.Request.Context(), agencyID, instanceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get instance")
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Instance not found",
		})
		return
	}

	// Get agent references for this instance
	agents, err := h.instanceService.ListInstanceAgents(c.Request.Context(), agencyID, instanceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list instance agents")
		agents = []*models.InstanceAgent{} // Continue with empty list
	}

	// Render the instance dashboard
	component := pages.InstanceDashboard(agency, instance, agents)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render instance dashboard")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to render page",
		})
		return
	}
}
