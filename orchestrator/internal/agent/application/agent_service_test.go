package application

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ztdp/orchestrator/internal/agent/domain"
)

// 🔴 RED PHASE: Test-driven development for Agent Application Services
// These tests define the expected behavior of our application layer

func TestAgentService_RegisterAgent(t *testing.T) {
	// Setup
	mockRepo := &MockAgentRepository{}
	service := NewAgentService(mockRepo)

	// Create test agent
	capabilities := []domain.AgentCapability{
		{Name: "excel-analysis", Description: "Excel file analysis"},
		{Name: "data-extraction", Description: "Extract data from Excel"},
	}

	agent, err := domain.NewAgent("excel-processor-001", "Excel Processor", "Processes Excel files", capabilities)
	require.NoError(t, err)

	// Setup mock expectations - agent doesn't exist yet, so GetByID returns error
	mockRepo.On("GetByID", mock.Anything, agent.ID).Return(nil, fmt.Errorf("agent not found"))
	mockRepo.On("Create", mock.Anything, agent).Return(nil)

	// Execute
	err = service.RegisterAgent(context.Background(), agent)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAgentService_RegisterAgent_DuplicateAgent(t *testing.T) {
	// 🔴 RED: Test duplicate agent registration should fail
	// Setup
	mockRepo := &MockAgentRepository{}
	service := NewAgentService(mockRepo)

	// Create test agent
	capabilities := []domain.AgentCapability{
		{Name: "excel-analysis", Description: "Excel file analysis"},
	}

	agent, err := domain.NewAgent("excel-processor-001", "Excel Processor", "Processes Excel files", capabilities)
	require.NoError(t, err)

	// Setup mock expectations - agent already exists
	existingAgent := agent // Simulate finding the existing agent
	mockRepo.On("GetByID", mock.Anything, agent.ID).Return(existingAgent, nil)

	// Execute
	err = service.RegisterAgent(context.Background(), agent)

	// Assert - should fail with duplicate error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestAgentService_DiscoverAgentsByCapability(t *testing.T) {
	// Setup
	mockRepo := &MockAgentRepository{}
	service := NewAgentService(mockRepo)

	// Create test agents
	excelAgent, _ := domain.NewAgent("excel-processor", "Excel Processor", "Excel processing",
		[]domain.AgentCapability{{Name: "excel-analysis", Description: "Excel analysis"}})
	// Set agent to online status for the test
	excelAgent.UpdateStatus(domain.AgentStatusOnline)

	// Note: pptAgent not used in this specific test but kept for other test cases
	_, _ = domain.NewAgent("ppt-creator", "PowerPoint Creator", "PPT creation",
		[]domain.AgentCapability{{Name: "powerpoint-creation", Description: "PowerPoint creation"}})

	mockRepo.On("GetByCapability", mock.Anything, "excel-analysis").Return([]*domain.Agent{excelAgent}, nil)

	// Execute
	agents, err := service.DiscoverAgentsByCapability(context.Background(), "excel-analysis")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, agents, 1)
	assert.Equal(t, "excel-processor", agents[0].ID)
	mockRepo.AssertExpectations(t)
}

func TestAgentService_GetAvailableAgents(t *testing.T) {
	// Setup
	mockRepo := &MockAgentRepository{}
	service := NewAgentService(mockRepo)

	// Create test agents
	onlineAgent, _ := domain.NewAgent("agent-1", "Agent 1", "Online agent",
		[]domain.AgentCapability{{Name: "test", Description: "Test capability"}})
	onlineAgent.UpdateStatus(domain.AgentStatusOnline)

	mockRepo.On("GetByStatus", mock.Anything, domain.AgentStatusOnline).Return([]*domain.Agent{onlineAgent}, nil)

	// Execute
	agents, err := service.GetAvailableAgents(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Len(t, agents, 1)
	assert.Equal(t, domain.AgentStatusOnline, agents[0].Status)
	mockRepo.AssertExpectations(t)
}

func TestAgentService_UpdateAgentStatus(t *testing.T) {
	// Setup
	mockRepo := &MockAgentRepository{}
	service := NewAgentService(mockRepo)

	// Setup expectations
	mockRepo.On("UpdateStatus", mock.Anything, "agent-1", domain.AgentStatusBusy).Return(nil)
	mockRepo.On("UpdateLastSeen", mock.Anything, "agent-1").Return(nil)

	// Execute
	err := service.UpdateAgentStatus(context.Background(), "agent-1", domain.AgentStatusBusy)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Mock implementation for testing
type MockAgentRepository struct {
	mock.Mock
}

func (m *MockAgentRepository) Create(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) GetByID(ctx context.Context, id string) (*domain.Agent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetAll(ctx context.Context) ([]*domain.Agent, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetByStatus(ctx context.Context, status domain.AgentStatus) ([]*domain.Agent, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetByCapability(ctx context.Context, capabilityName string) ([]*domain.Agent, error) {
	args := m.Called(ctx, capabilityName)
	return args.Get(0).([]*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) Update(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentRepository) UpdateStatus(ctx context.Context, id string, status domain.AgentStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockAgentRepository) UpdateLastSeen(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
