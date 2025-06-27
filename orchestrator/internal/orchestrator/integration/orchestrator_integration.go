package integration

import (
	"context"

	"github.com/ztdp/orchestrator/internal/logging"
	"github.com/ztdp/orchestrator/internal/orchestrator/application"
)

// OrchestratorIntegration provides a bridge between the old orchestrator interface and the new clean architecture
// This allows gradual migration without breaking existing code
type OrchestratorIntegration struct {
	orchestratorService *application.OrchestratorService
	logger              logging.Logger
}

// NewOrchestratorIntegration creates a new integration wrapper
func NewOrchestratorIntegration(orchestratorService *application.OrchestratorService, logger logging.Logger) *OrchestratorIntegration {
	return &OrchestratorIntegration{
		orchestratorService: orchestratorService,
		logger:              logger,
	}
}

// ProcessRequest provides backward compatibility with the old ProcessRequest interface
// This method maintains the same signature as the old orchestrator but uses the new clean architecture
func (oi *OrchestratorIntegration) ProcessRequest(ctx context.Context, userInput, userID string) (string, error) {
	// Convert to new request format
	request := &application.OrchestratorRequest{
		UserInput: userInput,
		UserID:    userID,
	}

	// Use the new orchestrator service
	result, err := oi.orchestratorService.ProcessUserRequest(ctx, request)
	if err != nil {
		return "", err
	}

	// Return error if the operation failed
	if !result.Success {
		return result.Error, nil
	}

	// Return the message for backward compatibility
	return result.Message, nil
}

// Start is provided for compatibility with the old orchestrator interface
func (oi *OrchestratorIntegration) Start(ctx context.Context) error {
	oi.logger.Info("🧠 New clean architecture orchestrator integration started")
	return nil
}

// Stop provides graceful shutdown
func (oi *OrchestratorIntegration) Stop(ctx context.Context) error {
	oi.logger.Info("🧠 Clean architecture orchestrator integration stopped")
	return nil
}
