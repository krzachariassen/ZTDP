package policies

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

// Test framework-based PolicyAgent implementation
func TestFrameworkPolicyAgent_AgentInterface(t *testing.T) {
	t.Run("implements agentRegistry.AgentInterface correctly", func(t *testing.T) {
		// Setup
		policyAgent := createTestFrameworkPolicyAgent(t)

		// Test AgentInterface compliance
		var _ agentRegistry.AgentInterface = policyAgent

		// Test capabilities
		capabilities := policyAgent.GetCapabilities()
		if len(capabilities) == 0 {
			t.Error("Policy agent should have capabilities")
		}

		// Verify policy-specific capabilities
		hasEvaluation := false
		for _, cap := range capabilities {
			if cap.Name == "policy_evaluation" {
				hasEvaluation = true
				break
			}
		}
		if !hasEvaluation {
			t.Error("Policy agent should have policy_evaluation capability")
		}

		// Test status
		status := policyAgent.GetStatus()
		if status.Type != "policy" {
			t.Errorf("Expected agent type 'policy', got %s", status.Type)
		}
	})
}

// Test event-based policy evaluation using framework
func TestFrameworkPolicyAgent_EventBasedEvaluation(t *testing.T) {
	tests := []struct {
		name           string
		event          *events.Event
		expectedStatus string
		shouldError    bool
	}{
		{
			name: "policy evaluation with node data",
			event: &events.Event{
				ID:      "test-event-1",
				Type:    events.EventTypeRequest,
				Subject: "policy.evaluation",
				Source:  "test",
				Payload: map[string]interface{}{
					"intent": "evaluate policy",
					"node": map[string]interface{}{
						"id":   "test-node",
						"kind": "application",
						"metadata": map[string]interface{}{
							"name": "test-app",
						},
					},
				},
				Timestamp: 1234567890,
			},
			expectedStatus: "success",
			shouldError:    false,
		},
		{
			name: "compliance check request",
			event: &events.Event{
				ID:      "test-event-2",
				Type:    events.EventTypeRequest,
				Subject: "compliance.check",
				Source:  "test",
				Payload: map[string]interface{}{
					"intent": "check compliance",
					"node": map[string]interface{}{
						"id":   "test-node",
						"kind": "service",
						"metadata": map[string]interface{}{
							"name": "test-service",
						},
					},
				},
				Timestamp: 1234567890,
			},
			expectedStatus: "success",
			shouldError:    false,
		},
		{
			name: "missing intent should error",
			event: &events.Event{
				ID:      "test-event-3",
				Type:    events.EventTypeRequest,
				Subject: "policy.evaluation",
				Source:  "test",
				Payload: map[string]interface{}{
					// No intent field
					"node": map[string]interface{}{
						"id":   "test-node",
						"kind": "application",
					},
				},
				Timestamp: 1234567890,
			},
			expectedStatus: "error",
			shouldError:    false, // Error response is valid, not a failure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create agent using framework
			baseAgent := createTestFrameworkPolicyAgent(t)

			// Cast to framework agent to access ProcessEvent
			agent, ok := baseAgent.(*agentFramework.BaseAgent)
			if !ok {
				t.Fatalf("Expected *agentFramework.BaseAgent, got %T", baseAgent)
			}

			// Process the event
			response, err := agent.ProcessEvent(context.Background(), tt.event)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response == nil {
				t.Fatal("Expected response but got nil")
			}

			// Check response status
			status, ok := response.Payload["status"].(string)
			if !ok {
				t.Error("Expected status field in response")
			}

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}

			// Log response for debugging
			t.Logf("🤖 Policy Agent Response:")
			t.Logf("   Status: %s", status)
			t.Logf("   Subject: %s", response.Subject)
			if msg, exists := response.Payload["message"]; exists {
				t.Logf("   Message: %v", msg)
			}
		})
	}
}

// Test business logic integration
func TestFrameworkPolicyAgent_BusinessLogicIntegration(t *testing.T) {
	t.Run("integrates with policy service correctly", func(t *testing.T) {
		// Create agent with policy service
		baseAgent := createTestFrameworkPolicyAgent(t)

		// Cast to framework agent
		agent, ok := baseAgent.(*agentFramework.BaseAgent)
		if !ok {
			t.Fatalf("Expected *agentFramework.BaseAgent, got %T", baseAgent)
		}

		// Create event with real node data
		event := &events.Event{
			ID:      "business-logic-test",
			Type:    events.EventTypeRequest,
			Subject: "policy.evaluation",
			Source:  "test",
			Payload: map[string]interface{}{
				"intent": "evaluate policy",
				"node": map[string]interface{}{
					"id":   "test-app",
					"kind": "application",
					"metadata": map[string]interface{}{
						"name": "test-application",
					},
				},
			},
			Timestamp: 1234567890,
		}

		// Process event
		response, err := agent.ProcessEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Business logic integration failed: %v", err)
		}

		// Verify response structure
		if response == nil {
			t.Fatal("Expected response from business logic")
		}

		status, ok := response.Payload["status"].(string)
		if !ok {
			t.Error("Expected status in business logic response")
		}

		if status != "success" {
			t.Errorf("Expected successful business logic execution, got status: %s", status)
		}

		// Check that policy evaluation data is included
		if _, hasDecision := response.Payload["decision"]; !hasDecision {
			t.Error("Expected decision in business logic response")
		}
	})
}

// TestPolicyAgentEndToEndScenario tests the policy agent with real AI scenarios
func TestPolicyAgentEndToEndScenario(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	backend := graph.NewMemoryGraph()
	globalGraph := graph.NewGlobalGraph(backend)
	graphStore := &graph.GraphStore{}
	mockPolicyStore := &MockPolicyStore{policies: make(map[string]*Policy)}

	// Create agent with real AI (created internally)
	baseAgent, err := NewPolicyAgent(graphStore, globalGraph, mockPolicyStore, eventBus, registry)
	assert.NoError(t, err, "Failed to create policy agent")

	agent, ok := baseAgent.(*agentFramework.BaseAgent)
	assert.True(t, ok, "Expected BaseAgent")

	// Test cases that require real AI processing
	testCases := []struct {
		name        string
		userMessage string
		expectType  string
	}{
		{
			name:        "Policy evaluation request",
			userMessage: "Evaluate deployment policy for app test-app to environment production",
			expectType:  "policy_evaluation",
		},
		{
			name:        "Compliance check request",
			userMessage: "Check if deployment of sensitive-app to production violates security policies",
			expectType:  "compliance_check",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create event for policy evaluation
			policyEvent := &events.Event{
				Type:    events.EventTypeRequest,
				Source:  "ai-chat",
				Subject: "policy.evaluation",
				Payload: map[string]interface{}{
					"intent":         "policy evaluation",
					"user_message":   tc.userMessage,
					"correlation_id": "policy-test-123",
					"request_id":     "req-policy-456",
				},
			}

			// Act - Process the event (real AI will be used internally)
			response, err := agent.ProcessEvent(context.Background(), policyEvent)

			// Assert
			t.Logf("🔍 User Message: %s", tc.userMessage)
			t.Logf("🔍 Response Error: %v", err)
			if response != nil {
				t.Logf("🔍 Response Subject: %s", response.Subject)
				t.Logf("🔍 Response Payload: %+v", response.Payload)

				if status, ok := response.Payload["status"]; ok {
					t.Logf("🔍 Status: %v", status)
				}
			}

			// The policy agent should process the request (success, error, or clarification)
			assert.NoError(t, err, "Expected no error processing policy event")
			assert.NotNil(t, response, "Expected response")

			// Verify response structure
			if response != nil {
				if status, ok := response.Payload["status"].(string); ok {
					assert.Contains(t, []string{"success", "error", "clarification"}, status,
						"Expected valid status from policy processing")
					t.Logf("✅ Policy agent processed with status: %s", status)
				}
			}
		})
	}
}

// Helper function to create test policy agent using framework
func createTestFrameworkPolicyAgent(t *testing.T) agentRegistry.AgentInterface {
	// Create test components
	mockPolicyStore := &MockPolicyStore{policies: make(map[string]*Policy)}
	realEventBus := events.NewEventBus(nil, false)
	registry := agentRegistry.NewInMemoryAgentRegistry()

	// Create memory graph and graph store
	backend := graph.NewMemoryGraph()
	globalGraph := graph.NewGlobalGraph(backend)
	graphStore := &graph.GraphStore{} // Create a graph store for the policy agent

	// Create agent using new framework (AI provider will be created internally)
	policyAgent, err := NewPolicyAgent(graphStore, globalGraph, mockPolicyStore, realEventBus, registry)
	if err != nil {
		t.Fatalf("Failed to create framework policy agent: %v", err)
	}

	return policyAgent
}
