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
	if node.Metadata.Environment != "dev" {
		t.Errorf("expected environment 'dev', got: %s", node.Metadata.Environment)
	}
}

func TestResolveContract_InvalidApplication(t *testing.T) {
	app := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:        "",
			Environment: "dev",
			Owner:       "team-x",
		},
		Spec: contracts.ApplicationSpec{
			Environments: []string{},
			Lifecycle:    map[string]contracts.LifecycleDefinition{},
		},
	}

	_, err := ResolveContract(app)
	if err == nil {
		t.Fatal("expected error for invalid contract, got nil")
	}
}

func TestResolveContract_MultiEnvironment(t *testing.T) {
	envs := []string{"dev", "qa", "prod"}
	for _, env := range envs {
		app := contracts.ApplicationContract{
			Metadata: contracts.Metadata{
				Name:        "checkout",
				Environment: env,
				Owner:       "team-x",
			},
			Spec: contracts.ApplicationSpec{
				Description:  "Handles checkout flows",
				Tags:         []string{"payments", "frontend"},
				Environments: envs,
				Lifecycle:    map[string]contracts.LifecycleDefinition{},
			},
		}

		node, err := ResolveContract(app)
		if err != nil {
			t.Fatalf("unexpected error for env %s: %v", env, err)
		}
		if node.Metadata.Environment != env {
			t.Errorf("expected environment %s, got: %s", env, node.Metadata.Environment)
		}
	}
}
