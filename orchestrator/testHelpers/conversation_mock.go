package testHelpers

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/ztdp/orchestrator/internal/conversation/domain"
	orchestratorDomain "github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// MockConversationRepository provides a testify-based mock for conversation repository operations
type MockConversationRepository struct {
	mock.Mock
}

// NewMockConversationRepository creates a new mock conversation repository instance
func NewMockConversationRepository() *MockConversationRepository {
	return &MockConversationRepository{}
}

func (m *MockConversationRepository) SaveConversation(ctx context.Context, conversation *domain.Conversation) error {
	args := m.Called(ctx, conversation)
	return args.Error(0)
}

func (m *MockConversationRepository) GetConversation(ctx context.Context, id string) (*domain.Conversation, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationRepository) GetConversationsByUser(ctx context.Context, userID string) ([]*domain.Conversation, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Conversation), args.Error(1)
}

// MockConversationService provides a testify-based mock for conversation service operations
type MockConversationService struct {
	mock.Mock
}

// NewMockConversationService creates a new mock conversation service instance
func NewMockConversationService() *MockConversationService {
	return &MockConversationService{}
}

func (m *MockConversationService) ProcessRequest(ctx context.Context, userID, request string) (*domain.Conversation, error) {
	args := m.Called(ctx, userID, request)
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationService) GetConversation(ctx context.Context, id string) (*domain.Conversation, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationService) SavePattern(ctx context.Context, pattern *orchestratorDomain.ConversationPattern) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}
