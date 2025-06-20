package application

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewService tests the service constructor
func TestNewService(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	// Test
	service := helpers.CreateTestApplicationService(t)

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, helpers.Graph, service.Graph)
}

// TestCreateEnvironment tests environment creation
func TestCreateEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		env         contracts.EnvironmentContract
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid environment",
			env: contracts.EnvironmentContract{
				Metadata: contracts.Metadata{
					Name:  "test-env",
					Owner: "test-owner",
				},
				Spec: contracts.EnvironmentSpec{
					Description: "Test environment",
				},
			},
			expectError: false,
		},
		{
			name: "missing name",
			env: contracts.EnvironmentContract{
				Metadata: contracts.Metadata{
					Owner: "test-owner",
				},
				Spec: contracts.EnvironmentSpec{
					Description: "Test environment",
				},
			},
			expectError: true,
			errorMsg:    "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers := CreateTestHelpers(t)
			defer helpers.CleanupTestData(t)

			service := helpers.CreateTestEnvironmentService(t)

			// Test
			err := service.CreateEnvironment(tt.env)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify environment was created in graph (we'll implement this)
				// For now, just check no error occurred
			}
		})
	}
}

// TestCreateEnvironmentFromContract tests the contract-based creation method
func TestCreateEnvironmentFromContract(t *testing.T) {
	tests := []struct {
		name        string
		env         *contracts.EnvironmentContract
		expectError bool
	}{
		{
			name: "valid environment contract",
			env: &contracts.EnvironmentContract{
				Metadata: contracts.Metadata{
					Name:  "test-env",
					Owner: "test-owner",
				},
				Spec: contracts.EnvironmentSpec{
					Description: "Test environment",
				},
			},
			expectError: false,
		},
		{
			name:        "nil contract",
			env:         nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers := CreateTestHelpers(t)
			defer helpers.CleanupTestData(t)

			ctx := context.Background()
			service := helpers.CreateTestEnvironmentService(t)

			// Test
			result, err := service.CreateEnvironmentFromContract(ctx, tt.env)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify result structure
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, tt.env.Metadata.Name, resultMap["name"])
				assert.Equal(t, "created", resultMap["status"])
				assert.Equal(t, tt.env.Spec.Description, resultMap["description"])
				assert.Equal(t, tt.env.Metadata.Owner, resultMap["owner"])
			}
		})
	}
}
