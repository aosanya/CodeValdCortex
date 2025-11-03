package ai_refine

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/sirupsen/logrus"
)

// Handler handles AI refinement requests for agency components
type Handler struct {
	agencyService        agency.Service
	roleService          registry.RoleService
	introductionRefiner  *ai.IntroductionRefiner
	goalRefiner          *ai.GoalRefiner
	goalConsolidator     *ai.GoalConsolidator
	workItemRefiner      *ai.WorkItemRefiner
	workItemConsolidator *ai.WorkItemConsolidator
	roleCreator          *ai.RoleCreator
	raciCreator          *ai.RACICreator
	designerService      *ai.AgencyDesignerService
	logger               *logrus.Logger
}

// NewHandler creates a new AI refine handler
func NewHandler(
	agencyService agency.Service,
	roleService registry.RoleService,
	introductionRefiner *ai.IntroductionRefiner,
	goalRefiner *ai.GoalRefiner,
	goalConsolidator *ai.GoalConsolidator,
	workItemRefiner *ai.WorkItemRefiner,
	workItemConsolidator *ai.WorkItemConsolidator,
	roleCreator *ai.RoleCreator,
	raciCreator *ai.RACICreator,
	designerService *ai.AgencyDesignerService,
	logger *logrus.Logger,
) *Handler {
	return &Handler{
		agencyService:        agencyService,
		roleService:          roleService,
		introductionRefiner:  introductionRefiner,
		goalRefiner:          goalRefiner,
		goalConsolidator:     goalConsolidator,
		workItemRefiner:      workItemRefiner,
		workItemConsolidator: workItemConsolidator,
		roleCreator:          roleCreator,
		raciCreator:          raciCreator,
		designerService:      designerService,
		logger:               logger,
	}
}
