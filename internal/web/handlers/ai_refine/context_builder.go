package ai_refine

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/sirupsen/logrus"
)

// AIContextBuilder provides methods to build AI context from agency data
type AIContextBuilder struct {
	agencyService agency.Service
	roleService   registry.RoleService
	logger        *logrus.Logger
}

// NewAIContextBuilder creates a new AI context builder
func NewAIContextBuilder(agencyService agency.Service, roleService registry.RoleService, logger *logrus.Logger) *AIContextBuilder {
	return &AIContextBuilder{
		agencyService: agencyService,
		roleService:   roleService,
		logger:        logger,
	}
}

// BuildAIContext gathers all agency context data and returns it as a structured AIContext
// This is the centralized function used by all AI operations to ensure consistent context
func (b *AIContextBuilder) BuildAIContext(ctx context.Context, agencyObj *agency.Agency, currentIntroduction string, userRequest string) (ai.AIContext, error) {
	b.logger.WithField("agency_id", agencyObj.ID).Debug("Building AI context data")

	// Get all goals for context
	goals, err := b.agencyService.GetGoals(ctx, agencyObj.ID)
	if err != nil {
		b.logger.WithError(err).Warn("Failed to fetch goals, continuing without them")
		goals = []*agency.Goal{}
	}

	// Get all units of work for context
	workItems, err := b.agencyService.GetWorkItems(ctx, agencyObj.ID)
	if err != nil {
		b.logger.WithError(err).Warn("Failed to fetch units of work, continuing without them")
		workItems = []*agency.WorkItem{}
	}

	// Get all roles for context
	roles, err := b.roleService.ListTypes(ctx)
	if err != nil {
		b.logger.WithError(err).Warn("Failed to fetch roles, continuing without them")
		roles = []*registry.Role{}
	}

	// Get RACI assignments for context
	assignments, err := b.agencyService.GetAllRACIAssignments(ctx, agencyObj.ID)
	if err != nil {
		b.logger.WithError(err).Warn("Failed to fetch RACI assignments, continuing without them")
		assignments = []*agency.RACIAssignment{}
	}

	aiContext := ai.AIContext{
		// Agency metadata
		AgencyName:        agencyObj.DisplayName,
		AgencyCategory:    agencyObj.Category,
		AgencyDescription: agencyObj.Description,

		// Agency working data
		Introduction: currentIntroduction,
		Goals:        goals,
		WorkItems:    workItems,
		Roles:        roles,
		Assignments:  assignments,
		UserInput:    userRequest,
	}

	b.logger.WithFields(logrus.Fields{
		"agency_id":         agencyObj.ID,
		"agency_name":       agencyObj.DisplayName,
		"goals_count":       len(goals),
		"work_items_count":  len(workItems),
		"roles_count":       len(roles),
		"assignments_count": len(assignments),
		"has_user_input":    userRequest != "",
	}).Debug("AI context data built successfully")

	return aiContext, nil
}
