package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RoutingServiceImpl implements the RoutingService interface using AI-native messaging
type RoutingServiceImpl struct {
	messageBus MessageBus
}

// NewRoutingServiceImpl creates a new routing service implementation
func NewRoutingServiceImpl(messageBus MessageBus) RoutingService {
	return &RoutingServiceImpl{
		messageBus: messageBus,
	}
}

// RouteExecutionStep routes an execution step to the specified agent via messaging
func (s *RoutingServiceImpl) RouteExecutionStep(ctx context.Context, agentID string, step *ExecutionStep) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	if step == nil {
		return fmt.Errorf("execution step cannot be nil")
	}

	// Create message payload
	payload := map[string]interface{}{
		"step_id":    step.ID,
		"plan_id":    step.PlanID,
		"action":     step.Action,
		"parameters": step.Parameters,
		"status":     step.Status,
	}

	// Create agent message
	message := &AgentMessage{
		ID:      uuid.New().String(),
		Type:    MessageTypeExecutionStep,
		From:    "ai-orchestrator",
		To:      agentID,
		Payload: payload,
		SentAt:  time.Now().UTC().Format(time.RFC3339),
	}

	// Send message via message bus
	err := s.messageBus.SendMessage(ctx, agentID, message)
	if err != nil {
		return fmt.Errorf("failed to route execution step to agent %s: %w", agentID, err)
	}

	return nil
}

// RouteStatusUpdate routes status updates from agents back to the orchestrator
func (s *RoutingServiceImpl) RouteStatusUpdate(ctx context.Context, update *StatusUpdate) error {
	if update == nil {
		return fmt.Errorf("status update cannot be nil")
	}

	// Create message payload
	payload := map[string]interface{}{
		"step_id":   update.StepID,
		"agent_id":  update.AgentID,
		"status":    update.Status,
		"result":    update.Result,
		"error":     update.Error,
		"timestamp": update.Timestamp,
	}

	// Create agent message
	message := &AgentMessage{
		ID:      uuid.New().String(),
		Type:    MessageTypeStatusUpdate,
		From:    update.AgentID,
		To:      "ai-orchestrator",
		Payload: payload,
		SentAt:  time.Now().UTC().Format(time.RFC3339),
	}

	// Send message to orchestrator via message bus
	err := s.messageBus.SendMessage(ctx, "ai-orchestrator", message)
	if err != nil {
		return fmt.Errorf("failed to route status update from agent %s: %w", update.AgentID, err)
	}

	return nil
}

// GetAvailableAgents retrieves the catalog of available agents for AI planning
func (s *RoutingServiceImpl) GetAvailableAgents(ctx context.Context) ([]*AgentInfo, error) {
	// TODO: In real implementation, this would query the agent registry
	// For now, return a mock list to satisfy the interface
	agents := []*AgentInfo{
		{
			ID:           "deployment-agent-001",
			Name:         "Deployment Agent",
			Capabilities: []string{"deploy", "rollback", "health-check"},
			Status:       "online",
		},
		{
			ID:           "monitoring-agent-001",
			Name:         "Monitoring Agent",
			Capabilities: []string{"monitor", "alert", "dashboard"},
			Status:       "online",
		},
	}

	return agents, nil
}
