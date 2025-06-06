package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/krzachariassen/ZTDP/internal/logging"
)

// ConversationEngine handles AI-powered conversational interactions
// This is where the revolutionary conversational AI capabilities live
type ConversationEngine struct {
	provider AIProvider
	logger   *logging.Logger
}

// NewConversationEngine creates a new conversation engine
func NewConversationEngine(provider AIProvider, logger *logging.Logger) *ConversationEngine {
	return &ConversationEngine{
		provider: provider,
		logger:   logger,
	}
}

// GenerateResponse creates a conversational response using AI
func (engine *ConversationEngine) GenerateResponse(
	ctx context.Context,
	query string,
	intent *Intent,
	actions []Action,
	platformContext *PlatformContext,
) (*ConversationalResponse, error) {
	engine.logger.Info("🗣️ Generating conversational response for intent: %s", intent.Type)

	// Build sophisticated system prompt for conversational response
	systemPrompt := engine.buildConversationalSystemPrompt(intent, platformContext)
	userPrompt := engine.buildConversationalUserPrompt(query, actions, platformContext)

	// Generate response using AI
	rawResponse, err := engine.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI conversation generation failed: %w", err)
	}

	// Parse and structure the response
	response, err := engine.parseConversationalResponse(rawResponse, intent, actions)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversational response: %w", err)
	}

	engine.logger.Info("✅ Conversational response generated: %d insights, %d suggestions",
		len(response.Insights), len(response.Actions))

	return response, nil
}

// buildConversationalSystemPrompt creates the system prompt for conversational AI
func (engine *ConversationEngine) buildConversationalSystemPrompt(intent *Intent, context *PlatformContext) string {
	return fmt.Sprintf(`You are the ZTDP Platform Agent, an AI-native infrastructure platform assistant.

Your role is to provide conversational, helpful responses about platform operations while maintaining technical accuracy.

CONTEXT:
- Platform has %d applications, %d services
- Intent detected: %s
- Platform health: %s

RESPONSE GUIDELINES:
1. Be conversational and helpful, not robotic
2. Provide actionable insights and next steps
3. Use emojis appropriately to make responses engaging
4. If actions were performed, summarize their results clearly
5. Offer proactive suggestions when relevant
6. Maintain technical accuracy while being accessible

TONE: Professional but friendly, knowledgeable but not condescending.`,
		len(context.Applications),
		len(context.Services),
		intent.Type,
		context.Health["status"])
}

// buildConversationalUserPrompt creates the user prompt with context and actions
func (engine *ConversationEngine) buildConversationalUserPrompt(
	query string,
	actions []Action,
	context *PlatformContext,
) string {
	prompt := fmt.Sprintf("User Query: %s\n\n", query)

	// Add action results if any
	if len(actions) > 0 {
		prompt += "Actions Performed:\n"
		for _, action := range actions {
			prompt += fmt.Sprintf("- %s: %s\n", action.Type, action.Status)
		}
		prompt += "\n"
	}

	// Add relevant platform context
	prompt += "Platform Context:\n"
	prompt += fmt.Sprintf("- Applications: %d\n", len(context.Applications))
	prompt += fmt.Sprintf("- Services: %d\n", len(context.Services))
	prompt += fmt.Sprintf("- Health Status: %s\n", context.Health["status"])

	prompt += "\nGenerate a conversational response that addresses the user's query and summarizes any actions taken."

	return prompt
}

// parseConversationalResponse parses the AI response into structured format
func (engine *ConversationEngine) parseConversationalResponse(
	rawResponse string,
	intent *Intent,
	actions []Action,
) (*ConversationalResponse, error) {
	// Parse the response into structured format
	response := &ConversationalResponse{
		Message:   strings.TrimSpace(rawResponse),
		Intent:    intent.Type,
		Actions:   actions,
		Insights:  engine.extractInsights(rawResponse),
		Timestamp: context.Background(), // Will be set properly in actual implementation
	}

	return response, nil
}

// extractInsights extracts key insights from the AI response
func (engine *ConversationEngine) extractInsights(response string) []string {
	insights := []string{}

	// Simple insight extraction (would be more sophisticated in production)
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "insight") || strings.Contains(line, "recommendation") {
			insights = append(insights, line)
		}
	}

	return insights
}
