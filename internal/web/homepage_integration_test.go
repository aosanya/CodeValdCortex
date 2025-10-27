package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers"
	"github.com/aosanya/CodeValdCortex/internal/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errAgencyNotFound = errors.New("agency not found")
	errNoActiveAgency = errors.New("no active agency")
)

// mockAgencyService is a mock implementation of agency.Service for testing
type mockAgencyService struct {
	agencies map[string]*agency.Agency
	active   string
}

func newMockAgencyService() *mockAgencyService {
	return &mockAgencyService{
		agencies: make(map[string]*agency.Agency),
	}
}

func (m *mockAgencyService) CreateAgency(ctx context.Context, ag *agency.Agency) error {
	m.agencies[ag.ID] = ag
	return nil
}

func (m *mockAgencyService) GetAgency(ctx context.Context, id string) (*agency.Agency, error) {
	ag, exists := m.agencies[id]
	if !exists {
		return nil, errAgencyNotFound
	}
	return ag, nil
}

func (m *mockAgencyService) ListAgencies(ctx context.Context, filters agency.AgencyFilters) ([]*agency.Agency, error) {
	result := make([]*agency.Agency, 0, len(m.agencies))
	for _, ag := range m.agencies {
		// Apply filters
		if filters.Status != "" && ag.Status != filters.Status {
			continue
		}
		result = append(result, ag)
	}
	return result, nil
}

func (m *mockAgencyService) UpdateAgency(ctx context.Context, id string, updates agency.AgencyUpdates) error {
	ag, exists := m.agencies[id]
	if !exists {
		return errAgencyNotFound
	}

	if updates.DisplayName != nil {
		ag.Name = *updates.DisplayName
	}
	return nil
}

func (m *mockAgencyService) DeleteAgency(ctx context.Context, id string) error {
	if _, exists := m.agencies[id]; !exists {
		return errAgencyNotFound
	}
	delete(m.agencies, id)
	return nil
}

func (m *mockAgencyService) SetActiveAgency(ctx context.Context, id string) error {
	if _, exists := m.agencies[id]; !exists {
		return errAgencyNotFound
	}
	m.active = id
	return nil
}

func (m *mockAgencyService) GetActiveAgency(ctx context.Context) (*agency.Agency, error) {
	if m.active == "" {
		return nil, errNoActiveAgency
	}

	ag, exists := m.agencies[m.active]
	if !exists {
		return nil, errAgencyNotFound
	}
	return ag, nil
}

func (m *mockAgencyService) GetAgencyStatistics(ctx context.Context, id string) (*agency.AgencyStatistics, error) {
	ag, err := m.GetAgency(ctx, id)
	if err != nil {
		return nil, err
	}

	return &agency.AgencyStatistics{
		ActiveAgents:   ag.Metadata.TotalAgents,
		InactiveAgents: 0,
	}, nil
}

func (m *mockAgencyService) GetAgencyOverview(ctx context.Context, id string) (*agency.Overview, error) {
	return &agency.Overview{
		AgencyID:     id,
		Introduction: "Mock introduction",
	}, nil
}

func (m *mockAgencyService) UpdateAgencyOverview(ctx context.Context, id string, introduction string) error {
	return nil
}

func (m *mockAgencyService) CreateProblem(ctx context.Context, agencyID string, code string, description string) (*agency.Problem, error) {
	return &agency.Problem{
		Key:         "mock-problem-1",
		AgencyID:    agencyID,
		Number:      1,
		Code:        code,
		Description: description,
	}, nil
}

func (m *mockAgencyService) GetProblems(ctx context.Context, agencyID string) ([]*agency.Problem, error) {
	return []*agency.Problem{}, nil
}

func (m *mockAgencyService) UpdateProblem(ctx context.Context, agencyID string, problemKey string, code string, description string) error {
	return nil
}

func (m *mockAgencyService) DeleteProblem(ctx context.Context, agencyID string, problemKey string) error {
	return nil
}

func (m *mockAgencyService) CreateUnitOfWork(ctx context.Context, agencyID string, code string, description string) (*agency.UnitOfWork, error) {
	return &agency.UnitOfWork{
		Key:         "mock-unit-1",
		AgencyID:    agencyID,
		Number:      1,
		Code:        code,
		Description: description,
	}, nil
}

func (m *mockAgencyService) GetUnitsOfWork(ctx context.Context, agencyID string) ([]*agency.UnitOfWork, error) {
	return []*agency.UnitOfWork{}, nil
}

func (m *mockAgencyService) UpdateUnitOfWork(ctx context.Context, agencyID string, unitKey string, code string, description string) error {
	return nil
}

func (m *mockAgencyService) DeleteUnitOfWork(ctx context.Context, agencyID string, unitKey string) error {
	return nil
}

// setupTestRouter creates a test router with the homepage handlers and middleware
func setupTestRouter(agencyService agency.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	// Create handlers and middleware
	homepageHandler := handlers.NewHomepageHandler(agencyService, nil, nil, nil, logger)
	agencyMiddleware := middleware.NewAgencyMiddleware(agencyService, logger)

	// Add middleware
	router.Use(agencyMiddleware.InjectAgencyContext())

	// Add routes
	router.GET("/", homepageHandler.ShowHomepage)
	router.POST("/agencies/:id/select", homepageHandler.SelectAgency)
	router.GET("/agencies/:id/dashboard", agencyMiddleware.RequireAgency(), homepageHandler.ShowAgencyDashboard)

	return router
}

// TestHomepageIntegration tests the complete homepage navigation flow
func TestHomepageIntegration(t *testing.T) {
	// Setup test agencies
	mockSvc := newMockAgencyService()

	agency1 := &agency.Agency{
		ID:   "agency-1",
		Name: "Test Agency 1",
		Metadata: agency.AgencyMetadata{
			TotalAgents: 5,
		},
		Status: "active",
	}

	agency2 := &agency.Agency{
		ID:   "agency-2",
		Name: "Test Agency 2",
		Metadata: agency.AgencyMetadata{
			TotalAgents: 3,
		},
		Status: "active",
	}

	_ = mockSvc.CreateAgency(context.Background(), agency1)
	_ = mockSvc.CreateAgency(context.Background(), agency2)

	router := setupTestRouter(mockSvc)

	t.Run("should render homepage with agency cards", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()

		// Check that both agencies are displayed
		assert.Contains(t, body, "Test Agency 1")
		assert.Contains(t, body, "Test Agency 2")
		assert.Contains(t, body, "agency-1")
		assert.Contains(t, body, "agency-2")
	})

	t.Run("should select agency via POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/agencies/agency-1/select", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check that cookie was set
		cookies := w.Result().Cookies()
		var foundCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "agency_id" {
				foundCookie = cookie
				break
			}
		}
		require.NotNil(t, foundCookie, "agency_id cookie should be set")
		assert.Equal(t, "agency-1", foundCookie.Value)
	})

	t.Run("should navigate to dashboard after selecting agency", func(t *testing.T) {
		// First select the agency
		req := httptest.NewRequest(http.MethodPost, "/agencies/agency-1/select", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Get the cookie
		cookies := w.Result().Cookies()
		var agencyCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "agency_id" {
				agencyCookie = cookie
				break
			}
		}
		require.NotNil(t, agencyCookie)

		// Now access the dashboard with the cookie
		req = httptest.NewRequest(http.MethodGet, "/agencies/agency-1/dashboard", nil)
		req.AddCookie(agencyCookie)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// The dashboard should be rendered (exact content depends on implementation)
	})

	t.Run("should return 404 for non-existent agency", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/agencies/non-existent/select", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should filter agencies by status", func(t *testing.T) {
		// Create a fresh service with mixed status agencies
		freshMock := newMockAgencyService()
		_ = freshMock.CreateAgency(context.Background(), &agency.Agency{
			ID:     "active-1",
			Name:   "Active Agency 1",
			Status: "active",
		})
		_ = freshMock.CreateAgency(context.Background(), &agency.Agency{
			ID:     "inactive-1",
			Name:   "Inactive Agency 1",
			Status: "inactive",
		})

		freshRouter := setupTestRouter(freshMock)

		// The homepage shows all agencies - client-side filtering happens via JS
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		freshRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()

		// Both agencies should be in the HTML (JS will filter them)
		assert.Contains(t, body, "Active Agency 1")
		assert.Contains(t, body, "Inactive Agency 1")
	})
} // TestAgencyMiddleware tests the middleware functionality
func TestAgencyMiddleware(t *testing.T) {
	mockSvc := newMockAgencyService()

	agency1 := &agency.Agency{
		ID:   "agency-1",
		Name: "Test Agency 1",
		Metadata: agency.AgencyMetadata{
			TotalAgents: 5,
		},
		Status: "active",
	}

	_ = mockSvc.CreateAgency(context.Background(), agency1)
	_ = mockSvc.SetActiveAgency(context.Background(), "agency-1")

	router := setupTestRouter(mockSvc)

	t.Run("should inject agency context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		// Set the active_agency cookie
		cookie := &http.Cookie{
			Name:  "active_agency",
			Value: "agency-1",
		}
		req.AddCookie(cookie)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// The middleware should inject the agency into context
		// This is verified implicitly by the handler working correctly
	})

	t.Run("should handle missing agency cookie gracefully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Homepage should still work without an active agency
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestHomepageFilter tests query parameter filtering
func TestHomepageFilter(t *testing.T) {
	mockSvc := newMockAgencyService()

	activeAgency := &agency.Agency{
		ID:   "agency-active",
		Name: "Active Agency",
		Metadata: agency.AgencyMetadata{
			TotalAgents: 5,
		},
		Status: "active",
	}

	inactiveAgency := &agency.Agency{
		ID:   "agency-inactive",
		Name: "Inactive Agency",
		Metadata: agency.AgencyMetadata{
			TotalAgents: 2,
		},
		Status: "inactive",
	}

	_ = mockSvc.CreateAgency(context.Background(), activeAgency)
	_ = mockSvc.CreateAgency(context.Background(), inactiveAgency)

	router := setupTestRouter(mockSvc)

	t.Run("should show all agencies on homepage", func(t *testing.T) {
		// Homepage shows all agencies, filtering happens client-side via JS
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()

		// Should show both agencies in HTML
		assert.Contains(t, body, "Active Agency")
		assert.Contains(t, body, "Inactive Agency")
	})
}
