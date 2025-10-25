package middleware

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	cookieAgencyID     = "agency_id"
	contextKeyAgency   = "current_agency"
	contextKeyAgencyID = "agency_id"
)

// AgencyMiddleware handles agency context injection
type AgencyMiddleware struct {
	agencyService agency.Service
	logger        *logrus.Logger
}

// NewAgencyMiddleware creates a new agency middleware
func NewAgencyMiddleware(agencyService agency.Service, logger *logrus.Logger) *AgencyMiddleware {
	return &AgencyMiddleware{
		agencyService: agencyService,
		logger:        logger,
	}
}

// InjectAgencyContext loads the agency from cookie/session and injects it into the request context
// This middleware is optional - it adds agency context if available but doesn't require it
func (m *AgencyMiddleware) InjectAgencyContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get agency ID from cookie
		agencyID, err := c.Cookie(cookieAgencyID)
		if err != nil || agencyID == "" {
			// No agency selected, continue without agency context
			c.Next()
			return
		}

		// Get agency details
		ag, err := m.agencyService.GetAgency(c.Request.Context(), agencyID)
		if err != nil {
			m.logger.Warnf("Failed to load agency %s: %v", agencyID, err)
			// Clear invalid cookie
			c.SetCookie(cookieAgencyID, "", -1, "/", "", false, true)
			c.Next()
			return
		}

		// Set active agency in service
		if err := m.agencyService.SetActiveAgency(c.Request.Context(), agencyID); err != nil {
			m.logger.Warnf("Failed to set active agency: %v", err)
		}

		// Store agency in context
		c.Set(contextKeyAgency, ag)
		c.Set(contextKeyAgencyID, agencyID)

		c.Next()
	}
}

// RequireAgency enforces that an agency must be selected
// Redirects to homepage if no agency is selected
func (m *AgencyMiddleware) RequireAgency() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if agency context exists
		_, exists := c.Get(contextKeyAgency)
		if !exists {
			// No agency selected - redirect to homepage
			m.logger.Debug("No agency selected, redirecting to homepage")

			// Check if it's an API request
			if isAPIRequest(c) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "agency_required",
					"message": "Please select an agency first",
				})
				c.Abort()
				return
			}

			// Redirect to homepage for web requests
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAgency is similar to InjectAgencyContext but also fetches the agency
// and makes it available via GetCurrentAgency helper
func (m *AgencyMiddleware) OptionalAgency() gin.HandlerFunc {
	return m.InjectAgencyContext()
}

// GetCurrentAgency retrieves the current agency from the Gin context
func GetCurrentAgency(c *gin.Context) (*agency.Agency, bool) {
	ag, exists := c.Get(contextKeyAgency)
	if !exists {
		return nil, false
	}

	currentAgency, ok := ag.(*agency.Agency)
	return currentAgency, ok
}

// GetCurrentAgencyID retrieves the current agency ID from the Gin context
func GetCurrentAgencyID(c *gin.Context) (string, bool) {
	id, exists := c.Get(contextKeyAgencyID)
	if !exists {
		return "", false
	}

	agencyID, ok := id.(string)
	return agencyID, ok
}

// isAPIRequest checks if the request is an API call
func isAPIRequest(c *gin.Context) bool {
	// Check if path starts with /api/
	if len(c.Request.URL.Path) >= 5 && c.Request.URL.Path[:5] == "/api/" {
		return true
	}

	// Check Accept header
	accept := c.GetHeader("Accept")
	if accept == "application/json" {
		return true
	}

	// Check if it's an HTMX request (usually expects HTML)
	if c.GetHeader("HX-Request") != "" {
		return false
	}

	return false
}
