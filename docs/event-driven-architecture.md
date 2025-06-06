# Event-Driven Architecture in ZTDP

## Introduction

ZTDP uses event-driven architecture to enable decoupled communication between domain services, AI agents, and external systems. This architecture provides the foundation for real-time collaboration, system observability, and multi-agent coordination.

## Core Event Architecture

### 1. Event Structure

All events in ZTDP follow a consistent structure:

```go
type Event struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`      // e.g., "deployment.started"
    Source    string                 `json:"source"`    // Service/Agent ID
    Target    string                 `json:"target"`    // Target service or "*" for broadcast
    Data      map[string]interface{} `json:"data"`      // Event payload
    Timestamp time.Time              `json:"timestamp"`
    Metadata  map[string]string      `json:"metadata"`  // Additional context
    CorrelationID string             `json:"correlation_id"` // Request tracing
}
```

### 2. Event Bus Interface

The event bus provides the core communication infrastructure:

```go
type EventBus interface {
    // Publishing events
    Publish(ctx context.Context, event *Event) error
    PublishBatch(ctx context.Context, events []*Event) error
    
    // Subscribing to events
    Subscribe(serviceID string, eventTypes []string, handler EventHandler) error
    Unsubscribe(serviceID string, eventTypes []string) error
    
    // Event streaming
    Stream(ctx context.Context, filter EventFilter) (<-chan *Event, error)
    
    // Management
    Close() error
}

type EventHandler func(ctx context.Context, event *Event) error

type EventFilter struct {
    Types     []string  `json:"types"`
    Sources   []string  `json:"sources"`
    Since     time.Time `json:"since"`
    Limit     int       `json:"limit"`
}
```

### 3. Event Categories

Events are categorized by domain and operation type:

#### Domain Events
- **Deployment**: `deployment.started`, `deployment.completed`, `deployment.failed`
- **Application**: `application.created`, `application.updated`, `application.deleted`
- **Policy**: `policy.evaluated`, `policy.violated`, `policy.updated`
- **Security**: `security.scan.completed`, `security.vulnerability.found`

#### System Events
- **Health**: `service.healthy`, `service.unhealthy`, `service.recovering`
- **Performance**: `performance.threshold.exceeded`, `performance.optimized`
- **Error**: `error.occurred`, `error.resolved`

#### AI Agent Events
- **Agent Communication**: `agent.request`, `agent.response`, `agent.collaboration`
- **AI Operations**: `ai.plan.generated`, `ai.decision.made`, `ai.feedback.received`

## Event Communication Patterns

### 1. Request-Response Pattern

Synchronous-style communication using events:

```go
// Core Agent requests deployment plan
func (ca *CoreAgent) RequestDeploymentPlan(ctx context.Context, app, env string) (*DeploymentPlan, error) {
    requestID := generateRequestID()
    
    // Publish request event
    requestEvent := &Event{
        Type:   "deployment.plan.request",
        Source: "core-agent",
        Target: "deployment-agent",
        Data: map[string]interface{}{
            "application": app,
            "environment": env,
            "requestID":   requestID,
        },
        CorrelationID: requestID,
    }
    
    // Subscribe to response
    responseChan := make(chan *Event, 1)
    defer close(responseChan)
    
    ca.eventBus.Subscribe("core-agent", []string{"deployment.plan.response"}, 
        func(ctx context.Context, event *Event) error {
            if event.Data["requestID"] == requestID {
                responseChan <- event
            }
            return nil
        })
    
    // Publish request
    if err := ca.eventBus.Publish(ctx, requestEvent); err != nil {
        return nil, err
    }
    
    // Wait for response with timeout
    select {
    case response := <-responseChan:
        return parseDeploymentPlan(response.Data["plan"])
    case <-time.After(30 * time.Second):
        return nil, errors.New("timeout waiting for deployment plan")
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

// Deployment Agent handles request
func (da *DeploymentAgent) HandlePlanRequest(ctx context.Context, event *Event) error {
    app := event.Data["application"].(string)
    env := event.Data["environment"].(string)
    requestID := event.Data["requestID"].(string)
    
    // Generate deployment plan
    plan, err := da.GeneratePlan(ctx, app, env)
    if err != nil {
        return da.publishError(requestID, err)
    }
    
    // Publish response
    responseEvent := &Event{
        Type:   "deployment.plan.response",
        Source: "deployment-agent",
        Target: "core-agent",
        Data: map[string]interface{}{
            "requestID": requestID,
            "plan":      plan,
            "confidence": plan.Confidence,
        },
        CorrelationID: requestID,
    }
    
    return da.eventBus.Publish(ctx, responseEvent)
}
```

### 2. Publish-Subscribe Pattern

Broadcast events for multiple subscribers:

```go
// Deployment service publishes deployment events
func (s *DeploymentService) ExecuteDeployment(ctx context.Context, plan *DeploymentPlan) error {
    // Emit deployment started event
    startEvent := &Event{
        Type:   "deployment.started",
        Source: "deployment-service",
        Target: "*", // Broadcast
        Data: map[string]interface{}{
            "deployment_id": plan.ID,
            "application":   plan.Application,
            "environment":   plan.Environment,
            "plan":          plan,
        },
        CorrelationID: plan.ID,
    }
    
    s.eventBus.Publish(ctx, startEvent)
    
    // Execute deployment steps
    err := s.executeDeploymentSteps(ctx, plan)
    
    // Emit completion or failure event
    if err != nil {
        failureEvent := &Event{
            Type:   "deployment.failed",
            Source: "deployment-service",
            Target: "*",
            Data: map[string]interface{}{
                "deployment_id": plan.ID,
                "application":   plan.Application,
                "error":         err.Error(),
                "failed_step":   plan.CurrentStep,
            },
            CorrelationID: plan.ID,
        }
        s.eventBus.Publish(ctx, failureEvent)
        return err
    }
    
    successEvent := &Event{
        Type:   "deployment.completed",
        Source: "deployment-service",
        Target: "*",
        Data: map[string]interface{}{
            "deployment_id": plan.ID,
            "application":   plan.Application,
            "duration":      time.Since(plan.StartTime),
            "version":       plan.TargetVersion,
        },
        CorrelationID: plan.ID,
    }
    
    return s.eventBus.Publish(ctx, successEvent)
}

// Multiple services can subscribe to deployment events
func (ps *PolicyService) SubscribeToDeploymentEvents() {
    ps.eventBus.Subscribe("policy-service", 
        []string{"deployment.started", "deployment.completed"}, 
        ps.handleDeploymentEvent)
}

func (ms *MonitoringService) SubscribeToDeploymentEvents() {
    ms.eventBus.Subscribe("monitoring-service",
        []string{"deployment.started", "deployment.completed", "deployment.failed"},
        ms.handleDeploymentEvent)
}
```

### 3. Event Sourcing Pattern

Capture all state changes as events:

```go
// Application aggregate emits events for all state changes
type Application struct {
    id       ApplicationID
    version  int
    events   []*Event
}

func (a *Application) AddService(service *Service) error {
    // Validate business rules
    if err := a.validateAddService(service); err != nil {
        return err
    }
    
    // Create event
    event := &Event{
        Type:   "application.service.added",
        Source: "application-domain",
        Data: map[string]interface{}{
            "application_id": a.id,
            "service":        service,
            "version":        a.version + 1,
        },
    }
    
    // Apply event
    a.applyEvent(event)
    return nil
}

func (a *Application) applyEvent(event *Event) {
    switch event.Type {
    case "application.service.added":
        service := event.Data["service"].(*Service)
        a.services = append(a.services, service)
        a.version++
    case "application.service.removed":
        serviceID := event.Data["service_id"].(string)
        a.removeServiceByID(serviceID)
        a.version++
    }
    
    a.events = append(a.events, event)
}

// Repository publishes events when saving
func (r *ApplicationRepository) Save(ctx context.Context, app *Application) error {
    // Save application state
    if err := r.saveToDatabase(app); err != nil {
        return err
    }
    
    // Publish all uncommitted events
    for _, event := range app.events {
        if err := r.eventBus.Publish(ctx, event); err != nil {
            // Rollback and return error
            return err
        }
    }
    
    // Clear events after successful publish
    app.events = nil
    return nil
}
```

## AI Agent Communication

### 1. Multi-Agent Coordination

Agents coordinate through events to accomplish complex tasks:

```go
// Core Agent orchestrates multi-step operation
func (ca *CoreAgent) DeployWithGovernance(ctx context.Context, request *DeploymentRequest) error {
    operationID := generateOperationID()
    
    // Step 1: Request policy evaluation
    policyEvent := &Event{
        Type:   "policy.evaluation.request",
        Source: "core-agent",
        Target: "governance-agent",
        Data: map[string]interface{}{
            "operation_id": operationID,
            "application":  request.Application,
            "environment":  request.Environment,
            "changes":      request.Changes,
        },
        CorrelationID: operationID,
    }
    
    if err := ca.eventBus.Publish(ctx, policyEvent); err != nil {
        return err
    }
    
    // Subscribe to policy response
    return ca.waitForPolicyApproval(ctx, operationID, request)
}

// Governance Agent evaluates policies
func (ga *GovernanceAgent) HandlePolicyEvaluation(ctx context.Context, event *Event) error {
    operationID := event.Data["operation_id"].(string)
    app := event.Data["application"].(string)
    env := event.Data["environment"].(string)
    
    // Evaluate policies
    result, err := ga.evaluatePolicies(ctx, app, env, event.Data["changes"])
    if err != nil {
        return ga.publishPolicyError(operationID, err)
    }
    
    // Publish result
    resultEvent := &Event{
        Type:   "policy.evaluation.result",
        Source: "governance-agent",
        Target: "core-agent",
        Data: map[string]interface{}{
            "operation_id": operationID,
            "approved":     result.Approved,
            "violations":   result.Violations,
            "requirements": result.Requirements,
        },
        CorrelationID: operationID,
    }
    
    return ga.eventBus.Publish(ctx, resultEvent)
}
```

### 2. Agent Learning and Feedback

Agents share learnings through events:

```go
// Deployment Agent shares learning from deployment outcomes
func (da *DeploymentAgent) ShareDeploymentLearning(deployment *Deployment, outcome *DeploymentOutcome) error {
    learningEvent := &Event{
        Type:   "ai.learning.deployment",
        Source: "deployment-agent",
        Target: "*", // All agents can learn
        Data: map[string]interface{}{
            "deployment_strategy": deployment.Strategy,
            "application_type":    deployment.ApplicationType,
            "environment":         deployment.Environment,
            "outcome":            outcome,
            "metrics":            outcome.Metrics,
            "lessons":            outcome.Lessons,
        },
    }
    
    return da.eventBus.Publish(context.Background(), learningEvent)
}

// Other agents can learn from deployment outcomes
func (oa *OptimizationAgent) HandleDeploymentLearning(ctx context.Context, event *Event) error {
    outcome := event.Data["outcome"].(*DeploymentOutcome)
    
    // Update optimization models based on deployment results
    if outcome.Success {
        oa.reinforceStrategy(event.Data["deployment_strategy"])
    } else {
        oa.penalizeStrategy(event.Data["deployment_strategy"], outcome.FailureReason)
    }
    
    return nil
}
```

## Real-Time Event Streaming

### 1. WebSocket Event Streaming

Provide real-time updates to clients:

```go
// WebSocket handler for event streaming
func (h *EventHandler) HandleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Error("WebSocket upgrade failed", "error", err)
        return
    }
    defer conn.Close()
    
    // Parse subscription filters from query parameters
    filter := parseEventFilter(r.URL.Query())
    
    // Create event stream
    eventChan, err := h.eventBus.Stream(r.Context(), filter)
    if err != nil {
        conn.WriteJSON(map[string]string{"error": err.Error()})
        return
    }
    
    // Stream events to client
    for {
        select {
        case event := <-eventChan:
            if err := conn.WriteJSON(event); err != nil {
                log.Error("WebSocket write failed", "error", err)
                return
            }
            
        case <-r.Context().Done():
            return
        }
    }
}

// Client-side event subscription
func subscribeToDeploymentEvents() {
    ws, err := websocket.Dial("ws://localhost:8080/events?types=deployment.*")
    if err != nil {
        log.Fatal("WebSocket connection failed", err)
    }
    defer ws.Close()
    
    for {
        var event Event
        if err := ws.ReadJSON(&event); err != nil {
            log.Error("Read error", err)
            break
        }
        
        handleDeploymentEvent(&event)
    }
}
```

### 2. Event Filtering and Routing

Advanced event filtering for efficient communication:

```go
// Event filter with complex conditions
type EventFilter struct {
    Types       []string            `json:"types"`        // Event type patterns
    Sources     []string            `json:"sources"`      // Source services
    Since       time.Time           `json:"since"`        // Time range
    Attributes  map[string]string   `json:"attributes"`   // Custom attributes
    Conditions  []FilterCondition   `json:"conditions"`   // Complex conditions
}

type FilterCondition struct {
    Field    string      `json:"field"`     // Data field path
    Operator string      `json:"operator"`  // eq, ne, gt, lt, contains
    Value    interface{} `json:"value"`     // Expected value
}

// Example: Subscribe to high-priority production deployments
filter := &EventFilter{
    Types: []string{"deployment.*"},
    Attributes: map[string]string{
        "environment": "production",
    },
    Conditions: []FilterCondition{
        {
            Field:    "data.priority",
            Operator: "gte",
            Value:    "high",
        },
    },
}
```

## Event Store and Persistence

### 1. Event Storage

Persist events for replay and analysis:

```go
type EventStore interface {
    Append(ctx context.Context, streamID string, events []*Event) error
    Read(ctx context.Context, streamID string, from int) ([]*Event, error)
    ReadAll(ctx context.Context, filter EventFilter) ([]*Event, error)
    Subscribe(ctx context.Context, filter EventFilter) (<-chan *Event, error)
}

// Event store implementation
func (es *NATSEventStore) Append(ctx context.Context, streamID string, events []*Event) error {
    for _, event := range events {
        data, err := json.Marshal(event)
        if err != nil {
            return err
        }
        
        subject := fmt.Sprintf("events.%s.%s", streamID, event.Type)
        if err := es.js.Publish(subject, data); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 2. Event Replay and Recovery

Replay events for system recovery:

```go
// Replay events to rebuild system state
func (s *ApplicationService) ReplayEvents(ctx context.Context, from time.Time) error {
    filter := EventFilter{
        Types: []string{"application.*"},
        Since: from,
    }
    
    events, err := s.eventStore.ReadAll(ctx, filter)
    if err != nil {
        return err
    }
    
    for _, event := range events {
        if err := s.applyEvent(event); err != nil {
            return fmt.Errorf("failed to apply event %s: %w", event.ID, err)
        }
    }
    
    return nil
}

func (s *ApplicationService) applyEvent(event *Event) error {
    switch event.Type {
    case "application.created":
        return s.handleApplicationCreated(event)
    case "application.updated":
        return s.handleApplicationUpdated(event)
    case "application.deleted":
        return s.handleApplicationDeleted(event)
    default:
        log.Warn("Unknown event type", "type", event.Type)
        return nil
    }
}
```

## Error Handling and Resilience

### 1. Event Processing Errors

Handle failures gracefully:

```go
// Resilient event handler with retry logic
func (h *ResilientEventHandler) HandleEvent(ctx context.Context, event *Event) error {
    var lastErr error
    
    for attempt := 0; attempt < h.maxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff
            delay := time.Duration(attempt) * time.Duration(attempt) * time.Second
            time.Sleep(delay)
        }
        
        if err := h.processEvent(ctx, event); err != nil {
            lastErr = err
            
            // Check if error is retryable
            if !h.isRetryableError(err) {
                return err
            }
            
            continue
        }
        
        return nil // Success
    }
    
    // Publish to dead letter queue after all retries
    dlqEvent := &Event{
        Type:   "event.processing.failed",
        Source: h.serviceID,
        Data: map[string]interface{}{
            "original_event": event,
            "error":         lastErr.Error(),
            "attempts":      h.maxRetries,
        },
    }
    
    h.eventBus.Publish(ctx, dlqEvent)
    return fmt.Errorf("event processing failed after %d attempts: %w", h.maxRetries, lastErr)
}
```

### 2. Circuit Breaker Pattern

Protect against cascading failures:

```go
type CircuitBreakerEventBus struct {
    eventBus EventBus
    breaker  *CircuitBreaker
}

func (cb *CircuitBreakerEventBus) Publish(ctx context.Context, event *Event) error {
    return cb.breaker.Execute(func() error {
        return cb.eventBus.Publish(ctx, event)
    })
}

// Circuit breaker opens after consecutive failures
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        failureThreshold: threshold,
        timeout:         timeout,
        state:           StateClosed,
    }
}
```

## Testing Event-Driven Systems

### 1. Event Testing Patterns

Test event publishing and handling:

```go
func TestDeploymentService_PublishesCorrectEvents(t *testing.T) {
    // Arrange
    mockEventBus := &MockEventBus{}
    service := NewDeploymentService(mockEventBus, nil, nil)
    
    plan := &DeploymentPlan{
        ID:          "plan-123",
        Application: "test-app",
        Environment: "production",
    }
    
    // Act
    err := service.ExecuteDeployment(context.Background(), plan)
    
    // Assert
    assert.NoError(t, err)
    
    // Verify events were published
    mockEventBus.AssertEventPublished(t, "deployment.started", func(event *Event) bool {
        return event.Data["deployment_id"] == "plan-123"
    })
    
    mockEventBus.AssertEventPublished(t, "deployment.completed", func(event *Event) bool {
        return event.Data["application"] == "test-app"
    })
}

// Mock event bus for testing
type MockEventBus struct {
    publishedEvents []*Event
}

func (m *MockEventBus) Publish(ctx context.Context, event *Event) error {
    m.publishedEvents = append(m.publishedEvents, event)
    return nil
}

func (m *MockEventBus) AssertEventPublished(t *testing.T, eventType string, validator func(*Event) bool) {
    for _, event := range m.publishedEvents {
        if event.Type == eventType && validator(event) {
            return // Found matching event
        }
    }
    t.Errorf("Expected event of type %s not found", eventType)
}
```

### 2. Integration Testing

Test event flows end-to-end:

```go
func TestDeploymentWorkflow_IntegrationTest(t *testing.T) {
    // Setup test environment
    eventBus := NewInMemoryEventBus()
    deploymentService := NewDeploymentService(eventBus, mockGraph, mockAI)
    policyService := NewPolicyService(eventBus, mockPolicyEngine)
    
    // Setup event subscriptions
    policyService.SubscribeToDeploymentEvents()
    
    // Execute test
    plan := &DeploymentPlan{Application: "test-app", Environment: "prod"}
    err := deploymentService.ExecuteDeployment(context.Background(), plan)
    
    // Verify end-to-end behavior
    assert.NoError(t, err)
    
    // Wait for async event processing
    time.Sleep(100 * time.Millisecond)
    
    // Verify policy service received and processed events
    assert.True(t, policyService.HasProcessedDeployment("test-app"))
}
```

## Performance and Scalability

### 1. Event Batching

Optimize throughput with batching:

```go
type BatchingEventBus struct {
    underlying EventBus
    batchSize  int
    batchTime  time.Duration
    pending    []*Event
    mu         sync.Mutex
}

func (b *BatchingEventBus) Publish(ctx context.Context, event *Event) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.pending = append(b.pending, event)
    
    if len(b.pending) >= b.batchSize {
        return b.flushBatch()
    }
    
    return nil
}

func (b *BatchingEventBus) flushBatch() error {
    if len(b.pending) == 0 {
        return nil
    }
    
    err := b.underlying.PublishBatch(context.Background(), b.pending)
    b.pending = b.pending[:0] // Clear batch
    return err
}
```

### 2. Event Partitioning

Scale event processing with partitioning:

```go
type PartitionedEventBus struct {
    partitions []EventBus
    hasher     hash.Hash32
}

func (p *PartitionedEventBus) Publish(ctx context.Context, event *Event) error {
    // Partition by source or correlation ID
    partitionKey := event.Source
    if event.CorrelationID != "" {
        partitionKey = event.CorrelationID
    }
    
    p.hasher.Reset()
    p.hasher.Write([]byte(partitionKey))
    partition := int(p.hasher.Sum32()) % len(p.partitions)
    
    return p.partitions[partition].Publish(ctx, event)
}
```

## Related Documentation

- **[Architecture Overview](architecture-overview.md)** - High-level system design
- **[Domain-Driven Design](domain-driven-design.md)** - Domain boundaries and communication
- **[Testing Strategies](testing-strategies.md)** - Testing event-driven systems
- **[AI Platform Architecture](ai-platform-architecture.md)** - AI agent communication patterns
