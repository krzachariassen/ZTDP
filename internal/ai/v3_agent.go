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

// V3Agent - PURE AI-NATIVE ORCHESTRATOR
// Philosophy: AI drives everything naturally, zero hardcoded domain logic
// This agent is completely domain-agnostic and routes by intent only
type V3Agent struct {
	provider AIProvider
	logger   *logging.Logger
	graph    *graph.GlobalGraph

	// Event-Driven Communication - ONLY dependencies for pure orchestration
	eventBus      *events.EventBus
	agentRegistry agents.AgentRegistry
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
- Policy checking/evaluation: respond with "policy check" 
- Deployment operations: respond with "deploy application"
- Status monitoring: respond with "check status"

RESOURCE CREATION (handle directly):
- Creating applications, services, resources, environments
- "create", "make", "build", "setup", "configure"

If OPERATIONAL INTENT detected, respond with:
INTENT: [exact intent from list above]

If RESOURCE CREATION, respond with:
RESOURCE_CREATION

Examples:
User: "Do a policy check for checkout" -> INTENT: policy check
User: "Check compliance for my app" -> INTENT: policy check  
User: "Evaluate policies for my deployment" -> INTENT: policy check
User: "Deploy test-app to production" -> INTENT: deploy application
User: "Create an API service" -> RESOURCE_CREATION`

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
			// Check if this is a timeout result
			if resultMap, ok := result.(map[string]interface{}); ok {
				if status, exists := resultMap["status"].(string); exists && status == "timeout" {
					intent := resultMap["intent"].(string)
					agentID := resultMap["selected_agent"].(string)
					responseMessage = fmt.Sprintf("I tried to %s but didn't get a response from the %s. This might be because the operation is taking longer than expected or the agent is busy. Please try again in a moment.", intent, agentID)
				} else if responseContent, ok := resultMap["response_content"].(string); ok {
					responseMessage = fmt.Sprintf("✅ %s\n\n%s", intent, responseContent)
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
		
		// Extract meaningful content from the agent response
		var responseContent string
		if decision, ok := response.Payload["decision"].(string); ok {
			if reasoning, ok := response.Payload["reasoning"].(string); ok {
				responseContent = fmt.Sprintf("Decision: %s. Reasoning: %s", decision, reasoning)
			} else {
				responseContent = fmt.Sprintf("Decision: %s", decision)
			}
		} else if message, ok := response.Payload["message"].(string); ok {
			responseContent = message
		} else {
			responseContent = fmt.Sprintf("Agent completed the %s request successfully", intent)
		}
		
		return map[string]interface{}{
			"status":           "completed",
			"intent":           intent,
			"selected_agent":   response.Source,
			"response_content": responseContent,
			"agent_response":   response.Payload,
		}, nil
	case <-time.After(10 * time.Second): // 10 second timeout
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
	// Get all available capabilities
	capabilities, err := agent.agentRegistry.GetAvailableCapabilities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available capabilities: %w", err)
	}

	var matchingAgents []agents.AgentStatus

	// Find agents whose capabilities match the intent
	for _, capability := range capabilities {
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

	// Remove duplicates (an agent might match multiple intents)
	return agent.deduplicate(matchingAgents), nil
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
