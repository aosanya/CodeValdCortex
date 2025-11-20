package work

import (
	"context"
	"fmt"
	"strings"
)

// autoMergeEngineImpl implements AutoMergeEngine interface
type autoMergeEngineImpl struct {
	prRepo       PRRepository
	checkService QualityCheckService
	apiClient    APIClient
}

// NewAutoMergeEngine creates a new auto-merge engine
func NewAutoMergeEngine(prRepo PRRepository, checkService QualityCheckService, apiClient APIClient) AutoMergeEngine {
	return &autoMergeEngineImpl{
		prRepo:       prRepo,
		checkService: checkService,
		apiClient:    apiClient,
	}
}

// ShouldAutoMerge evaluates whether a PR should be auto-merged
func (e *autoMergeEngineImpl) ShouldAutoMerge(ctx context.Context, prID string, config *AutoMergeConfig) (*AutoMergeDecision, error) {
	if !config.Enabled {
		return &AutoMergeDecision{
			ShouldMerge:   false,
			Reason:        "Auto-merge disabled",
			ChecksPassed:  make(map[string]bool),
			BlockedBy:     []string{"auto_merge_disabled"},
			MergeStrategy: config.MergeStrategy,
		}, nil
	}

	decision := &AutoMergeDecision{
		ChecksPassed:  make(map[string]bool),
		BlockedBy:     []string{},
		MergeStrategy: config.MergeStrategy,
	}

	// Get PR info (for future use in conflict checking)
	_, err := e.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	// Check for merge conflicts
	hasConflicts, conflicts, err := e.HasMergeConflicts(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to check conflicts: %w", err)
	}
	if hasConflicts {
		decision.ChecksPassed["no_conflicts"] = false
		decision.BlockedBy = append(decision.BlockedBy, fmt.Sprintf("merge_conflicts (%d files)", len(conflicts)))
	} else {
		decision.ChecksPassed["no_conflicts"] = true
	}

	// Get quality check status
	status, err := e.checkService.GetOverallStatus(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get check status: %w", err)
	}

	// Check tests
	if config.RequireTestsPass {
		decision.ChecksPassed["tests"] = status.TestsPassed
		if !status.TestsPassed {
			decision.BlockedBy = append(decision.BlockedBy, "tests_failing")
		}
	}

	// Check lint
	if config.RequireLintPass {
		decision.ChecksPassed["lint"] = status.LintPassed
		if !status.LintPassed {
			decision.BlockedBy = append(decision.BlockedBy, "lint_errors")
		}
	}

	// Check security
	if config.RequireSecurityScan {
		decision.ChecksPassed["security"] = status.SecurityOK
		if !status.SecurityOK && config.BlockOnHighVulns {
			decision.BlockedBy = append(decision.BlockedBy, "high_vulnerabilities")
		}
	}

	// Check coverage
	if config.MinCoveragePercent > 0 {
		decision.ChecksPassed["coverage"] = status.CoverageOK
		if !status.CoverageOK {
			decision.BlockedBy = append(decision.BlockedBy, "insufficient_coverage")
		}
	}

	// Check policy compliance
	decision.ChecksPassed["policy"] = status.PolicyOK
	if !status.PolicyOK {
		decision.BlockedBy = append(decision.BlockedBy, "policy_violation")
	}

	// Check approvals
	if config.RequireApproval {
		approvals, err := e.GetApprovalCount(ctx, prID)
		if err != nil {
			return nil, fmt.Errorf("failed to get approval count: %w", err)
		}
		decision.ChecksPassed["approvals"] = approvals >= config.MinApprovals
		if approvals < config.MinApprovals {
			decision.BlockedBy = append(decision.BlockedBy, fmt.Sprintf("pending_approval (%d/%d)", approvals, config.MinApprovals))
		}
	}

	// Final decision
	decision.ShouldMerge = len(decision.BlockedBy) == 0
	if decision.ShouldMerge {
		decision.Reason = "All quality checks passed"
	} else {
		decision.Reason = fmt.Sprintf("Blocked by: %s", strings.Join(decision.BlockedBy, ", "))
	}

	return decision, nil
}

// GetApprovalCount gets the number of approvals for a PR
// TODO: Integrate with Gitea PR review API
func (e *autoMergeEngineImpl) GetApprovalCount(ctx context.Context, prID string) (int, error) {
	// Stub implementation - returns 0
	// In production, this would:
	// 1. Get PR from repository
	// 2. Query Gitea for PR reviews
	// 3. Count approved reviews
	// 4. Return count

	return 0, nil
}

// HasMergeConflicts checks if a PR has merge conflicts
// TODO: Integrate with Gitea API to detect conflicts
func (e *autoMergeEngineImpl) HasMergeConflicts(ctx context.Context, prID string) (bool, []string, error) {
	// Stub implementation - returns no conflicts
	// In production, this would:
	// 1. Get PR from repository
	// 2. Query Gitea for mergeable status
	// 3. If not mergeable, get conflict files
	// 4. Return conflict status and files

	return false, []string{}, nil
}
