package work

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
)

const (
	// PRCollection is the ArangoDB collection name for PRs
	PRCollection = "agent_prs"
)

// prRepositoryImpl implements PRRepository interface
type prRepositoryImpl struct {
	db driver.Database
}

// NewPRRepository creates a new PR repository
func NewPRRepository(db driver.Database) PRRepository {
	return &prRepositoryImpl{
		db: db,
	}
}

// Create creates a new PR record
func (r *prRepositoryImpl) Create(ctx context.Context, pr *PRInfo) error {
	col, err := r.db.Collection(ctx, PRCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	_, err = col.CreateDocument(ctx, pr)
	if err != nil {
		return fmt.Errorf("failed to create PR document: %w", err)
	}

	return nil
}

// Update updates an existing PR
func (r *prRepositoryImpl) Update(ctx context.Context, prID string, updates map[string]interface{}) error {
	col, err := r.db.Collection(ctx, PRCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	_, err = col.UpdateDocument(ctx, prID, updates)
	if err != nil {
		return fmt.Errorf("failed to update PR %s: %w", prID, err)
	}

	return nil
}

// GetByID retrieves a PR by its ID
func (r *prRepositoryImpl) GetByID(ctx context.Context, prID string) (*PRInfo, error) {
	col, err := r.db.Collection(ctx, PRCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	var pr PRInfo
	_, err = col.ReadDocument(ctx, prID, &pr)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			return nil, fmt.Errorf("PR not found: %s", prID)
		}
		return nil, fmt.Errorf("failed to read PR %s: %w", prID, err)
	}

	return &pr, nil
}

// GetByNumber retrieves a PR by repository URL and PR number
func (r *prRepositoryImpl) GetByNumber(ctx context.Context, repoURL string, number int64) (*PRInfo, error) {
	query := `
		FOR pr IN @@collection
		FILTER pr.repository_url == @repoURL
		FILTER pr.number == @number
		LIMIT 1
		RETURN pr
	`

	bindVars := map[string]interface{}{
		"@collection": PRCollection,
		"repoURL":     repoURL,
		"number":      number,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query PR by number: %w", err)
	}
	defer cursor.Close()

	var pr PRInfo
	_, err = cursor.ReadDocument(ctx, &pr)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			return nil, fmt.Errorf("PR not found for repo %s, number %d", repoURL, number)
		}
		return nil, fmt.Errorf("failed to read PR document: %w", err)
	}

	return &pr, nil
}

// ListByAgent retrieves all PRs created by a specific agent
func (r *prRepositoryImpl) ListByAgent(ctx context.Context, agentID string) ([]*PRInfo, error) {
	query := `
		FOR pr IN @@collection
		FILTER pr.agent_id == @agentID
		SORT pr.created_at DESC
		RETURN pr
	`

	bindVars := map[string]interface{}{
		"@collection": PRCollection,
		"agentID":     agentID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query PRs by agent: %w", err)
	}
	defer cursor.Close()

	var prs []*PRInfo
	for {
		var pr PRInfo
		_, err := cursor.ReadDocument(ctx, &pr)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read PR document: %w", err)
		}
		prs = append(prs, &pr)
	}

	return prs, nil
}

// ListByIssue retrieves all PRs linked to a specific issue
func (r *prRepositoryImpl) ListByIssue(ctx context.Context, issueID string) ([]*PRInfo, error) {
	query := `
		FOR pr IN @@collection
		FILTER pr.linked_issue_id == @issueID
		SORT pr.created_at DESC
		RETURN pr
	`

	bindVars := map[string]interface{}{
		"@collection": PRCollection,
		"issueID":     issueID,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query PRs by issue: %w", err)
	}
	defer cursor.Close()

	var prs []*PRInfo
	for {
		var pr PRInfo
		_, err := cursor.ReadDocument(ctx, &pr)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read PR document: %w", err)
		}
		prs = append(prs, &pr)
	}

	return prs, nil
}

// ListByState retrieves all PRs with a specific state
func (r *prRepositoryImpl) ListByState(ctx context.Context, state string) ([]*PRInfo, error) {
	query := `
		FOR pr IN @@collection
		FILTER pr.state == @state
		SORT pr.created_at DESC
		RETURN pr
	`

	bindVars := map[string]interface{}{
		"@collection": PRCollection,
		"state":       state,
	}

	cursor, err := r.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query PRs by state: %w", err)
	}
	defer cursor.Close()

	var prs []*PRInfo
	for {
		var pr PRInfo
		_, err := cursor.ReadDocument(ctx, &pr)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read PR document: %w", err)
		}
		prs = append(prs, &pr)
	}

	return prs, nil
}

// Delete deletes a PR record
func (r *prRepositoryImpl) Delete(ctx context.Context, prID string) error {
	col, err := r.db.Collection(ctx, PRCollection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	_, err = col.RemoveDocument(ctx, prID)
	if err != nil {
		return fmt.Errorf("failed to delete PR %s: %w", prID, err)
	}

	return nil
}

// UpdateQualityChecks updates the quality check results for a PR
func (r *prRepositoryImpl) UpdateQualityChecks(ctx context.Context, prID string, checks *QualityCheckResults) error {
	updates := map[string]interface{}{
		"quality_checks": checks,
	}

	return r.Update(ctx, prID, updates)
}
