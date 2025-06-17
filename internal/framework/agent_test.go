package framework

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agents"
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/logging"
)

// TestAgentFramework tests the agent framework using patterns from deployment_agent.go
func TestAgentFramework(t *testing.T) {
	// Initialize logging for tests
	logging.InitializeLogger("test", logging.LevelInfo)
	
	t.Run("should create agent with auto-registration", func(t *testing.T) {
		// Given: A mock registry and event bus
		registry := agents.NewInMemoryAgentRegistry()
		eventBus := events.NewEventBus(nil, false)
		
		// When: Creating a new agent with the framework
		agent, err := NewBaseAgent(BaseAgentConfig{
			ID:            "test-agent",
			Type:          "test",
			Environment:   "test",
			EventBus:      eventBus,
			AgentRegistry: registry,
			Capabilities: []agents.AgentCapability{
				{
					Name:        "test_capability",
					Description: "Test capability for framework testing",
					Intents:     []string{"test intent"},
					RoutingKeys: []string{"test.route"},
					Version:     "1.0.0",
				},
			},
		})
		
		// Then: Agent should be created successfully
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if agent == nil {
			t.Fatal("Expected agent to be created")
		}
		
		// And: Agent should be registered
		if agent.GetID() != "test-agent" {
			t.Errorf("Expected agent ID 'test-agent', got: %s", agent.GetID())
		}
		
		// And: Agent should have correct status
		status := agent.GetStatus()
		if status.Type != "test" {
			t.Errorf("Expected agent type 'test', got: %s", status.Type)
		}
	})
	
	t.Run("should handle event processing with intent routing", func(t *testing.T) {
		// Given: An agent with intent handlers
		registry := agents.NewInMemoryAgentRegistry()
		eventBus := events.NewEventBus(nil, false)
		
		var processedIntent string
		var processedPayload map[string]interface{}
		
		agent, err := NewBaseAgent(BaseAgentConfig{
			ID:            "test-agent",
			Type:          "test",
			Environment:   "test",
			EventBus:      eventBus,
			AgentRegistry: registry,
			Capabilities: []agents.AgentCapability{
				{
					Name:        "test_capability",
					Intents:     []string{"test intent"},
					RoutingKeys: []string{"test.route"},
					Version:     "1.0.0",
				},
			},
			// Intent handler similar to deployment_agent.go
			IntentHandlers: map[string]IntentHandler{
				"test intent": func(ctx context.Context, event *events.Event) (*events.Event, error) {
					processedIntent = event.Payload["intent"].(string)
					processedPayload = event.Payload
					
					return &events.Event{
						Type:    events.EventTypeResponse,
						Source:  "test-agent",
						Subject: "Test completed",
						Payload: map[string]interface{}{
							"status": "success",
							"result": "test processed",
						},
					}, nil
				},
			},
		})
		
		if err != nil {
			t.Fatalf("Failed to create agent: %v", err)
		}
		
		// When: Processing an event with a matching intent
		event := &events.Event{
			Type:    events.EventTypeRequest,
			Source:  "test-source",
			Subject: "test.route",
			Payload: map[string]interface{}{
				"intent": "test intent",
				"data":   "test data",
			},
		}
		
		response, err := agent.ProcessEvent(context.Background(), event)
		
		// Then: Event should be processed successfully
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if response == nil {
			t.Fatal("Expected response, got nil")
		}
		
		// And: Intent should be processed correctly
		if processedIntent != "test intent" {
			t.Errorf("Expected processed intent 'test intent', got: %s", processedIntent)
		}
		
		// And: Response should have correct format
		if response.Payload["status"] != "success" {
			t.Errorf("Expected status 'success', got: %v", response.Payload["status"])
		}
	})
	
	t.Run("should support event subscription and auto-routing", func(t *testing.T) {
		// Given: An agent that subscribes to routing keys
		registry := agents.NewInMemoryAgentRegistry()
		eventBus := events.NewEventBus(nil, false)
		
		var receivedEvents []events.Event
		
		agent, err := NewBaseAgent(BaseAgentConfig{
			ID:            "subscriber-agent",
			Type:          "subscriber",
			Environment:   "test",
			EventBus:      eventBus,
			AgentRegistry: registry,
			Capabilities: []agents.AgentCapability{
				{
					Name:        "subscription_capability",
					Intents:     []string{"subscribe intent"},
					RoutingKeys: []string{"subscribe.test"},
					Version:     "1.0.0",
				},
			},
			IntentHandlers: map[string]IntentHandler{
				"subscribe intent": func(ctx context.Context, event *events.Event) (*events.Event, error) {
					receivedEvents = append(receivedEvents, *event)
					return nil, nil
				},
			},
		})
		
		if err != nil {
			t.Fatalf("Failed to create agent: %v", err)
		}
		
		// When: An event is emitted on the subscribed routing key
		eventBus.Emit(events.EventTypeNotify, "test-source", "subscribe.test", map[string]interface{}{
			"intent": "subscribe intent",
			"data":   "subscription test",
		})
		
		// Give some time for async processing
		time.Sleep(10 * time.Millisecond)
		
		// Then: Agent should have received the event
		if len(receivedEvents) != 1 {
			t.Errorf("Expected 1 received event, got: %d", len(receivedEvents))
		}
		
		if len(receivedEvents) > 0 {
			if receivedEvents[0].Payload["data"] != "subscription test" {
				t.Errorf("Expected data 'subscription test', got: %v", receivedEvents[0].Payload["data"])
			}
		}
	})
	
	t.Run("should provide logging and error handling", func(t *testing.T) {
		// Given: An agent with error-prone intent handler
		registry := agents.NewInMemoryAgentRegistry()
		eventBus := events.NewEventBus(nil, false)
		
		agent, err := NewBaseAgent(BaseAgentConfig{
			ID:            "error-agent",
			Type:          "error",
			Environment:   "test",
			EventBus:      eventBus,
			AgentRegistry: registry,
			Capabilities: []agents.AgentCapability{
				{
					Name:        "error_capability",
					Intents:     []string{"error intent"},
					RoutingKeys: []string{"error.test"},
					Version:     "1.0.0",
				},
			},
			IntentHandlers: map[string]IntentHandler{
				"error intent": func(ctx context.Context, event *events.Event) (*events.Event, error) {
					return nil, fmt.Errorf("simulated error")
				},
			},
		})
		
		if err != nil {
			t.Fatalf("Failed to create agent: %v", err)
		}
		
		// When: Processing an event that causes an error
		event := &events.Event{
			Type:    events.EventTypeRequest,
			Source:  "test-source",
			Subject: "error.test",
			Payload: map[string]interface{}{
				"intent": "error intent",
			},
		}
		
		response, err := agent.ProcessEvent(context.Background(), event)
		
		// Then: Error should be handled gracefully
		if err == nil {
			t.Error("Expected error, got nil")
		}
		
		// And: Response should be nil or error response
		if response != nil {
			if status, ok := response.Payload["status"]; ok && status != "error" {
				t.Errorf("Expected error status, got: %v", status)
			}
		}
	})
}
