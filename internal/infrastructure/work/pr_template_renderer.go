package work

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PRTemplateRenderer generates PR descriptions and comments
type PRTemplateRenderer struct {
	prRepo        PRRepository
	agentLinkRepo AgentIssueLinkRepository
}

// NewPRTemplateRenderer creates a new PR template renderer
func NewPRTemplateRenderer(
	prRepo PRRepository,
	agentLinkRepo AgentIssueLinkRepository,
) *PRTemplateRenderer {
	return &PRTemplateRenderer{
		prRepo:        prRepo,
		agentLinkRepo: agentLinkRepo,
	}
}

// RenderPRDescription generates a PR description with agent info and context
func (r *PRTemplateRenderer) RenderPRDescription(ctx context.Context, req *CreatePRRequest) (string, error) {
	var parts []string

	// Add title section
	parts = append(parts, fmt.Sprintf("## %s", req.Title))
	parts = append(parts, "")

	// Add description if provided
	if req.Description != "" {
		parts = append(parts, req.Description)
		parts = append(parts, "")
	}

	// Add agent attribution
	if req.AgentID != "" {
		parts = append(parts, "---")
		parts = append(parts, "### 🤖 Agent Information")
		parts = append(parts, fmt.Sprintf("- **Agent ID**: `%s`", req.AgentID))
		if req.IssueID != "" {
			parts = append(parts, fmt.Sprintf("- **Linked Issue**: #%s", req.IssueID))
		}
		parts = append(parts, "")
	}

	// Add changes summary
	if req.Changes != nil && len(req.Changes.Files) > 0 {
		parts = append(parts, "### 📝 Changes Summary")
		parts = append(parts, fmt.Sprintf("- **Files Changed**: %d", len(req.Changes.Files)))

		// Group changes by operation
		creates := 0
		updates := 0
		deletes := 0
		for _, file := range req.Changes.Files {
			switch file.Operation {
			case "create":
				creates++
			case "update":
				updates++
			case "delete":
				deletes++
			}
		}

		if creates > 0 {
			parts = append(parts, fmt.Sprintf("- **New Files**: %d", creates))
		}
		if updates > 0 {
			parts = append(parts, fmt.Sprintf("- **Modified Files**: %d", updates))
		}
		if deletes > 0 {
			parts = append(parts, fmt.Sprintf("- **Deleted Files**: %d", deletes))
		}
		parts = append(parts, "")

		// List files
		parts = append(parts, "#### Modified Files")
		for _, file := range req.Changes.Files {
			var emoji string
			switch file.Operation {
			case "create":
				emoji = "✨"
			case "update":
				emoji = "📝"
			case "delete":
				emoji = "🗑️"
			default:
				emoji = "📄"
			}
			parts = append(parts, fmt.Sprintf("- %s `%s`", emoji, file.Path))
		}
		parts = append(parts, "")
	}

	// Add auto-merge status
	if req.AutoMerge {
		parts = append(parts, "---")
		parts = append(parts, "### ⚡ Auto-Merge Enabled")
		parts = append(parts, "This PR will be automatically merged when:")
		parts = append(parts, "- ✅ All tests pass")
		parts = append(parts, "- ✅ Linting checks pass")
		parts = append(parts, "- ✅ Security scan is clean")
		parts = append(parts, "- ✅ Code coverage meets threshold (≥80%)")
		parts = append(parts, "- ✅ Required approvals obtained")
		parts = append(parts, "")
	}

	// Add labels if provided
	if len(req.Labels) > 0 {
		parts = append(parts, "---")
		parts = append(parts, fmt.Sprintf("**Labels**: %s", strings.Join(req.Labels, ", ")))
		parts = append(parts, "")
	}

	// Add footer
	parts = append(parts, "---")
	parts = append(parts, fmt.Sprintf("*Generated at %s*", time.Now().Format("2006-01-02 15:04:05")))

	return strings.Join(parts, "\n"), nil
}

// RenderQualityCheckStatus generates a markdown summary of quality check results
func (r *PRTemplateRenderer) RenderQualityCheckStatus(ctx context.Context, prID string) (string, error) {
	// Get PR info with quality checks
	prInfo, err := r.prRepo.GetByID(ctx, prID)
	if err != nil {
		return "", fmt.Errorf("failed to get PR: %w", err)
	}

	if prInfo.QualityChecks == nil {
		return "⏳ Quality checks not yet run", nil
	}

	checks := prInfo.QualityChecks
	var parts []string

	parts = append(parts, "## 🔍 Quality Check Results")
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("**Overall Status**: %s", r.formatOverallStatus(checks.OverallStatus)))
	parts = append(parts, fmt.Sprintf("**Last Checked**: %s", checks.CheckTimestamp.Format("2006-01-02 15:04:05")))
	parts = append(parts, "")

	// Test results
	if checks.TestResults != nil {
		parts = append(parts, r.formatTestResults(checks.TestResults))
		parts = append(parts, "")
	}

	// Lint results
	if checks.LintResults != nil {
		parts = append(parts, r.formatLintResults(checks.LintResults))
		parts = append(parts, "")
	}

	// Security scan
	if checks.SecurityScan != nil {
		parts = append(parts, r.formatSecurityResults(checks.SecurityScan))
		parts = append(parts, "")
	}

	// Code coverage
	if checks.Coverage != nil {
		parts = append(parts, r.formatCoverageResults(checks.Coverage))
		parts = append(parts, "")
	}

	// Policy compliance
	if checks.PolicyCheck != nil {
		parts = append(parts, r.formatPolicyResults(checks.PolicyCheck))
		parts = append(parts, "")
	}

	return strings.Join(parts, "\n"), nil
}

// RenderAutoMergeStatus generates a markdown summary of auto-merge decision
func (r *PRTemplateRenderer) RenderAutoMergeStatus(decision *AutoMergeDecision) string {
	var parts []string

	parts = append(parts, "## ⚡ Auto-Merge Status")
	parts = append(parts, "")

	if decision.ShouldMerge {
		parts = append(parts, "✅ **Ready to Auto-Merge**")
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("**Merge Strategy**: %s", decision.MergeStrategy))
		parts = append(parts, "")
		parts = append(parts, "All checks passed:")
		for check, passed := range decision.ChecksPassed {
			if passed {
				parts = append(parts, fmt.Sprintf("- ✅ %s", check))
			}
		}
	} else {
		parts = append(parts, "⏸️ **Auto-Merge Blocked**")
		parts = append(parts, "")
		if len(decision.BlockedBy) > 0 {
			parts = append(parts, "**Blocked by:**")
			for _, reason := range decision.BlockedBy {
				parts = append(parts, fmt.Sprintf("- ❌ %s", reason))
			}
		}
		parts = append(parts, "")
		if decision.Reason != "" {
			parts = append(parts, fmt.Sprintf("*Reason: %s*", decision.Reason))
		}
	}

	return strings.Join(parts, "\n")
}

// RenderPRCommentForIssue generates a comment to post on the linked issue
func (r *PRTemplateRenderer) RenderPRCommentForIssue(prInfo *PRInfo, action string) string {
	var parts []string

	switch action {
	case "opened":
		parts = append(parts, "## 🔀 Pull Request Opened")
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("**PR #%d**: %s", prInfo.Number, prInfo.Title))
		parts = append(parts, fmt.Sprintf("**Branch**: `%s` → `%s`", prInfo.SourceBranch, prInfo.TargetBranch))
		parts = append(parts, fmt.Sprintf("**Created by**: %s", prInfo.CreatedBy))
		parts = append(parts, "")
		if prInfo.AutoMergeEnabled {
			parts = append(parts, "⚡ Auto-merge is **enabled** for this PR")
		}
		parts = append(parts, "")
		parts = append(parts, "Quality checks will run automatically. Results will be posted here.")

	case "merged":
		parts = append(parts, "## ✅ Pull Request Merged")
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("**PR #%d** has been successfully merged!", prInfo.Number))
		if prInfo.MergedBy != "" {
			parts = append(parts, fmt.Sprintf("**Merged by**: %s", prInfo.MergedBy))
		}
		if prInfo.MergedAt != nil {
			parts = append(parts, fmt.Sprintf("**Merged at**: %s", prInfo.MergedAt.Format("2006-01-02 15:04:05")))
		}

	case "closed":
		parts = append(parts, "## ❌ Pull Request Closed")
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("**PR #%d** was closed without merging", prInfo.Number))

	default:
		parts = append(parts, fmt.Sprintf("## PR Update: %s", action))
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("**PR #%d**: %s", prInfo.Number, prInfo.Title))
	}

	return strings.Join(parts, "\n")
}

// Helper functions for formatting check results

func (r *PRTemplateRenderer) formatOverallStatus(status string) string {
	switch status {
	case "pass":
		return "✅ **PASS**"
	case "fail":
		return "❌ **FAIL**"
	case "pending":
		return "⏳ **PENDING**"
	default:
		return fmt.Sprintf("❓ %s", strings.ToUpper(status))
	}
}

func (r *PRTemplateRenderer) formatTestResults(results *TestResults) string {
	var parts []string
	parts = append(parts, "### 🧪 Tests")

	passed := results.PassedTests == results.TotalTests
	emoji := "❌"
	if passed {
		emoji = "✅"
	}

	parts = append(parts, fmt.Sprintf("%s **%d/%d tests passed** (%s)",
		emoji, results.PassedTests, results.TotalTests, results.Duration))

	if results.FailedTests > 0 && len(results.Failures) > 0 {
		parts = append(parts, "")
		parts = append(parts, "**Failed tests:**")
		for i, failure := range results.Failures {
			if i >= 5 {
				parts = append(parts, fmt.Sprintf("*...and %d more*", len(results.Failures)-5))
				break
			}
			parts = append(parts, fmt.Sprintf("- %s", failure))
		}
	}

	return strings.Join(parts, "\n")
}

func (r *PRTemplateRenderer) formatLintResults(results *LintResults) string {
	var parts []string
	parts = append(parts, "### 📋 Linting")

	clean := results.TotalIssues == 0
	emoji := "❌"
	if clean {
		emoji = "✅"
	}

	parts = append(parts, fmt.Sprintf("%s **%d total issues** (%d errors, %d warnings)",
		emoji, results.TotalIssues, results.Errors, results.Warnings))

	if results.TotalIssues > 0 && len(results.Issues) > 0 {
		parts = append(parts, "")
		parts = append(parts, "**Top issues:**")
		for i, issue := range results.Issues {
			if i >= 5 {
				parts = append(parts, fmt.Sprintf("*...and %d more*", len(results.Issues)-5))
				break
			}
			parts = append(parts, fmt.Sprintf("- `%s:%d` - %s", issue.File, issue.Line, issue.Message))
		}
	}

	return strings.Join(parts, "\n")
}

func (r *PRTemplateRenderer) formatSecurityResults(results *SecurityResults) string {
	var parts []string
	parts = append(parts, "### 🔒 Security Scan")

	clean := len(results.Vulnerabilities) == 0
	emoji := "❌"
	if clean {
		emoji = "✅"
	}

	riskEmoji := r.getRiskEmoji(results.RiskLevel)
	parts = append(parts, fmt.Sprintf("%s **%d vulnerabilities found** (Risk: %s %s)",
		emoji, len(results.Vulnerabilities), riskEmoji, strings.ToUpper(results.RiskLevel)))

	if len(results.Vulnerabilities) > 0 {
		parts = append(parts, "")
		parts = append(parts, "**Vulnerabilities:**")
		for i, vuln := range results.Vulnerabilities {
			if i >= 5 {
				parts = append(parts, fmt.Sprintf("*...and %d more*", len(results.Vulnerabilities)-5))
				break
			}
			parts = append(parts, fmt.Sprintf("- **%s** (%s): %s in `%s`",
				vuln.Severity, vuln.ID, vuln.Description, vuln.Package))
		}
	}

	return strings.Join(parts, "\n")
}

func (r *PRTemplateRenderer) formatCoverageResults(results *CoverageReport) string {
	var parts []string
	parts = append(parts, "### 📊 Code Coverage")

	emoji := "❌"
	if results.MeetsThreshold {
		emoji = "✅"
	}

	parts = append(parts, fmt.Sprintf("%s **%.1f%%** coverage (%d/%d lines covered)",
		emoji, results.Percentage, results.CoveredLines, results.TotalLines))

	if !results.MeetsThreshold {
		parts = append(parts, "⚠️ *Coverage is below the required threshold*")
	}

	return strings.Join(parts, "\n")
}

func (r *PRTemplateRenderer) formatPolicyResults(results *PolicyCheckResult) string {
	var parts []string
	parts = append(parts, "### 📜 Policy Compliance")

	emoji := "❌"
	if results.Compliant {
		emoji = "✅"
	}

	parts = append(parts, fmt.Sprintf("%s **Compliant**: %t", emoji, results.Compliant))

	if !results.Compliant && len(results.Violations) > 0 {
		parts = append(parts, "")
		parts = append(parts, "**Violations:**")
		for _, violation := range results.Violations {
			parts = append(parts, fmt.Sprintf("- %s", violation))
		}
	}

	parts = append(parts, fmt.Sprintf("*Checked at: %s*", results.CheckedAt.Format("2006-01-02 15:04:05")))

	return strings.Join(parts, "\n")
}

func (r *PRTemplateRenderer) getRiskEmoji(risk string) string {
	switch strings.ToLower(risk) {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}
