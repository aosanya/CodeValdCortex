package giteawebhook

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/infrastructure/work"
	driver "github.com/arangodb/go-driver"
)

// Repository implements work.Repository for ArangoDB persistence
type Repository struct {
	db            driver.Database
	issuesCol     driver.Collection
	prsCol        driver.Collection
	milestonesCol driver.Collection
}

// NewRepository creates a new ArangoDB repository for work items
func NewRepository(db driver.Database) (*Repository, error) {
	ctx := context.Background()

	// Get or create collections
	issuesCol, err := ensureCollection(ctx, db, "work_issues")
	if err != nil {
		return nil, fmt.Errorf("failed to create work_issues collection: %w", err)
	}

	prsCol, err := ensureCollection(ctx, db, "work_prs")
	if err != nil {
		return nil, fmt.Errorf("failed to create work_prs collection: %w", err)
	}

	milestonesCol, err := ensureCollection(ctx, db, "work_milestones")
	if err != nil {
		return nil, fmt.Errorf("failed to create work_milestones collection: %w", err)
	}

	return &Repository{
		db:            db,
		issuesCol:     issuesCol,
		prsCol:        prsCol,
		milestonesCol: milestonesCol,
	}, nil
}

// ensureCollection creates a collection if it doesn't exist
func ensureCollection(ctx context.Context, db driver.Database, name string) (driver.Collection, error) {
	exists, err := db.CollectionExists(ctx, name)
	if err != nil {
		return nil, err
	}

	if exists {
		return db.Collection(ctx, name)
	}

	return db.CreateCollection(ctx, name, nil)
}

// SaveIssue persists a work issue to the work_issues collection
// Uses upsert logic to handle duplicate webhooks (idempotent)
func (r *Repository) SaveIssue(ctx context.Context, issue *work.WorkIssue) error {
	if issue == nil {
		return fmt.Errorf("issue cannot be nil")
	}

	// Generate deterministic key from provider and issue ID
	key := generateKey(issue.Provider, issue.IssueID)

	// Check if document exists
	exists, err := r.issuesCol.DocumentExists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check if issue exists: %w", err)
	}

	// Create document structure
	doc := map[string]interface{}{
		"_key":              key,
		"provider":          issue.Provider,
		"issue_id":          issue.IssueID,
		"issue_number":      issue.IssueNumber,
		"title":             issue.Title,
		"body":              issue.Body,
		"state":             issue.State,
		"repo_url":          issue.RepoURL,
		"author_username":   issue.AuthorUsername,
		"author_email":      issue.AuthorEmail,
		"milestone":         issue.Milestone,
		"milestone_id":      issue.MilestoneID,
		"labels":            issue.Labels,
		"assignees":         issue.Assignees,
		"created_at":        issue.CreatedAt,
		"updated_at":        issue.UpdatedAt,
		"closed_at":         issue.ClosedAt,
		"provider_metadata": issue.ProviderMetadata,
	}

	if exists {
		// Update existing document
		_, err = r.issuesCol.UpdateDocument(ctx, key, doc)
		if err != nil {
			return fmt.Errorf("failed to update issue: %w", err)
		}
	} else {
		// Create new document
		_, err = r.issuesCol.CreateDocument(ctx, doc)
		if err != nil {
			return fmt.Errorf("failed to create issue: %w", err)
		}
	}

	return nil
}

// SavePullRequest persists a PR to the work_prs collection
func (r *Repository) SavePullRequest(ctx context.Context, pr *work.WorkPullRequest) error {
	if pr == nil {
		return fmt.Errorf("pull request cannot be nil")
	}

	key := generateKey(pr.Provider, pr.PRID)

	exists, err := r.prsCol.DocumentExists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check if PR exists: %w", err)
	}

	doc := map[string]interface{}{
		"_key":              key,
		"provider":          pr.Provider,
		"pr_id":             pr.PRID,
		"pr_number":         pr.PRNumber,
		"title":             pr.Title,
		"body":              pr.Body,
		"state":             pr.State,
		"head_branch":       pr.HeadBranch,
		"base_branch":       pr.BaseBranch,
		"repo_url":          pr.RepoURL,
		"merged":            pr.Merged,
		"mergeable":         pr.Mergeable,
		"author_username":   pr.AuthorUsername,
		"author_email":      pr.AuthorEmail,
		"labels":            pr.Labels,
		"reviewers":         pr.Reviewers,
		"created_at":        pr.CreatedAt,
		"updated_at":        pr.UpdatedAt,
		"merged_at":         pr.MergedAt,
		"closed_at":         pr.ClosedAt,
		"provider_metadata": pr.ProviderMetadata,
	}

	if exists {
		_, err = r.prsCol.UpdateDocument(ctx, key, doc)
		if err != nil {
			return fmt.Errorf("failed to update PR: %w", err)
		}
	} else {
		_, err = r.prsCol.CreateDocument(ctx, doc)
		if err != nil {
			return fmt.Errorf("failed to create PR: %w", err)
		}
	}

	return nil
}

// SaveMilestone persists a milestone to the work_milestones collection
func (r *Repository) SaveMilestone(ctx context.Context, milestone *work.WorkMilestone) error {
	if milestone == nil {
		return fmt.Errorf("milestone cannot be nil")
	}

	key := generateKey(milestone.Provider, milestone.MilestoneID)

	exists, err := r.milestonesCol.DocumentExists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check if milestone exists: %w", err)
	}

	doc := map[string]interface{}{
		"_key":              key,
		"provider":          milestone.Provider,
		"milestone_id":      milestone.MilestoneID,
		"title":             milestone.Title,
		"description":       milestone.Description,
		"state":             milestone.State,
		"repo_url":          milestone.RepoURL,
		"open_issues":       milestone.OpenIssues,
		"closed_issues":     milestone.ClosedIssues,
		"due_date":          milestone.DueDate,
		"created_at":        milestone.CreatedAt,
		"updated_at":        milestone.UpdatedAt,
		"closed_at":         milestone.ClosedAt,
		"provider_metadata": milestone.ProviderMetadata,
	}

	if exists {
		_, err = r.milestonesCol.UpdateDocument(ctx, key, doc)
		if err != nil {
			return fmt.Errorf("failed to update milestone: %w", err)
		}
	} else {
		_, err = r.milestonesCol.CreateDocument(ctx, doc)
		if err != nil {
			return fmt.Errorf("failed to create milestone: %w", err)
		}
	}

	return nil
}

// GetIssue retrieves an issue by provider and issue ID
func (r *Repository) GetIssue(ctx context.Context, provider string, issueID string) (*work.WorkIssue, error) {
	key := generateKey(provider, issueID)

	var issue work.WorkIssue
	_, err := r.issuesCol.ReadDocument(ctx, key, &issue)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, nil // Issue not found
		}
		return nil, fmt.Errorf("failed to read issue: %w", err)
	}

	return &issue, nil
}

// GetPullRequest retrieves a PR by provider and PR ID
func (r *Repository) GetPullRequest(ctx context.Context, provider string, prID string) (*work.WorkPullRequest, error) {
	key := generateKey(provider, prID)

	var pr work.WorkPullRequest
	_, err := r.prsCol.ReadDocument(ctx, key, &pr)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read PR: %w", err)
	}

	return &pr, nil
}

// GetMilestone retrieves a milestone by provider and milestone ID
func (r *Repository) GetMilestone(ctx context.Context, provider string, milestoneID string) (*work.WorkMilestone, error) {
	key := generateKey(provider, milestoneID)

	var milestone work.WorkMilestone
	_, err := r.milestonesCol.ReadDocument(ctx, key, &milestone)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read milestone: %w", err)
	}

	return &milestone, nil
}

// generateKey creates a deterministic key for ArangoDB documents
// Format: provider_hash(id)
func generateKey(provider string, id string) string {
	// Use a simple but deterministic key format
	// In production, consider using a hash function for very long IDs
	return fmt.Sprintf("%s_%s", provider, sanitizeKey(id))
}

// sanitizeKey removes characters that are invalid in ArangoDB keys
func sanitizeKey(s string) string {
	// ArangoDB keys can contain: a-z, A-Z, 0-9, _, -, :, @, (, ), +, ,, =, ;, $, !, *, ', %
	// For URLs, we'll use a simple encoding scheme
	// Replace / with _ and other special chars
	result := ""
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			result += string(r)
		case r == '-' || r == '_' || r == ':' || r == '@':
			result += string(r)
		default:
			result += "_"
		}
	}
	return result
}
