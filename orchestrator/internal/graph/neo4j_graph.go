package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jGraph implements Graph interface using Neo4j database
type Neo4jGraph struct {
	driver neo4j.DriverWithContext
	logger Logger
}

// NewNeo4jGraph creates a new Neo4j graph instance
func NewNeo4jGraph(ctx context.Context, config GraphConfig, logger Logger) (*Neo4jGraph, error) {
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

	graph := &Neo4jGraph{
		driver: driver,
		logger: logger,
	}

	if logger != nil {
		logger.Info("Connected to Neo4j", "url", config.Neo4jURL, "user", config.Neo4jUser)
	}

	return graph, nil
}

// Close closes the Neo4j connection
func (g *Neo4jGraph) Close(ctx context.Context) error {
	return g.driver.Close(ctx)
}

// AddNode adds a node to the Neo4j graph
func (g *Neo4jGraph) AddNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error {
	if nodeType == "" {
		return fmt.Errorf("node type cannot be empty")
	}
	if nodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Add type and id to properties
	props := make(map[string]interface{})
	for k, v := range properties {
		// Convert time.Time to string to avoid timezone issues
		if t, ok := v.(time.Time); ok {
			props[k] = t.UTC().Format(time.RFC3339)
		} else {
			props[k] = v
		}
	}
	props["type"] = nodeType
	props["id"] = nodeID
	props["created_at"] = time.Now().UTC().Format(time.RFC3339)
	props["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	cypher := fmt.Sprintf(`
		MERGE (n:%s {id: $id})
		SET n += $properties
		RETURN n
	`, nodeType)

	parameters := map[string]interface{}{
		"id":         nodeID,
		"properties": props,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, cypher, parameters)
		if err != nil {
			return nil, err
		}

		// Consume the result
		_, err = result.Consume(ctx)
		return nil, err
	})

	if err != nil {
		if g.logger != nil {
			g.logger.Error("Failed to add node to Neo4j", err, "type", nodeType, "id", nodeID)
		}
		return fmt.Errorf("failed to add node: %w", err)
	}

	if g.logger != nil {
		g.logger.Debug("Added node to Neo4j", "type", nodeType, "id", nodeID, "properties", properties)
	}

	return nil
}

// GetNode retrieves a node from Neo4j
func (g *Neo4jGraph) GetNode(ctx context.Context, nodeType, nodeID string) (map[string]interface{}, error) {
	if nodeType == "" {
		return nil, fmt.Errorf("node type cannot be empty")
	}
	if nodeID == "" {
		return nil, fmt.Errorf("node ID cannot be empty")
	}

	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	cypher := fmt.Sprintf("MATCH (n:%s {id: $id}) RETURN n", nodeType)
	parameters := map[string]interface{}{"id": nodeID}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, cypher, parameters)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			node := record.Values[0].(neo4j.Node)
			return node.Props, nil
		}

		return nil, fmt.Errorf("node not found")
	})

	if err != nil {
		return nil, err
	}

	return result.(map[string]interface{}), nil
}

// UpdateNode updates a node in Neo4j
func (g *Neo4jGraph) UpdateNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error {
	if nodeType == "" {
		return fmt.Errorf("node type cannot be empty")
	}
	if nodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Add updated_at timestamp
	props := make(map[string]interface{})
	for k, v := range properties {
		// Convert time.Time to string to avoid timezone issues
		if t, ok := v.(time.Time); ok {
			props[k] = t.UTC().Format(time.RFC3339)
		} else {
			props[k] = v
		}
	}
	props["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	cypher := fmt.Sprintf(`
		MATCH (n:%s {id: $id})
		SET n += $properties
		RETURN n
	`, nodeType)

	parameters := map[string]interface{}{
		"id":         nodeID,
		"properties": props,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, cypher, parameters)
		if err != nil {
			return nil, err
		}

		if !result.Next(ctx) {
			return nil, fmt.Errorf("node not found")
		}

		_, err = result.Consume(ctx)
		return nil, err
	})

	if err != nil {
		if g.logger != nil {
			g.logger.Error("Failed to update node in Neo4j", err, "type", nodeType, "id", nodeID)
		}
		return fmt.Errorf("failed to update node: %w", err)
	}

	if g.logger != nil {
		g.logger.Debug("Updated node in Neo4j", "type", nodeType, "id", nodeID, "properties", properties)
	}

	return nil
}

// DeleteNode deletes a node from Neo4j
func (g *Neo4jGraph) DeleteNode(ctx context.Context, nodeType, nodeID string) error {
	if nodeType == "" {
		return fmt.Errorf("node type cannot be empty")
	}
	if nodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	cypher := fmt.Sprintf("MATCH (n:%s {id: $id}) DETACH DELETE n", nodeType)
	parameters := map[string]interface{}{"id": nodeID}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, cypher, parameters)
		if err != nil {
			return nil, err
		}

		summary, err := result.Consume(ctx)
		if err != nil {
			return nil, err
		}

		if summary.Counters().NodesDeleted() == 0 {
			return nil, fmt.Errorf("node not found")
		}

		return nil, nil
	})

	if err != nil {
		if g.logger != nil {
			g.logger.Error("Failed to delete node from Neo4j", err, "type", nodeType, "id", nodeID)
		}
		return fmt.Errorf("failed to delete node: %w", err)
	}

	if g.logger != nil {
		g.logger.Debug("Deleted node from Neo4j", "type", nodeType, "id", nodeID)
	}

	return nil
}

// QueryNodes queries nodes from Neo4j
func (g *Neo4jGraph) QueryNodes(ctx context.Context, nodeType string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Build WHERE clause from filters
	whereClause := ""
	parameters := make(map[string]interface{})

	if len(filters) > 0 {
		var conditions []string
		for key, value := range filters {
			paramKey := "filter_" + key
			conditions = append(conditions, fmt.Sprintf("n.%s = $%s", key, paramKey))
			parameters[paramKey] = value
		}
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	cypher := fmt.Sprintf("MATCH (n:%s) %s RETURN n", nodeType, whereClause)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, cypher, parameters)
		if err != nil {
			return nil, err
		}

		var nodes []map[string]interface{}
		for result.Next(ctx) {
			record := result.Record()
			node := record.Values[0].(neo4j.Node)
			nodes = append(nodes, node.Props)
		}

		return nodes, result.Err()
	})

	if err != nil {
		if g.logger != nil {
			g.logger.Error("Failed to query nodes from Neo4j", err, "type", nodeType, "filters", filters)
		}
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}

	nodes := result.([]map[string]interface{})

	if g.logger != nil {
		g.logger.Debug("Queried nodes from Neo4j", "type", nodeType, "filters", filters, "count", len(nodes))
	}

	return nodes, nil
}

// GetStats returns statistics about the Neo4j graph
func (g *Neo4jGraph) GetStats() map[string]interface{} {
	ctx := context.Background()
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	stats := map[string]interface{}{
		"implementation":      "neo4j",
		"total_nodes":         0,
		"total_relationships": 0,
		"nodes_by_type":       make(map[string]int),
	}

	// Get total node count
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, "MATCH (n) RETURN count(n) as total", nil)
		if err != nil {
			return 0, err
		}

		if result.Next(ctx) {
			record := result.Record()
			total, _ := record.Values[0].(int64)
			return int(total), nil
		}
		return 0, nil
	})

	if err == nil {
		stats["total_nodes"] = result.(int)
	}

	// Get relationship count
	result, err = session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, "MATCH ()-[r]->() RETURN count(r) as total", nil)
		if err != nil {
			return 0, err
		}

		if result.Next(ctx) {
			record := result.Record()
			total, _ := record.Values[0].(int64)
			return int(total), nil
		}
		return 0, nil
	})

	if err == nil {
		stats["total_relationships"] = result.(int)
	}

	// Get nodes by label/type
	result, err = session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, "CALL db.labels() YIELD label CALL apoc.cypher.run('MATCH (n:' + label + ') RETURN count(n) as count', {}) YIELD value RETURN label, value.count as count", nil)
		if err != nil {
			// Fallback if APOC is not available
			result, err = tx.Run(ctx, "MATCH (n) RETURN labels(n)[0] as label, count(n) as count", nil)
			if err != nil {
				return nil, err
			}
		}

		nodesByType := make(map[string]int)
		for result.Next(ctx) {
			record := result.Record()
			if len(record.Values) >= 2 {
				label := record.Values[0].(string)
				count, _ := record.Values[1].(int64)
				nodesByType[label] = int(count)
			}
		}
		return nodesByType, result.Err()
	})

	if err == nil {
		stats["nodes_by_type"] = result.(map[string]int)
	}

	return stats
}
