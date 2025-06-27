package domain

import (
	"testing"
)

func TestNewExecutionPlan(t *testing.T) {
	t.Run("should create valid execution plan", func(t *testing.T) {
		// Given
		id := "plan-123"
		conversationID := "conv-456"
		userID := "user-789"
		userRequest := "Deploy my application"
		intent := "deployment"
		category := "deployment"

		// When
		plan, err := NewExecutionPlan(id, conversationID, userID, userRequest, intent, category)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if plan.ID != id {
			t.Errorf("Expected ID %s, got %s", id, plan.ID)
		}

		if plan.ConversationID != conversationID {
			t.Errorf("Expected ConversationID %s, got %s", conversationID, plan.ConversationID)
		}

		if plan.UserID != userID {
			t.Errorf("Expected UserID %s, got %s", userID, plan.UserID)
		}

		if plan.UserRequest != userRequest {
			t.Errorf("Expected UserRequest %s, got %s", userRequest, plan.UserRequest)
		}

		if plan.Status != ExecutionPlanStatusPending {
			t.Errorf("Expected Status %s, got %s", ExecutionPlanStatusPending, plan.Status)
		}

		if len(plan.Steps) != 0 {
			t.Errorf("Expected empty steps, got %d steps", len(plan.Steps))
		}

		if plan.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
	})

	t.Run("should fail with empty ID", func(t *testing.T) {
		// When
		_, err := NewExecutionPlan("", "conv-456", "user-789", "Deploy app", "deployment", "deployment")

		// Then
		if err == nil {
			t.Fatal("Expected validation error for empty ID")
		}

		validationErr, ok := err.(ExecutionPlanValidationError)
		if !ok {
			t.Errorf("Expected ExecutionPlanValidationError, got %T", err)
		}

		if validationErr.Field != "id" {
			t.Errorf("Expected field 'id', got '%s'", validationErr.Field)
		}
	})
}

func TestExecutionPlan_AddStep(t *testing.T) {
	t.Run("should add valid step", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		step := ExecutionStep{
			ID:          "step-1",
			Name:        "Build Application",
			Description: "Compile and build the application",
			AgentID:     "agent-1",
			AgentType:   "build-agent",
			Status:      ExecutionPlanStatusPending,
		}

		// When
		err := plan.AddStep(step)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(plan.Steps) != 1 {
			t.Errorf("Expected 1 step, got %d", len(plan.Steps))
		}

		addedStep := plan.Steps[0]
		if addedStep.ID != step.ID {
			t.Errorf("Expected step ID %s, got %s", step.ID, addedStep.ID)
		}
	})

	t.Run("should fail with duplicate step ID", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		step1 := ExecutionStep{
			ID:        "step-1",
			Name:      "Build Application",
			AgentID:   "agent-1",
			AgentType: "build-agent",
			Status:    ExecutionPlanStatusPending,
		}
		step2 := ExecutionStep{
			ID:        "step-1", // Same ID
			Name:      "Deploy Application",
			AgentID:   "agent-2",
			AgentType: "deploy-agent",
			Status:    ExecutionPlanStatusPending,
		}

		plan.AddStep(step1)

		// When
		err := plan.AddStep(step2)

		// Then
		if err == nil {
			t.Fatal("Expected error for duplicate step ID")
		}

		validationErr, ok := err.(ExecutionPlanValidationError)
		if !ok {
			t.Errorf("Expected ExecutionPlanValidationError, got %T", err)
		}

		if validationErr.Field != "step.id" {
			t.Errorf("Expected field 'step.id', got '%s'", validationErr.Field)
		}
	})
}

func TestExecutionPlan_Start(t *testing.T) {
	t.Run("should start pending plan", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")

		// When
		err := plan.Start()

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if plan.Status != ExecutionPlanStatusRunning {
			t.Errorf("Expected status %s, got %s", ExecutionPlanStatusRunning, plan.Status)
		}

		if plan.StartedAt == nil {
			t.Error("Expected StartedAt to be set")
		}
	})

	t.Run("should fail to start non-pending plan", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		plan.Status = ExecutionPlanStatusRunning

		// When
		err := plan.Start()

		// Then
		if err == nil {
			t.Fatal("Expected error when starting non-pending plan")
		}
	})
}

func TestExecutionPlan_Complete(t *testing.T) {
	t.Run("should complete running plan", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		plan.Start()
		result := "Deployment successful"

		// When
		err := plan.Complete(result)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if plan.Status != ExecutionPlanStatusCompleted {
			t.Errorf("Expected status %s, got %s", ExecutionPlanStatusCompleted, plan.Status)
		}

		if plan.Result != result {
			t.Errorf("Expected result %s, got %s", result, plan.Result)
		}

		if plan.CompletedAt == nil {
			t.Error("Expected CompletedAt to be set")
		}

		if plan.ActualTime == 0 {
			t.Error("Expected ActualTime to be calculated")
		}
	})

	t.Run("should fail to complete non-running plan", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")

		// When
		err := plan.Complete("result")

		// Then
		if err == nil {
			t.Fatal("Expected error when completing non-running plan")
		}
	})
}

func TestExecutionPlan_Fail(t *testing.T) {
	t.Run("should fail running plan", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		plan.Start()
		errorMsg := "Deployment failed due to network error"

		// When
		err := plan.Fail(errorMsg)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if plan.Status != ExecutionPlanStatusFailed {
			t.Errorf("Expected status %s, got %s", ExecutionPlanStatusFailed, plan.Status)
		}

		if plan.Error != errorMsg {
			t.Errorf("Expected error %s, got %s", errorMsg, plan.Error)
		}

		if plan.CompletedAt == nil {
			t.Error("Expected CompletedAt to be set")
		}
	})

	t.Run("should fail to fail completed plan", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		plan.Start()
		plan.Complete("success")

		// When
		err := plan.Fail("error")

		// Then
		if err == nil {
			t.Fatal("Expected error when failing completed plan")
		}
	})
}

func TestExecutionPlan_GetRunnableSteps(t *testing.T) {
	t.Run("should return steps with no dependencies", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")

		step1 := ExecutionStep{
			ID:        "step-1",
			Name:      "Build",
			AgentID:   "agent-1",
			AgentType: "build-agent",
			Status:    ExecutionPlanStatusPending,
		}

		step2 := ExecutionStep{
			ID:           "step-2",
			Name:         "Deploy",
			AgentID:      "agent-2",
			AgentType:    "deploy-agent",
			Status:       ExecutionPlanStatusPending,
			Dependencies: []string{"step-1"},
		}

		plan.AddStep(step1)
		plan.AddStep(step2)

		// When
		runnable := plan.GetRunnableSteps()

		// Then
		if len(runnable) != 1 {
			t.Errorf("Expected 1 runnable step, got %d", len(runnable))
		}

		if runnable[0].ID != "step-1" {
			t.Errorf("Expected step-1 to be runnable, got %s", runnable[0].ID)
		}
	})

	t.Run("should return dependent steps after dependencies complete", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")

		step1 := ExecutionStep{
			ID:        "step-1",
			Name:      "Build",
			AgentID:   "agent-1",
			AgentType: "build-agent",
			Status:    ExecutionPlanStatusCompleted, // Completed
		}

		step2 := ExecutionStep{
			ID:           "step-2",
			Name:         "Deploy",
			AgentID:      "agent-2",
			AgentType:    "deploy-agent",
			Status:       ExecutionPlanStatusPending,
			Dependencies: []string{"step-1"},
		}

		plan.AddStep(step1)
		plan.AddStep(step2)

		// When
		runnable := plan.GetRunnableSteps()

		// Then
		if len(runnable) != 1 {
			t.Errorf("Expected 1 runnable step, got %d", len(runnable))
		}

		if runnable[0].ID != "step-2" {
			t.Errorf("Expected step-2 to be runnable, got %s", runnable[0].ID)
		}
	})
}

func TestExecutionPlan_UpdateStepStatus(t *testing.T) {
	t.Run("should update step status", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		step := ExecutionStep{
			ID:        "step-1",
			Name:      "Build",
			AgentID:   "agent-1",
			AgentType: "build-agent",
			Status:    ExecutionPlanStatusPending,
		}
		plan.AddStep(step)

		result := map[string]interface{}{"build_id": "build-123"}

		// When
		err := plan.UpdateStepStatus("step-1", ExecutionPlanStatusCompleted, result, "")

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		updatedStep := plan.Steps[0]
		if updatedStep.Status != ExecutionPlanStatusCompleted {
			t.Errorf("Expected status %s, got %s", ExecutionPlanStatusCompleted, updatedStep.Status)
		}

		if updatedStep.Result == nil {
			t.Error("Expected result to be set")
		}

		if updatedStep.CompletedAt == nil {
			t.Error("Expected CompletedAt to be set")
		}
	})

	t.Run("should fail with non-existent step ID", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")

		// When
		err := plan.UpdateStepStatus("non-existent", ExecutionPlanStatusCompleted, nil, "")

		// Then
		if err == nil {
			t.Fatal("Expected error for non-existent step ID")
		}
	})
}

func TestExecutionPlan_IsCompleted(t *testing.T) {
	t.Run("should return true when all steps completed", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		step := ExecutionStep{
			ID:        "step-1",
			Name:      "Build",
			AgentID:   "agent-1",
			AgentType: "build-agent",
			Status:    ExecutionPlanStatusCompleted,
		}
		plan.AddStep(step)

		// When
		completed := plan.IsCompleted()

		// Then
		if !completed {
			t.Error("Expected plan to be completed")
		}
	})

	t.Run("should return false when steps are pending", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		step := ExecutionStep{
			ID:        "step-1",
			Name:      "Build",
			AgentID:   "agent-1",
			AgentType: "build-agent",
			Status:    ExecutionPlanStatusPending,
		}
		plan.AddStep(step)

		// When
		completed := plan.IsCompleted()

		// Then
		if completed {
			t.Error("Expected plan to not be completed")
		}
	})

	t.Run("should return false when no steps", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")

		// When
		completed := plan.IsCompleted()

		// Then
		if completed {
			t.Error("Expected plan with no steps to not be completed")
		}
	})
}

func TestExecutionPlan_HasFailed(t *testing.T) {
	t.Run("should return true when any step failed", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		step1 := ExecutionStep{
			ID:        "step-1",
			Name:      "Build",
			AgentID:   "agent-1",
			AgentType: "build-agent",
			Status:    ExecutionPlanStatusCompleted,
		}
		step2 := ExecutionStep{
			ID:        "step-2",
			Name:      "Deploy",
			AgentID:   "agent-2",
			AgentType: "deploy-agent",
			Status:    ExecutionPlanStatusFailed,
		}
		plan.AddStep(step1)
		plan.AddStep(step2)

		// When
		failed := plan.HasFailed()

		// Then
		if !failed {
			t.Error("Expected plan to have failed")
		}
	})

	t.Run("should return false when no steps failed", func(t *testing.T) {
		// Given
		plan, _ := NewExecutionPlan("plan-123", "conv-456", "user-789", "Deploy app", "deployment", "deployment")
		step := ExecutionStep{
			ID:        "step-1",
			Name:      "Build",
			AgentID:   "agent-1",
			AgentType: "build-agent",
			Status:    ExecutionPlanStatusCompleted,
		}
		plan.AddStep(step)

		// When
		failed := plan.HasFailed()

		// Then
		if failed {
			t.Error("Expected plan to not have failed")
		}
	})
}
