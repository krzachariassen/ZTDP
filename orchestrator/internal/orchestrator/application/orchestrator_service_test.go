package application

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	aiInfrastructure "github.com/ztdp/orchestrator/internal/ai/infrastructure"
	"github.com/ztdp/orchestrator/internal/logging"
	orchestratorDomain "github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// Mock implementations for testing (but we'll use real AI provider)
type MockGraphExplorer struct {
	mock.Mock
}

func (m *MockGraphExplorer) GetAgentContext(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

type MockExecutionCoordinator struct {
	mock.Mock
}

func (m *MockExecutionCoordinator) CreatePlan(ctx context.Context, decision *orchestratorDomain.Decision) (string, error) {
	args := m.Called(ctx, decision)
	return args.String(0), args.Error(1)
}

func (m *MockExecutionCoordinator) GetPlanStatus(ctx context.Context, planID string) (*orchestratorDomain.ExecutionPlan, error) {
	args := m.Called(ctx, planID)
	return args.Get(0).(*orchestratorDomain.ExecutionPlan), args.Error(1)
}

func (m *MockExecutionCoordinator) UpdateStatus(ctx context.Context, planID string, status orchestratorDomain.ExecutionStatus) error {
	args := m.Called(ctx, planID, status)
	return args.Error(0)
}

func (m *MockExecutionCoordinator) ExecutePlan(ctx context.Context, planID string) error {
	args := m.Called(ctx, planID)
	return args.Error(0)
}

type MockLearningService struct {
	mock.Mock
}

func (m *MockLearningService) StoreInsights(ctx context.Context, userRequest string, analysis *orchestratorDomain.Analysis, decision *orchestratorDomain.Decision) error {
	args := m.Called(ctx, userRequest, analysis, decision)
	return args.Error(0)
}

func (m *MockLearningService) AnalyzePatterns(ctx context.Context, sessionID string) (*orchestratorDomain.ConversationPattern, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*orchestratorDomain.ConversationPattern), args.Error(1)
}

// setupRealAIProvider creates a real OpenAI provider for testing
func setupRealAIProviderForOrchestrator(t *testing.T) *aiInfrastructure.OpenAIProvider {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY environment variable not set, skipping AI provider tests")
	}

	config := aiInfrastructure.DefaultOpenAIConfig()
	config.APIKey = apiKey
	config.Model = "gpt-3.5-turbo" // Use faster model for tests
	config.MaxTokens = 1000        // Limit tokens for faster tests

	logger, _ := logging.NewLogger(false) // Production logger for tests
	provider := aiInfrastructure.NewOpenAIProvider(config, logger)

	return provider
}

func TestOrchestratorService_ProcessUserRequest(t *testing.T) {
	t.Run("should process clarification request successfully", func(t *testing.T) {
		// Setup with real AI provider
		aiProvider := setupRealAIProviderForOrchestrator(t)
		aiEngine := NewAIDecisionEngine(aiProvider)

		// Setup mocks for other services
		mockExplorer := &MockGraphExplorer{}
		mockCoordinator := &MockExecutionCoordinator{}
		mockLearning := &MockLearningService{}

		service := NewOrchestratorService(aiEngine, mockExplorer, mockCoordinator, mockLearning)

		// Test data
		request := &OrchestratorRequest{
			UserInput: "Deploy something unclear",
			UserID:    "user-123",
		}

		agentContext := "Deploy Agent available"

		// Setup expectations
		mockExplorer.On("GetAgentContext", mock.Anything).Return(agentContext, nil)
		mockLearning.On("StoreInsights", mock.Anything, request.UserInput, mock.Anything, mock.Anything).Return(nil)

		// Execute
		result, err := service.ProcessUserRequest(context.Background(), request)

		// Verify
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotNil(t, result.Analysis)
		assert.NotNil(t, result.Decision)

		// The real AI should handle this request appropriately
		t.Logf("AI Response: %s", result.Message)
		t.Logf("Decision Type: %s", string(result.Decision.Type))

		// Verify mocks
		mockExplorer.AssertExpectations(t)
		mockLearning.AssertExpectations(t)
	})

	t.Run("should process execution request with action successfully", func(t *testing.T) {
		// Setup with real AI provider
		aiProvider := setupRealAIProviderForOrchestrator(t)
		aiEngine := NewAIDecisionEngine(aiProvider)

		// Setup mocks for other services
		mockExplorer := &MockGraphExplorer{}
		mockCoordinator := &MockExecutionCoordinator{}
		mockLearning := &MockLearningService{}

		service := NewOrchestratorService(aiEngine, mockExplorer, mockCoordinator, mockLearning)

		// Test data
		request := &OrchestratorRequest{
			UserInput: "Deploy my application to production environment",
			UserID:    "user-123",
		}

		agentContext := "Deploy Agent available with deploy capability"
		planID := "plan-123"

		// Setup expectations
		mockExplorer.On("GetAgentContext", mock.Anything).Return(agentContext, nil)
		mockCoordinator.On("CreatePlan", mock.Anything, mock.Anything).Return(planID, nil)
		mockLearning.On("StoreInsights", mock.Anything, request.UserInput, mock.Anything, mock.Anything).Return(nil)

		// Execute
		result, err := service.ProcessUserRequest(context.Background(), request)

		// Verify
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotNil(t, result.Analysis)
		assert.NotNil(t, result.Decision)

		// Log the AI's decision for inspection
		t.Logf("AI Response: %s", result.Message)
		t.Logf("Decision Type: %s", string(result.Decision.Type))
		t.Logf("Analysis Intent: %s", result.Analysis.Intent)
		t.Logf("Analysis Confidence: %d", result.Analysis.Confidence)

		// Verify mocks
		mockExplorer.AssertExpectations(t)
		mockLearning.AssertExpectations(t)

		// If AI made an execute decision with action, coordinator should be called
		if result.Decision.Type == orchestratorDomain.DecisionTypeExecute && result.Decision.HasAction() {
			mockCoordinator.AssertExpectations(t)
		}
	})

	t.Run("should handle agent context error", func(t *testing.T) {
		// Setup with real AI provider
		aiProvider := setupRealAIProviderForOrchestrator(t)
		aiEngine := NewAIDecisionEngine(aiProvider)

		// Setup mocks for other services
		mockExplorer := &MockGraphExplorer{}
		mockCoordinator := &MockExecutionCoordinator{}
		mockLearning := &MockLearningService{}

		service := NewOrchestratorService(aiEngine, mockExplorer, mockCoordinator, mockLearning)

		request := &OrchestratorRequest{
			UserInput: "Deploy app",
			UserID:    "user-123",
		}

		// Setup expectations
		mockExplorer.On("GetAgentContext", mock.Anything).Return("", assert.AnError)

		// Execute
		result, err := service.ProcessUserRequest(context.Background(), request)

		// Verify
		assert.NoError(t, err) // Service should not return Go error
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Failed to get agent context")

		// Verify mocks
		mockExplorer.AssertExpectations(t)
	})
}
