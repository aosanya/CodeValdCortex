# Kanban Workflow & Issue Lifecycle

**Status**: 📋 Planned - Architecture Defined  
**Related**: [git-based-document-system.md](./git-based-document-system.md), [work-item-schema.md](./work-item-schema.md)

## Overview

This document describes the complete workflow for work items (issues) in CodeValdCortex, from creation through completion. The system uses a **Kanban-based workflow** where issues progress through stages defined in agency specifications, with work performed by humans or AI agents, and all changes tracked in the internal Git-in-ArangoDB system.

**Key Concepts**:
- **Work Item Definitions**: Templates defined in agency specs (REQ1, IMPL1, etc.)
- **Issues**: Runtime instances created by users on Kanban board
- **Workflows**: Sequential steps that issues progress through
- **Agents/Humans**: Workers who claim and execute work on issues
- **Git Integration**: All work results in commits/PRs in internal Git system

---

## Complete Feature Request Workflow

### Use Case: "Implement JWT Authentication"

This traces a feature request from creation to completion through the software development workflow.

---

### Step 1: Issue Creation (User Action)

**Actor**: Human user  
**Location**: Agency Kanban Board UI

**Process**:
1. User navigates to agency instance Kanban board
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
- Issue appears as card in **REQ1** column of Kanban board
- Available for assignment/claiming

---

### Step 2: Work Assignment (Flexible Model)

**Actors**: Admin or Agent/Human worker  
**Location**: Kanban Board or Work Queue

**Two Assignment Modes**:

#### Mode A: Manual Assignment
1. Admin/manager clicks on issue card
2. Clicks **"Assign Agent"** or **"Assign Worker"** button
3. Selects from dropdown:
   - Agents: `REQUIREMENTS-ANALYST` (from agency roles)
   - Humans: `user-bob`, `user-charlie`
4. Worker assigned to issue

#### Mode B: Claim-Based (Pull Model)
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

### Step 3: Work Execution - Git Integration

**Actor**: Assigned agent or human  
**Location**: Agent runtime or human workspace

**Process**:

#### 3.1: Branch Creation
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

#### 3.2: File Creation/Editing
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

#### 3.3: Iterative Work
Agent/human continues:
- Multiple commits to branch
- Creates/edits multiple files
- All commits linked to issue via branch

---

### Step 4: Review Submission (Agent Auto-creates PR)

**Actor**: Agent (automated) or Human (manual)  
**Trigger**: Work completion

**Process**:

#### 4.1: Agent Determines Completion
Agent logic:
```
If all deliverables from Work Item Definition met:
  - Requirements document created ✓
  - Stakeholder sign-off obtained ✓ (simulated)
  - Technical feasibility notes added ✓
Then:
  Create Pull Request
```

#### 4.2: Pull Request Creation
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

### Step 5: Human Review

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
4. **Triggers Issue Progression** (critical)

---

### Step 6: Automated Issue Progression

**Trigger**: PR merge event  
**System**: Workflow orchestrator

**Process**:

#### 6.1: Detect Merge Event
```
Event: pull_request.merged
Data: {
  pr_id: "pr-456",
  issue_id: "issue-123",
  merged_at: "2025-11-26T11:00:00Z"
}
```

#### 6.2: Determine Next Step
Workflow logic:
```javascript
// Get current workflow and step
workflow = getWorkflow(issue.workflow_id)
currentStep = workflow.steps.find(s => s.work_item_id === issue.current_step)
currentStepIndex = workflow.steps.indexOf(currentStep)

// Get next step
nextStep = workflow.steps[currentStepIndex + 1]
// nextStep = REV1 (Review and validate technical specification)
```

#### 6.3: Move Issue to Next Column
```json
// Issue Update
{
  "_key": "issue-123",
  "current_step": "REV1",  // Moved from REQ1
  "status": "open",         // Ready for next assignment
  "assigned_to": null,      // Unassigned, ready to claim
  "branch_name": null,      // Reset for next work
  "pull_request_id": "pr-456",  // Keep history
  "completed_steps": ["REQ1"],
  "updated_at": "2025-11-26T11:00:00Z"
}
```

**UI Update**:
- Issue card moves from **REQ1** column → **REV1** column
- Available for next agent/human to claim

---

### Step 7: Workflow Continues (Repeat Steps 2-6)

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

## Workflow Configuration

### Agency Specification Structure

From actual agency spec JSON:

```json
{
  "work_items": [
    {
      "_key": "4268284c-ebfd-4e72-9f24-cb190ede3103",
      "code": "REQ1",
      "title": "Conduct stakeholder requirements gathering session",
      "description": "Execute structured interviews...",
      "deliverables": [
        "Requirements document",
        "Stakeholder sign-off",
        "Technical feasibility notes"
      ],
      "goal_keys": ["9188cb44-6aa5-450d-9164-ce9bb1334b2d"],
      "tags": ["requirements", "stakeholders", "analysis"]
    }
  ],
  "workflows": [
    {
      "_key": "workflow-001",
      "name": "Software Development Workflow",
      "steps": [
        {
          "order": 0,
          "items": [
            {"work_item_id": "REQ1"},  // Entry point
            {"work_item_id": "REV1"}
          ]
        },
        {
          "order": 1,
          "items": [
            {"work_item_id": "IMPL1"}
          ]
        }
      ]
    }
  ],
  "roles": [
    {
      "code": "REQUIREMENTS-ANALYST",
      "name": "Requirements Analyst",
      "description": "Specialized agent for gathering requirements",
      "autonomy_level": "semi-autonomous",
      "token_budget": 800000
    }
  ]
}
```

### Work Item Definition → Issue Mapping

**Work Item Definition** (Design Time):
- Defined in agency specification
- Template/blueprint for work
- Specifies deliverables, goals, tags
- Example: REQ1, IMPL1, TEST1

**Issue** (Runtime):
- Created by users on Kanban board
- Instance of actual work to be done
- References Work Item Definition implicitly (via column/step)
- Tracks progress, assignments, Git artifacts

**Relationship**:
```
Agency Spec                    Runtime
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Work Item: REQ1     ──────────► Issue #123 (in REQ1 column)
  - Deliverables               - Title: "Implement JWT auth"
  - Goals                      - Assigned to: agent-req-001
  - Tags                       - Branch: issue-123-jwt-auth
                               - PR: pr-456
                               - Status: ready_for_review
```

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

### Pull Request

```go
type PullRequest struct {
    Key          string    `json:"_key"`
    Title        string    `json:"title"`
    Description  string    `json:"description"`
    SourceBranch string    `json:"source_branch"`
    TargetBranch string    `json:"target_branch"`
    IssueID      string    `json:"issue_id"`         // Links to work issue
    Author       string    `json:"author"`           // agent-id or user-id
    Status       string    `json:"status"`           // open, approved, merged, closed
    Reviewers    []string  `json:"reviewers"`
    ApprovedBy   []string  `json:"approved_by"`
    MergedAt     time.Time `json:"merged_at"`
    CreatedAt    time.Time `json:"created_at"`
}
```

---

## API Endpoints

### Issue Management

```
POST   /api/v1/issues                    # Create issue
GET    /api/v1/issues/:id                # Get issue details
PATCH  /api/v1/issues/:id                # Update issue
POST   /api/v1/issues/:id/assign         # Assign worker
POST   /api/v1/issues/:id/claim          # Claim work
GET    /api/v1/issues/available          # Get claimable issues
```

### Git Operations

```
POST   /api/v1/git/branches              # Create branch for issue
POST   /api/v1/git/files                 # Create/edit file in branch
POST   /api/v1/git/commits               # Commit changes
GET    /api/v1/git/branches/:name        # Get branch details
```

### Pull Requests

```
POST   /api/v1/git/pull-requests         # Create PR
GET    /api/v1/git/pull-requests/:id     # Get PR details
POST   /api/v1/git/pull-requests/:id/approve    # Approve PR
POST   /api/v1/git/pull-requests/:id/merge      # Merge PR
POST   /api/v1/git/pull-requests/:id/comment    # Add review comment
```

---

## Integration Points

### 1. Workflow Orchestrator
- Listens to PR merge events
- Determines next workflow step
- Moves issues between columns
- Triggers notifications

### 2. Agent Factory
- Creates agent instances from role definitions
- Assigns agents to issues
- Monitors agent work completion

### 3. Git System
- Manages branches, commits, merges
- Stores all objects in ArangoDB
- Provides file explorer UI

### 4. Notification System
- Notifies on issue assignment
- Alerts reviewers when PR created
- Notifies on issue progression

---

## Implementation Phases

See [implementation-guide.md](./implementation-guide.md) for detailed 7-phase roadmap.

**Phase 1**: Git Core Layer (branches, commits, blobs)  
**Phase 2**: File Explorer UI  
**Phase 3**: Pull Requests & Review  
**Phase 4**: **Kanban Workflow Integration** ← This document  
**Phase 5**: AI-Assisted Merging  
**Phase 6**: Agent Integration  
**Phase 7**: Advanced Features  

---

## Related Documentation

- **Git Implementation**: [git-based-document-system.md](./git-based-document-system.md)
- **Work Item Schema**: [work-item-schema.md](./work-item-schema.md)
- **Pull Requests**: [pull-requests.md](./pull-requests.md)
- **Implementation Roadmap**: [implementation-guide.md](./implementation-guide.md)
- **Architectural Decisions**: [architecture/vcs-integration-decisions.md](./architecture/vcs-integration-decisions.md)

---

**Last Updated**: 2025-11-26  
**Status**: Architecture Defined - Ready for Implementation  
**Next Step**: Phase 4 Implementation - Kanban Workflow Integration
