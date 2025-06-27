package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztdp/orchestrator/internal/agent/domain"
)

// MockAgentService for testing
type MockAgentService struct {
	mock.Mock
}

func (m *MockAgentService) GetAvailableAgents(ctx context.Context) ([]*domain.Agent, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockAgentService) RegisterAgent(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentService) DiscoverAgentsByCapability(ctx context.Context, capability string) ([]*domain.Agent, error) {
	args := m.Called(ctx, capability)
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockAgentService) UpdateAgentStatus(ctx context.Context, agentID string, status domain.AgentStatus) error {
	args := m.Called(ctx, agentID, status)
	return args.Error(0)
}

func TestGraphExplorer_GetAgentContext(t *testing.T) {
	t.Run("should format agents for AI consumption", func(t *testing.T) {
		mockAgentService := &MockAgentService{}
		explorer := NewGraphExplorer(mockAgentService)

		// Create test agents
		agent1 := &domain.Agent{
			ID:     "deploy-agent-001",
			Name:   "Deploy Agent",
			Status: domain.AgentStatusOnline,
			Capabilities: []domain.AgentCapability{
				{Name: "deploy", Description: "Deploy applications"},
				{Name: "test", Description: "Run tests"},
			},
		}

		agent2 := &domain.Agent{
			ID:     "monitor-agent-001",
			Name:   "Monitor Agent",
			Status: domain.AgentStatusOnline,
			Capabilities: []domain.AgentCapability{
				{Name: "monitor", Description: "Monitor systems"},
			},
		}

		agents := []*domain.Agent{agent1, agent2}
		mockAgentService.On("GetAvailableAgents", mock.Anything).Return(agents, nil)

		context, err := explorer.GetAgentContext(context.Background())

		assert.NoError(t, err)
		assert.Contains(t, context, "Deploy Agent")
		assert.Contains(t, context, "deploy, test")
		assert.Contains(t, context, "Monitor Agent")
		assert.Contains(t, context, "monitor")
		assert.Contains(t, context, "online")
		mockAgentService.AssertExpectations(t)
	})

	t.Run("should handle no agents available", func(t *testing.T) {
		mockAgentService := &MockAgentService{}
		explorer := NewGraphExplorer(mockAgentService)

		mockAgentService.On("GetAvailableAgents", mock.Anything).Return([]*domain.Agent{}, nil)

		context, err := explorer.GetAgentContext(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, "No agents currently registered", context)
		mockAgentService.AssertExpectations(t)
	})
}

func TestGraphExplorer_FindCapableAgents(t *testing.T) {
	t.Run("should find agents with specific capabilities", func(t *testing.T) {
		mockAgentService := &MockAgentService{}
		explorer := NewGraphExplorer(mockAgentService)

		deployAgent := &domain.Agent{
			ID:   "deploy-agent-001",
			Name: "Deploy Agent",
			Capabilities: []domain.AgentCapability{
				{Name: "deploy", Description: "Deploy applications"},
			},
		}

		capabilities := []string{"deploy"}
		mockAgentService.On("DiscoverAgentsByCapability", mock.Anything, "deploy").Return([]*domain.Agent{deployAgent}, nil)

		agents, err := explorer.FindCapableAgents(context.Background(), capabilities)

		assert.NoError(t, err)
		assert.Len(t, agents, 1)
		assert.Equal(t, "deploy-agent-001", agents[0].ID)
		mockAgentService.AssertExpectations(t)
	})
}
