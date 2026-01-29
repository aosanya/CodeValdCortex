package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	// AccessTokenDuration is the lifetime of access tokens (15 minutes)
	AccessTokenDuration = 15 * time.Minute

	// RefreshTokenDuration is the lifetime of refresh tokens (7 days)
	RefreshTokenDuration = 7 * 24 * time.Hour

	// BcryptCost is the cost factor for bcrypt hashing
	BcryptCost = 12
)

// Service handles authentication business logic
type Service struct {
	repo      *Repository
	jwtSecret []byte
}

// NewService creates a new auth service
func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(passwordHash),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		fmt.Printf("DEBUG: GetUserByEmail failed: %v\n", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	fmt.Printf("DEBUG: User found - Email: %s, PasswordHash length: %d, IsActive: %v\n",
		user.Email, len(user.PasswordHash), user.IsActive)

	// Check if user is active
	if !user.IsActive {
		fmt.Printf("DEBUG: User is not active\n")
		return nil, fmt.Errorf("account is disabled")
	}

	// Verify password
	fmt.Printf("DEBUG: Comparing password, input length: %d\n", len(req.Password))
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		fmt.Printf("DEBUG: Password comparison failed: %v\n", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	fmt.Printf("DEBUG: Login successful!\n")

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateAndSaveRefreshToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenDuration.Seconds()),
		User:         user.ToResponse(),
	}, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (s *Service) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (*AuthResponse, error) {
	// Hash the refresh token to look it up
	tokenHash := s.hashToken(refreshTokenStr)

	// Get refresh token from database
	refreshToken, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if token is expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Optionally: Generate new refresh token (token rotation)
	// For now, we reuse the same refresh token
	newRefreshToken := refreshTokenStr

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenDuration.Seconds()),
		User:         user.ToResponse(),
	}, nil
}

// Logout revokes a refresh token
func (s *Service) Logout(ctx context.Context, refreshTokenStr string) error {
	tokenHash := s.hashToken(refreshTokenStr)
	return s.repo.RevokeRefreshToken(ctx, tokenHash)
}

// GetCurrentUser retrieves the authenticated user by ID
func (s *Service) GetCurrentUser(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// ValidateAccessToken validates a JWT access token and returns claims
func (s *Service) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Convert to TokenClaims
	userID, _ := claims["user_id"].(string)
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	exp, _ := claims["exp"].(float64)
	iat, _ := claims["iat"].(float64)

	return &TokenClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
		Exp:    int64(exp),
		Iat:    int64(iat),
	}, nil
}

// generateAccessToken creates a JWT access token
func (s *Service) generateAccessToken(user *User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(AccessTokenDuration)

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"iat":     now.Unix(),
		"exp":     expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// generateAndSaveRefreshToken creates and stores a refresh token
func (s *Service) generateAndSaveRefreshToken(ctx context.Context, user *User) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	tokenStr := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash token for storage
	tokenHash := s.hashToken(tokenStr)

	// Save to database
	refreshToken := &RefreshToken{
		UserID:    user.ID,
		Token:     tokenHash,
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
	}

	if err := s.repo.SaveRefreshToken(ctx, refreshToken); err != nil {
		return "", err
	}

	return tokenStr, nil
}

// hashToken creates a SHA-256 hash of a token
func (s *Service) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
