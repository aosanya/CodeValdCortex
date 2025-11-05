package ai_refine

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/sirupsen/logrus"
)

// BuilderContextBuilder provides methods to build AI context from agency data
type BuilderContextBuilder struct {
	agencyService agency.Service
	roleService   registry.RoleService
	logger        *logrus.Logger
}

// NewBuilderContextBuilder creates a new AI context builder
func NewBuilderContextBuilder(agencyService agency.Service, roleService registry.RoleService, logger *logrus.Logger) *BuilderContextBuilder {
	return &BuilderContextBuilder{
		agencyService: agencyService,
		roleService:   roleService,
		logger:        logger,
	}
}

// BuildBuilderContext gathers all agency context data and returns it as a structured BuilderContext
// This is the centralized function used by all AI operations to ensure consistent context
func (b *BuilderContextBuilder) BuildBuilderContext(ctx context.Context, agencyObj *agency.Agency, currentIntroduction string, userRequest string) (builder.BuilderContext, error) {
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

	builderContext := builder.BuilderContext{
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

	return builderContext, nil
}
