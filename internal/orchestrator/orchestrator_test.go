package orchestrator

import (
	"context"
	"strings"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestOrchestrator_NewGraphOrchestrator tests creating a new graph-based orchestrator
func TestOrchestrator_NewGraphOrchestrator(t *testing.T) {
	// Test creating orchestrator with graph backend
	tests := []struct {
		name    string
		setup   func() (graph.Graph, events.Bus, agentRegistry.Registry)
		wantErr bool
	}{
		{
			name: "successful_orchestrator_creation",
			setup: func() (graph.Graph, events.Bus, agentRegistry.Registry) {
				// Create test dependencies
				mockGraph := &MockGraph{}
				mockEventBus := &MockEventBus{}
				mockRegistry := &MockAgentRegistry{}
				return mockGraph, mockEventBus, mockRegistry
			},
			wantErr: false,
		},
		{
			name: "nil_graph_should_fail",
			setup: func() (graph.Graph, events.Bus, agentRegistry.Registry) {
				mockEventBus := &MockEventBus{}
				mockRegistry := &MockAgentRegistry{}
				return nil, mockEventBus, mockRegistry
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, eventBus, registry := tt.setup()

			orchestrator, err := NewGraphOrchestrator(graph, eventBus, registry)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, orchestrator)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, orchestrator)
				assert.Equal(t, graph, orchestrator.graph)
				assert.Equal(t, eventBus, orchestrator.eventBus)
				assert.Equal(t, registry, orchestrator.agentRegistry)
			}
		})
	}
}

// TestOrchestrator_RegisterAgent tests agent registration in graph
func TestOrchestrator_RegisterAgent(t *testing.T) {
	orchestrator, mocks := setupTestOrchestrator(t)

	tests := []struct {
		name      string
		agent     agentRegistry.AgentInterface
		setupMock func()
		wantErr   bool
	}{
		{
			name:  "successful_agent_registration",
			agent: &MockAgent{name: "test-agent", capabilities: []agentRegistry.AgentCapability{{Pattern: "test.*"}}},
			setupMock: func() {
				// Mock graph operations for agent registration
				mocks.graph.On("CreateNode", "agent", map[string]interface{}{
					"name":         "test-agent",
					"capabilities": []string{"test.*"},
					"status":       "active",
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "duplicate_agent_should_fail",
			agent: &MockAgent{name: "existing-agent"},
			setupMock: func() {
				mocks.registry.On("GetAgent", context.Background(), "existing-agent").Return(&MockAgent{}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := orchestrator.RegisterAgent(context.Background(), tt.agent)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOrchestrator_CreateWorkflow tests workflow creation in graph
func TestOrchestrator_CreateWorkflow(t *testing.T) {
	orchestrator, mocks := setupTestOrchestrator(t)

	tests := []struct {
		name      string
		intent    string
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "simple_intent_creates_workflow",
			intent: "create application testapp",
			setupMock: func() {
				// Mock workflow creation in graph
				mocks.graph.On("CreateNode", "workflow", map[string]interface{}{
					"intent":     "create application testapp",
					"status":     "created",
					"created_at": mock.Anything,
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "empty_intent_should_fail",
			intent: "",
			setupMock: func() {
				// No mocks needed for validation failure
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			workflow, err := orchestrator.CreateWorkflow(context.Background(), tt.intent)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, workflow)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, workflow)
				assert.Equal(t, tt.intent, workflow.Intent)
				assert.NotEmpty(t, workflow.ID)
			}
		})
	}
}

// TestOrchestrator_QueryAgents tests graph-based agent discovery
func TestOrchestrator_QueryAgents(t *testing.T) {
	orchestrator, mocks := setupTestOrchestrator(t)

	tests := []struct {
		name        string
		capability  string
		setupMock   func()
		expectedLen int
		wantErr     bool
	}{
		{
			name:       "find_agents_by_capability",
			capability: "application.*",
			setupMock: func() {
				// Mock graph query for agents with capability
				mocks.graph.On("Query", mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "MATCH (a:agent)")
				}), mock.Anything).Return(&graph.QueryResult{
					Nodes: []graph.Node{
						{Properties: map[string]interface{}{"name": "application-agent"}},
					},
				}, nil)
			},
			expectedLen: 1,
			wantErr:     false,
		},
		{
			name:       "no_agents_found",
			capability: "nonexistent.*",
			setupMock: func() {
				mocks.graph.On("Query", mock.Anything, mock.Anything).Return(&graph.QueryResult{
					Nodes: []graph.Node{},
				}, nil)
			},
			expectedLen: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			agents, err := orchestrator.QueryAgentsByCapability(context.Background(), tt.capability)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, agents, tt.expectedLen)
			}
		})
	}
}

// Helper functions and mocks
func setupTestOrchestrator(t *testing.T) (*GraphOrchestrator, *TestMocks) {
	mocks := &TestMocks{
		graph:    &MockGraph{},
		eventBus: &MockEventBus{},
		registry: &MockAgentRegistry{},
	}

	orchestrator, err := NewGraphOrchestrator(mocks.graph, mocks.eventBus, mocks.registry)
	require.NoError(t, err)

	return orchestrator, mocks
}

type TestMocks struct {
	graph    *MockGraph
	eventBus *MockEventBus
	registry *MockAgentRegistry
}

// Mock implementations will be defined in separate files
