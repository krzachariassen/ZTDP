package test

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/deployments"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

// TestDirectContractExecution tests when users send contracts directly
func TestDirectContractExecution(t *testing.T) {
	t.Run("V3Agent executes direct deployment contract via intent orchestration", func(t *testing.T) {
		// 1. Setup infrastructure
		backend := graph.NewMemoryGraph()
		graphStore := graph.NewGraphStore(backend)
		globalGraph := graph.NewGlobalGraph(backend)

		eventBus := events.NewEventBus(nil, false)
		agentRegistry := agents.NewInMemoryAgentRegistry()

		// 2. Initialize REAL AI provider
		realAI := createRealAIProviderForTest(t)
		if realAI == nil {
			t.Skip("Skipping direct contract test - set OPENAI_API_KEY environment variable to enable")
		}

		// 3. Create and auto-register agents
		policyService := policies.NewPolicyService(globalGraph)
		_, err := policies.NewPolicyAgent(policyService, eventBus)
		if err != nil {
			t.Fatalf("Failed to create PolicyAgent: %v", err)
		}

		_, err = deployments.NewDeploymentAgent(globalGraph, realAI, "test", &EventBusAdapter{eventBus}, agentRegistry)
		if err != nil {
			t.Fatalf("Failed to create DeploymentAgent: %v", err)
		}

		// 4. Create V3Agent (pure orchestrator)
		v3Agent := ai.NewV3Agent(
			realAI,
			globalGraph,
			eventBus,
			agentRegistry,
		)

		// 5. Test direct contract execution
		ctx := context.Background()

		// User sends a deployment contract directly (like in the API)
		userInput := `Execute this deployment contract: {
			"kind": "deployment",
			"application": "test-app",
			"environment": "production",
			"strategy": "rolling",
			"metadata": {
				"user": "test-user",
				"timestamp": "2025-06-16T08:00:00Z"
			}
		}`

		// 6. Execute via V3Agent Chat method
		response, err := v3Agent.Chat(ctx, userInput)
		if err != nil {
			t.Fatalf("Chat failed: %v", err)
		}

		// 7. Verify it was handled as a contract, not conversation
		if response == nil {
			t.Fatalf("No response received")
		}

		t.Logf("🎯 Response: %s", response.Message)
		t.Logf("🎯 Actions: %+v", response.Actions)

		// Check that it executed the contract rather than asking for clarification
		if len(response.Actions) > 0 {
			action := response.Actions[0]
			if action.Type == "contract_executed" {
				t.Logf("✅ Contract was executed directly via intent-based orchestration")
			} else if action.Type == "conversation_continue" {
				t.Errorf("❌ Contract was not detected - fell back to conversation mode")
			}
		}

		t.Log("✅ Direct contract execution test completed")
	})
}
