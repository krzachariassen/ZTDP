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

func (m *MockAIProvider) CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	args := m.Called(ctx, systemPrompt, userPrompt)
	return args.String(0), args.Error(1)
}

func (m *MockAIProvider) GetProviderInfo() *ProviderInfo {
	args := m.Called()
	return args.Get(0).(*ProviderInfo)
}

func (m *MockAIProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAIProvider) ChatWithPlatform(ctx context.Context, query string, context string) (*ConversationalResponse, error) {
	args := m.Called(ctx, query, context)
	return args.Get(0).(*ConversationalResponse), args.Error(1)
}

func (m *MockAIProvider) PredictImpact(ctx context.Context, request *ImpactAnalysisRequest) (*ImpactPrediction, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*ImpactPrediction), args.Error(1)
}

func (m *MockAIProvider) IntelligentTroubleshooting(ctx context.Context, incident *IncidentContext) (*TroubleshootingResponse, error) {
	args := m.Called(ctx, incident)
	return args.Get(0).(*TroubleshootingResponse), args.Error(1)
}

func (m *MockAIProvider) ProactiveOptimization(ctx context.Context, scope *OptimizationScope) (*OptimizationRecommendations, error) {
	args := m.Called(ctx, scope)
	return args.Get(0).(*OptimizationRecommendations), args.Error(1)
}

func (m *MockAIProvider) LearningFromFailures(ctx context.Context, outcome *DeploymentOutcome) (*LearningInsights, error) {
	args := m.Called(ctx, outcome)
	return args.Get(0).(*LearningInsights), args.Error(1)
}

// TestPlatformAgent tests the core platform agent functionality
func TestPlatformAgent(t *testing.T) {
	// Create test graph
	globalGraph := createTestGraph()

	// Create mock provider
	mockProvider := &MockAIProvider{}

	// Create platform agent
	agent := NewPlatformAgent(mockProvider, globalGraph, nil, nil)

	t.Run("ChatWithPlatform", func(t *testing.T) {
		// Setup mock response for chat
		expectedResponse := &ConversationalResponse{
			Message: "I can help you deploy your application",
			Intent:  "deployment_assistance",
		}

		mockProvider.On("ChatWithPlatform", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

		// Test chat functionality
		response, err := agent.ChatWithPlatform(context.Background(), "Help me deploy my app", "")

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Contains(t, response.Message, "deploy")

		mockProvider.AssertExpectations(t)
	})

	t.Run("Provider", func(t *testing.T) {
		// Test that we can get the provider
		provider := agent.Provider()
		assert.NotNil(t, provider)
		assert.Equal(t, mockProvider, provider)
	})

	t.Run("Close", func(t *testing.T) {
		// Test cleanup
		err := agent.Close()
		assert.NoError(t, err)
	})
}

// TestAIPlanner tests the deprecated AI planner adapter compatibility
func TestAIPlanner(t *testing.T) {
	// Create test components
	mockProvider := &MockAIProvider{}

	// Create test subgraph
	subgraph := createTestSubgraph()

	// Create AI planner (deprecated but still used in tests for compatibility)
	planner := NewAIPlanner(mockProvider, subgraph, "test-app")

	t.Run("PlannerCompatibility", func(t *testing.T) {
		// Test that the deprecated planner still provides basic functionality
		assert.NotNil(t, planner)
		assert.Equal(t, "test-app", planner.GetApplicationID())
		assert.NotNil(t, planner.GetSubgraph())
		assert.Equal(t, 3, len(planner.GetSubgraph().Nodes)) // app + 2 services
	})

	t.Run("ConvertPlanToOrder", func(t *testing.T) {
		// Test the plan conversion utility
		plan := &DeploymentPlan{
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
		}

		// Test the conversion helper function
		order, err := planner.convertPlanToOrder(plan)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(order))
		assert.Equal(t, "service-1", order[0])
		assert.Equal(t, "service-2", order[1])
	})

	t.Run("InvalidPlanHandling", func(t *testing.T) {
		// Test error handling for invalid plans
		_, err := planner.convertPlanToOrder(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid deployment plan")

		// Test empty plan
		emptyPlan := &DeploymentPlan{Steps: []*DeploymentStep{}}
		_, err = planner.convertPlanToOrder(emptyPlan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid deployment plan")
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
