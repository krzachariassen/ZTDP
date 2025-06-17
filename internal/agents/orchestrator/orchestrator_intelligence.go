package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
)

// buildDynamicPlatformKnowledge uses AI to analyze the agent registry and build dynamic platform knowledge
func (o *Orchestrator) buildDynamicPlatformKnowledge(ctx context.Context) (string, error) {
	if o.agentRegistry == nil {
		return "No agents available", nil
	}

	// Get all available capabilities from the agent registry
	capabilities, err := o.agentRegistry.GetAvailableCapabilities(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get capabilities: %w", err)
	}

	// Get platform state
	platformState := o.getPlatformState()

	// Build comprehensive platform knowledge using AI
	systemPrompt := `You are a platform intelligence analyzer. Your job is to analyze the available agent capabilities and current platform state to build comprehensive knowledge about what this platform can do.

TASK: Create a comprehensive, natural description of platform capabilities based on the agent registry.

GUIDELINES:
1. Analyze each agent capability and understand what it enables
2. Group related capabilities into logical domains
3. Explain the relationships between different operations
4. Describe what users can ask for and what will happen
5. Be specific about intents that are supported
6. Include examples of natural language requests users can make

OUTPUT FORMAT:
Return a structured knowledge base that another AI can use to help users.`

	capabilityData := fmt.Sprintf(`CURRENT PLATFORM STATE:
%s

AVAILABLE AGENT CAPABILITIES:
%s`, platformState, o.formatCapabilitiesForAI(capabilities))

	knowledge, err := o.aiProvider.CallAI(ctx, systemPrompt, capabilityData)
	if err != nil {
		return "", fmt.Errorf("failed to build platform knowledge: %w", err)
	}

	return knowledge, nil
}

// buildDynamicIntentDetectionPrompt creates an AI-generated prompt for intent detection
func (o *Orchestrator) buildDynamicIntentDetectionPrompt(ctx context.Context) (string, error) {
	if o.agentRegistry == nil {
		return o.getDefaultIntentDetectionPrompt(), nil
	}

	// Get all available capabilities
	capabilities, err := o.agentRegistry.GetAvailableCapabilities(ctx)
	if err != nil {
		return o.getDefaultIntentDetectionPrompt(), nil
	}

	// Build a comprehensive list of available capabilities and their intents
	var capabilityInfo []string
	for _, capability := range capabilities {
		info := fmt.Sprintf("- %s: %s (intents: %s)",
			capability.Name,
			capability.Description,
			strings.Join(capability.Intents, ", "))
		capabilityInfo = append(capabilityInfo, info)
	}

	// Create a precise prompt that instructs AI to match user requests to agent capabilities
	systemPrompt := `You are an intelligent agent router for a platform AI system.

TASK: Analyze user requests and determine which agent should handle them based on available capabilities.

AVAILABLE AGENT CAPABILITIES:
%s

ROUTING RULES:
1. Analyze the user's request and understand what they want to accomplish
2. Match the request to the most appropriate agent capability
3. Return the specific intent name that best matches their request
4. If no capability matches, return "general_conversation"

EXAMPLES:
- "Deploy myapp to production" → "deploy application"
- "Check if deployment is allowed" → "policy check"
- "Create a new service called checkout" → "create application"
- "What is this platform?" → "general_conversation"
- "Help me understand what I can do" → "general_conversation"

IMPORTANT: Return only the intent name, no prefix like "INTENT:" needed.

OUTPUT FORMAT: Just the intent name (e.g., "deploy application") or "general_conversation"`

	capabilityList := strings.Join(capabilityInfo, "\n")
	return fmt.Sprintf(systemPrompt, capabilityList), nil
}

// buildDynamicConversationPrompt creates an AI-generated prompt for general conversation
func (o *Orchestrator) buildDynamicConversationPrompt(ctx context.Context, platformKnowledge string) (string, error) {
	// Use AI to build the conversation prompt based on dynamic platform knowledge
	systemPrompt := `You are a prompt engineer. Your job is to create an optimized conversation prompt for a platform AI assistant.

TASK: Create a prompt that enables natural, helpful conversation about platform capabilities based on the dynamic platform knowledge.

REQUIREMENTS:
1. The prompt should enable the AI to help users understand what they can do
2. Should encourage users to ask for what they need in natural language
3. Should be conversational and helpful
4. Should reference actual platform capabilities, not hardcoded assumptions
5. Should guide users toward successful interactions

OUTPUT: Return ONLY the system prompt that will be used for conversation.`

	knowledge := fmt.Sprintf(`DYNAMIC PLATFORM KNOWLEDGE:
%s

CURRENT PLATFORM STATE:
%s`, platformKnowledge, o.getPlatformState())

	prompt, err := o.aiProvider.CallAI(ctx, systemPrompt, knowledge)
	if err != nil {
		return o.getDefaultConversationPrompt(), nil
	}

	return prompt, nil
}

// formatCapabilitiesForAI formats capabilities in a way that's easy for AI to understand
func (o *Orchestrator) formatCapabilitiesForAI(capabilities []agentRegistry.AgentCapability) string {
	if len(capabilities) == 0 {
		return "No capabilities available"
	}

	var formatted strings.Builder
	for _, cap := range capabilities {
		formatted.WriteString(fmt.Sprintf(`
CAPABILITY: %s
Description: %s
Supported Intents: %s
Input Types: %s
Output Types: %s
Routing Keys: %s
Version: %s
---`, cap.Name, cap.Description, strings.Join(cap.Intents, ", "),
			strings.Join(cap.InputTypes, ", "), strings.Join(cap.OutputTypes, ", "),
			strings.Join(cap.RoutingKeys, ", "), cap.Version))
	}

	return formatted.String()
}

// getDefaultIntentDetectionPrompt provides a fallback if dynamic generation fails
func (o *Orchestrator) getDefaultIntentDetectionPrompt() string {
	return `You are an intelligent agent router for a platform AI system.

Your job is to analyze user requests and determine which agent should handle them based on available capabilities.

TASK: Match the user request to the most appropriate agent capability.

GUIDELINES:
1. Look at the user's request and understand what they want to do
2. Match it to the most relevant capability from the available agents
3. Return the specific intent that matches their request
4. If no agent capability matches, return "general_conversation"

OUTPUT FORMAT: 
- For agent routing: Return just the intent name (e.g., "deploy application", "policy check", "create application")
- For general questions: Return "general_conversation"

EXAMPLES:
- "Deploy myapp to production" → "deploy application"
- "Check if deployment is allowed" → "policy check"  
- "Create a new service" → "create application"
- "What is ZTDP?" → "general_conversation"
- "Help me understand this platform" → "general_conversation"`
}

// getDefaultConversationPrompt provides a fallback if dynamic generation fails
func (o *Orchestrator) getDefaultConversationPrompt() string {
	return `You are a helpful platform AI assistant. Help users understand what they can do and respond to their requests naturally.`
}
