package services

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency"
)

// RACIService handles RACI assignment business logic
type RACIService struct {
	repo agency.Repository
}

// NewRACIService creates a new RACI service
func NewRACIService(repo agency.Repository) *RACIService {
	return &RACIService{
		repo: repo,
	}
}

// CreateRACIAssignment creates a new RACI assignment edge
func (s *RACIService) CreateRACIAssignment(ctx context.Context, agencyID string, assignment *agency.RACIAssignment) error {
	return s.repo.CreateRACIAssignment(ctx, agencyID, assignment)
}

// GetRACIAssignmentsForWorkItem retrieves all RACI assignments for a work item
func (s *RACIService) GetRACIAssignmentsForWorkItem(ctx context.Context, agencyID string, workItemKey string) ([]*agency.RACIAssignment, error) {
	return s.repo.GetRACIAssignmentsForWorkItem(ctx, agencyID, workItemKey)
}

// GetRACIAssignmentsForRole retrieves all RACI assignments for a role
func (s *RACIService) GetRACIAssignmentsForRole(ctx context.Context, agencyID string, roleID string) ([]*agency.RACIAssignment, error) {
	return s.repo.GetRACIAssignmentsForRole(ctx, agencyID, roleID)
}

// GetAllRACIAssignments retrieves all RACI assignments for an agency
func (s *RACIService) GetAllRACIAssignments(ctx context.Context, agencyID string) ([]*agency.RACIAssignment, error) {
	return s.repo.GetAllRACIAssignments(ctx, agencyID)
}

// UpdateRACIAssignment updates an existing RACI assignment
func (s *RACIService) UpdateRACIAssignment(ctx context.Context, agencyID string, key string, assignment *agency.RACIAssignment) error {
	return s.repo.UpdateRACIAssignment(ctx, agencyID, key, assignment)
}

// DeleteRACIAssignment deletes a RACI assignment by key
func (s *RACIService) DeleteRACIAssignment(ctx context.Context, agencyID string, key string) error {
	return s.repo.DeleteRACIAssignment(ctx, agencyID, key)
}

// DeleteRACIAssignmentsForWorkItem deletes all RACI assignments for a work item
func (s *RACIService) DeleteRACIAssignmentsForWorkItem(ctx context.Context, agencyID string, workItemKey string) error {
	return s.repo.DeleteRACIAssignmentsForWorkItem(ctx, agencyID, workItemKey)
}
