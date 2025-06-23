package graph

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockLogger for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})             {}
func (m *MockLogger) Error(msg string, err error, fields ...interface{}) {}
func (m *MockLogger) Debug(msg string, fields ...interface{})            {}

func TestEmbeddedGraph_AddNode(t *testing.T) {
	graph := NewEmbeddedGraph(&MockLogger{})
	ctx := context.Background()

	tests := []struct {
		name       string
		nodeType   string
		nodeID     string
		properties map[string]interface{}
		wantErr    bool
	}{
		{
			name:     "successful node addition",
			nodeType: "agent",
			nodeID:   "agent-1",
			properties: map[string]interface{}{
				"name":         "test-agent",
				"capabilities": []string{"deploy", "test"},
				"status":       "active",
			},
			wantErr: false,
		},
		{
			name:     "empty node type should fail",
			nodeType: "",
			nodeID:   "agent-1",
			wantErr:  true,
		},
		{
			name:     "empty node ID should fail",
			nodeType: "agent",
			nodeID:   "",
			wantErr:  true,
		},
		{
			name:       "duplicate node should fail",
			nodeType:   "agent",
			nodeID:     "agent-1", // Same as first test
			properties: map[string]interface{}{"name": "duplicate"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := graph.AddNode(ctx, tt.nodeType, tt.nodeID, tt.properties)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmbeddedGraph_GetNode(t *testing.T) {
	graph := NewEmbeddedGraph(&MockLogger{})
	ctx := context.Background()

	// Add a test node
	properties := map[string]interface{}{
		"name":         "test-agent",
		"capabilities": []string{"deploy", "test"},
		"status":       "active",
	}
	err := graph.AddNode(ctx, "agent", "agent-1", properties)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		nodeType string
		nodeID   string
		wantErr  bool
	}{
		{
			name:     "get existing node",
			nodeType: "agent",
			nodeID:   "agent-1",
			wantErr:  false,
		},
		{
			name:     "get non-existent node",
			nodeType: "agent",
			nodeID:   "agent-2",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := graph.GetNode(ctx, tt.nodeType, tt.nodeID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.nodeType, result["type"])
				assert.Equal(t, tt.nodeID, result["id"])
				assert.Equal(t, "test-agent", result["name"])
			}
		})
	}
}

func TestEmbeddedGraph_UpdateNode(t *testing.T) {
	graph := NewEmbeddedGraph(&MockLogger{})
	ctx := context.Background()

	// Add a test node
	err := graph.AddNode(ctx, "agent", "agent-1", map[string]interface{}{
		"name":   "original-name",
		"status": "inactive",
	})
	assert.NoError(t, err)

	// Update the node
	err = graph.UpdateNode(ctx, "agent", "agent-1", map[string]interface{}{
		"status":   "active",
		"endpoint": "http://localhost:8080",
	})
	assert.NoError(t, err)

	// Verify the update
	result, err := graph.GetNode(ctx, "agent", "agent-1")
	assert.NoError(t, err)
	assert.Equal(t, "original-name", result["name"])             // Original property preserved
	assert.Equal(t, "active", result["status"])                  // Updated property
	assert.Equal(t, "http://localhost:8080", result["endpoint"]) // New property
}

func TestEmbeddedGraph_QueryNodes(t *testing.T) {
	graph := NewEmbeddedGraph(&MockLogger{})
	ctx := context.Background()

	// Add multiple test nodes
	agents := []struct {
		id     string
		status string
		cap    string
	}{
		{"agent-1", "active", "deploy"},
		{"agent-2", "active", "test"},
		{"agent-3", "inactive", "deploy"},
	}

	for _, agent := range agents {
		err := graph.AddNode(ctx, "agent", agent.id, map[string]interface{}{
			"status":     agent.status,
			"capability": agent.cap,
		})
		assert.NoError(t, err)
	}

	tests := []struct {
		name        string
		nodeType    string
		filters     map[string]interface{}
		expectedLen int
	}{
		{
			name:        "query all agents",
			nodeType:    "agent",
			filters:     nil,
			expectedLen: 3,
		},
		{
			name:     "query active agents",
			nodeType: "agent",
			filters: map[string]interface{}{
				"status": "active",
			},
			expectedLen: 2,
		},
		{
			name:     "query deploy agents",
			nodeType: "agent",
			filters: map[string]interface{}{
				"capability": "deploy",
			},
			expectedLen: 2,
		},
		{
			name:     "query active deploy agents",
			nodeType: "agent",
			filters: map[string]interface{}{
				"status":     "active",
				"capability": "deploy",
			},
			expectedLen: 1,
		},
		{
			name:        "query non-existent type",
			nodeType:    "workflow",
			filters:     nil,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := graph.QueryNodes(ctx, tt.nodeType, tt.filters)
			assert.NoError(t, err)
			assert.Len(t, results, tt.expectedLen)
		})
	}
}

func TestEmbeddedGraph_DeleteNode(t *testing.T) {
	graph := NewEmbeddedGraph(&MockLogger{})
	ctx := context.Background()

	// Add a test node
	err := graph.AddNode(ctx, "agent", "agent-1", map[string]interface{}{
		"name": "test-agent",
	})
	assert.NoError(t, err)

	// Verify node exists
	_, err = graph.GetNode(ctx, "agent", "agent-1")
	assert.NoError(t, err)

	// Delete the node
	err = graph.DeleteNode(ctx, "agent", "agent-1")
	assert.NoError(t, err)

	// Verify node is deleted
	_, err = graph.GetNode(ctx, "agent", "agent-1")
	assert.Error(t, err)

	// Try to delete non-existent node
	err = graph.DeleteNode(ctx, "agent", "agent-2")
	assert.Error(t, err)
}

func TestEmbeddedGraph_GetStats(t *testing.T) {
	graph := NewEmbeddedGraph(&MockLogger{})
	ctx := context.Background()

	// Initially empty
	stats := graph.GetStats()
	assert.Equal(t, 0, stats["total_nodes"])
	assert.Equal(t, "embedded", stats["implementation"])

	// Add some nodes
	err := graph.AddNode(ctx, "agent", "agent-1", map[string]interface{}{})
	assert.NoError(t, err)

	err = graph.AddNode(ctx, "workflow", "workflow-1", map[string]interface{}{})
	assert.NoError(t, err)

	// Check updated stats
	stats = graph.GetStats()
	assert.Equal(t, 2, stats["total_nodes"])

	nodesByType := stats["nodes_by_type"].(map[string]int)
	assert.Equal(t, 1, nodesByType["agent"])
	assert.Equal(t, 1, nodesByType["workflow"])
}

func TestGraphFactory_CreateGraph(t *testing.T) {
	factory := NewGraphFactory(&MockLogger{})

	tests := []struct {
		name    string
		config  GraphConfig
		wantErr bool
		skip    bool
	}{
		{
			name: "create embedded graph",
			config: GraphConfig{
				Backend: GraphBackendEmbedded,
			},
			wantErr: false,
		},
		{
			name: "create neo4j graph",
			config: GraphConfig{
				Backend:       GraphBackendNeo4j,
				Neo4jURL:      "bolt://localhost:7687",
				Neo4jUser:     "neo4j",
				Neo4jPassword: "orchestrator123",
			},
			wantErr: false,
			skip:    true, // Skip unless Neo4j is available
		},
		{
			name: "unsupported backend",
			config: GraphConfig{
				Backend: "unsupported",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip && testing.Short() {
				t.Skip("skipping Neo4j test in short mode")
			}

			graph, err := factory.CreateGraph(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, graph)
			} else {
				if tt.skip {
					// For Neo4j, allow error if service is not available
					if err != nil {
						t.Skipf("Neo4j not available: %v", err)
					}
				}
				if err == nil {
					assert.NotNil(t, graph)
					// Cleanup Neo4j connections
					if neo4jGraph, ok := graph.(*Neo4jGraph); ok {
						defer neo4jGraph.Close(context.Background())
					}
				}
			}
		})
	}
}
