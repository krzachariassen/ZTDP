package agentRegistry

import (
	"context"
	"testing"
)

// MockAgent for testing - implements only what the registry needs
type MockAgent struct {
	id           string
	capabilities []AgentCapability
	status       AgentStatus
	health       HealthStatus
}

func (m *MockAgent) GetID() string {
	return m.id
}

func (m *MockAgent) GetStatus() AgentStatus {
	return m.status
}

func (m *MockAgent) GetCapabilities() []AgentCapability {
	return m.capabilities
}

func (m *MockAgent) Start(ctx context.Context) error {
	return nil
}

func (m *MockAgent) Stop(ctx context.Context) error {
	return nil
}

func (m *MockAgent) Health() HealthStatus {
	return m.health
}

// TestAgentRegistry_BasicOperations tests the core registry functionality
func TestAgentRegistry_BasicOperations(t *testing.T) {
	registry := NewInMemoryAgentRegistry()
	ctx := context.Background()

	// Create a mock agent
	mockAgent := &MockAgent{
		id: "test-agent-1",
		capabilities: []AgentCapability{
			{
				Name:        "test_capability",
				Description: "Test capability for registry testing",
				Intents:     []string{"test intent"},
				Version:     "1.0.0",
			},
		},
		status: AgentStatus{
			ID:     "test-agent-1",
			Type:   "test",
			Status: "running",
		},
		health: HealthStatus{
			Healthy: true,
			Status:  "healthy",
			Message: "Test agent is healthy",
		},
	}

	// Test 1: Register agent
	err := registry.RegisterAgent(ctx, mockAgent)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// Test 2: Find agent by ID
	foundAgent, err := registry.FindAgentByID(ctx, "test-agent-1")
	if err != nil {
		t.Fatalf("Failed to find agent by ID: %v", err)
	}
	if foundAgent.GetID() != "test-agent-1" {
		t.Errorf("Expected agent ID 'test-agent-1', got '%s'", foundAgent.GetID())
	}

	// Test 3: Find agents by capability
	agents, err := registry.FindAgentsByCapability(ctx, "test_capability")
	if err != nil {
		t.Fatalf("Failed to find agents by capability: %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent with test_capability, got %d", len(agents))
	}
	if agents[0].ID != "test-agent-1" {
		t.Errorf("Expected agent ID 'test-agent-1', got '%s'", agents[0].ID)
	}

	// Test 4: List all agents
	allAgents, err := registry.ListAllAgents(ctx)
	if err != nil {
		t.Fatalf("Failed to list all agents: %v", err)
	}
	if len(allAgents) != 1 {
		t.Errorf("Expected 1 registered agent, got %d", len(allAgents))
	}

	// Test 5: Get agent health
	health, err := registry.GetAgentHealth(ctx, "test-agent-1")
	if err != nil {
		t.Fatalf("Failed to get agent health: %v", err)
	}
	if !health.Healthy {
		t.Errorf("Expected agent to be healthy")
	}

	// Test 6: Unregister agent
	err = registry.UnregisterAgent(ctx, "test-agent-1")
	if err != nil {
		t.Fatalf("Failed to unregister agent: %v", err)
	}

	// Test 7: Verify agent is removed
	_, err = registry.FindAgentByID(ctx, "test-agent-1")
	if err == nil {
		t.Errorf("Expected error when finding unregistered agent, got nil")
	}
}

// TestAgentRegistry_CapabilityDiscovery tests capability-based discovery
func TestAgentRegistry_CapabilityDiscovery(t *testing.T) {
	registry := NewInMemoryAgentRegistry()
	ctx := context.Background()

	// Create agents with different capabilities
	policyAgent := &MockAgent{
		id: "policy-agent",
		capabilities: []AgentCapability{
			{
				Name:    "policy_evaluation",
				Intents: []string{"check policy", "evaluate policy"},
				Version: "1.0.0",
			},
		},
		status: AgentStatus{ID: "policy-agent", Type: "policy", Status: "running"},
		health: HealthStatus{Healthy: true, Status: "healthy"},
	}

	deploymentAgent := &MockAgent{
		id: "deployment-agent",
		capabilities: []AgentCapability{
			{
				Name:    "deployment_orchestration",
				Intents: []string{"deploy application", "create deployment"},
				Version: "1.0.0",
			},
		},
		status: AgentStatus{ID: "deployment-agent", Type: "deployment", Status: "running"},
		health: HealthStatus{Healthy: true, Status: "healthy"},
	}

	// Register both agents
	err := registry.RegisterAgent(ctx, policyAgent)
	if err != nil {
		t.Fatalf("Failed to register policy agent: %v", err)
	}

	err = registry.RegisterAgent(ctx, deploymentAgent)
	if err != nil {
		t.Fatalf("Failed to register deployment agent: %v", err)
	}

	// Test capability discovery
	policyAgents, err := registry.FindAgentsByCapability(ctx, "policy_evaluation")
	if err != nil {
		t.Fatalf("Failed to find policy agents: %v", err)
	}
	if len(policyAgents) != 1 || policyAgents[0].ID != "policy-agent" {
		t.Errorf("Expected 1 policy agent, got %d", len(policyAgents))
	}

	deploymentAgents, err := registry.FindAgentsByCapability(ctx, "deployment_orchestration")
	if err != nil {
		t.Fatalf("Failed to find deployment agents: %v", err)
	}
	if len(deploymentAgents) != 1 || deploymentAgents[0].ID != "deployment-agent" {
		t.Errorf("Expected 1 deployment agent, got %d", len(deploymentAgents))
	}

	// Test getting all capabilities
	capabilities, err := registry.GetAvailableCapabilities(ctx)
	if err != nil {
		t.Fatalf("Failed to get available capabilities: %v", err)
	}
	if len(capabilities) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(capabilities))
	}
}
