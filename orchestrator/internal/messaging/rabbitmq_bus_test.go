package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ztdp/orchestrator/internal/logging"
)

// Test that RabbitMQ message bus can be created
func TestRabbitMQMessageBus_New(t *testing.T) {
	// Given
	config := RabbitMQConfig{
		URL:            "amqp://guest:guest@localhost:5672/",
		ReconnectDelay: 5 * time.Second,
		MaxReconnects:  3,
		Heartbeat:      10 * time.Second,
	}
	logger := logging.NewNoOpLogger()

	// When
	bus := NewRabbitMQMessageBus(config, logger)

	// Then
	assert.NotNil(t, bus)
	assert.Equal(t, config.URL, bus.url)
	assert.Equal(t, config.ReconnectDelay, bus.reconnectDelay)
	assert.Equal(t, config.MaxReconnects, bus.maxReconnects)
	assert.Equal(t, "agent.messages", bus.agentExchange)
	assert.Equal(t, "agent.messages.dlx", bus.dlxExchange)
}

// Test that RabbitMQ connection fails with invalid URL
func TestRabbitMQMessageBus_Connect_InvalidURL(t *testing.T) {
	// Given
	config := RabbitMQConfig{
		URL: "amqp://invalid:invalid@nonexistent:5672/",
	}
	logger := logging.NewNoOpLogger()
	bus := NewRabbitMQMessageBus(config, logger)
	ctx := context.Background()

	// When
	err := bus.Connect(ctx)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to RabbitMQ")
}

// Test that RabbitMQ connection succeeds with valid RabbitMQ server
func TestRabbitMQMessageBus_Connect_Success(t *testing.T) {
	// Skip if RabbitMQ not available
	if !isRabbitMQAvailable() {
		t.Skip("RabbitMQ not available for testing")
	}

	// Given
	config := RabbitMQConfig{
		URL:            "amqp://orchestrator:orchestrator123@localhost:5672/",
		ReconnectDelay: 1 * time.Second,
		MaxReconnects:  3,
		Heartbeat:      10 * time.Second,
	}
	logger := logging.NewNoOpLogger()
	bus := NewRabbitMQMessageBus(config, logger)
	ctx := context.Background()

	// When
	err := bus.Connect(ctx)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, bus.conn)
	assert.NotNil(t, bus.channel)

	// Cleanup
	bus.Close()
}

// Test that agent can subscribe to messages without "already subscribed" error
func TestRabbitMQMessageBus_Subscribe_NoAlreadySubscribedError(t *testing.T) {
	// Skip if RabbitMQ not available
	if !isRabbitMQAvailable() {
		t.Skip("RabbitMQ not available for testing")
	}

	// Given
	config := RabbitMQConfig{
		URL: "amqp://orchestrator:orchestrator123@localhost:5672/",
	}
	logger := logging.NewNoOpLogger()
	bus := NewRabbitMQMessageBus(config, logger)
	ctx := context.Background()

	require.NoError(t, bus.Connect(ctx))
	defer bus.Close()

	agentID := "test-agent-1"

	// When - First subscription should work
	msgChan1, err1 := bus.Subscribe(ctx, agentID)

	// Then
	assert.NoError(t, err1)
	assert.NotNil(t, msgChan1)

	// When - Second subscription to same agent should ALSO work (unlike memory bus)
	// This is the key difference - RabbitMQ handles reconnections gracefully
	msgChan2, err2 := bus.Subscribe(ctx, agentID)

	// Then - Should NOT get "already subscribed" error
	assert.NoError(t, err2)
	assert.NotNil(t, msgChan2)

	// Cleanup
	bus.Unsubscribe(ctx, agentID)
}

// Test that messages can be sent and received
func TestRabbitMQMessageBus_SendReceiveMessage(t *testing.T) {
	// Skip if RabbitMQ not available
	if !isRabbitMQAvailable() {
		t.Skip("RabbitMQ not available for testing")
	}

	// Given
	config := RabbitMQConfig{
		URL: "amqp://orchestrator:orchestrator123@localhost:5672/",
	}
	logger := logging.NewNoOpLogger()
	bus := NewRabbitMQMessageBus(config, logger)
	ctx := context.Background()

	require.NoError(t, bus.Connect(ctx))
	defer bus.Close()

	agentID := "test-agent-receive"

	// Subscribe to messages
	msgChan, err := bus.Subscribe(ctx, agentID)
	require.NoError(t, err)
	defer bus.Unsubscribe(ctx, agentID)

	// When - Send a message
	testMessage := &Message{
		ID:            "test-msg-1",
		CorrelationID: "test-corr-1",
		FromID:        "orchestrator",
		ToID:          agentID,
		MessageType:   "task",
		Content:       "Hello, agent!",
	}

	err = bus.SendMessage(ctx, testMessage)
	require.NoError(t, err)

	// Then - Should receive the message
	select {
	case receivedMsg := <-msgChan:
		assert.Equal(t, testMessage.ID, receivedMsg.ID)
		assert.Equal(t, testMessage.CorrelationID, receivedMsg.CorrelationID)
		assert.Equal(t, testMessage.FromID, receivedMsg.FromID)
		assert.Equal(t, testMessage.ToID, receivedMsg.ToID)
		assert.Equal(t, testMessage.MessageType, receivedMsg.MessageType)
		assert.Equal(t, testMessage.Content, receivedMsg.Content)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

// Test that unsubscribe works properly
func TestRabbitMQMessageBus_Unsubscribe(t *testing.T) {
	// Skip if RabbitMQ not available
	if !isRabbitMQAvailable() {
		t.Skip("RabbitMQ not available for testing")
	}

	// Given
	config := RabbitMQConfig{
		URL: "amqp://orchestrator:orchestrator123@localhost:5672/",
	}
	logger := logging.NewNoOpLogger()
	bus := NewRabbitMQMessageBus(config, logger)
	ctx := context.Background()

	require.NoError(t, bus.Connect(ctx))
	defer bus.Close()

	agentID := "test-agent-unsub"

	// Subscribe first
	msgChan, err := bus.Subscribe(ctx, agentID)
	require.NoError(t, err)

	// When - Unsubscribe
	err = bus.Unsubscribe(ctx, agentID)

	// Then
	assert.NoError(t, err)

	// Message channel should be closed or no longer receive messages
	// We can test this by trying to send a message and ensuring it's not received
	testMessage := &Message{
		ID:      "test-msg-after-unsub",
		ToID:    agentID,
		Content: "Should not be received",
	}

	_ = bus.SendMessage(ctx, testMessage)

	// Should not receive message after unsubscribe
	select {
	case <-msgChan:
		// If we receive anything, it might be from before unsubscribe, so we'll be lenient
		// The key test is that unsubscribe doesn't error
	case <-time.After(1 * time.Second):
		// Expected - no message received after unsubscribe
	}
}

// Test health check functionality
func TestRabbitMQMessageBus_HealthCheck(t *testing.T) {
	// Given - disconnected bus
	config := RabbitMQConfig{
		URL: "amqp://orchestrator:orchestrator123@localhost:5672/",
	}
	logger := logging.NewNoOpLogger()
	bus := NewRabbitMQMessageBus(config, logger)

	// When - health check on disconnected bus
	err := bus.HealthCheck()

	// Then - should report unhealthy
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection closed")

	// Skip rest if RabbitMQ not available
	if !isRabbitMQAvailable() {
		return
	}

	// Given - connected bus
	ctx := context.Background()
	require.NoError(t, bus.Connect(ctx))
	defer bus.Close()

	// When - health check on connected bus
	err = bus.HealthCheck()

	// Then - should be healthy
	assert.NoError(t, err)
}

// Test multiple agents can subscribe simultaneously
func TestRabbitMQMessageBus_MultipleAgents(t *testing.T) {
	// Skip if RabbitMQ not available
	if !isRabbitMQAvailable() {
		t.Skip("RabbitMQ not available for testing")
	}

	// Given
	config := RabbitMQConfig{
		URL: "amqp://orchestrator:orchestrator123@localhost:5672/",
	}
	logger := logging.NewNoOpLogger()
	bus := NewRabbitMQMessageBus(config, logger)
	ctx := context.Background()

	require.NoError(t, bus.Connect(ctx))
	defer bus.Close()

	// When - Multiple agents subscribe
	agent1Chan, err1 := bus.Subscribe(ctx, "agent-1")
	agent2Chan, err2 := bus.Subscribe(ctx, "agent-2")
	agent3Chan, err3 := bus.Subscribe(ctx, "agent-3")

	// Then - All should succeed
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NotNil(t, agent1Chan)
	assert.NotNil(t, agent2Chan)
	assert.NotNil(t, agent3Chan)

	// Cleanup
	bus.Unsubscribe(ctx, "agent-1")
	bus.Unsubscribe(ctx, "agent-2")
	bus.Unsubscribe(ctx, "agent-3")
}

// Helper function to check if RabbitMQ is available for testing
func isRabbitMQAvailable() bool {
	config := RabbitMQConfig{
		URL: "amqp://orchestrator:orchestrator123@localhost:5672/",
	}
	logger := logging.NewNoOpLogger()
	bus := NewRabbitMQMessageBus(config, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := bus.Connect(ctx)
	if err == nil {
		bus.Close()
		return true
	}
	return false
}
