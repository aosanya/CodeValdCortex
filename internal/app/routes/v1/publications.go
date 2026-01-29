package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/auth"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterPublicationRoutes registers publication and activation endpoints
func RegisterPublicationRoutes(rg *gin.RouterGroup, publicationService services.PublicationService, activationService services.ActivationService, authMiddleware *auth.Middleware, logger *logrus.Logger) {
	// Publication endpoints - require authentication
	if publicationService != nil {
		pubHandler := handlers.NewPublicationHandler(publicationService, logger)
		protected := rg.Group("/agencies/:id")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.POST("/validate", pubHandler.ValidateForPublish)
			protected.POST("/publish", pubHandler.Publish)
			protected.POST("/activate", pubHandler.Activate)
			protected.POST("/deactivate", pubHandler.Deactivate)
			protected.GET("/publications", pubHandler.GetPublicationHistory)
		}
		pubProtected := rg.Group("/publications")
		pubProtected.Use(authMiddleware.RequireAuth())
		{
			pubProtected.POST("/:id/activate", pubHandler.ActivatePublication)
		}
		logger.Info("Publication endpoints registered (protected)")
	}

	// Activation/Lifecycle endpoints - require authentication
	if activationService != nil {
		activationHandler := handlers.NewActivationHandler(activationService, logger)
		lifecycleProtected := rg.Group("/agencies/:id/lifecycle")
		lifecycleProtected.Use(authMiddleware.RequireAuth())
		{
			lifecycleProtected.POST("/pause", activationHandler.PauseAgency)
			lifecycleProtected.POST("/resume", activationHandler.ResumeAgency)
			lifecycleProtected.POST("/drain", activationHandler.DrainAgency)
			lifecycleProtected.POST("/stop", activationHandler.StopAgency)
		}
		logger.Info("Activation lifecycle endpoints registered (protected)")
	}
}
