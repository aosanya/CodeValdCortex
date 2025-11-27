# Issue Management & Lifecycle

**Related Tasks**: MVP-WI-008 (Workbench/Kanban Board), MVP-030 (Work Item Definitions), MVP-031 (Agent Assignment)  
**Status**: 📋 Planned - Architecture Defined

## Overview

This document describes the complete lifecycle of work items (issues) in CodeValdCortex, from creation through completion. Issues are runtime instances of work that progress through workflow stages, with Git integration for all artifacts produced.

**UI Location**: Workbench (navbar link after Instances)

**Key Concepts**:
- **Issues**: Runtime work instances created by users on Workbench board
- **Work Items**: Templates defined in agency specs (REQ1, IMPL1, etc.)
- **Assignment**: Manual or claim-based worker allocation
- **Git Integration**: All work produces commits/PRs in internal Git system
- **Progression**: Automatic advancement through workflow stages on PR merge

---

## Complete Feature Request Workflow

### Use Case: "Implement JWT Authentication"

This traces a feature request from creation to completion through the software development workflow.

---

## Step 1: Issue Creation (User Action)

**Actor**: Human user  
**Location**: Workbench UI (navbar link)

**Process**:
1. User navigates to agency instance Workbench
2. Clicks **"Create Issue"** button
3. Fills out issue creation form:
   ```
   Title: "Implement JWT authentication"
   Description: "We need JWT-based authentication for API endpoints.
                 Requirements: HS256 algorithm, 1-hour expiry, refresh tokens..."
   ```
4. Submits issue

**System Response**:
- Creates `work_issue` document in ArangoDB `work_issues` collection
- Automatically places issue in **REQ1 column** (only entry point)
- Issue assigned unique ID (e.g., `issue-123`)
- Status: `open`, Stage: `REQ1`

**Data Model**:
```json
{
  "_key": "issue-123",
  "title": "Implement JWT authentication",
  "description": "We need JWT-based authentication...",
  "workflow_id": "software-dev-workflow-001",
  "current_step": "REQ1",
  "status": "open",
  "created_by": "user-alice",
  "created_at": "2025-11-26T10:00:00Z",
  "assigned_to": null,
  "branch_name": null,
  "pull_request_id": null
}
```

**UI State**:
- Issue appears as card in **REQ1** column of Workbench board
- Available for assignment/claiming

---

## Step 2: Work Assignment (Flexible Model)

**Actors**: Admin or Agent/Human worker  
**Location**: Workbench or Work Queue

**Two Assignment Modes**:

### Mode A: Manual Assignment
1. Admin/manager clicks on issue card
2. Clicks **"Assign Agent"** or **"Assign Worker"** button
3. Selects from dropdown:
   - Agents: `REQUIREMENTS-ANALYST` (from agency roles)
   - Humans: `user-bob`, `user-charlie`
4. Worker assigned to issue

### Mode B: Claim-Based (Pull Model)
1. Issue sits unassigned in REQ1 column
2. Available agents/humans see issue in "Available Work" queue
3. Agent or human clicks **"Claim Work"** on issue
4. First to claim gets assigned

**System Response**:
- Updates `work_issue.assigned_to` field
- If agent: Creates agent instance from role definition
- Notifies assigned worker
- Issue status: `assigned`

**Data Update**:
```json
{
  "_key": "issue-123",
  "assigned_to": "agent-req-001",  // or "user-bob"
  "assigned_at": "2025-11-26T10:05:00Z",
  "status": "assigned"
}
```

---

## Step 3: Work Execution - Git Integration

**Actor**: Assigned agent or human  
**Location**: Agent runtime or human workspace

**Process**:

### 3.1: Branch Creation
When agent/human starts work:
1. Agent calls API: `POST /api/v1/git/branches`
   ```json
   {
     "branch_name": "issue-123-jwt-auth",
     "base_branch": "main",
     "issue_id": "issue-123"
   }
   ```
2. System creates Git branch in ArangoDB
3. Updates issue with branch link

**Data Update**:
```json
{
  "_key": "issue-123",
  "branch_name": "issue-123-jwt-auth",
  "status": "in_progress"
}
```

### 3.2: File Creation/Editing
Agent/human performs work:
1. For **REQ1** (Requirements): Creates `docs/requirements/jwt-auth.md`
2. Agent calls File API: `POST /api/v1/git/files`
   ```json
   {
     "branch": "issue-123-jwt-auth",
     "path": "docs/requirements/jwt-auth.md",
     "content": "# JWT Authentication Requirements\n\n## Scope...",
     "message": "Add JWT authentication requirements",
     "author": "agent-req-001"
   }
   ```
3. System creates Git commit in branch

**Git Objects Created** (in ArangoDB):
```json
// Blob (file content)
{
  "_key": "blob-sha1-abc123",
  "type": "blob",
  "content": "# JWT Authentication Requirements...",
  "size": 2048
}

// Commit
{
  "_key": "commit-sha1-def456",
  "type": "commit",
  "tree": "tree-sha1-xyz789",
  "parent": "commit-sha1-prev",
  "author": "agent-req-001",
  "message": "Add JWT authentication requirements",
  "timestamp": "2025-11-26T10:15:00Z"
}
```

### 3.3: Iterative Work
Agent/human continues:
- Multiple commits to branch
- Creates/edits multiple files
- All commits linked to issue via branch

---

## Step 4: Review Submission (Agent Auto-creates PR)

**Actor**: Agent (automated) or Human (manual)  
**Trigger**: Work completion

**Process**:

### 4.1: Agent Determines Completion
Agent logic:
```
If all deliverables from Work Item Definition met:
  - Requirements document created ✓
  - Stakeholder sign-off obtained ✓ (simulated)
  - Technical feasibility notes added ✓
Then:
  Create Pull Request
```

### 4.2: Pull Request Creation
Agent calls API: `POST /api/v1/git/pull-requests`
```json
{
  "title": "Requirements: JWT Authentication",
  "description": "Completed requirements gathering for JWT auth feature.\n\nDeliverables:\n- Requirements document\n- Technical feasibility analysis\n\nRelated Issue: #123",
  "source_branch": "issue-123-jwt-auth",
  "target_branch": "main",
  "issue_id": "issue-123",
  "author": "agent-req-001"
}
```

**System Response**:
- Creates PR document in `pull_requests` collection
- Links PR to issue
- Updates issue status to `ready_for_review`
- Notifies reviewers

**Data Created**:
```json
// Pull Request
{
  "_key": "pr-456",
  "title": "Requirements: JWT Authentication",
  "source_branch": "issue-123-jwt-auth",
  "target_branch": "main",
  "issue_id": "issue-123",
  "author": "agent-req-001",
  "status": "open",
  "created_at": "2025-11-26T10:30:00Z",
  "reviewers": [],
  "approved_by": []
}

// Issue Update
{
  "_key": "issue-123",
  "pull_request_id": "pr-456",
  "status": "ready_for_review"
}
```

---

## Step 5: Human Review

**Actor**: Human reviewer  
**Location**: Pull Request Review UI

**Process**:
1. Reviewer navigates to PR #456
2. Reviews file changes:
   - `docs/requirements/jwt-auth.md` (diff view)
3. Checks against Work Item Definition deliverables
4. Options:
   - **Approve**: Clicks "Approve" button
   - **Request Changes**: Adds comments, agent revises
   - **Reject**: Closes PR, issue returns to REQ1

**Approval Flow**:
1. Reviewer clicks **"Approve"** button
2. System updates PR: `approved_by: ["user-alice"]`
3. Reviewer clicks **"Merge"** button

**System Response (Merge)**:
1. Performs Git merge operation:
   - Merges `issue-123-jwt-auth` → `main`
   - Creates merge commit
2. Updates PR status: `merged`
3. Deletes source branch (optional)
4. **Triggers Issue Progression** (critical - see [workflow-automation.md](workflow-automation.md))

---

## Step 6: Automated Issue Progression

**Trigger**: PR merge event  
**System**: Workflow orchestrator (see [workflow-automation.md](workflow-automation.md) for details)

**Quick Summary**:
1. Detect merge event
2. Determine next workflow step
3. Move issue to next column (REQ1 → REV1)
4. Reset issue for next assignment

**Full workflow automation logic**: See [workflow-automation.md](workflow-automation.md)

---

## Step 7: Workflow Continues (Repeat Steps 2-6)

**Next Stage: REV1 (Review and Validate)**

1. **Assignment**: SYSTEM-ARCHITECT agent claims issue
2. **Branch**: Agent creates `issue-123-jwt-auth-review`
3. **Work**: Agent reviews requirements doc, adds technical recommendations
4. **PR**: Agent creates PR with review findings
5. **Review**: Human approves
6. **Progression**: Issue moves to **ARCH1** (Design system architecture)

**Subsequent Stages**:
- **ARCH1**: System architecture design
- **IMPL1**: Go implementation
- **TEST1**: Testing
- **DEPLOY**: Deployment
- **DONE**: Completion

---

## Key Architecture Decisions

### 1. Entry Point Enforcement
- **All issues MUST start in REQ1** (requirements gathering)
- Cannot skip to IMPL1 directly
- Ensures proper requirements analysis for all work

### 2. Flexible Assignment Model
- **Manual Assignment**: Admin assigns specific agent/human
- **Claim-Based**: Workers pull work from queue
- Supports both managed and self-organizing teams

### 3. Git-Per-Issue Branching
- **Every issue gets a Git branch** when work starts
- Branch name: `issue-{id}-{slug}` (e.g., `issue-123-jwt-auth`)
- All commits associated with issue via branch
- Maintains clean Git history

### 4. PR-Driven Progression
- **PR merge triggers automatic issue progression**
- No manual "move to next column" needed
- Ensures work is reviewed and merged before advancement

### 5. Agent Auto-PR Creation
- **Agents automatically create PRs** when work complete
- Reduces manual overhead
- Ensures consistent review process

### 6. Pull-Based Workflow
- **Next worker pulls issue to their step**
- Not pushed automatically to next stage
- Allows capacity management and work prioritization

---

## Data Models

### Work Issue (Runtime)

```go
type WorkIssue struct {
    Key              string    `json:"_key"`
    Title            string    `json:"title"`
    Description      string    `json:"description"`
    WorkflowID       string    `json:"workflow_id"`       // Links to workflow
    CurrentStep      string    `json:"current_step"`      // REQ1, REV1, IMPL1, etc.
    Status           string    `json:"status"`            // open, assigned, in_progress, ready_for_review, completed
    AssignedTo       string    `json:"assigned_to"`       // agent-id or user-id
    BranchName       string    `json:"branch_name"`       // Git branch
    PullRequestID    string    `json:"pull_request_id"`   // Associated PR
    CompletedSteps   []string  `json:"completed_steps"`   // History
    CreatedBy        string    `json:"created_by"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

**Status Values**:
- `open` - Created, waiting for assignment
- `assigned` - Worker assigned but not started
- `in_progress` - Active work in Git branch
- `ready_for_review` - PR created, awaiting review
- `completed` - PR merged, work done

**CurrentStep Values**:
- Work item codes from agency spec (REQ1, REV1, IMPL1, TEST1, etc.)
- Determines which Kanban column issue appears in

---

## API Endpoints

### Issue Management

```
POST   /api/v1/issues
Request:
{
  "title": "Implement JWT authentication",
  "description": "We need JWT-based authentication...",
  "workflow_id": "software-dev-workflow-001"
}
Response:
{
  "_key": "issue-123",
  "current_step": "REQ1",
  "status": "open",
  "created_at": "2025-11-26T10:00:00Z"
}

GET    /api/v1/issues/:id
Response:
{
  "_key": "issue-123",
  "title": "Implement JWT authentication",
  "current_step": "REQ1",
  "status": "assigned",
  "assigned_to": "agent-req-001",
  "branch_name": "issue-123-jwt-auth"
}

PATCH  /api/v1/issues/:id
Request:
{
  "status": "in_progress",
  "branch_name": "issue-123-jwt-auth"
}

POST   /api/v1/issues/:id/assign
Request:
{
  "worker_id": "agent-req-001"  // or "user-bob"
}

POST   /api/v1/issues/:id/claim
Request: (empty - claims for authenticated user/agent)

GET    /api/v1/issues/available
Query params: ?step=REQ1&status=open
Response:
{
  "issues": [
    {
      "_key": "issue-123",
      "title": "Implement JWT authentication",
      "current_step": "REQ1",
      "status": "open"
    }
  ]
}
```

---

## Integration Points

### 1. Workflow Orchestrator
- Listens to PR merge events
- Determines next workflow step
- Moves issues between columns
- Triggers notifications

See [workflow-automation.md](workflow-automation.md) for orchestration logic.

### 2. Agent Factory
- Creates agent instances from role definitions
- Assigns agents to issues
- Monitors agent work completion

### 3. Git System
- Manages branches, commits, merges
- Stores all objects in ArangoDB
- Provides file explorer UI

See [git-operations.md](git-operations.md) and [file-explorer.md](file-explorer.md).

### 4. Notification System
- Notifies on issue assignment
- Alerts reviewers when PR created
- Notifies on issue progression

---

## UI Components

### Kanban Board View

```
┌─────────────────────────────────────────────────────────┐
│ Software Development Workflow                           │
├──────────┬──────────┬──────────┬──────────┬─────────────┤
│   REQ1   │   REV1   │  ARCH1   │  IMPL1   │    TEST1    │
├──────────┼──────────┼──────────┼──────────┼─────────────┤
│          │          │          │          │             │
│ ┌──────┐ │          │          │          │             │
│ │#123  │ │          │          │          │             │
│ │JWT   │ │          │          │          │             │
│ │Auth  │ │          │          │          │             │
│ │      │ │          │          │          │             │
│ │@agent│ │          │          │          │             │
│ └──────┘ │          │          │          │             │
│          │          │          │          │             │
│ [+ New]  │          │          │          │             │
└──────────┴──────────┴──────────┴──────────┴─────────────┘
```

### Issue Card

```
┌────────────────────────────────┐
│ #123 Implement JWT Auth        │
├────────────────────────────────┤
│ 📋 REQ1 - Requirements         │
│ 👤 @agent-req-001              │
│ 🌿 issue-123-jwt-auth          │
│ 📝 PR #456 (ready_for_review)  │
├────────────────────────────────┤
│ [View Details] [Claim]         │
└────────────────────────────────┘
```

---

## Related Documentation

- **Workflow Automation**: [workflow-automation.md](workflow-automation.md) - Progression logic, orchestration
- **Git Operations**: [git-operations.md](git-operations.md) - Branch/commit management
- **Pull Requests**: [pull-requests.md](pull-requests.md) - Review workflow
- **Work Item Schema**: [work-item-schema.md](work-item-schema.md) - Work item definitions
- **Implementation Guide**: [implementation-guide.md](implementation-guide.md) - Development roadmap

---

**Last Updated**: 2025-11-27  
**Status**: Architecture Defined - Ready for Implementation
