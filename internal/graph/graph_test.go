package graph

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

func TestGraph_AddAndGetNode(t *testing.T) {
	g := NewGraph()

	node := &Node{
		ID:   "checkout-api",
		Kind: "service",
		Metadata: contracts.Metadata{
			Name:        "checkout-api",
			Environment: "dev",
			Owner:       "team-x",
		},
		Spec: contracts.ServiceContract{
			Metadata: contracts.Metadata{
				Name:        "checkout-api",
				Environment: "dev",
				Owner:       "team-x",
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
		},
	}

	err := g.AddNode(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := g.GetNode("checkout-api")
	if err != nil {
		t.Fatalf("unexpected error retrieving node: %v", err)
	}

	if got.ID != "checkout-api" || got.Kind != "service" {
		t.Errorf("retrieved node mismatch: got %+v", got)
	}
}

func TestGraph_AddEdge(t *testing.T) {
	g := NewGraph()

	app := &Node{
		ID:   "checkout",
		Kind: "application",
		Metadata: contracts.Metadata{
			Name:        "checkout",
			Environment: "dev",
			Owner:       "team-x",
		},
	}

	svc := &Node{
		ID:   "checkout-api",
		Kind: "service",
		Metadata: contracts.Metadata{
			Name:        "checkout-api",
			Environment: "dev",
			Owner:       "team-x",
		},
	}

	g.AddNode(app)
	g.AddNode(svc)

	err := g.AddEdge("checkout-api", "checkout")
	if err != nil {
		t.Fatalf("unexpected error adding edge: %v", err)
	}

	if len(g.Edges["checkout-api"]) != 1 || g.Edges["checkout-api"][0] != "checkout" {
		t.Errorf("edge not properly stored")
	}
}
