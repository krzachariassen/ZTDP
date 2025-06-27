package domain

import (
	"errors"
	"fmt"
	"time"
)

// ExecutionPlanStatus represents the current status of an execution plan
type ExecutionPlanStatus string

const (
	ExecutionPlanStatusPending   ExecutionPlanStatus = "pending"
	ExecutionPlanStatusRunning   ExecutionPlanStatus = "running"
	ExecutionPlanStatusCompleted ExecutionPlanStatus = "completed"
	ExecutionPlanStatusFailed    ExecutionPlanStatus = "failed"
	ExecutionPlanStatusCancelled ExecutionPlanStatus = "cancelled"
)

// ExecutionStep represents a single step in an execution plan
type ExecutionStep struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	AgentID      string                 `json:"agent_id"`
	AgentType    string                 `json:"agent_type"`
	Parameters   map[string]interface{} `json:"parameters"`
	Status       ExecutionPlanStatus    `json:"status"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Result       map[string]interface{} `json:"result,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
}

// ExecutionPlan represents a plan for executing user requests through agents
type ExecutionPlan struct {
	ID             string              `json:"id"`
	ConversationID string              `json:"conversation_id"`
	UserID         string              `json:"user_id"`
	UserRequest    string              `json:"user_request"`
	Intent         string              `json:"intent"`
	Category       string              `json:"category"`
	Status         ExecutionPlanStatus `json:"status"`
	Steps          []ExecutionStep     `json:"steps"`
	CreatedAt      time.Time           `json:"created_at"`
	StartedAt      *time.Time          `json:"started_at,omitempty"`
	CompletedAt    *time.Time          `json:"completed_at,omitempty"`
	Error          string              `json:"error,omitempty"`
	Result         string              `json:"result,omitempty"`
	EstimatedTime  time.Duration       `json:"estimated_time,omitempty"`
	ActualTime     time.Duration       `json:"actual_time,omitempty"`
}

// ExecutionPlanValidationError represents validation errors for execution plans
type ExecutionPlanValidationError struct {
	Field   string
	Message string
}

func (e ExecutionPlanValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewExecutionPlan creates a new execution plan with validation
func NewExecutionPlan(id, conversationID, userID, userRequest, intent, category string) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		ID:             id,
		ConversationID: conversationID,
		UserID:         userID,
		UserRequest:    userRequest,
		Intent:         intent,
		Category:       category,
		Status:         ExecutionPlanStatusPending,
		Steps:          make([]ExecutionStep, 0),
		CreatedAt:      time.Now().UTC(),
	}

	if err := plan.Validate(); err != nil {
		return nil, err
	}

	return plan, nil
}

// Validate validates the execution plan
func (ep *ExecutionPlan) Validate() error {
	if ep.ID == "" {
		return ExecutionPlanValidationError{Field: "id", Message: "ID cannot be empty"}
	}

	if ep.ConversationID == "" {
		return ExecutionPlanValidationError{Field: "conversation_id", Message: "conversation ID cannot be empty"}
	}

	if ep.UserID == "" {
		return ExecutionPlanValidationError{Field: "user_id", Message: "user ID cannot be empty"}
	}

	if ep.UserRequest == "" {
		return ExecutionPlanValidationError{Field: "user_request", Message: "user request cannot be empty"}
	}

	if ep.Intent == "" {
		return ExecutionPlanValidationError{Field: "intent", Message: "intent cannot be empty"}
	}

	if ep.Category == "" {
		return ExecutionPlanValidationError{Field: "category", Message: "category cannot be empty"}
	}

	// Validate status
	if !ep.isValidStatus(ep.Status) {
		return ExecutionPlanValidationError{Field: "status", Message: "invalid status"}
	}

	// Validate steps
	for i, step := range ep.Steps {
		if err := ep.validateStep(step); err != nil {
			return fmt.Errorf("step %d validation failed: %w", i, err)
		}
	}

	return nil
}

// isValidStatus checks if the status is valid
func (ep *ExecutionPlan) isValidStatus(status ExecutionPlanStatus) bool {
	validStatuses := []ExecutionPlanStatus{
		ExecutionPlanStatusPending,
		ExecutionPlanStatusRunning,
		ExecutionPlanStatusCompleted,
		ExecutionPlanStatusFailed,
		ExecutionPlanStatusCancelled,
	}

	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// validateStep validates an execution step
func (ep *ExecutionPlan) validateStep(step ExecutionStep) error {
	if step.ID == "" {
		return ExecutionPlanValidationError{Field: "step.id", Message: "step ID cannot be empty"}
	}

	if step.Name == "" {
		return ExecutionPlanValidationError{Field: "step.name", Message: "step name cannot be empty"}
	}

	if step.AgentID == "" {
		return ExecutionPlanValidationError{Field: "step.agent_id", Message: "step agent ID cannot be empty"}
	}

	if step.AgentType == "" {
		return ExecutionPlanValidationError{Field: "step.agent_type", Message: "step agent type cannot be empty"}
	}

	if !ep.isValidStatus(step.Status) {
		return ExecutionPlanValidationError{Field: "step.status", Message: "invalid step status"}
	}

	return nil
}

// AddStep adds a step to the execution plan with validation
func (ep *ExecutionPlan) AddStep(step ExecutionStep) error {
	if err := ep.validateStep(step); err != nil {
		return err
	}

	// Check for duplicate step IDs
	for _, existingStep := range ep.Steps {
		if existingStep.ID == step.ID {
			return ExecutionPlanValidationError{Field: "step.id", Message: "step ID must be unique"}
		}
	}

	ep.Steps = append(ep.Steps, step)
	return nil
}

// Start starts the execution plan
func (ep *ExecutionPlan) Start() error {
	if ep.Status != ExecutionPlanStatusPending {
		return errors.New("can only start pending execution plans")
	}

	now := time.Now().UTC()
	ep.Status = ExecutionPlanStatusRunning
	ep.StartedAt = &now

	return nil
}

// Complete completes the execution plan
func (ep *ExecutionPlan) Complete(result string) error {
	if ep.Status != ExecutionPlanStatusRunning {
		return errors.New("can only complete running execution plans")
	}

	now := time.Now().UTC()
	ep.Status = ExecutionPlanStatusCompleted
	ep.CompletedAt = &now
	ep.Result = result

	if ep.StartedAt != nil {
		ep.ActualTime = now.Sub(*ep.StartedAt)
	}

	return nil
}

// Fail marks the execution plan as failed
func (ep *ExecutionPlan) Fail(errorMsg string) error {
	if ep.Status == ExecutionPlanStatusCompleted {
		return errors.New("cannot fail completed execution plans")
	}

	now := time.Now().UTC()
	ep.Status = ExecutionPlanStatusFailed
	ep.Error = errorMsg

	if ep.CompletedAt == nil {
		ep.CompletedAt = &now
	}

	if ep.StartedAt != nil {
		ep.ActualTime = now.Sub(*ep.StartedAt)
	}

	return nil
}

// Cancel cancels the execution plan
func (ep *ExecutionPlan) Cancel() error {
	if ep.Status == ExecutionPlanStatusCompleted {
		return errors.New("cannot cancel completed execution plans")
	}

	now := time.Now().UTC()
	ep.Status = ExecutionPlanStatusCancelled

	if ep.CompletedAt == nil {
		ep.CompletedAt = &now
	}

	if ep.StartedAt != nil {
		ep.ActualTime = now.Sub(*ep.StartedAt)
	}

	return nil
}

// GetPendingSteps returns all pending steps
func (ep *ExecutionPlan) GetPendingSteps() []ExecutionStep {
	var pending []ExecutionStep
	for _, step := range ep.Steps {
		if step.Status == ExecutionPlanStatusPending {
			pending = append(pending, step)
		}
	}
	return pending
}

// GetRunnableSteps returns steps that can be executed (pending with no pending dependencies)
func (ep *ExecutionPlan) GetRunnableSteps() []ExecutionStep {
	var runnable []ExecutionStep

	for _, step := range ep.Steps {
		if step.Status != ExecutionPlanStatusPending {
			continue
		}

		// Check if all dependencies are completed
		canRun := true
		for _, depID := range step.Dependencies {
			depCompleted := false
			for _, depStep := range ep.Steps {
				if depStep.ID == depID && depStep.Status == ExecutionPlanStatusCompleted {
					depCompleted = true
					break
				}
			}
			if !depCompleted {
				canRun = false
				break
			}
		}

		if canRun {
			runnable = append(runnable, step)
		}
	}

	return runnable
}

// UpdateStepStatus updates the status of a specific step
func (ep *ExecutionPlan) UpdateStepStatus(stepID string, status ExecutionPlanStatus, result map[string]interface{}, errorMsg string) error {
	for i, step := range ep.Steps {
		if step.ID == stepID {
			if !ep.isValidStatus(status) {
				return ExecutionPlanValidationError{Field: "status", Message: "invalid status"}
			}

			now := time.Now().UTC()
			ep.Steps[i].Status = status

			if status == ExecutionPlanStatusRunning && step.StartedAt == nil {
				ep.Steps[i].StartedAt = &now
			}

			if status == ExecutionPlanStatusCompleted || status == ExecutionPlanStatusFailed {
				ep.Steps[i].CompletedAt = &now
				if result != nil {
					ep.Steps[i].Result = result
				}
				if errorMsg != "" {
					ep.Steps[i].Error = errorMsg
				}
			}

			return nil
		}
	}

	return fmt.Errorf("step with ID %s not found", stepID)
}

// IsCompleted returns true if the execution plan is completed (all steps completed)
func (ep *ExecutionPlan) IsCompleted() bool {
	if len(ep.Steps) == 0 {
		return false
	}

	for _, step := range ep.Steps {
		if step.Status != ExecutionPlanStatusCompleted {
			return false
		}
	}

	return true
}

// HasFailed returns true if any step has failed
func (ep *ExecutionPlan) HasFailed() bool {
	for _, step := range ep.Steps {
		if step.Status == ExecutionPlanStatusFailed {
			return true
		}
	}
	return false
}
