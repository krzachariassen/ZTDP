package web

import (
	"context"

	"github.com/ztdp/orchestrator/internal/ai"
	"github.com/ztdp/orchestrator/internal/orchestrator/application"
)

// OrchestratorAdapter adapts the new clean architecture orchestrator
// to the web interface expectations
type OrchestratorAdapter struct {
	orchestratorService *application.OrchestratorService
}

// NewOrchestratorAdapter creates a new adapter
func NewOrchestratorAdapter(orchestratorService *application.OrchestratorService) *OrchestratorAdapter {
	return &OrchestratorAdapter{
		orchestratorService: orchestratorService,
	}
}

// ProcessRequest adapts the new ProcessUserRequest to the old web interface
func (w *OrchestratorAdapter) ProcessRequest(ctx context.Context, userInput, userID string) (*ai.ConversationalResponse, error) {
	request := &application.OrchestratorRequest{
		UserInput: userInput,
		UserID:    userID,
	}

	result, err := w.orchestratorService.ProcessUserRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// Convert OrchestratorResult to ConversationalResponse
	response := &ai.ConversationalResponse{
		Message:    result.Message,
		Intent:     result.Analysis.Intent,
		Confidence: float64(result.Analysis.Confidence) / 100.0, // Convert to 0.0-1.0 range
		Actions:    convertToLegacyActions(result),
	}

	return response, nil
}

// convertToLegacyActions converts the new execution plan to legacy actions
func convertToLegacyActions(result *application.OrchestratorResult) []ai.Action {
	var actions []ai.Action

	if result.ExecutionPlanID != "" {
		actions = append(actions, ai.Action{
			Type: "execute_plan",
			Parameters: map[string]interface{}{
				"plan_id": result.ExecutionPlanID,
			},
		})
	}

	if result.Analysis != nil {
		actions = append(actions, ai.Action{
			Type: "analysis_complete",
			Parameters: map[string]interface{}{
				"intent":     result.Analysis.Intent,
				"category":   result.Analysis.Category,
				"confidence": result.Analysis.Confidence,
				"agents":     result.Analysis.RequiredAgents,
			},
		})
	}

	if result.Decision != nil {
		actions = append(actions, ai.Action{
			Type: "decision_made",
			Parameters: map[string]interface{}{
				"decision_type": string(result.Decision.Type),
				"action":        result.Decision.Action,
				"reasoning":     result.Decision.Reasoning,
			},
		})
	}

	return actions
}
