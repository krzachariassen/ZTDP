package policies

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/agentFramework"
	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

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

// Helper function to create test policy agent using framework
func createTestFrameworkPolicyAgent(t *testing.T) agentRegistry.AgentInterface {
	// Create test components
	mockPolicyStore := &MockPolicyStore{policies: make(map[string]*Policy)}
	realEventBus := events.NewEventBus(nil, false)
	registry := agentRegistry.NewInMemoryAgentRegistry()

	// Create memory graph
	backend := graph.NewMemoryGraph()
	globalGraph := graph.NewGlobalGraph(backend)

	// Create agent using new framework
	policyAgent, err := NewPolicyAgent(nil, globalGraph, mockPolicyStore, realEventBus, registry)
	if err != nil {
		t.Fatalf("Failed to create framework policy agent: %v", err)
	}

	return policyAgent
}
