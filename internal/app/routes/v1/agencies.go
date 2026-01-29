package v1

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/auth"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterAgencyRoutes registers all agency-related endpoints
func RegisterAgencyRoutes(rg *gin.RouterGroup, agencyService agency.Service, authMiddleware *auth.Middleware, logger *logrus.Logger) {
	agencyHandler := handlers.NewAgencyHandler(agencyService, logger)

	// Protected agency routes - require authentication
	protected := rg.Group("/agencies")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Core agency CRUD
		protected.GET("", agencyHandler.ListAgencies)
		protected.GET("/:id", agencyHandler.GetAgency)
		protected.POST("", agencyHandler.CreateAgency)
		protected.PUT("/:id", agencyHandler.UpdateAgency)
		protected.DELETE("/:id", agencyHandler.DeleteAgency)
		protected.GET("/active", agencyHandler.GetActiveAgency)
		protected.GET("/:id/statistics", agencyHandler.GetAgencyStatistics)

		// Unified Specification endpoints
		protected.GET("/:id/specification", agencyHandler.GetSpecification)
		protected.PUT("/:id/specification", agencyHandler.UpdateSpecification)
		protected.PUT("/:id/specification/introduction", agencyHandler.UpdateIntroduction)
		protected.PUT("/:id/specification/goals", agencyHandler.UpdateGoals)
		protected.PUT("/:id/specification/work-items", agencyHandler.UpdateWorkItems)
		protected.PUT("/:id/specification/workflows", agencyHandler.UpdateWorkflows)
		protected.PUT("/:id/specification/roles", agencyHandler.UpdateRoles)
		protected.PUT("/:id/specification/raci-matrix", agencyHandler.UpdateRACIMatrixSection)

		// RACI Matrix CRUD
		protected.GET("/:id/raci-matrix", agencyHandler.GetRACIMatrix)
		protected.POST("/:id/raci-matrix", agencyHandler.SaveRACIMatrix)

		// Roles endpoints
		protected.GET("/:id/roles", agencyHandler.GetAgencyRoles)
		protected.GET("/:id/roles/html", agencyHandler.GetAgencyRolesHTML)
		protected.POST("/:id/roles", agencyHandler.CreateAgencyRole)
		protected.GET("/:id/roles/:key", agencyHandler.GetAgencyRole)
		protected.PUT("/:id/roles/:key", agencyHandler.UpdateAgencyRole)
		protected.DELETE("/:id/roles/:key", agencyHandler.DeleteAgencyRole)
	}

	logger.Info("Agency endpoints registered (protected)")
}
