package app

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/git/fileindex"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	webhandlers "github.com/aosanya/CodeValdCortex/internal/web/handlers"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers/ai_refine"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers/files"
	webmiddleware "github.com/aosanya/CodeValdCortex/internal/web/middleware"
	"github.com/aosanya/CodeValdCortex/internal/web/pages"
	"github.com/gin-gonic/gin"
)

// registerWebRoutes sets up all web UI routes
func (a *App) registerWebRoutes(router *gin.Engine) {
	// Register agent handler routes
	agentHandler := handlers.NewAgentHandler(a.runtimeManager, a.logger)
	agentHandler.RegisterRoutes(router)

	// Register task handler routes
	taskHandler := handlers.NewTaskHandler(a.runtimeManager)
	taskHandler.RegisterRoutes(router)

	// Register communication handler routes (if services are available)
	if a.messageService != nil && a.pubSubService != nil {
		commHandler := handlers.NewCommunicationHandler(a.messageService, a.pubSubService, a.logger)
		commHandler.RegisterRoutes(router)
		a.logger.Info("Communication endpoints registered")
	} else {
		a.logger.Warn("Communication services not available, endpoints not registered")
	}

	// Register web dashboard handler
	dashboardHandler := webhandlers.NewDashboardHandler(a.runtimeManager, a.logger)
	topologyVisualizerHandler := webhandlers.NewTopologyVisualizerHandler(a.runtimeManager, a.logger)
	// Initialize homepage handler
	homepageHandler := webhandlers.NewHomepageHandler(a.agencyService, a.runtimeManager, a.dbClient, a.registry, a.logger)

	// Initialize AI agency designer web handler (if service available)
	var aiDesignerWebHandler *webhandlers.AgencyDesignerWebHandler
	var chatHandler *webhandlers.ChatHandler
	var aiRefineHandler *ai_refine.Handler
	if a.aiDesignerService != nil && a.introductionRefiner != nil {
		// Create AI refine handler (needed by chat handler and API routes)
		aiRefineHandler = ai_refine.NewHandler(
			a.agencyService,
			a.workflowService,
			a.introductionRefiner,
			a.goalRefiner,
			a.workItemBuilder,
			a.roleBuilder,
			a.raciBuilder,
			a.workflowBuilder,
			a.aiDesignerService,
			a.logger,
		)

		aiDesignerWebHandler = webhandlers.NewAgencyDesignerWebHandler(a.aiDesignerService, a.agencyRepository, a.workflowService, a.logger)
		chatHandler = webhandlers.NewChatHandler(a.aiDesignerService, a.agencyService, a.introductionRefiner, a.goalRefiner, aiRefineHandler, a.logger)
		a.logger.Info("AI Agency Designer web handler initialized")
	}

	// Agency middleware
	agencyMiddleware := webmiddleware.NewAgencyMiddleware(a.agencyService, a.logger)

	// Serve static files
	router.Static("/static", "./static")

	// Web dashboard routes
	router.GET("/", homepageHandler.ShowHomepage)
	router.GET("/topology", topologyVisualizerHandler.ShowTopologyVisualizer)
	router.GET("/geo-network", topologyVisualizerHandler.ShowGeographicVisualizer)

	// Agency routes
	router.POST("/agencies/:id/select", homepageHandler.SelectAgency)
	router.GET("/agencies/:id", homepageHandler.RedirectToAgencyDashboard)

	// Agency-specific dashboard (with middleware to inject agency context)
	router.GET("/agencies/:id/dashboard", agencyMiddleware.InjectAgencyContext(), homepageHandler.ShowAgencyDashboard)

	// Instance management web routes (if available)
	if a.instanceService != nil && a.tagService != nil {
		instanceWebHandler := webhandlers.NewInstanceWebHandler(a.instanceService, a.agencyService, *a.tagService, a.logger)
		router.GET("/agencies/:id/instances", instanceWebHandler.ShowInstancesList)
		router.GET("/agencies/:id/instances/:instance_id", instanceWebHandler.ShowInstanceDashboard)
		a.logger.Info("Instance management web routes registered")
	}

	// Workbench web routes (MVP-WI-008)
	if a.workbenchService != nil && a.issueService != nil && a.instanceService != nil {
		workbenchHandler := webhandlers.NewWorkbenchHandler(a.issueService, a.workbenchService, a.instanceService, a.agencyService, a.logger)
		router.GET("/agencies/:id/workbench", workbenchHandler.ShowInstanceSelector)
		router.GET("/agencies/:id/instances/:instance_id/workbench", workbenchHandler.ShowWorkbench)
		a.logger.Info("Workbench web routes registered")
	}

	// File explorer web routes (if available)
	if a.fileIndexService != nil {
		filesHandler := files.NewHandler(a.fileIndexService, a.agencyRepository, a.logger)
		router.GET("/agencies/:id/instances/:instance_id/explorer", func(c *gin.Context) {
			agencyID := c.Param("id")
			instanceID := c.Param("instance_id")

			// Get path from query parameter (default to root)
			currentPath := c.DefaultQuery("path", "/")

			// Get agency
			agency, err := a.agencyService.GetAgency(c.Request.Context(), agencyID)
			if err != nil {
				c.String(http.StatusNotFound, "Agency not found")
				return
			}

			// Get agency database
			agencyDB := agency.ID
			if agency.Database != "" {
				agencyDB = agency.Database
			}

			// List directory
			entries, err := a.fileIndexService.ListDirectory(c.Request.Context(), agencyDB, instanceID, currentPath)
			if err != nil {
				a.logger.WithError(err).WithField("path", currentPath).Error("Failed to list directory")
				entries = []*fileindex.DirectoryEntry{} // Empty list on error
			}

			// Render file browser page using Templ
			component := pages.FileExplorerPage(agency, instanceID, currentPath, entries)
			component.Render(c.Request.Context(), c.Writer)
		})

		// API routes for file operations
		api := router.Group("/api")
		filesHandler.RegisterRoutes(api)
		a.logger.Info("File explorer routes registered")
	}

	// AI Agency Designer web routes (if available)
	if aiDesignerWebHandler != nil {
		aiDesignerWebHandler.RegisterRoutes(router.Group(""))
		a.logger.Info("AI Agency Designer web routes registered")
	}

	// AI Policy web routes (if available)
	if a.policyService != nil {
		aiPolicyHandler := webhandlers.NewAIPolicyWebHandler(a.policyService, a.agencyService, a.logger)
		aiPolicyHandler.RegisterRoutes(router.Group(""))
		a.logger.Info("AI Policy web routes registered")
	}

	// Chat routes for web interface (if available)
	if chatHandler != nil {
		// Web-specific chat routes (return HTML instead of JSON)
		router.POST("/api/v1/conversations/:conversationId/messages/web", chatHandler.SendMessage)
		router.POST("/api/v1/agencies/:id/designer/conversations/web", chatHandler.StartConversation)
		a.logger.Info("Web chat routes registered")
	}

	// Main dashboard route with agency context injection
	router.GET("/dashboard", agencyMiddleware.InjectAgencyContext(), dashboardHandler.ShowDashboard)

	// API routes for web dashboard (HTMX endpoints)
	webAPI := router.Group("/api/web")
	{
		webAPI.GET("/agents/live", dashboardHandler.GetAgentsLive)
		webAPI.GET("/agents/json", dashboardHandler.GetAgentsJSON) // JSON API for large datasets
		webAPI.POST("/agents/:id/:action", dashboardHandler.HandleAgentAction)

		// Topology visualizer endpoints
		webAPI.GET("/topology/data", topologyVisualizerHandler.GetTopologyData)
		webAPI.GET("/topology/updates", topologyVisualizerHandler.GetTopologyUpdates)
	}
}
