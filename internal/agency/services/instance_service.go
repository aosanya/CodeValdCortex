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
		"instance_name", req.InstanceName,
	)

	// Get the agency to validate it exists and get database name
	agencyModel, err := s.agencyRepo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	if agencyModel.Database == "" {
		return nil, fmt.Errorf("agency does not have a database configured")
	}

	// Get the tag to load the snapshot
	tag, err := s.tagService.GetTag(ctx, agencyID, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	// Create the instance model
	createdBy := ""
	if req.Metadata != nil {
		if cb, ok := req.Metadata["created_by"].(string); ok {
			createdBy = cb
		}
	}

	instance := &models.AgencyInstance{
		InstanceID:   uuid.New().String(),
		AgencyID:     agencyID,
		TagName:      tagName,
		InstanceName: req.InstanceName,
		State:        models.InstanceStateStarting,
		StateMessage: "Instance is being initialized",
		Snapshot:     tag.Snapshot,
		Config:       req.Config,
		Resources: models.InstanceResourceStatus{
			ActiveAgents:        0,
			SpawnedAgents:       0,
			TerminatedAgents:    0,
			HealthScore:         1.0,
			LastHealthCheck:     time.Now(),
			ConsecutiveFailures: 0,
		},
		StartedAt:   time.Now(),
		LastSeenAt:  time.Now(),
		CreatedBy:   createdBy,
		Environment: req.Environment,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
	}

	// Persist the instance
	if err := s.instanceRepo.Create(ctx, instance, agencyModel.Database); err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Spawn agents from the snapshot (async operation)
	go s.spawnInstanceAgents(context.Background(), instance, agencyModel.Database)

	// Update state to running
	instance.State = models.InstanceStateRunning
	instance.StateMessage = "Instance is operational"
	if err := s.instanceRepo.Update(ctx, instance, agencyModel.Database); err != nil {
		s.logger.Error("Failed to update instance state", "error", err)
		// Don't fail the operation, instance is created
	}

	s.logger.Info("Instance started successfully",
		"instance_id", instance.InstanceID,
		"agency_id", agencyID,
		"tag_name", tagName,
	)

	return instance, nil
}

// spawnInstanceAgents spawns agents for an instance based on the deployment manifest
func (s *instanceService) spawnInstanceAgents(ctx context.Context, instance *models.AgencyInstance, agencyDB string) {
	s.logger.Info("Spawning agents for instance", "instance_id", instance.InstanceID)

	// Get agent spawn plan from snapshot
	agentCount := len(instance.Snapshot.Specification.Roles)

	// Create instance agents (placeholder - actual agent spawning would integrate with agent factory)
	for i, role := range instance.Snapshot.Specification.Roles {
		agent := &models.InstanceAgent{
			AgentID:       fmt.Sprintf("%s-agent-%d", instance.InstanceID, i),
			InstanceID:    instance.InstanceID,
			AgencyID:      instance.AgencyID,
			RoleCode:      role.Code,
			Name:          role.Name,
			Type:          "worker", // Default type
			AutonomyLevel: "supervised",
			State:         "running",
			SpawnedAt:     time.Now(),
			TaskCount:     0,
			Metadata:      make(map[string]interface{}),
		}

		if err := s.instanceRepo.CreateAgent(ctx, agent, agencyDB); err != nil {
			s.logger.Error("Failed to create instance agent",
				"instance_id", instance.InstanceID,
				"agent_id", agent.AgentID,
				"error", err,
			)
			continue
		}

		s.logger.Info("Agent spawned",
			"instance_id", instance.InstanceID,
			"agent_id", agent.AgentID,
			"role", role.Code,
		)
	}

	// Update instance resource status
	instance.Resources.SpawnedAgents = agentCount
	instance.Resources.ActiveAgents = agentCount

	if err := s.instanceRepo.Update(ctx, instance, agencyDB); err != nil {
		s.logger.Error("Failed to update instance resources", "error", err)
	}

	s.logger.Info("Agent spawning completed",
		"instance_id", instance.InstanceID,
		"spawned_count", agentCount,
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

	// Update state
	instance.State = models.InstanceStateStopping
	instance.StateMessage = "Instance is shutting down"
	if err := s.instanceRepo.Update(ctx, instance, agencyModel.Database); err != nil {
		return fmt.Errorf("failed to update instance state: %w", err)
	}

	// Stop agents (placeholder - actual agent stopping would integrate with agent factory)
	// In a real implementation, this would gracefully stop all agents

	// Mark instance as stopped
	now := time.Now()
	instance.State = models.InstanceStateStopped
	instance.StateMessage = "Instance has been stopped"
	instance.StoppedAt = &now
	instance.Resources.ActiveAgents = 0

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

	// Update state
	instance.State = models.InstanceStateStarting
	instance.StateMessage = "Instance is restarting"
	instance.StoppedAt = nil
	instance.LastSeenAt = time.Now()

	if err := s.instanceRepo.Update(ctx, instance, agencyModel.Database); err != nil {
		return nil, fmt.Errorf("failed to update instance state: %w", err)
	}

	// Restart agents (placeholder - actual agent restarting would integrate with agent factory)
	go s.spawnInstanceAgents(context.Background(), instance, agencyModel.Database)

	// Update to running
	instance.State = models.InstanceStateRunning
	instance.StateMessage = "Instance has been restarted"

	if err := s.instanceRepo.Update(ctx, instance, agencyModel.Database); err != nil {
		s.logger.Error("Failed to update instance state after restart", "error", err)
	}

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
	uptime := time.Since(instance.StartedAt)
	if instance.StoppedAt != nil {
		uptime = instance.StoppedAt.Sub(instance.StartedAt)
	}

	// Determine health status
	healthy := instance.State == models.InstanceStateRunning && instance.Resources.HealthScore > 0.5

	health := &models.InstanceHealth{
		InstanceID:       instance.InstanceID,
		State:            instance.State,
		Healthy:          healthy,
		Message:          instance.StateMessage,
		AgentsHealthy:    instance.Resources.ActiveAgents > 0,
		WorkflowsHealthy: true, // Placeholder
		ResourcesHealthy: instance.Resources.HealthScore > 0.7,
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
