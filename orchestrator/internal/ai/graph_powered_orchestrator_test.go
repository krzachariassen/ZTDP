package ai

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestGraphPoweredOrchestrator demonstrates the graph-powered AI approach
func TestGraphPoweredOrchestrator(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Setup graph-powered orchestrator
	orchestrator, testGraph := SetupGraphPoweredOrchestrator(t)

	// Populate graph with test data
	SetupBasicTestGraph(t, testGraph)

	tests := []struct {
		name     string
		userID   string
		input    string
		wantType string
	}{
		{
			name:     "complex deployment with history",
			userID:   "user123",
			input:    "I need to deploy my e-commerce app to production, but I want to make sure it's done safely",
			wantType: "graph_aware_response",
		},
		{
			name:     "follow up question with context",
			userID:   "user123", // Same user - should leverage conversation history
			input:    "What about scaling it for Black Friday traffic?",
			wantType: "contextual_response",
		},
		{
			name:     "capability discovery",
			userID:   "newuser",
			input:    "What can this platform help me with?",
			wantType: "capability_exploration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()

			response, err := orchestrator.ProcessRequest(ctx, tt.input, tt.userID)
			if err != nil {
				t.Fatalf("ProcessRequest failed: %v", err)
			}

			if response.Message == "" {
				t.Error("Expected non-empty message")
			}

			t.Logf("=== %s ===", tt.name)
			t.Logf("User ID: %s", tt.userID)
			t.Logf("Input: %s", tt.input)
			t.Logf("Response: %s", response.Message)
			t.Logf("Confidence: %f", response.Confidence)

			if response.Context != nil {
				t.Logf("Graph Context Available: %v", response.Context["graph_context"] != nil)
			}
		})
	}
}

// TestGraphMemoryAndLearning tests that the AI learns from interactions
func TestGraphMemoryAndLearning(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Setup
	orchestrator, testGraph := SetupGraphPoweredOrchestrator(t)
	SetupBasicTestGraph(t, testGraph)
	ctx := context.Background()
	userID := "learning_user"

	// First interaction - establish context
	t.Log("=== First Interaction - Establishing Context ===")
	response1, err := orchestrator.ProcessRequest(ctx, "I'm working on a Node.js microservice that needs to be deployed", userID)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	t.Logf("First Response: %s", response1.Message)

	// Second interaction - should leverage learned context
	t.Log("\n=== Second Interaction - Leveraging Context ===")
	response2, err := orchestrator.ProcessRequest(ctx, "Now I need to set up monitoring for it", userID)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	t.Logf("Second Response: %s", response2.Message)

	// Third interaction - test learning
	t.Log("\n=== Third Interaction - Testing Learning ===")
	response3, err := orchestrator.ProcessRequest(ctx, "What's the best practice for this type of service?", userID)
	if err != nil {
		t.Fatalf("Third request failed: %v", err)
	}
	t.Logf("Third Response: %s", response3.Message)

	// The AI should show awareness of the accumulated context
	// This is hard to assert programmatically, but we can verify the responses are contextual
	if len(response3.Message) < 50 {
		t.Error("Expected more detailed response showing contextual awareness")
	}
}

// TestGraphExplorationCapabilities tests AI's ability to explore graph dynamically
func TestGraphExplorationCapabilities(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Setup
	orchestrator, testGraph := SetupGraphPoweredOrchestrator(t)
	SetupRichTestGraph(t, testGraph)
	ctx := context.Background()

	// Test AI's ability to discover capabilities through graph exploration
	response, err := orchestrator.ProcessRequest(ctx, "I need to understand what this platform can do for DevOps automation", "explorer_user")
	if err != nil {
		t.Fatalf("Exploration request failed: %v", err)
	}

	t.Log("=== Graph Exploration Response ===")
	t.Logf("Response: %s", response.Message)

	// The AI should demonstrate graph knowledge by mentioning specific agents, workflows, or capabilities
	// This is qualitative verification - the response should show deep system understanding
	if len(response.Message) < 100 {
		t.Error("Expected detailed response showing graph exploration results")
	}
}
