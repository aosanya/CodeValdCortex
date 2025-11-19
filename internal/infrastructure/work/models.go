package work

import "time"

// WorkIssue represents a normalized issue/ticket from any work tracking system
// Persisted to ArangoDB collection: work-issues
type WorkIssue struct {
	// Provider identifies the source system (gitea, github, gitlab, jira, etc.)
	Provider string `json:"provider"`

	// IssueID is the provider-specific unique identifier
	IssueID string `json:"issue_id"`

	// IssueNumber is the human-readable issue number
	IssueNumber int64 `json:"issue_number"`

	// Title is the issue title/summary
	Title string `json:"title"`

	// Body is the issue description/content
	Body string `json:"body"`

	// State represents the current status (open, closed, etc.)
	State string `json:"state"`

	// Milestone information
	Milestone   string `json:"milestone,omitempty"`
	MilestoneID string `json:"milestone_id,omitempty"`

	// Repository/Project information
	RepoURL     string `json:"repo_url"`
	ProjectKey  string `json:"project_key,omitempty"`
	ProjectName string `json:"project_name,omitempty"`

	// Labels/Tags
	Labels []string `json:"labels,omitempty"`

	// Assignees
	Assignees []string `json:"assignees,omitempty"`

	// Author information
	AuthorUsername string `json:"author_username"`
	AuthorEmail    string `json:"author_email,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`

	// Provider-specific metadata (JSON blob for provider-specific fields)
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
}

// WorkPullRequest represents a normalized pull request/merge request
// Persisted to ArangoDB collection: work-prs
type WorkPullRequest struct {
	// Provider identifies the source system
	Provider string `json:"provider"`

	// PRID is the provider-specific unique identifier
	PRID string `json:"pr_id"`

	// PRNumber is the human-readable PR number
	PRNumber int64 `json:"pr_number"`

	// Title is the PR title
	Title string `json:"title"`

	// Body is the PR description
	Body string `json:"body"`

	// State represents the current status (open, closed, merged, etc.)
	State string `json:"state"`

	// Branch information
	HeadBranch string `json:"head_branch"`
	BaseBranch string `json:"base_branch"`

	// Repository information
	RepoURL     string `json:"repo_url"`
	ProjectKey  string `json:"project_key,omitempty"`
	ProjectName string `json:"project_name,omitempty"`

	// Merge status
	Mergeable *bool      `json:"mergeable,omitempty"`
	Merged    bool       `json:"merged"`
	MergedAt  *time.Time `json:"merged_at,omitempty"`

	// Author and reviewers
	AuthorUsername string   `json:"author_username"`
	AuthorEmail    string   `json:"author_email,omitempty"`
	Reviewers      []string `json:"reviewers,omitempty"`

	// Labels
	Labels []string `json:"labels,omitempty"`

	// Linked issues
	LinkedIssues []string `json:"linked_issues,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`

	// Provider-specific metadata
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
}

// WorkMilestone represents a normalized milestone/sprint/release
// Persisted to ArangoDB collection: work-milestones
type WorkMilestone struct {
	// Provider identifies the source system
	Provider string `json:"provider"`

	// MilestoneID is the provider-specific unique identifier
	MilestoneID string `json:"milestone_id"`

	// Title is the milestone name
	Title string `json:"title"`

	// Description is the milestone description
	Description string `json:"description,omitempty"`

	// State represents the current status (open, closed, active, etc.)
	State string `json:"state"`

	// Repository/Project information
	RepoURL     string `json:"repo_url,omitempty"`
	ProjectKey  string `json:"project_key,omitempty"`
	ProjectName string `json:"project_name,omitempty"`

	// Due date
	DueDate *time.Time `json:"due_date,omitempty"`

	// Progress tracking
	OpenIssues   int `json:"open_issues"`
	ClosedIssues int `json:"closed_issues"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`

	// Provider-specific metadata
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
}

// WorkComment represents a comment on an issue or PR
type WorkComment struct {
	// Provider identifies the source system
	Provider string `json:"provider"`

	// CommentID is the provider-specific unique identifier
	CommentID string `json:"comment_id"`

	// ParentID references the issue or PR
	ParentID   string        `json:"parent_id"`
	ParentType WorkEventType `json:"parent_type"` // issue or pull_request

	// Body is the comment text
	Body string `json:"body"`

	// Author information
	AuthorUsername string `json:"author_username"`
	AuthorEmail    string `json:"author_email,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Provider-specific metadata
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
}

// WorkLabel represents a label/tag in a work tracking system
type WorkLabel struct {
	// Provider identifies the source system
	Provider string `json:"provider"`

	// LabelID is the provider-specific unique identifier
	LabelID string `json:"label_id"`

	// Name is the label name
	Name string `json:"name"`

	// Description is the label description
	Description string `json:"description,omitempty"`

	// Color is the label color (hex format)
	Color string `json:"color,omitempty"`

	// Provider-specific metadata
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
}

// WorkRepository represents a repository/project in a work tracking system
type WorkRepository struct {
	// Provider identifies the source system
	Provider string `json:"provider"`

	// RepoID is the provider-specific unique identifier
	RepoID string `json:"repo_id"`

	// Name is the repository name
	Name string `json:"name"`

	// FullName is the full repository path (e.g., "owner/repo")
	FullName string `json:"full_name"`

	// Description is the repository description
	Description string `json:"description,omitempty"`

	// Owner is the repository owner username
	Owner string `json:"owner"`

	// URL is the repository web URL
	URL string `json:"url"`

	// DefaultBranch is the default branch name
	DefaultBranch string `json:"default_branch,omitempty"`

	// IsPrivate indicates if the repository is private
	IsPrivate bool `json:"is_private"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Provider-specific metadata
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
}

// GetCollectionName returns the ArangoDB collection name for work items
func GetCollectionName(eventType WorkEventType) string {
	switch eventType {
	case EventTypeIssue:
		return "work-issues"
	case EventTypePullRequest:
		return "work-prs"
	case EventTypeMilestone:
		return "work-milestones"
	case EventTypeComment:
		return "work-comments"
	default:
		return "work-unknown"
	}
}
