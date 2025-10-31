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
- JavaScript should only handle:
  - Event handling
  - Data fetching
  - DOM manipulation (show/hide, updates)
  - State management
- Go handlers should:
  - Process business logic
  - Call services
  - Return data structures (JSON) or render `.templ` templates
  - Never contain HTML strings or fmt.Sprintf with HTML
- Pre-render all content sections in templates, then toggle visibility with JavaScript
- Benefits:
  - Type safety
  - Server-side rendering capability
  - Better maintainability
  - Separation of concerns
  - Easier testing

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

