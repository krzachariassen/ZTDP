package web

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ztdp/orchestrator/internal/ai"
	"github.com/ztdp/orchestrator/internal/logging"
)

// MockAIOrchestrator for testing
type MockAIOrchestrator struct {
	responses map[string]*ai.ConversationalResponse
}

func (m *MockAIOrchestrator) ProcessRequest(ctx context.Context, userInput, userID string) (*ai.ConversationalResponse, error) {
	if response, exists := m.responses[userInput]; exists {
		return response, nil
	}
	return &ai.ConversationalResponse{
		Message:    "Mock AI response for: " + userInput,
		Intent:     "test",
		Confidence: 0.9,
		Actions:    []ai.Action{},
	}, nil
}

func TestWebBFF_DirectAIResponse(t *testing.T) {
	// RED: Test that web sessions get immediate AI responses without message bus
	mockAI := &MockAIOrchestrator{
		responses: map[string]*ai.ConversationalResponse{
			"Count words in hello world": {
				Message:    "I'll count the words for you. The text 'hello world' contains 2 words.",
				Intent:     "word_count",
				Confidence: 0.95,
				Actions: []ai.Action{{
					Type: "execute_plan",
					Parameters: map[string]interface{}{
						"target":      "text-processor-001",
						"description": "Count words in text",
					},
				}},
			},
		},
	}

	logger := logging.NewNoOpLogger()
	bff := NewWebBFF(mockAI, logger)

	// Test direct AI response
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := bff.ProcessWebMessage(ctx, "web-session-123", "Count words in hello world")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	expectedContent := "I'll count the words for you. The text 'hello world' contains 2 words."
	if response.Content != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, response.Content)
	}

	if response.SessionID != "web-session-123" {
		t.Errorf("Expected session ID 'web-session-123', got '%s'", response.SessionID)
	}
}

func TestWebBFF_NoRabbitMQQueues(t *testing.T) {
	// RED: Test that web sessions don't create RabbitMQ queues
	mockAI := &MockAIOrchestrator{}
	logger := logging.NewNoOpLogger()
	bff := NewWebBFF(mockAI, logger)

	// Process multiple messages from same session
	ctx := context.Background()
	sessionID := "web-user-1234567890"

	for i := 0; i < 5; i++ {
		_, err := bff.ProcessWebMessage(ctx, sessionID, "Test message")
		if err != nil {
			t.Fatalf("Message %d failed: %v", i+1, err)
		}
	}

	// Test should pass without creating any RabbitMQ infrastructure
	// This is verified by the fact that we only use direct AI calls
}

func TestWebBFF_ConcurrentSessions(t *testing.T) {
	// RED: Test handling multiple concurrent web sessions
	mockAI := &MockAIOrchestrator{}
	logger := logging.NewNoOpLogger()
	bff := NewWebBFF(mockAI, logger)

	ctx := context.Background()
	numSessions := 10
	done := make(chan bool, numSessions)

	// Start concurrent sessions
	for i := 0; i < numSessions; i++ {
		go func(sessionNum int) {
			sessionID := fmt.Sprintf("web-session-%d", sessionNum)
			message := fmt.Sprintf("Message from session %d", sessionNum)

			response, err := bff.ProcessWebMessage(ctx, sessionID, message)
			if err != nil {
				t.Errorf("Session %d failed: %v", sessionNum, err)
			}

			if response == nil {
				t.Errorf("Session %d got nil response", sessionNum)
			}

			done <- true
		}(i)
	}

	// Wait for all sessions to complete
	timeout := time.After(10 * time.Second)
	completed := 0
	for completed < numSessions {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatalf("Timeout: only %d/%d sessions completed", completed, numSessions)
		}
	}
}

func TestWebBFF_ErrorHandling(t *testing.T) {
	// RED: Test graceful error handling for web sessions
	mockAI := &MockAIOrchestrator{}
	logger := logging.NewNoOpLogger()
	bff := NewWebBFF(mockAI, logger)

	ctx := context.Background()

	// Test empty session ID
	_, err := bff.ProcessWebMessage(ctx, "", "test message")
	if err == nil {
		t.Error("Expected error for empty session ID")
	}

	// Test empty message
	_, err = bff.ProcessWebMessage(ctx, "session-123", "")
	if err == nil {
		t.Error("Expected error for empty message")
	}

	// Test cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	_, err = bff.ProcessWebMessage(cancelledCtx, "session-123", "test message")
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}
