package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/lifecycle"
	"github.com/aosanya/CodeValdCortex/internal/orchestration"
)

// ActivationService manages agency lifecycle operations (activation, pause, resume, stop)
type ActivationService interface {
	// SpawnAgents spawns agents based on publication manifest
	SpawnAgents(ctx context.Context, publicationID string) (*AgentSpawnResult, error)

	// InitializeWorkflows starts workflows for the activated agency
	InitializeWorkflows(ctx context.Context, publicationID string) (*WorkflowInitResult, error)

	// StartMonitoring enables health monitoring for agency agents
	StartMonitoring(ctx context.Context, agencyID string) error

	// PauseAgency stops accepting work and pauses all agents
	PauseAgency(ctx context.Context, agencyID string) error

	// ResumeAgency resumes paused agency
	ResumeAgency(ctx context.Context, agencyID string) error

	// DrainAgency completes existing work, stops accepting new
	DrainAgency(ctx context.Context, agencyID string) error

	// StopAgency force stops all agents
	StopAgency(ctx context.Context, agencyID string, force bool) error
}

// AgentSpawnResult contains results of agent spawning operation
type AgentSpawnResult struct {
	TotalAgents   int                `json:"total_agents"`
	SpawnedAgents []SpawnedAgentInfo `json:"spawned_agents"`
	Failures      []SpawnFailure     `json:"failures,omitempty"`
}

// SpawnedAgentInfo contains information about a spawned agent
type SpawnedAgentInfo struct {
	AgentID   string    `json:"agent_id"`
	RoleCode  string    `json:"role_code"`
	State     string    `json:"state"`
	SpawnedAt time.Time `json:"spawned_at"`
}

// SpawnFailure represents a failed agent spawn attempt
type SpawnFailure struct {
	RoleCode string `json:"role_code"`
	Reason   string `json:"reason"`
}

// WorkflowInitResult contains results of workflow initialization
type WorkflowInitResult struct {
	TotalWorkflows       int      `json:"total_workflows"`
	InitializedWorkflows []string `json:"initialized_workflows"`
	Failures             []string `json:"failures,omitempty"`
}

// activationService is the concrete implementation
type activationService struct {
	pubRepo          PublicationRepository
	agencyRepo       agency.Repository
	lifecycleManager *lifecycle.Manager
	workflowEngine   *orchestration.Engine
	logger           *slog.Logger

	// Track active agents by agency
	mu           sync.RWMutex
	agencyAgents map[string][]string // agencyID -> []agentID
}

// NewActivationService creates a new activation service
func NewActivationService(
	pubRepo PublicationRepository,
	agencyRepo agency.Repository,
	lifecycleManager *lifecycle.Manager,
	workflowEngine *orchestration.Engine,
	logger *slog.Logger,
) ActivationService {
	return &activationService{
		pubRepo:          pubRepo,
		agencyRepo:       agencyRepo,
		lifecycleManager: lifecycleManager,
		workflowEngine:   workflowEngine,
		logger:           logger,
		agencyAgents:     make(map[string][]string),
	}
}

// SpawnAgents spawns agents based on the publication manifest
func (s *activationService) SpawnAgents(ctx context.Context, publicationID string) (*AgentSpawnResult, error) {
	s.logger.Info("spawning agents for publication", "publication_id", publicationID)

	// Get publication
	pub, err := s.pubRepo.GetByID(ctx, publicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get publication: %w", err)
	}

	// Get agency spawn plan from publication manifest
	spawnPlan := pub.Manifest.AgentSpawnPlan
	result := &AgentSpawnResult{
		TotalAgents:   spawnPlan.TotalAgents,
		SpawnedAgents: []SpawnedAgentInfo{},
		Failures:      []SpawnFailure{},
	}

	var agentIDs []string

	// Spawn each agent defined in the plan
	for _, agentDef := range spawnPlan.Agents {
		// Use default config values (can be extended later)
		agentConfig := agent.Config{
			MaxConcurrentTasks: 5,
			TaskQueueSize:      100,
			HeartbeatInterval:  30 * time.Second,
			TaskTimeout:        5 * time.Minute,
		}

		// Create agent via lifecycle manager
		createdAgent, err := s.lifecycleManager.CreateAgent(ctx, agentDef.Name, agentDef.Type, agentConfig)
		if err != nil {
			s.logger.Error("failed to spawn agent", "role_code", agentDef.RoleCode, "error", err)
			result.Failures = append(result.Failures, SpawnFailure{
				RoleCode: agentDef.RoleCode,
				Reason:   err.Error(),
			})
			continue
		}

		// Start the agent immediately
		if err := s.lifecycleManager.StartAgent(ctx, createdAgent.ID); err != nil {
			s.logger.Error("failed to start agent", "agent_id", createdAgent.ID, "error", err)
			result.Failures = append(result.Failures, SpawnFailure{
				RoleCode: agentDef.RoleCode,
				Reason:   fmt.Sprintf("start failed: %s", err.Error()),
			})
			continue
		}

		// Track spawned agent
		agentIDs = append(agentIDs, createdAgent.ID)
		result.SpawnedAgents = append(result.SpawnedAgents, SpawnedAgentInfo{
			AgentID:   createdAgent.ID,
			RoleCode:  agentDef.RoleCode,
			State:     string(createdAgent.State),
			SpawnedAt: time.Now(),
		})

		s.logger.Info("agent spawned and started",
			"agent_id", createdAgent.ID,
			"role_code", agentDef.RoleCode,
			"name", agentDef.Name)
	}

	// Track agents for this agency
	s.mu.Lock()
	s.agencyAgents[pub.AgencyID] = agentIDs
	s.mu.Unlock()

	s.logger.Info("agent spawning completed",
		"publication_id", publicationID,
		"total_spawned", len(result.SpawnedAgents),
		"total_failures", len(result.Failures))

	return result, nil
}

// InitializeWorkflows starts workflows for the agency
func (s *activationService) InitializeWorkflows(ctx context.Context, publicationID string) (*WorkflowInitResult, error) {
	s.logger.Info("initializing workflows for publication", "publication_id", publicationID)

	// Get publication
	pub, err := s.pubRepo.GetByID(ctx, publicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get publication: %w", err)
	}

	// Get workflow execution plan from manifest
	workflowExec := pub.Manifest.WorkflowExecution
	result := &WorkflowInitResult{
		TotalWorkflows:       len(workflowExec.Workflows),
		InitializedWorkflows: []string{},
		Failures:             []string{},
	}

	// Initialize each workflow
	for _, workflow := range workflowExec.Workflows {
		// TODO: Implement workflow initialization using workflowEngine
		// For now, just track workflow IDs
		result.InitializedWorkflows = append(result.InitializedWorkflows, workflow.Name)

		s.logger.Info("workflow initialized",
			"workflow_name", workflow.Name,
			"enabled", workflow.Enabled)
	}

	s.logger.Info("workflow initialization completed",
		"publication_id", publicationID,
		"total_initialized", len(result.InitializedWorkflows))

	return result, nil
}

// StartMonitoring enables health monitoring for agency
func (s *activationService) StartMonitoring(ctx context.Context, agencyID string) error {
	s.logger.Info("starting monitoring for agency", "agency_id", agencyID)

	// Get agents for this agency
	s.mu.RLock()
	agentIDs, exists := s.agencyAgents[agencyID]
	s.mu.RUnlock()

	if !exists || len(agentIDs) == 0 {
		return fmt.Errorf("no agents found for agency: %s", agencyID)
	}

	// TODO: Integrate with health monitoring system
	// For now, just log that monitoring is enabled
	s.logger.Info("monitoring enabled",
		"agency_id", agencyID,
		"agent_count", len(agentIDs))

	return nil
}

// PauseAgency pauses all agents in the agency
func (s *activationService) PauseAgency(ctx context.Context, agencyID string) error {
	s.logger.Info("pausing agency", "agency_id", agencyID)

	// Get agents for this agency
	s.mu.RLock()
	agentIDs, exists := s.agencyAgents[agencyID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agency not activated: %s", agencyID)
	}

	// Pause each agent
	var pauseErrors []string
	for _, agentID := range agentIDs {
		if err := s.lifecycleManager.PauseAgent(ctx, agentID); err != nil {
			s.logger.Error("failed to pause agent", "agent_id", agentID, "error", err)
			pauseErrors = append(pauseErrors, fmt.Sprintf("%s: %s", agentID, err.Error()))
		}
	}

	if len(pauseErrors) > 0 {
		return fmt.Errorf("failed to pause some agents: %v", pauseErrors)
	}

	s.logger.Info("agency paused", "agency_id", agencyID)
	return nil
}

// ResumeAgency resumes all paused agents in the agency
func (s *activationService) ResumeAgency(ctx context.Context, agencyID string) error {
	s.logger.Info("resuming agency", "agency_id", agencyID)

	// Get agents for this agency
	s.mu.RLock()
	agentIDs, exists := s.agencyAgents[agencyID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agency not activated: %s", agencyID)
	}

	// Resume each agent
	var resumeErrors []string
	for _, agentID := range agentIDs {
		if err := s.lifecycleManager.ResumeAgent(ctx, agentID); err != nil {
			s.logger.Error("failed to resume agent", "agent_id", agentID, "error", err)
			resumeErrors = append(resumeErrors, fmt.Sprintf("%s: %s", agentID, err.Error()))
		}
	}

	if len(resumeErrors) > 0 {
		return fmt.Errorf("failed to resume some agents: %v", resumeErrors)
	}

	s.logger.Info("agency resumed", "agency_id", agencyID)
	return nil
}

// DrainAgency gracefully drains all agents (complete work, no new tasks)
func (s *activationService) DrainAgency(ctx context.Context, agencyID string) error {
	s.logger.Info("draining agency", "agency_id", agencyID)

	// Get agents for this agency
	s.mu.RLock()
	agentIDs, exists := s.agencyAgents[agencyID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agency not activated: %s", agencyID)
	}

	// Stop each agent gracefully (allows completion of current work)
	var drainErrors []string
	for _, agentID := range agentIDs {
		if err := s.lifecycleManager.StopAgent(ctx, agentID); err != nil {
			s.logger.Error("failed to drain agent", "agent_id", agentID, "error", err)
			drainErrors = append(drainErrors, fmt.Sprintf("%s: %s", agentID, err.Error()))
		}
	}

	if len(drainErrors) > 0 {
		return fmt.Errorf("failed to drain some agents: %v", drainErrors)
	}

	s.logger.Info("agency drained", "agency_id", agencyID)
	return nil
}

// StopAgency stops all agents in the agency
func (s *activationService) StopAgency(ctx context.Context, agencyID string, force bool) error {
	s.logger.Info("stopping agency", "agency_id", agencyID, "force", force)

	// Get agents for this agency
	s.mu.RLock()
	agentIDs, exists := s.agencyAgents[agencyID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agency not activated: %s", agencyID)
	}

	// Stop each agent
	var stopErrors []string
	for _, agentID := range agentIDs {
		// lifecycle.Manager.StopAgent is always graceful
		// Force stop would require Delete operation
		if err := s.lifecycleManager.StopAgent(ctx, agentID); err != nil {
			s.logger.Error("failed to stop agent", "agent_id", agentID, "error", err)
			stopErrors = append(stopErrors, fmt.Sprintf("%s: %s", agentID, err.Error()))
		}

		// If force stop requested, delete the agent
		if force {
			if err := s.lifecycleManager.DeleteAgent(ctx, agentID); err != nil {
				s.logger.Error("failed to delete agent", "agent_id", agentID, "error", err)
			}
		}
	}

	if len(stopErrors) > 0 {
		return fmt.Errorf("failed to stop some agents: %v", stopErrors)
	}

	// Remove agency from tracking
	s.mu.Lock()
	delete(s.agencyAgents, agencyID)
	s.mu.Unlock()

	s.logger.Info("agency stopped", "agency_id", agencyID)
	return nil
}
