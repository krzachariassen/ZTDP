package types

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

// DeploymentMetadataKey is the key used to store deployment metadata in edge metadata
const DeploymentMetadataKey = "deployment"

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

// SetDeploymentStatus sets the deployment status in edge metadata
func SetDeploymentStatus(metadata map[string]interface{}, status DeploymentStatus, message string) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid deployment status: %s", status)
	}

	// Validate status transition if there's an existing status
	if err := ValidateStatusTransition(metadata, status); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Get or create deployment metadata
	var deploymentMeta map[string]interface{}
	if existing, exists := metadata[DeploymentMetadataKey]; exists {
		if existingMap, ok := existing.(map[string]interface{}); ok {
			deploymentMeta = existingMap
		} else {
			deploymentMeta = make(map[string]interface{})
		}
	} else {
		deploymentMeta = make(map[string]interface{})
	}

	// Update status and message
	deploymentMeta["status"] = string(status)
	deploymentMeta["message"] = message
	deploymentMeta["last_updated"] = time.Now().Format(time.RFC3339)

	// Store back in metadata
	metadata[DeploymentMetadataKey] = deploymentMeta

	return nil
}

// GetDeploymentStatus retrieves the deployment status from edge metadata
func GetDeploymentStatus(metadata map[string]interface{}) (DeploymentStatus, string, bool) {
	deploymentMeta, exists := metadata[DeploymentMetadataKey]
	if !exists {
		return "", "", false
	}

	deploymentMap, ok := deploymentMeta.(map[string]interface{})
	if !ok {
		return "", "", false
	}

	statusStr, hasStatus := deploymentMap["status"].(string)
	if !hasStatus {
		return "", "", false
	}

	message, _ := deploymentMap["message"].(string)

	return DeploymentStatus(statusStr), message, true
}

// SetDeploymentProgress sets the deployment progress in edge metadata
func SetDeploymentProgress(metadata map[string]interface{}, progress float64, message string) error {
	// Get or create deployment metadata
	var deploymentMeta map[string]interface{}
	if existing, exists := metadata[DeploymentMetadataKey]; exists {
		if existingMap, ok := existing.(map[string]interface{}); ok {
			deploymentMeta = existingMap
		} else {
			deploymentMeta = make(map[string]interface{})
		}
	} else {
		deploymentMeta = make(map[string]interface{})
	}

	// Update progress
	deploymentMeta["progress"] = progress
	if message != "" {
		deploymentMeta["progress_message"] = message
	}
	deploymentMeta["last_updated"] = time.Now().Format(time.RFC3339)

	// Store back in metadata
	metadata[DeploymentMetadataKey] = deploymentMeta

	return nil
}

// GetDeploymentProgress retrieves the deployment progress from edge metadata
func GetDeploymentProgress(metadata map[string]interface{}) (float64, string, bool) {
	deploymentMeta, exists := metadata[DeploymentMetadataKey]
	if !exists {
		return 0, "", false
	}

	deploymentMap, ok := deploymentMeta.(map[string]interface{})
	if !ok {
		return 0, "", false
	}

	progress, hasProgress := deploymentMap["progress"].(float64)
	if !hasProgress {
		return 0, "", false
	}

	message, _ := deploymentMap["progress_message"].(string)

	return progress, message, true
}

// AddDeploymentEvent adds an event to the deployment history in edge metadata
func AddDeploymentEvent(metadata map[string]interface{}, level, message string, timestamp time.Time) error {
	// Get or create deployment metadata
	var deploymentMeta map[string]interface{}
	if existing, exists := metadata[DeploymentMetadataKey]; exists {
		if existingMap, ok := existing.(map[string]interface{}); ok {
			deploymentMeta = existingMap
		} else {
			deploymentMeta = make(map[string]interface{})
		}
	} else {
		deploymentMeta = make(map[string]interface{})
	}

	// Get existing events or create new slice
	var events []interface{}
	if existingEvents, exists := deploymentMeta["events"]; exists {
		if eventSlice, ok := existingEvents.([]interface{}); ok {
			events = eventSlice
		}
	}

	// Create new event
	event := map[string]interface{}{
		"level":     level,
		"message":   message,
		"timestamp": timestamp.Format(time.RFC3339),
	}

	// Add event to slice
	events = append(events, event)

	// Store back in metadata
	deploymentMeta["events"] = events
	deploymentMeta["last_updated"] = time.Now().Format(time.RFC3339)
	metadata[DeploymentMetadataKey] = deploymentMeta

	return nil
}

// GetDeploymentEvents retrieves the deployment events from edge metadata
func GetDeploymentEvents(metadata map[string]interface{}) []map[string]interface{} {
	deploymentMeta, exists := metadata[DeploymentMetadataKey]
	if !exists {
		return []map[string]interface{}{}
	}

	deploymentMap, ok := deploymentMeta.(map[string]interface{})
	if !ok {
		return []map[string]interface{}{}
	}

	eventsInterface, exists := deploymentMap["events"]
	if !exists {
		return []map[string]interface{}{}
	}

	eventsSlice, ok := eventsInterface.([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}

	// Convert to proper type
	events := make([]map[string]interface{}, len(eventsSlice))
	for i, eventInterface := range eventsSlice {
		if eventMap, ok := eventInterface.(map[string]interface{}); ok {
			events[i] = eventMap
		}
	}

	return events
}

// ValidateStatusTransition validates that a status transition is allowed
func ValidateStatusTransition(metadata map[string]interface{}, newStatus DeploymentStatus) error {
	currentStatus, _, exists := GetDeploymentStatus(metadata)
	if !exists {
		// No existing status, any valid status is allowed
		return nil
	}

	if !currentStatus.CanTransitionTo(newStatus) {
		return fmt.Errorf("cannot transition from %s to %s", currentStatus, newStatus)
	}

	return nil
}
