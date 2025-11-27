# Workflow Integration & APIs

**Related Tasks**: MVP-WI-012 (Event-Driven Progression), MVP-031 (Goal Dashboard), MVP-032 (Communication Dashboard)  
**Status**: �� Planned

## Overview

This document describes API endpoints, event handling, and integration points for the workflow automation system. It covers REST APIs, event bus integration, notifications, and error handling patterns.

**For workflow orchestration logic**, see [Workflow Automation](workflow-automation.md).

---

## API Endpoints

### Workflow Management

```
GET    /api/v1/workflows
Response:
{
  "workflows": [
    {
      "_key": "workflow-001",
      "name": "Software Development Workflow",
      "steps": [
        {
          "order": 0,
          "items": [
            {"work_item_id": "REQ1"}
          ]
        }
      ]
    }
  ]
}

GET    /api/v1/workflows/:id
Response: (same structure as above)

GET    /api/v1/work-items
Response:
{
  "work_items": [
    {
      "_key": "uuid",
      "code": "REQ1",
      "title": "Conduct stakeholder requirements gathering",
      "deliverables": ["Requirements document", "Stakeholder sign-off"]
    }
  ]
}

GET    /api/v1/work-items/:code
Response: (same structure as above)
```

---

## Integration Points

### 1. Event Bus
- **Purpose**: Decouple PR merge from workflow progression
- **Events**:
  - `pull_request.merged` → Triggers orchestrator
  - `issue.progressed` → Notifies UI/agents
  - `issue.completed` → Archive/metrics

```go
type EventBus interface {
    Subscribe(eventType string, handler func(Event) error)
    Publish(eventType string, data interface{})
}
```

**Event Flow**:
```
PR Merge
    ↓
Event: pull_request.merged
    ↓
WorkflowOrchestrator.handlePRMerge()
    ↓
progressIssue()
    ↓
Event: issue.progressed
    ↓
[UI Update, Agent Notification, Metrics]
```

### 2. Notification System
- Alerts agents when new work available in their domain
- Notifies users of issue progression
- Sends completion notifications

```go
func (o *WorkflowOrchestrator) notifyWorkAvailable(workItemID string, issue *models.WorkIssue) {
    o.eventBus.Publish("work.available", WorkAvailableEvent{
        WorkItemID: workItemID,
        IssueID:    issue.Key,
        IssueTitle: issue.Title,
    })
}

type WorkAvailableEvent struct {
    WorkItemID string `json:"work_item_id"`  // e.g., "ARCH1"
    IssueID    string `json:"issue_id"`
    IssueTitle string `json:"issue_title"`
    Priority   int    `json:"priority"`
    AssignedTo string `json:"assigned_to,omitempty"`
}
```

**Notification Channels**:
- **In-app**: Browser notifications via WebSocket
- **Email**: Critical assignments (configurable)
- **Slack/Teams**: External integrations (future)

### 3. Agent Factory
- Listens to `work.available` events
- Matches agents to work items based on roles
- Auto-assigns or queues for claim-based assignment

```go
type AgentFactory struct {
    eventBus    EventBus
    agentPool   *pool.AgentPool
    issueRepo   repository.IssueRepository
}

func (f *AgentFactory) Start() {
    f.eventBus.Subscribe("work.available", f.handleWorkAvailable)
}

func (f *AgentFactory) handleWorkAvailable(event Event) error {
    workEvent := event.Data.(WorkAvailableEvent)
    
    // Find available agent for work item type
    agent, err := f.agentPool.ClaimAgent(workEvent.WorkItemID)
    if err != nil {
        return fmt.Errorf("no available agent for %s: %w", workEvent.WorkItemID, err)
    }
    
    // Assign issue to agent
    return f.issueRepo.Update(workEvent.IssueID, map[string]interface{}{
        "assigned_to": agent.ID,
        "status":      "in_progress",
    })
}
```

---

## Edge Cases & Error Handling

### 1. PR Merge Without Issue Link
**Problem**: PR merged but no `issue_id` in PR data

**Solution**:
```go
if event.IssueID == "" {
    o.logger.Warn("PR merged without issue link", "pr_id", event.PRID)
    return nil // Skip progression
}
```

**Prevention**: Enforce issue links in PR creation UI.

### 2. Workflow Step Not Found
**Problem**: Issue's `current_step` doesn't exist in workflow

**Solution**:
```go
if currentStepIndex == -1 {
    o.logger.Error("Current step not found in workflow",
        "issue_id", issue.Key,
        "current_step", issue.CurrentStep,
        "workflow_id", workflow.Key)
    return fmt.Errorf("workflow misconfiguration")
}
```

**Recovery**: Manual intervention required - check workflow definition vs issue state.

### 3. Concurrent PR Merges
**Problem**: Two PRs for same issue merged simultaneously

**Solution**: Use ArangoDB document versioning (optimistic locking)
```go
// Update with optimistic locking
updates := map[string]interface{}{
    "current_step": nextWorkItem,
    "_rev": issue.Rev,  // Ensures no concurrent updates
}
err := o.issueRepo.Update(issue.Key, updates)
if err == ErrConflict {
    // Another update happened, retry
    return o.progressIssue(issue, workflow)
}
```

**Result**: First merge wins, second merge retries with updated state.

### 4. Orphaned Issues
**Problem**: Workflow deleted but issues still reference it

**Solution**: Cascade delete prevention
```go
func (r *WorkflowRepository) Delete(workflowID string) error {
    // Check for active issues
    count, err := r.db.Collection("work_issues").CountDocuments(ctx, 
        map[string]interface{}{"workflow_id": workflowID, "status": bson.M{"$ne": "completed"}})
    
    if count > 0 {
        return fmt.Errorf("cannot delete workflow with %d active issues", count)
    }
    
    return r.db.Collection("workflows").DeleteDocument(ctx, workflowID)
}
```

---

## Performance Considerations

### 1. Workflow Caching
- Cache workflow definitions (rarely change)
- Invalidate on spec updates

```go
type CachedWorkflowRepo struct {
    repo  repository.WorkflowRepository
    cache *redis.Client
    ttl   time.Duration
}

func (r *CachedWorkflowRepo) GetByID(id string) (*models.Workflow, error) {
    // Check cache
    cached, err := r.cache.Get(ctx, fmt.Sprintf("workflow:%s", id)).Result()
    if err == nil {
        var workflow models.Workflow
        json.Unmarshal([]byte(cached), &workflow)
        return &workflow, nil
    }
    
    // Fetch from DB
    workflow, err := r.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    data, _ := json.Marshal(workflow)
    r.cache.Set(ctx, fmt.Sprintf("workflow:%s", id), data, r.ttl)
    
    return workflow, nil
}
```

**Cache Invalidation**:
```go
func (r *CachedWorkflowRepo) Update(id string, updates map[string]interface{}) error {
    err := r.repo.Update(id, updates)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    r.cache.Del(ctx, fmt.Sprintf("workflow:%s", id))
    return nil
}
```

### 2. Batch Event Processing
- Process multiple PR merges in batch
- Reduce DB round trips

```go
type BatchEventProcessor struct {
    orchestrator *WorkflowOrchestrator
    batchSize    int
    interval     time.Duration
    queue        chan PRMergeEvent
}

func (p *BatchEventProcessor) Start() {
    ticker := time.NewTicker(p.interval)
    batch := []PRMergeEvent{}
    
    for {
        select {
        case event := <-p.queue:
            batch = append(batch, event)
            
            if len(batch) >= p.batchSize {
                p.processBatch(batch)
                batch = []PRMergeEvent{}
            }
            
        case <-ticker.C:
            if len(batch) > 0 {
                p.processBatch(batch)
                batch = []PRMergeEvent{}
            }
        }
    }
}

func (p *BatchEventProcessor) processBatch(events []PRMergeEvent) {
    for _, event := range events {
        go p.orchestrator.handlePRMerge(event) // Parallel processing
    }
}
```

---

## Testing Strategy

### Unit Tests

```go
func TestProgressIssue(t *testing.T) {
    orchestrator := setupOrchestrator()
    
    issue := &models.WorkIssue{
        Key:         "issue-123",
        CurrentStep: "REQ1",
        WorkflowID:  "workflow-001",
    }
    
    workflow := &models.Workflow{
        Steps: []WorkflowStep{
            {Order: 0, Items: []WorkflowStepItem{{WorkItemID: "REQ1"}}},
            {Order: 1, Items: []WorkflowStepItem{{WorkItemID: "ARCH1"}}},
        },
    }
    
    err := orchestrator.progressIssue(issue, workflow)
    assert.NoError(t, err)
    
    // Verify issue updated
    updated, _ := orchestrator.issueRepo.GetByID("issue-123")
    assert.Equal(t, "ARCH1", updated.CurrentStep)
    assert.Equal(t, "open", updated.Status)
}
```

### Integration Tests

```go
func TestEndToEndProgression(t *testing.T) {
    // 1. Create issue
    issue := createTestIssue("REQ1", "workflow-001")
    
    // 2. Create PR
    pr := createTestPR(issue.Key)
    
    // 3. Merge PR
    mergePR(pr.Key)
    
    // 4. Wait for event processing
    time.Sleep(100 * time.Millisecond)
    
    // 5. Verify issue progressed
    updated := getIssue(issue.Key)
    assert.Equal(t, "ARCH1", updated.CurrentStep)
}
```

---

## Related Documentation

- **[Workflow Automation](workflow-automation.md)** - Orchestration logic and progression algorithms
- [Issue Management](issue-management.md) - Issue lifecycle, assignment
- [Git Operations](git-operations.md) - Branch/commit management
- [Pull Requests](pull-requests.md) - Review workflow, merge events
- [Work Item Schema](work-item-schema.md) - Work item definitions
- [Implementation Guide](implementation-guide.md) - Development roadmap

---

**Last Updated**: 2025-11-27  
**Status**: Architecture Defined - Ready for Implementation
