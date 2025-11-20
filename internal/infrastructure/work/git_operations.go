package work

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
)

// gitOperationsImpl implements GitOperations interface
type gitOperationsImpl struct {
	client *gitea.Client
}

// NewGitOperations creates a new Git operations handler
func NewGitOperations(client *gitea.Client) GitOperations {
	return &gitOperationsImpl{
		client: client,
	}
}

// CreateBranch creates a new branch from a base branch
func (g *gitOperationsImpl) CreateBranch(ctx context.Context, repo, baseBranch, newBranch string) error {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s (expected owner/repo)", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Create branch using Gitea API
	_, _, err := g.client.CreateBranch(owner, repoName, gitea.CreateBranchOption{
		BranchName:    newBranch,
		OldBranchName: baseBranch,
	})
	if err != nil {
		return fmt.Errorf("failed to create branch %s from %s: %w", newBranch, baseBranch, err)
	}

	return nil
}

// PushChanges pushes a set of file changes to a branch
func (g *gitOperationsImpl) PushChanges(ctx context.Context, repo, branch string, changes *ChangeSet) error {
	if changes == nil || len(changes.Files) == 0 {
		return fmt.Errorf("no changes to push")
	}

	// Process each file change
	for _, file := range changes.Files {
		var err error
		switch file.Operation {
		case "create":
			err = g.CreateFile(ctx, repo, branch, file.Path, file.Content)
		case "update":
			err = g.UpdateFile(ctx, repo, branch, file.Path, file.Content)
		case "delete":
			err = g.DeleteFile(ctx, repo, branch, file.Path)
		default:
			return fmt.Errorf("unsupported operation: %s", file.Operation)
		}

		if err != nil {
			return fmt.Errorf("failed to %s file %s: %w", file.Operation, file.Path, err)
		}
	}

	return nil
}

// DeleteBranch deletes a branch
func (g *gitOperationsImpl) DeleteBranch(ctx context.Context, repo, branch string) error {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	_, _, err := g.client.DeleteRepoBranch(owner, repoName, branch)
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", branch, err)
	}

	return nil
}

// CreateFile creates a new file in the repository
func (g *gitOperationsImpl) CreateFile(ctx context.Context, repo, branch, path, content string) error {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Encode content as base64
	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))

	_, _, err := g.client.CreateFile(owner, repoName, path, gitea.CreateFileOptions{
		FileOptions: gitea.FileOptions{
			Message:       fmt.Sprintf("Create %s", path),
			BranchName:    branch,
			NewBranchName: "",
		},
		Content: encodedContent,
	})
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}

	return nil
}

// UpdateFile updates an existing file in the repository
func (g *gitOperationsImpl) UpdateFile(ctx context.Context, repo, branch, path, content string) error {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Get current file to obtain SHA
	fileContents, _, err := g.client.GetContents(owner, repoName, branch, path)
	if err != nil {
		return fmt.Errorf("failed to get file contents for %s: %w", path, err)
	}

	// Encode new content as base64
	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))

	_, _, err = g.client.UpdateFile(owner, repoName, path, gitea.UpdateFileOptions{
		FileOptions: gitea.FileOptions{
			Message:    fmt.Sprintf("Update %s", path),
			BranchName: branch,
		},
		Content: encodedContent,
		SHA:     fileContents.SHA,
	})
	if err != nil {
		return fmt.Errorf("failed to update file %s: %w", path, err)
	}

	return nil
}

// DeleteFile deletes a file from the repository
func (g *gitOperationsImpl) DeleteFile(ctx context.Context, repo, branch, path string) error {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Get current file to obtain SHA
	fileContents, _, err := g.client.GetContents(owner, repoName, branch, path)
	if err != nil {
		return fmt.Errorf("failed to get file contents for %s: %w", path, err)
	}

	_, err = g.client.DeleteFile(owner, repoName, path, gitea.DeleteFileOptions{
		FileOptions: gitea.FileOptions{
			Message:    fmt.Sprintf("Delete %s", path),
			BranchName: branch,
		},
		SHA: fileContents.SHA,
	})
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}

	return nil
}

// CreateCommit creates a commit with multiple file changes
func (g *gitOperationsImpl) CreateCommit(ctx context.Context, repo, branch, message string, files []FileChange) error {
	// Gitea SDK doesn't have a direct CreateCommit method, so we'll use PushChanges
	changes := &ChangeSet{Files: files}
	return g.PushChanges(ctx, repo, branch, changes)
}
