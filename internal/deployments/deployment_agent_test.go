package deployments

import (
	"context"
	"strings"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/agentFramework"
	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// TestDeploymentAgentMigrationToFramework tests that the DeploymentAgent can be created using the new framework
func TestDeploymentAgentMigrationToFramework(t *testing.T) {
	// Arrange - Set up dependencies
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())

	// Mock AI provider
	mockAIProvider := &MockAIProvider{}

	// Act - Create DeploymentAgent using framework
	agent, err := NewDeploymentAgent(mockGraph, mockAIProvider, "test", eventBus, registry)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error creating framework deployment agent, got: %v", err)
	}

	if agent.GetID() != "deployment" {
		t.Errorf("Expected agent ID 'deployment', got: %s", agent.GetID())
	}

	// Verify auto-registration
	registeredAgent, err := registry.FindAgentByID(context.Background(), "deployment")
	if err != nil {
		t.Errorf("Expected agent to be auto-registered, got error: %v", err)
	}
	if registeredAgent.GetID() != "deployment" {
		t.Errorf("Expected registered agent ID 'deployment', got: %s", registeredAgent.GetID())
	}

	// Verify capabilities
	capabilities := agent.GetCapabilities()
	if len(capabilities) == 0 {
		t.Error("Expected agent to have capabilities")
	}

	foundDeploymentCapability := false
	for _, cap := range capabilities {
		if cap.Name == "deployment_orchestration" {
			foundDeploymentCapability = true
			// Verify intents
			expectedIntents := []string{"deploy application", "execute deployment", "start deployment", "run deployment"}
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
	if !foundDeploymentCapability {
		t.Error("Expected agent to have deployment_orchestration capability")
	}
}

// TestFrameworkDeploymentAgentEventHandling tests that the framework agent can handle deployment events
func TestFrameworkDeploymentAgentEventHandling(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())
	mockAIProvider := &MockAIProvider{}

	// Create agent using framework
	baseAgent, err := NewDeploymentAgent(mockGraph, mockAIProvider, "test", eventBus, registry)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Cast to framework agent to access ProcessEvent
	agent, ok := baseAgent.(*agentFramework.BaseAgent)
	if !ok {
		t.Fatalf("Expected BaseAgent, got %T", baseAgent)
	}

	// Create test deployment event
	deploymentEvent := &events.Event{
		Type:    events.EventTypeRequest,
		Source:  "test-source",
		Subject: "deployment.request",
		Payload: map[string]interface{}{
			"intent":         "deploy application",
			"user_message":   "Deploy test-app to production",
			"correlation_id": "test-123",
		},
	}

	// Act - Process the event
	response, err := agent.ProcessEvent(context.Background(), deploymentEvent)

	// Assert
	if err != nil {
		t.Errorf("Expected no error processing deployment event, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Verify response structure
	if response.Source != "deployment-agent" {
		t.Errorf("Expected response source 'deployment-agent', got: %s", response.Source)
	}

	// Verify correlation ID is preserved
	if correlationID, ok := response.Payload["correlation_id"]; !ok || correlationID != "test-123" {
		t.Errorf("Expected correlation_id 'test-123', got: %v", correlationID)
	}
}

// TestDeploymentAgentBusinessLogicIntegration tests that business logic is preserved after migration
func TestDeploymentAgentBusinessLogicIntegration(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)

	// Initialize global event bus for the engine
	events.InitializeEventBus(nil)

	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())

	// Add test application to graph
	testApp := &graph.Node{
		ID:   "test-app", // ID should match the application name
		Kind: "application",
		Metadata: map[string]interface{}{
			"name": "test-app",
		},
	}
	mockGraph.AddNode(testApp)

	// Add test environment to graph
	testEnv := &graph.Node{
		ID:   "production", // ID should match the environment name
		Kind: "environment",
		Metadata: map[string]interface{}{
			"name": "production",
		},
	}
	mockGraph.AddNode(testEnv)

	// Add allowed_in edge from application to environment
	err := mockGraph.AddEdge("test-app", "production", "allowed_in")
	if err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	mockAIProvider := &MockAIProvider{
		responses: map[string]string{
			"parse_deployment": `{"application": "test-app", "environment": "production"}`,
		},
	}

	// Create agent using framework
	baseAgent, err := NewDeploymentAgent(mockGraph, mockAIProvider, "test", eventBus, registry)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	agent, ok := baseAgent.(*agentFramework.BaseAgent)
	if !ok {
		t.Fatalf("Expected BaseAgent, got %T", baseAgent)
	}

	// Create deployment event with valid application
	deploymentEvent := &events.Event{
		Type:    events.EventTypeRequest,
		Source:  "test-source",
		Subject: "deployment.request",
		Payload: map[string]interface{}{
			"intent":       "deploy application",
			"user_message": "Deploy test-app to production",
		},
	}

	// Act - Process the event
	response, err := agent.ProcessEvent(context.Background(), deploymentEvent)

	// Assert
	if err != nil {
		t.Errorf("Expected no error processing deployment event, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Should process the event without panic/errors (deployment failure is expected due to no services)
	if status, ok := response.Payload["status"].(string); !ok || status != "error" {
		t.Errorf("Expected error status due to no services configured, got status: %v, payload: %v", status, response.Payload)
	}

	// Verify the error message indicates the expected business logic was executed
	if errorMsg, ok := response.Payload["error"].(string); ok {
		if !strings.Contains(errorMsg, "deployment failed") {
			t.Errorf("Expected deployment failure error message, got: %s", errorMsg)
		}
	} else {
		t.Error("Expected error message in response payload")
	}
}

// MockAIProvider for testing
type MockAIProvider struct {
	responses map[string]string
}

func (m *MockAIProvider) CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Return different responses based on prompt content
	if strings.Contains(systemPrompt, "deployment request") || strings.Contains(userPrompt, "Deploy") {
		if response, ok := m.responses["parse_deployment"]; ok {
			return response, nil
		}
		return `{"application": "test-app", "environment": "production"}`, nil
	}
	return "Mock AI response", nil
}

func (m *MockAIProvider) GetProviderInfo() *ai.ProviderInfo {
	return &ai.ProviderInfo{
		Name:    "mock",
		Version: "1.0.0",
	}
}

func (m *MockAIProvider) Close() error {
	return nil
}
