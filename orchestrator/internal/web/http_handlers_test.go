package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ztdp/orchestrator/internal/logging"
)

// TestWebBFFHTTPHandler_RED tests HTTP endpoint handling (RED phase)
func TestWebBFFHTTPHandler_RED(t *testing.T) {
	// Setup
	mockOrchestrator := &MockAIOrchestrator{}
	logger := logging.NewNoOpLogger()
	bff := NewWebBFF(mockOrchestrator, logger)

	// Test case: POST /api/chat should handle web messages
	t.Run("POST /api/chat handles chat messages", func(t *testing.T) {
		// This test should fail because we haven't implemented the HTTP handler yet
		handler := bff.ChatHandler() // This method doesn't exist yet

		requestBody := map[string]string{
			"session_id": "test-session-123",
			"message":    "Hello, what can you do?",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 200 OK with proper JSON response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response WebResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.SessionID != "test-session-123" {
			t.Errorf("Expected session_id 'test-session-123', got '%s'", response.SessionID)
		}

		if response.Content == "" {
			t.Error("Expected non-empty content in response")
		}
	})

	t.Run("POST /api/chat validates required fields", func(t *testing.T) {
		handler := bff.ChatHandler()

		// Test missing session_id
		requestBody := map[string]string{
			"message": "Hello",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 400 Bad Request
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("POST /api/chat handles invalid JSON", func(t *testing.T) {
		handler := bff.ChatHandler()

		req := httptest.NewRequest("POST", "/api/chat", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 400 Bad Request
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestWebBFFWebSocketHandler_RED tests WebSocket endpoint handling (RED phase)
func TestWebBFFWebSocketHandler_RED(t *testing.T) {
	// Setup
	mockOrchestrator := &MockAIOrchestrator{}
	logger := logging.NewNoOpLogger()
	bff := NewWebBFF(mockOrchestrator, logger)

	t.Run("WebSocket /ws/chat handles connections", func(t *testing.T) {
		// This test should fail because we haven't implemented the WebSocket handler yet
		handler := bff.WebSocketHandler() // This method doesn't exist yet

		// Create test server
		server := httptest.NewServer(handler)
		defer server.Close()

		// Convert http://127.0.0.1 to ws://127.0.0.1
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/chat"

		// Connect to WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		// Send a message
		message := map[string]string{
			"session_id": "ws-test-session",
			"message":    "Hello via WebSocket",
		}
		if err := conn.WriteJSON(message); err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Read response
		var response WebResponse
		if err := conn.ReadJSON(&response); err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if response.SessionID != "ws-test-session" {
			t.Errorf("Expected session_id 'ws-test-session', got '%s'", response.SessionID)
		}

		if response.Content == "" {
			t.Error("Expected non-empty content in response")
		}
	})

	t.Run("WebSocket handles invalid messages gracefully", func(t *testing.T) {
		handler := bff.WebSocketHandler()

		server := httptest.NewServer(handler)
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/chat"

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		// Send invalid message (missing required fields)
		message := map[string]string{
			"message": "Hello", // missing session_id
		}
		if err := conn.WriteJSON(message); err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Should receive error response
		var response WebResponse
		if err := conn.ReadJSON(&response); err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if response.Error == "" {
			t.Error("Expected error in response for invalid message")
		}
	})
}

// TestWebBFFServerIntegration_RED tests the complete server setup (RED phase)
func TestWebBFFServerIntegration_RED(t *testing.T) {
	// Setup
	mockOrchestrator := &MockAIOrchestrator{}
	logger := logging.NewNoOpLogger()
	bff := NewWebBFF(mockOrchestrator, logger)

	t.Run("CreateWebServer returns configured HTTP server", func(t *testing.T) {
		// This test should fail because CreateWebServer doesn't exist yet
		server := bff.CreateWebServer(":8080") // This method doesn't exist yet

		if server == nil {
			t.Error("Expected non-nil HTTP server")
		}

		// Server should be configured with proper routes
		// We'll test this by making requests to the server
	})

	t.Run("Web server serves static files", func(t *testing.T) {
		server := bff.CreateWebServer(":8080")

		// Start server in background
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				t.Logf("Server error: %v", err)
			}
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)
		defer server.Shutdown(context.Background())

		// Test that server serves static files (if any)
		resp, err := http.Get("http://localhost:8080/")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Should return 200 or 404 (depending on whether we have static files)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
		}
	})
}

// WebSocketMessage represents WebSocket message structure
type WebSocketMessage struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	Type      string `json:"type,omitempty"` // "chat", "ping", etc.
}
