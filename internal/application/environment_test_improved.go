package application

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironmentService_CreateEnvironment_WithHelpers demonstrates improved test structure
func TestEnvironmentService_CreateEnvironment_WithHelpers(t *testing.T) {
	// Setup: Single call to get all test infrastructure
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	// Create service under test
	envService := helpers.CreateTestEnvironmentService(t)

	tests := []struct {
		name        string
		environment contracts.EnvironmentContract
		expectError bool
	}{
		{
			name: "valid environment",
			environment: contracts.EnvironmentContract{
				Metadata: contracts.Metadata{
					Name:  "test-env",
					Owner: "test-team",
				},
				Spec: contracts.EnvironmentSpec{
					Description: "Test environment",
				},
			},
			expectError: false,
		},
		{
			name: "empty name",
			environment: contracts.EnvironmentContract{
				Metadata: contracts.Metadata{
					Name:  "",
					Owner: "test-team",
				},
				Spec: contracts.EnvironmentSpec{
					Description: "Test environment",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := envService.CreateEnvironment(tt.environment)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestApplicationService_WithHelpers demonstrates how helpers simplify application service testing
func TestApplicationService_WithHelpers(t *testing.T) {
	// Setup: All dependencies resolved in one call
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	// Create pre-populated test data easily
	app := helpers.CreateTestApplication(t, "test-app")
	service := helpers.CreateTestService(t, "test-app", "test-service")
	env := helpers.CreateTestEnvironment(t, "test-env")
	release := helpers.CreateTestRelease(t, "test-app", "test-release")

	// Create application service for testing
	appService := helpers.CreateTestApplicationService(t)

	// Now test business logic without setup overhead
	require.NotNil(t, appService)
	assert.Equal(t, "test-app", app.Metadata.Name)
	assert.Equal(t, "test-service", service.Metadata.Name)
	assert.Equal(t, "test-env", env.Metadata.Name)
	assert.Equal(t, "test-release", release.Metadata.Name)
}

// TestMultipleServices_Integration demonstrates cross-domain testing
func TestMultipleServices_Integration(t *testing.T) {
	// Setup: Single setup for integration testing
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	// Create all domain services easily
	appService := helpers.CreateTestApplicationService(t)
	serviceService := helpers.CreateTestServiceService(t)
	envService := helpers.CreateTestEnvironmentService(t)
	releaseService := helpers.CreateTestReleaseService(t)

	// Test cross-domain interactions
	require.NotNil(t, appService)
	require.NotNil(t, serviceService)
	require.NotNil(t, envService)
	require.NotNil(t, releaseService)

	// All services share the same graph and can interact
	assert.Equal(t, helpers.Graph, appService.Graph)
	assert.Equal(t, helpers.Graph, serviceService.Graph)
	assert.Equal(t, helpers.Graph, envService.Graph)
	assert.Equal(t, helpers.Graph, releaseService.Graph)
}
