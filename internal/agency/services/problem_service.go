package services

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency"
)

// ProblemService handles problem operations
type ProblemService struct {
	repo agency.Repository
}

// NewProblemService creates a new problem service
func NewProblemService(repo agency.Repository) *ProblemService {
	return &ProblemService{
		repo: repo,
	}
}

// CreateProblem creates a new problem for an agency
func (s *ProblemService) CreateProblem(ctx context.Context, agencyID string, description string) (*agency.Problem, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	problem := &agency.Problem{
		AgencyID:    agencyID,
		Description: description,
	}

	if err := s.repo.CreateProblem(ctx, problem); err != nil {
		return nil, fmt.Errorf("failed to create problem: %w", err)
	}

	return problem, nil
}

// GetProblems retrieves all problems for an agency
func (s *ProblemService) GetProblems(ctx context.Context, agencyID string) ([]*agency.Problem, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	problems, err := s.repo.GetProblems(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get problems: %w", err)
	}

	return problems, nil
}

// UpdateProblem updates a problem's description
func (s *ProblemService) UpdateProblem(ctx context.Context, agencyID string, key string, description string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	// Get the problem
	problem, err := s.repo.GetProblem(ctx, agencyID, key)
	if err != nil {
		return fmt.Errorf("failed to get problem: %w", err)
	}

	// Update description
	problem.Description = description

	// Save
	if err := s.repo.UpdateProblem(ctx, problem); err != nil {
		return fmt.Errorf("failed to update problem: %w", err)
	}

	return nil
}

// DeleteProblem deletes a problem
func (s *ProblemService) DeleteProblem(ctx context.Context, agencyID string, key string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	if err := s.repo.DeleteProblem(ctx, agencyID, key); err != nil {
		return fmt.Errorf("failed to delete problem: %w", err)
	}

	return nil
}
