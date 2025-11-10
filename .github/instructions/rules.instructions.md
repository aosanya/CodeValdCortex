---
applyTo: '**'
---

# Code Structure Rules

## Design Reference

**Main design concepts and styling reference:**

- **Location**: `/workspaces/CodeValdCortex/internal/web/designs/version1/`
- This directory contains HTML/CSS design references that show:
  - Complete page layouts and structure
  - Bulma CSS class usage patterns
  - Custom CSS implementations
  - Component hierarchy and organization
  - FontAwesome icon usage
  - Theme-related CSS classes
- Use these files as visual and structural references when implementing `.templ` templates
- **Note**: These are styling references only - do not copy JavaScript functionality
- Follow the template-first architecture when converting designs to `.templ` files

## Template-First Architecture

**Always prefer `.templ` files over Go/JavaScript for HTML generation.**

- HTML markup should be defined in `.templ` files (using the templ template engine)
- **NEVER generate HTML strings in Go handler files** - use `.templ` files instead
- **NEVER generate HTML strings in JavaScript files** - use `.templ` files or server-side rendering
- **NEVER put JavaScript logic in `.templ` files** - keep `.templ` files for HTML structure only
- JavaScript should only handle:
  - Event handling
  - Data fetching
  - DOM manipulation (show/hide, updates)
  - State management
  - Business logic for UI interactions
- JavaScript belongs in `.js` files in the `static/js/` directory
- Go handlers should:
  - Process business logic
  - Call services
  - Return data structures (JSON) or render `.templ` templates
  - Never contain HTML strings or fmt.Sprintf with HTML
- `.templ` files should only contain:
  - HTML structure and markup
  - Templ template directives (if/else, for loops, etc.)
  - Data attributes for JavaScript hooks
  - CSS classes
  - **NO inline JavaScript** (no `<script>` tags with logic)
  - **NO event handler logic** (onclick should call named functions defined in .js files)
- Pre-render all content sections in templates, then toggle visibility with JavaScript
- Benefits:
  - Type safety
  - Server-side rendering capability
  - Better maintainability
  - Clear separation of concerns
  - Easier testing
  - Reusable JavaScript functions

**Example of WRONG approach:**
```go
// ❌ NEVER DO THIS - HTML in Go handler
func (h *Handler) SomeHandler(c *gin.Context) {
    html := fmt.Sprintf(`
        <div class="content">
            <textarea>%s</textarea>
        </div>
    `, data)
    c.Data(200, "text/html", []byte(html))
}
```

```templ
// ❌ NEVER DO THIS - JavaScript logic in templ file
templ MyComponent(data string) {
    <div id="container">{ data }</div>
    <script>
        // Don't put JavaScript logic here!
        async function handleClick() {
            const response = await fetch('/api/data');
            // ... more logic
        }
    </script>
}
```

**Example of CORRECT approach:**
```go
// ✅ CORRECT - Use templ file
func (h *Handler) SomeHandler(c *gin.Context) {
    // Process data
    data := processData()
    
    // Render template
    component := templates.SomeComponent(data)
    component.Render(c.Request.Context(), c.Writer)
}
```

```templ
// ✅ CORRECT - Clean HTML structure only
templ MyComponent(data string) {
    <div id="container" data-initial-value={ data }>{ data }</div>
    <button onclick="handleMyComponentClick()" id="my-button">
        Click Me
    </button>
    <!-- JavaScript logic is in static/js/my-component.js -->
}
```

```javascript
// ✅ CORRECT - JavaScript in separate .js file
// static/js/my-component.js
window.handleMyComponentClick = async function() {
    const container = document.getElementById('container');
    const initialValue = container.dataset.initialValue;
    
    const response = await fetch('/api/data');
    // ... handle response
}
```

## CSS and Styling

**Minimize custom CSS by leveraging Bulma CSS framework.**

- Use Bulma's built-in classes whenever possible
- Only create custom CSS when Bulma doesn't provide the needed styling
- Keep custom CSS files minimal and focused
- Prefer Bulma utility classes over custom styles
- Benefits:
  - Consistent design language
  - Less CSS to maintain
  - Faster development
  - Better responsive design out of the box

## Code Quality and File Organization

**Keep files small and focused. Break down large files.**

- **Maximum file size**: ~500-700 lines
- If a file exceeds 700 lines, it should be broken down into smaller, focused modules
- Each file should have a single, clear responsibility
- Handler files should be split by domain/feature area
- Service files should be modular and composable

**Functions should be concise and testable.**

- **Maximum function length**: ~50 lines (prefer 20-30)
- Each function should do one thing well
- Use functional programming principles:
  - Pure functions when possible (no side effects)
  - Functions should be easily testable in isolation
  - Avoid deeply nested logic
  - Use composition over complex inheritance
- Extract complex logic into separate, named functions
- Prefer dependency injection for testability

**Example of breaking down a large handler file:**
```
❌ BEFORE (1 large file):
internal/web/handlers/ai_refine_handler.go (700+ lines)

✅ AFTER (split by domain):
internal/web/handlers/ai_refine/
├── handler.go           (main handler struct, ~100 lines)
├── introduction.go      (introduction refinement, ~150 lines)
├── goal.go             (goal refinement/generation, ~150 lines)
└── helpers.go          (shared utilities, ~100 lines)
```

**Example of functional, testable code:**
```go
// ❌ WRONG - Hard to test, does too much
func (h *Handler) ProcessData(c *gin.Context) {
    data := c.GetString("data")
    if data == "" {
        c.JSON(400, gin.H{"error": "missing data"})
        return
    }
    processed := strings.ToUpper(data)
    result := h.service.Save(processed)
    if result.Error != nil {
        h.logger.Error("failed", result.Error)
        c.JSON(500, gin.H{"error": "failed"})
        return
    }
    c.JSON(200, gin.H{"result": result.Data})
}

// ✅ CORRECT - Testable, single responsibility
func (h *Handler) ProcessData(c *gin.Context) {
    data, err := extractData(c)
    if err != nil {
        respondWithError(c, http.StatusBadRequest, err)
        return
    }
    
    processed := processInput(data)
    
    if err := h.service.Save(processed); err != nil {
        h.logger.Error("save failed", "error", err)
        respondWithError(c, http.StatusInternalServerError, err)
        return
    }
    
    respondWithSuccess(c, processed)
}

// Pure, testable functions
func extractData(c *gin.Context) (string, error) {
    data := c.GetString("data")
    if data == "" {
        return "", errors.New("missing data")
    }
    return data, nil
}

func processInput(input string) string {
    return strings.ToUpper(input)
}
```

## Task Management and Workflow

### Branch Management (MANDATORY)
For each new task:
```bash
# Create feature branch
git checkout -b feature/MVP-XXX_description

# Work on task implementation
# ... development work ...

# Build validation before merge
# - Follow coding standards
# - Run linting and validation tools
# - Verify code follows established patterns
# - Check for deprecated API usage
# - Remove unused code/imports/variables
# - Run build processes and tests
# - Fix any build errors or warnings

# Merge when complete and tested
git checkout main
git merge feature/MVP-XXX_description
git branch -d feature/MVP-XXX_description
```

### Task Completion Process (MANDATORY)
1. **Task Assignment**: Pick tasks based on priority (P0 first) and dependencies
2. **Implementation**: Update "Status" column as work progresses (Not Started → In Progress → Testing → Complete)
3. **Completion Process**:
   - Create detailed coding session document in `coding_sessions/` using format: `{TaskID}_{description}.md`
   - Add completed task to summary table in `mvp_done.md` with completion date
   - Remove completed task from active `mvp.md` file
   - Update any dependent task references
   - **Update README.md roadmap**: For milestone/major feature completion, update the roadmap section with brief, concise progress notes (move items from planned to completed, update current focus)
   - Merge feature branch to main (see Branch Management above)
4. **Dependencies**: Ensure prerequisite tasks are completed before starting dependent work

### README.md Update Guidelines
- **When to Update**: Major milestones, completed feature groups, significant architecture changes
- **Keep Changes Brief**: Update roadmap status, current focus, and key capabilities only
- **Areas to Update**: 
  - Roadmap section (✅ completed items, 🔄 current focus)
  - Key Features (if new major capabilities added)
  - Current development focus and next milestones
- **Avoid**: Detailed technical changes, minor bug fixes, work-in-progress updates

### Repository Structure
```
/workspaces/CodeValdCortex/
├── documents/3-SofwareDevelopment/
│   ├── mvp.md                    # Active tasks only
│   ├── mvp_done.md              # Completed tasks archive
│   └── coding_sessions/         # Detailed implementation logs
├── [project code structure]     # Implementation code
└── [other project folders]      # Additional project resources
```

## 🚨 Architectural Guidelines and Anti-Patterns

### CRITICAL: Prevent Type/Model Duplication

**❌ NEVER duplicate types across packages**
- **Problem**: Multiple definitions of the same concept (e.g., `WorkflowStatus` in 4+ places)
- **Rule**: Create shared types in `internal/shared/types/` for common enums and structs
- **Example**: Instead of defining `WorkflowStatus` in each package, import from shared location

```go
// ✅ CORRECT - Single source of truth
// internal/shared/types/workflow.go
type WorkflowStatus string
const (
    WorkflowStatusPending WorkflowStatus = "pending"
    WorkflowStatusRunning WorkflowStatus = "running"
    // ...
)

// ❌ WRONG - Duplicated across packages
// internal/orchestration/types.go - DON'T DO THIS
type WorkflowStatus string // Already exists elsewhere!
```

### Domain Boundaries and Package Organization

**Establish clear domain separation:**

```
internal/
├── shared/                  # Common types, utilities, errors
│   ├── types/              # Shared enums, constants, basic types
│   ├── errors/             # Common error types
│   └── utils/              # Utility functions
├── domain/                 # Business logic domains
│   ├── agency/
│   │   ├── models/         # Agency-specific models only
│   │   ├── repository/     # Data access interfaces
│   │   └── service/        # Business logic
│   ├── workflow/
│   │   ├── models/         # Workflow-specific models
│   │   ├── repository/     # Workflow data access
│   │   └── service/        # Workflow business logic
│   └── agent/
│       ├── models/
│       ├── repository/
│       └── service/
├── application/            # Use cases, orchestration
│   ├── usecases/          # Application-specific business flows
│   └── services/          # Cross-domain services
├── infrastructure/        # External concerns
│   ├── persistence/       # Database implementations
│   ├── messaging/         # Event/message handling
│   └── external/          # External service integrations
└── interfaces/            # Input/output adapters
    ├── http/              # HTTP handlers
    ├── cli/               # CLI commands
    └── grpc/              # gRPC services
```

### Repository Pattern Rules

**❌ NEVER create massive repository files (>300 lines)**
- **Current Issue**: `orchestration/repository.go` is 560+ lines
- **Rule**: Split by aggregate root or functional concern
- **Example**: 
  ```
  ✅ CORRECT:
  infrastructure/persistence/
  ├── workflow_repository.go     # Workflow CRUD only
  ├── execution_repository.go    # Execution CRUD only  
  └── workflow_stats_service.go  # Statistics as separate service
  ```

**Repository responsibilities (ONLY):**
- Data persistence (CRUD operations)
- Simple queries and filtering
- Data mapping/transformation

**❌ Repositories should NEVER contain:**
- Complex business logic
- Statistics calculations
- Cross-aggregate operations
- Event publishing

### Naming Conventions

**Package names must be consistent:**
```go
// ✅ CORRECT - Use singular form
internal/agent/
internal/workflow/ 
internal/task/

// ❌ WRONG - Mixed singular/plural
internal/handlers/    // Should be internal/handler/
internal/events/      // Should be internal/event/
```

**Type naming must be unambiguous:**
```go
// ✅ CORRECT - Domain-prefixed when necessary
type AgencyWorkflow struct{}     // Clear context
type ExecutionStatus string      // Specific to execution
type SharedWorkflowStatus string // Explicit shared type

// ❌ WRONG - Generic names that conflict
type Workflow struct{}    // Which domain?
type Status string        // Status of what?
```

### Cross-Package Dependencies

**❌ PREVENT circular dependencies:**
- Domain packages should not import each other
- Use dependency injection and interfaces
- Import direction: `interfaces -> application -> domain -> shared`

```go
// ✅ CORRECT - One-way dependency
// domain/workflow depends on shared
import "internal/shared/types"

// ❌ WRONG - Circular dependency  
// agency imports workflow AND workflow imports agency
```

### File Size Limits

**Enforce maximum file sizes:**
- **Handler files**: Max 300 lines (split by domain if larger)
- **Service files**: Max 400 lines (extract helper services)
- **Repository files**: Max 250 lines (split by aggregate)
- **Model files**: Max 200 lines (group related models)

### Pre-Development Checklist

**Before adding new code, ask:**
1. ✅ Does this type already exist elsewhere?
2. ✅ Which domain does this belong to?
3. ✅ Am I creating a circular dependency?
4. ✅ Is this file getting too large (check line count)?
5. ✅ Are my imports going in the correct direction?
6. ✅ Am I putting business logic in the right layer?

### Code Review Requirements

**Every PR must verify:**
- [ ] No duplicate types or constants
- [ ] Clear domain separation
- [ ] No files exceeding size limits
- [ ] No circular dependencies
- [ ] Consistent naming conventions
- [ ] Repository pattern compliance

**Automatic checks to implement:**
- Linter rules for file sizes
- Import cycle detection
- Duplicate type detection
- Package naming validation

