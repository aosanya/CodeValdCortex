package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterTagRoutes registers tag management endpoints
func RegisterTagRoutes(rg *gin.RouterGroup, tagService services.TagService, logger *logrus.Logger) {
	tagHandler := handlers.NewTagHandler(tagService, logger)

	rg.POST("/agencies/:id/tags", tagHandler.CreateTag)
	rg.GET("/agencies/:id/tags", tagHandler.ListTags)
	rg.GET("/agencies/:id/tags/:name", tagHandler.GetTag)
	rg.DELETE("/agencies/:id/tags/:name", tagHandler.DeleteTag)
	rg.POST("/agencies/:id/tags/:name/restore", tagHandler.RestoreFromTag)
	rg.GET("/tags/:tag1/compare/:tag2", tagHandler.CompareTags)

	logger.Info("Tag endpoints registered")
}
