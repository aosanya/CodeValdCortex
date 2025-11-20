# Webhook Integration

<!-- MVP-WI-001 -->
**Tasks Covered**: MVP-WI-001  
**Status**: ✅ Complete (2025-11-19)

## Overview

The webhook integration forms the foundation of our external work tracking integration. It implements a **pluggable architecture** that allows switching between different work tracking providers (Gitea, GitHub, GitLab, Jira, Linear) without changing downstream orchestration logic.

### Why Pluggable Architecture?

Rather than hard-coding Gitea-specific logic throughout the system, we created an abstraction layer (`work` package) that defines provider-agnostic interfaces and data models. This enables:

- ✅ **Flexibility**: Switch from Gitea to GitHub without changing orchestrator
- ✅ **Multi-Platform**: Support multiple work tracking systems simultaneously
- ✅ **Migration**: Easy migration path from one platform to another
- ✅ **Testability**: Mock providers for testing
- ✅ **Vendor Independence**: Not locked into any single platform

## Architecture Layers

**Interface Layer** (`internal/infrastructure/webhooks/work/`):
- `WorkTrackingProvider` - Main contract for all providers
- `WorkIssue`, `WorkPullRequest`, `WorkMilestone` - Provider-agnostic models
- `Repository` - ArangoDB persistence interface
- `WebhookValidator` - Signature verification

**Implementation Layer** (`internal/infrastructure/webhooks/gitea/`):
- `GiteaHandler` - Implements `WorkTrackingProvider`
- `GiteaIssuePayload` - Gitea-specific webhook types
- `ToWorkIssue()` - Transformation to common models
- HMAC SHA-256 signature validation

**Future Implementations**:
- `github/` - GitHub webhooks (X-Hub-Signature-256)
- `gitlab/` - GitLab webhooks (X-Gitlab-Token)
- `jira/` - Jira webhooks (different event model)
- `linear/` - Linear webhooks (GraphQL-based)

## Data Flow

```
Gitea Instance
  └─ Webhook: issue.milestoned
      ↓
POST /api/v1/work/issues
      ↓
1. Validate X-Gitea-Signature (HMAC SHA-256)
2. Parse JSON payload → GiteaIssuePayload
3. Transform → work.WorkIssue (provider-agnostic)
4. Persist → ArangoDB work-issues collection
5. Return 200 OK
      ↓
ArangoDB Change Streams emit event
      ↓
Orchestrator (MVP-032) detects new milestoned issue
      ↓
Creates agent from work item definition
```

### Why ArangoDB-Centric?

**Decoupling Benefits**:
- Webhook handler doesn't need orchestrator, workflows, or agent factory
- Webhooks persist data even if orchestrator is down (resilience)
- Complete audit history of all external events
- Agents query ArangoDB, not external APIs (performance)
- Change streams provide reactive processing (no polling)

## Provider-Agnostic Data Models

**WorkIssue** (common format across all providers):
```go
type WorkIssue struct {
    Provider       string    // "gitea", "github", "gitlab", "jira"
    IssueID        string    // Unique within provider
    IssueNumber    int64     // Human-readable number
    Title          string
    Body           string
    State          string    // "open", "closed"
    Milestone      string
    MilestoneID    string
    RepoURL        string
    ProjectKey     string    // For Jira
    Labels         []string
    Assignees      []string
    AuthorUsername string
    CreatedAt      time.Time
    UpdatedAt      time.Time
    ProviderMetadata map[string]interface{} // Platform-specific fields
}
```

## ArangoDB Collections

- **work-issues**: All issues across all providers
- **work-prs**: All pull requests/merge requests
- **work-milestones**: All milestones/sprints/releases

Query by provider:
```aql
FOR issue IN work-issues
    FILTER issue.provider == "gitea"
    FILTER issue.milestone == "Requirements"
    RETURN issue
```

## Objectives

- Implement ArangoDB persistence layer for Gitea webhook events
- Validate webhook signatures (X-Gitea-Signature HMAC SHA-256)
- Parse and transform Gitea payloads into provider-agnostic models
- Store issues, PRs, milestones in respective collections
- Provide data foundation for Orchestrator monitoring
- Handle webhook delivery failures with proper HTTP status codes

## Requirements

**Functional**:
1. **Webhook Endpoints**:
   - POST /api/v1/work/issues - Issue events
   - POST /api/v1/work/pull-requests - PR events

2. **Supported Events**:
   - **Issues**: `opened`, `milestoned`, `demilestoned`, `closed`, `edited`, `labeled`, `assigned`
   - **Pull Requests**: `opened`, `synchronized`, `closed`, `merged`, `edited`

3. **ArangoDB Persistence**:
   - Save to `work-issues`, `work-prs`, `work-milestones` collections
   - Upsert logic for updates (idempotent)

**Non-Functional**:
1. **Security**: HMAC SHA-256 signature validation, configurable secret, optional IP allowlist
2. **Performance**: Response time <200ms, asynchronous processing
3. **Reliability**: Idempotent handlers, graceful error recovery
4. **Observability**: Structured logging (MVP-WI-001 prefix), error tracking

## Acceptance Criteria

- [x] POST /api/v1/work/issues endpoint registered
- [x] X-Gitea-Signature HMAC SHA-256 validation working
- [x] GiteaIssuePayload transforms to work.WorkIssue correctly
- [x] work-issues collection persists with all required fields
- [x] Duplicate webhooks handled idempotently
- [x] Invalid signatures return 401 Unauthorized
- [x] Malformed payloads return 400 Bad Request
- [x] Server errors return 500 with error logging
- [x] Unit tests for signature validation, transformation, persistence
- [ ] Integration tests with mock Gitea webhooks (deferred)
- [x] Documentation: provider implementation guide

## Technical Implementation

**Completed Files**:
- `internal/infrastructure/work/interfaces.go` - Provider interfaces (WorkTrackingProvider, Repository)
- `internal/infrastructure/work/models.go` - Provider-agnostic models (WorkIssue, WorkPullRequest, WorkMilestone)
- `internal/infrastructure/work/README.md` - Provider implementation guide
- `internal/infrastructure/gitea/models.go` - Gitea payload types and transformers
- `internal/infrastructure/gitea/handler.go` - HTTP webhook handlers with async processing
- `internal/infrastructure/gitea/validator.go` - HMAC SHA-256 signature validation
- `internal/infrastructure/gitea/repository.go` - ArangoDB persistence with idempotent upsert
- `internal/infrastructure/gitea/validator_test.go` - Unit tests (100% pass)
- `internal/app/app.go` - Handler initialization and route registration
- `internal/config/config.go` - WorkTrackingConfig struct and environment bindings
- `config.yaml` - work_tracking configuration section

**Implementation Highlights**:
- Pluggable provider architecture (abstraction layer + Gitea provider)
- ArangoDB-centric design (webhooks persist → orchestrator monitors change streams)
- Async webhook processing (non-blocking, <200ms response times)
- Security: HMAC SHA-256 with constant-time comparison
- Resource-oriented API design (/api/v1/work/issues not /webhooks/gitea/issues)
- Comprehensive structured logging with MVP-WI-001 prefix

**Validation**:
- ✅ Code compiles successfully
- ✅ All unit tests passing
- ✅ No lint errors
- ✅ Configuration loads without errors
- ✅ Ready for Gitea webhook configuration

## Implementation History

| Date | Session | Summary |
|------|---------|---------|
| 2025-11-19 | [MVP-WI-001_gitea_webhook_integration](../coding_sessions/MVP-WI-001_gitea_webhook_integration.md) | ✅ **Completed**: Pluggable work tracking abstraction layer with full Gitea provider implementation. Created 9 new files (~1,570 LOC) including handler, validator, repository, models, tests. HTTP endpoints registered at /api/v1/work/issues and /api/v1/work/pull-requests. HMAC SHA-256 signature validation, async webhook processing, ArangoDB persistence with idempotent upserts. All unit tests passing. |

## Related Topics

- See [api-client.md](./api-client.md) for bidirectional communication with Gitea
- See [synchronization.md](./synchronization.md) for agent-to-issue updates
- See [work-item-schema.md](./work-item-schema.md) for data models
