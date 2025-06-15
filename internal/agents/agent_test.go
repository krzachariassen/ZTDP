package agents

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/events"
)

// Test for dynamic agent registration and discovery
func TestAgentRegistration_DynamicCapabilities(t *testing.T) {
	tests := []struct {
		name       string
		agentType  string
		setupAgent func() AgentInterface
		wantCaps   []string // Expected capability names
	}{
		{
			name:      "Policy Agent Registration",
			agentType: "policy",
			setupAgent: func() AgentInterface {
				return &MockPolicyAgent{
					id: "policy-agent-1",
					capabilities: []AgentCapability{
						{
							Name:        "policy_evaluation",
							Description: "Evaluates policies using AI reasoning",
							Intents:     []string{"evaluate policy", "check compliance", "validate rules"},
							Version:     "1.0.0",
						},
					},
				}
			},
			wantCaps: []string{"policy_evaluation"},
		},
		{
			name:      "Custom Agent Registration",
			agentType: "security",
			setupAgent: func() AgentInterface {
				return &MockSecurityAgent{
					id: "security-agent-1",
					capabilities: []AgentCapability{
						{
							Name:        "threat_detection",
							Description: "Detects security threats using AI",
							Intents:     []string{"scan for threats", "security analysis", "vulnerability check"},
							Version:     "1.0.0",
						},
						{
							Name:        "access_control",
							Description: "Manages access control policies",
							Intents:     []string{"check permissions", "validate access", "authorize user"},
							Version:     "1.0.0",
						},
					},
				}
			},
			wantCaps: []string{"threat_detection", "access_control"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create registry
			registry := NewInMemoryAgentRegistry()

			// Create and register agent
			agent := tt.setupAgent()
			err := registry.RegisterAgent(context.Background(), agent)
			if err != nil {
				t.Fatalf("Failed to register agent: %v", err)
			}

			// Test capability discovery
			for _, capName := range tt.wantCaps {
				agents, err := registry.FindAgentsByCapability(context.Background(), capName)
				if err != nil {
					t.Errorf("Failed to find agents by capability %s: %v", capName, err)
					continue
				}

				if len(agents) == 0 {
					t.Errorf("No agents found for capability %s", capName)
					continue
				}

				// Verify agent type matches
				found := false
				for _, agentStatus := range agents {
					if agentStatus.Type == tt.agentType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected agent type %s not found for capability %s", tt.agentType, capName)
				}
			}

			// Test intent-based discovery
			agent = tt.setupAgent()
			caps := agent.GetCapabilities()
			if len(caps) != len(tt.wantCaps) {
				t.Errorf("Expected %d capabilities, got %d", len(tt.wantCaps), len(caps))
			}
		})
	}
}

// Test dynamic intent routing without hardcoded types
func TestAgentCoordination_IntentBasedRouting(t *testing.T) {
	tests := []struct {
		name          string
		intent        string
		expectedAgent string
		expectedCap   string
	}{
		{
			name:          "Policy Intent Routing",
			intent:        "I need to check if this deployment violates any policies",
			expectedAgent: "policy",
			expectedCap:   "policy_evaluation",
		},
		{
			name:          "Security Intent Routing",
			intent:        "Scan this configuration for security vulnerabilities",
			expectedAgent: "security",
			expectedCap:   "threat_detection",
		},
		{
			name:          "Access Control Intent Routing",
			intent:        "Check if user has permission to access this resource",
			expectedAgent: "security",
			expectedCap:   "access_control",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup registry with multiple agents
			registry := NewInMemoryAgentRegistry()
			coordinator := NewAgentCoordinator(registry, nil) // nil event bus for unit test

			// Register policy agent
			policyAgent := &MockPolicyAgent{
				id: "policy-agent-1",
				capabilities: []AgentCapability{
					{
						Name:    "policy_evaluation",
						Intents: []string{"check policy", "validate rules", "deployment violates", "compliance"},
					},
				},
			}
			registry.RegisterAgent(context.Background(), policyAgent)

			// Register security agent
			securityAgent := &MockSecurityAgent{
				id: "security-agent-1",
				capabilities: []AgentCapability{
					{
						Name:    "threat_detection",
						Intents: []string{"scan", "security", "vulnerability", "threat"},
					},
					{
						Name:    "access_control",
						Intents: []string{"permission", "access", "authorize", "user"},
					},
				},
			}
			registry.RegisterAgent(context.Background(), securityAgent)

			// Test intent routing
			targetAgents, capability, err := coordinator.ResolveIntent(context.Background(), tt.intent)
			if err != nil {
				t.Fatalf("Failed to resolve intent: %v", err)
			}

			if len(targetAgents) == 0 {
				t.Fatal("No target agents found for intent")
			}

			// Verify correct agent type was selected
			found := false
			for _, agent := range targetAgents {
				if agent.Type == tt.expectedAgent {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected agent type %s not found, got agents: %v", tt.expectedAgent, targetAgents)
			}

			// Verify correct capability was identified
			if capability != tt.expectedCap {
				t.Errorf("Expected capability %s, got %s", tt.expectedCap, capability)
			}
		})
	}
}

// Mock implementations for testing

type MockPolicyAgent struct {
	id           string
	capabilities []AgentCapability
}

func (m *MockPolicyAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  m.id,
		Subject: "policy_result",
		Payload: map[string]interface{}{
			"decision": "allowed",
			"reason":   "No policy violations found",
		},
	}, nil
}

func (m *MockPolicyAgent) GetCapabilities() []AgentCapability {
	return m.capabilities
}

func (m *MockPolicyAgent) GetStatus() AgentStatus {
	return AgentStatus{
		ID:     m.id,
		Type:   "policy",
		Status: "running",
	}
}

func (m *MockPolicyAgent) Start(ctx context.Context) error { return nil }
func (m *MockPolicyAgent) Stop(ctx context.Context) error  { return nil }
func (m *MockPolicyAgent) Health() HealthStatus {
	return HealthStatus{Healthy: true, Status: "healthy"}
}

type MockSecurityAgent struct {
	id           string
	capabilities []AgentCapability
}

func (m *MockSecurityAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  m.id,
		Subject: "security_result",
		Payload: map[string]interface{}{
			"threats_found": false,
			"score":         "low_risk",
		},
	}, nil
}

func (m *MockSecurityAgent) GetCapabilities() []AgentCapability {
	return m.capabilities
}

func (m *MockSecurityAgent) GetStatus() AgentStatus {
	return AgentStatus{
		ID:     m.id,
		Type:   "security",
		Status: "running",
	}
}

func (m *MockSecurityAgent) Start(ctx context.Context) error { return nil }
func (m *MockSecurityAgent) Stop(ctx context.Context) error  { return nil }
func (m *MockSecurityAgent) Health() HealthStatus {
	return HealthStatus{Healthy: true, Status: "healthy"}
}
