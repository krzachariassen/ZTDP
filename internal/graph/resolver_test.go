package graph

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

func TestResolveContract_Application(t *testing.T) {
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  "checkout",
			Owner: "team-x",
		},
		Spec: contracts.ApplicationSpec{
			Description:  "Handles checkout flows",
			Tags:         []string{"payments", "frontend"},
			Environments: []string{"dev", "qa"},
			Lifecycle:    map[string]contracts.LifecycleDefinition{},
		},
	}

	node, err := ResolveContract(app)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if node.Kind != "application" {
		t.Errorf("expected kind 'application', got: %s", node.Kind)
	}
	if node.ID != "checkout" {
		t.Errorf("expected ID 'checkout', got: %s", node.ID)
	}
}
