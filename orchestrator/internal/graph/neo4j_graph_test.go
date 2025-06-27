package graph

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ztdp/orchestrator/internal/logging"
)

// TestNeo4jGraph_Integration tests Neo4j graph operations
// This test requires a running Neo4j instance (use docker-compose up neo4j)
func TestNeo4jGraph_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := logging.NewNoOpLogger()
	config := GraphConfig{
		Backend:       GraphBackendNeo4j,
		Neo4jURL:      "bolt://localhost:7687",
		Neo4jUser:     "neo4j",
		Neo4jPassword: "orchestrator123",
	}

	ctx := context.Background()
	graph, err := NewNeo4jGraph(ctx, config, logger)
	require.NoError(t, err)
	defer graph.Close(ctx)

	// Clear any existing test data
	err = graph.ClearTestData(ctx)
	require.NoError(t, err)

	// Test node operations
	t.Run("Node Operations", func(t *testing.T) {
		// AddNode
		properties := map[string]interface{}{
			"name":         "test-agent",
			"capabilities": []string{"deploy", "test"},
			"status":       "active",
		}
		err := graph.AddNode(ctx, "Agent", "agent-1", properties)
		assert.NoError(t, err)

		// GetNode
		result, err := graph.GetNode(ctx, "Agent", "agent-1")
		assert.NoError(t, err)
		assert.Equal(t, "Agent", result["type"])
		assert.Equal(t, "agent-1", result["id"])
		assert.Equal(t, "test-agent", result["name"])
		assert.Equal(t, "active", result["status"])

		// UpdateNode
		err = graph.UpdateNode(ctx, "Agent", "agent-1", map[string]interface{}{
			"status":   "inactive",
			"endpoint": "http://localhost:8080",
		})
		assert.NoError(t, err)

		// Verify update
		result, err = graph.GetNode(ctx, "Agent", "agent-1")
		assert.NoError(t, err)
		assert.Equal(t, "inactive", result["status"])
		assert.Equal(t, "http://localhost:8080", result["endpoint"])
		assert.Equal(t, "test-agent", result["name"]) // Original property preserved

		// QueryNodes
		err = graph.AddNode(ctx, "Agent", "agent-2", map[string]interface{}{
			"name":   "another-agent",
			"status": "active",
		})
		require.NoError(t, err)

		// Query all agents
		results, err := graph.QueryNodes(ctx, "Agent", nil)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)

		// Query active agents
		results, err = graph.QueryNodes(ctx, "Agent", map[string]interface{}{
			"status": "active",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "agent-2", results[0]["id"])

		// DeleteNode
		err = graph.DeleteNode(ctx, "Agent", "agent-1")
		assert.NoError(t, err)

		// Verify deletion
		_, err = graph.GetNode(ctx, "Agent", "agent-1")
		assert.Error(t, err)
	})

	// Test edge operations
	t.Run("Edge Operations", func(t *testing.T) {
		// Add two nodes to connect
		err := graph.AddNode(ctx, "Agent", "agent-source", map[string]interface{}{
			"name": "source-agent",
		})
		require.NoError(t, err)

		err = graph.AddNode(ctx, "Agent", "agent-target", map[string]interface{}{
			"name": "target-agent",
		})
		require.NoError(t, err)

		// AddEdge
		edgeProperties := map[string]interface{}{
			"relationship": "communicates_with",
			"created_at":   "2024-01-01",
		}
		err = graph.AddEdge(ctx, "Agent", "agent-source", "Agent", "agent-target", "CONNECTS", edgeProperties)
		assert.NoError(t, err)

		// GetEdges
		edges, err := graph.GetEdges(ctx, "Agent", "agent-source")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(edges))
		assert.Equal(t, "CONNECTS", edges[0]["type"])
		assert.Equal(t, "communicates_with", edges[0]["relationship"])

		// UpdateEdge
		err = graph.UpdateEdge(ctx, "Agent", "agent-source", "Agent", "agent-target", "CONNECTS", map[string]interface{}{
			"relationship": "depends_on",
			"weight":       10,
		})
		assert.NoError(t, err)

		// Verify update
		edges, err = graph.GetEdges(ctx, "Agent", "agent-source")
		assert.NoError(t, err)
		assert.Equal(t, "depends_on", edges[0]["relationship"])
		assert.Equal(t, 10, edges[0]["weight"])

		// DeleteEdge
		err = graph.DeleteEdge(ctx, "Agent", "agent-source", "Agent", "agent-target", "CONNECTS")
		assert.NoError(t, err)

		// Verify deletion
		edges, err = graph.GetEdges(ctx, "Agent", "agent-source")
		assert.NoError(t, err)
		assert.Equal(t, 0, len(edges))

		// Cleanup
		graph.DeleteNode(ctx, "Agent", "agent-source")
		graph.DeleteNode(ctx, "Agent", "agent-target")
		graph.DeleteNode(ctx, "Agent", "agent-2")
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := graph.GetStats()
		assert.Equal(t, "neo4j", stats["implementation"])
	})
}

// TestNeo4jGraph_ErrorHandling tests error scenarios
func TestNeo4jGraph_ErrorHandling(t *testing.T) {
	logger := logging.NewNoOpLogger()
	config := GraphConfig{
		Backend:       GraphBackendNeo4j,
		Neo4jURL:      "bolt://nonexistent:7687",
		Neo4jUser:     "neo4j",
		Neo4jPassword: "wrong",
	}

	ctx := context.Background()
	_, err := NewNeo4jGraph(ctx, config, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to Neo4j")
}

// TestGraphFactory_Neo4j tests the factory creation of Neo4j graphs
func TestGraphFactory_Neo4j(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := logging.NewNoOpLogger()
	factory := NewGraphFactory(logger)

	config := GraphConfig{
		Backend:       GraphBackendNeo4j,
		Neo4jURL:      "bolt://localhost:7687",
		Neo4jUser:     "neo4j",
		Neo4jPassword: "orchestrator123",
	}

	graph, err := factory.CreateGraph(config)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
	}

	assert.NotNil(t, graph)

	// Test basic operations
	ctx := context.Background()
	err = graph.AddNode(ctx, "TestNode", "test-1", map[string]interface{}{
		"name": "test",
	})
	assert.NoError(t, err)

	// Cleanup
	if neo4jGraph, ok := graph.(*Neo4jGraph); ok {
		defer neo4jGraph.Close(ctx)
		defer graph.DeleteNode(ctx, "TestNode", "test-1")
	}
}
