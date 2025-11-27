# Issue Data Models & API

**Related Tasks**: MVP-WI-008 (Workbench/Kanban Board), MVP-030 (Work Item Definitions), MVP-031 (Agent Assignment)  
**Status**: 📋 Planned - Architecture Defined

## Overview

This document defines data models, API endpoints, and UI components for issue management in CodeValdCortex. Issues are runtime instances of work items that progress through workflows.

**For complete workflow walkthrough**, see [Issue Lifecycle Workflow](issue-lifecycle-workflow.md).

---

## Data Models

### WorkIssue (Runtime Work Instance)

```go
package models

import "time"

type WorkIssue struct {
    // ArangoDB document fields
    Key string `json:"_key"`       // Issue identifier (issue-123)
    ID  string `json:"_id"`
    Rev string `json:"_rev"`
    
    // Issue metadata
    Title       string `json:"title"`              // "Implement JWT authentication"
    Description string `json:"description"`        // Detailed description (Markdown)
    
    // Workflow context
    WorkflowID  string `json:"workflow_id"`        // Links to workflow definition
    CurrentStep string `json:"current_step"`       // Current work item code (REQ1, IMPL1, etc.)
    Status      string `json:"status"`             // Issue status (see below)
    
    // Assignment
    AssignedTo string `json:"assigned_to,omitempty"` // Agent ID or user ID
    
    // Git integration
    BranchName     string `json:"branch_name,omitempty"`      // Git branch (issue-123-jwt-auth)
    PullRequestID  string `json:"pull_request_id,omitempty"`  // Associated PR
    
    // History tracking
    CompletedSteps []string `json:"completed_steps"` // Work items completed (REQ1, REV1, ...)
    
    // Audit fields
    CreatedBy string    `json:"created_by"`  // User/agent who created issue
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Status Values**:
- `open` - Created, waiting for assignment
- `assigned` - Worker assigned but work not started
- `in_progress` - Active work in Git branch
- `ready_for_review` - PR created, awaiting review
- `completed` - PR merged, all workflow steps done
- `blocked` - Cannot proceed (dependency or issue)

**CurrentStep Values**:
- Work item codes from agency specification (REQ1, REV1, ARCH1, IMPL1, TEST1, DEPLOY, etc.)
- Determines which Workbench column issue appears in

**Constraints**:
- `workflow_id` must reference existing workflow in `workflows` collection
- `current_step` must exist in workflow definition
- New issues MUST start at first workflow step (entry point enforcement)

---

### WorkbenchBoard (Generated from Workflow)

```go
type WorkbenchBoard struct {
    AgencyID    string   `json:"agency_id"`    // Agency identifier
    WorkflowID  string   `json:"workflow_id"`  // Workflow definition ID
    TagKey      string   `json:"tag_key"`      // Agency tag (specification snapshot)
    Columns     []Column `json:"columns"`      // Kanban columns
    GeneratedAt time.Time `json:"generated_at"`
}

type Column struct {
    ID           string      `json:"id"`             // Column identifier
    Name         string      `json:"name"`           // Display name (Requirements Gathering)
    WorkItemCode string      `json:"work_item_code"` // Work item code (REQ1)
    Order        int         `json:"order"`          // Display order (0, 1, 2, ...)
    Issues       []WorkIssue `json:"issues"`         // Issues in this column
}
```

**Generation Logic**:
1. Read workflow from published tag snapshot
2. For each workflow step:
   - Create column with work item code
   - Fetch work item details from specification
   - Query issues with `current_step == work_item_code`
3. Order columns by workflow step sequence

**Example Board** (software development workflow):
```json
{
  "agency_id": "UC-INFRA-001",
  "workflow_id": "software-dev-workflow-001",
  "tag_key": "v1.2.0",
  "columns": [
    {
      "id": "col-req1",
      "name": "Requirements Gathering",
      "work_item_code": "REQ1",
      "order": 0,
      "issues": [
        {
          "_key": "issue-123",
          "title": "Implement JWT authentication",
          "status": "open",
          "current_step": "REQ1"
        }
      ]
    },
    {
      "id": "col-rev1",
      "name": "Requirements Review",
      "work_item_code": "REV1",
      "order": 1,
      "issues": []
    }
  ],
  "generated_at": "2025-11-27T10:00:00Z"
}
```

---

## API Endpoints

### Issue Management

#### Create Issue

```
POST /api/v1/agencies/:agencyID/instances/:instanceID/issues
```

**Request Body**:
```json
{
  "title": "Implement JWT authentication",
  "description": "We need JWT-based authentication for API endpoints...",
  "workflow_id": "software-dev-workflow-001"
}
```

**Response** (201 Created):
```json
{
  "_key": "issue-123",
  "title": "Implement JWT authentication",
  "description": "We need JWT-based authentication...",
  "workflow_id": "software-dev-workflow-001",
  "current_step": "REQ1",
  "status": "open",
  "created_by": "user-alice",
  "created_at": "2025-11-27T10:00:00Z",
  "updated_at": "2025-11-27T10:00:00Z",
  "completed_steps": []
}
```

**Validation**:
- `workflow_id` must exist in workflows collection
- `current_step` automatically set to first workflow step
- `status` automatically set to `open`

---

#### Get Issue

```
GET /api/v1/agencies/:agencyID/instances/:instanceID/issues/:issueID
```

**Response** (200 OK):
```json
{
  "_key": "issue-123",
  "title": "Implement JWT authentication",
  "current_step": "REQ1",
  "status": "assigned",
  "assigned_to": "agent-req-001",
  "branch_name": "issue-123-jwt-auth",
  "pull_request_id": null,
  "completed_steps": [],
  "created_by": "user-alice",
  "created_at": "2025-11-27T10:00:00Z",
  "updated_at": "2025-11-27T10:05:00Z"
}
```

---

#### Update Issue

```
PATCH /api/v1/agencies/:agencyID/instances/:instanceID/issues/:issueID
```

**Request Body** (partial update):
```json
{
  "status": "in_progress",
  "branch_name": "issue-123-jwt-auth"
}
```

**Response** (200 OK):
```json
{
  "_key": "issue-123",
  "status": "in_progress",
  "branch_name": "issue-123-jwt-auth",
  "updated_at": "2025-11-27T10:10:00Z"
}
```

**Allowed Fields**:
- `status`, `assigned_to`, `branch_name`, `pull_request_id`
- `current_step` can only be updated by workflow orchestrator
- `completed_steps` managed by system only

---

#### Assign Issue

```
POST /api/v1/agencies/:agencyID/instances/:instanceID/issues/:issueID/assign
```

**Request Body**:
```json
{
  "worker_id": "agent-req-001"
}
```

**Response** (200 OK):
```json
{
  "_key": "issue-123",
  "assigned_to": "agent-req-001",
  "status": "assigned",
  "updated_at": "2025-11-27T10:05:00Z"
}
```

**Logic**:
- Creates agent instance if `worker_id` references agency role
- Updates issue status to `assigned`
- Sends notification to assigned worker

---

#### Claim Issue (Self-Assignment)

```
POST /api/v1/agencies/:agencyID/instances/:instanceID/issues/:issueID/claim
```

**Request**: Empty body (uses authenticated user/agent)

**Response** (200 OK):
```json
{
  "_key": "issue-123",
  "assigned_to": "user-bob",
  "status": "assigned",
  "updated_at": "2025-11-27T10:05:00Z"
}
```

**Constraints**:
- Issue must be `open` (not already assigned)
- Requester must have permission to work on this workflow step

---

#### List Available Issues

```
GET /api/v1/agencies/:agencyID/instances/:instanceID/issues/available
```

**Query Parameters**:
- `step` (optional) - Filter by workflow step (e.g., `REQ1`)
- `status` (optional) - Filter by status (e.g., `open`)
- `limit` (optional) - Max results (default: 50)

**Response** (200 OK):
```json
{
  "issues": [
    {
      "_key": "issue-123",
      "title": "Implement JWT authentication",
      "current_step": "REQ1",
      "status": "open",
      "created_at": "2025-11-27T10:00:00Z"
    },
    {
      "_key": "issue-124",
      "title": "Add API rate limiting",
      "current_step": "REQ1",
      "status": "open",
      "created_at": "2025-11-27T09:45:00Z"
    }
  ],
  "total": 2
}
```

**Use Case**: Work queue for agents/humans to claim available work

---

### Workbench Board

#### Get Workbench Board

```
GET /api/v1/agencies/:agencyID/instances/:instanceID/workbench
```

**Query Parameters**:
- `tag` (optional) - Agency tag key (defaults to latest published tag)

**Response** (200 OK):
```json
{
  "agency_id": "UC-INFRA-001",
  "workflow_id": "software-dev-workflow-001",
  "tag_key": "v1.2.0",
  "columns": [
    {
      "id": "col-req1",
      "name": "Requirements Gathering",
      "work_item_code": "REQ1",
      "order": 0,
      "issues": [
        {
          "_key": "issue-123",
          "title": "Implement JWT authentication",
          "status": "open",
          "assigned_to": null,
          "created_at": "2025-11-27T10:00:00Z"
        }
      ]
    },
    {
      "id": "col-rev1",
      "name": "Requirements Review",
      "work_item_code": "REV1",
      "order": 1,
      "issues": []
    },
    {
      "id": "col-arch1",
      "name": "Architecture Design",
      "work_item_code": "ARCH1",
      "order": 2,
      "issues": []
    }
  ],
  "generated_at": "2025-11-27T10:00:00Z"
}
```

**Generation Logic**:
1. Fetch agency tag snapshot (or latest published tag)
2. Extract workflow definition from tag
3. For each workflow step:
   - Query work item from specification
   - Query issues with `current_step == work_item_code`
4. Return generated board structure

---

## UI Components

### Workbench Board View

**Kanban-style board** with columns for each workflow step (REQ1, REV1, ARCH1, IMPL1, TEST1, etc.).

**Features**:
- Drag-and-drop disabled (auto-progression only via PR merge)
- Click card to view details
- **[+ New Issue]** button in first column only (entry point enforcement)
- Real-time updates (HTMX polling or WebSocket)
- Column headers show work item names
- Issue count per column visible

---

### Issue Card Component

**Card displays**:
- Issue number (#123) and title (truncated to 40 chars)
- Current workflow step (REQ1 - Requirements)
- Status icon: 📋 open, 🔄 in_progress, ✅ ready_for_review, ✅ completed
- Assigned worker (@agent-req-001 or @user-bob)
- Git branch name (if exists): `issue-123-jwt-auth`
- PR status (if exists): PR #456 (ready_for_review)

**Actions**:
- **View Details**: Open issue detail modal/page
- **Claim**: Self-assign issue (if unassigned)
- **Assign**: Admin action to assign specific worker

---

### Create Issue Modal

**Form fields**:
- **Title** (required, max 200 chars)
- **Description** (optional, Markdown editor)
- **Workflow** (dropdown selector, required)

**Validation**:
- Title cannot be empty
- Workflow must be selected
- Issue automatically created in first workflow step
- User sees confirmation with issue number

---

### Issue Detail View

**Modal/page displaying**:
- Issue metadata: Status (🔄 In Progress), Current step (REQ1), Assigned to (@agent-req-001), Branch name
- Description (Markdown rendered)
- Workflow progress tracker: ✅ completed steps, 🔄 current step, ⬜ pending steps
- Related artifacts: [View Files] [View PR #456]
- Actions: [Reassign] [Close Issue]

---

## Integration Points

### 1. Workflow Orchestrator
Listens to PR merge events. Determines next workflow step, updates issue `current_step` and `status`, clears assignment/branch/PR fields, triggers notifications. See [workflow-automation.md](workflow-automation.md).

### 2. Agent Factory
Listens to issue assignment events. Creates agent instances from role definitions, initializes context, starts execution loop.

### 3. Git System
Used by agents/humans for work execution. Operations: Create branch, commit changes, read/write files, create PRs. See [git-core-operations.md](git-core-operations.md), [file-explorer.md](file-explorer.md).

### 4. Notification System
**Triggers**: Issue created/assigned, PR created/merged, issue progressed.  
**Channels**: Email, webhooks, in-app notifications, dashboard alerts.

---

## Related Documentation

- **Issue Lifecycle Workflow**: [issue-lifecycle-workflow.md](issue-lifecycle-workflow.md) - Complete workflow walkthrough
- **Workflow Automation**: [workflow-automation.md](workflow-automation.md) - Orchestration logic
- **Git Core Operations**: [git-core-operations.md](git-core-operations.md) - Git integration
- **Pull Requests**: [pull-requests.md](pull-requests.md) - Review workflow
- **Kanban Workflow**: [kanban-workflow.md](kanban-workflow.md) - Board generation details

---

**Last Updated**: 2025-11-27  
**Status**: Architecture Defined - Ready for Implementation
