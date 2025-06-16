package deployments

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

func TestEngine_ExecuteApplicationDeployment(t *testing.T) {
	// Initialize event system for tests
	eventTransport := events.NewMemoryTransport()
	events.InitializeEventBus(eventTransport)

	// Create test graph
	globalGraph := &graph.GlobalGraph{
		Backend: graph.NewMemoryGraph(),
	}

	// Setup test application
	setupTestApplication(globalGraph)

	// For tests, create a real AI agent or skip if not available
	agent, err := createTestAIAgent(globalGraph)
	if err != nil {
		t.Skipf("AI platform agent not available for testing: %v", err)
	}

	// Create deployment engine with AI platform agent
	engine := NewEngine(globalGraph, agent)

	t.Run("Successful deployment", func(t *testing.T) {
		result, err := engine.ExecuteApplicationDeployment("test-app", "dev")
		if err != nil {
			t.Fatalf("Expected successful deployment, got error: %v", err)
		}

		if result.Application != "test-app" {
			t.Errorf("Expected application 'test-app', got '%s'", result.Application)
		}
		if result.Environment != "dev" {
			t.Errorf("Expected environment 'dev', got '%s'", result.Environment)
		}
		if !result.Summary.Success {
			t.Errorf("Expected successful deployment, got failed")
		}
	})

	t.Run("Application not found", func(t *testing.T) {
		_, err := engine.ExecuteApplicationDeployment("non-existent", "dev")
		if err == nil {
			t.Fatal("Expected error for non-existent application")
		}
		if err.Error() != "application 'non-existent' not found" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("Environment not found", func(t *testing.T) {
		_, err := engine.ExecuteApplicationDeployment("test-app", "non-existent")
		if err == nil {
			t.Fatal("Expected error for non-existent environment")
		}
		if err.Error() != "environment 'non-existent' not found" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("Unauthorized environment", func(t *testing.T) {
		_, err := engine.ExecuteApplicationDeployment("test-app", "prod")
		if err == nil {
			t.Fatal("Expected error for unauthorized environment")
		}
		expectedError := "application 'test-app' is not allowed to deploy to environment 'prod'"
		if err.Error() != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func setupTestApplication(globalGraph *graph.GlobalGraph) {
	// Create application
	appNode := &graph.Node{
		ID:   "test-app",
		Kind: "application",
		Metadata: map[string]interface{}{
			"name":  "test-app",
			"owner": "test-team",
		},
	}
	globalGraph.AddNode(appNode)

	// Create services
	serviceANode := &graph.Node{
		ID:   "service-a",
		Kind: "service",
		Metadata: map[string]interface{}{
			"name": "service-a",
		},
	}
	serviceBNode := &graph.Node{
		ID:   "service-b",
		Kind: "service",
		Metadata: map[string]interface{}{
			"name": "service-b",
		},
	}
	globalGraph.AddNode(serviceANode)
	globalGraph.AddNode(serviceBNode)

	// Create service versions
	serviceAv1Node := &graph.Node{
		ID:   "service-a:1.0.0",
		Kind: "service_version",
		Metadata: map[string]interface{}{
			"name":    "service-a",
			"version": "1.0.0",
		},
	}
	serviceBv1Node := &graph.Node{
		ID:   "service-b:1.0.0",
		Kind: "service_version",
		Metadata: map[string]interface{}{
			"name":    "service-b",
			"version": "1.0.0",
		},
	}
	globalGraph.AddNode(serviceAv1Node)
	globalGraph.AddNode(serviceBv1Node)

	// Create environments
	devNode := &graph.Node{
		ID:   "dev",
		Kind: "environment",
		Metadata: map[string]interface{}{
			"name": "dev",
		},
	}
	prodNode := &graph.Node{
		ID:   "prod",
		Kind: "environment",
		Metadata: map[string]interface{}{
			"name": "prod",
		},
	}
	globalGraph.AddNode(devNode)
	globalGraph.AddNode(prodNode)

	// Setup relationships
	globalGraph.AddEdge("test-app", "service-a", "owns")
	globalGraph.AddEdge("test-app", "service-b", "owns")
	globalGraph.AddEdge("service-a", "service-a:1.0.0", "has_version")
	globalGraph.AddEdge("service-b", "service-b:1.0.0", "has_version")

	// Setup environment access (only allow dev for testing)
	globalGraph.AddEdge("test-app", "dev", "allowed_in")
}

// createTestAIAgent creates a real AI agent for unit tests
func createTestAIAgent(globalGraph *graph.GlobalGraph) (*ai.V3Agent, error) {
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Create real OpenAI provider
	config := &ai.OpenAIConfig{
		APIKey:      apiKey,
		Model:       "gpt-4o-mini",
		BaseURL:     "https://api.openai.com/v1",
		Timeout:     90 * time.Second,
		MaxTokens:   4000,
		Temperature: 0.1,
	}
	aiProvider, err := ai.NewOpenAIProvider(config, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
	}

	// Create V3Agent with minimal dependencies for testing
	eventBus := events.NewEventBus(nil, false)
	agent := ai.NewV3Agent(
		aiProvider,
		globalGraph,
		eventBus, // eventBus
		nil,      // agentRegistry - use nil for tests
	)

	return agent, nil
}
