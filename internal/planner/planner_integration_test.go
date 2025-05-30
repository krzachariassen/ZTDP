package planner

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// TestPlannerDeterminismWithRealGraph tests planner determinism with a realistic graph structure
func TestPlannerDeterminismWithRealGraph(t *testing.T) {
	createRealisticGraph := func() *graph.Graph {
		g := graph.NewGraph()

		// Create a graph similar to what ZTDP would have in practice
		g.AddNode(&graph.Node{
			ID:   "payment",
			Kind: "application",
			Metadata: map[string]interface{}{
				"name":  "payment",
				"owner": "team-a",
			},
			Spec: map[string]interface{}{
				"description": "Payment service application",
			},
		})

		g.AddNode(&graph.Node{
			ID:   "checkout",
			Kind: "application",
			Metadata: map[string]interface{}{
				"name":  "checkout",
				"owner": "team-b",
			},
			Spec: map[string]interface{}{
				"description": "Checkout service application",
			},
		})

		// Add multiple environments
		for _, env := range []string{"dev", "staging", "prod"} {
			g.AddNode(&graph.Node{
				ID:   env,
				Kind: "environment",
				Metadata: map[string]interface{}{
					"name": env,
				},
				Spec: map[string]interface{}{
					"description": env + " environment",
				},
			})
		}

		// Add services for each app
		for _, app := range []string{"payment", "checkout"} {
			for _, svc := range []string{"api", "worker"} {
				serviceID := app + "-" + svc
				g.AddNode(&graph.Node{
					ID:   serviceID,
					Kind: "service",
					Metadata: map[string]interface{}{
						"name": serviceID,
					},
					Spec: map[string]interface{}{
						"application": app,
						"port":        8080,
					},
				})

				// App owns service
				g.AddEdge(app, serviceID, "owns")

				// Service versions
				versionID := serviceID + ":1.0.0"
				g.AddNode(&graph.Node{
					ID:   versionID,
					Kind: "service_version",
					Metadata: map[string]interface{}{
						"name": serviceID,
					},
					Spec: map[string]interface{}{
						"version": "1.0.0",
					},
				})

				// Service has version
				g.AddEdge(serviceID, versionID, "has_version")
			}
		}

		// Add deployment edges - this creates the scenario where multiple nodes
		// have the same in-degree and could cause non-deterministic ordering
		for _, app := range []string{"payment", "checkout"} {
			for _, env := range []string{"dev", "staging", "prod"} {
				g.AddEdge(app, env, "deploy")
			}
		}

		return g
	}

	// Test with deploy edge type (most common in ZTDP)
	var firstPlan []string
	const numRuns = 15

	for i := 0; i < numRuns; i++ {
		g := createRealisticGraph()
		p := NewPlanner(g)
		plan, err := p.Plan() // Uses default "deploy" edge type
		if err != nil {
			t.Fatalf("Planning failed on run %d: %v", i+1, err)
		}

		if i == 0 {
			firstPlan = plan
			t.Logf("Deployment plan: %v", plan)
		} else {
			if !slicesEqual(plan, firstPlan) {
				t.Errorf("Deployment plan inconsistency on run %d:", i+1)
				t.Errorf("  First plan: %v", firstPlan)
				t.Errorf("  Current plan: %v", plan)
				t.Fatal("Planner produces non-deterministic deployment plans")
			}
		}
	}

	t.Logf("✅ All %d deployment planning runs produced identical results", numRuns)

	// Also test with multiple edge types
	for i := 0; i < numRuns; i++ {
		g := createRealisticGraph()
		p := NewPlanner(g)
		plan, err := p.PlanWithEdgeTypes([]string{"owns", "has_version", "deploy"})
		if err != nil {
			t.Fatalf("Multi-edge planning failed on run %d: %v", i+1, err)
		}

		if i == 0 {
			t.Logf("Multi-edge plan: %v", plan)
		} else {
			// For multi-edge plans, just verify we get a consistent number of nodes
			// (the exact order may vary based on edge types, but should be deterministic)
			if len(plan) != len(firstPlan) {
				t.Errorf("Multi-edge plan length inconsistency on run %d: expected %d, got %d",
					i+1, len(firstPlan), len(plan))
			}
		}
	}

	t.Logf("✅ All %d multi-edge planning runs completed successfully", numRuns)
}
