package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/git/models"
	driver "github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
)

const (
	GitObjectsCollection   = "git_objects"
	GitRefsCollection      = "git_refs"
	RepositoriesCollection = "repositories"
)

// Repository handles Git object storage in ArangoDB
type Repository interface {
	// Object operations
	StoreObject(ctx context.Context, obj *models.GitObject) error
	GetObject(ctx context.Context, sha string) (*models.GitObject, error)
	ObjectExists(ctx context.Context, sha string) (bool, error)

	// Ref operations
	StoreRef(ctx context.Context, ref *models.GitRef) error
	GetRef(ctx context.Context, repoID, refName string) (*models.GitRef, error)
	UpdateRef(ctx context.Context, ref *models.GitRef) error
	ListRefs(ctx context.Context, repoID string, refType string) ([]*models.GitRef, error)

	// Repository operations
	CreateRepository(ctx context.Context, repo *models.Repository) error
	GetRepository(ctx context.Context, instanceID string) (*models.Repository, error)
}

// repository implements Repository interface
type repository struct {
	db     driver.Database
	logger *logrus.Logger
}

// NewRepository creates a new Git repository storage
func NewRepository(db driver.Database, logger *logrus.Logger) (Repository, error) {
	r := &repository{
		db:     db,
		logger: logger,
	}

	// Ensure collections exist
	if err := r.ensureCollections(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure collections: %w", err)
	}

	return r, nil
}

// ensureCollections creates Git collections if they don't exist
func (r *repository) ensureCollections(ctx context.Context) error {
	collections := []string{
		GitObjectsCollection,
		GitRefsCollection,
		RepositoriesCollection,
	}

	for _, collName := range collections {
		exists, err := r.db.CollectionExists(ctx, collName)
		if err != nil {
			return fmt.Errorf("failed to check collection %s: %w", collName, err)
		}

		if !exists {
			_, err = r.db.CreateCollection(ctx, collName, nil)
			if err != nil {
				return fmt.Errorf("failed to create collection %s: %w", collName, err)
			}
			r.logger.WithField("collection", collName).Info("Created Git collection")
		}
	}

	return nil
}

// StoreObject stores a Git object in the database
func (r *repository) StoreObject(ctx context.Context, obj *models.GitObject) error {
	collection, err := r.db.Collection(ctx, GitObjectsCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Key should already be set to SHA by caller (ops layer)
	// Content field contains actual file content (plain text for text files)

	// Try to create document (idempotent - same content = same SHA)
	_, err = collection.CreateDocument(ctx, obj)
	if err != nil {
		// Check if it's a duplicate key error (which is expected for same content)
		if driver.IsConflict(err) {
			r.logger.WithField("sha", obj.Key).Debug("Git object already exists (idempotent)")
			return nil // Not an error - object already stored
		}
		return fmt.Errorf("failed to create git object: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"sha":     obj.Key,
		"type":    obj.Type,
		"repo_id": obj.RepoID,
	}).Debug("Stored Git object")

	return nil
}

// GetObject retrieves a Git object by SHA
func (r *repository) GetObject(ctx context.Context, sha string) (*models.GitObject, error) {
	collection, err := r.db.Collection(ctx, GitObjectsCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	var obj models.GitObject
	_, err = collection.ReadDocument(ctx, sha, &obj)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, fmt.Errorf("git object not found: %s", sha)
		}
		return nil, fmt.Errorf("failed to read git object: %w", err)
	}

	return &obj, nil
}

// ObjectExists checks if a Git object exists
func (r *repository) ObjectExists(ctx context.Context, sha string) (bool, error) {
	collection, err := r.db.Collection(ctx, GitObjectsCollection)
	if err != nil {
		return false, fmt.Errorf("failed to get collection: %w", err)
	}

	exists, err := collection.DocumentExists(ctx, sha)
	if err != nil {
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return exists, nil
}

// StoreRef stores a Git reference
func (r *repository) StoreRef(ctx context.Context, ref *models.GitRef) error {
	collection, err := r.db.Collection(ctx, GitRefsCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// If Key is not already set, generate a safe one
	if ref.Key == "" {
		// Create safe key by replacing slashes and combining repo+name
		safeName := strings.ReplaceAll(ref.Name, "/", "_")
		safeRepoID := strings.ReplaceAll(ref.RepoID, "-", "_")
		ref.Key = safeRepoID + "_" + safeName
	}

	_, err = collection.CreateDocument(ctx, ref)
	if err != nil {
		return fmt.Errorf("failed to create ref: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"key":     ref.Key,
		"name":    ref.Name,
		"type":    ref.Type,
		"target":  ref.Target,
		"repo_id": ref.RepoID,
	}).Debug("Stored Git ref")

	return nil
}

// GetRef retrieves a Git reference
func (r *repository) GetRef(ctx context.Context, repoID, refName string) (*models.GitRef, error) {
	// Query by repoID and refName
	query := `
		FOR ref IN @@collection
			FILTER ref.repo_id == @repo_id
			FILTER ref.name == @ref_name
			LIMIT 1
			RETURN ref
	`

	bindVars := map[string]interface{}{
		"@collection": GitRefsCollection,
		"repo_id":     repoID,
		"ref_name":    refName,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query ref: %w", err)
	}
	defer cursor.Close()

	var ref models.GitRef
	_, err = cursor.ReadDocument(ctx, &ref)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			return nil, fmt.Errorf("ref not found: %s", refName)
		}
		return nil, fmt.Errorf("failed to read ref: %w", err)
	}

	return &ref, nil
}

// UpdateRef updates an existing Git reference
func (r *repository) UpdateRef(ctx context.Context, ref *models.GitRef) error {
	collection, err := r.db.Collection(ctx, GitRefsCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	_, err = collection.UpdateDocument(ctx, ref.Key, ref)
	if err != nil {
		return fmt.Errorf("failed to update ref: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"ref":    ref.Key,
		"target": ref.Target,
	}).Debug("Updated Git ref")

	return nil
}

// ListRefs lists all refs for a repository
func (r *repository) ListRefs(ctx context.Context, repoID string, refType string) ([]*models.GitRef, error) {
	query := `
		FOR ref IN @@collection
			FILTER ref.repo_id == @repo_id
			FILTER @ref_type == "" OR ref.type == @ref_type
			RETURN ref
	`

	bindVars := map[string]interface{}{
		"@collection": GitRefsCollection,
		"repo_id":     repoID,
		"ref_type":    refType,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query refs: %w", err)
	}
	defer cursor.Close()

	var refs []*models.GitRef
	for cursor.HasMore() {
		var ref models.GitRef
		_, err := cursor.ReadDocument(ctx, &ref)
		if err != nil {
			return nil, fmt.Errorf("failed to read ref: %w", err)
		}
		refs = append(refs, &ref)
	}

	return refs, nil
}

// CreateRepository creates a new repository record
func (r *repository) CreateRepository(ctx context.Context, repo *models.Repository) error {
	collection, err := r.db.Collection(ctx, RepositoriesCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Use instance ID as key (one repo per instance)
	repo.Key = repo.InstanceID

	_, err = collection.CreateDocument(ctx, repo)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"instance_id": repo.InstanceID,
		"name":        repo.Name,
	}).Info("Created Git repository")

	return nil
}

// GetRepository retrieves a repository by instance ID
func (r *repository) GetRepository(ctx context.Context, instanceID string) (*models.Repository, error) {
	collection, err := r.db.Collection(ctx, RepositoriesCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	var repo models.Repository
	_, err = collection.ReadDocument(ctx, instanceID, &repo)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, fmt.Errorf("repository not found for instance: %s", instanceID)
		}
		return nil, fmt.Errorf("failed to read repository: %w", err)
	}

	return &repo, nil
}
