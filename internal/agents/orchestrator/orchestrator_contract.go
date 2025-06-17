package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// executeContract executes a contract by extracting intent and routing to appropriate agents
func (o *Orchestrator) executeContract(ctx context.Context, contractJSON string, userMessage string) (interface{}, error) {
	var contractData map[string]interface{}
	if err := json.Unmarshal([]byte(contractJSON), &contractData); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// PURE ORCHESTRATOR: Extract intent using AI, no hardcoded domain knowledge
	intent, err := o.extractIntentFromContract(ctx, contractData, userMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to extract intent: %w", err)
	}

	o.logger.Info("🚀 Extracted intent '%s' from contract, orchestrating via agent discovery", intent)

	// Route purely by intent - completely domain-agnostic!
	return o.orchestrateViaIntentBasedAgents(ctx, intent, contractData)
}

// extractIntentFromContract uses AI to determine the intent from contract data
func (o *Orchestrator) extractIntentFromContract(ctx context.Context, contractData map[string]interface{}, userMessage string) (string, error) {
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

	response, err := o.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("AI intent extraction failed: %w", err)
	}

	// Clean up the response to get just the intent
	intent := strings.TrimSpace(strings.ToLower(response))
	intent = strings.Trim(intent, `"'`)

	return intent, nil
}

// detectDirectContract checks if user input contains a valid contract and executes it directly
func (o *Orchestrator) detectDirectContract(ctx context.Context, userMessage string) *ConversationalResponse {
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
	o.logger.Info("🎯 Detected direct contract in user input: %s", contractJSON)

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
	o.logger.Info("🚀 Executing direct contract via intent-based orchestration")
	result, err := o.executeContract(ctx, contractJSON, userMessage)
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

// mustMarshal helper function
func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
