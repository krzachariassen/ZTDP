package orchestrator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// TestOrchestratorBasicChat tests the fundamental chat functionality
func TestOrchestratorBasicChat(t *testing.T) {
	t.Run("should respond to simple user messages", func(t *testing.T) {
		// Arrange
		orchestrator := createTestOrchestrator(t)

		// Act
		response, err := orchestrator.Chat(context.Background(), "Hello, what can you help me with?")

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if response == nil {
			t.Fatal("Expected response, got nil")
		}

		if response.Message == "" {
			t.Error("Expected non-empty message in response")
		}

		// For general conversation, intent might be empty - that's ok
		t.Logf("✅ Response: %s", response.Message)
		t.Logf("✅ Intent: %s", response.Intent)
	})
}

// TestOrchestratorIntentDetection tests intent detection and routing
func TestOrchestratorIntentDetection(t *testing.T) {
	tests := []struct {
		name                 string
		userMessage          string
		expectedAgentType    string // Focus on which agent should handle this
		shouldRoute          bool
		expectedAgentPattern string // Pattern to match in agent ID
	}{
		{
			name:                 "deployment intent",
			userMessage:          "Deploy myapp to production",
			expectedAgentType:    "deployment",
			shouldRoute:          true,
			expectedAgentPattern: "deployment-agent",
		},
		{
			name:                 "policy intent",
			userMessage:          "Check if deployment is allowed",
			expectedAgentType:    "policy",
			shouldRoute:          true,
			expectedAgentPattern: "policy-agent",
		},
		// {
		// 	name:              "general question",
		// 	userMessage:       "What is ZTDP?",
		// 	expectedAgentType: "general",
		// 	shouldRoute:       false, // Should be handled directly - COMMENTED OUT: takes too long for orchestration tests
		// },
		{
			name:                 "resource creation",
			userMessage:          "Create a new application called test-app",
			expectedAgentType:    "application",
			shouldRoute:          true,
			expectedAgentPattern: "application-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			orchestrator := createTestOrchestrator(t)

			// Act
			response, err := orchestrator.Chat(context.Background(), tt.userMessage)

			// Assert
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if tt.shouldRoute {
				// Check that it was routed to the correct agent type
				if len(response.Actions) == 0 {
					t.Error("Expected actions to be recorded for agent routing")
				}

				// Check that the correct agent was selected (from test mode simulation)
				if !strings.Contains(strings.ToLower(response.Message), tt.expectedAgentPattern) {
					t.Errorf("Expected response to mention agent %s, got: %s", tt.expectedAgentPattern, response.Message)
				}
			}

			t.Logf("✅ Message: %s -> Intent: %s -> Response: %s", tt.userMessage, response.Intent, response.Message)
		})
	}
}

// TestOrchestratorAgentRouting tests routing to specialist agents
func TestOrchestratorAgentRouting(t *testing.T) {
	t.Run("should route deployment requests to deployment agent", func(t *testing.T) {
		// Arrange
		orchestrator := createTestOrchestrator(t)

		// Act
		response, err := orchestrator.Chat(context.Background(), "Deploy test-app to development environment")

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should contain evidence of agent routing (test mode simulation)
		if len(response.Actions) == 0 {
			t.Error("Expected actions to be recorded for agent routing")
		}

		// In test mode, should mention the deployment agent
		if !strings.Contains(strings.ToLower(response.Message), "deployment-agent") {
			t.Errorf("Expected response to mention deployment-agent routing, got: %s", response.Message)
		}

		t.Logf("✅ Routed deployment request: %s", response.Message)
	})

	t.Run("should route policy requests to policy agent", func(t *testing.T) {
		// Arrange
		orchestrator := createTestOrchestrator(t)

		// Act
		response, err := orchestrator.Chat(context.Background(), "Is it allowed to deploy to production?")

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should contain evidence of agent routing (test mode simulation)
		if len(response.Actions) == 0 {
			t.Error("Expected actions to be recorded for agent routing")
		}

		t.Logf("✅ Routed policy request: %s", response.Message)
	})
}

// TestOrchestratorResourceCreation tests direct resource creation
func TestOrchestratorResourceCreation(t *testing.T) {
	t.Run("should handle resource creation requests directly", func(t *testing.T) {
		// Arrange
		orchestrator := createTestOrchestrator(t)

		// Act
		response, err := orchestrator.Chat(context.Background(), "Create a new application called 'user-service' with a REST API")

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if response.Intent != "create application" {
			t.Errorf("Expected create application intent, got %s", response.Intent)
		}

		// Should contain evidence of resource creation
		if len(response.Actions) == 0 {
			t.Error("Expected actions to be recorded for resource creation")
		}

		t.Logf("✅ Created resource: %s", response.Message)
	})
}

// TestOrchestratorFrameworkIntegration tests framework integration
func TestOrchestratorFrameworkIntegration(t *testing.T) {
	t.Run("should work with agent framework", func(t *testing.T) {
		t.Skip("Framework integration test - will be implemented later")
	})
}

// Helper functions for test setup

func createTestOrchestrator(t *testing.T) *Orchestrator {
	provider := createRealAIProvider()
	if provider == nil {
		t.Skip("Skipping test - OpenAI API key not available")
	}

	orchestrator := NewOrchestrator(
		provider,
		createTestGraph(),
		events.NewEventBus(nil, false),
		NewMockAgentRegistry(), // Use the enhanced mock registry
	)

	// Enable test mode to avoid waiting for real agent responses
	orchestrator.testMode = true

	return orchestrator
}

func createRealAIProvider() ai.AIProvider {
	// Get API key from environment, same as main.go does
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Return nil if no API key is available - test should handle this gracefully
		return nil
	}

	provider, err := ai.NewOpenAIProvider(ai.DefaultOpenAIConfig(), apiKey)
	if err != nil || provider == nil {
		return nil
	}
	return provider
}

func createTestGraph() *graph.GlobalGraph {
	backend := graph.NewMemoryGraph()
	return graph.NewGlobalGraph(backend)
}

// MockAgentRegistry for testing
type MockAgentRegistry struct {
	agents       map[string]agentRegistry.AgentStatus
	capabilities map[string]agentRegistry.AgentCapability
}

func NewMockAgentRegistry() *MockAgentRegistry {
	registry := &MockAgentRegistry{
		agents:       make(map[string]agentRegistry.AgentStatus),
		capabilities: make(map[string]agentRegistry.AgentCapability),
	}

	// Register mock deployment agent
	registry.agents["deployment-agent"] = agentRegistry.AgentStatus{
		ID:     "deployment-agent",
		Type:   "deployment",
		Status: "running",
	}

	// Register mock policy agent
	registry.agents["policy-agent"] = agentRegistry.AgentStatus{
		ID:     "policy-agent",
		Type:   "policy",
		Status: "running",
	}

	// Register mock application agent
	registry.agents["application-agent"] = agentRegistry.AgentStatus{
		ID:     "application-agent",
		Type:   "application",
		Status: "running",
	}

	// Register capabilities
	registry.capabilities["deployment_management"] = agentRegistry.AgentCapability{
		Name:        "deployment_management",
		Description: "Handle deployment operations",
		Intents:     []string{"deploy application", "deployment"},
		RoutingKeys: []string{"deployment.request"},
		Version:     "1.0.0",
	}

	registry.capabilities["policy_evaluation"] = agentRegistry.AgentCapability{
		Name:        "policy_evaluation",
		Description: "Handle policy checks",
		Intents:     []string{"policy check", "policy"},
		RoutingKeys: []string{"policy.request"},
		Version:     "1.0.0",
	}

	registry.capabilities["application_management"] = agentRegistry.AgentCapability{
		Name:        "application_management",
		Description: "Handle application creation and management",
		Intents:     []string{"create application", "manage application", "list applications"}, // Removed generic "application"
		RoutingKeys: []string{"application.request"},
		Version:     "1.0.0",
	}

	return registry
}

func (m *MockAgentRegistry) FindAgentsByCapability(ctx context.Context, capability string) ([]agentRegistry.AgentStatus, error) {
	var agents []agentRegistry.AgentStatus

	switch capability {
	case "deployment_management":
		agents = append(agents, m.agents["deployment-agent"])
	case "policy_evaluation":
		agents = append(agents, m.agents["policy-agent"])
	case "application_management":
		agents = append(agents, m.agents["application-agent"])
	}

	return agents, nil
}

func (m *MockAgentRegistry) GetAvailableCapabilities(ctx context.Context) ([]agentRegistry.AgentCapability, error) {
	var caps []agentRegistry.AgentCapability
	for _, cap := range m.capabilities {
		caps = append(caps, cap)
	}
	return caps, nil
}

func (m *MockAgentRegistry) RegisterAgent(ctx context.Context, agent agentRegistry.AgentInterface) error {
	// No-op for tests
	return nil
}
