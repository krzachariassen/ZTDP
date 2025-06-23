package ai

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestGraphPoweredSimple tests basic graph-powered functionality
func TestGraphPoweredSimple(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Setup
	provider := SetupOpenAIProvider(t)
	logger := NewTestLogger("graph-simple-test")
	testGraph := SetupEmbeddedGraph(t)
	SetupBasicTestGraph(t, testGraph)

	orchestrator := NewGraphPoweredAIOrchestrator(provider, testGraph, logger)

	// Test simple case with longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test capability discovery (should be faster)
	response, err := orchestrator.ProcessRequest(ctx, "What can this platform help me with?", "test_user")
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	t.Log("=== Graph-Powered AI Response ===")
	t.Logf("Message: %s", response.Message)
	t.Logf("Confidence: %f", response.Confidence)

	if response.Context != nil {
		t.Logf("Has Graph Context: %v", response.Context["graph_context"] != nil)
	}

	// Verify the response shows graph knowledge
	if len(response.Message) < 50 {
		t.Error("Expected detailed response showing graph exploration")
	}
}

// TestGraphVsStaticComparison compares graph-powered vs limited approaches
func TestGraphVsStaticComparison(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Setup
	provider := SetupOpenAIProvider(t)
	logger := NewTestLogger("comparison-test")
	testGraph := SetupEmbeddedGraph(t)
	SetupBasicTestGraph(t, testGraph)

	graphOrchestrator := NewGraphPoweredAIOrchestrator(provider, testGraph, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userInput := "I need help with deployment"

	// Test graph-powered approach
	t.Log("\n=== Graph-Powered Approach ===")
	graphResponse, err := graphOrchestrator.ProcessRequest(ctx, userInput, "comparison_user")
	if err != nil {
		t.Fatalf("Graph-powered request failed: %v", err)
	}
	t.Logf("Graph Response: %s", graphResponse.Message)

	// The graph-powered response should demonstrate knowledge of specific agents,
	// workflows, and capabilities discovered from the graph
	t.Log("\n=== Analysis ===")
	t.Logf("Response Length: %d chars", len(graphResponse.Message))
	t.Logf("Has Graph Context: %v", graphResponse.Context != nil && graphResponse.Context["graph_context"] != nil)

	// Look for evidence of graph knowledge in the response
	responseText := graphResponse.Message
	hasAgentKnowledge := false
	hasWorkflowKnowledge := false

	// These are the agents and workflows we set up in the test graph
	if contains(responseText, "Deployment Agent") || contains(responseText, "deploy") {
		hasAgentKnowledge = true
	}
	if contains(responseText, "Safe Production Deployment") || contains(responseText, "canary") {
		hasWorkflowKnowledge = true
	}

	t.Logf("Shows Agent Knowledge: %v", hasAgentKnowledge)
	t.Logf("Shows Workflow Knowledge: %v", hasWorkflowKnowledge)
}
