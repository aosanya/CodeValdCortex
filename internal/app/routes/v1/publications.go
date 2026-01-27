package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterPublicationRoutes registers publication and activation endpoints
func RegisterPublicationRoutes(rg *gin.RouterGroup, publicationService services.PublicationService, activationService services.ActivationService, logger *logrus.Logger) {
	// Publication endpoints
	if publicationService != nil {
		pubHandler := handlers.NewPublicationHandler(publicationService, logger)
		rg.POST("/agencies/:id/validate", pubHandler.ValidateForPublish)
		rg.POST("/agencies/:id/publish", pubHandler.Publish)
		rg.POST("/agencies/:id/activate", pubHandler.Activate)
		rg.POST("/agencies/:id/deactivate", pubHandler.Deactivate)
		rg.GET("/agencies/:id/publications", pubHandler.GetPublicationHistory)
		rg.POST("/publications/:id/activate", pubHandler.ActivatePublication)
		logger.Info("Publication endpoints registered")
	}

	// Activation/Lifecycle endpoints
	if activationService != nil {
		activationHandler := handlers.NewActivationHandler(activationService, logger)
		rg.POST("/agencies/:id/lifecycle/pause", activationHandler.PauseAgency)
		rg.POST("/agencies/:id/lifecycle/resume", activationHandler.ResumeAgency)
		rg.POST("/agencies/:id/lifecycle/drain", activationHandler.DrainAgency)
		rg.POST("/agencies/:id/lifecycle/stop", activationHandler.StopAgency)
		logger.Info("Activation lifecycle endpoints registered")
	}
}
