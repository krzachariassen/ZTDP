package main

import (
	"fmt"
	"log"
	"time"

	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

func main() {
	// Initialize logging system
	logging.InitializeLogger("test-event-flow", logging.LevelInfo)
	logger := logging.GetLogger()

	// Initialize event system
	eventTransport := events.NewMemoryTransport()
	events.InitializeEventBus(eventTransport)

	// Setup event subscription
	events.GlobalEventBus.Subscribe(events.EventTypeApplicationCreated, func(event events.Event) error {
		logger.Info("🎯 Received Application Created Event: %s from %s", event.Subject, event.Source)
		for k, v := range event.Payload {
			logger.Info("  - %s: %v", k, v)
		}
		return nil
	})

	events.GlobalEventBus.Subscribe(events.EventTypeApplicationUpdated, func(event events.Event) error {
		logger.Info("✏️ Received Application Updated Event: %s from %s", event.Subject, event.Source)
		if payloadType, ok := event.Payload["type"].(string); ok {
			logger.Info("  - Event Type: %s", payloadType)
		}
		return nil
	})

	// Test event emission
	fmt.Println("Testing simplified event system...")

	// Test application created event
	err := events.GlobalEventBus.Emit(
		events.EventTypeApplicationCreated,
		"test-system",
		"my-app",
		map[string]interface{}{
			"owner":       "team-alpha",
			"description": "Test application for event system",
		},
	)
	if err != nil {
		log.Fatalf("Failed to emit application created event: %v", err)
	}

	// Test deployment event (using application.updated with deployment type)
	err = events.GlobalEventBus.Emit(
		events.EventTypeApplicationUpdated,
		"deployment-engine",
		"my-app",
		map[string]interface{}{
			"type":        "deployment_started",
			"environment": "dev",
			"version":     "1.0.0",
		},
	)
	if err != nil {
		log.Fatalf("Failed to emit deployment event: %v", err)
	}

	// Test policy event (using application.updated with policy type)
	err = events.GlobalEventBus.Emit(
		events.EventTypeApplicationUpdated,
		"policy-evaluator",
		"my-app",
		map[string]interface{}{
			"type":       "policy_check_result",
			"policy_id":  "must-deploy-dev-first",
			"success":    true,
			"check_name": "dev-deployment-check",
		},
	)
	if err != nil {
		log.Fatalf("Failed to emit policy event: %v", err)
	}

	// Wait a moment for events to process
	time.Sleep(100 * time.Millisecond)

	fmt.Println("✅ Event system test completed successfully!")
	fmt.Println("📊 All events were emitted and processed through the simplified event bus")
}
