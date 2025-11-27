# Issue Lifecycle Workflow

**Related Tasks**: MVP-WI-008 (Workbench/Kanban Board), MVP-030 (Work Item Definitions), MVP-031 (Agent Assignment)  
**Status**: 📋 Planned - Architecture Defined

## Overview

This document describes the complete end-to-end workflow of an issue (work item) progressing through a software development lifecycle, from creation to completion. It demonstrates how issues, work items, Git integration, pull requests, and agent orchestration work together.

**UI Location**: Workbench (navbar link after Instances)

**Key Concepts**:
- **Issues**: Runtime work instances created by users on Workbench board
- **Work Items**: Templates defined in agency specs (REQ1, IMPL1, TEST1, etc.)
- **Workflow**: Sequence of work item steps defining the process
- **Assignment**: Manual or claim-based worker allocation
- **Git Integration**: All work produces commits/PRs in internal Git system
- **Progression**: Automatic advancement through workflow stages on PR merge

**For data models and API**, see [Issue Data Models & API](issue-data-models-api.md).

---

## Complete Feature Request Workflow

### Use Case: "Implement JWT Authentication"

This traces a feature request from creation to completion through the software development workflow.

**Workflow Definition** (from agency specification):
```
Software Development Workflow:
  REQ1 (Requirements Gathering) → 
  REV1 (Requirements Review) → 
  ARCH1 (Architecture Design) → 
  IMPL1 (Implementation) → 
  TEST1 (Testing) → 
  DEPLOY (Deployment) → 
  DONE (Complete)
```

---

## Step 1: Issue Creation (User Action)

**Actor**: Human user  
**Location**: Workbench UI (navbar link after Instances)

### Process

1. User navigates to agency instance Workbench
2. Clicks **"Create Issue"** button
3. Fills out issue creation form:
   ```
   Title: "Implement JWT authentication"
   Description: "We need JWT-based authentication for API endpoints.
                 Requirements: HS256 algorithm, 1-hour expiry, refresh tokens..."
   ```
4. Submits issue

### System Response

- Creates `work_issue` document in ArangoDB `work_issues` collection
- **Automatically places issue in REQ1 column** (enforced entry point)
- Issue assigned unique ID (e.g., `issue-123`)
- Status: `open`, CurrentStep: `REQ1`

### Data State

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
  "pull_request_id": null,
  "completed_steps": []
}
```

### UI State

- Issue appears as card in **REQ1** column of Workbench board
- Available for assignment/claiming
- Card shows: title, status, creation date

---

## Step 2: Work Assignment (Flexible Model)

**Actors**: Admin/manager or Agent/human worker  
**Location**: Workbench or Work Queue

### Two Assignment Modes

#### Mode A: Manual Assignment

1. Admin/manager clicks on issue card
2. Clicks **"Assign Agent"** or **"Assign Worker"** button
3. Selects from dropdown:
   - **Agents**: `REQUIREMENTS-ANALYST` (from agency role definitions)
   - **Humans**: `user-bob`, `user-charlie`
4. Worker assigned to issue

#### Mode B: Claim-Based (Pull Model)

1. Issue sits unassigned in REQ1 column
2. Available agents/humans see issue in **"Available Work"** queue
3. Agent or human clicks **"Claim Work"** on issue
4. First to claim gets assigned

### System Response

- Updates `work_issue.assigned_to` field
- If agent: Creates agent instance from role definition in agency spec
- Notifies assigned worker (email/webhook/dashboard notification)
- Issue status: `assigned`

### Data State

```json
{
  "_key": "issue-123",
  "assigned_to": "agent-req-001",
  "status": "assigned",
  "updated_at": "2025-11-26T10:05:00Z"
}
```

---

## Step 3: Work Execution - Git Integration

**Actor**: Agent (`agent-req-001`) or human worker  
**Location**: Git repository, File Explorer UI

### 3.1: Branch Creation

Agent/worker starts work on issue:

1. System **automatically creates Git branch**: `issue-123-jwt-auth`
2. Branch created from current `main` branch head
3. Issue status: `in_progress`

**Git Operations**:
```go
// Get current main commit
mainCommit := gitOps.GetRef(ctx, repoID, "refs/heads/main")

// Create issue branch
gitOps.UpdateRef(ctx, repoID, "refs/heads/issue-123-jwt-auth", mainCommit)
```

**Data State**:
```json
{
  "_key": "issue-123",
  "status": "in_progress",
  "branch_name": "issue-123-jwt-auth",
  "updated_at": "2025-11-26T10:10:00Z"
}
```

### 3.2: File Creation/Editing

Agent creates requirements document:

1. Agent creates new file: `documents/jwt-authentication-requirements.md`
2. Agent writes requirements content:
   ```markdown
   # JWT Authentication Requirements
   
   ## Overview
   Implement JWT-based authentication for all API endpoints.
   
   ## Technical Requirements
   1. Algorithm: HS256
   2. Token expiry: 1 hour
   3. Refresh tokens: 7-day expiry
   4. Secret key: Environment variable JWT_SECRET
   
   ## Acceptance Criteria
   - [ ] POST /auth/login returns JWT token
   - [ ] All API endpoints validate JWT
   - [ ] Invalid tokens return 401 Unauthorized
   ```

3. Agent commits file to branch:
   ```go
   // Write blob
   blobSHA := gitOps.WriteBlob(ctx, repoID, content)
   
   // Create tree
   entries := []TreeEntry{
       {Mode: "100644", Type: "blob", SHA: blobSHA, Path: "jwt-authentication-requirements.md"},
   }
   treeSHA := gitOps.WriteTree(ctx, repoID, entries)
   
   // Create commit
   author := Author{Name: "agent-req-001", Email: "agent@codevaldcortex"}
   commitSHA := gitOps.Commit(ctx, repoID, treeSHA, []string{parentCommit}, author, 
       "Add JWT authentication requirements")
   
   // Update branch
   gitOps.UpdateRef(ctx, repoID, "refs/heads/issue-123-jwt-auth", commitSHA)
   ```

### 3.3: Iterative Work

- Agent can make multiple commits on the branch
- Each commit advances the branch pointer
- All work isolated from `main` branch
- Work visible in File Explorer UI (filtered by branch)

---

## Step 4: Review Submission (Agent Auto-creates PR)

**Actor**: Agent (`agent-req-001`)  
**Location**: Pull Request system

### 4.1: Agent Determines Completion

Agent uses AI reasoning to determine if work is complete:

```
Agent Self-Assessment:
- ✅ Requirements document created
- ✅ All acceptance criteria listed
- ✅ Technical specifications defined
- ✅ No missing requirements identified
→ Work complete, ready for review
```

### 4.2: Pull Request Creation

Agent **automatically creates pull request**:

```go
// Create PR
pr := &PullRequest{
    RepoID:       repoID,
    InstanceID:   instanceID,
    Title:        "Requirements: JWT Authentication",
    Description:  "Added comprehensive requirements document for JWT authentication implementation.",
    SourceBranch: "refs/heads/issue-123-jwt-auth",
    TargetBranch: "refs/heads/main",
    BaseCommit:   mainCommit,
    HeadCommit:   branchCommit,
    Status:       "open",
    CreatedBy:    "agent-req-001",
    Reviewers:    []string{"user-alice"},  // Defined in workflow or agency spec
}

prID := createPullRequest(ctx, pr)
```

### System Updates Issue

- Issue `pull_request_id`: `pr-456`
- Issue `status`: `ready_for_review`

**Data State**:
```json
{
  "_key": "issue-123",
  "status": "ready_for_review",
  "pull_request_id": "pr-456",
  "updated_at": "2025-11-26T10:30:00Z"
}
```

---

## Step 5: Human Review

**Actor**: Human reviewer (`user-alice`)  
**Location**: Pull Request UI

### Review Process

1. Reviewer navigates to PR #456
2. Views diff (changes in requirements document)
3. Reads agent's work:
   - Requirements coverage
   - Technical specifications
   - Acceptance criteria
4. Reviewer options:
   - **Approve**: Merge PR
   - **Request Changes**: Agent revises work
   - **Reject**: Close PR without merge

### Approval & Merge

Reviewer approves and merges PR:

```go
// Merge PR (fast-forward or three-way merge)
mergeCommit := mergePullRequest(ctx, prID)

// Update main branch
gitOps.UpdateRef(ctx, repoID, "refs/heads/main", mergeCommit)

// Update PR status
updatePRStatus(ctx, prID, "merged")
```

**Pull Request State**:
```json
{
  "_key": "pr-456",
  "status": "merged",
  "merge_commit": "abc789def012...",
  "approved_by": ["user-alice"],
  "updated_at": "2025-11-26T11:00:00Z"
}
```

---

## Step 6: Automated Issue Progression

**Actor**: Workflow Orchestrator (system)  
**Location**: Event handler listening to PR merge events

### Automatic Progression Logic

1. **Event Trigger**: PR #456 merged
2. **Lookup Issue**: Find issue linked to PR (`issue-123`)
3. **Determine Next Step**:
   - Current step: `REQ1`
   - Workflow definition: `REQ1 → REV1 → ARCH1 → ...`
   - Next step: `REV1` (Requirements Review)
4. **Update Issue**:
   ```go
   issue.CurrentStep = "REV1"
   issue.Status = "open"
   issue.AssignedTo = nil  // Clear assignment
   issue.CompletedSteps = append(issue.CompletedSteps, "REQ1")
   issue.BranchName = nil
   issue.PullRequestID = nil
   ```

### Pull-Based Next Step

- Issue **does NOT auto-assign** to next worker
- Issue appears in **REV1** column as `open`
- Next worker (reviewer agent or human) **claims or is assigned** to issue
- Allows capacity management and work prioritization

**Data State**:
```json
{
  "_key": "issue-123",
  "current_step": "REV1",
  "status": "open",
  "assigned_to": null,
  "branch_name": null,
  "pull_request_id": null,
  "completed_steps": ["REQ1"],
  "updated_at": "2025-11-26T11:05:00Z"
}
```

**UI State**:
- Issue card **moves from REQ1 column to REV1 column** on Workbench board
- Available for assignment/claiming in REV1

---

## Step 7: Workflow Continues (Repeat Steps 2-6)

Issue progresses through remaining workflow steps:

### REV1 (Requirements Review)

1. **Assignment**: Reviewer agent/human claims issue
2. **Branch**: `issue-123-jwt-auth-review`
3. **Work**: Agent reviews requirements doc, adds technical recommendations
4. **PR**: Agent creates PR with review findings
5. **Review**: Human approves
6. **Progression**: Issue moves to **ARCH1** (Architecture Design)

### ARCH1 (Architecture Design)

1. **Assignment**: Architecture agent/human claims issue
2. **Branch**: `issue-123-jwt-auth-arch`
3. **Work**: Agent designs system architecture (auth middleware, token service, database schema)
4. **PR**: Agent creates PR with architecture documents
5. **Review**: Human architect approves
6. **Progression**: Issue moves to **IMPL1** (Implementation)

### IMPL1 (Implementation)

1. **Assignment**: Developer agent/human claims issue
2. **Branch**: `issue-123-jwt-auth-impl`
3. **Work**: Agent implements JWT authentication code (Go handlers, middleware, tests)
4. **PR**: Agent creates PR with implementation
5. **Review**: Human developer reviews code
6. **Progression**: Issue moves to **TEST1** (Testing)

### TEST1 (Testing)

1. **Assignment**: QA agent/human claims issue
2. **Branch**: `issue-123-jwt-auth-test`
3. **Work**: Agent writes integration/E2E tests
4. **PR**: Agent creates PR with test suite
5. **Review**: Human QA approves
6. **Progression**: Issue moves to **DEPLOY** (Deployment)

### DEPLOY (Deployment)

1. **Assignment**: DevOps agent/human claims issue
2. **Work**: Deploy to staging/production
3. **Verification**: Smoke tests pass
4. **Progression**: Issue moves to **DONE** (Complete)

### DONE (Complete)

- Issue marked as **completed**
- No further work required
- Archived or removed from active Workbench

---

## Key Architecture Decisions

### 1. Entry Point Enforcement

- **All issues MUST start in REQ1** (requirements gathering)
- Cannot skip directly to IMPL1 or TEST1
- Ensures proper requirements analysis for all work
- Enforced at issue creation (backend validation)

### 2. Flexible Assignment Model

- **Manual Assignment**: Admin assigns specific agent/human
- **Claim-Based**: Workers pull work from queue
- Supports both managed teams and self-organizing teams
- No forced assignment on progression (pull-based workflow)

### 3. Git-Per-Issue Branching

- **Every issue gets a Git branch** when work starts
- Branch name: `issue-{id}-{slug}` (e.g., `issue-123-jwt-auth`)
- All commits associated with issue via branch
- Maintains clean Git history
- Branches deleted after PR merge (optional retention)

### 4. PR-Driven Progression

- **PR merge triggers automatic issue progression**
- No manual "move to next column" button needed
- Ensures work is reviewed and merged before advancement
- Prevents skipping review steps

### 5. Agent Auto-PR Creation

- **Agents automatically create PRs** when work complete
- Reduces manual overhead
- Ensures consistent review process
- Agents cannot merge own PRs (requires human approval)

### 6. Pull-Based Workflow

- **Next worker pulls issue to their step**
- Not pushed automatically to next worker
- Allows capacity management (don't overload workers)
- Supports work prioritization (workers choose critical issues first)

### 7. Audit Trail

- All Git commits preserved (full history)
- Pull requests retained (review history)
- Issue state transitions logged (`completed_steps` array)
- Complete traceability from request to deployment

---

## Related Documentation

- **Issue Data Models & API**: [issue-data-models-api.md](issue-data-models-api.md) - Models, endpoints, UI
- **Workflow Automation**: [workflow-automation.md](workflow-automation.md) - Orchestration logic
- **Git Core Operations**: [git-core-operations.md](git-core-operations.md) - Branch/commit management
- **Pull Requests**: [pull-requests.md](pull-requests.md) - Review workflow details
- **Work Item Schema**: [work-item-schema.md](work-item-schema.md) - Work item definitions
- **Kanban Workflow**: [kanban-workflow.md](kanban-workflow.md) - Board generation from workflows

---

**Last Updated**: 2025-11-27  
**Status**: Architecture Defined - Ready for Implementation
