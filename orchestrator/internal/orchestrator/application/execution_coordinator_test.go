package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	orchestratorDomain "github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// MockExecutionService for testing
type MockExecutionService struct {
	mock.Mock
}

func (m *MockExecutionService) CreateExecutionPlan(ctx context.Context, plan *orchestratorDomain.ExecutionPlan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}

func (m *MockExecutionService) GetExecutionPlan(ctx context.Context, planID string) (*orchestratorDomain.ExecutionPlan, error) {
	args := m.Called(ctx, planID)
	return args.Get(0).(*orchestratorDomain.ExecutionPlan), args.Error(1)
}

func (m *MockExecutionService) UpdateExecutionStatus(ctx context.Context, planID string, status orchestratorDomain.ExecutionStatus) error {
	args := m.Called(ctx, planID, status)
	return args.Error(0)
}

func TestExecutionCoordinator_CreatePlan(t *testing.T) {
	t.Run("should create execution plan from decision", func(t *testing.T) {
		mockExecutionService := &MockExecutionService{}
		coordinator := NewExecutionCoordinator(mockExecutionService)

		decision := &orchestratorDomain.Decision{
			Type:       orchestratorDomain.DecisionTypeExecute,
			Action:     "deploy-application",
			Parameters: map[string]interface{}{"app": "test-app", "env": "staging"},
			Reasoning:  "User requested deployment",
		}

		expectedPlan := &orchestratorDomain.ExecutionPlan{
			Action:     "deploy-application",
			Parameters: map[string]interface{}{"app": "test-app", "env": "staging"},
			Status:     orchestratorDomain.ExecutionStatusPending,
		}

		mockExecutionService.On("CreateExecutionPlan", mock.Anything, mock.MatchedBy(func(plan *orchestratorDomain.ExecutionPlan) bool {
			return plan.Action == expectedPlan.Action &&
				plan.Status == expectedPlan.Status &&
				plan.Parameters["app"] == "test-app"
		})).Return(nil)

		planID, err := coordinator.CreatePlan(context.Background(), decision)

		assert.NoError(t, err)
		assert.NotEmpty(t, planID)
		mockExecutionService.AssertExpectations(t)
	})

	t.Run("should reject clarification decisions", func(t *testing.T) {
		mockExecutionService := &MockExecutionService{}
		coordinator := NewExecutionCoordinator(mockExecutionService)

		decision := &orchestratorDomain.Decision{
			Type:      orchestratorDomain.DecisionTypeClarify,
			Action:    "clarify-requirements",
			Reasoning: "Need more information",
		}

		planID, err := coordinator.CreatePlan(context.Background(), decision)

		assert.Error(t, err)
		assert.Empty(t, planID)
		assert.Contains(t, err.Error(), "cannot create execution plan for clarification decision")
		mockExecutionService.AssertNotCalled(t, "CreateExecutionPlan")
	})
}

func TestExecutionCoordinator_GetPlanStatus(t *testing.T) {
	t.Run("should retrieve execution plan status", func(t *testing.T) {
		mockExecutionService := &MockExecutionService{}
		coordinator := NewExecutionCoordinator(mockExecutionService)

		planID := "plan-123"
		expectedPlan := &orchestratorDomain.ExecutionPlan{
			ID:     planID,
			Action: "deploy-application",
			Status: orchestratorDomain.ExecutionStatusInProgress,
		}

		mockExecutionService.On("GetExecutionPlan", mock.Anything, planID).Return(expectedPlan, nil)

		plan, err := coordinator.GetPlanStatus(context.Background(), planID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPlan, plan)
		mockExecutionService.AssertExpectations(t)
	})
}

func TestExecutionCoordinator_UpdateStatus(t *testing.T) {
	t.Run("should update execution plan status", func(t *testing.T) {
		mockExecutionService := &MockExecutionService{}
		coordinator := NewExecutionCoordinator(mockExecutionService)

		planID := "plan-123"
		newStatus := orchestratorDomain.ExecutionStatusCompleted

		mockExecutionService.On("UpdateExecutionStatus", mock.Anything, planID, newStatus).Return(nil)

		err := coordinator.UpdateStatus(context.Background(), planID, newStatus)

		assert.NoError(t, err)
		mockExecutionService.AssertExpectations(t)
	})
}
