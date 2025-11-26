package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/google/uuid"
)

// InstanceService defines the interface for agency instance management
type InstanceService interface {
	// StartInstance creates and starts a new instance from a tag
	StartInstance(ctx context.Context, agencyID string, tagName string, req *models.StartInstanceRequest) (*models.AgencyInstance, error)

	// StopInstance stops a running instance
	StopInstance(ctx context.Context, agencyID string, instanceID string) error

	// RestartInstance restarts a stopped instance
	RestartInstance(ctx context.Context, agencyID string, instanceID string) (*models.AgencyInstance, error)

	// DeleteInstance permanently deletes an instance
	DeleteInstance(ctx context.Context, agencyID string, instanceID string) error

	// GetInstance retrieves instance details
	GetInstance(ctx context.Context, agencyID string, instanceID string) (*models.AgencyInstance, error)

	// ListInstances lists all instances for an agency
	ListInstances(ctx context.Context, agencyID string) ([]*models.AgencyInstance, error)

	// ListInstancesByTag lists instances created from a specific tag
	ListInstancesByTag(ctx context.Context, agencyID string, tagName string) ([]*models.AgencyInstance, error)

	// GetInstanceHealth retrieves health status for an instance
	GetInstanceHealth(ctx context.Context, agencyID string, instanceID string) (*models.InstanceHealth, error)

	// ListInstanceAgents lists all agents for an instance
	ListInstanceAgents(ctx context.Context, agencyID string, instanceID string) ([]*models.InstanceAgent, error)
}

// InstanceRepository defines the interface for instance data persistence
type InstanceRepository interface {
	Create(ctx context.Context, instance *models.AgencyInstance, agencyDB string) error
	GetByID(ctx context.Context, instanceID string, agencyDB string) (*models.AgencyInstance, error)
	Update(ctx context.Context, instance *models.AgencyInstance, agencyDB string) error
	Delete(ctx context.Context, instanceID string, agencyDB string) error
	ExistsByName(ctx context.Context, agencyID string, name string, agencyDB string) (bool, error)
	ListByAgency(ctx context.Context, agencyID string, agencyDB string) ([]*models.AgencyInstance, error)
	ListByTag(ctx context.Context, agencyID string, tagName string, agencyDB string) ([]*models.AgencyInstance, error)
	ListByState(ctx context.Context, agencyID string, state models.InstanceState, agencyDB string) ([]*models.AgencyInstance, error)
	CreateAgent(ctx context.Context, agent *models.InstanceAgent, agencyDB string) error
	ListAgentsByInstance(ctx context.Context, instanceID string, agencyDB string) ([]*models.InstanceAgent, error)
	DeleteAgentsByInstance(ctx context.Context, instanceID string, agencyDB string) error
}

// instanceService implements the InstanceService interface
type instanceService struct {
	instanceRepo InstanceRepository
	tagService   TagService
	agencyRepo   agency.Repository
	logger       *slog.Logger
}

// NewInstanceService creates a new instance service
func NewInstanceService(
	instanceRepo InstanceRepository,
	tagService TagService,
	agencyRepo agency.Repository,
	logger *slog.Logger,
) InstanceService {
	return &instanceService{
		instanceRepo: instanceRepo,
		tagService:   tagService,
		agencyRepo:   agencyRepo,
		logger:       logger,
	}
}

// StartInstance creates and starts a new instance from a tag
func (s *instanceService) StartInstance(ctx context.Context, agencyID string, tagName string, req *models.StartInstanceRequest) (*models.AgencyInstance, error) {
	s.logger.Info("Starting instance from tag",
		"agency_id", agencyID,
		"tag_name", tagName,
		"instance_name", req.Name,
	)

	// Get the agency to validate it exists and get database name
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	if agencyModel.Database == "" {
		return nil, fmt.Errorf("agency does not have a database configured")
	}

	// Validate instance name is unique per agency
	exists, err := s.instanceRepo.ExistsByName(ctx, agencyID, req.Name, agencyModel.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("instance name '%s' already exists for this agency", req.Name)
	}

	// Get the tag to load the snapshot
	tag, err := s.tagService.GetTag(ctx, agencyID, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	// Create the instance model (optimistic start - immediately "running")
	now := time.Now()
	instance := &models.AgencyInstance{
		InstanceID:     uuid.New().String(),
		AgencyID:       agencyID,
		TagID:          tag.ID,
		TagName:        tagName,
		InstanceName:   req.Name,
		Description:    req.Description,
		State:          models.InstanceStateRunning, // Running immediately (optimistic start)
		DeployedAt:     now,
		StartedAt:      &now,
		DeployedBy:     "system", // TODO: Get from auth context
		AcceptsNewJobs: true,
		AgentCount:     0, // Will be updated as agents spawn
		WorkflowCount:  0,
		LastHeartbeat:  now,
		Tags:           req.Tags,
		Metadata:       req.Metadata,
	}

	// Persist the instance
	if err := s.instanceRepo.Create(ctx, instance, agencyModel.Database); err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Spawn agents from the tag snapshot (async operation - lazy initialization)
	go s.spawnInstanceAgents(context.Background(), instance, tag, agencyModel.Database)

	s.logger.Info("Instance started successfully",
		"instance_id", instance.InstanceID,
		"agency_id", agencyID,
		"tag_name", tagName,
	)

	return instance, nil
}

// spawnInstanceAgents spawns agents for an instance based on the tag's snapshot
// Note: Uses lazy initialization - agents are references, not physical entities
func (s *instanceService) spawnInstanceAgents(ctx context.Context, instance *models.AgencyInstance, tag *models.AgencyTag, agencyDB string) {
	s.logger.Info("Storing agent references for instance", "instance_id", instance.InstanceID)

	// Validate tag snapshot
	if tag.Snapshot.Specification.ID == "" {
		s.logger.Error("Tag snapshot specification is empty", "tag_id", tag.ID)
		return
	}

	// Get agent roles from specification
	roles := tag.Snapshot.Specification.Roles
	if len(roles) == 0 {
		s.logger.Warn("No roles defined in tag snapshot", "tag_id", tag.ID)
		return
	}

	agentCount := len(roles)

	// Create agent references (not physical agents - lazy initialization)
	for i, role := range roles {
		agent := &models.InstanceAgent{
			AgentID:       fmt.Sprintf("%s-agent-%d", instance.InstanceID, i),
			InstanceID:    instance.InstanceID,
			AgencyID:      instance.AgencyID,
			RoleCode:      role.Code,
			Name:          role.Name,
			Type:          "worker", // Default type
			AutonomyLevel: "supervised",
			State:         "registered", // Initial state (lazy spawn)
			SpawnedAt:     time.Now(),
			TaskCount:     0,
			Metadata:      make(map[string]interface{}),
		}

		if err := s.instanceRepo.CreateAgent(ctx, agent, agencyDB); err != nil {
			s.logger.Error("Failed to create agent reference",
				"instance_id", instance.InstanceID,
				"agent_id", agent.AgentID,
				"error", err,
			)
			continue
		}

		s.logger.Info("Agent reference stored",
			"instance_id", instance.InstanceID,
			"agent_id", agent.AgentID,
			"role", role.Code,
		)
	}

	// Update instance agent count
	instance.AgentCount = agentCount
	if err := s.instanceRepo.Update(ctx, instance, agencyDB); err != nil {
		s.logger.Error("Failed to update instance agent count", "error", err)
	}

	s.logger.Info("Agent reference storage completed",
		"instance_id", instance.InstanceID,
		"agent_count", agentCount,
	)
}

// StopInstance stops a running instance
func (s *instanceService) StopInstance(ctx context.Context, agencyID string, instanceID string) error {
	s.logger.Info("Stopping instance", "instance_id", instanceID, "agency_id", agencyID)

	// Get agency database
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Get the instance
	instance, err := s.instanceRepo.GetByID(ctx, instanceID, agencyModel.Database)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	if instance.State == models.InstanceStateStopped {
		return fmt.Errorf("instance is already stopped")
	}

	// Update state to "stopping" (reject new jobs)
	instance.State = models.InstanceStateStopping
	instance.AcceptsNewJobs = false
	if err := s.instanceRepo.Update(ctx, instance, agencyModel.Database); err != nil {
		return fmt.Errorf("failed to update instance state: %w", err)
	}

	// Stop agents (placeholder - actual agent stopping would integrate with agent factory)
	// In a real implementation, this would gracefully stop all agents

	// Mark instance as stopped
	now := time.Now()
	instance.State = models.InstanceStateStopped
	instance.StoppedAt = &now
	instance.AgentCount = 0 // All agents stopped

	if err := s.instanceRepo.Update(ctx, instance, agencyModel.Database); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	s.logger.Info("Instance stopped successfully", "instance_id", instanceID)
	return nil
}

// RestartInstance restarts a stopped instance
func (s *instanceService) RestartInstance(ctx context.Context, agencyID string, instanceID string) (*models.AgencyInstance, error) {
	s.logger.Info("Restarting instance", "instance_id", instanceID, "agency_id", agencyID)

	// Get agency database
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	// Get the instance
	instance, err := s.instanceRepo.GetByID(ctx, instanceID, agencyModel.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	if instance.State != models.InstanceStateStopped && instance.State != models.InstanceStateFailed {
		return nil, fmt.Errorf("instance must be stopped or failed to restart, current state: %s", instance.State)
	}

	// Get the tag to reload agent references
	tag, err := s.tagService.GetTag(ctx, agencyID, instance.TagName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	// Reset to running state
	now := time.Now()
	instance.State = models.InstanceStateRunning
	instance.StoppedAt = nil
	instance.StartedAt = &now
	instance.AcceptsNewJobs = true
	instance.LastHeartbeat = now

	if err := s.instanceRepo.Update(ctx, instance, agencyModel.Database); err != nil {
		return nil, fmt.Errorf("failed to update instance state: %w", err)
	}

	// Restart agents (re-store agent references)
	go s.spawnInstanceAgents(context.Background(), instance, tag, agencyModel.Database)

	s.logger.Info("Instance restarted successfully", "instance_id", instanceID)
	return instance, nil
}

// DeleteInstance permanently deletes an instance
func (s *instanceService) DeleteInstance(ctx context.Context, agencyID string, instanceID string) error {
	s.logger.Info("Deleting instance", "instance_id", instanceID, "agency_id", agencyID)

	// Get agency database
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get agency: %w", err)
	}

	// Get the instance to verify it exists and can be deleted
	instance, err := s.instanceRepo.GetByID(ctx, instanceID, agencyModel.Database)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// Stop the instance if it's running
	if instance.State == models.InstanceStateRunning {
		if err := s.StopInstance(ctx, agencyID, instanceID); err != nil {
			return fmt.Errorf("failed to stop instance before deletion: %w", err)
		}
	}

	// Delete all associated agents
	if err := s.instanceRepo.DeleteAgentsByInstance(ctx, instanceID, agencyModel.Database); err != nil {
		s.logger.Error("Failed to delete instance agents", "error", err)
		// Continue with instance deletion
	}

	// Delete the instance
	if err := s.instanceRepo.Delete(ctx, instanceID, agencyModel.Database); err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	s.logger.Info("Instance deleted successfully", "instance_id", instanceID)
	return nil
}

// GetInstance retrieves instance details
func (s *instanceService) GetInstance(ctx context.Context, agencyID string, instanceID string) (*models.AgencyInstance, error) {
	// Get agency database
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	instance, err := s.instanceRepo.GetByID(ctx, instanceID, agencyModel.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	return instance, nil
}

// ListInstances lists all instances for an agency
func (s *instanceService) ListInstances(ctx context.Context, agencyID string) ([]*models.AgencyInstance, error) {
	// Get agency database
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	instances, err := s.instanceRepo.ListByAgency(ctx, agencyID, agencyModel.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	return instances, nil
}

// ListInstancesByTag lists instances created from a specific tag
func (s *instanceService) ListInstancesByTag(ctx context.Context, agencyID string, tagName string) ([]*models.AgencyInstance, error) {
	// Get agency database
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	instances, err := s.instanceRepo.ListByTag(ctx, agencyID, tagName, agencyModel.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances by tag: %w", err)
	}

	return instances, nil
}

// GetInstanceHealth retrieves health status for an instance
func (s *instanceService) GetInstanceHealth(ctx context.Context, agencyID string, instanceID string) (*models.InstanceHealth, error) {
	// Get agency database
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	instance, err := s.instanceRepo.GetByID(ctx, instanceID, agencyModel.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Calculate uptime
	var uptime time.Duration
	if instance.StartedAt != nil {
		uptime = time.Since(*instance.StartedAt)
		if instance.StoppedAt != nil {
			uptime = instance.StoppedAt.Sub(*instance.StartedAt)
		}
	}

	// Determine health status based on state and agent count
	healthy := instance.State == models.InstanceStateRunning && instance.AgentCount > 0

	health := &models.InstanceHealth{
		InstanceID:       instance.InstanceID,
		State:            instance.State,
		Healthy:          healthy,
		Message:          instance.HealthStatus, // Use the health_status field
		AgentsHealthy:    instance.AgentCount > 0,
		WorkflowsHealthy: true, // Placeholder - would check workflow states
		ResourcesHealthy: instance.State == models.InstanceStateRunning,
		Uptime:           uptime,
		RequestCount:     0, // Placeholder - would come from metrics
		ErrorCount:       0, // Placeholder - would come from metrics
		LastErrorTime:    nil,
		CheckedAt:        time.Now(),
	}

	return health, nil
}

// ListInstanceAgents lists all agents for an instance
func (s *instanceService) ListInstanceAgents(ctx context.Context, agencyID string, instanceID string) ([]*models.InstanceAgent, error) {
	// Get agency database
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	agents, err := s.instanceRepo.ListAgentsByInstance(ctx, instanceID, agencyModel.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to list instance agents: %w", err)
	}

	return agents, nil
}
