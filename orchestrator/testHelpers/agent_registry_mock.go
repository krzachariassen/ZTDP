package testHelpers

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/ztdp/orchestrator/internal/agent/domain"
)

// MockRegistry provides a testify-based mock for registry operations
type MockRegistry struct {
	mock.Mock
}

// NewMockRegistry creates a new mock registry instance
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{}
}

func (m *MockRegistry) RegisterAgent(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockRegistry) UnregisterAgent(ctx context.Context, agentID string) error {
	args := m.Called(ctx, agentID)
	return args.Error(0)
}

func (m *MockRegistry) GetAgent(ctx context.Context, agentID string) (*domain.Agent, error) {
	args := m.Called(ctx, agentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockRegistry) GetAllAgents(ctx context.Context) ([]*domain.Agent, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockRegistry) GetAgentsByStatus(ctx context.Context, status domain.AgentStatus) ([]*domain.Agent, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockRegistry) GetAgentsByCapability(ctx context.Context, capability string) ([]*domain.Agent, error) {
	args := m.Called(ctx, capability)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockRegistry) UpdateAgentStatus(ctx context.Context, agentID string, status domain.AgentStatus) error {
	args := m.Called(ctx, agentID, status)
	return args.Error(0)
}

func (m *MockRegistry) UpdateAgentLastSeen(ctx context.Context, agentID string) error {
	args := m.Called(ctx, agentID)
	return args.Error(0)
}

func (m *MockRegistry) IsAgentHealthy(ctx context.Context, agentID string) (bool, error) {
	args := m.Called(ctx, agentID)
	return args.Bool(0), args.Error(1)
}
