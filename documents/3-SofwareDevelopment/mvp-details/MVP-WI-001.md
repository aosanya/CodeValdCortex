# MVP-WI-001: Gitea Webhook Integration

## Overview
**Priority**: P0  
**Effort**: Medium  
**Skills Required**: Go, Webhooks, REST API, Security  
**Dependencies**: MVP-032 (Work Items Assignment & Routing)  
**Status**: In Progress

## Description
Implement webhook handlers for Gitea issue milestone events that enable **automatic agent instantiation** based on Kanban workflow columns. When issues are assigned to milestones (representing Kanban columns), agents are automatically created from the work item definitions linked to those columns.

### Kanban-Based Conceptual Model

**Workflow = Kanban Board**:
- **Workflow**: A published agency contains workflows that define Kanban boards
- **Columns**: Each workflow has columns (e.g., "Requirements", "Design", "Implementation")
- **Work Item Definitions**: Each column links to a work item definition (agent blueprint)
- **Gitea Milestones**: Map to workflow columns (e.g., Milestone "Requirements Gathering" → Column "requirements")

**Integration Flow**:
```
Published Agency Workflow (Kanban Board)
  ├─ Column: Requirements → WorkItemDef: Requirements Agent
  ├─ Column: Design       → WorkItemDef: Design Agent
  ├─ Column: Implementation → WorkItemDef: Coding Agent
  └─ Column: Review       → WorkItemDef: Review Agent

Gitea Repository
  ├─ Milestone: "Requirements Gathering"
  ├─ Milestone: "System Design"
  ├─ Milestone: "Implementation"
  └─ Milestone: "Code Review"

Issue Flow:
1. Issue created: "Gather auth requirements"
2. Assigned to Milestone: "Requirements Gathering"
3. Webhook fires: issues.milestoned
4. CodeValdCortex:
   - Finds repository→workflow mapping
   - Maps milestone to workflow column
   - Gets work item definition for that column
   - Creates Requirements Agent instance
   - Agent starts work on the issue
5. Agent completes, moves issue to next milestone
6. Process repeats with next agent type
```

**Example**:
1. Agency publishes workflow with 4 columns (Requirements → Design → Code → Review)
2. Each column has an agent blueprint (work item definition)
3. Repository linked to workflow, milestones created matching columns
4. Issue: "Build auth system" assigned to "Requirements" milestone
5. Requirements Agent auto-created, gathers requirements, posts to issue
6. Agent moves issue to "Design" milestone when done
7. Design Agent auto-created, creates architecture, posts diagrams
8. Cycle continues through Implementation and Review

## Objectives
- Enable **Kanban-based workflow automation** where issues flow through workflow stages
- Map Gitea milestones to workflow columns for visual Kanban board management
- Automatically create agents when issues enter workflow columns (WIP-limited)
- Link repository → workflow → work item definitions for seamless integration
- Support agent handoff as issues progress through workflow stages
- Provide secure webhook signature validation to prevent unauthorized access
- Handle milestone events (assigned, changed, cleared) to manage agent lifecycle

## Requirements

### Functional Requirements
1. **Webhook Endpoint**: POST /api/v1/webhooks/gitea/issues
2. **Supported Events**:
   - `issues.milestoned` - Issue assigned to milestone (enters Kanban column)
   - `issues.demilestoned` - Issue removed from milestone (exits column)
   - `issues.closed` - Stop associated agent
3. **Repository-Workflow Mapping**:
   - Query repository → workflow mapping by repository URL
   - Validate workflow exists and is enabled
   - Retrieve milestone → column mappings
4. **Milestone-Column Mapping**:
   - Extract milestone name from webhook payload
   - Map to workflow column using stored configuration
   - Get work item definition ID from column config
5. **WIP Limit Enforcement**:
   - Count active agents in target column
   - Check against column's `max_concurrent` setting
   - Queue issue if WIP limit exceeded
6. **Agent Instantiation**:
   - Create agent instance from column's work item definition
   - Inject issue context (title, description, milestone, repository)
   - Start agent execution asynchronously
   - Link agent to issue and workflow column
7. **Agent Lifecycle Management**:
   - Track agent status per column
   - Stop agents when issues leave columns or close
   - Handle agent failures and retries
   - Move issues to next column when agent completes
8. **Signature Validation**: HMAC SHA-256 verification using X-Gitea-Signature header
9. **Error Handling**: Post comments to issues on errors, queue for retry

### Non-Functional Requirements
1. **Security**: 
   - HMAC SHA-256 signature validation on all webhooks
   - Configurable secret token stored in config.yaml
   - Optional IP allowlist for webhook sources
2. **Performance**:
   - Asynchronous processing (handle webhook in goroutine)
   - Response time <200ms for webhook acceptance
   - Non-blocking work item execution
3. **Reliability**:
   - Idempotent handlers (handle duplicate webhooks)
   - Graceful error recovery
   - Audit logging for all webhook events
4. **Observability**:
   - Structured logging with MVP-WI-001 prefix
   - Webhook delivery status tracking
   - Error rate monitoring

## Acceptance Criteria
- [ ] POST /api/v1/webhooks/gitea/issues endpoint created and registered in Gin router
- [ ] X-Gitea-Signature HMAC SHA-256 validation working (rejects invalid signatures)
- [ ] Repository→Workflow mapping stored in ArangoDB and queryable by repository URL
- [ ] Milestone→Column mapping extracted from webhook payload
- [ ] Issue "milestoned" event triggers workflow column lookup
- [ ] Column's work item definition retrieved from ArangoDB
- [ ] WIP (Work-In-Progress) limit checked before agent creation
- [ ] Issues queued when WIP limit exceeded for column
- [ ] Agent instance created from work item definition with milestone/issue context
- [ ] Agent starts execution asynchronously (non-blocking webhook response)
- [ ] Agent linked to issue, workflow, and column in tracking table
- [ ] Issue "demilestoned" event stops associated agent
- [ ] Issue "closed" event stops running agent and marks as completed
- [ ] Success/progress comments posted to Gitea issues by agents
- [ ] Configuration added to config.yaml (webhook secret, allowed IPs)
- [ ] All errors logged with MVP-WI-001 prefix
- [ ] Webhook processing returns 200 OK within 200ms
- [ ] Unit tests cover signature validation, milestone extraction, WIP limits, agent creation
- [ ] Integration tests with real Gitea instance verify Kanban flow (issue moves through milestones)
- [ ] Documentation: Kanban setup guide, repository-workflow linking, milestone configuration

## Technical Specifications

### Architecture

**Package Structure**:
```
internal/infrastructure/webhooks/gitea/
├── handler.go           # HTTP endpoint handlers
├── validator.go         # Signature validation
├── processor.go         # Event processing logic
├── matcher.go           # Label → WorkItemDefinition matching
├── factory.go           # Agent instantiation from definitions
└── handler_test.go      # Unit tests
```

**Dependencies**:
- `code.gitea.io/sdk/gitea` - Official Gitea Go SDK for types
- `crypto/hmac` - HMAC signature validation
- `github.com/gin-gonic/gin` - HTTP router
- `internal/agent` - Agent creation and lifecycle management
- `internal/domain/workflow/models` - WorkItemDefinition types
- `internal/domain/workflow/repository` - WorkItemDefinition queries

### Data Flow
```
Gitea Issue Event (with label: "requirements-gathering")
    ↓
Webhook POST → handler.go (signature validation)
    ↓
processor.go (parse payload, extract labels)
    ↓
matcher.go (label → work item definition lookup)
    ↓
    Query ArangoDB: work_item_definitions WHERE type = "requirements-gathering"
    ↓
factory.go (create agent from definition + issue context)
    ↓
    AgentFactory.CreateFromWorkItemDefinition()
    ↓
Agent Instance Created in ArangoDB (agents collection)
    ↓
Agent Started (agent.Start())
    ↓
Link Agent ↔ Issue (agent_issue_tracking table)
    ↓
Post comment to Gitea: "✅ Requirements Gathering Agent created (ID: abc-123)"
```

**Key Data Structures**:

```go
// Work Item Definition (blueprint in published agency)
type WorkItemDefinition struct {
    ID           string   `json:"id"`
    Type         string   `json:"type"`          // e.g., "requirements-gathering"
    Name         string   `json:"name"`
    Description  string   `json:"description"`
    Instructions string   `json:"instructions"`  // What the agent should do
    Tools        []string `json:"tools"`         // Tools agent can use
    AutonomyLevel string  `json:"autonomy_level"`// "supervised", "autonomous"
    AgencyID     string   `json:"agency_id"`
}

// Agent Instance (created from definition)
type Agent struct {
    ID               string    `json:"id"`
    Name             string    `json:"name"`
    Type             string    `json:"type"`           // From WorkItemDefinition.Type
    WorkItemDefID    string    `json:"work_item_def_id"` // Link to definition
    TriggerSource    string    `json:"trigger_source"` // "gitea"
    GiteaIssueID     int64     `json:"gitea_issue_id"`
    GiteaIssueURL    string    `json:"gitea_issue_url"`
    IssueContext     IssueContext `json:"issue_context"`
    Status           string    `json:"status"`         // "created", "running", "completed"
    CreatedAt        time.Time `json:"created_at"`
}

// Issue Context (injected into agent)
type IssueContext struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Repository  string `json:"repository"`
    Labels      []string `json:"labels"`
}
```

### Signature Validation

```go
func validateSignature(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedMAC := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
```

### Issue Label → WorkItemDefinition Matching

**Label Convention**:
- Work item type labels use format: `{work-item-type}`
- Examples: `requirements-gathering`, `code-review`, `security-audit`, `testing`

**Matching Logic**:
```go
func (m *Matcher) FindWorkItemDefinition(labels []string, agencyID string) (*WorkItemDefinition, error) {
    // 1. Extract work item type from labels
    workItemType := extractWorkItemType(labels)
    if workItemType == "" {
        return nil, errors.New("no work item type label found")
    }
    
    // 2. Query ArangoDB for definition
    query := `
        FOR def IN work_item_definitions
            FILTER def.type == @type
            AND def.agency_id == @agencyID
            AND def.enabled == true
            LIMIT 1
            RETURN def
    `
    
    cursor, err := m.db.Query(ctx, query, map[string]interface{}{
        "type":     workItemType,
        "agencyID": agencyID,
    })
    
    var def WorkItemDefinition
    cursor.ReadDocument(ctx, &def)
    return &def, nil
}

func extractWorkItemType(labels []string) string {
    // Known work item type labels
    workItemTypes := []string{
        "requirements-gathering",
        "code-review",
        "security-audit",
        "testing",
        "deployment",
        "documentation",
    }
    
    for _, label := range labels {
        for _, validType := range workItemTypes {
            if label == validType {
                return validType
            }
        }
    }
    return ""
}
```

| Gitea Label | WorkItemDefinition.Type | Agent Created |
|-------------|------------------------|---------------|
| `requirements-gathering` | `requirements-gathering` | Requirements Gathering Agent |
| `code-review` | `code-review` | Code Review Agent |
| `security-audit` | `security-audit` | Security Audit Agent |
| `testing` | `testing` | Testing Agent |

### Configuration Schema

```yaml
gitea:
  webhooks:
    secret: "${GITEA_WEBHOOK_SECRET}"  # HMAC secret token
    allowed_ips:                        # Optional IP allowlist
      - "10.0.0.0/8"
      - "172.16.0.0/12"
    timeout: 30s                        # Webhook processing timeout
    retry:
      max_attempts: 3
      backoff: exponential
```

### Error Handling

| Error Scenario | HTTP Status | Action |
|---------------|-------------|--------|
| Invalid signature | 401 Unauthorized | Log warning, reject webhook |
| Malformed JSON | 400 Bad Request | Log error, return details |
| Missing required fields | 422 Unprocessable Entity | Log validation errors |
| Database error | 500 Internal Server Error | Log error, trigger retry |
| Unknown event type | 200 OK | Log info, skip processing |

### Logging Format

```go
log.Info().
    Str("task", "MVP-WI-001").
    Str("event", "issues.opened").
    Int64("issue_id", payload.Issue.Index).
    Str("repository", payload.Repository.FullName).
    Msg("Processing Gitea webhook")
```

## Implementation Phases

### Phase 1: Foundation (Day 1)
- Create package structure
- Add Gitea SDK dependency
- Implement signature validation
- Create HTTP endpoint

### Phase 2: Event Handlers (Day 2)
- Implement issue.opened handler
- Implement issue.labeled handler
- Implement issue.closed handler
- Add payload → WorkItem mapping

### Phase 3: Configuration & Error Handling (Day 3)
- Add config.yaml integration
- Implement comprehensive error handling
- Add structured logging
- Add retry logic

### Phase 4: Testing & Documentation (Day 4)
- Write unit tests (signature, parsing, mapping)
- Set up Gitea in docker-compose
- Run integration tests
- Write setup documentation

## Testing Strategy

### Unit Tests
```go
func TestValidateSignature(t *testing.T) { /* ... */ }
func TestParseIssuePayload(t *testing.T) { /* ... */ }
func TestMapIssueToWorkItem(t *testing.T) { /* ... */ }
func TestHandleIssueOpened(t *testing.T) { /* ... */ }
```

### Integration Tests
1. Start Gitea in docker-compose
2. Create test repository
3. Configure webhook pointing to localhost:8080
4. Create issue with "work-item" label
5. Verify work item created in ArangoDB
6. Close issue, verify work item status updated

## Dependencies

**Upstream (Must Complete First)**:
- ✅ MVP-029: Work Items Architecture
- ✅ MVP-030: Work Items Core Schema & Registry
- ✅ MVP-031: Work Items Lifecycle & SLA
- ✅ MVP-032: Work Items Assignment & Routing

**Downstream (Blocks)**:
- MVP-WI-002: Gitea API Client (uses webhook data for bidirectional sync)
- MVP-WI-003: Work Item to Issue Sync (reverse direction)
- MVP-WI-004: Pull Request Automation (builds on webhook infrastructure)

## Success Metrics
- Webhook signature validation: 100% rejection of invalid signatures
- Event processing latency: <200ms to accept webhook, <2s to persist work item
- Error rate: <1% for valid webhooks
- Zero data loss: All valid webhooks result in work item creation/update

## References
- **Architecture**: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/gitea-integration.md`
- **Gitea SDK**: https://gitea.com/gitea/go-sdk
- **Webhook Documentation**: https://docs.gitea.io/en-us/webhooks/
- **Work Items System**: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/`
