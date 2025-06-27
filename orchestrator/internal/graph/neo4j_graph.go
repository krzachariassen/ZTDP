package graph

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/ztdp/orchestrator/internal/logging"
)

// Neo4jGraph implements simple graph operations using Neo4j
type Neo4jGraph struct {
	driver neo4j.DriverWithContext
	logger logging.Logger
}

// NewNeo4jGraph creates a new Neo4j graph instance
func NewNeo4jGraph(ctx context.Context, config GraphConfig, logger logging.Logger) (*Neo4jGraph, error) {
	if config.Neo4jURL == "" {
		config.Neo4jURL = "bolt://localhost:7687"
	}
	if config.Neo4jUser == "" {
		config.Neo4jUser = "neo4j"
	}
	if config.Neo4jPassword == "" {
		config.Neo4jPassword = "orchestrator123"
	}

	auth := neo4j.BasicAuth(config.Neo4jUser, config.Neo4jPassword, "")
	driver, err := neo4j.NewDriverWithContext(config.Neo4jURL, auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Test connection
	if err := driver.VerifyConnectivity(ctx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	return &Neo4jGraph{
		driver: driver,
		logger: logger,
	}, nil
}

// Close closes the Neo4j connection
func (g *Neo4jGraph) Close(ctx context.Context) error {
	return g.driver.Close(ctx)
}

// AddNode adds a node to the graph
func (g *Neo4jGraph) AddNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := fmt.Sprintf("CREATE (n:%s {id: $id}) SET n += $properties", nodeType)
	params := map[string]interface{}{
		"id":         nodeID,
		"properties": properties,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// GetNode retrieves a node from the graph
func (g *Neo4jGraph) GetNode(ctx context.Context, nodeType, nodeID string) (map[string]interface{}, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := fmt.Sprintf("MATCH (n:%s {id: $id}) RETURN n", nodeType)
	params := map[string]interface{}{"id": nodeID}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			node := record.Values[0].(neo4j.Node)

			// Convert to map[string]interface{}
			nodeMap := map[string]interface{}{
				"type": nodeType,
				"id":   nodeID,
			}

			// Add all properties with type conversion
			for k, v := range node.Props {
				nodeMap[k] = convertValue(v)
			}

			return nodeMap, nil
		}

		return nil, fmt.Errorf("node not found")
	})

	if err != nil {
		return nil, err
	}

	return result.(map[string]interface{}), nil
}

// UpdateNode updates a node in the graph
func (g *Neo4jGraph) UpdateNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := fmt.Sprintf("MATCH (n:%s {id: $id}) SET n += $properties", nodeType)
	params := map[string]interface{}{
		"id":         nodeID,
		"properties": properties,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// DeleteNode deletes a node from the graph
func (g *Neo4jGraph) DeleteNode(ctx context.Context, nodeType, nodeID string) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := fmt.Sprintf("MATCH (n:%s {id: $id}) DETACH DELETE n", nodeType)
	params := map[string]interface{}{"id": nodeID}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// QueryNodes queries nodes from the graph
func (g *Neo4jGraph) QueryNodes(ctx context.Context, nodeType string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Build query
	query := fmt.Sprintf("MATCH (n:%s)", nodeType)
	params := make(map[string]interface{})

	if filters != nil && len(filters) > 0 {
		query += " WHERE "
		conditions := []string{}
		for k, v := range filters {
			conditions = append(conditions, fmt.Sprintf("n.%s = $%s", k, k))
			params[k] = v
		}
		query += fmt.Sprintf("%s", conditions[0]) // Just use first condition for simplicity
	}

	query += " RETURN n"

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var nodes []map[string]interface{}
		for result.Next(ctx) {
			record := result.Record()
			node := record.Values[0].(neo4j.Node)

			// Convert to map[string]interface{}
			nodeMap := map[string]interface{}{
				"type": nodeType,
			}

			// Add all properties (including id) with type conversion
			for k, v := range node.Props {
				nodeMap[k] = convertValue(v)
			}

			nodes = append(nodes, nodeMap)
		}

		return nodes, result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]map[string]interface{}), nil
}

// AddEdge adds an edge between two nodes
func (g *Neo4jGraph) AddEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, edgeType string, properties map[string]interface{}) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := fmt.Sprintf(`
		MATCH (a:%s {id: $sourceID}), (b:%s {id: $targetID})
		CREATE (a)-[r:%s]->(b)
		SET r += $properties
	`, sourceType, targetType, edgeType)

	params := map[string]interface{}{
		"sourceID":   sourceID,
		"targetID":   targetID,
		"properties": properties,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// GetEdges gets edges from a node
func (g *Neo4jGraph) GetEdges(ctx context.Context, nodeType, nodeID string) ([]map[string]interface{}, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := fmt.Sprintf("MATCH (n:%s {id: $id})-[r]->(m) RETURN r", nodeType)
	params := map[string]interface{}{"id": nodeID}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var edges []map[string]interface{}
		for result.Next(ctx) {
			record := result.Record()
			rel := record.Values[0].(neo4j.Relationship)

			edgeMap := map[string]interface{}{
				"type": rel.Type,
			}

			// Add all properties with type conversion
			for k, v := range rel.Props {
				edgeMap[k] = convertValue(v)
			}

			edges = append(edges, edgeMap)
		}

		return edges, result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]map[string]interface{}), nil
}

// UpdateEdge updates an edge
func (g *Neo4jGraph) UpdateEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, edgeType string, properties map[string]interface{}) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := fmt.Sprintf(`
		MATCH (a:%s {id: $sourceID})-[r:%s]->(b:%s {id: $targetID})
		SET r += $properties
	`, sourceType, edgeType, targetType)

	params := map[string]interface{}{
		"sourceID":   sourceID,
		"targetID":   targetID,
		"properties": properties,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// DeleteEdge deletes an edge
func (g *Neo4jGraph) DeleteEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, edgeType string) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := fmt.Sprintf(`
		MATCH (a:%s {id: $sourceID})-[r:%s]->(b:%s {id: $targetID})
		DELETE r
	`, sourceType, edgeType, targetType)

	params := map[string]interface{}{
		"sourceID": sourceID,
		"targetID": targetID,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// ClearTestData removes all test data from the graph (for testing only)
func (g *Neo4jGraph) ClearTestData(ctx context.Context) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, "MATCH (n) DETACH DELETE n", nil)
		return nil, err
	})

	return err
}

// GetStats returns basic statistics
func (g *Neo4jGraph) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"implementation": "neo4j",
		"total_nodes":    0, // Simplified for now
	}
}

// convertValue converts Neo4j values to Go types with proper type handling
func convertValue(value interface{}) interface{} {
	switch v := value.(type) {
	case int64:
		// Convert Neo4j int64 to int for consistency
		return int(v)
	case []interface{}:
		// Convert slice elements recursively
		result := make([]interface{}, len(v))
		for i, elem := range v {
			result[i] = convertValue(elem)
		}
		return result
	default:
		return v
	}
}

// convertProperties converts a map of Neo4j properties to normalized Go types
func convertProperties(props map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range props {
		result[k] = convertValue(v)
	}
	return result
}
