package giteawebhook

import (
	"context"
	"time"

	"code.gitea.io/sdk/gitea"
)

// Client provides methods for interacting with Gitea API
// Enables bidirectional sync between agents and Gitea issues/PRs
type Client interface {
	// Issue Operations
	CreateIssue(ctx context.Context, owner, repo string, opts CreateIssueOptions) (*gitea.Issue, error)
	UpdateIssue(ctx context.Context, owner, repo string, index int64, opts UpdateIssueOptions) (*gitea.Issue, error)
	GetIssue(ctx context.Context, owner, repo string, index int64) (*gitea.Issue, error)
	ListIssues(ctx context.Context, owner, repo string, opts ListIssueOptions) ([]*gitea.Issue, error)
	CloseIssue(ctx context.Context, owner, repo string, index int64) error
	ReopenIssue(ctx context.Context, owner, repo string, index int64) error

	// Comment Operations
	PostComment(ctx context.Context, owner, repo string, index int64, body string) (*gitea.Comment, error)
	ListComments(ctx context.Context, owner, repo string, index int64) ([]*gitea.Comment, error)
	UpdateComment(ctx context.Context, owner, repo string, commentID int64, body string) (*gitea.Comment, error)
	DeleteComment(ctx context.Context, owner, repo string, commentID int64) error

	// Label Operations
	AddLabel(ctx context.Context, owner, repo string, index int64, labels []string) error
	RemoveLabel(ctx context.Context, owner, repo string, index int64, labelID int64) error
	ListLabels(ctx context.Context, owner, repo string, index int64) ([]*gitea.Label, error)

	// Pull Request Operations
	CreatePullRequest(ctx context.Context, owner, repo string, opts CreatePullRequestOptions) (*gitea.PullRequest, error)
	GetPullRequest(ctx context.Context, owner, repo string, index int64) (*gitea.PullRequest, error)
	ListPullRequests(ctx context.Context, owner, repo string, opts ListPullRequestOptions) ([]*gitea.PullRequest, error)
	MergePullRequest(ctx context.Context, owner, repo string, index int64, opts MergePullRequestOptions) error
	UpdatePullRequest(ctx context.Context, owner, repo string, index int64, opts UpdatePullRequestOptions) (*gitea.PullRequest, error)

	// Milestone Operations
	CreateMilestone(ctx context.Context, owner, repo string, opts CreateMilestoneOptions) (*gitea.Milestone, error)
	GetMilestone(ctx context.Context, owner, repo string, id int64) (*gitea.Milestone, error)
	ListMilestones(ctx context.Context, owner, repo string, opts ListMilestoneOptions) ([]*gitea.Milestone, error)
	UpdateMilestone(ctx context.Context, owner, repo string, id int64, opts UpdateMilestoneOptions) (*gitea.Milestone, error)

	// Repository Operations
	GetRepository(ctx context.Context, owner, repo string) (*gitea.Repository, error)
	ListRepositories(ctx context.Context, opts ListRepoOptions) ([]*gitea.Repository, error)
}

// CreateIssueOptions contains options for creating an issue
type CreateIssueOptions struct {
	Title     string
	Body      string
	Assignees []string
	Milestone int64
	Labels    []int64
	Closed    bool
}

// UpdateIssueOptions contains options for updating an issue
type UpdateIssueOptions struct {
	Title     *string
	Body      *string
	Assignees []string
	Milestone *int64
	State     *string // "open" or "closed"
}

// ListIssueOptions contains options for listing issues
type ListIssueOptions struct {
	State     string   // "open", "closed", "all"
	Labels    []string // Filter by labels
	Milestone string   // Filter by milestone name
	Since     *time.Time
	Page      int
	Limit     int
}

// CreatePullRequestOptions contains options for creating a pull request
type CreatePullRequestOptions struct {
	Title string
	Body  string
	Head  string // Branch name
	Base  string // Target branch (e.g., "main")
}

// UpdatePullRequestOptions contains options for updating a pull request
type UpdatePullRequestOptions struct {
	Title *string
	Body  *string
	State *string // "open" or "closed"
}

// MergePullRequestOptions contains options for merging a pull request
type MergePullRequestOptions struct {
	Style   string // "merge", "rebase", "squash"
	Message string
}

// ListPullRequestOptions contains options for listing pull requests
type ListPullRequestOptions struct {
	State string // "open", "closed", "all"
	Page  int
	Limit int
}

// CreateMilestoneOptions contains options for creating a milestone
type CreateMilestoneOptions struct {
	Title       string
	Description string
	DueDate     *time.Time
	State       string // "open" or "closed"
}

// UpdateMilestoneOptions contains options for updating a milestone
type UpdateMilestoneOptions struct {
	Title       *string
	Description *string
	DueDate     *time.Time
	State       *string // "open" or "closed"
}

// ListMilestoneOptions contains options for listing milestones
type ListMilestoneOptions struct {
	State string // "open", "closed", "all"
	Page  int
	Limit int
}

// ListRepoOptions contains options for listing repositories
type ListRepoOptions struct {
	Page  int
	Limit int
}
