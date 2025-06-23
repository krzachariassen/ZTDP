package graph

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EmbeddedGraph is a simple in-memory graph implementation
// This serves as a lightweight alternative to Neo4j for development and testing
type EmbeddedGraph struct {
	mu           sync.RWMutex
	nodes        map[string]*Node    // nodeType:nodeID -> Node
	edges        map[string]*Edge    // edgeID -> Edge
	nodesByType  map[string][]string // nodeType -> []nodeID
	nodesByLabel map[string][]string // label -> []nodeKey
	logger       Logger
}

// Node represents a graph node with properties
type Node struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	Labels     []string               `json:"labels"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// Edge represents a relationship between two nodes
type Edge struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	FromNode   string                 `json:"from_node"` // nodeType:nodeID
	ToNode     string                 `json:"to_node"`   // nodeType:nodeID
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Logger interface for graph operations
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewEmbeddedGraph creates a new embedded graph instance
func NewEmbeddedGraph(logger Logger) *EmbeddedGraph {
	return &EmbeddedGraph{
		nodes:        make(map[string]*Node),
		edges:        make(map[string]*Edge),
		nodesByType:  make(map[string][]string),
		nodesByLabel: make(map[string][]string),
		logger:       logger,
	}
}

// AddNode adds a node to the graph
func (g *EmbeddedGraph) AddNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if nodeType == "" || nodeID == "" {
		return fmt.Errorf("node type and ID cannot be empty")
	}

	nodeKey := g.nodeKey(nodeType, nodeID)

	// Check if node already exists
	if _, exists := g.nodes[nodeKey]; exists {
		return fmt.Errorf("node %s already exists", nodeKey)
	}

	// Create node
	node := &Node{
		Type:       nodeType,
		ID:         nodeID,
		Properties: make(map[string]interface{}),
		Labels:     []string{nodeType}, // Default label is the node type
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Copy properties
	for k, v := range properties {
		node.Properties[k] = v
	}

	// Store node
	g.nodes[nodeKey] = node

	// Update type index
	g.nodesByType[nodeType] = append(g.nodesByType[nodeType], nodeID)

	// Update label indices
	for _, label := range node.Labels {
		g.nodesByLabel[label] = append(g.nodesByLabel[label], nodeKey)
	}

	if g.logger != nil {
		g.logger.Debug("Added node to graph", "type", nodeType, "id", nodeID, "properties", properties)
	}

	return nil
}

// UpdateNode updates an existing node's properties
func (g *EmbeddedGraph) UpdateNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	nodeKey := g.nodeKey(nodeType, nodeID)

	node, exists := g.nodes[nodeKey]
	if !exists {
		return fmt.Errorf("node %s not found", nodeKey)
	}

	// Update properties
	for k, v := range properties {
		node.Properties[k] = v
	}

	node.UpdatedAt = time.Now()

	if g.logger != nil {
		g.logger.Debug("Updated node in graph", "type", nodeType, "id", nodeID, "properties", properties)
	}

	return nil
}

// GetNode retrieves a node by type and ID
func (g *EmbeddedGraph) GetNode(ctx context.Context, nodeType, nodeID string) (map[string]interface{}, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodeKey := g.nodeKey(nodeType, nodeID)

	node, exists := g.nodes[nodeKey]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeKey)
	}

	// Convert node to map format
	result := make(map[string]interface{})
	result["type"] = node.Type
	result["id"] = node.ID
	result["created_at"] = node.CreatedAt
	result["updated_at"] = node.UpdatedAt

	// Add all properties
	for k, v := range node.Properties {
		result[k] = v
	}

	return result, nil
}

// QueryNodes queries nodes by type and filters
func (g *EmbeddedGraph) QueryNodes(ctx context.Context, nodeType string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var results []map[string]interface{}

	// Get all nodes of the specified type
	nodeIDs, exists := g.nodesByType[nodeType]
	if !exists {
		return results, nil // No nodes of this type
	}

	// Filter nodes
	for _, nodeID := range nodeIDs {
		nodeKey := g.nodeKey(nodeType, nodeID)
		node, exists := g.nodes[nodeKey]
		if !exists {
			continue
		}

		// Apply filters
		if g.matchesFilters(node, filters) {
			// Convert to map format
			result := make(map[string]interface{})
			result["type"] = node.Type
			result["id"] = node.ID
			result["created_at"] = node.CreatedAt
			result["updated_at"] = node.UpdatedAt

			// Add all properties
			for k, v := range node.Properties {
				result[k] = v
			}

			results = append(results, result)
		}
	}

	if g.logger != nil {
		g.logger.Debug("Queried nodes", "type", nodeType, "filters", filters, "count", len(results))
	}

	return results, nil
}

// DeleteNode removes a node from the graph
func (g *EmbeddedGraph) DeleteNode(ctx context.Context, nodeType, nodeID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	nodeKey := g.nodeKey(nodeType, nodeID)

	node, exists := g.nodes[nodeKey]
	if !exists {
		return fmt.Errorf("node %s not found", nodeKey)
	}

	// Remove from nodes map
	delete(g.nodes, nodeKey)

	// Remove from type index
	if nodeList, exists := g.nodesByType[nodeType]; exists {
		g.nodesByType[nodeType] = g.removeFromSlice(nodeList, nodeID)
	}

	// Remove from label indices
	for _, label := range node.Labels {
		if labelList, exists := g.nodesByLabel[label]; exists {
			g.nodesByLabel[label] = g.removeFromSlice(labelList, nodeKey)
		}
	}

	// TODO: Remove related edges

	if g.logger != nil {
		g.logger.Debug("Deleted node from graph", "type", nodeType, "id", nodeID)
	}

	return nil
}

// Close closes the graph (no-op for embedded graph)
func (g *EmbeddedGraph) Close() error {
	if g.logger != nil {
		g.logger.Info("Closing embedded graph")
	}
	return nil
}

// GetStats returns statistics about the graph
func (g *EmbeddedGraph) GetStats() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := map[string]interface{}{
		"total_nodes":    len(g.nodes),
		"total_edges":    len(g.edges),
		"nodes_by_type":  make(map[string]int),
		"implementation": "embedded",
	}

	// Count nodes by type
	nodesByType := stats["nodes_by_type"].(map[string]int)
	for nodeType, nodeIDs := range g.nodesByType {
		nodesByType[nodeType] = len(nodeIDs)
	}

	return stats
}

// Helper methods

func (g *EmbeddedGraph) nodeKey(nodeType, nodeID string) string {
	return fmt.Sprintf("%s:%s", nodeType, nodeID)
}

func (g *EmbeddedGraph) matchesFilters(node *Node, filters map[string]interface{}) bool {
	if filters == nil {
		return true
	}

	for key, expectedValue := range filters {
		nodeValue, exists := node.Properties[key]
		if !exists {
			return false
		}

		// Simple equality check (can be enhanced for complex queries)
		if nodeValue != expectedValue {
			return false
		}
	}

	return true
}

func (g *EmbeddedGraph) removeFromSlice(slice []string, item string) []string {
	for i, val := range slice {
		if val == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// Future enhancement: Add relationship/edge operations
// AddEdge, QueryRelationships, TraverseGraph, etc.
