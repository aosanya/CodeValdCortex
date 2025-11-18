# Observability & Analytics

This document covers metrics, monitoring, tracing, and analytics for the work item system.

## Overview

Comprehensive observability through:
- **Metrics**: Performance and usage statistics
- **Traces**: Detailed execution flows
- **Logs**: Structured logging
- **Analytics**: Graph-based insights
- **Dashboards**: Real-time visualization

## Key Metrics

### Work Item Metrics

```go
type WorkItemMetrics struct {
    // Volume
    TotalCreated      int64         `json:"total_created"`
    TotalCompleted    int64         `json:"total_completed"`
    TotalFailed       int64         `json:"total_failed"`
    InProgress        int64         `json:"in_progress"`
    
    // By type
    ByType            map[string]TypeMetrics `json:"by_type"`
    
    // Performance
    AvgExecutionTime  time.Duration `json:"avg_execution_time"`
    P50ExecutionTime  time.Duration `json:"p50_execution_time"`
    P95ExecutionTime  time.Duration `json:"p95_execution_time"`
    P99ExecutionTime  time.Duration `json:"p99_execution_time"`
    
    // Success rate
    SuccessRate       float64       `json:"success_rate"`
    
    // Automation
    AutoMergeRate     float64       `json:"auto_merge_rate"`
    
    // Time range
    TimeRange         TimeRange     `json:"time_range"`
}

type TypeMetrics struct {
    Count             int64         `json:"count"`
    AvgDuration       time.Duration `json:"avg_duration"`
    SuccessRate       float64       `json:"success_rate"`
}
```

### Query Work Item Metrics

```go
func (r *Repo) GetWorkItemMetrics(timeRange TimeRange) (*WorkItemMetrics, error) {
    query := `
        LET items = (
            FOR wi IN work_items
                FILTER wi.created_at >= @start
                FILTER wi.created_at <= @end
                RETURN wi
        )
        
        LET byType = (
            FOR wi IN items
                COLLECT type = wi.work_type
                AGGREGATE 
                    count = COUNT(),
                    avgDuration = AVG(DATE_DIFF(wi.started_at, wi.completed_at, "s")),
                    completed = SUM(wi.status == "completed" ? 1 : 0)
                RETURN {
                    type: type,
                    count: count,
                    avg_duration: avgDuration,
                    success_rate: completed / count
                }
        )
        
        LET durations = items[* FILTER CURRENT.completed_at != null
            RETURN DATE_DIFF(CURRENT.started_at, CURRENT.completed_at, "s")]
        
        RETURN {
            total_created: LENGTH(items),
            total_completed: LENGTH(items[* FILTER CURRENT.status == "completed"]),
            total_failed: LENGTH(items[* FILTER CURRENT.status == "failed"]),
            in_progress: LENGTH(items[* FILTER CURRENT.status == "executing"]),
            by_type: byType,
            avg_execution_time: AVG(durations),
            p50_execution_time: PERCENTILE(durations, 50),
            p95_execution_time: PERCENTILE(durations, 95),
            p99_execution_time: PERCENTILE(durations, 99),
            success_rate: LENGTH(items[* FILTER CURRENT.status == "completed"]) / LENGTH(items)
        }
    `
    
    cursor, err := r.db.Query(context.Background(), query, map[string]interface{}{
        "start": timeRange.Start,
        "end":   timeRange.End,
    })
    
    var metrics WorkItemMetrics
    cursor.ReadDocument(context.Background(), &metrics)
    return &metrics, err
}
```

### LLM Usage Metrics

```go
type LLMMetrics struct {
    // Volume
    TotalRequests     int64         `json:"total_requests"`
    TotalTokens       int64         `json:"total_tokens"`
    PromptTokens      int64         `json:"prompt_tokens"`
    CompletionTokens  int64         `json:"completion_tokens"`
    
    // Cost
    TotalCost         float64       `json:"total_cost"`
    
    // Performance
    AvgResponseTime   time.Duration `json:"avg_response_time"`
    P95ResponseTime   time.Duration `json:"p95_response_time"`
    
    // By model
    ByModel           map[string]ModelMetrics `json:"by_model"`
    
    // By agent
    ByAgent           map[string]AgentMetrics `json:"by_agent"`
    
    // Error rate
    ErrorRate         float64       `json:"error_rate"`
}

type ModelMetrics struct {
    Requests          int64         `json:"requests"`
    Tokens            int64         `json:"tokens"`
    Cost              float64       `json:"cost"`
    AvgResponseTime   time.Duration `json:"avg_response_time"`
}

type AgentMetrics struct {
    Requests          int64         `json:"requests"`
    Tokens            int64         `json:"tokens"`
    Cost              float64       `json:"cost"`
    BudgetRemaining   float64       `json:"budget_remaining"`
}
```

### Query LLM Metrics

```go
query := `
    FOR usage IN llm_usage
        FILTER usage.timestamp >= @start
        FILTER usage.timestamp <= @end
        
        COLLECT AGGREGATE
            total_requests = COUNT(),
            total_tokens = SUM(usage.total_tokens),
            prompt_tokens = SUM(usage.prompt_tokens),
            completion_tokens = SUM(usage.completion_tokens),
            total_cost = SUM(usage.cost),
            avg_duration = AVG(usage.duration_ms),
            p95_duration = PERCENTILE(usage.duration_ms, 95)
        
        RETURN {
            total_requests,
            total_tokens,
            prompt_tokens,
            completion_tokens,
            total_cost,
            avg_response_time: avg_duration,
            p95_response_time: p95_duration
        }
`
```

### Git Operation Metrics

```go
type GitMetrics struct {
    // Volume
    TotalCommits      int64         `json:"total_commits"`
    TotalBranches     int64         `json:"total_branches"`
    TotalMerges       int64         `json:"total_merges"`
    
    // By agent
    CommitsByAgent    map[string]int64 `json:"commits_by_agent"`
    
    // Repository size
    TotalBlobs        int64         `json:"total_blobs"`
    TotalSize         int64         `json:"total_size_bytes"`
    
    // Code stats
    LinesByLanguage   map[string]int64 `json:"lines_by_language"`
}
```

## Execution Traces

### Trace Structure

```go
type WorkflowTrace struct {
    TraceID           string              `json:"trace_id"`
    WorkItemID        string              `json:"work_item_id"`
    Status            string              `json:"status"`
    
    // Timeline
    Timeline          []TraceEvent        `json:"timeline"`
    
    // Detailed steps
    Steps             []ExecutionStep     `json:"steps"`
    
    // LLM interactions
    LLMCalls          []LLMCall          `json:"llm_calls"`
    
    // Git operations
    GitOps            []GitOperation     `json:"git_operations"`
    
    // Errors
    Errors            []ErrorEvent       `json:"errors"`
    
    // Metrics
    Metrics           ExecutionMetrics   `json:"metrics"`
    
    // Timestamps
    StartedAt         time.Time          `json:"started_at"`
    CompletedAt       *time.Time         `json:"completed_at"`
}

type TraceEvent struct {
    Timestamp         time.Time          `json:"timestamp"`
    Event             string             `json:"event"`
    Details           map[string]interface{} `json:"details"`
    DurationMs        int64              `json:"duration_ms"`
}

type ErrorEvent struct {
    Timestamp         time.Time          `json:"timestamp"`
    Step              string             `json:"step"`
    Error             string             `json:"error"`
    Recoverable       bool               `json:"recoverable"`
}
```

### Trace Collection

```go
type TraceCollector struct {
    db       *arangodb.Database
    traceID  string
    events   []TraceEvent
    mu       sync.Mutex
}

func (c *TraceCollector) RecordEvent(event string, details map[string]interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.events = append(c.events, TraceEvent{
        Timestamp: time.Now(),
        Event:     event,
        Details:   details,
    })
}

func (c *TraceCollector) Flush() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    trace := WorkflowTrace{
        TraceID:   c.traceID,
        Timeline:  c.events,
        StartedAt: c.events[0].Timestamp,
    }
    
    if len(c.events) > 0 {
        lastEvent := c.events[len(c.events)-1]
        trace.CompletedAt = &lastEvent.Timestamp
    }
    
    _, err := c.db.Collection("workflow_traces").CreateDocument(context.Background(), trace)
    return err
}

// Usage in work item execution
func (e *Executor) Execute(ctx context.Context, issue *gitea.Issue) error {
    trace := NewTraceCollector(e.db, generateTraceID())
    defer trace.Flush()
    
    trace.RecordEvent("execution_started", map[string]interface{}{
        "issue_id": issue.Index,
        "title":    issue.Title,
    })
    
    // Classify work type
    trace.RecordEvent("classify_started", nil)
    workType := e.classifyWorkType(issue)
    trace.RecordEvent("classify_completed", map[string]interface{}{
        "work_type": workType,
    })
    
    // Generate content
    trace.RecordEvent("llm_generation_started", nil)
    content, err := e.llm.Generate(ctx, issue)
    if err != nil {
        trace.RecordEvent("llm_generation_failed", map[string]interface{}{
            "error": err.Error(),
        })
        return err
    }
    trace.RecordEvent("llm_generation_completed", map[string]interface{}{
        "content_length": len(content),
    })
    
    // ... rest of execution
    
    trace.RecordEvent("execution_completed", nil)
    return nil
}
```

## Graph Analytics

### File Importance (PageRank-style)

```go
query := `
    FOR file IN git_objects
        FILTER file.type == "blob"
        FILTER file.repo_id == @repoID
        
        // Count inbound dependencies (files that import this)
        LET inbound = LENGTH(
            FOR v IN 1..1 INBOUND file._id code_dependencies RETURN v
        )
        
        // Count outbound dependencies (files this imports)
        LET outbound = LENGTH(
            FOR v IN 1..1 OUTBOUND file._id code_dependencies RETURN v
        )
        
        // Calculate importance score
        // Files with many dependents are more important
        LET importance = inbound * 2 + outbound * 0.5
        
        // Count commits touching this file
        LET commits = LENGTH(
            FOR commit IN git_commits
                LET tree = DOCUMENT(CONCAT('git_objects/', commit.tree))
                FILTER file._key IN tree.entries[* RETURN CURRENT.hash]
                RETURN commit
        )
        
        // Combined score
        LET score = importance + (commits * 0.1)
        
        SORT score DESC
        LIMIT 50
        RETURN {
            file: file.path,
            language: file.language,
            importance: importance,
            dependents: inbound,
            dependencies: outbound,
            commits: commits,
            score: score
        }
`
```

### Agent Productivity

```go
query := `
    FOR agent IN agents
        FILTER agent.type == "llm"
        
        // Get work items
        LET workItems = (
            FOR wi IN 1..1 OUTBOUND agent._id agent_work_items
                RETURN wi
        )
        
        // Get commits via work items
        LET commits = FLATTEN(
            FOR wi IN workItems
                FOR commit IN 1..1 OUTBOUND wi._id work_item_commits
                    RETURN commit
        )
        
        // Get expertise
        LET expertise = (
            FOR file, edge IN 1..1 OUTBOUND agent._id agent_code_expertise
                RETURN {file: file.path, expertise: edge.expertise}
        )
        
        // Calculate productivity score
        LET productivity = LENGTH(workItems) * 1.0 + LENGTH(commits) * 0.5
        
        RETURN {
            agent: agent.name,
            work_items_total: LENGTH(workItems),
            work_items_completed: LENGTH(workItems[* FILTER CURRENT.status == "completed"]),
            commits: LENGTH(commits),
            files_known: LENGTH(expertise),
            avg_expertise: AVG(expertise[* RETURN CURRENT.expertise]),
            productivity_score: productivity,
            top_files: expertise[* SORT CURRENT.expertise DESC LIMIT 5]
        }
`
```

### Code Churn Analysis

```go
query := `
    FOR file IN git_objects
        FILTER file.type == "blob"
        FILTER file.repo_id == @repoID
        
        // Find all commits that touched this file
        LET commits = (
            FOR commit IN git_commits
                LET tree = DOCUMENT(CONCAT('git_objects/', commit.tree))
                FILTER file._key IN tree.entries[* RETURN CURRENT.hash]
                SORT commit.committed_at DESC
                RETURN {
                    hash: commit._key,
                    date: commit.committed_at,
                    author: commit.author.email
                }
        )
        
        // Calculate churn score
        LET churnScore = LENGTH(commits)
        
        // Recent activity (last 30 days)
        LET recentCommits = LENGTH(
            commits[* FILTER DATE_DIFF(CURRENT.date, DATE_NOW(), "d") >= -30]
        )
        
        FILTER churnScore > 5  // Only files with significant churn
        SORT churnScore DESC
        LIMIT 50
        
        RETURN {
            file: file.path,
            language: file.language,
            total_commits: churnScore,
            recent_commits: recentCommits,
            unique_authors: LENGTH(UNIQUE(commits[* RETURN CURRENT.author])),
            last_modified: commits[0].date,
            risk_score: churnScore * (recentCommits > 0 ? 1.5 : 1.0)
        }
`
```

## Dashboards

### Real-Time Dashboard

```go
type DashboardData struct {
    // Current state
    ActiveWorkItems   int64              `json:"active_work_items"`
    QueuedWorkItems   int64              `json:"queued_work_items"`
    
    // Recent activity (last 24h)
    CompletedToday    int64              `json:"completed_today"`
    FailedToday       int64              `json:"failed_today"`
    
    // LLM usage
    TokensUsedToday   int64              `json:"tokens_used_today"`
    CostToday         float64            `json:"cost_today"`
    
    // Top contributors
    TopAgents         []AgentSummary     `json:"top_agents"`
    
    // Recent completions
    RecentCompletions []WorkItemSummary  `json:"recent_completions"`
    
    // System health
    Health            SystemHealth       `json:"health"`
}

type SystemHealth struct {
    GiteaStatus       string             `json:"gitea_status"`
    ArangoDBStatus    string             `json:"arangodb_status"`
    LLMStatus         string             `json:"llm_status"`
    ErrorRate         float64            `json:"error_rate"`
}
```

### API Endpoint

```go
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
    data := DashboardData{
        ActiveWorkItems:  h.repo.CountWorkItems("executing"),
        QueuedWorkItems:  h.repo.CountWorkItems("pending"),
        CompletedToday:   h.repo.CountCompletedToday(),
        FailedToday:      h.repo.CountFailedToday(),
        TokensUsedToday:  h.repo.GetTokensUsedToday(),
        CostToday:        h.repo.GetCostToday(),
        TopAgents:        h.repo.GetTopAgents(10),
        RecentCompletions: h.repo.GetRecentCompletions(20),
        Health:           h.getSystemHealth(),
    }
    
    c.JSON(200, data)
}
```

## Alerting

### Alert Conditions

```go
type AlertRule struct {
    Name        string
    Condition   string  // AQL query
    Threshold   float64
    Severity    string  // "info", "warning", "critical"
    Actions     []AlertAction
}

var alertRules = []AlertRule{
    {
        Name:      "High Error Rate",
        Condition: "error_rate > @threshold",
        Threshold: 0.1,  // 10%
        Severity:  "critical",
        Actions:   []AlertAction{SendEmail, PostSlack},
    },
    {
        Name:      "Budget Exceeded",
        Condition: "cost_today > @threshold",
        Threshold: 100.0,  // $100
        Severity:  "warning",
        Actions:   []AlertAction{SendEmail},
    },
    {
        Name:      "Long Queue",
        Condition: "queued_work_items > @threshold",
        Threshold: 50,
        Severity:  "warning",
        Actions:   []AlertAction{PostSlack},
    },
}
```

### Alert Monitoring

```go
func (m *AlertMonitor) CheckAlerts() {
    for _, rule := range alertRules {
        triggered, value := m.evaluateRule(rule)
        
        if triggered {
            alert := Alert{
                Rule:      rule.Name,
                Severity:  rule.Severity,
                Value:     value,
                Threshold: rule.Threshold,
                Timestamp: time.Now(),
            }
            
            for _, action := range rule.Actions {
                m.executeAction(action, alert)
            }
        }
    }
}
```

## Logs

### Structured Logging

```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

logger.Info("work item execution started",
    "work_item_id", workItem.ID,
    "type", workItem.Type,
    "issue_id", workItem.IssueID,
)

logger.Error("LLM generation failed",
    "work_item_id", workItem.ID,
    "error", err.Error(),
    "model", agent.Model,
    "tokens", estimatedTokens,
)
```

### Log Aggregation

Store logs in ArangoDB for querying:

```go
type LogEntry struct {
    Level       string                 `json:"level"`
    Message     string                 `json:"message"`
    Attributes  map[string]interface{} `json:"attributes"`
    Timestamp   time.Time              `json:"timestamp"`
}

// Query recent errors
query := `
    FOR log IN logs
        FILTER log.level == "ERROR"
        FILTER log.timestamp >= DATE_SUBTRACT(DATE_NOW(), 1, "hour")
        SORT log.timestamp DESC
        LIMIT 100
        RETURN log
`
```

---

**See Also**:
- [Data Models](./data-models.md) - Schema for metrics collections
- [Deployment](./deployment.md) - Monitoring infrastructure setup
