package handlers

import (
	"github.com/krzachariassen/ZTDP/internal/agents/orchestrator"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

var (
	GlobalGraph        *graph.GlobalGraph
	globalOrchestrator *orchestrator.Orchestrator
)

// SetupGlobalOrchestrator sets the global orchestrator instance (called from main.go)
func SetupGlobalOrchestrator(o *orchestrator.Orchestrator) {
	globalOrchestrator = o
}

// GetGlobalOrchestrator returns the global orchestrator instance
func GetGlobalOrchestrator() *orchestrator.Orchestrator {
	return globalOrchestrator
}
