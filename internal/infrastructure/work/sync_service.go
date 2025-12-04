package work

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// DefaultSyncService implements SyncService for agent-to-issue synchronization
type DefaultSyncService struct {
	workClient   APIClient
	linkRepo     AgentIssueLinkRepository
	auditRepo    SyncAuditRepository
	templateRepo TemplateRenderer
	workflowSvc  WorkflowService
}

// NewSyncService creates a new sync service
func NewSyncService(
	workClient APIClient,
	linkRepo AgentIssueLinkRepository,
	auditRepo SyncAuditRepository,
	templateRepo TemplateRenderer,
	workflowSvc WorkflowService,
) SyncService {
	return &DefaultSyncService{
		workClient:   workClient,
		linkRepo:     linkRepo,
		auditRepo:    auditRepo,
		templateRepo: templateRepo,
		workflowSvc:  workflowSvc,
	}
}

// HandleAgentEvent processes an agent event and syncs it to the work tracking system
func (s *DefaultSyncService) HandleAgentEvent(ctx context.Context, event *SyncEventPayload) error {
	// Validate event
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	if event.AgentID == "" {
		return fmt.Errorf("event missing agent_id")
	}


	// Get agent-issue link
	link, err := s.linkRepo.GetByAgentID(ctx, event.AgentID)
	if err != nil {
		log.WithError(err).WithField("agent_id", event.AgentID).Warn("No agent-issue link found, skipping sync")
		return nil // Not an error - agent may not be linked to an issue
	}

	// Route to appropriate handler based on event type
	var syncErr error
	switch event.EventType {
	case "lifecycle":
		syncErr = s.handleLifecycleEvent(ctx, link, event)
	case "run":
		syncErr = s.handleRunEvent(ctx, link, event)
	case "progress":
		syncErr = s.handleProgressEvent(ctx, link, event)
	default:
		syncErr = fmt.Errorf("unknown event type: %s", event.EventType)
	}

	// Record audit trail
	audit := &SyncAuditRecord{
		AgentID:        event.AgentID,
		IssueID:        link.IssueID,
		EventType:      event.EventType,
		EventName:      event.EventName,
		EventTimestamp: event.EventTimestamp,
		SyncTimestamp:  time.Now(),
		Success:        syncErr == nil,
		ProviderType:   link.ProviderType,
		RepositoryURL:  link.RepositoryURL,
	}

	if syncErr != nil {
		audit.ErrorMessage = syncErr.Error()
		log.WithError(syncErr).Error("Sync operation failed")
	}

	if err := s.auditRepo.Create(ctx, audit); err != nil {
		log.WithError(err).Error("Failed to create audit record")
	}

	return syncErr
}

// handleLifecycleEvent processes agent lifecycle events
func (s *DefaultSyncService) handleLifecycleEvent(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	switch event.EventName {
	case "agent.lifecycle.registered":
		return s.postAgentStarted(ctx, link, event)
	case "agent.lifecycle.healthy":
		return s.postAgentHealthy(ctx, link, event)
	case "agent.lifecycle.degraded":
		return s.postAgentDegraded(ctx, link, event)
	case "agent.lifecycle.quarantined":
		return s.postAgentQuarantined(ctx, link, event)
	case "agent.lifecycle.stopped":
		return s.postAgentStopped(ctx, link, event)
	default:
		return nil
	}
}

// handleRunEvent processes run execution events
func (s *DefaultSyncService) handleRunEvent(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	switch event.EventName {
	case "run.execution.running":
		return s.postTaskRunning(ctx, link, event)
	case "run.execution.waiting_io":
		return s.postWaitingIO(ctx, link, event)
	case "run.execution.waiting_hitl":
		return s.postWaitingHITL(ctx, link, event)
	case "run.execution.succeeded":
		return s.postTaskCompleted(ctx, link, event)
	case "run.execution.failed":
		return s.postTaskFailed(ctx, link, event)
	default:
		return nil
	}
}

// handleProgressEvent processes custom progress events
func (s *DefaultSyncService) handleProgressEvent(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	switch event.EventName {
	case "agent.progress.update":
		return s.postProgressUpdate(ctx, link, event)
	case "agent.milestone.complete":
		return s.handleMilestoneComplete(ctx, link, event)
	default:
		return nil
	}
}

// postAgentStarted posts agent started comment
func (s *DefaultSyncService) postAgentStarted(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:       event.AgentID,
		AgentType:     event.AgentType,
		Timestamp:     event.EventTimestamp,
		FormattedTime: event.EventTimestamp.Format(time.RFC1123),
		ColumnName:    getStringFromMap(event.EventData, "column_name"),
	}

	return s.PostComment(ctx, link, "agent_started", data)
}

// postAgentHealthy posts agent healthy status
func (s *DefaultSyncService) postAgentHealthy(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:       event.AgentID,
		Timestamp:     event.EventTimestamp,
		FormattedTime: event.EventTimestamp.Format(time.RFC1123),
	}

	// Add label
	if err := s.UpdateLabels(ctx, link, []string{"agent-active"}, nil); err != nil {
		log.WithError(err).Warn("Failed to add agent-active label")
	}

	return s.PostComment(ctx, link, "agent_healthy", data)
}

// postAgentDegraded posts agent degraded status
func (s *DefaultSyncService) postAgentDegraded(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:       event.AgentID,
		Timestamp:     event.EventTimestamp,
		FormattedTime: event.EventTimestamp.Format(time.RFC1123),
		StatusMessage: getStringFromMap(event.EventData, "reason"),
	}

	return s.PostComment(ctx, link, "agent_degraded", data)
}

// postAgentQuarantined posts agent quarantined alert
func (s *DefaultSyncService) postAgentQuarantined(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:       event.AgentID,
		Timestamp:     event.EventTimestamp,
		FormattedTime: event.EventTimestamp.Format(time.RFC1123),
		ErrorMessage:  event.ErrorMessage,
	}

	// Add label
	if err := s.UpdateLabels(ctx, link, []string{"agent-quarantined"}, []string{"agent-active"}); err != nil {
		log.WithError(err).Warn("Failed to update quarantine labels")
	}

	return s.PostComment(ctx, link, "agent_quarantined", data)
}

// postAgentStopped posts agent stopped notification
func (s *DefaultSyncService) postAgentStopped(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:       event.AgentID,
		Timestamp:     event.EventTimestamp,
		FormattedTime: event.EventTimestamp.Format(time.RFC1123),
		StatusMessage: getStringFromMap(event.EventData, "reason"),
	}

	// Remove active label
	if err := s.UpdateLabels(ctx, link, nil, []string{"agent-active"}); err != nil {
		log.WithError(err).Warn("Failed to remove agent-active label")
	}

	// Mark link as completed if stopped successfully
	if err := s.linkRepo.MarkCompleted(ctx, link.AgentID); err != nil {
		log.WithError(err).Warn("Failed to mark link as completed")
	}

	return s.PostComment(ctx, link, "agent_stopped", data)
}

// postTaskRunning posts task execution started
func (s *DefaultSyncService) postTaskRunning(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:             event.AgentID,
		Timestamp:           event.EventTimestamp,
		FormattedTime:       event.EventTimestamp.Format(time.RFC1123),
		DetailedDescription: event.TaskDescription,
	}

	return s.PostComment(ctx, link, "task_running", data)
}

// postWaitingIO posts waiting for I/O status
func (s *DefaultSyncService) postWaitingIO(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:       event.AgentID,
		Timestamp:     event.EventTimestamp,
		FormattedTime: event.EventTimestamp.Format(time.RFC1123),
		StatusMessage: getStringFromMap(event.EventData, "io_description"),
	}

	return s.PostComment(ctx, link, "waiting_io", data)
}

// postWaitingHITL posts human-in-the-loop request
func (s *DefaultSyncService) postWaitingHITL(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:       event.AgentID,
		Timestamp:     event.EventTimestamp,
		FormattedTime: event.EventTimestamp.Format(time.RFC1123),
		StatusMessage: getStringFromMap(event.EventData, "approval_context"),
	}

	return s.PostComment(ctx, link, "waiting_hitl", data)
}

// postTaskCompleted posts task completion
func (s *DefaultSyncService) postTaskCompleted(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:       event.AgentID,
		Timestamp:     event.EventTimestamp,
		FormattedTime: event.EventTimestamp.Format(time.RFC1123),
		TaskName:      event.TaskName,
		TaskDuration:  event.TaskDuration,
		TaskSummary:   event.TaskSummary,
		Deliverables:  getStringFromMap(event.EventData, "deliverables"),
	}

	return s.PostComment(ctx, link, "task_completed", data)
}

// postTaskFailed posts task failure
func (s *DefaultSyncService) postTaskFailed(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:          event.AgentID,
		Timestamp:        event.EventTimestamp,
		FormattedTime:    event.EventTimestamp.Format(time.RFC1123),
		TaskName:         event.TaskName,
		ErrorMessage:     event.ErrorMessage,
		RemediationGuide: getStringFromMap(event.EventData, "remediation"),
	}

	return s.PostComment(ctx, link, "task_failed", data)
}

// postProgressUpdate posts progress update
func (s *DefaultSyncService) postProgressUpdate(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	data := &CommentTemplateData{
		AgentID:             event.AgentID,
		Timestamp:           event.EventTimestamp,
		FormattedTime:       event.EventTimestamp.Format(time.RFC1123),
		StatusMessage:       event.ProgressMessage,
		ProgressPercentage:  event.ProgressPercentage,
		DetailedDescription: getStringFromMap(event.EventData, "description"),
	}

	return s.PostComment(ctx, link, "progress_update", data)
}

// handleMilestoneComplete handles milestone completion and progression
func (s *DefaultSyncService) handleMilestoneComplete(ctx context.Context, link *AgentIssueLink, event *SyncEventPayload) error {
	// First, progress the milestone
	if err := s.ProgressMilestone(ctx, link); err != nil {
		return fmt.Errorf("failed to progress milestone: %w", err)
	}

	// Then post completion comment
	data := &CommentTemplateData{
		AgentID:       event.AgentID,
		Timestamp:     event.EventTimestamp,
		FormattedTime: event.EventTimestamp.Format(time.RFC1123),
		CurrentColumn: link.CurrentMilestone,
		WorkSummary:   getStringFromMap(event.EventData, "summary"),
	}

	return s.PostComment(ctx, link, "milestone_completed", data)
}

// PostComment creates an issue comment from a template
func (s *DefaultSyncService) PostComment(ctx context.Context, link *AgentIssueLink, template string, data *CommentTemplateData) error {
	// Render template
	comment, err := s.templateRepo.Render(ctx, template, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Parse repository owner and name from URL
	owner, repo, err := parseRepoURL(link.RepositoryURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %w", err)
	}

	// Post comment using work client (APIClient embeds CommentClient)
	workComment, err := s.workClient.PostComment(ctx, owner, repo, link.IssueID, comment)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	// Update link metadata
	commentID := workComment.CommentID
	if err := s.linkRepo.UpdateLastSync(ctx, link.AgentID, template, commentID); err != nil {
		log.WithError(err).Warn("Failed to update link sync metadata")
	}


	return nil
}

// UpdateLabels adds or removes labels from an issue
func (s *DefaultSyncService) UpdateLabels(ctx context.Context, link *AgentIssueLink, add []string, remove []string) error {
	// Parse repository owner and name from URL
	owner, repo, err := parseRepoURL(link.RepositoryURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %w", err)
	}

	// Add labels (APIClient embeds LabelClient)
	if len(add) > 0 {
		if err := s.workClient.AddLabel(ctx, owner, repo, link.IssueID, add); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"issue_id": link.IssueID,
				"labels":   add,
			}).Warn("Failed to add labels")
		}
	}

	// Remove labels
	for _, label := range remove {
		if err := s.workClient.RemoveLabel(ctx, owner, repo, link.IssueID, label); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"issue_id": link.IssueID,
				"label":    label,
			}).Warn("Failed to remove label")
		}
	}

	return nil
}

// ProgressMilestone moves an issue to the next workflow column
func (s *DefaultSyncService) ProgressMilestone(ctx context.Context, link *AgentIssueLink) error {
	// Get workflow
	workflow, err := s.workflowSvc.GetWorkflow(ctx, link.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Get next column
	nextColumn, err := s.workflowSvc.GetNextColumn(ctx, link.WorkflowID, link.ColumnID)
	if err != nil {
		return fmt.Errorf("failed to get next column: %w", err)
	}

	if nextColumn == nil {
		log.WithField("column_id", link.ColumnID).Info("No next column, workflow complete")
		return nil
	}

	// Get milestone name for next column
	milestoneName, err := s.workflowSvc.GetMilestoneMapping(ctx, link.WorkflowID, nextColumn.ID)
	if err != nil {
		return fmt.Errorf("failed to get milestone mapping: %w", err)
	}

	// Parse repository owner and name from URL
	owner, repo, err := parseRepoURL(link.RepositoryURL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %w", err)
	}

	// Update issue milestone (APIClient embeds IssueClient)
	_, err = s.workClient.UpdateIssue(ctx, owner, repo, link.IssueID, UpdateIssueOptions{
		Milestone: &milestoneName,
	})
	if err != nil {
		return fmt.Errorf("failed to update issue milestone: %w", err)
	}

	// Update link current milestone
	link.CurrentMilestone = milestoneName

	log.WithFields(log.Fields{
		"issue_id":  link.IssueID,
		"workflow":  workflow.Name,
		"milestone": milestoneName,
	}).Info("Progressed issue to next milestone")

	return nil
}

// RecordAudit logs a sync operation (called internally, can also be used externally)
func (s *DefaultSyncService) RecordAudit(ctx context.Context, record *SyncAuditRecord) error {
	return s.auditRepo.Create(ctx, record)
}

// GetStats returns sync service performance metrics
func (s *DefaultSyncService) GetStats(ctx context.Context) (*SyncStats, error) {
	// Get stats from last 24 hours
	since := time.Now().Add(-24 * time.Hour).Unix()
	return s.auditRepo.GetStats(ctx, since)
}

// Helper function to safely extract string from map
func getStringFromMap(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// parseRepoURL extracts owner and repo name from a repository URL
// Supports formats like:
// - https://gitea.example.com/owner/repo
// - https://github.com/owner/repo
// - git@github.com:owner/repo.git
func parseRepoURL(repoURL string) (owner string, repo string, err error) {
	if repoURL == "" {
		return "", "", fmt.Errorf("repository URL is empty")
	}

	// Handle SSH URLs (git@github.com:owner/repo.git)
	if len(repoURL) > 4 && repoURL[:4] == "git@" {
		// Extract the path after the colon
		colonIdx := len(repoURL)
		for i := len(repoURL) - 1; i >= 0; i-- {
			if repoURL[i] == ':' {
				colonIdx = i
				break
			}
		}
		if colonIdx < len(repoURL) {
			path := repoURL[colonIdx+1:]
			// Remove .git suffix if present
			if len(path) > 4 && path[len(path)-4:] == ".git" {
				path = path[:len(path)-4]
			}
			parts := splitPath(path)
			if len(parts) >= 2 {
				return parts[0], parts[1], nil
			}
		}
		return "", "", fmt.Errorf("invalid SSH repository URL format: %s", repoURL)
	}

	// Handle HTTPS URLs
	// Find the last two path segments
	slashCount := 0
	ownerStart := -1
	repoStart := -1

	for i := len(repoURL) - 1; i >= 0; i-- {
		if repoURL[i] == '/' {
			slashCount++
			if slashCount == 1 {
				repoStart = i + 1
			} else if slashCount == 2 {
				ownerStart = i + 1
				break
			}
		}
	}

	if ownerStart == -1 || repoStart == -1 {
		return "", "", fmt.Errorf("invalid repository URL format: %s", repoURL)
	}

	owner = repoURL[ownerStart : repoStart-1]
	repo = repoURL[repoStart:]

	// Remove .git suffix if present
	if len(repo) > 4 && repo[len(repo)-4:] == ".git" {
		repo = repo[:len(repo)-4]
	}

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("could not extract owner/repo from URL: %s", repoURL)
	}

	return owner, repo, nil
}

// splitPath splits a path by '/' character
func splitPath(path string) []string {
	var parts []string
	start := 0

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}

	if start < len(path) {
		parts = append(parts, path[start:])
	}

	return parts
}
