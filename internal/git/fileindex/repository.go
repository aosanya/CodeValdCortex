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
	// GitArtifactsCollection stores file and directory metadata
	GitArtifactsCollection = "git_artifacts"
)

// Repository handles file index operations in ArangoDB
type Repository interface {
	// Index operations (all methods now accept agencyDB parameter)
	IndexFile(ctx context.Context, agencyDB string, index *FileIndex) error
	GetByPath(ctx context.Context, agencyDB string, repoID, path string) (*FileIndex, error)
	ListDirectory(ctx context.Context, agencyDB string, repoID, path string) ([]*FileIndex, error)
	DeleteByPath(ctx context.Context, agencyDB string, repoID, path string) error
	DeleteDirectory(ctx context.Context, agencyDB string, repoID, path string) error
	UpdateIndex(ctx context.Context, agencyDB string, index *FileIndex) error

	// Batch operations
	RebuildIndex(ctx context.Context, agencyDB string, repoID, commitSHA string, entries []*FileIndex) error
}

// repository implements Repository interface
type repository struct {
	client driver.Client // ArangoDB client to access different databases
	logger *logrus.Logger
}

// NewRepository creates a new file index repository
func NewRepository(client driver.Client, logger *logrus.Logger) (Repository, error) {
	r := &repository{
		client: client,
		logger: logger,
	}

	return r, nil
}

// ensureCollectionExists creates the git_artifacts collection if it doesn't exist
// Call this only when we get a "collection not found" error
func (r *repository) ensureCollectionExists(ctx context.Context, db driver.Database) error {
	_, err := db.CreateCollection(ctx, GitArtifactsCollection, nil)
	if err != nil {
		// Ignore error if collection already exists (race condition)
		if !driver.IsConflict(err) {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	} else {
		r.logger.WithField("collection", GitArtifactsCollection).Info("Created git artifacts collection")
	}
	return nil
}

// IndexFile creates or updates a file index entry
func (r *repository) IndexFile(ctx context.Context, agencyDB string, index *FileIndex) error {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to access database %s: %w", agencyDB, err)
	}

	collection, err := db.Collection(ctx, GitArtifactsCollection)
	if err != nil {
		// If collection doesn't exist, create it and retry
		if driver.IsNotFoundGeneral(err) {
			if createErr := r.ensureCollectionExists(ctx, db); createErr != nil {
				return createErr
			}
			collection, err = db.Collection(ctx, GitArtifactsCollection)
			if err != nil {
				return fmt.Errorf("failed to get collection after creation: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get collection: %w", err)
		}
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

// GetByPath retrieves a file index entry by path
func (r *repository) GetByPath(ctx context.Context, agencyDB string, repoID, path string) (*FileIndex, error) {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to access database %s: %w", agencyDB, err)
	}

	collection, err := db.Collection(ctx, GitArtifactsCollection)
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

// ListDirectory retrieves all entries in a directory
func (r *repository) ListDirectory(ctx context.Context, agencyDB string, repoID, path string) ([]*FileIndex, error) {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return nil, fmt.Errorf("failed to access database %s: %w", agencyDB, err)
	}

	// Normalize path for root directory
	if path == "" {
		path = "/"
	}

	query := `
		FOR doc IN @@collection
		FILTER doc.repo_id == @repo_id
		FILTER doc.parent_path == @parent_path
		FILTER doc.deleted_at == null
		SORT doc.is_dir DESC, doc.name ASC
		RETURN doc
	`

	bindVars := map[string]interface{}{
		"@collection": GitArtifactsCollection,
		"repo_id":     repoID,
		"parent_path": path,
	}

	cursor, err := db.Query(ctx, query, bindVars)
	if err != nil {
		// If collection doesn't exist, return empty list instead of error
		if driver.IsNotFoundGeneral(err) {
			return []*FileIndex{}, nil
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
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

// DeleteByPath soft-deletes a file index entry
func (r *repository) DeleteByPath(ctx context.Context, agencyDB string, repoID, path string) error {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to access database %s: %w", agencyDB, err)
	}

	collection, err := db.Collection(ctx, GitArtifactsCollection)
	if err != nil {
		// If collection doesn't exist, nothing to delete
		if driver.IsNotFoundGeneral(err) {
			return nil
		}
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

// DeleteDirectory soft-deletes a directory and all its contents recursively
func (r *repository) DeleteDirectory(ctx context.Context, agencyDB string, repoID, path string) error {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to access database %s: %w", agencyDB, err)
	}

	// Query to find the directory and all children recursively
	query := `
		FOR doc IN @@collection
		FILTER doc.repo_id == @repo_id
		FILTER (doc.path == @path OR STARTS_WITH(doc.path, CONCAT(@path, "/")))
		FILTER doc.deleted_at == null
		UPDATE doc WITH { deleted_at: DATE_ISO8601(DATE_NOW()) } IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": GitArtifactsCollection,
		"repo_id":     repoID,
		"path":        path,
	}

	cursor, err := db.Query(ctx, query, bindVars)
	if err != nil {
		// If collection doesn't exist, nothing to delete
		if driver.IsNotFoundGeneral(err) {
			return nil
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close()

	r.logger.WithFields(logrus.Fields{
		"path":    path,
		"repo_id": repoID,
	}).Info("Deleted directory and contents")

	return nil
}

// UpdateIndex updates an existing file index entry
func (r *repository) UpdateIndex(ctx context.Context, agencyDB string, index *FileIndex) error {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to access database %s: %w", agencyDB, err)
	}

	collection, err := db.Collection(ctx, GitArtifactsCollection)
	if err != nil {
		// If collection doesn't exist, create it and retry
		if driver.IsNotFoundGeneral(err) {
			if createErr := r.ensureCollectionExists(ctx, db); createErr != nil {
				return createErr
			}
			collection, err = db.Collection(ctx, GitArtifactsCollection)
			if err != nil {
				return fmt.Errorf("failed to get collection after creation: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get collection: %w", err)
		}
	}

	index.Key = r.makeKey(index.RepoID, index.Path)
	_, err = collection.UpdateDocument(ctx, index.Key, index)
	if err != nil {
		return fmt.Errorf("failed to update file index: %w", err)
	}

	return nil
}

// RebuildIndex rebuilds the entire file index from a commit tree
func (r *repository) RebuildIndex(ctx context.Context, agencyDB string, repoID, commitSHA string, entries []*FileIndex) error {
	db, err := r.client.Database(ctx, agencyDB)
	if err != nil {
		return fmt.Errorf("failed to access database %s: %w", agencyDB, err)
	}

	// Delete all existing entries for this repo (collection might not exist yet)
	query := `
		FOR file IN @@collection
			FILTER file.repo_id == @repo_id
			REMOVE file IN @@collection
	`

	bindVars := map[string]interface{}{
		"@collection": GitArtifactsCollection,
		"repo_id":     repoID,
	}

	cursor, err := db.Query(ctx, query, bindVars)
	if err != nil {
		// If collection doesn't exist, that's fine - we'll create it when adding entries
		if !driver.IsNotFoundGeneral(err) {
			return fmt.Errorf("failed to clear index: %w", err)
		}
	} else {
		cursor.Close()
	}

	// Insert all new entries
	for _, entry := range entries {
		if err := r.IndexFile(ctx, agencyDB, entry); err != nil {
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

	// Replace all non-alphanumeric characters except dash and underscore with underscore
	// ArangoDB keys can only contain: a-z, A-Z, 0-9, _, -
	normalized = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, normalized)

	if normalized == "." || normalized == "" {
		normalized = "root"
	}

	// Ensure the key starts with a letter or underscore (ArangoDB requirement)
	if len(normalized) > 0 && normalized[0] >= '0' && normalized[0] <= '9' {
		normalized = "_" + normalized
	}

	return fmt.Sprintf("%s_%s", repoID, normalized)
}
