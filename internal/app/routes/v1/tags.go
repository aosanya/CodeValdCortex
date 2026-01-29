package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/auth"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterTagRoutes registers tag management endpoints
func RegisterTagRoutes(rg *gin.RouterGroup, tagService services.TagService, authMiddleware *auth.Middleware, logger *logrus.Logger) {
	tagHandler := handlers.NewTagHandler(tagService, logger)

	// Protected tag routes - require authentication
	protected := rg.Group("/agencies/:id/tags")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.POST("", tagHandler.CreateTag)
		protected.GET("", tagHandler.ListTags)
		protected.GET("/:name", tagHandler.GetTag)
		protected.DELETE("/:name", tagHandler.DeleteTag)
		protected.POST("/:name/restore", tagHandler.RestoreFromTag)
	}

	// Tag comparison endpoint (also protected)
	tagsProtected := rg.Group("/tags")
	tagsProtected.Use(authMiddleware.RequireAuth())
	{
		tagsProtected.GET("/:tag1/compare/:tag2", tagHandler.CompareTags)
	}

	logger.Info("Tag endpoints registered (protected)")
}
