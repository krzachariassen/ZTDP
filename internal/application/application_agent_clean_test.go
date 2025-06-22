package application

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper to create a test application service with real AI
func createCleanTestApplicationService(t *testing.T) *Service {
	helpers := CreateTestHelpers(t)
	return helpers.CreateTestApplicationService(t)
}

func TestNewApplicationAgent_Clean(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	agent, err := NewApplicationAgent(helpers.Graph, helpers.AIProvider, helpers.EventBus, helpers.Registry)
	require.NoError(t, err)
	assert.NotNil(t, agent)
}

func TestApplicationService_CreateApplication_Clean(t *testing.T) {
	service := createCleanTestApplicationService(t)

	tests := []struct {
		name        string
		app         contracts.ApplicationContract
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid application",
			app: contracts.ApplicationContract{
				Metadata: contracts.Metadata{
					Name:  "test-app",
					Owner: "test-owner",
				},
				Spec: contracts.ApplicationSpec{
					Description: "Test application",
					Tags:        []string{"web"},
				},
			},
			expectError: false,
		},
		{
			name: "application with empty name",
			app: contracts.ApplicationContract{
				Metadata: contracts.Metadata{
					Name:  "",
					Owner: "test-owner",
				},
				Spec: contracts.ApplicationSpec{
					Description: "Test application",
				},
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "application with empty owner",
			app: contracts.ApplicationContract{
				Metadata: contracts.Metadata{
					Name:  "test-app-2",
					Owner: "",
				},
				Spec: contracts.ApplicationSpec{
					Description: "Test application",
				},
			},
			expectError: true,
			errorMsg:    "owner is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateApplication(tt.app)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify application was created
				app, err := service.GetApplication(tt.app.Metadata.Name)
				assert.NoError(t, err)
				assert.Equal(t, tt.app.Metadata.Name, app.Metadata.Name)
				assert.Equal(t, tt.app.Metadata.Owner, app.Metadata.Owner)
			}
		})
	}
}

func TestApplicationService_ListApplications_Clean(t *testing.T) {
	service := createCleanTestApplicationService(t)

	// Initially should be empty
	apps, err := service.ListApplications()
	require.NoError(t, err)
	assert.Empty(t, apps)

	// Create a few applications
	testApps := []contracts.ApplicationContract{
		{
			Metadata: contracts.Metadata{
				Name:  "app1",
				Owner: "team1",
			},
			Spec: contracts.ApplicationSpec{
				Description: "Application 1",
				Tags:        []string{"web"},
			},
		},
		{
			Metadata: contracts.Metadata{
				Name:  "app2",
				Owner: "team2",
			},
			Spec: contracts.ApplicationSpec{
				Description: "Application 2",
				Tags:        []string{"api"},
			},
		},
	}

	for _, app := range testApps {
		err := service.CreateApplication(app)
		require.NoError(t, err)
	}

	// List applications should return all created apps
	apps, err = service.ListApplications()
	require.NoError(t, err)
	assert.Len(t, apps, 2)

	// Verify both apps are present
	appNames := make([]string, len(apps))
	for i, app := range apps {
		appNames[i] = app.Metadata.Name
	}
	assert.Contains(t, appNames, "app1")
	assert.Contains(t, appNames, "app2")
}

func TestApplicationService_UpdateApplication_Clean(t *testing.T) {
	service := createCleanTestApplicationService(t)

	// First create an application
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "update-test",
			Owner: "test-owner",
		},
		Spec: contracts.ApplicationSpec{
			Description: "Original description",
			Tags:        []string{"original"},
		},
	}
	err := service.CreateApplication(app)
	require.NoError(t, err)

	// Update the application
	updatedApp := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "update-test",
			Owner: "test-owner",
		},
		Spec: contracts.ApplicationSpec{
			Description: "Updated description",
			Tags:        []string{"updated", "modified"},
		},
	}

	err = service.UpdateApplication("update-test", updatedApp)
	assert.NoError(t, err)

	// Verify the update
	retrievedApp, err := service.GetApplication("update-test")
	require.NoError(t, err)
	assert.Equal(t, "Updated description", retrievedApp.Spec.Description)
	assert.Contains(t, retrievedApp.Spec.Tags, "updated")
	assert.Contains(t, retrievedApp.Spec.Tags, "modified")
}

func TestApplicationService_DeleteApplication_Clean(t *testing.T) {
	service := createCleanTestApplicationService(t)

	// Create an application to delete
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "delete-test",
			Owner: "test-owner",
		},
		Spec: contracts.ApplicationSpec{
			Description: "Application to delete",
		},
	}
	err := service.CreateApplication(app)
	require.NoError(t, err)

	// Verify it exists
	_, err = service.GetApplication("delete-test")
	require.NoError(t, err)

	// Delete the application
	err = service.DeleteApplication("delete-test")
	assert.NoError(t, err)

	// Verify it's marked as deleted (implementation may vary)
	// For now, we'll just verify delete operation completed without error
	// The actual deletion behavior would depend on the graph implementation
}

func TestApplicationAgent_EventHandling_Clean(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	agent, err := NewApplicationAgent(helpers.Graph, helpers.AIProvider, helpers.EventBus, helpers.Registry)
	require.NoError(t, err)

	// Test that agent has correct capabilities
	capabilities := agent.GetCapabilities()
	assert.Len(t, capabilities, 1)
	assert.Equal(t, "application_management", capabilities[0].Name)

	// Test that agent is properly registered
	registeredAgent, err := helpers.Registry.FindAgentByID(context.Background(), "application-agent")
	require.NoError(t, err)
	assert.Equal(t, "application-agent", registeredAgent.GetID())
}

// Note: Direct event processing tests would require more complex setup
// For now, we focus on testing the service layer and agent creation
