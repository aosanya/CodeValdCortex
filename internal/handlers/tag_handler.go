package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	tagService services.TagService
	logger     *logrus.Logger
}

// NewTagHandler creates a new tag handler
func NewTagHandler(tagService services.TagService, logger *logrus.Logger) *TagHandler {
	return &TagHandler{
		tagService: tagService,
		logger:     logger,
	}
}

// CreateTag handles POST /api/v1/agencies/:id/tags
func (h *TagHandler) CreateTag(c *gin.Context) {
	agencyID := c.Param("id")

	var req services.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set created by from context (if available)
	if req.CreatedBy == "" {
		if userID, exists := c.Get("user_id"); exists {
			req.CreatedBy = fmt.Sprintf("%v", userID)
		} else {
			req.CreatedBy = "system"
		}
	}

	// Create tag
	tag, err := h.tagService.CreateTag(c.Request.Context(), agencyID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create tag",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"tag_name":  req.Name,
		"tag_type":  req.Type,
	}).Info("created tag")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tag created successfully",
		"tag":     tag,
	})
}

// ListTags handles GET /api/v1/agencies/:id/tags
func (h *TagHandler) ListTags(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse filters from query params
	filters := &services.TagFilters{}

	// Type filter
	if tagType := c.Query("type"); tagType != "" {
		filters.Type = models.TagType(tagType)
	}

	// Name filter (LIKE)
	if nameLike := c.Query("name_like"); nameLike != "" {
		filters.NameLike = nameLike
	}

	// Date range filters
	if fromDate := c.Query("from_date"); fromDate != "" {
		if t, err := time.Parse(time.RFC3339, fromDate); err == nil {
			filters.FromDate = &t
		}
	}

	if toDate := c.Query("to_date"); toDate != "" {
		if t, err := time.Parse(time.RFC3339, toDate); err == nil {
			filters.ToDate = &t
		}
	}

	// Pagination
	if limit := c.Query("limit"); limit != "" {
		if l, err := parseInt(limit); err == nil {
			filters.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := parseInt(offset); err == nil {
			filters.Offset = o
		}
	}

	// List tags
	tags, err := h.tagService.ListTags(c.Request.Context(), agencyID, filters)
	if err != nil {
		h.logger.WithError(err).Error("failed to list tags")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list tags",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags":  tags,
		"count": len(tags),
	})
}

// GetTag handles GET /api/v1/agencies/:id/tags/:name
func (h *TagHandler) GetTag(c *gin.Context) {
	agencyID := c.Param("id")
	tagName := c.Param("name")

	tag, err := h.tagService.GetTag(c.Request.Context(), agencyID, tagName)
	if err != nil {
		h.logger.WithError(err).Error("failed to get tag")
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Tag not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tag": tag,
	})
}

// DeleteTag handles DELETE /api/v1/agencies/:id/tags/:name
func (h *TagHandler) DeleteTag(c *gin.Context) {
	agencyID := c.Param("id")
	tagName := c.Param("name")

	if err := h.tagService.DeleteTag(c.Request.Context(), agencyID, tagName); err != nil {
		h.logger.WithError(err).Error("failed to delete tag")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete tag",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"tag_name":  tagName,
	}).Info("deleted tag")

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag deleted successfully",
	})
}

// RestoreFromTag handles POST /api/v1/agencies/:id/tags/:name/restore
func (h *TagHandler) RestoreFromTag(c *gin.Context) {
	agencyID := c.Param("id")
	tagName := c.Param("name")

	if err := h.tagService.RestoreFromTag(c.Request.Context(), agencyID, tagName); err != nil {
		h.logger.WithError(err).Error("failed to restore from tag")

		statusCode := http.StatusInternalServerError
		if err.Error() == "can only restore to draft agency" || err.Error() == "invalid agency state for operation" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to restore from tag",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"tag_name":  tagName,
	}).Info("restored agency from tag")

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Agency restored from tag '%s'", tagName),
	})
}

// CompareTags handles GET /api/v1/agencies/:id/tags/:tag1/compare/:tag2
func (h *TagHandler) CompareTags(c *gin.Context) {
	agencyID := c.Param("id")
	tag1Name := c.Param("tag1")
	tag2Name := c.Param("tag2")

	comparison, err := h.tagService.CompareTags(c.Request.Context(), agencyID, tag1Name, tag2Name)
	if err != nil {
		h.logger.WithError(err).Error("failed to compare tags")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to compare tags",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comparison": comparison,
	})
}

// Helper function to parse integer from string
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
