// Package agent provides a simple SDK for building ZTDP agents
package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/ztdp/agents/text-processor/proto/orchestration"
)

// Handler defines the interface for processing agent tasks
type Handler interface {
	Process(ctx context.Context, task Task) (*Result, error)
	GetCapabilities() []string
}

// Task represents work sent to the agent
type Task struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Content string                 `json:"content"`
	Context map[string]interface{} `json:"context"`
}

// Result represents the agent's response
type Result struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Message string                 `json:"message"`
	Error   string                 `json:"error,omitempty"`
}

// Agent represents a ZTDP agent
type Agent struct {
	ID           string
	Name         string
	Capabilities []string
	Handler      Handler

	// gRPC connection
	conn   *grpc.ClientConn
	client pb.OrchestrationServiceClient

	// Internal
	ctx    context.Context
	cancel context.CancelFunc
}

// Config holds agent configuration
type Config struct {
	OrchestratorAddress string
	ReconnectInterval   time.Duration
}

// NewAgent creates a new agent with the given ID and handler
func NewAgent(id, name string, handler Handler) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		ID:           id,
		Name:         name,
		Capabilities: handler.GetCapabilities(),
		Handler:      handler,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start begins the agent's lifecycle
func (a *Agent) Start(config Config) error {
	log.Printf("🤖 Starting agent: %s (%s)", a.Name, a.ID)

	// Set default config
	if config.OrchestratorAddress == "" {
		config.OrchestratorAddress = "localhost:50051"
	}
	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 30 * time.Second
	}

	// Connect to orchestrator
	if err := a.connect(config.OrchestratorAddress); err != nil {
		return fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	// Register with orchestrator
	if err := a.register(); err != nil {
		return fmt.Errorf("failed to register with orchestrator: %w", err)
	}

	// Start work loop
	go a.workLoop()

	log.Printf("✅ Agent %s is ready and listening for work", a.Name)
	return nil
}

// Stop gracefully shuts down the agent
func (a *Agent) Stop() error {
	log.Printf("🛑 Stopping agent: %s", a.Name)

	a.cancel()

	if a.conn != nil {
		return a.conn.Close()
	}

	return nil
}

// connect establishes gRPC connection to orchestrator
func (a *Agent) connect(address string) error {
	log.Printf("🔗 Connecting to orchestrator at %s", address)

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	a.conn = conn
	a.client = pb.NewOrchestrationServiceClient(conn)

	log.Printf("✅ Connected to orchestrator")
	return nil
}

// register registers the agent with the orchestrator
func (a *Agent) register() error {
	log.Printf("📝 Registering agent with capabilities: %v", a.Capabilities)

	// Check if client is nil
	if a.client == nil {
		return fmt.Errorf("gRPC client is nil - connection failed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.RegisterAgentRequest{
		AgentId:           a.ID,
		Name:              a.Name,
		Type:              "text-processor",
		Capabilities:      a.Capabilities,
		Version:           "1.0.0",
		MaxConcurrentWork: 5,
	}

	log.Printf("🔄 Making gRPC call to RegisterAgent...")
	resp, err := a.client.RegisterAgent(ctx, req)
	if err != nil {
		log.Printf("❌ gRPC call failed: %v", err)
		return fmt.Errorf("failed to register agent: %w", err)
	}

	log.Printf("🔄 Received response: Success=%v, Message=%s", resp.Success, resp.Message)
	if !resp.Success {
		return fmt.Errorf("agent registration failed: %s", resp.Message)
	}

	log.Printf("✅ Agent registered successfully: %s", resp.Message)
	return nil
}

// workLoop continuously listens for conversations from the orchestrator
func (a *Agent) workLoop() {
	log.Printf("👂 Starting conversation loop for agent %s", a.Name)

	for {
		select {
		case <-a.ctx.Done():
			log.Printf("🛑 Conversation loop stopping for agent %s", a.Name)
			return
		default:
			// Try to open conversation with orchestrator
			if err := a.openConversation(); err != nil {
				log.Printf("⚠️ Conversation ended: %v", err)
				// Wait before retrying
				time.Sleep(10 * time.Second)
			}
		}
	}
}

// openConversation opens a conversation stream with the orchestrator
func (a *Agent) openConversation() error {
	ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
	defer cancel()

	// Create the conversation stream
	stream, err := a.client.OpenConversation(ctx)
	if err != nil {
		return fmt.Errorf("failed to start conversation stream: %w", err)
	}

	// Send initial identification message as required by orchestrator
	initMsg := &pb.ConversationMessage{
		MessageId:   fmt.Sprintf("%s-init-%d", a.ID, time.Now().UnixNano()),
		FromAgentId: a.ID,
		ToAgentId:   "", // Empty for AI
		Type:        pb.MessageType_MESSAGE_TYPE_STATUS_UPDATE,
		Content:     "Agent connected and ready for instructions",
		Timestamp:   timestamppb.Now(),
	}

	if err := stream.Send(initMsg); err != nil {
		return fmt.Errorf("failed to send identification message: %w", err)
	}

	log.Printf("🔄 Opened conversation stream with orchestrator, listening for AI instructions...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					log.Printf("📭 Conversation stream ended")
					return nil
				}
				return fmt.Errorf("failed to receive message: %w", err)
			}

			// Process the conversation message
			go a.processConversationMessage(message)
		}
	}
}

// processConversationMessage processes a conversation message from the orchestrator
func (a *Agent) processConversationMessage(msg *pb.ConversationMessage) {
	log.Printf("💬 Received AI instruction: %s", msg.Content)

	// Only process AI instructions targeted at this agent
	if msg.Type == pb.MessageType_MESSAGE_TYPE_AI_INSTRUCTION {

		// Parse the message content to extract task information
		task := Task{
			ID:      msg.MessageId,
			Type:    a.extractTaskType(msg.Content),
			Content: msg.Content,
			Context: make(map[string]interface{}),
		}

		// Add message context
		if msg.Context != nil {
			task.Context = msg.Context.AsMap()
		}

		// Add correlation ID for response tracking
		task.Context["correlation_id"] = msg.CorrelationId
		task.Context["from_agent_id"] = msg.FromAgentId

		// Process the task
		result := a.processTask(task)

		log.Printf("✅ Task result: %s", a.formatResult(result))

		// For now, just log the result - don't send it back to avoid feedback loops
		// TODO: Implement proper result reporting mechanism

	} else {
		log.Printf("ℹ️ Ignoring message type: %s", msg.Type.String())
	}
}

// extractTaskType extracts the task type from message content
func (a *Agent) extractTaskType(content string) string {
	content = strings.ToLower(content)

	// Simple keyword matching for now
	if strings.Contains(content, "count") && strings.Contains(content, "word") {
		return "word-count"
	}
	if strings.Contains(content, "count") && strings.Contains(content, "character") {
		return "character-count"
	}
	if strings.Contains(content, "analyze") || strings.Contains(content, "analysis") {
		return "text-analysis"
	}
	if strings.Contains(content, "format") || strings.Contains(content, "uppercase") ||
		strings.Contains(content, "lowercase") || strings.Contains(content, "title") {
		return "text-formatting"
	}
	if strings.Contains(content, "clean") || strings.Contains(content, "cleanup") {
		return "text-cleanup"
	}

	// Default to text analysis
	return "text-analysis"
}

// formatResult formats the task result for logging
func (a *Agent) formatResult(result *Result) string {
	if !result.Success {
		return fmt.Sprintf("❌ Task failed: %s", result.Error)
	}

	// Format based on result data
	if result.Data != nil {
		if wordCount, ok := result.Data["word_count"].(int); ok {
			return fmt.Sprintf("Word count: %d words", wordCount)
		}
		if charCount, ok := result.Data["character_count"].(int); ok {
			return fmt.Sprintf("Character count: %d characters", charCount)
		}
		if analysis, ok := result.Data["analysis"].(string); ok {
			return fmt.Sprintf("Text analysis: %s", analysis)
		}
		if formatted, ok := result.Data["formatted_text"].(string); ok {
			return fmt.Sprintf("Formatted text: %s", formatted)
		}
		if cleaned, ok := result.Data["cleaned_text"].(string); ok {
			return fmt.Sprintf("Cleaned text: %s", cleaned)
		}
	}

	return fmt.Sprintf("Task completed: %s", result.Message)
}

// sendResult sends the task result back to the orchestrator
// processTask handles a single task
func (a *Agent) processTask(task Task) *Result {
	log.Printf("⚡ Processing task: %s (type: %s)", task.ID, task.Type)

	// Call the handler
	result, err := a.Handler.Process(a.ctx, task)
	if err != nil {
		log.Printf("❌ Task failed: %v", err)
		return &Result{
			Success: false,
			Error:   err.Error(),
			Message: fmt.Sprintf("Task %s failed", task.ID),
		}
	}

	log.Printf("✅ Task completed successfully: %s", task.ID)
	return result
}

// SendMessage sends a message to the orchestrator (for agent-AI communication)
func (a *Agent) SendMessage(content string) error {
	log.Printf("💬 Sending message to orchestrator: %s", content)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.AgentMessage{
		FromAgentId:   a.ID,
		ToAgentId:     "", // Empty for AI
		CorrelationId: fmt.Sprintf("%s-%d", a.ID, time.Now().UnixNano()),
		Content:       content,
		Type:          pb.MessageType_MESSAGE_TYPE_STATUS_UPDATE,
	}

	resp, err := a.client.SendMessage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("message sending failed: %s", resp.Message)
	}

	log.Printf("✅ Message sent successfully")
	return nil
}

// WaitForShutdown blocks until the agent is shut down
func (a *Agent) WaitForShutdown() {
	<-a.ctx.Done()
}
