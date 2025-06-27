package application

import (
	"context"

	"github.com/ztdp/orchestrator/internal/conversation/domain"
)

// ConversationService defines the application service interface for conversation management
type ConversationService interface {
	// CreateConversation creates a new conversation with validation
	CreateConversation(ctx context.Context, userID string) (*domain.Conversation, error)

	// GetConversation retrieves a conversation by ID
	GetConversation(ctx context.Context, conversationID string) (*domain.Conversation, error)

	// GetUserConversations retrieves all conversations for a user
	GetUserConversations(ctx context.Context, userID string) ([]*domain.Conversation, error)

	// AddUserMessage adds a user message to a conversation
	AddUserMessage(ctx context.Context, conversationID, content string, metadata map[string]interface{}) error

	// AddAssistantMessage adds an assistant message to a conversation
	AddAssistantMessage(ctx context.Context, conversationID, content string, metadata map[string]interface{}) error

	// AnalyzeUserInput analyzes user input and returns conversation analysis
	AnalyzeUserInput(ctx context.Context, userInput string) (*ConversationAnalysis, error)

	// BuildResponse builds a response based on execution results
	BuildResponse(ctx context.Context, result *ExecutionResult) (*Response, error)
}

// ConversationRepository defines the repository interface for conversation persistence
type ConversationRepository interface {
	// Save stores or updates a conversation
	Save(ctx context.Context, conversation *domain.Conversation) error

	// GetByID retrieves a conversation by ID
	GetByID(ctx context.Context, conversationID string) (*domain.Conversation, error)

	// GetByUserID retrieves all conversations for a user
	GetByUserID(ctx context.Context, userID string) ([]*domain.Conversation, error)

	// Delete removes a conversation
	Delete(ctx context.Context, conversationID string) error
}

// ConversationAnalysis represents the result of analyzing user input
type ConversationAnalysis struct {
	Intent         string                 `json:"intent"`
	Category       string                 `json:"category"`
	Complexity     string                 `json:"complexity"`
	RequiredAgents []string               `json:"required_agents"`
	Parameters     map[string]interface{} `json:"parameters"`
	Confidence     float64                `json:"confidence"`
}

// ExecutionResult represents the result of executing a plan
type ExecutionResult struct {
	PlanID      string                 `json:"plan_id"`
	Status      string                 `json:"status"`
	Results     map[string]interface{} `json:"results"`
	Error       string                 `json:"error,omitempty"`
	CompletedAt string                 `json:"completed_at"`
}

// Response represents a response to be sent to the user
type Response struct {
	Content  string                 `json:"content"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Actions  []ResponseAction       `json:"actions,omitempty"`
}

// ResponseAction represents an action that can be taken by the user
type ResponseAction struct {
	Type       string                 `json:"type"`
	Label      string                 `json:"label"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}
