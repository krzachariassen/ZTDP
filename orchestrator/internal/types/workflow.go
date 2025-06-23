package types

import "time"

// WorkflowRequest represents a request to execute a workflow
type WorkflowRequest struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Steps       []WorkflowStep         `json:"steps"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Action       string                 `json:"action"`
	Capabilities []string               `json:"capabilities"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	RetryPolicy  *RetryPolicy           `json:"retry_policy,omitempty"`
}

// RetryPolicy defines how to handle step failures
type RetryPolicy struct {
	MaxRetries int           `json:"max_retries"`
	BackoffMs  int           `json:"backoff_ms"`
	Timeout    time.Duration `json:"timeout"`
}

// WorkflowExecution represents the execution state of a workflow
type WorkflowExecution struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      WorkflowStatus         `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       *string                `json:"error,omitempty"`
	Steps       []StepExecution        `json:"steps"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// StepExecution represents the execution state of a workflow step
type StepExecution struct {
	StepID      string                 `json:"step_id"`
	AgentID     string                 `json:"agent_id"`
	Status      StepStatus             `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       *string                `json:"error,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
}

// WorkflowStatus represents the status of a workflow execution
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// StepStatus represents the status of a step execution
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// WorkflowPlan represents a planned workflow execution
type WorkflowPlan struct {
	ID                  string                 `json:"id"`
	WorkflowID          string                 `json:"workflow_id"`
	Steps               []PlannedStep          `json:"steps"`
	EstimatedDurationMs int                    `json:"estimated_duration_ms"`
	Context             map[string]interface{} `json:"context,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
}

// PlannedStep represents a step in a workflow plan
type PlannedStep struct {
	StepID       string   `json:"step_id"`
	AgentID      string   `json:"agent_id"`
	Dependencies []string `json:"dependencies,omitempty"`
	Order        int      `json:"order"`
}
