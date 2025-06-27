package domain

import (
	"errors"
	"fmt"
	"time"
)

// ConversationStatus represents the current status of a conversation
type ConversationStatus string

const (
	ConversationStatusActive   ConversationStatus = "active"
	ConversationStatusPaused   ConversationStatus = "paused"
	ConversationStatusClosed   ConversationStatus = "closed"
	ConversationStatusArchived ConversationStatus = "archived"
)

// MessageRole represents the role of a message sender
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// ConversationMessage represents a single message in a conversation
type ConversationMessage struct {
	ID        string                 `json:"id"`
	Role      MessageRole            `json:"role"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Conversation represents a conversation between a user and the AI orchestrator
type Conversation struct {
	ID               string                 `json:"id"`
	UserID           string                 `json:"user_id"`
	Status           ConversationStatus     `json:"status"`
	Messages         []ConversationMessage  `json:"messages"`
	ExecutionPlanIDs []string               `json:"execution_plan_ids"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	LastActivityAt   time.Time              `json:"last_activity_at"`
	Context          map[string]interface{} `json:"context,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	Title            string                 `json:"title,omitempty"`
	Summary          string                 `json:"summary,omitempty"`
}

// ConversationValidationError represents validation errors for conversations
type ConversationValidationError struct {
	Field   string
	Message string
}

func (e ConversationValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewConversation creates a new conversation with validation
func NewConversation(id, userID string) (*Conversation, error) {
	now := time.Now().UTC()

	conversation := &Conversation{
		ID:               id,
		UserID:           userID,
		Status:           ConversationStatusActive,
		Messages:         make([]ConversationMessage, 0),
		ExecutionPlanIDs: make([]string, 0),
		CreatedAt:        now,
		UpdatedAt:        now,
		LastActivityAt:   now,
		Context:          make(map[string]interface{}),
		Tags:             make([]string, 0),
	}

	if err := conversation.Validate(); err != nil {
		return nil, err
	}

	return conversation, nil
}

// Validate validates the conversation
func (c *Conversation) Validate() error {
	if c.ID == "" {
		return ConversationValidationError{Field: "id", Message: "ID cannot be empty"}
	}

	if c.UserID == "" {
		return ConversationValidationError{Field: "user_id", Message: "user ID cannot be empty"}
	}

	// Validate status
	if !c.isValidStatus(c.Status) {
		return ConversationValidationError{Field: "status", Message: "invalid status"}
	}

	// Validate messages
	for i, message := range c.Messages {
		if err := c.validateMessage(message); err != nil {
			return fmt.Errorf("message %d validation failed: %w", i, err)
		}
	}

	return nil
}

// isValidStatus checks if the status is valid
func (c *Conversation) isValidStatus(status ConversationStatus) bool {
	validStatuses := []ConversationStatus{
		ConversationStatusActive,
		ConversationStatusPaused,
		ConversationStatusClosed,
		ConversationStatusArchived,
	}

	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// isValidRole checks if the message role is valid
func (c *Conversation) isValidRole(role MessageRole) bool {
	validRoles := []MessageRole{
		MessageRoleUser,
		MessageRoleAssistant,
		MessageRoleSystem,
	}

	for _, valid := range validRoles {
		if role == valid {
			return true
		}
	}
	return false
}

// validateMessage validates a conversation message
func (c *Conversation) validateMessage(message ConversationMessage) error {
	if message.ID == "" {
		return ConversationValidationError{Field: "message.id", Message: "message ID cannot be empty"}
	}

	if message.Content == "" {
		return ConversationValidationError{Field: "message.content", Message: "message content cannot be empty"}
	}

	if !c.isValidRole(message.Role) {
		return ConversationValidationError{Field: "message.role", Message: "invalid message role"}
	}

	if message.Timestamp.IsZero() {
		return ConversationValidationError{Field: "message.timestamp", Message: "message timestamp cannot be zero"}
	}

	return nil
}

// AddMessage adds a message to the conversation with validation
func (c *Conversation) AddMessage(message ConversationMessage) error {
	if err := c.validateMessage(message); err != nil {
		return err
	}

	// Check for duplicate message IDs
	for _, existingMessage := range c.Messages {
		if existingMessage.ID == message.ID {
			return ConversationValidationError{Field: "message.id", Message: "message ID must be unique"}
		}
	}

	c.Messages = append(c.Messages, message)
	c.UpdatedAt = time.Now().UTC()
	c.LastActivityAt = time.Now().UTC()

	return nil
}

// AddUserMessage adds a user message to the conversation
func (c *Conversation) AddUserMessage(messageID, content string, metadata map[string]interface{}) error {
	message := ConversationMessage{
		ID:        messageID,
		Role:      MessageRoleUser,
		Content:   content,
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
	}

	return c.AddMessage(message)
}

// AddAssistantMessage adds an assistant message to the conversation
func (c *Conversation) AddAssistantMessage(messageID, content string, metadata map[string]interface{}) error {
	message := ConversationMessage{
		ID:        messageID,
		Role:      MessageRoleAssistant,
		Content:   content,
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
	}

	return c.AddMessage(message)
}

// AddSystemMessage adds a system message to the conversation
func (c *Conversation) AddSystemMessage(messageID, content string, metadata map[string]interface{}) error {
	message := ConversationMessage{
		ID:        messageID,
		Role:      MessageRoleSystem,
		Content:   content,
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
	}

	return c.AddMessage(message)
}

// GetMessages returns all messages in chronological order
func (c *Conversation) GetMessages() []ConversationMessage {
	// Messages are already in chronological order as they're appended
	return c.Messages
}

// GetUserMessages returns only user messages
func (c *Conversation) GetUserMessages() []ConversationMessage {
	var userMessages []ConversationMessage
	for _, message := range c.Messages {
		if message.Role == MessageRoleUser {
			userMessages = append(userMessages, message)
		}
	}
	return userMessages
}

// GetAssistantMessages returns only assistant messages
func (c *Conversation) GetAssistantMessages() []ConversationMessage {
	var assistantMessages []ConversationMessage
	for _, message := range c.Messages {
		if message.Role == MessageRoleAssistant {
			assistantMessages = append(assistantMessages, message)
		}
	}
	return assistantMessages
}

// GetLastMessage returns the most recent message
func (c *Conversation) GetLastMessage() *ConversationMessage {
	if len(c.Messages) == 0 {
		return nil
	}
	return &c.Messages[len(c.Messages)-1]
}

// GetLastUserMessage returns the most recent user message
func (c *Conversation) GetLastUserMessage() *ConversationMessage {
	for i := len(c.Messages) - 1; i >= 0; i-- {
		if c.Messages[i].Role == MessageRoleUser {
			return &c.Messages[i]
		}
	}
	return nil
}

// AddExecutionPlan associates an execution plan with the conversation
func (c *Conversation) AddExecutionPlan(planID string) error {
	if planID == "" {
		return ConversationValidationError{Field: "plan_id", Message: "execution plan ID cannot be empty"}
	}

	// Check for duplicate plan IDs
	for _, existingPlanID := range c.ExecutionPlanIDs {
		if existingPlanID == planID {
			return ConversationValidationError{Field: "plan_id", Message: "execution plan ID already exists"}
		}
	}

	c.ExecutionPlanIDs = append(c.ExecutionPlanIDs, planID)
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// SetContext sets or updates context information
func (c *Conversation) SetContext(key string, value interface{}) error {
	if key == "" {
		return ConversationValidationError{Field: "context_key", Message: "context key cannot be empty"}
	}

	if c.Context == nil {
		c.Context = make(map[string]interface{})
	}

	c.Context[key] = value
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// GetContext retrieves context information
func (c *Conversation) GetContext(key string) (interface{}, bool) {
	if c.Context == nil {
		return nil, false
	}

	value, exists := c.Context[key]
	return value, exists
}

// AddTag adds a tag to the conversation
func (c *Conversation) AddTag(tag string) error {
	if tag == "" {
		return ConversationValidationError{Field: "tag", Message: "tag cannot be empty"}
	}

	// Check for duplicate tags
	for _, existingTag := range c.Tags {
		if existingTag == tag {
			return ConversationValidationError{Field: "tag", Message: "tag already exists"}
		}
	}

	c.Tags = append(c.Tags, tag)
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// RemoveTag removes a tag from the conversation
func (c *Conversation) RemoveTag(tag string) {
	for i, existingTag := range c.Tags {
		if existingTag == tag {
			c.Tags = append(c.Tags[:i], c.Tags[i+1:]...)
			c.UpdatedAt = time.Now().UTC()
			break
		}
	}
}

// HasTag checks if the conversation has a specific tag
func (c *Conversation) HasTag(tag string) bool {
	for _, existingTag := range c.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// SetTitle sets the conversation title
func (c *Conversation) SetTitle(title string) {
	c.Title = title
	c.UpdatedAt = time.Now().UTC()
}

// SetSummary sets the conversation summary
func (c *Conversation) SetSummary(summary string) {
	c.Summary = summary
	c.UpdatedAt = time.Now().UTC()
}

// Pause pauses the conversation
func (c *Conversation) Pause() error {
	if c.Status != ConversationStatusActive {
		return errors.New("can only pause active conversations")
	}

	c.Status = ConversationStatusPaused
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// Resume resumes the conversation
func (c *Conversation) Resume() error {
	if c.Status != ConversationStatusPaused {
		return errors.New("can only resume paused conversations")
	}

	c.Status = ConversationStatusActive
	c.UpdatedAt = time.Now().UTC()
	c.LastActivityAt = time.Now().UTC()

	return nil
}

// Close closes the conversation
func (c *Conversation) Close() error {
	if c.Status == ConversationStatusClosed || c.Status == ConversationStatusArchived {
		return errors.New("conversation is already closed or archived")
	}

	c.Status = ConversationStatusClosed
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// Archive archives the conversation
func (c *Conversation) Archive() error {
	if c.Status != ConversationStatusClosed {
		return errors.New("can only archive closed conversations")
	}

	c.Status = ConversationStatusArchived
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// IsActive returns true if the conversation is active
func (c *Conversation) IsActive() bool {
	return c.Status == ConversationStatusActive
}

// IsClosed returns true if the conversation is closed or archived
func (c *Conversation) IsClosed() bool {
	return c.Status == ConversationStatusClosed || c.Status == ConversationStatusArchived
}

// GetMessageCount returns the total number of messages
func (c *Conversation) GetMessageCount() int {
	return len(c.Messages)
}

// GetDuration returns the duration of the conversation
func (c *Conversation) GetDuration() time.Duration {
	return c.LastActivityAt.Sub(c.CreatedAt)
}
