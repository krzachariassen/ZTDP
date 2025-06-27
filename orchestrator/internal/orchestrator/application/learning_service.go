package application

import (
	"context"
	"fmt"
	"strings"

	orchestratorDomain "github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// ConversationService defines the interface for conversation operations
type ConversationService interface {
	StoreInteraction(ctx context.Context, userRequest string, analysis *orchestratorDomain.Analysis, decision *orchestratorDomain.Decision) error
	GetConversationHistory(ctx context.Context, sessionID string) ([]string, error)
	CreateSession(ctx context.Context) (string, error)
}

// LearningService handles learning and insights storage, fixing the architecture violation
// Replaces the storeInsightsToGraph() functionality from the old orchestrator
type LearningService struct {
	conversationService ConversationService
}

// NewLearningService creates a new LearningService instance
func NewLearningService(conversationService ConversationService) *LearningService {
	return &LearningService{
		conversationService: conversationService,
	}
}

// StoreInsights stores interaction insights using the conversation service
// This replaces direct graph access from the old orchestrator
func (ls *LearningService) StoreInsights(ctx context.Context, userRequest string, analysis *orchestratorDomain.Analysis, decision *orchestratorDomain.Decision) error {
	err := ls.conversationService.StoreInteraction(ctx, userRequest, analysis, decision)
	if err != nil {
		return fmt.Errorf("failed to store interaction insights: %w", err)
	}

	return nil
}

// AnalyzePatterns analyzes conversation patterns for a session
func (ls *LearningService) AnalyzePatterns(ctx context.Context, sessionID string) (*orchestratorDomain.ConversationPattern, error) {
	history, err := ls.conversationService.GetConversationHistory(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	pattern := orchestratorDomain.NewConversationPattern(sessionID)

	// Analyze conversation history for patterns
	for _, interaction := range history {
		pattern.IncrementInteractions()

		// Simple pattern recognition - in a real implementation this would be more sophisticated
		lowerInteraction := strings.ToLower(interaction)

		// Extract common intents
		if strings.Contains(lowerInteraction, "deploy") {
			pattern.AddIntent("deployment")
			pattern.AddAction("deploy")
		}
		if strings.Contains(lowerInteraction, "monitor") {
			pattern.AddIntent("monitoring")
			pattern.AddAction("monitor")
		}
		if strings.Contains(lowerInteraction, "status") || strings.Contains(lowerInteraction, "check") {
			pattern.AddIntent("status_check")
			pattern.AddAction("check")
		}

		// Extract entities (environments, app names, etc.)
		if strings.Contains(lowerInteraction, "staging") {
			pattern.AddEntity("staging")
		}
		if strings.Contains(lowerInteraction, "production") {
			pattern.AddEntity("production")
		}
		if strings.Contains(lowerInteraction, "app") {
			pattern.AddEntity("application")
		}
	}

	// Calculate success rate (simplified - assume all interactions are successful for now)
	pattern.UpdateSuccessRate(len(history))

	return pattern, nil
}

// CreateLearningSession creates a new learning session
func (ls *LearningService) CreateLearningSession(ctx context.Context) (string, error) {
	sessionID, err := ls.conversationService.CreateSession(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create learning session: %w", err)
	}

	return sessionID, nil
}
