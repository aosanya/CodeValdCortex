package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/arangodb"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/communication"
	"github.com/aosanya/CodeValdCortex/internal/config"
	"github.com/aosanya/CodeValdCortex/internal/database"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	webhandlers "github.com/aosanya/CodeValdCortex/internal/web/handlers"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers/ai_refine"
	webmiddleware "github.com/aosanya/CodeValdCortex/internal/web/middleware"
	"github.com/aosanya/CodeValdCortex/internal/workflow"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// App represents the main application
type App struct {
	config              *config.Config
	server              *http.Server
	logger              *logrus.Logger
	dbClient            *database.ArangoClient
	registry            *registry.Repository
	agencyService       agency.Service
	agencyRepository    agency.Repository
	runtimeManager      *runtime.Manager
	messageService      *communication.MessageService
	pubSubService       *communication.PubSubService
	aiDesignerService   *ai.AgencyDesignerService
	introductionRefiner *ai.IntroductionBuilder
	goalRefiner         *ai.GoalsBuilder
	workItemBuilder     *ai.WorkItemsBuilder
	roleBuilder         *ai.RolesBuilder
	raciBuilder         *ai.RACIBuilder
	workflowBuilder     *ai.WorkflowsBuilder
	workflowService     *workflow.Service
}

// New creates a new application instance
func New(cfg *config.Config) *App {
	logger := logrus.New()

	// Initialize ArangoDB client
	dbClient, err := database.NewArangoClient(&cfg.Database)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to ArangoDB")
	}

	// Verify database connection
	if err := dbClient.Ping(); err != nil {
		logger.WithError(err).Warn("Database ping failed, continuing with limited functionality")
	}

	// Initialize agent registry
	reg, err := registry.NewRepository(dbClient)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize agent registry")
	}

	// Note: Role registry removed - roles now managed via agency specifications

	// Initialize communication repository and services
	logger.Info("Initializing communication services")
	commRepo, err := communication.NewRepository(dbClient)
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize communication repository")
	}

	var messageService *communication.MessageService
	var pubSubService *communication.PubSubService

	if commRepo != nil {
		messageService = communication.NewMessageService(commRepo)
		pubSubService = communication.NewPubSubService(commRepo)
		logger.Info("Communication services initialized successfully")
	}

	// Create runtime manager with registry
	runtimeManager := runtime.NewManager(logger, runtime.ManagerConfig{
		MaxAgents:           100,
		HealthCheckInterval: 30 * time.Second,
		ShutdownTimeout:     30 * time.Second,
		EnableMetrics:       true,
	}, reg)

	// Initialize agency management
	logger.Info("Initializing agency management service")
	agencyRepo, err := arangodb.New(dbClient.Client(), dbClient.Database())
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize agency repository")
	}
	agencyValidator := agency.NewValidator()
	agencyDBInit := agency.NewDatabaseInitializer(dbClient.Client(), logger)
	agencyService := services.NewWithDBInit(agencyRepo, agencyValidator, agencyDBInit, logger)
	logger.Info("Agency management service initialized successfully")

	// Initialize AI services
	var aiDesignerService *ai.AgencyDesignerService
	var introductionRefiner *ai.IntroductionBuilder
	var goalRefiner *ai.GoalsBuilder
	var workItemBuilder *ai.WorkItemsBuilder
	var roleBuilder *ai.RolesBuilder
	var raciBuilder *ai.RACIBuilder
	var workflowBuilder *ai.WorkflowsBuilder
	if cfg.AI.Provider != "" {
		// Build LLM config from app config
		llmConfig := &ai.LLMConfig{
			Provider:    ai.Provider(cfg.AI.Provider),
			APIKey:      cfg.AI.APIKey,
			Model:       cfg.AI.Model,
			BaseURL:     cfg.AI.BaseURL,
			Temperature: cfg.AI.Temperature,
			MaxTokens:   cfg.AI.MaxTokens,
			Timeout:     cfg.AI.Timeout,
		}

		llmClient, err := ai.NewLLMClient(llmConfig)
		if err != nil {
			logger.WithError(err).Error("Failed to initialize LLM client")
		} else {
			aiDesignerService = ai.NewAgencyDesignerService(llmClient, logger)
			introductionRefiner = ai.NewAIIntroductionBuilder(llmClient, logger)
			goalRefiner = ai.NewGoalRefiner(llmClient, logger)
			workItemBuilder = ai.NewAIWorkItemsBuilder(llmClient, logger)
			roleBuilder = ai.NewAIRolesBuilder(llmClient, logger)
			raciBuilder = ai.NewAIRACIBuilder(llmClient, logger)
			workflowBuilder = ai.NewAIWorkflowsBuilder(llmClient, logger)
			logger.Info("AI agency designer service initialized successfully")
		}
	} else {
		logger.Info("AI configuration not provided, AI designer will not be available")
	}

	// Initialize workflow service
	workflowRepo, err := workflow.NewArangoRepository(dbClient.Database(), logger)
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize workflow repository")
	}
	workflowService := workflow.NewService(workflowRepo, logger)
	logger.Info("Workflow service initialized successfully")

	return &App{
		config:              cfg,
		logger:              logger,
		dbClient:            dbClient,
		registry:            reg,
		agencyService:       agencyService,
		agencyRepository:    agencyRepo,
		runtimeManager:      runtimeManager,
		messageService:      messageService,
		pubSubService:       pubSubService,
		aiDesignerService:   aiDesignerService,
		introductionRefiner: introductionRefiner,
		goalRefiner:         goalRefiner,
		workItemBuilder:     workItemBuilder,
		roleBuilder:         roleBuilder,
		raciBuilder:         raciBuilder,
		workflowBuilder:     workflowBuilder,
		workflowService:     workflowService,
	}
}

// Run starts the application
func (a *App) Run() error {
	// Setup HTTP server
	if err := a.setupServer(); err != nil {
		return fmt.Errorf("failed to setup server: %w", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		a.logger.WithFields(logrus.Fields{
			"host": a.config.Server.Host,
			"port": a.config.Server.Port,
		}).Info("Starting HTTP server")

		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down server...")

	// Shutdown runtime manager first
	a.logger.Info("Shutting down runtime manager")
	if err := a.runtimeManager.Shutdown(); err != nil {
		a.logger.WithError(err).Error("Runtime manager shutdown error")
	}

	// Close database connection
	a.logger.Info("Closing database connection")
	if err := a.dbClient.Close(); err != nil {
		a.logger.WithError(err).Error("Database close error")
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.WithError(err).Error("Server forced to shutdown")
		return err
	}

	a.logger.Info("Server exited")
	return nil
}

// setupServer configures the HTTP server
func (a *App) setupServer() error {
	// Set Gin mode based on log level
	if a.config.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

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
	} // Agency middleware
	agencyMiddleware := webmiddleware.NewAgencyMiddleware(a.agencyService, a.logger) // Serve static files
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

	// AI Agency Designer web routes (if available)
	if aiDesignerWebHandler != nil {
		aiDesignerWebHandler.RegisterRoutes(router.Group(""))
		a.logger.Info("AI Agency Designer web routes registered")
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

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "dev",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Agency endpoints
		agencyHandler := handlers.NewAgencyHandler(a.agencyService, a.logger)
		v1.GET("/agencies", agencyHandler.ListAgencies)
		v1.GET("/agencies/:id", agencyHandler.GetAgency)
		v1.POST("/agencies", agencyHandler.CreateAgency)
		v1.PUT("/agencies/:id", agencyHandler.UpdateAgency)
		v1.DELETE("/agencies/:id", agencyHandler.DeleteAgency)
		v1.POST("/agencies/:id/activate", agencyHandler.ActivateAgency)
		v1.GET("/agencies/active", agencyHandler.GetActiveAgency)
		v1.GET("/agencies/:id/statistics", agencyHandler.GetAgencyStatistics)

		// Unified Specification endpoints (replaces separate overview/goals/work-items)
		v1.GET("/agencies/:id/specification", agencyHandler.GetSpecification)
		v1.PUT("/agencies/:id/specification", agencyHandler.UpdateSpecification)
		v1.PUT("/agencies/:id/specification/introduction", agencyHandler.UpdateIntroduction)
		v1.PUT("/agencies/:id/specification/goals", agencyHandler.UpdateGoals)
		v1.PUT("/agencies/:id/specification/work-items", agencyHandler.UpdateWorkItems)
		v1.PUT("/agencies/:id/specification/roles", agencyHandler.UpdateRoles)
		v1.PUT("/agencies/:id/specification/raci-matrix", agencyHandler.UpdateRACIMatrixSection)

		// RACI Matrix CRUD endpoints
		v1.GET("/agencies/:id/raci-matrix", agencyHandler.GetRACIMatrix)
		v1.POST("/agencies/:id/raci-matrix", agencyHandler.SaveRACIMatrix)

		// Roles endpoints
		v1.GET("/agencies/:id/roles", agencyHandler.GetAgencyRoles)
		v1.GET("/agencies/:id/roles/html", agencyHandler.GetAgencyRolesHTML)
		v1.POST("/agencies/:id/roles", agencyHandler.CreateAgencyRole)
		v1.GET("/agencies/:id/roles/:key", agencyHandler.GetAgencyRole)
		v1.PUT("/agencies/:id/roles/:key", agencyHandler.UpdateAgencyRole)
		v1.DELETE("/agencies/:id/roles/:key", agencyHandler.DeleteAgencyRole)

		// RACI Matrix is now part of unified specification endpoint

		// Workflow endpoints
		if a.workflowService != nil {
			workflowHandler := handlers.NewWorkflowHandler(a.workflowService, a.logger)
			v1.POST("/agencies/:id/workflows", workflowHandler.CreateWorkflow)
			v1.GET("/agencies/:id/workflows", workflowHandler.GetWorkflows)
			v1.GET("/agencies/:id/workflows/html", workflowHandler.GetWorkflowsHTML)
			v1.GET("/workflows/:id", workflowHandler.GetWorkflow)
			v1.PUT("/workflows/:id", workflowHandler.UpdateWorkflow)
			v1.DELETE("/workflows/:id", workflowHandler.DeleteWorkflow)
			v1.POST("/workflows/:id/duplicate", workflowHandler.DuplicateWorkflow)
			v1.POST("/workflows/validate", workflowHandler.ValidateWorkflow)
			v1.POST("/workflows/:id/execute", workflowHandler.StartExecution)
			a.logger.Info("Workflow endpoints registered")
		}

		// AI Refine endpoints (if AI services are available)
		if aiRefineHandler != nil {
			v1.POST("/agencies/:id/overview/refine", aiRefineHandler.RefineIntroduction)
			if a.goalRefiner != nil {
				// Main dynamic router - handles all goal operations through natural language prompts
				v1.POST("/agencies/:id/goals/refine-dynamic", aiRefineHandler.RefineGoals)
				// Convenience routes that use RefineGoals with preset prompts
				v1.POST("/agencies/:id/goals/:goalKey/refine", aiRefineHandler.RefineSpecificGoal)
				v1.POST("/agencies/:id/goals/generate", aiRefineHandler.GenerateGoalWithPrompt)
				v1.POST("/agencies/:id/goals/consolidate", aiRefineHandler.ConsolidateGoalsWithPrompt)
			}
			if a.workItemBuilder != nil {
				// DISABLED: Work item AI handlers need refactoring for unified specification model
				// Main dynamic router - handles all work item operations through natural language prompts
				// v1.POST("/agencies/:id/work-items/refine-dynamic", aiRefineHandler.RefineWorkItems)
				// v1.POST("/agencies/:id/work-items/refine-specific", aiRefineHandler.RefineSpecificWorkItem)
				// v1.POST("/agencies/:id/work-items/generate", aiRefineHandler.GenerateWorkItemWithPrompt)
				// v1.POST("/agencies/:id/work-items/consolidate", aiRefineHandler.ConsolidateWorkItemsWithPrompt)
				// v1.POST("/agencies/:id/work-items/enhance-all", aiRefineHandler.EnhanceAllWorkItems)
				a.logger.Warn("Work item AI refine endpoints disabled - need refactoring for unified specification")
			}
			if a.roleBuilder != nil {
				// Main dynamic router - handles all role operations through natural language prompts
				v1.POST("/agencies/:id/roles/refine-dynamic", aiRefineHandler.RefineRoles)
				// Convenience routes that use RefineRoles with preset prompts
				v1.POST("/agencies/:id/roles/refine-specific", aiRefineHandler.RefineSpecificRole)
				v1.POST("/agencies/:id/roles/generate", aiRefineHandler.GenerateRoleWithPrompt)
				v1.POST("/agencies/:id/roles/consolidate", aiRefineHandler.ConsolidateRolesWithPrompt)
				v1.POST("/agencies/:id/roles/enhance-all", aiRefineHandler.EnhanceAllRolesWithPrompt)
			}
			if a.raciBuilder != nil {
				// Main dynamic router - handles all RACI operations through natural language prompts
				v1.POST("/agencies/:id/raci-matrix/refine-dynamic", aiRefineHandler.RefineRACIMappings)
				// Convenience routes that use RefineRACIMappings with preset prompts
				v1.POST("/agencies/:id/raci-matrix/refine-specific", aiRefineHandler.RefineSpecificRACIMapping)
				v1.POST("/agencies/:id/raci-matrix/generate", aiRefineHandler.GenerateRACIMappingWithPrompt)
				v1.POST("/agencies/:id/raci-matrix/consolidate", aiRefineHandler.ConsolidateRACIMappingsWithPrompt)
				v1.POST("/agencies/:id/raci-matrix/create-complete", aiRefineHandler.CreateCompleteRACIMatrixWithPrompt)
			}
			if a.workflowBuilder != nil {
				// DISABLED: Workflow handler needs refactoring for unified specification model
				// v1.POST("/agencies/:id/workflows/refine-dynamic", aiRefineHandler.RefineWorkflows)
				a.logger.Warn("Workflow AI refine endpoint disabled - needs refactoring for unified specification")
			}
			a.logger.Info("AI Refine endpoints registered")
		}

		// AI Agency Designer endpoints (if available)
		if a.aiDesignerService != nil {
			aiDesignerHandler := ai.NewAgencyDesignerHandler(a.aiDesignerService, a.logger)
			aiDesignerHandler.RegisterRoutes(v1)
			a.logger.Info("AI Agency Designer endpoints registered")
		}

		v1.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"app_name": a.config.AppName,
				"status":   "running",
				"version":  "dev",
			})
		})
	}

	// Create server
	a.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(a.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.config.Server.WriteTimeout) * time.Second,
	}

	return nil
}

// loadAgentInstancesFromDirectory loads agent instance definitions from JSON files in a directory
func loadAgentInstancesFromDirectory(ctx context.Context, dir string, repo *registry.Repository, logger *logrus.Logger) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	// Read all JSON files from directory
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read agent instances directory: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	logger.WithFields(logrus.Fields{
		"dir":   dir,
		"count": len(files),
	}).Info("Loading use case agent instances")

	loadedCount := 0
	skippedCount := 0

	// Load each agent instance file
	for _, file := range files {
		count, skipped, err := loadAgentInstancesFromFile(ctx, file, repo, logger)
		if err != nil {
			logger.WithError(err).WithField("file", file).Error("Failed to load agent instances")
			continue
		}
		loadedCount += count
		skippedCount += skipped
	}

	logger.WithFields(logrus.Fields{
		"loaded":  loadedCount,
		"skipped": skippedCount,
	}).Info("Completed loading agent instances")

	return nil
}

// loadAgentInstancesFromFile loads agent instances from a JSON file
func loadAgentInstancesFromFile(ctx context.Context, filename string, repo *registry.Repository, logger *logrus.Logger) (int, int, error) {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON as array of agents
	var agents []agent.Agent
	if err := json.Unmarshal(data, &agents); err != nil {
		return 0, 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	loadedCount := 0
	skippedCount := 0

	// Create each agent instance
	for i := range agents {
		ag := &agents[i]
		// Check if agent already exists
		existing, err := repo.Get(ctx, ag.ID)
		if err == nil && existing != nil {
			logger.WithFields(logrus.Fields{
				"id":   ag.ID,
				"name": ag.Name,
				"type": ag.Type,
			}).Info("Agent instance already exists, skipping")
			skippedCount++
			continue
		} // Create the agent
		if err := repo.Create(ctx, ag); err != nil {
			logger.WithError(err).WithFields(logrus.Fields{
				"id":   ag.ID,
				"name": ag.Name,
				"type": ag.Type,
			}).Error("Failed to create agent instance")
			continue
		}

		logger.WithFields(logrus.Fields{
			"id":   ag.ID,
			"name": ag.Name,
			"type": ag.Type,
			"file": filepath.Base(filename),
		}).Info("Loaded agent instance")

		loadedCount++
	}

	return loadedCount, skippedCount, nil
}
