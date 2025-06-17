package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/events"
)

// TestOrchestratorCoordination tests the core coordination functionality with mock agents
func TestOrchestratorCoordination(t *testing.T) {
	t.Run("should coordinate deployment request with mock deployment agent", func(t *testing.T) {
		// Arrange
		eventBus := events.NewEventBus(nil, false)
		registry := NewMockAgentRegistryWithResponders()

		// Create mock agents that respond to events
		deploymentAgent := NewMockDeploymentAgent(eventBus)
		policyAgent := NewMockPolicyAgent(eventBus)
		applicationAgent := NewMockApplicationAgent(eventBus)

		// Start the mock agents
		deploymentAgent.Start()
		policyAgent.Start()
		applicationAgent.Start()

		// Create orchestrator
		orchestrator := NewOrchestrator(
			createRealAIProvider(),
			createTestGraph(),
			eventBus,
			registry,
		)

		// Act
		response, err := orchestrator.Chat(context.Background(), "Deploy myapp to production")

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should successfully coordinate with deployment agent
		if !strings.Contains(response.Message, "deployment completed") && !strings.Contains(response.Message, "successfully") {
			t.Errorf("Expected successful deployment coordination, got: %s", response.Message)
		}

		// Should have the correct intent
		expectedIntents := []string{"deploy application", "deployment"}
		intentMatched := false
		for _, expectedIntent := range expectedIntents {
			if strings.Contains(response.Intent, expectedIntent) {
				intentMatched = true
				break
			}
		}

		if !intentMatched {
			t.Errorf("Expected deployment intent, got: %s", response.Intent)
		}

		t.Logf("✅ Deployment coordination: %s", response.Message)
		t.Logf("✅ Intent: %s", response.Intent)

		// Cleanup
		deploymentAgent.Stop()
		policyAgent.Stop()
		applicationAgent.Stop()
	})

	t.Run("should coordinate policy check with mock policy agent", func(t *testing.T) {
		// Arrange
		eventBus := events.NewEventBus(nil, false)
		registry := NewMockAgentRegistryWithResponders()

		// Create and start mock agents
		policyAgent := NewMockPolicyAgent(eventBus)
		policyAgent.Start()

		orchestrator := NewOrchestrator(
			createRealAIProvider(),
			createTestGraph(),
			eventBus,
			registry,
		)

		// Act
		response, err := orchestrator.Chat(context.Background(), "Is it allowed to deploy to production?")

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should successfully coordinate with policy agent
		if !strings.Contains(response.Message, "policy") && !strings.Contains(response.Message, "allowed") {
			t.Errorf("Expected successful policy coordination, got: %s", response.Message)
		}

		t.Logf("✅ Policy coordination: %s", response.Message)
		t.Logf("✅ Intent: %s", response.Intent)

		// Cleanup
		policyAgent.Stop()
	})

	t.Run("should coordinate application creation with mock application agent", func(t *testing.T) {
		// Arrange
		eventBus := events.NewEventBus(nil, false)
		registry := NewMockAgentRegistryWithResponders()

		// Create and start mock agents
		applicationAgent := NewMockApplicationAgent(eventBus)
		applicationAgent.Start()

		orchestrator := NewOrchestrator(
			createRealAIProvider(),
			createTestGraph(),
			eventBus,
			registry,
		)

		// Act
		response, err := orchestrator.Chat(context.Background(), "Create a new application called test-app")

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should successfully coordinate with application agent
		if !strings.Contains(response.Message, "application") && !strings.Contains(response.Message, "created") {
			t.Errorf("Expected successful application coordination, got: %s", response.Message)
		}

		t.Logf("✅ Application coordination: %s", response.Message)
		t.Logf("✅ Intent: %s", response.Intent)

		// Cleanup
		applicationAgent.Stop()
	})

	t.Run("should handle multiple agents coordination", func(t *testing.T) {
		// Arrange
		eventBus := events.NewEventBus(nil, false)
		registry := NewMockAgentRegistryWithResponders()

		// Create and start ALL mock agents
		deploymentAgent := NewMockDeploymentAgent(eventBus)
		policyAgent := NewMockPolicyAgent(eventBus)
		applicationAgent := NewMockApplicationAgent(eventBus)
		monitoringAgent := NewMockMonitoringAgent(eventBus)

		deploymentAgent.Start()
		policyAgent.Start()
		applicationAgent.Start()
		monitoringAgent.Start()

		orchestrator := NewOrchestrator(
			createRealAIProvider(),
			createTestGraph(),
			eventBus,
			registry,
		)

		// Test multiple coordination scenarios
		testCases := []struct {
			message  string
			expected string
		}{
			{"Deploy myapp to staging", "deployment"},
			{"Check security policies", "policy"},
			{"Create application user-service", "application"},
			{"Monitor system health", "monitoring"},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("coordinate_%s", tc.expected), func(t *testing.T) {
				response, err := orchestrator.Chat(context.Background(), tc.message)

				if err != nil {
					t.Fatalf("Expected no error for %s, got: %v", tc.message, err)
				}

				if response.Message == "" {
					t.Errorf("Expected response for %s", tc.message)
				}

				t.Logf("✅ %s: %s", tc.message, response.Message)
			})
		}

		// Cleanup
		deploymentAgent.Stop()
		policyAgent.Stop()
		applicationAgent.Stop()
		monitoringAgent.Stop()
	})
}

// MockAgent represents a mock agent that can respond to events
type MockAgent struct {
	id          string
	agentType   string
	eventBus    *events.EventBus
	routingKeys []string
	stopChan    chan bool
	mu          sync.Mutex
}

// NewMockDeploymentAgent creates a mock deployment agent
func NewMockDeploymentAgent(eventBus *events.EventBus) *MockAgent {
	return &MockAgent{
		id:          "deployment-agent",
		agentType:   "deployment",
		eventBus:    eventBus,
		routingKeys: []string{"deployment.request"},
		stopChan:    make(chan bool),
	}
}

// NewMockPolicyAgent creates a mock policy agent
func NewMockPolicyAgent(eventBus *events.EventBus) *MockAgent {
	return &MockAgent{
		id:          "policy-agent",
		agentType:   "policy",
		eventBus:    eventBus,
		routingKeys: []string{"policy.request"},
		stopChan:    make(chan bool),
	}
}

// NewMockApplicationAgent creates a mock application agent
func NewMockApplicationAgent(eventBus *events.EventBus) *MockAgent {
	return &MockAgent{
		id:          "application-agent",
		agentType:   "application",
		eventBus:    eventBus,
		routingKeys: []string{"application.request"},
		stopChan:    make(chan bool),
	}
}

// NewMockMonitoringAgent creates a mock monitoring agent
func NewMockMonitoringAgent(eventBus *events.EventBus) *MockAgent {
	return &MockAgent{
		id:          "monitoring-agent",
		agentType:   "monitoring",
		eventBus:    eventBus,
		routingKeys: []string{"monitoring.request"},
		stopChan:    make(chan bool),
	}
}

// Start begins the mock agent's event processing
func (m *MockAgent) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Subscribe to routing keys
	for _, routingKey := range m.routingKeys {
		m.eventBus.SubscribeToRoutingKey(routingKey, func(event events.Event) error {
			return m.handleEvent(event)
		})
	}
}

// Stop stops the mock agent
func (m *MockAgent) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case m.stopChan <- true:
	default:
	}
}

// handleEvent processes events and sends appropriate responses
func (m *MockAgent) handleEvent(event events.Event) error {
	// Extract correlation ID
	correlationID, ok := event.Payload["correlation_id"].(string)
	if !ok {
		return fmt.Errorf("no correlation_id in event")
	}

	// Create mock response based on agent type
	var responseMessage string
	var responseStatus string = "success"

	switch m.agentType {
	case "deployment":
		responseMessage = "deployment completed successfully to production environment"
	case "policy":
		responseMessage = "policy check passed - deployment is allowed"
	case "application":
		responseMessage = "application 'test-app' created successfully"
	case "monitoring":
		responseMessage = "monitoring dashboard shows all systems healthy"
	default:
		responseMessage = fmt.Sprintf("%s operation completed successfully", m.agentType)
	}

	// Send response event
	responseEvent := events.Event{
		Type:    events.EventTypeResponse,
		Source:  m.id,
		Subject: fmt.Sprintf("%s response", m.agentType),
		Payload: map[string]interface{}{
			"correlation_id": correlationID,
			"status":         responseStatus,
			"message":        responseMessage,
			"agent_id":       m.id,
			"agent_type":     m.agentType,
		},
		Timestamp: time.Now().Unix(),
		ID:        fmt.Sprintf("%s-resp-%d", m.id, time.Now().UnixNano()),
	}

	// Small delay to simulate processing
	go func() {
		time.Sleep(100 * time.Millisecond)
		m.eventBus.EmitEvent(responseEvent)
	}()

	return nil
}

// NewMockAgentRegistryWithResponders creates a registry with agents that can respond
func NewMockAgentRegistryWithResponders() *MockAgentRegistry {
	registry := &MockAgentRegistry{
		agents:       make(map[string]agentRegistry.AgentStatus),
		capabilities: make(map[string]agentRegistry.AgentCapability),
	}

	// Register agents
	registry.agents["deployment-agent"] = agentRegistry.AgentStatus{
		ID:     "deployment-agent",
		Type:   "deployment",
		Status: "running",
	}

	registry.agents["policy-agent"] = agentRegistry.AgentStatus{
		ID:     "policy-agent",
		Type:   "policy",
		Status: "running",
	}

	registry.agents["application-agent"] = agentRegistry.AgentStatus{
		ID:     "application-agent",
		Type:   "application",
		Status: "running",
	}

	registry.agents["monitoring-agent"] = agentRegistry.AgentStatus{
		ID:     "monitoring-agent",
		Type:   "monitoring",
		Status: "running",
	}

	// Register capabilities
	registry.capabilities["deployment_management"] = agentRegistry.AgentCapability{
		Name:        "deployment_management",
		Description: "Handle deployment operations",
		Intents:     []string{"deploy application", "deployment", "deploy"},
		RoutingKeys: []string{"deployment.request"},
		Version:     "1.0.0",
	}

	registry.capabilities["policy_evaluation"] = agentRegistry.AgentCapability{
		Name:        "policy_evaluation",
		Description: "Handle policy checks and compliance",
		Intents:     []string{"policy check", "policy", "compliance", "security"},
		RoutingKeys: []string{"policy.request"},
		Version:     "1.0.0",
	}

	registry.capabilities["application_management"] = agentRegistry.AgentCapability{
		Name:        "application_management",
		Description: "Handle application creation and management",
		Intents:     []string{"create application", "manage application", "list applications"},
		RoutingKeys: []string{"application.request"},
		Version:     "1.0.0",
	}

	registry.capabilities["system_monitoring"] = agentRegistry.AgentCapability{
		Name:        "system_monitoring",
		Description: "Monitor system health and performance",
		Intents:     []string{"monitor", "monitoring", "health", "status"},
		RoutingKeys: []string{"monitoring.request"},
		Version:     "1.0.0",
	}

	return registry
}

// Add missing interface methods to satisfy agentRegistry.AgentRegistry

func (m *MockAgentRegistry) FindAgentByID(ctx context.Context, agentID string) (agentRegistry.AgentInterface, error) {
	// For tests, we don't need a real agent implementation
	return nil, fmt.Errorf("agent %s not found", agentID)
}

func (m *MockAgentRegistry) ListAllAgents(ctx context.Context) ([]agentRegistry.AgentStatus, error) {
	var agents []agentRegistry.AgentStatus
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents, nil
}

func (m *MockAgentRegistry) UnregisterAgent(ctx context.Context, agentID string) error {
	delete(m.agents, agentID)
	return nil
}

func (m *MockAgentRegistry) GetAgentHealth(ctx context.Context, agentID string) (agentRegistry.HealthStatus, error) {
	return agentRegistry.HealthStatus{Status: "healthy"}, nil
}

// Helper functions from original test file
