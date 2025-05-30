package graph

import (
	"testing"
)

func TestAddEdge_ValidAndInvalidTypes(t *testing.T) {
	g := NewGraph()
	n1 := &Node{ID: "a", Kind: KindApplication, Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}}
	n2 := &Node{ID: "b", Kind: KindService, Metadata: map[string]interface{}{}, Spec: map[string]interface{}{}}
	g.AddNode(n1)
	g.AddNode(n2)

	// Valid edge type - applications can own services
	err := g.AddEdge("a", "b", EdgeTypeOwns)
	if err != nil {
		t.Errorf("expected valid edge type, got error: %v", err)
	}

	// Invalid edge type
	err = g.AddEdge("a", "b", "not_allowed")
	if err == nil {
		t.Error("expected error for invalid edge type, got nil")
	}
}

func TestEdge_MetadataState(t *testing.T) {
	edge := Edge{To: "b", Type: EdgeTypeDeploy, Metadata: map[string]interface{}{}}
	edge.Metadata["state"] = "deploying"
	if edge.Metadata["state"] != "deploying" {
		t.Errorf("expected state 'deploying', got %v", edge.Metadata["state"])
	}
}
