package domain

import (
	"testing"
	"time"
)

func TestAgent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		agent   Agent
		wantErr error
	}{
		{
			name: "valid agent",
			agent: Agent{
				ID:          "test-agent-1",
				Name:        "Test Agent",
				Description: "A test agent",
				Status:      AgentStatusOnline,
				Capabilities: []AgentCapability{
					{Name: "test-capability", Description: "Test capability"},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				LastSeen:  time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "empty ID",
			agent: Agent{
				ID:   "",
				Name: "Test Agent",
				Capabilities: []AgentCapability{
					{Name: "test-capability", Description: "Test capability"},
				},
			},
			wantErr: ErrInvalidAgentID,
		},
		{
			name: "invalid ID with special characters",
			agent: Agent{
				ID:   "test@agent!",
				Name: "Test Agent",
				Capabilities: []AgentCapability{
					{Name: "test-capability", Description: "Test capability"},
				},
			},
			wantErr: ErrInvalidAgentID,
		},
		{
			name: "empty name",
			agent: Agent{
				ID:   "test-agent",
				Name: "",
				Capabilities: []AgentCapability{
					{Name: "test-capability", Description: "Test capability"},
				},
			},
			wantErr: ErrInvalidAgentName,
		},
		{
			name: "name too long",
			agent: Agent{
				ID:   "test-agent",
				Name: string(make([]byte, 101)),
				Capabilities: []AgentCapability{
					{Name: "test-capability", Description: "Test capability"},
				},
			},
			wantErr: ErrInvalidAgentName,
		},
		{
			name: "description too long",
			agent: Agent{
				ID:          "test-agent",
				Name:        "Test Agent",
				Description: string(make([]byte, 501)),
				Capabilities: []AgentCapability{
					{Name: "test-capability", Description: "Test capability"},
				},
			},
			wantErr: ErrInvalidAgentDescription,
		},
		{
			name: "invalid status",
			agent: Agent{
				ID:     "test-agent",
				Name:   "Test Agent",
				Status: AgentStatus("invalid"),
				Capabilities: []AgentCapability{
					{Name: "test-capability", Description: "Test capability"},
				},
			},
			wantErr: ErrInvalidStatus,
		},
		{
			name: "no capabilities",
			agent: Agent{
				ID:           "test-agent",
				Name:         "Test Agent",
				Status:       AgentStatusOnline,
				Capabilities: []AgentCapability{},
			},
			wantErr: ErrNoCapabilities,
		},
		{
			name: "invalid capability name",
			agent: Agent{
				ID:     "test-agent",
				Name:   "Test Agent",
				Status: AgentStatusOnline,
				Capabilities: []AgentCapability{
					{Name: "", Description: "Empty name capability"},
				},
			},
			wantErr: ErrInvalidCapability,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.agent.Validate()
			if err != tt.wantErr {
				t.Errorf("Agent.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_IsHealthy(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		agent    Agent
		expected bool
	}{
		{
			name: "healthy online agent",
			agent: Agent{
				Status:   AgentStatusOnline,
				LastSeen: now.Add(-1 * time.Minute),
			},
			expected: true,
		},
		{
			name: "unhealthy offline agent",
			agent: Agent{
				Status:   AgentStatusOffline,
				LastSeen: now.Add(-1 * time.Minute),
			},
			expected: false,
		},
		{
			name: "unhealthy agent - last seen too long ago",
			agent: Agent{
				Status:   AgentStatusOnline,
				LastSeen: now.Add(-15 * time.Minute),
			},
			expected: false,
		},
		{
			name: "busy agent is healthy",
			agent: Agent{
				Status:   AgentStatusBusy,
				LastSeen: now.Add(-1 * time.Minute),
			},
			expected: true,
		},
		{
			name: "maintenance agent is not healthy",
			agent: Agent{
				Status:   AgentStatusMaintenance,
				LastSeen: now.Add(-1 * time.Minute),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.agent.IsHealthy()
			if result != tt.expected {
				t.Errorf("Agent.IsHealthy() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestAgent_HasCapability(t *testing.T) {
	agent := Agent{
		Capabilities: []AgentCapability{
			{Name: "text-processing", Description: "Process text"},
			{Name: "data-analysis", Description: "Analyze data"},
		},
	}

	tests := []struct {
		name       string
		capability string
		expected   bool
	}{
		{
			name:       "existing capability",
			capability: "text-processing",
			expected:   true,
		},
		{
			name:       "non-existing capability",
			capability: "image-processing",
			expected:   false,
		},
		{
			name:       "empty capability",
			capability: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.HasCapability(tt.capability)
			if result != tt.expected {
				t.Errorf("Agent.HasCapability(%s) = %v, expected %v", tt.capability, result, tt.expected)
			}
		})
	}
}

func TestAgent_UpdateStatus(t *testing.T) {
	agent := Agent{
		Status:    AgentStatusOffline,
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	originalUpdatedAt := agent.UpdatedAt

	err := agent.UpdateStatus(AgentStatusOnline)
	if err != nil {
		t.Errorf("Agent.UpdateStatus() error = %v, expected nil", err)
	}

	if agent.Status != AgentStatusOnline {
		t.Errorf("Agent.Status = %v, expected %v", agent.Status, AgentStatusOnline)
	}

	if !agent.UpdatedAt.After(originalUpdatedAt) {
		t.Errorf("Agent.UpdatedAt should be updated, but it wasn't")
	}
}

func TestAgent_UpdateStatus_Invalid(t *testing.T) {
	agent := Agent{
		Status: AgentStatusOnline,
	}

	err := agent.UpdateStatus(AgentStatus("invalid"))
	if err != ErrInvalidStatus {
		t.Errorf("Agent.UpdateStatus() error = %v, expected %v", err, ErrInvalidStatus)
	}

	if agent.Status != AgentStatusOnline {
		t.Errorf("Agent.Status should remain unchanged on error")
	}
}

func TestNewAgent(t *testing.T) {
	id := "test-agent"
	name := "Test Agent"
	description := "A test agent"
	capabilities := []AgentCapability{
		{Name: "test-capability", Description: "Test capability"},
	}

	agent, err := NewAgent(id, name, description, capabilities)
	if err != nil {
		t.Errorf("NewAgent() error = %v, expected nil", err)
	}

	if agent.ID != id {
		t.Errorf("Agent.ID = %v, expected %v", agent.ID, id)
	}

	if agent.Name != name {
		t.Errorf("Agent.Name = %v, expected %v", agent.Name, name)
	}

	if agent.Description != description {
		t.Errorf("Agent.Description = %v, expected %v", agent.Description, description)
	}

	if agent.Status != AgentStatusOffline {
		t.Errorf("Agent.Status = %v, expected %v", agent.Status, AgentStatusOffline)
	}

	if len(agent.Capabilities) != len(capabilities) {
		t.Errorf("Agent.Capabilities length = %v, expected %v", len(agent.Capabilities), len(capabilities))
	}

	if agent.CreatedAt.IsZero() {
		t.Errorf("Agent.CreatedAt should be set")
	}

	if agent.UpdatedAt.IsZero() {
		t.Errorf("Agent.UpdatedAt should be set")
	}
}

func TestNewAgent_ValidationError(t *testing.T) {
	// Test with invalid ID
	_, err := NewAgent("", "Test Agent", "Description", []AgentCapability{
		{Name: "test-capability", Description: "Test capability"},
	})
	if err != ErrInvalidAgentID {
		t.Errorf("NewAgent() with empty ID error = %v, expected %v", err, ErrInvalidAgentID)
	}

	// Test with no capabilities
	_, err = NewAgent("test-agent", "Test Agent", "Description", []AgentCapability{})
	if err != ErrNoCapabilities {
		t.Errorf("NewAgent() with no capabilities error = %v, expected %v", err, ErrNoCapabilities)
	}
}
