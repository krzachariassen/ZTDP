// Package example demonstrates how to use the simplified policy system in ZTDP
package example

import (
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

// Note: Legacy policy validators have been removed in favor of graph-based policies.
// Policy validation is now handled directly by the graph's IsTransitionAllowed method.
// This example demonstrates the modern approach to policy enforcement.

// Example usage function to demonstrate the integration
func ExampleIntegration() {
	// Create a graph backend
	backend := graph.NewMemoryGraph()

	// Create a global graph
	globalGraph := graph.NewGlobalGraph(backend)

	// Create a graph store
	graphStore := graph.NewGraphStore(backend)

	// Create the simplified policy evaluator
	// Note how it only needs the graph store and environment name
	policyEvaluator := policies.NewPolicyEvaluator(graphStore, "default")

	// Example of validating a transition
	fromNode := "service-version-1.0.0"
	toNode := "prod-environment"
	edgeType := graph.EdgeTypeDeploy
	user := "admin"

	// Validate the transition
	err := policyEvaluator.ValidateTransition(fromNode, toNode, edgeType, user)
	if err != nil {
		fmt.Printf("Policy evaluation failed: %v\n", err)
		return
	}

	fmt.Println("Transition allowed by policy")

	// To create a policy, add a policy node to the graph
	policyNode := &graph.Node{
		ID:   "policy-example",
		Kind: graph.KindPolicy,
		Metadata: map[string]interface{}{
			"name":        "Example Policy",
			"description": "Demonstrates policy creation",
			"type":        graph.PolicyTypeSystem,
			"status":      "active",
		},
		Spec: map[string]interface{}{
			"sourceKind": graph.KindServiceVersion,
			"targetKind": graph.KindEnvironment,
			// Add policy-specific configuration
		},
	}

	// Access the default graph
	g, _ := globalGraph.Graph, error(nil)

	// Add the policy node
	g.AddNode(policyNode)

	// Attach the policy to a transition
	g.AttachPolicyToTransition(fromNode, toNode, edgeType, policyNode.ID)

	// Create a check that satisfies the policy
	checkNode := &graph.Node{
		ID:   "check-example",
		Kind: graph.KindCheck,
		Metadata: map[string]interface{}{
			"name":   "Example Check",
			"type":   "verification",
			"status": graph.CheckStatusSucceeded,
		},
		Spec: map[string]interface{}{
			"verifiedBy": "system",
		},
	}

	// Add the check node
	g.AddNode(checkNode)

	// Mark the policy as satisfied by the check
	g.AddEdge(checkNode.ID, policyNode.ID, graph.EdgeTypeSatisfies)
}
