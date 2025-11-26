package fileindex

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/git/models"
	"github.com/aosanya/CodeValdCortex/internal/git/ops"
	"github.com/sirupsen/logrus"
)

// Service provides file browser operations
type Service interface {
	// File operations
	ListDirectory(ctx context.Context, instanceID, path string) ([]*DirectoryEntry, error)
	GetFileContent(ctx context.Context, instanceID, path string) (*FileContent, error)
	CreateFile(ctx context.Context, instanceID, path, content, author, message string) error
	UpdateFile(ctx context.Context, instanceID, path, content, author, message string) error
	DeleteFile(ctx context.Context, instanceID, path, author, message string) error

	// Directory operations
	CreateDirectory(ctx context.Context, instanceID, path, author, message string) error

	// Index operations
	RebuildIndex(ctx context.Context, instanceID string) error
}

// service implements Service interface
type service struct {
	gitOps    ops.GitOps
	indexRepo Repository
	logger    *logrus.Logger
}

// NewService creates a new file browser service
func NewService(gitOps ops.GitOps, indexRepo Repository, logger *logrus.Logger) Service {
	return &service{
		gitOps:    gitOps,
		indexRepo: indexRepo,
		logger:    logger,
	}
}

// ListDirectory lists files and folders in a directory
func (s *service) ListDirectory(ctx context.Context, instanceID, path string) ([]*DirectoryEntry, error) {
	// Normalize path
	if path == "" {
		path = "/"
	}

	// Get entries from index
	entries, err := s.indexRepo.ListDirectory(ctx, instanceID, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	// Convert to DirectoryEntry
	result := make([]*DirectoryEntry, 0, len(entries))
	for _, entry := range entries {
		entryType := "file"
		if entry.IsDir {
			entryType = "directory"
		}

		result = append(result, &DirectoryEntry{
			Name:      entry.Name,
			Path:      entry.Path,
			Type:      entryType,
			Size:      entry.Size,
			MimeType:  entry.MimeType,
			UpdatedBy: entry.UpdatedBy,
			UpdatedAt: entry.UpdatedAt,
			BlobSHA:   entry.BlobSHA,
			TreeSHA:   entry.TreeSHA,
		})
	}

	return result, nil
}

// GetFileContent retrieves file content
func (s *service) GetFileContent(ctx context.Context, instanceID, path string) (*FileContent, error) {
	// Get file index
	index, err := s.indexRepo.GetByPath(ctx, instanceID, path)
	if err != nil {
		return nil, err
	}

	if index.IsDir {
		return nil, fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Get blob content
	content, err := s.gitOps.GetBlob(ctx, index.BlobSHA)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	return &FileContent{
		Path:      index.Path,
		Content:   content,
		SHA:       index.BlobSHA,
		Size:      index.Size,
		MimeType:  index.MimeType,
		UpdatedBy: index.UpdatedBy,
		UpdatedAt: index.UpdatedAt,
	}, nil
}

// CreateFile creates a new file
func (s *service) CreateFile(ctx context.Context, instanceID, path, content, author, message string) error {
	// Check if file already exists
	_, err := s.indexRepo.GetByPath(ctx, instanceID, path)
	if err == nil {
		return fmt.Errorf("file already exists: %s", path)
	}

	// Create blob
	blobSHA, err := s.gitOps.WriteBlob(ctx, instanceID, content)
	if err != nil {
		return fmt.Errorf("failed to create blob: %w", err)
	}

	// Get current commit (HEAD)
	repo, err := s.gitOps.GetRepository(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	ref, err := s.gitOps.GetRef(ctx, instanceID, repo.DefaultRef)
	if err != nil {
		return fmt.Errorf("failed to get HEAD ref: %w", err)
	}

	currentCommit, err := s.gitOps.GetCommit(ctx, ref.Target)
	if err != nil {
		return fmt.Errorf("failed to get current commit: %w", err)
	}

	// Get current tree
	currentTree, err := s.gitOps.GetTree(ctx, currentCommit.Tree)
	if err != nil {
		return fmt.Errorf("failed to get current tree: %w", err)
	}

	// Add new file to tree
	newTree := s.addFileToTree(currentTree, path, blobSHA)

	// Create new tree
	newTreeSHA, err := s.gitOps.WriteTree(ctx, instanceID, newTree.Entries)
	if err != nil {
		return fmt.Errorf("failed to write tree: %w", err)
	}

	// Create commit
	if message == "" {
		message = fmt.Sprintf("Create %s", filepath.Base(path))
	}

	commitSHA, err := s.gitOps.Commit(ctx, instanceID, newTreeSHA, []string{currentCommit.SHA}, author, message)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	// Update ref
	err = s.gitOps.UpdateRef(ctx, instanceID, repo.DefaultRef, commitSHA)
	if err != nil {
		return fmt.Errorf("failed to update ref: %w", err)
	}

	// Update file index
	index := &FileIndex{
		RepoID:     instanceID,
		Path:       path,
		Name:       filepath.Base(path),
		ParentPath: filepath.Dir(path),
		IsDir:      false,
		BlobSHA:    blobSHA,
		Size:       len(content),
		MimeType:   s.detectMimeType(path),
		UpdatedBy:  author,
		UpdatedAt:  time.Now(),
		Created:    time.Now(),
	}

	err = s.indexRepo.IndexFile(ctx, index)
	if err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"path":       path,
		"instance":   instanceID,
		"commit_sha": commitSHA,
	}).Info("Created file")

	return nil
}

// UpdateFile updates an existing file
func (s *service) UpdateFile(ctx context.Context, instanceID, path, content, author, message string) error {
	// Get current file index
	index, err := s.indexRepo.GetByPath(ctx, instanceID, path)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if index.IsDir {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Create new blob
	blobSHA, err := s.gitOps.WriteBlob(ctx, instanceID, content)
	if err != nil {
		return fmt.Errorf("failed to create blob: %w", err)
	}

	// If content unchanged (same SHA), skip commit
	if blobSHA == index.BlobSHA {
		s.logger.WithField("path", path).Debug("File content unchanged, skipping commit")
		return nil
	}

	// Get repository and current commit
	repo, err := s.gitOps.GetRepository(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	ref, err := s.gitOps.GetRef(ctx, instanceID, repo.DefaultRef)
	if err != nil {
		return fmt.Errorf("failed to get HEAD ref: %w", err)
	}

	currentCommit, err := s.gitOps.GetCommit(ctx, ref.Target)
	if err != nil {
		return fmt.Errorf("failed to get current commit: %w", err)
	}

	// Get current tree and update
	currentTree, err := s.gitOps.GetTree(ctx, currentCommit.Tree)
	if err != nil {
		return fmt.Errorf("failed to get current tree: %w", err)
	}

	// Update file in tree
	newTree := s.updateFileInTree(currentTree, path, blobSHA)

	// Create new tree
	newTreeSHA, err := s.gitOps.WriteTree(ctx, instanceID, newTree.Entries)
	if err != nil {
		return fmt.Errorf("failed to write tree: %w", err)
	}

	// Create commit
	if message == "" {
		message = fmt.Sprintf("Update %s", filepath.Base(path))
	}

	commitSHA, err := s.gitOps.Commit(ctx, instanceID, newTreeSHA, []string{currentCommit.SHA}, author, message)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	// Update ref
	err = s.gitOps.UpdateRef(ctx, instanceID, repo.DefaultRef, commitSHA)
	if err != nil {
		return fmt.Errorf("failed to update ref: %w", err)
	}

	// Update file index
	index.BlobSHA = blobSHA
	index.Size = len(content)
	index.UpdatedBy = author
	index.UpdatedAt = time.Now()

	err = s.indexRepo.UpdateIndex(ctx, index)
	if err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"path":       path,
		"instance":   instanceID,
		"commit_sha": commitSHA,
	}).Info("Updated file")

	return nil
}

// DeleteFile deletes a file
func (s *service) DeleteFile(ctx context.Context, instanceID, path, author, message string) error {
	// Check file exists
	index, err := s.indexRepo.GetByPath(ctx, instanceID, path)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if index.IsDir {
		return fmt.Errorf("path is a directory, use DeleteDirectory: %s", path)
	}

	// Get repository and current commit
	repo, err := s.gitOps.GetRepository(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	ref, err := s.gitOps.GetRef(ctx, instanceID, repo.DefaultRef)
	if err != nil {
		return fmt.Errorf("failed to get HEAD ref: %w", err)
	}

	currentCommit, err := s.gitOps.GetCommit(ctx, ref.Target)
	if err != nil {
		return fmt.Errorf("failed to get current commit: %w", err)
	}

	// Get current tree and remove file
	currentTree, err := s.gitOps.GetTree(ctx, currentCommit.Tree)
	if err != nil {
		return fmt.Errorf("failed to get current tree: %w", err)
	}

	// Remove file from tree
	newTree := s.removeFileFromTree(currentTree, path)

	// Create new tree
	newTreeSHA, err := s.gitOps.WriteTree(ctx, instanceID, newTree.Entries)
	if err != nil {
		return fmt.Errorf("failed to write tree: %w", err)
	}

	// Create commit
	if message == "" {
		message = fmt.Sprintf("Delete %s", filepath.Base(path))
	}

	commitSHA, err := s.gitOps.Commit(ctx, instanceID, newTreeSHA, []string{currentCommit.SHA}, author, message)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	// Update ref
	err = s.gitOps.UpdateRef(ctx, instanceID, repo.DefaultRef, commitSHA)
	if err != nil {
		return fmt.Errorf("failed to update ref: %w", err)
	}

	// Delete from index
	err = s.indexRepo.DeleteByPath(ctx, instanceID, path)
	if err != nil {
		return fmt.Errorf("failed to delete from index: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"path":       path,
		"instance":   instanceID,
		"commit_sha": commitSHA,
	}).Info("Deleted file")

	return nil
}

// CreateDirectory creates a new directory
func (s *service) CreateDirectory(ctx context.Context, instanceID, path, author, message string) error {
	// Normalize path
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Check if directory already exists
	_, err := s.indexRepo.GetByPath(ctx, instanceID, path)
	if err == nil {
		return fmt.Errorf("directory already exists: %s", path)
	}

	// Create empty tree for directory
	treeSHA, err := s.gitOps.WriteTree(ctx, instanceID, []models.TreeEntry{})
	if err != nil {
		return fmt.Errorf("failed to create empty tree: %w", err)
	}

	// Index the directory
	index := &FileIndex{
		RepoID:     instanceID,
		Path:       path,
		Name:       filepath.Base(path),
		ParentPath: filepath.Dir(path),
		IsDir:      true,
		TreeSHA:    treeSHA,
		Size:       0,
		UpdatedBy:  author,
		UpdatedAt:  time.Now(),
		Created:    time.Now(),
	}

	err = s.indexRepo.IndexFile(ctx, index)
	if err != nil {
		return fmt.Errorf("failed to index directory: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"path":     path,
		"instance": instanceID,
	}).Info("Created directory")

	return nil
}

// RebuildIndex rebuilds the file index from the current commit
func (s *service) RebuildIndex(ctx context.Context, instanceID string) error {
	// Get repository
	repo, err := s.gitOps.GetRepository(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	// Get current commit
	ref, err := s.gitOps.GetRef(ctx, instanceID, repo.DefaultRef)
	if err != nil {
		return fmt.Errorf("failed to get HEAD ref: %w", err)
	}

	commit, err := s.gitOps.GetCommit(ctx, ref.Target)
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}

	// Walk tree and collect all files
	entries, err := s.walkTree(ctx, instanceID, commit.Tree, "/")
	if err != nil {
		return fmt.Errorf("failed to walk tree: %w", err)
	}

	// Rebuild index
	err = s.indexRepo.RebuildIndex(ctx, instanceID, commit.SHA, entries)
	if err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"instance":   instanceID,
		"commit_sha": commit.SHA,
		"files":      len(entries),
	}).Info("Rebuilt file index")

	return nil
}

// Helper functions

func (s *service) addFileToTree(tree *models.GitTree, path, blobSHA string) *models.GitTree {
	// Simple implementation - just add to root level
	// TODO: Handle nested paths properly
	name := filepath.Base(path)

	newEntry := models.TreeEntry{
		Mode: "100644",
		Type: "blob",
		SHA:  blobSHA,
		Path: name,
	}

	tree.Entries = append(tree.Entries, newEntry)
	return tree
}

func (s *service) updateFileInTree(tree *models.GitTree, path, blobSHA string) *models.GitTree {
	name := filepath.Base(path)

	for i, entry := range tree.Entries {
		if entry.Path == name {
			tree.Entries[i].SHA = blobSHA
			return tree
		}
	}

	// File not found, add it
	return s.addFileToTree(tree, path, blobSHA)
}

func (s *service) removeFileFromTree(tree *models.GitTree, path string) *models.GitTree {
	name := filepath.Base(path)

	var newEntries []models.TreeEntry
	for _, entry := range tree.Entries {
		if entry.Path != name {
			newEntries = append(newEntries, entry)
		}
	}

	tree.Entries = newEntries
	return tree
}

func (s *service) walkTree(ctx context.Context, instanceID, treeSHA, currentPath string) ([]*FileIndex, error) {
	tree, err := s.gitOps.GetTree(ctx, treeSHA)
	if err != nil {
		return nil, err
	}

	var entries []*FileIndex
	now := time.Now()

	for _, entry := range tree.Entries {
		entryPath := filepath.Join(currentPath, entry.Path)

		if entry.Type == "tree" {
			// Add directory entry
			entries = append(entries, &FileIndex{
				RepoID:     instanceID,
				Path:       entryPath,
				Name:       entry.Path,
				ParentPath: currentPath,
				IsDir:      true,
				TreeSHA:    entry.SHA,
				UpdatedBy:  "system",
				UpdatedAt:  now,
				Created:    now,
			})

			// Recursively walk subdirectory
			subEntries, err := s.walkTree(ctx, instanceID, entry.SHA, entryPath)
			if err != nil {
				return nil, err
			}
			entries = append(entries, subEntries...)
		} else {
			// Add file entry
			blob, _ := s.gitOps.GetBlob(ctx, entry.SHA)
			entries = append(entries, &FileIndex{
				RepoID:     instanceID,
				Path:       entryPath,
				Name:       entry.Path,
				ParentPath: currentPath,
				IsDir:      false,
				BlobSHA:    entry.SHA,
				Size:       len(blob),
				MimeType:   s.detectMimeType(entry.Path),
				UpdatedBy:  "system",
				UpdatedAt:  now,
				Created:    now,
			})
		}
	}

	return entries, nil
}

func (s *service) detectMimeType(filename string) string {
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {
		// Default to text/plain for common text extensions
		switch ext {
		case ".md", ".markdown":
			return "text/markdown"
		case ".go":
			return "text/x-go"
		case ".js":
			return "text/javascript"
		case ".json":
			return "application/json"
		case ".yaml", ".yml":
			return "application/x-yaml"
		default:
			return "text/plain"
		}
	}

	return mimeType
}
