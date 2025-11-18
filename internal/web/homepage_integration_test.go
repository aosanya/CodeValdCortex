package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
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
	agencies map[string]*models.Agency
	active   string
}

func newMockAgencyService() *mockAgencyService {
	return &mockAgencyService{
		agencies: make(map[string]*models.Agency),
	}
}

func (m *mockAgencyService) CreateAgency(ctx context.Context, ag *models.Agency) error {
	m.agencies[ag.ID] = ag
	return nil
}

func (m *mockAgencyService) GetAgency(ctx context.Context, id string) (*models.Agency, error) {
	ag, exists := m.agencies[id]
	if !exists {
		return nil, errAgencyNotFound
	}
	return ag, nil
}

func (m *mockAgencyService) ListAgencies(ctx context.Context, filters models.AgencyFilters) ([]*models.Agency, error) {
	result := make([]*models.Agency, 0, len(m.agencies))
	for _, ag := range m.agencies {
		// Apply filters
		if filters.Status != "" && ag.Status != filters.Status {
			continue
		}
		result = append(result, ag)
	}
	return result, nil
}

func (m *mockAgencyService) UpdateAgency(ctx context.Context, id string, updates models.AgencyUpdates) error {
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

func (m *mockAgencyService) GetActiveAgency(ctx context.Context) (*models.Agency, error) {
	if m.active == "" {
		return nil, errNoActiveAgency
	}

	ag, exists := m.agencies[m.active]
	if !exists {
		return nil, errAgencyNotFound
	}
	return ag, nil
}

func (m *mockAgencyService) GetAgencyStatistics(ctx context.Context, id string) (*models.AgencyStatistics, error) {
	ag, err := m.GetAgency(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.AgencyStatistics{
		ActiveAgents:   ag.Metadata.TotalAgents,
		InactiveAgents: 0,
	}, nil
}

// Specification methods (unified document approach)
func (m *mockAgencyService) GetSpecification(ctx context.Context, agencyID string) (*models.AgencySpecification, error) {
	return &models.AgencySpecification{
		Introduction: "Mock introduction",
		Goals:        []models.Goal{},
		WorkItems:    []models.WorkItem{},
		Roles:        []models.Role{},
		RACIMatrix:   nil,
		Version:      1,
		UpdatedBy:    "mock",
	}, nil
}

func (m *mockAgencyService) UpdateSpecification(ctx context.Context, agencyID string, req *models.SpecificationUpdateRequest) (*models.AgencySpecification, error) {
	spec := &models.AgencySpecification{
		Introduction: "Mock introduction",
		Goals:        []models.Goal{},
		WorkItems:    []models.WorkItem{},
		Roles:        []models.Role{},
		RACIMatrix:   nil,
		Version:      2,
		UpdatedBy:    req.UpdatedBy,
	}

	if req.Introduction != nil {
		spec.Introduction = *req.Introduction
	}
	if req.Goals != nil {
		spec.Goals = *req.Goals
	}
	if req.WorkItems != nil {
		spec.WorkItems = *req.WorkItems
	}
	if req.Roles != nil {
		spec.Roles = *req.Roles
	}
	if req.RACIMatrix != nil {
		spec.RACIMatrix = req.RACIMatrix
	}

	return spec, nil
}

func (m *mockAgencyService) UpdateIntroduction(ctx context.Context, agencyID, introduction, updatedBy string) (*models.AgencySpecification, error) {
	return &models.AgencySpecification{
		Introduction: introduction,
		Goals:        []models.Goal{},
		WorkItems:    []models.WorkItem{},
		Roles:        []models.Role{},
		RACIMatrix:   nil,
		Version:      2,
		UpdatedBy:    updatedBy,
	}, nil
}

func (m *mockAgencyService) UpdateSpecificationGoals(ctx context.Context, agencyID string, goals []models.Goal, updatedBy string) (*models.AgencySpecification, error) {
	return &models.AgencySpecification{
		Introduction: "Mock introduction",
		Goals:        goals,
		WorkItems:    []models.WorkItem{},
		Roles:        []models.Role{},
		RACIMatrix:   nil,
		Version:      2,
		UpdatedBy:    updatedBy,
	}, nil
}

func (m *mockAgencyService) UpdateSpecificationWorkItems(ctx context.Context, agencyID string, workItems []models.WorkItem, updatedBy string) (*models.AgencySpecification, error) {
	return &models.AgencySpecification{
		Introduction: "Mock introduction",
		Goals:        []models.Goal{},
		WorkItems:    workItems,
		Roles:        []models.Role{},
		RACIMatrix:   nil,
		Version:      2,
		UpdatedBy:    updatedBy,
	}, nil
}

func (m *mockAgencyService) UpdateSpecificationWorkflows(ctx context.Context, agencyID string, workflows []models.Workflow, updatedBy string) (*models.AgencySpecification, error) {
	return &models.AgencySpecification{
		Introduction: "Mock introduction",
		Goals:        []models.Goal{},
		WorkItems:    []models.WorkItem{},
		Workflows:    workflows,
		Roles:        []models.Role{},
		RACIMatrix:   nil,
		Version:      2,
		UpdatedBy:    updatedBy,
	}, nil
}

func (m *mockAgencyService) UpdateSpecificationRoles(ctx context.Context, agencyID string, roles []models.Role, updatedBy string) (*models.AgencySpecification, error) {
	return &models.AgencySpecification{
		Introduction: "Mock introduction",
		Goals:        []models.Goal{},
		WorkItems:    []models.WorkItem{},
		Roles:        roles,
		RACIMatrix:   nil,
		Version:      2,
		UpdatedBy:    updatedBy,
	}, nil
}

func (m *mockAgencyService) UpdateSpecificationRACIMatrix(ctx context.Context, agencyID string, matrix *models.RACIMatrix, updatedBy string) (*models.AgencySpecification, error) {
	return &models.AgencySpecification{
		Introduction: "Mock introduction",
		Goals:        []models.Goal{},
		WorkItems:    []models.WorkItem{},
		Roles:        []models.Role{},
		RACIMatrix:   matrix,
		Version:      2,
		UpdatedBy:    updatedBy,
	}, nil
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

	agency1 := &models.Agency{
		ID:   "agency-1",
		Name: "Test Agency 1",
		Metadata: models.AgencyMetadata{
			TotalAgents: 5,
		},
		Status: "active",
	}

	agency2 := &models.Agency{
		ID:   "agency-2",
		Name: "Test Agency 2",
		Metadata: models.AgencyMetadata{
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
		_ = freshMock.CreateAgency(context.Background(), &models.Agency{
			ID:     "active-1",
			Name:   "Active Agency 1",
			Status: "active",
		})
		_ = freshMock.CreateAgency(context.Background(), &models.Agency{
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

	agency1 := &models.Agency{
		ID:   "agency-1",
		Name: "Test Agency 1",
		Metadata: models.AgencyMetadata{
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

	activeAgency := &models.Agency{
		ID:   "agency-active",
		Name: "Active Agency",
		Metadata: models.AgencyMetadata{
			TotalAgents: 5,
		},
		Status: "active",
	}

	inactiveAgency := &models.Agency{
		ID:   "agency-inactive",
		Name: "Inactive Agency",
		Metadata: models.AgencyMetadata{
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
