package application

import (
	"context"
	"fmt"

	orchestratorDomain "github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// ExecutionService defines the interface for execution operations
type ExecutionService interface {
	CreateExecutionPlan(ctx context.Context, plan *orchestratorDomain.ExecutionPlan) error
	GetExecutionPlan(ctx context.Context, planID string) (*orchestratorDomain.ExecutionPlan, error)
	UpdateExecutionStatus(ctx context.Context, planID string, status orchestratorDomain.ExecutionStatus) error
}

// ExecutionCoordinator coordinates the execution of plans across agents
// Replaces the storeExecutionPlan() functionality from the old orchestrator
type ExecutionCoordinator struct {
	executionService ExecutionService
}

// NewExecutionCoordinator creates a new ExecutionCoordinator instance
func NewExecutionCoordinator(executionService ExecutionService) *ExecutionCoordinator {
	return &ExecutionCoordinator{
		executionService: executionService,
	}
}

// CreatePlan creates an execution plan from a decision
func (ec *ExecutionCoordinator) CreatePlan(ctx context.Context, decision *orchestratorDomain.Decision) (string, error) {
	if !decision.IsExecutable() {
		return "", fmt.Errorf("cannot create execution plan for clarification decision")
	}

	if !decision.HasAction() {
		return "", fmt.Errorf("decision must have action and parameters to create execution plan")
	}

	// Create execution plan from decision
	plan := orchestratorDomain.NewExecutionPlan(decision.Action, decision.Parameters)

	// Store the execution plan
	err := ec.executionService.CreateExecutionPlan(ctx, plan)
	if err != nil {
		return "", fmt.Errorf("failed to create execution plan: %w", err)
	}

	return plan.ID, nil
}

// GetPlanStatus retrieves the current status of an execution plan
func (ec *ExecutionCoordinator) GetPlanStatus(ctx context.Context, planID string) (*orchestratorDomain.ExecutionPlan, error) {
	plan, err := ec.executionService.GetExecutionPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution plan: %w", err)
	}

	return plan, nil
}

// UpdateStatus updates the status of an execution plan
func (ec *ExecutionCoordinator) UpdateStatus(ctx context.Context, planID string, status orchestratorDomain.ExecutionStatus) error {
	err := ec.executionService.UpdateExecutionStatus(ctx, planID, status)
	if err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	return nil
}

// ExecutePlan coordinates the execution of a plan with agents
func (ec *ExecutionCoordinator) ExecutePlan(ctx context.Context, planID string) error {
	plan, err := ec.GetPlanStatus(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to get plan for execution: %w", err)
	}

	if plan.Status != orchestratorDomain.ExecutionStatusPending {
		return fmt.Errorf("plan %s is not in pending status, current status: %s", planID, plan.Status)
	}

	// Update status to in progress
	err = ec.UpdateStatus(ctx, planID, orchestratorDomain.ExecutionStatusInProgress)
	if err != nil {
		return fmt.Errorf("failed to update plan status to in progress: %w", err)
	}

	// TODO: Implement actual agent coordination and step execution
	// This will involve:
	// 1. Analyzing the plan steps
	// 2. Coordinating with available agents
	// 3. Executing steps in dependency order
	// 4. Updating status as execution progresses

	return nil
}
