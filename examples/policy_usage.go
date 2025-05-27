// Package example demonstrates how to use the simplified policy system in ZTDP
package example

import (
	"fmt"

	"github.com/krzachariassen/ZTDP/internal/common"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

// CustomPolicyValidator demonstrates how to implement a validator using the graph-based policy system
type CustomPolicyValidator struct {
	// Add any needed dependencies here
}

// NewCustomPolicyValidator creates a new policy validator
func NewCustomPolicyValidator() *CustomPolicyValidator {
	return &CustomPolicyValidator{}
}

// ValidateMutation implements the common.PolicyValidator interface
// It uses the graph-based policy system to validate mutations
func (v *CustomPolicyValidator) ValidateMutation(view common.GraphView, mutation common.Mutation) error {
	// This is a compatibility adapter for the legacy interface
	// It performs validation using graph-based policies

	// In practice, this would call into graph-specific policy validation
	// The full graph-based policy implementation is handled in the graph package
	// And this adapter merely provides compatibility with the legacy interface

	return nil
}

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
