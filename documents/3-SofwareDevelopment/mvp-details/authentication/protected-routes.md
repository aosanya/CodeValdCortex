# MVP-AUTH-005: Protected Routes Integration

## Status: 📋 Not Started

## Overview
Apply authentication middleware to protected routes and update handlers to use real user context instead of hardcoded "system" user.

## Tasks

### 1. Apply Middleware to Protected Routes

**File**: Create `internal/app/routes_protected.go` (or update existing route files)

```go
// Apply auth middleware to protected route groups
protected := apiV1.Group("")
protected.Use(api.AuthMiddleware(a.authService))
{
    // Agency routes (require authentication)
    protected.POST("/agencies", agencyHandler.CreateAgency)
    protected.PUT("/agencies/:id", agencyHandler.UpdateAgency)
    protected.DELETE("/agencies/:id", agencyHandler.DeleteAgency)
    
    // Instance routes (require authentication)
    protected.POST("/agencies/:id/instances", instanceHandler.StartInstance)
    protected.DELETE("/agencies/:id/instances/:instance_id", instanceHandler.DeleteInstance)
    
    // Work item routes (require authentication)
    protected.POST("/agencies/:id/work-items", workItemHandler.CreateWorkItem)
    // ... etc
}
```

### 2. Update Handlers to Use User Context

**Pattern to Replace:**

**Before:**
```go
CreatedBy: "system", // TODO: Get from auth context
```

**After:**
```go
userID, exists := c.Get("user_id")
if !exists {
    c.JSON(401, gin.H{"error": "Unauthorized"})
    return
}
CreatedBy: userID.(string),
```

**Files to Update:**

#### `internal/handlers/agency_handler.go`
- Line 114: `CreateAgency` - Replace `CreatedBy: "system"`

#### `internal/handlers/instance_handler.go`
- Line 45: `StartInstance` - Replace metadata["created_by"] = "system"

#### `internal/handlers/workflow_handler.go`
- Line 43: `CreateWorkflow` - Replace `req.CreatedBy = "system"`

#### `internal/handlers/tag_handler.go`
- Lines 43-44: Already has context check, but uses fmt.Sprintf - verify

#### `internal/web/handlers/workbench_handler.go`
- Line 187: `CreateIssue` - Replace `createdBy := "system"`

### 3. Route Access Control Matrix

| Route | Method | Protected | User Context Required |
|-------|--------|-----------|----------------------|
| `/api/v1/auth/*` | * | No | No |
| `/api/v1/agencies` | GET | Maybe | Optional (for filtering) |
| `/api/v1/agencies` | POST | Yes | Required |
| `/api/v1/agencies/:id` | GET | Maybe | Optional |
| `/api/v1/agencies/:id` | PUT | Yes | Required (+ owner check) |
| `/api/v1/agencies/:id` | DELETE | Yes | Required (+ owner check) |
| `/api/v1/agencies/:id/instances` | POST | Yes | Required |
| `/api/v1/agencies/:id/instances/:id` | DELETE | Yes | Required |
| `/api/v1/agencies/:id/work-items` | * | Yes | Required |
| `/api/v1/workbench/*` | * | Yes | Required |

### 4. Permission Checking (Future)

**Add Permission Helper:**

```go
// File: internal/auth/permissions.go
package auth

func CanModifyAgency(userID, agencyID string, repo Repository) bool {
    agency, err := repo.GetAgency(agencyID)
    if err != nil {
        return false
    }
    return agency.CreatedBy == userID
}
```

**Use in Handler:**
```go
if !auth.CanModifyAgency(userID.(string), agencyID, h.repo) {
    c.JSON(403, gin.H{"error": "Forbidden"})
    return
}
```

### 5. Public vs Protected Routes

**Public Routes (No Auth Required):**
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout` (optional - can work without auth)
- `GET /health`
- `GET /api/v1/status`

**Optional Auth Routes (Better experience if authenticated):**
- `GET /api/v1/agencies` - Could filter to user's agencies if authenticated
- `GET /api/v1/agencies/:id` - Could show private info if owner

**Protected Routes (Auth Required):**
- All POST/PUT/DELETE operations
- All user-specific data endpoints
- All instance/work-item management

### 6. Testing Checklist

**Manual Testing:**

```bash
# 1. Register and login to get tokens
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.access_token')

# 2. Test protected route WITHOUT token (should fail)
curl -X POST http://localhost:8080/api/v1/agencies \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Agency"}'
# Expected: 401 Unauthorized

# 3. Test protected route WITH token (should succeed)
curl -X POST http://localhost:8080/api/v1/agencies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Agency","category":"other"}'
# Expected: 201 Created

# 4. Verify created_by is set to user ID (not "system")
curl -X GET http://localhost:8080/api/v1/agencies/AGENCY_ID \
  -H "Authorization: Bearer $TOKEN"
# Expected: created_by should be "users/user_abc123" (not "system")

# 5. Test with expired token (wait 15 minutes or use old token)
curl -X POST http://localhost:8080/api/v1/agencies \
  -H "Authorization: Bearer $EXPIRED_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Agency"}'
# Expected: 401 with "Invalid or expired token"
```

### 7. Database Migration

**No migration needed** - existing records with `created_by: "system"` can remain.

**Optional cleanup:**
```aql
// Update old records to have a system user
FOR doc IN agencies
  FILTER doc.created_by == "system"
  UPDATE doc WITH { created_by: "users/system" } IN agencies
```

## Dependencies
- MVP-AUTH-004 (AuthMiddleware) must be complete
- All handlers must have access to Gin context

## Implementation Steps

1. **Create route grouping** - Separate public and protected routes
2. **Apply middleware** - Add AuthMiddleware to protected group
3. **Update handlers** - Replace all "system" hardcoding with context extraction
4. **Test each route** - Verify auth works and user context is correct
5. **Update tests** - Add auth headers to integration tests
6. **Document** - Update API docs with auth requirements

## Security Notes

### Don't Forget
- [ ] All mutation operations (POST/PUT/DELETE) should be protected
- [ ] User context should be used for audit trails
- [ ] Validate user permissions before operations
- [ ] Log all authenticated actions with user ID

### Common Mistakes
- ❌ Forgetting to check `exists` before type assertion
- ❌ Not handling missing user context gracefully
- ❌ Applying middleware to public routes (breaks registration/login)
- ❌ Not updating all handlers that create/modify data

## Future Enhancements
- [ ] Role-based access control (admin, user, viewer)
- [ ] Resource ownership verification
- [ ] Permission-based middleware (e.g., RequireAdmin())
- [ ] API key authentication for service accounts
- [ ] Audit log of all authenticated actions
