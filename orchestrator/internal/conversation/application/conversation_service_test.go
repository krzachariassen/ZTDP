package application

import (
	"context"
	"errors"
	"testing"

	"github.com/ztdp/orchestrator/internal/conversation/domain"
)

// MockConversationRepository is a mock implementation of ConversationRepository
type MockConversationRepository struct {
	conversations map[string]*domain.Conversation
	saveErr       error
	getByIDErr    error
	getByUserErr  error
	deleteErr     error
}

func NewMockConversationRepository() *MockConversationRepository {
	return &MockConversationRepository{
		conversations: make(map[string]*domain.Conversation),
	}
}

func (m *MockConversationRepository) Save(ctx context.Context, conversation *domain.Conversation) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.conversations[conversation.ID] = conversation
	return nil
}

func (m *MockConversationRepository) GetByID(ctx context.Context, conversationID string) (*domain.Conversation, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if conversation, exists := m.conversations[conversationID]; exists {
		return conversation, nil
	}
	return nil, errors.New("conversation not found")
}

func (m *MockConversationRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Conversation, error) {
	if m.getByUserErr != nil {
		return nil, m.getByUserErr
	}

	var userConversations []*domain.Conversation
	for _, conversation := range m.conversations {
		if conversation.UserID == userID {
			userConversations = append(userConversations, conversation)
		}
	}
	return userConversations, nil
}

func (m *MockConversationRepository) Delete(ctx context.Context, conversationID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.conversations, conversationID)
	return nil
}

// Helper methods for testing
func (m *MockConversationRepository) SetSaveError(err error) {
	m.saveErr = err
}

func (m *MockConversationRepository) SetGetByIDError(err error) {
	m.getByIDErr = err
}

func (m *MockConversationRepository) SetGetByUserError(err error) {
	m.getByUserErr = err
}

func (m *MockConversationRepository) SetDeleteError(err error) {
	m.deleteErr = err
}

// TestConversationServiceImpl_CreateConversation tests the CreateConversation method
func TestConversationServiceImpl_CreateConversation(t *testing.T) {
	ctx := context.Background()
	repo := NewMockConversationRepository()
	service := NewConversationServiceImpl(repo)

	t.Run("should create conversation successfully", func(t *testing.T) {
		userID := "user123"

		conv, err := service.CreateConversation(ctx, userID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if conv == nil {
			t.Fatal("expected conversation, got nil")
		}

		if conv.UserID != userID {
			t.Errorf("expected user ID %s, got %s", userID, conv.UserID)
		}

		if conv.Status != domain.ConversationStatusActive {
			t.Errorf("expected status %s, got %s", domain.ConversationStatusActive, conv.Status)
		}
	})

	t.Run("should fail when repository save fails", func(t *testing.T) {
		repo.SetSaveError(errors.New("save failed"))
		userID := "user123"

		conv, err := service.CreateConversation(ctx, userID)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if conv != nil {
			t.Fatal("expected nil conversation on error")
		}
	})

	t.Run("should fail with empty user ID", func(t *testing.T) {
		repo.SetSaveError(nil) // Reset error

		conv, err := service.CreateConversation(ctx, "")

		if err == nil {
			t.Fatal("expected error for empty user ID")
		}

		if conv != nil {
			t.Fatal("expected nil conversation on validation error")
		}
	})
}

// TestConversationServiceImpl_GetConversation tests the GetConversation method
func TestConversationServiceImpl_GetConversation(t *testing.T) {
	ctx := context.Background()
	repo := NewMockConversationRepository()
	service := NewConversationServiceImpl(repo)

	// Setup test data
	testConv, _ := domain.NewConversation("conv123", "user123")
	repo.Save(ctx, testConv)

	t.Run("should retrieve conversation successfully", func(t *testing.T) {
		conv, err := service.GetConversation(ctx, "conv123")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if conv == nil {
			t.Fatal("expected conversation, got nil")
		}

		if conv.ID != "conv123" {
			t.Errorf("expected ID conv123, got %s", conv.ID)
		}
	})

	t.Run("should fail when conversation not found", func(t *testing.T) {
		conv, err := service.GetConversation(ctx, "nonexistent")

		if err == nil {
			t.Fatal("expected error for nonexistent conversation")
		}

		if conv != nil {
			t.Fatal("expected nil conversation when not found")
		}
	})

	t.Run("should fail with empty conversation ID", func(t *testing.T) {
		conv, err := service.GetConversation(ctx, "")

		if err == nil {
			t.Fatal("expected error for empty conversation ID")
		}

		if conv != nil {
			t.Fatal("expected nil conversation on validation error")
		}
	})
}

// TestConversationServiceImpl_AddUserMessage tests the AddUserMessage method
func TestConversationServiceImpl_AddUserMessage(t *testing.T) {
	ctx := context.Background()
	repo := NewMockConversationRepository()
	service := NewConversationServiceImpl(repo)

	// Setup test data
	testConv, _ := domain.NewConversation("conv123", "user123")
	repo.Save(ctx, testConv)

	t.Run("should add user message successfully", func(t *testing.T) {
		content := "Hello, world!"
		metadata := map[string]interface{}{"source": "test"}

		err := service.AddUserMessage(ctx, "conv123", content, metadata)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify message was added
		conv, _ := repo.GetByID(ctx, "conv123")
		messages := conv.GetUserMessages()

		if len(messages) != 1 {
			t.Fatalf("expected 1 user message, got %d", len(messages))
		}

		if messages[0].Content != content {
			t.Errorf("expected content %s, got %s", content, messages[0].Content)
		}
	})

	t.Run("should fail with empty conversation ID", func(t *testing.T) {
		err := service.AddUserMessage(ctx, "", "content", nil)

		if err == nil {
			t.Fatal("expected error for empty conversation ID")
		}
	})

	t.Run("should fail when conversation not found", func(t *testing.T) {
		err := service.AddUserMessage(ctx, "nonexistent", "content", nil)

		if err == nil {
			t.Fatal("expected error for nonexistent conversation")
		}
	})
}

// This will fail until we implement ConversationServiceImpl
func TestConversationServiceExists(t *testing.T) {
	// This test will fail because we haven't implemented ConversationServiceImpl yet
	// This is our RED phase - we write the test first, then implement the functionality
	t.Run("ConversationServiceImpl should exist", func(t *testing.T) {
		repo := NewMockConversationRepository()
		service := NewConversationServiceImpl(repo)

		if service == nil {
			t.Fatal("NewConversationServiceImpl should return a service instance")
		}
	})
}
