# Workflow Designer Enhancements - Implementation Plan

**Research Reference**: `/workspaces/CodeValdCortex/documents/3-SofwareDevelopment/research/workflow-designer/`

**Status**: Ready for Implementation  
**Priority**: High  
**Dependencies**: Existing workflow designer, properties panel, status bar components

---

## Overview

Enhance the workflow designer with:
1. Step-level autonomy controls (L0-L4)
2. Conditional flow routing (status-based branching)
3. Visual route indicators (color-coded arrows)
4. Properties panel integration for step configuration
5. Compact step cards with autonomy badges

---

## Data Model Changes

### 1. Update Step Model (`internal/agency/models/workflow.go`)

**Current Model**:
```go
type Step struct {
    ID    string     `json:"id"`
    Order int        `json:"order"`
    Items []StepItem `json:"items"`
}
```

**Enhanced Model**:
```go
type Step struct {
    ID                    string            `json:"id"`
    Order                 int               `json:"order"`
    Name                  string            `json:"name,omitempty"`
    Description           string            `json:"description,omitempty"`
    AutonomyLevel         string            `json:"autonomy_level" validate:"required,oneof=L0 L1 L2 L3 L4"`
    Items                 []StepItem        `json:"items"`
    Routes                map[string]string `json:"routes,omitempty"`
    Aggregation           string            `json:"aggregation,omitempty" validate:"omitempty,oneof=any all majority first"`
    DefaultRoute          string            `json:"default_route,omitempty"`
    RequiresHumanDecision bool              `json:"requires_human_decision,omitempty"`
    AvailableRoutes       []HumanRoute      `json:"available_routes,omitempty"`
}

type HumanRoute struct {
    ID          string `json:"id"`
    Label       string `json:"label"`
    Description string `json:"description,omitempty"`
    TargetStep  string `json:"target_step"`
    Icon        string `json:"icon,omitempty"`
    Severity    string `json:"severity,omitempty"` // info, warning, danger
    RequiresJustification bool `json:"requires_justification,omitempty"`
}
```

**Implementation Tasks**:
- [ ] Add new fields to Step struct with validation tags
- [ ] Create HumanRoute struct
- [ ] Update database migration scripts
- [ ] Add default AutonomyLevel = "L0" for existing workflows
- [ ] Update API endpoints to accept new fields

---

## Frontend Components

### 2. Update Step Card Display (`static/js/workflow-designer.js`)

**Current**: Basic step card with order and work item count

**Enhanced**: Compact card with autonomy badge and route indicator

**Layout**:
```
┌─────────────────────────────┐
│ [1] Review Documentation   │  ← Order + Name
│ 📋 3 items    [L2] 🔀 3    │  ← Work items + Autonomy + Routes
└─────────────────────────────┘
```

**Implementation Tasks**:
- [ ] Add step name input/display
- [ ] Create autonomy level badge component (color-coded L0-L4)
- [ ] Add route count indicator (🔀 icon + count)
- [ ] Implement hover tooltip showing work item names
- [ ] Style autonomy badges:
  - L0 (Manual): Gray `is-light`
  - L1 (Assisted): Blue `is-info`
  - L2 (Conditional): Yellow `is-warning`
  - L3 (High Auto): Orange `is-warning`
  - L4 (Full Auto): Green `is-success`

---

### 3. Visual Route Display (Canvas)

**Design Decision**: Combination approach (Q29-D)

**Components**:
1. **Step card indicator**: Route count badge
2. **Connection lines**: Color-coded arrows
3. **Route labels**: Status badges on arrows (hover)
4. **Properties panel**: Full route table

**Color Scheme**:
- 🟢 Green: Success routes
- 🔴 Red: Failure/error routes
- 🟡 Yellow: Retry/skip routes
- 🔵 Blue (dashed): Human decision routes

**Implementation Tasks**:
- [ ] Add SVG arrow rendering between steps
- [ ] Color-code arrows based on route status
- [ ] Add route label badges (show on hover or toggle)
- [ ] Implement dashed lines for human decision routes
- [ ] Add legend/key for route colors

---

### 4. Properties Panel Integration

**Reuse**: Existing `static/js/agency-designer/properties-panel.js` (940 lines)

**Step Properties Configuration**:
```javascript
const stepPropertiesConfig = {
  title: 'Step Properties',
  sections: [
    {
      title: 'Basic Information',
      fields: [
        { name: 'name', label: 'Step Name', type: 'text', required: false },
        { name: 'description', label: 'Description', type: 'textarea', rows: 3 },
        {
          name: 'autonomy_level',
          label: 'Autonomy Level',
          type: 'select',
          required: true,
          options: [
            { value: 'L0', label: 'L0 - Manual (Human executes)' },
            { value: 'L1', label: 'L1 - Assisted (AI suggests, human approves)' },
            { value: 'L2', label: 'L2 - Conditional (AI autonomous within constraints)' },
            { value: 'L3', label: 'L3 - High Automation (AI handles most scenarios)' },
            { value: 'L4', label: 'L4 - Full Autonomy (AI fully independent)' }
          ],
          default: 'L0'
        }
      ]
    },
    {
      title: 'Parallel Execution',
      fields: [
        {
          name: 'aggregation',
          label: 'Aggregation Rule',
          type: 'select',
          options: [
            { value: '', label: 'None (sequential)' },
            { value: 'any', label: 'Any (first completion)' },
            { value: 'all', label: 'All (wait for all)' },
            { value: 'majority', label: 'Majority (>50%)' },
            { value: 'first', label: 'First (immediate)' }
          ],
          description: 'How to proceed when multiple work items in this step complete',
          showIf: (step) => step.items && step.items.length > 1
        }
      ]
    },
    {
      title: 'Conditional Routing',
      fields: [
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
              options: 'dynamic' // Populated from workflow.steps
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
            { rule: 'noSelfLoop', message: 'Step cannot route to itself' }
          ]
        }
      ]
    },
    {
      title: 'Human Decisions',
      fields: [
        {
          name: 'requires_human_decision',
          label: 'Requires Human Decision',
          type: 'checkbox',
          description: 'Pause workflow for human input at this step'
        },
        {
          name: 'available_routes',
          label: 'Decision Routes',
          type: 'table',
          columns: [
            { key: 'label', label: 'Option Label', type: 'text' },
            { key: 'targetStep', label: 'Target Step', type: 'select', options: 'dynamic' },
            { key: 'icon', label: 'Icon', type: 'text', placeholder: 'fa-check' },
            { 
              key: 'severity', 
              label: 'Severity', 
              type: 'select',
              options: ['info', 'warning', 'danger']
            }
          ],
          showIf: (step) => step.requires_human_decision === true
        }
      ]
    }
  ],
  onSave: (stepData) => {
    // Update step in workflow
    updateStepInWorkflow(stepData);
    // Refresh canvas
    refreshCanvas();
    // Show success notification
    showNotification('Step updated successfully', 'success');
  }
};
```

**Implementation Tasks**:
- [ ] Configure step properties panel fields
- [ ] Add dynamic target step dropdown population
- [ ] Implement route table validation
- [ ] Add save handler to update workflow state
- [ ] Wire up panel to step selection event

---

## JavaScript Implementation

### 5. Update `workflow-designer.js`

**New State Properties**:
```javascript
Alpine.data('workflowDesigner', () => ({
  // ... existing state
  selectedStep: null,
  showPropertiesPanel: false,
  
  // Select step and show properties
  selectStep(stepId) {
    this.selectedStep = this.workflowSteps.find(s => s.id === stepId);
    if (this.selectedStep) {
      this.openStepProperties(this.selectedStep);
    }
  },
  
  // Open properties panel
  openStepProperties(step) {
    const config = createStepPropertiesConfig(step, this.workflowSteps);
    window.PropertiesPanel.showProperties(config);
  },
  
  // Update step from properties panel
  updateStepInWorkflow(stepData) {
    const index = this.workflowSteps.findIndex(s => s.id === stepData.id);
    if (index !== -1) {
      this.workflowSteps[index] = { ...this.workflowSteps[index], ...stepData };
      this.renderCanvas();
    }
  },
  
  // Get autonomy badge class
  getAutonomyBadgeClass(level) {
    const classes = {
      'L0': 'is-light',
      'L1': 'is-info',
      'L2': 'is-warning',
      'L3': 'is-warning',
      'L4': 'is-success'
    };
    return `tag ${classes[level] || 'is-light'}`;
  },
  
  // Get route count
  getRouteCount(step) {
    return step.routes ? Object.keys(step.routes).length : 0;
  }
}));
```

**Implementation Tasks**:
- [ ] Add step selection state management
- [ ] Implement `selectStep()` method
- [ ] Implement `openStepProperties()` integration
- [ ] Add autonomy badge helper methods
- [ ] Add route count calculation
- [ ] Update step card rendering to include new fields

---

### 6. Canvas Rendering Updates

**Route Visualization**:
```javascript
function renderRoutes(steps) {
  const svg = document.getElementById('route-connections');
  svg.innerHTML = ''; // Clear existing
  
  steps.forEach(step => {
    if (step.routes) {
      Object.entries(step.routes).forEach(([status, targetStepId]) => {
        const targetStep = steps.find(s => s.id === targetStepId);
        if (targetStep) {
          const arrow = createArrow(step, targetStep, status);
          svg.appendChild(arrow);
        }
      });
    }
  });
}

function createArrow(fromStep, toStep, status) {
  const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
  
  // Position calculation
  const fromPos = getStepPosition(fromStep);
  const toPos = getStepPosition(toStep);
  
  line.setAttribute('x1', fromPos.x);
  line.setAttribute('y1', fromPos.y);
  line.setAttribute('x2', toPos.x);
  line.setAttribute('y2', toPos.y);
  
  // Color coding
  const colors = {
    success: '#48c774',  // Green
    failure: '#f14668',  // Red
    error: '#f14668',    // Red
    retry: '#ffe08a',    // Yellow
    skip: '#ffe08a'      // Yellow
  };
  
  line.setAttribute('stroke', colors[status] || '#3273dc');
  line.setAttribute('stroke-width', '2');
  line.setAttribute('marker-end', 'url(#arrowhead)');
  
  // Add status badge on hover
  line.addEventListener('mouseenter', () => showRouteBadge(status, toPos));
  line.addEventListener('mouseleave', () => hideRouteBadge());
  
  return line;
}
```

**Implementation Tasks**:
- [ ] Add SVG container for route arrows
- [ ] Implement `renderRoutes()` function
- [ ] Create arrow SVG elements with color coding
- [ ] Add arrowhead markers
- [ ] Implement route badge tooltips
- [ ] Handle canvas resize/scroll

---

## Template Updates

### 7. Update `workflow_designer.templ`

**Add Properties Panel Integration**:
```templ
templ WorkflowDesigner(agency *models.Agency, workflow *models.Workflow) {
    @components.LayoutWithAgency("Workflow Designer", agency) {
        <section class="section">
            <div class="container is-fluid">
                <!-- Status Bar -->
                @components.StatusBarSimple()
                
                <!-- Main Canvas Area -->
                <div class="columns">
                    <div class="column is-9">
                        <div id="workflow-canvas" class="box">
                            <!-- Steps render here -->
                            <svg id="route-connections" style="position:absolute;top:0;left:0;width:100%;height:100%;pointer-events:none;">
                                <defs>
                                    <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                                        <polygon points="0 0, 10 3.5, 0 7" fill="#3273dc" />
                                    </marker>
                                </defs>
                            </svg>
                        </div>
                    </div>
                    
                    <!-- Properties Panel (hidden by default) -->
                    <div id="properties-panel-container" class="column is-3" style="display:none;">
                        <!-- Properties panel renders here via JavaScript -->
                    </div>
                </div>
                
                <!-- Route Legend -->
                <div class="box mt-4">
                    <h4 class="title is-6">Route Legend</h4>
                    <div class="tags">
                        <span class="tag" style="background:#48c774">Success</span>
                        <span class="tag" style="background:#f14668">Failure/Error</span>
                        <span class="tag" style="background:#ffe08a">Retry/Skip</span>
                        <span class="tag is-info" style="border:2px dashed #3273dc">Human Decision</span>
                    </div>
                </div>
            </div>
        </section>
    }
}
```

**Implementation Tasks**:
- [ ] Add SVG container for route arrows
- [ ] Add properties panel container column
- [ ] Add route legend section
- [ ] Wire up Alpine.js component

---

## Testing Plan

### Unit Tests
- [ ] Step model validation (AutonomyLevel required, valid values)
- [ ] Route validation (no self-loops, valid target steps)
- [ ] Aggregation rule validation
- [ ] HumanRoute model creation

### Integration Tests
- [ ] Properties panel opens on step click
- [ ] Step updates save to workflow state
- [ ] Route arrows render correctly
- [ ] Autonomy badges display with correct colors
- [ ] Route table validates uniqueness

### Manual Testing
- [ ] Create workflow with multiple steps
- [ ] Add routes between steps
- [ ] Configure autonomy levels
- [ ] Verify visual routes match configuration
- [ ] Test human decision route configuration
- [ ] Verify properties panel saves correctly

---

## Acceptance Criteria

**Step Configuration**:
- ✅ User can set step name and description
- ✅ User can select autonomy level (L0-L4) with default L0
- ✅ User can configure aggregation for parallel work items
- ✅ Autonomy badge displays on step card with correct color

**Conditional Routing**:
- ✅ User can define status-to-route mappings via table editor
- ✅ Routes validate (no duplicates, no self-loops)
- ✅ Visual arrows display between steps with color coding
- ✅ Route count badge shows on step card

**Human Decisions**:
- ✅ User can mark step as requiring human decision
- ✅ User can define available decision routes
- ✅ Human decision routes show as dashed lines
- ✅ Configuration stored correctly in workflow model

**UI/UX**:
- ✅ Step cards show compact information (order, name, items, autonomy, routes)
- ✅ Properties panel opens on step click
- ✅ Changes save and update canvas immediately
- ✅ Route legend explains color coding
- ✅ Workflow scales to 20+ steps without clutter

---

## Implementation Phases

### Phase 1: Data Model (1-2 days)
- Update Step struct in `workflow.go`
- Add HumanRoute struct
- Database migration
- API endpoint updates

### Phase 2: Step Cards (2-3 days)
- Add name/description fields
- Implement autonomy badges
- Add route count indicator
- Hover tooltips

### Phase 3: Properties Panel (3-4 days)
- Configure step properties fields
- Implement route table editor
- Add validation rules
- Wire up save handlers

### Phase 4: Visual Routes (3-4 days)
- SVG arrow rendering
- Color coding implementation
- Route labels/tooltips
- Legend component

### Phase 5: Testing & Polish (2-3 days)
- Unit tests
- Integration tests
- Bug fixes
- Documentation

**Total Estimate**: 11-16 days

---

## Related Documentation

- **Research Session**: `/workspaces/CodeValdCortex/documents/3-SofwareDevelopment/research/workflow-designer/`
- **Architecture**: `/workspaces/CodeValdCortex/documents/2-SoftwareDesignAndArchitecture/frontend-architecture-updated.md`
- **Workflow Models**: `/workspaces/CodeValdCortex/internal/agency/models/workflow.go`
- **Properties Panel**: `/workspaces/CodeValdCortex/static/js/agency-designer/properties-panel.js`
- **Workflow Designer**: `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/workflow_designer.templ`

---

## Notes

- **Execution UI is out of scope**: Human decision UI happens in the workbench, not the designer
- **Designer focus**: Configuration and visualization only
- **Component reuse**: Leverage existing PropertiesPanel, StatusBar, showNotification utilities
- **Validation**: Save with warnings OK for draft, strict validation before deployment (future)
- **Evolution path**: Start with table-based route editor, can add visual builder in v2
