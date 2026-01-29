package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/arangodb"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/agency/validation"
	"github.com/aosanya/CodeValdCortex/internal/auth"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/communication"
	"github.com/aosanya/CodeValdCortex/internal/config"
	"github.com/aosanya/CodeValdCortex/internal/database"
	"github.com/aosanya/CodeValdCortex/internal/git/fileindex"
	"github.com/aosanya/CodeValdCortex/internal/git/ops"
	gitstorage "github.com/aosanya/CodeValdCortex/internal/git/storage"
	giteaWork "github.com/aosanya/CodeValdCortex/internal/infrastructure/gitea"
	"github.com/aosanya/CodeValdCortex/internal/lifecycle"
	"github.com/aosanya/CodeValdCortex/internal/policy"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers/ai_refine"
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
	authService         *auth.Service
	agencyService       agency.Service
	agencyRepository    agency.Repository
	tagService          *services.TagService
	instanceService     services.InstanceService
	issueService        *services.IssueService
	workbenchService    *services.WorkbenchService
	publicationService  services.PublicationService
	activationService   services.ActivationService
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
	policyService       *policy.Service
	webhookHandler      *giteaWork.Handler
	fileIndexService    fileindex.Service
	gitOps              ops.GitOps
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

	// Initialize authentication service (MVP-AUTH-001, MVP-AUTH-002)
	logger.Info("Initializing authentication service")
	var authService *auth.Service
	{
		authRepo, err := auth.NewRepository(dbClient.Database())
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize auth repository")
		} else {
			jwtSecret := cfg.Auth.JWTSecret
			if jwtSecret == "" {
				logger.Warn("JWT secret not configured, using default (INSECURE for production)")
				jwtSecret = "default-secret-change-in-production"
			}
			authService = auth.NewService(authRepo, jwtSecret)
			logger.Info("Authentication service initialized successfully")
		}
	}

	// Initialize tag service
	logger.Info("Initializing tag service")
	tagRepo, err := arangodb.NewTagRepository(dbClient.Client())
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize tag repository")
	}
	var tagService services.TagService
	if tagRepo != nil {
		// Create slog logger from logrus logger
		slogger := slog.New(slog.NewJSONHandler(logger.WriterLevel(logrus.InfoLevel), nil))
		tagService = services.NewTagService(tagRepo, agencyRepo, slogger)
		logger.Info("Tag service initialized successfully")
	}

	// Initialize instance service (MVP-PUB-007)
	logger.Info("Initializing instance service")
	instanceRepo, err := arangodb.NewInstanceRepository(dbClient.Client())
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize instance repository")
	}
	var instanceService services.InstanceService
	if instanceRepo != nil && tagService != nil {
		// Create slog logger from logrus logger
		slogger := slog.New(slog.NewJSONHandler(logger.WriterLevel(logrus.InfoLevel), nil))
		instanceService = services.NewInstanceService(instanceRepo, tagService, agencyRepo, slogger)
		logger.Info("Instance service initialized successfully")
	}

	// Initialize workbench services (MVP-WI-008)
	logger.Info("Initializing workbench services")
	var issueService *services.IssueService
	var workbenchService *services.WorkbenchService
	{
		// Note: IssueRepository is initialized per agency database, not master database
		// The service will create agency-specific repositories when needed using dbClient
		issueService = services.NewIssueService(agencyRepo, dbClient)
		issueService.SetIssueRepositoryFactory(arangodb.NewIssueRepository)

		if tagService != nil && instanceService != nil {
			workbenchService = services.NewWorkbenchService(tagService, instanceService, dbClient, agencyRepo)
			// Set the factory function for creating issue repositories
			workbenchService.SetIssueRepositoryFactory(arangodb.NewIssueRepository)
			logger.Info("Workbench services initialized successfully")
		}
	}

	// Initialize Git services for file browser (MVP-WI-006)
	logger.Info("Initializing Git services")
	var gitOps ops.GitOps
	var fileIndexService fileindex.Service
	{
		gitStorage, err := gitstorage.NewRepository(dbClient.Database(), logger)
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize Git storage")
		} else {
			gitOps = ops.NewGitOps(gitStorage, logger)

			fileIndexRepo, err := fileindex.NewRepository(dbClient.Client(), logger)
			if err != nil {
				logger.WithError(err).Warn("Failed to initialize file index repository")
			} else {
				fileIndexService = fileindex.NewService(gitOps, fileIndexRepo, logger)
				logger.Info("Git services initialized successfully")
			}
		}
	}

	// Initialize publication service (MVP-PUB-003)
	logger.Info("Initializing publication service")
	slogger := slog.New(slog.NewJSONHandler(logger.WriterLevel(logrus.InfoLevel), nil))
	pubRepo := arangodb.NewPublicationRepository(dbClient.Database(), slogger)

	// Initialize lifecycle manager (needed for activation service)
	lifecycleRepo := lifecycle.NewInMemoryRepository()
	lifecycleManager := lifecycle.NewManager(lifecycleRepo)
	logger.Info("Lifecycle manager initialized successfully")

	// Initialize activation service (MVP-PUB-004)
	// Note: workflow engine is nil for MVP, will be added when workflow integration is complete
	var activationService services.ActivationService
	if pubRepo != nil {
		activationService = services.NewActivationService(pubRepo, agencyRepo, lifecycleManager, nil, slogger)
		logger.Info("Activation service initialized successfully")
	}

	var publicationService services.PublicationService
	if pubRepo != nil && activationService != nil {
		publisherValidator := validation.NewPublisherValidator(slogger)
		stateMachine := agency.NewAgencyStateMachine()
		publicationService = services.NewPublicationService(pubRepo, agencyRepo, stateMachine, publisherValidator, activationService, slogger)
		logger.Info("Publication service initialized successfully")
	} // Initialize AI services
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

	// Initialize policy service
	policyRepo, err := policy.NewArangoRepository(dbClient.Database())
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize policy repository")
	}
	policyService := policy.NewService(policyRepo)
	logger.Info("Policy service initialized successfully")

	// Initialize work tracking integration handler (Gitea)
	var webhookHandler *giteaWork.Handler
	if cfg.WorkTracking.Secret != "" && cfg.WorkTracking.Provider == "gitea" {
		validator := giteaWork.NewValidator(cfg.WorkTracking.Secret)
		webhookRepo, err := giteaWork.NewRepository(dbClient.Database())
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize work tracking repository")
		} else {
			webhookHandler = giteaWork.NewHandler(validator, webhookRepo)
			logger.WithField("provider", "gitea").Info("Work tracking integration handler initialized successfully")
		}
	} else {
		logger.Info("Work tracking integration not configured (set work_tracking.secret and work_tracking.provider)")
	}

	return &App{
		config:              cfg,
		logger:              logger,
		dbClient:            dbClient,
		registry:            reg,
		authService:         authService,
		agencyService:       agencyService,
		agencyRepository:    agencyRepo,
		tagService:          &tagService,
		instanceService:     instanceService,
		issueService:        issueService,
		workbenchService:    workbenchService,
		publicationService:  publicationService,
		activationService:   activationService,
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
		policyService:       policyService,
		webhookHandler:      webhookHandler,
		fileIndexService:    fileIndexService,
		gitOps:              gitOps,
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

	// Register web routes
	a.registerWebRoutes(router)

	// Get AI refine handler for API routes (if available)
	var aiRefineHandler interface{}
	if a.aiDesignerService != nil && a.introductionRefiner != nil {
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
	}

	// Register API routes
	a.registerAPIRoutes(router, aiRefineHandler)

	// Create server
	a.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(a.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.config.Server.WriteTimeout) * time.Second,
	}

	return nil
}
