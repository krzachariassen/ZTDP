package policies

import (
	"testing"
)

func TestPolicyService_EvaluateGraphPolicy_MaxAppsPerCustomer(t *testing.T) {
	t.Run("evaluates graph-wide policy with AI reasoning", func(t *testing.T) {
		_, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup graph policy
		policy := createMaxAppsPerCustomerPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Note: EvaluateGraphPolicy requires a graph parameter
		// For this test, we'll create a simple mock graph or skip if not available
		t.Skip("Graph policy evaluation requires graph parameter - implement when graph structure is available")
	})

	t.Run("handles graph policies with complex relationships", func(t *testing.T) {
		_, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Create a more complex graph policy
		policy := &Policy{
			ID:                  "complex-graph-policy",
			Name:                "Complex Graph Analysis",
			Description:         "Analyze complex graph relationships",
			Scope:               PolicyScopeGraph,
			NaturalLanguageRule: "The system should maintain proper architectural boundaries and prevent circular dependencies",
			Enforcement:         EnforcementWarn,
			RequiredConfidence:  0.7,
			Enabled:             true,
		}
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// AI should analyze the entire graph structure for architectural issues
		t.Skip("Complex graph policy evaluation requires graph parameter - implement when graph structure is available")
	})
}

func TestPolicyService_PolicyStore_GraphPolicies(t *testing.T) {
	t.Run("stores and retrieves graph policies", func(t *testing.T) {
		_, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Store a graph policy
		policy := createMaxAppsPerCustomerPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store graph policy: %v", err)
		}

		// Retrieve graph policies
		graphPolicies, err := store.GetGraphPolicies()
		if err != nil {
			t.Fatalf("Failed to retrieve graph policies: %v", err)
		}

		if len(graphPolicies) == 0 {
			t.Error("Expected at least one graph policy")
		}

		found := false
		for _, p := range graphPolicies {
			if p.ID == policy.ID {
				found = true
				break
			}
		}

		if !found {
			t.Error("Graph policy not found in retrieved policies")
		}
	})

	t.Run("filters policies by scope correctly", func(t *testing.T) {
		_, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Store different types of policies
		graphPolicy := createMaxAppsPerCustomerPolicy()
		err := store.Store(graphPolicy)
		if err != nil {
			t.Fatalf("Failed to store graph policy: %v", err)
		}

		nodePolicy := createApplicationServiceLimitPolicy()
		err = store.Store(nodePolicy)
		if err != nil {
			t.Fatalf("Failed to store node policy: %v", err)
		}

		edgePolicy := createNoDirectProdDeploymentPolicy()
		err = store.Store(edgePolicy)
		if err != nil {
			t.Fatalf("Failed to store edge policy: %v", err)
		}

		// Retrieve only graph policies
		graphPolicies, err := store.GetGraphPolicies()
		if err != nil {
			t.Fatalf("Failed to retrieve graph policies: %v", err)
		}

		// Should only return the graph policy
		if len(graphPolicies) != 1 {
			t.Errorf("Expected 1 graph policy, got %d", len(graphPolicies))
		}

		if graphPolicies[0].Scope != PolicyScopeGraph {
			t.Errorf("Expected graph scope, got %s", graphPolicies[0].Scope)
		}
	})
}
