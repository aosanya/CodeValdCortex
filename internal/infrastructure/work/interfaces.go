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
