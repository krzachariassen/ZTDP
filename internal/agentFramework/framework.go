package agentFramework

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// AgentDependencies contains the dependencies needed to build an agent
type AgentDependencies struct {
	Registry agentRegistry.AgentRegistry
	EventBus *events.EventBus
}

// BaseAgent represents the framework agent that implements common patterns
type BaseAgent struct {
	id           string
	agentType    string
	capabilities []agentRegistry.AgentCapability
	eventHandler func(ctx context.Context, event *events.Event) (*events.Event, error)

	// Dependencies
	registry  agentRegistry.AgentRegistry
	eventBus  *events.EventBus
	logger    *logging.Logger
	startTime time.Time
}

// AgentBuilder provides a fluent interface for building agents
type AgentBuilder struct {
	id           string
	agentType    string
	capabilities []agentRegistry.AgentCapability
	eventHandler func(ctx context.Context, event *events.Event) (*events.Event, error)
}

// NewAgent creates a new agent builder
func NewAgent(id string) *AgentBuilder {
	return &AgentBuilder{
		id:        id,
		agentType: "framework", // default type
	}
}

// WithCapabilities sets the agent capabilities
func (b *AgentBuilder) WithCapabilities(capabilities []agentRegistry.AgentCapability) *AgentBuilder {
	b.capabilities = capabilities
	return b
}

// WithEventHandler sets the event handling function
func (b *AgentBuilder) WithEventHandler(handler func(ctx context.Context, event *events.Event) (*events.Event, error)) *AgentBuilder {
	b.eventHandler = handler
	return b
}

// WithType sets the agent type
func (b *AgentBuilder) WithType(agentType string) *AgentBuilder {
	b.agentType = agentType
	return b
}

// Build creates the agent with the specified dependencies
func (b *AgentBuilder) Build(deps AgentDependencies) (agentRegistry.AgentInterface, error) {
	agent := &BaseAgent{
		id:           b.id,
		agentType:    b.agentType,
		capabilities: b.capabilities,
		eventHandler: b.eventHandler,
		registry:     deps.Registry,
		eventBus:     deps.EventBus,
		logger:       logging.GetLogger().ForComponent(b.id),
		startTime:    time.Now(),
	}

	// Auto-register the agent
	ctx := context.Background()
	if err := deps.Registry.RegisterAgent(ctx, agent); err != nil {
		return nil, err
	}

	// Auto-subscribe to routing keys based on capabilities
	if err := agent.subscribeToCapabilities(); err != nil {
		return nil, err
	}

	return agent, nil
}

// Implement agentRegistry.AgentInterface

// GetID returns the agent's unique identifier
func (a *BaseAgent) GetID() string {
	return a.id
}

// GetCapabilities returns the agent's capabilities
func (a *BaseAgent) GetCapabilities() []agentRegistry.AgentCapability {
	return a.capabilities
}

// GetStatus returns the current agent status
func (a *BaseAgent) GetStatus() agentRegistry.AgentStatus {
	return agentRegistry.AgentStatus{
		ID:           a.id,
		Type:         a.agentType,
		Status:       "running",
		LastActivity: time.Now(),
		LoadFactor:   0.1,
		Version:      "1.0.0",
		Metadata: map[string]interface{}{
			"uptime":         time.Since(a.startTime).String(),
			"framework_type": "base_agent",
		},
	}
}

// Start initializes the agent
func (a *BaseAgent) Start(ctx context.Context) error {
	a.logger.Info("🚀 Starting agent: %s", a.id)
	return nil
}

// Stop shuts down the agent
func (a *BaseAgent) Stop(ctx context.Context) error {
	a.logger.Info("🛑 Stopping agent: %s", a.id)
	return nil
}

// Health returns the agent's health status
func (a *BaseAgent) Health() agentRegistry.HealthStatus {
	return agentRegistry.HealthStatus{
		Healthy: true,
		Status:  "healthy",
		Message: "Agent is operational",
	}
}

// ProcessEvent handles incoming events using the configured handler
func (a *BaseAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🎯 Processing event: %s", event.Subject)

	if a.eventHandler == nil {
		return a.CreateErrorResponse(event, "No event handler configured"), nil
	}

	response, err := a.eventHandler(ctx, event)
	if err != nil {
		a.logger.Error("❌ Event processing failed: %v", err)
		return a.CreateErrorResponse(event, err.Error()), nil
	}

	a.logger.Info("✅ Event processed successfully")
	return response, nil
}

// HandleIncomingEvent handles events from the event bus and ensures responses are emitted
func (a *BaseAgent) HandleIncomingEvent(ctx context.Context, event events.Event) error {
	a.logger.Info("📨 Received event: %s from %s", event.Subject, event.Source)

	// Process the event using the agent's main processing logic
	responseEvent, err := a.ProcessEvent(ctx, &event)
	if err != nil {
		a.logger.Error("❌ Failed to process event: %v", err)
		return err
	}

	// CRITICAL: If we have a response event, emit it back to the event bus
	if responseEvent != nil && a.eventBus != nil {
		// Preserve correlation information for agent-to-agent communication
		if correlationID, ok := event.Payload["correlation_id"]; ok {
			if responseEvent.Payload == nil {
				responseEvent.Payload = make(map[string]interface{})
			}
			responseEvent.Payload["correlation_id"] = correlationID
		}
		if requestID, ok := event.Payload["request_id"]; ok {
			if responseEvent.Payload == nil {
				responseEvent.Payload = make(map[string]interface{})
			}
			responseEvent.Payload["request_id"] = requestID
		}

		// Add required event metadata
		if responseEvent.ID == "" {
			responseEvent.ID = fmt.Sprintf("%s-response-%d", a.id, time.Now().UnixNano())
		}
		if responseEvent.Type == "" {
			responseEvent.Type = events.EventTypeResponse
		}
		if responseEvent.Source == "" {
			responseEvent.Source = a.id
		}
		if responseEvent.Timestamp == 0 {
			responseEvent.Timestamp = time.Now().UnixNano()
		}

		// Emit the response event to the orchestrator
		err = a.eventBus.EmitEvent(*responseEvent)
		if err != nil {
			a.logger.Error("❌ Failed to emit response event: %v", err)
			return err
		}
		a.logger.Info("✅ %s sent response: %s", a.id, responseEvent.Subject)
	} else if responseEvent == nil {
		a.logger.Warn("⚠️ No response event generated from ProcessEvent")
	} else {
		a.logger.Warn("⚠️ No event bus available to send response")
	}

	return nil
}

// ==================================================================================
// FRAMEWORK HELPER METHODS FOR COMMON AGENT PATTERNS
// ==================================================================================

// ExtractIntent extracts the intent from an event payload with standardized error handling
func (a *BaseAgent) ExtractIntent(event *events.Event) (string, error) {
	intent, ok := event.Payload["intent"].(string)
	if !ok || intent == "" {
		return "", fmt.Errorf("intent field required in payload")
	}
	return intent, nil
}

// ExtractUserMessage extracts user message from event payload, checking both top-level and context
func (a *BaseAgent) ExtractUserMessage(event *events.Event) (string, bool) {
	// Check top-level payload first
	if msg, ok := event.Payload["user_message"].(string); ok && msg != "" {
		a.logger.Info("🔍 Found user_message at top level: %s", msg)
		return msg, true
	}

	// Check nested context
	if contextData, ok := event.Payload["context"].(map[string]interface{}); ok {
		a.logger.Info("🔍 Context keys: %v", GetPayloadKeys(contextData))
		if msg, ok := contextData["user_message"].(string); ok && msg != "" {
			a.logger.Info("🔍 Found user_message in context: %s", msg)
			return msg, true
		}
	}

	return "", false
}

// RouteByIntent provides a convenient way to route events based on intent patterns
func (a *BaseAgent) RouteByIntent(ctx context.Context, event *events.Event, intentHandlers map[string]func(context.Context, *events.Event) (*events.Event, error)) (*events.Event, error) {
	intent, err := a.ExtractIntent(event)
	if err != nil {
		return a.CreateErrorResponse(event, err.Error()), nil
	}

	// Try exact match first
	if handler, exists := intentHandlers[intent]; exists {
		return handler(ctx, event)
	}

	// Try pattern matching with strings.Contains
	for pattern, handler := range intentHandlers {
		if strings.Contains(intent, pattern) {
			return handler(ctx, event)
		}
	}

	// No handler found
	return a.CreateErrorResponse(event, fmt.Sprintf("no handler found for intent: %s", intent)), nil
}

// CallAI provides a standardized way to call AI providers (AI-native, no fallbacks)
func (a *BaseAgent) CallAI(ctx context.Context, aiProvider ai.AIProvider, systemPrompt, userPrompt string) (string, error) {
	if aiProvider == nil {
		return "", fmt.Errorf("AI provider not available")
	}

	response, err := aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("AI call failed: %w", err)
	}

	return response, nil
}

// ParseJSONResponse parses a JSON response from AI with cleaning and error handling
func (a *BaseAgent) ParseJSONResponse(response string, target interface{}) error {
	// Clean up the response - remove any markdown code blocks
	cleaned := strings.TrimSpace(response)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	return json.Unmarshal([]byte(cleaned), target)
}

// ExtractStructuredDataWithAI uses AI to extract structured data from user messages
// This is a generic pattern that any agent can use for parsing natural language into structured data
func (a *BaseAgent) ExtractStructuredDataWithAI(
	ctx context.Context,
	aiProvider ai.AIProvider,
	userMessage string,
	systemPrompt string,
	target interface{},
) error {
	if aiProvider == nil {
		return fmt.Errorf("AI provider not available")
	}

	userPrompt := fmt.Sprintf("Extract structured data from this message: %s", userMessage)

	response, err := aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("AI call failed: %w", err)
	}

	if err := a.ParseJSONResponse(response, target); err != nil {
		return fmt.Errorf("failed to parse AI response: %w", err)
	}

	return nil
}

// GetPayloadKeys returns the keys of a map for debugging purposes
func GetPayloadKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Framework helper methods

// CreateResponse creates a standardized success response
func (a *BaseAgent) CreateResponse(message string, payload map[string]interface{}, correlationEvent *events.Event) *events.Event {
	responsePayload := make(map[string]interface{})

	// Copy the original payload
	for k, v := range payload {
		responsePayload[k] = v
	}

	// Add standard fields
	responsePayload["status"] = "success"
	responsePayload["message"] = message
	responsePayload["agent_id"] = a.id

	// Add correlation ID if available
	if correlationEvent != nil {
		if correlationID, ok := correlationEvent.Payload["correlation_id"]; ok {
			responsePayload["correlation_id"] = correlationID
		}
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.id,
		Subject: "Response from " + a.id,
		Payload: responsePayload,
	}
}

// CreateErrorResponse creates a standardized error response
func (a *BaseAgent) CreateErrorResponse(correlationEvent *events.Event, errorMessage string) *events.Event {
	responsePayload := map[string]interface{}{
		"status":           "error",
		"error":            errorMessage,
		"agent_id":         a.id,
		"response_content": errorMessage,
	}

	// Add correlation ID if available
	if correlationEvent != nil {
		if correlationID, ok := correlationEvent.Payload["correlation_id"]; ok {
			responsePayload["correlation_id"] = correlationID
		}
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.id,
		Subject: "Error from " + a.id,
		Payload: responsePayload,
	}
}

// GetLogger returns the agent's logger
func (a *BaseAgent) GetLogger() *logging.Logger {
	return a.logger
}

// subscribeToCapabilities automatically subscribes to routing keys based on capabilities
func (a *BaseAgent) subscribeToCapabilities() error {
	if a.eventBus == nil {
		return nil // No event bus available
	}

	for _, capability := range a.capabilities {
		for _, routingKey := range capability.RoutingKeys {
			a.eventBus.SubscribeToRoutingKey(routingKey, func(event events.Event) error {
				// Use the comprehensive event handler that includes response emission
				return a.HandleIncomingEvent(context.Background(), event)
			})
			a.logger.Info("✅ Subscribed to routing key: %s", routingKey)
		}
	}

	return nil
}
