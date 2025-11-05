package ai_refine

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/sirupsen/logrus"
)

// Handler handles AI refinement requests for agency components
type Handler struct {
	agencyService       agency.Service
	roleService         registry.RoleService
	introductionRefiner *ai.IntroductionBuilder
	goalRefiner         *ai.GoalsBuilder
	workItemBuilder     *ai.WorkItemsBuilder
	roleBuilder         *ai.RolesBuilder
	raciBuilder         *ai.RACIBuilder
	designerService     *ai.AgencyDesignerService
	contextBuilder      *BuilderContextBuilder
	logger              *logrus.Logger
}

// NewHandler creates a new AI refine handler
func NewHandler(
	agencyService agency.Service,
	roleService registry.RoleService,
	introductionRefiner *ai.IntroductionBuilder,
	goalRefiner *ai.GoalsBuilder,
	workItemBuilder *ai.WorkItemsBuilder,
	roleBuilder *ai.RolesBuilder,
	raciBuilder *ai.RACIBuilder,
	designerService *ai.AgencyDesignerService,
	logger *logrus.Logger,
) *Handler {
	// Create context builder for shared AI context gathering
	contextBuilder := NewBuilderContextBuilder(agencyService, roleService, logger)

	return &Handler{
		agencyService:       agencyService,
		roleService:         roleService,
		introductionRefiner: introductionRefiner,
		goalRefiner:         goalRefiner,
		workItemBuilder:     workItemBuilder,
		roleBuilder:         roleBuilder,
		raciBuilder:         raciBuilder,
		designerService:     designerService,
		contextBuilder:      contextBuilder,
		logger:              logger,
	}
}
