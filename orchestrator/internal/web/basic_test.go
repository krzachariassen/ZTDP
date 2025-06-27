package web

import (
	"context"
	"testing"

	"github.com/ztdp/orchestrator/internal/ai"
	"github.com/ztdp/orchestrator/internal/logging"
)

// Simple test to validate basic functionality
func TestWebBFFBasic(t *testing.T) {
	mockOrchestrator := &MockAIOrchestrator{
		responses: make(map[string]*ai.ConversationalResponse),
	}
	logger := logging.NewNoOpLogger()
	bff := NewWebBFF(mockOrchestrator, logger)

	if bff == nil {
		t.Error("Expected non-nil WebBFF instance")
	}

	// Test basic message processing
	ctx := context.Background()
	response, err := bff.ProcessWebMessage(ctx, "test-session", "Hello")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response == nil {
		t.Error("Expected non-nil response")
	}

	if response.SessionID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", response.SessionID)
	}

	if response.Content == "" {
		t.Error("Expected non-empty content")
	}
}
