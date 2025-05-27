package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

func main() {
	// Configure event system
	var eventTransport events.EventTransport

	// Check if NATS is configured
	natsURL := os.Getenv("ZTDP_NATS_URL")
	if natsURL != "" {
		fmt.Println("🔔 Using NATS event transport:", natsURL)
		natsConfig := events.DefaultNATSConfig()
		natsConfig.URL = natsURL

		var err error
		eventTransport, err = events.NewNATSTransport(natsConfig)
		if err != nil {
			log.Printf("⚠️ Failed to connect to NATS, falling back to memory transport: %v", err)
			eventTransport = events.NewMemoryTransport()
		}
	} else {
		fmt.Println("🔔 Using in-memory event transport")
		eventTransport = events.NewMemoryTransport()
	}

	// Create the event bus
	eventBus := events.NewEventBus(eventTransport, true)

	// Create event services
	policyEvents := events.NewPolicyEventService(eventBus, "api-server")
	graphEvents := events.NewGraphEventService(eventBus, "api-server")

	// Create a validator that uses graph-based policies
	graphPolicyValidator := policies.NewGraphBasedPolicyValidator()

	// Set the policy validator for the graph package
	graph.SetPolicyValidator(graphPolicyValidator)

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

	// Set up handlers with event system
	handlers.SetupEventSystem(eventBus, policyEvents, graphEvents)

	// Wrap graph store with event emitter
	graphBackend := graph.NewGraphStore(backend)
	eventEmitter := graph.NewGraphEventEmitter(graphBackend, graphEvents)

	// Set up global handlers
	handlers.SetGraphEmitter(eventEmitter)

	r := server.NewRouter()
	log.Println("Starting API on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
