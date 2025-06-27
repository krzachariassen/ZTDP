package graph

import (
	"context"

	"github.com/ztdp/orchestrator/internal/logging"
)

// Graph defines a simple interface for basic graph operations
type Graph interface {
	// Node operations - basic CRUD
	AddNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error
	GetNode(ctx context.Context, nodeType, nodeID string) (map[string]interface{}, error)
	UpdateNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error
	DeleteNode(ctx context.Context, nodeType, nodeID string) error
	QueryNodes(ctx context.Context, nodeType string, filters map[string]interface{}) ([]map[string]interface{}, error)

	// Edge operations - basic CRUD
	AddEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, edgeType string, properties map[string]interface{}) error
	GetEdges(ctx context.Context, nodeType, nodeID string) ([]map[string]interface{}, error)
	UpdateEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, edgeType string, properties map[string]interface{}) error
	DeleteEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, edgeType string) error

	// Utility
	GetStats() map[string]interface{}
	Close(ctx context.Context) error
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
	logger logging.Logger
}

// NewGraphFactory creates a new graph factory
func NewGraphFactory(logger logging.Logger) *GraphFactory {
	return &GraphFactory{logger: logger}
}

// CreateGraph creates a graph instance based on configuration
func (f *GraphFactory) CreateGraph(config GraphConfig) (Graph, error) {
	ctx := context.Background()
	return NewNeo4jGraph(ctx, config, f.logger)
}
