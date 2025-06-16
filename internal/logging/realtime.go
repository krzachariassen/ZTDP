package logging

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// RealtimeLogSink is a sink that broadcasts log entries to WebSocket clients
type RealtimeLogSink struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

// NewRealtimeLogSink creates a new real-time log sink
func NewRealtimeLogSink() *RealtimeLogSink {
	sink := &RealtimeLogSink{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn, 10),
		unregister: make(chan *websocket.Conn, 10),
	}

	// Start the client manager goroutine
	go sink.run()

	return sink
}

// run manages WebSocket client connections
func (r *RealtimeLogSink) run() {
	for {
		select {
		case conn := <-r.register:
			r.mu.Lock()
			r.clients[conn] = true
			r.mu.Unlock()
			GetLogger().ForComponent("realtime-log").Debug("WebSocket client connected, total: %d", len(r.clients))

		case conn := <-r.unregister:
			r.mu.Lock()
			if _, ok := r.clients[conn]; ok {
				delete(r.clients, conn)
				conn.Close()
			}
			r.mu.Unlock()
			GetLogger().ForComponent("realtime-log").Debug("WebSocket client disconnected, total: %d", len(r.clients))
		}
	}
}

// RegisterClient registers a new WebSocket client for real-time logs
func (r *RealtimeLogSink) RegisterClient(conn *websocket.Conn) {
	r.register <- conn
}

// UnregisterClient unregisters a WebSocket client
func (r *RealtimeLogSink) UnregisterClient(conn *websocket.Conn) {
	r.unregister <- conn
}

// Write broadcasts a log entry to all connected WebSocket clients
func (r *RealtimeLogSink) Write(entry LogEntry) error {
	if len(r.clients) == 0 {
		return nil // No clients connected
	}

	// Transform the log entry for frontend consumption
	frontendEntry := map[string]interface{}{
		"timestamp": entry.Timestamp.Format(time.RFC3339),
		"level":     entry.Level,
		"message":   entry.Message,
		"component": entry.Component,
		"source":    entry.Source,
		"type":      "log.entry",
	}

	// Add operation if present
	if entry.Operation != "" {
		frontendEntry["operation"] = entry.Operation
	}

	// Add request ID if present
	if entry.RequestID != "" {
		frontendEntry["request_id"] = entry.RequestID
	}

	// Add duration if present
	if entry.Duration != nil {
		frontendEntry["duration_ms"] = entry.Duration.Milliseconds()
	}

	// Add error if present
	if entry.Error != "" {
		frontendEntry["error"] = entry.Error
	}

	// Add selected properties
	if len(entry.Properties) > 0 {
		details := make(map[string]interface{})
		for k, v := range entry.Properties {
			// Skip sensitive or redundant information
			if k == "request_id" || k == "operation" {
				continue
			}
			details[k] = v
		}
		if len(details) > 0 {
			frontendEntry["details"] = details
		}
	}

	// Broadcast to all clients
	r.mu.Lock() // Use write lock to prevent concurrent websocket writes
	var failedClients []*websocket.Conn
	for conn := range r.clients {
		if err := conn.WriteJSON(frontendEntry); err != nil {
			failedClients = append(failedClients, conn)
		}
	}
	r.mu.Unlock()

	// Remove failed clients (need to re-acquire lock)
	if len(failedClients) > 0 {
		r.mu.Lock()
		for _, conn := range failedClients {
			delete(r.clients, conn)
			conn.Close()
		}
		r.mu.Unlock()
	}

	return nil
}

// BroadcastEvent broadcasts a structured event directly to all connected WebSocket clients
func (r *RealtimeLogSink) BroadcastEvent(event map[string]interface{}) error {
	if len(r.clients) == 0 {
		return nil // No clients connected
	}

	// Add type indicator for frontend filtering
	event["type"] = "event.structured"

	// Broadcast to all clients
	r.mu.Lock() // Use write lock to prevent concurrent websocket writes
	var failedClients []*websocket.Conn
	for conn := range r.clients {
		if err := conn.WriteJSON(event); err != nil {
			failedClients = append(failedClients, conn)
		}
	}
	r.mu.Unlock()

	// Remove failed clients (need to re-acquire lock)
	if len(failedClients) > 0 {
		r.mu.Lock()
		for _, conn := range failedClients {
			delete(r.clients, conn)
			conn.Close()
		}
		r.mu.Unlock()
	}

	return nil
}

// Close closes the real-time log sink
func (r *RealtimeLogSink) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Close all client connections
	for conn := range r.clients {
		conn.Close()
	}
	r.clients = make(map[*websocket.Conn]bool)

	return nil
}

// GetClientCount returns the number of connected clients
func (r *RealtimeLogSink) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}
