package graph

import (
	"context"
	"fmt"
)

// Graph defines the interface for graph operations
type Graph interface {
	// Node operations
	AddNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error
	GetNode(ctx context.Context, nodeType, nodeID string) (map[string]interface{}, error)
	UpdateNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error
	DeleteNode(ctx context.Context, nodeType, nodeID string) error
	QueryNodes(ctx context.Context, nodeType string, filters map[string]interface{}) ([]map[string]interface{}, error)

	// Graph operations
	GetStats() map[string]interface{}
}

// GraphConfig defines configuration for graph backends
type GraphConfig struct {
	Backend string `json:"backend"`
	// Neo4j specific config
	Neo4jURL      string `json:"neo4j_url,omitempty"`
	Neo4jUser     string `json:"neo4j_user,omitempty"`
	Neo4jPassword string `json:"neo4j_password,omitempty"`
}

// Graph backend types
const (
	GraphBackendEmbedded = "embedded"
	GraphBackendNeo4j    = "neo4j"
)

// GraphFactory creates graph instances
type GraphFactory struct {
	logger Logger
}

// NewGraphFactory creates a new graph factory
func NewGraphFactory(logger Logger) *GraphFactory {
	return &GraphFactory{logger: logger}
}

// CreateGraph creates a graph instance based on configuration
func (f *GraphFactory) CreateGraph(config GraphConfig) (Graph, error) {
	switch config.Backend {
	case GraphBackendEmbedded:
		return NewEmbeddedGraph(f.logger), nil
	case GraphBackendNeo4j:
		ctx := context.Background()
		return NewNeo4jGraph(ctx, config, f.logger)
	default:
		return nil, fmt.Errorf("unsupported graph backend: %s", config.Backend)
	}
}
