package ai

import (
	"context"
	"testing"
	"time"
)

// TestSharedHelpers demonstrates the usage of all shared test helpers
func TestSharedHelpers(t *testing.T) {
	// Test 1: Basic orchestrator setup
	t.Run("Basic Setup", func(t *testing.T) {
		orchestrator, testGraph := SetupGraphPoweredOrchestrator(t)
		SetupBasicTestGraph(t, testGraph)

		// Quick validation
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		response, err := orchestrator.ProcessRequest(ctx, "What agents are available?", "helper_test_user")
		if err != nil {
			t.Fatalf("Basic setup failed: %v", err)
		}

		AssertGraphKnowledge(t, response, "basic setup test")
		t.Logf("Basic setup test passed with response: %s", TruncateString(response.Message, 100))
	})

	// Test 2: Rich graph setup with complex scenarios
	t.Run("Rich Graph Setup", func(t *testing.T) {
		orchestrator, testGraph := SetupGraphPoweredOrchestrator(t)
		SetupRichTestGraph(t, testGraph)

		// Test complex scenarios
		scenarios := GetComplexTestScenarios()

		for _, scenario := range scenarios {
			t.Run(scenario.Name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
				defer cancel()

				response, err := orchestrator.ProcessRequest(ctx, scenario.Input, scenario.UserID)
				if err != nil {
					t.Fatalf("Complex scenario %s failed: %v", scenario.Name, err)
				}

				AssertGraphKnowledge(t, response, scenario.Description)
				t.Logf("Complex scenario %s completed successfully", scenario.Name)
			})
		}
	})

	// Test 3: Basic scenarios with basic graph
	t.Run("Basic Scenarios", func(t *testing.T) {
		orchestrator, testGraph := SetupGraphPoweredOrchestrator(t)
		SetupBasicTestGraph(t, testGraph)

		scenarios := GetBasicTestScenarios()

		for _, scenario := range scenarios {
			t.Run(scenario.Name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				defer cancel()

				response, err := orchestrator.ProcessRequest(ctx, scenario.Input, scenario.UserID)
				if err != nil {
					t.Fatalf("Basic scenario %s failed: %v", scenario.Name, err)
				}

				AssertGraphKnowledge(t, response, scenario.Description)
				t.Logf("Basic scenario %s completed successfully", scenario.Name)
			})
		}
	})
}

// TestHelperConsistency tests that all helpers produce consistent results
func TestHelperConsistency(t *testing.T) {
	// Test that multiple setups work correctly
	t.Run("Multiple Setups", func(t *testing.T) {
		// Setup 1
		orchestrator1, graph1 := SetupGraphPoweredOrchestrator(t)
		SetupBasicTestGraph(t, graph1)

		// Setup 2
		orchestrator2, graph2 := SetupGraphPoweredOrchestrator(t)
		SetupBasicTestGraph(t, graph2)

		ctx := context.Background()

		// Both should respond similarly to the same input
		response1, err1 := orchestrator1.ProcessRequest(ctx, "What can you help with?", "consistency_user1")
		response2, err2 := orchestrator2.ProcessRequest(ctx, "What can you help with?", "consistency_user2")

		if err1 != nil {
			t.Fatalf("First orchestrator failed: %v", err1)
		}
		if err2 != nil {
			t.Fatalf("Second orchestrator failed: %v", err2)
		}

		// Both should have graph knowledge
		AssertGraphKnowledge(t, response1, "first orchestrator")
		AssertGraphKnowledge(t, response2, "second orchestrator")

		t.Logf("Both orchestrators demonstrate consistent graph knowledge")
	})
}

// TestHelperErrorHandling tests error handling in helpers
func TestHelperErrorHandling(t *testing.T) {
	// Test that helpers handle missing OpenAI key gracefully
	t.Run("Missing OpenAI Key", func(t *testing.T) {
		// This test will be skipped if OPENAI_API_KEY is not set
		// We're testing that the helper handles this gracefully
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Helper should not panic, but should skip test: %v", r)
				}
			}()

			// This should either work or skip, but not panic
			_ = SetupOpenAIProvider(t)
		}()
	})
}
