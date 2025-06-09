package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
	"github.com/krzachariassen/ZTDP/internal/resources"
)

// V3Agent - ULTRA simple AI-native agent inspired by ChatGPT example
// Philosophy: AI drives everything naturally, zero hardcoded logic
type V3Agent struct {
	provider AIProvider
	logger   *logging.Logger
	graph    *graph.GlobalGraph

	// Use actual service interfaces with correct method signatures
	applicationService ApplicationService
	serviceService     ServiceService
	resourceService    ResourceService
	environmentService EnvironmentService
	deploymentService  DeploymentService
	policyService      PolicyService
}

// NewV3Agent creates the ultra simple agent
func NewV3Agent(
	provider AIProvider,
	globalGraph *graph.GlobalGraph,
	applicationService ApplicationService,
	serviceService ServiceService,
	resourceService ResourceService,
	environmentService EnvironmentService,
	deploymentService DeploymentService,
	policyService PolicyService,
) *V3Agent {
	return &V3Agent{
		provider:           provider,
		logger:             logging.GetLogger().ForComponent("v3-agent"),
		graph:              globalGraph,
		applicationService: applicationService,
		serviceService:     serviceService,
		resourceService:    resourceService,
		environmentService: environmentService,
		deploymentService:  deploymentService,
		policyService:      policyService,
	}
}

// Chat - THE ONLY METHOD! Pure ChatGPT-style conversation
func (agent *V3Agent) Chat(ctx context.Context, userMessage string) (*ConversationalResponse, error) {
	agent.logger.Info("🤖 V3 User: %s", userMessage)

	// Get platform state
	state := agent.getPlatformState()

	// Get contract schemas to understand what's required vs optional
	contractSchemas := agent.loadAllContracts()
	// Simple, natural conversation - let AI be AI
	systemPrompt := fmt.Sprintf(`You are a platform AI assistant that creates resources through natural conversation.

CURRENT PLATFORM STATE:
%s

AVAILABLE CONTRACTS:
%s

CRITICAL: When users ask to create something, DO IT with smart defaults instead of asking for more details.

RESOURCE TYPES AND ARCHITECTURE:
- "application" - Container/grouping for related services (like a project boundary)
- "service" - Actual running code: APIs, consumers, workers, microservices  
- "resource" - Infrastructure: databases, storage, queues, caches
- "environment" - Deployment targets: dev, staging, prod

HIERARCHY AND LINKING:
Applications contain Services. Services use Resources. 
Services MUST have "app" field in metadata to link to parent application.

COMPLEX EXAMPLE:
User: "Create an application with an API that receives recipes and stores them in a database"
Should create:
1. Application container: {"kind":"application","metadata":{"name":"recipe-app"}}
2. API service: {"kind":"service","metadata":{"name":"recipe-api","app":"recipe-app"},"spec":{"type":"api"}}
3. Database resource: {"kind":"resource","metadata":{"name":"recipe-db","type":"database"}}

The service links to application via "app" field, and can be linked to resources via graph edges.

When creating something, respond naturally and include FINAL_CONTRACT for EACH resource needed.
For complex requests, create multiple contracts in sequence.

ACT with smart defaults. Only ask questions if truly ambiguous.`, state, contractSchemas)

	response, err := agent.provider.CallAI(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	agent.logger.Info("🤖 AI Response: %s", response)
	return agent.handleResponse(ctx, response, userMessage)
}

// handleResponse processes AI's natural response and executes if contract is ready
func (agent *V3Agent) handleResponse(ctx context.Context, aiResponse string, userMessage string) (*ConversationalResponse, error) {
	// Check if AI provided a final contract to execute
	if contractStart := strings.Index(aiResponse, "FINAL_CONTRACT:"); contractStart != -1 {
		// Extract everything after FINAL_CONTRACT:
		jsonPart := strings.TrimSpace(aiResponse[contractStart+len("FINAL_CONTRACT:"):])

		agent.logger.Info("🔍 Raw JSON part: %q", jsonPart)

		// Try to extract just the JSON object by finding the first { and matching }
		if startIdx := strings.Index(jsonPart, "{"); startIdx != -1 {
			// Find the matching closing brace
			braceCount := 0
			endIdx := -1
			for i := startIdx; i < len(jsonPart); i++ {
				switch jsonPart[i] {
				case '{':
					braceCount++
				case '}':
					braceCount--
					if braceCount == 0 {
						endIdx = i + 1
						break
					}
				}
			}

			if endIdx != -1 {
				cleanJSON := jsonPart[startIdx:endIdx]
				agent.logger.Info("🔍 Extracted JSON: %q", cleanJSON)

				// Try to execute the contract with user context
				if result, err := agent.executeContract(ctx, cleanJSON, userMessage); err == nil {
					// Remove the FINAL_CONTRACT part from the message
					cleanMessage := strings.TrimSpace(aiResponse[:contractStart])
					return &ConversationalResponse{
						Message: cleanMessage + "\n\n✅ Resource created successfully!",
						Actions: []Action{{Type: "resource_created", Result: result}},
					}, nil
				} else {
					agent.logger.Error("❌ Contract execution failed: %v", err)
					return &ConversationalResponse{
						Message: fmt.Sprintf("I created the contract but couldn't execute it: %v", err),
						Actions: []Action{{Type: "error", Result: err.Error()}},
					}, nil
				}
			}
		}

		// Fallback: if we couldn't extract clean JSON, return error
		return &ConversationalResponse{
			Message: "I tried to create a contract but couldn't parse the JSON properly.",
			Actions: []Action{{Type: "error", Result: "JSON parsing failed"}},
		}, nil
	}

	// For all other responses, just return the AI's natural response
	return &ConversationalResponse{
		Message: aiResponse,
		Actions: []Action{{Type: "conversation_continue", Result: "ai_response"}},
	}, nil
}

// executeContract executes a contract JSON using the appropriate service
func (agent *V3Agent) executeContract(ctx context.Context, contractJSON string, userMessage string) (interface{}, error) {
	var contractData map[string]interface{}
	if err := json.Unmarshal([]byte(contractJSON), &contractData); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Trust AI to specify resource type directly via "kind" field
	resourceType, ok := contractData["kind"].(string)
	if !ok {
		// If AI didn't specify kind, default to application (most common case)
		resourceType = "application"
		agent.logger.Info("🤖 No 'kind' specified, defaulting to application")
	}

	agent.logger.Info("🎯 AI specified resource type: %s", resourceType)

	switch resourceType {
	case "application":
		// Convert to ApplicationContract
		var appContract contracts.ApplicationContract
		if err := json.Unmarshal([]byte(contractJSON), &appContract); err != nil {
			return nil, fmt.Errorf("invalid application contract: %w", err)
		}

		// Use correct method signature: CreateApplication(contracts.ApplicationContract) error
		if err := agent.applicationService.CreateApplication(appContract); err != nil {
			return nil, err
		}

		return appContract, nil

	case "environment":
		// Use CreateEnvironmentFromData method which exists
		result, err := agent.environmentService.CreateEnvironmentFromData(contractData)
		return result, err

	case "service":
		// Extract app name and service data for CreateService
		if metadata, ok := contractData["metadata"].(map[string]interface{}); ok {
			if appName, ok := metadata["app"].(string); ok {
				// Use CreateService(appName string, serviceData map[string]interface{}) method
				result, err := agent.serviceService.CreateService(appName, contractData)
				return result, err
			}
		}
		return nil, fmt.Errorf("service contract missing app name in metadata")

	case "resource":
		// Convert to ResourceRequest for CreateResource
		var resourceReq resources.ResourceRequest
		if err := json.Unmarshal([]byte(contractJSON), &resourceReq); err != nil {
			return nil, fmt.Errorf("invalid resource contract: %w", err)
		}

		// Use CreateResource(resources.ResourceRequest) method
		result, err := agent.resourceService.CreateResource(resourceReq)
		return result, err

	default:
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

// loadAllContracts dynamically loads all contract definitions
func (agent *V3Agent) loadAllContracts() string {
	contractsDir := "/mnt/c/Work/git/ztdp/internal/contracts"

	contracts := ""
	contractFiles := []string{"application.go", "service.go", "environment.go", "resource.go"}

	for _, file := range contractFiles {
		if content, err := os.ReadFile(filepath.Join(contractsDir, file)); err == nil {
			contracts += fmt.Sprintf("\n// %s\n%s\n", file, string(content))
		}
	}

	return contracts
}

// getPlatformState gets current platform state with detailed information
func (agent *V3Agent) getPlatformState() string {
	if agent.graph == nil {
		return "Platform state: Not available"
	}

	// Get the current graph
	currentGraph, err := agent.graph.Graph()
	if err != nil {
		return "Platform state: Error loading graph"
	}

	// Get detailed lists
	applications := agent.getNodesByKind(currentGraph.Nodes, "application")
	services := agent.getNodesByKind(currentGraph.Nodes, "service")
	environments := agent.getNodesByKind(currentGraph.Nodes, "environment")
	resources := agent.getNodesByKind(currentGraph.Nodes, "resource")

	state := fmt.Sprintf(`Platform State:
- Total nodes: %d

APPLICATIONS (%d):`, len(currentGraph.Nodes), len(applications))

	if len(applications) == 0 {
		state += "\n  (No applications created yet)"
	} else {
		for _, app := range applications {
			name := agent.getNodeName(app)
			state += fmt.Sprintf("\n  - %s", name)
		}
	}

	state += fmt.Sprintf("\n\nSERVICES (%d):", len(services))
	if len(services) == 0 {
		state += "\n  (No services created yet)"
	} else {
		for _, service := range services {
			name := agent.getNodeName(service)
			state += fmt.Sprintf("\n  - %s", name)
		}
	}

	state += fmt.Sprintf("\n\nENVIRONMENTS (%d):", len(environments))
	if len(environments) == 0 {
		state += "\n  (No environments created yet)"
	} else {
		for _, env := range environments {
			name := agent.getNodeName(env)
			state += fmt.Sprintf("\n  - %s", name)
		}
	}

	state += fmt.Sprintf("\n\nRESOURCES (%d):", len(resources))
	if len(resources) == 0 {
		state += "\n  (No resources created yet)"
	} else {
		for _, resource := range resources {
			name := agent.getNodeName(resource)
			state += fmt.Sprintf("\n  - %s", name)
		}
	}

	return state
}

// getNodeName extracts the name from a node's metadata
func (agent *V3Agent) getNodeName(node *graph.Node) string {
	if node.Metadata != nil {
		if name, ok := node.Metadata["name"].(string); ok {
			return name
		}
	}
	return node.ID // fallback to ID if no name found
}

// getNodesByKind returns all nodes of a specific kind
func (agent *V3Agent) getNodesByKind(nodes map[string]*graph.Node, kind string) []*graph.Node {
	var result []*graph.Node
	for _, node := range nodes {
		if node.Kind == kind {
			result = append(result, node)
		}
	}
	return result
}

// countNodesByKind counts nodes of a specific kind
func (agent *V3Agent) countNodesByKind(nodes map[string]*graph.Node, kind string) int {
	count := 0
	for _, node := range nodes {
		if node.Kind == kind {
			count++
		}
	}
	return count
}

// Compatibility methods for existing code

// GetProviderInfo returns provider info for compatibility
func (agent *V3Agent) GetProviderInfo() *ProviderInfo {
	if agent.provider == nil {
		return &ProviderInfo{
			Name:         "V3 Agent (No Provider)",
			Version:      "3.0.0",
			Capabilities: []string{"chat"},
		}
	}

	return &ProviderInfo{
		Name:         "V3 Agent with AI",
		Version:      "3.0.0",
		Capabilities: []string{"chat", "create", "update", "list", "deploy"},
	}
}

// Provider returns the underlying AI provider for compatibility
func (agent *V3Agent) Provider() AIProvider {
	return agent.provider
}

// ChatWithPlatform provides compatibility with v1 endpoint
func (agent *V3Agent) ChatWithPlatform(ctx context.Context, query string, context string) (*ConversationalResponse, error) {
	return agent.Chat(ctx, query)
}
