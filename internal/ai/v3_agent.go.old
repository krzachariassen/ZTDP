package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// V3Agent - PURE AI-NATIVE ORCHESTRATOR with Agent Interface
// Philosophy: AI drives everything naturally, zero hardcoded domain logic
// This agent is completely domain-agnostic and routes by intent only
type V3Agent struct {
	provider AIProvider
	logger   *logging.Logger
	graph    *graph.GlobalGraph

	// Event-Driven Communication - ONLY dependencies for pure orchestration
	eventBus      *events.EventBus
	agentRegistry agents.AgentRegistry

	// Agent Interface Properties
	agentID   string
	startTime time.Time
}

// NewV3Agent creates the pure orchestrator agent with event-driven communication only
func NewV3Agent(
	provider AIProvider,
	globalGraph *graph.GlobalGraph,
	eventBus *events.EventBus,
	agentRegistry agents.AgentRegistry,
) *V3Agent {
	return &V3Agent{
		provider:      provider,
		logger:        logging.GetLogger().ForComponent("v3-agent"),
		graph:         globalGraph,
		eventBus:      eventBus,
		agentRegistry: agentRegistry,
	}
}

// Chat - THE ONLY METHOD! Pure ChatGPT-style conversation
func (agent *V3Agent) Chat(ctx context.Context, userMessage string) (*ConversationalResponse, error) {
	agent.logger.Info("🤖 V3 User: %s", userMessage)

	// FIRST: Check if user sent a contract directly (before AI processing)
	if contractResult := agent.detectDirectContract(ctx, userMessage); contractResult != nil {
		return contractResult, nil
	}

	// SECOND: Check if this is an operational intent (not resource creation)
	if intentResult, err := agent.checkForOperationalIntent(ctx, userMessage); intentResult != nil || err != nil {
		if err != nil {
			// Log error but continue with resource creation as fallback
			agent.logger.Warn("Intent detection failed, falling back to resource creation: %v", err)
		} else {
			return intentResult, nil
		}
	}

	// THIRD: Handle as resource creation conversation
	return agent.handleResourceCreationConversation(ctx, userMessage)
}

// checkForOperationalIntent detects and routes operational intents like policy checks, deployments, etc.
func (agent *V3Agent) checkForOperationalIntent(ctx context.Context, userMessage string) (*ConversationalResponse, error) {
	// Use AI to detect if this is an operational intent
	intentDetectionPrompt := `You are an intent classifier for a platform AI system.

Analyze the user message and determine if it's an OPERATIONAL INTENT or RESOURCE CREATION.

OPERATIONAL INTENTS (route to specialist agents):
- Application management: respond with "create application", "update application", or "list applications"
- Policy checking/evaluation: respond with "policy check" 
- Deployment operations: respond with "deploy application"
- Status monitoring: respond with "check status"

RESOURCE_CREATION (handle directly - only for complex multi-resource scenarios):
- Creating complex systems with multiple linked resources
- Multi-step resource setup requiring coordination

If OPERATIONAL INTENT detected, respond with:
INTENT: [exact intent from list above]

If RESOURCE CREATION, respond with:
RESOURCE_CREATION

Examples:
User: "Create an application called test-app" -> INTENT: create application
User: "Make a new application for my project" -> INTENT: create application
User: "List all applications" -> INTENT: list applications
User: "Do a policy check for checkout" -> INTENT: policy check
User: "Check compliance for my app" -> INTENT: policy check  
User: "Evaluate policies for my deployment" -> INTENT: policy check
User: "Deploy test-app to production" -> INTENT: deploy application
User: "Create a microservices architecture with API, database, and cache" -> RESOURCE_CREATION`

	response, err := agent.provider.CallAI(ctx, intentDetectionPrompt, userMessage)
	if err != nil {
		agent.logger.Error("Intent detection failed: %v", err)
		return nil, nil // Fall back to resource creation
	}

	// Check if AI detected an operational intent
	if strings.HasPrefix(response, "INTENT:") {
		intent := strings.TrimSpace(strings.TrimPrefix(response, "INTENT:"))
		agent.logger.Info("🎯 Detected operational intent: %s", intent)

		// Route to appropriate agent via intent-based orchestration
		result, err := agent.orchestrateViaIntentBasedAgents(ctx, intent, map[string]interface{}{
			"user_message": userMessage,
			"source":       "v3-agent-chat",
		})

		if err != nil {
			agent.logger.Error("Intent orchestration failed: %v", err)
			return &ConversationalResponse{
				Message: fmt.Sprintf("I understood you want to %s, but encountered an error: %v", intent, err),
				Answer:  fmt.Sprintf("I understood you want to %s, but encountered an error: %v", intent, err),
			}, nil
		}

		// Convert result to conversational response
		var responseMessage string
		if result != nil {
			// Check if this is an error result
			if resultMap, ok := result.(map[string]interface{}); ok {
				if status, exists := resultMap["status"].(string); exists && status == "error" {
					if responseContent, ok := resultMap["response_content"].(string); ok {
						responseMessage = responseContent
					} else {
						responseMessage = fmt.Sprintf("❌ %s request failed", intent)
					}
				} else if status, exists := resultMap["status"].(string); exists && status == "timeout" {
					intent := resultMap["intent"].(string)
					agentID := resultMap["selected_agent"].(string)
					responseMessage = fmt.Sprintf("I tried to %s but didn't get a response from the %s. This might be because the operation is taking longer than expected or the agent is busy. Please try again in a moment.", intent, agentID)
				} else if responseContent, ok := resultMap["response_content"].(string); ok {
					responseMessage = responseContent
				} else {
					responseMessage = fmt.Sprintf("✅ Successfully handled %s request", intent)
				}
			} else {
				responseMessage = fmt.Sprintf("✅ Successfully handled %s request", intent)
			}
		} else {
			responseMessage = fmt.Sprintf("✅ Successfully handled %s request", intent)
		}

		return &ConversationalResponse{
			Message: responseMessage,
			Answer:  responseMessage,
			Intent:  intent,
		}, nil
	}

	// Not an operational intent, let resource creation handle it
	return nil, nil
}

// handleResourceCreationConversation handles resource creation with the original logic
func (agent *V3Agent) handleResourceCreationConversation(ctx context.Context, userMessage string) (*ConversationalResponse, error) {
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
				result, err := agent.executeContract(ctx, cleanJSON, userMessage)
				if err != nil {
					agent.logger.Error("❌ Contract execution failed: %v", err)
					return &ConversationalResponse{
						Message: fmt.Sprintf("❌ %v", err),
						Actions: []Action{{Type: "error", Result: err.Error()}},
					}, nil
				}

				// Check if the result indicates an error from the underlying agent
				if resultMap, ok := result.(map[string]interface{}); ok {
					if status, ok := resultMap["status"].(string); ok && status == "error" {
						// Agent returned an error - use the error message directly
						if responseContent, ok := resultMap["response_content"].(string); ok {
							return &ConversationalResponse{
								Message: responseContent,
								Actions: []Action{{Type: "error", Result: resultMap}},
							}, nil
						}
						// Fallback error message
						return &ConversationalResponse{
							Message: "❌ The operation failed. Please check the application exists and try again.",
							Actions: []Action{{Type: "error", Result: resultMap}},
						}, nil
					}
				}

				// Success case - remove the FINAL_CONTRACT part from the message
				cleanMessage := strings.TrimSpace(aiResponse[:contractStart])
				return &ConversationalResponse{
					Message: cleanMessage + "\n\n✅ Resource created successfully!",
					Actions: []Action{{Type: "resource_created", Result: result}},
				}, nil
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

// executeContract executes a contract by extracting intent and routing to appropriate agents
func (agent *V3Agent) executeContract(ctx context.Context, contractJSON string, userMessage string) (interface{}, error) {
	var contractData map[string]interface{}
	if err := json.Unmarshal([]byte(contractJSON), &contractData); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// PURE ORCHESTRATOR: Extract intent using AI, no hardcoded domain knowledge
	intent, err := agent.extractIntentFromContract(ctx, contractData, userMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to extract intent: %w", err)
	}

	agent.logger.Info("🚀 Extracted intent '%s' from contract, orchestrating via agent discovery", intent)

	// Route purely by intent - completely domain-agnostic!
	return agent.orchestrateViaIntentBasedAgents(ctx, intent, contractData)
}

// extractIntentFromContract uses AI to determine the intent from contract data
func (agent *V3Agent) extractIntentFromContract(ctx context.Context, contractData map[string]interface{}, userMessage string) (string, error) {
	// Use AI to understand the intent from the contract and user message
	systemPrompt := `You are an intent extraction AI. Based on the user's message and the contract data, determine the specific intent.

Respond with ONLY the intent phrase, nothing else. Examples:
- "create application"
- "deploy application" 
- "create service"
- "create database resource"
- "validate policies"
- "setup environment"

The intent should be specific enough for agent discovery but generic enough to be domain-agnostic.`

	userPrompt := fmt.Sprintf(`User said: "%s"

Contract data: %s

What is the intent?`, userMessage, string(mustMarshal(contractData)))

	response, err := agent.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("AI intent extraction failed: %w", err)
	}

	// Clean up the response to get just the intent
	intent := strings.TrimSpace(strings.ToLower(response))
	intent = strings.Trim(intent, `"'`)

	return intent, nil
}

// mustMarshal helper function
func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
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

// orchestrateViaIntentBasedAgents - PURE ORCHESTRATOR: Discovers agents by intent and routes events
// This method contains NO domain-specific logic - it's completely generic!
func (agent *V3Agent) orchestrateViaIntentBasedAgents(ctx context.Context, intent string, context map[string]interface{}) (interface{}, error) {
	if agent.agentRegistry == nil {
		return nil, fmt.Errorf("agent registry not available - cannot discover agents")
	}

	agent.logger.Info("🔍 Discovering agents for intent: %s", intent)

	// STEP 1: Discover agents by intent (completely generic)
	availableAgents, err := agent.discoverAgentsByIntent(ctx, intent)
	if err != nil {
		return nil, fmt.Errorf("agent discovery failed for intent '%s': %w", intent, err)
	}

	if len(availableAgents) == 0 {
		return nil, fmt.Errorf("no agents found for intent '%s' - register appropriate agents first", intent)
	}

	agent.logger.Info("🎯 Found %d agents capable of handling intent: %s", len(availableAgents), intent)

	// STEP 2: Route to the best agent and get routing key
	selectedAgent := availableAgents[0] // Simple: use first available agent

	// STEP 2.5: Discover the appropriate routing key for this intent
	routingKey, err := agent.discoverRoutingKeyForIntent(ctx, intent, selectedAgent.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover routing key for intent '%s' and agent '%s': %w", intent, selectedAgent.ID, err)
	}

	agent.logger.Info("🔑 Using routing key '%s' for agent: %s", routingKey, selectedAgent.ID)

	// STEP 3: Create request-response correlation
	correlationID := fmt.Sprintf("orchestration-%d", time.Now().UnixNano())
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Create a channel to receive the response
	responseChan := make(chan *events.Event, 1)

	// Subscribe to response events for this correlation ID
	agent.eventBus.Subscribe(events.EventTypeResponse, func(event events.Event) error {
		// Check if this response is for our request
		if responseCorrelationID, ok := event.Payload["correlation_id"].(string); ok {
			if responseCorrelationID == correlationID {
				// This is our response!
				select {
				case responseChan <- &event:
					agent.logger.Info("📨 Received response for correlation ID: %s", correlationID)
				default:
					agent.logger.Warn("Response channel full for correlation ID: %s", correlationID)
				}
			}
		}
		return nil
	})

	// STEP 4: Emit targeted event using discovered routing key
	eventPayload := map[string]interface{}{
		"correlation_id": correlationID,
		"intent":         intent,
		"context":        context,
		"request_id":     requestID,
		"source_agent":   "v3-agent",
	}

	// Targeted event emission using specific routing key for this agent
	if err := agent.eventBus.Emit(events.EventTypeRequest, "v3-agent", routingKey, eventPayload); err != nil {
		return nil, fmt.Errorf("failed to emit intent request to routing key %s for agent %s: %w", routingKey, selectedAgent.ID, err)
	}

	agent.logger.Info("📤 Routed intent '%s' to agent: %s via routing key: %s", intent, selectedAgent.ID, routingKey)

	// STEP 5: Wait for response with timeout
	select {
	case response := <-responseChan:
		agent.logger.Info("✅ Received response from agent for intent: %s", intent)

		// Extract meaningful content from the agent response and check for errors
		var responseContent string
		var responseStatus string = "completed"

		// First, check if this is an error response
		if status, ok := response.Payload["status"].(string); ok && status == "error" {
			responseStatus = "error"
			if errorMsg, ok := response.Payload["error"].(string); ok {
				responseContent = fmt.Sprintf("❌ %s", errorMsg)
			} else {
				responseContent = fmt.Sprintf("❌ Agent reported an error for %s request", intent)
			}
		} else if decision, ok := response.Payload["decision"].(string); ok {
			if reasoning, ok := response.Payload["reasoning"].(string); ok {
				responseContent = fmt.Sprintf("Decision: %s. Reasoning: %s", decision, reasoning)
			} else {
				responseContent = fmt.Sprintf("Decision: %s", decision)
			}
		} else if message, ok := response.Payload["message"].(string); ok {
			responseContent = message
		} else {
			responseContent = fmt.Sprintf("✅ Agent completed the %s request successfully", intent)
		}

		return map[string]interface{}{
			"status":           responseStatus,
			"intent":           intent,
			"selected_agent":   response.Source,
			"response_content": responseContent,
			"agent_response":   response.Payload,
		}, nil
	case <-time.After(30 * time.Second): // 30 second timeout for AI operations
		agent.logger.Warn("⏰ Timeout waiting for response from agent for intent: %s", intent)
		return map[string]interface{}{
			"status":         "timeout",
			"intent":         intent,
			"selected_agent": selectedAgent.ID,
			"correlation_id": correlationID,
			"message":        fmt.Sprintf("Intent '%s' sent to agent %s but no response received within timeout", intent, selectedAgent.ID),
		}, nil
	}
}

// discoverAgentsByIntent - Generic agent discovery by matching intent to capabilities
func (agent *V3Agent) discoverAgentsByIntent(ctx context.Context, intent string) ([]agents.AgentStatus, error) {
	var matchingAgents []agents.AgentStatus

	// Get all available capabilities to find routing keys
	capabilities, err := agent.agentRegistry.GetAvailableCapabilities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available capabilities: %w", err)
	}

	// Find capabilities that match the intent
	for _, capability := range capabilities {
		// Check if this capability matches the intent
		for _, supportedIntent := range capability.Intents {
			if agent.intentMatches(intent, supportedIntent) {
				// Find agents with this capability
				agentsWithCapability, err := agent.agentRegistry.FindAgentsByCapability(ctx, capability.Name)
				if err != nil {
					agent.logger.Warn("⚠️ Failed to find agents for capability %s: %v", capability.Name, err)
					continue
				}
				matchingAgents = append(matchingAgents, agentsWithCapability...)
				break // Found a match, no need to check other intents for this capability
			}
		}
	}

	// Remove duplicates and exclude self (V3Agent should not route to itself during orchestration)
	deduplicated := agent.deduplicate(matchingAgents)
	return agent.excludeSelf(deduplicated), nil
}

// intentMatches - Simple intent matching (can be enhanced with AI/NLP)
func (agent *V3Agent) intentMatches(userIntent, supportedIntent string) bool {
	// Simple contains check - could be enhanced with semantic matching
	userWords := strings.ToLower(userIntent)
	supportedWords := strings.ToLower(supportedIntent)

	// Check if user intent contains key words from supported intent
	supportedKeywords := strings.Fields(supportedWords)
	for _, keyword := range supportedKeywords {
		if len(keyword) > 3 && strings.Contains(userWords, keyword) { // Only match meaningful words
			return true
		}
	}

	return false
}

// deduplicate - Remove duplicate agents from the list
func (agent *V3Agent) deduplicate(agentsList []agents.AgentStatus) []agents.AgentStatus {
	seen := make(map[string]bool)
	var result []agents.AgentStatus

	for _, a := range agentsList {
		if !seen[a.ID] {
			seen[a.ID] = true
			result = append(result, a)
		}
	}

	return result
}

// excludeSelf - Remove the V3Agent itself from agent selection during orchestration
func (agent *V3Agent) excludeSelf(agentsList []agents.AgentStatus) []agents.AgentStatus {
	var result []agents.AgentStatus

	for _, a := range agentsList {
		if a.ID != "v3-agent" { // V3Agent should not route to itself during orchestration
			result = append(result, a)
		}
	}

	return result
}

// discoverRoutingKeyForIntent finds the appropriate routing key for an intent and agent
func (agent *V3Agent) discoverRoutingKeyForIntent(ctx context.Context, intent string, agentID string) (string, error) {
	// Get all available capabilities to find routing keys
	capabilities, err := agent.agentRegistry.GetAvailableCapabilities(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get available capabilities: %w", err)
	}

	// Find the capability that matches both the intent and the agent
	for _, capability := range capabilities {
		// Check if this capability matches the intent
		for _, supportedIntent := range capability.Intents {
			if agent.intentMatches(intent, supportedIntent) {
				// Check if this capability belongs to the target agent
				agentsWithCapability, err := agent.agentRegistry.FindAgentsByCapability(ctx, capability.Name)
				if err != nil {
					continue
				}

				for _, agentStatus := range agentsWithCapability {
					if agentStatus.ID == agentID {
						// Found the right capability for this agent and intent
						if len(capability.RoutingKeys) > 0 {
							// Return the first routing key for this capability
							return capability.RoutingKeys[0], nil
						}
					}
				}
			}
		}
	}

	// Fallback to a default routing key based on intent type
	if strings.Contains(strings.ToLower(intent), "policy") {
		return "policy.request", nil
	} else if strings.Contains(strings.ToLower(intent), "deploy") {
		return "deployment.request", nil
	}

	// Ultimate fallback
	return "agent.request", nil
}

// detectDirectContract checks if user input contains a valid contract and executes it directly
func (agent *V3Agent) detectDirectContract(ctx context.Context, userMessage string) *ConversationalResponse {
	// Look for JSON-like patterns in the user message
	userMessage = strings.TrimSpace(userMessage)

	// Check if message starts with '{' or contains contract keywords
	if !strings.Contains(userMessage, "{") || !strings.Contains(userMessage, `"kind"`) {
		return nil // Not a direct contract
	}

	// Try to extract JSON from the message
	startIdx := strings.Index(userMessage, "{")
	if startIdx == -1 {
		return nil
	}

	// Find the matching closing brace
	braceCount := 0
	endIdx := -1
	for i := startIdx; i < len(userMessage); i++ {
		switch userMessage[i] {
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

	if endIdx == -1 {
		return nil // No complete JSON found
	}

	contractJSON := userMessage[startIdx:endIdx]
	agent.logger.Info("🎯 Detected direct contract in user input: %s", contractJSON)

	// Validate it's a proper contract
	var contractData map[string]interface{}
	if err := json.Unmarshal([]byte(contractJSON), &contractData); err != nil {
		return nil // Not valid JSON
	}

	// Check if it has the required 'kind' field
	if _, hasKind := contractData["kind"].(string); !hasKind {
		return nil // Not a contract
	}

	// Execute the contract directly via intent-based orchestration
	agent.logger.Info("🚀 Executing direct contract via intent-based orchestration")
	result, err := agent.executeContract(ctx, contractJSON, userMessage)
	if err != nil {
		return &ConversationalResponse{
			Message: fmt.Sprintf("I found a contract in your message but couldn't execute it: %v", err),
			Actions: []Action{{Type: "error", Result: err.Error()}},
		}
	}

	return &ConversationalResponse{
		Message: fmt.Sprintf("✅ Contract executed successfully via intent-based orchestration: %v", result),
		Actions: []Action{{Type: "contract_executed", Result: result}},
	}
}

// ============================================================================
// AgentInterface Implementation - Making V3Agent a first-class agent
// ============================================================================

// GetID returns the agent identifier
func (agent *V3Agent) GetID() string {
	if agent.agentID == "" {
		agent.agentID = "v3-agent"
	}
	return agent.agentID
}

// GetStatus returns the current agent status
func (agent *V3Agent) GetStatus() agents.AgentStatus {
	return agents.AgentStatus{
		ID:           agent.GetID(),
		Type:         "orchestrator",
		Status:       "running",
		LastActivity: time.Now(),
		LoadFactor:   0.5,
		Version:      "3.0.0",
		Metadata: map[string]interface{}{
			"ai_provider":  "openai",
			"role":         "orchestrator",
			"capabilities": "intent-based routing, resource creation, operational coordination",
		},
	}
}

// GetCapabilities returns the agent's capabilities for discovery
func (agent *V3Agent) GetCapabilities() []agents.AgentCapability {
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
			RoutingKeys: []string{"v3.chat", "v3.orchestrate", "v3.general"},
			Version:     "3.0.0",
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
			RoutingKeys: []string{"v3.create", "v3.resource", "v3.build"},
			Version:     "3.0.0",
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
			RoutingKeys: []string{"v3.route", "v3.delegate", "v3.coordinate"},
			Version:     "3.0.0",
		},
	}
}

// Start initializes the V3Agent as a registered agent
func (agent *V3Agent) Start(ctx context.Context) error {
	agent.startTime = time.Now()
	agent.agentID = "v3-agent"

	// Auto-register with the agent registry
	if agent.agentRegistry != nil {
		if err := agent.agentRegistry.RegisterAgent(ctx, agent); err != nil {
			agent.logger.Error("❌ Failed to auto-register V3Agent: %v", err)
			return fmt.Errorf("failed to auto-register V3Agent: %w", err)
		}
		agent.logger.Info("✅ V3Agent auto-registered successfully")
	}

	// Subscribe to V3Agent routing keys
	if agent.eventBus != nil {
		agent.subscribeToOwnRoutingKeys()
	}

	agent.logger.Info("🚀 V3Agent started as first-class agent")
	return nil
}

// Stop shuts down the V3Agent
func (agent *V3Agent) Stop(ctx context.Context) error {
	agent.logger.Info("🛑 V3Agent stopping")
	return nil
}

// Health returns the health status of the agent
func (agent *V3Agent) Health() agents.HealthStatus {
	return agents.HealthStatus{
		Healthy: true,
		Status:  "healthy",
		Message: "V3Agent operating normally",
		Checks: map[string]interface{}{
			"ai_provider_available": agent.provider != nil,
			"event_bus_connected":   agent.eventBus != nil,
			"registry_connected":    agent.agentRegistry != nil,
		},
		CheckedAt: time.Now(),
	}
}

// ProcessEvent handles events sent directly to V3Agent
func (agent *V3Agent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
	agent.logger.Info("📨 V3Agent received event: %s from %s", event.Subject, event.Source)

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
		return agent.createErrorResponse(event, "user_message required for V3Agent processing"), nil
	}

	// Process the message using the Chat method
	response, err := agent.Chat(ctx, userMessage)
	if err != nil {
		return agent.createErrorResponse(event, fmt.Sprintf("V3Agent processing failed: %v", err)), nil
	}

	// Create success response
	return agent.createSuccessResponse(event, response), nil
}

// subscribeToOwnRoutingKeys sets up event subscriptions for V3Agent's routing keys
func (agent *V3Agent) subscribeToOwnRoutingKeys() {
	routingKeys := []string{"v3.chat", "v3.orchestrate", "v3.general", "v3.create", "v3.resource", "v3.build", "v3.route", "v3.delegate", "v3.coordinate"}

	for _, routingKey := range routingKeys {
		agent.eventBus.SubscribeToRoutingKey(routingKey, func(event events.Event) error {
			agent.logger.Info("📨 V3Agent received event via routing key %s: %s", routingKey, event.Subject)

			// Process the event
			ctx := context.Background()
			response, err := agent.ProcessEvent(ctx, &event)
			if err != nil {
				agent.logger.Error("❌ Failed to process event: %v", err)
				return err
			}

			// Send response back
			if response != nil && agent.eventBus != nil {
				agent.eventBus.EmitEvent(*response)
			}

			return nil
		})
	}

	agent.logger.Info("✅ V3Agent subscribed to %d routing keys", len(routingKeys))
}

// createErrorResponse creates a standardized error response
func (agent *V3Agent) createErrorResponse(originalEvent *events.Event, errorMessage string) *events.Event {
	response := &events.Event{
		Type:    events.EventTypeResponse,
		Source:  agent.GetID(),
		Subject: "V3Agent processing failed",
		Payload: map[string]interface{}{
			"status":  "error",
			"error":   errorMessage,
			"context": "v3-agent",
		},
		Timestamp: time.Now().Unix(),
		ID:        fmt.Sprintf("v3-resp-%d", time.Now().UnixNano()),
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
func (agent *V3Agent) createSuccessResponse(originalEvent *events.Event, chatResponse *ConversationalResponse) *events.Event {
	response := &events.Event{
		Type:    events.EventTypeResponse,
		Source:  agent.GetID(),
		Subject: "V3Agent processing completed",
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
		ID:        fmt.Sprintf("v3-resp-%d", time.Now().UnixNano()),
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
