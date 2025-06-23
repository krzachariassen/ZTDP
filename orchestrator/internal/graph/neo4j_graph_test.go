package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNeo4jGraph_Integration tests Neo4j graph operations
// This test requires a running Neo4j instance (use docker-compose up neo4j)
func TestNeo4jGraph_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := &MockLogger{}
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

	t.Run("AddNode", func(t *testing.T) {
		properties := map[string]interface{}{
			"name":         "test-agent",
			"capabilities": []string{"deploy", "test"},
			"status":       "active",
		}

		err := graph.AddNode(ctx, "Agent", "agent-1", properties)
		assert.NoError(t, err)
	})

	t.Run("GetNode", func(t *testing.T) {
		result, err := graph.GetNode(ctx, "Agent", "agent-1")
		assert.NoError(t, err)
		assert.Equal(t, "Agent", result["type"])
		assert.Equal(t, "agent-1", result["id"])
		assert.Equal(t, "test-agent", result["name"])
		assert.Equal(t, "active", result["status"])
	})

	t.Run("UpdateNode", func(t *testing.T) {
		err := graph.UpdateNode(ctx, "Agent", "agent-1", map[string]interface{}{
			"status":   "inactive",
			"endpoint": "http://localhost:8080",
		})
		assert.NoError(t, err)

		// Verify update
		result, err := graph.GetNode(ctx, "Agent", "agent-1")
		assert.NoError(t, err)
		assert.Equal(t, "inactive", result["status"])
		assert.Equal(t, "http://localhost:8080", result["endpoint"])
		assert.Equal(t, "test-agent", result["name"]) // Original property preserved
	})

	t.Run("QueryNodes", func(t *testing.T) {
		// Add another agent for querying
		err := graph.AddNode(ctx, "Agent", "agent-2", map[string]interface{}{
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
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := graph.GetStats()
		assert.Equal(t, "neo4j", stats["implementation"])
		assert.GreaterOrEqual(t, stats["total_nodes"].(int), 2)
	})

	t.Run("DeleteNode", func(t *testing.T) {
		err := graph.DeleteNode(ctx, "Agent", "agent-1")
		assert.NoError(t, err)

		// Verify deletion
		_, err = graph.GetNode(ctx, "Agent", "agent-1")
		assert.Error(t, err)
	})

	// Cleanup
	t.Cleanup(func() {
		graph.DeleteNode(ctx, "Agent", "agent-2")
	})
}

// TestNeo4jGraph_ErrorHandling tests error scenarios
func TestNeo4jGraph_ErrorHandling(t *testing.T) {
	logger := &MockLogger{}
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

	logger := &MockLogger{}
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

// BenchmarkNeo4jVsEmbedded compares performance
func BenchmarkNeo4jVsEmbedded(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark test")
	}

	ctx := context.Background()
	logger := &MockLogger{}

	// Setup Neo4j
	neo4jConfig := GraphConfig{
		Backend:       GraphBackendNeo4j,
		Neo4jURL:      "bolt://localhost:7687",
		Neo4jUser:     "neo4j",
		Neo4jPassword: "orchestrator123",
	}

	neo4jGraph, err := NewNeo4jGraph(ctx, neo4jConfig, logger)
	if err != nil {
		b.Skipf("Neo4j not available: %v", err)
	}
	defer neo4jGraph.Close(ctx)

	// Setup Embedded
	embeddedGraph := NewEmbeddedGraph(logger)

	b.Run("Neo4j_AddNode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nodeID := fmt.Sprintf("bench-node-%d", i)
			neo4jGraph.AddNode(ctx, "BenchNode", nodeID, map[string]interface{}{
				"name":  "benchmark",
				"index": i,
			})
		}
	})

	b.Run("Embedded_AddNode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nodeID := fmt.Sprintf("bench-node-%d", i)
			embeddedGraph.AddNode(ctx, "BenchNode", nodeID, map[string]interface{}{
				"name":  "benchmark",
				"index": i,
			})
		}
	})
}
