package policies_test

import (
	"errors"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

// Mock implementation for GraphBackend
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

// Add a node - if it exists, update it
func (m *MockGraphBackend) AddNode(env string, node *graph.Node) error {
	g, ok := m.graphs[env]
	if !ok {
		m.graphs[env] = graph.NewGraph()
		g = m.graphs[env]
	}

	// First check if node exists, and if so, update it instead of adding a new one
	_, err := g.GetNode(node.ID)
	if err == nil {
		// Node exists, replace it
		return g.UpdateNode(node)
	}

	// Node doesn't exist, add it
	return g.AddNode(node)
}

func (m *MockGraphBackend) AddEdge(env, fromID, toID, relType string) error {
	g, ok := m.graphs[env]
	if !ok {
		return errors.New("environment not found")
	}
	return g.AddEdge(fromID, toID, relType)
}

func (m *MockGraphBackend) GetNode(env, id string) (*graph.Node, error) {
	g, ok := m.graphs[env]
	if !ok {
		return nil, errors.New("environment not found")
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

// Implement SaveGlobal and LoadGlobal to satisfy the GraphBackend interface
func (m *MockGraphBackend) SaveGlobal(g *graph.Graph) error {
	m.global = g
	return nil
}

func (m *MockGraphBackend) LoadGlobal() (*graph.Graph, error) {
	return m.global, nil
}

func TestPolicyEvaluator(t *testing.T) {
	// Initialize event system for testing
	eventTransport := events.NewMemoryTransport()
	events.InitializeEventBus(eventTransport)

	backend := NewMockGraphBackend()
	graphStore := graph.NewGraphStore(backend)

	t.Run("CreatePolicyNode", func(t *testing.T) {
		env := "default"

		evaluator := policies.NewPolicyEvaluator(graphStore, env)

		// Create a policy node
		policyNode, err := evaluator.CreatePolicyNode(
			"test-policy",
			"Test Policy Description",
			graph.PolicyTypeCheck,
			map[string]interface{}{
				"requiresApproval": true,
			},
		)

		if err != nil {
			t.Fatalf("Failed to create policy node: %v", err)
		}

		// Verify policy node was created correctly
		if policyNode.Kind != graph.KindPolicy {
			t.Errorf("Expected node kind %s, got %s", graph.KindPolicy, policyNode.Kind)
		}

		if name, ok := policyNode.Metadata["name"]; !ok || name != "test-policy" {
			t.Errorf("Expected policy name 'test-policy', got %v", name)
		}

		// Verify policy exists in graph
		retrievedNode, err := backend.GetNode(env, policyNode.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve policy node: %v", err)
		}

		if retrievedNode.ID != policyNode.ID {
			t.Errorf("Retrieved policy ID %s does not match created ID %s",
				retrievedNode.ID, policyNode.ID)
		}
	})

	t.Run("PolicySatisfactionFlow", func(t *testing.T) {
		env := "test-env"

		// Create a fresh graph for this test
		backend.graphs[env] = graph.NewGraph()

		evaluator := policies.NewPolicyEvaluator(graphStore, env)

		// Create application and environment nodes
		appNode := &graph.Node{
			ID:       "test-app",
			Kind:     graph.KindApplication,
			Metadata: map[string]interface{}{"name": "Test App"},
			Spec:     map[string]interface{}{},
		}

		envNode := &graph.Node{
			ID:       "prod-env",
			Kind:     graph.KindEnvironment,
			Metadata: map[string]interface{}{"name": "Production"},
			Spec:     map[string]interface{}{},
		}

		backend.AddNode(env, appNode)
		backend.AddNode(env, envNode)

		// Create a policy node
		policyNode, err := evaluator.CreatePolicyNode(
			"security-scan",
			"Security scan must pass",
			graph.PolicyTypeCheck,
			map[string]interface{}{
				"scanType": "security",
				"severity": "high",
			},
		)
		if err != nil {
			t.Fatalf("Failed to create policy node: %v", err)
		}

		// Get the graph
		g, err := backend.GetAll(env)
		if err != nil {
			t.Fatalf("Failed to get graph: %v", err)
		}

		// Attach policy to transition
		err = g.AttachPolicyToTransition(appNode.ID, envNode.ID, graph.EdgeTypeDeploy, policyNode.ID)
		if err != nil {
			t.Fatalf("Failed to attach policy to transition: %v", err)
		}

		// Since we're working directly with the graph, we need to make sure the changes
		// are reflected in the backend instance
		backend.graphs[env] = g

		// Create a check node
		checkNode, err := evaluator.CreateCheckNode(
			"security-scan-check-1",
			"Security Scan Execution",
			"security-scan",
			map[string]interface{}{
				"scanId": "scan-123",
				"tool":   "snyk",
			},
		)
		if err != nil {
			t.Fatalf("Failed to create check node: %v", err)
		}

		// Link check to policy
		err = evaluator.SatisfyPolicy(checkNode.ID, policyNode.ID)
		if err != nil {
			t.Fatalf("Failed to link check to policy: %v", err)
		}

		// Transition should not be allowed yet (check is pending)
		err = evaluator.ValidateTransition(appNode.ID, envNode.ID, graph.EdgeTypeDeploy, "test-user")
		if err == nil {
			t.Errorf("Expected transition to be blocked due to pending check")
		}

		// Update check status to succeeded
		err = evaluator.UpdateCheckStatus(checkNode.ID, graph.CheckStatusSucceeded, map[string]interface{}{
			"vulnerabilities": 0,
			"scanDuration":    "2m30s",
		})
		if err != nil {
			t.Fatalf("Failed to update check status: %v", err)
		}

		// Now transition should be allowed
		err = evaluator.ValidateTransition(appNode.ID, envNode.ID, graph.EdgeTypeDeploy, "test-user")
		if err != nil {
			t.Errorf("Expected transition to be allowed after check success, got error: %v", err)
		}
	})
}
