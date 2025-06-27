package testHelpers

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/ztdp/orchestrator/internal/messaging"
)

// MockMessageBus provides a testify-based mock for message bus operations
type MockMessageBus struct {
	mock.Mock
}

// NewMockMessageBus creates a new mock message bus instance
func NewMockMessageBus() *MockMessageBus {
	return &MockMessageBus{}
}

func (m *MockMessageBus) Publish(ctx context.Context, topic string, message interface{}) error {
	args := m.Called(ctx, topic, message)
	return args.Error(0)
}

func (m *MockMessageBus) Subscribe(ctx context.Context, topic string, handler func(interface{}) error) error {
	args := m.Called(ctx, topic, handler)
	return args.Error(0)
}

func (m *MockMessageBus) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockAIMessageBus provides a testify-based mock for AI message bus operations
type MockAIMessageBus struct {
	mock.Mock
}

// NewMockAIMessageBus creates a new mock AI message bus instance
func NewMockAIMessageBus() *MockAIMessageBus {
	return &MockAIMessageBus{}
}

func (m *MockAIMessageBus) SendToAgent(ctx context.Context, msg *messaging.AIToAgentMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockAIMessageBus) SendToAI(ctx context.Context, msg *messaging.AgentToAIMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockAIMessageBus) SendBetweenAgents(ctx context.Context, msg *messaging.AgentToAgentMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockAIMessageBus) SendUserToAI(ctx context.Context, msg *messaging.UserToAIMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockAIMessageBus) Subscribe(ctx context.Context, participantID string) (<-chan *messaging.Message, error) {
	args := m.Called(ctx, participantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan *messaging.Message), args.Error(1)
}

func (m *MockAIMessageBus) GetConversationHistory(ctx context.Context, correlationID string) ([]*messaging.Message, error) {
	args := m.Called(ctx, correlationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*messaging.Message), args.Error(1)
}
