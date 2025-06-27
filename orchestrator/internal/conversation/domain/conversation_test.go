package domain

import (
	"testing"
	"time"
)

func TestNewConversation(t *testing.T) {
	t.Run("should create valid conversation", func(t *testing.T) {
		// Given
		id := "conv-123"
		userID := "user-456"

		// When
		conversation, err := NewConversation(id, userID)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if conversation.ID != id {
			t.Errorf("Expected ID %s, got %s", id, conversation.ID)
		}

		if conversation.UserID != userID {
			t.Errorf("Expected UserID %s, got %s", userID, conversation.UserID)
		}

		if conversation.Status != ConversationStatusActive {
			t.Errorf("Expected Status %s, got %s", ConversationStatusActive, conversation.Status)
		}

		if len(conversation.Messages) != 0 {
			t.Errorf("Expected empty messages, got %d messages", len(conversation.Messages))
		}

		if len(conversation.ExecutionPlanIDs) != 0 {
			t.Errorf("Expected empty execution plan IDs, got %d IDs", len(conversation.ExecutionPlanIDs))
		}

		if conversation.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}

		if conversation.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}

		if conversation.LastActivityAt.IsZero() {
			t.Error("Expected LastActivityAt to be set")
		}
	})

	t.Run("should fail with empty ID", func(t *testing.T) {
		// When
		_, err := NewConversation("", "user-456")

		// Then
		if err == nil {
			t.Fatal("Expected validation error for empty ID")
		}

		validationErr, ok := err.(ConversationValidationError)
		if !ok {
			t.Errorf("Expected ConversationValidationError, got %T", err)
		}

		if validationErr.Field != "id" {
			t.Errorf("Expected field 'id', got '%s'", validationErr.Field)
		}
	})

	t.Run("should fail with empty user ID", func(t *testing.T) {
		// When
		_, err := NewConversation("conv-123", "")

		// Then
		if err == nil {
			t.Fatal("Expected validation error for empty user ID")
		}

		validationErr, ok := err.(ConversationValidationError)
		if !ok {
			t.Errorf("Expected ConversationValidationError, got %T", err)
		}

		if validationErr.Field != "user_id" {
			t.Errorf("Expected field 'user_id', got '%s'", validationErr.Field)
		}
	})
}

func TestConversation_AddUserMessage(t *testing.T) {
	t.Run("should add valid user message", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		messageID := "msg-1"
		content := "Hello, I need help with deployment"
		metadata := map[string]interface{}{"client": "web"}

		// When
		err := conversation.AddUserMessage(messageID, content, metadata)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(conversation.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(conversation.Messages))
		}

		message := conversation.Messages[0]
		if message.ID != messageID {
			t.Errorf("Expected message ID %s, got %s", messageID, message.ID)
		}

		if message.Role != MessageRoleUser {
			t.Errorf("Expected role %s, got %s", MessageRoleUser, message.Role)
		}

		if message.Content != content {
			t.Errorf("Expected content %s, got %s", content, message.Content)
		}

		if message.Timestamp.IsZero() {
			t.Error("Expected message timestamp to be set")
		}
	})

	t.Run("should fail with duplicate message ID", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		messageID := "msg-1"
		
		conversation.AddUserMessage(messageID, "First message", nil)

		// When
		err := conversation.AddUserMessage(messageID, "Second message", nil)

		// Then
		if err == nil {
			t.Fatal("Expected error for duplicate message ID")
		}

		validationErr, ok := err.(ConversationValidationError)
		if !ok {
			t.Errorf("Expected ConversationValidationError, got %T", err)
		}

		if validationErr.Field != "message.id" {
			t.Errorf("Expected field 'message.id', got '%s'", validationErr.Field)
		}
	})
}

func TestConversation_AddAssistantMessage(t *testing.T) {
	t.Run("should add valid assistant message", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		messageID := "msg-1"
		content := "I can help you with deployment. Let me analyze your request."

		// When
		err := conversation.AddAssistantMessage(messageID, content, nil)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(conversation.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(conversation.Messages))
		}

		message := conversation.Messages[0]
		if message.Role != MessageRoleAssistant {
			t.Errorf("Expected role %s, got %s", MessageRoleAssistant, message.Role)
		}
	})
}

func TestConversation_GetMessages(t *testing.T) {
	t.Run("should return messages in chronological order", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		
		// Add messages with slight delays to ensure different timestamps
		conversation.AddUserMessage("msg-1", "First message", nil)
		time.Sleep(1 * time.Millisecond)
		conversation.AddAssistantMessage("msg-2", "Second message", nil)
		time.Sleep(1 * time.Millisecond)
		conversation.AddUserMessage("msg-3", "Third message", nil)

		// When
		messages := conversation.GetMessages()

		// Then
		if len(messages) != 3 {
			t.Errorf("Expected 3 messages, got %d", len(messages))
		}

		if messages[0].ID != "msg-1" {
			t.Errorf("Expected first message ID 'msg-1', got %s", messages[0].ID)
		}

		if messages[1].ID != "msg-2" {
			t.Errorf("Expected second message ID 'msg-2', got %s", messages[1].ID)
		}

		if messages[2].ID != "msg-3" {
			t.Errorf("Expected third message ID 'msg-3', got %s", messages[2].ID)
		}
	})
}

func TestConversation_GetUserMessages(t *testing.T) {
	t.Run("should return only user messages", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		conversation.AddUserMessage("msg-1", "User message 1", nil)
		conversation.AddAssistantMessage("msg-2", "Assistant message", nil)
		conversation.AddUserMessage("msg-3", "User message 2", nil)

		// When
		userMessages := conversation.GetUserMessages()

		// Then
		if len(userMessages) != 2 {
			t.Errorf("Expected 2 user messages, got %d", len(userMessages))
		}

		for _, message := range userMessages {
			if message.Role != MessageRoleUser {
				t.Errorf("Expected user role, got %s", message.Role)
			}
		}
	})
}

func TestConversation_GetLastMessage(t *testing.T) {
	t.Run("should return nil for empty conversation", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")

		// When
		lastMessage := conversation.GetLastMessage()

		// Then
		if lastMessage != nil {
			t.Error("Expected nil for empty conversation")
		}
	})

	t.Run("should return last message", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		conversation.AddUserMessage("msg-1", "First message", nil)
		conversation.AddAssistantMessage("msg-2", "Last message", nil)

		// When
		lastMessage := conversation.GetLastMessage()

		// Then
		if lastMessage == nil {
			t.Fatal("Expected last message, got nil")
		}

		if lastMessage.ID != "msg-2" {
			t.Errorf("Expected last message ID 'msg-2', got %s", lastMessage.ID)
		}
	})
}

func TestConversation_AddExecutionPlan(t *testing.T) {
	t.Run("should add execution plan ID", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		planID := "plan-789"

		// When
		err := conversation.AddExecutionPlan(planID)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(conversation.ExecutionPlanIDs) != 1 {
			t.Errorf("Expected 1 execution plan ID, got %d", len(conversation.ExecutionPlanIDs))
		}

		if conversation.ExecutionPlanIDs[0] != planID {
			t.Errorf("Expected plan ID %s, got %s", planID, conversation.ExecutionPlanIDs[0])
		}
	})

	t.Run("should fail with duplicate plan ID", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		planID := "plan-789"
		
		conversation.AddExecutionPlan(planID)

		// When
		err := conversation.AddExecutionPlan(planID)

		// Then
		if err == nil {
			t.Fatal("Expected error for duplicate plan ID")
		}

		validationErr, ok := err.(ConversationValidationError)
		if !ok {
			t.Errorf("Expected ConversationValidationError, got %T", err)
		}

		if validationErr.Field != "plan_id" {
			t.Errorf("Expected field 'plan_id', got '%s'", validationErr.Field)
		}
	})
}

func TestConversation_SetContext(t *testing.T) {
	t.Run("should set context value", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		key := "deployment_env"
		value := "production"

		// When
		err := conversation.SetContext(key, value)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		retrievedValue, exists := conversation.GetContext(key)
		if !exists {
			t.Error("Expected context value to exist")
		}

		if retrievedValue != value {
			t.Errorf("Expected context value %s, got %v", value, retrievedValue)
		}
	})

	t.Run("should fail with empty key", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")

		// When
		err := conversation.SetContext("", "value")

		// Then
		if err == nil {
			t.Fatal("Expected error for empty context key")
		}

		validationErr, ok := err.(ConversationValidationError)
		if !ok {
			t.Errorf("Expected ConversationValidationError, got %T", err)
		}

		if validationErr.Field != "context_key" {
			t.Errorf("Expected field 'context_key', got '%s'", validationErr.Field)
		}
	})
}

func TestConversation_AddTag(t *testing.T) {
	t.Run("should add tag", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		tag := "deployment"

		// When
		err := conversation.AddTag(tag)

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if !conversation.HasTag(tag) {
			t.Error("Expected conversation to have tag")
		}
	})

	t.Run("should fail with duplicate tag", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		tag := "deployment"
		
		conversation.AddTag(tag)

		// When
		err := conversation.AddTag(tag)

		// Then
		if err == nil {
			t.Fatal("Expected error for duplicate tag")
		}

		validationErr, ok := err.(ConversationValidationError)
		if !ok {
			t.Errorf("Expected ConversationValidationError, got %T", err)
		}

		if validationErr.Field != "tag" {
			t.Errorf("Expected field 'tag', got '%s'", validationErr.Field)
		}
	})
}

func TestConversation_RemoveTag(t *testing.T) {
	t.Run("should remove existing tag", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		tag := "deployment"
		
		conversation.AddTag(tag)
		conversation.AddTag("monitoring")

		// When
		conversation.RemoveTag(tag)

		// Then
		if conversation.HasTag(tag) {
			t.Error("Expected tag to be removed")
		}

		if !conversation.HasTag("monitoring") {
			t.Error("Expected other tags to remain")
		}
	})
}

func TestConversation_StatusTransitions(t *testing.T) {
	t.Run("should pause active conversation", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")

		// When
		err := conversation.Pause()

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if conversation.Status != ConversationStatusPaused {
			t.Errorf("Expected status %s, got %s", ConversationStatusPaused, conversation.Status)
		}
	})

	t.Run("should resume paused conversation", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		conversation.Pause()

		// When
		err := conversation.Resume()

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if conversation.Status != ConversationStatusActive {
			t.Errorf("Expected status %s, got %s", ConversationStatusActive, conversation.Status)
		}
	})

	t.Run("should close conversation", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")

		// When
		err := conversation.Close()

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if conversation.Status != ConversationStatusClosed {
			t.Errorf("Expected status %s, got %s", ConversationStatusClosed, conversation.Status)
		}

		if !conversation.IsClosed() {
			t.Error("Expected conversation to be closed")
		}
	})

	t.Run("should archive closed conversation", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		conversation.Close()

		// When
		err := conversation.Archive()

		// Then
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if conversation.Status != ConversationStatusArchived {
			t.Errorf("Expected status %s, got %s", ConversationStatusArchived, conversation.Status)
		}

		if !conversation.IsClosed() {
			t.Error("Expected archived conversation to be considered closed")
		}
	})

	t.Run("should fail to archive non-closed conversation", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")

		// When
		err := conversation.Archive()

		// Then
		if err == nil {
			t.Fatal("Expected error when archiving non-closed conversation")
		}
	})
}

func TestConversation_IsActive(t *testing.T) {
	t.Run("should return true for active conversation", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")

		// When & Then
		if !conversation.IsActive() {
			t.Error("Expected new conversation to be active")
		}
	})

	t.Run("should return false for paused conversation", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		conversation.Pause()

		// When & Then
		if conversation.IsActive() {
			t.Error("Expected paused conversation to not be active")
		}
	})
}

func TestConversation_GetMessageCount(t *testing.T) {
	t.Run("should return correct message count", func(t *testing.T) {
		// Given
		conversation, _ := NewConversation("conv-123", "user-456")
		conversation.AddUserMessage("msg-1", "Message 1", nil)
		conversation.AddAssistantMessage("msg-2", "Message 2", nil)

		// When
		count := conversation.GetMessageCount()

		// Then
		if count != 2 {
			t.Errorf("Expected message count 2, got %d", count)
		}
	})
}
