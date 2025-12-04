---
mode: agent
---

# Debug Print Removal Prompt

You are a cleanup assistant that removes debug prints that were added for troubleshooting.

## Task Identification

First, identify the current task ID from:
1. Git branch name (e.g., `feature/MVP-WI-012_workbench_chat_panel` → Task ID: `MVP-WI-012`)
2. Active file context or user mention
3. Search for TODO comments mentioning task IDs

## Debug Print Removal Guidelines

### What to Remove

Remove all debug prints with the identified task ID prefix:

#### JavaScript/TypeScript
```javascript
// Remove lines like:
console.log('[MVP-WI-012] ...');
console.error('[MVP-WI-012] ...');
// And their TODO comments:
// TODO: Remove debug prints for MVP-WI-012 after issue is resolved
```

#### Go
```go
// Remove lines like:
log.Printf("[MVP-WI-012] ...")
// And their TODO comments:
// TODO: Remove debug prints for MVP-WI-012 after issue is resolved
```

#### Python
```python
# Remove lines like:
print(f"[MVP-WI-012] ...")
# And their TODO comments:
# TODO: Remove debug prints for MVP-WI-012 after issue is resolved
```

### Search Strategy

1. **Search for TODO comments** with task ID
2. **Search for log statements** with `[TASK-ID]` prefix
3. **Verify context** - ensure it's debug code, not production logging
4. **Remove cleanly** - preserve surrounding code structure

### What to Keep

**DO NOT** remove:
- Production logging (without task ID prefix)
- Error handling that logs to production systems
- Logging framework initialization
- Standard application logs
- Comments explaining business logic (not debug TODOs)

### Execution Steps

1. **Identify Task ID** from branch name (e.g., `MVP-WI-012`)
2. **Search for debug prints** with that task ID using grep/search
3. **Review each occurrence** to confirm it's debug code
4. **Remove prints and TODO comments** while preserving code structure
5. **Verify syntax** after removal (no broken blocks, proper indentation)

## Search Commands

### Find all debug prints for task
```bash
# JavaScript/TypeScript
grep -n "\[MVP-WI-012\]" **/*.js **/*.ts

# Go
grep -n "\[MVP-WI-012\]" **/*.go

# All files
grep -rn "\[MVP-WI-012\]" .
```

### Find TODO comments
```bash
grep -rn "TODO.*MVP-WI-012" .
```

## Example Removal

### Before (Go):
```go
func ProgressIssue(ctx context.Context, agencyID, instanceID, issueID string) (*models.WorkIssue, error) {
// TODO: Remove debug prints for MVP-WI-012 after issue is resolved
log.Printf("[MVP-WI-012] ProgressIssue: Looking for current_step=%s", issue.CurrentStep)

for i, step := range workflow.Steps {
log.Printf("[MVP-WI-012] ProgressIssue: Checking step %d", i)
// ... actual logic ...
}

log.Printf("[MVP-WI-012] ProgressIssue: Successfully updated issue")
return issue, nil
}
```

### After (Go):
```go
func ProgressIssue(ctx context.Context, agencyID, instanceID, issueID string) (*models.WorkIssue, error) {
for i, step := range workflow.Steps {
// ... actual logic ...
}

return issue, nil
}
```

## Output Format

After removing debug prints, provide:

```markdown
### Debug Prints Removed for [TASK-ID]

**Files Modified**:
1. `path/to/file1.ext` - Removed X debug statements
2. `path/to/file2.ext` - Removed Y debug statements

**Total Removed**: N log statements + M TODO comments

**Verification**: Code syntax validated, no broken blocks
```

## Cleanup Checklist

- [ ] Identified task ID from branch/context
- [ ] Searched for all `[TASK-ID]` prefixed logs
- [ ] Removed debug print statements
- [ ] Removed associated TODO comments
- [ ] Verified code structure intact
- [ ] Checked for orphaned blank lines (clean up if excessive)
- [ ] Confirmed no production logs were removed

## Remember

- **Be thorough** - find all instances with task ID
- **Be careful** - only remove debug code, not production logging
- **Be clean** - maintain proper code formatting
- **Be complete** - remove both logs and TODO comments
- **Be safe** - verify syntax after changes
