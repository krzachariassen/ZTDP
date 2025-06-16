package handlers

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/deployments"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

var (
	GlobalGraph   *graph.GlobalGraph
	GlobalV3Agent *ai.V3Agent // Using V3 Agent - the ultra simple one!
)

// InitializeGlobalV3Agent initializes the global V3 AI agent with pure orchestrator design
// This should be called once during application startup in main.go
func InitializeGlobalV3Agent() error {
	// Create AI provider the simple way!
	config := ai.DefaultOpenAIConfig()
	apiKey := os.Getenv("OPENAI_API_KEY")

	provider, err := ai.NewOpenAIProvider(config, apiKey)
	if err != nil {
		return err
	}

	// Initialize event infrastructure for agent-to-agent communication
	eventBus := events.NewEventBus(nil, false) // In-memory for now
	agentRegistry := agents.NewInMemoryAgentRegistry()

	// === AUTO-REGISTER DOMAIN AGENTS ===

	// 1. Register PolicyAgent for policy evaluation
	// Use the same backend as GlobalGraph to ensure data consistency
	if GlobalGraph == nil {
		return fmt.Errorf("GlobalGraph must be initialized before agents")
	}

	backend := GlobalGraph.Backend // Access the backend field directly
	graphStore := graph.NewGraphStore(backend)

	// Create EventBusAdapter for agent compatibility
	eventBusAdapter := &EventBusAdapter{eventBus}

	policyAgent, err := policies.NewPolicyAgent(graphStore, GlobalGraph, nil, "api", eventBusAdapter, agentRegistry)
	if err != nil {
		log.Printf("⚠️ Failed to create and register PolicyAgent: %v", err)
	} else {
		log.Printf("✅ PolicyAgent auto-registered successfully")
		
		// Subscribe PolicyAgent to its specific routing keys
		capabilities := policyAgent.GetCapabilities()
		for _, capability := range capabilities {
			for _, routingKey := range capability.RoutingKeys {
				eventBus.SubscribeToRoutingKey(routingKey, func(event events.Event) error {
					response, err := policyAgent.ProcessEvent(context.Background(), &event)
					if err != nil {
						log.Printf("⚠️ PolicyAgent failed to process event: %v", err)
					} else if response != nil {
						// Emit the response back to the event bus
						eventBus.Emit(response.Type, response.Source, response.Subject, response.Payload)
					}
					return nil
				})
				log.Printf("✅ PolicyAgent subscribed to routing key: %s", routingKey)
			}
		}
	}

	// 2. Register DeploymentAgent for deployment orchestration
	_, err = deployments.NewDeploymentAgent(GlobalGraph, provider, "api", eventBus, agentRegistry)
	if err != nil {
		log.Printf("⚠️ Failed to create and register DeploymentAgent: %v", err)
	} else {
		log.Printf("✅ DeploymentAgent auto-registered successfully")
	}

	// Create the V3 Agent with pure orchestrator design (no domain service dependencies)
	GlobalV3Agent = ai.NewV3Agent(
		provider,
		GlobalGraph,
		eventBus,
		agentRegistry,
	)

	return nil
}

// EventBusAdapter adapts events.EventBus to the agent EventBus interfaces
type EventBusAdapter struct {
	eventBus *events.EventBus
}

func (e *EventBusAdapter) Emit(eventType string, data map[string]interface{}) error {
	// Convert to events.EventType and call the underlying event bus
	return e.eventBus.Emit(events.EventTypeNotify, eventType, "api", data)
}

// GetGlobalV3Agent returns the initialized global V3 AI agent
// Returns nil if the agent hasn't been initialized
func GetGlobalV3Agent() *ai.V3Agent {
	return GlobalV3Agent
}
