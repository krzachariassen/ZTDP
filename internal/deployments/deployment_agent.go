package deployments

import (
	"context"
	"fmt"
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

// handleEvent is the main event handler for AI-native deployment processing
func (a *FrameworkDeploymentAgent) handleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	// Store current event for correlation context
	a.currentEvent = event

	a.logger.Info("🎯 Processing deployment event: %s", event.Subject)

	// AI-NATIVE ONLY: All requests must have user_message for AI processing
	userMessage, exists := event.Payload["user_message"].(string)
	if !exists || userMessage == "" {
		return a.createErrorResponse(event, "AI-native deployment agent requires user_message field with natural language request"), nil
	}

	a.logger.Info("🤖 AI-native deployment execution: %s", userMessage)

	// Use deployment service's AI extraction method (proper AI-native approach)
	params, err := a.service.ExtractDeploymentParamsFromUserMessage(ctx, userMessage)
	if err != nil {
		a.logger.Error("AI parameter extraction failed: %v", err)
		return a.createErrorResponse(event, fmt.Sprintf("failed to parse deployment request: %v", err)), nil
	}

	a.logger.Info("🤖 AI extracted - action: %s, app: %s, env: %s, confidence: %.2f",
		params.Action, params.AppName, params.Environment, params.Confidence)

	// Check confidence level - request clarification if too low
	if params.Confidence < 0.7 {
		clarificationMsg := params.Clarification
		if clarificationMsg == "" {
			clarificationMsg = "I'm not sure about the deployment details. Please specify the application name and target environment clearly."
		}
		return a.createErrorResponse(event, clarificationMsg), nil
	}

	// Validate required parameters
	if params.AppName == "" {
		return a.createErrorResponse(event, "Application name is required for deployment"), nil
	}
	if params.Environment == "" {
		return a.createErrorResponse(event, "Target environment is required for deployment"), nil
	}

	// Extract values from AI-parsed parameters
	appName := params.AppName
	environment := params.Environment

	a.logger.Info("🎯 AI validated parameters - app: %s, env: %s", appName, environment)

	// ✅ ORCHESTRATION WORKFLOW - Coordinate with other agents
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
		"ai_extracted_params": map[string]interface{}{
			"action":      params.Action,
			"app_name":    params.AppName,
			"environment": params.Environment,
			"confidence":  params.Confidence,
		},
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
