package graph

import (
	"os"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

func TestGlobalGraph_Apply_RedisBackend(t *testing.T) {
	addr := os.Getenv("REDIS_HOST")
	if addr == "" {
		t.Skip("REDIS_HOST not set, skipping Redis backend test")
	}
	backend := NewRedisGraph(RedisGraphConfig{Addr: addr, Password: os.Getenv("REDIS_PASSWORD")})
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

	if err := gg.AddEdge("checkout-api", "checkout", "owns"); err != nil {
		t.Fatalf("failed to add edge: %v", err)
	}

	applied, err := gg.Apply("dev")
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	if len(applied.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(applied.Nodes))
	}
	if len(applied.Edges["checkout-api"]) != 1 || applied.Edges["checkout-api"][0].To != "checkout" || applied.Edges["checkout-api"][0].Type != "owns" {
		t.Errorf("expected edge checkout-api --owns--> checkout not found")
	}
}
