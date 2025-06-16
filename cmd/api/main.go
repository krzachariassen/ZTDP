package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/ai"
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

	// Initialize AI infrastructure
	logger.Info("🤖 Initializing AI platform...")

	// Initialize global V3 AI agent as pure orchestrator (no domain service dependencies)
	if err := handlers.InitializeGlobalV3Agent(); err != nil {
		logger.Warn("⚠️  Failed to initialize V3 AI agent: %v. AI features will be limited.", err)
	} else {
		logger.Info("🤖 AI platform agent initialized successfully as pure orchestrator")
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

// createAIProvider creates the AI provider if configured
func createAIProvider() (ai.AIProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	config := ai.DefaultOpenAIConfig()
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		config.Model = model
	}
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}

	return ai.NewOpenAIProvider(config, apiKey)
}
