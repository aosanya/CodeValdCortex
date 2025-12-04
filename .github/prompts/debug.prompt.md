---
mode: agent
---

# Debug Workflow Save Button Not Firing

## Current Issue
The save button in the workflow editor card (`internal/web/pages/agency_designer/workflow_editor_card.templ`) is not triggering when clicked.

**Task ID**: MVP-054 (from branch: `feature/MVP-054_work_items_enhanced_deliverables`)

## Code Context

### Template File
**Location**: `internal/web/pages/agency_designer/workflow_editor_card.templ`

Button definition:
```html
<button 
id="save-workflow-btn"
class="button is-small is-primary"
onclick="saveWorkflowFromEditor()">
<span class="icon"><i class="fas fa-save"></i></span>
<span>Save</span>
</button>
```

### JavaScript Function
**Location**: `static/js/agency-designer/workflows.js` (lines 96-127)

```javascript
window.saveWorkflowFromEditor = function () {
    const name = document.getElementById('workflow-name-editor').value.trim();
    const description = document.getElementById('workflow-description-editor').value.trim();
    const version = document.getElementById('workflow-version-editor').value.trim();

    if (!name) {
        window.showNotification('Please provide a workflow name', 'error');
        return;
    }

    if (!description) {
        window.showNotification('Please provide a workflow description', 'error');
        return;
    }

    const data = { name, description, version, ... };

    window.saveEntity('workflows', workflowEditorState.mode, 
                     workflowEditorState.workflowId, data, 
                     'save-workflow-btn', () => {
        cancelWorkflowEdit();
        loadWorkflows();
    });
}
```

## Debug Instructions

### Step 1: Browser Console Checks
Open browser DevTools (F12) → Console tab, run:

```javascript
// 1. Check if function exists
console.log('[MVP-054] saveWorkflowFromEditor type:', typeof window.saveWorkflowFromEditor);

// 2. Check button exists
console.log('[MVP-054] Save button:', document.getElementById('save-workflow-btn'));

// 3. Check form fields exist
console.log('[MVP-054] Name field:', document.getElementById('workflow-name-editor'));
console.log('[MVP-054] Description field:', document.getElementById('workflow-description-editor'));
console.log('[MVP-054] Version field:', document.getElementById('workflow-version-editor'));

// 4. Check card visibility
const card = document.getElementById('workflow-editor-card');
console.log('[MVP-054] Card classes:', card?.className);
console.log('[MVP-054] Card is hidden:', card?.classList.contains('is-hidden'));

// 5. Check for duplicates
console.log('[MVP-054] Editor cards count:', document.querySelectorAll('#workflow-editor-card').length);
console.log('[MVP-054] Save buttons count:', document.querySelectorAll('#save-workflow-btn').length);
```

### Step 2: Manual Function Test
```javascript
// Try calling function directly
console.log('[MVP-054] Calling saveWorkflowFromEditor...');
window.saveWorkflowFromEditor();
```

### Step 3: Add Debug Logging
Add to `static/js/agency-designer/workflows.js`:

```javascript
window.saveWorkflowFromEditor = function () {
    console.log('[MVP-054] 🔵 saveWorkflowFromEditor CALLED');
    
    const name = document.getElementById('workflow-name-editor')?.value.trim();
    const description = document.getElementById('workflow-description-editor')?.value.trim();
    const version = document.getElementById('workflow-version-editor')?.value.trim();
    
    console.log('[MVP-054] 📝 Form values:', { name, description, version });
    console.log('[MVP-054] 🔧 State:', { 
        mode: workflowEditorState.mode, 
        workflowId: workflowEditorState.workflowId 
    });

    if (!name) {
        console.log('[MVP-054] ⚠️ VALIDATION FAILED: Missing name');
        window.showNotification('Please provide a workflow name', 'error');
        return;
    }

    if (!description) {
        console.log('[MVP-054] ⚠️ VALIDATION FAILED: Missing description');
        window.showNotification('Please provide a workflow description', 'error');
        return;
    }
    
    console.log('[MVP-054] ✅ Validation passed, calling saveEntity');
    
    const data = {
        name,
        description,
        version,
        nodes: workflowEditorState.originalData?.nodes || [],
        edges: workflowEditorState.originalData?.edges || [],
        variables: workflowEditorState.originalData?.variables || {}
    };
    
    console.log('[MVP-054] 📦 Data to save:', data);

    window.saveEntity('workflows', workflowEditorState.mode, workflowEditorState.workflowId, data, 'save-workflow-btn', () => {
        console.log('[MVP-054] ✅ Save callback executing');
        cancelWorkflowEdit();
        loadWorkflows();
    });
}
```

## Common Failure Scenarios

### Scenario 1: Form Validation Failing
**Symptom**: No console output, no notification
**Cause**: Name or description fields are empty
**Solution**: Fill in both fields before clicking save

### Scenario 2: Function Not Attached to Window
**Symptom**: `TypeError: window.saveWorkflowFromEditor is not a function`
**Cause**: Script not loaded or function not exported
**Check**: Bottom of `workflows.js` should have `window.saveWorkflowFromEditor = saveWorkflowFromEditor;`

### Scenario 3: Card is Hidden
**Symptom**: Button exists but not visible/clickable
**Cause**: Card has `is-hidden` class
**Solution**: Click "Add Workflow" button first to show editor

### Scenario 4: workflowEditorState Undefined
**Symptom**: Error accessing `workflowEditorState.mode`
**Cause**: State not initialized
**Check**: Top of `workflows.js` should initialize state object

### Scenario 5: saveEntity Function Missing
**Symptom**: `window.saveEntity is not a function`
**Cause**: `crud-helpers.js` not loaded before `workflows.js`
**Check**: Script load order in `agency_designer.templ`

## Debug Print Guidelines (General)

### Prefix Format
All debug prints MUST be prefixed with: `[TASK-ID]`

### JavaScript/TypeScript
```javascript
console.log('[MVP-054] Function called:', functionName, 'with args:', args);
console.log('[MVP-054] State before:', JSON.stringify(state, null, 2));
console.error('[MVP-054] Error in operation:', error.message);
```

### Go
```go
log.Printf("[MVP-054] Function called: %s with args: %+v", functionName, args)
log.Printf("[MVP-054] State before: %+v", state)
log.Printf("[MVP-054] Error in operation: %v", err)
```

### Python
```python
print(f"[MVP-054] Function called: {function_name} with args: {args}")
print(f"[MVP-054] State before: {json.dumps(state, indent=2)}")
print(f"[MVP-054] Error in operation: {error}", file=sys.stderr)
```

### Strategic Placement

Add debug prints at:

1. **Function Entry Points**
   - Log function name and key parameters
   - Example: `console.log('[MVP-052] splitEdge called:', edge.id, newNodeId);`

2. **State Changes**
   - Before and after critical state modifications
   - Example: `console.log('[MVP-052] Edges before split:', this.edges.length);`

3. **Conditional Branches**
   - Log which branch is taken and why
   - Example: `console.log('[MVP-052] Distance check:', distance, '> 50?', distance > 50);`

4. **Loop Iterations** (for complex loops)
   - Log iteration count and key variables
   - Example: `console.log('[MVP-052] Processing edge', i, '/', edges.length);`

5. **API Calls / Async Operations**
   - Before request and after response
   - Example: `console.log('[MVP-052] Fetching workflow:', workflowId);`

6. **Error Handling**
   - Log errors with context
   - Example: `console.error('[MVP-052] Failed to split edge:', error, 'edge:', edge);`

7. **Return Statements** (for complex functions)
   - Log what is being returned
   - Example: `console.log('[MVP-052] Returning nearEdge:', nearEdge);`

### What NOT to Debug

Avoid adding debug prints to:
- Simple getters/setters
- Trivial utility functions
- High-frequency event handlers (unless specifically investigating performance)
- Already well-instrumented code

### Debug Print Structure

Use descriptive messages that answer:
1. **WHERE**: Which function/block is executing
2. **WHAT**: What operation is happening
3. **VALUES**: Relevant variable values
4. **CONTEXT**: Why this matters (optional)

**Good Example:**
```javascript
console.log('[MVP-052] checkAutoDisconnect: Node moved', {
    nodeId,
    distance,
    threshold: 50,
    willDisconnect: distance > 50
});
```

**Bad Example:**
```javascript
console.log('here'); // No context, no task ID, not helpful
```

### Cleanup Instructions

Always add a comment above debug blocks:
```javascript
// TODO: Remove debug prints for MVP-052 after issue is resolved
console.log('[MVP-052] Debug info here');
```

## Execution Steps

1. **Identify Task ID** from branch or context
2. **Analyze Code** to understand the flow and issue
3. **Select Strategic Points** where debug prints will be most valuable
4. **Add Debug Prints** with proper format and task ID prefix
5. **Explain Placement** briefly why each debug print was added

## Output Format

After adding debug prints, provide:

```markdown
### Debug Prints Added for [TASK-ID]

**File**: `path/to/file.ext`

**Locations**:
1. Line XX: Function entry - logs parameters
2. Line YY: State change - logs before/after values
3. Line ZZ: Conditional check - logs decision logic

**Usage**: Run the code and filter logs with `grep "[MVP-052]"` or check browser console.
```

## Example Request Handling

**User**: "Add debug prints to track edge splitting"

**Your Response**:
1. Check git branch → Extract task ID (e.g., MVP-052)
2. Analyze edge splitting code flow
3. Add strategic prints at:
   - `findNearbyEdge` entry and result
   - `splitEdge` entry, edge removal, new connections
   - `checkAutoDisconnect` distance calculations
4. Use format: `console.log('[MVP-052] ...')`
5. Explain what each print reveals

## Remember

- **Always** use task ID prefix
- **Be strategic** - don't over-instrument
- **Be descriptive** - logs should tell a story
- **Be consistent** - use same format throughout
- **Be removable** - add TODO comments for cleanup
