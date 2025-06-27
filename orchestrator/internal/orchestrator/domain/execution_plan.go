package domain

import (
	"time"

	"github.com/google/uuid"
)

// ExecutionStatus represents the status of an execution plan
type ExecutionStatus string

const (
	ExecutionStatusPending    ExecutionStatus = "PENDING"
	ExecutionStatusInProgress ExecutionStatus = "IN_PROGRESS"
	ExecutionStatusCompleted  ExecutionStatus = "COMPLETED"
	ExecutionStatusFailed     ExecutionStatus = "FAILED"
	ExecutionStatusCancelled  ExecutionStatus = "CANCELLED"
)

// ExecutionStep represents a single step in an execution plan
type ExecutionStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	AgentID     string                 `json:"agent_id"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      ExecutionStatus        `json:"status"`
	DependsOn   []string               `json:"depends_on"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       *string                `json:"error,omitempty"`
}

// ExecutionPlan represents a plan for executing tasks across multiple agents
type ExecutionPlan struct {
	ID          string                 `json:"id"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Steps       []ExecutionStep        `json:"steps"`
	Status      ExecutionStatus        `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       *string                `json:"error,omitempty"`
}

// NewExecutionPlan creates a new execution plan
func NewExecutionPlan(action string, parameters map[string]interface{}) *ExecutionPlan {
	return &ExecutionPlan{
		ID:         uuid.New().String(),
		Action:     action,
		Parameters: parameters,
		Steps:      make([]ExecutionStep, 0),
		Status:     ExecutionStatusPending,
		CreatedAt:  time.Now(),
	}
}

// AddStep adds a step to the execution plan
func (ep *ExecutionPlan) AddStep(name, agentID, action string, parameters map[string]interface{}, dependsOn []string) {
	step := ExecutionStep{
		ID:         uuid.New().String(),
		Name:       name,
		AgentID:    agentID,
		Action:     action,
		Parameters: parameters,
		Status:     ExecutionStatusPending,
		DependsOn:  dependsOn,
		CreatedAt:  time.Now(),
	}
	ep.Steps = append(ep.Steps, step)
}

// UpdateStatus updates the plan's status
func (ep *ExecutionPlan) UpdateStatus(status ExecutionStatus) {
	ep.Status = status
	now := time.Now()

	switch status {
	case ExecutionStatusInProgress:
		if ep.StartedAt == nil {
			ep.StartedAt = &now
		}
	case ExecutionStatusCompleted, ExecutionStatusFailed, ExecutionStatusCancelled:
		if ep.CompletedAt == nil {
			ep.CompletedAt = &now
		}
	}
}

// SetError sets an error on the execution plan
func (ep *ExecutionPlan) SetError(err string) {
	ep.Error = &err
	ep.UpdateStatus(ExecutionStatusFailed)
}

// IsComplete returns true if the plan is completed (success or failure)
func (ep *ExecutionPlan) IsComplete() bool {
	return ep.Status == ExecutionStatusCompleted ||
		ep.Status == ExecutionStatusFailed ||
		ep.Status == ExecutionStatusCancelled
}
