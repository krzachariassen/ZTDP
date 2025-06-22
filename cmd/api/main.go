package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/krzachariassen/ZTDP/api/handlers"
	"github.com/krzachariassen/ZTDP/api/server"
	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/agents/orchestrator"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/application"
	"github.com/krzachariassen/ZTDP/internal/environment"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
	"github.com/krzachariassen/ZTDP/internal/policies"
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

	// Initialize Global Orchestrator at startup (Clean Architecture - Composition Root)
	logger.Info("🎯 Initializing Global Orchestrator...")

	// Create AI Provider
	logger.Info("🤖 Setting up AI Provider...")
	apiKey := os.Getenv("OPENAI_API_KEY")
	aiProvider, err := ai.NewOpenAIProvider(ai.DefaultOpenAIConfig(), apiKey)
	if err != nil || aiProvider == nil {
		logger.Warn("⚠️ AI Provider initialization failed: %v - AI features will be unavailable", err)
		// Continue without AI provider for now
	} else {
		logger.Info("✅ AI Provider initialized successfully")
	}

	// Create Agent Registry
	logger.Info("📋 Setting up Agent Registry...")
	agentRegistry := agentRegistry.NewInMemoryAgentRegistry()
	logger.Info("✅ Agent Registry initialized successfully")

	// Get the global event bus that was initialized earlier
	eventBus := events.GlobalEventBus

	// Create Orchestrator with all dependencies
	logger.Info("🎯 Creating Orchestrator...")
	orchestrator := orchestrator.NewOrchestrator(
		aiProvider,
		handlers.GlobalGraph,
		eventBus,
		agentRegistry,
	)
	logger.Info("✅ Global Orchestrator created successfully")

	// Inject orchestrator into handlers (Dependency Injection)
	handlers.SetupGlobalOrchestrator(orchestrator)

	// Initialize domain agents (environment-agnostic)
	logger.Info("🤖 Initializing domain agents...")

	// Initialize Application Agent
	logger.Info("📱 Creating Application Agent...")
	applicationAgent, err := application.NewApplicationAgent(
		handlers.GlobalGraph,
		aiProvider,
		eventBus,
		agentRegistry,
	)
	if err != nil {
		log.Fatalf("❌ Failed to create application agent: %v", err)
	}
	logger.Info("✅ Application Agent created successfully")

	// Initialize Environment Agent
	logger.Info("🚀 Creating Environment Agent...")
	deploymentAgent, err := environment.NewEnvironmentAgent(
		handlers.GlobalGraph,
		aiProvider,
		eventBus,
		agentRegistry,
	)
	if err != nil {
		log.Fatalf("❌ Failed to create Environment agent: %v", err)
	}
	logger.Info("✅ Environment Agent created successfully")

	// Initialize Policy Agent (with correct signature)
	logger.Info("🛡️ Creating Policy Agent...")
	policyAgent, err := policies.NewPolicyAgent(
		nil, // graphStore - using nil for now, will use global graph
		handlers.GlobalGraph,
		nil, // policyStore - using nil for default store
		eventBus,
		agentRegistry,
	)
	if err != nil {
		log.Fatalf("❌ Failed to create policy agent: %v", err)
	}
	logger.Info("✅ Policy Agent created successfully")

	// Start all agents
	logger.Info("▶️ Starting domain agents...")
	ctx := context.Background()

	if err := applicationAgent.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start application agent: %v", err)
	}
	logger.Info("✅ Application Agent started")

	if err := deploymentAgent.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start deployment agent: %v", err)
	}
	logger.Info("✅ Deployment Agent started")

	if err := policyAgent.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start policy agent: %v", err)
	}
	logger.Info("✅ Policy Agent started")

	logger.Info("🎯 All domain agents initialized and started successfully")

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
