package handlers

import (
	"sync"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/agents/orchestrator"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

var (
	globalOrchestrator *orchestrator.Orchestrator
	orchestratorOnce   sync.Once
)

// GetGlobalOrchestrator returns the global orchestrator instance
// This replaces the old GetGlobalV3Agent function
func GetGlobalOrchestrator() *orchestrator.Orchestrator {
	orchestratorOnce.Do(func() {
		// Initialize the orchestrator with real providers
		aiProvider, err := ai.NewOpenAIProvider(ai.DefaultOpenAIConfig(), "")
		if err != nil || aiProvider == nil {
			// Fallback to nil - handlers should check for this
			return
		}

		// Initialize graph
		backend := graph.NewMemoryGraph()
		globalGraph := graph.NewGlobalGraph(backend)

		// Initialize event bus
		eventBus := events.NewEventBus(nil, false)

		// Use the proper agent registry from internal/agentRegistry
		agentReg := agentRegistry.NewInMemoryAgentRegistry()

		globalOrchestrator = orchestrator.NewOrchestrator(
			aiProvider,
			globalGraph,
			eventBus,
			agentReg,
		)
	})

	return globalOrchestrator
}

// GetGlobalV3Agent is deprecated - use GetGlobalOrchestrator instead
// This is kept for backward compatibility during migration
func GetGlobalV3Agent() *orchestrator.Orchestrator {
	return GetGlobalOrchestrator()
}
