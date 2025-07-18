package application

import (
	"os"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/stretchr/testify/require"
)

// TestHelpers provides reusable test infrastructure for all application domain tests
type TestHelpers struct {
	Graph      *graph.GlobalGraph
	AIProvider ai.AIProvider
	Registry   agentRegistry.AgentRegistry
	EventBus   *events.EventBus
}

// CreateTestHelpers creates a standardized test environment for application domain testing
// This centralizes all test setup logic to avoid duplication across test files
func CreateTestHelpers(t *testing.T) *TestHelpers {
	t.Helper()

	// Create in-memory graph for testing
	g := graph.NewGlobalGraph(graph.NewMemoryGraph())
	require.NotNil(t, g)

	// Create real AI provider (skip test if not available)
	aiProvider := createRealAIProvider(t)

	// Create test registry and event bus
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(events.NewMemoryTransport(), false)

	return &TestHelpers{
		Graph:      g,
		AIProvider: aiProvider,
		Registry:   registry,
		EventBus:   eventBus,
	}
}

// createRealAIProvider creates a real OpenAI provider for testing
// Skips the test if no API key is available
func createRealAIProvider(t *testing.T) ai.AIProvider {
	t.Helper()

	// Get API key from environment, same as main.go does
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping test that requires real AI")
	}

	provider, err := ai.NewOpenAIProvider(ai.DefaultOpenAIConfig(), apiKey)
	if err != nil || provider == nil {
		t.Skipf("Failed to create AI provider: %v", err)
	}
	return provider
}

// CreateTestApplicationService creates a test application service with real AI
func (h *TestHelpers) CreateTestApplicationService(t *testing.T) *Service {
	t.Helper()
	return NewService(h.Graph, h.AIProvider)
}

// CreateTestApplication creates a test application in the graph
func (h *TestHelpers) CreateTestApplication(t *testing.T, name string) *contracts.ApplicationContract {
	t.Helper()

	app := &contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  name,
			Owner: "test-team",
		},
		Spec: contracts.ApplicationSpec{
			Description: "Test application",
			Tags:        []string{"test"},
		},
	}

	// Use the graph's AddNode method with resolved contract
	node, err := graph.ResolveContract(*app)
	require.NoError(t, err)
	h.Graph.AddNode(node)

	return app
}

// CleanupTestData removes all test data from the graph
func (h *TestHelpers) CleanupTestData(t *testing.T) {
	t.Helper()
	// For in-memory graph, this is automatically cleaned up
	// For persistent storage, we'd implement cleanup logic here
}
