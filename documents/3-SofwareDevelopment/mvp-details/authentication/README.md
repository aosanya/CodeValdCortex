# Authentication & User Management - Domain Overview

## Purpose
Implement complete user authentication system to support the Flutter frontend (CodeValdFortex). Provides JWT-based authentication with dual-token strategy (access + refresh tokens), user registration/login, and protected route middleware.

## Architecture Summary

### Components
```
┌─────────────────────────────────────────────────────────────┐
│                    Authentication Flow                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Frontend (Flutter)                                          │
│       │                                                      │
│       │ POST /api/v1/auth/register                          │
│       │ POST /api/v1/auth/login                             │
│       ▼                                                      │
│  ┌──────────────┐                                           │
│  │ Auth Handler │──────► Auth Service                       │
│  └──────────────┘              │                            │
│                                 ├─► User Repository         │
│                                 │   (ArangoDB)              │
│                                 │                            │
│                                 ├─► JWT Token Service       │
│                                 │   (Generate/Validate)     │
│                                 │                            │
│                                 └─► Refresh Token Repo      │
│                                     (ArangoDB)              │
│                                                              │
│  Protected Routes                                            │
│       │                                                      │
│       │ GET /api/v1/agencies (with Bearer token)            │
│       ▼                                                      │
│  ┌───────────────────┐                                      │
│  │ Auth Middleware   │──────► JWT Validation               │
│  └───────────────────┘              │                       │
│       │                              └─► Extract user_id   │
│       │                                                      │
│       ▼                                                      │
│  Protected Handler (has user context)                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Dual-Token Strategy
- **Access Token**: Short-lived (15 minutes), used for API authentication
- **Refresh Token**: Long-lived (7 days), used to obtain new access tokens
- **Security**: Refresh tokens stored hashed in database, revocable

### Data Models
- **User**: Email, name, password_hash, timestamps, is_active
- **RefreshToken**: user_id, token_hash, expires_at, is_revoked

## Task Index

| Task ID | Title | Status | File |
|---------|-------|--------|------|
| MVP-AUTH-001 | User Model & Repository | ✅ Complete | [user-model-repository.md](user-model-repository.md) |
| MVP-AUTH-002 | JWT Token Service | ✅ Complete | [jwt-token-service.md](jwt-token-service.md) |
| MVP-AUTH-003 | Authentication Endpoints | ✅ Complete | [auth-endpoints.md](auth-endpoints.md) |
| MVP-AUTH-004 | Authentication Middleware | ✅ Complete | [auth-middleware.md](auth-middleware.md) |
| MVP-AUTH-005 | Protected Routes Integration | 📋 Not Started | [protected-routes.md](protected-routes.md) |

## Implementation Files

### Core Package: `internal/auth/`
```
internal/auth/
├── models.go          # User, RefreshToken, Request/Response types
├── repository.go      # ArangoDB persistence layer
├── service.go         # Business logic, JWT generation/validation
└── handler.go         # HTTP handlers for auth endpoints
```

### Integration Points
- `internal/api/middleware.go` - AuthMiddleware for route protection
- `internal/app/app.go` - Service initialization
- `internal/app/routes_api.go` - Route registration
- `internal/config/config.go` - JWT secret configuration
- `config.yaml` - Configuration file

## API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/auth/register` | Create new user account | No |
| POST | `/api/v1/auth/login` | Login with email/password | No |
| POST | `/api/v1/auth/refresh` | Refresh access token | No |
| POST | `/api/v1/auth/logout` | Revoke refresh token | No |
| GET | `/api/v1/auth/me` | Get current user info | Yes |

## Configuration

**config.yaml:**
```yaml
auth:
  jwt_secret: "change-this-secret-in-production-use-env-var"
```

**Environment Variable (Production):**
```bash
export JWT_SECRET="your-secure-random-secret-key-here"
```

## Security Considerations

### Password Security
- ✅ Bcrypt hashing (cost factor: 12)
- ✅ Passwords never stored in plain text
- ✅ Password hashes never exposed in API responses

### Token Security
- ✅ Access tokens expire after 15 minutes
- ✅ Refresh tokens expire after 7 days
- ✅ Refresh tokens stored as SHA-256 hashes
- ✅ Token revocation support (logout)
- ✅ HMAC-SHA256 signing for JWTs

### Future Enhancements
- ⏳ Password strength requirements (uppercase, lowercase, number, special char)
- ⏳ Email verification on registration
- ⏳ Password reset flow
- ⏳ Rate limiting on auth endpoints
- ⏳ Brute force protection
- ⏳ Multi-factor authentication (MFA)
- ⏳ OAuth2/OIDC integration

## Dependencies

### Go Packages
```go
github.com/golang-jwt/jwt/v5      // JWT generation and parsing
golang.org/x/crypto/bcrypt        // Password hashing
github.com/arangodb/go-driver     // ArangoDB client
```

### Database Collections
- `users` - User accounts (unique index on email)
- `refresh_tokens` - Refresh token storage (indexes on user_id, token)

## Testing Strategy

### Unit Tests
- [ ] User repository CRUD operations
- [ ] JWT token generation and validation
- [ ] Password hashing and verification
- [ ] Refresh token rotation

### Integration Tests
- [ ] Registration flow (duplicate email handling)
- [ ] Login flow (invalid credentials)
- [ ] Token refresh flow (expired tokens)
- [ ] Logout flow (token revocation)
- [ ] Protected route access (with/without valid token)

### Manual Testing Checklist
- [ ] Register new user via Postman
- [ ] Login with correct credentials
- [ ] Login with incorrect credentials
- [ ] Access protected route without token (401)
- [ ] Access protected route with valid token (200)
- [ ] Access protected route with expired token (401)
- [ ] Refresh access token
- [ ] Logout and verify token revoked

## Frontend Integration

**CodeValdFortex** tasks that depend on this:
- ✅ MVP-FL-009: Authentication State Management
- ✅ MVP-FL-010: Login & Registration Screens
- ✅ MVP-FL-011: Protected Routes & Permissions

**Expected Frontend Flow:**
1. User fills registration form → POST `/api/v1/auth/register`
2. User fills login form → POST `/api/v1/auth/login` → Store tokens
3. For API calls → Add `Authorization: Bearer {access_token}` header
4. On 401 response → POST `/api/v1/auth/refresh` with refresh_token
5. On logout → POST `/api/v1/auth/logout` → Clear tokens

## Deployment Notes

### Pre-deployment Checklist
- [ ] Set JWT_SECRET environment variable (production)
- [ ] Verify ArangoDB users collection created
- [ ] Verify email index on users collection
- [ ] Test token expiration in production-like environment
- [ ] Review CORS settings for frontend origin

### Monitoring
- Monitor failed login attempts (potential brute force)
- Monitor token refresh rate (unusual patterns)
- Monitor auth endpoint latency
- Alert on database connection failures

## Related Documentation
- `/documents/2-SoftwareDesignAndArchitecture/flutter-designs/sign-in-design.md` - Frontend auth UI
- `/documents/3-SofwareDevelopment/mvp.md` - Task definitions
- `internal/auth/` - Implementation code
