package environment

import (
	"context"
	"os"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/agentFramework"
	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/stretchr/testify/assert"
)

// getRealAIProvider creates a real AI provider for testing
func getRealAIProvider(t *testing.T) ai.AIProvider {
	// Check if we have OpenAI API key for testing
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set - skipping AI integration tests")
	}

	// Use default config which includes proper BaseURL
	config := ai.DefaultOpenAIConfig()
	config.APIKey = apiKey
	config.Model = "gpt-4" // Use a reliable model

	provider, err := ai.NewOpenAIProvider(config, apiKey)
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	return provider
}

// TestEnvironmentAgentCreationWithFramework tests that the EnvironmentAgent can be created using the framework
func TestEnvironmentAgentCreationWithFramework(t *testing.T) {
	// Arrange - Set up dependencies
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())

	// Use real AI provider for authentic testing
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Act - Create EnvironmentAgent using framework
	agent, err := NewEnvironmentAgent(mockGraph, realAIProvider, eventBus, registry)

	// Assert
	assert.NoError(t, err, "Expected no error creating framework environment agent")
	assert.Equal(t, "environment-agent", agent.GetID(), "Expected agent ID 'environment-agent'")

	// Verify auto-registration
	registeredAgent, err := registry.FindAgentByID(context.Background(), "environment-agent")
	assert.NoError(t, err, "Expected agent to be auto-registered")
	assert.Equal(t, "environment-agent", registeredAgent.GetID(), "Expected registered agent ID 'environment-agent'")

	// Verify capabilities
	capabilities := agent.GetCapabilities()
	assert.Greater(t, len(capabilities), 0, "Expected agent to have capabilities")

	foundEnvironmentCapability := false
	for _, cap := range capabilities {
		if cap.Name == "environment_management" {
			foundEnvironmentCapability = true
			// Verify intents
			expectedIntents := []string{"create environment", "list environments", "show environment", "environment management"}
			for _, expectedIntent := range expectedIntents {
				found := false
				for _, intent := range cap.Intents {
					if intent == expectedIntent {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected capability to handle intent '%s'", expectedIntent)
			}
			break
		}
	}
	assert.True(t, foundEnvironmentCapability, "Expected agent to have environment_management capability")
}

// TestEnvironmentAgentEventHandlingIntegration tests that the framework agent can handle environment events with real AI
func TestEnvironmentAgentEventHandlingIntegration(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Create agent using framework
	baseAgent, err := NewEnvironmentAgent(mockGraph, realAIProvider, eventBus, registry)
	assert.NoError(t, err, "Failed to create agent")

	// Cast to framework agent to access ProcessEvent
	agent, ok := baseAgent.(*agentFramework.BaseAgent)
	assert.True(t, ok, "Expected BaseAgent, got %T", baseAgent)

	// Test cases for different environment operations
	testCases := []struct {
		name           string
		userMessage    string
		expectedAction string
		shouldSucceed  bool
	}{
		{
			name:           "create environment",
			userMessage:    "Create an environment called production owned by devops team",
			expectedAction: "create",
			shouldSucceed:  true,
		},
		{
			name:           "list environments",
			userMessage:    "list all environments",
			expectedAction: "list",
			shouldSucceed:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test environment event
			environmentEvent := &events.Event{
				Type:    events.EventTypeRequest,
				Source:  "test-source",
				Subject: "environment.request",
				Payload: map[string]interface{}{
					"intent":         "environment management",
					"user_message":   tc.userMessage,
					"correlation_id": "test-123",
					"request_id":     "req-456",
				},
			}

			// Act - Process the event
			response, err := agent.ProcessEvent(context.Background(), environmentEvent)

			// Assert
			if tc.shouldSucceed {
				assert.NoError(t, err, "Expected no error processing environment event")
				assert.NotNil(t, response, "Expected response")

				// Verify response structure
				assert.Equal(t, "environment-agent", response.Source, "Expected response source 'environment-agent'")

				// Verify correlation ID is preserved
				correlationID, ok := response.Payload["correlation_id"]
				assert.True(t, ok, "Expected correlation_id in response")
				assert.Equal(t, "test-123", correlationID, "Expected correlation_id 'test-123'")

				// Verify request ID is preserved
				requestID, ok := response.Payload["request_id"]
				assert.True(t, ok, "Expected request_id in response")
				assert.Equal(t, "req-456", requestID, "Expected request_id 'req-456'")
			} else {
				assert.Error(t, err, "Expected error processing environment event")
			}
		})
	}
}

// TestEnvironmentAgentBusinessLogicIntegration tests that business logic works end-to-end with real AI
func TestEnvironmentAgentBusinessLogicIntegration(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)

	// Initialize global event bus for the engine
	events.InitializeEventBus(nil)

	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Create agent
	baseAgent, err := NewEnvironmentAgent(mockGraph, realAIProvider, eventBus, registry)
	assert.NoError(t, err, "Failed to create environment agent")

	agent, ok := baseAgent.(*agentFramework.BaseAgent)
	assert.True(t, ok, "Expected BaseAgent")

	// Test environment creation with full business logic
	t.Run("create environment with business logic", func(t *testing.T) {
		createEvent := &events.Event{
			Type:    events.EventTypeRequest,
			Source:  "test-user",
			Subject: "environment.request",
			Payload: map[string]interface{}{
				"intent":         "create environment",
				"user_message":   "Create an environment called production owned by devops team",
				"correlation_id": "create-test-123",
				"request_id":     "req-create-456",
			},
		}

		response, err := agent.ProcessEvent(context.Background(), createEvent)
		assert.NoError(t, err, "Expected successful environment creation")
		assert.NotNil(t, response, "Expected response")

		// Verify the response indicates success
		status, ok := response.Payload["status"]
		assert.True(t, ok, "Expected status in response")

		// Should be either "success" or "clarification" (if confidence is low)
		statusStr, ok := status.(string)
		assert.True(t, ok, "Expected status to be string")
		assert.Contains(t, []string{"success", "clarification"}, statusStr, "Expected status to be success or clarification")

		// If successful, verify environment details
		if statusStr == "success" {
			envName, ok := response.Payload["environment_name"]
			assert.True(t, ok, "Expected environment_name in response")
			assert.Equal(t, "production", envName, "Expected environment name 'production'")
		}
	})

	// Test environment listing
	t.Run("list environments", func(t *testing.T) {
		listEvent := &events.Event{
			Type:    events.EventTypeRequest,
			Source:  "test-user",
			Subject: "environment.request",
			Payload: map[string]interface{}{
				"intent":         "list environments",
				"user_message":   "list all environments",
				"correlation_id": "list-test-123",
				"request_id":     "req-list-456",
			},
		}

		response, err := agent.ProcessEvent(context.Background(), listEvent)
		assert.NoError(t, err, "Expected successful environment listing")
		assert.NotNil(t, response, "Expected response")

		// Verify the response structure
		status, ok := response.Payload["status"]
		assert.True(t, ok, "Expected status in response")

		statusStr, ok := status.(string)
		assert.True(t, ok, "Expected status to be string")
		assert.Contains(t, []string{"success", "clarification"}, statusStr, "Expected status to be success or clarification")

		// If successful, verify environments list
		if statusStr == "success" {
			_, ok := response.Payload["environments"]
			assert.True(t, ok, "Expected environments in response")

			count, ok := response.Payload["count"]
			assert.True(t, ok, "Expected count in response")

			countInt, ok := count.(int)
			assert.True(t, ok, "Expected count to be integer")
			assert.GreaterOrEqual(t, countInt, 0, "Expected count to be non-negative")
		}
	})
}

// TestEnvironmentAgentParameterExtraction tests the AI parameter extraction in the agent context
func TestEnvironmentAgentParameterExtraction(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Create agent
	baseAgent, err := NewEnvironmentAgent(mockGraph, realAIProvider, eventBus, registry)
	assert.NoError(t, err, "Failed to create environment agent")

	agent, ok := baseAgent.(*agentFramework.BaseAgent)
	assert.True(t, ok, "Expected BaseAgent")

	// Test parameter extraction for different environment operations
	testCases := []struct {
		name               string
		userMessage        string
		expectedAction     string
		expectedConfidence float64
	}{
		{
			name:               "create environment with details",
			userMessage:        "Create an environment called production owned by devops team",
			expectedAction:     "create",
			expectedConfidence: 0.95,
		},
		{
			name:               "list environments",
			userMessage:        "list all environments",
			expectedAction:     "list",
			expectedConfidence: 0.9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &events.Event{
				Type:    events.EventTypeRequest,
				Source:  "test-user",
				Subject: "environment.request",
				Payload: map[string]interface{}{
					"intent":         "environment management",
					"user_message":   tc.userMessage,
					"correlation_id": "param-test-123",
					"request_id":     "req-param-456",
				},
			}

			response, err := agent.ProcessEvent(context.Background(), event)
			assert.NoError(t, err, "Expected successful parameter extraction")
			assert.NotNil(t, response, "Expected response")

			// The response should reflect the extracted parameters through business logic
			status, ok := response.Payload["status"]
			assert.True(t, ok, "Expected status in response")

			statusStr, ok := status.(string)
			assert.True(t, ok, "Expected status to be string")

			// With high confidence, we should get success, not clarification
			if tc.expectedConfidence >= 0.7 {
				assert.NotEqual(t, "clarification", statusStr, "Expected high confidence to not require clarification")
			}
		})
	}
}

// TestEnvironmentAgentEndToEndScenario tests the exact scenarios from the integration test
func TestEnvironmentAgentEndToEndScenario(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Create agent
	baseAgent, err := NewEnvironmentAgent(mockGraph, realAIProvider, eventBus, registry)
	assert.NoError(t, err, "Failed to create environment agent")

	agent, ok := baseAgent.(*agentFramework.BaseAgent)
	assert.True(t, ok, "Expected BaseAgent")

	// Test cases that match the exact integration test scenarios
	testCases := []struct {
		name        string
		userMessage string
		expectName  string
		expectOwner string
	}{
		{
			name:        "Create dev environment",
			userMessage: "Create a development environment called dev owned by platform-team for development work.",
			expectName:  "development", // Should be normalized from "dev" to canonical "development"
			expectOwner: "platform-team",
		},
		{
			name:        "Create staging environment",
			userMessage: "Create a staging environment for testing before production.",
			expectName:  "staging", // Should be inferred
			expectOwner: "",        // Not specified
		},
		{
			name:        "Create production environment",
			userMessage: "Create a production environment with strict policies for live workloads.",
			expectName:  "production", // Should be inferred
			expectOwner: "",           // Not specified
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create event that matches what the integration test sends
			environmentEvent := &events.Event{
				Type:    events.EventTypeRequest,
				Source:  "ai-chat",
				Subject: "environment.request",
				Payload: map[string]interface{}{
					"intent":         "environment management",
					"user_message":   tc.userMessage,
					"correlation_id": "integration-test-123",
					"request_id":     "req-integration-456",
				},
			}

			// Act - Process the event (this is what the agent framework does)
			response, err := agent.ProcessEvent(context.Background(), environmentEvent)

			// Assert - Debug the actual response to understand why it's failing
			t.Logf("🔍 User Message: %s", tc.userMessage)
			t.Logf("🔍 Response Error: %v", err)
			if response != nil {
				t.Logf("🔍 Response Type: %s", response.Type)
				t.Logf("🔍 Response Subject: %s", response.Subject)
				t.Logf("🔍 Response Payload: %+v", response.Payload)

				if status, ok := response.Payload["status"]; ok {
					t.Logf("🔍 Status: %v", status)
				}
				if message, ok := response.Payload["message"]; ok {
					t.Logf("🔍 Message: %v", message)
				}
				if envName, ok := response.Payload["environment_name"]; ok {
					t.Logf("🔍 Extracted Environment Name: %v", envName)
				}
			}

			// The integration test expects success, so let's see what we actually get
			assert.NoError(t, err, "Expected no error processing environment event")
			assert.NotNil(t, response, "Expected response")

			// Check if we got an error status (which would explain the integration test failure)
			if response != nil {
				if status, ok := response.Payload["status"].(string); ok {
					if status == "error" {
						if message, ok := response.Payload["message"].(string); ok {
							t.Errorf("❌ Integration test failure reproduced: %s", message)
							t.Logf("💡 This explains why the end-to-end test is failing!")
						}
					} else if status == "success" {
						t.Logf("✅ Success! Environment name should be: %v", response.Payload["environment_name"])
						if tc.expectName != "" {
							assert.Equal(t, tc.expectName, response.Payload["environment_name"], "Expected correct environment name extraction")
						}
					}
				}
			}
		})
	}
}
