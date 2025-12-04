# Workflow Designer Research

**Research Date**: December 4, 2025  
**Feature**: Workflow Designer Enhancements  
**Branch**: `feature/workflow-integration-tests`  

---

## 📋 Overview

Comprehensive research session exploring enhancements to the workflow designer, including:
- Status bar integration
- Properties panel for step/work item configuration
- Step-level autonomy control (L0-L4)
- Conditional flow control (branching, looping)
- Human-directed routing

---

## 📂 Research Documents

### Core Features
1. **[status-bar-properties.md](./status-bar-properties.md)** - Status bar and properties panel integration
2. **[autonomy-levels.md](./autonomy-levels.md)** - Step-level autonomy control (L0-L4)
3. **[conditional-flow.md](./conditional-flow.md)** - Conditional routing, looping, and branching
4. **[human-routing.md](./human-routing.md)** - Human-in-the-loop decision routing

### Implementation Details
5. **[data-models.md](./data-models.md)** - Enhanced Step and Workflow data models
6. **[ui-design.md](./ui-design.md)** - UI/UX patterns for workflow designer
7. **[reusable-components.md](./reusable-components.md)** - Inventory of components to reuse

---

## 🎯 Key Research Findings

### Design Decisions (Q&A Session)

**Properties Panel:**
- ✅ Read-only for work item details
- ✅ Editable for step properties (name, description, autonomy, routes)
- ✅ Minimal reference view for work items in steps

**Autonomy Levels:**
- ✅ Step-level control (not work item or role level)
- ✅ Default to L0 (Manual) for new steps
- ✅ Required field with validation (L0-L4)

**Conditional Flow:**
- ✅ Status-based routing (success/failure/error)
- ✅ Aggregation for parallel steps (any/all/majority/first)
- ✅ Pre-evaluated state properties (no runtime condition evaluation)
- ✅ Human-directed routing with predefined route options

**UI/UX:**
- ✅ Maintain vertical layout (no graph view)
- ✅ Show routes prominently on canvas
- ✅ Use existing components (StatusBar, PropertiesPanel, etc.)

---

## 🏗️ Implementation Phases

### Phase 1: Status Bar & Properties Panel (Completed in Research)
- Add VS Code-style status bar
- Integrate properties panel component
- Selection state and visual feedback

### Phase 2: Step-Level Autonomy (Completed in Research)
- Update Step data model
- Add autonomy level fields to properties panel
- Validation rules

### Phase 3: Conditional Flow Control (In Progress)
- Enhance Step model with routes
- Implement aggregation logic
- Visual route indicators on canvas

### Phase 4: Human-Directed Routing (In Progress)
- Add human decision routes
- Decision UI (modal/inline/notification)
- Audit trail and justification capture

---

## 📊 Data Model Changes

### Enhanced Step Model
```go
type Step struct {
    // Core fields
    ID            string     `json:"id" binding:"required"`
    Order         int        `json:"order" binding:"required"`
    Name          string     `json:"name,omitempty"`
    Description   string     `json:"description,omitempty"`
    AutonomyLevel string     `json:"autonomy_level" binding:"required,oneof=L0 L1 L2 L3 L4"`
    Items         []StepItem `json:"items" binding:"required,min=1"`
    
    // Conditional flow
    Routes              map[string]string `json:"routes,omitempty"`
    Aggregation         string            `json:"aggregation,omitempty"`
    DefaultRoute        string            `json:"default_route,omitempty"`
    
    // Human decisions
    RequiresHumanDecision bool         `json:"requires_human_decision,omitempty"`
    AvailableRoutes       []HumanRoute `json:"available_routes,omitempty"`
    
    // Runtime state
    Status        string `json:"status,omitempty"`
    ResolvedRoute string `json:"resolved_route,omitempty"`
}
```

See [data-models.md](./data-models.md) for complete specifications.

---

## 🔄 Next Steps

1. **Complete Q28**: Human decision UI pattern (modal/inline/notification)
2. **Document**: Visual design patterns for route display
3. **Prototype**: Canvas UI with route indicators
4. **Implement**: Phase 1 (Status bar + Properties panel)
5. **Iterate**: Test with real workflows

---

## 📚 Related Documentation

- [Workflow Designer Template](../../../../internal/web/pages/agency_designer/workflow_designer.templ)
- [Workflow Designer JavaScript](../../../../static/js/workflow-designer.js)
- [Properties Panel Component](../../../../static/js/agency-designer/properties-panel.js)
- [Status Bar Component](../../../../internal/web/components/status_bar.templ)
- [Agency Designer Reference](../../../../internal/web/pages/agency_designer/agency_designer.templ)

---

**Research Status**: 🔄 Active (28 questions completed, continuing exploration)
