# Work Tracking Integration Abstraction Layer

## Overview

The `work` package provides a **provider-agnostic abstraction layer** for integrating with external work tracking systems (Gitea, GitHub, GitLab, Jira, Linear, etc.). This allows the CodeValdCortex system to:

✅ Support multiple work tracking platforms simultaneously  
✅ Switch providers without changing orchestrator or core logic  
✅ Migrate from one platform to another seamlessly  
✅ Test with mock providers without external dependencies  
✅ Avoid vendor lock-in

## Architecture

```
┌─────────────────────────────────────────────────────┐
│         External Work Tracking Systems              │
│  Gitea | GitHub | GitLab | Jira | Linear            │
└───────────────────┬─────────────────────────────────┘
                    │ Webhooks / API
                    ▼
┌─────────────────────────────────────────────────────┐
│     Work Abstraction Layer (this package)           │
│  - WorkTrackingProvider interface                   │
│  - WorkIssue, WorkPullRequest, WorkMilestone        │
│  - Common webhook handling patterns                 │
└───────────────────┬─────────────────────────────────┘
                    │ Implements
        ┌───────────┼───────────┐
        ▼           ▼           ▼
   ┌────────┐  ┌────────┐  ┌────────┐
   │ Gitea  │  │ GitHub │  │ GitLab │
   │Provider│  │Provider│  │Provider│
   └───┬────┘  └───┬────┘  └───┬────┘
       │           │           │
       └───────────┼───────────┘
                   │ Persist
                   ▼
            ┌────────────┐
            │  ArangoDB  │
            │work-issues │
            │ work-prs   │
            └────────────┘
```

## Core Interfaces

### WorkTrackingProvider

Main contract for all work tracking system integrations:

```go
type WorkTrackingProvider interface {
    // GetProviderName returns the unique identifier (e.g., "gitea", "github")
    GetProviderName() string

    // ValidateWebhookSignature verifies webhook authenticity
    ValidateWebhookSignature(payload []byte, signature string, secret string) error

    // HandleWebhook processes incoming webhooks and returns normalized results
    HandleWebhook(ctx context.Context, r *http.Request) (*WebhookResult, error)
}
```

### Repository

Persistence layer for work items:

```go
type Repository interface {
    SaveIssue(ctx context.Context, issue *WorkIssue) error
    SavePullRequest(ctx context.Context, pr *WorkPullRequest) error
    SaveMilestone(ctx context.Context, milestone *WorkMilestone) error
    GetIssue(ctx context.Context, provider string, issueID string) (*WorkIssue, error)
    // ... other CRUD operations
}
```

### Specialized Handlers

Optional interfaces for granular control:

```go
type IssueHandler interface {
    ParseIssueWebhook(payload []byte) (*WorkIssue, error)
    ValidateIssueEvent(issue *WorkIssue, action string) bool
}

type PullRequestHandler interface {
    ParsePRWebhook(payload []byte) (*WorkPullRequest, error)
    ValidatePREvent(pr *WorkPullRequest, action string) bool
}

type WebhookValidator interface {
    ValidateSignature(payload []byte, signature string, secret string) error
    GetSignatureHeader() string
}
```

## Common Data Models

### WorkIssue

Provider-agnostic representation of an issue/ticket:

```go
type WorkIssue struct {
    // Provider identification
    Provider string `json:"provider"` // "gitea", "github", "gitlab", "jira"

    // Universal fields
    IssueID     string `json:"issue_id"`     // Unique within provider
    IssueNumber int64  `json:"issue_number"` // Human-readable number
    Title       string `json:"title"`
    Body        string `json:"body"`
    State       string `json:"state"`         // "open", "closed", etc.

    // Milestone/Sprint
    Milestone   string `json:"milestone,omitempty"`
    MilestoneID string `json:"milestone_id,omitempty"`

    // Repository/Project
    RepoURL     string `json:"repo_url"`
    ProjectKey  string `json:"project_key,omitempty"`  // For Jira/Linear
    ProjectName string `json:"project_name,omitempty"`

    // Labels/Tags
    Labels []string `json:"labels,omitempty"`

    // Assignees
    Assignees []string `json:"assignees,omitempty"`

    // Author
    AuthorUsername string `json:"author_username"`
    AuthorEmail    string `json:"author_email,omitempty"`

    // Timestamps
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    ClosedAt  *time.Time `json:"closed_at,omitempty"`

    // Provider-specific metadata (JSON blob)
    ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
}
```

### WorkPullRequest

Normalized pull request/merge request:

```go
type WorkPullRequest struct {
    Provider       string
    PRID           string
    PRNumber       int64
    Title          string
    Body           string
    State          string
    HeadBranch     string
    BaseBranch     string
    RepoURL        string
    Mergeable      *bool
    Merged         bool
    MergedAt       *time.Time
    AuthorUsername string
    Reviewers      []string
    Labels         []string
    LinkedIssues   []string // Cross-reference to issues
    CreatedAt      time.Time
    UpdatedAt      time.Time
    ProviderMetadata map[string]interface{}
}
```

### WorkMilestone

Sprint/milestone/release representation:

```go
type WorkMilestone struct {
    Provider     string
    MilestoneID  string
    Title        string
    Description  string
    State        string
    RepoURL      string
    DueDate      *time.Time
    OpenIssues   int
    ClosedIssues int
    CreatedAt    time.Time
    UpdatedAt    time.Time
    ProviderMetadata map[string]interface{}
}
```

## Implementing a New Provider

### Step 1: Create Provider Package

```
internal/infrastructure/webhooks/yourprovider/
├── handler.go       # Implements WorkTrackingProvider
├── validator.go     # Signature validation logic
├── models.go        # Provider-specific types and transformations
└── handler_test.go  # Tests
```

### Step 2: Implement WorkTrackingProvider

```go
package yourprovider

import (
    "github.com/aosanya/CodeValdCortex/internal/infrastructure/webhooks/work"
)

type YourProvider struct {
    secret     string
    repository work.Repository
}

func NewYourProvider(secret string, repo work.Repository) *YourProvider {
    return &YourProvider{
        secret:     secret,
        repository: repo,
    }
}

func (p *YourProvider) GetProviderName() string {
    return "yourprovider"
}

func (p *YourProvider) ValidateWebhookSignature(payload []byte, signature string, secret string) error {
    // Implement provider-specific signature validation
    // e.g., HMAC SHA-256, SHA-1, custom algorithm
}

func (p *YourProvider) HandleWebhook(ctx context.Context, r *http.Request) (*work.WebhookResult, error) {
    // 1. Read request body
    // 2. Validate signature
    // 3. Parse event type
    // 4. Transform to work.WorkIssue/WorkPullRequest/WorkMilestone
    // 5. Persist via repository
    // 6. Return WebhookResult
}
```

### Step 3: Transform Provider Types to Work Models

```go
// Provider-specific payload struct
type YourProviderIssuePayload struct {
    Action string          `json:"action"`
    Issue  *YourIssueType  `json:"issue"`
    // ... provider-specific fields
}

// Transformation method
func (p *YourProviderIssuePayload) ToWorkIssue() *work.WorkIssue {
    return &work.WorkIssue{
        Provider:       "yourprovider",
        IssueID:        p.Issue.ID,
        IssueNumber:    p.Issue.Number,
        Title:          p.Issue.Title,
        Body:           p.Issue.Description,
        State:          p.Issue.Status,
        RepoURL:        p.Issue.ProjectURL,
        Labels:         extractLabels(p.Issue),
        Assignees:      extractAssignees(p.Issue),
        CreatedAt:      p.Issue.CreatedDate,
        UpdatedAt:      p.Issue.ModifiedDate,
        ProviderMetadata: map[string]interface{}{
            "raw_issue_id": p.Issue.RawID,
            "webhook_event": p.Action,
        },
    }
}
```

### Step 4: Register HTTP Endpoints

```go
// In your main router setup
func RegisterYourProviderRoutes(router *gin.Engine, provider *YourProvider) {
    router.POST("/api/v1/webhooks/yourprovider/issues", func(c *gin.Context) {
        result, err := provider.HandleWebhook(c.Request.Context(), c.Request)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, result)
    })
}
```

### Step 5: Test Your Provider

```go
func TestYourProviderHandleWebhook(t *testing.T) {
    mockRepo := &MockRepository{}
    provider := NewYourProvider("secret", mockRepo)

    payload := `{"action":"opened","issue":{...}}`
    req := httptest.NewRequest("POST", "/webhook", strings.NewReader(payload))
    req.Header.Set("X-YourProvider-Signature", "valid-signature")

    result, err := provider.HandleWebhook(context.Background(), req)
    assert.NoError(t, err)
    assert.Equal(t, "yourprovider", result.Provider)
    assert.NotNil(t, result.Issue)
}
```

## Example: Gitea Provider

See `internal/infrastructure/webhooks/gitea/` for a complete reference implementation:

- **models.go**: Gitea SDK types → work.WorkIssue transformation
- **handler.go**: Webhook endpoint handling
- **validator.go**: X-Gitea-Signature HMAC SHA-256 validation

## Provider-Specific Metadata

Use `ProviderMetadata` field to store provider-specific information that doesn't fit the common model:

```go
issue := &work.WorkIssue{
    // ... common fields
    ProviderMetadata: map[string]interface{}{
        // Gitea-specific
        "html_url": "https://gitea.example.com/org/repo/issues/45",
        "issue_id": 12345,

        // Jira-specific
        "issue_key": "PROJ-123",
        "epic_link": "PROJ-100",
        "story_points": 5,

        // GitHub-specific
        "node_id": "MDU6SXNzdWU1",
        "reactions": {"+1": 5, "heart": 2},
    },
}
```

## ArangoDB Collection Naming

All work items use the `work-` prefix:

- **work-issues** - All issues across all providers
- **work-prs** - All pull requests/merge requests
- **work-milestones** - All milestones/sprints/releases
- **work-comments** - Comments on issues/PRs

Query by provider:
```aql
FOR issue IN work-issues
    FILTER issue.provider == "gitea"
    FILTER issue.milestone == "Requirements"
    RETURN issue
```

## Multi-Provider Support

Run multiple providers simultaneously:

```go
// Initialize providers
giteaProvider := gitea.NewGiteaProvider(giteaSecret, repo)
githubProvider := github.NewGitHubProvider(githubSecret, repo)
gitlabProvider := gitlab.NewGitLabProvider(gitlabSecret, repo)

// Register routes
RegisterGiteaRoutes(router, giteaProvider)
RegisterGitHubRoutes(router, githubProvider)
RegisterGitLabRoutes(router, gitlabProvider)

// Orchestrator monitors all work-issues regardless of provider
orchestrator.MonitorWorkIssues() // Watches work-issues collection
```

## Testing Guidelines

### Mock Provider

```go
type MockProvider struct {
    Name string
}

func (m *MockProvider) GetProviderName() string {
    return m.Name
}

func (m *MockProvider) HandleWebhook(ctx context.Context, r *http.Request) (*work.WebhookResult, error) {
    return &work.WebhookResult{
        Provider:  m.Name,
        EventType: work.EventTypeIssue,
        Action:    "opened",
        Issue: &work.WorkIssue{
            Provider:    m.Name,
            IssueID:     "test-123",
            IssueNumber: 123,
            Title:       "Test Issue",
        },
    }, nil
}
```

### Mock Repository

```go
type MockRepository struct {
    SavedIssues []*work.WorkIssue
}

func (m *MockRepository) SaveIssue(ctx context.Context, issue *work.WorkIssue) error {
    m.SavedIssues = append(m.SavedIssues, issue)
    return nil
}
```

## Future Provider Roadmap

| Provider | Priority | Complexity | Notes |
|----------|----------|------------|-------|
| **Gitea** | P0 (MVP) | Medium | ✅ Implemented |
| **GitHub** | P1 | Medium | X-Hub-Signature-256 HMAC |
| **GitLab** | P2 | Medium | X-Gitlab-Token validation |
| **Jira Cloud** | P2 | High | OAuth, different event model |
| **Linear** | P3 | Medium | GraphQL webhooks |
| **Azure DevOps** | P3 | High | Complex REST API |

## Best Practices

1. **Always transform at the boundary**: Convert provider-specific types to `work` models immediately
2. **Use ProviderMetadata wisely**: Don't duplicate common fields, only provider-unique data
3. **Validate early**: Signature validation should be first step in webhook handling
4. **Idempotency**: Webhooks may be delivered multiple times, handle duplicates gracefully
5. **Logging**: Include provider name in all log messages for debugging multi-provider setups
6. **Error handling**: Return appropriate HTTP status codes (401 for auth, 400 for bad payload, 500 for server errors)

## Related Documentation

- [MVP-WI-001](../../../documents/3-SofwareDevelopment/mvp-details/MVP-WI-001.md) - Gitea Implementation
- [Orchestration Architecture](../../../documents/2-SoftwareDesignAndArchitecture/orchestration-architecture.md) - How orchestrator consumes work items
- [Gitea Integration](../../../documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/gitea-integration.md) - Detailed Gitea webhook flow
