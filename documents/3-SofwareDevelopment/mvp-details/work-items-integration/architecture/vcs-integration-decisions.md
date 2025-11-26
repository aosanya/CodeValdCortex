# Git-Based Document System - Architectural Decisions

**Date**: 2025-11-26  
**Research Session**: Gitea Integration → VCS-Agnostic Design → **Git-in-ArangoDB**  
**Related**: [git-based-document-system.md](../git-based-document-system.md)

## Overview

This document captures the architectural decisions made during the research session, showing the **evolution from external VCS integration to internal Git implementation**.

**⚠️ Important**: Decisions 1-10 represent **historical exploration** of external VCS integration (Gitea/GitLab/GitHub). These were ultimately **superseded by Decision 11**, which pivots to implementing Git directly in ArangoDB. They are preserved here to document the decision-making process and rationale for the pivot.

**Current Architecture**: See **Decisions 11-15** for the active implementation strategy.

---

## Decisions 1-10: Historical External VCS Exploration

**⚠️ Status**: These decisions explored external VCS integration but were **superseded by Decision 11**. Preserved for historical context and to document why internal Git was chosen.

---

## Decision 1: VCS-Agnostic Design ❌ Superseded

**Decision**: Abstract VCS integration to support multiple providers (Gitea, GitLab, GitHub, Bitbucket)

**Status**: **NOT IMPLEMENTED** - Decision 11 pivoted to internal Git implementation

**Original Rationale**:
- Organizations use different VCS platforms
- Lock-in to single provider limits adoption
- Provider abstraction enables future flexibility
- Core workflow logic should be independent of VCS specifics

**Implementation**:
- `VCSProvider` interface for all VCS operations
- Provider-specific implementations in `internal/infrastructure/vcs/{provider}/`
- Generic data models: `VCSIssue`, `VCSRepository`, `VCSPullRequest`
- Provider-specific data stored in `provider_data` JSON field

**Alternative Considered**: Gitea-only implementation (simpler initially, but creates technical debt)

---

## Decision 2: ArangoDB as Source of Truth (Not VCS Webhooks)

**Decision**: Issues are created and managed in ArangoDB; VCS is optional sync target

**Rationale**:
- System must work without VCS (offline development, testing)
- Single source of truth eliminates sync conflicts
- VCS integration is enhancement for human visibility, not core dependency
- Agents operate purely within CodeValdCortex ecosystem

**Data Flow**:
```
User/Agent → Create Issue in ArangoDB → (Optional) Sync to VCS
                     ↓
            Change Stream Detection
                     ↓
            Agent Creation & Execution
```

**Alternative Considered**: Webhook-driven (VCS as source of truth)
- **Rejected because**: Creates VCS dependency, complicates offline scenarios, sync conflicts

---

## Decision 3: Instance-Scoped Data Segregation

**Decision**: Each agency instance has independent VCS configuration and issue tracking

**Rationale**:
- Multiple instances of same agency can connect to different repositories
- Example: `prod-instance` → production repo, `staging-instance` → staging repo
- Enables testing/staging scenarios without production interference
- Supports scenario: Team A uses Gitea, Team B uses GitLab, both using same agency blueprint

**Configuration Model**:
```go
AgencyInstance {
    InstanceID: "prod-001"
    VCSConfig: {
        Provider: "gitea"
        RepositoryIDs: ["repo-001", "repo-002"]
    }
}

VCSRepository {
    InstanceID: "prod-001"
    Provider: "gitea"
    RepositoryURL: "https://gitea.example.com/team/project"
    WorkflowID: "software-dev-workflow-001"
}

VCSIssue {
    InstanceID: "prod-001"  // Isolates data per instance
    RepositoryID: "repo-001"
}
```

**Alternative Considered**: Agency-level VCS config (all instances share same repos)
- **Rejected because**: No isolation, can't test independently, no prod/staging separation

---

## Decision 4: Field-Based Data Segregation (Not Graph Edges or Separate Databases)

**Decision**: Use `instance_id` field for data isolation, not graph edges or separate databases

**Comparison**:

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| **Field-based** (`instance_id` field) | Simple queries, fast indexed lookups, flexible | Requires discipline (always filter by instance_id) | ✅ **CHOSEN** |
| **Graph edges** (user→instance, instance→issue) | Relationship modeling, referential integrity | Overkill for simple lookups, slower queries | ❌ Use only for access control |
| **Collection-per-instance** | Physical isolation, impossible to query wrong data | Dynamic collections, hard to query across instances | ❌ Too complex |
| **Database-per-instance** | Strongest isolation, independent lifecycle | Database proliferation, complex cross-instance queries | ❌ Reserve for extreme cases |

**Rationale**:
- Instances within same agency are trusted (same team)
- Simple filtering is sufficient for most queries
- Graph edges add complexity without benefit for one-to-many relationships
- Database-per-instance can be manually configured if extreme isolation needed

**Query Pattern**:
```aql
FOR issue IN vcs_issues
  FILTER issue.instance_id == @instanceId
  FILTER issue.state == "open"
  RETURN issue
```

**Graph edges used ONLY for**: User access control (`user_access_instance`)

---

## Decision 5: VCS Configuration at Instance Creation (Not Post-Creation)

**Decision**: VCS config provided when creating instance (Option A)

**Rationale**:
- Declarative, infrastructure-as-code approach
- Instance fully configured from start
- Enables automated instance provisioning
- Immediate webhook setup possible

**API**:
```json
POST /api/v1/agencies/{agencyID}/instances
{
  "instance_name": "production-001",
  "vcs_config": {
    "provider": "gitea",
    "repositories": [{...}]
  }
}
```

**Alternative Considered**: Configure after instance creation (Option B)
- **Rejected because**: Extra API call, incomplete instance state, harder to automate

---

## Decision 6: Webhook Endpoint per Instance

**Decision**: Instance-specific webhook URLs for explicit routing

**URL Pattern**: `/api/v1/webhooks/vcs/{instance_id}`

**Rationale**:
- Explicit routing (no ambiguity about which instance handles webhook)
- Easy to configure in VCS UI
- Clear in logs which instance received event
- Can disable/enable per instance independently
- Supports multiple instances per agency cleanly

**Alternative Considered**: 
- Repository URL matching (query instance by repo URL) - **Rejected**: slower, requires DB lookup
- Webhook secret mapping - **Rejected**: complex, harder to debug

---

## Decision 7: Asynchronous Webhook Processing

**Decision**: Webhook handler responds immediately, processes in background

**Rationale**:
- Fast webhook response (VCS won't timeout/retry)
- Agent creation can be slow (policy checks, resource allocation)
- Failed processing doesn't fail webhook delivery
- Can implement retry logic independently

**Flow**:
```go
Webhook POST → Validate signature → Save to DB → Respond 200 OK
                                         ↓
                                    Background worker
                                         ↓
                                  Process & create agent
```

**Alternative Considered**: Synchronous processing
- **Rejected because**: Slow webhook responses, timeout risks, blocks VCS webhook delivery

---

## Decision 8: Collections in Agency Database

**Decision**: VCS collections created in agency database (not master database)

**Collections**:
```go
// Agency database: UC-INFRA-001
collections := []string{
    "agents",
    "tasks",
    "work_items",
    "roles",
    "events",
    "logs",
    "metrics",
    "agency_instances",
    
    // VCS integration (NEW)
    "vcs_issues",         // Generic issue tracking
    "vcs_pull_requests",  // Generic PR/MR tracking
    "vcs_repositories",   // Repository configurations
}
```

**Rationale**:
- VCS data is agency-specific (not platform-wide)
- Follows database-per-agency pattern
- Data isolation at agency boundary
- Instance-level isolation via `instance_id` field

**Alternative Considered**: Master database storage
- **Rejected because**: Breaks agency isolation, cross-agency data leakage risk

---

## Decision 9: Optional VCS Integration

**Decision**: VCS integration is optional; instances can run without it

**Scenarios**:
1. **No VCS**: Pure CodeValdCortex workflow (issues created internally)
2. **VCS for visibility**: Issues created in CodeValdCortex, synced to VCS for team visibility
3. **Bi-directional sync**: Issues can be created in either system, kept in sync

**Instance Config**:
```go
// No VCS
VCSConfig: nil

// VCS enabled
VCSConfig: {
    Provider: "gitea",
    RepositoryIDs: ["repo-001"]
}
```

**Rationale**:
- Not all teams use VCS for issue tracking
- Development/testing doesn't require VCS
- Reduces dependencies for simple deployments

---

## Decision 10: Change Stream Monitoring (Not Webhook-Driven Agent Creation)

**Decision**: Orchestrator monitors ArangoDB change streams to detect new issues

**Rationale**:
- Decouples webhook ingestion from agent creation
- Works for both internal and VCS-created issues
- Single code path for agent spawning logic
- Can implement queuing, rate limiting, WIP limits easily

**Architecture**:
```
Issue Created (any source) → vcs_issues collection
                                    ↓
                            Change Stream Event
                                    ↓
                          Orchestrator Detects
                                    ↓
                      Process Milestone → Column Mapping
                                    ↓
                         Create Agent from Blueprint
```

**Alternative Considered**: Direct agent creation in webhook handler
- **Rejected because**: Tight coupling, duplicated logic, harder to test

---

## Summary Table

| Decision | Chosen Approach | Key Benefit |
|----------|-----------------|-------------|
| VCS Provider | Abstraction layer | Multi-vendor support |
| Source of Truth | ArangoDB | Works offline, no sync conflicts |
| Data Scope | Instance-level | Independent configurations |
| Segregation | Field-based (`instance_id`) | Simple, performant |
| Configuration | At instance creation | Declarative, complete setup |
| Webhook Routing | Instance-specific URLs | Explicit, traceable |
| Processing | Asynchronous | Fast response, resilient |
| Database | Agency database | Agency isolation |
| VCS Requirement | Optional | Flexible deployment |
| Agent Trigger | Change streams | Decoupled, single code path |

---

## Decision 11: Internal Git Implementation (No External VCS)

**Date**: 2025-11-26  
**Status**: ✅ Accepted  
**Participants**: System Architecture Team

### Context
Originally planned to integrate with external VCS (Gitea/GitLab/GitHub) via webhooks. However, analysis revealed this creates:
- Sync complexity and circular event loops
- Two sources of truth (ArangoDB + external VCS)
- Limited control over merge strategies
- External dependencies for core versioning functionality

**Key insight**: "Do I really need VCS? My key requirement is to version control documents and handle merge conflicts."

### Decision
Implement full Git object model (commits, trees, blobs, refs) directly in ArangoDB. No external VCS integration.

### Rationale
✅ **Single source of truth**: All data in ArangoDB  
✅ **Custom merge strategies**: AI-assisted conflict resolution  
✅ **Tight integration**: Direct AQL queries on Git history  
✅ **No sync delays**: Immediate consistency  
✅ **No external dependencies**: Simpler deployment  

### Consequences
- **Positive**: Complete control over versioning, merging, and collaboration
- **Positive**: Can implement AI-driven merge strategies
- **Positive**: Git history queryable with AQL
- **Negative**: Must implement Git internals (commits, trees, merges)
- **Negative**: Higher initial development effort

### Implementation
```go
// Collections
git_objects     // Content-addressable storage (SHA-1)
git_refs        // Branches, tags, HEAD
repositories    // Repository metadata
pull_requests   // PR tracking with AI conflict resolution
file_index      // Fast path lookups
```

---

## Decision 12: Section-Based Documents for Granular Merging

**Date**: 2025-11-26  
**Status**: ✅ Accepted  
**Participants**: System Architecture Team

### Context
**Critical requirement**: "Merge conflicts are important" - must support collaborative editing by multiple humans and AI agents simultaneously.

**Problem with monolithic documents**:
```
Agent edits line 50 of requirements.md
User edits line 100 of requirements.md
→ Git merge conflict (entire file)
→ Manual resolution required
```

### Decision
Structure all documents (Markdown, YAML) as sections with unique IDs:

```markdown
---
id: requirements-doc-001
sections:
  - id: introduction
  - id: functional-requirements
  - id: non-functional-requirements
---

<!-- section: introduction -->
# Introduction
Content...

<!-- section: functional-requirements -->
# Functional Requirements
Content...
```

### Rationale
✅ **Granular merging**: Agent edits "Introduction" section, user edits "Functional Requirements" → auto-merge  
✅ **Parallel editing**: Multiple actors can work on same document simultaneously  
✅ **Reduced conflicts**: Conflicts only when same section modified  
✅ **AI-friendly**: Sections have clear semantic boundaries  

### Consequences
- **Positive**: Dramatic reduction in merge conflicts
- **Positive**: True collaborative editing (humans + AI)
- **Positive**: Section-level history and attribution
- **Negative**: Documents must follow section structure
- **Negative**: More complex merge logic

### Implementation
```go
func MergeSectionedDocument(base, source, target string) (merged string, conflicts []Conflict) {
    // Parse into sections
    baseSections := parseDocument(base)
    sourceSections := parseDocument(source)
    targetSections := parseDocument(target)
    
    // Merge at section level (auto-merge different sections)
    for sectionID := range allSections {
        if sourceSections[sectionID] == targetSections[sectionID] {
            merged[sectionID] = sourceSections[sectionID]
        } else if baseSections[sectionID] == targetSections[sectionID] {
            merged[sectionID] = sourceSections[sectionID] // Only source changed
        } else if baseSections[sectionID] == sourceSections[sectionID] {
            merged[sectionID] = targetSections[sectionID] // Only target changed
        } else {
            conflicts = append(conflicts, sectionID) // Both changed
        }
    }
}
```

---

## Decision 13: AI-Assisted Conflict Resolution

**Date**: 2025-11-26  
**Status**: ✅ Accepted  
**Participants**: System Architecture Team

### Context
Even with section-based documents, conflicts will occur when multiple actors edit the same section. Traditional Git conflict markers require manual resolution.

**Opportunity**: Use AI to intelligently merge conflicting changes.

### Decision
When Git merge detects conflicts, invoke AI to analyze both versions and propose merged content:

```go
type AIConflictResolution struct {
    ConflictFiles []string  `json:"conflict_files"`
    ProposedMerge string    `json:"proposed_merge"`
    Confidence    float64   `json:"confidence"`
    Reasoning     string    `json:"reasoning"`
}
```

### Rationale
✅ **Intelligent merging**: AI understands semantic intent  
✅ **Reduced manual work**: Auto-resolve when AI is confident  
✅ **Human oversight**: Human reviews AI proposal before applying  
✅ **Audit trail**: AI reasoning captured for transparency  

### Consequences
- **Positive**: Faster conflict resolution
- **Positive**: AI learns from merge patterns
- **Positive**: Humans can override AI decisions
- **Negative**: AI API costs for conflict resolution
- **Negative**: Requires high-quality prompts

### Implementation Workflow
```
1. Git merge detects conflict in section X
2. Extract: base version, source version, target version
3. Send to AI: "Intelligently merge these two edits"
4. AI analyzes both changes and proposes merged version
5. Store AI proposal in PullRequest.AIResolution
6. Human reviews AI proposal in UI
7. Human approves (apply AI merge) or manually edits
```

---

## Decision 14: File Explorer UX (Abstract Git Mechanics)

**Date**: 2025-11-26  
**Status**: ✅ Accepted  
**Participants**: System Architecture Team, UX Team

### Context
**User requirement**: "The Git mechanics should be abstracted from the users. They should just have the file and folder structure experience."

Users (especially non-technical stakeholders) should not need Git knowledge. Present familiar file explorer interface.

### Decision
Hide all Git terminology and mechanics behind user-friendly actions:

**Users see**:
- 📁 File/folder tree view
- "Save" button (creates commit)
- "Request Review" button (creates PR)
- "Approve" button (merges PR)
- "View History" (shows commit log as timeline)

**Users do NOT see**:
- ❌ `git commit`, `git push`, `git merge`
- ❌ Branch names (except as "Draft", "Published")
- ❌ SHA-1 hashes
- ❌ Commit messages (auto-generated from context)

### Rationale
✅ **Lower barrier to entry**: Non-technical users can collaborate  
✅ **Familiar UX**: Like Google Docs or Notion  
✅ **Git benefits**: Full versioning, branching, merging behind scenes  
✅ **Progressive disclosure**: Power users can access Git details if needed  

### Consequences
- **Positive**: Wider adoption (non-developers)
- **Positive**: Simpler onboarding
- **Positive**: Reduced training requirements
- **Negative**: Loss of Git power user features in primary UI
- **Negative**: Must design intuitive abstractions

### Implementation
```go
// High-level API (what users see)
PUT /api/v1/instances/{id}/files/{path}
{
  "content": "...",
  "message": "Update requirements"  // Optional
}

// Backend creates Git commit
func (h *FileHandler) UpdateFile(c *gin.Context) {
    // 1. Create blob
    blobSHA := gitOps.WriteBlob(content)
    
    // 2. Update tree
    treeSHA := gitOps.UpdateTreeEntry(path, blobSHA)
    
    // 3. Create commit (auto-generate message if not provided)
    msg := generateCommitMessage(path, author)
    commitSHA := gitOps.Commit(treeSHA, []string{parentCommit}, author, msg)
    
    // 4. Update branch
    gitOps.UpdateRef("refs/heads/main", commitSHA)
}
```

---

## Decision 15: Unified Human/AI Access (No Distinction)

**Date**: 2025-11-26  
**Status**: ✅ Accepted  
**Participants**: System Architecture Team

### Context
Should AI agents have different APIs or permissions than humans? Initial consideration was to separate concerns.

**Insight**: "AI and humans should use the same API. No distinction."

### Decision
Humans and AI agents use identical APIs, authentication, and permissions. Only difference is `author` field.

```go
type Author struct {
    Name  string  // "user-alice" OR "agent-requirements-001"
    Email string
    When  time.Time
}
```

### Rationale
✅ **Simpler architecture**: One API, one permission model  
✅ **Natural collaboration**: AI and humans work together seamlessly  
✅ **Same workflows**: AI can create PRs, humans can review  
✅ **Unified audit trail**: All changes tracked consistently  

### Consequences
- **Positive**: Reduced code complexity
- **Positive**: AI can participate in human workflows naturally
- **Positive**: Easier testing (same API for both)
- **Negative**: Must be careful with AI permissions (prevent runaway agents)

### Implementation
```go
// Same API for human and AI
PUT /api/v1/instances/{id}/files/{path}
{
  "content": "...",
  "author": "agent-xyz-001"  // or "user-alice"
}

// Pull request approval
POST /api/v1/instances/{id}/pull-requests/{prID}/approve
{
  "reviewer": "user-alice"  // or "agent-reviewer-001"
}
```

---

## References

- **Git-in-ArangoDB Spec**: [../git-based-document-system.md](../git-based-document-system.md)
- **Implementation Guide**: [../implementation-guide.md](../implementation-guide.md)
- **Research Session**: [../research-sessions/2025-11-26_git-in-arangodb.md](../research-sessions/2025-11-26_git-in-arangodb.md)
- **Original Gitea Design**: ❌ Deleted (see Decision 11 for why we pivoted away from external VCS)
- **Instance Management**: MVP-PUB-007
- **Database Architecture**: `/workspaces/CodeValdCortex/documents/2-SoftwareDesignAndArchitecture/backend-architecture.md`
