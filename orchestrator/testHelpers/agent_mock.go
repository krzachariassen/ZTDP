package testHelpers

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/ztdp/orchestrator/internal/agent/domain"
)

// MockAgentService provides a testify-based mock for agent service operations
type MockAgentService struct {
	mock.Mock
}

// NewMockAgentService creates a new mock agent service instance
func NewMockAgentService() *MockAgentService {
	return &MockAgentService{}
}

func (m *MockAgentService) GetAgentsByCapability(ctx context.Context, capability string) ([]*domain.Agent, error) {
	args := m.Called(ctx, capability)
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockAgentService) UpdateAgentStatus(ctx context.Context, agentID, status string) error {
	args := m.Called(ctx, agentID, status)
	return args.Error(0)
}

// MockAgentRepository provides a testify-based mock for agent repository operations
type MockAgentRepository struct {
	mock.Mock
}

// NewMockAgentRepository creates a new mock agent repository instance
func NewMockAgentRepository() *MockAgentRepository {
	return &MockAgentRepository{}
}

func (m *MockAgentRepository) RegisterAgent(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) GetAgent(ctx context.Context, id string) (*domain.Agent, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetAgentsByCapability(ctx context.Context, capability string) ([]*domain.Agent, error) {
	args := m.Called(ctx, capability)
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) UpdateAgentStatus(ctx context.Context, agentID, status string) error {
	args := m.Called(ctx, agentID, status)
	return args.Error(0)
}

func (m *MockAgentRepository) IsAgentHealthy(ctx context.Context, agentID string) (bool, error) {
	args := m.Called(ctx, agentID)
	return args.Bool(0), args.Error(1)
}
