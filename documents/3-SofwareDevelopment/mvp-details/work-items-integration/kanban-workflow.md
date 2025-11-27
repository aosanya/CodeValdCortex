# Kanban Workflow & Issue Lifecycle

**Status**: 📋 Planned - Architecture Defined  
**Related**: [git-based-document-system.md](./git-based-document-system.md), [work-item-schema.md](./work-item-schema.md)

## Overview

This document provides an overview of the Kanban workflow system in CodeValdCortex. The system uses a workflow-based approach where issues progress through stages defined in agency specifications, with work performed by humans or AI agents, and all changes tracked in the internal Git-in-ArangoDB system.

**Key Concepts**:
- **Work Item Definitions**: Templates defined in agency specs (REQ1, IMPL1, etc.)
- **Issues**: Runtime instances created by users on Kanban board
- **Workflows**: Sequential steps that issues progress through
- **Agents/Humans**: Workers who claim and execute work on issues
- **Git Integration**: All work results in commits/PRs in internal Git system
- **Automation**: PR merge events trigger automatic issue progression

---

## Topic Documentation

This overview has been split into focused topic areas:

### 1. Issue Management & Lifecycle
**File**: [issue-management.md](issue-management.md)

**Covers**:
- Issue creation process
- Assignment models (manual vs claim-based)
- Git integration (branch creation, commits)
- Pull request creation workflow
- Human review process
- Issue data models and API endpoints
- UI components (Kanban board, issue cards)

**Read this for**: Understanding how issues are created, assigned, worked on, and reviewed.

---

### 2. Workflow Automation & Orchestration
**File**: [workflow-automation.md](workflow-automation.md)

**Covers**:
- Automated issue progression on PR merge
- Workflow configuration from agency specs
- Step progression logic and algorithms
- Entry point enforcement (REQ1)
- Event-driven orchestration
- Workflow data models
- Edge cases and error handling
- Performance considerations (caching, batching)

**Read this for**: Understanding how the system automatically moves issues through workflow stages.

---

## Quick Reference

### Complete Workflow Example

```
1. User creates issue "Implement JWT Authentication"
   → Issue placed in REQ1 column

2. Agent claims issue
   → Issue assigned to agent-req-001

3. Agent creates branch and commits work
   → Branch: issue-123-jwt-auth
   → Commits requirements document

4. Agent creates Pull Request
   → PR #456 linked to issue #123
   → Issue status: ready_for_review

5. Human reviews and merges PR
   → PR merged to main
   → Trigger: pull_request.merged event

6. Workflow Orchestrator runs
   → Detects issue #123 in REQ1
   → Determines next step: ARCH1
   → Moves issue to ARCH1 column
   → Resets issue for next assignment

7. Process repeats for ARCH1, IMPL1, TEST1...
   → Until workflow complete
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
- Branch name: `issue-{id}-{slug}`
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

## Data Models Summary

### Work Issue (Runtime)

```go
type WorkIssue struct {
    Key              string    `json:"_key"`
    Title            string    `json:"title"`
    Description      string    `json:"description"`
    WorkflowID       string    `json:"workflow_id"`
    CurrentStep      string    `json:"current_step"`      // REQ1, REV1, IMPL1, etc.
    Status           string    `json:"status"`            // open, assigned, in_progress, ready_for_review
    AssignedTo       string    `json:"assigned_to"`
    BranchName       string    `json:"branch_name"`
    PullRequestID    string    `json:"pull_request_id"`
    CompletedSteps   []string  `json:"completed_steps"`
    CreatedBy        string    `json:"created_by"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

### Workflow Definition

```go
type Workflow struct {
    Key   string         `json:"_key"`
    Name  string         `json:"name"`
    Steps []WorkflowStep `json:"steps"`
}

type WorkflowStep struct {
    Order int                 `json:"order"`
    Items []WorkflowStepItem  `json:"items"`
}

type WorkflowStepItem struct {
    WorkItemID string `json:"work_item_id"`  // REQ1, REV1, etc.
}
```

---

## API Endpoints Summary

### Issue Management
```
POST   /api/v1/issues                    # Create issue
GET    /api/v1/issues/:id                # Get issue details
PATCH  /api/v1/issues/:id                # Update issue
POST   /api/v1/issues/:id/assign         # Assign worker
POST   /api/v1/issues/:id/claim          # Claim work
GET    /api/v1/issues/available          # Get claimable issues
```

### Workflow Management
```
GET    /api/v1/workflows                 # List workflows
GET    /api/v1/workflows/:id             # Get workflow details
GET    /api/v1/work-items                # List work item definitions
GET    /api/v1/work-items/:code          # Get work item by code
```

---

## Integration Points

### 1. Workflow Orchestrator
- Listens to PR merge events
- Determines next workflow step
- Moves issues between columns
- Triggers notifications

**Details**: See [workflow-automation.md](workflow-automation.md)

### 2. Agent Factory
- Creates agent instances from role definitions
- Assigns agents to issues
- Monitors agent work completion

### 3. Git System
- Manages branches, commits, merges
- Stores all objects in ArangoDB
- Provides file explorer UI

**Details**: See [git-operations.md](git-operations.md) and [file-explorer.md](file-explorer.md)

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

- **Issue Management**: [issue-management.md](issue-management.md) - Detailed issue lifecycle, assignment, Git integration
- **Workflow Automation**: [workflow-automation.md](workflow-automation.md) - Orchestration logic, progression algorithms
- **Git Implementation**: [git-based-document-system.md](./git-based-document-system.md) - Git architecture
- **Git Operations**: [git-operations.md](git-operations.md) - Low-level Git operations
- **Work Item Schema**: [work-item-schema.md](./work-item-schema.md) - Work item definitions
- **Pull Requests**: [pull-requests.md](./pull-requests.md) - Review workflow
- **Implementation Roadmap**: [implementation-guide.md](./implementation-guide.md) - Development phases

---

**Last Updated**: 2025-11-27  
**Status**: Architecture Defined - Ready for Implementation  
**Next Step**: Phase 4 Implementation - Kanban Workflow Integration
