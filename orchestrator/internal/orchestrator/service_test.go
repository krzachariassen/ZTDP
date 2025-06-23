package orchestrator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztdp/orchestrator/internal/graph"
	"github.com/ztdp/orchestrator/internal/types"
)

// Mock implementations for testing
type MockRegistry struct {
	mock.Mock
}

func (m *MockRegistry) GetAgentsByCapability(ctx context.Context, capability string) ([]*types.Agent, error) {
	args := m.Called(ctx, capability)
	return args.Get(0).([]*types.Agent), args.Error(1)
}

func (m *MockRegistry) UpdateAgentStatus(ctx context.Context, agentID, status string) error {
	args := m.Called(ctx, agentID, status)
	return args.Error(0)
}

type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})             {}
func (m *MockLogger) Error(msg string, err error, fields ...interface{}) {}
func (m *MockLogger) Debug(msg string, fields ...interface{})            {}

func TestService_ExecuteWorkflow(t *testing.T) {
	// Setup
	graphBackend := graph.NewEmbeddedGraph(&MockLogger{})
	mockRegistry := &MockRegistry{}
	service := NewService(graphBackend, mockRegistry, &MockLogger{})
	ctx := context.Background()

	// Mock agents for deploy capability
	deployAgent := &types.Agent{
		ID:           "agent-1",
		Name:         "deploy-agent",
		Status:       types.AgentStatusActive,
		Capabilities: []string{"deploy"},
	}

	mockRegistry.On("GetAgentsByCapability", ctx, "deploy").Return([]*types.Agent{deployAgent}, nil)

	// Test workflow request
	request := &types.WorkflowRequest{
		ID:   "workflow-1",
		Name: "Deploy Application",
		Steps: []types.WorkflowStep{
			{
				ID:           "step-1",
				Name:         "Deploy",
				Action:       "deploy",
				Capabilities: []string{"deploy"},
			},
		},
	}

	// Execute
	execution, err := service.ExecuteWorkflow(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, request.ID, execution.WorkflowID)
	assert.Equal(t, types.WorkflowStatusCompleted, execution.Status)
	assert.Len(t, execution.Steps, 1)
	assert.Equal(t, types.StepStatusCompleted, execution.Steps[0].Status)
	assert.Equal(t, deployAgent.ID, execution.Steps[0].AgentID)

	mockRegistry.AssertExpectations(t)
}

func TestService_PlanWorkflow(t *testing.T) {
	// Setup
	graphBackend := graph.NewEmbeddedGraph(&MockLogger{})
	mockRegistry := &MockRegistry{}
	service := NewService(graphBackend, mockRegistry, &MockLogger{})
	ctx := context.Background()

	tests := []struct {
		name      string
		request   *types.WorkflowRequest
		agents    []*types.Agent
		wantErr   bool
		wantSteps int
	}{
		{
			name: "successful planning with available agents",
			request: &types.WorkflowRequest{
				ID:   "workflow-1",
				Name: "Test Workflow",
				Steps: []types.WorkflowStep{
					{
						ID:           "step-1",
						Name:         "Deploy",
						Capabilities: []string{"deploy"},
					},
				},
			},
			agents: []*types.Agent{
				{
					ID:           "agent-1",
					Name:         "deploy-agent",
					Status:       types.AgentStatusActive,
					Capabilities: []string{"deploy"},
				},
			},
			wantErr:   false,
			wantSteps: 1,
		},
		{
			name: "planning fails when no agents available",
			request: &types.WorkflowRequest{
				ID:   "workflow-2",
				Name: "Test Workflow",
				Steps: []types.WorkflowStep{
					{
						ID:           "step-1",
						Name:         "Deploy",
						Capabilities: []string{"deploy"},
					},
				},
			},
			agents:    []*types.Agent{}, // No agents
			wantErr:   true,
			wantSteps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRegistry.ExpectedCalls = nil // Reset previous calls
			mockRegistry.On("GetAgentsByCapability", ctx, "deploy").Return(tt.agents, nil)

			// Execute
			plan, err := service.PlanWorkflow(ctx, tt.request)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
				assert.Equal(t, tt.request.ID, plan.WorkflowID)
				assert.Len(t, plan.Steps, tt.wantSteps)
			}
		})
	}
}
