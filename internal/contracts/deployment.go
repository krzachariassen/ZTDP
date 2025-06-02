package contracts

import (
	"fmt"
	"time"
)

// DeploymentStatus represents the current state of a deployment
type DeploymentStatus string

// Deployment status constants
const (
	StatusPending    DeploymentStatus = "pending"
	StatusInProgress DeploymentStatus = "in_progress"
	StatusSucceeded  DeploymentStatus = "succeeded"
	StatusFailed     DeploymentStatus = "failed"
	StatusCancelled  DeploymentStatus = "cancelled"
)

// DeploymentContract represents a deployment operation with its status and lifecycle
type DeploymentContract struct {
	EdgeID    string                 `json:"edge_id"`
	Status    DeploymentStatus       `json:"status"`
	Message   string                 `json:"message"`
	Progress  float64                `json:"progress"`
	StartTime *time.Time             `json:"start_time,omitempty"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Events    []DeploymentEvent      `json:"events"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DeploymentEvent represents an event in the deployment lifecycle
type DeploymentEvent struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ID returns the deployment contract ID (edge ID)
func (d *DeploymentContract) ID() string {
	return d.EdgeID
}

// Kind returns the contract kind
func (d *DeploymentContract) Kind() string {
	return "deployment"
}

// Validate validates the deployment contract
func (d *DeploymentContract) Validate() error {
	if d.EdgeID == "" {
		return fmt.Errorf("edge_id is required")
	}

	if !d.Status.IsValid() {
		return fmt.Errorf("invalid deployment status: %s", d.Status)
	}

	if d.Progress < 0 || d.Progress > 1 {
		return fmt.Errorf("progress must be between 0 and 1, got %f", d.Progress)
	}

	return nil
}

// GetMetadata returns contract metadata
func (d *DeploymentContract) GetMetadata() Metadata {
	return Metadata{
		Name:  fmt.Sprintf("deployment-%s", d.EdgeID),
		Owner: "deployment-engine",
	}
}

// IsValid checks if the deployment status is a valid status
func (s DeploymentStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusSucceeded, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the status represents a terminal state
func (s DeploymentStatus) IsTerminal() bool {
	switch s {
	case StatusSucceeded, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

// IsSuccess returns true if the status represents a successful deployment
func (s DeploymentStatus) IsSuccess() bool {
	return s == StatusSucceeded
}

// CanTransitionTo checks if this status can transition to the target status
func (s DeploymentStatus) CanTransitionTo(target DeploymentStatus) bool {
	// Same status transitions are always allowed (idempotent)
	if s == target {
		return true
	}

	// From pending
	if s == StatusPending {
		switch target {
		case StatusInProgress, StatusCancelled:
			return true
		default:
			return false
		}
	}

	// From in_progress
	if s == StatusInProgress {
		switch target {
		case StatusSucceeded, StatusFailed, StatusCancelled:
			return true
		default:
			return false
		}
	}

	// Terminal states cannot transition to other states
	if s.IsTerminal() {
		return false
	}

	return false
}

// String returns the string representation of the deployment status
func (s DeploymentStatus) String() string {
	return string(s)
}
