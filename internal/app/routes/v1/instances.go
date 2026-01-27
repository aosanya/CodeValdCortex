package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterInstanceRoutes registers instance management endpoints
func RegisterInstanceRoutes(rg *gin.RouterGroup, instanceService services.InstanceService, logger *logrus.Logger) {
	instanceHandler := handlers.NewInstanceHandler(instanceService, logger)

	rg.POST("/agencies/:id/tags/:name/instances", instanceHandler.StartInstance)
	rg.GET("/agencies/:id/instances", instanceHandler.ListInstances)
	rg.GET("/agencies/:id/instances/:instance_id", instanceHandler.GetInstance)
	rg.DELETE("/agencies/:id/instances/:instance_id", instanceHandler.DeleteInstance)
	rg.POST("/agencies/:id/instances/:instance_id/stop", instanceHandler.StopInstance)
	rg.POST("/agencies/:id/instances/:instance_id/restart", instanceHandler.RestartInstance)
	rg.GET("/agencies/:id/instances/:instance_id/health", instanceHandler.GetInstanceHealth)
	rg.GET("/agencies/:id/instances/:instance_id/agents", instanceHandler.GetInstanceAgents)
	rg.POST("/agencies/:id/instances/:instance_id/accept-job", instanceHandler.AcceptJob)
	rg.GET("/agencies/:id/tags/:name/instances", instanceHandler.ListInstancesByTag)

	logger.Info("Instance endpoints registered")
}
