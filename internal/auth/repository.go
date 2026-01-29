package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
)

const (
	UsersCollection         = "users"
	RefreshTokensCollection = "refresh_tokens"
)

// Repository handles user and refresh token persistence
type Repository struct {
	db driver.Database
}

// NewRepository creates a new auth repository
func NewRepository(db driver.Database) (*Repository, error) {
	repo := &Repository{db: db}

	// Ensure collections exist
	if err := repo.ensureCollections(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure collections: %w", err)
	}

	return repo, nil
}

// ensureCollections creates required collections if they don't exist
func (r *Repository) ensureCollections(ctx context.Context) error {
	// Create users collection
	exists, err := r.db.CollectionExists(ctx, UsersCollection)
	if err != nil {
		return fmt.Errorf("failed to check users collection: %w", err)
	}
	if !exists {
		collection, err := r.db.CreateCollection(ctx, UsersCollection, nil)
		if err != nil {
			return fmt.Errorf("failed to create users collection: %w", err)
		}

		// Create unique index on email
		_, _, err = collection.EnsurePersistentIndex(ctx, []string{"email"}, &driver.EnsurePersistentIndexOptions{
			Unique: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create email index: %w", err)
		}
	}

	// Create refresh_tokens collection
	exists, err = r.db.CollectionExists(ctx, RefreshTokensCollection)
	if err != nil {
		return fmt.Errorf("failed to check refresh_tokens collection: %w", err)
	}
	if !exists {
		collection, err := r.db.CreateCollection(ctx, RefreshTokensCollection, nil)
		if err != nil {
			return fmt.Errorf("failed to create refresh_tokens collection: %w", err)
		}

		// Create index on user_id for faster lookup
		_, _, err = collection.EnsurePersistentIndex(ctx, []string{"user_id"}, nil)
		if err != nil {
			return fmt.Errorf("failed to create user_id index: %w", err)
		}

		// Create index on token for faster lookup
		_, _, err = collection.EnsurePersistentIndex(ctx, []string{"token"}, nil)
		if err != nil {
			return fmt.Errorf("failed to create token index: %w", err)
		}
	}

	return nil
}

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	collection, err := r.db.Collection(ctx, UsersCollection)
	if err != nil {
		return fmt.Errorf("failed to get users collection: %w", err)
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true

	meta, err := collection.CreateDocument(ctx, user)
	if err != nil {
		if driver.IsConflict(err) {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.Key = meta.Key
	user.ID = meta.ID.String()
	user.Rev = meta.Rev

	return nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := "FOR u IN @@collection FILTER u.email == @email LIMIT 1 RETURN u"
	bindVars := map[string]interface{}{
		"@collection": UsersCollection,
		"email":       email,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	defer cursor.Close()

	var user User
	if _, err := cursor.ReadDocument(ctx, &user); err != nil {
		if driver.IsNoMoreDocuments(err) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to read user: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	collection, err := r.db.Collection(ctx, UsersCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get users collection: %w", err)
	}

	// Extract key from ID (e.g., "users/123" -> "123")
	// If id is already just the key, use it as-is
	key := id
	if strings.Contains(id, "/") {
		parts := strings.Split(id, "/")
		if len(parts) == 2 {
			key = parts[1]
		}
	}

	var user User
	if _, err := collection.ReadDocument(ctx, key, &user); err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to read user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (r *Repository) UpdateUser(ctx context.Context, user *User) error {
	collection, err := r.db.Collection(ctx, UsersCollection)
	if err != nil {
		return fmt.Errorf("failed to get users collection: %w", err)
	}

	user.UpdatedAt = time.Now()

	meta, err := collection.UpdateDocument(ctx, user.Key, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	user.Rev = meta.Rev
	return nil
}

// SaveRefreshToken stores a refresh token
func (r *Repository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	collection, err := r.db.Collection(ctx, RefreshTokensCollection)
	if err != nil {
		return fmt.Errorf("failed to get refresh_tokens collection: %w", err)
	}

	token.CreatedAt = time.Now()
	token.IsRevoked = false

	meta, err := collection.CreateDocument(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	token.Key = meta.Key
	token.ID = meta.ID.String()
	token.Rev = meta.Rev

	return nil
}

// GetRefreshToken retrieves a refresh token by token value
func (r *Repository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := "FOR t IN @@collection FILTER t.token == @token AND t.is_revoked == false LIMIT 1 RETURN t"
	bindVars := map[string]interface{}{
		"@collection": RefreshTokensCollection,
		"token":       tokenHash,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query refresh token: %w", err)
	}
	defer cursor.Close()

	var token RefreshToken
	if _, err := cursor.ReadDocument(ctx, &token); err != nil {
		if driver.IsNoMoreDocuments(err) {
			return nil, fmt.Errorf("refresh token not found or revoked")
		}
		return nil, fmt.Errorf("failed to read refresh token: %w", err)
	}

	return &token, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	token, err := r.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	collection, err := r.db.Collection(ctx, RefreshTokensCollection)
	if err != nil {
		return fmt.Errorf("failed to get refresh_tokens collection: %w", err)
	}

	now := time.Now()
	token.IsRevoked = true
	token.RevokedAt = &now

	if _, err := collection.UpdateDocument(ctx, token.Key, token); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *Repository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	query := `
		FOR t IN @@collection 
		FILTER t.user_id == @user_id AND t.is_revoked == false
		UPDATE t WITH { is_revoked: true, revoked_at: @now } IN @@collection
	`
	bindVars := map[string]interface{}{
		"@collection": RefreshTokensCollection,
		"user_id":     userID,
		"now":         time.Now(),
	}

	if _, err := r.db.Query(ctx, query, bindVars); err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes expired refresh tokens
func (r *Repository) CleanupExpiredTokens(ctx context.Context) error {
	query := `
		FOR t IN @@collection 
		FILTER t.expires_at < @now
		REMOVE t IN @@collection
	`
	bindVars := map[string]interface{}{
		"@collection": RefreshTokensCollection,
		"now":         time.Now(),
	}

	if _, err := r.db.Query(ctx, query, bindVars); err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
}
