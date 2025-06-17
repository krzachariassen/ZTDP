package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/events"
)

// orchestrateViaIntentBasedAgents - PURE ORCHESTRATOR: Discovers agents by intent and routes events
// This method contains NO domain-specific logic - it's completely generic!
func (o *Orchestrator) orchestrateViaIntentBasedAgents(ctx context.Context, intent string, context map[string]interface{}) (interface{}, error) {
	if o.agentRegistry == nil {
		return nil, fmt.Errorf("agent registry not available - cannot discover agents")
	}

	o.logger.Info("🔍 Discovering agents for intent: %s", intent)

	// STEP 1: Discover agents by intent (completely generic)
	availableAgents, err := o.discoverAgentsByIntent(ctx, intent)
	if err != nil {
		return nil, fmt.Errorf("agent discovery failed for intent '%s': %w", intent, err)
	}

	if len(availableAgents) == 0 {
		return nil, fmt.Errorf("no agents found for intent '%s' - register appropriate agents first", intent)
	}

	o.logger.Info("🎯 Found %d agents capable of handling intent: %s", len(availableAgents), intent)

	// STEP 2: Route to the best agent and get routing key
	selectedAgent := availableAgents[0] // Simple: use first available agent

	// STEP 2.5: Discover the appropriate routing key for this intent
	routingKey, err := o.discoverRoutingKeyForIntent(ctx, intent, selectedAgent.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover routing key for intent '%s' and agent '%s': %w", intent, selectedAgent.ID, err)
	}

	o.logger.Info("🔑 Using routing key '%s' for agent: %s", routingKey, selectedAgent.ID)

	// STEP 3: Create request-response correlation
	correlationID := fmt.Sprintf("orchestration-%d", time.Now().UnixNano())
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Create a channel to receive the response
	responseChan := make(chan *events.Event, 1)

	// Subscribe to response events for this correlation ID
	o.eventBus.Subscribe(events.EventTypeResponse, func(event events.Event) error {
		// Check if this response is for our request
		if responseCorrelationID, ok := event.Payload["correlation_id"].(string); ok {
			if responseCorrelationID == correlationID {
				// This is our response!
				select {
				case responseChan <- &event:
					o.logger.Info("📨 Received response for correlation ID: %s", correlationID)
				default:
					o.logger.Warn("Response channel full for correlation ID: %s", correlationID)
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
		"source_agent":   "orchestrator",
	}

	// Targeted event emission using specific routing key for this agent
	if err := o.eventBus.Emit(events.EventTypeRequest, "orchestrator", routingKey, eventPayload); err != nil {
		return nil, fmt.Errorf("failed to emit intent request to routing key %s for agent %s: %w", routingKey, selectedAgent.ID, err)
	}

	o.logger.Info("📤 Routed intent '%s' to agent: %s via routing key: %s", intent, selectedAgent.ID, routingKey)

	// STEP 5: Handle test mode vs real mode
	if o.testMode {
		// In test mode, simulate successful routing without waiting for real responses
		o.logger.Info("🧪 Test mode: Simulating successful routing to agent: %s", selectedAgent.ID)
		return map[string]interface{}{
			"status":           "completed",
			"intent":           intent,
			"selected_agent":   selectedAgent.ID,
			"response_content": fmt.Sprintf("✅ Successfully routed %s request to %s (test mode)", intent, selectedAgent.ID),
			"agent_response":   map[string]interface{}{"test_mode": true},
			"correlation_id":   correlationID,
			"routing_key":      routingKey,
		}, nil
	}

	// STEP 5: Wait for response with timeout (real mode)
	select {
	case response := <-responseChan:
		o.logger.Info("✅ Received response from agent for intent: %s", intent)

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
		o.logger.Warn("⏰ Timeout waiting for response from agent for intent: %s", intent)
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
func (o *Orchestrator) discoverAgentsByIntent(ctx context.Context, intent string) ([]agentRegistry.AgentStatus, error) {
	var matchingAgents []agentRegistry.AgentStatus

	// Get all available capabilities to find routing keys
	capabilities, err := o.agentRegistry.GetAvailableCapabilities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available capabilities: %w", err)
	}

	// Find capabilities that match the intent
	for _, capability := range capabilities {
		// Check if this capability matches the intent
		for _, supportedIntent := range capability.Intents {
			if o.intentMatches(intent, supportedIntent) {
				// Find agents with this capability
				agentsWithCapability, err := o.agentRegistry.FindAgentsByCapability(ctx, capability.Name)
				if err != nil {
					o.logger.Warn("⚠️ Failed to find agents for capability %s: %v", capability.Name, err)
					continue
				}
				matchingAgents = append(matchingAgents, agentsWithCapability...)
				break // Found a match, no need to check other intents for this capability
			}
		}
	}

	// Remove duplicates and exclude self (Orchestrator should not route to itself during orchestration)
	deduplicated := o.deduplicate(matchingAgents)
	return o.excludeSelf(deduplicated), nil
}

// intentMatches - Simplified exact matching since AI provides precise intent names
func (o *Orchestrator) intentMatches(userIntent, supportedIntent string) bool {
	// AI provides precise intent names, so we can do exact matching
	userIntent = strings.ToLower(strings.TrimSpace(userIntent))
	supportedIntent = strings.ToLower(strings.TrimSpace(supportedIntent))

	return userIntent == supportedIntent
}

// deduplicate - Remove duplicate agents from the list
func (o *Orchestrator) deduplicate(agentsList []agentRegistry.AgentStatus) []agentRegistry.AgentStatus {
	seen := make(map[string]bool)
	var result []agentRegistry.AgentStatus

	for _, a := range agentsList {
		if !seen[a.ID] {
			seen[a.ID] = true
			result = append(result, a)
		}
	}

	return result
}

// excludeSelf - Remove the Orchestrator itself from agent selection during orchestration
func (o *Orchestrator) excludeSelf(agentsList []agentRegistry.AgentStatus) []agentRegistry.AgentStatus {
	var result []agentRegistry.AgentStatus

	for _, a := range agentsList {
		if a.ID != "orchestrator" { // Orchestrator should not route to itself during orchestration
			result = append(result, a)
		}
	}

	return result
}

// discoverRoutingKeyForIntent finds the appropriate routing key for an intent and agent
func (o *Orchestrator) discoverRoutingKeyForIntent(ctx context.Context, intent string, agentID string) (string, error) {
	// Get all available capabilities to find routing keys
	capabilities, err := o.agentRegistry.GetAvailableCapabilities(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get available capabilities: %w", err)
	}

	// Find the capability that matches both the intent and the agent
	for _, capability := range capabilities {
		// Check if this capability matches the intent
		for _, supportedIntent := range capability.Intents {
			if o.intentMatches(intent, supportedIntent) {
				// Check if this capability belongs to the target agent
				agentsWithCapability, err := o.agentRegistry.FindAgentsByCapability(ctx, capability.Name)
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
