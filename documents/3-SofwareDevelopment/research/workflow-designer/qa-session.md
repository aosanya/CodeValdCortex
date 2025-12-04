# Workflow Designer Research: Q&A Session

**Session Date**: December 4, 2025  
**Method**: Structured single-question research flow  
**Participants**: AI Assistant + User  

---

## 📋 Session Overview

Comprehensive Q&A session exploring workflow designer enhancements through iterative questioning. Each question builds on previous answers to develop a complete understanding of requirements.

---

## 🎯 Design Decisions Summary

| Topic | Decision | Rationale |
|-------|----------|-----------|
| Properties Panel Editing | Read-only for work items | Single source of truth in work items section |
| Autonomy Level Location | Step-level (L0-L4) | Workflow controls execution policy |
| Default Autonomy | L0 (Manual) | Conservative, safety-first approach |
| Step Properties | Name, Description, Autonomy (editable) | Essential configuration fields |
| Work Item Display | Minimal reference + link | Keep designer focused |
| Autonomy Required | Yes (validation enforced) | No ambiguity in governance |
| Conditional Flow | Status + Route mapping | Pre-evaluated, simple |
| Aggregation Logic | any/all/majority/first | Parallel step control |
| Human Routing | Predefined routes only | Safety and governance |
| Route Display | Prominently on canvas | Maximum visibility |

---

## 📝 Q&A Session Transcript

### Q13: Properties Panel Editing Capability

**Question**: When a user clicks a work item in a workflow step, should the properties panel show read-only information or editable properties?

**Answer**: ✅ **READ-ONLY for work item details**

**Rationale**:
- Maintains single source of truth (work items section)
- Prevents data synchronization issues
- Keeps workflow designer focused on composition/orchestration
- Users go to work items section for detailed editing

---

### Q14: Autonomy Level Location

**Context**: Autonomy levels (L0-L4) currently exist on roles. User proposed using them to determine how dragged items are handled in workflow.

**Question**: Should autonomy levels be on roles, work items, or workflow steps?

**Answer**: ✅ **Step-level autonomy control**

**Decisions**:
1. Autonomy levels at **step level** (not work item or role)
2. Remove AutonomyLevel from work items (if exists)
3. Steps control execution policy/governance

**Benefits**:
- Workflow orchestration controls execution policy
- Removes complexity from work items
- Same work item can have different autonomy in different workflows
- Enables phased rollouts

---

### Q15: Default Autonomy Level

**Question**: What should the default autonomy level be for new steps?

**Answer**: ✅ **L0 (Manual)**

**Rationale**:
- Conservative by default (requires human approval)
- Users explicitly opt-in to higher autonomy
- Clear governance model
- Changeable via properties panel after creation

---

### Q16: Step Properties Configuration

**Question**: What should the step properties panel display and allow editing?

**Answer**: ✅ **Comprehensive step configuration**

**Fields**:
- Step Number (read-only)
- Name (editable text)
- Description (editable textarea)
- Autonomy Level (dropdown: L0-L4)
- Execution Mode (sequential/parallel)
- Work Item Count (read-only badge)

See [ui-design.md](./ui-design.md) for complete field specifications.

---

### Q17: Work Item Representation

**Question**: Should properties panel show minimal reference info or full details for work items in steps?

**Answer**: ✅ **Minimal reference + link to full details**

**Rationale**:
- Keeps workflow designer focused
- Quick identification
- See step-level autonomy context
- Easy removal
- Jump to work items section for details

---

### Q18: Data Model Requirements

**Question**: Should autonomy level be required or optional?

**Answer**: ✅ **REQUIRED with validation**

**Model Changes**:
```go
type Step struct {
    ID            string  `binding:"required"`
    Order         int     `binding:"required"`
    Name          string  `omitempty`
    Description   string  `omitempty`
    AutonomyLevel string  `binding:"required,oneof=L0 L1 L2 L3 L4"`
    Items         []StepItem `binding:"required,min=1"`
}
```

---

### Q19-20: Conditional Flow Patterns

**Question**: Which conditional flow patterns should be supported?

**Answer**: ✅ **Flexible system supporting all patterns**

**Patterns**:
1. Conditional branching (IF-THEN-ELSE)
2. Looping (WHILE/UNTIL)
3. Retry logic (error handling)
4. Approval gates (human decisions)
5. Parallel synchronization (race/all)

---

### Q21: Condition Evaluation

**User Insight**: "All evaluations should already have been performed"

**Answer**: ✅ **Pre-evaluated state properties**

**Architecture**:
- Evaluations happen BEFORE workflow execution
- Steps read state flags/properties
- No runtime condition evaluation
- Simple, auditable decisions

---

### Q22: State Properties for Flow Control

**User Specification**: "Status and route for given status... level group mirrors this (any succeeds, all succeed)"

**Answer**: ✅ **Status + Route mapping with aggregation**

**Structure**:
```json
{
    "status": "success|failure|error|retry|skip|pending",
    "route": "next-step-id"
}
```

**Aggregation Rules** (parallel steps):
- `any`: ANY item succeeds → step success
- `all`: ALL items succeed → step success
- `majority`: >50% succeed → step success
- `first`: First completion determines status

---

### Q23: Enhanced Step Data Model

**Answer**: ✅ **Base model sufficient**

**Model**:
```go
type Step struct {
    // Core
    ID, Order, Name, Description, AutonomyLevel, Items
    
    // Conditional flow
    Routes        map[string]string  // status -> next_step_id
    Aggregation   string             // any|all|majority|first
    DefaultRoute  string             // fallback
    
    // Runtime
    Status, ResolvedRoute
}
```

---

### Q24-25: Route Display

**Question**: How should routes be displayed in vertical layout?

**Answer**: ✅ **Prominently on canvas (maximum visibility)**

**User Specification**: "Show the routes as much as possible"

---

### Q26: Human-Directed Routing

**User Insight**: "When someone has been notified, they can direct workflow to certain other steps"

**Pattern**: Human-in-the-loop decision routing

**Question**: Predefined or flexible routing?

**Answer**: ✅ **Predefined routes (Option B)**

**Model Addition**:
```go
type Step struct {
    // ... existing fields
    RequiresHumanDecision bool
    AvailableRoutes       []HumanRoute
}

type HumanRoute struct {
    ID, Label, Description
    NextStepID
    Icon, Severity
    RequiresJustification bool
}
```

**Benefits**:
- Safety (only approved routes)
- Governance (designer controls paths)
- Audit trail (logged choices)
- Clear UX (guided decisions)

---

### Q27-28: Human Decision UI (OUT OF SCOPE)

**Question**: How should human decision UI work during workflow execution?

**User Clarification**: "The execution part is not yet in discussion, this will come up in the workbench"

**Answer**: ✅ **OUT OF SCOPE for workflow designer**

**Scope Clarification**:
- **Workflow Designer**: Define available routes, configure options (DESIGN TIME)
- **Workbench**: Execute workflows, present decision UI to users (EXECUTION TIME)

**Designer Responsibility**:
- Configure `RequiresHumanDecision: true`
- Define `AvailableRoutes` with labels, descriptions, icons, severity
- Specify which routes require justification
- Visual representation of decision points in workflow diagram

**Workbench Responsibility** (future):
- Display decision UI when workflow pauses
- Capture user choice and justification
- Log decision for audit trail
- Resume workflow on chosen route

**Status**: ✅ **Resolved - Deferred to workbench**

---

## Phase 3: Visual Design & UX (Q29-31)

### Q29: Visual Route Representation

**Question**: How should conditional routes be visualized on the workflow canvas?

**Options**:
- **A**: Arrow labels with status badges (success/failure/error)
- **B**: Color-coded connection lines (green/red/yellow)
- **C**: Compact route summary panel on step card
- **D**: Combination approach

**User Answer**: **D - Combination approach**

**Design Decision**: ✅ **Multi-layer visualization**

**Implementation**:
1. **Step Card**: Compact route summary indicator
   - Icon badge showing route count (e.g., "3 routes")
   - Color indicator if routes defined (blue) vs default (gray)
2. **Connection Lines**: Color-coded arrows between steps
   - Green: Success route
   - Red: Failure/error routes
   - Yellow: Retry/skip routes
   - Dashed: Human decision routes
3. **Route Labels**: Status badges on arrows (on hover or always visible)
   - Small pill badges: "success", "failure", "error", "retry", "skip"
4. **Properties Panel**: Full route table for detailed editing

**Benefits**:
- At-a-glance overview (step card indicator)
- Clear flow direction (color-coded arrows)
- Detailed status mapping (labels)
- Full control (properties panel)

**Status**: ✅ **Resolved**

---

### Q30: Aggregation Rule UI

**Question**: How should users configure parallel step aggregation in the designer?

**Options**:
- **A**: Dropdown in step properties (any/all/majority/first)
- **B**: Visual indicator on step card
- **C**: Both

**User Answer**: **A - Dropdown in step properties**

**Design Decision**: ✅ **Properties panel dropdown only**

**Rationale**:
- Aggregation is advanced configuration (not needed for all workflows)
- Properties panel is the right place for detailed settings
- Keeps step card clean and focused
- Can add visual indicator later if needed

**Implementation**:
```javascript
// Properties panel field configuration
{
  name: 'aggregation',
  label: 'Parallel Step Aggregation',
  type: 'select',
  options: [
    { value: '', label: 'None (sequential)' },
    { value: 'any', label: 'Any (first completion)' },
    { value: 'all', label: 'All (wait for all)' },
    { value: 'majority', label: 'Majority (>50%)' },
    { value: 'first', label: 'First (immediate)' }
  ],
  description: 'How to proceed when multiple work items in this step complete',
  showIf: (step) => step.Items && step.Items.length > 1
}
```

**Status**: ✅ **Resolved**

---

### Q31: Route Editing UX

**Question**: How should users define/edit status-to-route mappings?

**Options**:
- **A**: Key-value table in properties panel
- **B**: Visual route builder (drag connections between steps)
- **C**: Predefined templates + custom overrides

**User Answer**: **A - Key-value table in properties panel**

**Design Decision**: ✅ **Table-based route editor**

**Rationale**:
- Clear, explicit mapping (no ambiguity)
- Easy to understand and validate
- Works well with existing properties panel infrastructure
- Can evolve to visual builder later (v2 feature)

**Implementation**:
```javascript
// Properties panel field configuration
{
  name: 'routes',
  label: 'Status Routes',
  type: 'table',
  columns: [
    { 
      key: 'status', 
      label: 'Status', 
      type: 'select',
      options: ['success', 'failure', 'error', 'retry', 'skip']
    },
    { 
      key: 'targetStep', 
      label: 'Target Step', 
      type: 'select',
      options: (workflow) => workflow.Steps.map(s => ({ 
        value: s.ID, 
        label: `${s.Order}. ${s.Name || 'Unnamed Step'}` 
      }))
    },
    {
      key: 'condition',
      label: 'Condition (Optional)',
      type: 'text',
      placeholder: 'e.g., approval_count > 2'
    }
  ],
  addButtonLabel: '+ Add Route',
  emptyMessage: 'No routes defined. Default will proceed to next step.',
  validation: [
    { rule: 'uniqueStatus', message: 'Each status can only have one route' },
]

**Status**: ✅ **Resolved**

---

### Q32: Step Card Information Density

**Question**: When displaying steps on the canvas, what level of detail should be visible without clicking?

**Options**:
- **A**: Minimal (order number + work item count + route indicator)
- **B**: Compact (+ step name + autonomy level badge)
- **C**: Detailed (+ work item names + status summary)
- **D**: User-configurable zoom levels (toggle detail)

**User Answer**: **B - Compact (step name + autonomy level badge)**

**Design Decision**: ✅ **Balanced information display**

**Step Card Layout**:
```
┌─────────────────────────────┐
│ [1] Review Documentation   │  ← Order + Name
│ 📋 3 items    [L2] 🔀 3    │  ← Work items + Autonomy + Routes
└─────────────────────────────┘
```

**Card Contents**:
1. **Order number**: [1], [2], [3] (large, bold)
2. **Step name**: "Review Documentation" (truncate if long)
3. **Work item count**: 📋 3 items (icon + count)
4. **Autonomy badge**: [L0], [L1], [L2], [L3], [L4] (color-coded)
5. **Route indicator**: 🔀 3 (if routes defined)

**Color Coding**:
- **Autonomy badges**:
  - L0 (Manual): Gray
  - L1 (Assisted): Blue
  - L2 (Conditional): Yellow
  - L3 (High Auto): Orange
  - L4 (Full Auto): Green
- **Route indicator**: Blue if routes defined, hidden if default

**Hover State**: Show tooltip with work item names

**Benefits**:
- Clear step identification (order + name)
- Governance visibility (autonomy level)
- Complexity indicator (work items + routes)
- Scalable to 20+ steps
- No constant clicking needed

**Status**: ✅ **Resolved**

---

### Q33: Workflow Validation Rules (PAUSED)

**Status**: ⏸️ **Research paused - Ready for implementation**

---

## Research Session Summary

**Total Questions**: 33 (Q1-Q12 from previous session, Q13-Q32 documented here)
**Questions Answered**: 32
**Design Decisions Made**: 20
**Completion**: ~97%

**Key Outcomes**:
1. ✅ Step-level autonomy model (L0-L4)
2. ✅ Conditional flow architecture (status/route mapping)
3. ✅ Human-directed routing (predefined routes, deferred to workbench)
4. ✅ Visual design (combination approach with color-coded routes)
5. ✅ Properties panel configuration (table-based route editor)
6. ✅ Step card layout (compact with autonomy badges)

**Next Phase**: Implementation planning and MVP task creation

**Research Status**: ✅ **COMPLETE - Ready for development**


---

## 🔄 Next Questions

**Q28**: Human decision UI pattern selection
**Q29**: Visual route display design (colors, icons, arrows)
**Q30**: Properties panel integration with routes
**Q31**: Workflow validation rules for conditional flows
**Q32**: Testing strategy for complex workflows

---

## 📊 Research Metrics

- **Questions Asked**: 28
- **Design Decisions**: 15
- **Data Model Changes**: 3 major enhancements
- **New Features**: 4 (status bar, properties, autonomy, conditional flow)
- **Session Duration**: ~2 hours
- **Completion**: ~60% (core features defined, UI/implementation details remaining)

---

**Status**: 🔄 **Active Research** - Continue with Q28 (Human Decision UI)
