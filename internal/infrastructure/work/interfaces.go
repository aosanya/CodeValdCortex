package work

import (
	"context"
	"net/http"
	"time"
)

// WorkTrackingProvider defines the contract for work tracking system integrations
// (Gitea, GitHub, GitLab, Jira, Linear, etc.)
type WorkTrackingProvider interface {
	// GetProviderName returns the name of the provider (e.g., "gitea", "github", "gitlab")
	GetProviderName() string

	// ValidateWebhookSignature verifies the authenticity of webhook requests
	ValidateWebhookSignature(payload []byte, signature string, secret string) error

	// HandleWebhook processes incoming webhook requests and returns normalized work items
	HandleWebhook(ctx context.Context, r *http.Request) (*WebhookResult, error)
}

// IssueHandler processes issue-related webhooks
type IssueHandler interface {
	// ParseIssueWebhook extracts issue data from webhook payload
	ParseIssueWebhook(payload []byte) (*WorkIssue, error)

	// ValidateIssueEvent checks if the issue event should be processed
	ValidateIssueEvent(issue *WorkIssue, action string) bool
}

// PullRequestHandler processes pull request/merge request webhooks
type PullRequestHandler interface {
	// ParsePRWebhook extracts PR data from webhook payload
	ParsePRWebhook(payload []byte) (*WorkPullRequest, error)

	// ValidatePREvent checks if the PR event should be processed
	ValidatePREvent(pr *WorkPullRequest, action string) bool
}

// MilestoneHandler processes milestone-related webhooks
type MilestoneHandler interface {
	// ParseMilestoneWebhook extracts milestone data from webhook payload
	ParseMilestoneWebhook(payload []byte) (*WorkMilestone, error)

	// ValidateMilestoneEvent checks if the milestone event should be processed
	ValidateMilestoneEvent(milestone *WorkMilestone, action string) bool
}

// WebhookValidator handles signature validation for different providers
type WebhookValidator interface {
	// ValidateSignature verifies the webhook signature
	ValidateSignature(payload []byte, signature string, secret string) error

	// GetSignatureHeader returns the header name used for signatures
	GetSignatureHeader() string
}

// Repository handles persistence to ArangoDB
type Repository interface {
	// SaveIssue persists a work issue to the work-issues collection
	SaveIssue(ctx context.Context, issue *WorkIssue) error

	// SavePullRequest persists a PR to the work-prs collection
	SavePullRequest(ctx context.Context, pr *WorkPullRequest) error

	// SaveMilestone persists a milestone to the work-milestones collection
	SaveMilestone(ctx context.Context, milestone *WorkMilestone) error

	// GetIssue retrieves an issue by provider and issue ID
	GetIssue(ctx context.Context, provider string, issueID string) (*WorkIssue, error)

	// GetPullRequest retrieves a PR by provider and PR ID
	GetPullRequest(ctx context.Context, provider string, prID string) (*WorkPullRequest, error)

	// GetMilestone retrieves a milestone by provider and milestone ID
	GetMilestone(ctx context.Context, provider string, milestoneID string) (*WorkMilestone, error)
}

// WebhookResult represents the normalized result of processing a webhook
type WebhookResult struct {
	// Provider identifies the source system
	Provider string

	// EventType indicates what kind of event occurred (issue, pr, milestone, etc.)
	EventType WorkEventType

	// Action describes what happened (created, updated, closed, etc.)
	Action string

	// Issue contains issue data if EventType is EventTypeIssue
	Issue *WorkIssue

	// PullRequest contains PR data if EventType is EventTypePullRequest
	PullRequest *WorkPullRequest

	// Milestone contains milestone data if EventType is EventTypeMilestone
	Milestone *WorkMilestone

	// ProcessedAt timestamp when the webhook was processed
	ProcessedAt time.Time
}

// WorkEventType represents the type of work item event
type WorkEventType string

const (
	EventTypeIssue       WorkEventType = "issue"
	EventTypePullRequest WorkEventType = "pull_request"
	EventTypeMilestone   WorkEventType = "milestone"
	EventTypeComment     WorkEventType = "comment"
	EventTypeUnknown     WorkEventType = "unknown"
)

// WebhookConfig holds configuration for webhook handlers
type WebhookConfig struct {
	// Secret is the webhook secret for signature validation
	Secret string

	// EnabledEvents lists which event types to process
	EnabledEvents []WorkEventType

	// ProviderSpecificConfig holds provider-specific settings
	ProviderSpecificConfig map[string]interface{}
}

// ============================================================================
// API Client Interfaces (Outbound Communication)
// ============================================================================

// APIClient provides methods for bidirectional sync with work tracking systems
// This is the provider-agnostic interface that abstracts Gitea, GitHub, GitLab, etc.
type APIClient interface {
	IssueClient
	PullRequestClient
	MilestoneClient
	CommentClient
	LabelClient
	RepositoryClient
}

// IssueClient handles issue operations
type IssueClient interface {
	// CreateIssue creates a new issue
	CreateIssue(ctx context.Context, owner, repo string, opts CreateIssueOptions) (*WorkIssue, error)

	// UpdateIssue updates an existing issue
	UpdateIssue(ctx context.Context, owner, repo string, issueID string, opts UpdateIssueOptions) (*WorkIssue, error)

	// GetIssue retrieves an issue by ID
	GetIssue(ctx context.Context, owner, repo string, issueID string) (*WorkIssue, error)

	// ListIssues lists issues with filtering
	ListIssues(ctx context.Context, owner, repo string, opts ListIssueOptions) ([]*WorkIssue, error)

	// CloseIssue closes an issue
	CloseIssue(ctx context.Context, owner, repo string, issueID string) error

	// ReopenIssue reopens a closed issue
	ReopenIssue(ctx context.Context, owner, repo string, issueID string) error
}

// PullRequestClient handles pull request operations
type PullRequestClient interface {
	// CreatePullRequest creates a new pull request
	CreatePullRequest(ctx context.Context, owner, repo string, opts CreatePullRequestOptions) (*WorkPullRequest, error)

	// UpdatePullRequest updates an existing pull request
	UpdatePullRequest(ctx context.Context, owner, repo string, prID string, opts UpdatePullRequestOptions) (*WorkPullRequest, error)

	// GetPullRequest retrieves a pull request by ID
	GetPullRequest(ctx context.Context, owner, repo string, prID string) (*WorkPullRequest, error)

	// ListPullRequests lists pull requests with filtering
	ListPullRequests(ctx context.Context, owner, repo string, opts ListPullRequestOptions) ([]*WorkPullRequest, error)

	// MergePullRequest merges a pull request
	MergePullRequest(ctx context.Context, owner, repo string, prID string, opts MergePullRequestOptions) error
}

// MilestoneClient handles milestone operations
type MilestoneClient interface {
	// CreateMilestone creates a new milestone
	CreateMilestone(ctx context.Context, owner, repo string, opts CreateMilestoneOptions) (*WorkMilestone, error)

	// UpdateMilestone updates an existing milestone
	UpdateMilestone(ctx context.Context, owner, repo string, milestoneID string, opts UpdateMilestoneOptions) (*WorkMilestone, error)

	// GetMilestone retrieves a milestone by ID
	GetMilestone(ctx context.Context, owner, repo string, milestoneID string) (*WorkMilestone, error)

	// ListMilestones lists milestones with filtering
	ListMilestones(ctx context.Context, owner, repo string, opts ListMilestoneOptions) ([]*WorkMilestone, error)
}

// CommentClient handles comment operations
type CommentClient interface {
	// PostComment posts a comment on an issue or PR
	PostComment(ctx context.Context, owner, repo string, issueID string, body string) (*WorkComment, error)

	// ListComments lists comments on an issue or PR
	ListComments(ctx context.Context, owner, repo string, issueID string) ([]*WorkComment, error)

	// UpdateComment updates a comment
	UpdateComment(ctx context.Context, owner, repo string, commentID string, body string) (*WorkComment, error)

	// DeleteComment deletes a comment
	DeleteComment(ctx context.Context, owner, repo string, commentID string) error
}

// LabelClient handles label operations
type LabelClient interface {
	// AddLabel adds labels to an issue or PR
	AddLabel(ctx context.Context, owner, repo string, issueID string, labels []string) error

	// RemoveLabel removes a label from an issue or PR
	RemoveLabel(ctx context.Context, owner, repo string, issueID string, labelID string) error

	// ListLabels lists labels on an issue or PR
	ListLabels(ctx context.Context, owner, repo string, issueID string) ([]*WorkLabel, error)
}

// RepositoryClient handles repository operations
type RepositoryClient interface {
	// GetRepository retrieves repository information
	GetRepository(ctx context.Context, owner, repo string) (*WorkRepository, error)

	// ListRepositories lists repositories for a user or organization
	ListRepositories(ctx context.Context, opts ListRepoOptions) ([]*WorkRepository, error)
}

// ============================================================================
// API Options Structs
// ============================================================================

// CreateIssueOptions contains options for creating an issue
type CreateIssueOptions struct {
	Title        string
	Body         string
	Assignees    []string
	Milestone    string
	Labels       []string
	Closed       bool
	DueDate      *time.Time
	CustomFields map[string]interface{} // Provider-specific fields
}

// UpdateIssueOptions contains options for updating an issue
type UpdateIssueOptions struct {
	Title        *string
	Body         *string
	Assignees    []string
	Milestone    *string
	State        *string
	Labels       []string
	CustomFields map[string]interface{}
}

// ListIssueOptions contains options for listing issues
type ListIssueOptions struct {
	State     string
	Labels    []string
	Milestone string
	Assignee  string
	Creator   string
	Since     *time.Time
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
}

// CreatePullRequestOptions contains options for creating a PR
type CreatePullRequestOptions struct {
	Title        string
	Body         string
	Base         string
	Head         string
	Assignees    []string
	Labels       []string
	Milestone    string
	Draft        bool
	CustomFields map[string]interface{}
}

// UpdatePullRequestOptions contains options for updating a PR
type UpdatePullRequestOptions struct {
	Title        *string
	Body         *string
	Base         *string
	State        *string
	Assignees    []string
	Labels       []string
	Milestone    *string
	CustomFields map[string]interface{}
}

// ListPullRequestOptions contains options for listing PRs
type ListPullRequestOptions struct {
	State     string
	Base      string
	Head      string
	Labels    []string
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

// MergePullRequestOptions contains options for merging a PR
type MergePullRequestOptions struct {
	Style        string // "merge", "rebase", "squash"
	Message      string
	DeleteBranch bool
	CustomFields map[string]interface{}
}

// CreateMilestoneOptions contains options for creating a milestone
type CreateMilestoneOptions struct {
	Title       string
	Description string
	DueDate     *time.Time
	State       string
}

// UpdateMilestoneOptions contains options for updating a milestone
type UpdateMilestoneOptions struct {
	Title       *string
	Description *string
	DueDate     *time.Time
	State       *string
}

// ListMilestoneOptions contains options for listing milestones
type ListMilestoneOptions struct {
	State     string
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

// ListRepoOptions contains options for listing repositories
type ListRepoOptions struct {
	Owner     string
	Type      string // "all", "owner", "member", "public", "private"
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}
