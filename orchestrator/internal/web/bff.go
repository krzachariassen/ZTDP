package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ztdp/orchestrator/internal/ai"
	"github.com/ztdp/orchestrator/internal/logging"
)

// ChatRequest represents a chat request from the web UI
type ChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// WebResponse represents a response from the WebBFF to the web client
type WebResponse struct {
	Content   string `json:"content"`
	SessionID string `json:"session_id"`
	Intent    string `json:"intent,omitempty"`
	Error     string `json:"error,omitempty"`
}

// AIOrchestrator defines the interface for AI orchestration operations
type AIOrchestrator interface {
	ProcessRequest(ctx context.Context, userInput, userID string) (*ai.ConversationalResponse, error)
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for now
		// In production, implement proper origin checking
		return true
	},
}

// WebBFF (Backend for Frontend) handles web session communication
// It provides a clean separation between web UI concerns and agent orchestration
type WebBFF struct {
	orchestrator AIOrchestrator
	logger       logging.Logger
	sessions     map[string]*WebSession
	sessionMutex sync.RWMutex
}

// WebSession represents a web user session
type WebSession struct {
	SessionID string
	UserID    string
	CreatedAt int64
}

// NewWebBFF creates a new WebBFF instance
func NewWebBFF(orchestrator AIOrchestrator, logger logging.Logger) *WebBFF {
	return &WebBFF{
		orchestrator: orchestrator,
		logger:       logger,
		sessions:     make(map[string]*WebSession),
		sessionMutex: sync.RWMutex{},
	}
}

// ProcessWebMessage processes a message from a web session
// This method handles web-specific concerns and delegates AI processing to the orchestrator
func (w *WebBFF) ProcessWebMessage(ctx context.Context, sessionID, message string) (*WebResponse, error) {
	// Validate input
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}

	// Check context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Get or create session
	session := w.getOrCreateSession(sessionID)

	w.logger.Debug("Processing web message", "sessionID", sessionID, "message", message)

	// Process request through AI orchestrator
	// Note: For web sessions, we use the sessionID as userID to maintain session isolation
	aiResponse, err := w.orchestrator.ProcessRequest(ctx, message, session.UserID)
	if err != nil {
		w.logger.Error("Failed to process AI request", err, "sessionID", sessionID)
		return &WebResponse{
			Content:   "I'm sorry, I encountered an error processing your request.",
			SessionID: sessionID,
			Error:     err.Error(),
		}, nil // Return nil error to indicate graceful error handling
	}

	// Convert AI response to web response
	webResponse := &WebResponse{
		Content:   aiResponse.Message,
		SessionID: sessionID,
		Intent:    aiResponse.Intent,
	}

	w.logger.Info("Web message processed successfully", "sessionID", sessionID)

	return webResponse, nil
}

// getOrCreateSession retrieves an existing session or creates a new one
func (w *WebBFF) getOrCreateSession(sessionID string) *WebSession {
	w.sessionMutex.RLock()
	session, exists := w.sessions[sessionID]
	w.sessionMutex.RUnlock()

	if exists {
		return session
	}

	// Create new session
	w.sessionMutex.Lock()
	defer w.sessionMutex.Unlock()

	// Double-check after acquiring write lock
	if session, exists := w.sessions[sessionID]; exists {
		return session
	}

	session = &WebSession{
		SessionID: sessionID,
		UserID:    sessionID, // Use sessionID as userID for web sessions
		CreatedAt: 0,         // Could add timestamp here if needed
	}
	w.sessions[sessionID] = session

	w.logger.Info("Created new web session", "sessionID", sessionID)

	return session
}

// ChatHandler returns an HTTP handler for chat API endpoints
func (w *WebBFF) ChatHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request
		var chatReq ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
			w.logger.Error("Failed to decode chat request", err)
			http.Error(rw, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate request
		if chatReq.SessionID == "" {
			http.Error(rw, "session_id is required", http.StatusBadRequest)
			return
		}
		if chatReq.Message == "" {
			http.Error(rw, "message is required", http.StatusBadRequest)
			return
		}

		// Process message
		response, err := w.ProcessWebMessage(r.Context(), chatReq.SessionID, chatReq.Message)
		if err != nil {
			w.logger.Error("Failed to process web message", err)
			http.Error(rw, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return response
		rw.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(rw).Encode(response); err != nil {
			w.logger.Error("Failed to encode response", err)
			http.Error(rw, "Internal server error", http.StatusInternalServerError)
			return
		}
	})
}

// WebSocketHandler returns a WebSocket handler for real-time chat
func (w *WebBFF) WebSocketHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Upgrade connection to WebSocket
		conn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			w.logger.Error("Failed to upgrade to WebSocket", err)
			return
		}
		defer conn.Close()

		// Get session ID from query params (optional, can be overridden by message)
		defaultSessionID := r.URL.Query().Get("session_id")

		w.logger.Info("WebSocket connection established", "default_session_id", defaultSessionID)

		// Handle WebSocket messages
		for {
			var message ChatRequest
			if err := conn.ReadJSON(&message); err != nil {
				w.logger.Error("Failed to read WebSocket message", err)
				break
			}

			// Use session ID from message, or fall back to default from URL
			sessionID := message.SessionID
			if sessionID == "" {
				sessionID = defaultSessionID
			}

			// Validate session ID
			if sessionID == "" {
				w.logger.Error("Missing session_id in WebSocket message", nil)
				conn.WriteJSON(map[string]string{"error": "session_id is required"})
				continue
			}

			// Process message
			response, err := w.ProcessWebMessage(r.Context(), sessionID, message.Message)
			if err != nil {
				w.logger.Error("Failed to process WebSocket message", err)
				conn.WriteJSON(map[string]string{"error": "Failed to process message"})
				continue
			}

			// Send response
			if err := conn.WriteJSON(response); err != nil {
				w.logger.Error("Failed to send WebSocket response", err)
				break
			}
		}

		w.logger.Info("WebSocket connection closed")
	})
}

// CreateWebServer creates and configures an HTTP server with WebBFF routes
func (w *WebBFF) CreateWebServer(addr string) *http.Server {
	mux := http.NewServeMux()

	// Add routes
	mux.Handle("/api/chat", w.ChatHandler())
	mux.Handle("/ws", w.WebSocketHandler())

	// Add health check
	mux.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, `{"status":"ok","service":"web-bff"}`)
	})

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
