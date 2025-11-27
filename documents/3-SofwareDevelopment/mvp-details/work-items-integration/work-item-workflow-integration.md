# Work Item & Workflow Integration

<!-- MVP-030, MVP-031, MVP-032 -->
**Tasks Covered**: MVP-030 (Workflow-WorkItem Integration), MVP-031 (Issue Lifecycle), MVP-032 (Agent Assignment)  
**Status**: ✅ Architecture Aligned - Tag-Based Single Source of Truth

## Overview

This document defines how **WorkItems from Agency Specifications** integrate with **Workflows** to create the **Workbench** (Kanban-style board accessible via navbar after Instances) in CodeValdCortex. The system uses published agency tags as the single source of truth.

**Key Architectural Principle**: WorkItems are defined in the Agency Specification (not as separate types). When an agency is published/tagged, these WorkItems become Workbench columns through Workflow definitions that reference them.

### Architecture Flow

```
Agency Specification (internal/agency/models/specification.go)
  ├─> WorkItems: Array of work definitions
  │   Example: [
  │     {code: "WI-001", title: "Planning", description: "...", deliverables: [...], goal_keys: [...]}
  │     {code: "WI-002", title: "Development", description: "...", deliverables: [...], goal_keys: [...]}
  │     {code: "WI-003", title: "Review", description: "...", deliverables: [...], goal_keys: [...]}
  │   ]
  │
  └─> Workflows: Array of process definitions
      Example: [
        {
          name: "Development Workflow",
          description: "Standard development process",
          steps: [
            { work_item_code: "WI-001", name: "Planning Phase" },
            { work_item_code: "WI-002", name: "Implementation Phase" },
            { work_item_code: "WI-003", name: "Review Phase" }
          ]
        }
      ]
       ↓
  Publish/Tag (creates immutable snapshot)
       ↓
  Tag Snapshot (Single Source of Truth)
    specification: {
      work_items: [...],    # WorkItem definitions
      workflows: [...]      # Workflow definitions referencing WorkItems
    }
       ↓
  Workbench Board Generation (UI: navbar link after Instances)
    ↓
  Workbench Columns (created from Workflow steps)
    └─> Each column represents a WorkItem from the specification
         └─> Issues/Tickets move through these columns
              └─> Agent assignment based on WorkItem metadata
```

### Existing Models

**WorkItem** (`internal/agency/models/work_item.go`):
```go
type WorkItem struct {
    Key          string    `json:"_key,omitempty"`
    ID           string    `json:"_id,omitempty"`
    AgencyID     string    `json:"agency_id"`
    Number       int       `json:"number"`
    Code         string    `json:"code"`         // e.g., "WI-001"
    Title        string    `json:"title"`        // e.g., "Planning"
    Description  string    `json:"description"`  // Detailed explanation
    Deliverables []string  `json:"deliverables"` // Expected outputs
    GoalKeys     []string  `json:"goal_keys"`    // Links to agency goals
    Tags         []string  `json:"tags"`         // Categorization
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

**Workflow** (`internal/agency/models/workflow.go`):
```go
type Workflow struct {
    Key         string `json:"_key,omitempty"`
    ID          string `json:"_id,omitempty"`
    Name        string `json:"name"`              // e.g., "Development Workflow"
    Description string `json:"description"`
    AgencyID    string `json:"agency_id"`
    Version     string `json:"version"`
    Steps       Steps  `json:"steps"`             // Array of workflow steps
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Step struct {
    ID              string          `json:"id"`
    Name            string          `json:"name"`
    Type            string          `json:"type"`           // "work_item", "decision", "parallel"
    WorkItemCode    string          `json:"work_item_code"` // ← Links to WorkItem.Code
    Items           []ProcessItem   `json:"items,omitempty"`
    Transitions     []Transition    `json:"transitions,omitempty"`
}
```

### Relationship to Other Systems

- **Agency Specification**: WorkItems defined as part of agency design in specification document
- **Workflows**: Reference WorkItems by code to create ordered process flow
- **Tags**: Published snapshots containing complete specification (WorkItems + Workflows + Goals + Roles + RACI)
- **Workbench**: Generated from tag's workflow.steps, each step maps to a Workbench column (UI: navbar link after Instances)
- **Git System**: Issues created in Workbench link to branches/PRs for code/document changes
- **Agent Orchestration**: Issues trigger agent creation based on WorkItem deliverables and requirements

---

## MVP-030: Workflow-WorkItem Integration

### Objective

Connect Workflows to Agency Specification WorkItems to enable Workbench board generation from published agency tags.

### Problem

Currently:
- ✅ WorkItems exist in Agency Specification (`specification.work_items`)
- ✅ Workflows exist in Agency Specification (`specification.workflows`)
- ❌ Workflows don't explicitly reference WorkItems
- ❌ No clear mapping from Workflow steps to Workbench columns
- ❌ Workbench board generation logic missing

### Solution

**1. Enhance Workflow Steps to Reference WorkItems**

Update `Step` model to include WorkItem reference:
```go
type Step struct {
    ID              string   `json:"id"`
    Name            string   `json:"name"`
    Type            string   `json:"type"`           // "work_item", "decision", "parallel"
    WorkItemCode    string   `json:"work_item_code"` // NEW: Reference to WorkItem.Code
    Description     string   `json:"description,omitempty"`
    // ... rest of fields
}
```

**2. Agency Designer UI Enhancement**

Add workflow designer interface that:
- Lists all WorkItems from specification
- Allows drag-and-drop WorkItems into workflow steps
- Visualizes workflow as a Workbench board preview
- Validates WorkItem references (ensure Code exists in specification)

**3. Workbench Board Generation Service**

When an agency is activated (tag is used):
```go
// Pseudo-code
func GenerateWorkbenchBoard(tagSnapshot *models.TagSnapshot) (*WorkbenchBoard, error) {
    workItems := tagSnapshot.Specification.WorkItems
    workflows := tagSnapshot.Specification.Workflows
    
    // For each workflow, create a Workbench board
    for _, workflow := range workflows {
        board := &WorkbenchBoard{
            Name: workflow.Name,
            Columns: []Column{},
        }
        
        // Each workflow step becomes a Kanban column
        for _, step := range workflow.Steps {
            if step.Type == "work_item" {
                // Find the WorkItem by code
                workItem := findWorkItemByCode(workItems, step.WorkItemCode)
                
                column := &Column{
                    ID:            step.ID,
                    Name:          step.Name,
                    WorkItemCode:  step.WorkItemCode,
                    Deliverables:  workItem.Deliverables,
                    GoalKeys:      workItem.GoalKeys,
                }
                board.Columns = append(board.Columns, column)
            }
        }
    }
}
```

### Implementation Tasks

- [ ] Update `Step` model with `WorkItemCode` field
- [ ] Create workflow designer UI in Agency Designer
  - [ ] WorkItem list panel (shows all specification.work_items)
  - [ ] Workflow canvas (drag-drop to create steps)
  - [ ] Step configuration modal (link to WorkItem)
  - [ ] Workflow validation (ensure all WorkItem codes exist)
- [ ] Implement Workbench board generation service
  - [ ] Read tag snapshot
  - [ ] Parse workflows and WorkItems
  - [ ] Create board structure
  - [ ] Store in agency-specific database
- [ ] API endpoints for workflow-workitem management
  - [ ] GET /api/v1/agencies/:id/workflows (already exists)
  - [ ] POST /api/v1/agencies/:id/workflows (already exists)
  - [ ] PUT /api/v1/workflows/:id (already exists)
  - [ ] GET /api/v1/agencies/:id/workbench (NEW - from tag)

### Acceptance Criteria

- [ ] Workflow steps can reference WorkItems by code
- [ ] Agency Designer shows WorkItem-Workflow integration UI
- [ ] Workflow validation prevents invalid WorkItem references
- [ ] Workbench boards are generated from published tag workflows
- [ ] Each Workbench column maps to a WorkItem from specification
- [ ] Workflow CRUD operations maintain WorkItem references
- [ ] Tag snapshots preserve WorkItem-Workflow relationships

---

## MVP-031: Issue Lifecycle & SLA

### Problem

Issues (tickets) created in Workbench need:
- State management (as they move through columns)
- SLA tracking (time limits per WorkItem)
- Escalation policies (when SLAs are breached)
- History/audit trail

### Lifecycle States

```
Backlog → Ready → In-Progress → Review → Done
                       ↓
                    Blocked (waiting on dependency)
                       ↓
                  In-Progress (dependency resolved)

Special transitions:
- Review → In-Progress (changes requested)
- Any state → Cancelled (work abandoned)
```

### Issue Model

```go
type Issue struct {
    Key             string    `json:"_key,omitempty"`
    ID              string    `json:"_id,omitempty"`
    AgencyID        string    `json:"agency_id"`
    Number          int       `json:"number"`          // Auto-increment per agency
    Title           string    `json:"title"`
    Description     string    `json:"description"`
    
    // Workflow/Column tracking
    WorkflowID      string    `json:"workflow_id"`     // Which Kanban board
    CurrentStep     string    `json:"current_step"`    // Current column (step ID)
    WorkItemCode    string    `json:"work_item_code"`  // Current WorkItem being worked on
    
    // State and lifecycle
    State           string    `json:"state"`           // Backlog, Ready, In-Progress, Review, Done, Blocked, Cancelled
    Priority        string    `json:"priority"`        // Low, Medium, High, Critical
    
    // Assignment
    AssignedTo      string    `json:"assigned_to,omitempty"`      // Agent ID
    AssignedAt      *time.Time `json:"assigned_at,omitempty"`
    
    // Git integration
    BranchName      string    `json:"branch_name,omitempty"`
    PullRequestID   string    `json:"pull_request_id,omitempty"`
    
    // SLA tracking
    CreatedAt       time.Time  `json:"created_at"`
    StartedAt       *time.Time `json:"started_at,omitempty"`     // When moved to In-Progress
    CompletedAt     *time.Time `json:"completed_at,omitempty"`   // When moved to Done
    DueDate         *time.Time `json:"due_date,omitempty"`
    SLABreach       bool       `json:"sla_breach"`
    
    // Metadata
    Tags            []string   `json:"tags,omitempty"`
    UpdatedAt       time.Time  `json:"updated_at"`
}
```

### SLA Configuration (from WorkItem)

WorkItems can be extended with SLA metadata:
```go
type WorkItem struct {
    // ... existing fields
    
    // Optional SLA configuration
    SLA *WorkItemSLA `json:"sla,omitempty"`
}

type WorkItemSLA struct {
    ResponseTime   string `json:"response_time"`   // e.g., "1h" - time to start
    CompletionTime string `json:"completion_time"` // e.g., "24h" - time to complete
    WarningThreshold float64 `json:"warning_threshold"` // e.g., 0.8 (80% of time elapsed)
}
```

### Implementation Tasks

- [ ] Create Issue model and ArangoDB collection
- [ ] Implement state machine for issue transitions
- [ ] Build SLA tracking service
  - [ ] Calculate time in each state
  - [ ] Detect SLA breaches
  - [ ] Send notifications on warning thresholds
- [ ] Issue movement API
  - [ ] POST /api/v1/issues (create issue)
  - [ ] PUT /api/v1/issues/:id/move (move to different column/step)
  - [ ] PUT /api/v1/issues/:id/assign (assign to agent)
  - [ ] GET /api/v1/issues/:id/history (state change history)

### Acceptance Criteria

- [ ] Issues can be created in Kanban boards
- [ ] Issues move through workflow steps (columns)
- [ ] State transitions are validated
- [ ] SLA timers track time in each state
- [ ] SLA breaches are detected and flagged
- [ ] Issue history is complete and queryable

---

## MVP-032: Agent Assignment & Routing

### Problem

When an issue is created or reaches a specific column:
- Which agent should work on it?
- What skills/capabilities are needed?
- Should a new agent instance be created?
- How to handle WIP (Work In Progress) limits?

### Agent Assignment Model

```go
type AgentAssignment struct {
    IssueID        string    `json:"issue_id"`
    AgentID        string    `json:"agent_id"`
    WorkItemCode   string    `json:"work_item_code"`   // What type of work
    AssignedAt     time.Time `json:"assigned_at"`
    Status         string    `json:"status"`           // assigned, working, completed
    Priority       int       `json:"priority"`
}
```

### Assignment Strategy

**Option 1: Manual Assignment**
- User selects agent from available pool
- Agent accepts or rejects assignment

**Option 2: Auto-Assignment (Based on WorkItem)**
- WorkItem specifies required capabilities/tools
- System finds/creates agent with those capabilities
- Agent is automatically assigned

**Option 3: Claim-Based (Pull Model)**
- Issues sit in "Ready" column
- Agents claim work when available
- First-come-first-served or priority-based

### Implementation Tasks

- [ ] Agent assignment service
- [ ] Agent capability matching
- [ ] Agent instance creation from WorkItem metadata
- [ ] WIP limit enforcement
- [ ] Assignment notification system

### Acceptance Criteria

- [ ] Issues can be assigned to agents (manual or auto)
- [ ] Agent capabilities match WorkItem requirements
- [ ] New agent instances can be created for issues
- [ ] WIP limits are enforced per agent/column
- [ ] Assignment history is tracked

---

## Next Steps

1. **Update Workflow model** to include `WorkItemCode` in steps
2. **Build Workflow Designer UI** with WorkItem integration
3. **Implement Kanban generation** from published tags
4. **Create Issue model** and lifecycle management
5. **Build Agent assignment** routing logic
