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
		var authMiddleware *auth.Middleware
		if a.authService != nil {
			authHandler := auth.NewHandler(a.authService, a.logger)
			authHandler.RegisterRoutes(apiV1)
			
			// Create auth middleware for protected routes (MVP-AUTH-005)
			authMiddleware = auth.NewMiddleware(a.authService, a.logger)
			a.logger.Info("Authentication endpoints registered")
		}

		// Agency endpoints - Core CRUD (protected)
		if authMiddleware != nil {
			v1routes.RegisterAgencyRoutes(apiV1, a.agencyService, authMiddleware, a.logger)
		}

		// Tag endpoints (protected)
		if a.tagService != nil && authMiddleware != nil {
			v1routes.RegisterTagRoutes(apiV1, *a.tagService, authMiddleware, a.logger)
		}

		// Instance endpoints (protected)
		if a.instanceService != nil && authMiddleware != nil {
			v1routes.RegisterInstanceRoutes(apiV1, a.instanceService, authMiddleware, a.logger)
		}

		// Workbench and issue management endpoints (protected)
		if a.workbenchService != nil && a.issueService != nil && a.instanceService != nil && authMiddleware != nil {
			v1routes.RegisterWorkbenchRoutes(apiV1, a.issueService, a.workbenchService, a.instanceService, a.agencyService, authMiddleware, a.logger)
		}

		// Publication and activation endpoints (protected)
		if (a.publicationService != nil || a.activationService != nil) && authMiddleware != nil {
			v1routes.RegisterPublicationRoutes(apiV1, a.publicationService, a.activationService, authMiddleware, a.logger)
		}

		// Workflow endpoints (protected)
		if a.workflowService != nil && authMiddleware != nil {
			v1routes.RegisterWorkflowRoutes(apiV1, a.workflowService, a.agencyService, authMiddleware, a.logger)
		}

		// Work Items REST API (protected)
		if authMiddleware != nil {
			if err := v1routes.RegisterWorkItemsRoutes(apiV1, a.dbClient.Database(), authMiddleware, a.logger); err != nil {
				a.logger.WithError(err).Warn("Failed to register work items REST API endpoints")
			}
		}

		// AI Refine and Builder endpoints (protected)
		if aiRefineHandler != nil && authMiddleware != nil {
			v1routes.RegisterAIRefineRoutes(
				apiV1,
				aiRefineHandler.(*ai_refine.Handler),
				a.goalRefiner,
				a.workItemBuilder,
				a.roleBuilder,
				a.raciBuilder,
				a.workflowBuilder,
				authMiddleware,
				a.logger,
			)
		}

		// AI Agency Designer endpoints (protected - requires own handler registration)
		if a.aiDesignerService != nil && authMiddleware != nil {
			aiDesignerHandler := ai.NewAgencyDesignerHandler(a.aiDesignerService, a.logger)
			// Apply auth middleware to designer routes
			protected := apiV1.Group("")
			protected.Use(authMiddleware.RequireAuth())
			aiDesignerHandler.RegisterRoutes(protected)
			a.logger.Info("AI Agency Designer endpoints registered (protected)")
		}

		// Webhook endpoints for work item integration (public - called by external systems)
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
