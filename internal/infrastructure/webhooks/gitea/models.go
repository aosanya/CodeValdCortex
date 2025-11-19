package giteawebhook

import (
	gitea "code.gitea.io/sdk/gitea"
	"github.com/aosanya/CodeValdCortex/internal/infrastructure/webhooks/work"
)

// GiteaIssuePayload represents the Gitea-specific webhook payload for issues
type GiteaIssuePayload struct {
	Action     string            `json:"action"`
	Number     int64             `json:"number"`
	Issue      *gitea.Issue      `json:"issue"`
	Repository *gitea.Repository `json:"repository"`
	Sender     *gitea.User       `json:"sender"`
}

// GiteaPullRequestPayload represents the Gitea-specific webhook payload for PRs
type GiteaPullRequestPayload struct {
	Action      string             `json:"action"`
	Number      int64              `json:"number"`
	PullRequest *gitea.PullRequest `json:"pull_request"`
	Repository  *gitea.Repository  `json:"repository"`
	Sender      *gitea.User        `json:"sender"`
}

// GiteaMilestonePayload represents the Gitea-specific webhook payload for milestones
type GiteaMilestonePayload struct {
	Action     string            `json:"action"`
	Milestone  *gitea.Milestone  `json:"milestone"`
	Repository *gitea.Repository `json:"repository"`
	Sender     *gitea.User       `json:"sender"`
}

// ToWorkIssue converts a Gitea issue payload to the normalized WorkIssue format
func (p *GiteaIssuePayload) ToWorkIssue() *work.WorkIssue {
	if p.Issue == nil || p.Repository == nil {
		return nil
	}

	issue := &work.WorkIssue{
		Provider:       "gitea",
		IssueID:        p.Issue.HTMLURL, // Use HTML URL as unique ID
		IssueNumber:    p.Issue.Index,
		Title:          p.Issue.Title,
		Body:           p.Issue.Body,
		State:          string(p.Issue.State),
		RepoURL:        p.Repository.HTMLURL,
		AuthorUsername: p.Issue.Poster.UserName,
		AuthorEmail:    p.Issue.Poster.Email,
		CreatedAt:      p.Issue.Created,
		UpdatedAt:      p.Issue.Updated,
	}

	// Handle milestone
	if p.Issue.Milestone != nil {
		issue.Milestone = p.Issue.Milestone.Title
		issue.MilestoneID = p.Issue.Milestone.Title // Gitea uses title as identifier
	}

	// Extract labels
	if len(p.Issue.Labels) > 0 {
		issue.Labels = make([]string, len(p.Issue.Labels))
		for i, label := range p.Issue.Labels {
			issue.Labels[i] = label.Name
		}
	}

	// Extract assignees
	if len(p.Issue.Assignees) > 0 {
		issue.Assignees = make([]string, len(p.Issue.Assignees))
		for i, assignee := range p.Issue.Assignees {
			issue.Assignees[i] = assignee.UserName
		}
	}

	// Handle closed timestamp
	if p.Issue.Closed != nil {
		issue.ClosedAt = p.Issue.Closed
	}

	// Store Gitea-specific metadata
	issue.ProviderMetadata = map[string]interface{}{
		"html_url":      p.Issue.HTMLURL,
		"webhook_event": p.Action,
		"issue_id":      p.Issue.ID,
	}

	return issue
}

// ToWorkPullRequest converts a Gitea PR payload to the normalized WorkPullRequest format
func (p *GiteaPullRequestPayload) ToWorkPullRequest() *work.WorkPullRequest {
	if p.PullRequest == nil || p.Repository == nil {
		return nil
	}

	pr := &work.WorkPullRequest{
		Provider:       "gitea",
		PRID:           p.PullRequest.HTMLURL,
		PRNumber:       p.PullRequest.Index,
		Title:          p.PullRequest.Title,
		Body:           p.PullRequest.Body,
		State:          string(p.PullRequest.State),
		HeadBranch:     p.PullRequest.Head.Ref,
		BaseBranch:     p.PullRequest.Base.Ref,
		RepoURL:        p.Repository.HTMLURL,
		Merged:         p.PullRequest.HasMerged,
		AuthorUsername: p.PullRequest.Poster.UserName,
		AuthorEmail:    p.PullRequest.Poster.Email,
	}

	// Handle timestamps
	if p.PullRequest.Created != nil {
		pr.CreatedAt = *p.PullRequest.Created
	}
	if p.PullRequest.Updated != nil {
		pr.UpdatedAt = *p.PullRequest.Updated
	}

	// Handle mergeable status
	if p.PullRequest.Mergeable {
		mergeable := true
		pr.Mergeable = &mergeable
	}

	// Handle merged timestamp
	if p.PullRequest.Merged != nil {
		pr.MergedAt = p.PullRequest.Merged
	}

	// Extract labels
	if len(p.PullRequest.Labels) > 0 {
		pr.Labels = make([]string, len(p.PullRequest.Labels))
		for i, label := range p.PullRequest.Labels {
			pr.Labels[i] = label.Name
		}
	}

	// Extract assignees/reviewers
	if len(p.PullRequest.Assignees) > 0 {
		pr.Reviewers = make([]string, len(p.PullRequest.Assignees))
		for i, assignee := range p.PullRequest.Assignees {
			pr.Reviewers[i] = assignee.UserName
		}
	}

	// Handle closed timestamp
	if p.PullRequest.Closed != nil {
		pr.ClosedAt = p.PullRequest.Closed
	}

	// Store Gitea-specific metadata
	pr.ProviderMetadata = map[string]interface{}{
		"html_url":      p.PullRequest.HTMLURL,
		"webhook_event": p.Action,
		"pr_id":         p.PullRequest.ID,
	}

	return pr
}

// ToWorkMilestone converts a Gitea milestone payload to the normalized WorkMilestone format
func (p *GiteaMilestonePayload) ToWorkMilestone() *work.WorkMilestone {
	if p.Milestone == nil || p.Repository == nil {
		return nil
	}

	milestone := &work.WorkMilestone{
		Provider:     "gitea",
		MilestoneID:  p.Milestone.Title, // Gitea uses title as identifier
		Title:        p.Milestone.Title,
		Description:  p.Milestone.Description,
		State:        string(p.Milestone.State),
		RepoURL:      p.Repository.HTMLURL,
		OpenIssues:   p.Milestone.OpenIssues,
		ClosedIssues: p.Milestone.ClosedIssues,
		CreatedAt:    p.Milestone.Created,
	}

	// Handle updated timestamp
	if p.Milestone.Updated != nil {
		milestone.UpdatedAt = *p.Milestone.Updated
	}

	// Handle due date
	if p.Milestone.Deadline != nil {
		milestone.DueDate = p.Milestone.Deadline
	}

	// Handle closed timestamp
	if p.Milestone.Closed != nil {
		milestone.ClosedAt = p.Milestone.Closed
	}

	// Store Gitea-specific metadata
	milestone.ProviderMetadata = map[string]interface{}{
		"webhook_event": p.Action,
		"milestone_id":  p.Milestone.ID,
	}

	return milestone
}
