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
