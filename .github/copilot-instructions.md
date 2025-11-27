# CodeValdCortex - AI Agent Development Instructions

## Project Overview

**CodeValdCortex** is an enterprise multi-agent AI orchestration platform - the "Kubernetes of AI Agents". It enables organizations to deploy autonomous AI agent teams safely with complete visibility, multi-vendor interoperability (A2A Protocol), and cloud-native architecture.

**Core Concept**: Each "Agency" is an isolated workspace with its own ArangoDB database, agents, workflows, and policies. The system uses database-per-agency multi-tenancy for complete data isolation.

## Critical Architecture Patterns

### 1. Template-First Architecture (MANDATORY)

**HTML MUST be in `.templ` files, NEVER in Go handlers or JavaScript:**

```go
// ❌ WRONG - HTML in Go handler
func (h *Handler) Handle(c *gin.Context) {
    html := fmt.Sprintf(`<div>%s</div>`, data)
    c.Data(200, "text/html", []byte(html))
}

// ✅ CORRECT - Use templ template
func (h *Handler) Handle(c *gin.Context) {
    component := templates.SomeComponent(data)
    component.Render(c.Request.Context(), c.Writer)
}
```

**Build workflow**: Always run `templ generate` after editing `.templ` files to regenerate Go code.

### 2. Component Reuse (Repository Size Critical)

**ALWAYS use shared layouts - NEVER duplicate page structures:**

```templ
// ✅ CORRECT - Use shared layout
templ MyPage(agency *models.Agency) {
    @components.LayoutWithAgency("Page Title", agency) {
        <section class="section">
            <!-- page content only -->
        </section>
    }
}

// ❌ WRONG - Duplicating entire HTML structure
templ MyPage(agency *models.Agency) {
    <!DOCTYPE html>
    <html>
        <head>...</head>
        <body>...</body>
    </html>
}
```

**Before creating ANY new component**: Search `internal/web/components/` for existing similar components.

### 3. Database-Per-Agency Architecture

**Each agency has its own ArangoDB database** (e.g., `UC-INFRA-001`, `UC-CHAR-001`):

```go
// Get agency-specific database
agencyDB := agencyID  // Agency ID IS the database name
if agency.Database != "" {
    agencyDB = agency.Database
}

db, err := h.dbClient.GetDatabase(ctx, agencyDB)
// Use this db for ALL agency-scoped operations

// Create agency-scoped repositories
agencyRegistry, err := registry.NewRepositoryWithDB(db)
```

**Master database** (`codevaldcortex`) stores only agency metadata in the `agencies` collection. All other data (agents, tasks, workflows, etc.) lives in agency-specific databases.

### 4. Service Layer Architecture

**Standard pattern for services across codebase:**

```
internal/
├── agency/
│   ├── models/          # Domain models (Agency, Specification, etc.)
│   ├── arangodb/        # Repository implementation (data access)
│   ├── services/        # Business logic (AgencyService, TagService, etc.)
│   └── validation/      # Input validation logic
├── agent/               # Agent domain (same structure)
├── workflow/            # Workflow domain (same structure)
└── policy/              # Policy domain (same structure)
```

**Dependency flow**: `Handlers → Services → Repositories → Database`

## Developer Workflows

### Build & Run Commands

```bash
# Full build and run (regenerates templates)
make run

# Kill any running instances first
make kill

# Development mode (with hot reload)
make run-dev

# Run with specific use case config
make run-water  # UC-INFRA-001 water distribution network

# Build only
make build

# Run tests
make test

# Clean build artifacts
make clean
```

**Important**: After editing `.templ` files, run `templ generate` or `make run` (which includes template generation).

### Frontend Asset Management

```bash
# Download self-hosted assets (HTMX, Alpine.js, Chart.js, Bulma CSS)
./scripts/download-assets.sh

# Verify assets are present
./scripts/verify-assets.sh

# Build Tailwind CSS (if using custom styles)
./bin/tailwindcss -i ./static/css/input.css -o ./static/css/tailwind.min.css --minify
```

**Hard refresh browser** (Ctrl+Shift+R / Cmd+Shift+R) after JavaScript changes to clear cache.

### Task Management Workflow

**Every task follows strict branch management:**

```bash
# 1. Create feature branch
git checkout -b feature/MVP-XXX_description

# 2. Implement changes
# ... development work ...

# 3. Build validation (before merge)
make build              # Must succeed
templ generate          # Must succeed
# Check for linting issues, unused code, deprecated APIs

# 4. Merge when complete
git checkout main
git merge feature/MVP-XXX_description --no-ff
git branch -d feature/MVP-XXX_description
```

**Task completion checklist**:
- [ ] Create coding session doc in `documents/3-SofwareDevelopment/coding_sessions/{TaskID}_{description}.md`
- [ ] Add to `documents/3-SofwareDevelopment/mvp_done.md`
- [ ] Remove from `documents/3-SofwareDevelopment/mvp.md`
- [ ] Merge feature branch to main

## Technology Stack & Integration Points

### Backend Stack
- **Language**: Go 1.21+ (native concurrency for agent coordination)
- **Web Framework**: Gin (HTTP routing)
- **Database**: ArangoDB (multi-model graph database)
- **Templates**: Templ (type-safe HTML templates, compile-time)
- **Frontend Interactivity**: HTMX (server-driven UI updates)

### Frontend Stack
- **CSS Framework**: Bulma 1.0.2 (prefer built-in classes over custom CSS)
- **JavaScript**: Alpine.js 3.13.3 (lightweight reactivity)
- **Icons**: FontAwesome 6.x
- **Charts**: Chart.js 4.4.1
- **Total bundle**: ~310KB (all self-hosted, see `static/` directory)

### External Integrations
- **Work Tracking**: Gitea (webhook-based issue sync for Kanban workflows)
- **AI Providers**: OpenAI, Claude, or custom (configurable via `config.yaml`)
- **A2A Protocol**: Multi-vendor agent interoperability (planned integration with `a2a-go` SDK)

## Code Quality Rules

### File Organization
- **Max file size**: 500-700 lines (split if exceeding)
- **Max function length**: 50 lines (prefer 20-30)
- **Principle**: One responsibility per file/function

### Naming Conventions
- **Package names**: Singular form (`agent`, not `agents`)
- **Collections**: Plural form in database (`agents`, `workflows`)
- **Services**: `{Domain}Service` (e.g., `AgencyService`, `TagService`)
- **Handlers**: `{Domain}Handler` (e.g., `AgencyHandler`, `WorkflowHandler`)
- **Repositories**: `Repository` in domain package

### Anti-Patterns to Avoid
- ❌ **Duplicate types** across packages (use `internal/shared/types/` for common enums)
- ❌ **Circular dependencies** between domain packages
- ❌ **Repository files >300 lines** (split by aggregate root)
- ❌ **Business logic in repositories** (belongs in services)
- ❌ **HTML generation in Go/JavaScript** (use `.templ` files)
- ❌ **JavaScript logic in `.templ` files** (use `static/js/` files)

## Configuration Management

**Main config**: `config.yaml` - Application-wide settings  
**Environment overrides**: `.env` files in use case directories

**Key config sections**:
```yaml
database:
  host: "localhost"
  port: 8529
  database: "codevaldcortex"  # Master database for agency metadata
  
ai:
  provider: "openai"  # openai, claude, local, or custom
  api_key: ""         # Use environment variable
  model: ""           # Uses provider default if empty
  
work_tracking:
  provider: "gitea"
  gitea_base_url: "https://gitea.example.com"
  gitea_api_token: ""  # Use environment variable
```

## Testing & Validation

**Run before committing**:
```bash
make test                 # Unit and integration tests
make test-coverage        # Generate coverage report
go vet ./...              # Static analysis
go fmt ./...              # Format code
templ generate            # Ensure templates compile
```

**Integration test setup**: Uses `codevaldcortex_test` database (auto-created)

## Common Patterns & Examples

### Creating a New Handler
```go
// internal/web/handlers/my_handler.go
type MyHandler struct {
    service     *myservice.Service
    agencyRepo  agency.Repository
    logger      *logrus.Logger
}

func NewMyHandler(service *myservice.Service, agencyRepo agency.Repository, logger *logrus.Logger) *MyHandler {
    return &MyHandler{
        service:    service,
        agencyRepo: agencyRepo,
        logger:     logger,
    }
}

func (h *MyHandler) HandleRequest(c *gin.Context) {
    // 1. Extract and validate input
    // 2. Call service layer
    // 3. Render template
    component := templates.MyTemplate(data)
    component.Render(c.Request.Context(), c.Writer)
}
```

### Agency-Scoped Operations
```go
// ALWAYS get agency-specific database for agency operations
func (h *Handler) GetAgencyData(c *gin.Context) {
    agencyID := c.Param("agencyID")
    
    // Get agency metadata from master DB
    agency, err := h.agencyService.GetAgency(ctx, agencyID)
    
    // Get agency-specific database
    db, err := h.dbClient.GetDatabase(ctx, agency.Database)
    
    // Create agency-scoped repository
    repo, err := registry.NewRepositoryWithDB(db)
    
    // Use repo for all agency operations
    agents, err := repo.List(ctx)
}
```

## Documentation References

**Architecture docs**: `documents/2-SoftwareDesignAndArchitecture/`
- `backend-architecture.md` - Microservices architecture
- `frontend-architecture-updated.md` - Templ/HTMX patterns
- `agency-operation-framework/` - Goals, work items, RACI concepts
- `a2a-protocol-integration.md` - Multi-vendor agent interoperability

**Development docs**: `documents/3-SofwareDevelopment/`
- `mvp.md` - Active tasks with priorities and dependencies
- `mvp_done.md` - Completed tasks archive
- `coding_sessions/` - Detailed implementation logs

**Requirements**: `documents/1-SoftwareRequirements/`

## Quick Reference: Key Directories

```
/workspaces/CodeValdCortex/
├── cmd/main.go                    # Application entry point
├── config.yaml                    # Main configuration
├── Makefile                       # Build commands
├── internal/
│   ├── app/app.go                 # Application bootstrap
│   ├── agency/                    # Agency domain (multi-tenant core)
│   ├── agent/                     # Agent domain
│   ├── workflow/                  # Workflow orchestration
│   ├── policy/                    # AI policy layer (governance)
│   ├── database/                  # ArangoDB client
│   ├── web/
│   │   ├── handlers/              # HTTP request handlers
│   │   ├── pages/                 # Templ page templates
│   │   └── components/            # Reusable Templ components
│   └── infrastructure/
│       └── gitea/                 # Gitea integration (work tracking)
├── static/
│   ├── css/                       # Bulma CSS + custom styles
│   ├── js/                        # JavaScript (Alpine.js, HTMX, Chart.js)
│   └── img/                       # Static images
├── documents/                     # Comprehensive documentation
└── usecases/                      # Use case configurations
    ├── UC-INFRA-001-water-distribution-network/
    ├── UC-CHAR-001-tumaini/
    └── ...
```

## When in Doubt

1. **Check existing patterns**: Search codebase for similar implementations
2. **Follow DRY principle**: Reuse aggressively to prevent repository bloat
3. **Consult architecture docs**: `documents/2-SoftwareDesignAndArchitecture/`
4. **Review coding sessions**: `documents/3-SofwareDevelopment/coding_sessions/`
5. **Ask before duplicating**: Especially for layouts, components, and types


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

## Component Reuse and DRY Principles

**🚨 CRITICAL: Maximize code reuse to prevent repository bloat.**

- **ALWAYS check for existing components before creating new ones**
- **NEVER duplicate layouts, navbars, or page structures** - use shared components
- **Repository size is a concern** - reuse aggressively to keep codebase maintainable
- Before creating ANY new template file, search for similar existing components
- Use `@components.LayoutWithAgency()` for all agency-scoped pages
- Use shared components from `internal/web/components/` whenever possible
- Extract repeated markup patterns into reusable components
- Benefits:
  - Smaller repository size
  - Consistent UI/UX across all pages
  - Single source of truth for layouts
  - Easier maintenance and bug fixes
  - Faster development

**Example of WRONG approach:**
```templ
// ❌ NEVER DO THIS - Duplicating entire page structure
templ MyPage(agency *models.Agency) {
    <!DOCTYPE html>
    <html>
        <head>
            <meta charset="UTF-8"/>
            <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.2/css/bulma.min.css"/>
            <!-- ... duplicate all CSS/JS includes ... -->
        </head>
        <body>
            @components.NavbarWithAgency(agency)  // Even if using shared navbar, structure is duplicated!
            <section class="section">
                <!-- page content -->
            </section>
        </body>
    </html>
}
```

**Example of CORRECT approach:**
```templ
// ✅ CORRECT - Use shared layout component
templ MyPage(agency *models.Agency) {
    @components.LayoutWithAgency("My Page Title", agency) {
        <section class="section">
            <!-- page content only -->
        </section>
    }
}
```

**Checklist before creating new templates:**
1. ✅ Does a similar component already exist in `internal/web/components/`?
2. ✅ Can I use `@components.LayoutWithAgency()` instead of building full HTML?
3. ✅ Is there a page template I can reference in `internal/web/pages/`?
4. ✅ Can I extract repeated markup into a new shared component?
5. ✅ Am I about to duplicate any HTML structure that exists elsewhere?

## Template-First Architecture

**Always prefer `.templ` files over Go/JavaScript for HTML generation.**

- HTML markup should be defined in `.templ` files (using the templ template engine)
- **NEVER generate HTML strings in Go handler files** - use `.templ` files instead
- **NEVER generate HTML strings in JavaScript files** - use `.templ` files or server-side rendering
- **NEVER put JavaScript logic in `.templ` files** - keep `.templ` files for HTML structure only
- **NEVER duplicate page layouts** - always use shared layout components
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

**🚨 CRITICAL: Keep files small and focused. Break down large files immediately.**

- **MAXIMUM FILE SIZE: 500-700 lines** (HARD LIMIT - NO EXCEPTIONS)
- **If a file exceeds 700 lines, it MUST be broken down into smaller, focused modules IMMEDIATELY**
- **Each file MUST have a single, clear responsibility**
- **Handler files MUST be split by domain/feature area**
- **Service files MUST be modular and composable**
- **⚠️ WARNING: We currently have several non-compliant files that violate these rules**
  - These MUST be refactored as they are encountered
  - Do NOT create new files that violate these limits
  - Do NOT add to existing files that already exceed limits
- **Enforcement**: Check line count before editing. If file is >600 lines, split it FIRST before adding new code

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

