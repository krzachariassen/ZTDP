package application

import (
	"context"

	"github.com/ztdp/orchestrator/internal/planning/domain"
)

// PlanningService defines the application service interface for execution planning
type PlanningService interface {
	// CreateExecutionPlan creates a new execution plan based on analysis
	CreateExecutionPlan(ctx context.Context, request *PlanningRequest) (*domain.ExecutionPlan, error)

	// GetExecutionPlan retrieves an execution plan by ID
	GetExecutionPlan(ctx context.Context, planID string) (*domain.ExecutionPlan, error)

	// GetUserExecutionPlans retrieves all execution plans for a user
	GetUserExecutionPlans(ctx context.Context, userID string) ([]*domain.ExecutionPlan, error)

	// ValidatePlan validates an execution plan for correctness and feasibility
	ValidatePlan(ctx context.Context, plan *domain.ExecutionPlan) error

	// UpdatePlanStatus updates the status of an execution plan
	UpdatePlanStatus(ctx context.Context, planID string, status domain.ExecutionPlanStatus) error

	// AddStepToPlan adds a new step to an existing execution plan
	AddStepToPlan(ctx context.Context, planID string, step *domain.ExecutionStep) error
}

// PlanningRepository defines the repository interface for execution plan persistence
type PlanningRepository interface {
	// Save stores or updates an execution plan
	Save(ctx context.Context, plan *domain.ExecutionPlan) error

	// GetByID retrieves an execution plan by ID
	GetByID(ctx context.Context, planID string) (*domain.ExecutionPlan, error)

	// GetByUserID retrieves all execution plans for a user
	GetByUserID(ctx context.Context, userID string) ([]*domain.ExecutionPlan, error)

	// GetByStatus retrieves execution plans by status
	GetByStatus(ctx context.Context, status domain.ExecutionPlanStatus) ([]*domain.ExecutionPlan, error)

	// Delete removes an execution plan
	Delete(ctx context.Context, planID string) error
}

// PlanningRequest represents a request to create an execution plan
type PlanningRequest struct {
	ConversationID string                 `json:"conversation_id"`
	UserID         string                 `json:"user_id"`
	UserRequest    string                 `json:"user_request"`
	Intent         string                 `json:"intent"`
	Category       string                 `json:"category"`
	RequiredAgents []string               `json:"required_agents"`
	Parameters     map[string]interface{} `json:"parameters"`
}

// PlanValidationResult represents the result of plan validation
type PlanValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}
