package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Response represents the standard API response format
type Response struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Error    *ErrorInfo  `json:"error,omitempty"`
	Metadata *Metadata   `json:"metadata"`
}

// ErrorInfo represents error information in API responses
type ErrorInfo struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id"`
}

// Metadata represents response metadata
type Metadata struct {
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id"`
	Version   string    `json:"version"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Success    bool           `json:"success"`
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
	Metadata   *Metadata      `json:"metadata"`
	Error      *ErrorInfo     `json:"error,omitempty"`
}

// Common error codes
const (
	ErrorCodeBadRequest         = "BAD_REQUEST"
	ErrorCodeUnauthorized       = "UNAUTHORIZED"
	ErrorCodeForbidden          = "FORBIDDEN"
	ErrorCodeNotFound           = "NOT_FOUND"
	ErrorCodeConflict           = "CONFLICT"
	ErrorCodeValidation         = "VALIDATION_ERROR"
	ErrorCodeInternalError      = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// SuccessResponse creates a successful API response
func SuccessResponse(c *gin.Context, data interface{}) {
	response := Response{
		Success: true,
		Data:    data,
		Metadata: &Metadata{
			Timestamp: time.Now(),
			RequestID: getRequestID(c),
			Version:   "v1",
		},
	}
	c.JSON(200, response)
}

// SuccessListResponse creates a successful paginated list response
func SuccessListResponse(c *gin.Context, data interface{}, pagination PaginationInfo) {
	response := ListResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Metadata: &Metadata{
			Timestamp: time.Now(),
			RequestID: getRequestID(c),
			Version:   "v1",
		},
	}
	c.JSON(200, response)
}

// ErrorResponse creates an error API response
func ErrorResponse(c *gin.Context, statusCode int, errorCode, message string, details interface{}) {
	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:      errorCode,
			Message:   message,
			Details:   details,
			Timestamp: time.Now(),
			RequestID: getRequestID(c),
		},
		Metadata: &Metadata{
			Timestamp: time.Now(),
			RequestID: getRequestID(c),
			Version:   "v1",
		},
	}
	c.JSON(statusCode, response)
}

// BadRequestError creates a 400 Bad Request error response
func BadRequestError(c *gin.Context, message string, details interface{}) {
	ErrorResponse(c, 400, ErrorCodeBadRequest, message, details)
}

// UnauthorizedError creates a 401 Unauthorized error response
func UnauthorizedError(c *gin.Context, message string) {
	ErrorResponse(c, 401, ErrorCodeUnauthorized, message, nil)
}

// ForbiddenError creates a 403 Forbidden error response
func ForbiddenError(c *gin.Context, message string) {
	ErrorResponse(c, 403, ErrorCodeForbidden, message, nil)
}

// NotFoundError creates a 404 Not Found error response
func NotFoundError(c *gin.Context, message string) {
	ErrorResponse(c, 404, ErrorCodeNotFound, message, nil)
}

// ConflictError creates a 409 Conflict error response
func ConflictError(c *gin.Context, message string, details interface{}) {
	ErrorResponse(c, 409, ErrorCodeConflict, message, details)
}

// ValidationError creates a 422 Validation Error response
func ValidationError(c *gin.Context, message string, details interface{}) {
	ErrorResponse(c, 422, ErrorCodeValidation, message, details)
}

// InternalError creates a 500 Internal Server Error response
func InternalError(c *gin.Context, message string, details interface{}) {
	ErrorResponse(c, 500, ErrorCodeInternalError, message, details)
}

// ServiceUnavailableError creates a 503 Service Unavailable error response
func ServiceUnavailableError(c *gin.Context, message string) {
	ErrorResponse(c, 503, ErrorCodeServiceUnavailable, message, nil)
}

// getRequestID extracts or generates a request ID for tracing
func getRequestID(c *gin.Context) string {
	// First try to get from header
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Try to get from context (set by middleware)
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}

	// Generate a new one
	return uuid.New().String()
}

// Pagination parameters from query
type PaginationParams struct {
	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"per_page" binding:"min=1,max=1000"`
}

// DefaultPaginationParams returns default pagination parameters
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:    1,
		PerPage: 50,
	}
}

// CalculatePagination calculates pagination info
func CalculatePagination(params PaginationParams, total int) PaginationInfo {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PerPage == 0 {
		params.PerPage = 50
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage
	if totalPages == 0 {
		totalPages = 1
	}

	return PaginationInfo{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// GetOffset calculates the database offset for pagination
func (p PaginationParams) GetOffset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.PerPage
}

// SortParams represents sorting parameters
type SortParams struct {
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// DefaultSortParams returns default sort parameters
func DefaultSortParams() SortParams {
	return SortParams{
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// FilterParams represents common filtering parameters
type FilterParams struct {
	Search    string            `form:"search"`
	Status    string            `form:"status"`
	Type      string            `form:"type"`
	Labels    map[string]string `form:"labels"`
	CreatedAt struct {
		From *time.Time `form:"created_from" time_format:"2006-01-02T15:04:05Z07:00"`
		To   *time.Time `form:"created_to" time_format:"2006-01-02T15:04:05Z07:00"`
	} `form:"created_at"`
	UpdatedAt struct {
		From *time.Time `form:"updated_from" time_format:"2006-01-02T15:04:05Z07:00"`
		To   *time.Time `form:"updated_to" time_format:"2006-01-02T15:04:05Z07:00"`
	} `form:"updated_at"`
}

// RequestParams combines common request parameters
type RequestParams struct {
	PaginationParams
	SortParams
	FilterParams
}

// GetDefaultRequestParams returns default request parameters
func GetDefaultRequestParams() RequestParams {
	return RequestParams{
		PaginationParams: DefaultPaginationParams(),
		SortParams:       DefaultSortParams(),
		FilterParams:     FilterParams{},
	}
}

// HealthStatus represents system health information
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
	Uptime    string            `json:"uptime"`
}

// SystemInfo represents system information
type SystemInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	BuildTime   string            `json:"build_time"`
	GoVersion   string            `json:"go_version"`
	Platform    string            `json:"platform"`
	Features    []string          `json:"features"`
	Config      map[string]string `json:"config,omitempty"`
}
