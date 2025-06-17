package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/ai"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/graph"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// Orchestrator - Pure AI-native orchestrator following Clean Architecture
// This is a 1:1 functional replacement of V3Agent with proper domain separation
type Orchestrator struct {
	aiProvider    ai.AIProvider
	logger        *logging.Logger
	graph         *graph.GlobalGraph
	eventBus      *events.EventBus
	agentRegistry agentRegistry.AgentRegistry

	// Agent interface properties
	agentID   string
	startTime time.Time

	// Raw interfaces for framework compatibility
	rawAIProvider    interface{}
	rawGraph         interface{}
	rawEventBus      interface{}
	rawAgentRegistry interface{}

	// Test mode flag - when true, don't wait for agent responses
	testMode bool
}

// ConversationalResponse represents the response structure for chat interactions
type ConversationalResponse struct {
	Message    string   `json:"message"`
	Answer     string   `json:"answer,omitempty"`
	Intent     string   `json:"intent,omitempty"`
	Actions    []Action `json:"actions,omitempty"`
	Insights   []string `json:"insights,omitempty"`
	Confidence float64  `json:"confidence,omitempty"`
}

// Action represents an action taken by the orchestrator
type Action struct {
	Type   string      `json:"type"`
	Result interface{} `json:"result"`
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(
	aiProvider ai.AIProvider,
	globalGraph *graph.GlobalGraph,
	eventBus *events.EventBus,
	agentRegistry agentRegistry.AgentRegistry,
) *Orchestrator {
	return &Orchestrator{
		aiProvider:    aiProvider,
		logger:        logging.GetLogger().ForComponent("orchestrator"),
		graph:         globalGraph,
		eventBus:      eventBus,
		agentRegistry: agentRegistry,
		agentID:       "orchestrator",
	}
}

// Chat - Simplified AI-native orchestration interface
func (o *Orchestrator) Chat(ctx context.Context, userMessage string) (*ConversationalResponse, error) {
	o.logger.Info("🤖 Orchestrator Chat: %s", userMessage)

	// STEP 1: Use AI to determine intent and route accordingly
	return o.routeUserRequest(ctx, userMessage)
}

// routeUserRequest - Simplified routing using AI to determine intent and route accordingly
func (o *Orchestrator) routeUserRequest(ctx context.Context, userMessage string) (*ConversationalResponse, error) {
	// Use AI to determine the intent based on available agent capabilities
	intentDetectionPrompt, err := o.buildDynamicIntentDetectionPrompt(ctx)
	if err != nil {
		o.logger.Warn("Failed to build dynamic intent detection prompt, using fallback: %v", err)
		intentDetectionPrompt = o.getDefaultIntentDetectionPrompt()
	}

	response, err := o.aiProvider.CallAI(ctx, intentDetectionPrompt, userMessage)
	if err != nil {
		o.logger.Error("Intent detection failed: %v", err)
		// Fall back to general conversation
		return o.handleGeneralConversation(ctx, userMessage)
	}

	// Clean up the response
	intent := strings.TrimSpace(response)

	// Check if this is a general conversation (not an agent intent)
	if intent == "general_conversation" || intent == "" {
		return o.handleGeneralConversation(ctx, userMessage)
	}

	o.logger.Info("🎯 Detected operational intent: %s", intent)

	// Route to appropriate agent via intent-based orchestration
	result, err := o.orchestrateViaIntentBasedAgents(ctx, intent, map[string]interface{}{
		"user_message": userMessage,
		"source":       "orchestrator-chat",
	})

	if err != nil {
		o.logger.Error("Intent orchestration failed: %v", err)
		return &ConversationalResponse{
			Message: fmt.Sprintf("I understood you want to %s, but encountered an error: %v", intent, err),
			Answer:  fmt.Sprintf("I understood you want to %s, but encountered an error: %v", intent, err),
		}, nil
	}

	// Convert result to conversational response
	var responseMessage string
	if result != nil {
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
		Actions: []Action{{Type: "orchestration", Result: result}},
	}, nil
}

// handleGeneralConversation - Simplified general conversation handling
func (o *Orchestrator) handleGeneralConversation(ctx context.Context, userMessage string) (*ConversationalResponse, error) {
	// Build dynamic platform knowledge from agent registry
	platformKnowledge, err := o.buildDynamicPlatformKnowledge(ctx)
	if err != nil {
		o.logger.Warn("Failed to build dynamic platform knowledge, using fallback: %v", err)
		platformKnowledge = "Platform knowledge unavailable"
	}

	// Build dynamic conversation prompt based on platform knowledge
	conversationPrompt, err := o.buildDynamicConversationPrompt(ctx, platformKnowledge)
	if err != nil {
		o.logger.Warn("Failed to build dynamic conversation prompt, using fallback: %v", err)
		conversationPrompt = o.getDefaultConversationPrompt()
	}

	response, err := o.aiProvider.CallAI(ctx, conversationPrompt, userMessage)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	o.logger.Info("🤖 AI Response: %s", response)

	// Determine intent for general conversation
	intent := "general_conversation"
	if strings.Contains(strings.ToLower(userMessage), "help") || strings.Contains(strings.ToLower(userMessage), "what") {
		intent = "help_request"
	}

	return &ConversationalResponse{
		Message: response,
		Answer:  response,
		Intent:  intent,
		Actions: []Action{{Type: "conversation", Result: "general_help"}},
	}, nil
}
