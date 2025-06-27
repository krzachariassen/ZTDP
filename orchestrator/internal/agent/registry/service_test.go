package registry_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ztdp/orchestrator/internal/agent/domain"
	"github.com/ztdp/orchestrator/internal/agent/registry"
	"github.com/ztdp/orchestrator/internal/logging"
	"github.com/ztdp/orchestrator/testHelpers"
)

func TestAgentRegistry_RegisterAgent_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewStructuredLogger(logging.LevelError) // Reduce noise in tests

	// Create mock graph for testing
	testGraph := testHelpers.NewCleanMockGraph()

	registryService := registry.NewService(testGraph, logger)

	// Create test agent
	agent := &domain.Agent{
		ID:          "test-agent-1",
		Name:        "Test Agent",
		Description: "A test agent for unit testing",
		Status:      domain.AgentStatusOnline,
		Capabilities: []domain.AgentCapability{
			{
				Name:        "text-processing",
				Description: "Can process text",
				Parameters:  map[string]string{"model": "gpt-4"},
			},
		},
		Metadata:  map[string]string{"env": "test"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	// Act
	err := registryService.RegisterAgent(ctx, agent)

	// Assert
	assert.NoError(t, err)

	// Verify agent was registered
	retrievedAgent, err := registryService.GetAgent(ctx, agent.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.ID, retrievedAgent.ID)
	assert.Equal(t, agent.Name, retrievedAgent.Name)
	assert.Equal(t, agent.Status, retrievedAgent.Status)
}

func TestAgentRegistry_RegisterAgent_ValidationErrors(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewStructuredLogger(logging.LevelError)

	testGraph := testHelpers.NewCleanMockGraph()

	registryService := registry.NewService(testGraph, logger)

	tests := []struct {
		name        string
		agent       *domain.Agent
		expectedErr string
	}{
		{
			name:        "nil agent",
			agent:       nil,
			expectedErr: "agent cannot be nil",
		},
		{
			name: "empty ID",
			agent: &domain.Agent{
				Name: "Test Agent",
			},
			expectedErr: "agent ID cannot be empty",
		},
		{
			name: "empty name",
			agent: &domain.Agent{
				ID: "test-id",
			},
			expectedErr: "agent name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := registryService.RegisterAgent(ctx, tt.agent)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestAgentRegistry_GetAgentsByCapability(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewStructuredLogger(logging.LevelError)

	testGraph := testHelpers.NewCleanMockGraph()

	registryService := registry.NewService(testGraph, logger)

	// Register multiple agents with different capabilities
	agents := []*domain.Agent{
		{
			ID:     "agent-1",
			Name:   "Text Processor",
			Status: domain.AgentStatusOnline,
			Capabilities: []domain.AgentCapability{
				{Name: "text-processing", Description: "Process text"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:     "agent-2",
			Name:   "Image Processor",
			Status: domain.AgentStatusOnline,
			Capabilities: []domain.AgentCapability{
				{Name: "image-processing", Description: "Process images"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:     "agent-3",
			Name:   "Multi Processor",
			Status: domain.AgentStatusOnline,
			Capabilities: []domain.AgentCapability{
				{Name: "text-processing", Description: "Process text"},
				{Name: "image-processing", Description: "Process images"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, agent := range agents {
		err := registryService.RegisterAgent(ctx, agent)
		require.NoError(t, err)
	}

	// Act
	textProcessors, err := registryService.GetAgentsByCapability(ctx, "text-processing")

	// Assert
	require.NoError(t, err)
	assert.Len(t, textProcessors, 2) // agent-1 and agent-3

	agentIDs := make([]string, len(textProcessors))
	for i, agent := range textProcessors {
		agentIDs[i] = agent.ID
	}
	assert.Contains(t, agentIDs, "agent-1")
	assert.Contains(t, agentIDs, "agent-3")
}

func TestAgentRegistry_UpdateAgentStatus(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewStructuredLogger(logging.LevelError)

	testGraph := testHelpers.NewCleanMockGraph()

	registryService := registry.NewService(testGraph, logger)

	// Register an agent
	agent := &domain.Agent{
		ID:        "test-agent",
		Name:      "Test Agent",
		Status:    domain.AgentStatusOnline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := registryService.RegisterAgent(ctx, agent)
	require.NoError(t, err)

	// Act - Update status
	err = registryService.UpdateAgentStatus(ctx, agent.ID, domain.AgentStatusBusy)

	// Assert
	require.NoError(t, err)

	// Verify status was updated
	updatedAgent, err := registryService.GetAgent(ctx, agent.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.AgentStatusBusy, updatedAgent.Status)
}

func TestAgentRegistry_IsAgentHealthy(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewStructuredLogger(logging.LevelError)

	testGraph := testHelpers.NewCleanMockGraph()

	registryService := registry.NewService(testGraph, logger)

	// Register an agent
	agent := &domain.Agent{
		ID:        "healthy-agent",
		Name:      "Healthy Agent",
		Status:    domain.AgentStatusOnline,
		LastSeen:  time.Now(), // Recent last seen
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := registryService.RegisterAgent(ctx, agent)
	require.NoError(t, err)

	// Act
	isHealthy, err := registryService.IsAgentHealthy(ctx, agent.ID)

	// Assert
	require.NoError(t, err)
	assert.True(t, isHealthy)
}

// Interface compliance test
func TestAgentRegistry_ImplementsInterface(t *testing.T) {
	// Arrange
	logger := logging.NewStructuredLogger(logging.LevelError)
	testGraph := testHelpers.NewCleanMockGraph()

	// Act & Assert - This will fail to compile if Service doesn't implement AgentRegistry
	var _ domain.AgentRegistry = registry.NewService(testGraph, logger)
}
