package ops

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/git/models"
	"github.com/aosanya/CodeValdCortex/internal/git/storage"
	"github.com/sirupsen/logrus"
)

// GitOps provides high-level Git operations
type GitOps interface {
	// Blob operations
	WriteBlob(ctx context.Context, repoID, content string) (string, error)
	GetBlob(ctx context.Context, sha string) (string, error)

	// Tree operations
	WriteTree(ctx context.Context, repoID string, entries []models.TreeEntry) (string, error)
	GetTree(ctx context.Context, sha string) (*models.GitTree, error)

	// Commit operations
	Commit(ctx context.Context, repoID, treeSHA string, parents []string, author, message string) (string, error)
	GetCommit(ctx context.Context, sha string) (*models.GitCommit, error)

	// Ref operations
	UpdateRef(ctx context.Context, repoID, refName, commitSHA string) error
	GetRef(ctx context.Context, repoID, refName string) (*models.GitRef, error)
	CreateBranch(ctx context.Context, repoID, branchName, fromCommit string) error

	// Repository operations
	InitRepository(ctx context.Context, instanceID, name string) error
	GetRepository(ctx context.Context, instanceID string) (*models.Repository, error)
}

// gitOps implements GitOps interface
type gitOps struct {
	storage storage.Repository
	logger  *logrus.Logger
}

// NewGitOps creates a new GitOps instance
func NewGitOps(storage storage.Repository, logger *logrus.Logger) GitOps {
	return &gitOps{
		storage: storage,
		logger:  logger,
	}
}

// WriteBlob stores file content as a Git blob
func (g *gitOps) WriteBlob(ctx context.Context, repoID, content string) (string, error) {
	// Calculate SHA-1 hash (Git blob format: "blob <size>\0<content>")
	header := fmt.Sprintf("blob %d\x00", len(content))
	hash := sha1.Sum([]byte(header + content))
	sha := hex.EncodeToString(hash[:])

	// Create Git object
	obj := &models.GitObject{
		Key:     sha,
		Type:    "blob",
		Size:    len(content),
		Content: content,
		RepoID:  repoID,
		Created: time.Now(),
	}

	// Store object (idempotent - same content = same SHA)
	if err := g.storage.StoreObject(ctx, obj); err != nil {
		return "", fmt.Errorf("failed to store blob: %w", err)
	}

	return sha, nil
}

// GetBlob retrieves blob content
func (g *gitOps) GetBlob(ctx context.Context, sha string) (string, error) {
	obj, err := g.storage.GetObject(ctx, sha)
	if err != nil {
		return "", fmt.Errorf("failed to get blob: %w", err)
	}

	if obj.Type != "blob" {
		return "", fmt.Errorf("object %s is not a blob (type: %s)", sha, obj.Type)
	}

	return obj.Content, nil
}

// WriteTree stores directory structure as a Git tree
func (g *gitOps) WriteTree(ctx context.Context, repoID string, entries []models.TreeEntry) (string, error) {
	// Serialize tree (Git tree format)
	var buf bytes.Buffer
	for _, e := range entries {
		fmt.Fprintf(&buf, "%s %s %s\t%s\n", e.Mode, e.Type, e.SHA, e.Path)
	}
	content := buf.String()

	// Calculate SHA-1 hash
	header := fmt.Sprintf("tree %d\x00", len(content))
	hash := sha1.Sum([]byte(header + content))
	sha := hex.EncodeToString(hash[:])

	// Create Git object
	obj := &models.GitObject{
		Key:     sha,
		Type:    "tree",
		Size:    len(content),
		Content: content,
		RepoID:  repoID,
		Created: time.Now(),
	}

	// Store object
	if err := g.storage.StoreObject(ctx, obj); err != nil {
		return "", fmt.Errorf("failed to store tree: %w", err)
	}

		return sha, nil
}

// GetTree retrieves tree structure
func (g *gitOps) GetTree(ctx context.Context, sha string) (*models.GitTree, error) {
	obj, err := g.storage.GetObject(ctx, sha)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	if obj.Type != "tree" {
		return nil, fmt.Errorf("object %s is not a tree (type: %s)", sha, obj.Type)
	}

	// Parse tree content
	tree := &models.GitTree{
		SHA:     sha,
		Entries: []models.TreeEntry{},
	}

	lines := strings.Split(strings.TrimSpace(obj.Content), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		tree.Entries = append(tree.Entries, models.TreeEntry{
			Mode: parts[0],
			Type: parts[1],
			SHA:  parts[2],
			Path: parts[3],
		})
	}

	return tree, nil
}

// Commit creates a version snapshot
func (g *gitOps) Commit(ctx context.Context, repoID, treeSHA string, parents []string, authorName, message string) (string, error) {
	now := time.Now()

	commit := &models.GitCommit{
		Tree:    treeSHA,
		Parents: parents,
		Author: models.Author{
			Name:  authorName,
			Email: authorName + "@codevaldcortex",
			When:  now,
		},
		Committer: models.Author{
			Name:  "codevaldcortex",
			Email: "system@codevaldcortex",
			When:  now,
		},
		Message:   message,
		Timestamp: now,
	}

	// Serialize commit (Git commit format)
	content := serializeCommit(commit)

	// Calculate SHA-1 hash
	header := fmt.Sprintf("commit %d\x00", len(content))
	hash := sha1.Sum([]byte(header + content))
	sha := hex.EncodeToString(hash[:])

	// Store commit object
	obj := &models.GitObject{
		Key:     sha,
		Type:    "commit",
		Size:    len(content),
		Content: content,
		RepoID:  repoID,
		Created: now,
	}

	if err := g.storage.StoreObject(ctx, obj); err != nil {
		return "", fmt.Errorf("failed to store commit: %w", err)
	}

		return sha, nil
}

// GetCommit retrieves commit object
func (g *gitOps) GetCommit(ctx context.Context, sha string) (*models.GitCommit, error) {
	obj, err := g.storage.GetObject(ctx, sha)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	if obj.Type != "commit" {
		return nil, fmt.Errorf("object %s is not a commit (type: %s)", sha, obj.Type)
	}

	commit := deserializeCommit(obj.Content)
	commit.SHA = sha

	return commit, nil
}

// UpdateRef moves a branch pointer to a new commit
func (g *gitOps) UpdateRef(ctx context.Context, repoID, refName, commitSHA string) error {
	// Try to get existing ref
	ref, err := g.storage.GetRef(ctx, repoID, refName)
	if err != nil {
		// Ref doesn't exist, create it
		// Create a safe document key by replacing slashes with underscores
		safeKey := strings.ReplaceAll(refName, "/", "_")

		ref = &models.GitRef{
			Key:    safeKey,
			RepoID: repoID,
			Name:   refName,
			Type:   "branch",
			Target: commitSHA,
		}

		if err := g.storage.StoreRef(ctx, ref); err != nil {
			return fmt.Errorf("failed to create ref: %w", err)
		}

				return nil
	}

	// Update existing ref
	ref.Target = commitSHA
	if err := g.storage.UpdateRef(ctx, ref); err != nil {
		return fmt.Errorf("failed to update ref: %w", err)
	}

		return nil
}

// GetRef retrieves a Git reference
func (g *gitOps) GetRef(ctx context.Context, repoID, refName string) (*models.GitRef, error) {
	return g.storage.GetRef(ctx, repoID, refName)
}

// CreateBranch creates a new branch from a commit
func (g *gitOps) CreateBranch(ctx context.Context, repoID, branchName, fromCommit string) error {
	ref := &models.GitRef{
		RepoID: repoID,
		Type:   "branch",
		Target: fromCommit,
	}

	if err := g.storage.StoreRef(ctx, ref); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	g.logger.WithFields(logrus.Fields{
		"branch":  branchName,
		"from":    fromCommit,
		"repo_id": repoID,
	}).Info("Created branch")

	return nil
}

// InitRepository initializes a new Git repository
func (g *gitOps) InitRepository(ctx context.Context, instanceID, name string) error {
	// Create empty tree (no files yet)
	emptyTreeSHA, err := g.WriteTree(ctx, instanceID, []models.TreeEntry{})
	if err != nil {
		return fmt.Errorf("failed to create empty tree: %w", err)
	}

	// Create initial commit
	initialCommitSHA, err := g.Commit(
		ctx,
		instanceID,
		emptyTreeSHA,
		[]string{}, // No parents
		"system",
		"Initial commit",
	)
	if err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	// Create main branch pointing to initial commit
	if err := g.UpdateRef(ctx, instanceID, "refs/heads/main", initialCommitSHA); err != nil {
		return fmt.Errorf("failed to create main branch: %w", err)
	}

	// Create repository record
	repo := &models.Repository{
		InstanceID:  instanceID,
		Name:        name,
		Description: "Git repository for " + name,
		DefaultRef:  "refs/heads/main",
		RootPath:    "/",
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	if err := g.storage.CreateRepository(ctx, repo); err != nil {
		return fmt.Errorf("failed to create repository record: %w", err)
	}

	g.logger.WithFields(logrus.Fields{
		"instance_id":    instanceID,
		"name":           name,
		"initial_commit": initialCommitSHA,
	}).Info("Initialized Git repository")

	return nil
}

// GetRepository retrieves repository metadata
func (g *gitOps) GetRepository(ctx context.Context, instanceID string) (*models.Repository, error) {
	return g.storage.GetRepository(ctx, instanceID)
}

// serializeCommit converts a GitCommit to Git format
func serializeCommit(commit *models.GitCommit) string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "tree %s\n", commit.Tree)

	for _, parent := range commit.Parents {
		fmt.Fprintf(&buf, "parent %s\n", parent)
	}

	fmt.Fprintf(&buf, "author %s <%s> %d +0000\n",
		commit.Author.Name,
		commit.Author.Email,
		commit.Author.When.Unix())

	fmt.Fprintf(&buf, "committer %s <%s> %d +0000\n",
		commit.Committer.Name,
		commit.Committer.Email,
		commit.Committer.When.Unix())

	fmt.Fprintf(&buf, "\n%s\n", commit.Message)

	return buf.String()
}

// deserializeCommit parses a Git commit from serialized format
func deserializeCommit(content string) *models.GitCommit {
	commit := &models.GitCommit{
		Parents: []string{},
	}

	lines := strings.Split(content, "\n")
	var messageLines []string
	inMessage := false

	for _, line := range lines {
		if inMessage {
			messageLines = append(messageLines, line)
			continue
		}

		if line == "" {
			inMessage = true
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "tree":
			commit.Tree = parts[1]
		case "parent":
			commit.Parents = append(commit.Parents, parts[1])
		case "author":
			commit.Author = parseAuthor(line)
		case "committer":
			commit.Committer = parseAuthor(line)
		}
	}

	commit.Message = strings.Join(messageLines, "\n")

	return commit
}

// parseAuthor parses author line from Git commit
func parseAuthor(line string) models.Author {
	// Format: "author Name <email> timestamp +0000"
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 2 {
		return models.Author{}
	}

	remainder := parts[1]
	emailStart := strings.Index(remainder, "<")
	emailEnd := strings.Index(remainder, ">")

	if emailStart == -1 || emailEnd == -1 {
		return models.Author{}
	}

	name := strings.TrimSpace(remainder[:emailStart])
	email := remainder[emailStart+1 : emailEnd]

	return models.Author{
		Name:  name,
		Email: email,
		When:  time.Now(), // Simplified - would need to parse timestamp
	}
}
