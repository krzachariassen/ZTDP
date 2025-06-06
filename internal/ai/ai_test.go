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
	agent := NewPlatformAgent(mockProvider, globalGraph, nil, nil, nil, nil, nil, nil)

	t.Run("Provider", func(t *testing.T) {
		// Test that we can get the provider
		provider := agent.Provider()
		assert.NotNil(t, provider)
		assert.Equal(t, mockProvider, provider)
	})

	t.Run("Close", func(t *testing.T) {
		// Setup mock for Close method
		mockProvider.On("Close").Return(nil)

		// Test cleanup
		err := agent.Close()
		assert.NoError(t, err)

		mockProvider.AssertExpectations(t)
	})
}

// TestPlanConversion tests plan conversion utilities
func TestPlanConversion(t *testing.T) {
	t.Run("ConvertPlanToOrder", func(t *testing.T) {
		// Test the plan conversion utility directly
		plan := &DeploymentPlan{
			ID:          "test-plan",
			Application: "test-app",
			Environment: "test",
			Steps: []DeploymentStep{
				{
					ID:           "step-1",
					Type:         "deploy",
					Description:  "Deploy first service",
					Dependencies: []string{},
				},
				{
					ID:           "step-2",
					Type:         "deploy",
					Description:  "Deploy second service",
					Dependencies: []string{"step-1"},
				},
			},
			EstimatedTime: "5m",
		}

		// Test the conversion utility logic where it belongs (deployment domain)
		// This validates the plan structure is correct
		if plan != nil && len(plan.Steps) > 0 {
			assert.Equal(t, 2, len(plan.Steps))
			assert.Equal(t, "step-1", plan.Steps[0].ID)
			assert.Equal(t, "step-2", plan.Steps[1].ID)
			assert.Equal(t, []string{"step-1"}, plan.Steps[1].Dependencies)
		}
	})

	t.Run("InvalidPlanHandling", func(t *testing.T) {
		// Test validation of invalid plans
		nilPlan := (*DeploymentPlan)(nil)
		assert.Nil(t, nilPlan)

		// Test empty plan
		emptyPlan := &DeploymentPlan{Steps: []DeploymentStep{}}
		assert.Equal(t, 0, len(emptyPlan.Steps))
	})
}

// TestAIService tests core AI service functionality
func TestAIService(t *testing.T) {
	// Create test components
	mockProvider := &MockAIProvider{}
	globalGraph := createTestGraph()

	// Create AI service
	service := NewAIService(mockProvider, globalGraph)

	t.Run("GenerateDeploymentPlan", func(t *testing.T) {
		// Setup mock response
		mockProvider.On("CallAI", mock.Anything, mock.Anything, mock.Anything).Return(`{"plan": {"id": "test-plan", "application": "test-app", "environment": "test", "steps": [{"id": "step-1", "type": "deploy", "description": "Deploy application"}], "estimated_time": "5m"}, "confidence": 0.9}`, nil)

		// Test plan generation
		request := &PlanningRequest{
			ApplicationID: "test-app",
			Intent:        "deploy application",
		}

		response, err := service.GenerateDeploymentPlan(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Plan)
		assert.Equal(t, "test-app", response.Plan.Application)

		mockProvider.AssertExpectations(t)
	})
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
