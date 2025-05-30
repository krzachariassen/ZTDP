package graph

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

func TestGlobalGraph_Apply_MemoryBackend(t *testing.T) {
	backend := NewMemoryGraph()
	gg := NewGlobalGraph(backend)

	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "checkout",
			Owner: "team-x",
		},
		Spec: contracts.ApplicationSpec{
			Description: "Checkout app",
			Tags:        []string{"payments"},
			Lifecycle:   map[string]contracts.LifecycleDefinition{},
		},
	}
	appNode, _ := ResolveContract(app)
	gg.AddNode(appNode)

	svc := contracts.ServiceContract{
		Metadata: contracts.Metadata{
			Name:  "checkout-api",
			Owner: "team-x",
		},
		Spec: struct {
			Application string `json:"application"`
			Port        int    `json:"port"`
			Public      bool   `json:"public"`
		}{
			Application: "checkout",
			Port:        8080,
			Public:      true,
		},
	}
	svcNode, _ := ResolveContract(svc)
	gg.AddNode(svcNode)

	if err := gg.AddEdge("checkout", "checkout-api", "owns"); err != nil {
		t.Fatalf("failed to add edge: %v", err)
	}

	// Save the graph to backend
	if err := gg.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Get the current graph state
	currentGraph, err := gg.Graph()
	if err != nil {
		t.Fatalf("getting current graph failed: %v", err)
	}

	if len(currentGraph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(currentGraph.Nodes))
	}
	if len(currentGraph.Edges["checkout"]) != 1 || currentGraph.Edges["checkout"][0].To != "checkout-api" || currentGraph.Edges["checkout"][0].Type != "owns" {
		t.Errorf("expected edge checkout --> checkout-api not found")
	}
}
