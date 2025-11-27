# Workflow Automation & Orchestration

**Related Tasks**: MVP-WI-009 (Workflow Automation), MVP-030 (Work Item Definitions)  
**Status**: 📋 Planned - Architecture Defined

## Overview

This document describes the workflow orchestration system that automatically progresses issues through workflow stages based on PR merge events. The system interprets agency specification workflows and manages issue state transitions.

**Key Responsibilities**:
- Detect PR merge events
- Determine next workflow step from agency spec
- Move issues between workflow stages (columns)
- Trigger notifications and agent assignments
- Enforce workflow rules and entry points

---

## Automated Issue Progression

### Trigger: PR Merge Event

When a pull request is merged to `main`:

```
Event: pull_request.merged
Data: {
  pr_id: "pr-456",
  issue_id: "issue-123",
  merged_at: "2025-11-26T11:00:00Z",
  merged_by: "user-alice"
}
```

### Workflow Orchestrator Logic

```javascript
// Pseudo-code for orchestrator
async function handlePRMerge(event) {
  // 1. Get issue
  const issue = await getIssue(event.issue_id);
  
  // 2. Get workflow configuration
  const workflow = await getWorkflow(issue.workflow_id);
  
  // 3. Find current step
  const currentStepIndex = workflow.steps.findIndex(step => 
    step.items.some(item => item.work_item_id === issue.current_step)
  );
  
  // 4. Determine next step
  const nextStepGroup = workflow.steps[currentStepIndex + 1];
  
  if (!nextStepGroup) {
    // Workflow complete
    await completeIssue(issue);
    return;
  }
  
  // 5. Get first work item in next step
  const nextWorkItem = nextStepGroup.items[0].work_item_id;
  
  // 6. Update issue
  await updateIssue(issue._key, {
    current_step: nextWorkItem,
    status: "open",
    assigned_to: null,
    branch_name: null,
    completed_steps: [...issue.completed_steps, issue.current_step],
    updated_at: new Date()
  });
  
  // 7. Trigger notifications
  await notifyWorkAvailable(nextWorkItem, issue);
}
```

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
    },
    {
      "_key": "5379395d-fcfe-4e83-a035-dc2a0fef4e14",
      "code": "REV1",
      "title": "Review and validate technical specification",
      "description": "Systematic review of technical documentation...",
      "deliverables": [
        "Review report",
        "Validation checklist",
        "Recommendations"
      ]
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
            {"work_item_id": "ARCH1"},
            {"work_item_id": "IMPL1"}
          ]
        },
        {
          "order": 2,
          "items": [
            {"work_item_id": "TEST1"}
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

---

## Work Item Definition → Issue Mapping

### Design Time vs Runtime

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

## Step Progression Logic

### Example Workflow Progression

**Workflow**: Software Development  
**Issue**: #123 "Implement JWT Authentication"

```
Step 0 (REQ1, REV1)
  │
  ├─► Issue created in REQ1
  │   - Status: open
  │   - Current step: REQ1
  │
  ├─► Agent claims issue
  │   - Status: assigned
  │   - Assigned to: agent-req-001
  │
  ├─► Agent creates branch and commits
  │   - Status: in_progress
  │   - Branch: issue-123-jwt-auth
  │
  ├─► Agent creates PR
  │   - Status: ready_for_review
  │   - PR: pr-456
  │
  ├─► Human reviews and merges PR
  │   - PR merged to main
  │   - Event: pull_request.merged
  │
  └─► ORCHESTRATOR TRIGGERED
      │
      ├─► Find current step: REQ1 (in step 0)
      ├─► Get next step: Step 1
      ├─► Get first work item in step 1: ARCH1
      ├─► Update issue:
      │     - Current step: ARCH1
      │     - Status: open
      │     - Assigned to: null
      │     - Completed steps: ["REQ1"]
      │
      └─► Issue moved to ARCH1 column on Kanban board

Step 1 (ARCH1, IMPL1)
  │
  └─► Process repeats for ARCH1...
```

---

## Orchestrator Implementation

### Event Listener

```go
package orchestration

type WorkflowOrchestrator struct {
    issueRepo   repository.IssueRepository
    workflowRepo repository.WorkflowRepository
    eventBus    events.EventBus
    logger      *logrus.Logger
}

func (o *WorkflowOrchestrator) Start() {
    // Subscribe to PR merge events
    o.eventBus.Subscribe("pull_request.merged", o.handlePRMerge)
}

func (o *WorkflowOrchestrator) handlePRMerge(event events.Event) error {
    data := event.Data.(PRMergedEvent)
    
    // Get issue
    issue, err := o.issueRepo.GetByID(data.IssueID)
    if err != nil {
        return fmt.Errorf("failed to get issue: %w", err)
    }
    
    // Get workflow
    workflow, err := o.workflowRepo.GetByID(issue.WorkflowID)
    if err != nil {
        return fmt.Errorf("failed to get workflow: %w", err)
    }
    
    // Progress issue
    return o.progressIssue(issue, workflow)
}
```

---

### Progression Algorithm

```go
func (o *WorkflowOrchestrator) progressIssue(issue *models.WorkIssue, workflow *models.Workflow) error {
    // 1. Find current step index
    currentStepIndex := -1
    for i, step := range workflow.Steps {
        for _, item := range step.Items {
            if item.WorkItemID == issue.CurrentStep {
                currentStepIndex = i
                break
            }
        }
        if currentStepIndex >= 0 {
            break
        }
    }
    
    if currentStepIndex == -1 {
        return fmt.Errorf("current step %s not found in workflow", issue.CurrentStep)
    }
    
    // 2. Check if workflow complete
    if currentStepIndex >= len(workflow.Steps)-1 {
        return o.completeIssue(issue)
    }
    
    // 3. Get next step
    nextStep := workflow.Steps[currentStepIndex+1]
    nextWorkItem := nextStep.Items[0].WorkItemID
    
    // 4. Update issue
    updates := map[string]interface{}{
        "current_step":    nextWorkItem,
        "status":          "open",
        "assigned_to":     nil,
        "branch_name":     nil,
        "completed_steps": append(issue.CompletedSteps, issue.CurrentStep),
        "updated_at":      time.Now(),
    }
    
    if err := o.issueRepo.Update(issue.Key, updates); err != nil {
        return fmt.Errorf("failed to update issue: %w", err)
    }
    
    // 5. Trigger notifications
    o.notifyWorkAvailable(nextWorkItem, issue)
    
    o.logger.WithFields(logrus.Fields{
        "issue_id":   issue.Key,
        "from_step":  issue.CurrentStep,
        "to_step":    nextWorkItem,
    }).Info("Issue progressed to next workflow step")
    
    return nil
}
```

---

### Completion Handler

```go
func (o *WorkflowOrchestrator) completeIssue(issue *models.WorkIssue) error {
    updates := map[string]interface{}{
        "status":          "completed",
        "completed_steps": append(issue.CompletedSteps, issue.CurrentStep),
        "completed_at":    time.Now(),
    }
    
    if err := o.issueRepo.Update(issue.Key, updates); err != nil {
        return fmt.Errorf("failed to complete issue: %w", err)
    }
    
    o.logger.WithField("issue_id", issue.Key).Info("Issue completed")
    
    // Trigger completion events
    o.eventBus.Publish("issue.completed", IssueCompletedEvent{
        IssueID:      issue.Key,
        CompletedAt:  time.Now(),
    })
    
    return nil
}
```

---

## Entry Point Enforcement

### REQ1 as Single Entry Point

**Rule**: All issues MUST start in REQ1 (requirements gathering)

**Enforcement**:

```go
func (s *IssueService) CreateIssue(req CreateIssueRequest) (*models.WorkIssue, error) {
    // Get workflow
    workflow, err := s.workflowRepo.GetByID(req.WorkflowID)
    if err != nil {
        return nil, err
    }
    
    // Find entry point (first step, first work item)
    entryStep := workflow.Steps[0]
    entryWorkItem := entryStep.Items[0].WorkItemID
    
    issue := &models.WorkIssue{
        Title:       req.Title,
        Description: req.Description,
        WorkflowID:  req.WorkflowID,
        CurrentStep: entryWorkItem,  // Always REQ1
        Status:      "open",
        CreatedBy:   req.CreatedBy,
        CreatedAt:   time.Now(),
    }
    
    return s.issueRepo.Create(issue)
}
```

**Why REQ1 Entry Point?**
- ✅ Ensures all work has proper requirements analysis
- ✅ Prevents skipping critical planning phases
- ✅ Maintains consistent quality standards
- ✅ Provides complete audit trail

---

## Workflow Data Models

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
    WorkItemID string `json:"work_item_id"`  // References work item code (REQ1, REV1, etc.)
}
```

### Work Item Definition

```go
type WorkItemDefinition struct {
    Key          string   `json:"_key"`
    Code         string   `json:"code"`           // REQ1, REV1, IMPL1, etc.
    Title        string   `json:"title"`
    Description  string   `json:"description"`
    Deliverables []string `json:"deliverables"`
    GoalKeys     []string `json:"goal_keys"`
    Tags         []string `json:"tags"`
}
```

---

