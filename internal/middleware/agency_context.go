package middleware

import (
	"context"
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
)

// AgencyContext is middleware that injects agency context into requests
type AgencyContext struct {
	contextManager *agency.ContextManager
}

// NewAgencyContext creates a new agency context middleware
func NewAgencyContext(contextManager *agency.ContextManager) *AgencyContext {
	return &AgencyContext{
		contextManager: contextManager,
	}
}

// Handler wraps an http.Handler with agency context injection
func (m *AgencyContext) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get agency ID from various sources
		agencyID := m.getAgencyID(r)
		
		if agencyID != "" {
			// Inject agency context
			ctx, err := m.contextManager.WithAgency(r.Context(), agencyID)
			if err != nil {
				// Log error but continue without agency context
				// TODO: Add proper logging
				next.ServeHTTP(w, r)
				return
			}
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAgency is middleware that requires an agency to be set
func (m *AgencyContext) RequireAgency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !agency.HasAgencyContext(r.Context()) {
			http.Error(w, "Agency context required", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// getAgencyID extracts agency ID from the request
func (m *AgencyContext) getAgencyID(r *http.Request) string {
	// 1. Check URL path parameter (e.g., /agencies/{id}/...)
	// This would typically be extracted by the router (mux.Vars)
	// For now, check query parameter and header as fallbacks
	
	// 2. Check query parameter
	if agencyID := r.URL.Query().Get("agency_id"); agencyID != "" {
		return agencyID
	}
	
	// 3. Check header
	if agencyID := r.Header.Get("X-Agency-ID"); agencyID != "" {
		return agencyID
	}
	
	// 4. Check session/cookie
	if cookie, err := r.Cookie("agency_id"); err == nil {
		return cookie.Value
	}
	
	return ""
}

// SetAgencyCookie sets the agency ID in a cookie
func SetAgencyCookie(w http.ResponseWriter, agencyID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "agency_id",
		Value:    agencyID,
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true, // Set to true in production with HTTPS
	})
}

// ClearAgencyCookie clears the agency ID cookie
func ClearAgencyCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "agency_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// GetAgencyFromRequest is a helper to get agency from request context
func GetAgencyFromRequest(r *http.Request) (*agency.Agency, error) {
	return agency.GetAgencyFromContext(r.Context())
}

// WithAgencyContext wraps a context with agency information
func WithAgencyContext(ctx context.Context, agencyID string, svc agency.Service) (context.Context, error) {
	agencyObj, err := svc.GetAgency(ctx, agencyID)
	if err != nil {
		return nil, err
	}
	
	ctx = context.WithValue(ctx, agency.AgencyContextKey, agencyObj)
	ctx = context.WithValue(ctx, agency.AgencyIDContextKey, agencyID)
	
	return ctx, nil
}
