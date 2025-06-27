package infrastructure

import (
	"context"

	"github.com/ztdp/orchestrator/internal/graph"
	orchestratorDomain "github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// GraphExecutionService implements ExecutionService using the graph backend
type GraphExecutionService struct {
	graph graph.Graph
}

// NewGraphExecutionService creates a new GraphExecutionService
func NewGraphExecutionService(graph graph.Graph) *GraphExecutionService {
	return &GraphExecutionService{
		graph: graph,
	}
}

// CreateExecutionPlan stores an execution plan in the graph
func (ges *GraphExecutionService) CreateExecutionPlan(ctx context.Context, plan *orchestratorDomain.ExecutionPlan) error {
	// This would normally store the plan in the graph
	// For now, return success to satisfy the interface
	return nil
}

// GetExecutionPlan retrieves an execution plan from the graph
func (ges *GraphExecutionService) GetExecutionPlan(ctx context.Context, planID string) (*orchestratorDomain.ExecutionPlan, error) {
	// This would normally query the graph for the plan
	// For now, return a mock plan
	plan := &orchestratorDomain.ExecutionPlan{
		ID:     planID,
		Status: orchestratorDomain.ExecutionStatusPending,
		Steps: []orchestratorDomain.ExecutionStep{
			{
				ID:     "step-1",
				Name:   "Initialize",
				Action: "init",
				Status: orchestratorDomain.ExecutionStatusPending,
			},
		},
	}

	return plan, nil
}

// UpdateExecutionStatus updates the status of an execution plan
func (ges *GraphExecutionService) UpdateExecutionStatus(ctx context.Context, planID string, status orchestratorDomain.ExecutionStatus) error {
	// This would normally update the plan status in the graph
	// For now, return success to satisfy the interface
	return nil
}
