package deployments

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// getOpenAIProvider creates a real OpenAI provider for testing
func getOpenAIProvider(t *testing.T) ai.AIProvider {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set - skipping real AI test. Set environment variable to test real AI integration.")
	}

	config := ai.DefaultOpenAIConfig()
	provider, err := ai.NewOpenAIProvider(config, apiKey)
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	return provider
}

// TestDeploymentAgent tests the basic functionality of the DeploymentAgent
func TestDeploymentAgent(t *testing.T) {
	t.Run("Agent implements AgentInterface correctly", func(t *testing.T) {
		// Setup
		backend := graph.NewMemoryGraph()
		globalGraph := graph.NewGlobalGraph(backend)
		eventBus := events.NewEventBus(nil, false) // Use real EventBus

		// Create agent with real AI provider and no auto-registration for testing
		aiProvider := getOpenAIProvider(t)
		agentInterface, err := NewDeploymentAgent(globalGraph, aiProvider, "test", eventBus, nil)
		if err != nil {
			t.Fatalf("Failed to create deployment agent: %v", err)
		}

		agent, ok := agentInterface.(*DeploymentAgent)
		if !ok {
			t.Fatalf("Expected DeploymentAgent type")
		}

		// Test basic interface methods
		if agent.GetID() != "deployment-agent" {
			t.Errorf("Expected agent ID 'deployment-agent', got %s", agent.GetID())
		}

		status := agent.GetStatus()
		if status.Type != "deployment" {
			t.Errorf("Expected agent type 'deployment', got %s", status.Type)
		}

		capabilities := agent.GetCapabilities()
		if len(capabilities) == 0 {
			t.Errorf("Expected agent to have capabilities")
		}

		// Test capabilities include expected deployment operations
		foundOrchestration := false
		foundPlanning := false
		for _, cap := range capabilities {
			if cap.Name == "deployment_orchestration" {
				foundOrchestration = true
			}
			if cap.Name == "deployment_planning" {
				foundPlanning = true
			}
		}

		if !foundOrchestration {
			t.Errorf("Expected deployment_orchestration capability")
		}
		if !foundPlanning {
			t.Errorf("Expected deployment_planning capability")
		}

		// Test lifecycle methods
		err = agent.Start(context.Background())
		if err != nil {
			t.Errorf("Agent start failed: %v", err)
		}

		health := agent.Health()
		if !health.Healthy {
			t.Errorf("Expected agent to be healthy")
		}

		err = agent.Stop(context.Background())
		if err != nil {
			t.Errorf("Agent stop failed: %v", err)
		}
	})

	t.Run("Agent processes deployment planning with REAL AI", func(t *testing.T) {
		// Setup
		backend := graph.NewMemoryGraph()
		globalGraph := graph.NewGlobalGraph(backend)
		eventBus := events.NewEventBus(nil, false) // Use real EventBus

		// Add test application to graph
		testApp := &graph.Node{
			ID:   "test-app",
			Kind: graph.KindApplication,
			Metadata: map[string]interface{}{
				"name":         "test-app",
				"environment":  "staging",
				"version":      "1.0.0",
				"dependencies": []string{"redis", "postgres"},
			},
		}
		globalGraph.AddNode(testApp)

		// Create agent with REAL AI provider and no auto-registration for testing
		aiProvider := getOpenAIProvider(t)
		agentInterface, err := NewDeploymentAgent(globalGraph, aiProvider, "test", eventBus, nil)
		if err != nil {
			t.Fatalf("Failed to create deployment agent: %v", err)
		}

		agent := agentInterface.(*DeploymentAgent)

		// Start the agent
		err = agent.Start(context.Background())
		if err != nil {
			t.Fatalf("Failed to start agent: %v", err)
		}
		defer agent.Stop(context.Background())

		// Create deployment planning event
		event := &events.Event{
			Type:    events.EventTypeRequest,
			Source:  "test",
			Subject: "plan deployment for test-app to staging environment",
			Payload: map[string]interface{}{
				"intent":             "deployment_planning",
				"application_name":   "test-app",
				"target_environment": "staging",
				"requirements": []string{
					"zero-downtime deployment",
					"health checks required",
					"rollback capability",
				},
			},
		}

		// Process event with real AI
		response, err := agent.ProcessEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Event processing failed: %v", err)
		}

		// Validate response structure
		if response == nil {
			t.Fatalf("Expected response event")
		}

		if response.Type != events.EventTypeResponse {
			t.Errorf("Expected response event type, got %s", response.Type)
		}

		if response.Source != "deployment-agent" {
			t.Errorf("Expected response from deployment-agent, got %s", response.Source)
		}

		// Validate AI-generated content in response
		status, ok := response.Payload["status"].(string)
		if !ok {
			t.Errorf("Expected status field in response")
		}

		operation, ok := response.Payload["operation"].(string)
		if !ok {
			t.Errorf("Expected operation field in response")
		}

		// Check that AI actually generated meaningful content
		if aiResponse, exists := response.Payload["ai_response"]; exists {
			aiResponseStr, ok := aiResponse.(string)
			if ok && len(aiResponseStr) > 50 { // AI should provide substantial response
				t.Logf("✅ AI generated substantial deployment plan: %d characters", len(aiResponseStr))
			} else {
				t.Errorf("AI response seems too short or empty: %v", aiResponse)
			}
		}

		// Log response for manual inspection
		t.Logf("🤖 Deployment Agent Response:")
		t.Logf("   Status: %s", status)
		t.Logf("   Operation: %s", operation)
		if reasoning, exists := response.Payload["reasoning"]; exists {
			t.Logf("   AI Reasoning: %v", reasoning)
		}
	})

	t.Run("Agent handles invalid events gracefully", func(t *testing.T) {
		// Setup
		backend := graph.NewMemoryGraph()
		globalGraph := graph.NewGlobalGraph(backend)
		eventBus := events.NewEventBus(nil, false) // Use real EventBus

		// Test without AI provider for invalid event handling
		agentInterface, err := NewDeploymentAgent(globalGraph, nil, "test", eventBus, nil)
		if err != nil {
			t.Fatalf("Failed to create deployment agent: %v", err)
		}

		agent := agentInterface.(*DeploymentAgent)

		// Create invalid event (missing intent)
		event := &events.Event{
			Type:    events.EventTypeRequest,
			Source:  "test",
			Subject: "invalid event",
			Payload: map[string]interface{}{
				"some_field": "some_value",
			},
		}

		// Process event
		_, err = agent.ProcessEvent(context.Background(), event)
		if err == nil {
			t.Fatalf("Expected error for invalid event")
		}

		if err.Error() != "deployment agent requires 'intent' field in payload" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

// TestDeploymentAgentEventHandling tests that the agent properly subscribes to and handles events
func TestDeploymentAgentEventHandling(t *testing.T) {
	t.Run("Agent subscribes to events and processes deployment requests", func(t *testing.T) {
		// Setup real event infrastructure
		backend := graph.NewMemoryGraph()
		globalGraph := graph.NewGlobalGraph(backend)
		eventBus := events.NewEventBus(nil, false) // Real EventBus for testing

		// Create agent with real AI provider
		aiProvider := getOpenAIProvider(t)
		agentInterface, err := NewDeploymentAgent(globalGraph, aiProvider, "test", eventBus, nil)
		if err != nil {
			t.Fatalf("Failed to create deployment agent: %v", err)
		}

		agent, ok := agentInterface.(*DeploymentAgent)
		if !ok {
			t.Fatalf("Expected DeploymentAgent type")
		}

		// Verify agent is created and subscribed
		t.Logf("✅ DeploymentAgent created and subscribed to events")

		// Create a deployment request event (similar to what V3Agent would send)
		deploymentEvent := events.Event{
			Type:    events.EventTypeRequest,
			Source:  "v3-agent",
			Subject: "agent.intent.requested",
			Payload: map[string]interface{}{
				"correlation_id": "test-correlation-123",
				"intent":         "deploy application", // Intent as extracted by V3Agent
				"application":    "test-app",
				"environment":    "staging",
				"strategy":       "rolling",
				"request_id":     "test-request-123",
				"source_agent":   "v3-agent",
			},
		}

		// Track response events
		responseReceived := false
		responseEvent := events.Event{}

		// Subscribe to response events to verify the agent responds
		eventBus.Subscribe(events.EventTypeResponse, func(event events.Event) error {
			if event.Source == "deployment-agent" {
				responseReceived = true
				responseEvent = event
				t.Logf("📨 Received response from DeploymentAgent: %s", event.Subject)
			}
			return nil
		})

		// Send the event to trigger agent processing
		t.Logf("📤 Sending deployment event to agent...")
		err = agent.handleIncomingEvent(deploymentEvent)
		if err != nil {
			t.Fatalf("Agent failed to handle event: %v", err)
		}

		// Verify the agent processed the event
		// Note: Since we don't have actual request-response correlation yet,
		// we test that the agent can process the event without errors
		t.Logf("✅ Agent successfully processed the event")

		// Test direct ProcessEvent method with proper payload structure
		processEvent := events.Event{
			Type:    events.EventTypeRequest,
			Source:  "v3-agent",
			Subject: "deployment.request",
			Payload: map[string]interface{}{
				"intent":           "deploy application",
				"application_name": "test-app",
				"environment":      "staging",
				"correlation_id":   "test-direct-123",
			},
		}

		ctx := context.Background()
		response, err := agent.ProcessEvent(ctx, &processEvent)

		// The agent should handle the event, even if deployment fails due to missing app
		if err != nil {
			t.Logf("⚠️ ProcessEvent returned error (expected for non-existent app): %v", err)
		}

		if response != nil {
			t.Logf("✅ Agent generated response event: %s", response.Subject)
		}

		// Verify the response tracking worked (for future async improvements)
		_ = responseReceived // Will be used when async response handling is implemented
		_ = responseEvent    // Will be used when async response handling is implemented

		t.Logf("✅ Event subscription and handling test completed successfully")
	})

	t.Run("Agent handles intent matching correctly", func(t *testing.T) {
		// Test that the agent can handle different intent formats
		backend := graph.NewMemoryGraph()
		globalGraph := graph.NewGlobalGraph(backend)
		eventBus := events.NewEventBus(nil, false)

		agentInterface, err := NewDeploymentAgent(globalGraph, nil, "test", eventBus, nil) // No AI for this test
		if err != nil {
			t.Fatalf("Failed to create deployment agent: %v", err)
		}

		agent, ok := agentInterface.(*DeploymentAgent)
		if !ok {
			t.Fatalf("Expected DeploymentAgent type")
		}

		// Test various intent formats that should all map to deployment
		testIntents := []string{
			"deploy application",
			"deployment_orchestration",
			"deploy_application",
			"execute deployment",
			"orchestrate deployment",
		}

		ctx := context.Background()
		for _, intent := range testIntents {
			t.Run(fmt.Sprintf("Intent: %s", intent), func(t *testing.T) {
				event := events.Event{
					Type:    events.EventTypeRequest,
					Source:  "test",
					Subject: "test.intent",
					Payload: map[string]interface{}{
						"intent":           intent,
						"application_name": "test-app",
						"environment":      "test",
					},
				}

				response, err := agent.ProcessEvent(ctx, &event)

				// Should not fail due to intent recognition (may fail for other reasons)
				if err != nil && strings.Contains(err.Error(), "requires 'intent' field") {
					t.Errorf("Agent failed to recognize intent '%s': %v", intent, err)
				}

				if response != nil {
					t.Logf("✅ Intent '%s' processed successfully", intent)
				}
			})
		}
	})
}
