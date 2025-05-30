package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now (in production, this should be more restrictive)
		return true
	},
}

// Global real-time log sink for WebSocket broadcasting
var realtimeLogSink *logging.RealtimeLogSink

// InitLogManager initializes the log manager and sets up event subscriptions
func InitLogManager() {
	// Get the real-time sink from the global logger
	logger := logging.GetLogger()

	// Create a new real-time sink if one doesn't exist
	if realtimeLogSink == nil {
		realtimeLogSink = logging.NewRealtimeLogSink()
		logger.AddSink(realtimeLogSink)
	}

	// Subscribe to all event types for logging
	subscribeToEvents()

	logger.Info("📊 Real-time log manager initialized")
}

// subscribeToEvents sets up event subscriptions for real-time logging using the simplified event system
func subscribeToEvents() {
	logger := logging.GetLogger().ForComponent("event-subscriber")

	// Subscribe to the three main application event types we have
	events.GlobalEventBus.Subscribe(events.EventTypeApplicationCreated, createEventHandler(logger, "application", "Application Created"))
	events.GlobalEventBus.Subscribe(events.EventTypeApplicationUpdated, createEventHandler(logger, "application", "Application Updated"))
	events.GlobalEventBus.Subscribe(events.EventTypeApplicationDeleted, createEventHandler(logger, "application", "Application Deleted"))
}

// createEventHandler creates a generic event handler for the simplified event system
func createEventHandler(logger *logging.Logger, component, message string) events.EventHandler {
	return func(event events.Event) error {
		// Use the centralized logger with appropriate context
		eventLogger := logger.
			ForComponent(component).
			WithContext("event_type", string(event.Type)).
			WithContext("event_source", event.Source).
			WithContext("event_subject", event.Subject)

		// Add payload as properties
		for k, v := range event.Payload {
			eventLogger = eventLogger.WithContext(k, v)
		}

		// Determine the appropriate log level and message based on event type and payload
		eventType := string(event.Type)
		if eventType == "application.created" {
			eventLogger.Info("✅ %s: %s", message, event.Subject)
		} else if eventType == "application.updated" {
			// Check if this is actually a deployment or policy event from the payload
			if eventPayloadType, ok := event.Payload["type"].(string); ok {
				switch eventPayloadType {
				case "deployment_requested":
					eventLogger.Info("🚀 Deployment Requested: %s", event.Subject)
				case "deployment_started":
					eventLogger.Info("📦 Deployment Started: %s", event.Subject)
				case "deployment_completed":
					eventLogger.Info("✅ Deployment Completed: %s", event.Subject)
				case "deployment_failed":
					eventLogger.Warn("❌ Deployment Failed: %s", event.Subject)
				case "transition_attempt":
					eventLogger.Info("🔒 Policy Transition Attempt: %s", event.Subject)
				case "transition_success":
					eventLogger.Info("✅ Policy Transition Approved: %s", event.Subject)
				case "transition_failure":
					eventLogger.Warn("❌ Policy Transition Rejected: %s", event.Subject)
				case "policy_check":
					eventLogger.Info("🔍 Policy Check: %s", event.Subject)
				case "policy_check_result":
					if success, ok := event.Payload["success"].(bool); ok && success {
						eventLogger.Info("✅ Policy Check Passed: %s", event.Subject)
					} else {
						eventLogger.Warn("❌ Policy Check Failed: %s", event.Subject)
					}
				case "resource_provision_completed":
					eventLogger.Info("🔧 Resource Provisioned: %s", event.Subject)
				default:
					eventLogger.Info("📝 %s: %s", message, event.Subject)
				}
			} else {
				eventLogger.Info("📝 %s: %s", message, event.Subject)
			}
		} else if eventType == "application.deleted" {
			eventLogger.Info("🗑️  %s: %s", message, event.Subject)
		} else {
			eventLogger.Info("📝 %s: %s", message, event.Subject)
		}

		return nil
	}
}

// LogsWebSocket godoc
// @Summary      WebSocket endpoint for real-time logs
// @Description  Establishes a WebSocket connection to stream real-time platform logs and events
// @Tags         logs
// @Accept       json
// @Produce      json
// @Success      101  {string}  string  "Switching Protocols"
// @Router       /v1/logs/stream [get]
func LogsWebSocket(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger().ForComponent("logs-websocket")

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ErrorWithErr(err, "WebSocket upgrade failed")
		return
	}
	defer conn.Close()

	// Register the new client with the real-time sink
	if realtimeLogSink != nil {
		realtimeLogSink.RegisterClient(conn)
		defer realtimeLogSink.UnregisterClient(conn)
	}

	// Send a welcome message
	welcomeMessage := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     "INFO",
		"message":   "🌟 Real-time log stream connected - ZTDP Platform Events",
		"component": "websocket",
		"type":      "connection.established",
		"details": map[string]interface{}{
			"client_ip":  r.RemoteAddr,
			"user_agent": r.Header.Get("User-Agent"),
		},
	}

	if err := conn.WriteJSON(welcomeMessage); err != nil {
		logger.ErrorWithErr(err, "Failed to send welcome message")
		return
	}

	logger.Info("WebSocket client connected from %s", r.RemoteAddr)

	// Keep the connection alive and handle ping/pong
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Send periodic pings
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Read loop to handle client messages and keep connection alive
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Warn("WebSocket error: %v", err)
				}
				return
			}
		}
	}()

	// Ping loop
	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Debug("WebSocket ping failed: %v", err)
				return
			}
		}
	}
}
