package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/ztdp/orchestrator/internal/graph"
	"github.com/ztdp/orchestrator/internal/types"
)

// Service handles workflow orchestration
type Service struct {
	graph    graph.Graph
	registry RegistryService
	logger   graph.Logger
}

// RegistryService interface for agent registry operations
type RegistryService interface {
	GetAgentsByCapability(ctx context.Context, capability string) ([]*types.Agent, error)
	UpdateAgentStatus(ctx context.Context, agentID, status string) error
}

// NewService creates a new orchestrator service
func NewService(g graph.Graph, registry RegistryService, logger graph.Logger) *Service {
	return &Service{
		graph:    g,
		registry: registry,
		logger:   logger,
	}
}

// ExecuteWorkflow executes a workflow
func (s *Service) ExecuteWorkflow(ctx context.Context, request *types.WorkflowRequest) (*types.WorkflowExecution, error) {
	if request == nil {
		return nil, fmt.Errorf("workflow request cannot be nil")
	}

	// Create execution
	execution := &types.WorkflowExecution{
		ID:         fmt.Sprintf("exec-%d", time.Now().UnixNano()),
		WorkflowID: request.ID,
		Status:     types.WorkflowStatusPending,
		StartedAt:  time.Now(),
		Context:    request.Context,
		Steps:      make([]types.StepExecution, 0, len(request.Steps)),
	}

	// Plan workflow first
	plan, err := s.PlanWorkflow(ctx, request)
	if err != nil {
		execution.Status = types.WorkflowStatusFailed
		errMsg := fmt.Sprintf("workflow planning failed: %v", err)
		execution.Error = &errMsg
		return execution, err
	}

	execution.Status = types.WorkflowStatusRunning

	// Execute steps based on plan
	for _, plannedStep := range plan.Steps {
		stepExecution := types.StepExecution{
			StepID:  plannedStep.StepID,
			AgentID: plannedStep.AgentID,
			Status:  types.StepStatusRunning,
		}
		now := time.Now()
		stepExecution.StartedAt = &now

		// Simulate step execution (in real implementation, this would call the agent)
		time.Sleep(10 * time.Millisecond) // Simulate work

		// Mark step as completed
		stepExecution.Status = types.StepStatusCompleted
		completedAt := time.Now()
		stepExecution.CompletedAt = &completedAt
		stepExecution.Output = map[string]interface{}{
			"result": "success",
			"agent":  plannedStep.AgentID,
		}

		execution.Steps = append(execution.Steps, stepExecution)
	}

	execution.Status = types.WorkflowStatusCompleted
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	// Store execution in graph
	err = s.storeExecution(ctx, execution)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to store execution", err, "execution_id", execution.ID)
		}
	}

	if s.logger != nil {
		s.logger.Info("Workflow executed successfully", "execution_id", execution.ID, "workflow_id", request.ID)
	}

	return execution, nil
}

// PlanWorkflow creates an execution plan for a workflow
func (s *Service) PlanWorkflow(ctx context.Context, request *types.WorkflowRequest) (*types.WorkflowPlan, error) {
	if request == nil {
		return nil, fmt.Errorf("workflow request cannot be nil")
	}

	plan := &types.WorkflowPlan{
		ID:         fmt.Sprintf("plan-%d", time.Now().UnixNano()),
		WorkflowID: request.ID,
		Context:    request.Context,
		CreatedAt:  time.Now(),
		Steps:      make([]types.PlannedStep, 0, len(request.Steps)),
	}

	// Plan each step
	for i, step := range request.Steps {
		// Find agents with required capabilities
		var selectedAgent *types.Agent
		for _, capability := range step.Capabilities {
			agents, err := s.registry.GetAgentsByCapability(ctx, capability)
			if err != nil {
				continue
			}

			// Select first available agent
			for _, agent := range agents {
				if agent.Status == types.AgentStatusActive {
					selectedAgent = agent
					break
				}
			}

			if selectedAgent != nil {
				break
			}
		}

		if selectedAgent == nil {
			return nil, fmt.Errorf("no available agent found for step %s with capabilities %v", step.ID, step.Capabilities)
		}

		plannedStep := types.PlannedStep{
			StepID:       step.ID,
			AgentID:      selectedAgent.ID,
			Dependencies: step.Dependencies,
			Order:        i,
		}

		plan.Steps = append(plan.Steps, plannedStep)
	}

	if s.logger != nil {
		s.logger.Info("Workflow planned successfully", "plan_id", plan.ID, "workflow_id", request.ID, "steps", len(plan.Steps))
	}

	return plan, nil
}

// GetWorkflowExecution retrieves a workflow execution
func (s *Service) GetWorkflowExecution(ctx context.Context, executionID string) (*types.WorkflowExecution, error) {
	if executionID == "" {
		return nil, fmt.Errorf("execution ID cannot be empty")
	}

	// Try to get from graph
	nodeData, err := s.graph.GetNode(ctx, "workflow_execution", executionID)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %w", err)
	}

	execution, err := s.nodeToExecution(executionID, nodeData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert node to execution: %w", err)
	}

	return execution, nil
}

// Helper methods

func (s *Service) storeExecution(ctx context.Context, execution *types.WorkflowExecution) error {
	properties := map[string]interface{}{
		"workflow_id": execution.WorkflowID,
		"status":      string(execution.Status),
		"started_at":  execution.StartedAt,
		"context":     execution.Context,
		"created_at":  time.Now(),
	}

	if execution.CompletedAt != nil {
		properties["completed_at"] = *execution.CompletedAt
	}

	if execution.Error != nil {
		properties["error"] = *execution.Error
	}

	return s.graph.AddNode(ctx, "workflow_execution", execution.ID, properties)
}

func (s *Service) nodeToExecution(executionID string, nodeData map[string]interface{}) (*types.WorkflowExecution, error) {
	execution := &types.WorkflowExecution{
		ID: executionID,
	}

	if workflowID, ok := nodeData["workflow_id"].(string); ok {
		execution.WorkflowID = workflowID
	}

	if statusStr, ok := nodeData["status"].(string); ok {
		execution.Status = types.WorkflowStatus(statusStr)
	}

	if startedAt, ok := nodeData["started_at"].(time.Time); ok {
		execution.StartedAt = startedAt
	}

	if completedAt, ok := nodeData["completed_at"].(time.Time); ok {
		execution.CompletedAt = &completedAt
	}

	if errorStr, ok := nodeData["error"].(string); ok {
		execution.Error = &errorStr
	}

	if context, ok := nodeData["context"].(map[string]interface{}); ok {
		execution.Context = context
	}

	return execution, nil
}
