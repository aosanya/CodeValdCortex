package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/validation"
)

// PublicationRepository defines the interface for publication data access
// This is redefined here to avoid import cycles with arangodb package
type PublicationRepository interface {
	Create(ctx context.Context, pub *models.AgencyPublication) error
	GetByID(ctx context.Context, pubID string) (*models.AgencyPublication, error)
	ListByAgency(ctx context.Context, agencyID string) ([]*models.AgencyPublication, error)
	GetLatest(ctx context.Context, agencyID string) (*models.AgencyPublication, error)
	Update(ctx context.Context, pub *models.AgencyPublication) error
	GetByVersion(ctx context.Context, agencyID string, version string) (*models.AgencyPublication, error)
}

// PublishRequest contains the parameters for publishing an agency
type PublishRequest struct {
	Version      string            `json:"version" binding:"required"`
	Description  string            `json:"description" binding:"required"`
	AutoActivate bool              `json:"auto_activate,omitempty"`
	CreateTag    bool              `json:"create_tag,omitempty"`
	TagName      string            `json:"tag_name,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	PublishedBy  string            `json:"published_by,omitempty"`
}

// ActivationResult contains the results of agency activation
type ActivationResult struct {
	AgentsSpawned        int       `json:"agents_spawned"`
	WorkflowsInitialized int       `json:"workflows_initialized"`
	MonitoringEnabled    bool      `json:"monitoring_enabled"`
	ActivatedAt          time.Time `json:"activated_at"`
}

// PublicationService defines the interface for publication operations
type PublicationService interface {
	// ValidateForPublish checks if agency is ready to be published
	ValidateForPublish(ctx context.Context, agencyID string) (*validation.ValidationResult, error)

	// Publish creates an immutable publication from current agency state
	Publish(ctx context.Context, agencyID string, req *PublishRequest) (*models.AgencyPublication, error)

	// Activate transitions agency from published to active (stub for MVP-PUB-004)
	Activate(ctx context.Context, publicationID string) (*ActivationResult, error)

	// Deactivate transitions agency from active to stopped
	Deactivate(ctx context.Context, agencyID string, graceful bool) error

	// GetPublicationHistory retrieves all publications for an agency
	GetPublicationHistory(ctx context.Context, agencyID string) ([]*models.AgencyPublication, error)

	// RepublishFromTag creates new publication from a tag (future)
	RepublishFromTag(ctx context.Context, tagID string) (*models.AgencyPublication, error)
}

// publicationService is the concrete implementation of PublicationService
type publicationService struct {
	pubRepo      PublicationRepository
	agencyRepo   agency.Repository
	stateMachine *agency.AgencyStateMachine
	validator    *validation.PublisherValidator
	logger       *slog.Logger
}

// NewPublicationService creates a new publication service instance
func NewPublicationService(
	pubRepo PublicationRepository,
	agencyRepo agency.Repository,
	stateMachine *agency.AgencyStateMachine,
	validator *validation.PublisherValidator,
	logger *slog.Logger,
) PublicationService {
	return &publicationService{
		pubRepo:      pubRepo,
		agencyRepo:   agencyRepo,
		stateMachine: stateMachine,
		validator:    validator,
		logger:       logger,
	}
}

// ValidateForPublish checks if agency is ready for publication
func (s *publicationService) ValidateForPublish(ctx context.Context, agencyID string) (*validation.ValidationResult, error) {
	s.logger.Info("validating agency for publication", "agency_id", agencyID)

	// Retrieve agency
	agencyDoc, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Retrieve specification
	spec, err := s.agencyRepo.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get specification: %w", err)
	}

	// Run validation
	result, err := s.validator.ValidateForPublish(agencyDoc, spec)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	s.logger.Info("validation completed",
		"agency_id", agencyID,
		"valid", result.Valid,
		"errors", len(result.Errors),
		"warnings", len(result.Warnings))

	return result, nil
}

// Publish creates an immutable publication from current agency state
func (s *publicationService) Publish(ctx context.Context, agencyID string, req *PublishRequest) (*models.AgencyPublication, error) {
	s.logger.Info("publishing agency",
		"agency_id", agencyID,
		"version", req.Version,
		"published_by", req.PublishedBy)

	// Step 1: Retrieve current agency
	agencyDoc, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Step 2: Validate for publishing
	spec, err := s.agencyRepo.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get specification: %w", err)
	}

	validationResult, err := s.validator.ValidateForPublish(agencyDoc, spec)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if !validationResult.Valid {
		return nil, fmt.Errorf("validation failed: %d errors found", len(validationResult.Errors))
	}

	// Step 3: Check for duplicate version
	existingPub, err := s.pubRepo.GetByVersion(ctx, agencyID, req.Version)
	if err == nil && existingPub != nil {
		return nil, fmt.Errorf("publication with version %s already exists", req.Version)
	}

	// Step 4: Transition to validated if in draft
	if agencyDoc.State == models.AgencyStateDraft {
		if err := s.stateMachine.Transition(agencyDoc, "validate"); err != nil {
			s.logger.Warn("failed to transition to validated", "error", err)
			// Continue anyway - validation checks passed
		}
		// Update agency state in database
		if err := s.agencyRepo.Update(ctx, agencyDoc); err != nil {
			return nil, fmt.Errorf("failed to update agency state: %w", err)
		}
	}

	// Step 5: Generate snapshot
	snapshot := s.generateSnapshot(agencyDoc, spec)

	// Step 6: Generate deployment manifest
	manifest := s.generateDeploymentManifest(spec, agencyDoc)

	// Step 7: Create publication
	publication := &models.AgencyPublication{
		AgencyID:    agencyID,
		Version:     req.Version,
		Description: req.Description,
		Snapshot:    snapshot,
		Manifest:    manifest,
		PublishedAt: time.Now(),
		PublishedBy: req.PublishedBy,
	}

	// Add metadata if provided
	if req.Metadata != nil {
		publication.Metadata = make(map[string]interface{})
		for k, v := range req.Metadata {
			publication.Metadata[k] = v
		}
	}

	// Step 8: Save publication
	if err := s.pubRepo.Create(ctx, publication); err != nil {
		return nil, fmt.Errorf("failed to create publication: %w", err)
	}

	// Step 9: Update agency state to published
	if err := s.stateMachine.Transition(agencyDoc, "publish"); err != nil {
		s.logger.Warn("failed to transition to published", "error", err)
		// Continue - publication already saved
	}

	// Step 10: Update agency metadata
	agencyDoc.PublishedAt = &publication.PublishedAt
	agencyDoc.PublishedBy = req.PublishedBy
	agencyDoc.PublicationID = publication.Key

	if err := s.agencyRepo.Update(ctx, agencyDoc); err != nil {
		s.logger.Error("failed to update agency metadata", "error", err)
		// Non-critical - publication exists
	}

	s.logger.Info("agency published successfully",
		"agency_id", agencyID,
		"publication_id", publication.Key,
		"version", publication.Version)

	// Step 11: Auto-activate if requested
	if req.AutoActivate {
		s.logger.Info("auto-activating publication", "publication_id", publication.Key)
		_, err := s.Activate(ctx, publication.Key)
		if err != nil {
			s.logger.Warn("auto-activation failed", "error", err)
			// Non-critical - publication succeeded
		}
	}

	return publication, nil
}

// Activate transitions agency from published to active (stub for MVP-PUB-004)
func (s *publicationService) Activate(ctx context.Context, publicationID string) (*ActivationResult, error) {
	s.logger.Info("activating publication", "publication_id", publicationID)

	// Retrieve publication
	publication, err := s.pubRepo.GetByID(ctx, publicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get publication: %w", err)
	}

	// Retrieve agency
	agencyDoc, err := s.agencyRepo.GetByID(ctx, publication.AgencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Check state (must be published)
	if agencyDoc.State != models.AgencyStatePublished {
		return nil, fmt.Errorf("agency must be in published state to activate (current: %s)", agencyDoc.State)
	}

	// Transition to active
	if err := s.stateMachine.Transition(agencyDoc, "activate"); err != nil {
		return nil, fmt.Errorf("failed to transition to active: %w", err)
	}

	// Update agency
	now := time.Now()
	agencyDoc.ActivatedAt = &now
	publication.ActivatedAt = &now

	if err := s.agencyRepo.Update(ctx, agencyDoc); err != nil {
		return nil, fmt.Errorf("failed to update agency: %w", err)
	}

	// Update publication
	if err := s.pubRepo.Update(ctx, publication); err != nil {
		s.logger.Warn("failed to update publication", "error", err)
	}

	s.logger.Info("publication activated",
		"publication_id", publicationID,
		"agency_id", publication.AgencyID)

	// Return stub result (actual agent spawning in MVP-PUB-004)
	result := &ActivationResult{
		AgentsSpawned:        0, // Stub: agents not spawned yet
		WorkflowsInitialized: 0, // Stub: workflows not initialized yet
		MonitoringEnabled:    agencyDoc.Settings.MonitoringEnabled,
		ActivatedAt:          now,
	}

	return result, nil
}

// Deactivate transitions agency from active to stopped
func (s *publicationService) Deactivate(ctx context.Context, agencyID string, graceful bool) error {
	s.logger.Info("deactivating agency", "agency_id", agencyID, "graceful", graceful)

	// Retrieve agency
	agencyDoc, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Check state (must be active or paused)
	if agencyDoc.State != models.AgencyStateActive && agencyDoc.State != models.AgencyStatePaused {
		return fmt.Errorf("agency must be in active or paused state to deactivate (current: %s)", agencyDoc.State)
	}

	// Choose transition event
	event := "force_stop"
	if graceful {
		event = "drain"
	}

	// Transition state
	if err := s.stateMachine.Transition(agencyDoc, event); err != nil {
		return fmt.Errorf("failed to transition: %w", err)
	}

	// Update agency
	if err := s.agencyRepo.Update(ctx, agencyDoc); err != nil {
		return fmt.Errorf("failed to update agency: %w", err)
	}

	// If graceful, wait for drain completion (stub for MVP-PUB-004)
	if graceful {
		s.logger.Info("draining agency", "agency_id", agencyID)
		// TODO (MVP-PUB-004): Wait for work to complete
		// For now, immediately transition to stopped
		if err := s.stateMachine.Transition(agencyDoc, "drain_complete"); err != nil {
			return fmt.Errorf("failed to complete drain: %w", err)
		}
		if err := s.agencyRepo.Update(ctx, agencyDoc); err != nil {
			return fmt.Errorf("failed to update agency after drain: %w", err)
		}
	}

	s.logger.Info("agency deactivated",
		"agency_id", agencyID,
		"final_state", agencyDoc.State)

	return nil
}

// GetPublicationHistory retrieves all publications for an agency
func (s *publicationService) GetPublicationHistory(ctx context.Context, agencyID string) ([]*models.AgencyPublication, error) {
	s.logger.Info("retrieving publication history", "agency_id", agencyID)

	publications, err := s.pubRepo.ListByAgency(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list publications: %w", err)
	}

	s.logger.Info("publication history retrieved",
		"agency_id", agencyID,
		"count", len(publications))

	return publications, nil
}

// RepublishFromTag creates new publication from a tag (future)
func (s *publicationService) RepublishFromTag(ctx context.Context, tagID string) (*models.AgencyPublication, error) {
	// TODO (MVP-PUB-006): Implement tag-based republishing
	return nil, fmt.Errorf("republish from tag not yet implemented")
}

// ===== Helper Functions =====

// generateSnapshot creates a snapshot of agency state
func (s *publicationService) generateSnapshot(agencyDoc *models.Agency, spec *models.AgencySpecification) models.AgencySnapshot {
	snapshot := models.AgencySnapshot{
		Specification: *spec,
		Settings:      agencyDoc.Settings,
		Metadata:      agencyDoc.Metadata,
	}

	// Add AI policy if present in spec
	if spec.AIPolicy != nil {
		snapshot.AIPolicy = models.AIPolicy{
			Enabled:          true,
			RiskLevel:        "medium",
			ApprovalRequired: false,
			Rules:            []models.PolicyRule{},
		}
	}

	return snapshot
}

// generateDeploymentManifest creates deployment plan from specification
func (s *publicationService) generateDeploymentManifest(spec *models.AgencySpecification, agencyDoc *models.Agency) models.DeploymentManifest {
	// Generate agent spawn plan from roles
	agentPlan := s.generateAgentSpawnPlan(spec)

	// Generate workflow execution plan
	workflowPlan := s.generateWorkflowExecution(spec)

	// Calculate resource allocation
	resources := s.calculateResourceAllocation(agentPlan)

	// Generate monitoring configuration
	monitoring := s.generateMonitoringConfig(agencyDoc)

	manifest := models.DeploymentManifest{
		AgentSpawnPlan:     agentPlan,
		WorkflowExecution:  workflowPlan,
		ResourceAllocation: resources,
		MonitoringConfig:   monitoring,
	}

	return manifest
}

// generateAgentSpawnPlan creates agent spawn plan from roles
func (s *publicationService) generateAgentSpawnPlan(spec *models.AgencySpecification) models.AgentSpawnPlan {
	agents := []models.AgentDefinition{}

	for _, role := range spec.Roles {
		// Default resource limits
		cpuLimit := "500m"  // 0.5 CPU cores
		memLimit := "512Mi" // 512 MB

		// Adjust based on role type or autonomy level
		if role.AutonomyLevel == "L3" || role.AutonomyLevel == "L4" {
			cpuLimit = "1000m" // 1 CPU core for higher autonomy
			memLimit = "1Gi"   // 1 GB
		}

		// Default token budget
		tokenBudget := 10000
		if role.TokenBudget > 0 {
			tokenBudget = int(role.TokenBudget)
		}

		agent := models.AgentDefinition{
			RoleCode:      role.Code,
			Name:          role.Name,
			Type:          "general", // Default type
			AutonomyLevel: role.AutonomyLevel,
			ResourceLimits: models.ResourceLimits{
				CPULimit:    cpuLimit,
				MemoryLimit: memLimit,
				TokenBudget: tokenBudget,
			},
			Configuration: map[string]interface{}{
				"role_description": role.Description,
				"role_code":        role.Code,
			},
		}

		agents = append(agents, agent)
	}

	return models.AgentSpawnPlan{
		TotalAgents: len(agents),
		Agents:      agents,
	}
}

// generateWorkflowExecution creates workflow execution plan
func (s *publicationService) generateWorkflowExecution(spec *models.AgencySpecification) models.WorkflowExecution {
	workflows := []models.WorkflowConfig{}

	for _, wf := range spec.Workflows {
		config := models.WorkflowConfig{
			WorkflowID: wf.ID,
			Name:       wf.Name,
			Enabled:    true,
			AutoStart:  false, // Default: manual start
		}
		workflows = append(workflows, config)
	}

	return models.WorkflowExecution{
		Workflows: workflows,
	}
}

// calculateResourceAllocation computes resource quotas
func (s *publicationService) calculateResourceAllocation(agentPlan models.AgentSpawnPlan) models.ResourceAllocation {
	totalCPU := 0
	totalMemory := 0

	for _, agent := range agentPlan.Agents {
		// Parse CPU (assume "500m" format)
		switch agent.ResourceLimits.CPULimit {
		case "500m":
			totalCPU += 500
		case "1000m":
			totalCPU += 1000
		}

		// Parse memory (assume "512Mi" or "1Gi" format)
		switch agent.ResourceLimits.MemoryLimit {
		case "512Mi":
			totalMemory += 512
		case "1Gi":
			totalMemory += 1024
		}
	}

	// Format totals
	totalCPUStr := fmt.Sprintf("%dm", totalCPU)
	totalMemStr := fmt.Sprintf("%dMi", totalMemory)

	return models.ResourceAllocation{
		TotalCPU:    totalCPUStr,
		TotalMemory: totalMemStr,
		MaxAgents:   agentPlan.TotalAgents,
	}
}

// generateMonitoringConfig creates monitoring configuration
func (s *publicationService) generateMonitoringConfig(agencyDoc *models.Agency) models.MonitoringConfig {
	return models.MonitoringConfig{
		Enabled:         agencyDoc.Settings.MonitoringEnabled,
		MetricsEndpoint: fmt.Sprintf("/api/v1/agencies/%s/metrics", agencyDoc.ID),
		Alerts: []string{
			"agent_failure",
			"high_error_rate",
			"resource_exhaustion",
		},
	}
}
