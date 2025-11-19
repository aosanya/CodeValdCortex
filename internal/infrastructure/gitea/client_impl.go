package giteawebhook

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"code.gitea.io/sdk/gitea"
	"golang.org/x/time/rate"
)

// GiteaClient implements the Client interface
type GiteaClient struct {
	client      *gitea.Client
	rateLimiter *rate.Limiter
	baseURL     string
	token       string
}

// ClientConfig contains configuration for Gitea API client
type ClientConfig struct {
	BaseURL        string        // Gitea instance URL (e.g., "https://gitea.example.com")
	Token          string        // API token for authentication
	Timeout        time.Duration // HTTP client timeout
	RateLimit      int           // Max requests per second (0 = no limit)
	RateLimitBurst int           // Burst size for rate limiter
}

// NewClient creates a new Gitea API client
func NewClient(config ClientConfig) (*GiteaClient, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("API token is required")
	}

	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RateLimit == 0 {
		config.RateLimit = 10 // Default: 10 requests/second
	}
	if config.RateLimitBurst == 0 {
		config.RateLimitBurst = 20 // Default: burst of 20
	}

	// Create Gitea SDK client
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	giteaClient, err := gitea.NewClient(
		config.BaseURL,
		gitea.SetToken(config.Token),
		gitea.SetHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	// Create rate limiter
	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), config.RateLimitBurst)

	return &GiteaClient{
		client:      giteaClient,
		rateLimiter: limiter,
		baseURL:     config.BaseURL,
		token:       config.Token,
	}, nil
}

// waitForRateLimit blocks until rate limiter allows the request
func (c *GiteaClient) waitForRateLimit(ctx context.Context) error {
	return c.rateLimiter.Wait(ctx)
}

// CreateIssue creates a new issue in the specified repository
func (c *GiteaClient) CreateIssue(ctx context.Context, owner, repo string, opts CreateIssueOptions) (*gitea.Issue, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.CreateIssueOption{
		Title:     opts.Title,
		Body:      opts.Body,
		Assignees: opts.Assignees,
		Closed:    opts.Closed,
	}

	if opts.Milestone > 0 {
		giteaOpts.Milestone = opts.Milestone
	}

	if len(opts.Labels) > 0 {
		giteaOpts.Labels = opts.Labels
	}

	issue, _, err := c.client.CreateIssue(owner, repo, giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return issue, nil
}

// UpdateIssue updates an existing issue
func (c *GiteaClient) UpdateIssue(ctx context.Context, owner, repo string, index int64, opts UpdateIssueOptions) (*gitea.Issue, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.EditIssueOption{}

	// Only set fields that are provided (non-nil)
	if opts.Title != nil {
		giteaOpts.Title = *opts.Title
	}
	if opts.Body != nil {
		giteaOpts.Body = opts.Body // Body is *string in EditIssueOption
	}
	if opts.Assignees != nil {
		giteaOpts.Assignees = opts.Assignees
	}
	if opts.Milestone != nil {
		giteaOpts.Milestone = opts.Milestone
	}
	if opts.State != nil {
		state := gitea.StateType(*opts.State)
		giteaOpts.State = &state
	}

	issue, _, err := c.client.EditIssue(owner, repo, index, giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue %d: %w", index, err)
	}

	return issue, nil
}

// GetIssue retrieves an issue by its index
func (c *GiteaClient) GetIssue(ctx context.Context, owner, repo string, index int64) (*gitea.Issue, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	issue, _, err := c.client.GetIssue(owner, repo, index)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %d: %w", index, err)
	}

	return issue, nil
}

// ListIssues lists issues in a repository with optional filters
func (c *GiteaClient) ListIssues(ctx context.Context, owner, repo string, opts ListIssueOptions) ([]*gitea.Issue, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	giteaOpts := gitea.ListIssueOption{
		ListOptions: gitea.ListOptions{
			Page:     opts.Page,
			PageSize: opts.Limit,
		},
		State:  gitea.StateType(opts.State),
		Labels: opts.Labels,
	}

	if opts.Since != nil {
		giteaOpts.Since = *opts.Since
	}

	if giteaOpts.State == "" {
		giteaOpts.State = gitea.StateOpen
	}

	issues, _, err := c.client.ListRepoIssues(owner, repo, giteaOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	return issues, nil
}

// CloseIssue closes an issue
func (c *GiteaClient) CloseIssue(ctx context.Context, owner, repo string, index int64) error {
	if err := c.waitForRateLimit(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	state := gitea.StateClosed
	_, _, err := c.client.EditIssue(owner, repo, index, gitea.EditIssueOption{
		State: &state,
	})
	if err != nil {
		return fmt.Errorf("failed to close issue %d: %w", index, err)
	}

	return nil
}

// ReopenIssue reopens a closed issue
func (c *GiteaClient) ReopenIssue(ctx context.Context, owner, repo string, index int64) error {
	if err := c.waitForRateLimit(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	state := gitea.StateOpen
	_, _, err := c.client.EditIssue(owner, repo, index, gitea.EditIssueOption{
		State: &state,
	})
	if err != nil {
		return fmt.Errorf("failed to reopen issue %d: %w", index, err)
	}

	return nil
}

// PostComment posts a comment on an issue
func (c *GiteaClient) PostComment(ctx context.Context, owner, repo string, index int64, body string) (*gitea.Comment, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	comment, _, err := c.client.CreateIssueComment(owner, repo, index, gitea.CreateIssueCommentOption{
		Body: body,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to post comment on issue %d: %w", index, err)
	}

	return comment, nil
}

// ListComments lists all comments on an issue
func (c *GiteaClient) ListComments(ctx context.Context, owner, repo string, index int64) ([]*gitea.Comment, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	comments, _, err := c.client.ListIssueComments(owner, repo, index, gitea.ListIssueCommentOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list comments for issue %d: %w", index, err)
	}

	return comments, nil
}

// UpdateComment updates an existing comment
func (c *GiteaClient) UpdateComment(ctx context.Context, owner, repo string, commentID int64, body string) (*gitea.Comment, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	comment, _, err := c.client.EditIssueComment(owner, repo, commentID, gitea.EditIssueCommentOption{
		Body: body,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update comment %d: %w", commentID, err)
	}

	return comment, nil
}

// DeleteComment deletes a comment
func (c *GiteaClient) DeleteComment(ctx context.Context, owner, repo string, commentID int64) error {
	if err := c.waitForRateLimit(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	_, err := c.client.DeleteIssueComment(owner, repo, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment %d: %w", commentID, err)
	}

	return nil
}

// AddLabel adds labels to an issue
func (c *GiteaClient) AddLabel(ctx context.Context, owner, repo string, index int64, labels []string) error {
	if err := c.waitForRateLimit(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Convert label names to label IDs
	labelIDs := make([]int64, 0, len(labels))
	for _, labelName := range labels {
		// Get all repo labels to find IDs
		repoLabels, _, err := c.client.ListRepoLabels(owner, repo, gitea.ListLabelsOptions{})
		if err != nil {
			return fmt.Errorf("failed to list repo labels: %w", err)
		}

		for _, label := range repoLabels {
			if label.Name == labelName {
				labelIDs = append(labelIDs, label.ID)
				break
			}
		}
	}

	_, _, err := c.client.AddIssueLabels(owner, repo, index, gitea.IssueLabelsOption{
		Labels: labelIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to add labels to issue %d: %w", index, err)
	}

	return nil
}

// RemoveLabel removes a label from an issue
func (c *GiteaClient) RemoveLabel(ctx context.Context, owner, repo string, index int64, labelID int64) error {
	if err := c.waitForRateLimit(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	_, err := c.client.DeleteIssueLabel(owner, repo, index, labelID)
	if err != nil {
		return fmt.Errorf("failed to remove label %d from issue %d: %w", labelID, index, err)
	}

	return nil
}

// ListLabels lists all labels on an issue
func (c *GiteaClient) ListLabels(ctx context.Context, owner, repo string, index int64) ([]*gitea.Label, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	labels, _, err := c.client.GetIssueLabels(owner, repo, index, gitea.ListLabelsOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list labels for issue %d: %w", index, err)
	}

	return labels, nil
}
