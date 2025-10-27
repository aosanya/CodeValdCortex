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
	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/aosanya/CodeValdCortex/internal/communication"
	"github.com/aosanya/CodeValdCortex/internal/config"
	"github.com/aosanya/CodeValdCortex/internal/database"
	"github.com/aosanya/CodeValdCortex/internal/handlers"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	webhandlers "github.com/aosanya/CodeValdCortex/internal/web/handlers"
	webmiddleware "github.com/aosanya/CodeValdCortex/internal/web/middleware"
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
	agentTypeService    registry.AgentTypeService
	agentTypeRepository registry.AgentTypeRepository
	agencyService       agency.Service
	agencyRepository    agency.Repository
	runtimeManager      *runtime.Manager
	messageService      *communication.MessageService
	pubSubService       *communication.PubSubService
	aiDesignerService   *ai.AgencyDesignerService
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

	// Initialize agent type registry with ArangoDB persistence
	logger.Info("Initializing agent type repository with ArangoDB")
	agentTypeRepo, err := registry.NewArangoAgentTypeRepository(dbClient)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize agent type repository")
	}
	agentTypeService := registry.NewAgentTypeService(agentTypeRepo, logger)

	// Register default agent types
	ctx := context.Background()
	if err := registry.InitializeDefaultAgentTypes(ctx, agentTypeService, logger); err != nil {
		logger.WithError(err).Warn("Failed to initialize default agent types")
	}

	// Load use case-specific agent types from config directory
	useCaseConfigDir := os.Getenv("USECASE_CONFIG_DIR")
	if useCaseConfigDir != "" {
		agentTypesDir := filepath.Join(useCaseConfigDir, "config", "agents")
		if err := loadAgentTypesFromDirectory(ctx, agentTypesDir, agentTypeService, logger); err != nil {
			logger.WithError(err).Warn("Failed to load use case agent types")
		}

		// Load use case-specific agent instances from data directory
		agentDataDir := filepath.Join(useCaseConfigDir, "data")
		if err := loadAgentInstancesFromDirectory(ctx, agentDataDir, reg, logger); err != nil {
			logger.WithError(err).Warn("Failed to load use case agent instances")
		}
	}

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
	agencyRepo, err := agency.NewArangoRepository(dbClient.Client(), dbClient.Database())
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize agency repository")
	}
	agencyValidator := agency.NewValidator()
	agencyDBInit := agency.NewDatabaseInitializer(dbClient.Client(), logger)
	agencyService := agency.NewServiceWithDBInit(agencyRepo, agencyValidator, agencyDBInit)
	logger.Info("Agency management service initialized successfully")

	// Initialize AI agency designer service (if configured)
	var aiDesignerService *ai.AgencyDesignerService
	if cfg.AI.Provider != "" && cfg.AI.APIKey != "" {
		logger.WithField("provider", cfg.AI.Provider).Info("Initializing AI agency designer service")
		aiConfig := &ai.LLMConfig{
			Provider:    ai.Provider(cfg.AI.Provider),
			APIKey:      cfg.AI.APIKey,
			Model:       cfg.AI.Model,
			BaseURL:     cfg.AI.BaseURL,
			Temperature: cfg.AI.Temperature,
			MaxTokens:   cfg.AI.MaxTokens,
			Timeout:     cfg.AI.Timeout,
		}
		llmClient, err := ai.NewLLMClient(aiConfig)
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize LLM client, AI designer will not be available")
		} else {
			aiDesignerService = ai.NewAgencyDesignerService(llmClient, logger)
			logger.Info("AI agency designer service initialized successfully")
		}
	} else {
		logger.Info("AI configuration not provided, AI designer will not be available")
	}

	return &App{
		config:              cfg,
		logger:              logger,
		dbClient:            dbClient,
		registry:            reg,
		agentTypeRepository: agentTypeRepo,
		agentTypeService:    agentTypeService,
		agencyService:       agencyService,
		agencyRepository:    agencyRepo,
		runtimeManager:      runtimeManager,
		messageService:      messageService,
		pubSubService:       pubSubService,
		aiDesignerService:   aiDesignerService,
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
	agentTypesWebHandler := webhandlers.NewAgentTypesWebHandler(a.agentTypeService, a.logger)
	topologyVisualizerHandler := webhandlers.NewTopologyVisualizerHandler(a.runtimeManager, a.logger)
	// Initialize homepage handler
	homepageHandler := webhandlers.NewHomepageHandler(a.agencyService, a.runtimeManager, a.dbClient, a.registry, a.logger)

	// Initialize AI agency designer web handler (if service available)
	var aiDesignerWebHandler *webhandlers.AgencyDesignerWebHandler
	if a.aiDesignerService != nil {
		aiDesignerWebHandler = webhandlers.NewAgencyDesignerWebHandler(a.aiDesignerService, a.agencyRepository, a.logger)
		a.logger.Info("AI Agency Designer web handler initialized")
	}

	// Agency middleware
	agencyMiddleware := webmiddleware.NewAgencyMiddleware(a.agencyService, a.logger)

	// Serve static files
	router.Static("/static", "./static")

	// Web dashboard routes
	router.GET("/", homepageHandler.ShowHomepage)
	router.GET("/agent-types", agentTypesWebHandler.ShowAgentTypes)
	router.GET("/topology", topologyVisualizerHandler.ShowTopologyVisualizer)
	router.GET("/geo-network", topologyVisualizerHandler.ShowGeographicVisualizer)

	// Agency routes
	router.POST("/agencies/:id/select", homepageHandler.SelectAgency)
	router.GET("/agencies/:id", homepageHandler.RedirectToAgencyDashboard)

	// Agency-specific dashboard (with middleware to inject agency context)
	router.GET("/agencies/:id/dashboard", agencyMiddleware.InjectAgencyContext(), homepageHandler.ShowAgencyDashboard)
	router.GET("/agencies/switch", homepageHandler.ShowAgencySwitcher)

	// AI Agency Designer web routes (if available)
	if aiDesignerWebHandler != nil {
		aiDesignerWebHandler.RegisterRoutes(router.Group(""))
		a.logger.Info("AI Agency Designer web routes registered")
	}

	// Main dashboard route with agency context injection
	router.GET("/dashboard", agencyMiddleware.InjectAgencyContext(), dashboardHandler.ShowDashboard)

	// API routes for web dashboard (HTMX endpoints)
	webAPI := router.Group("/api/web")
	{
		webAPI.GET("/agents/live", dashboardHandler.GetAgentsLive)
		webAPI.GET("/agents/json", dashboardHandler.GetAgentsJSON) // JSON API for large datasets
		webAPI.POST("/agents/:id/:action", dashboardHandler.HandleAgentAction)

		// Agent types web endpoints
		webAPI.GET("/agent-types", agentTypesWebHandler.GetAgentTypesLive)
		webAPI.POST("/agent-types/:id/:action", agentTypesWebHandler.HandleAgentTypeAction)

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
		// Agent types endpoints
		agentTypeHandler := handlers.NewAgentTypeHandler(a.agentTypeService, a.logger)
		v1.GET("/agent-types", agentTypeHandler.ListAgentTypes)
		v1.GET("/agent-types/:id", agentTypeHandler.GetAgentType)
		v1.POST("/agent-types", agentTypeHandler.CreateAgentType)
		v1.PUT("/agent-types/:id", agentTypeHandler.UpdateAgentType)
		v1.DELETE("/agent-types/:id", agentTypeHandler.DeleteAgentType)
		v1.POST("/agent-types/:id/enable", agentTypeHandler.EnableAgentType)
		v1.POST("/agent-types/:id/disable", agentTypeHandler.DisableAgentType)

		// Agency endpoints
		agencyHandler := handlers.NewAgencyHandler(a.agencyService)
		v1.GET("/agencies", agencyHandler.ListAgencies)
		v1.GET("/agencies/:id", agencyHandler.GetAgency)
		v1.POST("/agencies", agencyHandler.CreateAgency)
		v1.PUT("/agencies/:id", agencyHandler.UpdateAgency)
		v1.DELETE("/agencies/:id", agencyHandler.DeleteAgency)
		v1.POST("/agencies/:id/activate", agencyHandler.ActivateAgency)
		v1.GET("/agencies/active", agencyHandler.GetActiveAgency)
		v1.GET("/agencies/:id/statistics", agencyHandler.GetAgencyStatistics)
		v1.GET("/agencies/:id/overview", agencyHandler.GetOverview)
		v1.PUT("/agencies/:id/overview", agencyHandler.UpdateOverview)
		v1.GET("/agencies/:id/problems", agencyHandler.GetProblems)
		v1.GET("/agencies/:id/problems/html", agencyHandler.GetProblemsHTML)
		v1.POST("/agencies/:id/problems", agencyHandler.CreateProblem)
		v1.PUT("/agencies/:id/problems/:problemKey", agencyHandler.UpdateProblem)
		v1.DELETE("/agencies/:id/problems/:problemKey", agencyHandler.DeleteProblem)

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

// loadAgentTypesFromDirectory loads agent type definitions from JSON files in a directory
func loadAgentTypesFromDirectory(ctx context.Context, dir string, service registry.AgentTypeService, logger *logrus.Logger) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.WithField("dir", dir).Debug("Agent types directory does not exist, skipping")
		return nil
	}

	// Read all JSON files from directory
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read agent types directory: %w", err)
	}

	if len(files) == 0 {
		logger.WithField("dir", dir).Debug("No agent type files found")
		return nil
	}

	logger.WithFields(logrus.Fields{
		"dir":   dir,
		"count": len(files),
	}).Info("Loading use case agent types")

	// Load each agent type file
	for _, file := range files {
		if err := loadAgentTypeFromFile(ctx, file, service, logger); err != nil {
			logger.WithError(err).WithField("file", file).Error("Failed to load agent type")
			continue
		}
	}

	return nil
}

// loadAgentTypeFromFile loads a single agent type from a JSON file
func loadAgentTypeFromFile(ctx context.Context, filename string, service registry.AgentTypeService, logger *logrus.Logger) error {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var agentType registry.AgentType
	if err := json.Unmarshal(data, &agentType); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Register agent type
	if err := service.RegisterType(ctx, &agentType); err != nil {
		return fmt.Errorf("failed to register agent type: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"id":       agentType.ID,
		"name":     agentType.Name,
		"category": agentType.Category,
		"file":     filepath.Base(filename),
	}).Info("Loaded agent type")

	return nil
}

// loadAgentInstancesFromDirectory loads agent instance definitions from JSON files in a directory
func loadAgentInstancesFromDirectory(ctx context.Context, dir string, repo *registry.Repository, logger *logrus.Logger) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.WithField("dir", dir).Debug("Agent instances directory does not exist, skipping")
		return nil
	}

	// Read all JSON files from directory
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read agent instances directory: %w", err)
	}

	if len(files) == 0 {
		logger.WithField("dir", dir).Debug("No agent instance files found")
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
			}).Debug("Agent instance already exists, skipping")
			skippedCount++
			continue
		}

		// Create the agent
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
