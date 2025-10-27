package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/database"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	"github.com/aosanya/CodeValdCortex/internal/web/pages"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	cookieAgencyID = "agency_id"
)

// HomepageHandler handles homepage and agency selection routes
type HomepageHandler struct {
	agencyService  agency.Service
	runtimeManager *runtime.Manager
	dbClient       *database.ArangoClient
	registry       *registry.Repository
	logger         *logrus.Logger
}

// NewHomepageHandler creates a new homepage handler
func NewHomepageHandler(
	agencyService agency.Service,
	runtimeManager *runtime.Manager,
	dbClient *database.ArangoClient,
	registry *registry.Repository,
	logger *logrus.Logger,
) *HomepageHandler {
	return &HomepageHandler{
		agencyService:  agencyService,
		runtimeManager: runtimeManager,
		dbClient:       dbClient,
		registry:       registry,
		logger:         logger,
	}
}

// ShowHomepage renders the agency selection homepage
func (h *HomepageHandler) ShowHomepage(c *gin.Context) {
	// List all agencies
	filters := agency.AgencyFilters{
		// Default: show all active and inactive agencies
		// Exclude archived by default
	}

	agencies, err := h.agencyService.ListAgencies(c.Request.Context(), filters)
	if err != nil {
		h.logger.Errorf("Failed to list agencies: %v", err)
		c.String(http.StatusInternalServerError, "Failed to load agencies")
		return
	}

	// Render homepage
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = pages.Homepage(agencies).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.Errorf("Failed to render homepage: %v", err)
		c.String(http.StatusInternalServerError, "Failed to render page")
		return
	}
}

// SelectAgency sets the selected agency in session and cookie
func (h *HomepageHandler) SelectAgency(c *gin.Context) {
	agencyID := c.Param("id")
	if agencyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agency ID required"})
		return
	}

	// Verify agency exists
	_, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.Errorf("Failed to get agency %s: %v", agencyID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "agency not found"})
		return
	}

	// Set active agency in service
	err = h.agencyService.SetActiveAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.Errorf("Failed to set active agency: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set active agency"})
		return
	}

	// Store in session (if session middleware is available)
	// For now, we'll use cookies
	c.SetCookie(
		cookieAgencyID,
		agencyID,
		3600*24*30, // 30 days
		"/",
		"",
		false, // secure (set to true in production with HTTPS)
		true,  // httpOnly
	)

	h.logger.Infof("Agency %s selected", agencyID)
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"agency_id": agencyID,
		"message":   "Agency selected successfully",
	})
}

// RedirectToAgencyDashboard redirects to the agency-specific dashboard
func (h *HomepageHandler) RedirectToAgencyDashboard(c *gin.Context) {
	agencyID := c.Param("id")
	if agencyID == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Verify agency exists
	_, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.Errorf("Failed to get agency %s: %v", agencyID, err)
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Set active agency
	err = h.agencyService.SetActiveAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.Errorf("Failed to set active agency: %v", err)
	}

	// Set cookie
	c.SetCookie(
		cookieAgencyID,
		agencyID,
		3600*24*30,
		"/",
		"",
		false,
		true,
	)

	// Redirect to dashboard
	c.Redirect(http.StatusFound, fmt.Sprintf("/agencies/%s/dashboard", agencyID))
}

// ShowAgencyDashboard renders the dashboard for a specific agency
func (h *HomepageHandler) ShowAgencyDashboard(c *gin.Context) {
	agencyID := c.Param("id")
	if agencyID == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Get agency details
	ag, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.Errorf("Failed to get agency %s: %v", agencyID, err)
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Set as active agency
	err = h.agencyService.SetActiveAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.Errorf("Failed to set active agency: %v", err)
	}

	// Store agency in context for the dashboard
	c.Set("agency", ag)
	c.Set("agency_id", agencyID)

	h.logger.Infof("Rendering agency dashboard: %s (%s)", ag.Name, ag.ID)

	// Get the agency-specific database
	agencyDB := agencyID
	if ag.Database != "" {
		agencyDB = ag.Database
	}

	// Get database connection for this agency
	db, err := h.dbClient.GetDatabase(c.Request.Context(), agencyDB)
	if err != nil {
		h.logger.Errorf("Failed to connect to agency database %s: %v", agencyDB, err)
		c.String(http.StatusInternalServerError, "Failed to connect to agency database")
		return
	}

	// Create a registry for this agency's database
	agencyRegistry, err := registry.NewRepositoryWithDB(db)
	if err != nil {
		h.logger.Errorf("Failed to create registry for agency %s: %v", agencyID, err)
		c.String(http.StatusInternalServerError, "Failed to initialize agency registry")
		return
	}

	// Create a runtime manager for this agency
	agencyRuntimeManager := runtime.NewManager(h.logger, runtime.ManagerConfig{
		MaxAgents:           100,
		HealthCheckInterval: 30 * time.Second,
		ShutdownTimeout:     30 * time.Second,
		EnableMetrics:       true,
	}, agencyRegistry)

	// Get all agents from the agency-specific database
	agencyAgents := agencyRuntimeManager.ListAgents()

	stats := h.calculateStats(agencyAgents)

	// Render the dashboard with agency context
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = pages.Dashboard(agencyAgents, stats, ag).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.Errorf("Failed to render agency dashboard: %v", err)
		c.String(http.StatusInternalServerError, "Failed to render dashboard")
		return
	}
}

// calculateStats calculates dashboard statistics from agents
func (h *HomepageHandler) calculateStats(agents []*agent.Agent) pages.DashboardStats {
	stats := pages.DashboardStats{
		Total: len(agents),
	}

	for _, a := range agents {
		state := a.GetState()
		switch state {
		case agent.StateRunning:
			stats.Running++
		case agent.StateStopped:
			stats.Stopped++
		case agent.StatePaused:
			stats.Paused++
		}

		if a.IsHealthy() {
			stats.Healthy++
		} else {
			stats.Unhealthy++
		}
	}

	return stats
}

// GetActiveAgency returns the currently active agency
func (h *HomepageHandler) GetActiveAgency(c *gin.Context) {
	ag, err := h.agencyService.GetActiveAgency(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active agency"})
		return
	}

	c.JSON(http.StatusOK, ag)
}

// ShowAgencySwitcher renders the agency switcher modal
func (h *HomepageHandler) ShowAgencySwitcher(c *gin.Context) {
	// List all agencies
	filters := agency.AgencyFilters{}
	agencies, err := h.agencyService.ListAgencies(c.Request.Context(), filters)
	if err != nil {
		h.logger.Errorf("Failed to list agencies: %v", err)
		c.String(http.StatusInternalServerError, "Failed to load agencies")
		return
	}

	// Get current agency
	currentAgency, _ := h.agencyService.GetActiveAgency(c.Request.Context())

	// Render agency switcher modal
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = pages.AgencySwitcherModal(agencies, currentAgency).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.Errorf("Failed to render agency switcher: %v", err)
		c.String(http.StatusInternalServerError, "Failed to render modal")
		return
	}
}
