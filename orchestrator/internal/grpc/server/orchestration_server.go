package server

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ztdp/orchestrator/internal/agent/domain"
	"github.com/ztdp/orchestrator/internal/logging"
	"github.com/ztdp/orchestrator/internal/messaging"
	pb "github.com/ztdp/orchestrator/proto/orchestration"
)

// OrchestrationServer implements the gRPC OrchestrationService as a stateless proxy.
// It delegates:
// - Agent registration/unregistration to the registry service (domain logic)
// - Message streaming to the AI Message Bus (communication)
// It contains NO AI logic or business logic.
type OrchestrationServer struct {
	pb.UnimplementedOrchestrationServiceServer

	messageBus      messaging.AIMessageBus
	registryService domain.AgentRegistry
	logger          logging.Logger

	// Track active streams for cleanup
	activeStreams map[string]context.CancelFunc
	streamsMutex  sync.RWMutex
}

// NewOrchestrationServer creates a new gRPC server that acts as a stateless proxy
func NewOrchestrationServer(messageBus messaging.AIMessageBus, registryService domain.AgentRegistry, logger logging.Logger) *OrchestrationServer {
	return &OrchestrationServer{
		messageBus:      messageBus,
		registryService: registryService,
		logger:          logger,
		activeStreams:   make(map[string]context.CancelFunc),
	}
}

// RegisterAgent delegates agent registration to the registry service (domain logic)
func (s *OrchestrationServer) RegisterAgent(ctx context.Context, req *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	// Input validation
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request cannot be nil")
	}

	if req.AgentId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "agent ID cannot be empty")
	}

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "agent name cannot be empty")
	}

	if len(req.Capabilities) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "agent must have at least one capability")
	}

	s.logger.Info("Registering agent via gRPC",
		"agent_id", req.AgentId,
		"capabilities", req.Capabilities)

	// Convert gRPC message to internal domain.Agent format
	agent := &domain.Agent{
		ID:           req.AgentId,
		Name:         req.Name,
		Description:  "Agent registered via gRPC",
		Capabilities: convertCapabilities(req.Capabilities),
		Status:       domain.AgentStatusOnline,
		Metadata:     convertStructToStringMap(req.Metadata),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	// Delegate to registry service (domain logic)
	err := s.registryService.RegisterAgent(ctx, agent)
	if err != nil {
		s.logger.Error("Failed to register agent", err,
			"agent_id", req.AgentId)
		return nil, status.Errorf(codes.Internal, "failed to register agent: %v", err)
	}

	s.logger.Info("Successfully registered agent",
		"agent_id", req.AgentId)

	return &pb.RegisterAgentResponse{
		Success:      true,
		Message:      "Agent registered successfully",
		RegisteredAt: timestamppb.Now(),
	}, nil
}

// UnregisterAgent delegates agent unregistration to the registry service (domain logic)
func (s *OrchestrationServer) UnregisterAgent(ctx context.Context, req *pb.UnregisterAgentRequest) (*pb.UnregisterAgentResponse, error) {
	// Input validation
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request cannot be nil")
	}

	if req.AgentId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "agent ID cannot be empty")
	}

	s.logger.Info("Unregistering agent via gRPC",
		"agent_id", req.AgentId,
		"reason", req.Reason)

	// Clean up any active streams for this agent
	s.streamsMutex.Lock()
	if cancel, exists := s.activeStreams[req.AgentId]; exists {
		cancel()
		delete(s.activeStreams, req.AgentId)
	}
	s.streamsMutex.Unlock()

	// Delegate to registry service (domain logic)
	err := s.registryService.UnregisterAgent(ctx, req.AgentId)
	if err != nil {
		s.logger.Error("Failed to unregister agent", err,
			"agent_id", req.AgentId)
		return nil, status.Errorf(codes.Internal, "failed to unregister agent: %v", err)
	}

	s.logger.Info("Successfully unregistered agent",
		"agent_id", req.AgentId)

	return &pb.UnregisterAgentResponse{
		Success: true,
		Message: "Agent unregistered successfully",
	}, nil
}

// OpenConversation creates a bidirectional stream between the agent and AI Message Bus
func (s *OrchestrationServer) OpenConversation(stream pb.OrchestrationService_OpenConversationServer) error {
	ctx := stream.Context()

	s.logger.Info("Opening conversation stream")

	// Wait for the first message to identify the agent
	firstMsg, err := stream.Recv()
	if err != nil {
		s.logger.Error("Failed to receive first message", err)
		return status.Errorf(codes.InvalidArgument, "failed to receive agent identification: %v", err)
	}

	if firstMsg.FromAgentId == "" {
		return status.Errorf(codes.InvalidArgument, "agent ID is required in first message")
	}

	agentID := firstMsg.FromAgentId
	s.logger.Info("Agent opened conversation", "agent_id", agentID)

	// Subscribe to message bus for agent communication
	s.logger.Debug("Subscribing to message bus", "agent_id", agentID)
	messageChan, err := s.messageBus.Subscribe(ctx, agentID)
	if err != nil {
		s.logger.Error("Failed to subscribe to message bus", err, "agent_id", agentID)
		return status.Errorf(codes.Internal, "failed to subscribe to message bus: %v", err)
	}

	// Track this stream for cleanup
	streamCtx, cancel := context.WithCancel(ctx)
	s.streamsMutex.Lock()
	s.activeStreams[agentID] = cancel
	s.streamsMutex.Unlock()

	// Cleanup on exit
	defer func() {
		s.streamsMutex.Lock()
		if _, exists := s.activeStreams[agentID]; exists {
			cancel()
			delete(s.activeStreams, agentID)
		}
		s.streamsMutex.Unlock()
		s.logger.Info("Conversation stream closed", "agent_id", agentID)
	}()

	// Process the first message
	if err := s.processIncomingMessage(streamCtx, firstMsg); err != nil {
		s.logger.Error("Failed to process first message", err, "agent_id", agentID)
		return err
	}

	// Channel for incoming messages from the stream
	incomingChan := make(chan *pb.ConversationMessage, 10)
	errorChan := make(chan error, 1)

	// Goroutine to receive messages from the stream
	go func() {
		defer close(incomingChan)
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					s.logger.Debug("Stream closed by client", "agent_id", agentID)
					return
				}
				errorChan <- err
				return
			}

			select {
			case incomingChan <- msg:
			case <-streamCtx.Done():
				return
			}
		}
	}()

	// Main event loop - only for real agents
	for {
		// Real agents: Listen for both incoming messages and message bus
		select {
		case <-streamCtx.Done():
			s.logger.Debug("Stream context cancelled", "agent_id", agentID)
			return nil

		case err := <-errorChan:
			s.logger.Error("Stream error", err, "agent_id", agentID)
			return status.Errorf(codes.Internal, "stream error: %v", err)

		case msg := <-incomingChan:
			if msg == nil {
				// Channel closed, client disconnected
				return nil
			}

			if err := s.processIncomingMessage(streamCtx, msg); err != nil {
				s.logger.Error("Failed to process incoming message", err, "agent_id", agentID)
				// Continue processing other messages
			}

		case busMsg := <-messageChan:
			if busMsg == nil {
				// Real agent message bus closed - this is an error
				return status.Errorf(codes.Internal, "message bus closed")
			}

			// Convert message bus message to protobuf and send to agent
			pbMsg := s.convertToPbMessage(busMsg)
			if err := stream.Send(pbMsg); err != nil {
				s.logger.Error("Failed to send message to agent", err, "agent_id", agentID)
				return status.Errorf(codes.Internal, "failed to send message: %v", err)
			}
		}
	}
}

// processIncomingMessage handles messages received from the agent
func (s *OrchestrationServer) processIncomingMessage(ctx context.Context, msg *pb.ConversationMessage) error {
	s.logger.Debug("Processing incoming message",
		"from_agent", msg.FromAgentId,
		"to_agent", msg.ToAgentId,
		"type", msg.Type,
		"correlation_id", msg.CorrelationId)

	switch msg.Type {
	case pb.MessageType_MESSAGE_TYPE_AGENT_TO_AGENT:
		// Agent-to-agent message
		if msg.ToAgentId == "" {
			return status.Errorf(codes.InvalidArgument, "to_agent_id is required for agent-to-agent messages")
		}

		agentMsg := &messaging.AgentToAgentMessage{
			FromAgentID:   msg.FromAgentId,
			ToAgentID:     msg.ToAgentId,
			Content:       msg.Content,
			CorrelationID: msg.CorrelationId,
			Context:       convertStructToMap(msg.Context),
		}

		return s.messageBus.SendBetweenAgents(ctx, agentMsg)

	case pb.MessageType_MESSAGE_TYPE_CLARIFICATION:
		// Agent asking AI for clarification
		aiMsg := &messaging.AgentToAIMessage{
			AgentID:       msg.FromAgentId,
			Content:       msg.Content,
			MessageType:   messaging.MessageTypeClarification,
			CorrelationID: msg.CorrelationId,
			Context:       convertStructToMap(msg.Context),
		}

		return s.messageBus.SendToAI(ctx, aiMsg)

	case pb.MessageType_MESSAGE_TYPE_STATUS_UPDATE:
		// Agent status update
		aiMsg := &messaging.AgentToAIMessage{
			AgentID:       msg.FromAgentId,
			Content:       msg.Content,
			MessageType:   messaging.MessageTypeNotification,
			CorrelationID: msg.CorrelationId,
			Context:       convertStructToMap(msg.Context),
		}

		return s.messageBus.SendToAI(ctx, aiMsg)

	default:
		s.logger.Warn("Unknown message type", "type", msg.Type)
		return nil // Don't fail on unknown message types
	}
}

// convertToPbMessage converts internal message to protobuf message
func (s *OrchestrationServer) convertToPbMessage(msg *messaging.Message) *pb.ConversationMessage {
	return &pb.ConversationMessage{
		MessageId:     msg.ID,
		CorrelationId: msg.CorrelationID,
		FromAgentId:   msg.FromID,
		ToAgentId:     msg.ToID,
		Type:          convertMessageType(msg.MessageType),
		Content:       msg.Content,
		Context:       nil, // Simplified for now
		Timestamp:     timestamppb.New(msg.Timestamp),
	}
}

// convertMessageType converts internal message type to protobuf type
func convertMessageType(msgType messaging.MessageType) pb.MessageType {
	switch msgType {
	case messaging.MessageTypeClarification:
		return pb.MessageType_MESSAGE_TYPE_CLARIFICATION
	case messaging.MessageTypeNotification:
		return pb.MessageType_MESSAGE_TYPE_STATUS_UPDATE
	case messaging.MessageTypeAgentToAgent:
		return pb.MessageType_MESSAGE_TYPE_AGENT_TO_AGENT
	case messaging.MessageTypeAIToAgent:
		return pb.MessageType_MESSAGE_TYPE_AI_INSTRUCTION
	default:
		return pb.MessageType_MESSAGE_TYPE_UNKNOWN
	}
}

// Helper functions for struct conversion
func convertStructToMap(s interface{}) map[string]interface{} {
	if s == nil {
		return make(map[string]interface{})
	}

	// Check if it's a protobuf Struct
	if pbStruct, ok := s.(*structpb.Struct); ok {
		return pbStruct.AsMap()
	}

	// For other types, return empty map to avoid type errors
	return make(map[string]interface{})
}

func convertStructToStringMap(s interface{}) map[string]string {
	if s == nil {
		return make(map[string]string)
	}

	// Check if it's a protobuf Struct
	if pbStruct, ok := s.(*structpb.Struct); ok {
		result := make(map[string]string)
		for key, value := range pbStruct.AsMap() {
			result[key] = convertValueToString(value)
		}
		return result
	}

	// For other types, return empty map to avoid type errors
	return make(map[string]string)
}

// Helper function to convert any value to string
func convertValueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		// For complex types (arrays, objects), convert to string representation
		return fmt.Sprintf("%v", v)
	}
}

// SendMessage handles single message sends (non-streaming)
func (s *OrchestrationServer) SendMessage(ctx context.Context, req *pb.AgentMessage) (*pb.MessageResponse, error) {
	// Input validation
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request cannot be nil")
	}

	if req.FromAgentId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "from_agent_id cannot be empty")
	}

	if req.Content == "" {
		return nil, status.Errorf(codes.InvalidArgument, "content cannot be empty")
	}

	s.logger.Info("Processing single message send",
		"from_agent", req.FromAgentId,
		"to_agent", req.ToAgentId,
		"type", req.Type,
		"correlation_id", req.CorrelationId)

	// Route message based on type
	switch req.Type {
	case pb.MessageType_MESSAGE_TYPE_AGENT_TO_AGENT:
		// Agent-to-agent message
		if req.ToAgentId == "" {
			return nil, status.Errorf(codes.InvalidArgument, "to_agent_id is required for agent-to-agent messages")
		}

		agentMsg := &messaging.AgentToAgentMessage{
			FromAgentID:   req.FromAgentId,
			ToAgentID:     req.ToAgentId,
			Content:       req.Content,
			CorrelationID: req.CorrelationId,
			Context:       convertStructToMap(req.Context),
		}

		err := s.messageBus.SendBetweenAgents(ctx, agentMsg)
		if err != nil {
			s.logger.Error("Failed to send agent-to-agent message", err,
				"from_agent", req.FromAgentId,
				"to_agent", req.ToAgentId)
			return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
		}

	case pb.MessageType_MESSAGE_TYPE_CLARIFICATION:
		// Agent asking AI for clarification
		aiMsg := &messaging.AgentToAIMessage{
			AgentID:       req.FromAgentId,
			Content:       req.Content,
			MessageType:   messaging.MessageTypeClarification,
			CorrelationID: req.CorrelationId,
			Context:       convertStructToMap(req.Context),
		}

		err := s.messageBus.SendToAI(ctx, aiMsg)
		if err != nil {
			s.logger.Error("Failed to send agent-to-AI clarification", err,
				"agent_id", req.FromAgentId)
			return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
		}

	case pb.MessageType_MESSAGE_TYPE_STATUS_UPDATE:
		// Agent status update to AI
		aiMsg := &messaging.AgentToAIMessage{
			AgentID:       req.FromAgentId,
			Content:       req.Content,
			MessageType:   messaging.MessageTypeNotification,
			CorrelationID: req.CorrelationId,
			Context:       convertStructToMap(req.Context),
		}

		err := s.messageBus.SendToAI(ctx, aiMsg)
		if err != nil {
			s.logger.Error("Failed to send agent status update", err,
				"agent_id", req.FromAgentId)
			return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
		}

	default:
		s.logger.Warn("Unknown message type for SendMessage", "type", req.Type)
		return nil, status.Errorf(codes.InvalidArgument, "unsupported message type: %v", req.Type)
	}

	s.logger.Debug("Message sent successfully",
		"from_agent", req.FromAgentId,
		"correlation_id", req.CorrelationId)

	return &pb.MessageResponse{
		Success:       true,
		Message:       "Message sent successfully",
		CorrelationId: req.CorrelationId,
	}, nil
}

// Legacy methods - these should probably be deprecated in favor of AI-native approach
func (s *OrchestrationServer) PullWork(stream pb.OrchestrationService_PullWorkServer) error {
	return status.Errorf(codes.Unimplemented, "PullWork is deprecated - use OpenConversation for AI-native interaction")
}

func (s *OrchestrationServer) ReportResult(ctx context.Context, req *pb.ReportResultRequest) (*pb.ReportResultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ReportResult is deprecated - use OpenConversation for AI-native interaction")
}

func (s *OrchestrationServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	// Input validation
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request cannot be nil")
	}

	if req.AgentId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "agent ID is required")
	}

	// Convert protobuf status to string
	statusStr := "healthy"
	switch req.Status {
	case pb.AgentStatus_AGENT_STATUS_HEALTHY:
		statusStr = "healthy"
	case pb.AgentStatus_AGENT_STATUS_BUSY:
		statusStr = "busy"
	case pb.AgentStatus_AGENT_STATUS_ERROR:
		statusStr = "error"
	case pb.AgentStatus_AGENT_STATUS_SHUTTING_DOWN:
		statusStr = "shutting_down"
	default:
		statusStr = "unknown"
	}

	// Update heartbeat in registry - update last seen time
	if err := s.registryService.UpdateAgentLastSeen(ctx, req.AgentId); err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to update agent heartbeat", err, "agent_id", req.AgentId)
		}
		return &pb.HeartbeatResponse{
			Success:    false,
			ServerTime: timestamppb.Now(),
		}, status.Errorf(codes.Internal, "failed to update heartbeat: %v", err)
	}

	if s.logger != nil {
		s.logger.Debug("Agent heartbeat received", "agent_id", req.AgentId, "status", statusStr)
	}

	return &pb.HeartbeatResponse{
		Success:    true,
		ServerTime: timestamppb.Now(),
	}, nil
}

func (s *OrchestrationServer) RequestClarification(ctx context.Context, req *pb.ClarificationRequest) (*pb.ClarificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "RequestClarification is deprecated - use OpenConversation for AI-native interaction")
}

// Helper functions

// convertCapabilities converts protobuf capabilities to domain capabilities
func convertCapabilities(pbCapabilities []string) []domain.AgentCapability {
	capabilities := make([]domain.AgentCapability, len(pbCapabilities))
	for i, cap := range pbCapabilities {
		capabilities[i] = domain.AgentCapability{
			Name:        cap,
			Description: "Capability: " + cap,
		}
	}
	return capabilities
}
