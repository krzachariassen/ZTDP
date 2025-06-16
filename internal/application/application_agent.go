package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

// ApplicationAgent implements AgentInterface for application lifecycle management
type ApplicationAgent struct {
	service   *Service
	agentID   string
	env       string
	eventBus  EventBus
	startTime time.Time
}

// EventBus interface for the application agent
type EventBus interface {
	Emit(eventType string, data map[string]interface{}) error
}

// NewApplicationAgent creates a new ApplicationAgent that implements AgentInterface
func NewApplicationAgent(graph *graph.GlobalGraph, env string, eventBus EventBus) agents.AgentInterface {
	service := NewService(graph)
	return &ApplicationAgent{
		service:   service,
		agentID:   "application-agent",
		env:       env,
		eventBus:  eventBus,
		startTime: time.Now(),
	}
}

// GetID returns the agent's unique identifier
func (a *ApplicationAgent) GetID() string {
	return a.agentID
}

// GetStatus returns current agent status information
func (a *ApplicationAgent) GetStatus() agents.AgentStatus {
	return agents.AgentStatus{
		ID:           a.agentID,
		Type:         "application",
		Status:       "running",
		LastActivity: time.Now(),
		LoadFactor:   0.1,
		Version:      "1.0.0",
		Metadata: map[string]interface{}{
			"environment":          a.env,
			"supported_operations": []string{"create", "read", "update", "list"},
			"graph_integration":    true,
		},
	}
}

// GetCapabilities returns the agent's capabilities
func (a *ApplicationAgent) GetCapabilities() []agents.AgentCapability {
	return []agents.AgentCapability{
		{
			Name:        "application_lifecycle",
			Description: "Manages application lifecycle operations (create, update, delete, query)",
			Intents:     []string{"create application", "update application", "delete application", "list applications"},
			InputTypes:  []string{"ApplicationContract", "application_name"},
			OutputTypes: []string{"application_status", "application_list", "application_details"},
			Version:     "1.0.0",
		},
		{
			Name:        "application_validation",
			Description: "Validates application configurations and constraints",
			Intents:     []string{"validate application", "check application", "verify application"},
			InputTypes:  []string{"ApplicationContract"},
			OutputTypes: []string{"validation_result", "error_list"},
			Version:     "1.0.0",
		},
		{
			Name:        "application_query",
			Description: "Queries application information and relationships",
			Intents:     []string{"get application", "find application", "search applications"},
			InputTypes:  []string{"application_name", "query_criteria"},
			OutputTypes: []string{"application_details", "application_list"},
			Version:     "1.0.0",
		},
	}
}

// ProcessEvent handles incoming events for the application agent
func (a *ApplicationAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	// Validate event has required intent field
	intent, ok := event.Payload["intent"].(string)
	if !ok {
		return nil, fmt.Errorf("application agent requires 'intent' field in payload")
	}

	// Route based on intent
	switch intent {
	case "application_create":
		return a.handleCreateApplication(ctx, event)
	case "application_read":
		return a.handleReadApplication(ctx, event)
	case "application_list":
		return a.handleListApplications(ctx, event)
	case "application_update":
		return a.handleUpdateApplication(ctx, event)
	default:
		return a.handleGenericQuestion(ctx, event, intent)
	}
}

// handleCreateApplication processes application creation requests
func (a *ApplicationAgent) handleCreateApplication(ctx context.Context, event *events.Event) (*events.Event, error) {
	// Extract application contract from payload
	appData, ok := event.Payload["application"]
	if !ok {
		return a.createErrorResponse(event, "application data required for creation")
	}

	// Convert to ApplicationContract
	var appContract contracts.ApplicationContract
	if err := a.convertToApplicationContract(appData, &appContract); err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("invalid application contract: %v", err))
	}

	// Create application using service
	err := a.service.CreateApplication(appContract)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to create application: %v", err))
	}

	// Emit success event
	if a.eventBus != nil {
		a.eventBus.Emit("application_created", map[string]interface{}{
			"application_name": appContract.Metadata.Name,
			"owner":            appContract.Metadata.Owner,
			"status":           "created",
		})
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Application created successfully",
		Payload: map[string]interface{}{
			"status":           "success",
			"operation":        "create",
			"application_name": appContract.Metadata.Name,
			"message":          fmt.Sprintf("Application %s created successfully", appContract.Metadata.Name),
		},
	}, nil
}

// handleReadApplication processes application read requests
func (a *ApplicationAgent) handleReadApplication(ctx context.Context, event *events.Event) (*events.Event, error) {
	appName, ok := event.Payload["application_name"].(string)
	if !ok {
		return a.createErrorResponse(event, "application_name required for read operation")
	}

	app, err := a.service.GetApplication(appName)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to get application: %v", err))
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Application retrieved successfully",
		Payload: map[string]interface{}{
			"status":      "success",
			"operation":   "read",
			"application": app,
		},
	}, nil
}

// handleListApplications processes application list requests
func (a *ApplicationAgent) handleListApplications(ctx context.Context, event *events.Event) (*events.Event, error) {
	apps, err := a.service.ListApplications()
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to list applications: %v", err))
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Applications listed successfully",
		Payload: map[string]interface{}{
			"status":       "success",
			"operation":    "list",
			"applications": apps,
			"count":        len(apps),
		},
	}, nil
}

// handleUpdateApplication processes application update requests
func (a *ApplicationAgent) handleUpdateApplication(ctx context.Context, event *events.Event) (*events.Event, error) {
	appName, ok := event.Payload["application_name"].(string)
	if !ok {
		return a.createErrorResponse(event, "application_name required for update operation")
	}

	appData, ok := event.Payload["application"]
	if !ok {
		return a.createErrorResponse(event, "application data required for update")
	}

	var appContract contracts.ApplicationContract
	if err := a.convertToApplicationContract(appData, &appContract); err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("invalid application contract: %v", err))
	}

	err := a.service.UpdateApplication(appName, appContract)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("failed to update application: %v", err))
	}

	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Application updated successfully",
		Payload: map[string]interface{}{
			"status":           "success",
			"operation":        "update",
			"application_name": appName,
			"message":          fmt.Sprintf("Application %s updated successfully", appName),
		},
	}, nil
}

// handleGenericQuestion processes general application-related questions
func (a *ApplicationAgent) handleGenericQuestion(ctx context.Context, event *events.Event, intent string) (*events.Event, error) {
	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Application agent response",
		Payload: map[string]interface{}{
			"status":  "processed",
			"agent":   "application",
			"intent":  intent,
			"message": fmt.Sprintf("Application agent received intent '%s' but needs specific application operation (create, read, update, list)", intent),
		},
	}, nil
}

// Start initializes the agent
func (a *ApplicationAgent) Start(ctx context.Context) error {
	// Agent is already ready to process events
	return nil
}

// Stop gracefully shuts down the agent
func (a *ApplicationAgent) Stop(ctx context.Context) error {
	// Clean shutdown logic here if needed
	return nil
}

// Health returns the agent's health status
func (a *ApplicationAgent) Health() agents.HealthStatus {
	return agents.HealthStatus{
		Healthy: true,
		Status:  "healthy",
		Message: "Application agent is running normally",
		Checks: map[string]interface{}{
			"graph_connection": "connected",
			"event_bus":        "connected",
			"service":          "ready",
		},
		CheckedAt: time.Now(),
	}
}

// Helper functions

func (a *ApplicationAgent) createErrorResponse(originalEvent *events.Event, errorMsg string) (*events.Event, error) {
	return &events.Event{
		Type:    events.EventTypeResponse,
		Source:  a.agentID,
		Subject: "Application operation failed",
		Payload: map[string]interface{}{
			"status":  "error",
			"agent":   "application",
			"message": errorMsg,
		},
	}, nil
}

func (a *ApplicationAgent) convertToApplicationContract(data interface{}, contract *contracts.ApplicationContract) error {
	// Convert via JSON marshaling/unmarshaling for type safety
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal application data: %v", err)
	}

	err = json.Unmarshal(jsonData, contract)
	if err != nil {
		return fmt.Errorf("failed to unmarshal to ApplicationContract: %v", err)
	}

	return nil
}
