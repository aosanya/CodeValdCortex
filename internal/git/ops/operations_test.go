package ops

import (
	"context"
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/git/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of storage.Repository
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) StoreObject(ctx context.Context, obj *models.GitObject) error {
	args := m.Called(ctx, obj)
	return args.Error(0)
}

func (m *MockStorage) GetObject(ctx context.Context, sha string) (*models.GitObject, error) {
	args := m.Called(ctx, sha)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GitObject), args.Error(1)
}

func (m *MockStorage) ObjectExists(ctx context.Context, sha string) (bool, error) {
	args := m.Called(ctx, sha)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) StoreRef(ctx context.Context, ref *models.GitRef) error {
	args := m.Called(ctx, ref)
	return args.Error(0)
}

func (m *MockStorage) GetRef(ctx context.Context, repoID, refName string) (*models.GitRef, error) {
	args := m.Called(ctx, repoID, refName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GitRef), args.Error(1)
}

func (m *MockStorage) UpdateRef(ctx context.Context, ref *models.GitRef) error {
	args := m.Called(ctx, ref)
	return args.Error(0)
}

func (m *MockStorage) ListRefs(ctx context.Context, repoID string, refType string) ([]*models.GitRef, error) {
	args := m.Called(ctx, repoID, refType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GitRef), args.Error(1)
}

func (m *MockStorage) CreateRepository(ctx context.Context, repo *models.Repository) error {
	args := m.Called(ctx, repo)
	return args.Error(0)
}

func (m *MockStorage) GetRepository(ctx context.Context, instanceID string) (*models.Repository, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Repository), args.Error(1)
}

func TestWriteBlob_PlainText(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	gitOps := NewGitOps(mockStorage, logger)
	ctx := context.Background()

	// Test data - plain text content
	repoID := "test-repo-001"
	content := "# My Document\n\nThis is plain text content for a text file.\n"

	// Mock storage to capture the object
	var capturedObj *models.GitObject
	mockStorage.On("StoreObject", ctx, mock.MatchedBy(func(obj *models.GitObject) bool {
		capturedObj = obj
		return true
	})).Return(nil)

	// Write blob
	sha, err := gitOps.WriteBlob(ctx, repoID, content)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, sha)
	assert.Len(t, sha, 40) // SHA-1 is 40 hex characters

	// Verify the stored object
	assert.NotNil(t, capturedObj)
	assert.Equal(t, "blob", capturedObj.Type)
	assert.Equal(t, content, capturedObj.Content) // ✅ Content stored as plain text
	assert.Equal(t, len(content), capturedObj.Size)
	assert.Equal(t, repoID, capturedObj.RepoID)
	assert.Equal(t, sha, capturedObj.Key)

	mockStorage.AssertExpectations(t)
}

func TestWriteTree(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	gitOps := NewGitOps(mockStorage, logger)
	ctx := context.Background()

	repoID := "test-repo-001"
	entries := []models.TreeEntry{
		{Mode: "100644", Type: "blob", SHA: "abc123", Path: "README.md"},
		{Mode: "040000", Type: "tree", SHA: "def456", Path: "docs"},
	}

	var capturedObj *models.GitObject
	mockStorage.On("StoreObject", ctx, mock.MatchedBy(func(obj *models.GitObject) bool {
		capturedObj = obj
		return true
	})).Return(nil)

	sha, err := gitOps.WriteTree(ctx, repoID, entries)

	assert.NoError(t, err)
	assert.NotEmpty(t, sha)
	assert.Len(t, sha, 40)

	assert.Equal(t, "tree", capturedObj.Type)
	assert.Contains(t, capturedObj.Content, "README.md")
	assert.Contains(t, capturedObj.Content, "docs")

	mockStorage.AssertExpectations(t)
}

func TestCommit(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	gitOps := NewGitOps(mockStorage, logger)
	ctx := context.Background()

	repoID := "test-repo-001"
	treeSHA := "tree123456"
	parents := []string{"parent123"}
	author := "user-alice"
	message := "Initial commit"

	var capturedObj *models.GitObject
	mockStorage.On("StoreObject", ctx, mock.MatchedBy(func(obj *models.GitObject) bool {
		capturedObj = obj
		return true
	})).Return(nil)

	sha, err := gitOps.Commit(ctx, repoID, treeSHA, parents, author, message)

	assert.NoError(t, err)
	assert.NotEmpty(t, sha)
	assert.Len(t, sha, 40)

	assert.Equal(t, "commit", capturedObj.Type)
	assert.Contains(t, capturedObj.Content, treeSHA)
	assert.Contains(t, capturedObj.Content, "parent123")
	assert.Contains(t, capturedObj.Content, author)
	assert.Contains(t, capturedObj.Content, message)

	mockStorage.AssertExpectations(t)
}

func TestGetBlob_RetrievesPlainText(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	gitOps := NewGitOps(mockStorage, logger)
	ctx := context.Background()

	// Original plain text content
	expectedContent := "# Documentation\n\nPlain text stored as-is.\n"
	sha := "test-sha-123"

	mockStorage.On("GetObject", ctx, sha).Return(&models.GitObject{
		Key:     sha,
		Type:    "blob",
		Content: expectedContent, // Stored as plain text
		Size:    len(expectedContent),
		RepoID:  "test-repo",
		Created: time.Now(),
	}, nil)

	// Retrieve blob
	content, err := gitOps.GetBlob(ctx, sha)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, content) // ✅ Retrieved as plain text

	mockStorage.AssertExpectations(t)
}

func TestInitRepository(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	gitOps := NewGitOps(mockStorage, logger)
	ctx := context.Background()

	instanceID := "instance-001"
	name := "Test Repository"

	// Mock expectations
	mockStorage.On("StoreObject", ctx, mock.AnythingOfType("*models.GitObject")).Return(nil).Times(2) // tree + commit
	mockStorage.On("GetRef", ctx, instanceID, "refs/heads/main").Return(nil, assert.AnError)
	mockStorage.On("StoreRef", ctx, mock.AnythingOfType("*models.GitRef")).Return(nil)
	mockStorage.On("CreateRepository", ctx, mock.MatchedBy(func(repo *models.Repository) bool {
		return repo.InstanceID == instanceID && repo.Name == name
	})).Return(nil)

	err := gitOps.InitRepository(ctx, instanceID, name)

	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}
