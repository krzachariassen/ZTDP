# AI Agent-to-Agent Communication Implementation Plan
## Event-Driven Architecture for Decoupled Agent Coordination

**Date:** June 15, 2025  
**Status:** Planning Phase  
**Priority:** High  
**Complexity:** High  

---

## 🎯 **EXECUTIVE SUMMARY**

Transform ZTDP from hardcoded service-to-service calls to a truly AI-native, event-driven agent ecosystem where:
- Platform Agent coordinates through events, not direct calls
- Specialized agents (Policy, Deployment, Security) operate independently 
- All agent interactions are observable and auditable
- System is highly decoupled, scalable, and resilient

**Key Decision: Hybrid Architecture**
- **Synchronous Agent-to-Agent** for immediate user interactions requiring real-time responses
- **Event-Driven Architecture** for coordination, background processing, and complex workflows

---

## 📊 **CURRENT STATE ANALYSIS**

### **❌ Problems with Current Architecture**

1. **Hardcoded Service Interfaces**
   ```go
   // Current: Tightly coupled
   type PolicyService interface {
       ValidateDeployment(ctx context.Context, app, env string) error
   }
   ```

2. **Direct Method Calls**
   ```go
   // Current: Direct coupling
   err := agent.policyService.ValidateDeployment(ctx, appName, env)
   ```

3. **Platform Agent has Domain Knowledge**
   - Knows about deployment validation
   - Knows about policy evaluation
   - Knows about specific service methods

4. **No Agent Discoverability**
   - Agents can't discover each other's capabilities
   - No dynamic agent routing
   - Hard to add new agents

5. **Limited Observability**
   - Agent interactions are not visible
   - No audit trail of agent decisions
   - Hard to debug agent workflows

---

## 🎯 **TARGET ARCHITECTURE**

### **Core Principles**

1. **Agent Autonomy**: Each agent operates independently with its own AI brain
2. **Event-First**: All agent coordination happens through events
3. **Intent-Based**: Agents communicate through natural language intents, not APIs
4. **Observable**: Every agent interaction is auditable through events
5. **Scalable**: Easy to add new agents without changing existing ones

### **Architecture Overview**

```
┌─────────────────┐     Events     ┌─────────────────┐
│  Platform Agent │◄──────────────►│   Event Bus     │
│   (Coordinator) │                │  (Redis/NATS)   │
└─────────────────┘                └─────────────────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    │                      │                      │
                    ▼                      ▼                      ▼
        ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
        │  Policy Agent   │    │Deployment Agent │    │ Security Agent  │
        │ (Specialized)   │    │ (Specialized)   │    │ (Specialized)   │
        └─────────────────┘    └─────────────────┘    └─────────────────┘
```

### **Communication Patterns**

1. **Request-Response via Events**
   ```
   Platform Agent → [policy.evaluation.requested] → Event Bus
   Event Bus → Policy Agent → [policy.evaluation.completed] → Event Bus
   Event Bus → Platform Agent (receives response)
   ```

2. **Broadcast Notifications**
   ```
   Policy Agent → [policy.violation.detected] → Event Bus
   Event Bus → All Interested Agents (Security, Platform, etc.)
   ```

3. **Workflow Orchestration**
   ```
   Platform Agent → [deployment.workflow.started] → Event Bus
   Event Bus → Policy Agent → [policy.check.completed] → Event Bus
   Event Bus → Deployment Agent → [deployment.executed] → Event Bus
   ```

---

## 📋 **EVENT SCHEMA DESIGN**

### **Core Event Structure**

```go
type AgentEvent struct {
    // Event Identity
    ID          string    `json:"id"`          // Unique event ID
    Type        string    `json:"type"`        // Event type (policy.evaluation.requested)
    Source      string    `json:"source"`      // Source agent ID
    Target      string    `json:"target"`      // Target agent ID (empty for broadcast)
    Timestamp   time.Time `json:"timestamp"`   // Event timestamp
    
    // Correlation
    CorrelationID string  `json:"correlation_id"` // Links related events
    ConversationID string `json:"conversation_id"` // User conversation context
    WorkflowID    string  `json:"workflow_id"`     // Multi-step workflow ID
    
    // Agent Communication
    Intent        string                 `json:"intent"`        // Natural language intent
    Context       map[string]interface{} `json:"context"`       // Rich context data
    GraphSnapshot map[string]interface{} `json:"graph_snapshot"` // Relevant graph data
    
    // Metadata
    Priority      string                 `json:"priority"`      // high/medium/low
    TTL           int                    `json:"ttl"`           // Time to live (seconds)
    Retry         bool                   `json:"retry"`         // Retry on failure
    Metadata      map[string]interface{} `json:"metadata"`      // Additional metadata
}

type AgentResponse struct {
    // Response Identity  
    ID            string    `json:"id"`            // Response ID
    RequestID     string    `json:"request_id"`    // Original request ID
    Source        string    `json:"source"`        // Responding agent
    Timestamp     time.Time `json:"timestamp"`     // Response timestamp
    
    // Response Data
    Success       bool                   `json:"success"`       // Success/failure
    Decision      string                 `json:"decision"`      // allowed/blocked/conditional
    Reasoning     string                 `json:"reasoning"`     // AI reasoning
    Confidence    float64               `json:"confidence"`    // Confidence score
    Suggestions   []string              `json:"suggestions"`   // Recommendations
    
    // Result Data
    Result        map[string]interface{} `json:"result"`        // Structured result
    Error         string                 `json:"error"`         // Error message if failed
    Metadata      map[string]interface{} `json:"metadata"`      // Additional metadata
}
```

### **Event Types**

#### **Policy Events**
```
- policy.evaluation.requested
- policy.evaluation.completed
- policy.violation.detected
- policy.compliance.checked
- policy.recommendation.generated
```

#### **Deployment Events**
```
- deployment.plan.requested
- deployment.plan.generated
- deployment.execution.requested
- deployment.execution.completed
- deployment.rollback.requested
```

#### **Platform Coordination Events**
```
- platform.agent.query.received
- platform.workflow.started
- platform.workflow.completed
- platform.agent.registration
- platform.capability.discovered
```

#### **Monitoring & Observability Events**
```
- agent.health.check
- agent.performance.metric
- agent.error.occurred
- agent.capability.updated
```

---

## 🔧 **AGENT INTERFACE DESIGN**

### **Generic Agent Interface**

```go
// AgentInterface - Universal interface for all AI agents
type AgentInterface interface {
    // Core Agent Operations
    ProcessEvent(ctx context.Context, event *AgentEvent) (*AgentResponse, error)
    GetCapabilities() []AgentCapability
    GetStatus() AgentStatus
    
    // Lifecycle
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() HealthStatus
}

type AgentCapability struct {
    Type        string   `json:"type"`        // "policy_evaluation", "deployment_planning"
    Description string   `json:"description"` // Human-readable description
    Intents     []string `json:"intents"`     // Supported intent patterns
    InputTypes  []string `json:"input_types"` // Expected input data types
    OutputTypes []string `json:"output_types"` // Response data types
}

type AgentStatus struct {
    ID           string    `json:"id"`
    Type         string    `json:"type"`         // "platform", "policy", "deployment"
    Status       string    `json:"status"`       // "running", "idle", "busy", "error"
    LastActivity time.Time `json:"last_activity"`
    LoadFactor   float64   `json:"load_factor"`  // 0.0 to 1.0
    Version      string    `json:"version"`
}
```

### **Event Bus Interface**

```go
// EventBus - Decoupled event communication
type EventBus interface {
    // Publishing
    Publish(ctx context.Context, event *AgentEvent) error
    PublishResponse(ctx context.Context, response *AgentResponse) error
    
    // Subscribing
    Subscribe(ctx context.Context, agentID string, eventTypes []string, handler EventHandler) error
    SubscribeToResponses(ctx context.Context, agentID string, handler ResponseHandler) error
    
    // Request-Response Pattern
    Request(ctx context.Context, event *AgentEvent, timeout time.Duration) (*AgentResponse, error)
    
    // Agent Discovery
    RegisterAgent(ctx context.Context, agent AgentInterface) error
    UnregisterAgent(ctx context.Context, agentID string) error
    DiscoverAgents(ctx context.Context, capability string) ([]AgentStatus, error)
    
    // Management
    GetMetrics() EventBusMetrics
    Close() error
}

type EventHandler func(ctx context.Context, event *AgentEvent) error
type ResponseHandler func(ctx context.Context, response *AgentResponse) error
```

---

## 🚀 **IMPLEMENTATION PHASES**

### **Phase 1: Foundation (Week 1-2)**

**Deliverables:**
- [ ] Event bus implementation (Redis/NATS)
- [ ] Core event types and schemas
- [ ] Generic agent interface
- [ ] Agent registration and discovery
- [ ] Basic event publishing/subscribing

**Files to Create:**
```
internal/events/
├── bus.go              # Event bus implementation
├── types.go            # Event schemas
├── agent_registry.go   # Agent discovery
└── metrics.go          # Event metrics

internal/agents/
├── interface.go        # Generic agent interface
├── base_agent.go       # Base agent implementation
└── registry.go         # Agent registry
```

### **Phase 2: Policy Agent Integration (Week 3)**

**Deliverables:**
- [ ] Refactor Policy service to PolicyAgent
- [ ] Event-based policy evaluation
- [ ] Policy Agent event handlers
- [ ] Integration with existing policy system

**Files to Modify:**
```
internal/policies/
├── agent.go            # NEW: Policy Agent implementation
├── event_handlers.go   # NEW: Event processing
└── service.go          # MODIFY: Integrate with agent

internal/ai/
└── platform_agent.go  # MODIFY: Use events for policy calls
```

### **Phase 3: Platform Agent Refactoring (Week 4)**

**Deliverables:**
- [ ] Remove hardcoded PolicyService interface
- [ ] Implement event-based coordination
- [ ] AI-driven intent formulation for policy questions
- [ ] Request-response correlation

**Changes:**
```go
// REMOVE
type PolicyService interface {
    ValidateDeployment(ctx context.Context, app, env string) error
}

// ADD
func (agent *PlatformAgent) consultPolicyAgent(ctx context.Context, intent string, context map[string]interface{}) (*AgentResponse, error) {
    event := &AgentEvent{
        Type:   "policy.evaluation.requested",
        Intent: intent,
        Context: context,
        // ...
    }
    return agent.eventBus.Request(ctx, event, 30*time.Second)
}
```

### **Phase 4: Additional Agents (Week 5-6)**

**Deliverables:**
- [ ] Deployment Agent implementation
- [ ] Security Agent implementation
- [ ] Agent capability discovery
- [ ] Multi-agent workflows

### **Phase 5: Advanced Features (Week 7-8)**

**Deliverables:**
- [ ] Agent performance monitoring
- [ ] Event replay and debugging
- [ ] Agent load balancing
- [ ] Failure recovery and retry logic

---

## 🔄 **MIGRATION STRATEGY**

### **Backward Compatibility Approach**

1. **Dual Implementation**
   ```go
   // Keep old interface during transition
   type PolicyService interface {
       ValidateDeployment(ctx context.Context, app, env string) error
   }
   
   // Add new event-based wrapper
   func (ps *PolicyServiceEventWrapper) ValidateDeployment(ctx context.Context, app, env string) error {
       event := &AgentEvent{
           Type: "policy.deployment.validation.requested",
           Context: map[string]interface{}{
               "application": app,
               "environment": env,
           },
       }
       response, err := ps.eventBus.Request(ctx, event, 30*time.Second)
       if err != nil {
           return err
       }
       if !response.Success {
           return errors.New(response.Error)
       }
       return nil
   }
   ```

2. **Feature Flags**
   ```go
   if os.Getenv("ENABLE_AGENT_EVENTS") == "true" {
       // Use event-based communication
       return agent.consultPolicyAgentViaEvents(ctx, intent, context)
   } else {
       // Use legacy service calls
       return agent.policyService.ValidateDeployment(ctx, app, env)
   }
   ```

3. **Gradual Rollout**
   - Phase 1: Events alongside existing calls (observe only)
   - Phase 2: Events with fallback to service calls
   - Phase 3: Events only, remove service interfaces

---

## 🧪 **TESTING STRATEGY**

### **Unit Testing**

```go
func TestPolicyAgentEventProcessing(t *testing.T) {
    // Test event processing in isolation
    agent := NewPolicyAgent(mockAIProvider, mockGraph, mockEventBus)
    
    event := &AgentEvent{
        Type: "policy.evaluation.requested",
        Intent: "Validate deployment of app-x to production",
        Context: map[string]interface{}{
            "application": "app-x",
            "environment": "production",
        },
    }
    
    response, err := agent.ProcessEvent(ctx, event)
    assert.NoError(t, err)
    assert.Equal(t, "blocked", response.Decision)
}
```

### **Integration Testing**

```go
func TestPlatformToPolicyAgentWorkflow(t *testing.T) {
    // Test full agent-to-agent communication
    eventBus := NewTestEventBus()
    platformAgent := NewPlatformAgent(..., eventBus)
    policyAgent := NewPolicyAgent(..., eventBus)
    
    // Register agents
    eventBus.RegisterAgent(ctx, platformAgent)
    eventBus.RegisterAgent(ctx, policyAgent)
    
    // Test policy consultation
    result, err := platformAgent.ChatWithPlatform(ctx, 
        "Deploy my-app to production", 
        "")
    
    assert.NoError(t, err)
    assert.Contains(t, result.Answer, "policy")
}
```

### **End-to-End Testing**

```go
func TestFullUserJourneyWithAgents(t *testing.T) {
    // Test complete user interaction through agents
    // User Query → Platform Agent → Policy Agent → Response
}
```

---

## ⚠️ **RISK MITIGATION**

### **Technical Risks**

1. **Event Bus Failure**
   - **Risk:** Single point of failure
   - **Mitigation:** Event bus clustering, fallback to direct calls
   
2. **Message Ordering**
   - **Risk:** Out-of-order event processing
   - **Mitigation:** Event sequencing, correlation IDs
   
3. **Agent Failures**
   - **Risk:** Agent becomes unresponsive
   - **Mitigation:** Health checks, circuit breakers, timeouts

4. **Performance Overhead**
   - **Risk:** Events slower than direct calls
   - **Mitigation:** Hybrid approach, caching, optimized serialization

### **Architectural Risks**

1. **Complexity Growth**
   - **Risk:** System becomes too complex
   - **Mitigation:** Clear documentation, testing, monitoring

2. **Debugging Difficulty**
   - **Risk:** Hard to trace multi-agent workflows
   - **Mitigation:** Correlation IDs, event tracing, debugging tools

---

## 📈 **SUCCESS METRICS**

### **Decoupling Metrics**
- [ ] Zero direct agent-to-agent dependencies
- [ ] All agent interactions observable via events
- [ ] New agents can be added without modifying existing agents

### **Performance Metrics**
- [ ] Agent response times < 500ms for 95th percentile
- [ ] Event processing latency < 100ms
- [ ] System throughput maintains or improves

### **Reliability Metrics**
- [ ] Agent failure recovery < 5 seconds
- [ ] Event delivery success rate > 99.9%
- [ ] No data loss during agent failures

---

## 🎯 **IMPLEMENTATION PRIORITIES**

### **Must Have (Critical)**
1. ✅ Event bus implementation
2. ✅ Policy Agent event integration
3. ✅ Platform Agent refactoring
4. ✅ Basic request-response patterns

### **Should Have (Important)**
1. ⚠️ Agent discovery and registration
2. ⚠️ Event correlation and tracing
3. ⚠️ Performance monitoring
4. ⚠️ Failure recovery

### **Could Have (Nice to Have)**
1. 💡 Event replay for debugging
2. 💡 Agent load balancing
3. 💡 Advanced workflow orchestration
4. 💡 Event analytics and insights

---

## 📝 **NEXT STEPS**

### **Immediate Actions (This Week)**
1. [ ] Create event bus implementation in `internal/events/`
2. [ ] Design and implement core event schemas
3. [ ] Create generic agent interface in `internal/agents/`
4. [ ] Set up basic agent registration

### **Next Week**
1. [ ] Refactor Policy service to Policy Agent
2. [ ] Implement event handlers for policy evaluation
3. [ ] Create event-based wrapper for backward compatibility
4. [ ] Add feature flags for gradual rollout

### **Technical Decisions Needed**
1. **Event Bus Technology:** Redis Streams vs NATS vs Apache Kafka
2. **Serialization:** JSON vs Protocol Buffers vs MessagePack  
3. **Agent Discovery:** Registry vs Service Discovery vs DNS
4. **Event Storage:** In-memory vs Persistent vs Hybrid

---

## 💭 **DESIGN DECISIONS & RATIONALE**

### **Why Hybrid Architecture?**
- **Synchronous** for user-facing interactions requiring immediate responses
- **Asynchronous Events** for coordination, workflows, and background processing
- Best of both worlds: responsiveness + decoupling

### **Why Natural Language Intents?**
- Leverages AI capabilities of agents
- More flexible than rigid API contracts
- Easier to extend and modify
- Better debuggability and observability

### **Why Event-First Design?**
- Enables true decoupling between agents
- Provides audit trail of all agent interactions
- Supports complex multi-agent workflows
- Makes system observable and debuggable

---

**This implementation plan transforms ZTDP into a truly AI-native, event-driven agent ecosystem where agents coordinate intelligently without hardcoded dependencies.**
