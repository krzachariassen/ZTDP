package policies

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// Test PolicyAgent implementation of AgentInterface
func TestPolicyAgent_AgentInterface(t *testing.T) {
	t.Run("implements AgentInterface correctly", func(t *testing.T) {
		// Setup
		policyAgent := createTestPolicyAgent(t)

		// Test AgentInterface compliance
		var _ agents.AgentInterface = policyAgent

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

// Test event-based policy evaluation
func TestPolicyAgent_EventBasedEvaluation(t *testing.T) {
	tests := []struct {
		name           string
		eventType      events.EventType
		eventSubject   string
		eventPayload   map[string]interface{}
		expectResponse bool
		expectDecision string
	}{
		{
			name:         "Policy evaluation request",
			eventType:    events.EventTypeRequest,
			eventSubject: "evaluate policy for deployment",
			eventPayload: map[string]interface{}{
				"intent":      "Check if deploying app-x to production violates policies",
				"application": "app-x",
				"environment": "production",
				"action":      "deployment",
			},
			expectResponse: true,
			expectDecision: "allowed", // or "blocked" depending on policy
		},
		{
			name:         "Policy compliance check",
			eventType:    events.EventTypeRequest,
			eventSubject: "check compliance",
			eventPayload: map[string]interface{}{
				"intent": "Validate application configuration against security policies",
				"application_config": map[string]interface{}{
					"name":     "test-app",
					"services": []string{"api", "db"},
				},
			},
			expectResponse: true,
		},
		{
			name:           "Invalid request type",
			eventType:      events.EventTypeBroadcast,
			eventSubject:   "broadcast message",
			eventPayload:   map[string]interface{}{},
			expectResponse: true, // Should still respond but indicate not handled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			policyAgent := createTestPolicyAgent(t)

			// Create event
			event := &events.Event{
				Type:    tt.eventType,
				Source:  "test-coordinator",
				Subject: tt.eventSubject,
				Payload: tt.eventPayload,
			}

			// Process event
			response, err := policyAgent.ProcessEvent(context.Background(), event)

			if tt.expectResponse {
				if err != nil {
					t.Fatalf("Expected successful response, got error: %v", err)
				}
				if response == nil {
					t.Fatal("Expected response, got nil")
				}

				// Verify response structure
				if response.Type != events.EventTypeResponse {
					t.Errorf("Expected response type %s, got %s", events.EventTypeResponse, response.Type)
				}

				if response.Source != policyAgent.GetStatus().ID {
					t.Errorf("Expected response source %s, got %s", policyAgent.GetStatus().ID, response.Source)
				}

				// Check for decision in payload (for policy evaluation requests)
				if tt.expectDecision != "" {
					if decision, ok := response.Payload["decision"].(string); ok {
						// Decision should be one of the valid policy statuses
						validDecisions := []string{"allowed", "blocked", "conditional", "warning"}
						found := false
						for _, valid := range validDecisions {
							if decision == valid {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Invalid decision %s, must be one of %v", decision, validDecisions)
						}
					}
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid request")
				}
			}
		})
	}
}

// Test integration with existing policy system
func TestPolicyAgent_Integration(t *testing.T) {
	t.Run("integrates with existing policy evaluation", func(t *testing.T) {
		// Setup
		policyAgent := createTestPolicyAgent(t)

		// Create a policy evaluation request that matches existing system
		event := &events.Event{
			Type:    events.EventTypeRequest,
			Source:  "platform-agent",
			Subject: "policy evaluation",
			Payload: map[string]interface{}{
				"intent": "Evaluate node policy for application service limit",
				"node": map[string]interface{}{
					"id":   "test-app",
					"kind": "application",
					"metadata": map[string]interface{}{
						"name": "Over-Complicated Application",
					},
					"spec": map[string]interface{}{
						"services": []string{
							"service-1", "service-2", "service-3", "service-4", "service-5",
							"service-6", "service-7", "service-8", "service-9", "service-10",
							"service-11", "service-12", "service-13", "service-14", "service-15",
						},
					},
				},
				"policy": map[string]interface{}{
					"id":                    "app-service-limit",
					"name":                  "Application Service Limit",
					"natural_language_rule": "Applications must have fewer than 10 services",
				},
			},
		}

		// Process event
		response, err := policyAgent.ProcessEvent(context.Background(), event)

		// Verify response
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		if response == nil {
			t.Fatal("Expected policy evaluation response")
		}

		// Should detect violation (15 services > 10 limit)
		if decision, ok := response.Payload["decision"].(string); ok {
			if decision != "blocked" {
				t.Errorf("Expected 'blocked' decision for service limit violation, got %s", decision)
			}
		} else {
			t.Error("Response should contain decision field")
		}

		// Should include reasoning
		if reasoning, ok := response.Payload["reasoning"].(string); ok {
			if reasoning == "" {
				t.Error("Response should include reasoning")
			}
		} else {
			t.Error("Response should contain reasoning field")
		}
	})
}

// Test agent lifecycle
func TestPolicyAgent_Lifecycle(t *testing.T) {
	t.Run("starts and stops correctly", func(t *testing.T) {
		// Setup
		policyAgent := createTestPolicyAgent(t)

		// Test initial state
		status := policyAgent.GetStatus()
		if status.Status == "running" {
			t.Error("Agent should not be running initially")
		}

		// Test start
		err := policyAgent.Start(context.Background())
		if err != nil {
			t.Fatalf("Failed to start agent: %v", err)
		}

		status = policyAgent.GetStatus()
		if status.Status != "running" {
			t.Errorf("Expected status 'running', got %s", status.Status)
		}

		// Test health
		health := policyAgent.Health()
		if !health.Healthy {
			t.Error("Running agent should be healthy")
		}

		// Test stop
		err = policyAgent.Stop(context.Background())
		if err != nil {
			t.Fatalf("Failed to stop agent: %v", err)
		}

		status = policyAgent.GetStatus()
		if status.Status != "stopped" {
			t.Errorf("Expected status 'stopped', got %s", status.Status)
		}
	})
}

// createTestPolicyAgent creates a policy agent for testing
func createTestPolicyAgent(t *testing.T) agents.AgentInterface {
	// Use real components for testing
	mockPolicyStore := &MockPolicyStore{}
	realEventBus := events.NewEventBus(nil, false) // Use real EventBus

	// Create an adapter to match the EventBus interface expected by PolicyAgent
	eventBusAdapter := &EventBusAdapter{realEventBus}

	// Create mock components that the PolicyAgent needs
	backend := graph.NewMemoryGraph()
	globalGraph := graph.NewGlobalGraph(backend)

	policyAgent, err := NewPolicyAgent(nil, globalGraph, mockPolicyStore, "test", eventBusAdapter, nil)
	if err != nil {
		t.Fatalf("Failed to create policy agent: %v", err)
	}

	return policyAgent
}

// EventBusAdapter adapts the real EventBus to match the interface expected by PolicyAgent
type EventBusAdapter struct {
	bus *events.EventBus
}

func (a *EventBusAdapter) Emit(eventType string, data map[string]interface{}) error {
	// Convert to the format expected by the real EventBus
	return a.bus.Emit(events.EventType(eventType), "policy-agent", "policy.event", data)
}
