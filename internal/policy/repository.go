package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/arangodb/go-driver"
)

// Repository defines the interface for AI policy data access
type Repository interface {
	// Create creates a new AI policy
	Create(ctx context.Context, policy *AIPolicy) error

	// Get retrieves an AI policy by agency ID
	Get(ctx context.Context, agencyID string) (*AIPolicy, error)

	// Update updates an existing AI policy
	Update(ctx context.Context, policy *AIPolicy) error

	// Delete deletes an AI policy
	Delete(ctx context.Context, agencyID string) error

	// ListAll lists all AI policies
	ListAll(ctx context.Context) ([]*AIPolicy, error)

	// LogEvaluation logs a policy evaluation result
	LogEvaluation(ctx context.Context, evaluation *PolicyEvaluation) error

	// LogViolation logs a policy violation
	LogViolation(ctx context.Context, violation *PolicyViolation) error

	// GetViolations retrieves policy violations for an agency
	GetViolations(ctx context.Context, agencyID string, limit int) ([]*PolicyViolation, error)

	// GetEvaluations retrieves policy evaluations for an agency
	GetEvaluations(ctx context.Context, agencyID string, limit int) ([]*PolicyEvaluation, error)
}

// ArangoRepository implements Repository using ArangoDB
type ArangoRepository struct {
	policiesCol    driver.Collection
	evaluationsCol driver.Collection
	violationsCol  driver.Collection
}

// NewArangoRepository creates a new ArangoDB-backed policy repository
func NewArangoRepository(db driver.Database) (*ArangoRepository, error) {
	// Get or create ai_policies collection
	policiesCol, err := getOrCreateCollection(db, "ai_policies")
	if err != nil {
		return nil, fmt.Errorf("failed to get ai_policies collection: %w", err)
	}

	// Get or create policy_evaluations collection
	evaluationsCol, err := getOrCreateCollection(db, "policy_evaluations")
	if err != nil {
		return nil, fmt.Errorf("failed to get policy_evaluations collection: %w", err)
	}

	// Get or create policy_violations collection
	violationsCol, err := getOrCreateCollection(db, "policy_violations")
	if err != nil {
		return nil, fmt.Errorf("failed to get policy_violations collection: %w", err)
	}

	// Create indexes
	if err := createIndexes(policiesCol, evaluationsCol, violationsCol); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return &ArangoRepository{
		policiesCol:    policiesCol,
		evaluationsCol: evaluationsCol,
		violationsCol:  violationsCol,
	}, nil
}

// Create creates a new AI policy
func (r *ArangoRepository) Create(ctx context.Context, policy *AIPolicy) error {
	if policy.AgencyID == "" {
		return fmt.Errorf("agency_id is required")
	}

	// Set timestamps
	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	// Set version if not provided
	if policy.Version == "" {
		policy.Version = "1.0"
	}

	// Use agency_id as document key for easy lookup
	policy.Key = policy.AgencyID

	meta, err := r.policiesCol.CreateDocument(ctx, policy)
	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	policy.ID = meta.ID.String()
	policy.Rev = meta.Rev

	return nil
}

// Get retrieves an AI policy by agency ID
func (r *ArangoRepository) Get(ctx context.Context, agencyID string) (*AIPolicy, error) {
	if agencyID == "" {
		return nil, fmt.Errorf("agency_id is required")
	}

	var policy AIPolicy
	meta, err := r.policiesCol.ReadDocument(ctx, agencyID, &policy)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, fmt.Errorf("policy not found for agency: %s", agencyID)
		}
		return nil, fmt.Errorf("failed to read policy: %w", err)
	}

	policy.Key = meta.Key
	policy.ID = meta.ID.String()
	policy.Rev = meta.Rev

	return &policy, nil
}

// Update updates an existing AI policy
func (r *ArangoRepository) Update(ctx context.Context, policy *AIPolicy) error {
	if policy.AgencyID == "" {
		return fmt.Errorf("agency_id is required")
	}

	// Update timestamp
	policy.UpdatedAt = time.Now()

	// Increment version
	policy.Version = incrementVersion(policy.Version)

	meta, err := r.policiesCol.UpdateDocument(ctx, policy.AgencyID, policy)
	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	policy.Rev = meta.Rev

	return nil
}

// Delete deletes an AI policy
func (r *ArangoRepository) Delete(ctx context.Context, agencyID string) error {
	if agencyID == "" {
		return fmt.Errorf("agency_id is required")
	}

	_, err := r.policiesCol.RemoveDocument(ctx, agencyID)
	if err != nil {
		if driver.IsNotFound(err) {
			return fmt.Errorf("policy not found for agency: %s", agencyID)
		}
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	return nil
}

// ListAll lists all AI policies
func (r *ArangoRepository) ListAll(ctx context.Context) ([]*AIPolicy, error) {
	query := `
		FOR policy IN ai_policies
		SORT policy.updated_at DESC
		RETURN policy
	`

	cursor, err := r.policiesCol.Database().Query(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer cursor.Close()

	var policies []*AIPolicy
	for {
		var policy AIPolicy
		meta, err := cursor.ReadDocument(ctx, &policy)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read policy document: %w", err)
		}

		policy.Key = meta.Key
		policy.ID = meta.ID.String()
		policy.Rev = meta.Rev

		policies = append(policies, &policy)
	}

	return policies, nil
}

// LogEvaluation logs a policy evaluation result
func (r *ArangoRepository) LogEvaluation(ctx context.Context, evaluation *PolicyEvaluation) error {
	if evaluation.AgencyID == "" {
		return fmt.Errorf("agency_id is required")
	}

	// Set timestamp
	evaluation.EvaluatedAt = time.Now()

	meta, err := r.evaluationsCol.CreateDocument(ctx, evaluation)
	if err != nil {
		return fmt.Errorf("failed to log evaluation: %w", err)
	}

	evaluation.Key = meta.Key
	evaluation.ID = meta.ID.String()
	evaluation.Rev = meta.Rev

	return nil
}

// LogViolation logs a policy violation
func (r *ArangoRepository) LogViolation(ctx context.Context, violation *PolicyViolation) error {
	if violation.AgencyID == "" {
		return fmt.Errorf("agency_id is required")
	}

	// Set timestamp
	violation.ViolatedAt = time.Now()

	meta, err := r.violationsCol.CreateDocument(ctx, violation)
	if err != nil {
		return fmt.Errorf("failed to log violation: %w", err)
	}

	violation.Key = meta.Key
	violation.ID = meta.ID.String()
	violation.Rev = meta.Rev

	return nil
}

// GetViolations retrieves policy violations for an agency
func (r *ArangoRepository) GetViolations(ctx context.Context, agencyID string, limit int) ([]*PolicyViolation, error) {
	if agencyID == "" {
		return nil, fmt.Errorf("agency_id is required")
	}

	if limit <= 0 {
		limit = 100
	}

	query := `
		FOR violation IN policy_violations
		FILTER violation.agency_id == @agency_id
		SORT violation.violated_at DESC
		LIMIT @limit
		RETURN violation
	`

	bindVars := map[string]interface{}{
		"agency_id": agencyID,
		"limit":     limit,
	}

	cursor, err := r.violationsCol.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query violations: %w", err)
	}
	defer cursor.Close()

	var violations []*PolicyViolation
	for {
		var violation PolicyViolation
		meta, err := cursor.ReadDocument(ctx, &violation)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read violation document: %w", err)
		}

		violation.Key = meta.Key
		violation.ID = meta.ID.String()
		violation.Rev = meta.Rev

		violations = append(violations, &violation)
	}

	return violations, nil
}

// GetEvaluations retrieves policy evaluations for an agency
func (r *ArangoRepository) GetEvaluations(ctx context.Context, agencyID string, limit int) ([]*PolicyEvaluation, error) {
	if agencyID == "" {
		return nil, fmt.Errorf("agency_id is required")
	}

	if limit <= 0 {
		limit = 100
	}

	query := `
		FOR evaluation IN policy_evaluations
		FILTER evaluation.agency_id == @agency_id
		SORT evaluation.evaluated_at DESC
		LIMIT @limit
		RETURN evaluation
	`

	bindVars := map[string]interface{}{
		"agency_id": agencyID,
		"limit":     limit,
	}

	cursor, err := r.evaluationsCol.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("failed to query evaluations: %w", err)
	}
	defer cursor.Close()

	var evaluations []*PolicyEvaluation
	for {
		var evaluation PolicyEvaluation
		meta, err := cursor.ReadDocument(ctx, &evaluation)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read evaluation document: %w", err)
		}

		evaluation.Key = meta.Key
		evaluation.ID = meta.ID.String()
		evaluation.Rev = meta.Rev

		evaluations = append(evaluations, &evaluation)
	}

	return evaluations, nil
}

// Helper functions

func getOrCreateCollection(db driver.Database, name string) (driver.Collection, error) {
	ctx := context.Background()

	exists, err := db.CollectionExists(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}

	if exists {
		col, err := db.Collection(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("failed to get collection: %w", err)
		}
		return col, nil
	}

	col, err := db.CreateCollection(ctx, name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return col, nil
}

func createIndexes(policiesCol, evaluationsCol, violationsCol driver.Collection) error {
	ctx := context.Background()

	// Index on policies: agency_id (unique)
	_, _, err := policiesCol.EnsurePersistentIndex(ctx, []string{"agency_id"}, &driver.EnsurePersistentIndexOptions{
		Unique: true,
		Name:   "idx_agency_id",
	})
	if err != nil {
		return fmt.Errorf("failed to create agency_id index on policies: %w", err)
	}

	// Index on evaluations: agency_id + evaluated_at
	_, _, err = evaluationsCol.EnsurePersistentIndex(ctx, []string{"agency_id", "evaluated_at"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_agency_evaluated",
	})
	if err != nil {
		return fmt.Errorf("failed to create agency_evaluated index on evaluations: %w", err)
	}

	// Index on violations: agency_id + violated_at
	_, _, err = violationsCol.EnsurePersistentIndex(ctx, []string{"agency_id", "violated_at"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_agency_violated",
	})
	if err != nil {
		return fmt.Errorf("failed to create agency_violated index on violations: %w", err)
	}

	// Index on violations: severity
	_, _, err = violationsCol.EnsurePersistentIndex(ctx, []string{"severity"}, &driver.EnsurePersistentIndexOptions{
		Name: "idx_severity",
	})
	if err != nil {
		return fmt.Errorf("failed to create severity index on violations: %w", err)
	}

	return nil
}

func incrementVersion(version string) string {
	// Simple version increment: "1.0" -> "1.1", "1.9" -> "1.10", etc.
	// For MVP purposes, this is sufficient
	var major, minor int
	fmt.Sscanf(version, "%d.%d", &major, &minor)
	minor++
	return fmt.Sprintf("%d.%d", major, minor)
}
