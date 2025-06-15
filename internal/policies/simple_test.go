package policies

import (
	"testing"
)

func TestSimple(t *testing.T) {
	t.Log("Simple test starting")

	// Test that we can create a basic policy
	policy := &Policy{
		ID:                  "simple-test",
		Name:                "Simple Test Policy",
		Description:         "A simple test policy",
		Scope:               PolicyScopeNode,
		NaturalLanguageRule: "Test rule",
		Enforcement:         EnforcementBlock,
		Enabled:             true,
	}

	if policy.ID != "simple-test" {
		t.Errorf("Expected policy ID to be 'simple-test', got %s", policy.ID)
	}

	t.Log("Simple test completed successfully")
}
