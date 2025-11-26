package fileindex

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
)

const (
	FileIndexCollection = "file_index"
)

// Repository handles file index operations in ArangoDB
type Repository interface {
	// Index operations
	IndexFile(ctx context.Context, index *FileIndex) error
	GetByPath(ctx context.Context, repoID, path string) (*FileIndex, error)
	ListDirectory(ctx context.Context, repoID, path string) ([]*FileIndex, error)
	DeleteByPath(ctx context.Context, repoID, path string) error
	DeleteDirectory(ctx context.Context, repoID, path string) error
	UpdateIndex(ctx context.Context, index *FileIndex) error

	// Batch operations
	RebuildIndex(ctx context.Context, repoID, commitSHA string, entries []*FileIndex) error
}

// repository implements Repository interface
type repository struct {
	db     driver.Database
	logger *logrus.Logger
}

// NewRepository creates a new file index repository
func NewRepository(db driver.Database, logger *logrus.Logger) (Repository, error) {
	r := &repository{
		db:     db,
		logger: logger,
	}

	// Ensure collection exists
	if err := r.ensureCollection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	return r, nil
}

// ensureCollection creates file_index collection if it doesn't exist
func (r *repository) ensureCollection(ctx context.Context) error {
	exists, err := r.db.CollectionExists(ctx, FileIndexCollection)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}

	if !exists {
		_, err = r.db.CreateCollection(ctx, FileIndexCollection, nil)
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
		r.logger.WithField("collection", FileIndexCollection).Info("Created file index collection")
	}

	return nil
}

// IndexFile stores or updates a file index entry
func (r *repository) IndexFile(ctx context.Context, index *FileIndex) error {
	collection, err := r.db.Collection(ctx, FileIndexCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Use path as key for fast lookups
	index.Key = r.makeKey(index.RepoID, index.Path)

	// Try to create, update if exists
	_, err = collection.CreateDocument(ctx, index)
	if err != nil {
		if driver.IsConflict(err) {
			// Document exists, update it
			_, err = collection.UpdateDocument(ctx, index.Key, index)
			if err != nil {
				return fmt.Errorf("failed to update file index: %w", err)
			}
			r.logger.WithFields(logrus.Fields{
				"path":    index.Path,
				"repo_id": index.RepoID,
			}).Debug("Updated file index")
		} else {
			return fmt.Errorf("failed to create file index: %w", err)
		}
	} else {
		r.logger.WithFields(logrus.Fields{
			"path":    index.Path,
			"repo_id": index.RepoID,
		}).Debug("Created file index")
	}

	return nil
}

// GetByPath retrieves a file index by path
func (r *repository) GetByPath(ctx context.Context, repoID, path string) (*FileIndex, error) {
	collection, err := r.db.Collection(ctx, FileIndexCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	key := r.makeKey(repoID, path)
	var index FileIndex
	_, err = collection.ReadDocument(ctx, key, &index)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read file index: %w", err)
	}

	return &index, nil
}

// ListDirectory lists all files and folders in a directory
func (r *repository) ListDirectory(ctx context.Context, repoID, path string) ([]*FileIndex, error) {
	// Normalize path
	if path == "" {
		path = "/"
	}
	if !strings.HasSuffix(path, "/") && path != "/" {
		path = path + "/"
	}

	query := `
		FOR file IN @@collection
			FILTER file.repo_id == @repo_id
			FILTER file.parent_path == @parent_path
			SORT file.is_dir DESC, file.name ASC
			RETURN file
	`

	bindVars := map[string]interface{}{
		"@collection": FileIndexCollection,
		"repo_id":     repoID,
		"parent_path": strings.TrimSuffix(path, "/"),
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query directory: %w", err)
	}
	defer cursor.Close()

	var entries []*FileIndex
	for cursor.HasMore() {
		var entry FileIndex
		_, err := cursor.ReadDocument(ctx, &entry)
		if err != nil {
			return nil, fmt.Errorf("failed to read entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

// DeleteByPath deletes a file index entry
func (r *repository) DeleteByPath(ctx context.Context, repoID, path string) error {
	collection, err := r.db.Collection(ctx, FileIndexCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	key := r.makeKey(repoID, path)
	_, err = collection.RemoveDocument(ctx, key)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete file index: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"path":    path,
		"repo_id": repoID,
	}).Debug("Deleted file index")

	return nil
}

// DeleteDirectory deletes all entries under a directory path
func (r *repository) DeleteDirectory(ctx context.Context, repoID, path string) error {
	// Normalize path
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	query := `
		FOR file IN @@collection
			FILTER file.repo_id == @repo_id
			FILTER STARTS_WITH(file.path, @path_prefix)
			REMOVE file IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": FileIndexCollection,
		"repo_id":     repoID,
		"path_prefix": path,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}
	defer cursor.Close()

	r.logger.WithFields(logrus.Fields{
		"path":    path,
		"repo_id": repoID,
	}).Info("Deleted directory and contents")

	return nil
}

// UpdateIndex updates an existing file index
func (r *repository) UpdateIndex(ctx context.Context, index *FileIndex) error {
	collection, err := r.db.Collection(ctx, FileIndexCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	index.Key = r.makeKey(index.RepoID, index.Path)
	_, err = collection.UpdateDocument(ctx, index.Key, index)
	if err != nil {
		return fmt.Errorf("failed to update file index: %w", err)
	}

	return nil
}

// RebuildIndex rebuilds the entire file index from a commit tree
func (r *repository) RebuildIndex(ctx context.Context, repoID, commitSHA string, entries []*FileIndex) error {
	// Delete all existing entries for this repo
	query := `
		FOR file IN @@collection
			FILTER file.repo_id == @repo_id
			REMOVE file IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": FileIndexCollection,
		"repo_id":     repoID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("failed to clear index: %w", err)
	}
	cursor.Close()

	// Insert all new entries
	for _, entry := range entries {
		if err := r.IndexFile(ctx, entry); err != nil {
			return fmt.Errorf("failed to index file %s: %w", entry.Path, err)
		}
	}

	r.logger.WithFields(logrus.Fields{
		"repo_id":    repoID,
		"commit_sha": commitSHA,
		"count":      len(entries),
	}).Info("Rebuilt file index")

	return nil
}

// makeKey creates a unique key for file index entry
func (r *repository) makeKey(repoID, path string) string {
	// Normalize path separators and remove leading slash
	normalized := filepath.Clean(path)
	normalized = strings.TrimPrefix(normalized, "/")
	normalized = strings.ReplaceAll(normalized, "/", "_")

	if normalized == "." || normalized == "" {
		normalized = "root"
	}

	return fmt.Sprintf("%s_%s", repoID, normalized)
}
