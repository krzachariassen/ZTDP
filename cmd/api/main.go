package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

func main() {
	// Create a new, empty registry (policies will register themselves dynamically)
	policyRegistry := policies.NewPolicyRegistry()
	graph.SetPolicyRegistry(policyRegistry)

	var backend graph.GraphBackend
	switch os.Getenv("ZTDP_GRAPH_BACKEND") {
	case "redis":
		fmt.Println("⚙️  Using backend: Redis")
		backend = graph.NewRedisGraph(graph.RedisGraphConfig{})
	default:
		fmt.Println("⚙️  Using backend: Memory")
		backend = graph.NewMemoryGraph()
	}
	handlers.GlobalGraph = graph.NewGlobalGraph(backend)

	// Load persisted graph from backend (Redis)
	if err := handlers.GlobalGraph.Load(); err != nil {
		fmt.Println("No existing global graph found, starting fresh")
	}

	r := server.NewRouter()
	log.Println("Starting API on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
