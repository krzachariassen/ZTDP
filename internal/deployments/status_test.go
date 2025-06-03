package deployments

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetDeploymentStatus(t *testing.T) {
	tests := []struct {
		name        string
		metadata    map[string]interface{}
		status      DeploymentStatus
		message     string
		expectError bool
	}{
		{
			name:        "Set status on empty metadata",
			metadata:    map[string]interface{}{},
			status:      StatusPending,
			message:     "Deployment queued",
			expectError: false,
		},
		{
			name: "Update existing status with valid transition",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: map[string]interface{}{
					"status": string(StatusPending),
				},
			},
			status:      StatusInProgress,
			message:     "Starting deployment",
			expectError: false,
		},
		{
			name:        "Invalid status should fail",
			metadata:    map[string]interface{}{},
			status:      DeploymentStatus("invalid"),
			message:     "Should fail",
			expectError: true,
		},
		{
			name: "Invalid transition should fail",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: map[string]interface{}{
					"status": string(StatusSucceeded),
				},
			},
			status:      StatusInProgress,
			message:     "Should fail",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetDeploymentStatus(tt.metadata, tt.status, tt.message)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify the status was set correctly
			status, message, exists := GetDeploymentStatus(tt.metadata)
			require.True(t, exists)
			assert.Equal(t, tt.status, status)
			assert.Equal(t, tt.message, message)

			// Verify timestamp was set
			deploymentMeta := tt.metadata[DeploymentMetadataKey].(map[string]interface{})
			_, hasTimestamp := deploymentMeta["last_updated"]
			assert.True(t, hasTimestamp)
		})
	}
}

func TestGetDeploymentStatus(t *testing.T) {
	tests := []struct {
		name            string
		metadata        map[string]interface{}
		expectedStatus  DeploymentStatus
		expectedMessage string
		expectedExists  bool
	}{
		{
			name:            "No deployment metadata",
			metadata:        map[string]interface{}{},
			expectedStatus:  "",
			expectedMessage: "",
			expectedExists:  false,
		},
		{
			name: "Valid deployment metadata",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: map[string]interface{}{
					"status":  string(StatusInProgress),
					"message": "Deploying resources",
				},
			},
			expectedStatus:  StatusInProgress,
			expectedMessage: "Deploying resources",
			expectedExists:  true,
		},
		{
			name: "Invalid deployment metadata structure",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: "invalid",
			},
			expectedStatus:  "",
			expectedMessage: "",
			expectedExists:  false,
		},
		{
			name: "Missing status field",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: map[string]interface{}{
					"message": "Some message",
				},
			},
			expectedStatus:  "",
			expectedMessage: "",
			expectedExists:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message, exists := GetDeploymentStatus(tt.metadata)

			assert.Equal(t, tt.expectedStatus, status)
			assert.Equal(t, tt.expectedMessage, message)
			assert.Equal(t, tt.expectedExists, exists)
		})
	}
}

func TestValidateStatusTransition(t *testing.T) {
	tests := []struct {
		name        string
		metadata    map[string]interface{}
		newStatus   DeploymentStatus
		expectError bool
	}{
		{
			name:        "No existing status - any valid status allowed",
			metadata:    map[string]interface{}{},
			newStatus:   StatusPending,
			expectError: false,
		},
		{
			name: "Valid transition: pending to in_progress",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: map[string]interface{}{
					"status": string(StatusPending),
				},
			},
			newStatus:   StatusInProgress,
			expectError: false,
		},
		{
			name: "Valid transition: in_progress to succeeded",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: map[string]interface{}{
					"status": string(StatusInProgress),
				},
			},
			newStatus:   StatusSucceeded,
			expectError: false,
		},
		{
			name: "Valid transition: in_progress to failed",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: map[string]interface{}{
					"status": string(StatusInProgress),
				},
			},
			newStatus:   StatusFailed,
			expectError: false,
		},
		{
			name: "Invalid transition: succeeded to in_progress",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: map[string]interface{}{
					"status": string(StatusSucceeded),
				},
			},
			newStatus:   StatusInProgress,
			expectError: true,
		},
		{
			name: "Invalid transition: failed to in_progress",
			metadata: map[string]interface{}{
				DeploymentMetadataKey: map[string]interface{}{
					"status": string(StatusFailed),
				},
			},
			newStatus:   StatusInProgress,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatusTransition(tt.metadata, tt.newStatus)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeploymentStatusWorkflow(t *testing.T) {
	metadata := map[string]interface{}{}

	// Start with pending
	err := SetDeploymentStatus(metadata, StatusPending, "Deployment queued")
	require.NoError(t, err)

	status, message, exists := GetDeploymentStatus(metadata)
	assert.True(t, exists)
	assert.Equal(t, StatusPending, status)
	assert.Equal(t, "Deployment queued", message)

	// Move to in_progress
	err = SetDeploymentStatus(metadata, StatusInProgress, "Starting deployment")
	require.NoError(t, err)

	status, message, exists = GetDeploymentStatus(metadata)
	assert.True(t, exists)
	assert.Equal(t, StatusInProgress, status)
	assert.Equal(t, "Starting deployment", message)

	// Complete successfully
	err = SetDeploymentStatus(metadata, StatusSucceeded, "Deployment completed")
	require.NoError(t, err)

	status, message, exists = GetDeploymentStatus(metadata)
	assert.True(t, exists)
	assert.Equal(t, StatusSucceeded, status)
	assert.Equal(t, "Deployment completed", message)

	// Verify we can't go back to in_progress
	err = SetDeploymentStatus(metadata, StatusInProgress, "Should fail")
	assert.Error(t, err)
}

func TestDeploymentStatusTimestamp(t *testing.T) {
	metadata := map[string]interface{}{}

	// Set initial status
	before := time.Now()
	err := SetDeploymentStatus(metadata, StatusPending, "Test")
	require.NoError(t, err)
	after := time.Now()

	// Check timestamp was set and is reasonable
	deploymentMeta := metadata[DeploymentMetadataKey].(map[string]interface{})
	timestampStr, exists := deploymentMeta["last_updated"].(string)
	require.True(t, exists)

	timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
	require.NoError(t, err)

	// The timestamp should be between before and after (inclusive)
	assert.True(t, timestamp.After(before) || timestamp.Equal(before))
	assert.True(t, timestamp.Before(after) || timestamp.Equal(after))
}
