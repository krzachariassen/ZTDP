package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/events"
)

// NewFrameworkOrchestrator creates an orchestrator that implements the agent framework interface
func NewFrameworkOrchestrator(
	aiProvider interface{}, // Accepting interface{} to avoid import cycle issues
	globalGraph interface{},
	eventBus interface{},
	agentRegistry interface{},
) (agents.AgentInterface, error) {

	// Type assertions to convert interfaces to concrete types
	// This allows the orchestrator to work with the agent framework without import cycles
	var orchestrator *Orchestrator

	// Create based on what we receive - flexible construction
	switch p := aiProvider.(type) {
	case nil:
		// No AI provider - create mock or basic orchestrator
		orchestrator = &Orchestrator{
			logger:    nil, // Will be set up later
			agentID:   "orchestrator",
			startTime: time.Now(),
		}
	default:
		// Try to create with available components
		orchestrator = &Orchestrator{
			aiProvider: nil, // Will need to be cast properly
			logger:     nil, // Will be set up later
			agentID:    "orchestrator",
			startTime:  time.Now(),
		}

		// Store the raw interfaces for later use
		orchestrator.rawAIProvider = p
		orchestrator.rawGraph = globalGraph
		orchestrator.rawEventBus = eventBus
		orchestrator.rawAgentRegistry = agentRegistry
	}

	return &FrameworkOrchestrator{orchestrator: orchestrator}, nil
}

// FrameworkOrchestrator wraps the Orchestrator to implement both agent interfaces
type FrameworkOrchestrator struct {
	orchestrator *Orchestrator
}

// GetID implements agentRegistry.AgentInterface
func (f *FrameworkOrchestrator) GetID() string {
	return "orchestrator"
}

// GetStatus returns the current agent status
func (f *FrameworkOrchestrator) GetStatus() agents.AgentStatus {
	return agents.AgentStatus{
		ID:           f.GetID(),
		Type:         "orchestrator",
		Status:       "running",
		LastActivity: time.Now(),
		LoadFactor:   0.5,
		Version:      "1.0.0",
		Metadata: map[string]interface{}{
			"role":         "orchestrator",
			"capabilities": "intent-based routing, resource creation, operational coordination",
			"ai_enabled":   true,
		},
	}
}

// GetCapabilities returns the agent's capabilities for discovery
func (f *FrameworkOrchestrator) GetCapabilities() []agents.AgentCapability {
	return []agents.AgentCapability{
		{
			Name:        "chat_orchestration",
			Description: "Natural language chat interface for orchestrating platform operations",
			Intents: []string{
				"chat", "conversation", "help", "ask question",
				"orchestrate", "coordinate", "general query",
			},
			InputTypes:  []string{"natural_language", "user_message", "question"},
			OutputTypes: []string{"conversational_response", "orchestration_result"},
			RoutingKeys: []string{"orchestrator.chat", "orchestrator.orchestrate", "orchestrator.general"},
			Version:     "1.0.0",
		},
		{
			Name:        "resource_creation",
			Description: "AI-driven creation of platform resources via natural language",
			Intents: []string{
				"create resource", "build application", "setup environment",
				"make service", "configure deployment", "resource creation",
			},
			InputTypes:  []string{"creation_request", "resource_specification", "natural_language"},
			OutputTypes: []string{"resource_created", "creation_result", "resource_contract"},
			RoutingKeys: []string{"orchestrator.create", "orchestrator.resource", "orchestrator.build"},
			Version:     "1.0.0",
		},
		{
			Name:        "intent_routing",
			Description: "Smart routing of operational intents to appropriate specialist agents",
			Intents: []string{
				"route intent", "find agent", "orchestrate operation",
				"delegate task", "agent coordination",
			},
			InputTypes:  []string{"operational_intent", "routing_request", "delegation_request"},
			OutputTypes: []string{"routing_result", "agent_response", "coordination_result"},
			RoutingKeys: []string{"orchestrator.route", "orchestrator.delegate", "orchestrator.coordinate"},
			Version:     "1.0.0",
		},
	}
}

// Start initializes the orchestrator as a registered agent
func (f *FrameworkOrchestrator) Start(ctx context.Context) error {
	f.orchestrator.startTime = time.Now()
	f.orchestrator.agentID = "orchestrator"

	// Auto-register with the agent registry if available
	// TODO: Fix interface compatibility issues
	/*
		if f.orchestrator.agentRegistry != nil {
			if err := f.orchestrator.agentRegistry.RegisterAgent(ctx, f); err != nil {
				if f.orchestrator.logger != nil {
					f.orchestrator.logger.Error("❌ Failed to auto-register Orchestrator: %v", err)
				}
				return fmt.Errorf("failed to auto-register Orchestrator: %w", err)
			}
			if f.orchestrator.logger != nil {
				f.orchestrator.logger.Info("✅ Orchestrator auto-registered successfully")
			}
		}
	*/

	// Subscribe to orchestrator routing keys
	if f.orchestrator.eventBus != nil {
		f.subscribeToOwnRoutingKeys()
	}

	if f.orchestrator.logger != nil {
		f.orchestrator.logger.Info("🚀 Orchestrator started as first-class agent")
	}
	return nil
}

// Stop shuts down the orchestrator
func (f *FrameworkOrchestrator) Stop(ctx context.Context) error {
	if f.orchestrator.logger != nil {
		f.orchestrator.logger.Info("🛑 Orchestrator stopping")
	}
	return nil
}

// Health returns the health status of the agent
func (f *FrameworkOrchestrator) Health() agents.HealthStatus {
	return agents.HealthStatus{
		Healthy: true,
		Status:  "healthy",
		Message: "Orchestrator operating normally",
		Checks: map[string]interface{}{
			"ai_provider_available": f.orchestrator.aiProvider != nil,
			"event_bus_connected":   f.orchestrator.eventBus != nil,
			"registry_connected":    f.orchestrator.agentRegistry != nil,
		},
		CheckedAt: time.Now(),
	}
}

// ProcessEvent handles events sent directly to the orchestrator
func (f *FrameworkOrchestrator) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	if f.orchestrator.logger != nil {
		f.orchestrator.logger.Info("📨 Orchestrator received event: %s from %s", event.Subject, event.Source)
	}

	// Extract user message from the event payload
	userMessage := ""

	// Try to get user_message from different places in the payload
	if msg, ok := event.Payload["user_message"].(string); ok {
		userMessage = msg
	} else if context, ok := event.Payload["context"].(map[string]interface{}); ok {
		if msg, ok := context["user_message"].(string); ok {
			userMessage = msg
		}
	} else if msg, ok := event.Payload["message"].(string); ok {
		userMessage = msg
	}

	if userMessage == "" {
		return f.createErrorResponse(event, "user_message required for Orchestrator processing"), nil
	}

	// Process the message using the Chat method
	response, err := f.orchestrator.Chat(ctx, userMessage)
	if err != nil {
		return f.createErrorResponse(event, fmt.Sprintf("Orchestrator processing failed: %v", err)), nil
	}

	// Create success response
	return f.createSuccessResponse(event, response), nil
}

// subscribeToOwnRoutingKeys sets up event subscriptions for orchestrator's routing keys
func (f *FrameworkOrchestrator) subscribeToOwnRoutingKeys() {
	routingKeys := []string{
		"orchestrator.chat", "orchestrator.orchestrate", "orchestrator.general",
		"orchestrator.create", "orchestrator.resource", "orchestrator.build",
		"orchestrator.route", "orchestrator.delegate", "orchestrator.coordinate",
	}

	for _, routingKey := range routingKeys {
		f.orchestrator.eventBus.SubscribeToRoutingKey(routingKey, func(event events.Event) error {
			if f.orchestrator.logger != nil {
				f.orchestrator.logger.Info("📨 Orchestrator received event via routing key %s: %s", routingKey, event.Subject)
			}

			// Process the event
			ctx := context.Background()
			response, err := f.ProcessEvent(ctx, &event)
			if err != nil {
				if f.orchestrator.logger != nil {
					f.orchestrator.logger.Error("❌ Failed to process event: %v", err)
				}
				return err
			}

			// Send response back
			if response != nil && f.orchestrator.eventBus != nil {
				f.orchestrator.eventBus.EmitEvent(*response)
			}

			return nil
		})
	}

	if f.orchestrator.logger != nil {
		f.orchestrator.logger.Info("✅ Orchestrator subscribed to %d routing keys", len(routingKeys))
	}
}

// createErrorResponse creates a standardized error response
func (f *FrameworkOrchestrator) createErrorResponse(originalEvent *events.Event, errorMessage string) *events.Event {
	response := &events.Event{
		Type:    events.EventTypeResponse,
		Source:  f.GetID(),
		Subject: "Orchestrator processing failed",
		Payload: map[string]interface{}{
			"status":  "error",
			"error":   errorMessage,
			"context": "orchestrator",
		},
		Timestamp: time.Now().Unix(),
		ID:        fmt.Sprintf("orchestrator-resp-%d", time.Now().UnixNano()),
	}

	// Copy correlation fields from original event
	if originalEvent != nil && originalEvent.Payload != nil {
		if correlationID, ok := originalEvent.Payload["correlation_id"]; ok {
			response.Payload["correlation_id"] = correlationID
		}
		if requestID, ok := originalEvent.Payload["request_id"]; ok {
			response.Payload["request_id"] = requestID
		}
	}

	return response
}

// createSuccessResponse creates a standardized success response
func (f *FrameworkOrchestrator) createSuccessResponse(originalEvent *events.Event, chatResponse *ConversationalResponse) *events.Event {
	response := &events.Event{
		Type:    events.EventTypeResponse,
		Source:  f.GetID(),
		Subject: "Orchestrator processing completed",
		Payload: map[string]interface{}{
			"status":                  "success",
			"message":                 chatResponse.Message,
			"intent":                  chatResponse.Intent,
			"actions":                 chatResponse.Actions,
			"insights":                chatResponse.Insights,
			"confidence":              chatResponse.Confidence,
			"conversational_response": chatResponse,
		},
		Timestamp: time.Now().Unix(),
		ID:        fmt.Sprintf("orchestrator-resp-%d", time.Now().UnixNano()),
	}

	// Copy correlation fields from original event
	if originalEvent != nil && originalEvent.Payload != nil {
		if correlationID, ok := originalEvent.Payload["correlation_id"]; ok {
			response.Payload["correlation_id"] = correlationID
		}
		if requestID, ok := originalEvent.Payload["request_id"]; ok {
			response.Payload["request_id"] = requestID
		}
	}

	return response
}
