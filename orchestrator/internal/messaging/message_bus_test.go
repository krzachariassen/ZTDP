package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ztdp/orchestrator/internal/logging"
)

func TestMemoryMessageBus_TDD(t *testing.T) {
	t.Run("can_create_message_bus", func(t *testing.T) {
		bus := NewMemoryMessageBus(logging.NewNoOpLogger())
		require.NotNil(t, bus)
	})

	t.Run("can_subscribe_and_receive_messages", func(t *testing.T) {
		bus := NewMemoryMessageBus(logging.NewNoOpLogger())
		ctx := context.Background()

		// Subscribe agent to receive messages
		agentID := "test-agent-1"
		messageChan, err := bus.Subscribe(ctx, agentID)
		require.NoError(t, err)
		require.NotNil(t, messageChan)

		// Send message to agent
		message := &Message{
			ID:            "msg-1",
			CorrelationID: "conv-1",
			FromID:        "ai-orchestrator",
			ToID:          agentID,
			Content:       "Hello, can you deploy this application?",
			MessageType:   MessageTypeAIToAgent,
			Metadata: map[string]interface{}{
				"action": "deploy",
				"app":    "test-app",
			},
			Timestamp: time.Now(),
		}

		err = bus.SendMessage(ctx, message)
		require.NoError(t, err)

		// Agent should receive the message
		select {
		case receivedMessage := <-messageChan:
			assert.Equal(t, message.ID, receivedMessage.ID)
			assert.Equal(t, message.Content, receivedMessage.Content)
			assert.Equal(t, MessageTypeAIToAgent, receivedMessage.MessageType)
			assert.Equal(t, "deploy", receivedMessage.Metadata["action"])
		case <-time.After(1 * time.Second):
			t.Fatal("Expected to receive message within 1 second")
		}
	})

	t.Run("can_handle_bidirectional_conversation", func(t *testing.T) {
		bus := NewMemoryMessageBus(logging.NewNoOpLogger())
		ctx := context.Background()

		// Setup participants
		aiID := "ai-orchestrator"
		agentID := "deployment-agent"

		// Both participants subscribe
		aiChan, err := bus.Subscribe(ctx, aiID)
		require.NoError(t, err)
		agentChan, err := bus.Subscribe(ctx, agentID)
		require.NoError(t, err)

		// AI sends request to agent
		aiRequest := &Message{
			ID:            "req-1",
			CorrelationID: "conversation-123",
			FromID:        aiID,
			ToID:          agentID,
			Content:       "Deploy application xyz to production",
			MessageType:   MessageTypeAIToAgent,
			Metadata: map[string]interface{}{
				"action":      "deploy",
				"application": "xyz",
				"environment": "production",
			},
			Timestamp: time.Now(),
		}

		err = bus.SendMessage(ctx, aiRequest)
		require.NoError(t, err)

		// Agent receives request
		var receivedRequest *Message
		select {
		case receivedRequest = <-agentChan:
			assert.Equal(t, aiRequest.Content, receivedRequest.Content)
		case <-time.After(1 * time.Second):
			t.Fatal("Agent should have received AI request")
		}

		// Agent asks for clarification
		clarificationRequest := &Message{
			ID:            "clarif-1",
			CorrelationID: "conversation-123",
			FromID:        agentID,
			ToID:          aiID,
			Content:       "What is the desired instance count for production?",
			MessageType:   MessageTypeClarification,
			Metadata: map[string]interface{}{
				"clarification_type": "parameter_missing",
				"missing_parameter":  "instance_count",
			},
			Timestamp: time.Now(),
		}

		err = bus.SendMessage(ctx, clarificationRequest)
		require.NoError(t, err)

		// AI receives clarification request
		select {
		case receivedClarification := <-aiChan:
			assert.Equal(t, clarificationRequest.Content, receivedClarification.Content)
			assert.Equal(t, MessageTypeClarification, receivedClarification.MessageType)
		case <-time.After(1 * time.Second):
			t.Fatal("AI should have received clarification request")
		}

		// AI responds with clarification
		clarificationResponse := &Message{
			ID:            "clarif-resp-1",
			CorrelationID: "conversation-123",
			FromID:        aiID,
			ToID:          agentID,
			Content:       "Use 3 instances for production deployment",
			MessageType:   MessageTypeResponse,
			Metadata: map[string]interface{}{
				"instance_count": 3,
				"response_to":    "clarif-1",
			},
			Timestamp: time.Now(),
		}

		err = bus.SendMessage(ctx, clarificationResponse)
		require.NoError(t, err)

		// Agent receives clarification response
		select {
		case receivedResponse := <-agentChan:
			assert.Equal(t, clarificationResponse.Content, receivedResponse.Content)
			assert.Equal(t, 3, receivedResponse.Metadata["instance_count"])
		case <-time.After(1 * time.Second):
			t.Fatal("Agent should have received clarification response")
		}
	})

	t.Run("can_manage_conversation_history", func(t *testing.T) {
		bus := NewMemoryMessageBus(logging.NewNoOpLogger())
		ctx := context.Background()

		conversationID := "test-conversation"
		participants := []string{"ai-orchestrator", "build-agent", "deploy-agent"}

		// Subscribe all participants first
		for _, participantID := range participants {
			_, err := bus.Subscribe(ctx, participantID)
			require.NoError(t, err)
		}

		// Create conversation context
		conversation, err := bus.CreateConversation(ctx, participants, map[string]interface{}{
			"workflow_id": "deploy-workflow-123",
			"environment": "production",
		})
		require.NoError(t, err)
		require.NotNil(t, conversation)
		assert.Equal(t, participants, conversation.Participants)

		// Send multiple messages in conversation
		messages := []*Message{
			{
				ID:            "msg-1",
				CorrelationID: conversationID,
				FromID:        "ai-orchestrator",
				ToID:          "build-agent",
				Content:       "Build application version 1.2.3",
				MessageType:   MessageTypeRequest,
				Timestamp:     time.Now(),
			},
			{
				ID:            "msg-2",
				CorrelationID: conversationID,
				FromID:        "build-agent",
				ToID:          "ai-orchestrator",
				Content:       "Build completed successfully",
				MessageType:   MessageTypeResponse,
				Timestamp:     time.Now().Add(1 * time.Minute),
			},
			{
				ID:            "msg-3",
				CorrelationID: conversationID,
				FromID:        "ai-orchestrator",
				ToID:          "deploy-agent",
				Content:       "Deploy built application to production",
				MessageType:   MessageTypeRequest,
				Timestamp:     time.Now().Add(2 * time.Minute),
			},
		}

		// Send all messages
		for _, msg := range messages {
			err = bus.SendMessage(ctx, msg)
			require.NoError(t, err)
		}

		// Retrieve conversation history
		history, err := bus.GetConversationHistory(ctx, conversationID)
		require.NoError(t, err)
		require.Len(t, history, 3)

		// Verify chronological order and content
		assert.Equal(t, "msg-1", history[0].ID)
		assert.Equal(t, "Build application version 1.2.3", history[0].Content)
		assert.Equal(t, "msg-2", history[1].ID)
		assert.Equal(t, "Build completed successfully", history[1].Content)
		assert.Equal(t, "msg-3", history[2].ID)
		assert.Equal(t, "Deploy built application to production", history[2].Content)
	})

	t.Run("can_broadcast_to_multiple_recipients", func(t *testing.T) {
		bus := NewMemoryMessageBus(logging.NewNoOpLogger())
		ctx := context.Background()

		// Setup multiple agents
		agents := []string{"agent-1", "agent-2", "agent-3"}
		channels := make(map[string]<-chan *Message)

		for _, agentID := range agents {
			ch, err := bus.Subscribe(ctx, agentID)
			require.NoError(t, err)
			channels[agentID] = ch
		}

		// Broadcast message to all agents
		broadcastMessage := &Message{
			ID:            "broadcast-1",
			CorrelationID: "system-notification",
			FromID:        "system",
			ToID:          "", // Will be set per recipient
			Content:       "System maintenance starting in 10 minutes",
			MessageType:   MessageTypeNotification,
			Metadata: map[string]interface{}{
				"maintenance_type": "scheduled",
				"duration":         "30 minutes",
			},
			Timestamp: time.Now(),
		}

		err := bus.PublishMessage(ctx, broadcastMessage, agents)
		require.NoError(t, err)

		// All agents should receive the message
		for agentID, ch := range channels {
			select {
			case receivedMessage := <-ch:
				assert.Equal(t, broadcastMessage.Content, receivedMessage.Content)
				assert.Equal(t, agentID, receivedMessage.ToID)
				assert.Equal(t, MessageTypeNotification, receivedMessage.MessageType)
			case <-time.After(1 * time.Second):
				t.Fatalf("Agent %s should have received broadcast message", agentID)
			}
		}
	})

	t.Run("handles_unsubscribe_correctly", func(t *testing.T) {
		bus := NewMemoryMessageBus(logging.NewNoOpLogger())
		ctx := context.Background()

		agentID := "test-agent"

		// Subscribe
		_, err := bus.Subscribe(ctx, agentID)
		require.NoError(t, err)

		// Unsubscribe
		err = bus.Unsubscribe(ctx, agentID)
		require.NoError(t, err)

		// Sending message should fail
		message := &Message{
			ID:          "msg-after-unsubscribe",
			FromID:      "ai",
			ToID:        agentID,
			Content:     "This should fail",
			MessageType: MessageTypeRequest,
			Timestamp:   time.Now(),
		}

		err = bus.SendMessage(ctx, message)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no subscriber found")
	})

	t.Run("handles_agent_to_agent_communication", func(t *testing.T) {
		bus := NewMemoryMessageBus(logging.NewNoOpLogger())
		ctx := context.Background()

		// Setup two agents
		agent1ID := "build-agent"
		agent2ID := "deploy-agent"

		agent1Chan, err := bus.Subscribe(ctx, agent1ID)
		require.NoError(t, err)
		agent2Chan, err := bus.Subscribe(ctx, agent2ID)
		require.NoError(t, err)

		// Agent 1 asks Agent 2 for status
		agentMessage := &Message{
			ID:            "agent-msg-1",
			CorrelationID: "agent-coordination",
			FromID:        agent1ID,
			ToID:          agent2ID,
			Content:       "Are you ready to receive the built artifact?",
			MessageType:   MessageTypeAgentToAgent,
			Metadata: map[string]interface{}{
				"artifact_size": "150MB",
				"build_id":      "build-789",
			},
			Timestamp: time.Now(),
		}

		err = bus.SendMessage(ctx, agentMessage)
		require.NoError(t, err)

		// Agent 2 receives message
		select {
		case receivedMessage := <-agent2Chan:
			assert.Equal(t, agentMessage.Content, receivedMessage.Content)
			assert.Equal(t, MessageTypeAgentToAgent, receivedMessage.MessageType)
			assert.Equal(t, "build-789", receivedMessage.Metadata["build_id"])
		case <-time.After(1 * time.Second):
			t.Fatal("Agent 2 should have received message from Agent 1")
		}

		// Agent 2 responds
		responseMessage := &Message{
			ID:            "agent-resp-1",
			CorrelationID: "agent-coordination",
			FromID:        agent2ID,
			ToID:          agent1ID,
			Content:       "Yes, ready to receive. Please use HTTPS endpoint.",
			MessageType:   MessageTypeResponse,
			Metadata: map[string]interface{}{
				"endpoint": "https://deploy-agent/artifacts",
				"ready":    true,
			},
			Timestamp: time.Now(),
		}

		err = bus.SendMessage(ctx, responseMessage)
		require.NoError(t, err)

		// Agent 1 receives response
		select {
		case receivedResponse := <-agent1Chan:
			assert.Equal(t, responseMessage.Content, receivedResponse.Content)
			assert.True(t, receivedResponse.Metadata["ready"].(bool))
		case <-time.After(1 * time.Second):
			t.Fatal("Agent 1 should have received response from Agent 2")
		}
	})
}
