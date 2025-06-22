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

// mustMarshal helper function
func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
