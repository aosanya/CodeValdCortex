package agency

// Re-export all types from the types subpackage for backward compatibility

// This allows existing code to continue using agency.Agency, agency.Goal, etc.

// without needing to import agency/types// Re-export all types from the types subpackage for backward compatibility



import (// This allows existing code to continue using models.Agency, models.Goal, etc.// Re-export all types from the types subpackage for backward compatibility// Re-export all types from the types subpackage for backward compatibility

	"github.com/aosanya/CodeValdCortex/internal/agency/types"

)



// Agency typesimport (// This allows existing code to continue using models.Agency, models.Goal, etc.// This allows existing code to continue using models.Agency, models.Goal, etc.

type (

	Agency           = types.Agency	

	AgencyStatus     = types.AgencyStatus

	AgencyMetadata   = types.AgencyMetadata)// without needing to import agency/types// without needing to import agency/types

	AgencySettings   = types.AgencySettings

	AgencyFilters    = types.AgencyFilters

	AgencyUpdates    = types.AgencyUpdates

	AgencyStatistics = types.AgencyStatistics// Agency types

	

	CreateAgencyRequest = types.CreateAgencyRequesttype (

	UpdateAgencyRequest = types.UpdateAgencyRequest

)	Agency           = types.Agencyimport (import (



// Agency status constants	AgencyStatus     = types.AgencyStatus

const (

	AgencyStatusActive   = types.AgencyStatusActive	AgencyMetadata   = types.AgencyMetadata	"github.com/aosanya/CodeValdCortex/internal/agency/types"	"github.com/aosanya/CodeValdCortex/internal/agency/types"

	AgencyStatusInactive = types.AgencyStatusInactive

	AgencyStatusPaused   = types.AgencyStatusPaused	AgencySettings   = types.AgencySettings

	AgencyStatusArchived = types.AgencyStatusArchived

)	AgencyFilters    = types.AgencyFilters))



// Overview types	AgencyUpdates    = types.AgencyUpdates

type (

	Overview              = types.Overview	AgencyStatistics = types.AgencyStatistics

	UpdateOverviewRequest = types.UpdateOverviewRequest

)	



// Goal types	CreateAgencyRequest = types.CreateAgencyRequest// Agency types// Agency types

type (

	Goal              = types.Goal	UpdateAgencyRequest = types.UpdateAgencyRequest

	CreateGoalRequest = types.CreateGoalRequest

	UpdateGoalRequest = types.UpdateGoalRequest)type (type (

	GoalRefineRequest = types.GoalRefineRequest

)



// Work item types// Agency status constants	Agency           = types.Agency	Agency           = types.Agency

type (

	WorkItem                      = types.WorkItemconst (

	CreateWorkItemRequest         = types.CreateWorkItemRequest

	UpdateWorkItemRequest         = types.UpdateWorkItemRequest	AgencyStatusActive   = types.AgencyStatusActive	AgencyStatus     = types.AgencyStatus	AgencyStatus     = types.AgencyStatus

	WorkItemRefineRequest         = types.WorkItemRefineRequest

	WorkItemGoalLink              = types.WorkItemGoalLink	AgencyStatusInactive = types.AgencyStatusInactive

	CreateWorkItemGoalLinkRequest = types.CreateWorkItemGoalLinkRequest

)	AgencyStatusPaused   = types.AgencyStatusPaused	AgencyMetadata   = types.AgencyMetadata	AgencyMetadata   = types.AgencyMetadata



// RACI types	AgencyStatusArchived = types.AgencyStatusArchived

type (

	RACIRole     = types.RACIRole)	AgencySettings   = types.AgencySettings	AgencySettings   = types.AgencySettings

	RACIActivity = types.RACIActivity

	RACIMatrix   = types.RACIMatrix

	RACITemplate = types.RACITemplate

	// Overview types	AgencyFilters    = types.AgencyFilters	AgencyFilters    = types.AgencyFilters

	CreateRACIMatrixRequest = types.CreateRACIMatrixRequest

	UpdateRACIMatrixRequest = types.UpdateRACIMatrixRequesttype (

	

	RACIValidationResult  = types.RACIValidationResult	Overview              = types.Overview	AgencyUpdates    = types.AgencyUpdates	AgencyUpdates    = types.AgencyUpdates

	RACIValidationError   = types.RACIValidationError

	RACIValidationWarning = types.RACIValidationWarning	UpdateOverviewRequest = types.UpdateOverviewRequest

	RACIValidationSummary = types.RACIValidationSummary

	)	AgencyStatistics = types.AgencyStatistics	AgencyStatistics = types.AgencyStatistics

	RACIExportFormat = types.RACIExportFormat

	

	RACIAssignment              = types.RACIAssignment

	CreateRACIAssignmentRequest = types.CreateRACIAssignmentRequest// Goal types		

	

	AgencyRACIAssignments = types.AgencyRACIAssignmentstype (

	RoleAssignment        = types.RoleAssignment

)	Goal              = types.Goal	CreateAgencyRequest = types.CreateAgencyRequest	CreateAgencyRequest = types.CreateAgencyRequest



// RACI role constants	CreateGoalRequest = types.CreateGoalRequest

const (

	RACIResponsible = types.RACIResponsible	UpdateGoalRequest = types.UpdateGoalRequest	UpdateAgencyRequest = types.UpdateAgencyRequest	UpdateAgencyRequest = types.UpdateAgencyRequest

	RACIAccountable = types.RACIAccountable

	RACIConsulted   = types.RACIConsulted	GoalRefineRequest = types.GoalRefineRequest

	RACIInformed    = types.RACIInformed

))))



// RACI export format constants

const (

	RACIExportPDF      = types.RACIExportPDF// Work item types

	RACIExportMarkdown = types.RACIExportMarkdown

	RACIExportJSON     = types.RACIExportJSONtype (

)

	WorkItem                      = types.WorkItem// Agency status constants// Agency status constants

	CreateWorkItemRequest         = types.CreateWorkItemRequest

	UpdateWorkItemRequest         = types.UpdateWorkItemRequestconst (const (

	WorkItemRefineRequest         = types.WorkItemRefineRequest

	WorkItemGoalLink              = types.WorkItemGoalLink	AgencyStatusActive   = types.AgencyStatusActive	AgencyStatusActive   = types.AgencyStatusActive

	CreateWorkItemGoalLinkRequest = types.CreateWorkItemGoalLinkRequest

)	AgencyStatusInactive = types.AgencyStatusInactive	AgencyStatusInactive = types.AgencyStatusInactive



// RACI types	AgencyStatusPaused   = types.AgencyStatusPaused	AgencyStatusPaused   = types.AgencyStatusPaused

type (

	models.RACIRole     = types.models.RACIRole	AgencyStatusArchived = types.AgencyStatusArchived	AgencyStatusArchived = types.AgencyStatusArchived

	RACIActivity = types.RACIActivity

	RACIMatrix   = types.RACIMatrix))

	RACITemplate = types.RACITemplate

	

	CreateRACIMatrixRequest = types.CreateRACIMatrixRequest

	UpdateRACIMatrixRequest = types.UpdateRACIMatrixRequest// Overview types// Overview types

	

	RACIValidationResult  = types.RACIValidationResulttype (type (

	RACIValidationError   = types.RACIValidationError

	RACIValidationWarning = types.RACIValidationWarning	Overview              = types.Overview	Overview              = types.Overview

	RACIValidationSummary = types.RACIValidationSummary

		UpdateOverviewRequest = types.UpdateOverviewRequest	UpdateOverviewRequest = types.UpdateOverviewRequest

	RACIExportFormat = types.RACIExportFormat

	))

	RACIAssignment              = types.RACIAssignment

	CreateRACIAssignmentRequest = types.CreateRACIAssignmentRequest

	

	AgencyRACIAssignments = types.AgencyRACIAssignments// Goal types// Goal types

	RoleAssignment        = types.RoleAssignment

)type (type (



// RACI role constants	Goal              = types.Goal	Goal              = types.Goal

const (

	models.RACIResponsible = types.models.RACIResponsible	CreateGoalRequest = types.CreateGoalRequest	CreateGoalRequest = types.CreateGoalRequest

	RACIAccountable = types.RACIAccountable

	RACIConsulted   = types.RACIConsulted	UpdateGoalRequest = types.UpdateGoalRequest	UpdateGoalRequest = types.UpdateGoalRequest

	RACIInformed    = types.RACIInformed

)	GoalRefineRequest = types.GoalRefineRequest	GoalRefineRequest = types.GoalRefineRequest



// RACI export format constants))

const (

	RACIExportPDF      = types.RACIExportPDF

	RACIExportMarkdown = types.RACIExportMarkdown

	RACIExportJSON     = types.RACIExportJSON// Work item types// Work item types

)

type (type (

	WorkItem                      = types.WorkItem	WorkItem                      = types.WorkItem

	CreateWorkItemRequest         = types.CreateWorkItemRequest	CreateWorkItemRequest         = types.CreateWorkItemRequest

	UpdateWorkItemRequest         = types.UpdateWorkItemRequest	UpdateWorkItemRequest         = types.UpdateWorkItemRequest

	WorkItemRefineRequest         = types.WorkItemRefineRequest	WorkItemRefineRequest         = types.WorkItemRefineRequest

	WorkItemGoalLink              = types.WorkItemGoalLink	WorkItemGoalLink              = types.WorkItemGoalLink

	CreateWorkItemGoalLinkRequest = types.CreateWorkItemGoalLinkRequest	CreateWorkItemGoalLinkRequest = types.CreateWorkItemGoalLinkRequest

))



// RACI types// RACI types

type (type (

	models.RACIRole     = types.models.RACIRole	models.RACIRole     = types.models.RACIRole

	RACIActivity = types.RACIActivity	RACIActivity = types.RACIActivity

	RACIMatrix   = types.RACIMatrix	RACIMatrix   = types.RACIMatrix

	RACITemplate = types.RACITemplate	RACITemplate = types.RACITemplate

		

	CreateRACIMatrixRequest = types.CreateRACIMatrixRequest	CreateRACIMatrixRequest = types.CreateRACIMatrixRequest

	UpdateRACIMatrixRequest = types.UpdateRACIMatrixRequest	UpdateRACIMatrixRequest = types.UpdateRACIMatrixRequest

		

	RACIValidationResult  = types.RACIValidationResult	RACIValidationResult  = types.RACIValidationResult

	RACIValidationError   = types.RACIValidationError	RACIValidationError   = types.RACIValidationError

	RACIValidationWarning = types.RACIValidationWarning	RACIValidationWarning = types.RACIValidationWarning

	RACIValidationSummary = types.RACIValidationSummary	RACIValidationSummary = types.RACIValidationSummary

		

	RACIExportFormat = types.RACIExportFormat	RACIExportFormat = types.RACIExportFormat

		

	RACIAssignment              = types.RACIAssignment	RACIAssignment              = types.RACIAssignment

	CreateRACIAssignmentRequest = types.CreateRACIAssignmentRequest	CreateRACIAssignmentRequest = types.CreateRACIAssignmentRequest

		

	AgencyRACIAssignments = types.AgencyRACIAssignments	AgencyRACIAssignments = types.AgencyRACIAssignments

	RoleAssignment        = types.RoleAssignment	RoleAssignment        = types.RoleAssignment

))



// RACI role constants// RACI role constants

const (const (

	models.RACIResponsible = types.models.RACIResponsible	models.RACIResponsible = types.models.RACIResponsible

	RACIAccountable = types.RACIAccountable	RACIAccountable = types.RACIAccountable

	RACIConsulted   = types.RACIConsulted	RACIConsulted   = types.RACIConsulted

	RACIInformed    = types.RACIInformed	RACIInformed    = types.RACIInformed

))



// RACI export format constants// RACI export format constants

const (const (

	RACIExportPDF      = types.RACIExportPDF	RACIExportPDF      = types.RACIExportPDF

	RACIExportMarkdown = types.RACIExportMarkdown	RACIExportMarkdown = types.RACIExportMarkdown

	RACIExportJSON     = types.RACIExportJSON	RACIExportJSON     = types.RACIExportJSON

))

