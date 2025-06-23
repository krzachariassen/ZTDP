package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ztdp/orchestrator/internal/graph"
	"github.com/ztdp/orchestrator/internal/types"
)

type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})             {}
func (m *MockLogger) Error(msg string, err error, fields ...interface{}) {}
func (m *MockLogger) Debug(msg string, fields ...interface{})            {}

func TestService_RegisterAgent(t *testing.T) {
	// Setup
	graphBackend := graph.NewEmbeddedGraph(&MockLogger{})
	service := NewService(graphBackend, &MockLogger{})
	ctx := context.Background()

	tests := []struct {
		name    string
		agent   *types.Agent
		wantErr bool
	}{
		{
			name: "successful agent registration",
			agent: &types.Agent{
				ID:           "agent-1",
				Name:         "test-agent",
				Capabilities: []string{"deploy", "test"},
			},
			wantErr: false,
		},
		{
			name: "empty agent ID should fail",
			agent: &types.Agent{
				ID:   "",
				Name: "test-agent",
			},
			wantErr: true,
		},
		{
			name: "empty agent name should fail",
			agent: &types.Agent{
				ID:   "agent-1",
				Name: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RegisterAgent(ctx, tt.agent)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetAgentsByCapability(t *testing.T) {
	// Setup
	graphBackend := graph.NewEmbeddedGraph(&MockLogger{})
	service := NewService(graphBackend, &MockLogger{})
	ctx := context.Background()

	// Register test agents
	agents := []*types.Agent{
		{
			ID:           "agent-1",
			Name:         "deploy-agent",
			Capabilities: []string{"deploy", "validate"},
			Status:       types.AgentStatusActive,
		},
		{
			ID:           "agent-2",
			Name:         "test-agent",
			Capabilities: []string{"test", "validate"},
			Status:       types.AgentStatusActive,
		},
		{
			ID:           "agent-3",
			Name:         "monitor-agent",
			Capabilities: []string{"monitor"},
			Status:       types.AgentStatusInactive,
		},
	}

	for _, agent := range agents {
		err := service.RegisterAgent(ctx, agent)
		assert.NoError(t, err)
	}

	tests := []struct {
		name           string
		capability     string
		expectedAgents []string
	}{
		{
			name:           "find deploy agents",
			capability:     "deploy",
			expectedAgents: []string{"agent-1"},
		},
		{
			name:           "find validate agents",
			capability:     "validate",
			expectedAgents: []string{"agent-1", "agent-2"},
		},
		{
			name:           "find monitor agents",
			capability:     "monitor",
			expectedAgents: []string{"agent-3"}, // Include inactive agents
		},
		{
			name:           "find non-existent capability",
			capability:     "non-existent",
			expectedAgents: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundAgents, err := service.GetAgentsByCapability(ctx, tt.capability)
			assert.NoError(t, err)

			var foundIDs []string
			for _, agent := range foundAgents {
				foundIDs = append(foundIDs, agent.ID)
			}

			assert.ElementsMatch(t, tt.expectedAgents, foundIDs)
		})
	}
}

func TestService_UpdateAgentStatus(t *testing.T) {
	// Setup
	graphBackend := graph.NewEmbeddedGraph(&MockLogger{})
	service := NewService(graphBackend, &MockLogger{})
	ctx := context.Background()

	// Register test agent
	agent := &types.Agent{
		ID:     "agent-1",
		Name:   "test-agent",
		Status: types.AgentStatusInactive,
	}
	err := service.RegisterAgent(ctx, agent)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		agentID string
		status  string
		wantErr bool
	}{
		{
			name:    "successful status update",
			agentID: "agent-1",
			status:  types.AgentStatusActive,
			wantErr: false,
		},
		{
			name:    "empty agent ID should fail",
			agentID: "",
			status:  types.AgentStatusActive,
			wantErr: true,
		},
		{
			name:    "invalid status should fail",
			agentID: "agent-1",
			status:  "invalid-status",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateAgentStatus(ctx, tt.agentID, tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetActiveAgents(t *testing.T) {
	// Setup
	graphBackend := graph.NewEmbeddedGraph(&MockLogger{})
	service := NewService(graphBackend, &MockLogger{})
	ctx := context.Background()

	// Register agents with different statuses
	agents := []*types.Agent{
		{ID: "agent-1", Name: "active-1", Status: types.AgentStatusActive},
		{ID: "agent-2", Name: "active-2", Status: types.AgentStatusActive},
		{ID: "agent-3", Name: "inactive-1", Status: types.AgentStatusInactive},
	}

	for _, agent := range agents {
		err := service.RegisterAgent(ctx, agent)
		assert.NoError(t, err)
	}

	// Get active agents
	activeAgents, err := service.GetActiveAgents(ctx)
	assert.NoError(t, err)
	assert.Len(t, activeAgents, 2)

	var activeIDs []string
	for _, agent := range activeAgents {
		activeIDs = append(activeIDs, agent.ID)
	}
	assert.ElementsMatch(t, []string{"agent-1", "agent-2"}, activeIDs)
}
