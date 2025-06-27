package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecision_Creation(t *testing.T) {
	t.Run("should create clarify decision", func(t *testing.T) {
		decision := NewClarifyDecision("What environment would you like to deploy to?", "Need more specifics about deployment target")

		assert.Equal(t, DecisionTypeClarify, decision.Type)
		assert.Equal(t, "What environment would you like to deploy to?", decision.ClarificationQuestion)
		assert.Equal(t, "Need more specifics about deployment target", decision.Reasoning)
		assert.Empty(t, decision.ExecutionPlan)
		assert.Empty(t, decision.AgentCoordination)
	})

	t.Run("should create execute decision", func(t *testing.T) {
		executionPlan := "Step 1: Deploy to staging\nStep 2: Run tests"
		coordination := "Primary Agent: deploy-agent"

		decision := NewExecuteDecision(executionPlan, coordination, "Clear deployment request with sufficient details")

		assert.Equal(t, DecisionTypeExecute, decision.Type)
		assert.Equal(t, executionPlan, decision.ExecutionPlan)
		assert.Equal(t, coordination, decision.AgentCoordination)
		assert.Equal(t, "Clear deployment request with sufficient details", decision.Reasoning)
		assert.Empty(t, decision.ClarificationQuestion)
	})

	t.Run("should create execute decision with action and parameters", func(t *testing.T) {
		action := "deploy-application"
		parameters := map[string]interface{}{"app": "test-app", "env": "staging"}

		decision := NewExecuteDecisionWithAction(action, parameters, "Clear deployment request")

		assert.Equal(t, DecisionTypeExecute, decision.Type)
		assert.Equal(t, action, decision.Action)
		assert.Equal(t, parameters, decision.Parameters)
		assert.Equal(t, "Clear deployment request", decision.Reasoning)
		assert.Empty(t, decision.ClarificationQuestion)
		assert.True(t, decision.HasAction())
	})
}

func TestDecision_IsExecutable(t *testing.T) {
	t.Run("should return true for execute decision", func(t *testing.T) {
		decision := NewExecuteDecision("test plan", "test coordination", "test reasoning")
		assert.True(t, decision.IsExecutable())
	})

	t.Run("should return false for clarify decision", func(t *testing.T) {
		decision := NewClarifyDecision("test question", "test reasoning")
		assert.False(t, decision.IsExecutable())
	})
	t.Run("should create execute decision with action and parameters", func(t *testing.T) {
		action := "deploy-application"
		parameters := map[string]interface{}{"app": "test-app", "env": "staging"}

		decision := NewExecuteDecisionWithAction(action, parameters, "Clear deployment request")

		assert.Equal(t, DecisionTypeExecute, decision.Type)
		assert.Equal(t, action, decision.Action)
		assert.Equal(t, parameters, decision.Parameters)
		assert.Equal(t, "Clear deployment request", decision.Reasoning)
		assert.Empty(t, decision.ClarificationQuestion)
		assert.True(t, decision.HasAction())
	})
}

func TestDecision_NeedsClarification(t *testing.T) {
	t.Run("should return true for clarify decision", func(t *testing.T) {
		decision := NewClarifyDecision("test question", "test reasoning")
		assert.True(t, decision.NeedsClarification())
	})

	t.Run("should return false for execute decision", func(t *testing.T) {
		decision := NewExecuteDecision("test plan", "test coordination", "test reasoning")
		assert.False(t, decision.NeedsClarification())
	})
}

func TestDecision_HasAction(t *testing.T) {
	t.Run("should return true for decision with action", func(t *testing.T) {
		decision := NewExecuteDecisionWithAction("deploy", map[string]interface{}{}, "test")
		assert.True(t, decision.HasAction())
	})

	t.Run("should return false for decision without action", func(t *testing.T) {
		decision := NewExecuteDecision("test plan", "test coordination", "test reasoning")
		assert.False(t, decision.HasAction())
	})

	t.Run("should return false for clarify decision", func(t *testing.T) {
		decision := NewClarifyDecision("test question", "test reasoning")
		assert.False(t, decision.HasAction())
	})
}
