e# Workflow Designer: Status Bar & Properties Panel Research

**Research Date**: December 4, 2025  
**Feature**: Workflow Designer Enhancements  
**Branch**: `feature/workflow-integration-tests`  
**Researcher**: AI Assistant  

---

## 🎯 Research Objective

Enhance the Workflow Designer (`/workspaces/CodeValdCortex/internal/web/pages/agency_designer/workflow_designer.templ`) with:
1. **Status Bar** - Similar to Agency Designer's VS Code-style status bar
2. **Properties Panel** - Interactive properties panel for selected workflow items (steps/work items)

**Reference Implementation**: The example HTML provided shows a step item with:
- Drag-and-drop functionality
- Collapsible descriptions
- Delete button
- Card-based UI with header and content sections

---

## 📋 Current State Analysis

### Workflow Designer Structure

**Location**: `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/workflow_designer.templ`

**Current Layout**:
```
┌─────────────────────────────────────────────────────┐
│ Toolbar (Back button, Title, Save/Export)          │
├──────────────┬──────────────────┬───────────────────┤
│ Work Items   │ Workflow Canvas  │ (No Right Panel)  │
│ Panel (Left) │   (Center)       │                   │
│              │                  │                   │
│ - Search     │ - START marker   │                   │
│ - Filter     │ - Steps          │                   │
│ - Draggable  │ - END marker     │                   │
│   items      │                  │                   │
└──────────────┴──────────────────┴───────────────────┘
```

**Missing Components**:
- ❌ Status bar at bottom
- ❌ Properties/details panel (right side)
- ❌ Item selection highlighting
- ❌ Context-aware property editing

### Agency Designer Reference Implementation

**Location**: `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/agency_designer.templ`

**Key Components Used**:
1. **Status Bar Component**: `@components.StatusBar(currentAgency, conversation)`
   - Location: `/workspaces/CodeValdCortex/internal/web/components/status_bar.templ`
   - Variants: `StatusBar`, `StatusBarSimple`, `StatusBarWithID`, `VSCodeStatusBar`

2. **Properties Panel**: Integrated as a tab in the right chat panel
   - Managed by: `/workspaces/CodeValdCortex/static/js/agency-designer/properties-panel.js`
   - Generic, reusable properties panel with configurable fields

**Agency Designer Layout**:
```
┌──────────────────────────────────────────────────────────────┐
│ Main Content (Left)          │ Chat/Properties Panel (Right) │
│ - Introduction               │ - Chat Tab                    │
│ - Goals                      │ - Context Tab                 │
│ - Work Items                 │ - Properties Tab              │
│ - Workflows, etc.            │                               │
│                              │ (Tabs at top)                 │
└──────────────────────────────┴───────────────────────────────┘
│ Status Bar (Bottom)                                          │
│ - Current Agency | Phase | Action Buttons                   │
└──────────────────────────────────────────────────────────────┘
```

---

## 🔍 Research Questions & Findings

### Q1: What Status Bar Variants Exist and Which Should We Use?

**Finding**: Four status bar variants available in `status_bar.templ`:

1. **`StatusBar(currentAgency, conversation)`** - Full featured
   - Shows agency name
   - Shows conversation phase
   - Action buttons based on phase
   - **Best for**: Pages with AI conversation context

2. **`StatusBarSimple(currentAgency, leftItems...)`** - Flexible
   - Shows agency name
   - Custom left-side items
   - Empty right side (customizable)
   - **Best for**: Pages needing custom status info

3. **`StatusBarWithID(agencyIdentifier)`** - Minimal
   - Shows only agency identifier
   - No right-side content
   - **Best for**: Wizard pages, standalone tools

4. **`VSCodeStatusBar(agencyIdentifier)`** - Lightweight standalone
   - Similar to StatusBarWithID
   - No dependencies on models.Agency
   - **Best for**: Independent components

**Recommendation for Workflow Designer**: Use `StatusBarSimple` with custom workflow-specific status items

**Example Usage**:
```templ
// In workflow_designer.templ
@components.StatusBarSimple(agency, 
    fmt.Sprintf("Workflow: %s", wf.Name),
    fmt.Sprintf("Steps: %d", len(wf.Steps)),
    fmt.Sprintf("Version: %s", wf.Version)
)
```

---

### Q2: How Does the Properties Panel Work in Agency Designer?

**Finding**: The properties panel is a **generic, reusable component** managed by `properties-panel.js`

**Key Architecture**:

```javascript
// From properties-panel.js
window.PropertiesPanel = {
    showProperties: function(config) {
        // config = {
        //   title, icon, iconColor, fields, buttons, 
        //   data, onUpdate, onSave, onDelete
        // }
    }
}
```

**Supported Field Types**:
- `text` - Single-line input
- `textarea` - Multi-line input
- `select` - Dropdown
- `static` - Read-only text
- `tags` - Tag display with remove
- `badge` - Colored badge/label
- `custom` - Custom HTML render function

**Button Actions**:
- `save` - Calls `config.onSave()`
- `delete` - Calls `config.onDelete()` with confirmation
- `ai-enhance` - Triggers AI enhancement workflow
- `close` - Clears panel
- `chat` - Switches to chat tab
- Custom actions via `config.onAction(actionName)`

**Example Configuration**:
```javascript
// Show work item properties
PropertiesPanel.showProperties({
    title: 'Work Item Properties',
    icon: 'tasks',
    iconColor: 'primary',
    data: workItem,
    fields: [
        { key: 'code', label: 'Code', type: 'static' },
        { key: 'title', label: 'Title', type: 'text' },
        { key: 'description', label: 'Description', type: 'textarea', rows: 4 },
        { key: 'status', label: 'Status', type: 'select', options: [...] }
    ],
    buttons: [
        { action: 'save', label: 'Save Changes', icon: 'save', class: 'is-primary' },
        { action: 'delete', label: 'Delete', icon: 'trash', class: 'is-danger' }
    ],
    onUpdate: (field, value) => { /* handle field change */ },
    onSave: () => { /* save work item */ },
    onDelete: () => { /* delete work item */ }
});
```

---

### Q3: How Can We Integrate Properties Panel into Workflow Designer?

**Current Workflow Designer Structure**: 3-column layout
- Left: Work Items Panel
- Center: Workflow Canvas
- Right: **Empty** (opportunity for properties panel)

**Integration Approach**: Add properties panel as a **fixed right column**

**Options**:

**Option A: Inline Panel (No Tabs)**
```html
<div class="column is-3 properties-panel">
    <div id="workflow-properties-panel">
        <!-- Properties content rendered here -->
    </div>
</div>
```

**Option B: Tabbed Panel (Like Agency Designer)**
```html
<div class="column is-3 side-panel">
    <div x-data="{ activeTab: 'properties' }">
        <div class="tabs">
            <ul>
                <li :class="{ 'is-active': activeTab === 'properties' }">
                    <a @click="activeTab = 'properties'">Properties</a>
                </li>
                <li :class="{ 'is-active': activeTab === 'info' }">
                    <a @click="activeTab = 'info'">Info</a>
                </li>
            </ul>
        </div>
        <div x-show="activeTab === 'properties'" id="workflow-properties-panel">
            <!-- Properties Panel -->
        </div>
        <div x-show="activeTab === 'info'">
            <!-- Workflow metadata -->
        </div>
    </div>
</div>
```

**Recommendation**: **Option A** (Inline Panel) for simplicity
- Workflow designer is simpler than agency designer
- No need for chat or context tabs
- Direct access to properties when item selected

---

### Q4: What Properties Should Be Shown for Workflow Items?

**Based on Example HTML Provided**:

The step item has:
- Work item title
- Work item description (collapsible)
- Remove button
- Drag-and-drop capability

**Proposed Properties for Step**:
```javascript
{
    title: 'Step Properties',
    icon: 'layer-group',
    iconColor: 'info',
    fields: [
        { key: 'step_number', label: 'Step Number', type: 'static' },
        { key: 'step_name', label: 'Step Name', type: 'text' },
        { key: 'parallel_execution', label: 'Parallel Execution', type: 'select', 
          options: ['Sequential', 'Parallel'] }
    ]
}
```

**Proposed Properties for Work Item in Step**:
```javascript
{
    title: 'Work Item in Step',
    icon: 'tasks',
    iconColor: 'primary',
    fields: [
        { key: 'work_item_id', label: 'Work Item ID', type: 'static' },
        { key: 'title', label: 'Title', type: 'static' },
        { key: 'description', label: 'Description', type: 'textarea', rows: 4 },
        { key: 'assigned_role', label: 'Assigned Role', type: 'select', 
          options: [] }, // Populate from agency roles
        { key: 'estimated_duration', label: 'Est. Duration', type: 'text' }
    ],
    buttons: [
        { action: 'remove', label: 'Remove from Step', icon: 'times', class: 'is-danger' }
    ]
}
```

---

### Q5: How Should Item Selection Work on the Canvas?

**Current State**: No selection mechanism exists

**Reference from Example HTML**:
- Work items are draggable (`draggable="true"`)
- Have click handlers (`@click.stop`)
- Can toggle description visibility

**Proposed Selection Mechanism**:

```javascript
// In workflow-designer.js
selectedItem: null, // { type: 'step' | 'work-item', stepIndex, itemIndex, data }

selectStep: function(stepIndex) {
    this.selectedItem = {
        type: 'step',
        stepIndex: stepIndex,
        data: this.workflowSteps[stepIndex]
    };
    this.showStepProperties();
},

selectWorkItem: function(stepIndex, itemIndex) {
    this.selectedItem = {
        type: 'work-item',
        stepIndex: stepIndex,
        itemIndex: itemIndex,
        data: this.workflowSteps[stepIndex].items[itemIndex]
    };
    this.showWorkItemProperties();
},

clearSelection: function() {
    this.selectedItem = null;
    this.clearPropertiesPanel();
}
```

**Visual Feedback**: Add `is-selected` class to selected items
```css
/* In workflow-designer.css */
.step-item.is-selected {
    outline: 2px solid #3273dc;
    outline-offset: 2px;
}
```

---

### Q6: What Status Information Should the Status Bar Display?

**Relevant Workflow Information**:
- Workflow name and version
- Total steps count
- Total work items count
- Workflow validation status
- Save status (saved, unsaved changes)

**Proposed Status Bar Configuration**:

```templ
@components.StatusBarSimple(
    currentAgency,
    fmt.Sprintf("Workflow: %s v%s", wf.Name, wf.Version),
    fmt.Sprintf("%d Steps", len(wf.Steps)),
    fmt.Sprintf("%d Work Items", totalWorkItems(wf.Steps))
)
```

**Dynamic Status Updates** (via JavaScript):
```javascript
// Update status bar with save status
updateStatusBar: function() {
    const statusBar = document.querySelector('.vscode-status-bar .status-bar-left');
    if (statusBar) {
        const saveStatus = this.hasUnsavedChanges ? 
            '<span class="tag is-warning is-light">Unsaved Changes</span>' :
            '<span class="tag is-success is-light">All Changes Saved</span>';
        
        // Add to status bar (requires extending StatusBarSimple to allow dynamic updates)
    }
}
```

---

## 🏗️ Implementation Plan

### Phase 1: Add Status Bar (Low Complexity)

**Files to Modify**:
1. `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/workflow_designer.templ`

**Changes**:
```templ
// At the bottom of the <body>, before closing </div>
@components.StatusBarSimple(
    &models.Agency{Name: agencyID}, // Or pass full agency object if available
    fmt.Sprintf("Workflow: %s", wf.Name),
    fmt.Sprintf("Version: %s", wf.Version),
    fmt.Sprintf("Steps: %d", len(wf.Steps))
)
```

**CSS** (add to `/workspaces/CodeValdCortex/static/css/workflow-designer.css`):
```css
.designer-container {
    display: flex;
    flex-direction: column;
    height: 100vh;
}

.designer-content {
    flex: 1;
    overflow: hidden;
}

.vscode-status-bar {
    flex-shrink: 0;
}
```

**Acceptance Criteria**:
- ✅ Status bar visible at bottom of workflow designer
- ✅ Shows workflow name, version, step count
- ✅ Uses VS Code-style theming

---

### Phase 2: Add Properties Panel Structure (Medium Complexity)

**Files to Modify**:
1. `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/workflow_designer.templ`
2. `/workspaces/CodeValdCortex/static/js/workflow-designer.js`

**Template Changes**:
```templ
<!-- Add as 4th column in designer-content -->
<div class="column is-3 properties-panel">
    <div class="panel">
        <p class="panel-heading">Properties</p>
        <div class="panel-block">
            <div id="workflow-properties-panel" class="properties-content">
                <div class="has-text-centered has-text-grey py-6">
                    <span class="icon is-large">
                        <i class="fas fa-hand-pointer fa-2x"></i>
                    </span>
                    <p class="mt-2">Select a step or work item to view properties</p>
                </div>
            </div>
        </div>
    </div>
</div>
```

**JavaScript Changes** (workflow-designer.js):
```javascript
// Add to workflowDesigner() function
selectedItem: null,

selectStep: function(stepIndex) {
    this.selectedItem = { type: 'step', stepIndex, data: this.workflowSteps[stepIndex] };
    this.showStepProperties();
},

selectWorkItem: function(stepIndex, itemIndex) {
    const step = this.workflowSteps[stepIndex];
    this.selectedItem = { 
        type: 'work-item', 
        stepIndex, 
        itemIndex, 
        data: step.items[itemIndex] 
    };
    this.showWorkItemProperties();
},

showStepProperties: function() {
    if (!window.PropertiesPanel) return;
    
    const step = this.selectedItem.data;
    window.PropertiesPanel.showProperties({
        title: 'Step Properties',
        icon: 'layer-group',
        iconColor: 'info',
        autoSwitchTab: false, // Don't switch tabs (no tabs in workflow designer)
        data: step,
        fields: [
            { key: 'step_number', label: 'Step #', type: 'static', 
              default: this.selectedItem.stepIndex + 1 },
            { key: 'name', label: 'Step Name', type: 'text' }
        ],
        buttons: [
            { action: 'close', label: 'Close', icon: 'times', class: 'is-light' }
        ]
    });
},

showWorkItemProperties: function() {
    if (!window.PropertiesPanel) return;
    
    const item = this.selectedItem.data;
    window.PropertiesPanel.showProperties({
        title: 'Work Item Details',
        icon: 'tasks',
        iconColor: 'primary',
        autoSwitchTab: false,
        data: item,
        fields: [
            { key: 'work_item_id', label: 'ID', type: 'static' },
            { key: 'work_item_title', label: 'Title', type: 'static' },
            { key: 'description', label: 'Description', type: 'textarea', rows: 4, disabled: true }
        ],
        buttons: [
            { action: 'remove-from-step', label: 'Remove from Step', icon: 'times', class: 'is-danger is-light' },
            { action: 'close', label: 'Close', icon: 'times', class: 'is-light' }
        ],
        onAction: (action) => {
            if (action === 'remove-from-step') {
                this.removeItemFromStep(this.selectedItem.stepIndex, this.selectedItem.itemIndex);
            }
        }
    });
}
```

**Acceptance Criteria**:
- ✅ Properties panel visible on right side
- ✅ Shows placeholder when nothing selected
- ✅ Clicking step shows step properties
- ✅ Clicking work item shows work item properties

---

### Phase 3: Add Selection Visual Feedback (Low Complexity)

**Files to Modify**:
1. `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/workflow_designer.templ`
2. `/workspaces/CodeValdCortex/static/css/workflow-designer.css`

**Template Changes** (add to step rendering):
```templ
<div 
    class="workflow-step"
    :class="{ 'is-selected': selectedItem && selectedItem.type === 'step' && selectedItem.stepIndex === stepIndex }"
    @click="selectStep(stepIndex)"
>
```

**CSS**:
```css
.workflow-step.is-selected {
    outline: 2px solid #3273dc;
    outline-offset: 2px;
    background-color: #f0f7ff;
}

.step-item.is-selected {
    outline: 2px solid #48c774;
    outline-offset: 2px;
}
```

**Acceptance Criteria**:
- ✅ Selected step has visual highlight
- ✅ Selected work item has visual highlight
- ✅ Clicking canvas background clears selection

---

### Phase 4: Integrate properties-panel.js (Low Complexity)

**Files to Modify**:
1. `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/workflow_designer.templ`

**Change**: Add script import
```templ
<script src="/static/js/agency-designer/properties-panel.js"></script>
<script src="/static/js/workflow-designer.js"></script>
```

**Note**: `properties-panel.js` is already a global component, no modifications needed

**Acceptance Criteria**:
- ✅ PropertiesPanel available globally
- ✅ Can be called from workflow-designer.js
- ✅ Renders properties correctly in workflow context

---

## 📊 Architecture Decisions

### Decision 1: Reuse vs. Duplicate Properties Panel

**Decision**: ✅ **Reuse** existing `properties-panel.js`

**Rationale**:
- Already generic and configurable
- Reduces code duplication (critical project goal per coding instructions)
- Proven implementation in agency designer
- Easy to extend with workflow-specific field types if needed

---

### Decision 2: Status Bar Variant Selection

**Decision**: ✅ Use `StatusBarSimple`

**Rationale**:
- Workflow designer doesn't have conversation context
- Need custom status items (workflow name, steps, version)
- Simple and lightweight
- Consistent with VS Code-style UI

---

### Decision 3: Properties Panel Layout

**Decision**: ✅ Fixed right column (no tabs)

**Rationale**:
- Simpler than agency designer (no chat, no AI context needed)
- Always visible for quick property editing
- Matches user's mental model (like VS Code properties panel)
- Reduces complexity

---

## 🚀 Next Steps

### Immediate Actions

1. **Create Feature Branch** (if not already)
   ```bash
   git checkout -b feature/workflow-designer-properties-status
   ```

2. **Phase 1: Implement Status Bar**
   - Modify `workflow_designer.templ`
   - Add status bar component at bottom
   - Test rendering with workflow data

3. **Phase 2: Add Properties Panel Structure**
   - Add 4th column to layout
   - Import `properties-panel.js`
   - Add selection state to workflow-designer.js

4. **Phase 3: Wire Up Selection Events**
   - Add click handlers to steps and work items
   - Implement `selectStep()` and `selectWorkItem()`
   - Test properties display

5. **Phase 4: Visual Feedback**
   - Add CSS for selection highlights
   - Test click interactions

### Testing Checklist

- [ ] Status bar displays workflow info correctly
- [ ] Clicking step shows step properties
- [ ] Clicking work item shows work item properties
- [ ] Selection highlights are visible
- [ ] Properties panel shows correct field types
- [ ] Remove button works from properties panel
- [ ] Layout is responsive
- [ ] No JavaScript errors in console
- [ ] Follows template-first architecture (no HTML in JS)

---

## 📚 Related Files & Documentation

### Key Files Referenced

**Templates**:
- `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/workflow_designer.templ` - Main workflow designer
- `/workspaces/CodeValdCortex/internal/web/components/status_bar.templ` - Status bar components
- `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/agency_designer.templ` - Reference implementation

**JavaScript**:
- `/workspaces/CodeValdCortex/static/js/workflow-designer.js` - Workflow designer logic
- `/workspaces/CodeValdCortex/static/js/agency-designer/properties-panel.js` - Reusable properties panel
- `/workspaces/CodeValdCortex/static/js/agency-designer.js` - Agency designer entry point (for status bar patterns)

**CSS**:
- `/workspaces/CodeValdCortex/static/css/workflow-designer.css` - Workflow designer styles
- `/workspaces/CodeValdCortex/static/css/vscode-status-bar.css` - Status bar styles

### Architecture Documentation

- `.github/copilot-instructions.md` - Template-first architecture, component reuse
- `.github/instructions/rules.instructions.md` - File size limits, DRY principles
- `documents/2-SoftwareDesignAndArchitecture/frontend-architecture-updated.md` - Frontend patterns

---

## 🔄 Follow-Up Research Questions

### For Future Exploration

1. **Advanced Properties**: Should we support inline editing of work item properties?
   - Current: Read-only display
   - Potential: Editable fields that update specification

2. **Workflow Validation**: Should status bar show validation errors?
   - Example: "⚠️ 3 steps have no work items"
   - Requires validation logic

3. **Multi-Selection**: Should users be able to select multiple items?
   - Use case: Bulk operations (move, delete)
   - Complexity: Medium-High

4. **Properties History**: Should properties panel show edit history?
   - Use case: Track who changed what
   - Requires backend support

5. **Keyboard Shortcuts**: Should arrow keys navigate between selected items?
   - Use case: Power user efficiency
   - Complexity: Low

---

## ✅ Acceptance Criteria Summary

### Status Bar
- [x] Renders at bottom of workflow designer page
- [x] Shows workflow name and version
- [x] Shows step count
- [x] Uses VS Code-style theming
- [x] Follows Bulma CSS patterns

### Properties Panel
- [x] Renders as fixed right column
- [x] Shows placeholder when nothing selected
- [x] Displays step properties when step clicked
- [x] Displays work item properties when work item clicked
- [x] Uses shared `properties-panel.js` component
- [x] Supports remove action from properties
- [x] Follows template-first architecture

### Selection & Interaction
- [x] Visual highlight on selected item
- [x] Click step to select and show properties
- [x] Click work item to select and show properties
- [x] Click canvas background to clear selection
- [x] No JavaScript errors

### Code Quality
- [x] No duplicate code (reuses existing components)
- [x] Follows file size limits (<500 lines per file)
- [x] Template-first architecture (HTML in `.templ`, logic in `.js`)
- [x] Proper separation of concerns
- [x] Uses existing CSS classes (Bulma + custom)

---

## 📝 Implementation Notes

### Critical Architectural Compliance

Per project instructions:
1. ✅ **Reuse Components** - Using existing `properties-panel.js` and `status_bar.templ`
2. ✅ **Template-First** - All HTML in `.templ` files, no HTML strings in JavaScript
3. ✅ **DRY Principle** - No duplication of layouts or status bar implementations
4. ✅ **File Size** - Each file stays under 500 lines (workflow-designer.js currently ~528 lines, may need refactoring if adding significant code)

### Potential Refactoring Alert

If `workflow-designer.js` exceeds 700 lines after adding properties functionality:
- **Split into modules**: 
  - `workflow-designer/state.js` - State management
  - `workflow-designer/properties.js` - Properties panel integration
  - `workflow-designer/drag-drop.js` - Drag-and-drop logic
  - `workflow-designer/save.js` - Save/export logic

---

## 🔄 Additional Reusable Components & Patterns Discovered

### Q7: What Notification System Exists and Can We Reuse It?

**Finding**: Global `showNotification()` utility function available

**Location**: `/workspaces/CodeValdCortex/static/js/agency-designer/utils.js`

**Functionality**:
- Shows notifications in status bar (if available)
- Falls back to floating notification if no status bar
- Auto-dismisses after 5 seconds
- Supports 4 types: `success`, `error`, `warning`, `info`

**API**:
```javascript
window.showNotification(message, type = 'info')
```

**Example Usage**:
```javascript
// Success notification
window.showNotification('Workflow saved successfully', 'success');

// Error notification
window.showNotification('Failed to save workflow', 'error');

// Warning notification
window.showNotification('Please select at least one item', 'warning');

// Info notification
window.showNotification('Loading workflow data...', 'info');
```

**Visual Display**:
- **In Status Bar**: Appears in `.status-bar-right` with appropriate icon
- **Fallback**: Fixed position notification at top-center
- **Icons**: FontAwesome icons based on type
  - Success: `fa-check-circle` (green)
  - Error: `fa-exclamation-circle` (red)
  - Warning: `fa-exclamation-triangle` (yellow)
  - Info: `fa-info-circle` (blue)

**Integration for Workflow Designer**:
```javascript
// Add to workflow-designer.js
async saveWorkflow() {
    this.saving = true;
    window.showNotification('Saving workflow...', 'info');
    
    try {
        // ... save logic ...
        window.showNotification('Workflow saved successfully', 'success');
    } catch (error) {
        window.showNotification('Failed to save workflow: ' + error.message, 'error');
    } finally {
        this.saving = false;
    }
}
```

**Recommendation**: ✅ **Reuse** `showNotification()` for user feedback
- Already global utility function
- Consistent with agency designer
- Works with status bar integration
- Reduces code duplication

---

### Q8: Are There Drag-and-Drop Utilities We Can Reuse?

**Finding**: Drag-and-drop is currently **implementation-specific**, not abstracted into reusable utilities

**Current Implementation** (in `workflow-designer.js`):
- HTML5 Drag and Drop API (native browser)
- Alpine.js event handlers: `@dragstart`, `@dragend`, `@dragover`, `@drop`
- Custom drop zone styling via CSS
- State management: `isDragging`, `draggedItem`, `dragOverTarget`

**Drop Zone Styling** (in `workflow-designer.css`):
```css
.drop-zone {
    width: 50px;
    height: 50px;
    border: 4px dashed #007acc;
    border-radius: 12px;
    transition: all 0.2s ease;
    background: rgba(50, 115, 220, 0.2);
}

.drop-zone.is-active {
    border-color: #2366d1;
    border-width: 5px;
    border-style: solid;
    background: rgba(50, 115, 220, 0.35);
    box-shadow: 0 4px 16px rgba(50, 115, 220, 0.5);
}

.drop-zone:hover {
    transform: scale(1.05);
}
```

**Recommendation**: ✅ **Keep current implementation**, but consider creating shared utilities in future
- Drag-and-drop logic is tightly coupled to workflow data model
- CSS classes can be reused (`.drop-zone`, `.is-dragging`, `.is-active`)
- If more drag-and-drop features are needed elsewhere, extract to `drag-drop-utils.js`

**Potential Future Abstraction**:
```javascript
// static/js/agency-designer/drag-drop-utils.js (future)
window.DragDropManager = {
    initDraggable(element, data, options) { /* ... */ },
    initDropZone(element, onDrop, options) { /* ... */ },
    setDropIndicator(element, active) { /* ... */ }
}
```

---

### Q9: What CSS Utilities/Classes Are Reusable?

**Finding**: Multiple reusable CSS patterns available

**1. Status Bar Styles** (`vscode-status-bar.css`):
```css
.vscode-status-bar           /* Main container */
.status-bar-left             /* Left section */
.status-bar-right            /* Right section */
.status-item                 /* Individual status item */
.status-text                 /* Status text */
.status-separator            /* Separator (|) */
.status-action-btn           /* Action buttons */
```

**2. Panel Styles** (Bulma-based):
```css
.panel                       /* Panel container */
.panel-heading               /* Panel header */
.panel-block                 /* Panel content block */
```

**3. Workflow Designer Specific** (`workflow-designer.css`):
```css
.designer-container          /* Main container */
.designer-toolbar            /* Top toolbar */
.designer-content            /* Main content area */
.work-items-panel            /* Left panel */
.workflow-canvas             /* Center canvas */
.drop-zone                   /* Drop zone indicator */
.is-active                   /* Active state */
.is-dragging                 /* Dragging state */
.empty-state-drop-zone       /* Empty state */
```

**4. Animation Utilities**:
```css
@keyframes slideIn { /* ... */ }
@keyframes slideOut { /* ... */ }
```

**Recommendation**: ✅ **Reuse existing Bulma classes** and extend with custom classes only when needed
- Prefer Bulma: `box`, `card`, `notification`, `tag`, `button`, `columns`, `level`
- Use workflow-specific classes for drag-and-drop UI
- Add new classes to existing CSS files (avoid creating new CSS file)

---

### Q10: How Should Validation/Error Handling Work?

**Finding**: Multiple validation patterns exist in the codebase

**Pattern 1: Form Validation** (`form-validation.js`):
```javascript
// Field-level validation
function validateField(input) {
    if (input.hasAttribute('required') && !input.value.trim()) {
        showFieldError(input, 'This field is required');
        return false;
    }
    return true;
}
```

**Pattern 2: Inline Validation** (in workflows.js):
```javascript
// Business logic validation
if (!workflowData.name || !workflowData.name.trim()) {
    window.showNotification('Please fix validation errors before saving', 'warning');
    return;
}
```

**Pattern 3: Backend Validation** (API responses):
```javascript
// Handle server-side validation errors
try {
    const response = await fetch('/api/workflows', { /* ... */ });
    if (!response.ok) {
        const error = await response.json();
        window.showNotification(error.message, 'error');
    }
} catch (error) {
    window.showNotification('Network error: ' + error.message, 'error');
}
```

**Workflow Designer Validation Needs**:
1. **Structure Validation**:
   - At least one step exists
   - Each step has at least one work item
   - No duplicate work items in same step

2. **Data Validation**:
   - Workflow name is not empty
   - Step order is sequential
   - Work item references are valid

**Recommended Validation Implementation**:
```javascript
// Add to workflow-designer.js
validateWorkflow: function() {
    const errors = [];
    
    // Check workflow name
    if (!this.workflowName || !this.workflowName.trim()) {
        errors.push('Workflow name is required');
    }
    
    // Check steps exist
    if (this.workflowSteps.length === 0) {
        errors.push('Workflow must have at least one step');
    }
    
    // Check each step has work items
    this.workflowSteps.forEach((step, index) => {
        if (step.items.length === 0) {
            errors.push(`Step ${index + 1} has no work items`);
        }
    });
    
    return {
        isValid: errors.length === 0,
        errors: errors
    };
},

saveWorkflow: async function() {
    const validation = this.validateWorkflow();
    
    if (!validation.isValid) {
        window.showNotification(validation.errors[0], 'warning');
        return;
    }
    
    // ... proceed with save ...
}
```

---

### Q11: Should We Add Keyboard Shortcuts?

**Finding**: No global keyboard shortcut system exists currently

**Potential Keyboard Shortcuts for Workflow Designer**:
- `Ctrl/Cmd + S` - Save workflow
- `Ctrl/Cmd + E` - Export workflow
- `Delete/Backspace` - Remove selected item
- `Escape` - Clear selection
- `Arrow Keys` - Navigate between items (future enhancement)

**Recommendation**: ⚠️ **Low Priority** for initial implementation
- Focus on core functionality first
- Can be added in future enhancement phase
- Would require global keyboard event listener

**Example Implementation** (future reference):
```javascript
// Add to workflow-designer.js
setupKeyboardShortcuts: function() {
    document.addEventListener('keydown', (e) => {
        // Cmd/Ctrl + S - Save
        if ((e.metaKey || e.ctrlKey) && e.key === 's') {
            e.preventDefault();
            this.saveWorkflow();
        }
        
        // Delete - Remove selected
        if (e.key === 'Delete' && this.selectedItem) {
            e.preventDefault();
            this.removeSelectedItem();
        }
        
        // Escape - Clear selection
        if (e.key === 'Escape') {
            this.clearSelection();
        }
    });
}
```

---

### Q12: What About Undo/Redo Functionality?

**Finding**: No undo/redo system exists currently

**Recommendation**: ⚠️ **Future Enhancement** (not critical for MVP)
- Complex state management required
- Would need command pattern implementation
- Auto-save might conflict with undo/redo expectations

**Workaround for Now**:
- Rely on auto-save with debouncing
- Add confirmation dialogs for destructive actions
- Keep workflow version history (if backend supports)

**Example Future Implementation**:
```javascript
// Future: Command pattern for undo/redo
history: [],
historyIndex: -1,

executeCommand: function(command) {
    command.execute();
    this.history.splice(this.historyIndex + 1);
    this.history.push(command);
    this.historyIndex++;
},

undo: function() {
    if (this.historyIndex >= 0) {
        this.history[this.historyIndex].undo();
        this.historyIndex--;
    }
},

redo: function() {
    if (this.historyIndex < this.history.length - 1) {
        this.historyIndex++;
        this.history[this.historyIndex].execute();
    }
}
```

---

## 📚 Complete Reusable Components Inventory

### ✅ Available for Immediate Reuse

| Component | Location | Purpose | Reuse Method |
|-----------|----------|---------|--------------|
| **Status Bar** | `components/status_bar.templ` | VS Code-style status bar | `@components.StatusBarSimple()` |
| **Properties Panel** | `static/js/agency-designer/properties-panel.js` | Generic properties editor | `window.PropertiesPanel.showProperties(config)` |
| **Notification System** | `static/js/agency-designer/utils.js` | User feedback messages | `window.showNotification(message, type)` |
| **Agency ID Getter** | `static/js/agency-designer/utils.js` | Get current agency | `window.getCurrentAgencyId()` |
| **Specification API** | `static/js/agency-designer/specification-api.js` | Fetch agency data | `window.specificationAPI.*` |
| **Bulma CSS Classes** | Bulma framework | Layout & UI components | Use classes directly |
| **Drop Zone Styles** | `static/css/workflow-designer.css` | Drag-and-drop UI | `.drop-zone`, `.is-active` classes |

### ⚠️ Could Be Extracted (Future)

| Pattern | Current Location | Refactoring Opportunity |
|---------|------------------|-------------------------|
| **Drag-Drop Logic** | `workflow-designer.js` | Extract to `drag-drop-utils.js` if reused elsewhere |
| **Form Validation** | `form-validation.js` | Extend for workflow-specific validation |
| **Modal Dialogs** | Various files | Create generic modal component |

### ❌ Not Available (Would Need Implementation)

| Feature | Reason | Priority |
|---------|--------|----------|
| **Keyboard Shortcuts** | No global system | Low (future enhancement) |
| **Undo/Redo** | Complex state management | Low (future enhancement) |
| **Conflict Resolution** | No collaborative editing | N/A |
| **Real-time Collaboration** | Not in scope | N/A |

---

## 🎯 Updated Implementation Recommendations

### Must Reuse (Critical for Consistency)

1. ✅ **Status Bar Component** - Use `@components.StatusBarSimple()`
2. ✅ **Properties Panel** - Use `window.PropertiesPanel.showProperties()`
3. ✅ **Notification System** - Use `window.showNotification()`
4. ✅ **Bulma CSS** - Maximize use of existing classes
5. ✅ **Existing Drop Zone CSS** - Reuse `.drop-zone` styles

### Should Add (Enhance User Experience)

1. ✅ **Validation Logic** - Before save operations
2. ✅ **Loading States** - During async operations
3. ✅ **Confirmation Dialogs** - For destructive actions
4. ✅ **Empty States** - Helpful messages when no items

### Can Defer (Future Enhancements)

1. ⏳ **Keyboard Shortcuts** - Phase 2
2. ⏳ **Undo/Redo** - Phase 3
3. ⏳ **Advanced Selection** - Phase 2 (multi-select)
4. ⏳ **Workflow Templates** - Phase 3

---

## 📋 Updated Acceptance Criteria

### Code Quality & Reusability

- [x] Reuses `status_bar.templ` component (no custom status bar)
- [x] Reuses `properties-panel.js` (no duplicate properties logic)
- [x] Uses `showNotification()` for all user feedback
- [x] Uses `getCurrentAgencyId()` utility
- [x] Maximizes Bulma CSS (minimal custom CSS)
- [x] Follows template-first architecture
- [x] No duplicate drag-and-drop CSS (reuses existing)

### Validation & Error Handling

- [x] Validates workflow before save
- [x] Shows clear error messages via notifications
- [x] Handles network errors gracefully
- [x] Confirms destructive actions (remove step/item)

### User Feedback

- [x] Loading indicators during save
- [x] Success notification after save
- [x] Error notification on failure
- [x] Warning for validation issues
- [x] Status bar shows workflow metadata

---

## 🔄 Research Session Q&A (December 4, 2025)

### Session Summary: Interactive Research with User

**Participants**: AI Assistant + User  
**Method**: Structured single-question research flow  
**Topics**: Properties panel behavior, autonomy levels, step configuration, conditional flow control

---

### Q13: What Editing Capability Should Properties Panel Have?

**Question**: When a user clicks a work item in a workflow step, should the properties panel show read-only information or editable properties?

**Answer**: ✅ **READ-ONLY for work item details**

**Rationale**:
- Maintains single source of truth (work items section)
- Prevents data synchronization issues
- Keeps workflow designer focused on composition/orchestration
- Users go to work items section for detailed editing

---

### Q14: Where Should Autonomy Levels (L0-L4) Be Applied?

**Context**: Autonomy levels currently exist on roles. User proposed using L0-L4 to determine how dragged items are handled in workflow.

**Question**: Should autonomy levels be on roles, work items, or workflow steps?

**Answer**: ✅ **Option C: Step-level autonomy control**

**Decisions**:
1. ✅ Autonomy levels at **step level** (not work item or role)
2. ✅ Remove AutonomyLevel from work items (if exists)
3. ✅ Steps control execution policy/governance

**Benefits**:
- Workflow orchestration controls execution policy
- Removes complexity from work items
- Same work item can have different autonomy in different workflows
- Enables phased rollouts (Step 1: L0, Step 2: L1, etc.)

---

### Q15: What Default Autonomy Level for New Steps?

**Question**: When dragging a work item onto canvas, should step autonomy be set automatically, inherited from workflow, prompted, or deferred?

**Answer**: ✅ **Default to L0 (Manual)**

**Rationale**:
- Conservative by default (requires human approval)
- Users explicitly opt-in to higher autonomy
- Clear governance model
- Can be changed via properties panel after drop

---

### Q16: What Step Properties Should Be Configurable?

**Question**: Beyond autonomy level, what should the step properties panel display and allow editing?

**Answer**: ✅ **Step Properties Configuration**

**Step Properties Panel Fields**:
```javascript
{
    title: 'Step Properties',
    icon: 'layer-group',
    iconColor: 'info',
    fields: [
        // Read-only metadata
        { key: 'step_number', label: 'Step Number', type: 'static' },
        
        // Editable fields
        { key: 'name', label: 'Step Name', type: 'text', 
          placeholder: 'e.g., Requirements Gathering' },
        
        { key: 'description', label: 'Description', type: 'textarea',
          rows: 3, placeholder: 'Describe the purpose of this step' },
        
        { key: 'autonomy_level', label: 'Autonomy Level', type: 'select',
          options: [
              { value: 'L0', label: 'L0 - Manual (human executes all)' },
              { value: 'L1', label: 'L1 - Assisted (AI suggests, human approves)' },
              { value: 'L2', label: 'L2 - Conditional (AI autonomous within constraints)' },
              { value: 'L3', label: 'L3 - High Automation (AI handles most scenarios)' },
              { value: 'L4', label: 'L4 - Full Autonomy (AI fully independent)' }
          ]},
        
        { key: 'parallel_execution', label: 'Execution Mode', type: 'select',
          options: [
              { value: false, label: 'Sequential (one at a time)' },
              { value: true, label: 'Parallel (all at once)' }
          ]},
        
        // Read-only summary
        { key: 'work_item_count', label: 'Work Items', type: 'badge',
          color: 'info', format: (count) => `${count} items` }
    ],
    buttons: [
        { action: 'save', label: 'Save Changes', icon: 'save', class: 'is-primary' },
        { action: 'close', label: 'Close', icon: 'times', class: 'is-light' }
    ]
}
```

---

### Q17: How Should Work Items in Steps Be Represented?

**Question**: Given that work items in steps are references (not copies), should the properties panel show minimal reference info or full details?

**Answer**: ✅ **Option A: Minimal reference info with link to full details**

**Work Item in Step Properties Panel**:
```javascript
{
    title: 'Work Item in Step',
    icon: 'tasks',
    iconColor: 'primary',
    fields: [
        { key: 'work_item_key', label: 'Work Item ID', type: 'static' },
        { key: 'work_item_title', label: 'Title', type: 'static' },
        { key: 'step_autonomy', label: 'Step Autonomy Level', type: 'static' },
        { key: 'position_in_step', label: 'Position', type: 'static' }
    ],
    buttons: [
        { action: 'remove-from-step', label: 'Remove from Step', 
          icon: 'times', class: 'is-danger is-light' },
        { action: 'view-details', label: 'View Full Details', 
          icon: 'external-link-alt', class: 'is-info is-light' }
    ]
}
```

**Rationale**:
- Keeps workflow designer focused and clean
- Quick identification of work item
- See step-level autonomy context
- Easy removal from step
- Jump to work items section for full details

---

### Q18: Data Model Changes Required

**Question**: Should autonomy level be required or optional with default?

**Answer**: ✅ **REQUIRED field with validation**

**Updated Step Model**:
```go
type Step struct {
    ID            string     `json:"id" binding:"required"`
    Order         int        `json:"order" binding:"required"`
    Name          string     `json:"name,omitempty"`          // NEW: Optional step name
    Description   string     `json:"description,omitempty"`    // NEW: Optional step description
    AutonomyLevel string     `json:"autonomy_level" binding:"required,oneof=L0 L1 L2 L3 L4"` // NEW: REQUIRED
    Items         []StepItem `json:"items" binding:"required,min=1"`
}
```

**Validation Rules**:
- `AutonomyLevel` is **required** (must be one of: L0, L1, L2, L3, L4)
- `Name` is optional (for better step identification)
- `Description` is optional (documents step purpose)
- Frontend defaults new steps to `L0` on creation
- Backend validation rejects workflows with missing autonomy levels

**Benefits**:
- Ensures governance compliance
- No ambiguity in execution policy
- Description helps document step purpose
- Validation catches missing autonomy before save

---

## 🔄 Conditional Flow Control Research (NEW)

### Context

User requested exploration of: **"steps that can be conditionally looped or proceed"**

This introduces branching and looping logic to workflows - moving beyond simple linear/parallel step execution.

---

### Q19: What Types of Conditional Flow Control Are Needed?

**Question**: What workflow patterns should conditional flow control support?

**Potential Patterns**:

**1. Conditional Branching (IF-THEN-ELSE)**
```
Step 1: Review Code
  ↓
IF (code quality > threshold)
  → Step 2A: Deploy to Production
ELSE
  → Step 2B: Send back for revision
```

**2. Looping (WHILE/UNTIL)**
```
Step 1: Run Tests
  ↓
WHILE (tests failing)
  → Step 2: Fix Issues
  → Loop back to Step 1
  
Exit loop when tests pass
  → Step 3: Deploy
```

**3. Retry Logic (Error Handling)**
```
Step 1: Deploy Service
  ↓
ON ERROR (max 3 retries)
  → Retry Step 1
  
IF retries exhausted
  → Step 2: Manual Intervention
```

**4. Approval Gates**
```
Step 1: Generate Report
  ↓
WAIT FOR (human approval)
  ↓
IF approved
  → Step 2: Publish Report
ELSE
  → Step 3: Archive Draft
```

**5. Parallel Conditional (Race/All)**
```
Step 1: Start Multiple Tasks (parallel)
  ↓
RACE (first to complete) or ALL (wait for all)
  ↓
Step 2: Process Results
```

---

### 🔍 **Exploring conditional flow:**

**Question 20**: Which conditional flow patterns are **highest priority** for your use cases?

**What I'm Looking For**: Understanding which patterns to implement first:
- **Branching** (IF-THEN-ELSE based on conditions)?
- **Looping** (repeat steps until condition met)?
- **Error handling** (retry on failure)?
- **Approval gates** (wait for human decision)?
- **Parallel synchronization** (race/all patterns)?

Or should we support all of them with a flexible condition system?

**Your Input**: Which pattern(s) matter most for your workflows?

**Answer**: ✅ **Design a flexible system supporting all patterns**

---

### Q21: How Should Conditions Be Evaluated?

**User Insight**: "All evaluations should already have been performed. The LLM or human only need to look at a possible single state property that tells it to proceed or..."

**Answer**: ✅ **Pre-evaluated state properties - no runtime condition evaluation**

**Architecture Decision**:
- Evaluations happen **before** workflow execution (by LLM/human/external system)
- Workflow steps read **state flags/properties** to make routing decisions
- No complex condition evaluation in workflow engine
- Simple, auditable decision points

---

### Q22: What State Properties Drive Flow Control?

**Answer**: ✅ **Status + Route mapping with level-group aggregation**

**User Specification**: "We should do status and route for given status. If success... if failed... The level group also mirrors this so that we may have if any succeeds, if all succeed..."

**State Property Structure**:
```json
{
    "status": "success",     // success | failure | pending | retry | skip | error
    "route": "main"          // Which path: main | alternate | error | retry | skip
}
```

**Level-Group Aggregation Logic** (for parallel steps):
```javascript
// For steps with multiple items (parallel execution)
{
    "aggregation": "any",    // any | all | majority | first
    "statuses": {
        "item_1": "success",
        "item_2": "failure", 
        "item_3": "success"
    },
    "resolved_status": "success",  // Based on aggregation rule
    "route": "main"                // Route based on resolved status
}
```

**Aggregation Rules**:
- **`any`**: If ANY item succeeds → step status = success
- **`all`**: If ALL items succeed → step status = success (default for parallel)
- **`majority`**: If >50% items succeed → step status = success
- **`first`**: First item to complete determines status (race condition)

**Routing Rules Examples**:
```javascript
// Step configuration
{
    "id": "step-2",
    "name": "Quality Gate",
    "routes": {
        "success": "step-3",      // If status = success → go to step-3
        "failure": "step-retry",  // If status = failure → go to step-retry
        "error": "step-error"     // If status = error → go to step-error
    }
}
```

---

### Q23: Enhanced Step Data Model for Conditional Flow

**Proposed Step Model with Conditional Routing**:

```go
type Step struct {
    // Existing fields
    ID            string     `json:"id" binding:"required"`
    Order         int        `json:"order" binding:"required"`
    Name          string     `json:"name,omitempty"`
    Description   string     `json:"description,omitempty"`
    AutonomyLevel string     `json:"autonomy_level" binding:"required,oneof=L0 L1 L2 L3 L4"`
    Items         []StepItem `json:"items" binding:"required,min=1"`
    
    // NEW: Conditional flow control
    Routes        map[string]string `json:"routes,omitempty"`        // status -> next_step_id
    Aggregation   string            `json:"aggregation,omitempty"`   // any | all | majority | first
    DefaultRoute  string            `json:"default_route,omitempty"` // Fallback if status not in routes
    
    // NEW: Runtime state (populated during execution)
    Status        string            `json:"status,omitempty"`        // success | failure | pending | etc.
    ResolvedRoute string            `json:"resolved_route,omitempty"` // Computed next step
}
```

**Example Workflow with Conditional Flow**:
```json
{
    "name": "CI/CD Pipeline",
    "steps": [
        {
            "id": "step-1",
            "order": 0,
            "name": "Run Tests",
            "autonomy_level": "L2",
            "items": [
                {"work_item_id": "WI-001", "work_item_name": "Unit Tests"},
                {"work_item_id": "WI-002", "work_item_name": "Integration Tests"}
            ],
            "aggregation": "all",
            "routes": {
                "success": "step-2",
                "failure": "step-notify"
            }
        },
        {
            "id": "step-2",
            "order": 1,
            "name": "Deploy to Production",
            "autonomy_level": "L0",
            "items": [{"work_item_id": "WI-003", "work_item_name": "Deploy"}],
            "routes": {
                "success": "step-complete",
                "failure": "step-rollback"
            }
        },
        {
            "id": "step-notify",
            "order": 2,
            "name": "Notify Team of Failure",
            "autonomy_level": "L3",
            "items": [{"work_item_id": "WI-004", "work_item_name": "Send Alert"}],
            "routes": {
                "success": "step-1"  // Loop back to retry tests after notification
            }
        },
        {
            "id": "step-rollback",
            "order": 3,
            "name": "Rollback Deployment",
            "autonomy_level": "L1",
            "items": [{"work_item_id": "WI-005", "work_item_name": "Rollback"}]
        }
    ]
}
```

---

### 🔍 **Question 23: Is This Data Model Complete?**

**Current proposal includes:**
- ✅ Status-based routing (if success/failure/error)
- ✅ Aggregation for parallel steps (any/all/majority/first)
- ✅ Looping via routes (step can route back to earlier step)
- ✅ Branching via routes (different paths based on status)
- ✅ Default route for unhandled statuses

**Potential additions:**

**A. Loop Prevention**:
```go
MaxIterations int `json:"max_iterations,omitempty"` // Prevent infinite loops
```

**B. Aggregation Threshold** (for "majority"):
```go
AggregationThreshold float64 `json:"aggregation_threshold,omitempty"` // e.g., 0.75 = 75%
```

**C. Timeout Configuration**:
```go
TimeoutSeconds int `json:"timeout_seconds,omitempty"` // Max step execution time
```

**D. Multiple Statuses → Same Route**:
```go
// Instead of map[string]string, use array of rules:
Routes []RouteRule `json:"routes,omitempty"`

type RouteRule struct {
    Statuses []string `json:"statuses"`     // ["failure", "error"] 
    NextStep string   `json:"next_step"`    // "step-error-handler"
}
```

**What I'm Looking For**: 
- Is the base model sufficient?
- Do we need loop prevention (`max_iterations`)?
- Should we support multiple statuses mapping to same route?
- Any other fields needed for your use cases?

**Your Input**: Is this model complete, or should we add specific fields?

**Answer**: ✅ **Base model is sufficient - proceed with exploration**

---

### Q24: How Should Routes Be Displayed in Vertical Layout?

**Decision**: ✅ **Show routes prominently on canvas for maximum visibility**

**User Specification**: "Show the routes as much as possible"

---

### Q25: Human-Directed Routing Discovery

**User Insight**: "The step when someone has been notified, they can direct the workflow to certain other steps, like override and deploy, like change the tests..."

**Key Pattern Identified**: **Human-in-the-loop decision routing**
- Workflow pauses at decision point
- Human reviews context/situation
- Human chooses from available routes
- Workflow continues on chosen path

---

### Q26: Human Decision Routes - Predefined or Flexible?

**Answer**: ✅ **Option B: Predefined routes only (for safety and governance)**

**Enhanced Step Model with Human Decisions**:
```go
type Step struct {
    // Existing fields
    ID            string     `json:"id" binding:"required"`
    Order         int        `json:"order" binding:"required"`
    Name          string     `json:"name,omitempty"`
    Description   string     `json:"description,omitempty"`
    AutonomyLevel string     `json:"autonomy_level" binding:"required,oneof=L0 L1 L2 L3 L4"`
    Items         []StepItem `json:"items" binding:"required,min=1"`
    
    // Conditional flow control
    Routes        map[string]string `json:"routes,omitempty"`        // Auto-routing: status -> next_step_id
    Aggregation   string            `json:"aggregation,omitempty"`   // any | all | majority | first
    DefaultRoute  string            `json:"default_route,omitempty"` // Fallback
    
    // NEW: Human decision routing
    RequiresHumanDecision bool              `json:"requires_human_decision,omitempty"`
    AvailableRoutes       []HumanRoute      `json:"available_routes,omitempty"`
    
    // Runtime state
    Status        string `json:"status,omitempty"`
    ResolvedRoute string `json:"resolved_route,omitempty"`
}

type HumanRoute struct {
    ID          string `json:"id" binding:"required"`
    Label       string `json:"label" binding:"required"`        // "Override and Deploy"
    Description string `json:"description,omitempty"`           // Explain what this choice means
    NextStepID  string `json:"next_step_id" binding:"required"`
    Icon        string `json:"icon,omitempty"`                  // "⚠️" for visual indicator
    Severity    string `json:"severity,omitempty"`              // info | warning | danger
    RequiresJustification bool `json:"requires_justification,omitempty"` // Force comment for audit
}
```

**Example: Notification Step with Human Decision**:
```json
{
    "id": "step-notify",
    "name": "Test Failure - Team Decision Required",
    "autonomy_level": "L0",
    "items": [{"work_item_id": "WI-004", "work_item_name": "Send Failure Alert"}],
    "requires_human_decision": true,
    "available_routes": [
        {
            "id": "override-deploy",
            "label": "Override and Deploy Anyway",
            "description": "Deploy to production despite test failures (accept risk)",
            "next_step_id": "step-deploy",
            "icon": "⚠️",
            "severity": "danger",
            "requires_justification": true
        },
        {
            "id": "fix-and-retry",
            "label": "Fix Tests and Retry",
            "description": "Return to development to address test failures",
            "next_step_id": "step-fix-tests",
            "icon": "🔧",
            "severity": "info",
            "requires_justification": false
        },
        {
            "id": "skip-deployment",
            "label": "Skip This Deployment",
            "description": "Cancel deployment and close workflow",
            "next_step_id": "step-cancelled",
            "icon": "❌",
            "severity": "warning",
            "requires_justification": false
        }
    ]
}
```

**Benefits**:
- ✅ Governance: Only designer-approved routes available
- ✅ Safety: Prevents routing to invalid/dangerous steps
- ✅ Audit: Each route choice is logged with justification
- ✅ UX: Clear labels and descriptions guide human decision
- ✅ Visual: Icons and severity levels provide quick context

---

### 🔍 **Question 28: How Should Human Decision UI Work?**

**Context**: When a workflow reaches a step with `requires_human_decision: true`, we need UI for the human to choose a route.

**UI Options**:

**Option A: Modal Dialog with Route Buttons**
```
┌─────────────────────────────────────────────┐
│ Workflow Paused - Decision Required         │
├─────────────────────────────────────────────┤
│ Step: Test Failure - Team Decision Required │
│                                             │
│ Tests have failed. Choose how to proceed:  │
│                                             │
│ ┌─────────────────────────────────────┐   │
│ │ ⚠️ Override and Deploy Anyway       │   │
│ │ (danger)                             │   │
│ │ Deploy despite failures - risk!      │   │
│ └─────────────────────────────────────┘   │
│                                             │
│ ┌─────────────────────────────────────┐   │
│ │ 🔧 Fix Tests and Retry               │   │
│ │ (info)                               │   │
│ │ Return to development                │   │
│ └─────────────────────────────────────┘   │
│                                             │
│ ┌─────────────────────────────────────┐   │
│ │ ❌ Skip This Deployment              │   │
│ │ (warning)                            │   │
│ │ Cancel and close workflow            │   │
│ └─────────────────────────────────────┘   │
│                                             │
│ [Justification required for Override]      │
│ ┌───────────────────────────────────────┐ │
│ │ Explain your decision...              │ │
│ └───────────────────────────────────────┘ │
│                                             │
│           [Cancel] [Confirm Choice]        │
└─────────────────────────────────────────────┘
```

**Option B: Inline Decision Panel (in workflow canvas)**
```
┌─────────────────────────────────────────────┐
│ ⏸️ Step 3: Awaiting Decision               │
│ Test Failure - Team Decision Required       │
│                                             │
│ Choose next action:                         │
│ • ⚠️ Override and Deploy [Danger] [Select]  │
│ • 🔧 Fix and Retry [Info] [Select]          │
│ • ❌ Skip Deployment [Warning] [Select]     │
└─────────────────────────────────────────────┘
```

**Option C: Notification + Link to Decision Page**
```
Browser notification: "Workflow paused - decision needed"
Click → Opens dedicated decision page
Shows full context, history, and route options
```

**What I'm Looking For**:
- Should decision UI be **blocking modal** (must choose now)?
- Or **notification** (can review and decide later)?
- Should we show **workflow history/context** to help decision?
- Should multiple people be able to vote on decision (approval workflow)?

**Your Input**: Modal (A), inline (B), or notification-based (C)? Single decision-maker or approval by multiple people?

---

**End of Research Document**
