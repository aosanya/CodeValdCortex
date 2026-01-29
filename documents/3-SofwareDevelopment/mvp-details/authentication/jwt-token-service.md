# MVP-AUTH-002: JWT Token Service

## Status: ✅ Complete

## Overview
Implement JWT token generation/validation service with access tokens (15min expiry), refresh tokens (7 day expiry), and secure token management.

## Implementation

### File: `internal/auth/service.go`

**Service Structure:**
```go
type Service struct {
    repo      *Repository
    jwtSecret []byte
}

const (
    AccessTokenDuration  = 15 * time.Minute  // 15 minutes
    RefreshTokenDuration = 7 * 24 * time.Hour // 7 days
    BcryptCost          = 12
)
```

**Core Methods:**

#### 1. User Registration
```go
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*User, error)
```
- Validates email format and password length
- Hashes password with bcrypt (cost 12)
- Creates user in database
- Returns User (without password hash)

#### 2. User Login
```go
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
```
- Looks up user by email
- Verifies password with bcrypt
- Checks user is_active status
- Generates access token (JWT)
- Generates and stores refresh token
- Returns both tokens + user info

#### 3. Token Refresh
```go
func (s *Service) RefreshAccessToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
```
- Validates refresh token exists and not expired
- Checks user is still active
- Generates new access token
- Returns new access token (reuses refresh token)

#### 4. Logout
```go
func (s *Service) Logout(ctx context.Context, refreshToken string) error
```
- Marks refresh token as revoked in database
- Prevents further use of that token

#### 5. JWT Validation
```go
func (s *Service) ValidateAccessToken(tokenString string) (*TokenClaims, error)
```
- Parses JWT token
- Validates signature with secret
- Checks expiration
- Returns user claims (user_id, email, name)

## Token Generation Details

### Access Token (JWT)
**Algorithm**: HMAC-SHA256 (HS256)

**Claims:**
```json
{
  "user_id": "users/user_abc123",
  "email": "user@example.com",
  "name": "John Doe",
  "iat": 1706438400,  // Issued at (Unix timestamp)
  "exp": 1706439300   // Expires at (iat + 15 minutes)
}
```

**Example Token:**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlcnMvdXNlcl9hYmMxMjMiLCJlbWFpbCI6InVzZXJAZXhhbXBsZS5jb20iLCJuYW1lIjoiSm9obiBEb2UiLCJpYXQiOjE3MDY0Mzg0MDAsImV4cCI6MTcwNjQzOTMwMH0.signature
```

### Refresh Token
**Format**: Random 32-byte value, base64 URL-encoded

**Storage**: SHA-256 hash stored in database (not raw token)

**Security Rationale:**
- Even if database is compromised, attacker cannot use hashed tokens
- Only valid tokens can be matched via hash comparison
- Tokens are revocable (database lookup required)

## Security Features

### Password Security
- Bcrypt hashing with cost factor 12
- Constant-time comparison
- Password hash never returned from service methods

### Token Security
- Access tokens short-lived (15 minutes)
- Refresh tokens long-lived but revocable
- Refresh tokens stored as hashes
- JWT signed with HMAC-SHA256
- Signature validation on every request
- Expiration checked automatically

### Error Messages
- Generic "invalid credentials" on login failure (prevents email enumeration)
- Separate error for disabled accounts
- Token validation returns specific error types

## Authentication Response

**AuthResponse structure:**
```go
type AuthResponse struct {
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    TokenType    string       `json:"token_type"`    // "Bearer"
    ExpiresIn    int          `json:"expires_in"`    // 900 (seconds)
    User         UserResponse `json:"user"`
}
```

**Example JSON:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "xK9mP2nQ4rS6tU8vW0yA1bC3dE5fG7hI9jK",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "users/user_abc123",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2026-01-28T10:00:00Z",
    "updated_at": "2026-01-28T10:00:00Z",
    "is_active": true
  }
}
```

## Configuration

**Required:**
- JWT secret configured in `config.yaml` or environment variable
- Minimum 32 characters recommended for production

**config.yaml:**
```yaml
auth:
  jwt_secret: "your-production-secret-key-min-32-chars"
```

**Environment Variable (overrides config):**
```bash
export JWT_SECRET="your-production-secret-key-min-32-chars"
```

## Token Lifecycle

### Registration Flow
1. User submits registration form
2. Service hashes password
3. User created in database
4. Return success (no tokens issued yet)
5. User must login to get tokens

### Login Flow
1. User submits email + password
2. Service validates credentials
3. Service generates access token (JWT)
4. Service generates refresh token (random bytes)
5. Service stores hashed refresh token in database
6. Return both tokens + user info to client

### Token Refresh Flow
1. Client detects access token expired (or preemptively refreshes)
2. Client sends refresh token to `/api/v1/auth/refresh`
3. Service validates refresh token (hash lookup, expiration check)
4. Service generates new access token
5. Service returns new access token (same refresh token)

### Logout Flow
1. Client sends refresh token to `/api/v1/auth/logout`
2. Service marks token as revoked in database
3. Future refresh attempts with that token fail

## Testing

**Manual Test - Registration:**
```go
service := auth.NewService(repo, "test-secret")
user, err := service.Register(ctx, &auth.RegisterRequest{
    Email:    "test@example.com",
    Name:     "Test User",
    Password: "password123",
})
// Verify user created, password not returned
```

**Manual Test - Login:**
```go
authResp, err := service.Login(ctx, &auth.LoginRequest{
    Email:    "test@example.com",
    Password: "password123",
})
// Verify access_token and refresh_token present
// Verify expires_in = 900
```

**Manual Test - Token Validation:**
```go
claims, err := service.ValidateAccessToken(authResp.AccessToken)
// Verify user_id, email, name extracted correctly
// Try with expired token - should fail
// Try with tampered token - should fail
```

**Manual Test - Token Refresh:**
```go
newAuth, err := service.RefreshAccessToken(ctx, authResp.RefreshToken)
// Verify new access_token issued
// Old access token still valid until expiry
```

## Dependencies
- `github.com/golang-jwt/jwt/v5` - JWT library
- `golang.org/x/crypto/bcrypt` - Password hashing
- `crypto/rand`, `crypto/sha256` - Token generation and hashing
- `encoding/base64` - Token encoding

## Integration Points
- Used by `auth.Handler` for all auth operations
- Requires `auth.Repository` for user/token persistence
- JWT secret from `config.Config.Auth.JWTSecret`

## Future Enhancements
- [ ] Token rotation (issue new refresh token on each refresh)
- [ ] Configurable token expiration times
- [ ] Support for multiple concurrent sessions per user
- [ ] Token blacklist for immediate revocation
- [ ] Refresh token family tracking (detect token reuse attacks)
- [ ] JWT key rotation support
- [ ] Support for RSA/ECDSA signing (public key verification)
