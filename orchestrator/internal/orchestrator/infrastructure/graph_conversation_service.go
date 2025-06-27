package infrastructure

import (
	"context"

	"github.com/ztdp/orchestrator/internal/graph"
	orchestratorDomain "github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// GraphConversationService implements ConversationService using the graph backend
type GraphConversationService struct {
	graph graph.Graph
}

// NewGraphConversationService creates a new GraphConversationService
func NewGraphConversationService(graph graph.Graph) *GraphConversationService {
	return &GraphConversationService{
		graph: graph,
	}
}

// StoreInteraction stores a user interaction in the graph
func (gcs *GraphConversationService) StoreInteraction(ctx context.Context, userRequest string, analysis *orchestratorDomain.Analysis, decision *orchestratorDomain.Decision) error {
	// This would normally store the interaction in the graph for learning
	// For now, return success to satisfy the interface
	return nil
}

// GetConversationHistory retrieves conversation history for a session
func (gcs *GraphConversationService) GetConversationHistory(ctx context.Context, sessionID string) ([]string, error) {
	// This would normally query the graph for conversation history
	// For now, return empty history
	return []string{}, nil
}

// CreateSession creates a new conversation session
func (gcs *GraphConversationService) CreateSession(ctx context.Context) (string, error) {
	// This would normally create a session in the graph
	// For now, return a mock session ID
	return "session-123", nil
}
