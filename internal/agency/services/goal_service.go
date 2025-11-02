package services

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency"
)

// GoalService handles goal operations
type GoalService struct {
	repo agency.Repository
}

// NewGoalService creates a new goal service
func NewGoalService(repo agency.Repository) *GoalService {
	return &GoalService{
		repo: repo,
	}
}

// CreateGoal creates a new goal for an agency
func (s *GoalService) CreateGoal(ctx context.Context, agencyID string, code string, description string) (*agency.Goal, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	goal := &agency.Goal{
		AgencyID:    agencyID,
		Code:        code,
		Description: description,
	}

	if err := s.repo.CreateGoal(ctx, goal); err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	return goal, nil
}

// GetGoals retrieves all goals for an agency
func (s *GoalService) GetGoals(ctx context.Context, agencyID string) ([]*agency.Goal, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	goals, err := s.repo.GetGoals(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goals: %w", err)
	}

	return goals, nil
}

// GetGoal retrieves a single goal by key
func (s *GoalService) GetGoal(ctx context.Context, agencyID string, key string) (*agency.Goal, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	goal, err := s.repo.GetGoal(ctx, agencyID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	return goal, nil
}

// UpdateGoal updates a goal's code and description
func (s *GoalService) UpdateGoal(ctx context.Context, agencyID string, key string, code string, description string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	// Get the goal
	goal, err := s.repo.GetGoal(ctx, agencyID, key)
	if err != nil {
		return fmt.Errorf("failed to get goal: %w", err)
	}

	// Update code and description
	goal.Code = code
	goal.Description = description

	// Save
	if err := s.repo.UpdateGoal(ctx, goal); err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}

	return nil
}

// DeleteGoal deletes a goal
func (s *GoalService) DeleteGoal(ctx context.Context, agencyID string, key string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	if err := s.repo.DeleteGoal(ctx, agencyID, key); err != nil {
		return fmt.Errorf("failed to delete goal: %w", err)
	}

	return nil
}
