# Debug Prints Added for MVP-054: Deliverables Save Flow

## Task ID: MVP-054
**Task**: Work Items Enhanced Deliverables

## Purpose
Added strategic debug prints to trace the complete flow of deliverables data from the tree builder UI to the database persistence layer.

---

## Debug Prints Added

### 1. Frontend - work-items.js (Client-side Save Logic)

**File**: `/workspaces/CodeValdCortex/static/js/agency-designer/work-items.js`

**Locations**:
- **Line ~183**: Function entry - logs start of deliverables extraction
- **Line ~194**: Logs which mode is being used (simple text vs tree builder)
- **Line ~197**: Logs retrieved structured deliverables data
- **Line ~230**: Logs complete data object being sent to API (sanitized for console)
- **Line ~244**: Logs save operation mode (add vs update)
- **Line ~246-250**: Logs which API method is being called (addWorkItem or updateWorkItem)
- **Line ~253**: Logs successful save result
- **Line ~261**: Logs any errors during save

**What to look for**:
```javascript
[MVP-054] saveWorkItemFromEditor: Starting deliverables extraction
[MVP-054] Using tree builder mode, getDeliverablesStructuredData exists? true
[MVP-054] Retrieved structured deliverables: [array of nodes]
[MVP-054] Data object being sent to API: {...}
[MVP-054] About to save work item, mode: add
[MVP-054] Calling addWorkItem with data
[MVP-054] Work item saved successfully: {...}
```

---

### 2. Frontend - specification-api.js (API Client Layer)

**File**: `/workspaces/CodeValdCortex/static/js/agency-designer/specification-api.js`

**Locations**:
- **Line ~331**: addWorkItem - logs work item being added
- **Line ~334**: addWorkItem - logs updated work items array size
- **Line ~341**: updateWorkItem - logs work item key and update data
- **Line ~349**: updateWorkItem - logs which work item is being updated
- **Line ~354**: updateWorkItem - logs the merged work item data
- **Line ~133**: updateWorkItems - logs count of work items being sent
- **Line ~134-139**: updateWorkItems - logs summary of each work item (code, title, has_deliverables_structured)
- **Line ~146**: updateWorkItems - logs full JSON payload being sent to server
- **Line ~158**: updateWorkItems - logs API error responses
- **Line ~162**: updateWorkItems - logs successful API response
- **Line ~164**: updateWorkItems - logs any errors in the process

**What to look for**:
```javascript
[MVP-054] SpecificationAPI.addWorkItem called with: {...}
[MVP-054] Updated work items array (after add): 5 items
[MVP-054] SpecificationAPI.updateWorkItems: Sending 5 work items to API
[MVP-054] Work items being sent: [{code: 'WI-001', title: 'Test...', has_deliverables_structured: true, deliverables_structured_count: 3}]
[MVP-054] Full payload to /work-items endpoint: {"work_items": [...], "updated_by": "user"}
[MVP-054] API response received: {...}
```

---

### 3. Backend - agency_handler.go (HTTP Handler Layer)

**File**: `/workspaces/CodeValdCortex/internal/handlers/agency_handler.go`

**Locations**:
- **Line ~371**: Logs JSON binding errors
- **Line ~376**: Logs handler invocation with agency ID and work item count
- **Line ~380-383**: Logs each work item's deliverables data (simple and structured)
- **Line ~383**: Logs full structured deliverables for items that have them
- **Line ~389**: Logs errors from service layer
- **Line ~394**: Logs successful update completion

**What to look for in server logs**:
```
[MVP-054] UpdateWorkItems called for agency: UC-INFRA-001 with 5 work items
[MVP-054] WorkItem[0]: code=WI-001, title=Test Work Item, deliverables=0, deliverables_structured=3
[MVP-054] WorkItem[0] structured deliverables: [{id:folder-1 type:folder name:Frontend children:[...]}]
[MVP-054] Work items updated successfully, returning specification
```

---

### 4. Backend - specification_service.go (Business Logic Layer)

**File**: `/workspaces/CodeValdCortex/internal/agency/services/specification_service.go`

**Locations**:
- **Line ~62**: Logs service method invocation with context
- **Line ~68-71**: Logs each work item's structured deliverables count
- **Line ~75**: Logs repository errors
- **Line ~79**: Logs successful save completion with count

**What to look for in server logs**:
```
[MVP-054] UpdateWorkItems service called agency_id=UC-INFRA-001 work_item_count=5 updated_by=user
[MVP-054] Service layer - WorkItem[0]: code=WI-001, deliverables_structured=3 items
[MVP-054] UpdateWorkItems completed, saved 5 work items
```

---

## How to Use These Debug Prints

### Testing Deliverables Save Flow:

1. **Start the application** with logging enabled:
   ```bash
   make run
   ```

2. **Open browser console** (F12) and filter for `[MVP-054]`

3. **Create or edit a work item** with deliverables:
   - Add folders and files in the deliverable tree builder
   - Click "Save Work Item"

4. **Watch the console output** for the complete flow:
   ```
   Frontend (work-items.js):
   [MVP-054] saveWorkItemFromEditor: Starting deliverables extraction
   [MVP-054] Using tree builder mode, getDeliverablesStructuredData exists? true
   [MVP-054] Retrieved structured deliverables: [3 nodes]
   [MVP-054] Data object being sent to API: {...}
   
   Frontend (specification-api.js):
   [MVP-054] SpecificationAPI.addWorkItem called with: {...}
   [MVP-054] SpecificationAPI.updateWorkItems: Sending 1 work items to API
   [MVP-054] Full payload to /work-items endpoint: {...}
   [MVP-054] API response received: {...}
   ```

5. **Check server logs** for backend processing:
   ```bash
   # Filter logs for MVP-054
   tail -f codevaldcortex.log | grep "\[MVP-054\]"
   # OR if logs go to stdout
   # Output will appear in the terminal where you ran 'make run'
   ```

### Expected Output Flow:

**✅ Success Path**:
1. Frontend extracts deliverables from Alpine.js component
2. Frontend sends data to API with `deliverables_structured` field populated
3. Backend handler receives work items with structured deliverables
4. Service layer processes and persists to database
5. All logs show deliverables count > 0 where expected

**❌ Failure Indicators**:
- `deliverables_structured: null` in frontend logs → Alpine.js data extraction failed
- `deliverables_structured_count: 0` in API logs → Data not being sent
- `deliverables_structured=0` in backend logs → Data not reaching server
- Error logs at any layer → Check error messages for details

---

## Comparing with Goals Save Flow

**Goals are being saved successfully**, so we can compare:

### Goals Flow (Working):
```javascript
// work-items.js
const selectedGoals = Array.from(checkboxes).map(cb => cb.value);
const data = { goal_keys: selectedGoals, ... };

// Backend receives and saves
goal_keys: ['G-001', 'G-002']
```

### Deliverables Flow (Testing):
```javascript
// work-items.js
const deliverablesStructured = window.getDeliverablesStructuredData();
const data = { deliverables_structured: deliverablesStructured, ... };

// Backend should receive and save
deliverables_structured: [{id: 'folder-1', type: 'folder', name: 'Frontend', children: [...]}]
```

**Key Difference**: Goals use simple checkboxes (always reliable), deliverables use Alpine.js component (needs careful data extraction).

---

## Next Steps

1. **Run the application**: `make run`
2. **Open agency designer** in a browser
3. **Create/edit a work item** with deliverables
4. **Check browser console** for `[MVP-054]` logs
5. **Check server logs** for backend processing
6. **Compare the data** at each layer to find where deliverables might be dropped

## Cleanup

After debugging is complete, remove debug prints by searching for `[MVP-054]` and deleting those console.log/logger statements.

```bash
# Find all debug prints
grep -r "\[MVP-054\]" internal/ static/

# Or use the remove-debug-logs.sh script (if it supports task ID filtering)
./scripts/remove-debug-logs.sh MVP-054
```

---

## Files Modified

- `/workspaces/CodeValdCortex/static/js/agency-designer/work-items.js`
- `/workspaces/CodeValdCortex/static/js/agency-designer/specification-api.js`
- `/workspaces/CodeValdCortex/internal/handlers/agency_handler.go`
- `/workspaces/CodeValdCortex/internal/agency/services/specification_service.go`

Total debug print locations: **~20 strategic points** across the entire save flow.
