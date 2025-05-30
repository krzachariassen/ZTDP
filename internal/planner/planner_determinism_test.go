package planner

import (
	"fmt"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// TestPlannerDeterminism tests that the planner produces the same plan for identical inputs
func TestPlannerDeterminism(t *testing.T) {
	// Create a graph with multiple nodes that have the same in-degree
	// This should expose any non-deterministic behavior
	createTestGraph := func() *graph.Graph {
		g := graph.NewGraph()

		// Add nodes - these will all have in-degree 0 initially
		g.AddNode(&graph.Node{ID: "app1", Kind: "application", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
		g.AddNode(&graph.Node{ID: "app2", Kind: "application", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
		g.AddNode(&graph.Node{ID: "svc1", Kind: "service", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
		g.AddNode(&graph.Node{ID: "svc2", Kind: "service", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
		g.AddNode(&graph.Node{ID: "svc3", Kind: "service", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
		g.AddNode(&graph.Node{ID: "db1", Kind: "resource", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
		g.AddNode(&graph.Node{ID: "db2", Kind: "resource", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})

		// Add edges to create some dependencies
		g.AddEdge("app1", "svc1", "owns")
		g.AddEdge("app1", "svc2", "owns")
		g.AddEdge("app2", "svc3", "owns")
		g.AddEdge("svc1", "db1", "uses")
		g.AddEdge("svc2", "db1", "uses")
		g.AddEdge("svc3", "db2", "uses")

		return g
	}

	// Run the planner multiple times and ensure we get the same result
	var firstPlan []string
	const numRuns = 10

	for i := 0; i < numRuns; i++ {
		g := createTestGraph()
		p := NewPlanner(g)
		plan, err := p.PlanWithEdgeTypes([]string{"owns", "uses"})
		if err != nil {
			t.Fatalf("Planning failed on run %d: %v", i+1, err)
		}

		if i == 0 {
			firstPlan = plan
			t.Logf("First plan: %v", plan)
		} else {
			if !slicesEqual(plan, firstPlan) {
				t.Errorf("Plan inconsistency detected on run %d:", i+1)
				t.Errorf("  First plan: %v", firstPlan)
				t.Errorf("  Current plan: %v", plan)
				t.Fatal("Planner produces non-deterministic results")
			}
		}
	}

	t.Logf("✅ All %d runs produced identical plans: %v", numRuns, firstPlan)
}

// TestPlannerDeterminismComplexGraph tests with a more complex scenario
func TestPlannerDeterminismComplexGraph(t *testing.T) {
	createComplexGraph := func() *graph.Graph {
		g := graph.NewGraph()

		// Create many nodes with same in-degree to stress test
		for i := 1; i <= 20; i++ {
			g.AddNode(&graph.Node{
				ID:       fmt.Sprintf("node%d", i),
				Kind:     "test",
				Metadata: map[string]interface{}{},
				Spec:     map[string]interface{}{},
			})
		}

		// Add some edges but leave many nodes with the same in-degree
		g.AddEdge("node1", "node10", "deploy")
		g.AddEdge("node2", "node11", "deploy")
		g.AddEdge("node3", "node12", "deploy")
		g.AddEdge("node10", "node15", "deploy")
		g.AddEdge("node11", "node16", "deploy")
		g.AddEdge("node12", "node17", "deploy")

		return g
	}

	// Test determinism with multiple runs
	var firstPlan []string
	const numRuns = 5

	for i := 0; i < numRuns; i++ {
		g := createComplexGraph()
		p := NewPlanner(g)
		plan, err := p.PlanWithEdgeTypes([]string{"deploy"})
		if err != nil {
			t.Fatalf("Planning failed on run %d: %v", i+1, err)
		}

		if i == 0 {
			firstPlan = plan
		} else {
			if !slicesEqual(plan, firstPlan) {
				t.Errorf("Complex graph plan inconsistency on run %d:", i+1)
				t.Errorf("  First plan: %v", firstPlan)
				t.Errorf("  Current plan: %v", plan)
				t.Fatal("Planner produces non-deterministic results for complex graphs")
			}
		}
	}

	t.Logf("✅ Complex graph: All %d runs produced identical plans", numRuns)
}

// Helper function to compare slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
