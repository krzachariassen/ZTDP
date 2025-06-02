package ai

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAIProvider is a mock implementation of AIProvider for testing
type MockAIProvider struct {
	mock.Mock
}

func (m *MockAIProvider) GeneratePlan(ctx context.Context, request *PlanningRequest) (*PlanningResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*PlanningResponse), args.Error(1)
}

func (m *MockAIProvider) EvaluatePolicy(ctx context.Context, policyContext interface{}) (*PolicyEvaluation, error) {
	args := m.Called(ctx, policyContext)
	return args.Get(0).(*PolicyEvaluation), args.Error(1)
}

func (m *MockAIProvider) OptimizePlan(ctx context.Context, plan *DeploymentPlan, context *PlanningContext) (*PlanningResponse, error) {
	args := m.Called(ctx, plan, context)
	return args.Get(0).(*PlanningResponse), args.Error(1)
}

func (m *MockAIProvider) GetProviderInfo() *ProviderInfo {
	args := m.Called()
	return args.Get(0).(*ProviderInfo)
}

func (m *MockAIProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestAIBrain tests the core AI brain functionality
func TestAIBrain(t *testing.T) {
	// Create test graph
	globalGraph := createTestGraph()

	// Create mock provider
	mockProvider := &MockAIProvider{}

	// Create AI brain
	brain := NewAIBrain(mockProvider, globalGraph)

	t.Run("GenerateDeploymentPlan", func(t *testing.T) {
		// Setup mock response
		expectedResponse := &PlanningResponse{
			Plan: &DeploymentPlan{
				Steps: []*DeploymentStep{
					{
						ID:        "step-1",
						Action:    "deploy",
						Target:    "test-service",
						Reasoning: "Deploy service first",
					},
					{
						ID:           "step-2",
						Action:       "validate",
						Target:       "test-app",
						Dependencies: []string{"step-1"},
						Reasoning:    "Validate deployment",
					},
				},
				Strategy: "rolling",
			},
			Reasoning:  "Deploy services in dependency order",
			Confidence: 0.95,
		}

		// Setup mock expectations
		mockProvider.On("GeneratePlan", mock.Anything, mock.Anything).Return(expectedResponse, nil)
		mockProvider.On("GetProviderInfo").Return(&ProviderInfo{
			Name:         "mock-provider",
			Version:      "1.0.0",
			Capabilities: []string{"planning", "evaluation"},
		})

		// Test plan generation
		response, err := brain.GenerateDeploymentPlan(context.Background(), "test-app", []string{"deploy", "owns"})

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 2, len(response.Plan.Steps))
		assert.Equal(t, "rolling", response.Plan.Strategy)
		assert.Equal(t, 0.95, response.Confidence)

		mockProvider.AssertExpectations(t)
	})

	t.Run("EvaluateDeploymentPolicies", func(t *testing.T) {
		// Setup mock response
		expectedEvaluation := &PolicyEvaluation{
			Compliant:   true,
			Violations:  []string{},
			Suggestions: []string{"Consider adding health checks"},
			Reasoning:   "All policies satisfied",
			Confidence:  0.9,
		}

		mockProvider.On("EvaluatePolicy", mock.Anything, mock.Anything).Return(expectedEvaluation, nil)

		// Test policy evaluation
		evaluation, err := brain.EvaluateDeploymentPolicies(context.Background(), "test-app", "production")

		assert.NoError(t, err)
		assert.NotNil(t, evaluation)
		assert.True(t, evaluation.Compliant)
		assert.Equal(t, 0, len(evaluation.Violations))
		assert.Equal(t, 0.9, evaluation.Confidence)

		mockProvider.AssertExpectations(t)
	})
}

// TestAIPlanner tests the AI planner adapter
func TestAIPlanner(t *testing.T) {
	// Create test components
	globalGraph := createTestGraph()
	mockProvider := &MockAIProvider{}
	brain := NewAIBrain(mockProvider, globalGraph)

	// Create test subgraph
	subgraph := createTestSubgraph()

	// Create AI planner
	planner := NewAIPlanner(brain, subgraph, "test-app")

	t.Run("PlanWithEdgeTypes", func(t *testing.T) {
		// Setup mock response
		expectedResponse := &PlanningResponse{
			Plan: &DeploymentPlan{
				Steps: []*DeploymentStep{
					{
						ID:           "step-1",
						Action:       "deploy",
						Target:       "service-1",
						Dependencies: []string{},
						Reasoning:    "Deploy first service",
					},
					{
						ID:           "step-2",
						Action:       "deploy",
						Target:       "service-2",
						Dependencies: []string{"step-1"},
						Reasoning:    "Deploy second service after first",
					},
				},
				Strategy: "rolling",
			},
			Reasoning:  "Sequential deployment for safety",
			Confidence: 0.9,
		}

		// Setup mock expectations
		mockProvider.On("GeneratePlan", mock.Anything, mock.Anything).Return(expectedResponse, nil)
		mockProvider.On("GetProviderInfo").Return(&ProviderInfo{
			Name:         "mock-provider",
			Version:      "1.0.0",
			Capabilities: []string{"planning", "evaluation"},
		})

		// Test planning
		order, err := planner.PlanWithEdgeTypes([]string{"deploy", "owns"})

		assert.NoError(t, err)
		assert.Equal(t, 2, len(order))
		assert.Equal(t, "service-1", order[0])
		assert.Equal(t, "service-2", order[1])

		mockProvider.AssertExpectations(t)
	})

	t.Run("FallbackPlan", func(t *testing.T) {
		// Test fallback when AI fails
		mockProvider.On("GeneratePlan", mock.Anything, mock.Anything).Return(nil, assert.AnError)

		order, err := planner.PlanWithEdgeTypes([]string{"deploy"})

		// Should not error - should use fallback
		assert.NoError(t, err)
		assert.Greater(t, len(order), 0)

		mockProvider.AssertExpectations(t)
	})
}

// TestExtractApplicationSubgraph tests the subgraph extraction
func TestExtractApplicationSubgraph(t *testing.T) {
	globalGraph := createTestGraph()

	subgraph, err := ExtractApplicationSubgraph(globalGraph, "test-app")

	assert.NoError(t, err)
	assert.NotNil(t, subgraph)
	assert.Greater(t, len(subgraph.Nodes), 0)
}

// createTestGraph creates a test graph for testing
func createTestGraph() *graph.GlobalGraph {
	backend := graph.NewMemoryGraph()
	globalGraph := graph.NewGlobalGraph(backend)

	// Add test application
	appNode := &graph.Node{
		ID:   "test-app",
		Kind: "application",
		Metadata: map[string]interface{}{
			"name":  "test-app",
			"owner": "test-team",
		},
		Spec: map[string]interface{}{
			"description": "Test application",
		},
	}
	globalGraph.AddNode(appNode)

	// Add test services
	service1 := &graph.Node{
		ID:   "service-1",
		Kind: "service",
		Metadata: map[string]interface{}{
			"name":  "service-1",
			"owner": "test-team",
		},
		Spec: map[string]interface{}{
			"application": "test-app",
			"port":        8080,
		},
	}
	globalGraph.AddNode(service1)

	service2 := &graph.Node{
		ID:   "service-2",
		Kind: "service",
		Metadata: map[string]interface{}{
			"name":  "service-2",
			"owner": "test-team",
		},
		Spec: map[string]interface{}{
			"application": "test-app",
			"port":        8081,
		},
	}
	globalGraph.AddNode(service2)

	// Add edges
	globalGraph.AddEdge("test-app", "service-1", "owns")
	globalGraph.AddEdge("test-app", "service-2", "owns")
	globalGraph.AddEdge("service-2", "service-1", "depends")

	return globalGraph
}

// createTestSubgraph creates a test subgraph
func createTestSubgraph() *graph.Graph {
	subgraph := graph.NewGraph()

	// Add test nodes
	appNode := &graph.Node{
		ID:   "test-app",
		Kind: "application",
		Metadata: map[string]interface{}{
			"name": "test-app",
		},
	}
	subgraph.AddNode(appNode)

	service1 := &graph.Node{
		ID:   "service-1",
		Kind: "service",
		Metadata: map[string]interface{}{
			"name": "service-1",
		},
	}
	subgraph.AddNode(service1)

	service2 := &graph.Node{
		ID:   "service-2",
		Kind: "service",
		Metadata: map[string]interface{}{
			"name": "service-2",
		},
	}
	subgraph.AddNode(service2)

	// Add edges
	subgraph.AddEdge("test-app", "service-1", "owns")
	subgraph.AddEdge("test-app", "service-2", "owns")

	return subgraph
}
