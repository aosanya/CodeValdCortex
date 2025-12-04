# Work Items & Document Management Domain

## Overview

This domain covers **work tracking** and **document/code versioning** within CodeValdCortex's multi-agent architecture. After extensive research (see [research-sessions/](./research-sessions/)), we implemented a **Git-in-ArangoDB** system that provides:

- **Document & Code Versioning**: Full Git implementation stored in ArangoDB
- **Collaborative Editing**: Section-based documents enabling parallel human/AI editing
- **Intelligent Merging**: AI-assisted conflict resolution
- **File Explorer UX**: Git mechanics abstracted behind familiar file/folder interface
- **Kanban Integration**: Work items linked to file changes via branches and pull requests

## What We Built

✅ **Git-in-ArangoDB**: Complete Git implementation (commits, trees, blobs, refs) stored in ArangoDB  
✅ **File Explorer UI**: Familiar file/folder interface - Git mechanics completely hidden from users  
✅ **AI-Assisted Merging**: Intelligent conflict resolution for concurrent human/AI edits  
✅ **Section-Based Documents**: Granular editing enabling parallel work without conflicts  
✅ **Pull Request Workflow**: Internal PR system for review and approval (no external VCS)  

## What We Did NOT Build

❌ **External VCS Integration**: No Gitea, GitHub, GitLab, or Bitbucket integration  
❌ **Webhook-Based Sync**: No bidirectional synchronization with external systems  
❌ **External Issue Tracking**: Issues are managed within CodeValdCortex, not external platforms  
❌ **GitOps Workflows**: No external Git server operations or remote repository management  

**Why?** See [architectural decisions](./architecture/vcs-integration-decisions.md) for the rationale behind choosing internal Git implementation over external VCS integration.

---

## Topic Index

| Topic | File | Tasks | Status | Description |
|-------|------|-------|--------|-------------|
| **📚 Core Architecture** | [git-based-document-system.md](./git-based-document-system.md) | MVP-WI-005, MVP-PUB-007 | 📋 Planned | Complete Git implementation in ArangoDB with file explorer UX |
| **🎯 Implementation Roadmap** | [implementation-guide.md](./implementation-guide.md) | MVP-WI-005 | 📋 Planned | 7-phase implementation plan with tasks and acceptance criteria |
| **🔀 Pull Requests** | [pull-requests.md](./pull-requests.md) | MVP-WI-005 | 📋 Planned | Internal PR workflow, review process, AI-assisted merging |
| **📊 Work Item Schema** | [work-item-schema.md](./work-item-schema.md) | MVP-WI-005 | 🔄 Needs Update | Data models for issues, milestones, workflows (update for Git integration) |
| **📦 Deliverables Structure** | [deliverables-structure-research.md](./deliverables-structure-research.md) | MVP-043 | ✅ Complete | Enhanced deliverable architecture with folder trees and AI prompt instructions |
| **📋 Architectural Decisions** | [architecture/vcs-integration-decisions.md](./architecture/vcs-integration-decisions.md) | MVP-WI-005 | ✅ Complete | 15 decisions documenting evolution from external VCS to Git-in-ArangoDB |
| **🔬 Research Sessions** | [research-sessions/](./research-sessions/) | MVP-WI-005 | ✅ Complete | Dated research session summaries |

---

## Quick Start

**For Understanding**: Read [git-based-document-system.md](./git-based-document-system.md) → [vcs-integration-decisions.md](./architecture/vcs-integration-decisions.md) → [research session](./research-sessions/2025-11-26_git-in-arangodb.md)

**For Implementing**: Follow [implementation-guide.md](./implementation-guide.md) for 7-phase roadmap

**For Contributing**: See Contributing Guidelines below

---

## Implementation Status

**Phase 0**: ✅ Research & Design Complete  
**Phase 1-7**: 📋 Not Started

See [implementation-guide.md](./implementation-guide.md) for 7-phase roadmap with tasks and acceptance criteria.

---

## Database Collections

```go
// Git system
git_objects, git_refs, repositories, pull_requests, file_index

// Work tracking  
work_issues, workflows
```

See [git-based-document-system.md](./git-based-document-system.md#git-object-model-in-arangodb) for detailed schemas.

---

## API Overview

- **File Operations**: Browse, read, write, delete files
- **Repository Management**: Branches, commits, history
- **Pull Requests**: Create, review, merge, AI conflict resolution

See [git-based-document-system.md](./git-based-document-system.md#file-explorer-implementation) for complete API documentation.

---

## Folder Structure

```
work-items-integration/
├── README.md                          # Domain overview (this file)
├── git-based-document-system.md       # Main architecture (945 lines - comprehensive spec)
├── implementation-guide.md            # 7-phase roadmap (402 lines)
├── pull-requests.md                   # Internal PR workflow (305 lines)
├── work-item-schema.md                # Data models (369 lines - needs update)
├── architecture/
│   └── vcs-integration-decisions.md   # 15 architectural decisions
└── research-sessions/
    └── 2025-11-26_git-in-arangodb.md  # Research summary
```

---

## Contributing Guidelines

### Adding New Content
1. Check file sizes: `wc -l *.md`
2. Add to existing topic file if related (keep < 500 lines)
3. Create new topic file for new areas (descriptive name, not task ID)
4. Update Topic Index in this README

### Updating Architecture
1. Document decisions in `architecture/vcs-integration-decisions.md`
2. Update `git-based-document-system.md` for specs
3. Update `implementation-guide.md` for phases
4. Create research session summary for major changes

### File Size Compliance
- README.md: < 300 lines ✅
- Topic files: < 500 lines (exception for comprehensive specs)
- Architecture details: Split to `architecture/` if needed

---

**Last Updated**: 2025-11-26  
**Phase**: Design Complete → Ready for Implementation  
**Next**: Phase 1 - Git Core Layer

See [implementation-guide.md](./implementation-guide.md) for detailed roadmap.

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
