package graph_test

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

func TestPolicyHelpers(t *testing.T) {
	t.Run("AttachPolicyToTransition", func(t *testing.T) {
		g := graph.NewGraph()

		// Create source and target nodes
		srcNode := &graph.Node{
			ID:       "app-1",
			Kind:     graph.KindApplication,
			Metadata: map[string]interface{}{"name": "App 1"},
			Spec:     map[string]interface{}{},
		}

		targetNode := &graph.Node{
			ID:       "env-prod",
			Kind:     graph.KindEnvironment,
			Metadata: map[string]interface{}{"name": "Production"},
			Spec:     map[string]interface{}{},
		}

		policyNode := &graph.Node{
			ID:   "policy-1",
			Kind: graph.KindPolicy,
			Metadata: map[string]interface{}{
				"name":        "Must scan before deployment",
				"description": "Application must be scanned before deployment",
				"type":        graph.PolicyTypeCheck,
			},
			Spec: map[string]interface{}{},
		}

		g.AddNode(srcNode)
		g.AddNode(targetNode)
		g.AddNode(policyNode)

		// Attach policy to transition
		err := g.AttachPolicyToTransition(srcNode.ID, targetNode.ID, graph.EdgeTypeDeploy, policyNode.ID)
		if err != nil {
			t.Fatalf("Failed to attach policy: %v", err)
		}

		// Verify policy is now required for transition
		policies, err := g.FindPoliciesRequiredForTransition(srcNode.ID, targetNode.ID, graph.EdgeTypeDeploy)
		if err != nil {
			t.Fatalf("Failed to find policies: %v", err)
		}

		if len(policies) != 1 {
			t.Errorf("Expected 1 policy, got %d", len(policies))
		}

		if policies[0].ID != policyNode.ID {
			t.Errorf("Expected policy ID %s, got %s", policyNode.ID, policies[0].ID)
		}
	})

	t.Run("IsTransitionAllowed", func(t *testing.T) {
		g := graph.NewGraph()

		// Create source and target nodes
		srcNode := &graph.Node{
			ID:       "app-1",
			Kind:     graph.KindApplication,
			Metadata: map[string]interface{}{"name": "App 1"},
			Spec:     map[string]interface{}{},
		}

		targetNode := &graph.Node{
			ID:       "env-prod",
			Kind:     graph.KindEnvironment,
			Metadata: map[string]interface{}{"name": "Production"},
			Spec:     map[string]interface{}{},
		}

		policyNode := &graph.Node{
			ID:   "policy-1",
			Kind: graph.KindPolicy,
			Metadata: map[string]interface{}{
				"name":        "Must scan before deployment",
				"description": "Application must be scanned before deployment",
				"type":        graph.PolicyTypeCheck,
			},
			Spec: map[string]interface{}{},
		}

		checkNode := &graph.Node{
			ID:   "check-1",
			Kind: graph.KindCheck,
			Metadata: map[string]interface{}{
				"name":   "Security scan",
				"type":   "security_scan",
				"status": graph.CheckStatusPending,
			},
			Spec: map[string]interface{}{},
		}

		g.AddNode(srcNode)
		g.AddNode(targetNode)
		g.AddNode(policyNode)
		g.AddNode(checkNode)

		// Attach policy to transition
		err := g.AttachPolicyToTransition(srcNode.ID, targetNode.ID, graph.EdgeTypeDeploy, policyNode.ID)
		if err != nil {
			t.Fatalf("Failed to attach policy: %v", err)
		}

		// Check that transition is not allowed (no check satisfies policy)
		err = g.IsTransitionAllowed(srcNode.ID, targetNode.ID, graph.EdgeTypeDeploy)
		if err == nil {
			t.Errorf("Expected error for unsatisfied policy, got nil")
		}

		// Mark check as satisfying policy, but still pending
		err = g.MarkPolicySatisfiedByCheck(checkNode.ID, policyNode.ID)
		if err != nil {
			t.Fatalf("Failed to mark policy as satisfied: %v", err)
		}

		// Check still not allowed (check is pending)
		err = g.IsTransitionAllowed(srcNode.ID, targetNode.ID, graph.EdgeTypeDeploy)
		if err == nil {
			t.Errorf("Expected error for unsatisfied policy (pending check), got nil")
		}

		// Update check status to succeeded
		checkNode.Metadata["status"] = graph.CheckStatusSucceeded

		// Check should be allowed now
		err = g.IsTransitionAllowed(srcNode.ID, targetNode.ID, graph.EdgeTypeDeploy)
		if err != nil {
			t.Errorf("Expected no error for satisfied policy, got %v", err)
		}
	})
}
