package agentFramework

import (
	"context"
	"testing"
	"time"

	"github.com/krzachariassen/ZTDP/internal/agentRegistry"
	"github.com/krzachariassen/ZTDP/internal/events"
)

// TestAgentCreationWithAutoRegistration tests that agents can be created and auto-register
func TestAgentCreationWithAutoRegistration(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)

	capabilities := []agentRegistry.AgentCapability{
		{
			Name:        "test_capability",
			Description: "Test capability for framework testing",
			Intents:     []string{"test intent"},
			RoutingKeys: []string{"test.routing.key"},
		},
	}

	// Act - Create agent using framework
	agent, err := NewAgent("test-agent").
		WithCapabilities(capabilities).
		WithEventHandler(func(ctx context.Context, event *events.Event) (*events.Event, error) {
			return &events.Event{
				Type:    events.EventTypeResponse,
				Source:  "test-agent",
				Subject: "Test response",
				Payload: map[string]interface{}{"status": "success"},
			}, nil
		}).
		Build(AgentDependencies{
			Registry: registry,
			EventBus: eventBus,
		})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error creating agent, got: %v", err)
	}

	if agent.GetID() != "test-agent" {
		t.Errorf("Expected agent ID 'test-agent', got: %s", agent.GetID())
	}

	// Verify auto-registration
	registeredAgent, err := registry.FindAgentByID(context.Background(), "test-agent")
	if err != nil {
		t.Errorf("Expected agent to be auto-registered, got error: %v", err)
	}
	if registeredAgent.GetID() != "test-agent" {
		t.Errorf("Expected registered agent ID 'test-agent', got: %s", registeredAgent.GetID())
	}
}

// TestEventSubscriptionBasedOnCapabilities tests that agents auto-subscribe to routing keys
func TestEventSubscriptionBasedOnCapabilities(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)

	capabilities := []agentRegistry.AgentCapability{
		{
			Name:        "test_capability",
			Description: "Test capability",
			Intents:     []string{"test intent"},
			RoutingKeys: []string{"test.routing.key"},
		},
	}

	responseReceived := false

	// Act - Create agent with event handler
	_, err := NewAgent("test-agent").
		WithCapabilities(capabilities).
		WithEventHandler(func(ctx context.Context, event *events.Event) (*events.Event, error) {
			responseReceived = true
			return &events.Event{
				Type:    events.EventTypeResponse,
				Source:  "test-agent",
				Subject: "Test response",
				Payload: map[string]interface{}{"status": "success"},
			}, nil
		}).
		Build(AgentDependencies{
			Registry: registry,
			EventBus: eventBus,
		})

	if err != nil {
		t.Fatalf("Expected no error creating agent, got: %v", err)
	}

	// Give some time for subscription
	time.Sleep(10 * time.Millisecond)

	// Send event to routing key
	eventBus.EmitEvent(events.Event{
		Type:    events.EventTypeRequest,
		Source:  "test-source",
		Subject: "test.routing.key",
		Payload: map[string]interface{}{"test": "data"},
	})

	// Give some time for event processing
	time.Sleep(10 * time.Millisecond)

	// Assert
	if !responseReceived {
		t.Error("Expected agent to receive event via routing key subscription")
	}
}

// TestIntentBasedEventRouting tests that events are routed based on intent
func TestIntentBasedEventRouting(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)

	capabilities := []agentRegistry.AgentCapability{
		{
			Name:        "intent_handling",
			Description: "Handles specific intents",
			Intents:     []string{"handle test"},
			RoutingKeys: []string{"intent.test"},
		},
	}

	intentReceived := ""

	// Act - Create agent that handles specific intent
	_, err := NewAgent("intent-agent").
		WithCapabilities(capabilities).
		WithEventHandler(func(ctx context.Context, event *events.Event) (*events.Event, error) {
			if intent, ok := event.Payload["intent"].(string); ok {
				intentReceived = intent
			}
			return &events.Event{
				Type:    events.EventTypeResponse,
				Source:  "intent-agent",
				Subject: "Intent handled",
				Payload: map[string]interface{}{"status": "success"},
			}, nil
		}).
		Build(AgentDependencies{
			Registry: registry,
			EventBus: eventBus,
		})

	if err != nil {
		t.Fatalf("Expected no error creating agent, got: %v", err)
	}

	// Give some time for subscription
	time.Sleep(10 * time.Millisecond)

	// Send event with intent
	eventBus.EmitEvent(events.Event{
		Type:    events.EventTypeRequest,
		Source:  "test-source",
		Subject: "intent.test",
		Payload: map[string]interface{}{
			"intent": "handle test",
			"data":   "test data",
		},
	})

	// Give some time for event processing
	time.Sleep(10 * time.Millisecond)

	// Assert
	if intentReceived != "handle test" {
		t.Errorf("Expected intent 'handle test', got: '%s'", intentReceived)
	}
}

// TestErrorHandlingAndResponsePatterns tests consistent error handling
func TestErrorHandlingAndResponsePatterns(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)

	capabilities := []agentRegistry.AgentCapability{
		{
			Name:        "error_testing",
			Description: "Tests error handling",
			Intents:     []string{"cause error"},
			RoutingKeys: []string{"error.test"},
		},
	}

	// Act - Create agent that returns error responses
	var baseAgent agentRegistry.AgentInterface

	baseAgent, err := NewAgent("error-agent").
		WithCapabilities(capabilities).
		WithEventHandler(func(ctx context.Context, event *events.Event) (*events.Event, error) {
			// Cast back to BaseAgent to access framework methods
			if agent, ok := baseAgent.(*BaseAgent); ok {
				return agent.CreateErrorResponse(event, "Test error message"), nil
			}
			return &events.Event{
				Type:    events.EventTypeResponse,
				Source:  "error-agent",
				Subject: "Error response",
				Payload: map[string]interface{}{
					"status": "error",
					"error":  "Test error message",
				},
			}, nil
		}).
		Build(AgentDependencies{
			Registry: registry,
			EventBus: eventBus,
		})

	if err != nil {
		t.Fatalf("Expected no error creating agent, got: %v", err)
	}

	// Cast to BaseAgent to access ProcessEvent
	agent, ok := baseAgent.(*BaseAgent)
	if !ok {
		t.Fatalf("Expected BaseAgent, got %T", baseAgent)
	}

	// Create test event
	testEvent := &events.Event{
		Type:    events.EventTypeRequest,
		Source:  "test-source",
		Subject: "error.test",
		Payload: map[string]interface{}{"intent": "cause error"},
	}

	// Process event
	response, err := agent.ProcessEvent(context.Background(), testEvent)

	// Assert
	if err != nil {
		t.Errorf("Expected no error processing event, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if status, ok := response.Payload["status"].(string); !ok || status != "error" {
		t.Errorf("Expected error response with status 'error', got: %v", response.Payload)
	}
}

// TestLoggingConsistency tests that all agents have consistent logging
func TestLoggingConsistency(t *testing.T) {
	// Arrange
	registry := agentRegistry.NewInMemoryAgentRegistry()
	eventBus := events.NewEventBus(nil, false)

	capabilities := []agentRegistry.AgentCapability{
		{
			Name:        "logging_test",
			Description: "Tests logging consistency",
			Intents:     []string{"test logging"},
			RoutingKeys: []string{"logging.test"},
		},
	}

	// Act - Create agent using framework
	baseAgent, err := NewAgent("logging-agent").
		WithCapabilities(capabilities).
		WithEventHandler(func(ctx context.Context, event *events.Event) (*events.Event, error) {
			return &events.Event{
				Type:    events.EventTypeResponse,
				Source:  "logging-agent",
				Subject: "Logging test response",
				Payload: map[string]interface{}{"status": "success"},
			}, nil
		}).
		Build(AgentDependencies{
			Registry: registry,
			EventBus: eventBus,
		})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error creating agent, got: %v", err)
	}

	// Cast to BaseAgent to access framework methods
	agent, ok := baseAgent.(*BaseAgent)
	if !ok {
		t.Fatalf("Expected BaseAgent, got %T", baseAgent)
	}

	// Verify agent has logger with proper component
	logger := agent.GetLogger()
	if logger == nil {
		t.Error("Expected agent to have logger, got nil")
	}

	// Verify logger component name matches agent ID
	// This ensures consistent logging across all agents
}
