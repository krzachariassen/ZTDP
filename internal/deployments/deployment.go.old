package deployments

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// DeploymentResult represents the result of a deployment operation
type DeploymentResult struct {
	Application  string                   `json:"application"`
	Environment  string                   `json:"environment"`
	DeploymentID string                   `json:"deployment_id"`
	Deployments  []string                 `json:"deployments"`
	Skipped      []string                 `json:"skipped"`
	Failed       []map[string]interface{} `json:"failed"`
	Summary      DeploymentSummary        `json:"summary"`
	Status       string                   `json:"status"` // "initiated", "in_progress", "completed", "failed"
}

// DeploymentSummary provides a high-level summary of the deployment
type DeploymentSummary struct {
	TotalServices int    `json:"total_services"`
	Deployed      int    `json:"deployed"`
	Skipped       int    `json:"skipped"`
	Failed        int    `json:"failed"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}

// Engine handles application deployments with policy enforcement using Event-Driven Architecture
type Engine struct {
	graph    *graph.GlobalGraph
	provider ai.AIProvider
	planner  DeploymentPlanner
	logger   *logging.Logger
}

// NewEngine creates a new event-driven deployment engine with an AI provider
func NewEngine(g *graph.GlobalGraph, provider ai.AIProvider) *Engine {
	logger := logging.GetLogger().ForComponent("deployment-engine")

	// AI provider is optional - deployment engine can work without AI
	if provider == nil {
		logger.Warn("⚠️ No AI provider available - using basic deployment planning")
	}

	planner := NewAIDeploymentPlanner(g, provider)

	return &Engine{
		graph:    g,
		provider: provider,
		planner:  planner,
		logger:   logger,
	}
}

// NewEngineWithProvider creates a new deployment engine with AI provider directly
// This avoids the AI Brain dependency and uses clean domain separation
func NewEngineWithProvider(g *graph.GlobalGraph, provider ai.AIProvider) *Engine {
	logger := logging.GetLogger().ForComponent("deployment-engine")

	// Create AI deployment planner - AI provider is required for AI-native platform
	if provider == nil {
		logger.Error("❌ AI provider is required for deployment engine - this is an AI-native platform")
		// Return nil engine since we can't operate without AI
		return nil
	}

	planner := NewAIDeploymentPlanner(g, provider)

	return &Engine{
		graph:    g,
		provider: provider,
		planner:  planner,
		logger:   logger,
	}
}

// isResourceInstance checks if a resource node is an instance (has application context)
// rather than a catalog resource
func (e *Engine) isResourceInstance(node *graph.Node) bool {
	if node.Kind != "resource" {
		return false
	}

	// Resource instances have application metadata, catalog resources do not
	if app, hasApp := node.Metadata["application"]; hasApp && app != nil {
		return true
	}

	return false
}

// ExecuteApplicationDeployment initiates an event-driven deployment for an entire application
func (e *Engine) ExecuteApplicationDeployment(appName, environment string) (*DeploymentResult, error) {
	// Generate unique deployment ID
	deploymentID := uuid.New().String()

	// Validate application exists
	currentGraph, err := e.graph.Graph()
	if err != nil {
		return nil, fmt.Errorf("failed to get current graph: %v", err)
	}

	if _, ok := currentGraph.Nodes[appName]; !ok {
		return nil, fmt.Errorf("application '%s' not found", appName)
	}

	// Validate environment exists
	if _, ok := currentGraph.Nodes[environment]; !ok {
		return nil, fmt.Errorf("environment '%s' not found", environment)
	}

	// Check if application is allowed to deploy to this environment
	allowedEnvs := e.getAllowedEnvironmentsForApp(appName)
	allowed := false
	for _, env := range allowedEnvs {
		if env == environment {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("application '%s' is not allowed to deploy to environment '%s'", appName, environment)
	}

	// Note: We do NOT create application→environment deploy edges.
	// Deployment edges are created from individual service_version and resource nodes to environments.

	// Generate execution plan for the application - AI automatically discovers dependencies
	plan, err := e.generateDeploymentPlan(appName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI deployment plan: %v", err)
	}

	// Collect services and resources for the deployment
	var services []string
	var resources []string

	for _, nodeID := range plan {
		if node, exists := currentGraph.Nodes[nodeID]; exists {
			switch node.Kind {
			case "service_version":
				services = append(services, nodeID)
			case "resource":
				resources = append(resources, nodeID)
			}
		}
	}

	// ==================================================================================
	// EDA COMPLIANCE: Emit deployment requested event instead of calling RPs directly
	// Resource Providers subscribe to this event and react accordingly
	// ==================================================================================
	events.GlobalEventBus.Emit(events.EventTypeNotify, "deployment-engine", "deployment_requested", map[string]interface{}{
		"type":          "deployment_requested",
		"application":   appName,
		"environment":   environment,
		"services":      services,
		"resources":     resources,
		"deployment_id": deploymentID,
		"plan":          plan,
	})

	// Execute immediate synchronous operations (graph updates)
	result := &DeploymentResult{
		Application:  appName,
		Environment:  environment,
		DeploymentID: deploymentID,
		Deployments:  []string{},
		Skipped:      []string{},
		Failed:       []map[string]interface{}{},
		Status:       "initiated",
	}

	var deployments []string
	var skipped []string
	var failed []map[string]interface{}

	// Process service version and resource deployments (synchronous graph operations)
	for _, nodeID := range plan {
		node, exists := currentGraph.Nodes[nodeID]
		if !exists {
			failed = append(failed, map[string]interface{}{
				"node":   nodeID,
				"reason": "Node not found in graph",
			})
			continue
		}

		// Deploy both service versions and resource instances to the target environment
		// Note: Only deploy resource INSTANCES, not catalog resources
		if node.Kind == "service_version" || (node.Kind == "resource" && e.isResourceInstance(node)) {
			// Check if already deployed to this environment using backend-aware method
			alreadyDeployed, err := e.graph.HasEdge(nodeID, environment, "deploy")
			if err != nil {
				failed = append(failed, map[string]interface{}{
					"node":   nodeID,
					"reason": "Failed to check deployment status: " + err.Error(),
				})
				continue
			}

			if alreadyDeployed {
				skipped = append(skipped, nodeID)
				continue
			}

			// Attempt to create deployment edge (this will enforce policies)
			if err := e.graph.AddEdge(nodeID, environment, "deploy"); err != nil {
				failed = append(failed, map[string]interface{}{
					"node":   nodeID,
					"reason": err.Error(),
				})
				continue
			}

			deployments = append(deployments, nodeID)
		} else {
			// For other node types (application, service), just record as skipped
			skipped = append(skipped, nodeID)
		}
	}

	// Save graph changes if any deployments were successful
	if len(deployments) > 0 {
		if err := e.graph.Save(); err != nil {
			// Emit failure event
			events.GlobalEventBus.Emit(events.EventTypeNotify, "deployment-engine", "deployment_failed", map[string]interface{}{
				"type":          "deployment_failed",
				"application":   appName,
				"environment":   environment,
				"deployment_id": deploymentID,
				"error":         "failed to save graph changes: " + err.Error(),
				"failed":        []string{},
			})
			return nil, fmt.Errorf("failed to save deployments: %v", err)
		}

		// Update policy checks related to this deployment
		e.updatePolicyChecks(appName, environment)

		// Emit deployment started event (Resource Providers will begin actual deployment)
		events.GlobalEventBus.Emit(events.EventTypeNotify, "deployment-engine", "deployment_started", map[string]interface{}{
			"type":          "deployment_started",
			"application":   appName,
			"environment":   environment,
			"deployment_id": deploymentID,
		})
	}

	// Build final result
	result.Deployments = deployments
	result.Skipped = skipped
	result.Failed = failed
	result.Summary = DeploymentSummary{
		TotalServices: len(plan),
		Deployed:      len(deployments),
		Skipped:       len(skipped),
		Failed:        len(failed),
		Success:       len(failed) == 0,
		Message:       e.generateDeploymentMessage(appName, environment, len(deployments), len(failed)),
	}

	// Status depends on whether we have failures or if everything was skipped
	if len(failed) > 0 {
		result.Status = "failed"
		events.GlobalEventBus.Emit(events.EventTypeNotify, "deployment-engine", "deployment_failed", map[string]interface{}{
			"type":          "deployment_failed",
			"application":   appName,
			"environment":   environment,
			"deployment_id": deploymentID,
			"error":         "Some services failed to deploy",
			"failed":        e.extractServiceNames(failed),
			"failures":      failed,
		})

		// Check if any failures are due to policy violations
		for _, failure := range failed {
			if reason, ok := failure["reason"].(string); ok {
				if strings.Contains(reason, "Policy not satisfied") || strings.Contains(reason, "policy validation failed") {
					return result, fmt.Errorf("deployment blocked by policy: %s", reason)
				}
			}
		}

		// If not policy-related, return generic deployment failure error
		return result, fmt.Errorf("deployment failed: %d services failed to deploy", len(failed))
	} else if len(deployments) > 0 {
		result.Status = "in_progress" // Resource Providers will complete asynchronously
	} else {
		result.Status = "completed" // Nothing to deploy (all skipped)
		events.GlobalEventBus.Emit(events.EventTypeNotify, "deployment-engine", "deployment_completed", map[string]interface{}{
			"type":          "deployment_completed",
			"application":   appName,
			"environment":   environment,
			"deployment_id": deploymentID,
			"deployments":   deployments,
			"all_skipped":   true,
		})
	}

	return result, nil
}

// getAllowedEnvironmentsForApp returns the environments this application is allowed to deploy to
func (e *Engine) getAllowedEnvironmentsForApp(appName string) []string {
	currentGraph, err := e.graph.Graph()
	if err != nil {
		return []string{} // Return empty if can't get graph
	}

	var allowedEnvs []string
	if edges, ok := currentGraph.Edges[appName]; ok {
		for _, edge := range edges {
			if edge.Type == "allowed_in" {
				allowedEnvs = append(allowedEnvs, edge.To)
			}
		}
	}
	return allowedEnvs
}

// generateDeploymentMessage creates user-friendly deployment status messages
func (e *Engine) generateDeploymentMessage(appName, environment string, deployed, failed int) string {
	if failed > 0 {
		return "Deployment of " + appName + " to " + environment + " initiated with " +
			strconv.Itoa(failed) + " failures"
	}
	if deployed == 0 {
		return "All services of " + appName + " were already deployed to " + environment
	}
	return "Successfully initiated deployment of " + appName + " to " + environment +
		" (Resource Providers will complete asynchronously)"
}

// updatePolicyChecks updates relevant policy checks after a successful deployment
func (e *Engine) updatePolicyChecks(appName, environment string) {
	currentGraph, err := e.graph.Graph()
	if err != nil {
		return // Can't update if we can't get graph
	}

	// Look for checks that might be satisfied by this deployment
	for _, node := range currentGraph.Nodes {
		if node.Kind == "check" {
			// Check if this is a deployment prerequisite check for this application
			if checkApp, ok := node.Spec["application"].(string); ok && checkApp == appName {
				if requiredEnv, ok := node.Spec["required_env"].(string); ok && requiredEnv == environment {
					// This deployment satisfies the check's requirement
					node.Metadata["status"] = "succeeded"
					fmt.Printf("[DEBUG] Updated check %s status to succeeded due to %s deployment to %s\n",
						node.ID, appName, environment)
				}
			}
		}
	}
}

// extractServiceNames extracts service names from failed deployments for event emission
func (e *Engine) extractServiceNames(failed []map[string]interface{}) []string {
	var services []string
	for _, failure := range failed {
		if service, ok := failure["service"].(string); ok {
			services = append(services, service)
		}
	}
	return services
}

// ==================================================================================
// RESOURCE PROVIDER INTEGRATION HOOKS
// These would be called by Resource Providers when they complete their work
// ==================================================================================

// HandleResourceProvisionCompleted is called when a Resource Provider completes provisioning
func (e *Engine) HandleResourceProvisionCompleted(resourceID, resourceType string, connectionInfo map[string]interface{}) {
	// Update graph with resource connection information
	currentGraph, err := e.graph.Graph()
	if err != nil {
		return // Can't update if we can't get graph
	}

	if node, exists := currentGraph.Nodes[resourceID]; exists {
		if node.Metadata == nil {
			node.Metadata = make(map[string]interface{})
		}
		node.Metadata["connection_info"] = connectionInfo
		node.Metadata["status"] = "provisioned"

		// Emit resource provision completed event
		events.GlobalEventBus.Emit(events.EventTypeNotify, "deployment-engine", "resource_provision_completed", map[string]interface{}{
			"type":            "resource_provision_completed",
			"resource_id":     resourceID,
			"resource_type":   resourceType,
			"provision_id":    uuid.New().String(), // provision ID would come from the RP
			"connection_info": connectionInfo,
		})
	}
}

// HandleServiceDeploymentCompleted is called when a Resource Provider completes service deployment
func (e *Engine) HandleServiceDeploymentCompleted(serviceID, environment string, metadata map[string]interface{}) {
	// Update graph with deployment information
	currentGraph, err := e.graph.Graph()
	if err != nil {
		return // Can't update if we can't get graph
	}

	if node, exists := currentGraph.Nodes[serviceID]; exists {
		if node.Metadata == nil {
			node.Metadata = make(map[string]interface{})
		}
		node.Metadata["deployment_status"] = "deployed"
		node.Metadata["deployment_environment"] = environment
		node.Metadata["deployment_metadata"] = metadata

		// This would be part of a larger orchestration to track when all services are complete
		// For now, we just update the individual service status
	}
}

// generateDeploymentPlan uses simplified AI planner - automatically discovers edges
func (e *Engine) generateDeploymentPlan(appName string) ([]string, error) {
	e.logger.Info("🧠 Generating simplified AI deployment plan for application: %s", appName)

	// Allow longer timeout for AI plan generation (configurable via environment)
	timeout := 2 * time.Minute
	if timeoutEnv := os.Getenv("ZTDP_AI_TIMEOUT"); timeoutEnv != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutEnv); err == nil {
			timeout = parsedTimeout
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use simplified planner - AI automatically discovers all dependencies
	return e.planner.Deploy(ctx, appName)
}
