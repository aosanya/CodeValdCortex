# Work Items Integration Domain

## Overview

The Work Items Integration system provides comprehensive infrastructure for managing work tracking across external systems (Gitea, GitHub, GitLab, Jira) and internal agent workflows. This domain encompasses webhook ingestion, work item definitions, lifecycle management, routing, and bidirectional synchronization with external platforms.

### Vision

Enable CodeValdCortex agents to be automatically spawned and managed based on work items from external tracking systems, creating a seamless bridge between traditional project management tools and autonomous agent execution. Work items serve as both triggers and blueprints for agent creation, with full bidirectional sync keeping external systems updated on progress.

### Architecture

```
External Systems (Gitea/GitHub/etc.)
    ↓ Webhooks
┌─────────────────────────────────────────┐
│  Work Tracking Abstraction Layer        │
│  (pluggable provider architecture)      │
└──────────────┬──────────────────────────┘
               ↓ Persist
         ┌──────────────┐
         │  ArangoDB    │
         │ work-issues  │
         │ work-prs     │
         └──────┬───────┘
                ↓ Change Streams
         ┌─────────────────┐
         │  Orchestrator   │
         │ (MVP-032)       │
         └──────┬──────────┘
                ↓ Creates
         ┌─────────────────┐
         │    Agents       │
         └──────┬──────────┘
                ↓ Sync Back
    External Systems (Comments, Status)
```

### Key Concepts

- **Work Items**: Abstract representations of work (issues, PRs, tasks, jobs) from any provider
- **Work Item Definitions**: Blueprints for creating agents from work items
- **Providers**: Pluggable integrations (Gitea, GitHub, GitLab, Jira)
- **Lifecycle States**: State machines governing work item progression
- **Routing Rules**: Intelligent assignment of work items to qualified agents
- **Bidirectional Sync**: Keep external systems updated with agent progress

---

## Topic Index

This domain is organized into topic-based files covering related functionality:

| Topic | File | Tasks Covered | Status | Description |
|-------|------|---------------|--------|-------------|
| **Webhook Integration** | [webhooks.md](./webhooks.md) | MVP-WI-001 | ✅ Complete | Ingest events from Gitea/GitHub/GitLab |
| **API Client** | [api-client.md](./api-client.md) | MVP-WI-002 | ✅ Complete | Bidirectional communication with Gitea |
| **Agent Synchronization** | [synchronization.md](./synchronization.md) | MVP-WI-003 | ✅ Complete | Update issues with agent progress |
| **Pull Request Automation** | [pull-requests.md](./pull-requests.md) | MVP-WI-004 | 🔄 In Progress | Automated PR creation and merging |
| **Work Item Schema** | [work-item-schema.md](./work-item-schema.md) | MVP-030, MVP-031, MVP-032 | 📋 Planned | Core schema, lifecycle, routing |

### Quick Navigation

- **New to this domain?** → Start with [webhooks.md](./webhooks.md) to understand event ingestion
- **Need API integration?** → See [api-client.md](./api-client.md) for Gitea API patterns
- **Building sync features?** → Review [synchronization.md](./synchronization.md) for event-driven updates
- **Working on PR automation?** → Explore [pull-requests.md](./pull-requests.md) for current implementation
- **Designing work items?** → Reference [work-item-schema.md](./work-item-schema.md) for data models

---

## Adding New Content to This Domain

### Topic-Based Organization Rules

1. **Use topic names, not task IDs**: Files should be named `webhooks.md`, `authentication.md`, NOT `MVP-001.md`
2. **Group related tasks**: If 2+ tasks cover the same topic, put them in ONE file (e.g., MVP-030, MVP-031, MVP-032 → `work-item-schema.md`)
3. **File size limits**:
   - Topic files: MAX 500 lines
   - This README: MAX 300 lines
4. **Split when needed**: If topic file exceeds 500 lines, extract detailed designs to `architecture/` folder

### Adding a New Topic

1. **Create topic file**: `{topic-name}.md` (e.g., `authentication.md`)
2. **Add to Topic Index**: Update the table above with new topic, tasks covered, status
3. **Keep focused**: Each topic file covers ONE coherent area of functionality
4. **Cross-reference**: Link between related topics using relative paths

### Example Topic File Structure

```markdown
# {Topic Name}

<!-- MVP-XXX, MVP-YYY -->
**Tasks Covered**: MVP-XXX, MVP-YYY  
**Status**: ✅ Complete / 🔄 In Progress / 📋 Planned

## Overview
[What problem does this solve?]

## Requirements
[Detailed requirements from all covered tasks]

## Architecture
[Technical approach, diagrams]

## Implementation
[Code structure, key files]

## Testing
[Test strategy, validation]

## Related Topics
- See [api-client.md](./api-client.md) for API integration
- See [webhooks.md](./webhooks.md) for event handling
```

---

## Architecture Details

### Pluggable Provider System

The work tracking integration uses a **provider-agnostic architecture** that abstracts common concepts across different platforms:

**Core Abstraction** (`internal/infrastructure/work/`):
```go
type APIClient interface {
    // Issues
    GetIssue(ctx context.Context, number int) (*WorkIssue, error)
    CreateIssue(ctx context.Context, req *CreateIssueRequest) (*WorkIssue, error)
    UpdateIssue(ctx context.Context, number int, req *UpdateIssueRequest) error
    PostComment(ctx context.Context, number int, comment string) (*WorkComment, error)
    
    // Labels
    AddLabel(ctx context.Context, number int, label string) error
    RemoveLabel(ctx context.Context, number int, label string) error
    
    // Milestones
    GetMilestone(ctx context.Context, number int) (*WorkMilestone, error)
    
    // Pull Requests
    CreatePR(ctx context.Context, req *CreatePRRequest) (*WorkPR, error)
    MergePR(ctx context.Context, number int, method string) error
}
```

**Provider Implementations**:
- `GiteaClient` (implemented) - Gitea API
- `GitHubClient` (planned) - GitHub API
- `GitLabClient` (planned) - GitLab API
- `JiraClient` (planned) - Jira REST API

**Benefits**:
- ✅ Single orchestration codebase works with all providers
- ✅ Easy to add new platforms
- ✅ Testable with mock implementations
- ✅ No vendor lock-in

### Data Flow

**Inbound (External → CodeValdCortex)**:
```
1. Gitea webhook fired (issue created/updated)
2. HTTP handler validates signature
3. Transform Gitea payload → WorkIssue
4. Persist to ArangoDB (gitea_issues collection)
5. ArangoDB change stream triggers orchestrator
6. Orchestrator creates agent from work item definition
```

**Outbound (CodeValdCortex → External)**:
```
1. Agent lifecycle event (started, completed, failed)
2. SyncService handles event
3. Render comment template
4. GiteaClient.PostComment() via APIClient interface
5. Update labels/milestone via GiteaClient
6. Record audit trail in ArangoDB
```

### ArangoDB Collections

| Collection | Purpose | Key Fields | Indexes |
|------------|---------|------------|---------|
| `gitea_issues` | Webhook issue storage | repository_url, issue_id, state | repository_url, issue_id |
| `gitea_pull_requests` | Webhook PR storage | repository_url, pr_number, state | repository_url, pr_number |
| `agent_issue_links` | Agent↔Issue mapping | agent_id, issue_id, status | agent_id (unique), issue_id |
| `sync_audit` | Sync operation audit | agent_id, sync_type, success | agent_id, issue_id, sync_timestamp |
| `pr_metadata` | PR quality/auto-merge | pr_number, quality_checks, auto_merge_decision | pr_number, repository_url |

### Event System

**Event Types**:
- `EventTypeAgentCreated` - Agent spawned from work item
- `EventTypeAgentStarted` - Agent began execution
- `EventTypeAgentHealthChanged` - Agent degraded/quarantined
- `EventTypeAgentStopped` - Agent completed/terminated
- `EventTypeTaskStarted` - Agent started task execution
- `EventTypeTaskCompleted` - Task finished successfully
- `EventTypeTaskFailed` - Task failed with error
- Custom: `progress_update`, `milestone_complete`

**Event Flow**:
```
Agent Lifecycle Event
    ↓
Event Bus (events.Processor)
    ↓
SyncEventHandler
    ↓
SyncService.HandleAgentEvent()
    ↓
├─ PostComment (template rendering)
├─ UpdateLabels (state-based labels)
├─ ProgressMilestone (workflow advancement)
└─ RecordAudit (compliance logging)
```

---

## Integration Points

### With Other Domains

- **Agent Lifecycle** (`agent-lifecycle/`): Uses work items as agent blueprints
- **Orchestration** (`orchestration/`): Monitors work item changes, spawns agents
- **AI Policy Layer** (`ai-policy/`): Enforces policies on agent actions triggered by work items
- **Communication** (`communication/`): Agents report progress back to work tracking systems

### External Systems

- **Gitea**: Primary integration (webhooks + API client implemented)
- **GitHub**: Planned (similar API structure)
- **GitLab**: Planned (different webhook format)
- **Jira**: Planned (different data model)

---

## Success Metrics

- ✅ **Phase 1 Complete**: 3/3 foundation tasks operational
- 🚀 **Phase 2 Active**: PR automation in development
- 📊 **Performance**: <200ms webhook processing, <300ms sync latency
- 🔒 **Security**: HMAC signature validation, secure token storage
- 📈 **Scalability**: ArangoDB change streams for async orchestration
- ✅ **Provider Agnostic**: Abstract interface supports multiple platforms

---

## References

- **Architecture Docs**: `architecture/` folder (detailed designs)
- **API Specifications**: Individual task files contain API schemas
- **Code Location**: `internal/infrastructure/work/`, `internal/infrastructure/webhooks/`
- **Related Domains**: Agent lifecycle, Orchestration, AI policy
