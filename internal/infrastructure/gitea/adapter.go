package giteawebhook

import (
	"context"
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/aosanya/CodeValdCortex/internal/infrastructure/work"
)

// APIClientAdapter wraps the Gitea client and implements work.APIClient
// This adapter converts between Gitea-specific types and provider-agnostic work types
type APIClientAdapter struct {
	client  Client
	baseURL string
}

// NewAPIClientAdapter creates a new adapter that implements work.APIClient
func NewAPIClientAdapter(client Client, baseURL string) work.APIClient {
	return &APIClientAdapter{
		client:  client,
		baseURL: baseURL,
	}
}

// ============================================================================
// Issue Operations
// ============================================================================

func (a *APIClientAdapter) CreateIssue(ctx context.Context, owner, repo string, opts work.CreateIssueOptions) (*work.WorkIssue, error) {
	giteaOpts := CreateIssueOptions{
		Title:     opts.Title,
		Body:      opts.Body,
		Assignees: opts.Assignees,
		Labels:    convertLabelsToIDs(opts.Labels),
		Closed:    opts.Closed,
	}

	if opts.Milestone != "" {
		milestoneID, err := strconv.ParseInt(opts.Milestone, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid milestone ID: %w", err)
		}
		giteaOpts.Milestone = milestoneID
	}

	issue, err := a.client.CreateIssue(ctx, owner, repo, giteaOpts)
	if err != nil {
		return nil, err
	}

	return a.convertIssueToWork(issue, owner, repo), nil
}

func (a *APIClientAdapter) UpdateIssue(ctx context.Context, owner, repo string, issueID string, opts work.UpdateIssueOptions) (*work.WorkIssue, error) {
	index, err := strconv.ParseInt(issueID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid issue ID: %w", err)
	}

	giteaOpts := UpdateIssueOptions{
		Title:     opts.Title,
		Body:      opts.Body,
		Assignees: opts.Assignees,
		State:     opts.State,
	}

	if opts.Milestone != nil {
		milestoneID, err := strconv.ParseInt(*opts.Milestone, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid milestone ID: %w", err)
		}
		giteaOpts.Milestone = &milestoneID
	}

	issue, err := a.client.UpdateIssue(ctx, owner, repo, index, giteaOpts)
	if err != nil {
		return nil, err
	}

	return a.convertIssueToWork(issue, owner, repo), nil
}

func (a *APIClientAdapter) GetIssue(ctx context.Context, owner, repo string, issueID string) (*work.WorkIssue, error) {
	index, err := strconv.ParseInt(issueID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid issue ID: %w", err)
	}

	issue, err := a.client.GetIssue(ctx, owner, repo, index)
	if err != nil {
		return nil, err
	}

	return a.convertIssueToWork(issue, owner, repo), nil
}

func (a *APIClientAdapter) ListIssues(ctx context.Context, owner, repo string, opts work.ListIssueOptions) ([]*work.WorkIssue, error) {
	giteaOpts := ListIssueOptions{
		State:     opts.State,
		Labels:    opts.Labels,
		Milestone: opts.Milestone,
		Since:     opts.Since,
		Page:      opts.Page,
		Limit:     opts.Limit,
	}

	issues, err := a.client.ListIssues(ctx, owner, repo, giteaOpts)
	if err != nil {
		return nil, err
	}

	workIssues := make([]*work.WorkIssue, len(issues))
	for i, issue := range issues {
		workIssues[i] = a.convertIssueToWork(issue, owner, repo)
	}

	return workIssues, nil
}

func (a *APIClientAdapter) CloseIssue(ctx context.Context, owner, repo string, issueID string) error {
	index, err := strconv.ParseInt(issueID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid issue ID: %w", err)
	}

	return a.client.CloseIssue(ctx, owner, repo, index)
}

func (a *APIClientAdapter) ReopenIssue(ctx context.Context, owner, repo string, issueID string) error {
	index, err := strconv.ParseInt(issueID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid issue ID: %w", err)
	}

	return a.client.ReopenIssue(ctx, owner, repo, index)
}

// ============================================================================
// Pull Request Operations
// ============================================================================

func (a *APIClientAdapter) CreatePullRequest(ctx context.Context, owner, repo string, opts work.CreatePullRequestOptions) (*work.WorkPullRequest, error) {
	giteaOpts := CreatePullRequestOptions{
		Title: opts.Title,
		Body:  opts.Body,
		Base:  opts.Base,
		Head:  opts.Head,
	}
	// Note: Gitea CreatePR doesn't support assignees in creation
	// TODO: Add assignees in a separate call after PR creation

	pr, err := a.client.CreatePullRequest(ctx, owner, repo, giteaOpts)
	if err != nil {
		return nil, err
	}

	return a.convertPRToWork(pr, owner, repo), nil
}

func (a *APIClientAdapter) UpdatePullRequest(ctx context.Context, owner, repo string, prID string, opts work.UpdatePullRequestOptions) (*work.WorkPullRequest, error) {
	index, err := strconv.ParseInt(prID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid PR ID: %w", err)
	}

	giteaOpts := UpdatePullRequestOptions{
		Title: opts.Title,
		Body:  opts.Body,
	}
	// Note: Gitea UpdatePR doesn't support assignees in update
	// TODO: Add assignees in a separate call

	pr, err := a.client.UpdatePullRequest(ctx, owner, repo, index, giteaOpts)
	if err != nil {
		return nil, err
	}

	return a.convertPRToWork(pr, owner, repo), nil
}

func (a *APIClientAdapter) GetPullRequest(ctx context.Context, owner, repo string, prID string) (*work.WorkPullRequest, error) {
	index, err := strconv.ParseInt(prID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid PR ID: %w", err)
	}

	pr, err := a.client.GetPullRequest(ctx, owner, repo, index)
	if err != nil {
		return nil, err
	}

	return a.convertPRToWork(pr, owner, repo), nil
}

func (a *APIClientAdapter) ListPullRequests(ctx context.Context, owner, repo string, opts work.ListPullRequestOptions) ([]*work.WorkPullRequest, error) {
	giteaOpts := ListPullRequestOptions{
		State: opts.State,
		Page:  opts.Page,
		Limit: opts.Limit,
	}

	prs, err := a.client.ListPullRequests(ctx, owner, repo, giteaOpts)
	if err != nil {
		return nil, err
	}

	workPRs := make([]*work.WorkPullRequest, len(prs))
	for i, pr := range prs {
		workPRs[i] = a.convertPRToWork(pr, owner, repo)
	}

	return workPRs, nil
}

func (a *APIClientAdapter) MergePullRequest(ctx context.Context, owner, repo string, prID string, opts work.MergePullRequestOptions) error {
	index, err := strconv.ParseInt(prID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid PR ID: %w", err)
	}

	giteaOpts := MergePullRequestOptions{
		Style:   opts.Style,
		Message: opts.Message,
	}

	return a.client.MergePullRequest(ctx, owner, repo, index, giteaOpts)
}

// ============================================================================
// Milestone Operations
// ============================================================================

func (a *APIClientAdapter) CreateMilestone(ctx context.Context, owner, repo string, opts work.CreateMilestoneOptions) (*work.WorkMilestone, error) {
	giteaOpts := CreateMilestoneOptions{
		Title:       opts.Title,
		Description: opts.Description,
		DueDate:     opts.DueDate,
		State:       opts.State,
	}

	milestone, err := a.client.CreateMilestone(ctx, owner, repo, giteaOpts)
	if err != nil {
		return nil, err
	}

	return a.convertMilestoneToWork(milestone, owner, repo), nil
}

func (a *APIClientAdapter) UpdateMilestone(ctx context.Context, owner, repo string, milestoneID string, opts work.UpdateMilestoneOptions) (*work.WorkMilestone, error) {
	id, err := strconv.ParseInt(milestoneID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid milestone ID: %w", err)
	}

	giteaOpts := UpdateMilestoneOptions{
		Title:       opts.Title,
		Description: opts.Description,
		DueDate:     opts.DueDate,
		State:       opts.State,
	}

	milestone, err := a.client.UpdateMilestone(ctx, owner, repo, id, giteaOpts)
	if err != nil {
		return nil, err
	}

	return a.convertMilestoneToWork(milestone, owner, repo), nil
}

func (a *APIClientAdapter) GetMilestone(ctx context.Context, owner, repo string, milestoneID string) (*work.WorkMilestone, error) {
	id, err := strconv.ParseInt(milestoneID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid milestone ID: %w", err)
	}

	milestone, err := a.client.GetMilestone(ctx, owner, repo, id)
	if err != nil {
		return nil, err
	}

	return a.convertMilestoneToWork(milestone, owner, repo), nil
}

func (a *APIClientAdapter) ListMilestones(ctx context.Context, owner, repo string, opts work.ListMilestoneOptions) ([]*work.WorkMilestone, error) {
	giteaOpts := ListMilestoneOptions{
		State: opts.State,
		Page:  opts.Page,
		Limit: opts.Limit,
	}

	milestones, err := a.client.ListMilestones(ctx, owner, repo, giteaOpts)
	if err != nil {
		return nil, err
	}

	workMilestones := make([]*work.WorkMilestone, len(milestones))
	for i, milestone := range milestones {
		workMilestones[i] = a.convertMilestoneToWork(milestone, owner, repo)
	}

	return workMilestones, nil
}

// ============================================================================
// Comment Operations
// ============================================================================

func (a *APIClientAdapter) PostComment(ctx context.Context, owner, repo string, issueID string, body string) (*work.WorkComment, error) {
	index, err := strconv.ParseInt(issueID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid issue ID: %w", err)
	}

	comment, err := a.client.PostComment(ctx, owner, repo, index, body)
	if err != nil {
		return nil, err
	}

	return a.convertCommentToWork(comment, issueID), nil
}

func (a *APIClientAdapter) ListComments(ctx context.Context, owner, repo string, issueID string) ([]*work.WorkComment, error) {
	index, err := strconv.ParseInt(issueID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid issue ID: %w", err)
	}

	comments, err := a.client.ListComments(ctx, owner, repo, index)
	if err != nil {
		return nil, err
	}

	workComments := make([]*work.WorkComment, len(comments))
	for i, comment := range comments {
		workComments[i] = a.convertCommentToWork(comment, issueID)
	}

	return workComments, nil
}

func (a *APIClientAdapter) UpdateComment(ctx context.Context, owner, repo string, commentID string, body string) (*work.WorkComment, error) {
	id, err := strconv.ParseInt(commentID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid comment ID: %w", err)
	}

	comment, err := a.client.UpdateComment(ctx, owner, repo, id, body)
	if err != nil {
		return nil, err
	}

	return a.convertCommentToWork(comment, ""), nil
}

func (a *APIClientAdapter) DeleteComment(ctx context.Context, owner, repo string, commentID string) error {
	id, err := strconv.ParseInt(commentID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid comment ID: %w", err)
	}

	return a.client.DeleteComment(ctx, owner, repo, id)
}

// ============================================================================
// Label Operations
// ============================================================================

func (a *APIClientAdapter) AddLabel(ctx context.Context, owner, repo string, issueID string, labels []string) error {
	index, err := strconv.ParseInt(issueID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid issue ID: %w", err)
	}

	return a.client.AddLabel(ctx, owner, repo, index, labels)
}

func (a *APIClientAdapter) RemoveLabel(ctx context.Context, owner, repo string, issueID string, labelID string) error {
	index, err := strconv.ParseInt(issueID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid issue ID: %w", err)
	}

	id, err := strconv.ParseInt(labelID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid label ID: %w", err)
	}

	return a.client.RemoveLabel(ctx, owner, repo, index, id)
}

func (a *APIClientAdapter) ListLabels(ctx context.Context, owner, repo string, issueID string) ([]*work.WorkLabel, error) {
	index, err := strconv.ParseInt(issueID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid issue ID: %w", err)
	}

	labels, err := a.client.ListLabels(ctx, owner, repo, index)
	if err != nil {
		return nil, err
	}

	workLabels := make([]*work.WorkLabel, len(labels))
	for i, label := range labels {
		workLabels[i] = a.convertLabelToWork(label)
	}

	return workLabels, nil
}

// ============================================================================
// Repository Operations
// ============================================================================

func (a *APIClientAdapter) GetRepository(ctx context.Context, owner, repo string) (*work.WorkRepository, error) {
	repository, err := a.client.GetRepository(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return a.convertRepositoryToWork(repository), nil
}

func (a *APIClientAdapter) ListRepositories(ctx context.Context, opts work.ListRepoOptions) ([]*work.WorkRepository, error) {
	giteaOpts := ListRepoOptions{
		Page:  opts.Page,
		Limit: opts.Limit,
	}

	repos, err := a.client.ListRepositories(ctx, giteaOpts)
	if err != nil {
		return nil, err
	}

	workRepos := make([]*work.WorkRepository, len(repos))
	for i, repo := range repos {
		workRepos[i] = a.convertRepositoryToWork(repo)
	}

	return workRepos, nil
}

// ============================================================================
// Type Conversion Helpers
// ============================================================================

func (a *APIClientAdapter) convertIssueToWork(issue *gitea.Issue, owner, repo string) *work.WorkIssue {
	if issue == nil {
		return nil
	}

	workIssue := &work.WorkIssue{
		Provider:       "gitea",
		IssueID:        strconv.FormatInt(issue.Index, 10),
		IssueNumber:    issue.Index,
		Title:          issue.Title,
		Body:           issue.Body,
		State:          string(issue.State),
		RepoURL:        fmt.Sprintf("%s/%s/%s", a.baseURL, owner, repo),
		AuthorUsername: issue.Poster.UserName,
		CreatedAt:      issue.Created,
		UpdatedAt:      issue.Updated,
		ClosedAt:       issue.Closed,
	}

	if issue.Poster.Email != "" {
		workIssue.AuthorEmail = issue.Poster.Email
	}

	if issue.Milestone != nil {
		workIssue.Milestone = issue.Milestone.Title
		workIssue.MilestoneID = strconv.FormatInt(issue.Milestone.ID, 10)
	}

	// Convert labels
	if len(issue.Labels) > 0 {
		workIssue.Labels = make([]string, len(issue.Labels))
		for i, label := range issue.Labels {
			workIssue.Labels[i] = label.Name
		}
	}

	// Convert assignees
	if len(issue.Assignees) > 0 {
		workIssue.Assignees = make([]string, len(issue.Assignees))
		for i, assignee := range issue.Assignees {
			workIssue.Assignees[i] = assignee.UserName
		}
	}

	return workIssue
}

func (a *APIClientAdapter) convertPRToWork(pr *gitea.PullRequest, owner, repo string) *work.WorkPullRequest {
	if pr == nil {
		return nil
	}

	workPR := &work.WorkPullRequest{
		Provider:       "gitea",
		PRID:           strconv.FormatInt(pr.Index, 10),
		PRNumber:       pr.Index,
		Title:          pr.Title,
		Body:           pr.Body,
		State:          string(pr.State),
		HeadBranch:     pr.Head.Name,
		BaseBranch:     pr.Base.Name,
		RepoURL:        fmt.Sprintf("%s/%s/%s", a.baseURL, owner, repo),
		Merged:         pr.Merged != nil,
		MergedAt:       pr.Merged,
		AuthorUsername: pr.Poster.UserName,
		ClosedAt:       pr.Closed,
	}

	// Handle pointer time fields
	if pr.Created != nil {
		workPR.CreatedAt = *pr.Created
	}
	if pr.Updated != nil {
		workPR.UpdatedAt = *pr.Updated
	}

	if pr.Poster.Email != "" {
		workPR.AuthorEmail = pr.Poster.Email
	}

	if pr.Mergeable {
		mergeable := true
		workPR.Mergeable = &mergeable
	}

	// Convert labels
	if len(pr.Labels) > 0 {
		workPR.Labels = make([]string, len(pr.Labels))
		for i, label := range pr.Labels {
			workPR.Labels[i] = label.Name
		}
	}

	// Note: Gitea SDK doesn't expose RequestedReviewers in current version
	// TODO: Add reviewer support when SDK is updated

	return workPR
}

func (a *APIClientAdapter) convertMilestoneToWork(milestone *gitea.Milestone, owner, repo string) *work.WorkMilestone {
	if milestone == nil {
		return nil
	}

	workMilestone := &work.WorkMilestone{
		Provider:     "gitea",
		MilestoneID:  strconv.FormatInt(milestone.ID, 10),
		Title:        milestone.Title,
		Description:  milestone.Description,
		State:        string(milestone.State),
		RepoURL:      fmt.Sprintf("%s/%s/%s", a.baseURL, owner, repo),
		DueDate:      milestone.Deadline,
		OpenIssues:   int(milestone.OpenIssues),
		ClosedIssues: int(milestone.ClosedIssues),
		CreatedAt:    milestone.Created,
		ClosedAt:     milestone.Closed,
	}

	// Handle pointer time field
	if milestone.Updated != nil {
		workMilestone.UpdatedAt = *milestone.Updated
	}

	return workMilestone
}

func (a *APIClientAdapter) convertCommentToWork(comment *gitea.Comment, parentID string) *work.WorkComment {
	if comment == nil {
		return nil
	}

	workComment := &work.WorkComment{
		Provider:       "gitea",
		CommentID:      strconv.FormatInt(comment.ID, 10),
		ParentID:       parentID,
		ParentType:     work.EventTypeIssue, // Gitea uses same endpoint for issue/PR comments
		Body:           comment.Body,
		AuthorUsername: comment.Poster.UserName,
		CreatedAt:      comment.Created,
		UpdatedAt:      comment.Updated,
	}

	if comment.Poster.Email != "" {
		workComment.AuthorEmail = comment.Poster.Email
	}

	return workComment
}

func (a *APIClientAdapter) convertLabelToWork(label *gitea.Label) *work.WorkLabel {
	if label == nil {
		return nil
	}

	return &work.WorkLabel{
		Provider:    "gitea",
		LabelID:     strconv.FormatInt(label.ID, 10),
		Name:        label.Name,
		Description: label.Description,
		Color:       label.Color,
	}
}

func (a *APIClientAdapter) convertRepositoryToWork(repo *gitea.Repository) *work.WorkRepository {
	if repo == nil {
		return nil
	}

	workRepo := &work.WorkRepository{
		Provider:      "gitea",
		RepoID:        strconv.FormatInt(repo.ID, 10),
		Name:          repo.Name,
		FullName:      repo.FullName,
		Description:   repo.Description,
		Owner:         repo.Owner.UserName,
		URL:           repo.HTMLURL,
		DefaultBranch: repo.DefaultBranch,
		IsPrivate:     repo.Private,
		CreatedAt:     repo.Created,
		UpdatedAt:     repo.Updated,
	}

	return workRepo
}

// convertLabelsToIDs converts label names to IDs (placeholder - needs actual label lookup)
// In a real implementation, this would query Gitea API to get label IDs by name
func convertLabelsToIDs(labels []string) []int64 {
	// TODO: Implement actual label name -> ID conversion
	// For now, return empty slice (labels will be added by name in real implementation)
	return []int64{}
}
