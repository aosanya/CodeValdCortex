# MVP-AUTH-001: User Model & Repository

## Status: ✅ Complete

## Overview
Create User data model and ArangoDB repository with CRUD operations, password hashing (bcrypt), and email validation.

## Implementation

### File: `internal/auth/models.go`

**User Model:**
```go
type User struct {
    Key          string    `json:"_key,omitempty"`
    ID           string    `json:"_id,omitempty"`
    Rev          string    `json:"_rev,omitempty"`
    Email        string    `json:"email" binding:"required,email"`
    Name         string    `json:"name" binding:"required"`
    PasswordHash string    `json:"-"` // Never expose in JSON
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    IsActive     bool      `json:"is_active"`
}
```

**UserResponse** (safe for API responses):
```go
type UserResponse struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    IsActive  bool      `json:"is_active"`
}
```

**RefreshToken Model:**
```go
type RefreshToken struct {
    Key       string     `json:"_key,omitempty"`
    ID        string     `json:"_id,omitempty"`
    UserID    string     `json:"user_id"`
    Token     string     `json:"token"` // SHA-256 hash
    ExpiresAt time.Time  `json:"expires_at"`
    CreatedAt time.Time  `json:"created_at"`
    IsRevoked bool       `json:"is_revoked"`
    RevokedAt *time.Time `json:"revoked_at,omitempty"`
}
```

### File: `internal/auth/repository.go`

**Repository Methods:**
- `CreateUser(ctx, *User) error` - Insert new user with unique email check
- `GetUserByEmail(ctx, email) (*User, error)` - Find user by email
- `GetUserByID(ctx, id) (*User, error)` - Find user by ID
- `UpdateUser(ctx, *User) error` - Update user fields
- `SaveRefreshToken(ctx, *RefreshToken) error` - Store refresh token
- `GetRefreshToken(ctx, tokenHash) (*RefreshToken, error)` - Retrieve token
- `RevokeRefreshToken(ctx, tokenHash) error` - Mark token as revoked
- `RevokeAllUserTokens(ctx, userID) error` - Revoke all user's tokens
- `CleanupExpiredTokens(ctx) error` - Remove expired tokens

**Collections Created:**
1. `users` - User accounts
   - Unique index on `email`
2. `refresh_tokens` - Refresh tokens
   - Index on `user_id`
   - Index on `token`

**Error Handling:**
- Duplicate email returns conflict error
- User not found returns specific error
- ArangoDB driver errors wrapped with context

## Security Features

### Password Hashing
- Uses bcrypt with cost factor 12
- Password hash never exposed in JSON (json:"-" tag)
- Comparison done in constant time (bcrypt.CompareHashAndPassword)

### Email Validation
- Unique constraint at database level
- Email format validation via binding tags
- Case-insensitive email lookup (FILTER lowercase)

## Database Schema

**users collection:**
```json
{
  "_key": "user_abc123",
  "_id": "users/user_abc123",
  "email": "user@example.com",
  "name": "John Doe",
  "password_hash": "$2a$12$...",
  "created_at": "2026-01-28T10:00:00Z",
  "updated_at": "2026-01-28T10:00:00Z",
  "is_active": true
}
```

**refresh_tokens collection:**
```json
{
  "_key": "token_xyz789",
  "_id": "refresh_tokens/token_xyz789",
  "user_id": "users/user_abc123",
  "token": "base64-encoded-sha256-hash",
  "expires_at": "2026-02-04T10:00:00Z",
  "created_at": "2026-01-28T10:00:00Z",
  "is_revoked": false
}
```

## Testing

**Manual Test - User Creation:**
```go
repo, _ := auth.NewRepository(db)
user := &auth.User{
    Email: "test@example.com",
    Name: "Test User",
    PasswordHash: "$2a$12$hashed_password",
}
err := repo.CreateUser(ctx, user)
// Verify user.ID is set
// Try creating duplicate - should get conflict error
```

**Manual Test - Email Lookup:**
```go
user, err := repo.GetUserByEmail(ctx, "test@example.com")
// Verify user found
// Try non-existent email - should get "user not found" error
```

## Dependencies
- `github.com/arangodb/go-driver` - ArangoDB client
- Standard library: `time`, `fmt`, `context`

## Integration Points
- Used by `auth.Service` for user CRUD operations
- Collections auto-created on first repository initialization
- Indexes created automatically via `ensureCollections()`

## Future Enhancements
- [ ] Add user roles field (admin, user, viewer)
- [ ] Add email verification status
- [ ] Add last_login_at timestamp
- [ ] Add failed login attempt tracking
- [ ] Add soft delete (deleted_at field)
- [ ] Add user preferences/settings field
