package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   DeploymentStatus
		expected string
	}{
		{"Pending status", StatusPending, "pending"},
		{"InProgress status", StatusInProgress, "in_progress"},
		{"Succeeded status", StatusSucceeded, "succeeded"},
		{"Failed status", StatusFailed, "failed"},
		{"Cancelled status", StatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestDeploymentStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   DeploymentStatus
		expected bool
	}{
		{"Valid pending", StatusPending, true},
		{"Valid in_progress", StatusInProgress, true},
		{"Valid succeeded", StatusSucceeded, true},
		{"Valid failed", StatusFailed, true},
		{"Valid cancelled", StatusCancelled, true},
		{"Invalid empty", DeploymentStatus(""), false},
		{"Invalid unknown", DeploymentStatus("unknown"), false},
		{"Invalid case", DeploymentStatus("PENDING"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status.IsValid() != tt.expected {
				t.Errorf("Expected %v, got %v for status %s", tt.expected, tt.status.IsValid(), tt.status)
			}
		})
	}
}

func TestDeploymentStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name     string
		status   DeploymentStatus
		expected bool
	}{
		{"Pending is not terminal", StatusPending, false},
		{"InProgress is not terminal", StatusInProgress, false},
		{"Succeeded is terminal", StatusSucceeded, true},
		{"Failed is terminal", StatusFailed, true},
		{"Cancelled is terminal", StatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status.IsTerminal() != tt.expected {
				t.Errorf("Expected %v, got %v for status %s", tt.expected, tt.status.IsTerminal(), tt.status)
			}
		})
	}
}

func TestDeploymentStatus_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		status   DeploymentStatus
		expected bool
	}{
		{"Succeeded is success", StatusSucceeded, true},
		{"Failed is not success", StatusFailed, false},
		{"Pending is not success", StatusPending, false},
		{"InProgress is not success", StatusInProgress, false},
		{"Cancelled is not success", StatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status.IsSuccess() != tt.expected {
				t.Errorf("Expected %v, got %v for status %s", tt.expected, tt.status.IsSuccess(), tt.status)
			}
		})
	}
}

func TestDeploymentStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     DeploymentStatus
		to       DeploymentStatus
		expected bool
	}{
		// Valid transitions from pending
		{"Pending to InProgress", StatusPending, StatusInProgress, true},
		{"Pending to Cancelled", StatusPending, StatusCancelled, true},

		// Valid transitions from in_progress
		{"InProgress to Succeeded", StatusInProgress, StatusSucceeded, true},
		{"InProgress to Failed", StatusInProgress, StatusFailed, true},
		{"InProgress to Cancelled", StatusInProgress, StatusCancelled, true},

		// Invalid transitions from terminal states
		{"Succeeded to Failed", StatusSucceeded, StatusFailed, false},
		{"Failed to Succeeded", StatusFailed, StatusSucceeded, false},
		{"Cancelled to InProgress", StatusCancelled, StatusInProgress, false},

		// Invalid backwards transitions
		{"InProgress to Pending", StatusInProgress, StatusPending, false},
		{"Succeeded to Pending", StatusSucceeded, StatusPending, false},

		// Same status transitions (should be allowed for idempotency)
		{"Pending to Pending", StatusPending, StatusPending, true},
		{"InProgress to InProgress", StatusInProgress, StatusInProgress, true},
		{"Succeeded to Succeeded", StatusSucceeded, StatusSucceeded, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.from.CanTransitionTo(tt.to) != tt.expected {
				t.Errorf("Expected %v for transition from %s to %s", tt.expected, tt.from, tt.to)
			}
		})
	}
}

func TestDeploymentContract_Validate(t *testing.T) {
	tests := []struct {
		name        string
		contract    *DeploymentContract
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid contract",
			contract: &DeploymentContract{
				EdgeID:   "edge1",
				Status:   StatusPending,
				Progress: 0.5,
			},
			expectError: false,
		},
		{
			name: "Missing edge ID",
			contract: &DeploymentContract{
				Status:   StatusPending,
				Progress: 0.5,
			},
			expectError: true,
			errorMsg:    "edge_id is required",
		},
		{
			name: "Invalid status",
			contract: &DeploymentContract{
				EdgeID:   "edge1",
				Status:   DeploymentStatus("invalid"),
				Progress: 0.5,
			},
			expectError: true,
			errorMsg:    "invalid deployment status",
		},
		{
			name: "Progress too low",
			contract: &DeploymentContract{
				EdgeID:   "edge1",
				Status:   StatusPending,
				Progress: -0.1,
			},
			expectError: true,
			errorMsg:    "progress must be between 0 and 1",
		},
		{
			name: "Progress too high",
			contract: &DeploymentContract{
				EdgeID:   "edge1",
				Status:   StatusPending,
				Progress: 1.1,
			},
			expectError: true,
			errorMsg:    "progress must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contract.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeploymentContract_ContractInterface(t *testing.T) {
	contract := &DeploymentContract{
		EdgeID: "edge1",
		Status: StatusPending,
	}

	// Test that it implements the Contract interface
	assert.Equal(t, "edge1", contract.ID())
	assert.Equal(t, "deployment", contract.Kind())

	metadata := contract.GetMetadata()
	assert.Equal(t, "deployment-edge1", metadata.Name)
	assert.Equal(t, "deployment-engine", metadata.Owner)
}
