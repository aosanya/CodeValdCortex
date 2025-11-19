# Gitea API Client Adapter

## Overview

The Gitea API Client Adapter bridges the Gitea-specific client (`internal/infrastructure/gitea`) with the provider-agnostic `work.APIClient` interface. This enables services to work with any provider (Gitea, GitHub, GitLab) through a common interface.

## Architecture

```
┌────────────────────────────────────────┐
│  Service Layer (Agent Sync, etc.)     │
│  Depends on: work.APIClient interface │
└──────────────┬─────────────────────────┘
               │
               ▼
┌────────────────────────────────────────┐
│  APIClientAdapter (this file)         │
│  - Implements work.APIClient           │
│  - Converts types (Gitea ↔ work)      │
└──────────────┬─────────────────────────┘
               │
               ▼
┌────────────────────────────────────────┐
│  Gitea Client (client.go)             │
│  - Gitea SDK-specific implementation   │
│  - Returns gitea.Issue, gitea.PR, etc.│
└────────────────────────────────────────┘
```

## Usage

### Creating the Adapter

```go
import (
    "time"
    giteawebhook "github.com/aosanya/CodeValdCortex/internal/infrastructure/gitea"
    "github.com/aosanya/CodeValdCortex/internal/infrastructure/work"
)

// Step 1: Create the Gitea client
giteaClient, err := giteawebhook.NewClient(giteawebhook.ClientConfig{
    BaseURL:   "https://gitea.example.com",
    Token:     "your-api-token",
    Timeout:   30 * time.Second,
    RateLimit: 10,
})
if err != nil {
    log.Fatal(err)
}

// Step 2: Wrap with adapter
apiClient := giteawebhook.NewAPIClientAdapter(giteaClient, "https://gitea.example.com")

// Step 3: Use through work.APIClient interface
var client work.APIClient = apiClient
```

### Using the Adapter

#### Creating an Issue

```go
issue, err := apiClient.CreateIssue(ctx, "myorg", "myrepo", work.CreateIssueOptions{
    Title:     "Bug found by agent",
    Body:      "Agent detected an error during execution",
    Assignees: []string{"devops-team"},
    Labels:    []string{"bug", "agent-generated"},
    Milestone: "42",  // String ID (Gitea uses int64 internally)
})
// Returns: *work.WorkIssue (not *gitea.Issue)
```

#### Posting a Comment

```go
comment, err := apiClient.PostComment(ctx, "myorg", "myrepo", "123", 
    "Agent progress: Task 50% complete")
// Returns: *work.WorkComment
```

#### Merging a Pull Request

```go
err := apiClient.MergePullRequest(ctx, "myorg", "myrepo", "456", work.MergePullRequestOptions{
    Style:   "squash",
    Message: "Automated merge by agent",
})
```

## Type Conversions

The adapter handles automatic conversion between Gitea and work package types:

### ID Conversion

| Gitea Type | Work Type | Conversion |
|------------|-----------|------------|
| `int64` (issue.Index) | `string` (WorkIssue.IssueID) | `strconv.FormatInt()` |
| `int64` (pr.Index) | `string` (WorkPullRequest.PRID) | `strconv.FormatInt()` |
| `int64` (milestone.ID) | `string` (WorkMilestone.MilestoneID) | `strconv.FormatInt()` |

**Why strings?** Different providers use different ID types:
- Gitea/GitHub: `int64` 
- Jira/Linear: `string` UUIDs
- Using strings as common denominator

### Time Handling

Gitea SDK uses pointer times (`*time.Time`) in some places:

```go
// Adapter handles nil checks and dereferencing
if pr.Created != nil {
    workPR.CreatedAt = *pr.Created
}
```

### Label Conversion

```go
// Gitea: []gitea.Label with ID, Name, Color, Description
// Work: []string (just names)

workIssue.Labels = make([]string, len(issue.Labels))
for i, label := range issue.Labels {
    workIssue.Labels[i] = label.Name
}
```

### State Normalization

```go
// Gitea: gitea.StateType ("open", "closed")
// Work: string

workIssue.State = string(issue.State)
```

## Limitations & TODOs

### Current Limitations

1. **Label Creation**: `AddLabel()` currently accepts label names but Gitea API requires label IDs. TODO: Add label lookup by name.

2. **PR Assignees**: Gitea's CreatePR/UpdatePR endpoints don't support assignees in the request. TODO: Add separate API call to set assignees after creation.

3. **Reviewers**: Gitea SDK v0.19.0 doesn't expose `RequestedReviewers` field. TODO: Update when SDK adds support.

4. **Custom Fields**: `work.CreateIssueOptions.CustomFields` not yet implemented. TODO: Map to Gitea-specific fields.

### Future Enhancements

1. **Caching**: Cache frequently accessed data (labels, milestones)
2. **Batch Operations**: Support bulk operations for efficiency
3. **Retry Logic**: Add exponential backoff for failed requests
4. **Metrics**: Track conversion overhead, API call patterns

## Testing

### Unit Tests

```bash
go test ./internal/infrastructure/gitea/ -v -run TestAPIClientAdapter
```

Tests verify:
- ✅ Adapter implements `work.APIClient`
- ✅ Adapter implements all sub-interfaces
- ✅ Adapter creation succeeds

### Integration Tests (Future)

TODO: Add integration tests that:
- Create actual Gitea issues through adapter
- Verify type conversions are correct
- Test round-trip (create → get → verify)

## Example: Agent Sync Service

```go
type AgentSyncService struct {
    apiClient work.APIClient  // Uses interface, not Gitea-specific type
}

func (s *AgentSyncService) OnAgentProgress(agent *Agent, progress int) error {
    // Works with ANY provider (Gitea, GitHub, GitLab)
    comment, err := s.apiClient.PostComment(
        context.Background(),
        agent.RepoOwner,
        agent.RepoName,
        agent.IssueID,
        fmt.Sprintf("🤖 Agent %s: %d%% complete", agent.ID, progress),
    )
    return err
}

func (s *AgentSyncService) OnAgentComplete(agent *Agent) error {
    // Close the issue
    return s.apiClient.CloseIssue(
        context.Background(),
        agent.RepoOwner,
        agent.RepoName,
        agent.IssueID,
    )
}

func (s *AgentSyncService) OnAgentError(agent *Agent, err error) error {
    // Create new issue for the error
    issue, err := s.apiClient.CreateIssue(
        context.Background(),
        agent.RepoOwner,
        agent.RepoName,
        work.CreateIssueOptions{
            Title:  fmt.Sprintf("Agent %s encountered error", agent.ID),
            Body:   fmt.Sprintf("Error: %v", err),
            Labels: []string{"agent-error", "bug"},
        },
    )
    return err
}
```

## Related Documentation

- [Gitea Client](./client.go) - Gitea-specific API client
- [Work Interfaces](../work/interfaces.go) - Provider-agnostic interfaces
- [Work Models](../work/models.go) - Common data models
- [MVP-WI-002](../../../documents/3-SofwareDevelopment/coding_sessions/MVP-WI-002_gitea_api_client.md) - Gitea client implementation
- [MVP-WI-003](../../../documents/3-SofwareDevelopment/mvp-details/work-items-integration.md) - Agent-to-Issue Sync (next task)
