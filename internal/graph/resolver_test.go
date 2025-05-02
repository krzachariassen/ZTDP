package graph

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

func TestResolveContract_Application(t *testing.T) {
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:        "checkout",
			Environment: "dev",
			Owner:       "team-x",
		},
		Spec: struct {
			Description string   `json:"description"`
			Tags        []string `json:"tags,omitempty"`
		}{
			Description: "Handles checkout flows",
			Tags:        []string{"payments", "frontend"},
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
	if node.Metadata.Environment != "dev" {
		t.Errorf("expected environment 'dev', got: %s", node.Metadata.Environment)
	}
}
