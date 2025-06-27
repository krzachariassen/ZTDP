package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockMessageBus is a mock implementation of MessageBus for testing
type MockMessageBus struct {
	sentMessages []*AgentMessage
	sendError    error
}

func (m *MockMessageBus) SendMessage(ctx context.Context, targetID string, message *AgentMessage) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.sentMessages = append(m.sentMessages, message)
	return nil
}

func (m *MockMessageBus) GetMessagesForAgent(agentID string) []*AgentMessage {
	var messages []*AgentMessage
	for _, msg := range m.sentMessages {
		if msg.To == agentID {
			messages = append(messages, msg)
		}
	}
	return messages
}

func (m *MockMessageBus) SetSendError(err error) {
	m.sendError = err
}

func (m *MockMessageBus) GetSentMessages() []*AgentMessage {
	return m.sentMessages
}

func (m *MockMessageBus) ClearMessages() {
	m.sentMessages = nil
}

// TestRoutingServiceImpl_RouteExecutionStep tests routing execution steps to agents
func TestRoutingServiceImpl_RouteExecutionStep(t *testing.T) {
	ctx := context.Background()

	t.Run("should route execution step to agent successfully", func(t *testing.T) {
		messageBus := &MockMessageBus{}
		service := NewRoutingServiceImpl(messageBus)

		step := &ExecutionStep{
			ID:         "step-123",
			PlanID:     "plan-456",
			AgentID:    "deployment-agent-001",
			Action:     "deploy",
			Parameters: map[string]interface{}{"environment": "production"},
			Status:     "pending",
		}
		agentID := "deployment-agent-001"

		err := service.RouteExecutionStep(ctx, agentID, step)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify message was sent
		messages := messageBus.GetSentMessages()
		if len(messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(messages))
		}

		message := messages[0]
		if message.Type != MessageTypeExecutionStep {
			t.Errorf("expected message type %s, got %s", MessageTypeExecutionStep, message.Type)
		}

		if message.To != agentID {
			t.Errorf("expected message to %s, got %s", agentID, message.To)
		}

		if message.From != "ai-orchestrator" {
			t.Errorf("expected message from ai-orchestrator, got %s", message.From)
		}
	})

	t.Run("should fail with empty agent ID", func(t *testing.T) {
		messageBus := &MockMessageBus{}
		service := NewRoutingServiceImpl(messageBus)

		step := &ExecutionStep{
			ID:      "step-123",
			AgentID: "agent-123",
			Action:  "deploy",
			Status:  "pending",
		}

		err := service.RouteExecutionStep(ctx, "", step)

		if err == nil {
			t.Fatal("expected error for empty agent ID")
		}
	})

	t.Run("should fail with nil execution step", func(t *testing.T) {
		messageBus := &MockMessageBus{}
		service := NewRoutingServiceImpl(messageBus)

		err := service.RouteExecutionStep(ctx, "agent-123", nil)

		if err == nil {
			t.Fatal("expected error for nil execution step")
		}
	})

	t.Run("should fail when message bus returns error", func(t *testing.T) {
		messageBus := &MockMessageBus{}
		messageBus.SetSendError(errors.New("message bus error"))
		service := NewRoutingServiceImpl(messageBus)

		step := &ExecutionStep{
			ID:      "step-123",
			AgentID: "agent-123",
			Action:  "deploy",
			Status:  "pending",
		}

		err := service.RouteExecutionStep(ctx, "agent-123", step)

		if err == nil {
			t.Fatal("expected error when message bus fails")
		}
	})
}

// TestRoutingServiceImpl_RouteStatusUpdate tests routing status updates from agents
func TestRoutingServiceImpl_RouteStatusUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("should route status update to orchestrator successfully", func(t *testing.T) {
		messageBus := &MockMessageBus{}
		service := NewRoutingServiceImpl(messageBus)

		update := &StatusUpdate{
			StepID:    "step-123",
			AgentID:   "deployment-agent-001",
			Status:    "completed",
			Result:    map[string]interface{}{"deployed_version": "1.2.3"},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		err := service.RouteStatusUpdate(ctx, update)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify message was sent
		messages := messageBus.GetSentMessages()
		if len(messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(messages))
		}

		message := messages[0]
		if message.Type != MessageTypeStatusUpdate {
			t.Errorf("expected message type %s, got %s", MessageTypeStatusUpdate, message.Type)
		}

		if message.To != "ai-orchestrator" {
			t.Errorf("expected message to ai-orchestrator, got %s", message.To)
		}

		if message.From != update.AgentID {
			t.Errorf("expected message from %s, got %s", update.AgentID, message.From)
		}
	})

	t.Run("should fail with nil status update", func(t *testing.T) {
		messageBus := &MockMessageBus{}
		service := NewRoutingServiceImpl(messageBus)

		err := service.RouteStatusUpdate(ctx, nil)

		if err == nil {
			t.Fatal("expected error for nil status update")
		}
	})

	t.Run("should fail when message bus returns error", func(t *testing.T) {
		messageBus := &MockMessageBus{}
		messageBus.SetSendError(errors.New("message bus error"))
		service := NewRoutingServiceImpl(messageBus)

		update := &StatusUpdate{
			StepID:  "step-123",
			AgentID: "agent-123",
			Status:  "completed",
		}

		err := service.RouteStatusUpdate(ctx, update)

		if err == nil {
			t.Fatal("expected error when message bus fails")
		}
	})
}

// TestRoutingServiceImpl_GetAvailableAgents tests retrieving agent catalog for AI planning
func TestRoutingServiceImpl_GetAvailableAgents(t *testing.T) {
	ctx := context.Background()

	t.Run("should return available agents for AI planning", func(t *testing.T) {
		messageBus := &MockMessageBus{}
		service := NewRoutingServiceImpl(messageBus)

		agents, err := service.GetAvailableAgents(ctx)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(agents) == 0 {
			t.Fatal("expected at least one agent")
		}

		// Verify agent structure
		agent := agents[0]
		if agent.ID == "" {
			t.Error("expected agent to have ID")
		}

		if agent.Name == "" {
			t.Error("expected agent to have name")
		}

		if len(agent.Capabilities) == 0 {
			t.Error("expected agent to have capabilities")
		}

		if agent.Status == "" {
			t.Error("expected agent to have status")
		}
	})
}

// TestRoutingServiceExists tests that the routing service can be created
func TestRoutingServiceExists(t *testing.T) {
	t.Run("RoutingServiceImpl should exist", func(t *testing.T) {
		messageBus := &MockMessageBus{}
		service := NewRoutingServiceImpl(messageBus)

		if service == nil {
			t.Fatal("NewRoutingServiceImpl should return a service instance")
		}
	})
}
