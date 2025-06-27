package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ztdp/orchestrator/internal/conversation/domain"
)

// ConversationServiceImpl implements the ConversationService interface
type ConversationServiceImpl struct {
	repo ConversationRepository
}

// NewConversationServiceImpl creates a new conversation service implementation
func NewConversationServiceImpl(repo ConversationRepository) ConversationService {
	return &ConversationServiceImpl{
		repo: repo,
	}
}

// CreateConversation creates a new conversation with validation
func (s *ConversationServiceImpl) CreateConversation(ctx context.Context, userID string) (*domain.Conversation, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	// Generate a unique ID for the conversation
	conversationID := uuid.New().String()

	// Create the conversation using domain logic
	conversation, err := domain.NewConversation(conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Save the conversation
	err = s.repo.Save(ctx, conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	return conversation, nil
}

// GetConversation retrieves a conversation by ID
func (s *ConversationServiceImpl) GetConversation(ctx context.Context, conversationID string) (*domain.Conversation, error) {
	if conversationID == "" {
		return nil, errors.New("conversation ID cannot be empty")
	}

	conversation, err := s.repo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conversation, nil
}

// GetUserConversations retrieves all conversations for a user
func (s *ConversationServiceImpl) GetUserConversations(ctx context.Context, userID string) ([]*domain.Conversation, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	conversations, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user conversations: %w", err)
	}

	return conversations, nil
}

// AddUserMessage adds a user message to a conversation
func (s *ConversationServiceImpl) AddUserMessage(ctx context.Context, conversationID, content string, metadata map[string]interface{}) error {
	if conversationID == "" {
		return errors.New("conversation ID cannot be empty")
	}

	// Get the conversation
	conversation, err := s.repo.GetByID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// Generate a unique message ID
	messageID := uuid.New().String()

	// Add the user message using domain logic
	err = conversation.AddUserMessage(messageID, content, metadata)
	if err != nil {
		return fmt.Errorf("failed to add user message: %w", err)
	}

	// Save the updated conversation
	err = s.repo.Save(ctx, conversation)
	if err != nil {
		return fmt.Errorf("failed to save conversation after adding message: %w", err)
	}

	return nil
}

// AddAssistantMessage adds an assistant message to a conversation
func (s *ConversationServiceImpl) AddAssistantMessage(ctx context.Context, conversationID, content string, metadata map[string]interface{}) error {
	if conversationID == "" {
		return errors.New("conversation ID cannot be empty")
	}

	// Get the conversation
	conversation, err := s.repo.GetByID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// Generate a unique message ID
	messageID := uuid.New().String()

	// Add the assistant message using domain logic
	err = conversation.AddAssistantMessage(messageID, content, metadata)
	if err != nil {
		return fmt.Errorf("failed to add assistant message: %w", err)
	}

	// Save the updated conversation
	err = s.repo.Save(ctx, conversation)
	if err != nil {
		return fmt.Errorf("failed to save conversation after adding message: %w", err)
	}

	return nil
}

// AnalyzeUserInput analyzes user input and returns conversation analysis
func (s *ConversationServiceImpl) AnalyzeUserInput(ctx context.Context, userInput string) (*ConversationAnalysis, error) {
	// TODO: Implement AI-powered conversation analysis
	// For now, return a basic analysis to satisfy the interface
	return &ConversationAnalysis{
		Intent:         "unknown",
		Category:       "general",
		Complexity:     "simple",
		RequiredAgents: []string{},
		Parameters:     make(map[string]interface{}),
		Confidence:     0.5,
	}, nil
}

// BuildResponse builds a response based on execution results
func (s *ConversationServiceImpl) BuildResponse(ctx context.Context, result *ExecutionResult) (*Response, error) {
	// TODO: Implement AI-powered response building
	// For now, return a basic response to satisfy the interface
	return &Response{
		Content:  "Task completed successfully",
		Type:     "text",
		Metadata: make(map[string]interface{}),
		Actions:  []ResponseAction{},
	}, nil
}
