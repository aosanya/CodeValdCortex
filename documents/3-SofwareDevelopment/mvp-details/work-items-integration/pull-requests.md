# Pull Request Automation

<!-- MVP-WI-004 -->
**Tasks Covered**: MVP-WI-004  
**Status**: 🔄 In Progress

## Overview

Automate the complete pull request workflow for agent-generated code changes, from branch creation through merge and issue closure. This component enables agents to submit their work for review, handles automated quality checks, and completes the development cycle by merging approved changes and updating related issues.

## Problem Statement

When agents complete coding tasks, they need a streamlined way to:
1. Submit code changes for human review (create PR)
2. Link their work to the originating issue
3. Trigger automated quality checks (tests, linting, security scans)
4. Automatically merge when all checks pass and approval received
5. Update issue state to reflect merged code
6. Track which agent generated which code changes

**Without this capability**, agents would need manual intervention to:
- Create branches and PRs manually
- Copy/paste agent output into Git
- Manually link PRs to issues
- Wait for human to merge even if checks pass
- Manually close/update issues after merge

**With automation**, the entire code submission → review → merge → issue update cycle becomes autonomous.

## Objectives

1. **PR Creation Automation**: Agents create PRs directly from their work outputs
2. **Issue Linking**: Automatic bi-directional linking between PRs and issues
3. **Quality Gate Integration**: Run tests, linting, security scans before merge
4. **Auto-Merge**: Merge approved PRs when all checks pass
5. **Issue State Updates**: Update issue milestone/status on PR events
6. **Agent Attribution**: Track code changes by agent for audit/analytics

## Architecture

See [architecture/pr-workflow.md](./architecture/pr-workflow.md) for detailed component designs and data flow diagrams.

### Components

**1. PR Service** (`internal/infrastructure/work/pr_service.go`)
**2. Git Operations Handler** (`internal/infrastructure/work/git_operations.go`)
**3. PR Event Handler** (`internal/infrastructure/work/pr_event_handler.go`)
**4. Quality Check Service** (`internal/infrastructure/work/quality_check_service.go`)
**5. Auto-Merge Engine** (`internal/infrastructure/work/auto_merge_engine.go`)

See [architecture/pr-workflow.md](./architecture/pr-workflow.md) for interface definitions and implementation details.

## Data Models

### Pull Request Info
```go
type PRInfo struct {
    ID             string    `json:"id"`
    Number         int64     `json:"number"`
    Title          string    `json:"title"`
    Description    string    `json:"description"`
    RepositoryURL  string    `json:"repository_url"`
    SourceBranch   string    `json:"source_branch"`
    TargetBranch   string    `json:"target_branch"`
    State          string    `json:"state"` // open, merged, closed
    CreatedBy      string    `json:"created_by"` // agent ID
    AgentID        string    `json:"agent_id"`
    LinkedIssueID  string    `json:"linked_issue_id"`
    CommitSHA      string    `json:"commit_sha"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
    MergedAt       *time.Time `json:"merged_at,omitempty"`
}
```

### Quality Check Results
```go
type QualityCheckResults struct {
    PRID           string              `json:"pr_id"`
    CheckTimestamp time.Time           `json:"check_timestamp"`
    TestResults    *TestResults        `json:"test_results"`
    LintResults    *LintResults        `json:"lint_results"`
    SecurityScan   *SecurityResults    `json:"security_scan"`
    Coverage       *CoverageReport     `json:"coverage"`
    PolicyCheck    *PolicyCheckResult  `json:"policy_check"`
    OverallStatus  string              `json:"overall_status"` // pass, fail, pending
}
```

### Auto-Merge Decision
```go
type AutoMergeDecision struct {
    ShouldMerge  bool               `json:"should_merge"`
    Reason       string             `json:"reason"`
    ChecksPassed map[string]bool    `json:"checks_passed"`
    BlockedBy    []string           `json:"blocked_by,omitempty"`
    MergeStrategy string            `json:"merge_strategy"` // merge, squash, rebase
}
```

See [architecture/pr-workflow.md](./architecture/pr-workflow.md) for complete model definitions.

## Workflow

### 1. Agent Creates PR

Agent completes coding task → Creates branch → Pushes changes → Creates PR → Links to issue

### 2. PR Webhook Handler

Handles events:
- `pull_request.opened` → Trigger quality checks
- `pull_request.synchronized` → Re-run checks on new commits
- `pull_request.approved` → Evaluate auto-merge
- `pull_request.merged` → Update issue state
- `pull_request.closed` → Clean up branch

### 3. Quality Checks

Runs in parallel:
- Test execution (CI/CD pipeline)
- Linting (golangci-lint, eslint)
- Security scanning (trivy, snyk, gosec)
- Code coverage analysis
- Policy compliance checks

### 4. Auto-Merge Evaluation

Decision based on:
- ✅ All tests passing
- ✅ No lint errors
- ✅ No high/critical vulnerabilities
- ✅ Coverage threshold met (default: 80%)
- ✅ Approvals received (if required)

### 5. PR Merged → Issue Update

- Update issue milestone
- Post merge notification comment
- Update issue labels (add "merged", remove "in-progress")
- Record audit trail

## ArangoDB Schema

**Collection**: `agent_prs`

**Example Document**:
```json
{
  "_key": "pr-123",
  "pr_number": 45,
  "title": "Add authentication module",
  "repository_url": "https://gitea.example.com/org/repo",
  "source_branch": "agent/agent-uuid-abc/auth-feature",
  "target_branch": "main",
  "state": "merged",
  "agent_id": "agent-uuid-abc",
  "linked_issue_id": "123",
  "created_at": "2024-12-29T10:00:00Z",
  "merged_at": "2024-12-29T10:30:00Z",
  "quality_checks": {
    "tests": {"status": "passed"},
    "lint": {"status": "passed"},
    "security": {"status": "passed"},
    "coverage": {"percentage": 87.5}
  }
}
```

**Indexes**:
- `agent_id` - Query PRs by agent
- `linked_issue_id` - Find PR for issue
- `state` - Filter by PR state
- `repository_url + state` - Active PRs per repo

## Configuration

```yaml
auto_merge:
  enabled: true
  require_approval: false  # For agent PRs
  min_approvals: 0
  require_tests_pass: true
  require_lint_pass: true
  require_security_scan: true
  min_coverage_percent: 80.0
  block_on_high_vulns: true
  merge_strategy: "squash"
  delete_branch_after: true
```

## Templates

### PR Description Template
```markdown
## 🤖 Agent-Generated Pull Request

**Agent**: `{{ .AgentID }}`  
**Linked Issue**: #{{ .IssueID }}  
**Created**: {{ .CreatedAt }}

### Changes Summary
{{ .ChangesSummary }}

### Files Modified
{{ range .Files }}
- `{{ .Path }}` ({{ .LinesAdded }}+ / {{ .LinesDeleted }}-)
{{ end }}

### Quality Checks
- ✅ Tests: {{ .TestResults.PassedTests }}/{{ .TestResults.TotalTests }} passed
- ✅ Lint: {{ .LintResults.Errors }} errors
- ✅ Security: {{ .SecurityResults.Vulnerabilities | len }} vulnerabilities
- ✅ Coverage: {{ .Coverage.Percentage }}%

### Auto-Merge
{{ if .AutoMergeEnabled }}
⚡ Auto-merge is **enabled**. This PR will merge automatically when all checks pass.
{{ else }}
👤 Manual review required before merge.
{{ end }}
```

## Error Handling

### PR Creation Failures
- `BRANCH_EXISTS` - Branch name collision
- `PUSH_FAILED` - Git push failed
- `API_ERROR` - Gitea API error
- `VALIDATION_FAIL` - Invalid PR request
- `MERGE_CONFLICT` - Conflicts detected

Retry logic: 3 attempts with exponential backoff (except validation errors)

### Merge Conflicts
1. Detect conflicts before merge
2. Post conflict details to PR
3. Request agent to resolve
4. Block auto-merge until resolved

## Acceptance Criteria

- [ ] PRService interface implemented with all methods
- [ ] GitOperations handler supports branch/file operations
- [ ] PR event handler processes all PR webhook events
- [ ] Quality check service runs tests, lint, security scans
- [ ] Auto-merge engine evaluates merge readiness correctly
- [ ] PRs automatically linked to originating issues
- [ ] Issue state updates on PR merge/close
- [ ] Agent attribution tracked in PR metadata
- [ ] ArangoDB collection created with indexes
- [ ] PR description templates render correctly
- [ ] Configuration supports auto-merge rules
- [ ] Error handling covers all failure scenarios
- [ ] Merge conflict detection and notification
- [ ] Branch cleanup after successful merge
- [ ] Audit trail records all PR operations
- [ ] Integration tests for complete PR workflow
- [ ] API endpoints for PR management
- [ ] Documentation for PR automation setup

## Success Metrics

- **PR Creation Latency**: <5 seconds from agent request
- **Auto-Merge Accuracy**: >95% correct merge decisions
- **Quality Check Speed**: <10 minutes for complete check suite
- **Merge Success Rate**: >90% PRs merge successfully
- **Conflict Detection**: 100% conflicts detected before merge
- **Issue Linking Accuracy**: 100% PRs linked to correct issue

## Integration Points

**With MVP-WI-003** (Agent-to-Issue Sync):
- PR events trigger issue comment updates
- Issue milestone progression on PR merge
- Agent health status reflected in PR

**With MVP-032** (Agent Factory & Lifecycle):
- Agent completion triggers PR creation
- PR merge completes agent lifecycle
- Agent state machine transitions on PR events

**With Quality Systems**:
- CI/CD pipeline integration (GitHub Actions, GitLab CI)
- Linter integration (golangci-lint, eslint)
- Security scanner integration (trivy, snyk, gosec)
- Code coverage tools (go test -cover, codecov)

## Future Enhancements

1. **Intelligent Auto-Merge**: ML-based merge decision (learn from past PRs)
2. **Conflict Auto-Resolution**: Agent attempts to resolve simple conflicts
3. **Code Review AI**: Automated code review comments before human review
4. **Performance Testing**: Run performance benchmarks on PR
5. **Changelog Generation**: Auto-generate release notes from merged PRs
6. **Multi-Repository PRs**: Support changes spanning multiple repos
7. **Draft PR Support**: Create draft PRs for work-in-progress
8. **PR Templates**: Custom PR templates per repository/agency

## Related Topics

- See [api-client.md](./api-client.md) for Gitea API integration
- See [synchronization.md](./synchronization.md) for issue update workflows
- See [webhooks.md](./webhooks.md) for PR webhook handling
- See [architecture/pr-workflow.md](./architecture/pr-workflow.md) for detailed component designs
