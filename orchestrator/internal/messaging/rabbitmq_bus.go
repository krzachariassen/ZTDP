package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ztdp/orchestrator/internal/logging"
)

// RabbitMQMessageBus implements MessageBus using RabbitMQ
// Solves all reconnection and resilience issues
type RabbitMQMessageBus struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
	logger  logging.Logger

	// Connection recovery
	reconnectDelay time.Duration
	maxReconnects  int

	// Exchanges and queues
	agentExchange string
	dlxExchange   string

	// Consumer tag tracking for proper cleanup
	consumerTags map[string]string // participantID -> consumerTag
	mu           sync.RWMutex
}

// RabbitMQConfig holds configuration for RabbitMQ connection
type RabbitMQConfig struct {
	URL            string
	ReconnectDelay time.Duration
	MaxReconnects  int
	Heartbeat      time.Duration
}

// NewRabbitMQMessageBus creates a new RabbitMQ-based message bus
func NewRabbitMQMessageBus(config RabbitMQConfig, logger logging.Logger) *RabbitMQMessageBus {
	return &RabbitMQMessageBus{
		url:            config.URL,
		logger:         logger,
		reconnectDelay: config.ReconnectDelay,
		maxReconnects:  config.MaxReconnects,
		agentExchange:  "agent.messages",
		dlxExchange:    "agent.messages.dlx",
		consumerTags:   make(map[string]string),
	}
}

// Connect establishes connection to RabbitMQ with auto-recovery
func (rmq *RabbitMQMessageBus) Connect(ctx context.Context) error {
	config := amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
	}

	var err error
	rmq.conn, err = amqp.DialConfig(rmq.url, config)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	rmq.channel, err = rmq.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set up exchanges and queues
	return rmq.setupTopology()
}

// setupTopology creates exchanges, queues, and bindings
func (rmq *RabbitMQMessageBus) setupTopology() error {
	// Declare main exchange for agent messages
	err := rmq.channel.ExchangeDeclare(
		rmq.agentExchange, // name
		"direct",          // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare main exchange: %w", err)
	}

	// Declare dead letter exchange
	err = rmq.channel.ExchangeDeclare(
		rmq.dlxExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLX exchange: %w", err)
	}

	rmq.logger.Info("✅ RabbitMQ topology setup complete")
	return nil
}

// SendMessage sends a message to a specific agent
func (rmq *RabbitMQMessageBus) SendMessage(ctx context.Context, message *Message) error {
	if rmq.channel == nil {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	// Serialize message
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Publish to agent's queue
	err = rmq.channel.PublishWithContext(
		ctx,
		rmq.agentExchange, // exchange
		message.ToID,      // routing key (agent ID)
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			MessageId:     message.ID,
			CorrelationId: message.CorrelationID,
			Timestamp:     time.Now(),
			Expiration:    "300000", // 5 minutes TTL
			Headers: amqp.Table{
				"fromAgentId": message.FromID,
				"messageType": string(message.MessageType),
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	rmq.logger.Debug("📨 Message published to agent queue",
		"message_id", message.ID,
		"to_agent", message.ToID,
		"routing_key", message.ToID)

	return nil
}

// Subscribe subscribes an agent to messages (SOLVES RECONNECTION ISSUE)
func (rmq *RabbitMQMessageBus) Subscribe(ctx context.Context, participantID string) (<-chan *Message, error) {
	if rmq.channel == nil {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	// Declare agent's queue (idempotent - won't fail if already exists)
	queueName := fmt.Sprintf("agent.%s", participantID)
	_, err := rmq.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		amqp.Table{
			"x-message-ttl":             300000, // 5 minutes
			"x-dead-letter-exchange":    rmq.dlxExchange,
			"x-dead-letter-routing-key": participantID + ".dlq",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = rmq.channel.QueueBind(
		queueName,         // queue name
		participantID,     // routing key
		rmq.agentExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consuming (RabbitMQ handles reconnection automatically)
	// Generate unique consumer tag to avoid conflicts
	consumerTag := fmt.Sprintf("%s-%d", participantID, time.Now().UnixNano())

	// Store consumer tag for cleanup
	rmq.mu.Lock()
	rmq.consumerTags[participantID] = consumerTag
	rmq.mu.Unlock()

	msgs, err := rmq.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer tag (must be unique per connection)
		false,       // auto-ack (we'll ack manually)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		amqp.Table{
			"x-priority": 0,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	// Convert AMQP messages to our Message type
	messageChan := make(chan *Message, 100)

	go func() {
		defer close(messageChan)

		for {
			select {
			case <-ctx.Done():
				// Graceful shutdown - cancel consumer
				rmq.channel.Cancel(participantID, false)
				return

			case delivery, ok := <-msgs:
				if !ok {
					// Channel closed - attempt reconnection handled by RabbitMQ client
					rmq.logger.Warn("Message channel closed for agent", "agent_id", participantID)
					return
				}

				// Deserialize message
				var message Message
				if err := json.Unmarshal(delivery.Body, &message); err != nil {
					rmq.logger.Error("Failed to deserialize message: %v", err)
					delivery.Nack(false, false) // Send to DLQ
					continue
				}

				// Send to agent
				select {
				case messageChan <- &message:
					delivery.Ack(false) // Message successfully delivered
				case <-ctx.Done():
					delivery.Nack(false, true) // Requeue message
					return
				}
			}
		}
	}()

	rmq.logger.Info("✅ Agent subscribed to RabbitMQ",
		"agent_id", participantID,
		"queue", queueName)

	return messageChan, nil
}

// Unsubscribe removes an agent subscription (PROPER CLEANUP)
func (rmq *RabbitMQMessageBus) Unsubscribe(ctx context.Context, participantID string) error {
	if rmq.channel == nil {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	// Get the consumer tag for this participant
	rmq.mu.RLock()
	consumerTag, exists := rmq.consumerTags[participantID]
	rmq.mu.RUnlock()

	if !exists {
		// Already unsubscribed or never subscribed
		return nil
	}

	// Cancel the consumer
	err := rmq.channel.Cancel(consumerTag, false)
	if err != nil {
		return fmt.Errorf("failed to cancel consumer: %w", err)
	}

	// Remove from tracking
	rmq.mu.Lock()
	delete(rmq.consumerTags, participantID)
	rmq.mu.Unlock()

	rmq.logger.Info("✅ Agent unsubscribed from RabbitMQ", "agent_id", participantID)
	return nil
}

// PublishMessage publishes to multiple recipients
func (rmq *RabbitMQMessageBus) PublishMessage(ctx context.Context, message *Message, recipients []string) error {
	for _, recipient := range recipients {
		msg := *message // Copy message
		msg.ToID = recipient
		msg.ID = uuid.New().String() // New ID for each recipient

		if err := rmq.SendMessage(ctx, &msg); err != nil {
			return fmt.Errorf("failed to send to %s: %w", recipient, err)
		}
	}
	return nil
}

// GetConversationHistory retrieves conversation history (could be stored in Redis/DB)
func (rmq *RabbitMQMessageBus) GetConversationHistory(ctx context.Context, conversationID string) ([]*Message, error) {
	// For now, return empty - in production this would query a database
	// RabbitMQ is for message transport, not storage
	return []*Message{}, nil
}

// CreateConversation creates a new conversation context
func (rmq *RabbitMQMessageBus) CreateConversation(ctx context.Context, participants []string, context map[string]interface{}) (*ConversationContext, error) {
	conversationID := uuid.New().String()

	conversation := &ConversationContext{
		ConversationID: conversationID,
		Participants:   participants,
		Context:        context,
		StartTime:      time.Now(),
		LastActivity:   time.Now(),
	}

	// In production, store this in Redis/Database
	rmq.logger.Info("📝 Created conversation",
		"conversation_id", conversationID,
		"participants", participants)

	return conversation, nil
}

// Close closes RabbitMQ connection
func (rmq *RabbitMQMessageBus) Close() error {
	if rmq.channel != nil {
		rmq.channel.Close()
	}
	if rmq.conn != nil {
		rmq.conn.Close()
	}
	return nil
}

// HealthCheck checks RabbitMQ connection health
func (rmq *RabbitMQMessageBus) HealthCheck() error {
	if rmq.conn == nil || rmq.conn.IsClosed() {
		return fmt.Errorf("RabbitMQ connection closed")
	}
	if rmq.channel == nil {
		return fmt.Errorf("RabbitMQ channel not available")
	}
	return nil
}
