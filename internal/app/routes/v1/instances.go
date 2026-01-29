package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/auth"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterInstanceRoutes registers instance management endpoints
func RegisterInstanceRoutes(rg *gin.RouterGroup, instanceService services.InstanceService, authMiddleware *auth.Middleware, logger *logrus.Logger) {
	instanceHandler := handlers.NewInstanceHandler(instanceService, logger)

	// Protected instance routes - require authentication
	protected := rg.Group("/agencies/:id")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.POST("/tags/:name/instances", instanceHandler.StartInstance)
		protected.GET("/instances", instanceHandler.ListInstances)
		protected.GET("/instances/:instance_id", instanceHandler.GetInstance)
		protected.DELETE("/instances/:instance_id", instanceHandler.DeleteInstance)
		protected.POST("/instances/:instance_id/stop", instanceHandler.StopInstance)
		protected.POST("/instances/:instance_id/restart", instanceHandler.RestartInstance)
		protected.GET("/instances/:instance_id/health", instanceHandler.GetInstanceHealth)
		protected.GET("/instances/:instance_id/agents", instanceHandler.GetInstanceAgents)
		protected.POST("/instances/:instance_id/accept-job", instanceHandler.AcceptJob)
		protected.GET("/tags/:name/instances", instanceHandler.ListInstancesByTag)
	}

	logger.Info("Instance endpoints registered (protected)")
}
