package work

import "time"

// PRInfo represents a pull request in the system
type PRInfo struct {
	ID            string     `json:"_key"`
	Number        int64      `json:"number"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	RepositoryURL string     `json:"repository_url"`
	SourceBranch  string     `json:"source_branch"`
	TargetBranch  string     `json:"target_branch"`
	State         string     `json:"state"` // open, merged, closed
	CreatedBy     string     `json:"created_by"`
	AgentID       string     `json:"agent_id"`
	LinkedIssueID string     `json:"linked_issue_id"`
	CommitSHA     string     `json:"commit_sha"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	MergedAt      *time.Time `json:"merged_at,omitempty"`
	MergedBy      string     `json:"merged_by,omitempty"`

	// Auto-merge configuration
	AutoMergeEnabled bool `json:"auto_merge_enabled"`

	// Quality check results (embedded for convenience)
	QualityChecks *QualityCheckResults `json:"quality_checks,omitempty"`
}

// CreatePRRequest represents a request to create a new pull request
type CreatePRRequest struct {
	RepositoryURL string                 `json:"repository_url"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	SourceBranch  string                 `json:"source_branch"`
	TargetBranch  string                 `json:"target_branch"` // default: main/master
	AgentID       string                 `json:"agent_id"`
	IssueID       string                 `json:"issue_id"`
	Changes       *ChangeSet             `json:"changes"`
	AutoMerge     bool                   `json:"auto_merge"` // enable auto-merge
	Labels        []string               `json:"labels,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ChangeSet represents a set of file changes
type ChangeSet struct {
	Files []FileChange `json:"files"`
}

// FileChange represents a single file change
type FileChange struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Operation string `json:"operation"` // create, update, delete
	OldPath   string `json:"old_path,omitempty"`
}

// PRResult represents the result of PR creation
type PRResult struct {
	PRID       string `json:"pr_id"`
	PRNumber   int64  `json:"pr_number"`
	URL        string `json:"url"`
	BranchName string `json:"branch_name"`
}

// PRUpdates represents updates to apply to a PR
type PRUpdates struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	State       *string  `json:"state,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}

// MergeOptions represents options for merging a PR
type MergeOptions struct {
	Strategy     string `json:"strategy"` // merge, squash, rebase
	DeleteBranch bool   `json:"delete_branch"`
	Message      string `json:"message,omitempty"`
}

// QualityCheckResults represents the results of all quality checks
type QualityCheckResults struct {
	PRID           string             `json:"pr_id"`
	CheckTimestamp time.Time          `json:"check_timestamp"`
	TestResults    *TestResults       `json:"test_results,omitempty"`
	LintResults    *LintResults       `json:"lint_results,omitempty"`
	SecurityScan   *SecurityResults   `json:"security_scan,omitempty"`
	Coverage       *CoverageReport    `json:"coverage,omitempty"`
	PolicyCheck    *PolicyCheckResult `json:"policy_check,omitempty"`
	OverallStatus  string             `json:"overall_status"` // pass, fail, pending
}

// TestResults represents test execution results
type TestResults struct {
	TotalTests  int      `json:"total_tests"`
	PassedTests int      `json:"passed_tests"`
	FailedTests int      `json:"failed_tests"`
	Duration    string   `json:"duration"`
	Failures    []string `json:"failures,omitempty"`
}

// LintResults represents linting results
type LintResults struct {
	TotalIssues int         `json:"total_issues"`
	Errors      int         `json:"errors"`
	Warnings    int         `json:"warnings"`
	Issues      []LintIssue `json:"issues"`
}

// LintIssue represents a single linting issue
type LintIssue struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
}

// SecurityResults represents security scan results
type SecurityResults struct {
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	RiskLevel       string          `json:"risk_level"` // low, medium, high, critical
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string `json:"id"`
	Package     string `json:"package"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	FixVersion  string `json:"fix_version,omitempty"`
}

// CoverageReport represents code coverage results
type CoverageReport struct {
	TotalLines     int     `json:"total_lines"`
	CoveredLines   int     `json:"covered_lines"`
	Percentage     float64 `json:"percentage"`
	MeetsThreshold bool    `json:"meets_threshold"`
}

// PolicyCheckResult represents policy compliance check result
type PolicyCheckResult struct {
	Compliant  bool      `json:"compliant"`
	Violations []string  `json:"violations,omitempty"`
	CheckedAt  time.Time `json:"checked_at"`
}

// AutoMergeDecision represents the decision whether to auto-merge
type AutoMergeDecision struct {
	ShouldMerge   bool            `json:"should_merge"`
	Reason        string          `json:"reason"`
	ChecksPassed  map[string]bool `json:"checks_passed"`
	BlockedBy     []string        `json:"blocked_by,omitempty"`
	MergeStrategy string          `json:"merge_strategy"` // merge, squash, rebase
}

// AutoMergeConfig represents auto-merge configuration
type AutoMergeConfig struct {
	Enabled             bool    `json:"enabled"`
	RequireApproval     bool    `json:"require_approval"`
	MinApprovals        int     `json:"min_approvals"`
	RequireTestsPass    bool    `json:"require_tests_pass"`
	RequireLintPass     bool    `json:"require_lint_pass"`
	RequireSecurityScan bool    `json:"require_security_scan"`
	MinCoveragePercent  float64 `json:"min_coverage_percent"`
	BlockOnHighVulns    bool    `json:"block_on_high_vulns"`
	MergeStrategy       string  `json:"merge_strategy"` // merge, squash, rebase
	DeleteBranchAfter   bool    `json:"delete_branch_after"`
}

// CheckStatus represents the current status of quality checks
type CheckStatus struct {
	PRID          string    `json:"pr_id"`
	TestsPassed   bool      `json:"tests_passed"`
	LintPassed    bool      `json:"lint_passed"`
	SecurityOK    bool      `json:"security_ok"`
	CoverageOK    bool      `json:"coverage_ok"`
	PolicyOK      bool      `json:"policy_ok"`
	AllPassed     bool      `json:"all_passed"`
	LastCheckedAt time.Time `json:"last_checked_at"`
}

// PRCreationError represents errors during PR creation
type PRCreationError struct {
	Reason string
	Code   string
}

func (e *PRCreationError) Error() string {
	return e.Reason
}

// Error codes for PR operations
const (
	ErrBranchExists   = "BRANCH_EXISTS"
	ErrPushFailed     = "PUSH_FAILED"
	ErrAPIError       = "API_ERROR"
	ErrValidationFail = "VALIDATION_FAIL"
	ErrConflict       = "MERGE_CONFLICT"
)
