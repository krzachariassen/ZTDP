package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ztdp/orchestrator/internal/ai"
	"github.com/ztdp/orchestrator/internal/logging"
)

// TestMockOrchestrator for focused testing
type TestMockOrchestrator struct{}

func (m *TestMockOrchestrator) ProcessRequest(ctx context.Context, userInput, userID string) (*ai.ConversationalResponse, error) {
	return &ai.ConversationalResponse{
		Message: "Test response to: " + userInput,
		Intent:  "test_intent",
	}, nil
}

// TestWebBFFGreenPhase tests that our GREEN implementation works
func TestWebBFFGreenPhase(t *testing.T) {
	mockOrchestrator := &TestMockOrchestrator{}
	logger := logging.NewNoOpLogger()
	bff := NewWebBFF(mockOrchestrator, logger)

	t.Run("ProcessWebMessage works", func(t *testing.T) {
		ctx := context.Background()
		response, err := bff.ProcessWebMessage(ctx, "test-session", "Hello")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if response == nil {
			t.Fatal("Expected non-nil response")
		}

		if response.SessionID != "test-session" {
			t.Errorf("Expected session ID 'test-session', got '%s'", response.SessionID)
		}

		if response.Content != "Test response to: Hello" {
			t.Errorf("Expected 'Test response to: Hello', got '%s'", response.Content)
		}
	})

	t.Run("ChatHandler exists and responds", func(t *testing.T) {
		handler := bff.ChatHandler()

		if handler == nil {
			t.Fatal("Expected non-nil chat handler")
		}

		// Test HTTP request
		requestBody := map[string]string{
			"session_id": "test-session-123",
			"message":    "Hello, test",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var response WebResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.SessionID != "test-session-123" {
			t.Errorf("Expected session_id 'test-session-123', got '%s'", response.SessionID)
		}
	})

	t.Run("WebSocketHandler exists", func(t *testing.T) {
		handler := bff.WebSocketHandler()

		if handler == nil {
			t.Error("Expected non-nil WebSocket handler")
		}
	})

	t.Run("CreateWebServer returns configured server", func(t *testing.T) {
		server := bff.CreateWebServer(":8080")

		if server == nil {
			t.Error("Expected non-nil HTTP server")
		}

		if server.Addr != ":8080" {
			t.Errorf("Expected server address ':8080', got '%s'", server.Addr)
		}
	})
}
