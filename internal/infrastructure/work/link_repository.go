package work

import (
	"context"
	"fmt"
	"time"

	driver "github.com/arangodb/go-driver"
	log "github.com/sirupsen/logrus"
)

const (
	CollectionAgentIssueLinks = "agent_issue_links"
	CollectionSyncAudit       = "sync_audit"
)

// ArangoAgentIssueLinkRepository implements AgentIssueLinkRepository using ArangoDB
type ArangoAgentIssueLinkRepository struct {
	db  driver.Database
	col driver.Collection
}

// NewAgentIssueLinkRepository creates a new ArangoDB-backed agent-issue link repository
func NewAgentIssueLinkRepository(db driver.Database) (AgentIssueLinkRepository, error) {
	ctx := context.Background()

	// Ensure collection exists
	col, err := ensureCollection(ctx, db, CollectionAgentIssueLinks)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure agent_issue_links collection: %w", err)
	}

	// Create indexes for efficient queries
	if err := ensureAgentIssueLinkIndexes(ctx, col); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	log.WithField("collection", CollectionAgentIssueLinks).Info("Agent-issue link repository initialized")

	return &ArangoAgentIssueLinkRepository{
		db:  db,
		col: col,
	}, nil
}

// ensureCollection creates a collection if it doesn't exist
func ensureCollection(ctx context.Context, db driver.Database, name string) (driver.Collection, error) {
	exists, err := db.CollectionExists(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	if exists {
		return db.Collection(ctx, name)
	}

	col, err := db.CreateCollection(ctx, name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	log.WithField("collection", name).Info("Collection created")
	return col, nil
}

// ensureAgentIssueLinkIndexes creates necessary indexes
func ensureAgentIssueLinkIndexes(ctx context.Context, col driver.Collection) error {
	// Index on agent_id for fast agent lookups
	_, _, err := col.EnsurePersistentIndex(ctx, []string{"agent_id"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
		Name:   "idx_agent_id",
	})
	if err != nil {
		return fmt.Errorf("failed to create agent_id index: %w", err)
	}

	// Index on issue_id for reverse lookups
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"issue_id"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_issue_id",
	})
	if err != nil {
		return fmt.Errorf("failed to create issue_id index: %w", err)
	}

	// Index on status for filtering active links
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"status"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_status",
	})
	if err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	// Compound index on repository_url and status for repository queries
	_, _, err = col.EnsurePersistentIndex(ctx, []string{"repository_url", "status"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_repo_status",
	})
	if err != nil {
		return fmt.Errorf("failed to create repository_url+status index: %w", err)
	}

	return nil
}

// Create creates a new agent-issue link
func (r *ArangoAgentIssueLinkRepository) Create(ctx context.Context, link *AgentIssueLink) error {
	// Use agent_id as document key for uniqueness
	link.Key = link.AgentID
	link.CreatedAt = time.Now()
	link.LastSyncAt = time.Now()
	link.SyncCount = 0

	if link.Status == "" {
		link.Status = "active"
	}

	_, err := r.col.CreateDocument(ctx, link)
	if err != nil {
		return fmt.Errorf("failed to create agent-issue link: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": link.AgentID,
		"issue_id": link.IssueID,
	}).Info("Agent-issue link created")

	return nil
}

// GetByAgentID retrieves the link for a specific agent
func (r *ArangoAgentIssueLinkRepository) GetByAgentID(ctx context.Context, agentID string) (*AgentIssueLink, error) {
	var link AgentIssueLink
	_, err := r.col.ReadDocument(ctx, agentID, &link)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("agent-issue link not found for agent: %s", agentID)
		}
		return nil, fmt.Errorf("failed to read agent-issue link: %w", err)
	}

	return &link, nil
}

// GetByIssueID retrieves the link for a specific issue
func (r *ArangoAgentIssueLinkRepository) GetByIssueID(ctx context.Context, issueID string) (*AgentIssueLink, error) {
	query := `
		FOR link IN @@collection
		FILTER link.issue_id == @issue_id
		LIMIT 1
		RETURN link
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionAgentIssueLinks,
		"issue_id":    issueID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent-issue link: %w", err)
	}
	defer cursor.Close()

	if !cursor.HasMore() {
		return nil, fmt.Errorf("agent-issue link not found for issue: %s", issueID)
	}

	var link AgentIssueLink
	_, err = cursor.ReadDocument(ctx, &link)
	if err != nil {
		return nil, fmt.Errorf("failed to read link document: %w", err)
	}

	return &link, nil
}

// UpdateStatus updates the link status
func (r *ArangoAgentIssueLinkRepository) UpdateStatus(ctx context.Context, agentID string, status string) error {
	patch := map[string]interface{}{
		"status": status,
	}

	_, err := r.col.UpdateDocument(ctx, agentID, patch)
	if err != nil {
		return fmt.Errorf("failed to update link status: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"status":   status,
	}).Debug("Agent-issue link status updated")

	return nil
}

// UpdateLastSync updates the last sync timestamp and metadata
func (r *ArangoAgentIssueLinkRepository) UpdateLastSync(ctx context.Context, agentID string, eventType string, commentID string) error {
	// Read current link to increment sync count
	var link AgentIssueLink
	_, err := r.col.ReadDocument(ctx, agentID, &link)
	if err != nil {
		return fmt.Errorf("failed to read link for sync update: %w", err)
	}

	patch := map[string]interface{}{
		"last_sync_at":    time.Now(),
		"sync_count":      link.SyncCount + 1,
		"last_event_type": eventType,
		"last_comment_id": commentID,
	}

	_, err = r.col.UpdateDocument(ctx, agentID, patch)
	if err != nil {
		return fmt.Errorf("failed to update last sync: %w", err)
	}

	return nil
}

// MarkCompleted marks the link as completed with timestamp
func (r *ArangoAgentIssueLinkRepository) MarkCompleted(ctx context.Context, agentID string) error {
	now := time.Now()
	patch := map[string]interface{}{
		"status":       "completed",
		"completed_at": now,
	}

	_, err := r.col.UpdateDocument(ctx, agentID, patch)
	if err != nil {
		return fmt.Errorf("failed to mark link completed: %w", err)
	}

	log.WithField("agent_id", agentID).Info("Agent-issue link marked completed")

	return nil
}

// ListActive returns all active agent-issue links
func (r *ArangoAgentIssueLinkRepository) ListActive(ctx context.Context) ([]*AgentIssueLink, error) {
	query := `
		FOR link IN @@collection
		FILTER link.status == "active"
		SORT link.created_at DESC
		RETURN link
	`

	bindVars := map[string]interface{}{
		"@collection": CollectionAgentIssueLinks,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query active links: %w", err)
	}
	defer cursor.Close()

	var links []*AgentIssueLink
	for {
		var link AgentIssueLink
		_, err := cursor.ReadDocument(ctx, &link)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read link document: %w", err)
		}
		links = append(links, &link)
	}

	return links, nil
}

// Delete removes a link
func (r *ArangoAgentIssueLinkRepository) Delete(ctx context.Context, agentID string) error {
	_, err := r.col.RemoveDocument(ctx, agentID)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("agent-issue link not found: %s", agentID)
		}
		return fmt.Errorf("failed to delete agent-issue link: %w", err)
	}

	log.WithField("agent_id", agentID).Info("Agent-issue link deleted")

	return nil
}
