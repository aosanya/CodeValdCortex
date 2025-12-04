package giteawebhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/infrastructure/work"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// TODO THIS File may be redundant
// Handler implements the Gitea webhook HTTP endpoint handlers
type Handler struct {
	validator  *Validator
	repository *Repository
	logger     *log.Entry
}

// NewHandler creates a new Gitea webhook handler
func NewHandler(validator *Validator, repository *Repository) *Handler {
	return &Handler{
		validator:  validator,
		repository: repository,
		logger: log.WithFields(log.Fields{
			"component": "gitea-webhook-handler",
			"prefix":    "MVP-WI-001",
		}),
	}
}

// HandleIssueWebhook processes incoming issue webhook events
// Endpoint: POST /api/v1/work/issues
func (h *Handler) HandleIssueWebhook(c *gin.Context) {
	startTime := time.Now()
	h.logger.Info("Received issue webhook request")

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Validate signature
	signature := c.GetHeader(h.validator.GetSignatureHeader())
	if err := h.validator.ValidateSignature(body, signature); err != nil {
		h.logger.WithError(err).Warn("Signature validation failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse payload
	var payload GiteaIssuePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.WithError(err).Error("Failed to parse JSON payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Log webhook event
	h.logger.WithFields(log.Fields{
		"action":       payload.Action,
		"issue_number": payload.Number,
		"repo":         payload.Repository.FullName,
	}).Info("Processing issue webhook")

	// Transform to normalized work issue
	workIssue := payload.ToWorkIssue()
	if workIssue == nil {
		h.logger.Error("Failed to transform payload to work issue")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue payload"})
		return
	}

	// Process asynchronously (non-blocking webhook response)
	go h.processIssueAsync(c.Request.Context(), workIssue, payload.Action)

	// Return 200 OK immediately
	duration := time.Since(startTime)
	h.logger.WithFields(log.Fields{
		"duration_ms":  duration.Milliseconds(),
		"issue_number": payload.Number,
	}).Info("Webhook accepted successfully")

	c.JSON(http.StatusOK, gin.H{
		"status":  "accepted",
		"message": "Webhook processed successfully",
	})
}

// HandlePullRequestWebhook processes incoming pull request webhook events
// Endpoint: POST /api/v1/work/pull-requests
func (h *Handler) HandlePullRequestWebhook(c *gin.Context) {
	startTime := time.Now()
	h.logger.Info("Received pull request webhook request")

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Validate signature
	signature := c.GetHeader(h.validator.GetSignatureHeader())
	if err := h.validator.ValidateSignature(body, signature); err != nil {
		h.logger.WithError(err).Warn("Signature validation failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse payload
	var payload GiteaPullRequestPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.WithError(err).Error("Failed to parse JSON payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Log webhook event
	h.logger.WithFields(log.Fields{
		"action":    payload.Action,
		"pr_number": payload.Number,
		"repo":      payload.Repository.FullName,
	}).Info("Processing PR webhook")

	// Transform to normalized work PR
	workPR := payload.ToWorkPullRequest()
	if workPR == nil {
		h.logger.Error("Failed to transform payload to work PR")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PR payload"})
		return
	}

	// Process asynchronously
	go h.processPRAsync(c.Request.Context(), workPR, payload.Action)

	// Return 200 OK immediately
	duration := time.Since(startTime)
	h.logger.WithFields(log.Fields{
		"duration_ms": duration.Milliseconds(),
		"pr_number":   payload.Number,
	}).Info("Webhook accepted successfully")

	c.JSON(http.StatusOK, gin.H{
		"status":  "accepted",
		"message": "Webhook processed successfully",
	})
}

// processIssueAsync handles issue persistence asynchronously
func (h *Handler) processIssueAsync(ctx context.Context, issue *work.WorkIssue, action string) {
	// Use background context with timeout to avoid cancellation from HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	h.logger.WithFields(log.Fields{
		"issue_id": issue.IssueID,
		"action":   action,
	}).Info("Starting async issue processing")

	// Save to ArangoDB
	if err := h.repository.SaveIssue(ctx, issue); err != nil {
		h.logger.WithError(err).WithFields(log.Fields{
			"issue_id": issue.IssueID,
		}).Error("Failed to save issue to ArangoDB")
		return
	}

	h.logger.WithFields(log.Fields{
		"issue_id":     issue.IssueID,
		"issue_number": issue.IssueNumber,
		"action":       action,
	}).Info("Issue saved to ArangoDB successfully")

	// TODO: Trigger orchestrator notification via ArangoDB change stream
	// The orchestrator (MVP-032) will monitor work_issues collection
	// and react to new/updated issues automatically
}

// processPRAsync handles PR persistence asynchronously
func (h *Handler) processPRAsync(ctx context.Context, pr *work.WorkPullRequest, action string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	h.logger.WithFields(log.Fields{
		"pr_id":  pr.PRID,
		"action": action,
	}).Info("Starting async PR processing")

	// Save to ArangoDB
	if err := h.repository.SavePullRequest(ctx, pr); err != nil {
		h.logger.WithError(err).WithFields(log.Fields{
			"pr_id": pr.PRID,
		}).Error("Failed to save PR to ArangoDB")
		return
	}

	h.logger.WithFields(log.Fields{
		"pr_id":     pr.PRID,
		"pr_number": pr.PRNumber,
		"action":    action,
	}).Info("PR saved to ArangoDB successfully")
}

// GetProviderName returns the provider name for this handler
func (h *Handler) GetProviderName() string {
	return "gitea"
}

// ValidateWebhookSignature validates the webhook signature
func (h *Handler) ValidateWebhookSignature(payload []byte, signature string, secret string) error {
	validator := NewValidator(secret)
	return validator.ValidateSignature(payload, signature)
}

// HandleWebhook is the main webhook processing method implementing work.WorkTrackingProvider
func (h *Handler) HandleWebhook(ctx context.Context, r *http.Request) (*work.WebhookResult, error) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Validate signature
	signature := r.Header.Get(h.validator.GetSignatureHeader())
	if err := h.validator.ValidateSignature(body, signature); err != nil {
		return nil, fmt.Errorf("signature validation failed: %w", err)
	}

	// Determine event type from X-Gitea-Event header
	eventType := r.Header.Get("X-Gitea-Event")

	result := &work.WebhookResult{
		Provider:    "gitea",
		ProcessedAt: time.Now(),
	}

	switch eventType {
	case "issues":
		var payload GiteaIssuePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse issue payload: %w", err)
		}

		result.EventType = work.EventTypeIssue
		result.Action = payload.Action
		result.Issue = payload.ToWorkIssue()

	case "pull_request":
		var payload GiteaPullRequestPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse PR payload: %w", err)
		}

		result.EventType = work.EventTypePullRequest
		result.Action = payload.Action
		result.PullRequest = payload.ToWorkPullRequest()

	default:
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}

	return result, nil
}
