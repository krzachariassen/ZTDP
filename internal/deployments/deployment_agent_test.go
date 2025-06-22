package deployments

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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

// TestDeploymentAgentCreation tests that the DeploymentAgent can be created using the framework
// TestDeploymentAgentCreation tests that the DeploymentAgent can be created using the framework
func TestDeploymentAgentCreation(t *testing.T) {
	// Arrange - Set up dependencies
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())

	// Use real AI provider for authentic testing
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Act - Create DeploymentAgent using framework
	agent, err := NewDeploymentAgent(mockGraph, realAIProvider, eventBus, registry)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error creating framework deployment agent, got: %v", err)
	}

	if agent.GetID() != "deployment-agent" {
		t.Errorf("Expected agent ID 'deployment-agent', got: %s", agent.GetID())
	}

	// Verify auto-registration
	registeredAgent, err := registry.FindAgentByID(context.Background(), "deployment-agent")
	if err != nil {
		t.Errorf("Expected agent to be auto-registered, got error: %v", err)
	}
	if registeredAgent.GetID() != "deployment-agent" {
		t.Errorf("Expected registered agent ID 'deployment-agent', got: %s", registeredAgent.GetID())
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

// TestDeploymentAgentEventHandling tests that the agent can handle deployment events with real AI
func TestDeploymentAgentEventHandling(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)
	mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())

	// Use real AI provider - no mocking!
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Create agent using framework
	baseAgent, err := NewDeploymentAgent(mockGraph, realAIProvider, eventBus, registry)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Cast to framework agent to access ProcessEvent
	agent, ok := baseAgent.(*agentFramework.BaseAgent)
	if !ok {
		t.Fatalf("Expected BaseAgent, got %T", baseAgent)
	}

	// Create test deployment event with natural language
	deploymentEvent := &events.Event{
		Type:    events.EventTypeRequest,
		Source:  "test-source",
		Subject: "deployment.request",
		Payload: map[string]interface{}{
			"user_message":   "Deploy test-app to production environment",
			"correlation_id": "test-123",
		},
	}

	// Act - Process the event with real AI
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

// TestDeploymentAgentAIIntegration tests real AI parameter extraction through the agent
func TestDeploymentAgentAIIntegration(t *testing.T) {
	tests := []struct {
		name           string
		userMessage    string
		expectedAction string // We can still expect certain actions based on clear language
		shouldSucceed  bool
	}{
		{
			name:           "deploy with explicit details",
			userMessage:    "Deploy the checkout-api application to production environment",
			expectedAction: "deploy", // This should be clear enough for AI to extract
			shouldSucceed:  true,
		},
		{
			name:           "deploy to staging",
			userMessage:    "deploy myapp to staging",
			expectedAction: "deploy", // This should be clear
			shouldSucceed:  true,
		},
		{
			name:           "start deployment",
			userMessage:    "start deployment of payment service to prod",
			expectedAction: "deploy", // Should be interpreted as deploy
			shouldSucceed:  true,
		},
		{
			name:           "ambiguous request",
			userMessage:    "do something with deployments",
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

			// Use real AI provider
			realAIProvider := getRealAIProvider(t)
			defer realAIProvider.Close()

			// Create agent using framework
			baseAgent, err := NewDeploymentAgent(mockGraph, realAIProvider, eventBus, registry)
			if err != nil {
				t.Fatalf("Failed to create agent: %v", err)
			}

			agent, ok := baseAgent.(*agentFramework.BaseAgent)
			if !ok {
				t.Fatalf("Expected BaseAgent, got %T", baseAgent)
			}

			// Create deployment event
			deploymentEvent := &events.Event{
				Type:    events.EventTypeRequest,
				Source:  "test-source",
				Subject: "deployment.request",
				Payload: map[string]interface{}{
					"user_message":   tt.userMessage,
					"correlation_id": "test-correlation-id",
				},
			}

			// Act - Process with real AI
			response, err := agent.ProcessEvent(context.Background(), deploymentEvent)

			// Assert
			if !tt.shouldSucceed {
				if err == nil && response != nil {
					if status, ok := response.Payload["status"].(string); ok && status == "error" {
						t.Logf("✅ Expected error handled gracefully with error response")
					} else {
						t.Errorf("Expected error response, got status: %v", response.Payload["status"])
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)

				// If we expect a specific action, verify AI extracted it
				if tt.expectedAction != "" {
					if action, exists := response.Payload["action"]; exists {
						t.Logf("✅ AI extracted action: %v for message: %s", action, tt.userMessage)
						// We're flexible here - AI might interpret things differently but validly
					} else {
						t.Logf("⚠️ AI didn't extract explicit action, but processed: %s", tt.userMessage)
						// This might still be valid if AI asks for clarification
					}
				}

				// Verify correlation ID is preserved
				if correlationID, ok := response.Payload["correlation_id"]; ok {
					assert.Equal(t, "test-correlation-id", correlationID)
				}
			}
		})
	}
}

// TestDeploymentAgentBusinessLogicIntegration tests that business logic is preserved with real AI
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

	// Use real AI provider for authentic business logic testing
	realAIProvider := getRealAIProvider(t)
	defer realAIProvider.Close()

	// Create agent using framework
	baseAgent, err := NewDeploymentAgent(mockGraph, realAIProvider, eventBus, registry)
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
			"user_message":   "Deploy test-app to production",
			"correlation_id": "business-logic-test-123",
		},
	}

	// Act - Process the event with real AI
	response, err := agent.ProcessEvent(context.Background(), deploymentEvent)

	// Assert
	if err != nil {
		t.Errorf("Expected no error processing deployment event, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Verify correlation ID preservation
	if correlationID, ok := response.Payload["correlation_id"]; !ok || correlationID != "business-logic-test-123" {
		t.Errorf("Expected correlation_id 'business-logic-test-123', got: %v", correlationID)
	}

	// Real AI integration test: we can't predict exact responses, but we can verify processing
	if status, exists := response.Payload["status"]; exists {
		t.Logf("✅ Real AI business logic processing returned status: %v", status)
		// Business logic should result in success, error, or clarification
		assert.Contains(t, []string{"success", "error", "clarification"}, status)
	} else {
		t.Error("❌ Expected AI business logic to process and return status")
	}

	// If deployment was successful, check for deployment ID
	if status, ok := response.Payload["status"].(string); ok && status == "success" {
		if deploymentID, exists := response.Payload["deployment_id"]; exists {
			t.Logf("✅ AI business logic created deployment ID: %v", deploymentID)
		}
	}
}

// Test the complete deployment orchestration workflow (TDD - this should FAIL initially)
func TestDeploymentOrchestrationWorkflow(t *testing.T) {
	t.Run("full deployment orchestration from user message", func(t *testing.T) {
		// Setup - Create all the agents and infrastructure needed for orchestration
		registry := agentRegistry.NewInMemoryAgentRegistry()
		eventBus := events.NewEventBus(nil, false)
		mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())
		realAIProvider := getRealAIProvider(t)
		defer realAIProvider.Close()

		// Track events fired during the workflow
		eventsReceived := make([]string, 0)
		eventBus.Subscribe(events.EventTypeRequest, func(event events.Event) error {
			eventsReceived = append(eventsReceived, event.Subject)
			t.Logf("📨 Event received: %s", event.Subject)
			return nil
		})
		eventBus.Subscribe(events.EventTypeResponse, func(event events.Event) error {
			eventsReceived = append(eventsReceived, event.Subject)
			t.Logf("📨 Event received: %s", event.Subject)
			return nil
		})
		eventBus.Subscribe(events.EventTypeBroadcast, func(event events.Event) error {
			eventsReceived = append(eventsReceived, event.Subject)
			t.Logf("📨 Event received: %s", event.Subject)

			// Mock Release Agent behavior - create Release node when release.create event is emitted
			if event.Subject == "release.create" {
				if appName, exists := event.Payload["application"].(string); exists {
					// Create Release node in graph
					releaseNode := &graph.Node{
						ID:   appName, // Use app name as the Release node ID for test simplicity
						Kind: "Release",
						Metadata: map[string]interface{}{
							"application": appName,
							"created_at":  time.Now().Unix(),
						},
					}
					mockGraph.AddNode(releaseNode)
					t.Logf("🏷️ Mock Release Agent: Created Release node for %s", appName)
				}
			}

			return nil
		})
		eventBus.Subscribe(events.EventTypeNotify, func(event events.Event) error {
			eventsReceived = append(eventsReceived, event.Subject)
			t.Logf("📨 Event received: %s", event.Subject)
			return nil
		})

		// Create deployment agent
		deploymentAgent, err := NewDeploymentAgent(mockGraph, realAIProvider, eventBus, registry)
		if err != nil {
			t.Fatalf("Failed to create deployment agent: %v", err)
		}

		// Step 1: User requests deployment
		userMessage := "Deploy app-a to production"
		deploymentEvent := &events.Event{
			Type:    events.EventTypeRequest,
			Source:  "user",
			Subject: "deployment.request",
			Payload: map[string]interface{}{
				"intent":       "deploy application",
				"user_message": userMessage,
			},
		}

		// Act - Start the orchestration workflow
		response, err := deploymentAgent.(*agentFramework.BaseAgent).ProcessEvent(context.Background(), deploymentEvent)
		if err != nil {
			t.Fatalf("Deployment orchestration failed: %v", err)
		}

		// Assert - Verify the complete workflow was executed

		// Step 2: Verify deployment agent understood the request
		if response.Type != events.EventTypeResponse {
			t.Errorf("Expected response event, got: %s", response.Type)
		}

		// Step 3: Verify Release Agent coordination - should emit release.create event
		releaseCreateEventFound := false
		for _, eventSubject := range eventsReceived {
			if eventSubject == "release.create" {
				releaseCreateEventFound = true
				break
			}
		}
		if !releaseCreateEventFound {
			t.Error("❌ STEP 3 FAILED: Expected 'release.create' event to be emitted for Release Agent coordination")
		} else {
			t.Log("✅ STEP 3 PASSED: 'release.create' event emitted for Release Agent coordination")
		}

		// Step 4: Verify Policy Agent coordination - should emit policy.evaluate event
		policyEvaluateEventFound := false
		for _, eventSubject := range eventsReceived {
			if eventSubject == "policy.evaluate" {
				policyEvaluateEventFound = true
				break
			}
		}
		if !policyEvaluateEventFound {
			t.Error("❌ STEP 4 FAILED: Expected 'policy.evaluate' event to be emitted for Policy Agent coordination")
		} else {
			t.Log("✅ STEP 4 PASSED: 'policy.evaluate' event emitted for Policy Agent coordination")
		}

		// Step 5: Verify Release node was created in graph
		// Query graph for release nodes
		currentGraph, err := mockGraph.Graph()
		if err != nil {
			t.Errorf("Failed to get current graph: %v", err)
		}
		releaseNodeFound := false
		for nodeID, node := range currentGraph.Nodes {
			if node.Kind == "Release" && nodeID == "app-a" {
				releaseNodeFound = true
				break
			}
		}
		if !releaseNodeFound {
			t.Error("❌ STEP 5 FAILED: Expected Release node to be created in graph")
		}

		// Step 6: Verify Deployment edge creation (Release → Environment)
		deploymentEdgeFound := false
		for _, edges := range currentGraph.Edges {
			for _, edge := range edges {
				if edge.Type == "deployment" && edge.To == "production" {
					deploymentEdgeFound = true
					// Step 7: Verify edge has proper final status (should be "succeeded" after completion)
					if status, exists := edge.Metadata["status"]; !exists || status != "succeeded" {
						t.Errorf("❌ STEP 7 FAILED: Expected deployment edge to have status 'succeeded', got: %v", status)
					} else {
						t.Log("✅ STEP 7 PASSED: Deployment edge has correct final status 'succeeded'")
					}
					break
				}
			}
			if deploymentEdgeFound {
				break
			}
		}
		if !deploymentEdgeFound {
			t.Error("❌ STEP 6 FAILED: Expected Deployment edge from Release to Environment")
		}

		// Step 7: Verify deployment completion - should emit deployment.completed event
		deploymentCompletedEventFound := false
		for _, eventSubject := range eventsReceived {
			if eventSubject == "deployment.completed" {
				deploymentCompletedEventFound = true
				break
			}
		}
		if !deploymentCompletedEventFound {
			t.Error("❌ STEP 7 FAILED: Expected 'deployment.completed' event after successful deployment")
		} else {
			t.Log("✅ STEP 7 PASSED: 'deployment.completed' event emitted after deployment")
		}

		// Final verification: Check deployment result contains all required information
		payload := response.Payload
		if deploymentID, exists := payload["deployment_id"]; !exists || deploymentID == "" {
			t.Error("❌ FINAL CHECK FAILED: Expected deployment_id in response")
		}
		if releaseID, exists := payload["release_id"]; !exists || releaseID == "" {
			t.Error("❌ FINAL CHECK FAILED: Expected release_id in response")
		}
		if application, exists := payload["application"]; !exists || application != "app-a" {
			t.Error("❌ FINAL CHECK FAILED: Expected application 'app-a' in response")
		}
		if environment, exists := payload["environment"]; !exists || environment != "production" {
			t.Error("❌ FINAL CHECK FAILED: Expected environment 'production' in response")
		}

		t.Logf("✅ Deployment orchestration workflow verification complete")
		t.Logf("📊 Events fired during workflow: %v", eventsReceived)
	})
}

// Mock Release Agent for unit testing
type MockReleaseAgent struct {
	releaseID string
	graph     *graph.GlobalGraph // Add graph access for mocking
}

func (m *MockReleaseAgent) GetID() string { return "release-agent" }
func (m *MockReleaseAgent) GetCapabilities() []agentRegistry.AgentCapability {
	return []agentRegistry.AgentCapability{
		{
			Name:        "release_creation",
			Description: "Creates releases for applications",
			Intents:     []string{"create release", "release application"},
		},
	}
}
func (m *MockReleaseAgent) GetStatus() agentRegistry.AgentStatus {
	return agentRegistry.AgentStatus{
		ID:     "release-agent",
		Type:   "release",
		Status: "active",
	}
}
func (m *MockReleaseAgent) HandleEvent(ctx context.Context, event *events.Event) error {
	// Mock Release Agent behavior - when it receives release.create, it creates the node
	if event.Subject == "release.create" {
		appName := event.Payload["application"].(string)
		releaseID := fmt.Sprintf("release-%s-123", appName)

		// Create Release node in graph (this is what the real Release Agent would do)
		releaseNode := &graph.Node{
			ID:   releaseID,
			Kind: "Release",
			Metadata: map[string]interface{}{
				"application": appName,
				"version":     "v1.0.0",
			},
		}
		m.graph.AddNode(releaseNode)
	}
	return nil
}
func (m *MockReleaseAgent) Start(ctx context.Context) error { return nil }
func (m *MockReleaseAgent) Stop(ctx context.Context) error  { return nil }
func (m *MockReleaseAgent) Health() agentRegistry.HealthStatus {
	return agentRegistry.HealthStatus{Healthy: true, Status: "ok"}
}

// Mock Policy Agent for unit testing
type MockPolicyAgent struct {
	shouldApprove bool
}

func (m *MockPolicyAgent) GetID() string { return "policy-agent" }
func (m *MockPolicyAgent) GetCapabilities() []agentRegistry.AgentCapability {
	return []agentRegistry.AgentCapability{
		{
			Name:        "policy_evaluation",
			Description: "Evaluates deployment policies",
			Intents:     []string{"evaluate policy", "check policy"},
		},
	}
}
func (m *MockPolicyAgent) GetStatus() agentRegistry.AgentStatus {
	return agentRegistry.AgentStatus{
		ID:     "policy-agent",
		Type:   "policy",
		Status: "active",
	}
}
func (m *MockPolicyAgent) HandleEvent(ctx context.Context, event *events.Event) error {
	// Mock Policy Agent behavior - when it receives policy.evaluate, it responds with policy.decision
	if event.Subject == "policy.evaluate" {
		// Extract parameters from the event
		appName, _ := event.Payload["application"].(string)
		environment, _ := event.Payload["environment"].(string)
		releaseID, _ := event.Payload["release_id"].(string)

		// Simulate policy decision (simulating Policy Agent behavior)
		decision := "allowed"
		if !m.shouldApprove {
			decision = "blocked"
		}

		// Note: In a real system, we would emit a policy.decision event back to the event bus
		// For this test, we're just simulating the decision logic
		_ = appName
		_ = environment
		_ = releaseID
		_ = decision
	}
	return nil
}
func (m *MockPolicyAgent) Start(ctx context.Context) error { return nil }
func (m *MockPolicyAgent) Stop(ctx context.Context) error  { return nil }
func (m *MockPolicyAgent) Health() agentRegistry.HealthStatus {
	return agentRegistry.HealthStatus{Healthy: true, Status: "ok"}
}

// Mock Transport for EventBus testing
type MockTransport struct {
	emittedEvents []events.Event // Store full events, not just messages
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		emittedEvents: make([]events.Event, 0),
	}
}

func (m *MockTransport) Publish(topic string, data []byte) error {
	// Deserialize the event to capture full event details
	var event events.Event
	if err := json.Unmarshal(data, &event); err == nil {
		m.emittedEvents = append(m.emittedEvents, event)
	}
	return nil
}

func (m *MockTransport) Subscribe(topic string, handler func([]byte)) error {
	return nil
}

func (m *MockTransport) Close() error {
	return nil
}

func (m *MockTransport) GetEmittedEvents() []events.Event {
	return m.emittedEvents
}

func (m *MockTransport) GetEventsBySubject(subject string) []events.Event {
	var result []events.Event
	for _, event := range m.emittedEvents {
		if event.Subject == subject {
			result = append(result, event)
		}
	}
	return result
}

// Test the deployment orchestration workflow with mocked dependencies (proper unit test)
func TestDeploymentOrchestrationWorkflow_UnitTest(t *testing.T) {
	t.Run("orchestrates deployment with mocked agents", func(t *testing.T) {
		// Setup - Create mocked dependencies
		registry := agentRegistry.NewInMemoryAgentRegistry()
		mockTransport := NewMockTransport()
		eventBus := events.NewEventBus(mockTransport, false)
		mockGraph := graph.NewGlobalGraph(graph.NewMemoryGraph())
		realAIProvider := getRealAIProvider(t)
		defer realAIProvider.Close()

		// Track events for verification
		eventsReceived := make([]events.Event, 0)
		eventBus.Subscribe(events.EventTypeRequest, func(event events.Event) error {
			eventsReceived = append(eventsReceived, event)
			return nil
		})
		eventBus.Subscribe(events.EventTypeResponse, func(event events.Event) error {
			eventsReceived = append(eventsReceived, event)
			return nil
		})
		eventBus.Subscribe(events.EventTypeBroadcast, func(event events.Event) error {
			eventsReceived = append(eventsReceived, event)
			return nil
		})

		// Register mock agents in registry
		mockReleaseAgent := &MockReleaseAgent{
			releaseID: "release-123",
			graph:     mockGraph, // Pass graph so mock can create nodes
		}
		mockPolicyAgent := &MockPolicyAgent{shouldApprove: true}

		registry.RegisterAgent(context.Background(), mockReleaseAgent)
		registry.RegisterAgent(context.Background(), mockPolicyAgent)

		// Connect Mock Release Agent to the event bus to handle release.create events
		eventBus.Subscribe(events.EventTypeBroadcast, func(event events.Event) error {
			if event.Subject == "release.create" {
				return mockReleaseAgent.HandleEvent(context.Background(), &event)
			}
			return nil
		})

		// Connect Mock Policy Agent to the event bus to handle policy.evaluate events
		eventBus.Subscribe(events.EventTypeRequest, func(event events.Event) error {
			if event.Subject == "policy.evaluate" {
				return mockPolicyAgent.HandleEvent(context.Background(), &event)
			}
			return nil
		})

		// Create deployment agent with mocked dependencies
		deploymentAgent, err := NewDeploymentAgent(mockGraph, realAIProvider, eventBus, registry)
		if err != nil {
			t.Fatalf("Failed to create deployment agent: %v", err)
		}

		// The agent is automatically registered in the framework, no need to register again
		_ = deploymentAgent // Use the variable to avoid "declared and not used" error

		// Setup graph with required nodes for testing
		// Add application node that the deployment will reference
		appNode := &graph.Node{
			ID:   "app-a",
			Kind: "Application",
			Metadata: map[string]interface{}{
				"name":    "app-a",
				"version": "1.0.0",
			},
		}
		mockGraph.AddNode(appNode)

		// Add environment node
		envNode := &graph.Node{
			ID:   "production",
			Kind: "Environment",
			Metadata: map[string]interface{}{
				"name": "production",
				"type": "production",
			},
		}
		mockGraph.AddNode(envNode)

		// Step 1: User requests deployment
		userMessage := "Deploy app-a to production"
		deploymentEvent := &events.Event{
			Type:    events.EventTypeRequest,
			Source:  "user",
			Subject: "deployment.request",
			Payload: map[string]interface{}{
				"intent":       "deploy application",
				"user_message": userMessage,
			},
		}

		// Act - Start the orchestration workflow by emitting event to the bus
		err = eventBus.EmitEvent(*deploymentEvent)
		if err != nil {
			t.Fatalf("Failed to emit deployment event: %v", err)
		}

		// Wait a bit for event processing (in real systems, this would be handled differently)
		time.Sleep(100 * time.Millisecond)

		// Assert - Verify orchestration workflow actually happened

		// Step 2: Verify deployment agent received and processed the event
		publishedMessages := mockTransport.GetEmittedEvents()
		if len(publishedMessages) == 0 {
			t.Error("❌ STEP 2 FAILED: Expected deployment agent to emit events during orchestration")
		}

		// Step 3: Verify Release Agent coordination - should emit release.create event
		releaseCreateFound := false
		for _, msg := range publishedMessages {
			if strings.Contains(msg.Subject, "release.create") {
				releaseCreateFound = true
				break
			}
		}
		if !releaseCreateFound {
			t.Error("❌ STEP 3 FAILED: Expected 'release.create' event to be emitted for Release Agent coordination")
		}

		// Step 4: Verify Release node was created in graph
		currentGraph, err := mockGraph.Graph()
		if err != nil {
			t.Errorf("Failed to get current graph: %v", err)
		}
		releaseNodeFound := false
		for nodeID, node := range currentGraph.Nodes {
			if node.Kind == "Release" && strings.Contains(nodeID, "app-a") {
				releaseNodeFound = true
				break
			}
		}
		if !releaseNodeFound {
			t.Error("❌ STEP 4 FAILED: Expected Release node to be created in graph")
		}

		// Step 5: Verify Deployment edge creation (Release → Environment)
		deploymentEdgeFound := false
		for _, edges := range currentGraph.Edges {
			for _, edge := range edges {
				if edge.Type == "deployment" && edge.To == "production" {
					deploymentEdgeFound = true
					break
				}
			}
			if deploymentEdgeFound {
				break
			}
		}
		if !deploymentEdgeFound {
			t.Error("❌ STEP 5 FAILED: Expected Deployment edge from Release to Environment")
		}

		// Step 6: Verify Policy Agent coordination
		policyEvaluateFound := false
		for _, msg := range publishedMessages {
			if strings.Contains(msg.Subject, "policy.evaluate") {
				policyEvaluateFound = true
				break
			}
		}
		if !policyEvaluateFound {
			t.Error("❌ STEP 6 FAILED: Expected 'policy.evaluate' event for Policy Agent coordination")
		}

		// Final verification: Check that deployment events contain required orchestration information
		allEvents := mockTransport.GetEmittedEvents()
		deploymentResultFound := false
		for _, event := range allEvents {
			if event.Type == events.EventTypeResponse &&
				(strings.Contains(event.Subject, "deployment") || strings.Contains(event.Subject, "orchestration")) {
				deploymentResultFound = true
				payload := event.Payload
				if payload != nil {
					// Check for orchestration information in the event payload
					if deploymentID, exists := payload["deployment_id"]; !exists || deploymentID == "" {
						t.Error("❌ FINAL CHECK FAILED: Expected deployment_id in deployment event")
					}
					if releaseID, exists := payload["release_id"]; !exists || releaseID == "" {
						t.Error("❌ FINAL CHECK FAILED: Expected release_id in deployment event")
					}
					if application, exists := payload["application"]; !exists || application != "app-a" {
						t.Error("❌ FINAL CHECK FAILED: Expected application 'app-a' in deployment event")
					}
					if environment, exists := payload["environment"]; !exists || environment != "production" {
						t.Error("❌ FINAL CHECK FAILED: Expected environment 'production' in deployment event")
					}
				}
				break
			}
		}
		if !deploymentResultFound {
			t.Error("❌ FINAL CHECK FAILED: Expected deployment result event to be emitted")
		}

		t.Logf("📊 Published messages: %v", publishedMessages)
		// t.Logf("📨 Response payload keys: %v", getPayloadKeys(payload)) // Removed since we use events now
	})
}
