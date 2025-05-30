package main

import (
	"log"
	"net/http"
	"os"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

func main() {
	// Initialize centralized logging system
	logLevel := logging.LevelInfo
	if os.Getenv("ZTDP_LOG_LEVEL") == "debug" {
		logLevel = logging.LevelDebug
	}
	logging.InitializeLogger("ztdp-api", logLevel)

	// Create real-time log sink for WebSocket broadcasting
	realtimeSink := logging.NewRealtimeLogSink()
	logging.GetLogger().AddSink(realtimeSink)

	logger := logging.GetLogger()
	logger.Info("🚀 Starting ZTDP API Server")

	// Configure event system
	var eventTransport events.EventTransport

	// Check if NATS is configured
	natsURL := os.Getenv("ZTDP_NATS_URL")
	if natsURL != "" {
		logger.Info("🔔 Using NATS event transport: %s", natsURL)
		natsConfig := events.DefaultNATSConfig()
		natsConfig.URL = natsURL

		var err error
		eventTransport, err = events.NewNATSTransport(natsConfig)
		if err != nil {
			logger.Warn("⚠️ Failed to connect to NATS, falling back to memory transport: %v", err)
			eventTransport = events.NewMemoryTransport()
		}
	} else {
		logger.Info("🔔 Using in-memory event transport")
		eventTransport = events.NewMemoryTransport()
	}

	// Initialize simple event system
	events.InitializeEventBus(eventTransport)
	logger.Info("🔔 Event system initialized")

	// Initialize log manager for real-time WebSocket streaming
	handlers.InitLogManager()
	logger.Info("📊 Log manager initialized")

	var backend graph.GraphBackend
	switch os.Getenv("ZTDP_GRAPH_BACKEND") {
	case "redis":
		logger.Info("⚙️  Using backend: Redis")
		backend = graph.NewRedisGraph(graph.RedisGraphConfig{})
	default:
		logger.Info("⚙️  Using backend: Memory")
		backend = graph.NewMemoryGraph()
	}
	handlers.GlobalGraph = graph.NewGlobalGraph(backend)

	// Load persisted graph from backend (Redis)
	if err := handlers.GlobalGraph.Load(); err != nil {
		logger.Info("No existing global graph found, starting fresh")
	}

	r := server.NewRouter()

	// Add logging middleware to router
	loggedRouter := logging.CreateHTTPLoggingMiddleware("api-server")(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("🌐 Starting API server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, loggedRouter))
}
