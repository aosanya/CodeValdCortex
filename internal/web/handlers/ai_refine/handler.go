package ai_refine

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/sirupsen/logrus"
)

// Handler handles AI refinement requests for agency components
type Handler struct {
	agencyService       agency.Service
	introductionRefiner *ai.IntroductionRefiner
	goalRefiner         *ai.GoalRefiner
	designerService     *ai.AgencyDesignerService
	logger              *logrus.Logger
}

// NewHandler creates a new AI refine handler
func NewHandler(
	agencyService agency.Service,
	introductionRefiner *ai.IntroductionRefiner,
	goalRefiner *ai.GoalRefiner,
	designerService *ai.AgencyDesignerService,
	logger *logrus.Logger,
) *Handler {
	return &Handler{
		agencyService:       agencyService,
		introductionRefiner: introductionRefiner,
		goalRefiner:         goalRefiner,
		designerService:     designerService,
		logger:              logger,
	}
}
