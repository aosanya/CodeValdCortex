package handlers

import (
	"fmt"
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// InstanceHandler handles instance-related HTTP requests
type InstanceHandler struct {
	instanceService services.InstanceService
	logger          *logrus.Logger
}

// NewInstanceHandler creates a new instance handler
func NewInstanceHandler(instanceService services.InstanceService, logger *logrus.Logger) *InstanceHandler {
	return &InstanceHandler{
		instanceService: instanceService,
		logger:          logger,
	}
}

// StartInstance handles POST /api/v1/agencies/:id/tags/:name/instances
func (h *InstanceHandler) StartInstance(c *gin.Context) {
	agencyID := c.Param("id")
	tagName := c.Param("name")

	var req models.StartInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set created_by from context if available
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	if userID, exists := c.Get("user_id"); exists {
		req.Metadata["created_by"] = fmt.Sprintf("%v", userID)
	} else {
		req.Metadata["created_by"] = "system"
	}

	// Default environment if not provided
	if req.Environment == "" {
		req.Environment = "development"
	}

	// Start the instance
	instance, err := h.instanceService.StartInstance(c.Request.Context(), agencyID, tagName, &req)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"tag_name":  tagName,
			"error":     err.Error(),
		}).Error("failed to start instance")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start instance",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":   agencyID,
		"tag_name":    tagName,
		"instance_id": instance.InstanceID,
	}).Info("instance started")

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Instance started successfully",
		"instance": instance,
	})
}

// ListInstances handles GET /api/v1/agencies/:id/instances
func (h *InstanceHandler) ListInstances(c *gin.Context) {
	agencyID := c.Param("id")

	instances, err := h.instanceService.ListInstances(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("failed to list instances")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list instances",
			"details": err.Error(),
		})
		return
	}

	// Ensure instances is never nil for JSON marshaling
	if instances == nil {
		instances = []*models.AgencyInstance{}
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instances,
		"count":     len(instances),
	})
}

// GetInstance handles GET /api/v1/agencies/:id/instances/:instanceId
func (h *InstanceHandler) GetInstance(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instanceId")

	instance, err := h.instanceService.GetInstance(c.Request.Context(), agencyID, instanceID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":   agencyID,
			"instance_id": instanceID,
			"error":       err.Error(),
		}).Error("failed to get instance")

		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Instance not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instance": instance,
	})
}

// StopInstance handles DELETE /api/v1/agencies/:id/instances/:instanceId
func (h *InstanceHandler) StopInstance(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instanceId")

	err := h.instanceService.StopInstance(c.Request.Context(), agencyID, instanceID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":   agencyID,
			"instance_id": instanceID,
			"error":       err.Error(),
		}).Error("failed to stop instance")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to stop instance",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":   agencyID,
		"instance_id": instanceID,
	}).Info("instance stopped")

	c.JSON(http.StatusOK, gin.H{
		"message": "Instance stopped successfully",
	})
}

// RestartInstance handles POST /api/v1/agencies/:id/instances/:instanceId/restart
func (h *InstanceHandler) RestartInstance(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instanceId")

	instance, err := h.instanceService.RestartInstance(c.Request.Context(), agencyID, instanceID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":   agencyID,
			"instance_id": instanceID,
			"error":       err.Error(),
		}).Error("failed to restart instance")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to restart instance",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":   agencyID,
		"instance_id": instanceID,
	}).Info("instance restarted")

	c.JSON(http.StatusOK, gin.H{
		"message":  "Instance restarted successfully",
		"instance": instance,
	})
}

// GetInstanceHealth handles GET /api/v1/agencies/:id/instances/:instanceId/health
func (h *InstanceHandler) GetInstanceHealth(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instanceId")

	health, err := h.instanceService.GetInstanceHealth(c.Request.Context(), agencyID, instanceID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":   agencyID,
			"instance_id": instanceID,
			"error":       err.Error(),
		}).Error("failed to get instance health")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get instance health",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"health": health,
	})
}

// GetInstanceAgents handles GET /api/v1/agencies/:id/instances/:instanceId/agents
func (h *InstanceHandler) GetInstanceAgents(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instanceId")

	agents, err := h.instanceService.ListInstanceAgents(c.Request.Context(), agencyID, instanceID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":   agencyID,
			"instance_id": instanceID,
			"error":       err.Error(),
		}).Error("failed to get instance agents")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get instance agents",
			"details": err.Error(),
		})
		return
	}

	// Ensure agents is never nil for JSON marshaling
	if agents == nil {
		agents = []*models.InstanceAgent{}
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"count":  len(agents),
	})
}

// ListInstancesByTag handles GET /api/v1/agencies/:id/tags/:name/instances
func (h *InstanceHandler) ListInstancesByTag(c *gin.Context) {
	agencyID := c.Param("id")
	tagName := c.Param("name")

	instances, err := h.instanceService.ListInstancesByTag(c.Request.Context(), agencyID, tagName)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"tag_name":  tagName,
			"error":     err.Error(),
		}).Error("failed to list instances by tag")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list instances by tag",
			"details": err.Error(),
		})
		return
	}

	// Ensure instances is never nil for JSON marshaling
	if instances == nil {
		instances = []*models.AgencyInstance{}
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instances,
		"count":     len(instances),
		"tag_name":  tagName,
	})
}
