package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ztdp/orchestrator/internal/planning/domain"
)

// PlanningServiceImpl implements the PlanningService interface
type PlanningServiceImpl struct {
	repo PlanningRepository
}

// NewPlanningServiceImpl creates a new planning service implementation
func NewPlanningServiceImpl(repo PlanningRepository) PlanningService {
	return &PlanningServiceImpl{
		repo: repo,
	}
}

// CreateExecutionPlan creates a new execution plan based on analysis
func (s *PlanningServiceImpl) CreateExecutionPlan(ctx context.Context, request *PlanningRequest) (*domain.ExecutionPlan, error) {
	if request == nil {
		return nil, errors.New("planning request cannot be nil")
	}

	if request.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if request.UserRequest == "" {
		return nil, errors.New("user request cannot be empty")
	}

	// Generate a unique ID for the execution plan
	planID := uuid.New().String()

	// Create the execution plan using domain logic
	plan, err := domain.NewExecutionPlan(
		planID,
		request.ConversationID,
		request.UserID,
		request.UserRequest,
		request.Intent,
		request.Category,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution plan: %w", err)
	}

	// Save the plan
	err = s.repo.Save(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to save execution plan: %w", err)
	}

	return plan, nil
}

// GetExecutionPlan retrieves an execution plan by ID
func (s *PlanningServiceImpl) GetExecutionPlan(ctx context.Context, planID string) (*domain.ExecutionPlan, error) {
	if planID == "" {
		return nil, errors.New("plan ID cannot be empty")
	}

	plan, err := s.repo.GetByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution plan: %w", err)
	}

	return plan, nil
}

// GetUserExecutionPlans retrieves all execution plans for a user
func (s *PlanningServiceImpl) GetUserExecutionPlans(ctx context.Context, userID string) ([]*domain.ExecutionPlan, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	plans, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user execution plans: %w", err)
	}

	return plans, nil
}

// ValidatePlan validates an execution plan for correctness and feasibility
func (s *PlanningServiceImpl) ValidatePlan(ctx context.Context, plan *domain.ExecutionPlan) error {
	if plan == nil {
		return errors.New("execution plan cannot be nil")
	}

	// Use domain validation
	return plan.Validate()
}

// UpdatePlanStatus updates the status of an execution plan
func (s *PlanningServiceImpl) UpdatePlanStatus(ctx context.Context, planID string, status domain.ExecutionPlanStatus) error {
	if planID == "" {
		return errors.New("plan ID cannot be empty")
	}

	// Get the plan
	plan, err := s.repo.GetByID(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to get execution plan: %w", err)
	}

	// Update status using domain logic
	switch status {
	case domain.ExecutionPlanStatusRunning:
		err = plan.Start()
	case domain.ExecutionPlanStatusCompleted:
		err = plan.Complete("Manual completion")
	case domain.ExecutionPlanStatusFailed:
		err = plan.Fail("Manual status update")
	case domain.ExecutionPlanStatusCancelled:
		err = plan.Cancel()
	default:
		return fmt.Errorf("unsupported status transition to %s", status)
	}

	if err != nil {
		return fmt.Errorf("failed to update plan status: %w", err)
	}

	// Save the updated plan
	err = s.repo.Save(ctx, plan)
	if err != nil {
		return fmt.Errorf("failed to save plan after status update: %w", err)
	}

	return nil
}

// AddStepToPlan adds a new step to an existing execution plan
func (s *PlanningServiceImpl) AddStepToPlan(ctx context.Context, planID string, step *domain.ExecutionStep) error {
	if planID == "" {
		return errors.New("plan ID cannot be empty")
	}

	if step == nil {
		return errors.New("execution step cannot be nil")
	}

	// Get the plan
	plan, err := s.repo.GetByID(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to get execution plan: %w", err)
	}

	// Add step using domain logic
	err = plan.AddStep(*step)
	if err != nil {
		return fmt.Errorf("failed to add step to plan: %w", err)
	}

	// Save the updated plan
	err = s.repo.Save(ctx, plan)
	if err != nil {
		return fmt.Errorf("failed to save plan after adding step: %w", err)
	}

	return nil
}
