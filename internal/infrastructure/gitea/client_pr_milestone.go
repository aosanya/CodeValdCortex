package giteawebhook

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
)

// CreatePullRequest creates a new pull request
func (c *GiteaClient) CreatePullRequest(ctx context.Context, owner, repo string, opts CreatePullRequestOptions) (*gitea.PullRequest, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.CreatePullRequestOption{
		Head:  opts.Head,
		Base:  opts.Base,
		Title: opts.Title,
		Body:  opts.Body,
	}

	pr, _, err := c.client.CreatePullRequest(owner, repo, giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return pr, nil
}

// GetPullRequest retrieves a pull request by its index
func (c *GiteaClient) GetPullRequest(ctx context.Context, owner, repo string, index int64) (*gitea.PullRequest, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	pr, _, err := c.client.GetPullRequest(owner, repo, index)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request %d: %w", index, err)
	}

	return pr, nil
}

// ListPullRequests lists pull requests in a repository
func (c *GiteaClient) ListPullRequests(ctx context.Context, owner, repo string, opts ListPullRequestOptions) ([]*gitea.PullRequest, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.ListPullRequestsOptions{
		ListOptions: gitea.ListOptions{
			Page:     opts.Page,
			PageSize: opts.Limit,
		},
		State: gitea.StateType(opts.State),
	}

	if giteaOpts.State == "" {
		giteaOpts.State = gitea.StateOpen
	}

	prs, _, err := c.client.ListRepoPullRequests(owner, repo, giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	return prs, nil
}

// MergePullRequest merges a pull request
func (c *GiteaClient) MergePullRequest(ctx context.Context, owner, repo string, index int64, opts MergePullRequestOptions) error {
	if err := c.waitForRateLimit(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Convert merge style string to Gitea merge style
	var style gitea.MergeStyle
	switch opts.Style {
	case "rebase":
		style = gitea.MergeStyleRebase
	case "squash":
		style = gitea.MergeStyleSquash
	case "merge", "":
		style = gitea.MergeStyleMerge
	default:
		return fmt.Errorf("invalid merge style: %s (must be merge, rebase, or squash)", opts.Style)
	}

	giteaOpts := gitea.MergePullRequestOption{
		Style:   style,
		Message: opts.Message,
	}

	_, _, err := c.client.MergePullRequest(owner, repo, index, giteaOpts)
	if err != nil {
		return fmt.Errorf("failed to merge pull request %d: %w", index, err)
	}

	return nil
}

// UpdatePullRequest updates an existing pull request
func (c *GiteaClient) UpdatePullRequest(ctx context.Context, owner, repo string, index int64, opts UpdatePullRequestOptions) (*gitea.PullRequest, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.EditPullRequestOption{}

	// Only set fields that are provided
	if opts.Title != nil {
		giteaOpts.Title = *opts.Title // Title is string in EditPullRequestOption
	}
	if opts.Body != nil {
		giteaOpts.Body = opts.Body // Body is *string in EditPullRequestOption
	}
	if opts.State != nil {
		state := gitea.StateType(*opts.State)
		giteaOpts.State = &state
	}

	pr, _, err := c.client.EditPullRequest(owner, repo, index, giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to update pull request %d: %w", index, err)
	}

	return pr, nil
}

// CreateMilestone creates a new milestone
func (c *GiteaClient) CreateMilestone(ctx context.Context, owner, repo string, opts CreateMilestoneOptions) (*gitea.Milestone, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.CreateMilestoneOption{
		Title:       opts.Title,
		Description: opts.Description,
		State:       gitea.StateType(opts.State),
	}

	if opts.DueDate != nil {
		giteaOpts.Deadline = opts.DueDate
	}

	if giteaOpts.State == "" {
		giteaOpts.State = gitea.StateOpen
	}

	milestone, _, err := c.client.CreateMilestone(owner, repo, giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create milestone: %w", err)
	}

	return milestone, nil
}

// GetMilestone retrieves a milestone by its ID
func (c *GiteaClient) GetMilestone(ctx context.Context, owner, repo string, id int64) (*gitea.Milestone, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	milestone, _, err := c.client.GetMilestone(owner, repo, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone %d: %w", id, err)
	}

	return milestone, nil
}

// ListMilestones lists milestones in a repository
func (c *GiteaClient) ListMilestones(ctx context.Context, owner, repo string, opts ListMilestoneOptions) ([]*gitea.Milestone, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.ListMilestoneOption{
		ListOptions: gitea.ListOptions{
			Page:     opts.Page,
			PageSize: opts.Limit,
		},
		State: gitea.StateType(opts.State),
	}

	if giteaOpts.State == "" {
		giteaOpts.State = gitea.StateOpen
	}

	milestones, _, err := c.client.ListRepoMilestones(owner, repo, giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list milestones: %w", err)
	}

	return milestones, nil
}

// UpdateMilestone updates an existing milestone
func (c *GiteaClient) UpdateMilestone(ctx context.Context, owner, repo string, id int64, opts UpdateMilestoneOptions) (*gitea.Milestone, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.EditMilestoneOption{}

	// Only set fields that are provided
	if opts.Title != nil {
		giteaOpts.Title = *opts.Title // Title is string in EditMilestoneOption
	}
	if opts.Description != nil {
		giteaOpts.Description = opts.Description // Description is *string
	}
	if opts.DueDate != nil {
		giteaOpts.Deadline = opts.DueDate
	}
	if opts.State != nil {
		state := gitea.StateType(*opts.State)
		giteaOpts.State = &state
	}

	milestone, _, err := c.client.EditMilestone(owner, repo, id, giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to update milestone %d: %w", id, err)
	}

	return milestone, nil
}

// GetRepository retrieves repository information
func (c *GiteaClient) GetRepository(ctx context.Context, owner, repo string) (*gitea.Repository, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	repository, _, err := c.client.GetRepo(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, repo, err)
	}

	return repository, nil
}

// ListRepositories lists all repositories accessible to the authenticated user
func (c *GiteaClient) ListRepositories(ctx context.Context, opts ListRepoOptions) ([]*gitea.Repository, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{
			Page:     opts.Page,
			PageSize: opts.Limit,
		},
	}

	repos, _, err := c.client.ListMyRepos(giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return repos, nil
}
