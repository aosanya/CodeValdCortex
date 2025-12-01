# Work Item Deliverables Structure - Research Session

**Date**: 2025-12-01  
**Research Focus**: Enhanced deliverables architecture for work items with folder structures and AI prompt instructions

---

## Executive Summary

Work item deliverables are evolving from simple string arrays to hierarchical folder/file structures with embedded AI prompt instructions. This enables work items to define not just WHAT should be delivered, but also WHERE it should be organized and HOW AI agents should approach creating each deliverable.

---

## Key Architectural Decisions

### 1. Hierarchical Tree Structure (vs Flat Array)

**Decision**: Use nested tree structure with `DeliverableNode` containing children array.

**Rationale**:
- ✅ Natural representation of folder hierarchies
- ✅ Agents can traverse structure programmatically
- ✅ UI can render tree/folder views
- ✅ Supports future features (templates, validation)
- ✅ Easier exploration and navigation

**Example**:
```
Software Requirements (folder)
├─ introduction/ (folder)
│  ├─ stakeholders.md (file)
│  ├─ vision.md (file)
│  └─ scope.md (file)
└─ requirements/ (folder)
   ├─ functional.md (file)
   └─ non-functional.md (file)
```

### 2. User-Defined Structure (vs Auto-Generated)

**Decision**: Users define folder structures manually through UI.

**Rationale**:
- Flexibility for different project types
- No enforced templates or conventions
- Users know their domain best
- System provides guidance but doesn't constrain

### 3. Prompt Instructions Per Node

**Decision**: Every folder AND file has `prompt_instructions` field.

**Rationale**:
- This is the primary communication mechanism for "what to achieve"
- Folder-level instructions provide context
- File-level instructions provide specific tasks
- AI agents consume these as structured prompts

**Example**:
```go
{
  "name": "stakeholders.md",
  "type": "file",
  "prompt_instructions": "List all stakeholders with roles, influence levels, and key interests. Use markdown tables."
}
```

### 4. Inheritance Model

**Decision**: Child nodes inherit parent folder prompt instructions.

**Rationale**:
- Reduces redundancy (don't repeat context)
- Natural hierarchical flow (guidelines → specifics)
- AI gets full context from root to leaf
- Supports consistent tone/style across sections

### 5. Prompt Composition Format

**Decision**: Hierarchical sections format (Option B).

**Format**:
```
PROJECT GUIDELINES:
[Root deliverable prompt]

FOLDER GUIDELINES:
[Parent folder prompts]

YOUR TASK:
[Specific file prompt]
```

**Rationale**:
- Clear visual hierarchy for AI
- Easy to understand scope and context
- Extensible for multiple levels
- Separates constraints from tasks

### 6. No Backward Compatibility

**Decision**: Replace old `Deliverables []string` directly with new structure.

**Rationale**:
- System not in production use yet
- Clean break, simpler implementation
- No migration complexity
- Can iterate faster

### 7. Markdown-Only Initially

**Decision**: Only `.md` files supported in first version.

**Rationale**:
- Focused scope for MVP
- Covers primary use case (documentation)
- Can extend to other file types later
- Simpler validation rules

---

## Data Model

```go
type DeliverableNode struct {
    ID                 string            `json:"id"`                   // UUID for UI tracking
    Name               string            `json:"name"`                 // "stakeholders.md" or "introduction"
    Description        string            `json:"description"`          // Brief description
    Path               string            `json:"path"`                 // Computed full path
    Type               DeliverableType   `json:"type"`                 // "folder" or "file"
    PromptInstructions string            `json:"prompt_instructions"`  // AI instructions
    Children           []DeliverableNode `json:"children,omitempty"`   // Nested items
    FileExtension      string            `json:"file_extension"`       // ".md"
}

type DeliverableType string
const (
    DeliverableTypeFolder DeliverableType = "folder"
    DeliverableTypeFile   DeliverableType = "file"
)
```

---

## Validation Constraints

Comprehensive validation rules to maintain data integrity:

```go
const (
    MaxNestingDepth     = 10    // Prevent excessive nesting
    MaxNameLength       = 255   // File/folder name limit
    MaxPromptLength     = 5000  // Prompt instruction limit
    MaxChildrenPerNode  = 100   // Prevent overly complex trees
)
```

**Validation Rules**:
1. **Required Fields**: Name, Type (always required)
2. **Path Validation**:
   - No duplicate paths within work item
   - Valid characters only (alphanumeric, dash, underscore, slash)
   - Folders cannot have file extensions
   - Files must have `.md` extension
3. **Prompt Instructions**: Recommended (supports inheritance if empty)
4. **Nesting**: Max 10 levels deep
5. **Complexity**: Max 100 children per node

---

## UI Design: Tree Builder

**Selected Approach**: Visual tree builder with drag-and-drop (Option A)

**Features**:
- Visual tree/folder view
- Add Folder / Add File buttons
- Drag-and-drop to reorder and nest
- Inline editing for names, descriptions, prompt instructions
- Expand/collapse navigation
- File type indicators (icons)
- Path preview (auto-computed)

**Benefits**:
- Intuitive for users familiar with file systems
- Easy for simple cases (single file)
- Powerful for complex structures (nested folders)
- Visual feedback on hierarchy
- Works well with AI assistance

---

## Prompt Composition Service

**Purpose**: Generate hierarchical AI prompts from deliverable tree structure.

```go
type PromptComposer interface {
    // Compose hierarchical prompt from deliverable node and ancestors
    ComposePrompt(node *DeliverableNode) (string, error)
}
```

**Algorithm**:
1. Walk from leaf node up to root
2. Collect prompt instructions at each level
3. Format as hierarchical sections
4. Return composed prompt string

**Example Output**:
```
PROJECT GUIDELINES:
Create professional software requirements documentation following IEEE 830 standards. Use formal tone.

FOLDER GUIDELINES (introduction/):
Cover project background, stakeholders, and objectives. Each document should be comprehensive but concise.

YOUR TASK (stakeholders.md):
List all stakeholders with roles, influence levels, and key interests. Use markdown tables.
```

---

## Integration Points

### 1. Agency Designer
- Work Items editor gets tree builder UI
- Replace old textarea with tree component
- AI refinement suggests folder structures

### 2. AI Builder Service
- Use prompt instructions when generating content
- Compose hierarchical prompts for agents
- Suggest deliverable structures based on work item type

### 3. Workbench / Kanban
- Display deliverable tree structure per issue
- Show completion status per deliverable node
- Track which files have been created

### 4. Agent Execution Engine
- Agents read deliverables to know what to create
- Use composed prompts as instructions
- Create files in specified paths

### 5. Git Integration
- Deliverables map to file paths in agency repository
- File structure mirrors deliverable tree
- Validation ensures paths are valid

---

## Implementation Roadmap

### Phase 1: Data Model & Backend (Week 1-2)
- [ ] Update `WorkItem` model with `Deliverables []DeliverableNode`
- [ ] Create `DeliverableNode` model
- [ ] Implement `DeliverableValidator` service
- [ ] Create `PromptComposer` service
- [ ] Add path computation logic
- [ ] Update API endpoints

### Phase 2: UI Components (Week 2-3)
- [ ] Build tree builder component (Templ)
- [ ] Add drag-and-drop functionality (Alpine.js)
- [ ] Create inline editors for node properties
- [ ] Implement expand/collapse navigation
- [ ] Add folder/file type indicators
- [ ] Path preview display

### Phase 3: AI Integration (Week 3-4)
- [ ] Integrate prompt composition with AI builder
- [ ] Test hierarchical prompt inheritance
- [ ] Add deliverable structure suggestions
- [ ] Implement AI-powered template generation

### Phase 4: Agency Designer Integration (Week 4-5)
- [ ] Update Work Items editor UI
- [ ] Replace deliverables textarea with tree builder
- [ ] Update AI refinement for structure suggestions
- [ ] Add export functionality for deliverable templates
- [ ] Testing and validation

---

## Example Use Case: Software Requirements Work Item

**Work Item**: "Requirements Documentation"

**Deliverable Structure**:
```
Software Requirements (folder)
  Path: documents/1-SoftwareRequirements
  Prompt: "Create professional software requirements documentation following IEEE 830 standards. Use formal tone."
  
  ├─ introduction/ (folder)
  │   Prompt: "Cover project background, stakeholders, and objectives. Each document should be comprehensive but concise."
  │   
  │   ├─ stakeholders.md (file)
  │   │   Prompt: "List all stakeholders with roles, influence levels, and key interests. Use markdown tables."
  │   │
  │   ├─ vision.md (file)
  │   │   Prompt: "Document the project vision, goals, and success criteria."
  │   │
  │   └─ scope.md (file)
  │       Prompt: "Define project scope, in-scope items, out-of-scope items, and constraints."
  │
  └─ requirements/ (folder)
      Prompt: "Document functional and non-functional requirements with clear acceptance criteria."
      
      ├─ functional.md (file)
      │   Prompt: "List all functional requirements using standard format: ID, Description, Priority, Acceptance Criteria."
      │
      └─ non-functional.md (file)
          Prompt: "Document performance, security, scalability, and usability requirements."
```

**When Agent Works on `stakeholders.md`**:

Composed Prompt:
```
PROJECT GUIDELINES:
Create professional software requirements documentation following IEEE 830 standards. Use formal tone.

FOLDER GUIDELINES (introduction/):
Cover project background, stakeholders, and objectives. Each document should be comprehensive but concise.

YOUR TASK (stakeholders.md):
List all stakeholders with roles, influence levels, and key interests. Use markdown tables.
```

---

## Files to Create/Update

### New Files
```
internal/agency/models/deliverable_node.go
internal/agency/services/deliverable_validator.go
internal/agency/services/prompt_composer.go
internal/web/pages/agency_designer/deliverable_tree_builder.templ
static/js/agency-designer/deliverable-tree.js
```

### Files to Update
```
internal/agency/models/work_item.go
internal/web/pages/agency_designer/work_item_editor_card.templ
static/js/agency-designer/work-items.js
documents/3-SofwareDevelopment/mvp-details/MVP-043.md (✅ DONE)
```

---

## Open Questions / Future Enhancements

### Short-term (Defer to Later)
- Template library for common deliverable structures
- Import/export deliverable structures
- Visual diff for deliverable tree changes
- Completion tracking per deliverable node

### Medium-term (Post-MVP)
- Support for other file types (.json, .yaml, .go, etc.)
- Deliverable versioning and history
- Collaborative editing of deliverable structures
- AI-suggested folder structures based on industry patterns

### Long-term (Future Research)
- Integration with actual file system (sync)
- Deliverable dependency graphs
- Quality metrics per deliverable
- Automated deliverable validation

---

## Success Metrics

**Definition of Done**:
- ✅ Users can create hierarchical deliverable structures
- ✅ Tree builder UI is intuitive and functional
- ✅ Validation enforces all constraints
- ✅ Prompt composition generates correct hierarchical format
- ✅ AI agents can consume deliverable structures
- ✅ Integration with existing work item workflows
- ✅ Export includes deliverable tree structure
- ✅ Documentation updated with new architecture

---

## References

- **MVP-043**: Work Items UI Module (updated with enhanced deliverables)
- **Research Prompt**: `/workspaces/CodeValdCortex/.github/prompts/research.prompt.md`
- **Existing Work Item Model**: `internal/agency/models/work_item.go`
- **Example Structure**: `documents/1-SoftwareRequirements/introduction/`

---

**Research Session Complete**: 2025-12-01  
**Next Action**: Begin Phase 1 implementation (data model & backend services)
