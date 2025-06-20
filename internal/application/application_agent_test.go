package application

import (
	"context"
	"testing"
	"time"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper to create a test application service with real AI
func createTestApplicationService(t *testing.T) *Service {
	helpers := CreateTestHelpers(t)
	return helpers.CreateTestApplicationService(t)
}

func TestNewApplicationAgent(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	agent, err := NewApplicationAgent(helpers.Graph, helpers.AIProvider, helpers.EventBus, helpers.Registry)
	require.NoError(t, err)
	assert.NotNil(t, agent)
}

func TestApplicationAgent_BasicLifecycle(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	agent, err := NewApplicationAgent(helpers.Graph, helpers.AIProvider, helpers.EventBus, helpers.Registry)
	require.NoError(t, err)

	// For interface testing, we'll focus on the agent creation
	assert.NotNil(t, agent)
}

func TestApplicationService_CreateApplication(t *testing.T) {
	service := createTestApplicationService(t)

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
					Tags:        []string{"web"},
				},
			},
			expectError: true,
			errorMsg:    "name is required",
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
				apps, err := service.ListApplications()
				assert.NoError(t, err)
				assert.Len(t, apps, 1)
				assert.Equal(t, tt.app.Metadata.Name, apps[0].Metadata.Name)
			}
		})
	}
}

func TestApplicationService_ServiceManagement(t *testing.T) {
	service := createTestApplicationService(t)

	// First create an application
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "test-app",
			Owner: "test-owner",
		},
		Spec: contracts.ApplicationSpec{
			Description: "Test application",
			Tags:        []string{"web"},
		},
	}
	err := service.CreateApplication(app)
	require.NoError(t, err)

	tests := []struct {
		name        string
		appName     string
		serviceName string
		serviceType string
		description string
		port        int
		replicas    int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid service",
			appName:     "test-app",
			serviceName: "api-service",
			serviceType: "web",
			description: "API service",
			port:        8080,
			replicas:    3,
			expectError: false,
		},
		{
			name:        "service for non-existent app",
			appName:     "non-existent-app",
			serviceName: "api-service",
			serviceType: "web",
			description: "API service",
			port:        8080,
			replicas:    1,
			expectError: true,
			errorMsg:    "application not found",
		},
		{
			name:        "service with empty name",
			appName:     "test-app",
			serviceName: "",
			serviceType: "web",
			description: "API service",
			port:        8080,
			replicas:    1,
			expectError: true,
			errorMsg:    "service name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateApplicationService(tt.appName, tt.serviceName, tt.serviceType, tt.description, tt.port, tt.replicas)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify service was created
				services, err := service.ListApplicationServices(tt.appName)
				assert.NoError(t, err)
				assert.Len(t, services, 1)
				assert.Equal(t, tt.serviceName, services[0].Name)
				assert.Equal(t, tt.port, services[0].Port)
				assert.Equal(t, tt.replicas, services[0].Replicas)
			}
		})
	}
}

func TestApplicationService_ReleaseManagement(t *testing.T) {
	service := createTestApplicationService(t)

	// Create application and service
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "test-app",
			Owner: "test-owner",
		},
		Spec: contracts.ApplicationSpec{
			Description: "Test application",
			Tags:        []string{"web"},
		},
	}
	err := service.CreateApplication(app)
	require.NoError(t, err)

	err = service.CreateApplicationService("test-app", "api-service", "web", "API service", 8080, 1)
	require.NoError(t, err)

	tests := []struct {
		name        string
		appName     string
		environment string
		version     string
		services    map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid release",
			appName:     "test-app",
			environment: "staging",
			version:     "1.0.0",
			services:    map[string]string{"api-service": "1.0.0"},
			expectError: false,
		},
		{
			name:        "release for non-existent app",
			appName:     "non-existent-app",
			environment: "staging",
			version:     "1.0.0",
			services:    map[string]string{"api-service": "1.0.0"},
			expectError: true,
			errorMsg:    "application not found",
		},
		{
			name:        "release with empty version",
			appName:     "test-app",
			environment: "staging",
			version:     "",
			services:    map[string]string{"api-service": "1.0.0"},
			expectError: true,
			errorMsg:    "release version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			release, err := service.CreateApplicationRelease(tt.appName, tt.environment, tt.version, tt.services)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, release)
				assert.Equal(t, tt.version, release.Version)
				assert.Equal(t, tt.appName, release.Application)
				assert.Equal(t, tt.environment, release.Environment)

				// Verify release was created
				releases, err := service.ListApplicationReleases(tt.appName)
				assert.NoError(t, err)
				assert.Len(t, releases, 1)
				assert.Equal(t, tt.version, releases[0].Version)
			}
		})
	}
}

func TestApplicationService_CompleteUserFlow(t *testing.T) {
	service := createTestApplicationService(t)

	// 1. Create an application
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "checkout",
			Owner: "test-owner",
		},
		Spec: contracts.ApplicationSpec{
			Description: "Checkout service application",
			Tags:        []string{"microservice"},
		},
	}
	err := service.CreateApplication(app)
	require.NoError(t, err)

	// 2. Create services for the application
	err = service.CreateApplicationService("checkout", "checkout-api", "web", "REST API service", 8080, 3)
	require.NoError(t, err)

	err = service.CreateApplicationService("checkout", "checkout-worker", "worker", "Background worker service", 0, 2)
	require.NoError(t, err)

	// 3. Verify services were created
	services, err := service.ListApplicationServices("checkout")
	require.NoError(t, err)
	assert.Len(t, services, 2)

	// 4. Create service versions
	apiVersion, err := service.CreateServiceVersion("checkout", "checkout-api", "1.0.0", "checkout-api:1.0.0")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", apiVersion.Version)

	workerVersion, err := service.CreateServiceVersion("checkout", "checkout-worker", "1.0.0", "checkout-worker:1.0.0")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", workerVersion.Version)

	// 5. Create a release
	serviceVersions := map[string]string{
		"checkout-api":    "1.0.0",
		"checkout-worker": "1.0.0",
	}
	release, err := service.CreateApplicationRelease("checkout", "production", "v1.0.0", serviceVersions)
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", release.Version)
	assert.Equal(t, "production", release.Environment)
	assert.Equal(t, serviceVersions, release.Services)

	// 6. Verify the complete setup
	releases, err := service.ListApplicationReleases("checkout")
	require.NoError(t, err)
	assert.Len(t, releases, 1)
	assert.Equal(t, "v1.0.0", releases[0].Version)

	// 7. Verify application can be retrieved
	retrievedApp, err := service.GetApplication("checkout")
	require.NoError(t, err)
	assert.Equal(t, "checkout", retrievedApp.Metadata.Name)
	assert.Contains(t, retrievedApp.Spec.Tags, "microservice")
}

func TestApplicationAgent_IntegratedDomainCapabilities(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	agent, err := NewApplicationAgent(helpers.Graph, helpers.AIProvider, helpers.EventBus, helpers.Registry)
	require.NoError(t, err)
	assert.NotNil(t, agent)

	// Test that the agent has all the correct capabilities
	capabilities := agent.GetCapabilities()
	assert.Len(t, capabilities, 4) // application, service, environment, release management

	capabilityNames := make([]string, len(capabilities))
	for i, cap := range capabilities {
		capabilityNames[i] = cap.Name
	}

	assert.Contains(t, capabilityNames, "application_management")
	assert.Contains(t, capabilityNames, "service_management")
	assert.Contains(t, capabilityNames, "environment_management")
	assert.Contains(t, capabilityNames, "release_management")

	// Test that the agent starts and stops correctly
	ctx := context.Background()
	err = agent.Start(ctx)
	assert.NoError(t, err)

	status := agent.GetStatus()
	assert.Equal(t, "application-agent", status.ID)
	assert.Equal(t, "application", status.Type)

	err = agent.Stop(ctx)
	assert.NoError(t, err)
}

func TestApplicationAgent_CrossDomainIntegration(t *testing.T) {
	// Test the integrated services directly since we have access to them
	service := createTestApplicationService(t)

	t.Run("Complete Application Lifecycle", func(t *testing.T) {
		// 1. Create Application
		app := contracts.ApplicationContract{
			Metadata: contracts.Metadata{
				Name:  "integration-app",
				Owner: "test-user",
			},
			Spec: contracts.ApplicationSpec{
				Description: "Integration test application",
				Tags:        []string{"integration", "test"},
			},
		}
		err := service.CreateApplication(app)
		require.NoError(t, err)

		// 2. Create Services
		err = service.CreateApplicationService("integration-app", "api-service", "web", "API service", 8080, 3)
		require.NoError(t, err)

		err = service.CreateApplicationService("integration-app", "worker-service", "worker", "Background worker", 0, 2)
		require.NoError(t, err)

		// 3. Create Service Versions
		apiVersion, err := service.CreateServiceVersion("integration-app", "api-service", "1.0.0", "integration-app/api:1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", apiVersion.Version)

		workerVersion, err := service.CreateServiceVersion("integration-app", "worker-service", "1.0.0", "integration-app/worker:1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", workerVersion.Version)

		// 4. Create Release
		serviceVersions := map[string]string{
			"api-service":    "1.0.0",
			"worker-service": "1.0.0",
		}
		release, err := service.CreateApplicationRelease("integration-app", "production", "v1.0.0", serviceVersions)
		require.NoError(t, err)
		assert.Equal(t, "v1.0.0", release.Version)
		assert.Equal(t, "production", release.Environment)

		// 5. Verify all components exist
		apps, err := service.ListApplications()
		require.NoError(t, err)
		assert.Len(t, apps, 1)
		assert.Equal(t, "integration-app", apps[0].Metadata.Name)

		services, err := service.ListApplicationServices("integration-app")
		require.NoError(t, err)
		assert.Len(t, services, 2)

		releases, err := service.ListApplicationReleases("integration-app")
		require.NoError(t, err)
		assert.Len(t, releases, 1)
		assert.Equal(t, "v1.0.0", releases[0].Version)

		// 6. Test retrieval
		retrievedApp, err := service.GetApplication("integration-app")
		require.NoError(t, err)
		assert.Equal(t, "integration-app", retrievedApp.Metadata.Name)
		assert.Contains(t, retrievedApp.Spec.Tags, "integration")
	})
}

// TestApplicationAgentOrchestrationIntegration tests the critical orchestrator integration
// This is what we missed in our TDD - agents must respond to orchestrator intents!
func TestApplicationAgentOrchestrationIntegration(t *testing.T) {
	tests := []struct {
		name               string
		routingKey         string
		intent             string
		expectedResponse   string
		shouldEmitResponse bool
	}{
		{
			name:               "should handle list applications intent from orchestrator",
			routingKey:         "application.request",
			intent:             "list applications",
			expectedResponse:   "Applications retrieved successfully",
			shouldEmitResponse: true,
		},
		{
			name:               "should handle create application intent from orchestrator",
			routingKey:         "application.request",
			intent:             "create application",
			expectedResponse:   "Application creation completed",
			shouldEmitResponse: true,
		},
		{
			name:               "should handle list environments intent from orchestrator",
			routingKey:         "environment.request",
			intent:             "list environments",
			expectedResponse:   "Environments retrieved successfully",
			shouldEmitResponse: true,
		},
		{
			name:               "should handle create environment intent from orchestrator",
			routingKey:         "environment.request",
			intent:             "create environment",
			expectedResponse:   "Environment creation completed",
			shouldEmitResponse: true,
		},
		{
			name:               "should handle list services intent from orchestrator",
			routingKey:         "service.request",
			intent:             "list services",
			expectedResponse:   "Services retrieved successfully",
			shouldEmitResponse: true,
		},
		{
			name:               "should handle create service intent from orchestrator",
			routingKey:         "service.request",
			intent:             "create service",
			expectedResponse:   "Service creation completed",
			shouldEmitResponse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup using test helpers
			helpers := CreateTestHelpers(t)
			defer helpers.CleanupTestData(t)

			agent, err := NewApplicationAgent(helpers.Graph, helpers.AIProvider, helpers.EventBus, helpers.Registry)
			require.NoError(t, err)
			require.NotNil(t, agent)

			// Start the agent
			err = agent.Start(context.Background())
			require.NoError(t, err)

			// Simulate orchestrator request (this is what was missing!)
			correlationID := "orchestration-test-123"

			// Track response events
			responseReceived := false
			var responseMessage string
			var responseCorrelationID string

			// Subscribe to response events (what orchestrator expects!)
			helpers.EventBus.Subscribe(events.EventTypeResponse, func(event events.Event) error {
				// Check if this response is for our request (same pattern as orchestrator)
				if respCorrelationID, ok := event.Payload["correlation_id"].(string); ok {
					if respCorrelationID == correlationID {
						responseReceived = true

						// Extract the actual message from the response payload
						if status, exists := event.Payload["status"].(string); exists && status == "success" {
							if msg, msgExists := event.Payload["message"].(string); msgExists {
								responseMessage = msg
							} else {
								responseMessage = "Success response without message"
							}
						} else if errorMsg, exists := event.Payload["error"].(string); exists {
							responseMessage = "Error: " + errorMsg
						} else {
							responseMessage = "Unknown response format"
						}

						responseCorrelationID = respCorrelationID
					}
				}
				return nil
			})

			requestPayload := map[string]interface{}{
				"correlation_id": correlationID,
				"intent":         tt.intent, // This is key - orchestrator sends INTENT not ACTION!
				"request_id":     "req-test-123",
				"source_agent":   "orchestrator",
				"context": map[string]interface{}{
					"source":       "orchestrator-chat",
					"user_message": "Test message for " + tt.intent,
				},
			}

			// Emit the request event with proper signature
			err = helpers.EventBus.Emit(events.EventTypeRequest, "orchestrator", tt.routingKey, requestPayload)
			require.NoError(t, err)

			// Wait for response (this should NOT timeout!)
			require.Eventually(t, func() bool {
				return responseReceived
			}, 5*time.Second, 100*time.Millisecond, "Agent should respond to orchestrator intent within timeout")

			// Validate response
			assert.True(t, responseReceived, "Agent must respond to orchestrator requests")
			assert.Contains(t, responseMessage, tt.expectedResponse, "Response should contain expected message")
			assert.Equal(t, correlationID, responseCorrelationID, "Correlation ID must match for orchestrator tracking")
		})
	}
}

// TestApplicationAgentIntentHandling specifically tests intent processing (not action)
func TestApplicationAgentIntentHandling(t *testing.T) {
	helpers := CreateTestHelpers(t)
	defer helpers.CleanupTestData(t)

	_, err := NewApplicationAgent(helpers.Graph, helpers.AIProvider, helpers.EventBus, helpers.Registry)
	require.NoError(t, err)

	// Test that agent handles intent field correctly
	t.Run("should process intent field from orchestrator", func(t *testing.T) {
		// This simulates what orchestrator actually sends
		requestPayload := map[string]interface{}{
			"intent":         "list applications", // Orchestrator sends INTENT
			"correlation_id": "test-correlation",
			"request_id":     "test-request",
			"context": map[string]interface{}{
				"source":       "orchestrator-chat",
				"user_message": "List all applications",
			},
		}

		// This should NOT panic or fail - agent must handle intent-based requests
		require.NotPanics(t, func() {
			// Simulate the event processing
			err := helpers.EventBus.Emit(events.EventTypeRequest, "orchestrator", "application.request", requestPayload)
			require.NoError(t, err)
		})
	})

	t.Run("should handle missing action field gracefully", func(t *testing.T) {
		// Old tests might send "action" but orchestrator sends "intent"
		requestPayload := map[string]interface{}{
			"intent":         "create application", // No "action" field!
			"correlation_id": "test-correlation",
			"request_id":     "test-request",
		}

		// This should work - agents must be flexible with orchestrator format
		require.NotPanics(t, func() {
			err := helpers.EventBus.Emit(events.EventTypeRequest, "orchestrator", "application.request", requestPayload)
			require.NoError(t, err)
		})
	})
}
