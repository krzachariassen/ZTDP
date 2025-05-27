package policies_test

import (
	"testing"

	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

// MockGraphBackend implements graph.GraphBackend for testing
type MockGraphBackend struct {
	graphs map[string]*graph.Graph
	global *graph.Graph
}

func NewMockGraphBackend() *MockGraphBackend {
	return &MockGraphBackend{
		graphs: map[string]*graph.Graph{
			"default": graph.NewGraph(),
		},
		global: graph.NewGraph(),
	}
}

func (m *MockGraphBackend) AddNode(env string, node *graph.Node) error {
	g, ok := m.graphs[env]
	if !ok {
		m.graphs[env] = graph.NewGraph()
		g = m.graphs[env]
	}
	return g.AddNode(node)
}

func (m *MockGraphBackend) AddEdge(env, fromID, toID, relType string) error {
	g, ok := m.graphs[env]
	if !ok {
		return nil // For simplicity in test
	}
	return g.AddEdge(fromID, toID, relType)
}

func (m *MockGraphBackend) GetNode(env, id string) (*graph.Node, error) {
	g, ok := m.graphs[env]
	if !ok {
		return nil, nil // For simplicity in test
	}
	return g.GetNode(id)
}

func (m *MockGraphBackend) GetAll(env string) (*graph.Graph, error) {
	g, ok := m.graphs[env]
	if !ok {
		return graph.NewGraph(), nil
	}
	return g, nil
}

func (m *MockGraphBackend) SaveGlobal(g *graph.Graph) error {
	m.global = g
	return nil
}

func (m *MockGraphBackend) LoadGlobal() (*graph.Graph, error) {
	return m.global, nil
}

// A simple test to debug and verify policy enforcement
func TestGraphPolicyEnforcement(t *testing.T) {
	// Create a memory-backed graph with test backend
	backend := NewMockGraphBackend()
	graphStore := graph.NewGraphStore(backend)
	env := "default"
	
	// Get the environment graph (already exists in mock)
	g, err := backend.GetAll(env)
	if err != nil {
		t.Fatalf("Failed to get graph: %v", err)
	}

	// Create service, environment nodes
	serviceNode := &graph.Node{
		ID:       "test-service:1.0.0",
		Kind:     graph.KindServiceVersion,
		Metadata: map[string]interface{}{"name": "Test Service"},
		Spec:     map[string]interface{}{},
	}

	devEnvNode := &graph.Node{
		ID:       "dev-env",
		Kind:     graph.KindEnvironment,
		Metadata: map[string]interface{}{"name": "Development"},
		Spec:     map[string]interface{}{},
	}

	prodEnvNode := &graph.Node{
		ID:       "prod-env",
		Kind:     graph.KindEnvironment,
		Metadata: map[string]interface{}{"name": "Production"},
		Spec:     map[string]interface{}{},
	}

	// Add nodes to graph
	err = g.AddNode(serviceNode)
	if err != nil {
		t.Fatalf("Failed to add service node: %v", err)
	}

	err = g.AddNode(devEnvNode)
	if err != nil {
		t.Fatalf("Failed to add dev env node: %v", err)
	}

	err = g.AddNode(prodEnvNode)
	if err != nil {
		t.Fatalf("Failed to add prod env node: %v", err)
	}

	// Create a policy node
	policyNode := &graph.Node{
		ID:   "policy-dev-before-prod",
		Kind: graph.KindPolicy,
		Metadata: map[string]interface{}{
			"name":        "Must Deploy To Dev Before Prod",
			"description": "Requires a service version to be deployed to dev before it can be deployed to prod",
			"type":        graph.PolicyTypeSystem, // Using PolicyTypeSystem since PolicyTypeReachability is not available
		},
		Spec: map[string]interface{}{
			"sourceKind":      graph.KindServiceVersion,
			"targetKind":      graph.KindEnvironment,
			"targetID":        "prod-env",
			"requiredPathIDs": []string{"dev-env"},
		},
	}

	err = g.AddNode(policyNode)
	if err != nil {
		t.Fatalf("Failed to add policy node: %v", err)
	}

	// Attach policy to the transition from service to prod environment
	err = g.AttachPolicyToTransition(serviceNode.ID, prodEnvNode.ID, graph.EdgeTypeDeploy, policyNode.ID)
	if err != nil {
		t.Fatalf("Failed to attach policy to transition: %v", err)
	}

	// Create policy evaluator
	evaluator := policies.NewPolicyEvaluator(graphStore, env)

	// Test 1: Try deploying directly to prod (should fail)
	err = evaluator.ValidateTransition(serviceNode.ID, prodEnvNode.ID, graph.EdgeTypeDeploy, "test-user")
	if err == nil {
		t.Errorf("Expected error when deploying directly to production, got nil")
	} else {
		t.Logf("Got expected error: %v", err)
	}

	// Test 2: Deploy to dev first, then to prod (should succeed)
	err = g.AddEdge(serviceNode.ID, devEnvNode.ID, graph.EdgeTypeDeploy)
	if err != nil {
		t.Fatalf("Failed to deploy to dev: %v", err)
	}

	// Just deploying to dev isn't enough, we need a check to satisfy the policy
	var testErr error
	testErr = evaluator.ValidateTransition(serviceNode.ID, prodEnvNode.ID, graph.EdgeTypeDeploy, "test-user")
	if testErr == nil {
		t.Errorf("Expected error when deploying to prod (policy not yet satisfied)")
	} else {
		t.Logf("Got expected error: %v", testErr)
	}
	
	// Test 3: Create a check node that satisfies the policy
	checkNode := &graph.Node{
		ID:   "check-dev-deployment",
		Kind: graph.KindCheck,
		Metadata: map[string]interface{}{
			"name":   "Dev Deployment Check",
			"type":   "deployment-verification",
			"status": graph.CheckStatusSucceeded, // Mark as succeeded
		},
		Spec: map[string]interface{}{},
	}

	err = g.AddNode(checkNode)
	if err != nil {
		t.Fatalf("Failed to add check node: %v", err)
	}

	// Create a satisfies relationship from the check to policy
	err = g.AddEdge(checkNode.ID, policyNode.ID, graph.EdgeTypeSatisfies)
	if err != nil {
		t.Fatalf("Failed to create satisfies relationship: %v", err)
	}

	// Now try to deploy to prod again (should succeed)
	testErr = evaluator.ValidateTransition(serviceNode.ID, prodEnvNode.ID, graph.EdgeTypeDeploy, "test-user")
	if testErr != nil {
		t.Errorf("Expected to be able to deploy to prod with satisfied policy, got error: %v", testErr)
	} else {
		t.Logf("Successfully validated deployment to prod with satisfied policy")
	}
}
