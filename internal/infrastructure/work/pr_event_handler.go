package work

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// PREventHandler handles PR-related webhook events and triggers appropriate actions
type PREventHandler struct {
	prService       PRService
	qualityCheckSvc QualityCheckService
	autoMergeEngine AutoMergeEngine
	syncService     SyncService
	prRepository    PRRepository
	agentLinkRepo   AgentIssueLinkRepository
}

// NewPREventHandler creates a new PR event handler
func NewPREventHandler(
	prService PRService,
	qualityCheckSvc QualityCheckService,
	autoMergeEngine AutoMergeEngine,
	syncService SyncService,
	prRepository PRRepository,
	agentLinkRepo AgentIssueLinkRepository,
) *PREventHandler {
	return &PREventHandler{
		prService:       prService,
		qualityCheckSvc: qualityCheckSvc,
		autoMergeEngine: autoMergeEngine,
		syncService:     syncService,
		prRepository:    prRepository,
		agentLinkRepo:   agentLinkRepo,
	}
}

// PRWebhookEvent represents a PR webhook event payload
type PRWebhookEvent struct {
	Action      string           // opened, synchronized, reopened, closed, merged, edited, approved, review_requested
	PullRequest *WorkPullRequest // PR data
	Repository  *RepositoryInfo  // Repository info
	RequestedBy string           // User who triggered the action
	Review      *ReviewInfo      // Review data (for approved/changes_requested events)
	Changes     *PRChanges       // What changed (for edited events)
}

// RepositoryInfo contains repository metadata
type RepositoryInfo struct {
	FullName      string // e.g., "owner/repo"
	CloneURL      string
	HTMLURL       string
	DefaultBranch string
}

// ReviewInfo contains PR review data
type ReviewInfo struct {
	State       string // approved, changes_requested, commented
	SubmittedBy string
	Body        string
	SubmittedAt time.Time
}

// PRChanges represents what changed in an edit event
type PRChanges struct {
	Title       *FieldChange
	Description *FieldChange
	BaseBranch  *FieldChange
}

// FieldChange represents a field modification
type FieldChange struct {
	From string
	To   string
}

// HandlePREvent processes a PR webhook event
func (h *PREventHandler) HandlePREvent(ctx context.Context, event *PRWebhookEvent) error {
	log.WithFields(log.Fields{
		"action":    event.Action,
		"pr_id":     event.PullRequest.PRID,
		"pr_number": event.PullRequest.PRNumber,
	}).Info("Processing PR webhook event")

	switch event.Action {
	case "opened":
		return h.handlePROpened(ctx, event)
	case "synchronized":
		return h.handlePRSynchronized(ctx, event)
	case "reopened":
		return h.handlePRReopened(ctx, event)
	case "closed":
		return h.handlePRClosed(ctx, event)
	case "merged":
		return h.handlePRMerged(ctx, event)
	case "edited":
		return h.handlePREdited(ctx, event)
	case "approved":
		return h.handlePRApproved(ctx, event)
	case "changes_requested":
		return h.handleChangesRequested(ctx, event)
	case "review_requested":
		return h.handleReviewRequested(ctx, event)
	default:
		log.WithField("action", event.Action).Warn("Unknown PR action, ignoring")
		return nil
	}
}

// handlePROpened processes a newly opened PR
func (h *PREventHandler) handlePROpened(ctx context.Context, event *PRWebhookEvent) error {
	log.WithField("pr_id", event.PullRequest.PRID).Info("PR opened, storing metadata")

	// Store PR in database
	prInfo := h.convertWorkPRToPRInfo(event.PullRequest)
	prInfo.CreatedAt = time.Now()
	prInfo.UpdatedAt = time.Now()

	// Find linked agent (if PR was created by an agent)
	agentID, err := h.findAgentFromPR(ctx, event.PullRequest)
	if err != nil {
		log.WithError(err).Warn("Could not find linked agent, continuing anyway")
	} else if agentID != "" {
		prInfo.AgentID = agentID
	}

	if err := h.prRepository.Create(ctx, prInfo); err != nil {
		return fmt.Errorf("failed to store PR: %w", err)
	}

	// Trigger quality checks asynchronously
	go func() {
		checkCtx := context.Background()
		if err := h.runQualityChecksAndEvaluate(checkCtx, prInfo.ID); err != nil {
			log.WithError(err).WithField("pr_id", prInfo.ID).Error("Failed to run quality checks")
		}
	}()

	return nil
}

// handlePRSynchronized processes push events to PR branch
func (h *PREventHandler) handlePRSynchronized(ctx context.Context, event *PRWebhookEvent) error {
	log.WithField("pr_id", event.PullRequest.PRID).Info("PR synchronized (new commits pushed)")

	// Update timestamp in database
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if err := h.prRepository.Update(ctx, event.PullRequest.PRID, updates); err != nil {
		return fmt.Errorf("failed to update PR: %w", err)
	}

	// Re-run quality checks on new commits
	go func() {
		checkCtx := context.Background()
		if err := h.runQualityChecksAndEvaluate(checkCtx, event.PullRequest.PRID); err != nil {
			log.WithError(err).WithField("pr_id", event.PullRequest.PRID).Error("Failed to re-run quality checks")
		}
	}()

	return nil
}

// handlePRReopened processes a reopened PR
func (h *PREventHandler) handlePRReopened(ctx context.Context, event *PRWebhookEvent) error {
	log.WithField("pr_id", event.PullRequest.PRID).Info("PR reopened")

	updates := map[string]interface{}{
		"state":      "open",
		"updated_at": time.Now(),
	}

	if err := h.prRepository.Update(ctx, event.PullRequest.PRID, updates); err != nil {
		return fmt.Errorf("failed to update PR state: %w", err)
	}

	// Re-run quality checks
	go func() {
		checkCtx := context.Background()
		if err := h.runQualityChecksAndEvaluate(checkCtx, event.PullRequest.PRID); err != nil {
			log.WithError(err).WithField("pr_id", event.PullRequest.PRID).Error("Failed to re-run quality checks")
		}
	}()

	return nil
}

// handlePRClosed processes a closed (not merged) PR
func (h *PREventHandler) handlePRClosed(ctx context.Context, event *PRWebhookEvent) error {
	log.WithFields(log.Fields{
		"pr_id":  event.PullRequest.PRID,
		"merged": event.PullRequest.Merged,
	}).Info("PR closed")

	// If PR was merged, handlePRMerged will be called instead
	if event.PullRequest.Merged {
		return h.handlePRMerged(ctx, event)
	}

	updates := map[string]interface{}{
		"state":      "closed",
		"updated_at": time.Now(),
	}

	if err := h.prRepository.Update(ctx, event.PullRequest.PRID, updates); err != nil {
		return fmt.Errorf("failed to update PR state: %w", err)
	}

	return nil
}

// handlePRMerged processes a merged PR
func (h *PREventHandler) handlePRMerged(ctx context.Context, event *PRWebhookEvent) error {
	log.WithField("pr_id", event.PullRequest.PRID).Info("PR merged")

	now := time.Now()
	updates := map[string]interface{}{
		"state":      "merged",
		"merged_at":  now,
		"merged_by":  event.RequestedBy,
		"updated_at": now,
	}

	if err := h.prRepository.Update(ctx, event.PullRequest.PRID, updates); err != nil {
		return fmt.Errorf("failed to update PR state: %w", err)
	}

	return nil
}

// handlePREdited processes PR title/description edits
func (h *PREventHandler) handlePREdited(ctx context.Context, event *PRWebhookEvent) error {
	log.WithField("pr_id", event.PullRequest.PRID).Info("PR edited")

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	// Update fields that changed
	if event.Changes != nil {
		if event.Changes.Title != nil {
			updates["title"] = event.Changes.Title.To
		}
		if event.Changes.Description != nil {
			updates["description"] = event.Changes.Description.To
		}
		if event.Changes.BaseBranch != nil {
			updates["target_branch"] = event.Changes.BaseBranch.To
		}
	}

	return h.prRepository.Update(ctx, event.PullRequest.PRID, updates)
}

// handlePRApproved processes a PR approval
func (h *PREventHandler) handlePRApproved(ctx context.Context, event *PRWebhookEvent) error {
	log.WithFields(log.Fields{
		"pr_id":    event.PullRequest.PRID,
		"reviewer": event.Review.SubmittedBy,
	}).Info("PR approved")

	// Get PR info to check auto-merge status
	prInfo, err := h.prRepository.GetByID(ctx, event.PullRequest.PRID)
	if err != nil {
		return fmt.Errorf("failed to get PR: %w", err)
	}

	// Check if auto-merge should be triggered
	if prInfo.AutoMergeEnabled {
		go func() {
			mergeCtx := context.Background()
			if err := h.evaluateAndAutoMerge(mergeCtx, prInfo.ID); err != nil {
				log.WithError(err).WithField("pr_id", prInfo.ID).Error("Auto-merge evaluation failed")
			}
		}()
	}

	return nil
}

// handleChangesRequested processes a request for changes
func (h *PREventHandler) handleChangesRequested(ctx context.Context, event *PRWebhookEvent) error {
	log.WithFields(log.Fields{
		"pr_id":    event.PullRequest.PRID,
		"reviewer": event.Review.SubmittedBy,
	}).Info("Changes requested on PR")

	// Auto-merge should not proceed if changes are requested
	// Quality checks will reflect this
	return nil
}

// handleReviewRequested processes a review request
func (h *PREventHandler) handleReviewRequested(ctx context.Context, event *PRWebhookEvent) error {
	log.WithField("pr_id", event.PullRequest.PRID).Info("Review requested on PR")

	// Update metadata if needed
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	return h.prRepository.Update(ctx, event.PullRequest.PRID, updates)
}

// runQualityChecksAndEvaluate runs all quality checks and evaluates auto-merge
func (h *PREventHandler) runQualityChecksAndEvaluate(ctx context.Context, prID string) error {
	log.WithField("pr_id", prID).Info("Running quality checks")

	// Run all quality checks via PR service
	checks, err := h.prService.RunQualityChecks(ctx, prID)
	if err != nil {
		return fmt.Errorf("failed to run quality checks: %w", err)
	}

	// Get PR info
	prInfo, err := h.prRepository.GetByID(ctx, prID)
	if err != nil {
		return fmt.Errorf("failed to get PR info: %w", err)
	}

	// Evaluate auto-merge if enabled and checks passed
	if prInfo.AutoMergeEnabled && checks.OverallStatus == "pass" {
		return h.evaluateAndAutoMerge(ctx, prID)
	}

	return nil
}

// evaluateAndAutoMerge evaluates auto-merge criteria and merges if appropriate
func (h *PREventHandler) evaluateAndAutoMerge(ctx context.Context, prID string) error {
	log.WithField("pr_id", prID).Info("Evaluating auto-merge criteria")

	// Get auto-merge configuration (use defaults if not specified)
	config := &AutoMergeConfig{
		Enabled:             true,
		RequireApproval:     true,
		MinApprovals:        1,
		RequireTestsPass:    true,
		RequireLintPass:     true,
		RequireSecurityScan: true,
		MinCoveragePercent:  80.0,
		BlockOnHighVulns:    true,
		MergeStrategy:       "squash",
		DeleteBranchAfter:   true,
	}

	// Evaluate auto-merge via auto-merge engine
	decision, err := h.autoMergeEngine.ShouldAutoMerge(ctx, prID, config)
	if err != nil {
		return fmt.Errorf("failed to evaluate auto-merge: %w", err)
	}

	if !decision.ShouldMerge {
		log.WithFields(log.Fields{
			"pr_id":      prID,
			"blocked_by": decision.BlockedBy,
		}).Info("Auto-merge blocked")
		return nil
	}

	log.WithFields(log.Fields{
		"pr_id":          prID,
		"merge_strategy": config.MergeStrategy,
	}).Info("Auto-merging PR")

	// Trigger merge via PR service
	opts := &MergeOptions{
		Strategy:     config.MergeStrategy,
		DeleteBranch: config.DeleteBranchAfter,
		Message:      "Auto-merged by system",
	}

	err = h.prService.MergePR(ctx, prID, opts)
	if err != nil {
		return fmt.Errorf("failed to auto-merge PR: %w", err)
	}

	return nil
}

// findAgentFromPR attempts to find the agent ID from PR metadata
func (h *PREventHandler) findAgentFromPR(ctx context.Context, pr *WorkPullRequest) (string, error) {
	// Strategy 1: Check if PR branch follows agent/{agentID}/... pattern
	// Strategy 2: Check if PR is linked to an issue, then check agent-issue links
	// Strategy 3: Check PR creator metadata

	// For now, return empty (will be implemented when integration is tested)
	return "", nil
}

// convertWorkPRToPRInfo converts a WorkPullRequest to PRInfo
func (h *PREventHandler) convertWorkPRToPRInfo(pr *WorkPullRequest) *PRInfo {
	return &PRInfo{
		ID:            pr.PRID,
		Number:        pr.PRNumber,
		Title:         pr.Title,
		Description:   pr.Body,
		RepositoryURL: pr.RepoURL,
		SourceBranch:  pr.HeadBranch,
		TargetBranch:  pr.BaseBranch,
		State:         pr.State,
		CreatedBy:     pr.AuthorUsername,
		CreatedAt:     pr.CreatedAt,
		UpdatedAt:     pr.UpdatedAt,
		MergedAt:      pr.MergedAt,
	}
}
