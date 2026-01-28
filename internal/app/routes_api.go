package app

import (
	"net/http"
	"time"

	v1routes "github.com/aosanya/CodeValdCortex/internal/app/routes/v1"
	"github.com/aosanya/CodeValdCortex/internal/auth"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers/ai_refine"
	"github.com/gin-gonic/gin"
)

// registerAPIRoutes sets up all API routes
func (a *App) registerAPIRoutes(router *gin.Engine, aiRefineHandler interface{}) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "dev",
		})
	})

	// API v1 routes
	// Handlers live in /internal/handlers (unversioned - business logic)
	// Routes are versioned here via /api/v1 prefix
	// Route registration delegated to /internal/app/routes/v1/* for cleaner organization
	// Future: Add v2 package for breaking changes
	apiV1 := router.Group("/api/v1")
	{
		// Authentication endpoints (MVP-AUTH-003) - Public routes
		if a.authService != nil {
			authHandler := auth.NewHandler(a.authService, a.logger)
			authHandler.RegisterRoutes(apiV1)
			a.logger.Info("Authentication endpoints registered")
		}

		// Agency endpoints - Core CRUD always available
		v1routes.RegisterAgencyRoutes(apiV1, a.agencyService, a.logger)

		// Tag endpoints (if tag service is available)
		if a.tagService != nil {
			v1routes.RegisterTagRoutes(apiV1, *a.tagService, a.logger)
		}

		// Instance endpoints (MVP-PUB-007)
		if a.instanceService != nil {
			v1routes.RegisterInstanceRoutes(apiV1, a.instanceService, a.logger)
		}

		// Workbench and issue management endpoints (MVP-WI-008)
		if a.workbenchService != nil && a.issueService != nil && a.instanceService != nil {
			v1routes.RegisterWorkbenchRoutes(apiV1, a.issueService, a.workbenchService, a.instanceService, a.agencyService, a.logger)
		}

		// Publication and activation endpoints (MVP-PUB-003, MVP-PUB-004)
		if a.publicationService != nil || a.activationService != nil {
			v1routes.RegisterPublicationRoutes(apiV1, a.publicationService, a.activationService, a.logger)
		}

		// Workflow endpoints
		if a.workflowService != nil {
			v1routes.RegisterWorkflowRoutes(apiV1, a.workflowService, a.agencyService, a.logger)
		}

		// Work Items REST API (MVP-RM-003)
		if err := v1routes.RegisterWorkItemsRoutes(apiV1, a.dbClient.Database(), a.logger); err != nil {
			a.logger.WithError(err).Warn("Failed to register work items REST API endpoints")
		}

		// AI Refine and Builder endpoints (if AI services are available)
		if aiRefineHandler != nil {
			v1routes.RegisterAIRefineRoutes(
				apiV1,
				aiRefineHandler.(*ai_refine.Handler),
				a.goalRefiner,
				a.workItemBuilder,
				a.roleBuilder,
				a.raciBuilder,
				a.workflowBuilder,
				a.logger,
			)
		}

		// AI Agency Designer endpoints (if available)
		if a.aiDesignerService != nil {
			aiDesignerHandler := ai.NewAgencyDesignerHandler(a.aiDesignerService, a.logger)
			aiDesignerHandler.RegisterRoutes(apiV1)
			a.logger.Info("AI Agency Designer endpoints registered")
		}

		// Webhook endpoints for work item integration (if available)
		if a.webhookHandler != nil {
			v1routes.RegisterWebhookRoutes(apiV1, a.webhookHandler, a.logger)
		}

		apiV1.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"app_name": a.config.AppName,
				"status":   "running",
				"version":  "dev",
			})
		})
	}
}
