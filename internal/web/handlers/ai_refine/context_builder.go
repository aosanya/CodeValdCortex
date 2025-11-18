package ai_refine

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/sirupsen/logrus"
)

// BuilderContextBuilder provides methods to build AI context from agency data
type BuilderContextBuilder struct {
	agencyService agency.Service
	logger        *logrus.Logger
}

// NewBuilderContextBuilder creates a new AI context builder
func NewBuilderContextBuilder(agencyService agency.Service, logger *logrus.Logger) *BuilderContextBuilder {
	return &BuilderContextBuilder{
		agencyService: agencyService,
		logger:        logger,
	}
}

// BuildBuilderContext gathers all agency context data and returns it as a structured BuilderContext
// This is the centralized function used by all AI operations to ensure consistent context
func (b *BuilderContextBuilder) BuildBuilderContext(ctx context.Context, agencyObj *models.Agency, currentIntroduction string, userRequest string) (builder.BuilderContext, error) {

	// Get unified specification (replaces separate GetGoals, GetWorkItems, GetOverview calls)
	spec, err := b.agencyService.GetSpecification(ctx, agencyObj.ID)
	if err != nil {
		b.logger.WithError(err).Warn("Failed to fetch specification, using empty context")
		spec = &models.AgencySpecification{
			Goals:     []models.Goal{},
			WorkItems: []models.WorkItem{},
		}
	}

	// Convert goals from []Goal to []*Goal for compatibility
	goals := make([]*models.Goal, len(spec.Goals))
	for i := range spec.Goals {
		goals[i] = &spec.Goals[i]
	}

	// Convert work items from []WorkItem to []*WorkItem for compatibility
	workItems := make([]*models.WorkItem, len(spec.WorkItems))
	for i := range spec.WorkItems {
		workItems[i] = &spec.WorkItems[i]
	}

	// Convert roles from []Role to []*Role for compatibility
	roles := make([]*models.Role, len(spec.Roles))
	for i := range spec.Roles {
		roles[i] = &spec.Roles[i]
	}

	// Convert workflows from []Workflow to []*Workflow for compatibility
	workflows := make([]*models.Workflow, len(spec.Workflows))
	for i := range spec.Workflows {
		workflows[i] = &spec.Workflows[i]
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
		Workflows:    workflows,
		Roles:        roles,
		Assignments:  []*models.RACIAssignment{}, // RACI now in specification.RACIMatrix
		UserInput:    userRequest,
	}

	b.logger.WithFields(logrus.Fields{
		"agency_id":        agencyObj.ID,
		"agency_name":      agencyObj.DisplayName,
		"goals_count":      len(goals),
		"work_items_count": len(workItems),
		"workflows_count":  len(workflows),
		"roles_count":      len(roles),
		"has_user_input":   userRequest != "",
	}).Info("Built refine context")

	return builderContext, nil
}
