package planner

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

func TestExtractApplicationSubgraph(t *testing.T) {
	g := graph.NewGraph()
	// Application node
	g.AddNode(&graph.Node{ID: "app1", Kind: "application", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	// Service nodes
	g.AddNode(&graph.Node{ID: "svc1", Kind: "service", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	g.AddNode(&graph.Node{ID: "svc2", Kind: "service", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})
	// Resource node (this should be a resource instance, not catalog resource)
	g.AddNode(&graph.Node{
		ID:   "db1",
		Kind: "resource",
		Metadata: map[string]interface{}{
			"application": "app1",
			"catalog_ref": "db-catalog",
		},
		Spec: map[string]interface{}{},
	})
	// Edges
	g.AddEdge("app1", "svc1", "owns")
	g.AddEdge("app1", "svc2", "owns")
	g.AddEdge("svc1", "db1", "uses")
	g.AddEdge("svc2", "db1", "uses")
	// Unrelated node
	g.AddNode(&graph.Node{ID: "other", Kind: "application", Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}})

	// Use the ExtractApplicationSubgraph utility
	subgraph := ExtractApplicationSubgraph("app1", g)

	if len(subgraph.Nodes) != 4 {
		t.Errorf("expected 4 nodes in subgraph, got %d", len(subgraph.Nodes))
	}
	if _, ok := subgraph.Nodes["other"]; ok {
		t.Error("unrelated node 'other' should not be in subgraph")
	}
	if len(subgraph.Edges["app1"]) != 2 {
		t.Error("expected 2 edges from app1 in subgraph")
	}
	if len(subgraph.Edges["svc1"]) != 1 || len(subgraph.Edges["svc2"]) != 1 {
		t.Error("expected 1 edge from each service in subgraph")
	}
}
