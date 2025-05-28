package planner

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

func TestPlanner_Plan(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(&graph.Node{ID: "A", Kind: "test", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	g.AddNode(&graph.Node{ID: "B", Kind: "test", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	g.AddNode(&graph.Node{ID: "C", Kind: "test", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	g.AddEdge("A", "B", "deploy")
	g.AddEdge("B", "C", "deploy")

	p := NewPlanner(g)
	order, err := p.Plan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(order))
	}
	// A must come before B, B before C
	pos := map[string]int{}
	for i, id := range order {
		pos[id] = i
	}
	if !(pos["A"] < pos["B"] && pos["B"] < pos["C"]) {
		t.Errorf("incorrect order: %v", order)
	}
}

func TestPlanner_Cycle(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(&graph.Node{ID: "A", Kind: "test", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	g.AddNode(&graph.Node{ID: "B", Kind: "test", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	g.AddEdge("A", "B", "deploy")
	g.AddEdge("B", "A", "deploy")
	p := NewPlanner(g)
	_, err := p.Plan()
	if err == nil {
		t.Error("expected error for cycle, got nil")
	}
}

func TestPlanner_PlanWithEdgeTypes(t *testing.T) {
	g := graph.NewGraph()
	g.AddNode(&graph.Node{ID: "A", Kind: "test", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	g.AddNode(&graph.Node{ID: "B", Kind: "test", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	g.AddNode(&graph.Node{ID: "C", Kind: "test", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	g.AddEdge("A", "B", graph.EdgeTypeDeploy)
	g.AddEdge("B", "C", graph.EdgeTypeCreate)

	p := NewPlanner(g)
	order, err := p.PlanWithEdgeTypes([]string{graph.EdgeTypeDeploy, graph.EdgeTypeCreate})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(order))
	}
	pos := map[string]int{}
	for i, id := range order {
		pos[id] = i
	}
	if !(pos["A"] < pos["B"] && pos["B"] < pos["C"]) {
		t.Errorf("incorrect order: %v", order)
	}

	// Only deploy edge
	order, err = p.PlanWithEdgeTypes([]string{graph.EdgeTypeDeploy})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(order))
	}
	if !(pos["A"] < pos["B"]) {
		t.Errorf("A should come before B with only deploy edge")
	}
}
