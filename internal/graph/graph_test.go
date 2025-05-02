package graph

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

func TestGlobalGraph_Apply(t *testing.T) {
	gg := NewGlobalGraph(NewMemoryGraph())

	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "checkout",
			Owner: "team-x",
		},
		Spec: contracts.ApplicationSpec{
			Description:  "Checkout app",
			Tags:         []string{"payments"},
			Environments: []string{"dev", "qa"},
			Lifecycle:    map[string]contracts.LifecycleDefinition{},
		},
	}
	appNode, _ := ResolveContract(app)
	if err := gg.AddNode(appNode); err != nil {
		t.Fatalf("failed to add app node: %v", err)
	}

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
	if err := gg.AddNode(svcNode); err != nil {
		t.Fatalf("failed to add service node: %v", err)
	}
	if err := gg.AddEdge("checkout-api", "checkout"); err != nil {
		t.Fatalf("failed to add edge: %v", err)
	}

	applied, err := gg.Apply("dev")
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	if len(applied.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(applied.Nodes))
	}
	if len(applied.Edges["checkout-api"]) != 1 || applied.Edges["checkout-api"][0] != "checkout" {
		t.Errorf("expected edge checkout-api --> checkout not found")
	}
}
