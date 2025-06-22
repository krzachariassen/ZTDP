package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agentFramework"
	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// AIResponse represents the structure of AI responses for parameter extraction
type AIResponse struct {
	Action          string  `json:"action"`
	ApplicationName string  `json:"application_name,omitempty"`
	Details         string  `json:"details,omitempty"`
	Confidence      float64 `json:"confidence"`
	Clarification   string  `json:"clarification,omitempty"`
}

// ApplicationAgent - AI-native application agent following best practices
type ApplicationAgent struct {
	service    *Service
	aiProvider ai.AIProvider
	logger     *logging.Logger
	eventBus   *events.EventBus
}

// NewApplicationAgent creates a new AI-native application agent
func NewApplicationAgent(
	graph *graph.GlobalGraph,
	aiProvider ai.AIProvider,
	eventBus *events.EventBus,
	registry agentRegistry.AgentRegistry,
) (agentRegistry.AgentInterface, error) {
	// Validate required dependencies - AI provider is mandatory for AI-native agent
	if graph == nil {
		return nil, fmt.Errorf("graph is required")
	}
	if aiProvider == nil {
		return nil, fmt.Errorf("aiProvider is required for AI-native agent")
	}
	if eventBus == nil {
		return nil, fmt.Errorf("eventBus is required")
	}
	if registry == nil {
		return nil, fmt.Errorf("registry is required")
	}

	// Create the application service for business logic
	service := NewService(graph, aiProvider)

	// Create the agent wrapper
	wrapper := &ApplicationAgent{
		service:    service,
		aiProvider: aiProvider,
		logger:     logging.GetLogger().ForComponent("application-agent"),
		eventBus:   eventBus,
	}

	// Create dependencies for the framework
	deps := agentFramework.AgentDependencies{
		Registry: registry,
		EventBus: eventBus,
	}

	// Build the agent using the framework
	agent, err := agentFramework.NewAgent("application-agent").
		WithType("application").
		WithCapabilities(getApplicationCapabilities()).
		WithEventHandler(wrapper.handleEvent).
		Build(deps)

	if err != nil {
		return nil, fmt.Errorf("failed to build application agent: %w", err)
	}

	wrapper.logger.Info("✅ AI-native ApplicationAgent created successfully")
	return agent, nil
}

// getApplicationCapabilities returns the capabilities for the application agent
func getApplicationCapabilities() []agentRegistry.AgentCapability {
	return []agentRegistry.AgentCapability{
		{
			Name:        "application_management",
			Description: "AI-native application lifecycle management with no fallback logic",
			Intents: []string{
				"create application", "list applications", "get application", "update application",
				"delete application", "application management", "app management",
				"application creation", "application discovery", "manage apps",
				"show applications", "find applications",
			},
			InputTypes:  []string{"user_message"},
			OutputTypes: []string{"application_result", "application_status", "application_list", "clarification"},
			RoutingKeys: []string{"application.request", "application.create", "application.list", "application.management"},
			Version:     "2.0.0",
		},
	}
}

// handleEvent processes events for the application agent (only handles application domain events)
func (a *ApplicationAgent) handleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🎯 Processing application domain event: %s", event.Subject)

	// Only handle application events - other domains have their own agents
	if a.matchesPattern(event.Subject, "application.*") {
		return a.handleApplicationEvent(ctx, event)
	}

	// If we get here, this event shouldn't be routed to the application agent
	a.logger.Warn("Application agent received non-application event: %s", event.Subject)
	return a.createErrorResponse(event, fmt.Sprintf("Application agent only handles application events, not: %s", event.Subject)), nil
}

// Helper method to match event subject patterns
func (a *ApplicationAgent) matchesPattern(subject, pattern string) bool {
	// Simple pattern matching for application events only
	if pattern == "application.*" {
		return subject == "application.request" || subject == "application.create" ||
			subject == "application.list" || subject == "application.management"
	}
	return false
}

// handleApplicationEvent processes application-specific events
func (a *ApplicationAgent) handleApplicationEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🎯 Processing application event: %s", event.Subject)

	// Extract user message from event payload
	userMessage, exists := event.Payload["message"].(string)
	if !exists {
		userMessage, exists = event.Payload["query"].(string)
		if !exists {
			userMessage, exists = event.Payload["request"].(string)
			if !exists {
				userMessage = "list applications" // Default for empty requests
			}
		}
	}

	// Use AI to extract intent and parameters - pure AI-native approach
	aiResponse, err := a.extractIntentAndParameters(ctx, userMessage)
	if err != nil {
		a.logger.Error("AI extraction failed: %v", err)
		return a.createErrorResponse(event, fmt.Sprintf("I'm having trouble understanding your request: %v", err)), nil
	}

	a.logger.Info("🤖 AI extracted - action: %s, app: %s, confidence: %.2f",
		aiResponse.Action, aiResponse.ApplicationName, aiResponse.Confidence)

	// Check confidence level - request clarification if too low
	if aiResponse.Confidence < 0.7 {
		clarificationMsg := aiResponse.Clarification
		if clarificationMsg == "" {
			clarificationMsg = fmt.Sprintf("I'm not completely sure what you want to do (confidence: %.0f%%). Could you please clarify your request?", aiResponse.Confidence*100)
		}
		return a.createClarificationResponse(event, clarificationMsg), nil
	}

	// Route to appropriate handler based on AI-extracted action
	switch aiResponse.Action {
	case "list", "show", "get":
		return a.handleApplicationList(ctx, event, aiResponse)
	case "create", "add":
		return a.handleApplicationCreate(ctx, event, aiResponse)
	case "update", "modify":
		return a.handleApplicationUpdate(ctx, event, aiResponse)
	case "delete", "remove":
		return a.handleApplicationDelete(ctx, event, aiResponse)
	default:
		return a.createClarificationResponse(event, fmt.Sprintf("I'm not sure how to '%s' applications. I can list, create, update, or delete applications.", aiResponse.Action)), nil
	}
}

// extractIntentAndParameters uses AI to parse user message and extract structured parameters
func (a *ApplicationAgent) extractIntentAndParameters(ctx context.Context, userMessage string) (*AIResponse, error) {
	systemPrompt := `You are an application management assistant. Parse the user's request and extract the action and parameters.

Available actions: list, create, update, delete, show, get

Response format must be valid JSON:
{
  "action": "list|create|update|delete|show|get",
  "application_name": "name if specified or null",
  "details": "any additional context",
  "confidence": 0.0-1.0,
  "clarification": "what to ask if confidence < 0.8"
}

Set confidence < 0.8 if:
- Action is unclear
- Required parameters are missing for create/update/delete
- Request is ambiguous

Examples:
- "list all applications" -> {"action": "list", "confidence": 0.9}
- "create app called myapp" -> {"action": "create", "application_name": "myapp", "confidence": 0.9}
- "do something" -> {"action": "unknown", "confidence": 0.2, "clarification": "What would you like to do with applications?"}`

	userPrompt := fmt.Sprintf("Parse this application request: %s", userMessage)

	aiResponseText, err := a.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	var response AIResponse
	if err := json.Unmarshal([]byte(aiResponseText), &response); err != nil {
		a.logger.Warn("Failed to parse AI response as JSON: %v", err)
		// If AI response isn't valid JSON, return low confidence instead of fallback logic
		return &AIResponse{
			Action:        "unknown",
			Confidence:    0.1,
			Clarification: "I had trouble understanding your application request. Could you please rephrase what you want to do?",
		}, nil
	}

	a.logger.Info("🤖 AI extracted - action: %s, confidence: %.2f", response.Action, response.Confidence)
	return &response, nil
}

// AI-native handler methods

// handleApplicationList processes application listing requests
func (a *ApplicationAgent) handleApplicationList(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("📋 AI-native application listing")

	// Use service to list applications
	applications, err := a.service.ListApplications()
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to list applications: %v", err)), nil
	}

	// Create success response
	payload := map[string]interface{}{
		"action":       "list",
		"applications": applications,
		"count":        len(applications),
		"ai_response":  aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleApplicationCreate processes application creation requests
func (a *ApplicationAgent) handleApplicationCreate(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("🆕 AI-native application creation")

	// Validate required parameters
	if aiResponse.ApplicationName == "" {
		return a.createClarificationResponse(event, "What would you like to name the new application?"), nil
	}

	// Create application contract
	appContract := contracts.ApplicationContract{
		Metadata: contracts.Metadata{
			Name:  aiResponse.ApplicationName,
			Owner: "user", // Could be extracted from context in real implementation
		},
		Spec: contracts.ApplicationSpec{
			Description: fmt.Sprintf("Application %s created via AI", aiResponse.ApplicationName),
		},
	}

	// Use service to create application
	err := a.service.CreateApplication(appContract)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to create application: %v", err)), nil
	}

	// Create success response
	payload := map[string]interface{}{
		"action":           "create",
		"application_name": aiResponse.ApplicationName,
		"status":           "created",
		"ai_response":      aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleApplicationUpdate processes application update requests
func (a *ApplicationAgent) handleApplicationUpdate(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("✏️ AI-native application update")

	// Validate required parameters
	if aiResponse.ApplicationName == "" {
		return a.createClarificationResponse(event, "Which application would you like to update?"), nil
	}

	// For now, return a placeholder since update logic depends on what fields to update
	payload := map[string]interface{}{
		"action":           "update",
		"application_name": aiResponse.ApplicationName,
		"status":           "not_implemented",
		"message":          "Application update feature coming soon",
		"ai_response":      aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleApplicationDelete processes application deletion requests
func (a *ApplicationAgent) handleApplicationDelete(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("🗑️ AI-native application deletion")

	// Validate required parameters
	if aiResponse.ApplicationName == "" {
		return a.createClarificationResponse(event, "Which application would you like to delete?"), nil
	}

	// Use service to delete application
	err := a.service.DeleteApplication(aiResponse.ApplicationName)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to delete application: %v", err)), nil
	}

	// Create success response
	payload := map[string]interface{}{
		"action":           "delete",
		"application_name": aiResponse.ApplicationName,
		"status":           "deleted",
		"ai_response":      aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// Response helper methods

// getCorrelationID extracts the correlation ID from the orchestrator's event payload
func (a *ApplicationAgent) getCorrelationID(originalEvent *events.Event) string {
	// First, try to extract correlation_id from the event payload (set by orchestrator)
	if originalEvent.Payload != nil {
		if correlationID, exists := originalEvent.Payload["correlation_id"]; exists {
			if correlationStr, ok := correlationID.(string); ok {
				return correlationStr
			}
		}
	}

	// Fallback to event ID if no correlation_id found in payload
	return originalEvent.ID
}

func (a *ApplicationAgent) createSuccessResponse(originalEvent *events.Event, payload map[string]interface{}) *events.Event {
	// Extract correlation ID from incoming event payload (set by orchestrator)
	correlationID := a.getCorrelationID(originalEvent)

	return &events.Event{
		ID:        fmt.Sprintf("response-%d", time.Now().UnixNano()),
		Subject:   fmt.Sprintf("application.response.%s", originalEvent.ID),
		Type:      "response", // Use standard event type for orchestrator compatibility
		Source:    "application-agent",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":           "success",
			"correlation_id":   correlationID,
			"original_subject": originalEvent.Subject,
			"data":             payload,
		},
	}
}

func (a *ApplicationAgent) createErrorResponse(originalEvent *events.Event, errorMessage string) *events.Event {
	// Extract correlation ID from incoming event payload (set by orchestrator)
	correlationID := a.getCorrelationID(originalEvent)

	return &events.Event{
		ID:        fmt.Sprintf("error-%d", time.Now().UnixNano()),
		Subject:   fmt.Sprintf("application.error.%s", originalEvent.ID),
		Type:      "response", // Use standard event type for orchestrator compatibility
		Source:    "application-agent",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":           "error",
			"error":            errorMessage,
			"correlation_id":   correlationID,
			"original_subject": originalEvent.Subject,
		},
	}
}

func (a *ApplicationAgent) createClarificationResponse(originalEvent *events.Event, clarificationMessage string) *events.Event {
	// Extract correlation ID from incoming event payload (set by orchestrator)
	correlationID := a.getCorrelationID(originalEvent)

	return &events.Event{
		ID:        fmt.Sprintf("clarification-%d", time.Now().UnixNano()),
		Subject:   fmt.Sprintf("application.clarification.%s", originalEvent.ID),
		Type:      "response", // Use standard event type for orchestrator compatibility
		Source:    "application-agent",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"status":           "clarification_needed",
			"clarification":    clarificationMessage,
			"correlation_id":   correlationID,
			"original_subject": originalEvent.Subject,
		},
	}
}
