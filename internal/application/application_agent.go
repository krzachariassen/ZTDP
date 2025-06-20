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
	ServiceName     string  `json:"service_name,omitempty"`
	ReleaseName     string  `json:"release_name,omitempty"`
	Environment     string  `json:"environment,omitempty"`
	Version         string  `json:"version,omitempty"`
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
		{
			Name:        "service_management",
			Description: "Manages application services and service versioning with AI assistance",
			Intents: []string{
				"create service", "list services", "get service", "update service", "version service",
				"service management", "service versioning", "manage service versions",
				"deploy service", "service configuration", "service discovery",
			},
			InputTypes:  []string{"service", "service_config", "service_version"},
			OutputTypes: []string{"service_result", "service_status", "service_list", "version_info"},
			RoutingKeys: []string{"service.request", "service.create", "service.list", "service.version", "service.management"},
			Version:     "1.0.0",
		},
		{
			Name:        "environment_management",
			Description: "Manages application environments with AI assistance",
			Intents: []string{
				"create environment", "list environments", "get environment", "update environment",
				"environment management", "env management", "environment configuration",
				"setup environment", "environment discovery", "manage environments",
			},
			InputTypes:  []string{"environment", "environment_config", "environment_metadata"},
			OutputTypes: []string{"environment_result", "environment_status", "environment_list"},
			RoutingKeys: []string{"environment.request", "environment.create", "environment.list", "environment.management"},
			Version:     "1.0.0",
		},
		{
			Name:        "release_management",
			Description: "Manages application releases and deployment coordination with AI assistance",
			Intents: []string{
				"create release", "list releases", "get release", "update release", "delete release",
				"release management", "release coordination", "release planning",
				"deploy release", "release versioning", "manage releases",
			},
			InputTypes:  []string{"release", "release_config", "release_plan", "service_versions"},
			OutputTypes: []string{"release_result", "release_status", "release_list", "deployment_plan"},
			RoutingKeys: []string{"release.request", "release.create", "release.list", "release.management", "release.deploy"},
			Version:     "1.0.0",
		},
	}
}

// handleEvent processes events for the application agent (handles all application domain events)
func (a *ApplicationAgent) handleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🎯 Processing application domain event: %s", event.Subject)

	// Route based on event subject pattern to determine domain type
	switch {
	case a.matchesPattern(event.Subject, "application.*"):
		return a.handleApplicationEvent(ctx, event)
	case a.matchesPattern(event.Subject, "service.*"):
		return a.handleServiceEvent(ctx, event)
	case a.matchesPattern(event.Subject, "environment.*"):
		return a.handleEnvironmentEvent(ctx, event)
	case a.matchesPattern(event.Subject, "release.*"):
		return a.handleReleaseEvent(ctx, event)
	default:
		a.logger.Warn("Unknown event subject pattern: %s", event.Subject)
		return a.createErrorResponse(event, fmt.Sprintf("I don't know how to handle events with subject: %s", event.Subject)), nil
	}
}

// Helper method to match event subject patterns
func (a *ApplicationAgent) matchesPattern(subject, pattern string) bool {
	// Simple pattern matching - could be enhanced with regex if needed
	if pattern == "application.*" {
		return subject == "application.request" || subject == "application.create" ||
			subject == "application.list" || subject == "application.management"
	}
	if pattern == "service.*" {
		return subject == "service.request" || subject == "service.create" ||
			subject == "service.list" || subject == "service.version" || subject == "service.management"
	}
	if pattern == "environment.*" {
		return subject == "environment.request" || subject == "environment.create" ||
			subject == "environment.list" || subject == "environment.management"
	}
	if pattern == "release.*" {
		return subject == "release.request" || subject == "release.create" ||
			subject == "release.list" || subject == "release.management" || subject == "release.deploy"
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
	aiResponse, err := a.extractIntentAndParameters(ctx, userMessage, "application")
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
func (a *ApplicationAgent) extractIntentAndParameters(ctx context.Context, userMessage, domainType string) (*AIResponse, error) {
	var systemPrompt string

	switch domainType {
	case "application":
		systemPrompt = `You are an application management assistant. Parse the user's request and extract the action and parameters.

Available actions: list, create, update, delete, show, get

Response format must be valid JSON:
{
  "action": "list|create|update|delete|show|get",
  "application_name": "name if specified or null",
  "environment": "environment if specified or null", 
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

	case "service":
		systemPrompt = `You are a service management assistant. Parse the user's request and extract the action and parameters.

Available actions: list, create, update, delete, show, get, version

Response format must be valid JSON:
{
  "action": "list|create|update|delete|show|get|version",
  "application_name": "parent application name if specified or null",
  "service_name": "service name if specified or null",
  "version": "version if specified or null",
  "details": "any additional context",
  "confidence": 0.0-1.0,
  "clarification": "what to ask if confidence < 0.8"
}

Examples:
- "list services for myapp" -> {"action": "list", "application_name": "myapp", "confidence": 0.9}
- "create service api in myapp" -> {"action": "create", "application_name": "myapp", "service_name": "api", "confidence": 0.9}`

	case "environment":
		systemPrompt = `You are an environment management assistant. Parse the user's request and extract the action and parameters.

Available actions: list, create, update, delete, show, get

Response format must be valid JSON:
{
  "action": "list|create|update|delete|show|get",
  "environment": "environment name if specified or null",
  "application_name": "application name if specified or null",
  "details": "any additional context",
  "confidence": 0.0-1.0,
  "clarification": "what to ask if confidence < 0.8"
}

Examples:
- "list environments" -> {"action": "list", "confidence": 0.9}
- "create environment staging" -> {"action": "create", "environment": "staging", "confidence": 0.9}`

	case "release":
		systemPrompt = `You are a release management assistant. Parse the user's request and extract the action and parameters.

Available actions: list, create, update, delete, show, get, deploy

Response format must be valid JSON:
{
  "action": "list|create|update|delete|show|get|deploy",
  "application_name": "application name if specified or null",
  "release_name": "release name if specified or null",
  "version": "version if specified or null",
  "environment": "target environment if specified or null",
  "details": "any additional context",
  "confidence": 0.0-1.0,
  "clarification": "what to ask if confidence < 0.8"
}

Examples:
- "list releases for myapp" -> {"action": "list", "application_name": "myapp", "confidence": 0.9}
- "create release v1.0 for myapp" -> {"action": "create", "application_name": "myapp", "release_name": "v1.0", "confidence": 0.9}`

	default:
		return nil, fmt.Errorf("unsupported domain type: %s", domainType)
	}

	userPrompt := fmt.Sprintf("Parse this %s request: %s", domainType, userMessage)

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
			Clarification: fmt.Sprintf("I had trouble understanding your %s request. Could you please rephrase what you want to do?", domainType),
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

// Service Event Handlers

// handleServiceEvent processes service-specific events
func (a *ApplicationAgent) handleServiceEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🔧 Processing service event: %s", event.Subject)

	// Extract user message from event payload
	userMessage, exists := event.Payload["message"].(string)
	if !exists {
		userMessage, exists = event.Payload["query"].(string)
		if !exists {
			userMessage, exists = event.Payload["request"].(string)
			if !exists {
				userMessage = "list services" // Default for empty requests
			}
		}
	}

	// Use AI to extract intent and parameters for service domain
	aiResponse, err := a.extractIntentAndParameters(ctx, userMessage, "service")
	if err != nil {
		a.logger.Error("AI extraction failed: %v", err)
		return a.createErrorResponse(event, fmt.Sprintf("I'm having trouble understanding your service request: %v", err)), nil
	}

	a.logger.Info("🤖 Service AI extracted - action: %s, confidence: %.2f", aiResponse.Action, aiResponse.Confidence)

	// Check confidence level
	if aiResponse.Confidence < 0.7 {
		clarificationMsg := aiResponse.Clarification
		if clarificationMsg == "" {
			clarificationMsg = fmt.Sprintf("I'm not completely sure what you want to do with services (confidence: %.0f%%). Could you please clarify?", aiResponse.Confidence*100)
		}
		return a.createClarificationResponse(event, clarificationMsg), nil
	}

	// Route to appropriate service handler
	switch aiResponse.Action {
	case "list", "show", "get":
		return a.handleServiceList(ctx, event, aiResponse)
	case "create", "add":
		return a.handleServiceCreate(ctx, event, aiResponse)
	case "version":
		return a.handleServiceVersion(ctx, event, aiResponse)
	default:
		return a.createClarificationResponse(event, fmt.Sprintf("I'm not sure how to '%s' services. I can list, create, or version services.", aiResponse.Action)), nil
	}
}

// handleServiceList processes service listing requests
func (a *ApplicationAgent) handleServiceList(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("📋 AI-native service listing")

	// Create service service instance
	serviceService := NewServiceService(a.service.Graph)

	var services []map[string]interface{}
	var err error

	// If application name is specified, list services for that app
	if aiResponse.ApplicationName != "" {
		services, err = serviceService.ListServices(aiResponse.ApplicationName)
	} else {
		// List all services (would need a method for this, for now return empty)
		services = []map[string]interface{}{}
	}

	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to list services: %v", err)), nil
	}

	payload := map[string]interface{}{
		"action":           "list",
		"services":         services,
		"application_name": aiResponse.ApplicationName,
		"count":            len(services),
		"ai_response":      aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleServiceCreate processes service creation requests
func (a *ApplicationAgent) handleServiceCreate(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("🆕 AI-native service creation")

	// Validate required parameters
	if aiResponse.ApplicationName == "" {
		return a.createClarificationResponse(event, "Which application should I create the service in?"), nil
	}

	// Create service service instance
	serviceService := NewServiceService(a.service.Graph)

	// Create basic service data (could be enhanced with AI to extract more details)
	serviceData := map[string]interface{}{
		"name":        fmt.Sprintf("service-%d", time.Now().Unix()),
		"description": fmt.Sprintf("Service created via AI for %s", aiResponse.ApplicationName),
		"type":        "microservice",
	}

	// Use service service to create service
	result, err := serviceService.CreateService(aiResponse.ApplicationName, serviceData)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to create service: %v", err)), nil
	}

	payload := map[string]interface{}{
		"action":           "create",
		"service":          result,
		"application_name": aiResponse.ApplicationName,
		"status":           "created",
		"ai_response":      aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleServiceVersion processes service versioning requests
func (a *ApplicationAgent) handleServiceVersion(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("🔢 AI-native service versioning")

	// For now, return a placeholder
	payload := map[string]interface{}{
		"action":      "version",
		"status":      "not_implemented",
		"message":     "Service versioning feature coming soon",
		"ai_response": aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// Environment Event Handlers

// handleEnvironmentEvent processes environment-specific events
func (a *ApplicationAgent) handleEnvironmentEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🌍 Processing environment event: %s", event.Subject)

	// Extract user message from event payload
	userMessage, exists := event.Payload["message"].(string)
	if !exists {
		userMessage, exists = event.Payload["query"].(string)
		if !exists {
			userMessage, exists = event.Payload["request"].(string)
			if !exists {
				userMessage = "list environments" // Default for empty requests
			}
		}
	}

	// Use AI to extract intent and parameters for environment domain
	aiResponse, err := a.extractIntentAndParameters(ctx, userMessage, "environment")
	if err != nil {
		a.logger.Error("AI extraction failed: %v", err)
		return a.createErrorResponse(event, fmt.Sprintf("I'm having trouble understanding your environment request: %v", err)), nil
	}

	a.logger.Info("🤖 Environment AI extracted - action: %s, confidence: %.2f", aiResponse.Action, aiResponse.Confidence)

	// Check confidence level
	if aiResponse.Confidence < 0.7 {
		clarificationMsg := aiResponse.Clarification
		if clarificationMsg == "" {
			clarificationMsg = fmt.Sprintf("I'm not completely sure what you want to do with environments (confidence: %.0f%%). Could you please clarify?", aiResponse.Confidence*100)
		}
		return a.createClarificationResponse(event, clarificationMsg), nil
	}

	// Route to appropriate environment handler
	switch aiResponse.Action {
	case "list", "show", "get":
		return a.handleEnvironmentList(ctx, event, aiResponse)
	case "create", "add":
		return a.handleEnvironmentCreate(ctx, event, aiResponse)
	default:
		return a.createClarificationResponse(event, fmt.Sprintf("I'm not sure how to '%s' environments. I can list or create environments.", aiResponse.Action)), nil
	}
}

// handleEnvironmentList processes environment listing requests
func (a *ApplicationAgent) handleEnvironmentList(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("📋 AI-native environment listing")

	// Create environment service instance
	envService := NewEnvironmentService(a.service.Graph)

	environments, err := envService.ListEnvironmentsAsData()
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to list environments: %v", err)), nil
	}

	payload := map[string]interface{}{
		"action":       "list",
		"environments": environments,
		"count":        len(environments),
		"ai_response":  aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleEnvironmentCreate processes environment creation requests
func (a *ApplicationAgent) handleEnvironmentCreate(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("🆕 AI-native environment creation")

	// Validate required parameters
	envName := aiResponse.Environment
	if envName == "" {
		return a.createClarificationResponse(event, "What would you like to name the new environment?"), nil
	}

	// Create environment service instance
	envService := NewEnvironmentService(a.service.Graph)

	// Create basic environment data
	envData := map[string]interface{}{
		"name":        envName,
		"description": fmt.Sprintf("Environment %s created via AI", envName),
		"owner":       "user",
	}

	// Use environment service to create environment
	result, err := envService.CreateEnvironmentFromData(envData)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to create environment: %v", err)), nil
	}

	payload := map[string]interface{}{
		"action":      "create",
		"environment": result,
		"status":      "created",
		"ai_response": aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// Release Event Handlers

// handleReleaseEvent processes release-specific events
func (a *ApplicationAgent) handleReleaseEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🚀 Processing release event: %s", event.Subject)

	// Extract user message from event payload
	userMessage, exists := event.Payload["message"].(string)
	if !exists {
		userMessage, exists = event.Payload["query"].(string)
		if !exists {
			userMessage, exists = event.Payload["request"].(string)
			if !exists {
				userMessage = "list releases" // Default for empty requests
			}
		}
	}

	// Use AI to extract intent and parameters for release domain
	aiResponse, err := a.extractIntentAndParameters(ctx, userMessage, "release")
	if err != nil {
		a.logger.Error("AI extraction failed: %v", err)
		return a.createErrorResponse(event, fmt.Sprintf("I'm having trouble understanding your release request: %v", err)), nil
	}

	a.logger.Info("🤖 Release AI extracted - action: %s, confidence: %.2f", aiResponse.Action, aiResponse.Confidence)

	// Check confidence level
	if aiResponse.Confidence < 0.7 {
		clarificationMsg := aiResponse.Clarification
		if clarificationMsg == "" {
			clarificationMsg = fmt.Sprintf("I'm not completely sure what you want to do with releases (confidence: %.0f%%). Could you please clarify?", aiResponse.Confidence*100)
		}
		return a.createClarificationResponse(event, clarificationMsg), nil
	}

	// Route to appropriate release handler
	switch aiResponse.Action {
	case "list", "show", "get":
		return a.handleReleaseList(ctx, event, aiResponse)
	case "create", "add":
		return a.handleReleaseCreate(ctx, event, aiResponse)
	case "deploy":
		return a.handleReleaseDeploy(ctx, event, aiResponse)
	default:
		return a.createClarificationResponse(event, fmt.Sprintf("I'm not sure how to '%s' releases. I can list, create, or deploy releases.", aiResponse.Action)), nil
	}
}

// handleReleaseList processes release listing requests
func (a *ApplicationAgent) handleReleaseList(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("📋 AI-native release listing")

	// Create release service instance
	releaseService := NewReleaseService(a.service.Graph)

	releases, err := releaseService.ListReleases(aiResponse.ApplicationName)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to list releases: %v", err)), nil
	}

	payload := map[string]interface{}{
		"action":           "list",
		"releases":         releases,
		"application_name": aiResponse.ApplicationName,
		"count":            len(releases),
		"ai_response":      aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleReleaseCreate processes release creation requests
func (a *ApplicationAgent) handleReleaseCreate(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("🆕 AI-native release creation")

	// Validate required parameters
	if aiResponse.ApplicationName == "" {
		return a.createClarificationResponse(event, "Which application should I create the release for?"), nil
	}

	// Create release service instance
	releaseService := NewReleaseService(a.service.Graph)

	// Create basic release contract
	releaseContract := contracts.ReleaseContract{
		Metadata: contracts.Metadata{
			Name:  fmt.Sprintf("release-%d", time.Now().Unix()),
			Owner: "user",
		},
		Spec: contracts.ReleaseSpec{
			Application:     aiResponse.ApplicationName,
			Version:         "1.0.0",    // Could be extracted from AI response
			ServiceVersions: []string{}, // Would need service versions
			Status:          "planned",
			Strategy:        "rolling",
			Notes:           "Release created via AI",
			Timestamp:       time.Now(),
		},
	}

	// Use release service to create release
	err := releaseService.CreateRelease(releaseContract)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("Failed to create release: %v", err)), nil
	}

	payload := map[string]interface{}{
		"action":           "create",
		"release":          releaseContract,
		"application_name": aiResponse.ApplicationName,
		"status":           "created",
		"ai_response":      aiResponse,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleReleaseDeploy processes release deployment requests
func (a *ApplicationAgent) handleReleaseDeploy(ctx context.Context, event *events.Event, aiResponse *AIResponse) (*events.Event, error) {
	a.logger.Info("🚀 AI-native release deployment")

	// For now, return a placeholder
	payload := map[string]interface{}{
		"action":      "deploy",
		"status":      "not_implemented",
		"message":     "Release deployment feature coming soon",
		"ai_response": aiResponse,
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
