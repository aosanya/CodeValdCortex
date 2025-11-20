package work

import "context"

// PRService handles pull request operations
type PRService interface {
	// PR Creation
	CreatePR(ctx context.Context, req *CreatePRRequest) (*PRResult, error)

	// PR Management
	UpdatePR(ctx context.Context, prID string, updates *PRUpdates) error
	MergePR(ctx context.Context, prID string, opts *MergeOptions) error
	ClosePR(ctx context.Context, prID string, reason string) error
	GetPR(ctx context.Context, prID string) (*PRInfo, error)

	// Quality Checks
	RunQualityChecks(ctx context.Context, prID string) (*QualityCheckResults, error)
	GetCheckStatus(ctx context.Context, prID string) (*CheckStatus, error)

	// Auto-Merge Logic
	EvaluateAutoMerge(ctx context.Context, prID string) (*AutoMergeDecision, error)

	// Agent Attribution
	LinkAgentToPR(ctx context.Context, agentID, prID string) error
	GetPRsByAgent(ctx context.Context, agentID string) ([]*PRInfo, error)
}

// GitOperations handles Git operations for PR workflow
type GitOperations interface {
	// Branch Management
	CreateBranch(ctx context.Context, repo, baseBranch, newBranch string) error
	PushChanges(ctx context.Context, repo, branch string, changes *ChangeSet) error
	DeleteBranch(ctx context.Context, repo, branch string) error

	// File Operations
	CreateFile(ctx context.Context, repo, branch, path, content string) error
	UpdateFile(ctx context.Context, repo, branch, path, content string) error
	DeleteFile(ctx context.Context, repo, branch, path string) error

	// Commit Operations
	CreateCommit(ctx context.Context, repo, branch, message string, files []FileChange) error
}

// QualityCheckService handles quality checks for PRs
type QualityCheckService interface {
	// Test Execution
	RunTests(ctx context.Context, prID string) (*TestResults, error)

	// Code Quality
	RunLinter(ctx context.Context, prID string) (*LintResults, error)
	RunSecurityScan(ctx context.Context, prID string) (*SecurityResults, error)

	// Coverage Analysis
	CheckCodeCoverage(ctx context.Context, prID string) (*CoverageReport, error)

	// Policy Compliance
	CheckPolicyCompliance(ctx context.Context, prID string, agentID string) (*PolicyCheckResult, error)

	// Get overall status
	GetOverallStatus(ctx context.Context, prID string) (*CheckStatus, error)
}

// AutoMergeEngine evaluates whether a PR should be auto-merged
type AutoMergeEngine interface {
	// Evaluate if PR is ready for auto-merge
	ShouldAutoMerge(ctx context.Context, prID string, config *AutoMergeConfig) (*AutoMergeDecision, error)

	// Get approval count for PR
	GetApprovalCount(ctx context.Context, prID string) (int, error)

	// Check for merge conflicts
	HasMergeConflicts(ctx context.Context, prID string) (bool, []string, error)
}

// PRRepository handles persistence of PR data
type PRRepository interface {
	// Create new PR record
	Create(ctx context.Context, pr *PRInfo) error

	// Update existing PR
	Update(ctx context.Context, prID string, updates map[string]interface{}) error

	// Get PR by ID
	GetByID(ctx context.Context, prID string) (*PRInfo, error)

	// Get PR by number
	GetByNumber(ctx context.Context, repoURL string, number int64) (*PRInfo, error)

	// List PRs by agent
	ListByAgent(ctx context.Context, agentID string) ([]*PRInfo, error)

	// List PRs by issue
	ListByIssue(ctx context.Context, issueID string) ([]*PRInfo, error)

	// List PRs by state
	ListByState(ctx context.Context, state string) ([]*PRInfo, error)

	// Delete PR record
	Delete(ctx context.Context, prID string) error

	// Update quality checks
	UpdateQualityChecks(ctx context.Context, prID string, checks *QualityCheckResults) error
}
