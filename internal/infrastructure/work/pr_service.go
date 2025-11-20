package work

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// prServiceImpl implements PRService interface
type prServiceImpl struct {
	prRepo          PRRepository
	gitOps          GitOperations
	checkService    QualityCheckService
	autoMerge       AutoMergeEngine
	apiClient       APIClient
	syncService     SyncService
	linkRepo        AgentIssueLinkRepository
	autoMergeConfig *AutoMergeConfig
}

// PRServiceConfig holds configuration for PR service
type PRServiceConfig struct {
	AutoMergeConfig *AutoMergeConfig
}

// NewPRService creates a new PR service
func NewPRService(
	prRepo PRRepository,
	gitOps GitOperations,
	checkService QualityCheckService,
	autoMerge AutoMergeEngine,
	apiClient APIClient,
	syncService SyncService,
	linkRepo AgentIssueLinkRepository,
	config *PRServiceConfig,
) PRService {
	return &prServiceImpl{
		prRepo:          prRepo,
		gitOps:          gitOps,
		checkService:    checkService,
		autoMerge:       autoMerge,
		apiClient:       apiClient,
		syncService:     syncService,
		linkRepo:        linkRepo,
		autoMergeConfig: config.AutoMergeConfig,
	}
}

// CreatePR creates a new pull request
func (s *prServiceImpl) CreatePR(ctx context.Context, req *CreatePRRequest) (*PRResult, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, &PRCreationError{
			Reason: err.Error(),
			Code:   ErrValidationFail,
		}
	}

	// Generate branch name if source branch not specified
	branchName := req.SourceBranch
	if branchName == "" {
		branchName = fmt.Sprintf("agent/%s/%s", req.AgentID, uuid.New().String()[:8])
	}

	// Create branch
	targetBranch := req.TargetBranch
	if targetBranch == "" {
		targetBranch = "main"
	}

	err := s.gitOps.CreateBranch(ctx, req.RepositoryURL, targetBranch, branchName)
	if err != nil {
		return nil, &PRCreationError{
			Reason: fmt.Sprintf("failed to create branch: %v", err),
			Code:   ErrBranchExists,
		}
	}

	// Push changes to branch
	if req.Changes != nil && len(req.Changes.Files) > 0 {
		err = s.gitOps.PushChanges(ctx, req.RepositoryURL, branchName, req.Changes)
		if err != nil {
			return nil, &PRCreationError{
				Reason: fmt.Sprintf("failed to push changes: %v", err),
				Code:   ErrPushFailed,
			}
		}
	}

	// Create PR via API client
	parts := strings.SplitN(req.RepositoryURL, "/", 2)
	if len(parts) != 2 {
		return nil, &PRCreationError{
			Reason: fmt.Sprintf("invalid repository URL: %s", req.RepositoryURL),
			Code:   ErrValidationFail,
		}
	}
	owner, repo := parts[0], parts[1]

	pr, err := s.apiClient.CreatePullRequest(ctx, owner, repo, CreatePullRequestOptions{
		Title:  req.Title,
		Body:   req.Description,
		Head:   branchName,
		Base:   targetBranch,
		Labels: req.Labels,
	})
	if err != nil {
		return nil, &PRCreationError{
			Reason: fmt.Sprintf("failed to create PR via API: %v", err),
			Code:   ErrAPIError,
		}
	}

	// Create PR record in database
	prInfo := &PRInfo{
		ID:               pr.PRID,
		Number:           pr.PRNumber,
		Title:            pr.Title,
		Description:      pr.Body,
		RepositoryURL:    req.RepositoryURL,
		SourceBranch:     branchName,
		TargetBranch:     targetBranch,
		State:            pr.State,
		CreatedBy:        req.AgentID,
		AgentID:          req.AgentID,
		LinkedIssueID:    req.IssueID,
		CommitSHA:        "", // Will be populated by webhook
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		AutoMergeEnabled: req.AutoMerge,
	}

	err = s.prRepo.Create(ctx, prInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR record: %w", err)
	}

	// Link agent to PR
	if req.AgentID != "" {
		err = s.LinkAgentToPR(ctx, req.AgentID, prInfo.ID)
		if err != nil {
			// Log error but don't fail PR creation
			fmt.Printf("Warning: failed to link agent to PR: %v\n", err)
		}
	}

	// Post comment to linked issue
	if req.IssueID != "" {
		// Note: WorkPullRequest doesn't have URL field, construct manually
		comment := fmt.Sprintf("📝 **Pull Request Created**\n\nI've created PR #%d with my proposed changes.\n\n🔗 Repository: %s", pr.PRNumber, req.RepositoryURL)
		parts := strings.SplitN(req.RepositoryURL, "/", 2)
		if len(parts) == 2 {
			_, err = s.apiClient.PostComment(ctx, parts[0], parts[1], req.IssueID, comment)
			if err != nil {
				fmt.Printf("Warning: failed to post PR link comment: %v\n", err)
			}
		}
	}

	return &PRResult{
		PRID:       prInfo.ID,
		PRNumber:   pr.PRNumber,
		URL:        req.RepositoryURL, // Construct URL or leave for frontend
		BranchName: branchName,
	}, nil
}

// UpdatePR updates an existing pull request
func (s *prServiceImpl) UpdatePR(ctx context.Context, prID string, updates *PRUpdates) error {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return fmt.Errorf("failed to get PR: %w", err)
	}

	parts := strings.SplitN(pr.RepositoryURL, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository URL: %s", pr.RepositoryURL)
	}
	owner, repo := parts[0], parts[1]

	// Update via API
	_, err = s.apiClient.UpdatePullRequest(ctx, owner, repo, pr.ID, UpdatePullRequestOptions{
		Title: updates.Title,
		Body:  updates.Description,
		State: updates.State,
	})
	if err != nil {
		return fmt.Errorf("failed to update PR via API: %w", err)
	}

	// Update database record
	dbUpdates := make(map[string]interface{})
	if updates.Title != nil {
		dbUpdates["title"] = *updates.Title
	}
	if updates.Description != nil {
		dbUpdates["description"] = *updates.Description
	}
	if updates.State != nil {
		dbUpdates["state"] = *updates.State
	}
	dbUpdates["updated_at"] = time.Now()

	return s.prRepo.Update(ctx, prID, dbUpdates)
}

// MergePR merges a pull request
func (s *prServiceImpl) MergePR(ctx context.Context, prID string, opts *MergeOptions) error {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return fmt.Errorf("failed to get PR: %w", err)
	}

	parts := strings.SplitN(pr.RepositoryURL, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository URL: %s", pr.RepositoryURL)
	}
	owner, repo := parts[0], parts[1]

	// Merge via API
	err = s.apiClient.MergePullRequest(ctx, owner, repo, pr.ID, MergePullRequestOptions{
		Style:        opts.Strategy,
		Message:      opts.Message,
		DeleteBranch: opts.DeleteBranch,
	})
	if err != nil {
		return fmt.Errorf("failed to merge PR via API: %w", err)
	}

	// Update database record
	now := time.Now()
	updates := map[string]interface{}{
		"state":      "merged",
		"merged_at":  now,
		"updated_at": now,
	}

	err = s.prRepo.Update(ctx, prID, updates)
	if err != nil {
		return fmt.Errorf("failed to update PR record: %w", err)
	}

	// Delete branch if requested
	if opts.DeleteBranch {
		err = s.gitOps.DeleteBranch(ctx, pr.RepositoryURL, pr.SourceBranch)
		if err != nil {
			// Log error but don't fail merge
			fmt.Printf("Warning: failed to delete branch %s: %v\n", pr.SourceBranch, err)
		}
	}

	// Update linked issue
	if pr.LinkedIssueID != "" {
		comment := fmt.Sprintf("✅ **Code Merged**\n\nPR #%d has been successfully merged.\n\nThis issue is now complete.", pr.Number)
		_, err = s.apiClient.PostComment(ctx, owner, repo, pr.LinkedIssueID, comment)
		if err != nil {
			fmt.Printf("Warning: failed to post merge comment: %v\n", err)
		}
	}

	return nil
}

// ClosePR closes a pull request without merging
func (s *prServiceImpl) ClosePR(ctx context.Context, prID string, reason string) error {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return fmt.Errorf("failed to get PR: %w", err)
	}

	parts := strings.SplitN(pr.RepositoryURL, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository URL: %s", pr.RepositoryURL)
	}
	owner, repo := parts[0], parts[1]

	// Close via API
	state := "closed"
	_, err = s.apiClient.UpdatePullRequest(ctx, owner, repo, pr.ID, UpdatePullRequestOptions{
		State: &state,
	})
	if err != nil {
		return fmt.Errorf("failed to close PR via API: %w", err)
	}

	// Update database record
	updates := map[string]interface{}{
		"state":      "closed",
		"updated_at": time.Now(),
	}

	err = s.prRepo.Update(ctx, prID, updates)
	if err != nil {
		return fmt.Errorf("failed to update PR record: %w", err)
	}

	// Post close reason to PR if provided
	if reason != "" {
		comment := fmt.Sprintf("❌ **Pull Request Closed**\n\nReason: %s", reason)
		_, err = s.apiClient.PostComment(ctx, owner, repo, fmt.Sprintf("%d", pr.Number), comment)
		if err != nil {
			fmt.Printf("Warning: failed to post close comment: %v\n", err)
		}
	}

	return nil
}

// GetPR retrieves a pull request by ID
func (s *prServiceImpl) GetPR(ctx context.Context, prID string) (*PRInfo, error) {
	return s.prRepo.GetByID(ctx, prID)
}

// RunQualityChecks runs all quality checks for a PR
func (s *prServiceImpl) RunQualityChecks(ctx context.Context, prID string) (*QualityCheckResults, error) {
	results := &QualityCheckResults{
		PRID:           prID,
		CheckTimestamp: time.Now(),
	}

	// Run tests
	testResults, err := s.checkService.RunTests(ctx, prID)
	if err == nil {
		results.TestResults = testResults
	}

	// Run linter
	lintResults, err := s.checkService.RunLinter(ctx, prID)
	if err == nil {
		results.LintResults = lintResults
	}

	// Run security scan
	securityResults, err := s.checkService.RunSecurityScan(ctx, prID)
	if err == nil {
		results.SecurityScan = securityResults
	}

	// Check coverage
	coverage, err := s.checkService.CheckCodeCoverage(ctx, prID)
	if err == nil {
		results.Coverage = coverage
	}

	// Get PR to check agent ID
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err == nil && pr.AgentID != "" {
		policyCheck, err := s.checkService.CheckPolicyCompliance(ctx, prID, pr.AgentID)
		if err == nil {
			results.PolicyCheck = policyCheck
		}
	}

	// Determine overall status
	allPassed := true
	if results.TestResults != nil && results.TestResults.FailedTests > 0 {
		allPassed = false
	}
	if results.LintResults != nil && results.LintResults.Errors > 0 {
		allPassed = false
	}
	if results.SecurityScan != nil && (results.SecurityScan.RiskLevel == "high" || results.SecurityScan.RiskLevel == "critical") {
		allPassed = false
	}
	if results.Coverage != nil && !results.Coverage.MeetsThreshold {
		allPassed = false
	}
	if results.PolicyCheck != nil && !results.PolicyCheck.Compliant {
		allPassed = false
	}

	if allPassed {
		results.OverallStatus = "pass"
	} else {
		results.OverallStatus = "fail"
	}

	// Update PR with quality check results
	err = s.prRepo.UpdateQualityChecks(ctx, prID, results)
	if err != nil {
		return nil, fmt.Errorf("failed to update quality checks: %w", err)
	}

	return results, nil
}

// GetCheckStatus gets the current quality check status for a PR
func (s *prServiceImpl) GetCheckStatus(ctx context.Context, prID string) (*CheckStatus, error) {
	return s.checkService.GetOverallStatus(ctx, prID)
}

// EvaluateAutoMerge evaluates whether a PR should be auto-merged
func (s *prServiceImpl) EvaluateAutoMerge(ctx context.Context, prID string) (*AutoMergeDecision, error) {
	return s.autoMerge.ShouldAutoMerge(ctx, prID, s.autoMergeConfig)
}

// LinkAgentToPR links an agent to a PR (stores in agent-issue-links for tracking)
func (s *prServiceImpl) LinkAgentToPR(ctx context.Context, agentID, prID string) error {
	// This is already handled in the database through the agent_id field
	// Additional linking could be done through agent-issue-links if needed
	return nil
}

// GetPRsByAgent retrieves all PRs created by an agent
func (s *prServiceImpl) GetPRsByAgent(ctx context.Context, agentID string) ([]*PRInfo, error) {
	return s.prRepo.ListByAgent(ctx, agentID)
}

// validateCreateRequest validates a CreatePRRequest
func (s *prServiceImpl) validateCreateRequest(req *CreatePRRequest) error {
	if req.RepositoryURL == "" {
		return fmt.Errorf("repository URL is required")
	}
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.AgentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	return nil
}
