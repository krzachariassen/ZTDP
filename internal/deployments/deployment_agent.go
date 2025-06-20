package deployments

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agentFramework"
	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// FrameworkDeploymentAgent wraps the deployment business logic in the new agent framework
type FrameworkDeploymentAgent struct {
	service      *Service
	env          string
	logger       *logging.Logger
	eventBus     *events.EventBus // Store EventBus for emitting events
	currentEvent *events.Event    // Store current event context for correlation
}

// NewDeploymentAgent creates a DeploymentAgent using the agent framework
func NewDeploymentAgent(
	graph *graph.GlobalGraph,
	aiProvider ai.AIProvider,
	eventBus *events.EventBus,
	registry agentRegistry.AgentRegistry,
) (agentRegistry.AgentInterface, error) {
	// Create the deployment service for business logic
	service := NewDeploymentService(graph, aiProvider)

	// Create the wrapper that contains the business logic
	wrapper := &FrameworkDeploymentAgent{
		service:  service,
		env:      "", // Agents are environment-agnostic
		logger:   logging.GetLogger().ForComponent("deployment-agent"),
		eventBus: eventBus,
	}

	// Create dependencies for the framework
	deps := agentFramework.AgentDependencies{
		Registry: registry,
		EventBus: eventBus,
	}

	// Build the agent using the framework
	agent, err := agentFramework.NewAgent("deployment-agent").
		WithType("deployment").
		WithCapabilities(getDeploymentCapabilities()).
		WithEventHandler(wrapper.handleEvent).
		Build(deps)

	if err != nil {
		return nil, fmt.Errorf("failed to build framework deployment agent: %w", err)
	}

	wrapper.logger.Info("✅ FrameworkDeploymentAgent created successfully")
	return agent, nil
}

// getDeploymentCapabilities returns the capabilities for the deployment agent
func getDeploymentCapabilities() []agentRegistry.AgentCapability {
	return []agentRegistry.AgentCapability{
		{
			Name:        "deployment_orchestration",
			Description: "Orchestrates application deployments with AI planning and execution",
			Intents: []string{
				"deploy application", "execute deployment", "start deployment", "run deployment",
				"deploy to environment", "perform deployment", "application deployment",
				"deployment execution", "deploy app", "deploy service",
			},
			InputTypes:  []string{"application", "environment", "deployment_plan", "deployment_config"},
			OutputTypes: []string{"deployment_result", "deployment_status", "deployment_plan"},
			RoutingKeys: []string{"deployment.request", "deployment.execute", "deployment.orchestration"},
			Version:     "1.0.0",
		},
		{
			Name:        "deployment_planning",
			Description: "Generates AI-enhanced deployment plans and strategies",
			Intents: []string{
				"plan deployment", "generate deployment plan", "deployment strategy",
				"create deployment plan", "deployment planning", "plan application deployment",
			},
			InputTypes:  []string{"application", "environment", "dependencies", "configuration"},
			OutputTypes: []string{"deployment_plan", "deployment_order", "deployment_strategy"},
			RoutingKeys: []string{"deployment.planning", "deployment.plan"},
			Version:     "1.0.0",
		},
		{
			Name:        "deployment_status",
			Description: "Provides deployment status monitoring and reporting",
			Intents: []string{
				"deployment status", "check deployment", "deployment progress",
				"get deployment status", "deployment health", "deployment monitoring",
			},
			InputTypes:  []string{"application", "environment", "deployment_id"},
			OutputTypes: []string{"deployment_status", "deployment_health", "status_report"},
			RoutingKeys: []string{"deployment.status", "deployment.monitoring"},
			Version:     "1.0.0",
		},
	}
}

// handleEvent is the main event handler that preserves the existing business logic
func (a *FrameworkDeploymentAgent) handleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	// Store current event for correlation context
	a.currentEvent = event

	a.logger.Info("🎯 Processing deployment event: %s", event.Subject)

	// Extract intent from event payload using framework pattern
	intent, ok := event.Payload["intent"].(string)
	if !ok || intent == "" {
		return a.createErrorResponse(event, "intent field required in payload"), nil
	}

	// Route based on intent - using a cleaner pattern
	intentHandlers := map[string]func(context.Context, *events.Event) (*events.Event, error){
		"deploy":  a.handleDeployment,
		"plan":    a.handleDeploymentPlan,
		"status":  a.handleDeploymentStatus,
		"execute": a.handleDeployment,
		"start":   a.handleDeployment,
		"run":     a.handleDeployment,
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

	// Default to generic handler
	return a.handleGenericDeploymentQuestion(ctx, event, intent)
}

// handleDeployment processes deployment execution requests
func (a *FrameworkDeploymentAgent) handleDeployment(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("🚀 Deployment execution payload keys: %v", agentFramework.GetPayloadKeys(event.Payload))

	// Try to extract from user message first (AI-native approach)
	if userMessage, exists := event.Payload["user_message"].(string); exists {
		return a.handleAINativeDeploymentExecution(ctx, userMessage, event)
	}

	// Extract required parameters (fallback for explicit parameters)
	appName, ok := event.Payload["application"].(string)
	if !ok {
		appName, ok = event.Payload["app"].(string)
		if !ok {
			return a.createErrorResponse(event, "deployment requires application field or user_message"), nil
		}
	}

	environment, ok := event.Payload["environment"].(string)
	if !ok {
		environment, ok = event.Payload["env"].(string)
		if !ok {
			return a.createErrorResponse(event, "deployment requires environment field or user_message"), nil
		}
	}

	// Execute deployment via service
	result, err := a.service.DeployApplication(ctx, appName, environment)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("deployment failed: %v", err)), nil
	}

	// Create success response
	payload := map[string]interface{}{
		"deployment_id":     result.DeploymentID,
		"application":       result.Application,
		"environment":       result.Environment,
		"deployment_result": result,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleDeploymentPlan processes deployment planning requests
func (a *FrameworkDeploymentAgent) handleDeploymentPlan(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("📋 Deployment planning payload keys: %v", agentFramework.GetPayloadKeys(event.Payload))

	// Extract required parameters
	appName, ok := event.Payload["application"].(string)
	if !ok {
		appName, ok = event.Payload["app"].(string)
		if !ok {
			return a.createErrorResponse(event, "deployment planning requires application field"), nil
		}
	}

	// Generate deployment plan via service
	plan, err := a.service.GenerateDeploymentPlan(ctx, appName)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("deployment planning failed: %v", err)), nil
	}

	// Create success response
	payload := map[string]interface{}{
		"application":     appName,
		"deployment_plan": plan,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleDeploymentStatus processes deployment status requests
func (a *FrameworkDeploymentAgent) handleDeploymentStatus(ctx context.Context, event *events.Event) (*events.Event, error) {
	a.logger.Info("📊 Deployment status payload keys: %v", agentFramework.GetPayloadKeys(event.Payload))

	// Extract required parameters
	appName, ok := event.Payload["application"].(string)
	if !ok {
		appName, ok = event.Payload["app"].(string)
		if !ok {
			return a.createErrorResponse(event, "deployment status requires application field"), nil
		}
	}

	environment, ok := event.Payload["environment"].(string)
	if !ok {
		environment, ok = event.Payload["env"].(string)
		if !ok {
			return a.createErrorResponse(event, "deployment status requires environment field"), nil
		}
	}

	// Get deployment status via service
	status, err := a.service.GetDeploymentStatus(appName, environment)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("deployment status failed: %v", err)), nil
	}

	// Create success response
	payload := map[string]interface{}{
		"application":       appName,
		"environment":       environment,
		"deployment_status": status,
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleGenericDeploymentQuestion handles general deployment-related questions
func (a *FrameworkDeploymentAgent) handleGenericDeploymentQuestion(ctx context.Context, event *events.Event, intent string) (*events.Event, error) {
	a.logger.Info("❓ Generic deployment question with intent: %s", intent)

	// For generic deployment questions, provide helpful guidance
	payload := map[string]interface{}{
		"intent":  intent,
		"message": "I can help with deployment operations. Please specify 'deploy', 'plan', or 'status' with application and environment details.",
		"capabilities": []string{
			"deploy application to environment",
			"generate deployment plan for application",
			"check deployment status of application in environment",
		},
	}

	return a.createSuccessResponse(event, payload), nil
}

// handleAINativeDeploymentExecution uses AI to parse user message and extract deployment intent
func (a *FrameworkDeploymentAgent) handleAINativeDeploymentExecution(ctx context.Context, userMessage string, event *events.Event) (*events.Event, error) {
	a.logger.Info("🤖 AI-native deployment execution: %s", userMessage)

	// Use AI to parse the user message and extract application and environment
	systemPrompt := `You are a deployment assistant. Extract the application name and environment from the user's deployment request.
	
Response format must be JSON:
{
  "application": "app-name",
  "environment": "env-name",
  "action": "deploy|plan|status",
  "confidence": 0.0-1.0
}

If you cannot determine the application or environment with high confidence, set confidence < 0.8.`

	userPrompt := fmt.Sprintf("Parse this deployment request: %s", userMessage)

	// Call AI provider (assuming it exists in the agent)
	if a.service.aiProvider == nil {
		return a.createErrorResponse(event, "AI provider not available for parsing user message"), nil
	}

	response, err := a.service.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		a.logger.Error("AI parsing failed: %v", err)
		return a.createErrorResponse(event, fmt.Sprintf("failed to parse deployment request: %v", err)), nil
	}

	// Parse AI response to extract application and environment
	// For now, use a simple parsing approach (in real implementation, would use JSON parsing)
	var appName, environment string

	// Simple heuristic parsing for demo (should be replaced with proper JSON parsing)
	if strings.Contains(strings.ToLower(userMessage), "to production") || strings.Contains(strings.ToLower(userMessage), "prod") {
		environment = "production"
	} else if strings.Contains(strings.ToLower(userMessage), "to staging") || strings.Contains(strings.ToLower(userMessage), "staging") {
		environment = "staging"
	} else if strings.Contains(strings.ToLower(userMessage), "to dev") || strings.Contains(strings.ToLower(userMessage), "development") {
		environment = "development"
	} else if strings.Contains(strings.ToLower(userMessage), "to test") || strings.Contains(strings.ToLower(userMessage), "testing") {
		environment = "test"
	}

	// Simple app name extraction (look for "deploy <app-name>")
	words := strings.Fields(strings.ToLower(userMessage))
	for i, word := range words {
		if (word == "deploy" || word == "deployment") && i+1 < len(words) {
			appName = words[i+1]
			break
		}
	}

	// Validate extracted values
	if appName == "" {
		return a.createErrorResponse(event, "could not determine application name from message"), nil
	}
	if environment == "" {
		return a.createErrorResponse(event, "could not determine environment from message"), nil
	}

	a.logger.Info("🎯 AI extracted - app: %s, env: %s", appName, environment)

	// ✅ ORCHESTRATION WORKFLOW - Coordinate with other agents
	// Step 1: Request release creation from Release Agent
	result, err := a.orchestrateDeployment(ctx, appName, environment, userMessage)
	if err != nil {
		return a.createErrorResponse(event, fmt.Sprintf("deployment orchestration failed: %v", err)), nil
	}

	// Create success response
	payload := map[string]interface{}{
		"deployment_id":     result.DeploymentID,
		"release_id":        result.ReleaseID,
		"application":       result.Application,
		"environment":       result.Environment,
		"deployment_result": result,
		"parsed_from":       userMessage,
		"ai_response":       response,
	}

	return a.createSuccessResponse(event, payload), nil
}

// orchestrateDeployment implements the full multi-agent deployment workflow
func (a *FrameworkDeploymentAgent) orchestrateDeployment(ctx context.Context, appName, environment, userMessage string) (*DeploymentResult, error) {
	a.logger.Info("🎭 Orchestrating deployment: %s → %s", appName, environment)

	// Step 1: Create deployment plan (simple for TDD)
	plan := []string{"validate", "create-release", "evaluate-policies", "execute"}
	a.logger.Info("📋 Created simple deployment plan for %s", appName)

	// Step 2: Request Release Agent to create a release
	releaseID, err := a.requestReleaseCreation(ctx, appName, plan)
	if err != nil {
		return nil, fmt.Errorf("release creation failed: %w", err)
	}

	// Step 3: Create deployment edge from Release to Environment
	deploymentID, err := a.createDeploymentEdge(ctx, releaseID, environment, "pending")
	if err != nil {
		return nil, fmt.Errorf("deployment edge creation failed: %w", err)
	}

	// Step 4: Request Policy Agent validation
	policyDecision, err := a.requestPolicyValidation(ctx, appName, environment, releaseID)
	if err != nil {
		// Update deployment status to failed
		a.updateDeploymentStatus(ctx, deploymentID, "failed", fmt.Sprintf("Policy validation failed: %v", err))
		return nil, fmt.Errorf("policy validation failed: %w", err)
	}

	if policyDecision != "allowed" {
		// Update deployment status to blocked
		a.updateDeploymentStatus(ctx, deploymentID, "blocked", "Deployment blocked by policy")
		return nil, fmt.Errorf("deployment blocked by policy: %s", policyDecision)
	}

	// Step 5: Update status to in-progress and execute deployment
	a.updateDeploymentStatus(ctx, deploymentID, "in-progress", "Executing deployment")

	// Step 6: Execute actual deployment (currently mocked)
	result, err := a.executeDeployment(ctx, appName, environment, releaseID, deploymentID)
	if err != nil {
		// Update deployment status to failed
		a.updateDeploymentStatus(ctx, deploymentID, "failed", fmt.Sprintf("Deployment execution failed: %v", err))
		return nil, fmt.Errorf("deployment execution failed: %w", err)
	}

	// Step 7: Update final status to succeeded
	a.updateDeploymentStatus(ctx, deploymentID, "succeeded", "Deployment completed successfully")

	// Step 8: Emit deployment.completed event
	completionEvent := events.Event{
		Subject: "deployment.completed",
		Source:  "deployment-agent",
		Type:    events.EventTypeNotify,
		Payload: map[string]interface{}{
			"deployment_id": deploymentID,
			"application":   appName,
			"environment":   environment,
			"release_id":    releaseID,
			"status":        "succeeded",
			"timestamp":     time.Now().Unix(),
		},
	}

	if err := a.eventBus.EmitEvent(completionEvent); err != nil {
		a.logger.Error("Failed to emit deployment.completed event: %v", err)
		// Don't fail the deployment for event emission errors
	} else {
		a.logger.Info("📤 Emitted deployment.completed event for %s → %s", appName, environment)
	}

	a.logger.Info("✅ Deployment orchestration completed: %s", deploymentID)
	return result, nil
}

// requestReleaseCreation coordinates with Release Agent to create a release
func (a *FrameworkDeploymentAgent) requestReleaseCreation(ctx context.Context, appName string, plan []string) (string, error) {
	a.logger.Info("📦 Requesting release creation for %s", appName)

	// Step 1: Emit "release.create" event to Release Agent
	releaseCreateEvent := events.Event{
		Type:    events.EventTypeBroadcast,
		Source:  "deployment-agent",
		Subject: "release.create",
		Payload: map[string]interface{}{
			"application": appName,
			"plan":        plan,
			"timestamp":   time.Now().Unix(),
		},
	}

	err := a.eventBus.EmitEvent(releaseCreateEvent)
	if err != nil {
		return "", fmt.Errorf("failed to emit release.create event: %w", err)
	}

	a.logger.Info("📤 Emitted release.create event for %s", appName)

	// TODO: Wait for "release.created" response from Release Agent
	// For now, generate a release ID (should be replaced with actual Release Agent coordination)
	releaseID := fmt.Sprintf("release-%s-%d", appName, time.Now().Unix())

	a.logger.Info("📦 Release created: %s", releaseID)
	return releaseID, nil
}

// createDeploymentEdge creates a deployment edge from Release to Environment in the graph
func (a *FrameworkDeploymentAgent) createDeploymentEdge(ctx context.Context, releaseID, environment, status string) (string, error) {
	a.logger.Info("🔗 Creating deployment edge: %s → %s", releaseID, environment)

	deploymentID := fmt.Sprintf("deployment-%s-%s-%d", releaseID, environment, time.Now().Unix())

	// Get current graph
	currentGraph, err := a.service.globalGraph.Graph()
	if err != nil {
		return "", fmt.Errorf("failed to get graph: %w", err)
	}

	// Create deployment edge with metadata
	edge := graph.Edge{
		To:   environment,
		Type: "deployment",
		Metadata: map[string]interface{}{
			"deployment_id": deploymentID,
			"status":        status,
			"created_at":    time.Now().Format(time.RFC3339),
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	// Add edge to graph
	if currentGraph.Edges == nil {
		currentGraph.Edges = make(map[string][]graph.Edge)
	}
	currentGraph.Edges[releaseID] = append(currentGraph.Edges[releaseID], edge)

	// Save graph
	if err := a.service.globalGraph.Save(); err != nil {
		return "", fmt.Errorf("failed to save graph: %w", err)
	}

	a.logger.Info("🔗 Deployment edge created: %s", deploymentID)
	return deploymentID, nil
}

// requestPolicyValidation coordinates with Policy Agent for deployment validation
func (a *FrameworkDeploymentAgent) requestPolicyValidation(ctx context.Context, appName, environment, releaseID string) (string, error) {
	a.logger.Info("🛡️ Requesting policy validation for %s → %s", appName, environment)

	// Emit policy.evaluate event to Policy Agent
	policyEvent := events.Event{
		Type:    events.EventTypeRequest,
		Source:  "deployment-agent",
		Subject: "policy.evaluate",
		Payload: map[string]interface{}{
			"application": appName,
			"environment": environment,
			"release_id":  releaseID,
			"timestamp":   time.Now().Unix(),
		},
	}

	if err := a.eventBus.EmitEvent(policyEvent); err != nil {
		return "error", fmt.Errorf("failed to emit policy evaluation event: %w", err)
	}

	a.logger.Info("📤 Emitted policy.evaluate event for %s → %s", appName, environment)

	// TODO: In full implementation, wait for "policy.decision" response
	// For now, simulate policy validation for TDD purposes

	// Simple validation for demo
	if environment == "production" && appName == "critical-app" {
		return "blocked", fmt.Errorf("critical application requires manual approval for production")
	}

	a.logger.Info("🛡️ Policy validation passed")
	return "allowed", nil
}

// updateDeploymentStatus updates the deployment edge status in the graph
func (a *FrameworkDeploymentAgent) updateDeploymentStatus(ctx context.Context, deploymentID, status, message string) error {
	a.logger.Info("📊 Updating deployment status: %s → %s", deploymentID, status)

	// Get current graph
	currentGraph, err := a.service.globalGraph.Graph()
	if err != nil {
		return fmt.Errorf("failed to get graph: %w", err)
	}

	// Find and update the deployment edge
	for from, edges := range currentGraph.Edges {
		for i, edge := range edges {
			if edge.Type == "deployment" {
				if deploymentIDVal, ok := edge.Metadata["deployment_id"].(string); ok && deploymentIDVal == deploymentID {
					// Update status and timestamp
					edge.Metadata["status"] = status
					edge.Metadata["updated_at"] = time.Now().Format(time.RFC3339)
					edge.Metadata["message"] = message
					currentGraph.Edges[from][i] = edge

					// Save graph
					if err := a.service.globalGraph.Save(); err != nil {
						return fmt.Errorf("failed to save graph: %w", err)
					}

					a.logger.Info("📊 Deployment status updated: %s", status)
					return nil
				}
			}
		}
	}

	return fmt.Errorf("deployment edge not found: %s", deploymentID)
}

// executeDeployment performs the actual deployment (currently mocked)
func (a *FrameworkDeploymentAgent) executeDeployment(ctx context.Context, appName, environment, releaseID, deploymentID string) (*DeploymentResult, error) {
	a.logger.Info("🚀 Executing deployment: %s → %s", appName, environment)

	// TODO: Implement actual deployment logic
	// For now, simulate deployment execution

	// This would normally:
	// 1. Apply Kubernetes manifests
	// 2. Update load balancers
	// 3. Run health checks
	// 4. Monitor deployment progress
	// 5. Handle rollback if needed

	result := &DeploymentResult{
		DeploymentID: deploymentID,
		Application:  appName,
		Environment:  environment,
		ReleaseID:    releaseID,
		Status:       "completed",
		Message:      "Deployment completed successfully",
	}

	a.logger.Info("🚀 Deployment execution completed")
	return result, nil
}

// createErrorResponse creates a standardized error response
func (a *FrameworkDeploymentAgent) createErrorResponse(originalEvent *events.Event, errorMsg string) *events.Event {
	payload := map[string]interface{}{
		"status":      "error",
		"error":       errorMsg,
		"original_id": originalEvent.ID,
		"timestamp":   time.Now().Unix(),
		"agent_id":    "deployment-agent",
	}

	// Preserve correlation_id if it exists
	if correlationID, ok := originalEvent.Payload["correlation_id"]; ok {
		payload["correlation_id"] = correlationID
	}

	return &events.Event{
		ID:        fmt.Sprintf("response-%s", originalEvent.ID),
		Type:      events.EventTypeResponse,
		Subject:   "deployment.response.error",
		Source:    "deployment-agent",
		Timestamp: time.Now().Unix(),
		Payload:   payload,
	}
}

// createSuccessResponse creates a standardized success response
func (a *FrameworkDeploymentAgent) createSuccessResponse(originalEvent *events.Event, payload map[string]interface{}) *events.Event {
	// Ensure required fields
	payload["original_id"] = originalEvent.ID
	payload["agent_id"] = "deployment-agent"
	payload["status"] = "success"
	payload["timestamp"] = time.Now().Unix()

	// Preserve correlation_id if it exists
	if correlationID, ok := originalEvent.Payload["correlation_id"]; ok {
		payload["correlation_id"] = correlationID
	}

	return &events.Event{
		ID:        fmt.Sprintf("response-%s", originalEvent.ID),
		Type:      events.EventTypeResponse,
		Subject:   "deployment.response.success",
		Source:    "deployment-agent",
		Timestamp: time.Now().Unix(),
		Payload:   payload,
	}
}
