package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	orchestratorDomain "github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// MockConversationService for testing
type MockConversationService struct {
	mock.Mock
}

func (m *MockConversationService) StoreInteraction(ctx context.Context, userRequest string, analysis *orchestratorDomain.Analysis, decision *orchestratorDomain.Decision) error {
	args := m.Called(ctx, userRequest, analysis, decision)
	return args.Error(0)
}

func (m *MockConversationService) GetConversationHistory(ctx context.Context, sessionID string) ([]string, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockConversationService) CreateSession(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockConversationService) AnalyzePatterns(ctx context.Context, sessionID string) (*orchestratorDomain.ConversationPattern, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*orchestratorDomain.ConversationPattern), args.Error(1)
}

func TestLearningService_StoreInsights(t *testing.T) {
	t.Run("should store interaction insights", func(t *testing.T) {
		mockConversationService := &MockConversationService{}
		learningService := NewLearningService(mockConversationService)

		userRequest := "Deploy the application to staging"
		analysis := &orchestratorDomain.Analysis{
			Intent:     "deployment",
			Category:   "infrastructure",
			Confidence: 95,
		}
		decision := orchestratorDomain.NewExecuteDecisionWithAction("deploy",
			map[string]interface{}{"env": "staging"}, "Clear deployment request")

		mockConversationService.On("StoreInteraction", mock.Anything, userRequest, analysis, decision).Return(nil)

		err := learningService.StoreInsights(context.Background(), userRequest, analysis, decision)

		assert.NoError(t, err)
		mockConversationService.AssertExpectations(t)
	})

	t.Run("should handle storage failure", func(t *testing.T) {
		mockConversationService := &MockConversationService{}
		learningService := NewLearningService(mockConversationService)

		userRequest := "Deploy the application"
		analysis := &orchestratorDomain.Analysis{
			Intent:     "deployment",
			Confidence: 80,
		}
		decision := orchestratorDomain.NewClarifyDecision("Which environment?", "Need environment")

		mockConversationService.On("StoreInteraction", mock.Anything, userRequest, analysis, decision).Return(assert.AnError)

		err := learningService.StoreInsights(context.Background(), userRequest, analysis, decision)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store interaction insights")
		mockConversationService.AssertExpectations(t)
	})
}

func TestLearningService_AnalyzePattern(t *testing.T) {
	t.Run("should analyze patterns from conversation history", func(t *testing.T) {
		mockConversationService := &MockConversationService{}
		learningService := NewLearningService(mockConversationService)

		sessionID := "session-123"
		history := []string{
			"Deploy app to staging",
			"Deploy app to production",
			"Check deployment status",
		}

		mockConversationService.On("GetConversationHistory", mock.Anything, sessionID).Return(history, nil)

		patterns, err := learningService.AnalyzePatterns(context.Background(), sessionID)

		assert.NoError(t, err)
		assert.NotNil(t, patterns)
		assert.Contains(t, patterns.CommonIntents, "deployment")
		assert.GreaterOrEqual(t, patterns.TotalInteractions, 3)
		mockConversationService.AssertExpectations(t)
	})

	t.Run("should handle empty conversation history", func(t *testing.T) {
		mockConversationService := &MockConversationService{}
		learningService := NewLearningService(mockConversationService)

		sessionID := "empty-session"
		history := []string{}

		mockConversationService.On("GetConversationHistory", mock.Anything, sessionID).Return(history, nil)

		patterns, err := learningService.AnalyzePatterns(context.Background(), sessionID)

		assert.NoError(t, err)
		assert.NotNil(t, patterns)
		assert.Empty(t, patterns.CommonIntents)
		assert.Equal(t, 0, patterns.TotalInteractions)
		mockConversationService.AssertExpectations(t)
	})
}
