package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExecutionPlan(t *testing.T) {
	t.Run("should create execution plan with default values", func(t *testing.T) {
		action := "deploy-application"
		parameters := map[string]interface{}{"app": "test-app"}

		plan := NewExecutionPlan(action, parameters)

		assert.NotEmpty(t, plan.ID)
		assert.Equal(t, action, plan.Action)
		assert.Equal(t, parameters, plan.Parameters)
		assert.Equal(t, ExecutionStatusPending, plan.Status)
		assert.Empty(t, plan.Steps)
		assert.False(t, plan.CreatedAt.IsZero())
		assert.Nil(t, plan.StartedAt)
		assert.Nil(t, plan.CompletedAt)
		assert.Nil(t, plan.Error)
	})
}

func TestExecutionPlan_AddStep(t *testing.T) {
	t.Run("should add step to execution plan", func(t *testing.T) {
		plan := NewExecutionPlan("deploy", map[string]interface{}{})

		plan.AddStep("deploy-app", "deploy-agent-001", "deploy",
			map[string]interface{}{"app": "test-app"}, []string{})

		assert.Len(t, plan.Steps, 1)
		step := plan.Steps[0]
		assert.NotEmpty(t, step.ID)
		assert.Equal(t, "deploy-app", step.Name)
		assert.Equal(t, "deploy-agent-001", step.AgentID)
		assert.Equal(t, "deploy", step.Action)
		assert.Equal(t, ExecutionStatusPending, step.Status)
		assert.False(t, step.CreatedAt.IsZero())
	})
}

func TestExecutionPlan_UpdateStatus(t *testing.T) {
	t.Run("should update status to in progress and set started time", func(t *testing.T) {
		plan := NewExecutionPlan("deploy", map[string]interface{}{})

		plan.UpdateStatus(ExecutionStatusInProgress)

		assert.Equal(t, ExecutionStatusInProgress, plan.Status)
		assert.NotNil(t, plan.StartedAt)
		assert.Nil(t, plan.CompletedAt)
	})

	t.Run("should update status to completed and set completed time", func(t *testing.T) {
		plan := NewExecutionPlan("deploy", map[string]interface{}{})

		plan.UpdateStatus(ExecutionStatusCompleted)

		assert.Equal(t, ExecutionStatusCompleted, plan.Status)
		assert.NotNil(t, plan.CompletedAt)
	})
}

func TestExecutionPlan_SetError(t *testing.T) {
	t.Run("should set error and update status to failed", func(t *testing.T) {
		plan := NewExecutionPlan("deploy", map[string]interface{}{})
		errorMsg := "deployment failed"

		plan.SetError(errorMsg)

		assert.Equal(t, ExecutionStatusFailed, plan.Status)
		assert.NotNil(t, plan.Error)
		assert.Equal(t, errorMsg, *plan.Error)
		assert.NotNil(t, plan.CompletedAt)
	})
}

func TestExecutionPlan_IsComplete(t *testing.T) {
	testCases := []struct {
		name     string
		status   ExecutionStatus
		expected bool
	}{
		{"pending should not be complete", ExecutionStatusPending, false},
		{"in progress should not be complete", ExecutionStatusInProgress, false},
		{"completed should be complete", ExecutionStatusCompleted, true},
		{"failed should be complete", ExecutionStatusFailed, true},
		{"cancelled should be complete", ExecutionStatusCancelled, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan := NewExecutionPlan("test", map[string]interface{}{})
			plan.UpdateStatus(tc.status)

			assert.Equal(t, tc.expected, plan.IsComplete())
		})
	}
}
