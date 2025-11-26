package files

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/git/fileindex"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler handles file browser API requests
type Handler struct {
	fileService  fileindex.Service
	agencyRepo   agency.Repository
	instanceRepo interface {
		GetByID(ctx context.Context, instanceID string, agencyDB string) (interface{}, error)
	}
	logger *logrus.Logger
}

// NewHandler creates a new file handler
func NewHandler(fileService fileindex.Service, agencyRepo agency.Repository, logger *logrus.Logger) *Handler {
	return &Handler{
		fileService: fileService,
		agencyRepo:  agencyRepo,
		logger:      logger,
	}
}

// getAgencyDatabase retrieves the agency database name
func (h *Handler) getAgencyDatabase(ctx context.Context, agencyID string) (string, error) {
	agency, err := h.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return "", fmt.Errorf("failed to get agency: %w", err)
	}

	// Use agency.Database field, fallback to agency ID if not set
	if agency.Database != "" {
		return agency.Database, nil
	}
	return agency.ID, nil
}

// RegisterRoutes registers file browser API routes
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// API routes remain simple, but handlers will lookup agency from instance
	files := rg.Group("/files")
	{
		files.GET("", h.ListFiles)                   // List directory contents
		files.GET("/*path", h.GetFile)               // Get file content
		files.POST("", h.CreateFile)                 // Create new file
		files.PUT("/*path", h.UpdateFile)            // Update file content
		files.DELETE("/*path", h.DeleteFile)         // Delete file or directory
		files.POST("/directory", h.CreateDirectory)  // Create directory
		files.POST("/rebuild-index", h.RebuildIndex) // Rebuild file index
	}
}

// ListFilesRequest represents directory listing request
type ListFilesRequest struct {
	InstanceID string `form:"instance_id" binding:"required"`
	AgencyID   string `form:"agency_id" binding:"required"`
	Path       string `form:"path"`
}

// FileRequest represents file operation request
type FileRequest struct {
	InstanceID string `json:"instance_id" binding:"required"`
	AgencyID   string `json:"agency_id" binding:"required"`
	Path       string `json:"path" binding:"required"`
	Content    string `json:"content"`
	Author     string `json:"author" binding:"required"`
	Message    string `json:"message"`
}

// DirectoryRequest represents directory creation request
type DirectoryRequest struct {
	InstanceID string `json:"instance_id" binding:"required"`
	AgencyID   string `json:"agency_id" binding:"required"`
	Path       string `json:"path" binding:"required"`
	Author     string `json:"author" binding:"required"`
	Message    string `json:"message"`
}

// RebuildIndexRequest represents index rebuild request
type RebuildIndexRequest struct {
	InstanceID string `json:"instance_id" binding:"required"`
	AgencyID   string `json:"agency_id" binding:"required"`
}

// ListFiles lists directory contents
func (h *Handler) ListFiles(c *gin.Context) {
	var req ListFilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get agency database
	agencyDB, err := h.getAgencyDatabase(c.Request.Context(), req.AgencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agency"})
		return
	}

	// Default to root if no path specified
	if req.Path == "" {
		req.Path = "/"
	}

	entries, err := h.fileService.ListDirectory(c.Request.Context(), agencyDB, req.InstanceID, req.Path)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list directory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list directory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":    req.Path,
		"entries": entries,
	})
}

// GetFile retrieves file content or lists directory
func (h *Handler) GetFile(c *gin.Context) {
	instanceID := c.Query("instance_id")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instance_id is required"})
		return
	}

	agencyID := c.Query("agency_id")
	if agencyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agency_id is required"})
		return
	}

	// Get agency database
	agencyDB, err := h.getAgencyDatabase(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agency"})
		return
	}

	path := c.Param("path")
	if path == "" {
		path = "/"
	}

	// Try to get as file first
	fileContent, err := h.fileService.GetFileContent(c.Request.Context(), agencyDB, instanceID, path)
	if err == nil {
		c.JSON(http.StatusOK, fileContent)
		return
	}

	// If not a file, try as directory
	entries, err := h.fileService.ListDirectory(c.Request.Context(), agencyDB, instanceID, path)
	if err != nil {
		h.logger.WithError(err).WithField("path", path).Error("Failed to get file or directory")
		c.JSON(http.StatusNotFound, gin.H{"error": "File or directory not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":    path,
		"type":    "directory",
		"entries": entries,
	})
}

// CreateFile creates a new file
func (h *Handler) CreateFile(c *gin.Context) {
	var req FileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get agency database
	agencyDB, err := h.getAgencyDatabase(c.Request.Context(), req.AgencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agency"})
		return
	}

	// Validate path
	if err := h.validatePath(req.Path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.fileService.CreateFile(
		c.Request.Context(),
		agencyDB,
		req.InstanceID,
		req.Path,
		req.Content,
		req.Author,
		req.Message,
	)

	if err != nil {
		h.logger.WithError(err).Error("Failed to create file")
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "File created successfully",
		"path":    req.Path,
	})
}

// UpdateFile updates file content
func (h *Handler) UpdateFile(c *gin.Context) {
	var req FileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get agency database
	agencyDB, err := h.getAgencyDatabase(c.Request.Context(), req.AgencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agency"})
		return
	}

	// Get path from URL if not in body
	if req.Path == "" {
		req.Path = c.Param("path")
	}

	// Validate path
	if err := h.validatePath(req.Path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.fileService.UpdateFile(
		c.Request.Context(),
		agencyDB,
		req.InstanceID,
		req.Path,
		req.Content,
		req.Author,
		req.Message,
	)

	if err != nil {
		h.logger.WithError(err).Error("Failed to update file")
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File updated successfully",
		"path":    req.Path,
	})
}

// DeleteFile deletes a file
func (h *Handler) DeleteFile(c *gin.Context) {
	instanceID := c.Query("instance_id")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instance_id is required"})
		return
	}

	agencyID := c.Query("agency_id")
	if agencyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agency_id is required"})
		return
	}

	// Get agency database
	agencyDB, err := h.getAgencyDatabase(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agency"})
		return
	}

	author := c.Query("author")
	if author == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "author is required"})
		return
	}

	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	message := c.Query("message")

	// Try to delete as file first
	err = h.fileService.DeleteFile(
		c.Request.Context(),
		agencyDB,
		instanceID,
		path,
		author,
		message,
	)

	// If it's a directory, try DeleteDirectory instead
	if err != nil && strings.Contains(err.Error(), "path is a directory") {
		err = h.fileService.DeleteDirectory(
			c.Request.Context(),
			agencyDB,
			instanceID,
			path,
			author,
			message,
		)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "Directory deleted successfully",
				"path":    path,
			})
			return
		}
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to delete file")
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
		"path":    path,
	})
}

// CreateDirectory creates a new directory
func (h *Handler) CreateDirectory(c *gin.Context) {
	var req DirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get agency database
	agencyDB, err := h.getAgencyDatabase(c.Request.Context(), req.AgencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agency"})
		return
	}

	// Validate path
	if err := h.validatePath(req.Path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.fileService.CreateDirectory(
		c.Request.Context(),
		agencyDB,
		req.InstanceID,
		req.Path,
		req.Author,
		req.Message,
	)

	if err != nil {
		h.logger.WithError(err).Error("Failed to create directory")
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Directory created successfully",
		"path":    req.Path,
	})
}

// RebuildIndex rebuilds the file index from Git tree
func (h *Handler) RebuildIndex(c *gin.Context) {
	var req RebuildIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get agency database
	agencyDB, err := h.getAgencyDatabase(c.Request.Context(), req.AgencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agency"})
		return
	}

	err = h.fileService.RebuildIndex(c.Request.Context(), agencyDB, req.InstanceID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to rebuild index")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rebuild index"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Index rebuilt successfully"})
}

// Helper functions

func (h *Handler) validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Clean path
	path = filepath.Clean(path)

	// Prevent directory traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid path: directory traversal not allowed")
	}

	// Ensure absolute path
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must be absolute (start with /)")
	}

	return nil
}

// UploadFile handles file upload (multipart form)
func (h *Handler) UploadFile(c *gin.Context) {
	instanceID := c.PostForm("instance_id")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instance_id is required"})
		return
	}

	agencyID := c.PostForm("agency_id")
	if agencyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agency_id is required"})
		return
	}

	// Get agency database
	agencyDB, err := h.getAgencyDatabase(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agency"})
		return
	}

	author := c.PostForm("author")
	if author == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "author is required"})
		return
	}

	path := c.PostForm("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	message := c.PostForm("message")

	// Get uploaded file
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Read file content
	contentBytes, err := io.ReadAll(file)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read uploaded file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	content := string(contentBytes)

	// Create or update file
	ctx := context.Background()
	_, getErr := h.fileService.GetFileContent(ctx, agencyDB, instanceID, path)

	var opErr error
	if getErr != nil {
		// File doesn't exist, create it
		opErr = h.fileService.CreateFile(ctx, agencyDB, instanceID, path, content, author, message)
	} else {
		// File exists, update it
		opErr = h.fileService.UpdateFile(ctx, agencyDB, instanceID, path, content, author, message)
	}

	if opErr != nil {
		h.logger.WithError(opErr).Error("Failed to upload file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"path":    path,
	})
}
