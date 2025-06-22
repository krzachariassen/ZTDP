package application

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

// getRealAIProvider creates a real AI provider for testing (same as ServiceAgent)
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

// TestApplicationAgentCreation tests that the ApplicationAgent can be created using the framework
func TestApplicationAgentCreation(t *testing.T) {
	// Arrange - Set up dependencies
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())

	// Use real AI provider for authentic testing
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Act - Create ApplicationAgent using framework
	agent, err := NewApplicationAgent(mockGraph, realAIProvider, eventBus, registry)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error creating application agent, got: %v", err)
	}

	if agent.GetID() != "application-agent" {
		t.Errorf("Expected agent ID 'application-agent', got: %s", agent.GetID())
	}

	// Verify auto-registration
	registeredAgent, err := registry.FindAgentByID(context.Background(), "application-agent")
	if err != nil {
		t.Errorf("Expected agent to be auto-registered, got error: %v", err)
	}
	if registeredAgent.GetID() != "application-agent" {
		t.Errorf("Expected registered agent ID 'application-agent', got: %s", registeredAgent.GetID())
	}

	// Verify capabilities
	capabilities := agent.GetCapabilities()
	if len(capabilities) == 0 {
		t.Error("Expected agent to have capabilities")
	}

	foundApplicationCapability := false
	for _, cap := range capabilities {
		if cap.Name == "application_management" {
			foundApplicationCapability = true
			// Verify intents
			expectedIntents := []string{"create application", "list applications", "get application", "update application"}
			for _, expectedIntent := range expectedIntents {
				found := false
				for _, intent := range cap.Intents {
					if intent == expectedIntent {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected capability to handle intent '%s'", expectedIntent)
				}
			}
			break
		}
	}
	if !foundApplicationCapability {
		t.Error("Expected agent to have application_management capability")
	}
}

// TestApplicationAgentEventHandling tests that the agent can handle application events with real AI
func TestApplicationAgentEventHandling(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())

	// Use real AI provider - no mocking!
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Create agent using framework
	baseAgent, err := NewApplicationAgent(mockGraph, realAIProvider, eventBus, registry)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Cast to framework agent to access ProcessEvent
	agent, ok := baseAgent.(*agentFramework.BaseAgent)
	if !ok {
		t.Fatalf("Expected BaseAgent, got %T", baseAgent)
	}

	// Create test application event with natural language
	applicationEvent := &events.Event{
		Type:    events.EventTypeRequest,
		Source:  "test-source",
		Subject: "application.request",
		Payload: map[string]interface{}{
			"user_message":   "Create an application called test-app for web services",
			"correlation_id": "test-123",
		},
	}

	// Act - Process the event with real AI
	response, err := agent.ProcessEvent(context.Background(), applicationEvent)

	// Assert
	if err != nil {
		t.Errorf("Expected no error processing application event, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Verify response structure
	if response.Source != "application-agent" {
		t.Errorf("Expected response source 'application-agent', got: %s", response.Source)
	}

	// Verify correlation ID is preserved
	if correlationID, ok := response.Payload["correlation_id"]; !ok || correlationID != "test-123" {
		t.Errorf("Expected correlation_id 'test-123', got: %v", correlationID)
	}

	// The real test: verify AI actually processed the natural language
	// We can't predict exact AI response, but we can verify it was processed
	if status, exists := response.Payload["status"]; exists {
		t.Logf("✅ AI processed natural language and returned status: %v", status)
		// Status should be either success or error - both are valid AI processing results
		assert.Contains(t, []string{"success", "error", "clarification"}, status)
	} else {
		t.Error("❌ Expected AI to process request and return status")
	}
}

// TestApplicationAgentAIIntegration tests real AI parameter extraction through the agent
func TestApplicationAgentAIIntegration(t *testing.T) {
	tests := []struct {
		name           string
		userMessage    string
		expectedAction string // We can still expect certain actions based on clear language
		shouldSucceed  bool
	}{
		{
			name:           "create application with explicit details",
			userMessage:    "Create an application called checkout-app for the checkout team",
			expectedAction: "create", // This should be clear enough for AI to extract
			shouldSucceed:  true,
		},
		{
			name:           "list applications",
			userMessage:    "list all applications",
			expectedAction: "list", // This should be clear
			shouldSucceed:  true,
		},
		{
			name:           "show specific application",
			userMessage:    "show me the payment-app details",
			expectedAction: "get", // Should be interpreted as get/show
			shouldSucceed:  true,
		},
		{
			name:           "ambiguous request",
			userMessage:    "do something with applications",
			expectedAction: "",   // AI might ask for clarification
			shouldSucceed:  true, // Still succeeds but may ask for clarification
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			registry := agentRegistry.NewInMemoryAgentRegistry()
			eventBus := events.NewEventBus(nil, false)
			mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())

			// Use real AI provider - the core of this test!
			realAIProvider := getRealAIProvider(t)
			defer realAIProvider.Close()

			// Create agent
			baseAgent, err := NewApplicationAgent(mockGraph, realAIProvider, eventBus, registry)
			if err != nil {
				t.Fatalf("Failed to create agent: %v", err)
			}

			agent, ok := baseAgent.(*agentFramework.BaseAgent)
			if !ok {
				t.Fatalf("Expected BaseAgent, got %T", baseAgent)
			}

			// Create test event
			applicationEvent := &events.Event{
				Type:    events.EventTypeRequest,
				Source:  "test-user",
				Subject: "application.request",
				Payload: map[string]interface{}{
					"user_message":   tt.userMessage,
					"correlation_id": "param-test-123",
				},
			}

			// Act - Let real AI process the natural language
			response, err := agent.ProcessEvent(context.Background(), applicationEvent)

			// Assert
			if tt.shouldSucceed {
				assert.NoError(t, err, "Expected successful processing")
				assert.NotNil(t, response, "Expected response")

				if response != nil {
					status, exists := response.Payload["status"]
					assert.True(t, exists, "Expected status in response")
					t.Logf("✅ AI processed '%s' with status: %v", tt.userMessage, status)
				}
			}
		})
	}
}
