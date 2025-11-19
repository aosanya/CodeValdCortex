# MVP-WI-001: Work Tracking System Webhook Integration (Gitea Implementation)

## Overview
**Priority**: P0  
**Effort**: Medium  
**Skills Required**: Go, Webhooks, ArangoDB, REST API, Security  
**Dependencies**: None (Persistence layer only - Orchestrator in MVP-032 consumes persisted data)  
**Status**: In Progress

## Description
Implement a **pluggable webhook integration system** with Gitea as the first implementation. The system uses an abstraction layer (`work` package) that defines interfaces for work tracking systems (issues, PRs, milestones), allowing easy integration with GitHub, GitLab, Jira, or other platforms in the future.

### Pluggable Architecture

**Interface Layer** (`internal/infrastructure/webhooks/work/`):
- Defines common interfaces for work tracking concepts
- Provider-agnostic data models
- Webhook handler interface

**Implementation Layer** (`internal/infrastructure/webhooks/gitea/`):
- Gitea-specific webhook handlers
- Implements `work` interfaces
- Transforms Gitea payloads to common models

**Future Implementations**:
- `internal/infrastructure/webhooks/github/` - GitHub webhooks
- `internal/infrastructure/webhooks/gitlab/` - GitLab webhooks
- `internal/infrastructure/webhooks/jira/` - Jira webhooks
- `internal/infrastructure/webhooks/linear/` - Linear webhooks

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    External Work Tracking Systems                │
│  ┌───────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐ │
│  │  Gitea    │  │  GitHub    │  │  GitLab    │  │   Jira     │ │
│  │ (MVP-WI-001)│  │  (Future)  │  │  (Future)  │  │  (Future)  │ │
│  └─────┬─────┘  └──────┬─────┘  └──────┬─────┘  └──────┬─────┘ │
│        │ Webhooks      │ Webhooks      │ Webhooks      │ API    │
└────────┼───────────────┼───────────────┼───────────────┼────────┘
         │               │               │               │
         ▼               ▼               ▼               ▼
┌─────────────────────────────────────────────────────────────────┐
│          Work Tracking Integration Layer (Abstraction)           │
│                                                                   │
│  internal/infrastructure/webhooks/work/                           │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  Interfaces:                                                 ││
│  │  - WorkTrackingProvider                                      ││
│  │  - IssueHandler                                              ││
│  │  - PullRequestHandler                                        ││
│  │  - WebhookValidator                                          ││
│  │                                                               ││
│  │  Common Models:                                               ││
│  │  - WorkIssue (provider-agnostic)                             ││
│  │  - WorkPullRequest                                           ││
│  │  - WorkMilestone                                             ││
│  └─────────────────────────────────────────────────────────────┘│
└────────────────────────────┬────────────────────────────────────┘
                             │ Implements
         ┌───────────────────┼───────────────────┐
         ▼                   ▼                   ▼
┌──────────────────┐  ┌──────────────┐  ┌──────────────┐
│ Gitea Provider   │  │GitHub Provider│  │GitLab Provider│
│ (MVP-WI-001)     │  │   (Future)   │  │   (Future)   │
│                  │  │              │  │              │
│ - GiteaHandler   │  │- GitHubHandler│  │- GitLabHandler│
│ - Signature Val  │  │- Signature Val│  │- Signature Val│
│ - Payload Parser │  │- Payload Parser│ │- Payload Parser│
└────────┬─────────┘  └──────┬───────┘  └──────┬───────┘
         │                   │                  │
         └───────────────────┼──────────────────┘
                             │ Persist to
                             ▼
                    ┌─────────────────┐
                    │    ArangoDB     │
                    │  ┌────────────┐ │
                    │  │work-issues │ │
                    │  │work-prs    │ │
                    │  │work-milestones││
                    │  └────────────┘ │
                    └─────────────────┘
```

### Why Pluggable Architecture?

**Benefits**:
- ✅ **Flexibility**: Switch from Gitea to GitHub/GitLab without changing orchestrator
- ✅ **Multi-Platform**: Support multiple work tracking systems simultaneously
- ✅ **Migration**: Easy migration path from one platform to another
- ✅ **Testability**: Mock providers for testing without external dependencies
- ✅ **Vendor Independence**: Not locked into any single platform

### ArangoDB-Centric Architecture

**Separation of Concerns**:
- **MVP-WI-001 (This Task)**: Receive webhooks → Validate signatures → Persist to ArangoDB
- **MVP-032 (Orchestrator)**: Monitor ArangoDB change streams → Create agents from work item definitions

**Data Flow**:
```
Gitea Instance
  └─ Webhooks (issues, PRs, milestones)
      ↓
MVP-WI-001: Webhook Handler
  ├─ Validate X-Gitea-Signature (HMAC SHA-256)
  ├─ Parse JSON payload
  ├─ Transform to ArangoDB document
  └─ Save to work-issues / work-prs collections
      ↓
ArangoDB Collections:
  ├─ work-issues (issue_number, title, body, milestone, state, repo_url, etc.)
  ├─ work-prs (pr_number, title, head_branch, base_branch, state, etc.)
  └─ work-milestones (milestone_id, title, description, due_date, etc.)
      ↓
MVP-032: Orchestrator (Change Stream Monitoring)
  ├─ Watches work-issues collection for inserts/updates
  ├─ Detects milestoned issues
  ├─ Queries workflow/column mappings
  ├─ Checks WIP limits
  └─ Creates agents from work item definitions
```

**Why ArangoDB-Centric?**:
- ✅ **Decoupling**: Webhook handler doesn't need orchestrator, workflows, or agent factory
- ✅ **Resilience**: Webhooks persist data even if orchestrator is down
- ✅ **Auditability**: Complete history of all Gitea events in database
- ✅ **Queryability**: Agents query ArangoDB for issue details, not Gitea API
- ✅ **Change Streams**: Orchestrator reactively processes new issues via ArangoDB native feature

## Objectives
- Implement **ArangoDB persistence layer** for Gitea webhook events
- Validate webhook signatures to ensure authenticity (X-Gitea-Signature)
- Parse and transform Gitea payloads into ArangoDB documents
- Store issues, PRs, milestones, and comments in respective collections
- Provide data foundation for Orchestrator to monitor and create agents
- Handle webhook delivery failures with proper HTTP status codes

## Requirements

### Functional Requirements
1. **Webhook Endpoints**:
   - POST /api/v1/work/issues - Issue events
   - POST /api/v1/work/pull-requests - PR events
2. **Supported Events**:
   - **Issues**: `opened`, `milestoned`, `demilestoned`, `closed`, `edited`, `labeled`, `assigned`
   - **Pull Requests**: `opened`, `synchronized`, `closed`, `merged`, `edited`
   - **Comments**: `issue_comment.created`, `pull_request_comment.created`
3. **ArangoDB Persistence**:
   - Save issue data to `work-issues` collection
   - Save PR data to `work-prs` collection
   - Save milestone data to `work-milestones` collection
   - Update existing documents on webhook updates (upsert logic)
4. **Document Schema**:
   ```json
   // work-issues collection
   {
     "_key": "gitea-issue-<issue_id>",
     "issue_id": 123,
     "issue_number": 45,
     "title": "Implement authentication",
     "body": "Need OAuth2 support...",
     "state": "open",
     "milestone": "Requirements Gathering",
     "milestone_id": 5,
     "repo_url": "https://gitea.example.com/org/repo",
     "repo_owner": "org",
     "repo_name": "repo",
     "labels": ["feature", "security"],
     "assignees": ["user1", "user2"],
     "created_at": "2025-11-19T10:00:00Z",
     "updated_at": "2025-11-19T11:30:00Z",
     "synced_at": "2025-11-19T11:30:05Z"
   }
   ```
5. **Signature Validation**: HMAC SHA-256 verification using X-Gitea-Signature header
6. **Error Handling**: Return appropriate HTTP status codes (401, 400, 500)

### Non-Functional Requirements
1. **Security**: 
   - HMAC SHA-256 signature validation on all webhooks
   - Configurable secret token stored in config.yaml
   - Optional IP allowlist for webhook sources
2. **Performance**:
   - Asynchronous processing (handle webhook in goroutine)
   - Response time <200ms for webhook acceptance
   - Non-blocking work item execution
3. **Reliability**:
   - Idempotent handlers (handle duplicate webhooks)
   - Graceful error recovery
   - Audit logging for all webhook events
4. **Observability**:
   - Structured logging with MVP-WI-001 prefix
   - Webhook delivery status tracking
   - Error rate monitoring

## Acceptance Criteria
- [ ] POST /api/v1/work/issues endpoint created and registered in Gin router
- [ ] X-Gitea-Signature HMAC SHA-256 validation working (rejects invalid signatures)
- [ ] Repository→Workflow mapping stored in ArangoDB and queryable by repository URL
- [ ] Milestone→Column mapping extracted from webhook payload
- [ ] Issue "milestoned" event triggers workflow column lookup
- [ ] Column's work item definition retrieved from ArangoDB
- [ ] WIP (Work-In-Progress) limit checked before agent creation
- [ ] Issues queued when WIP limit exceeded for column
- [ ] Agent instance created from work item definition with milestone/issue context
- [ ] Agent starts execution asynchronously (non-blocking webhook response)
- [ ] Agent linked to issue, workflow, and column in tracking table
- [ ] Issue "demilestoned" event stops associated agent
- [ ] Issue "closed" event stops running agent and marks as completed
- [ ] Success/progress comments posted to Gitea issues by agents
- [ ] Configuration added to config.yaml (webhook secret, allowed IPs)
- [ ] All errors logged with MVP-WI-001 prefix
- [ ] Webhook processing returns 200 OK within 200ms
- [ ] Unit tests cover signature validation, milestone extraction, WIP limits, agent creation
- [ ] Integration tests with real Gitea instance verify Kanban flow (issue moves through milestones)
- [ ] Documentation: Kanban setup guide, repository-workflow linking, milestone configuration

## Technical Specifications

### Package Structure

**Work Package (Abstraction Layer)**:
```
internal/infrastructure/webhooks/work/
├── interfaces.go        # WorkTrackingProvider, IssueHandler, Repository
├── models.go           # WorkIssue, WorkPullRequest, WorkMilestone (provider-agnostic)
└── README.md           # Provider implementation guide
```

**Gitea Package (Implementation)**:
```
internal/infrastructure/webhooks/gitea/
├── handler.go          # HTTP endpoint handlers implementing work.WorkTrackingProvider
├── validator.go        # HMAC SHA-256 signature validation
├── models.go           # Gitea-specific payload types and transformation to work models
├── repository.go       # ArangoDB persistence for work items
└── handler_test.go     # Unit tests
```

### Core Interfaces

**work.WorkTrackingProvider** - Main contract for all work tracking integrations:
```go
type WorkTrackingProvider interface {
    GetProviderName() string
    ValidateWebhookSignature(payload []byte, signature string, secret string) error
    HandleWebhook(ctx context.Context, r *http.Request) (*WebhookResult, error)
}
```

**work.Repository** - Persistence abstraction:
```go
type Repository interface {
    SaveIssue(ctx context.Context, issue *WorkIssue) error
    SavePullRequest(ctx context.Context, pr *WorkPullRequest) error
    SaveMilestone(ctx context.Context, milestone *WorkMilestone) error
}
```

### Data Flow

```
Gitea Webhook Event (issue.milestoned)
    ↓
POST /api/v1/work/issues → handler.HandleWebhook()
    ↓
validator.ValidateSignature(X-Gitea-Signature)
    ↓
Parse JSON → GiteaIssuePayload
    ↓
GiteaIssuePayload.ToWorkIssue() → work.WorkIssue (normalized)
    ↓
repository.SaveIssue(workIssue) → ArangoDB work-issues collection
    ↓
Return 200 OK
    ↓
[Later] Orchestrator monitors ArangoDB change streams
    ↓
Detects new milestoned issue → Creates agent from work item definition
```

### Key Data Transformations

**Gitea → Common Model**:
```go
// Gitea-specific payload
type GiteaIssuePayload struct {
    Action     string            `json:"action"`
    Issue      *gitea.Issue      `json:"issue"`
    Repository *gitea.Repository `json:"repository"`
}

// Transform to provider-agnostic model
func (p *GiteaIssuePayload) ToWorkIssue() *work.WorkIssue {
    return &work.WorkIssue{
        Provider:       "gitea",
        IssueID:        p.Issue.HTMLURL,
        IssueNumber:    p.Issue.Index,
        Title:          p.Issue.Title,
        Body:           p.Issue.Body,
        State:          string(p.Issue.State),
        Milestone:      p.Issue.Milestone.Title,
        RepoURL:        p.Repository.HTMLURL,
        Labels:         extractLabels(p.Issue.Labels),
        Assignees:      extractAssignees(p.Issue.Assignees),
        ProviderMetadata: map[string]interface{}{
            "html_url":      p.Issue.HTMLURL,
            "webhook_event": p.Action,
        },
    }
}
```

### ArangoDB Collections

**work-issues** (Common issue format):
```json
{
    "provider": "gitea",
    "issue_id": "https://gitea.example.com/org/repo/issues/45",
    "issue_number": 45,
    "title": "Implement authentication",
    "body": "Need OAuth2 support...",
    "state": "open",
    "milestone": "Requirements",
    "milestone_id": "Requirements",
    "repo_url": "https://gitea.example.com/org/repo",
    "labels": ["feature", "security"],
    "assignees": ["user1", "user2"],
    "author_username": "developer1",
    "author_email": "dev@example.com",
    "created_at": "2025-11-19T10:00:00Z",
    "updated_at": "2025-11-19T11:30:00Z",
    "provider_metadata": {
        "html_url": "https://gitea.example.com/org/repo/issues/45",
        "webhook_event": "milestoned",
        "issue_id": 12345
    }
}
```

### Architecture

**Package Structure**:
```
internal/infrastructure/webhooks/gitea/
├── handler.go           # HTTP endpoint handlers
├── validator.go         # Signature validation
├── processor.go         # Event processing logic
├── matcher.go           # Label → WorkItemDefinition matching
├── factory.go           # Agent instantiation from definitions
└── handler_test.go      # Unit tests
```

**Dependencies**:
- `code.gitea.io/sdk/gitea` - Official Gitea Go SDK for types
- `crypto/hmac` - HMAC signature validation
- `github.com/gin-gonic/gin` - HTTP router
- `internal/agent` - Agent creation and lifecycle management
- `internal/domain/workflow/models` - WorkItemDefinition types
- `internal/domain/workflow/repository` - WorkItemDefinition queries

### Data Flow
```
Gitea Issue Event (with label: "requirements-gathering")
    ↓
Webhook POST → handler.go (signature validation)
    ↓
processor.go (parse payload, extract labels)
    ↓
matcher.go (label → work item definition lookup)
    ↓
    Query ArangoDB: work_item_definitions WHERE type = "requirements-gathering"
    ↓
factory.go (create agent from definition + issue context)
    ↓
    AgentFactory.CreateFromWorkItemDefinition()
    ↓
Agent Instance Created in ArangoDB (agents collection)
    ↓
Agent Started (agent.Start())
    ↓
Link Agent ↔ Issue (agent_issue_tracking table)
    ↓
Post comment to Gitea: "✅ Requirements Gathering Agent created (ID: abc-123)"
```

**Key Data Structures**:

```go
// Work Item Definition (blueprint in published agency)
type WorkItemDefinition struct {
    ID           string   `json:"id"`
    Type         string   `json:"type"`          // e.g., "requirements-gathering"
    Name         string   `json:"name"`
    Description  string   `json:"description"`
    Instructions string   `json:"instructions"`  // What the agent should do
    Tools        []string `json:"tools"`         // Tools agent can use
    AutonomyLevel string  `json:"autonomy_level"`// "supervised", "autonomous"
    AgencyID     string   `json:"agency_id"`
}

// Agent Instance (created from definition)
type Agent struct {
    ID               string    `json:"id"`
    Name             string    `json:"name"`
    Type             string    `json:"type"`           // From WorkItemDefinition.Type
    WorkItemDefID    string    `json:"work_item_def_id"` // Link to definition
    TriggerSource    string    `json:"trigger_source"` // "gitea"
    GiteaIssueID     int64     `json:"gitea_issue_id"`
    GiteaIssueURL    string    `json:"gitea_issue_url"`
    IssueContext     IssueContext `json:"issue_context"`
    Status           string    `json:"status"`         // "created", "running", "completed"
    CreatedAt        time.Time `json:"created_at"`
}

// Issue Context (injected into agent)
type IssueContext struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Repository  string `json:"repository"`
    Labels      []string `json:"labels"`
}
```

### Signature Validation

```go
func validateSignature(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedMAC := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
```

### Issue Label → WorkItemDefinition Matching

**Label Convention**:
- Work item type labels use format: `{work-item-type}`
- Examples: `requirements-gathering`, `code-review`, `security-audit`, `testing`

**Matching Logic**:
```go
func (m *Matcher) FindWorkItemDefinition(labels []string, agencyID string) (*WorkItemDefinition, error) {
    // 1. Extract work item type from labels
    workItemType := extractWorkItemType(labels)
    if workItemType == "" {
        return nil, errors.New("no work item type label found")
    }
    
    // 2. Query ArangoDB for definition
    query := `
        FOR def IN work_item_definitions
            FILTER def.type == @type
            AND def.agency_id == @agencyID
            AND def.enabled == true
            LIMIT 1
            RETURN def
    `
    
    cursor, err := m.db.Query(ctx, query, map[string]interface{}{
        "type":     workItemType,
        "agencyID": agencyID,
    })
    
    var def WorkItemDefinition
    cursor.ReadDocument(ctx, &def)
    return &def, nil
}

func extractWorkItemType(labels []string) string {
    // Known work item type labels
    workItemTypes := []string{
        "requirements-gathering",
        "code-review",
        "security-audit",
        "testing",
        "deployment",
        "documentation",
    }
    
    for _, label := range labels {
        for _, validType := range workItemTypes {
            if label == validType {
                return validType
            }
        }
    }
    return ""
}
```

| Gitea Label | WorkItemDefinition.Type | Agent Created |
|-------------|------------------------|---------------|
| `requirements-gathering` | `requirements-gathering` | Requirements Gathering Agent |
| `code-review` | `code-review` | Code Review Agent |
| `security-audit` | `security-audit` | Security Audit Agent |
| `testing` | `testing` | Testing Agent |

### Configuration Schema

```yaml
gitea:
  webhooks:
    secret: "${GITEA_WEBHOOK_SECRET}"  # HMAC secret token
    allowed_ips:                        # Optional IP allowlist
      - "10.0.0.0/8"
      - "172.16.0.0/12"
    timeout: 30s                        # Webhook processing timeout
    retry:
      max_attempts: 3
      backoff: exponential
```

### Error Handling

| Error Scenario | HTTP Status | Action |
|---------------|-------------|--------|
| Invalid signature | 401 Unauthorized | Log warning, reject webhook |
| Malformed JSON | 400 Bad Request | Log error, return details |
| Missing required fields | 422 Unprocessable Entity | Log validation errors |
| Database error | 500 Internal Server Error | Log error, trigger retry |
| Unknown event type | 200 OK | Log info, skip processing |

### Logging Format

```go
log.Info().
    Str("task", "MVP-WI-001").
    Str("event", "issues.opened").
    Int64("issue_id", payload.Issue.Index).
    Str("repository", payload.Repository.FullName).
    Msg("Processing Gitea webhook")
```

## Implementation Phases

### Phase 1: Foundation (Day 1)
- Create package structure
- Add Gitea SDK dependency
- Implement signature validation
- Create HTTP endpoint

### Phase 2: Event Handlers (Day 2)
- Implement issue.opened handler
- Implement issue.labeled handler
- Implement issue.closed handler
- Add payload → WorkItem mapping

### Phase 3: Configuration & Error Handling (Day 3)
- Add config.yaml integration
- Implement comprehensive error handling
- Add structured logging
- Add retry logic

### Phase 4: Testing & Documentation (Day 4)
- Write unit tests (signature, parsing, mapping)
- Set up Gitea in docker-compose
- Run integration tests
- Write setup documentation

## Testing Strategy

### Unit Tests
```go
func TestValidateSignature(t *testing.T) { /* ... */ }
func TestParseIssuePayload(t *testing.T) { /* ... */ }
func TestMapIssueToWorkItem(t *testing.T) { /* ... */ }
func TestHandleIssueOpened(t *testing.T) { /* ... */ }
```

### Integration Tests
1. Start Gitea in docker-compose
2. Create test repository
3. Configure webhook pointing to localhost:8080
4. Create issue with "work-item" label
5. Verify work item created in ArangoDB
6. Close issue, verify work item status updated

## Dependencies

**Upstream (Must Complete First)**:
- ✅ MVP-029: Work Items Architecture
- ✅ MVP-030: Work Items Core Schema & Registry
- ✅ MVP-031: Work Items Lifecycle & SLA
- ✅ MVP-032: Work Items Assignment & Routing

**Downstream (Blocks)**:
- MVP-WI-002: Gitea API Client (uses webhook data for bidirectional sync)
- MVP-WI-003: Work Item to Issue Sync (reverse direction)
- MVP-WI-004: Pull Request Automation (builds on webhook infrastructure)

## Success Metrics
- Webhook signature validation: 100% rejection of invalid signatures
- Event processing latency: <200ms to accept webhook, <2s to persist work item
- Error rate: <1% for valid webhooks
- Zero data loss: All valid webhooks result in work item creation/update

## References
- **Architecture**: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/gitea-integration.md`
- **Gitea SDK**: https://gitea.com/gitea/go-sdk
- **Webhook Documentation**: https://docs.gitea.io/en-us/webhooks/
- **Work Items System**: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/`
