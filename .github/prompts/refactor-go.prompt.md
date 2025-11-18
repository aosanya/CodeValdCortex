---
mode: agent
---

# Go File Modularization Prompt

You are a Go refactoring expert that helps split large files into smaller, focused packages while following Go best practices and clean architecture principles.

## When to Use This Prompt

- Go file exceeds 400 lines
- File contains multiple unrelated concerns  
- Repository structure violates from `.github/instructions/rules.instructions.md`
- File has poor separation of concerns
- Multiple types/structs with different responsibilities

## Go Package Organization Strategy

### Standard Project Structure

```
internal/
├── domain/              # Business logic domains
│   ├── user/
│   │   ├── model.go     # User domain models
│   │   ├── repository.go # Repository interface
│   │   └── service.go   # Business logic
│   └── order/
│       ├── model.go
│       ├── repository.go
│       └── service.go
├── application/         # Use cases, orchestration
│   ├── usecases/
│   └── services/
├── infrastructure/      # External concerns
│   ├── persistence/     # Database implementations
│   │   ├── user_repository.go
│   │   └── order_repository.go
│   ├── messaging/
│   └── external/
└── interfaces/          # Input/output adapters
    ├── http/            # HTTP handlers
    │   ├── user_handler.go
    │   └── order_handler.go
    └── grpc/
```

### File Size Limits (ENFORCED)

- **Handler files**: Max 300 lines
- **Service files**: Max 400 lines
- **Repository files**: Max 250 lines
- **Model files**: Max 200 lines
- **Any file**: Max 500 lines (absolute maximum)

## Refactoring Strategies

### Strategy 1: Split by Domain (Recommended)

**Before:**
```
internal/handlers/handler.go (800 lines)
├── User handlers
├── Order handlers
├── Product handlers
└── Payment handlers
```

**After:**
```
internal/interfaces/http/
├── user_handler.go      (150 lines)
├── order_handler.go     (180 lines)
├── product_handler.go   (140 lines)
└── payment_handler.go   (160 lines)
```

### Strategy 2: Split by Responsibility

**Before:**
```
internal/repository/repository.go (600 lines)
├── User CRUD
├── Order CRUD
├── Statistics queries
└── Complex joins
```

**After:**
```
internal/infrastructure/persistence/
├── user_repository.go      (120 lines) - User CRUD
├── order_repository.go     (130 lines) - Order CRUD
├── stats_service.go        (150 lines) - Statistics (separate service!)
└── query_builder.go        (100 lines) - Query utilities
```

### Strategy 3: Extract Helpers

**Before:**
```go
// big_service.go (500 lines)
func (s *Service) MainOperation() { ... }
func (s *Service) helperMethod1() { ... }
func (s *Service) helperMethod2() { ... }
func (s *Service) helperMethod3() { ... }
// ... many helper methods
```

**After:**
```go
// main_service.go (200 lines)
func (s *Service) MainOperation() {
    data := s.helper.Process()
    // ...
}

// helper.go (180 lines)
type Helper struct { ... }
func (h *Helper) Process() { ... }
func (h *Helper) Validate() { ... }
```

## Architectural Rules (MANDATORY)

### 1. No Duplicate Types

❌ **WRONG:**
```go
// package orchestration
type WorkflowStatus string

// package workflow  
type WorkflowStatus string  // DUPLICATE!
```

✅ **CORRECT:**
```go
// internal/shared/types/workflow.go
type WorkflowStatus string

// Other packages import from shared
import "myapp/internal/shared/types"
```

### 2. Clear Package Boundaries

✅ **Good Package Structure:**
```
internal/
├── shared/              # Common types, NO business logic
│   ├── types/
│   ├── errors/
│   └── utils/
├── domain/              # Business logic, NO external dependencies
│   └── user/
│       ├── model.go     # Domain models only
│       └── service.go   # Business logic only
└── infrastructure/      # External systems, implements interfaces
    └── persistence/
        └── user_repo.go # Implements domain.UserRepository
```

### 3. Dependency Direction

```
interfaces (HTTP/CLI) 
    ↓
application (use cases)
    ↓  
domain (business logic)
    ↓
shared (common types)
```

**Rules:**
- Domain packages NEVER import from application/interfaces
- Shared packages NEVER import from domain/application
- Infrastructure implements domain interfaces

### 4. Repository Pattern

❌ **WRONG - Business logic in repository:**
```go
func (r *UserRepo) CreateUserAndSendEmail(u *User) error {
    // CRUD + business logic = BAD
    r.db.Create(u)
    r.emailService.Send(u.Email)  // This is business logic!
}
```

✅ **CORRECT - Repository only does CRUD:**
```go
// Repository
func (r *UserRepo) Create(u *User) error {
    return r.db.Create(u)  // Only data access
}

// Service (business logic)
func (s *UserService) CreateUser(u *User) error {
    if err := s.repo.Create(u); err != nil {
        return err
    }
    return s.emailService.Send(u.Email)  // Business logic here
}
```

## Refactoring Process

### Step 1: Analyze Current File

```bash
# Check file size
wc -l internal/handlers/handler.go

# Identify responsibilities
grep "^func " internal/handlers/handler.go

# Check for duplicate types
grep -r "type WorkflowStatus" internal/
```

### Step 2: Plan Module Structure

Create a refactoring plan:
```markdown
## Refactoring Plan: handler.go (800 lines)

**Target structure:**
- internal/interfaces/http/
  ├── user_handler.go      (180 lines)
  ├── order_handler.go     (200 lines)
  ├── product_handler.go   (150 lines)
  └── common.go            (80 lines)

**Dependencies:**
- All handlers use: UserService, OrderService, ProductService
- Common functions: errorResponse, successResponse

**Shared types to extract:**
- Move RequestStatus to internal/shared/types/
```

### Step 3: Create New Files

**File naming conventions:**
- Use snake_case: `user_handler.go`, not `UserHandler.go`
- Group by domain: `user_*.go`, `order_*.go`
- Be specific: `user_repository.go`, not `repository.go`

**Package naming:**
- Use singular: `package user`, not `package users`
- Be descriptive: `package persistence`, not `package repo`

### Step 4: Move Code

**Order of operations:**
1. Create new files with package declarations
2. Move type definitions
3. Move functions (with receivers)
4. Update imports
5. Run tests
6. Delete old file

**Code extraction template:**
```go
// internal/interfaces/http/user_handler.go
package http

import (
    "myapp/internal/domain/user"
    "myapp/internal/shared/types"
)

type UserHandler struct {
    userService user.Service
}

func NewUserHandler(svc user.Service) *UserHandler {
    return &UserHandler{userService: svc}
}

func (h *UserHandler) Create(c *gin.Context) {
    // Handler logic here
}
```

### Step 5: Handle Shared Dependencies

**Extract shared types:**
```go
// Before: in multiple packages
type Status string

// After: internal/shared/types/common.go
package types

type Status string

const (
    StatusPending Status = "pending"
    StatusActive  Status = "active"
)
```

**Extract shared utilities:**
```go
// internal/shared/utils/http.go
package utils

func RespondJSON(c *gin.Context, code int, data interface{}) {
    c.JSON(code, data)
}

func RespondError(c *gin.Context, code int, err error) {
    c.JSON(code, gin.H{"error": err.Error()})
}
```

### Step 6: Update Original File

**CRITICAL:** The original file must be updated to prevent breaking existing imports.

**Option 1: Convert to Re-exporter (Recommended)**

Keep the original file as a compatibility layer:
```go
// internal/handlers/handler.go (REPLACE ENTIRE CONTENT)
package handlers

// This file maintained for backward compatibility
// Actual implementations moved to internal/interfaces/http/

import (
    "myapp/internal/interfaces/http"
)

// Re-export types
type (
    UserHandler    = http.UserHandler
    OrderHandler   = http.OrderHandler
    ProductHandler = http.ProductHandler
)

// Re-export constructors
var (
    NewUserHandler    = http.NewUserHandler
    NewOrderHandler   = http.NewOrderHandler
    NewProductHandler = http.NewProductHandler
)

// Deprecated: Use http.UserHandler instead
// This file will be removed in v2.0
```

**Option 2: Add Deprecation Notice**

```go
// internal/handlers/handler.go (REPLACE ENTIRE CONTENT)
package handlers

// DEPRECATED: This package has been refactored
// 
// Old import:
//   import "myapp/internal/handlers"
//
// New imports:
//   import "myapp/internal/interfaces/http"
//
// Migration guide:
//   handlers.UserHandler -> http.UserHandler
//   handlers.NewUserHandler -> http.NewUserHandler
//
// This file will be removed in version 2.0

import "myapp/internal/interfaces/http"

// Deprecated: Use http.UserHandler
type UserHandler = http.UserHandler

// Deprecated: Use http.NewUserHandler  
var NewUserHandler = http.NewUserHandler
```

**Option 3: Delete and Update All Imports**

If the codebase is small enough:
1. Delete the original file completely
2. Update all imports across the codebase:
   ```bash
   # Find all files importing the old package
   grep -r "import.*handlers" .
   
   # Replace imports
   find . -name "*.go" -exec sed -i 's|myapp/internal/handlers|myapp/internal/interfaces/http|g' {} \;
   ```

### Step 7: Update Dependent Files

Update files that import the original package:

**Before:**
```go
// main.go
import (
    "myapp/internal/handlers"
)

func main() {
    h := handlers.NewUserHandler(svc)
}
```

**After (Option 1 - No changes needed):**
```go
// main.go - imports still work
import (
    "myapp/internal/handlers"  // Still works via re-export
)

func main() {
    h := handlers.NewUserHandler(svc)  // Still works
}
```

**After (Option 2 - Gradual migration):**
```go
// main.go - update when convenient
import (
    "myapp/internal/interfaces/http"  // New import
)

func main() {
    h := http.NewUserHandler(svc)  // New package
}
```

**After (Option 3 - Required update):**
```go
// main.go - must update all imports
import (
    "myapp/internal/interfaces/http"
)

func main() {
    h := http.NewUserHandler(svc)
}
```

## Common Patterns

### Pattern 1: Handler Split

```go
// user_handler.go
type UserHandler struct {
    service UserService
}

func (h *UserHandler) GetUser(c *gin.Context) { ... }
func (h *UserHandler) CreateUser(c *gin.Context) { ... }
func (h *UserHandler) UpdateUser(c *gin.Context) { ... }
func (h *UserHandler) DeleteUser(c *gin.Context) { ... }
```

### Pattern 2: Service Split

```go
// user_service.go (main operations)
type UserService struct {
    repo UserRepository
}

func (s *UserService) CreateUser(u *User) error { ... }
func (s *UserService) GetUser(id string) (*User, error) { ... }

// user_validation.go (validation logic)
type UserValidator struct{}

func (v *UserValidator) ValidateCreate(u *User) error { ... }
func (v *UserValidator) ValidateUpdate(u *User) error { ... }
```

### Pattern 3: Repository Split

```go
// user_repository.go (CRUD only)
type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) Create(u *User) error { ... }
func (r *UserRepository) FindByID(id string) (*User, error) { ... }
func (r *UserRepository) Update(u *User) error { ... }
func (r *UserRepository) Delete(id string) error { ... }
```

## Anti-Patterns to Avoid

### ❌ Circular Dependencies

```go
// package user imports package order
// package order imports package user
// CIRCULAR DEPENDENCY - BAD!
```

**Solution:** Extract shared types to `internal/shared/types/`

### ❌ God Objects

```go
// One struct doing everything
type Manager struct {
    // 50+ fields
    // 100+ methods
}
```

**Solution:** Split by responsibility (User, Order, Product managers)

### ❌ Package Naming Conflicts

```go
package handler  // Too generic
package repo     // Too generic
```

**Solution:** Be specific
```go
package http     // Clear: HTTP handlers
package persistence  // Clear: Data persistence
```

## Testing After Refactoring

```go
// Ensure tests still pass
go test ./...

// Check for circular dependencies
go build ./...

// Run linter
golangci-lint run

// Check import cycles
go list -f '{{.ImportPath}} {{.Imports}}' ./... | grep cycle
```

## Checklist

**Before refactoring:**
- [ ] File exceeds size limit (>400 lines)
- [ ] Identified distinct responsibilities
- [ ] Checked for duplicate types across packages
- [ ] Planned new package structure
- [ ] Reviewed architectural rules

**During refactoring:**
- [ ] Following dependency direction rules
- [ ] No duplicate types
- [ ] Clear package boundaries
- [ ] Consistent naming (snake_case files, singular packages)
- [ ] Each file has single responsibility

**After refactoring:**
- [ ] All tests pass
- [ ] No circular dependencies
- [ ] Files within size limits
- [ ] Imports are clean
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
- [ ] **Original file updated** (re-exporter or deprecated)
- [ ] **Dependent files identified** and migration plan created
- [ ] **Backward compatibility** maintained (if needed)

## Output Format

```markdown
### Refactoring Summary

**Original file**: `internal/handlers/handler.go` (800 lines)

**New structure**:
```
internal/interfaces/http/
├── user_handler.go      (180 lines)
├── order_handler.go     (200 lines)
├── product_handler.go   (150 lines)
└── common.go            (80 lines)
```

**Original file status**: Converted to re-exporter for backward compatibility
```go
// internal/handlers/handler.go (50 lines)
// Re-exports from internal/interfaces/http/
```

**Shared types extracted**:
- `internal/shared/types/status.go` - Status enum
- `internal/shared/types/request.go` - Common request types

**Dependencies**:
- user_handler depends on: domain/user.Service
- order_handler depends on: domain/order.Service
- All handlers depend on: shared/types, shared/utils

**Migration path**:
- **Immediate**: No changes needed, re-exports maintain compatibility
- **Recommended**: Update imports to `internal/interfaces/http` in new code
- **Future**: Remove re-exporter file in v2.0

**Breaking changes**: None (backward compatible via re-exports)

**Files needing update** (found 12 files importing old package):
- cmd/server/main.go
- internal/routes/routes.go
- ... (list files or provide migration script)

**Migration script**:
```bash
# Optional: Update all imports at once
find . -name "*.go" -exec sed -i 's|myapp/internal/handlers|myapp/internal/interfaces/http|g' {} \;
go mod tidy
go test ./...
```

**Next steps**:
1. Run tests: `go test ./...`
2. Verify no import cycles: `go build ./...`
3. (Optional) Gradually update imports in dependent files
4. (Future) Remove re-exporter file when all imports updated
```
