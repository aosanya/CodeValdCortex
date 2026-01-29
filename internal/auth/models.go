package auth

import (
	"time"
)

// User represents a user account in the system
type User struct {
	Key          string    `json:"_key,omitempty"`
	ID           string    `json:"_id,omitempty"`
	Rev          string    `json:"_rev,omitempty"`
	Email        string    `json:"email" binding:"required,email"`
	Name         string    `json:"name" binding:"required"`
	PasswordHash string    `json:"password_hash,omitempty"` // Store in DB, but omitempty so responses can omit
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
}

// UserResponse represents user data for API responses (no sensitive data)
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

// ToResponse converts User to UserResponse (removes sensitive fields)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		IsActive:  u.IsActive,
	}
}

// RefreshToken represents a refresh token for JWT token rotation
type RefreshToken struct {
	Key       string     `json:"_key,omitempty"`
	ID        string     `json:"_id,omitempty"`
	Rev       string     `json:"_rev,omitempty"`
	UserID    string     `json:"user_id" binding:"required"`
	Token     string     `json:"token" binding:"required"` // Hashed refresh token
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	IsRevoked bool       `json:"is_revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"` // seconds until access token expires
	User         UserResponse `json:"user"`
}

// TokenClaims represents the JWT claims
type TokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}
