package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterAgencyRoutes registers all agency-related endpoints
func RegisterAgencyRoutes(rg *gin.RouterGroup, agencyService agency.Service, logger *logrus.Logger) {
	agencyHandler := handlers.NewAgencyHandler(agencyService, logger)

	// Core agency CRUD
	rg.GET("/agencies", agencyHandler.ListAgencies)
	rg.GET("/agencies/:id", agencyHandler.GetAgency)
	rg.POST("/agencies", agencyHandler.CreateAgency)
	rg.PUT("/agencies/:id", agencyHandler.UpdateAgency)
	rg.DELETE("/agencies/:id", agencyHandler.DeleteAgency)
	rg.GET("/agencies/active", agencyHandler.GetActiveAgency)
	rg.GET("/agencies/:id/statistics", agencyHandler.GetAgencyStatistics)

	// Unified Specification endpoints
	rg.GET("/agencies/:id/specification", agencyHandler.GetSpecification)
	rg.PUT("/agencies/:id/specification", agencyHandler.UpdateSpecification)
	rg.PUT("/agencies/:id/specification/introduction", agencyHandler.UpdateIntroduction)
	rg.PUT("/agencies/:id/specification/goals", agencyHandler.UpdateGoals)
	rg.PUT("/agencies/:id/specification/work-items", agencyHandler.UpdateWorkItems)
	rg.PUT("/agencies/:id/specification/workflows", agencyHandler.UpdateWorkflows)
	rg.PUT("/agencies/:id/specification/roles", agencyHandler.UpdateRoles)
	rg.PUT("/agencies/:id/specification/raci-matrix", agencyHandler.UpdateRACIMatrixSection)

	// RACI Matrix CRUD
	rg.GET("/agencies/:id/raci-matrix", agencyHandler.GetRACIMatrix)
	rg.POST("/agencies/:id/raci-matrix", agencyHandler.SaveRACIMatrix)

	// Roles endpoints
	rg.GET("/agencies/:id/roles", agencyHandler.GetAgencyRoles)
	rg.GET("/agencies/:id/roles/html", agencyHandler.GetAgencyRolesHTML)
	rg.POST("/agencies/:id/roles", agencyHandler.CreateAgencyRole)
	rg.GET("/agencies/:id/roles/:key", agencyHandler.GetAgencyRole)
	rg.PUT("/agencies/:id/roles/:key", agencyHandler.UpdateAgencyRole)
	rg.DELETE("/agencies/:id/roles/:key", agencyHandler.DeleteAgencyRole)

	logger.Info("Agency endpoints registered")
}
