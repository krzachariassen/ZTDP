package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ztdp/orchestrator/internal/graph"
	"github.com/ztdp/orchestrator/internal/logging"
)

// Simple mock graph for messaging tests
type mockGraph struct{}

func (m *mockGraph) AddNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error {
	return nil
}

func (m *mockGraph) GetNode(ctx context.Context, nodeType, nodeID string) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (m *mockGraph) UpdateNode(ctx context.Context, nodeType, nodeID string, properties map[string]interface{}) error {
	return nil
}

func (m *mockGraph) DeleteNode(ctx context.Context, nodeType, nodeID string) error {
	return nil
}

func (m *mockGraph) QueryNodes(ctx context.Context, nodeType string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *mockGraph) AddEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, edgeType string, properties map[string]interface{}) error {
	return nil
}

func (m *mockGraph) GetEdges(ctx context.Context, nodeType, nodeID string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *mockGraph) UpdateEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, edgeType string, properties map[string]interface{}) error {
	return nil
}

func (m *mockGraph) DeleteEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, edgeType string) error {
	return nil
}

func (m *mockGraph) GetStats() map[string]interface{} {
	return make(map[string]interface{})
}

func (m *mockGraph) Close(ctx context.Context) error {
	return nil
}

func newMockGraph() graph.Graph {
	return &mockGraph{}
}

func TestAIMessageBus_TDD(t *testing.T) {
	t.Run("can_create_ai_message_bus", func(t *testing.T) {
		// Setup dependencies
		messageBus := NewMemoryMessageBus(logging.NewNoOpLogger())
		mockGraph := newMockGraph()

		// Create AI message bus
		aiMessageBus := NewAIMessageBus(messageBus, mockGraph, &TestLogger{t: t})
		require.NotNil(t, aiMessageBus)
	})

	t.Run("ai_can_send_instructions_to_agent", func(t *testing.T) {
		// Setup
		messageBus := NewMemoryMessageBus(logging.NewNoOpLogger())
		mockGraph := newMockGraph()
		aiMessageBus := NewAIMessageBus(messageBus, mockGraph, &TestLogger{t: t})

		ctx := context.Background()
		agentID := "deployment-agent"

		// Agent subscribes to AI messages
		agentChan, err := aiMessageBus.Subscribe(ctx, agentID)
		require.NoError(t, err)

		// AI sends instruction to agent
		instruction := &AIToAgentMessage{
			AgentID:       agentID,
			Content:       "Deploy application 'web-app' to production environment",
			Intent:        "deployment",
			CorrelationID: "workflow-123",
			Context: map[string]interface{}{
				"application": "web-app",
				"environment": "production",
				"version":     "1.2.3",
			},
		}

		err = aiMessageBus.SendToAgent(ctx, instruction)
		require.NoError(t, err)

		// Agent receives instruction
		select {
		case message := <-agentChan:
			assert.Equal(t, instruction.Content, message.Content)
			assert.Equal(t, MessageTypeAIToAgent, message.MessageType)
			assert.Equal(t, agentID, message.ToID)
			assert.Equal(t, "ai-orchestrator", message.FromID)
		case <-time.After(1 * time.Second):
			t.Fatal("Agent should have received AI instruction")
		}
	})

	t.Run("agent_can_request_clarification_from_ai", func(t *testing.T) {
		// Setup
		messageBus := NewMemoryMessageBus(logging.NewNoOpLogger())
		mockGraph := newMockGraph()
		aiMessageBus := NewAIMessageBus(messageBus, mockGraph, &TestLogger{t: t})

		ctx := context.Background()
		agentID := "deployment-agent"
		aiID := "ai-orchestrator"

		// Both subscribe
		agentChan, err := aiMessageBus.Subscribe(ctx, agentID)
		require.NoError(t, err)
		aiChan, err := aiMessageBus.Subscribe(ctx, aiID)
		require.NoError(t, err)

		// Agent requests clarification
		clarification := &AgentToAIMessage{
			AgentID:       agentID,
			Content:       "What is the desired number of replicas for the production deployment?",
			MessageType:   MessageTypeClarification,
			CorrelationID: "workflow-123",
			Context: map[string]interface{}{
				"step":        "deployment",
				"missing":     "replica_count",
				"application": "web-app",
			},
		}

		err = aiMessageBus.SendToAI(ctx, clarification)
		require.NoError(t, err)

		// AI receives clarification request
		select {
		case message := <-aiChan:
			assert.Equal(t, clarification.Content, message.Content)
			assert.Equal(t, MessageTypeClarification, message.MessageType)
			assert.Equal(t, agentID, message.FromID)
			assert.Equal(t, aiID, message.ToID)
		case <-time.After(1 * time.Second):
			t.Fatal("AI should have received clarification request")
		}

		// AI responds with clarification
		response := &AIToAgentMessage{
			AgentID:       agentID,
			Content:       "Use 3 replicas for production deployment",
			Intent:        "clarification_response",
			CorrelationID: "workflow-123",
			Context: map[string]interface{}{
				"replica_count": 3,
				"response_to":   clarification.MessageType,
			},
		}

		err = aiMessageBus.SendToAgent(ctx, response)
		require.NoError(t, err)

		// Agent receives clarification response
		select {
		case message := <-agentChan:
			assert.Equal(t, response.Content, message.Content)
			assert.Equal(t, 3, message.Metadata["replica_count"])
		case <-time.After(1 * time.Second):
			t.Fatal("Agent should have received clarification response")
		}
	})

	t.Run("agents_can_communicate_through_ai_mediation", func(t *testing.T) {
		// Setup
		messageBus := NewMemoryMessageBus(logging.NewNoOpLogger())
		mockGraph := newMockGraph()
		aiMessageBus := NewAIMessageBus(messageBus, mockGraph, &TestLogger{t: t})

		ctx := context.Background()
		buildAgentID := "build-agent"
		deployAgentID := "deploy-agent"

		// Both agents subscribe
		_, err := aiMessageBus.Subscribe(ctx, buildAgentID)
		require.NoError(t, err)
		deployChan, err := aiMessageBus.Subscribe(ctx, deployAgentID)
		require.NoError(t, err)

		// Build agent wants to communicate with deploy agent
		agentMessage := &AgentToAgentMessage{
			FromAgentID:   buildAgentID,
			ToAgentID:     deployAgentID,
			Content:       "Build artifact ready for deployment. Size: 150MB, Location: /artifacts/web-app-v1.2.3.tar.gz",
			CorrelationID: "workflow-123",
			Context: map[string]interface{}{
				"artifact_path": "/artifacts/web-app-v1.2.3.tar.gz",
				"size_mb":       150,
				"checksum":      "sha256:abc123...",
			},
		}

		err = aiMessageBus.SendBetweenAgents(ctx, agentMessage)
		require.NoError(t, err)

		// Deploy agent receives the message (AI mediates routing)
		select {
		case message := <-deployChan:
			assert.Equal(t, agentMessage.Content, message.Content)
			assert.Equal(t, MessageTypeAgentToAgent, message.MessageType)
			assert.Equal(t, buildAgentID, message.FromID)
			assert.Equal(t, deployAgentID, message.ToID)
			assert.Equal(t, "/artifacts/web-app-v1.2.3.tar.gz", message.Metadata["artifact_path"])
		case <-time.After(1 * time.Second):
			t.Fatal("Deploy agent should have received message from build agent")
		}
	})

	t.Run("ai_stores_conversation_context_in_graph", func(t *testing.T) {
		// Setup
		messageBus := NewMemoryMessageBus(logging.NewNoOpLogger())
		mockGraph := newMockGraph()
		aiMessageBus := NewAIMessageBus(messageBus, mockGraph, &TestLogger{t: t})

		ctx := context.Background()
		agentID := "test-agent"

		// Agent subscribes
		_, err := aiMessageBus.Subscribe(ctx, agentID)
		require.NoError(t, err)

		// Send message
		instruction := &AIToAgentMessage{
			AgentID:       agentID,
			Content:       "Test message for graph storage",
			Intent:        "test",
			CorrelationID: "test-conversation",
			Context: map[string]interface{}{
				"test": "graph_storage",
			},
		}

		err = aiMessageBus.SendToAgent(ctx, instruction)
		require.NoError(t, err)

		// Verify conversation history can be retrieved
		history, err := aiMessageBus.GetConversationHistory(ctx, "test-conversation")
		require.NoError(t, err)
		require.NotEmpty(t, history)

		// Verify message was stored with correct structure
		storedMessage := history[0]
		assert.Equal(t, instruction.Content, storedMessage.Content)
		assert.Equal(t, "test-conversation", storedMessage.CorrelationID)
	})

	t.Run("ai_handles_user_requests", func(t *testing.T) {
		// Setup
		messageBus := NewMemoryMessageBus(logging.NewNoOpLogger())
		mockGraph := newMockGraph()
		aiMessageBus := NewAIMessageBus(messageBus, mockGraph, &TestLogger{t: t})

		ctx := context.Background()
		userID := "user-123"
		aiID := "ai-orchestrator"

		// AI subscribes to user messages
		aiChan, err := aiMessageBus.Subscribe(ctx, aiID)
		require.NoError(t, err)

		// User sends request to AI
		userRequest := &UserToAIMessage{
			UserID:        userID,
			Content:       "I need to deploy my web application to production",
			CorrelationID: "user-session-456",
			Context: map[string]interface{}{
				"application_type": "web",
				"target_env":       "production",
				"urgency":          "normal",
			},
		}

		err = aiMessageBus.SendUserToAI(ctx, userRequest)
		require.NoError(t, err)

		// AI receives user request
		select {
		case message := <-aiChan:
			assert.Equal(t, userRequest.Content, message.Content)
			assert.Equal(t, MessageTypeRequest, message.MessageType)
			assert.Equal(t, userID, message.FromID)
			assert.Equal(t, aiID, message.ToID)
			assert.Equal(t, "production", message.Metadata["target_env"])
		case <-time.After(1 * time.Second):
			t.Fatal("AI should have received user request")
		}
	})
}

// Test logger for AI message bus tests
type TestLogger struct {
	t *testing.T
}

func (l *TestLogger) Info(msg string, keysAndValues ...interface{}) {
	l.t.Logf("INFO: %s %v", msg, keysAndValues)
}

func (l *TestLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.t.Logf("DEBUG: %s %v", msg, keysAndValues)
}

func (l *TestLogger) Error(msg string, err error, keysAndValues ...interface{}) {
	l.t.Logf("ERROR: %s: %v %v", msg, err, keysAndValues)
}

func (l *TestLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.t.Logf("WARN: %s %v", msg, keysAndValues)
}
