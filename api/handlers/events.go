package handlers

import (
	"github.com/krzachariassen/ZTDP/internal/events"
	"github.com/krzachariassen/ZTDP/internal/policies"
)

var (
	// GlobalEventBus is the shared event bus for the application
	GlobalEventBus *events.EventBus

	// PolicyEventService handles policy-related events
	PolicyEventService *events.PolicyEventService
)

// GlobalGraphEventService is the event service for graph operations
var GlobalGraphEventService *events.GraphEventService

// SetupEventSystem initializes the global event system
func SetupEventSystem(eventBus *events.EventBus, policyEvents *events.PolicyEventService, graphEvents *events.GraphEventService) {
	GlobalEventBus = eventBus
	PolicyEventService = policyEvents
	GlobalGraphEventService = graphEvents

	// Set up policy evaluators to use the event service
	// Start with the default environment
	environments := []string{"default"}

	// If we have a global graph available, check for actual environments
	if GlobalGraph != nil && GlobalGraph.Graph != nil {
		environments = GlobalGraph.Graph.GetEnvironments()
	}

	for _, env := range environments {
		// Get the policy evaluator for this environment
		graphStore := getGraphStore()
		evaluator := policies.NewPolicyEvaluator(graphStore, env)
		// Set the event service
		evaluator.SetEventService(policyEvents)
	}

	// Register event handlers
	setupEventHandlers(eventBus)
}

// setupEventHandlers registers handlers for various events
func setupEventHandlers(eventBus *events.EventBus) {
	// Register policy check event handler
	eventBus.Subscribe(events.EventTypePolicyCheck, func(event events.Event) error {
		// Log policy checks
		logger.Printf("Policy check for %s: %s", event.Subject, event.Status)
		return nil
	})

	// Register transition event handlers
	eventBus.Subscribe(events.EventTypeTransitionAttempt, func(event events.Event) error {
		// Log transition attempts
		logger.Printf("Transition attempt: %s", event.ID)
		return nil
	})

	eventBus.Subscribe(events.EventTypeTransitionApproved, func(event events.Event) error {
		// Log approved transitions
		logger.Printf("Transition approved: %s", event.ID)
		return nil
	})

	eventBus.Subscribe(events.EventTypeTransitionRejected, func(event events.Event) error {
		// Log rejected transitions
		logger.Printf("Transition rejected: %s", event.ID)
		return nil
	})
}
