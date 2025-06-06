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

	// Get contracts dynamically
	contracts := agent.loadAllContracts()
	
	// Get platform state
	state := agent.getPlatformState()

	// Pure natural conversation like the ChatGPT example
	systemPrompt := fmt.Sprintf(`You are a platform AI that helps users create and manage resources using contracts.

AVAILABLE CONTRACTS:
%s

CURRENT PLATFORM STATE:
%s

INSTRUCTIONS:
- Drive the conversation naturally to understand what the user wants
- For resource creation, guide them through the contract fields step by step
- Ask clarifying questions to collect missing information
- When you have a complete valid contract, output the final JSON with "FINAL_CONTRACT:" prefix
- For listing/info operations, provide the information directly
- For updates, help modify existing contracts
- Be conversational and helpful

Remember: Drive the conversation naturally. Just be helpful and guide them to complete contracts.`, contracts, state)

	response, err := agent.provider.CallAI(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	return agent.handleResponse(ctx, response)
}

// handleResponse processes AI's natural response and executes if contract is ready
func (agent *V3Agent) handleResponse(ctx context.Context, aiResponse string) (*ConversationalResponse, error) {
	// Check if AI provided a final contract to execute
	if contractStart := strings.Index(aiResponse, "FINAL_CONTRACT:"); contractStart != -1 {
		// Extract the JSON contract
		jsonPart := strings.TrimSpace(aiResponse[contractStart+len("FINAL_CONTRACT:"):])
		
		// Try to execute the contract
		if result, err := agent.executeContract(ctx, jsonPart); err == nil {
			// Remove the FINAL_CONTRACT part from the message
			cleanMessage := strings.TrimSpace(aiResponse[:contractStart])
			return &ConversationalResponse{
				Message: cleanMessage + "\n\n✅ Resource created successfully!",
				Actions: []Action{{Type: "resource_created", Result: result}},
			}, nil
		} else {
			return &ConversationalResponse{
				Message: fmt.Sprintf("I created the contract but couldn't execute it: %v", err),
				Actions: []Action{{Type: "error", Result: err.Error()}},
			}, nil
		}
	}
	
	// For all other responses, just return the AI's natural response
	return &ConversationalResponse{
		Message: aiResponse,
		Actions: []Action{{Type: "conversation_continue", Result: "ai_response"}},
	}, nil
}

// executeContract executes a contract JSON using the appropriate service
func (agent *V3Agent) executeContract(ctx context.Context, contractJSON string) (interface{}, error) {
	var contractData map[string]interface{}
	if err := json.Unmarshal([]byte(contractJSON), &contractData); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	resourceType := agent.detectResourceType(contractData)
	
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

// detectResourceType determines resource type from contract structure
func (agent *V3Agent) detectResourceType(contractData map[string]interface{}) string {
	// Look for metadata.name and spec structure to determine type
	if spec, ok := contractData["spec"].(map[string]interface{}); ok {
		// Check spec fields to determine type
		if _, hasLifecycle := spec["lifecycle"]; hasLifecycle {
			return "application"
		}
		if _, hasEnvironments := spec["environments"]; hasEnvironments {
			return "service"
		}
		if _, hasKind := contractData["kind"]; hasKind {
			return "resource"
		}
		// Simple environment if only has description
		if _, hasDescription := spec["description"]; hasDescription && len(spec) == 1 {
			return "environment"
		}
	}
	return ""
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

// getPlatformState gets current platform state
func (agent *V3Agent) getPlatformState() string {
	if agent.graph == nil {
		return "Platform state: Not available"
	}
	
	// Get the current graph
	currentGraph, err := agent.graph.Graph()
	if err != nil {
		return "Platform state: Error loading graph"
	}
	
	return fmt.Sprintf(`
Platform State:
- Total nodes: %d
- Applications: %d
- Services: %d  
- Environments: %d
- Resources: %d
`, len(currentGraph.Nodes), 
	agent.countNodesByKind(currentGraph.Nodes, "application"),
	agent.countNodesByKind(currentGraph.Nodes, "service"),
	agent.countNodesByKind(currentGraph.Nodes, "environment"),
	agent.countNodesByKind(currentGraph.Nodes, "resource"))
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
