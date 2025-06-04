# ZTDP AI-Native Platform Architecture

## Executive Summary

ZTDP is transitioning from an API-first platform to an **AI-native platform** where artificial intelligence is the primary interface for developer interactions. This document outlines the complete AI architecture vision, current implementation state, and roadmap to achieve a multi-agent ecosystem.

## Table of Contents

1. [Vision & Strategic Direction](#vision--strategic-direction)
2. [Current State Analysis](#current-state-analysis)
3. [Target Architecture](#target-architecture)
4. [User Experience Patterns](#user-experience-patterns)
5. [Multi-Agent System Design](#multi-agent-system-design)
6. [Implementation Roadmap](#implementation-roadmap)
7. [Technical Architecture](#technical-architecture)
8. [Domain Separation](#domain-separation)
9. [Event-Driven Agent Communication](#event-driven-agent-communication)
10. [Bring Your Own Agent (BYOA)](#bring-your-own-agent-byoa)
11. [Critical Decisions & Trade-offs](#critical-decisions--trade-offs)
12. [**Developer Principles & Best Practices**](#developer-principles--best-practices)
13. [**AI Agent Instructions & Guidelines**](#ai-agent-instructions--guidelines)
14. [**Coding Standards & Architecture Guidelines**](#coding-standards--architecture-guidelines)
15. [Backlog & Next Steps](#backlog--next-steps)

---

## Vision & Strategic Direction

### Business Context

**Original Problem**: API-first approach was deemed a poor investor case - too similar to existing IDP solutions.

**Strategic Pivot**: Shift to AI-native platform where AI is the **primary interface**, not a feature add-on.

### Core AI Vision

ZTDP will become a **conversational infrastructure platform** where:

1. **Developers primarily interact through natural language** with a core AI agent
2. **Specialized AI agents** handle domain-specific operations (deployment, governance, security)
3. **Multi-agent coordination** enables complex, cross-domain automation
4. **"Bring Your Own Agent"** allows customers to integrate custom AI agents
5. **Event-driven architecture** enables agent-to-agent communication

### Success Metrics

- **Primary Interface**: 80%+ of developer interactions happen through AI conversation
- **Agent Ecosystem**: Multiple specialized agents working in coordination
- **Customer Extension**: Customers successfully deploy custom agents
- **Automation Level**: Complex multi-step operations executed with single AI requests

---

## Current State Analysis

### What We Have (MVP v1)

#### ✅ Foundational Infrastructure
- **Graph-based platform** modeling applications, infrastructure, and policies
- **Event-driven architecture** with real-time WebSocket streaming
- **Policy enforcement engine** with graph-based validation
- **AI provider interface** supporting OpenAI GPT models
- **Comprehensive API layer** with deployment, policy, and graph operations

#### ✅ Basic AI Integration
- **AI-enhanced deployment planning** using GPT-4 for intelligent plan generation
- **AI provider abstraction** allowing multiple AI backends
- **Graceful fallback** to traditional planning when AI unavailable
- **Conversational endpoints** for basic platform interaction

#### ⚠️ Current Limitations
- **Single AI instance** instead of multi-agent system
- **API-first UX** - developers still primarily use REST endpoints
- **Limited conversation scope** - basic Q&A rather than action-oriented chat
- **No agent-to-agent communication** - monolithic AI brain approach
- **Mixed architectural patterns** - some business logic in AI layer vs domain services

### Current File Structure
```
internal/ai/
├── ai_brain.go          # 🚀 Core Platform Agent (PlatformAI) - THE AI-NATIVE INTERFACE
├── ai_provider.go       # ✅ Clean infrastructure interface
├── openai_provider.go   # ✅ OpenAI implementation
└── ai_test.go          # ✅ Comprehensive test coverage

api/handlers/
├── ai.go               # ⚠️ Mixed patterns - should primarily route through PlatformAI

internal/deployments/
├── service.go          # ⚠️ Needs better integration with PlatformAI orchestration

internal/policies/
├── service.go          # ⚠️ Needs better integration with PlatformAI orchestration
```

---

## Target Architecture

### Multi-Agent Ecosystem

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Developer     │───▶│  Core Platform  │───▶│  Deployment     │
│   "Deploy my    │    │  Agent          │    │  Agent          │
│    app"         │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │  Governance     │    │  Security       │
                       │  Agent          │    │  Agent          │
                       └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │  Customer       │    │  Event Bus      │
                       │  Custom Agent   │    │  (NATS/Events)  │
                       └─────────────────┘    └─────────────────┘
```

### Agent Responsibilities

#### Core Platform Agent (PlatformAI)
- **Primary Interface**: Receives all developer requests via chat
- **Intent Recognition**: Understands and routes requests to appropriate agents
- **Orchestration**: Coordinates multi-agent workflows
- **Response Synthesis**: Combines agent responses into coherent developer responses

#### Deployment Agent
- **Domain Expertise**: Deployment planning, execution, rollback
- **AI Provider**: Uses AI for intelligent deployment strategies
- **Business Logic**: Owns all deployment-related business rules
- **Event Integration**: Publishes deployment events, subscribes to relevant platform events

#### Governance Agent (Future)
- **Policy Management**: Creation, evaluation, enforcement of policies
- **Compliance Monitoring**: Continuous compliance checking
- **Risk Assessment**: Proactive risk analysis and mitigation

#### Security Agent (Future)
- **Threat Detection**: Real-time security monitoring
- **Access Control**: Dynamic permission management
- **Audit Trail**: Security event correlation and analysis

---

## User Experience Patterns

### Interaction Models

#### 1. Chat-First Interface (Primary)
```
Developer: "Deploy my checkout service to production with zero downtime"

Core Agent: "I'll coordinate a blue-green deployment for checkout service. 
            Let me check with the deployment and governance agents..."

Deployment Agent: "I can execute a blue-green strategy with these steps..."
Governance Agent: "Production deployment approved, all policies satisfied"

Core Agent: "✅ Deployment initiated with 8 steps, estimated 12 minutes. 
            I'll monitor progress and notify you of completion."
```

#### 2. API with AI Backend (Secondary)
```bash
# API call that internally uses AI agents
curl -X POST /v1/deployments \
  -d '{"app": "checkout-service", "environment": "production"}' \
  -H "Content-Type: application/json"

# Behind the scenes: API → Core Agent → Deployment Agent → Execution
```

#### 3. Hybrid Approach (Recommended)
- **Developer Choice**: Use chat OR API based on preference/context
- **Consistent Backend**: Both interfaces use the same AI agent system
- **No Functional Difference**: Chat and API provide identical capabilities

### Conversation Flow Examples

#### Complex Multi-Domain Request
```
Developer: "I need to deploy the new user service, but first check if it complies 
           with our security policies and create the required infrastructure"

Core Agent Analysis:
1. Security policy validation (Security Agent)
2. Infrastructure provisioning (Infrastructure Agent) 
3. Deployment execution (Deployment Agent)
4. Monitoring setup (Observability Agent)

Response: "I've identified a 4-step process involving security, infrastructure, 
          deployment, and monitoring. Proceeding with coordination..."
```

#### Learning and Improvement
```
Developer: "The last deployment failed, can you analyze what went wrong?"

Core Agent: "I'll analyze the deployment failure with the troubleshooting agent..."
Deployment Agent: "Failure occurred at step 3 - database migration timeout"
Core Agent: "Root cause identified. I've updated deployment patterns to extend 
           migration timeouts. Future deployments will include this fix."
```

---

## Multi-Agent System Design

### Agent Architecture Patterns

#### Individual Agent Structure
```go
type Agent interface {
    // Core agent identity and capabilities
    ID() string
    Capabilities() []Capability
    
    // Request processing
    CanHandle(request *Request) bool
    Process(ctx context.Context, request *Request) (*Response, error)
    
    // Event-driven communication
    Subscribe(eventTypes []string) error
    Publish(event *Event) error
    
    // AI integration
    GetAIProvider() AIProvider
    SetContext(context *AgentContext)
}

type DeploymentAgent struct {
    id          string
    service     *deployments.Service  // Domain service
    aiProvider  ai.AIProvider         // AI infrastructure
    eventBus    *events.Bus          // Agent communication
    logger      *logging.Logger
}

func (a *DeploymentAgent) Process(ctx context.Context, req *Request) (*Response, error) {
    // Use domain service + AI provider directly
    plan, err := a.service.GenerateDeploymentPlan(ctx, req.AppName)
    if err != nil {
        return nil, err
    }
    
    // Publish events for other agents
    a.eventBus.Publish(&Event{
        Type: "deployment.plan.generated",
        Data: plan,
    })
    
    return &Response{Plan: plan}, nil
}
```

#### Core Platform Agent (Orchestrator)
```go
type CorePlatformAgent struct {
    agents      map[string]Agent      // Registry of available agents
    eventBus    *events.Bus          // Central communication
    aiProvider  ai.AIProvider         // For conversation and routing
    intentRecognizer *IntentRecognizer
}

func (c *CorePlatformAgent) ProcessConversation(ctx context.Context, query string) (*ConversationalResponse, error) {
    // 1. Understand intent using AI
    intent, err := c.intentRecognizer.Analyze(query)
    
    // 2. Route to appropriate agents
    agents := c.findCapableAgents(intent)
    
    // 3. Coordinate agent execution
    responses := c.executeAgentWorkflow(ctx, intent, agents)
    
    // 4. Synthesize final response
    return c.synthesizeResponse(responses)
}
```

### Agent Discovery and Registration

#### Agent Registry
```go
type AgentRegistry struct {
    agents map[string]Agent
    capabilities map[Capability][]string  // capability -> agent IDs
}

func (r *AgentRegistry) Register(agent Agent) error {
    r.agents[agent.ID()] = agent
    
    // Index by capabilities
    for _, cap := range agent.Capabilities() {
        r.capabilities[cap] = append(r.capabilities[cap], agent.ID())
    }
}

func (r *AgentRegistry) FindByCapability(cap Capability) []Agent {
    agentIDs := r.capabilities[cap]
    agents := make([]Agent, 0, len(agentIDs))
    for _, id := range agentIDs {
        agents = append(agents, r.agents[id])
    }
    return agents
}
```

---

## Implementation Roadmap

### Phase 1: Core Platform Agent Enhancement (Current Priority)

**Goal**: Enhance PlatformAI (AI Brain) as the Core Platform Agent while maintaining clean architecture

**Tasks**:
1. ✅ **Enhance PlatformAI** (`/internal/ai/ai_brain.go`) - Keep revolutionary AI capabilities, refactor architecture
2. **Refactor Business Logic** - Move domain-specific logic to services, keep orchestration in PlatformAI
3. **Complete Domain Service Integration** - Services provide business logic, PlatformAI provides AI orchestration
4. **Enhance API Handlers** - Use PlatformAI for conversation interface, domain services for direct API calls
5. **Clean AI Provider Interface** - Keep only `CallAI()` infrastructure method

**Expected Outcome**: PlatformAI serves as Core Platform Agent with clean domain service integration

### Phase 2: Core Platform Agent

**Goal**: Implement the central orchestrating agent

**Tasks**:
1. **Create Core Platform Agent** (`/internal/agents/core/`)
2. **Implement Intent Recognition** using AI to understand user requests
3. **Build Agent Registry** for discovering and routing to specialized agents
4. **Enhanced Conversation Interface** supporting complex, multi-turn conversations
5. **Event-Driven Orchestration** coordinating multiple agents

**Expected Outcome**: Single conversational interface that can handle complex requests

### Phase 3: Specialized Agents

**Goal**: Break domain services into independent agents

**Tasks**:
1. **Deployment Agent** (`/internal/agents/deployment/`)
2. **Policy/Governance Agent** (`/internal/agents/governance/`)
3. **Security Agent** (`/internal/agents/security/`)
4. **Agent Communication Protocol** via events
5. **Agent Health Monitoring** and failure recovery

**Expected Outcome**: Multi-agent ecosystem with specialized capabilities

### Phase 4: Customer Extensibility

**Goal**: Enable "Bring Your Own Agent" capabilities

**Tasks**:
1. **Agent SDK** for building custom agents
2. **Agent Marketplace** for discovering and installing agents
3. **MCP Integration** (Model Context Protocol) for external AI tools
4. **Agent Security Model** for sandboxing and permissions
5. **Agent Lifecycle Management** (install, update, remove)

**Expected Outcome**: Customers can build and deploy custom agents

---

## Technical Architecture

### Current AI Integration Pattern

#### Correct Pattern (Domain Services)
```go
// Deployment service owns deployment business logic
func (s *DeploymentService) GenerateDeploymentPlan(ctx context.Context, app string) (*Plan, error) {
    // Business logic here
    if s.aiProvider == nil {
        return s.generateBasicPlan(app)
    }
    
    prompt := s.buildDeploymentPrompt(app)
    response, err := s.aiProvider.CallAI(ctx, prompt.System, prompt.User)
    if err != nil {
        return s.generateBasicPlan(app) // Fallback
    }
    
    return s.parseAndValidatePlan(response)
}
```

#### Incorrect Pattern (AI Brain - Being Removed)
```go
// AI Brain incorrectly contains deployment business logic
func (brain *AIBrain) GenerateDeploymentPlan(app string) (*Plan, error) {
    // ❌ Deployment logic in AI layer - WRONG!
    // This violates separation of concerns
}
```

### AI Provider Interface (Infrastructure Only)

```go
type AIProvider interface {
    // Pure infrastructure - only communication with AI services
    CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error)
    
    // Provider metadata
    GetProviderInfo() *ProviderInfo
    Close() error
}

// Clean implementation focused on infrastructure
type OpenAIProvider struct {
    client *openai.Client
    config *OpenAIConfig
}

func (p *OpenAIProvider) CallAI(ctx context.Context, system, user string) (string, error) {
    resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: p.config.Model,
        Messages: []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleSystem, Content: system},
            {Role: openai.ChatMessageRoleUser, Content: user},
        },
    })
    return resp.Choices[0].Message.Content, err
}
```

### Future Agent Interface

```go
type Agent interface {
    // Agent identity
    ID() string
    Name() string
    Description() string
    Version() string
    
    // Capabilities
    Capabilities() []Capability
    CanHandle(request *Request) (bool, float64) // can handle + confidence
    
    // Request processing
    Process(ctx context.Context, request *Request) (*Response, error)
    
    // Event-driven communication
    Subscribe(eventTypes []string, handler EventHandler) error
    Publish(event *Event) error
    
    // Lifecycle
    Initialize(config *AgentConfig) error
    Shutdown() error
    HealthCheck() error
}
```

---

## Domain Separation

### Principles

1. **Domain Services Own Business Logic**: Deployment, Policy, Security services contain all domain-specific logic
2. **AI as Infrastructure Tool**: AI providers are pure infrastructure for communicating with AI services
3. **No Business Logic in AI Layer**: AI components handle only AI communication, not business decisions
4. **Clean Interfaces**: Clear separation between domain logic and AI infrastructure

### Current Cleanup Required

#### Files to Enhance
- `/internal/ai/ai_brain.go` (997 lines) - **Keep as Core Platform Agent**, refactor for clean architecture
- `/internal/deployments/service.go` - Integrate with PlatformAI orchestration
- `/internal/policies/service.go` - Integrate with PlatformAI orchestration
- `/api/handlers/ai.go` - Route conversation through PlatformAI, direct API through services
- `/internal/ai/ai_provider.go` - Keep business methods for PlatformAI orchestration

#### Files to Keep
- `/internal/ai/openai_provider.go` - Pure infrastructure implementation
- `/internal/ai/prompts.go` - Reusable prompt templates (if domain-agnostic)

---

## Event-Driven Agent Communication

### Event Architecture

```go
type Event struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Source    string                 `json:"source"`    // Agent ID
    Target    string                 `json:"target"`    // Agent ID or "*" for broadcast
    Data      map[string]interface{} `json:"data"`
    Timestamp time.Time              `json:"timestamp"`
    Metadata  map[string]string      `json:"metadata"`
}

type EventBus interface {
    Publish(event *Event) error
    Subscribe(agentID string, eventTypes []string, handler EventHandler) error
    Unsubscribe(agentID string, eventTypes []string) error
}
```

### Agent Communication Patterns

#### Request-Response Pattern
```go
// Core Agent requests deployment plan
event := &Event{
    Type: "deployment.plan.request",
    Source: "core-agent",
    Target: "deployment-agent",
    Data: map[string]interface{}{
        "application": "checkout-service",
        "environment": "production",
        "requestID": "req-123",
    },
}

// Deployment Agent responds
responseEvent := &Event{
    Type: "deployment.plan.response",
    Source: "deployment-agent", 
    Target: "core-agent",
    Data: map[string]interface{}{
        "requestID": "req-123",
        "plan": deploymentPlan,
        "confidence": 0.89,
    },
}
```

#### Broadcast Notification Pattern
```go
// Deployment Agent notifies all interested agents
event := &Event{
    Type: "deployment.started",
    Source: "deployment-agent",
    Target: "*", // Broadcast
    Data: map[string]interface{}{
        "application": "checkout-service",
        "environment": "production",
        "deploymentID": "deploy-456",
    },
}
```

### Integration with Existing Event System

ZTDP already has a robust event system (`/internal/events/`). Agent communication will extend this:

```go
// Extend existing event system for agent communication
type AgentEventBus struct {
    *events.Bus  // Existing ZTDP event bus
    agentRegistry *AgentRegistry
}

func (bus *AgentEventBus) PublishToAgent(targetAgent string, event *Event) error {
    // Route directly to specific agent
    agent := bus.agentRegistry.Get(targetAgent)
    return agent.HandleEvent(event)
}
```

---

## Bring Your Own Agent (BYOA)

### Vision

Enable customers to deploy custom agents that integrate seamlessly with the ZTDP platform.

### Agent SDK

```go
package agents

// Customer implements this interface for their custom agent
type CustomerAgent interface {
    Agent // Inherits standard agent interface
    
    // Customer-specific methods
    Configure(config map[string]string) error
    GetRequiredPermissions() []Permission
    GetDependencies() []Dependency
}

// Example customer agent
type CustomSecurityAgent struct {
    customerConfig map[string]string
    permissions    []Permission
}

func (a *CustomSecurityAgent) ID() string { return "customer.security" }
func (a *CustomSecurityAgent) Capabilities() []Capability {
    return []Capability{
        CapabilitySecurityScan,
        CapabilityComplianceCheck,
        CapabilityThreatDetection,
    }
}

func (a *CustomSecurityAgent) Process(ctx context.Context, req *Request) (*Response, error) {
    // Customer's custom security logic
    return a.performSecurityAnalysis(req)
}
```

### Agent Marketplace

```yaml
# agent-manifest.yaml
apiVersion: agents.ztdp.io/v1
kind: Agent
metadata:
  name: advanced-security-agent
  vendor: acme-corp
  version: 1.2.0
spec:
  capabilities:
    - security.scan
    - security.compliance
    - security.threat-detection
  permissions:
    - read:applications
    - read:policies
    - write:security-events
  dependencies:
    - ai-provider: openai
    - event-bus: nats
  configuration:
    - name: api-key
      type: secret
      required: true
    - name: scan-frequency
      type: string
      default: "hourly"
```

### MCP Integration

```go
// Model Context Protocol integration for external AI tools
type MCPAgent struct {
    agentID    string
    mcpClient  *mcp.Client
    connector  *mcp.Connector
}

func (a *MCPAgent) Process(ctx context.Context, req *Request) (*Response, error) {
    // Convert ZTDP request to MCP format
    mcpRequest := a.convertToMCPRequest(req)
    
    // Send to external MCP server
    mcpResponse, err := a.mcpClient.SendRequest(ctx, mcpRequest)
    if err != nil {
        return nil, err
    }
    
    // Convert MCP response back to ZTDP format
    return a.convertFromMCPResponse(mcpResponse)
}
```

---

## Critical Decisions & Trade-offs

### 1. API vs Chat Primary Interface

**Decision**: Hybrid approach where both chat and API use the same AI agent backend

**Rationale**: 
- Developers have different preferences and contexts
- No functional difference between interfaces
- Easier migration path from current API-first approach
- Internal consistency with single agent system

**Alternative Considered**: Chat-only interface - rejected due to migration complexity

### 2. Monolithic vs Microservice Agents

**Decision**: Start with agents as internal modules, evolve to microservices

**Rationale**:
- Lower initial complexity for MVP
- Easier debugging and development
- Clear evolution path to distributed agents
- Allows validation of agent communication patterns

**Alternative Considered**: Microservices from day 1 - rejected due to operational complexity

### 3. Event-Driven vs Direct Agent Communication

**Decision**: Event-driven communication using existing ZTDP event infrastructure

**Rationale**:
- Leverages existing robust event system
- Enables async agent workflows
- Supports broadcast/multicast patterns
- Natural integration with current platform events

**Alternative Considered**: Direct agent-to-agent calls - rejected due to coupling concerns

### 4. PlatformAI Enhancement vs Domain Separation

**Decision**: Enhance PlatformAI as Core Platform Agent while maintaining clean architecture

**Rationale**:
- PlatformAI contains exactly the revolutionary AI capabilities investors want to see
- 997 lines of conversational AI, intent recognition, and platform orchestration
- Core Platform Agent is the centerpiece of our AI-native vision
- Clean architecture achieved through refactoring, not deletion

**Alternative Considered**: Delete AI Brain entirely - rejected as it removes our core AI-native differentiator

---

## Backlog & Next Steps

### Immediate Priority (Sprint 1-2)

1. **🔥 Enhance Core Platform Agent**
   - [ ] Refactor `/internal/ai/ai_brain.go` to separate orchestration from business logic
   - [ ] Complete domain service integration with PlatformAI
   - [ ] Enhance conversation capabilities and intent recognition
   - [ ] Fix compilation errors while preserving AI capabilities

2. **🔥 Validate Current AI Integration**
   - [ ] Ensure all existing AI endpoints work with domain services
   - [ ] Verify graceful fallback when AI unavailable
   - [ ] Test end-to-end deployment with AI planning

### Short Term (Sprint 3-6)

3. **🚀 Core Platform Agent**
   - [ ] Design and implement Core Platform Agent
   - [ ] Build intent recognition using AI
   - [ ] Create agent registry and discovery
   - [ ] Enhanced conversation interface

4. **🎯 Agent Communication**
   - [ ] Extend event system for agent communication
   - [ ] Define agent communication protocols
   - [ ] Implement request-response and broadcast patterns

### Medium Term (Sprint 7-12)

5. **🏗️ Specialized Agents**
   - [ ] Deployment Agent (`/internal/agents/deployment/`)
   - [ ] Policy/Governance Agent (`/internal/agents/governance/`)
   - [ ] Agent health monitoring and recovery

6. **💬 Enhanced UX**
   - [ ] Multi-turn conversation support
   - [ ] Context persistence across conversations
   - [ ] Rich response formatting (markdown, code blocks, etc.)

### Long Term (Sprint 13+)

7. **🔌 Customer Extensibility**
   - [ ] Agent SDK and documentation
   - [ ] Agent marketplace infrastructure
   - [ ] MCP integration for external tools
   - [ ] Agent security and sandboxing

8. **📊 Advanced Features**
   - [ ] Agent performance analytics
   - [ ] A/B testing for AI responses
   - [ ] Learning and improvement loops
   - [ ] Multi-provider AI support

### Success Criteria

#### Phase 1 Success (Domain Separation)
- [ ] All API endpoints functional without AI Brain
- [ ] Clean separation: business logic in domain services, AI as infrastructure
- [ ] Zero compilation errors
- [ ] All existing tests passing

#### Phase 2 Success (Core Agent)
- [ ] Natural language deployment requests working
- [ ] Intent recognition accuracy >90%
- [ ] Multi-agent coordination for complex requests
- [ ] Performance: AI responses <5 seconds

#### Phase 3 Success (Agent Ecosystem)
- [ ] 3+ specialized agents operational
- [ ] Event-driven agent communication working
- [ ] Agent failure recovery implemented
- [ ] Customer can deploy custom agent (proof of concept)

---

## Conclusion

ZTDP's evolution to an AI-native platform represents a fundamental shift in how developers interact with infrastructure. By building a multi-agent ecosystem with clean separation of concerns, we create a platform that is:

- **Conversational**: Natural language as the primary interface
- **Intelligent**: AI-driven decision making across all domains  
- **Extensible**: Customer agents integrate seamlessly
- **Reliable**: Graceful fallbacks and robust error handling
- **Maintainable**: Clean architecture with proper domain separation

This architecture document serves as the definitive guide for implementing ZTDP's AI-native future. Each phase builds on the previous, ensuring we maintain platform stability while evolving toward the multi-agent vision.

**Next Action**: Begin Phase 1 domain separation by completing the work outlined in `/DOMAIN_SEPARATION_PLAN.md`.

---

## Developer Principles & Best Practices

### Core Development Principles

#### 1. Clean Architecture
- **Dependency Direction**: Core business logic should not depend on external layers
- **Interface Segregation**: Use interfaces to break dependencies
- **Single Responsibility**: Each package should have a clear purpose
- **Testability**: Design for easy unit testing and mocking

#### 2. Domain-Driven Design
- **Domain Services Own Business Logic**: Deployment, Policy, Security services contain all domain-specific logic
- **No Business Logic in API Handlers**: API handlers are thin, delegating to domain services
- **AI as Infrastructure Tool**: AI providers are pure infrastructure for communicating with AI services
- **No Business Logic in AI Layer**: AI components handle only AI communication, not business decisions

#### 3. Event-Driven Architecture
- **Consistent Emission**: Ensure all operations emit appropriate events
- **Structured Events**: Follow the standard event schema
- **Error Handling**: Emit events for both success and failure cases
- **Performance**: Consider event volume and processing capacity

#### 4. Policy-First Development
- **Early Validation**: Check policies before making changes
- **Clear Errors**: Provide detailed policy violation messages
- **Event Integration**: Emit events for all policy decisions
- **Test Coverage**: Include policy scenarios in all tests

### Testing Philosophy

#### Test-Driven Development (TDD)
- **Test First**: Every feature starts with a test
- **Red-Green-Refactor**: Follow TDD cycle religiously
- **API-First**: Test the public interface, not implementation details
- **Coverage Requirements**: Maintain high test coverage across all layers

#### Test Organization
- **Unit Tests**: Co-located with source code (`*_test.go`)
- **Integration Tests**: API-level testing (`test/api/`)
- **End-to-End Tests**: Control plane validation (`test/controlplane/`)
- **Policy Tests**: Comprehensive policy enforcement validation

#### Test Patterns
```go
// Unit test pattern - test business logic in isolation
func TestDeploymentService_GenerateDeploymentPlan(t *testing.T) {
    // Arrange
    mockGraph := &mockGraph{}
    mockAI := &mockAIProvider{}
    service := NewDeploymentService(mockGraph, mockAI)
    
    // Act
    plan, err := service.GenerateDeploymentPlan(ctx, "test-app")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, plan)
    // Verify business logic, not AI integration
}

// Integration test pattern - test end-to-end API flow
func TestCreateAndDeployApplication(t *testing.T) {
    router := newTestRouter(t)
    
    // Create application
    createApplication(t, router, "test-app")
    
    // Deploy application
    deployApplication(t, router, "test-app", "dev")
    
    // Verify deployment was successful
    verifyDeploymentStatus(t, router, "test-app", "dev", "deployed")
}
```

### Error Handling Standards

#### HTTP Error Responses
```go
// Policy violations - return 403 Forbidden
if errors.Is(err, &graph.PolicyNotSatisfiedError{}) {
    WriteJSONError(w, "Policy not satisfied", http.StatusForbidden)
    return
}

// Business validation errors - return 400 Bad Request
if errors.Is(err, &ValidationError{}) {
    WriteJSONError(w, err.Error(), http.StatusBadRequest)
    return
}

// System errors - return 500 Internal Server Error
WriteJSONError(w, "Internal server error", http.StatusInternalServerError)
```

#### Event Emission for Errors
```go
// Emit events for both success and failure cases
func (s *DeploymentService) DeployApplication(ctx context.Context, app, env string) error {
    // Emit deployment started event
    s.eventBus.Publish(&Event{
        Type: "deployment.started",
        Subject: fmt.Sprintf("%s/%s", app, env),
    })
    
    err := s.performDeployment(ctx, app, env)
    if err != nil {
        // Emit failure event
        s.eventBus.Publish(&Event{
            Type: "deployment.failed",
            Subject: fmt.Sprintf("%s/%s", app, env),
            Payload: map[string]interface{}{
                "error": err.Error(),
            },
        })
        return err
    }
    
    // Emit success event
    s.eventBus.Publish(&Event{
        Type: "deployment.completed",
        Subject: fmt.Sprintf("%s/%s", app, env),
    })
    return nil
}
```

### Code Organization Standards

#### Package Structure
```
internal/
├── application/     # Application domain service
├── deployment/      # Deployment domain service  
├── policies/        # Policy domain service
├── ai/             # AI infrastructure only
├── graph/          # Graph data model and operations
├── events/         # Event system infrastructure
└── contracts/      # Data contracts and schemas

api/
├── handlers/       # HTTP handlers (thin layer)
└── server/         # Routing and middleware setup

test/
├── api/           # Integration tests
└── controlplane/  # End-to-end tests
```

#### File Naming Conventions
- `service.go` - Main domain service implementation
- `*_test.go` - Unit tests co-located with source
- `contracts.go` - Data structures and interfaces
- `errors.go` - Domain-specific error types

#### Import Organization
```go
// Standard library imports first
import (
    "context"
    "fmt"
    "time"
)

// External library imports second
import (
    "github.com/go-chi/chi/v5"
    "github.com/gorilla/websocket"
)

// Internal imports last
import (
    "github.com/krzachariassen/ZTDP/internal/graph"
    "github.com/krzachariassen/ZTDP/internal/events"
)
```

---

## AI Agent Instructions & Guidelines

### For AI Assistants Working on ZTDP

#### Understanding the Codebase
1. **Read This Document First**: This document is the single source of truth for architecture and development practices
2. **Domain Separation is Critical**: Never put business logic in AI components or API handlers
3. **Event-Driven by Design**: All operations should emit structured events
4. **Policy-Aware**: Consider policy implications for every change

#### When Making Changes

##### 1. Research Phase
```bash
# Use semantic search to understand existing patterns
semantic_search("deployment service AI integration")
semantic_search("policy enforcement patterns")
semantic_search("event emission examples")

# Read related files to understand context
read_file("/internal/deployments/service.go")
read_file("/internal/policies/service.go")
```

##### 2. Implementation Phase
- **Follow TDD**: Write tests before implementation
- **Use Domain Services**: Call domain services, not AI Brain
- **Emit Events**: Ensure operations emit structured events
- **Handle Errors**: Follow established error handling patterns

##### 3. Testing Phase
- **Run All Tests**: `go test ./...` must pass
- **Test Different Scenarios**: Success, failure, policy violations
- **Integration Tests**: Test end-to-end API flows

#### Common Patterns to Follow

##### API Handler Pattern (Thin Layer)
```go
func CreateApplication(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var app contracts.ApplicationContract
    if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
        WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Delegate to domain service
    appService := application.NewService(GlobalGraph)
    if err := appService.CreateApplication(app); err != nil {
        WriteJSONError(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Return response
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(app)
}
```

##### Domain Service Pattern (Business Logic)
```go
func (s *DeploymentService) GenerateDeploymentPlan(ctx context.Context, app string) (*Plan, error) {
    // Business logic validation
    if app == "" {
        return nil, errors.New("application name required")
    }
    
    // Check policies
    if err := s.validateDeploymentPolicies(app); err != nil {
        return nil, err
    }
    
    // Use AI as tool (if available)
    if s.aiProvider != nil {
        prompt := s.buildDeploymentPrompt(app)
        response, err := s.aiProvider.CallAI(ctx, prompt.System, prompt.User)
        if err == nil {
            if plan, err := s.parseAIPlan(response); err == nil {
                return plan, nil
            }
        }
    }
    
    // Fallback to traditional planning
    return s.generateBasicPlan(app)
}
```

##### Event Emission Pattern
```go
// Emit structured events for observability
func (s *DeploymentService) emitDeploymentEvent(eventType, app, env string, payload map[string]interface{}) {
    event := &events.Event{
        Type:    eventType,
        Source:  "deployment-service",
        Subject: fmt.Sprintf("%s/%s", app, env),
        Payload: payload,
    }
    s.eventBus.Publish(event)
}
```

#### What NOT to Do

##### ❌ Business Logic in API Handlers
```go
// WRONG - business logic in handler
func CreateApplication(w http.ResponseWriter, r *http.Request) {
    // Don't do complex validation here
    // Don't call multiple services here
    // Don't handle policy enforcement here
}
```

##### ❌ Business Logic in AI Components
```go
// WRONG - business logic in AI layer
func (brain *AIBrain) GenerateDeploymentPlan(app string) (*Plan, error) {
    // Don't put deployment logic in AI layer
    // Don't validate business rules in AI layer
    // Don't handle policy enforcement in AI layer
}
```

##### ❌ Direct AI Provider Calls from Handlers
```go
// WRONG - skipping domain services
func AIGeneratePlan(w http.ResponseWriter, r *http.Request) {
    // Don't call AI provider directly
    response, err := aiProvider.CallAI(ctx, system, user)
    // Domain logic belongs in domain services
}
```

#### Required Steps for AI-Related Changes

1. **Identify the Domain**: Which business domain does this belong to?
2. **Check Domain Service**: Does the domain service have the required method?
3. **Add to Domain Service**: If missing, add business logic to domain service
4. **Use AI as Tool**: Domain service can use AI provider for enhancement
5. **Update Handler**: Handler calls domain service, not AI directly
6. **Add Tests**: Unit tests for domain service, integration tests for API
7. **Emit Events**: Ensure operations emit structured events

#### File Change Patterns

When modifying AI-related functionality:

##### For New AI-Enhanced Features:
1. Add method to appropriate domain service (`internal/[domain]/service.go`)
2. Use AI provider as infrastructure tool within domain service
3. Update API handler to call domain service
4. Add comprehensive tests
5. Update this architecture document if needed

##### For Refactoring Existing AI Code:
1. Identify business logic currently in AI layer
2. Move business logic to appropriate domain service
3. Update AI component to be pure infrastructure
4. Update callers to use domain service
5. Remove any business logic from AI layer

---

## Coding Standards & Architecture Guidelines

### Go Language Standards

#### Code Style
- **gofmt**: All code must be formatted with `gofmt`
- **goimports**: Organize imports using `goimports`
- **golint**: Follow `golint` recommendations
- **go vet**: Code must pass `go vet` checks

#### Naming Conventions
```go
// Interfaces: -er suffix
type Deployer interface {
    Deploy(ctx context.Context, app string) error
}

// Structs: Clear, descriptive names
type DeploymentService struct {
    graph      graph.Graph
    aiProvider ai.Provider
    eventBus   events.Bus
}

// Methods: Verb + Object
func (s *DeploymentService) GenerateDeploymentPlan(ctx context.Context, app string) (*Plan, error)
func (s *DeploymentService) ValidateDeploymentPolicy(ctx context.Context, app, env string) error

// Constants: ALL_CAPS or CamelCase for exported
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
)
```

#### Error Handling
```go
// Define domain-specific error types
type PolicyNotSatisfiedError struct {
    Policy string
    Reason string
}

func (e *PolicyNotSatisfiedError) Error() string {
    return fmt.Sprintf("policy %s not satisfied: %s", e.Policy, e.Reason)
}

// Use errors.Is for error checking
if errors.Is(err, &PolicyNotSatisfiedError{}) {
    // Handle policy violation
}

// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to generate deployment plan for %s: %w", app, err)
}
```

### Architecture Patterns

#### Dependency Injection
```go
// Constructor pattern for dependency injection
func NewDeploymentService(graph graph.Graph, aiProvider ai.Provider, eventBus events.Bus) *DeploymentService {
    return &DeploymentService{
        graph:      graph,
        aiProvider: aiProvider,
        eventBus:   eventBus,
    }
}

// Interface segregation
type GraphReader interface {
    GetApplication(name string) (*Application, error)
    GetServices(app string) ([]*Service, error)
}

type GraphWriter interface {
    AddNode(node *Node) error
    AddEdge(from, to, relation string) error
}
```

#### Context Usage
```go
// Always use context for cancellation and timeouts
func (s *DeploymentService) GenerateDeploymentPlan(ctx context.Context, app string) (*Plan, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Pass context to dependencies
    response, err := s.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
    if err != nil {
        return nil, err
    }
    
    return s.parsePlan(response)
}
```

#### Interface Design
```go
// Small, focused interfaces
type PlanGenerator interface {
    GeneratePlan(ctx context.Context, app string) (*Plan, error)
}

type PolicyValidator interface {
    ValidateTransition(ctx context.Context, from, to, relation string) error
}

// Compose interfaces when needed
type DeploymentOrchestrator interface {
    PlanGenerator
    PolicyValidator
}
```

### Documentation Standards

#### Code Documentation
```go
// Package documentation
// Package deployment provides deployment planning and execution capabilities.
// It integrates with AI providers for intelligent plan generation while
// maintaining fallback to traditional planning methods.
package deployment

// Function documentation
// GenerateDeploymentPlan creates a deployment plan for the specified application.
// It uses AI-enhanced planning when available, falling back to basic planning
// when AI is unavailable or fails.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - app: Application name to generate plan for
//
// Returns:
//   - *Plan: Generated deployment plan
//   - error: Error if plan generation fails
func (s *DeploymentService) GenerateDeploymentPlan(ctx context.Context, app string) (*Plan, error) {
    // Implementation...
}
```

#### API Documentation (Swagger)
```go
// CreateApplication godoc
// @Summary      Create a new application
// @Description  Creates a new application resource in the platform
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        application  body      contracts.ApplicationContract  true  "Application payload"
// @Success      201  {object}  contracts.ApplicationContract
// @Failure      400  {object}  map[string]string
// @Router       /v1/applications [post]
func CreateApplication(w http.ResponseWriter, r *http.Request) {
    // Implementation...
}
```

### Performance Guidelines

#### Context and Timeouts
```go
// Set reasonable timeouts for AI operations
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := s.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
```

#### Resource Management
```go
// Close resources properly
func (p *OpenAIProvider) Close() error {
    if p.client != nil {
        // Close HTTP client if needed
    }
    return nil
}

// Use defer for cleanup
func (s *DeploymentService) processDeployment(ctx context.Context) error {
    resources, err := s.allocateResources()
    if err != nil {
        return err
    }
    defer s.releaseResources(resources)
    
    // Process deployment...
}
```

#### Event System Performance
```go
// Batch events when possible
func (s *DeploymentService) emitBatchEvents(events []*Event) {
    for _, event := range events {
        s.eventBus.Publish(event)
    }
}

// Consider event volume for high-frequency operations
func (s *DeploymentService) shouldEmitDetailedEvents() bool {
    return s.config.DetailedEvents && s.eventBus.Load() < s.config.MaxEventLoad
}
```

### Security Guidelines

#### Input Validation
```go
// Validate all inputs at service boundaries
func (s *DeploymentService) GenerateDeploymentPlan(ctx context.Context, app string) (*Plan, error) {
    if app == "" {
        return nil, errors.New("application name is required")
    }
    
    if !isValidAppName(app) {
        return nil, errors.New("invalid application name format")
    }
    
    // Continue with business logic...
}
```

#### AI Provider Security
```go
// Sanitize inputs to AI providers
func (s *DeploymentService) buildPrompt(app string) *Prompt {
    // Sanitize application name to prevent prompt injection
    sanitizedApp := sanitizeForAI(app)
    
    return &Prompt{
        System: "You are a deployment planning assistant...",
        User:   fmt.Sprintf("Generate deployment plan for application: %s", sanitizedApp),
    }
}
```

#### Policy Enforcement
```go
// Always check policies before making changes
func (s *DeploymentService) DeployApplication(ctx context.Context, app, env string) error {
    // Policy check before any action
    if err := s.validateDeploymentPolicy(ctx, app, env); err != nil {
        return fmt.Errorf("deployment policy violation: %w", err)
    }
    
    // Proceed with deployment...
}
```

## 12. Domain Separation Requirements

### Current Architecture Problem

The AI components have **absorbed business logic** that belongs to domain services, violating clean architecture principles.

#### Wrong: AI Brain as Business Controller

```
API Handler → AI Brain → AI Provider
            ↑ (Business Logic)
```

#### Correct: Domains Own Business Logic

```
API Handler → Domain Service → AI Provider
            ↑ (Business Logic)   ↑ (Infrastructure)
```

### Domain Responsibility Matrix

| Component | Current Responsibility | Correct Responsibility |
|-----------|----------------------|----------------------|
| **AI Provider** | Infrastructure only | ✅ Infrastructure only |
| **AI Brain** | ❌ Business logic controller | 🗑️ DELETE - Not needed |
| **Deployment Service** | Partial business logic | ✅ All deployment business logic |
| **Policy Service** | Partial business logic | ✅ All policy business logic |

### Required Changes

#### 1. Delete AI Brain

- `/internal/ai/ai_brain.go` - Remove entirely (997 lines of misplaced logic)
- All business methods belong in domain services

#### 2. Update Domain Services

- **Deployment Service**: Call AI Provider directly for planning
- **Policy Service**: Call AI Provider directly for evaluation
- Remove dependencies on AI Brain

#### 3. Update API Handlers

- Call domain services instead of AI Brain
- Domain services will use AI as needed

#### 4. Clean AI Provider Interface

- Keep only infrastructure methods: `CallAI()`
- Remove business methods that were added for AI Brain

### Correct Implementation Pattern

```go
// Deployment domain owns deployment logic
func (s *DeploymentService) PlanDeployment(app string) (*Plan, error) {
    // Business logic here
    prompt := s.buildDeploymentPrompt(app)
    response, err := s.aiProvider.CallAI(ctx, prompt)
    // Parse and validate response
    return s.parseDeploymentPlan(response)
}
```

## 13. AI Provider Architecture Refactoring

### Critical Issue Identified

The `/internal/ai/openai_provider.go` file contains **852 lines** with **business logic mixed into infrastructure code**, violating separation of concerns.

### Current Problems

#### 1. Business Logic in Infrastructure Provider

```go
// WRONG: Business logic in OpenAI provider
func (p *OpenAIProvider) GeneratePlan(ctx context.Context, request *PlanningRequest) (*PlanningResponse, error) {
    // Complex deployment planning business rules
    // Domain-specific validation logic
    // Graph analysis logic
    // Plan optimization logic
}
```

#### 2. Dead Code (~500+ Lines)

- Revolutionary AI methods implemented but unused in API:
  - `ChatWithPlatform()` - Conversational AI (unused)
  - `PredictImpact()` - Impact prediction (unused)
  - `IntelligentTroubleshooting()` - Root cause analysis (unused)
  - `ProactiveOptimization()` - Continuous optimization (unused)

#### 3. Provider Lock-in

- Cannot swap AI providers without reimplementing all business logic
- OpenAI-specific implementation contains generic domain logic

### Solution: Architectural Separation

#### Correct Architecture Pattern

```
┌─────────────────────────────────────┐
│          AI SERVICE                 │
│  (Domain Business Logic)            │
│  ├── GenerateDeploymentPlan()       │
│  ├── EvaluateDeploymentPolicy()     │
│  ├── OptimizeDeploymentPlan()       │
│  └── Domain-agnostic AI logic       │
└─────────────────┬───────────────────┘
                  │ Uses
┌─────────────────▼───────────────────┐
│       AI PROVIDER INTERFACE        │
│  (Infrastructure Contract)          │
│  ├── CallAI(prompts) → response     │
│  ├── GetProviderInfo()              │
│  └── Close()                        │
└─────────────────┬───────────────────┘
                  │ Implements
┌─────────────────▼───────────────────┐
│      OPENAI PROVIDER               │
│  (Pure Infrastructure)             │
│  ├── HTTP communication with API   │
│  ├── JSON parsing/marshaling       │
│  ├── Error handling                │
│  └── ~150 lines max                │
└─────────────────────────────────────┘
```

#### New File Structure

**1. `/internal/ai/ai_service.go` (Business Logic)**

```go
// Domain-agnostic AI business logic
type AIService struct {
    provider AIProvider  // Infrastructure provider
    graph    *graph.GlobalGraph
    // Domain-specific components
}

func (s *AIService) GenerateDeploymentPlan(ctx, request) (*PlanningResponse, error) {
    // 1. Build domain prompts
    // 2. Call provider.CallAI(systemPrompt, userPrompt) 
    // 3. Parse response with domain logic
    // 4. Validate business rules
    // 5. Return domain object
}
```

**2. `/internal/ai/ai_provider.go` (Interface)**

```go
// Infrastructure-only interface
type AIProvider interface {
    CallAI(ctx, systemPrompt, userPrompt string) (string, error)
    GetProviderInfo() *ProviderInfo
    Close() error
}
```

**3. `/internal/ai/openai_provider_clean.go` (Infrastructure)**

```go
// Pure OpenAI HTTP communication (~150 lines)
func (p *OpenAIProvider) CallAI(ctx, systemPrompt, userPrompt string) (string, error) {
    // 1. Build HTTP request
    // 2. Call OpenAI API
    // 3. Parse HTTP response
    // 4. Return raw AI response
}
```

### Benefits of Refactored Architecture

#### 🔧 Separation of Concerns

- **Business Logic**: In `AIService` (domain-agnostic)
- **Infrastructure**: In providers (communication only)
- **Clear boundaries**: No business rules in infrastructure

#### 🔄 Provider Flexibility

- Easy to add Anthropic, Azure OpenAI, local models
- Business logic works with any AI provider
- No code duplication across providers

#### 🧪 Testability

- Mock AI providers for unit tests
- Test business logic independently
- Validate infrastructure separately

## 14. Developer Principles & Best Practices

### Clean Architecture Principles

#### 1. Dependency Direction

- **Rule**: Core business logic should not depend on external layers
- **Implementation**: Domain services use interfaces, not concrete implementations
- **Validation**: Dependencies point inward toward business logic

#### 2. Domain-Driven Design

- **Domain Services Own Business Logic**: Deployment, Policy, Security services contain all domain-specific logic
- **API Handlers Are Thin**: Only handle HTTP concerns, delegate to domain services
- **AI Components Are Infrastructure**: Tools used by domain services, not business logic owners

#### 3. Event-Driven Architecture

- **Consistent Emission**: Ensure all operations emit appropriate events
- **Decoupled Communication**: Services communicate through events, not direct calls
- **Audit Trail**: Events provide complete system state history

#### 4. Policy-First Development

- **Early Validation**: Check policies before making changes
- **Fail Fast**: Policy violations should prevent operations immediately
- **Comprehensive Coverage**: All state changes must respect policies

### Test-Driven Development (TDD)

#### Test First Approach

- **Test First**: Every feature starts with a test
- **Red-Green-Refactor**: Write failing test, make it pass, improve code
- **Comprehensive Coverage**: Unit, integration, and end-to-end tests

#### Test Organization

- **Unit Tests**: Co-located with source code (`*_test.go`)
- **Integration Tests**: `/tests/integration/`
- **End-to-End Tests**: `/tests/e2e/`

#### Test Patterns

```go
func TestDeploymentService_PlanDeployment(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Plan
        wantErr bool
    }{
        {
            name:  "valid application",
            input: "test-app",
            want:  &Plan{...},
        },
        {
            name:    "invalid application",
            input:   "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewDeploymentService(mockProvider, mockGraph)
            got, err := service.PlanDeployment(ctx, tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

This comprehensive guide ensures all development work follows consistent patterns and maintains the architectural integrity of ZTDP's AI-native platform vision.

---

## Development Workflow & Git Practices

### Git Workflow

#### Branch Strategy
- **Main Branch**: Always deployable production code
- **Feature Branches**: `feature/domain-description` (e.g., `feature/deployment-ai-integration`)
- **Hotfix Branches**: `hotfix/issue-description` for critical production fixes
- **Release Branches**: `release/version` for coordinated releases

#### Commit Standards
```bash
# Commit message format
type(scope): description

# Examples
feat(deployment): integrate AI for deployment plan generation
fix(policy): resolve policy validation edge case
docs(architecture): update AI integration guidelines
test(deployment): add comprehensive deployment service tests
refactor(handlers): extract common error handling patterns
```

#### Pull Request Process
1. **Feature Branch**: Create from latest main
2. **Tests First**: Ensure all tests pass locally
3. **Documentation**: Update relevant docs
4. **Code Review**: Minimum one reviewer approval
5. **Integration Tests**: Verify CI/CD pipeline passes
6. **Merge**: Use squash merge for clean history

### Development Environment Setup

#### Prerequisites
```bash
# Install Go 1.21+
go version

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest

# Install testing tools
go install gotest.tools/gotestsum@latest
```

#### Local Development Workflow
```bash
# 1. Setup development environment
make dev-setup

# 2. Run tests during development
make test-watch

# 3. Run linter
make lint

# 4. Generate documentation
make docs

# 5. Run integration tests
make test-integration
```

### Code Review Guidelines

#### What to Look For
1. **Architecture Compliance**: Follows clean architecture principles
2. **Domain Separation**: Business logic in appropriate domain services
3. **Event Emission**: All operations emit structured events
4. **Error Handling**: Consistent error patterns and HTTP responses
5. **Test Coverage**: Comprehensive unit and integration tests
6. **AI Integration**: Proper use of AI providers without business logic mixing
7. **Documentation**: Clear comments and updated API docs

#### Review Checklist
- [ ] Tests cover happy path and error cases
- [ ] Error messages are clear and actionable
- [ ] Events are emitted for all significant operations
- [ ] Domain services contain business logic, not handlers
- [ ] AI calls have proper context and timeouts
- [ ] Policy validation occurs before state changes
- [ ] Documentation is updated for public APIs
- [ ] Code follows Go style guidelines

### Deployment Process

#### Staging Deployment
```bash
# Deploy to staging environment
make deploy-staging

# Run smoke tests
make test-smoke-staging

# Verify AI integrations
make test-ai-integration-staging
```

#### Production Deployment
```bash
# Deploy to production (requires approval)
make deploy-production

# Monitor deployment
make monitor-deployment

# Rollback if needed
make rollback-production
```

### Monitoring & Observability

#### Key Metrics to Monitor
- **API Response Times**: Track handler performance
- **AI Provider Latency**: Monitor AI service response times
- **Event Processing**: Track event emission and processing rates
- **Error Rates**: Monitor error patterns and frequencies
- **Policy Violations**: Track policy enforcement effectiveness

#### Logging Standards
```go
// Use structured logging
log.Info("deployment plan generated",
    "application", app,
    "environment", env,
    "plan_id", plan.ID,
    "generation_time_ms", time.Since(start).Milliseconds(),
)

// Log errors with context
log.Error("AI provider call failed",
    "provider", provider.Name(),
    "error", err,
    "context", ctx.Value("request_id"),
)
```

### Troubleshooting Guide

#### Common Issues

**AI Provider Timeouts**
```bash
# Check AI provider status
curl -H "Authorization: Bearer $API_KEY" https://api.openai.com/v1/models

# Increase timeout in configuration
export AI_TIMEOUT=60s
```

**Event System Backlog**
```bash
# Monitor event queue depth
make monitor-events

# Scale event processors
kubectl scale deployment event-processor --replicas=5
```

**Policy Validation Failures**
```bash
# Debug policy evaluation
go run cmd/debug/policy-validator.go --app=myapp --env=production

# Check policy configuration
make validate-policies
```

#### Debug Commands
```bash
# Run with debug logging
DEBUG=true go run cmd/server/main.go

# Profile memory usage
go tool pprof http://localhost:8080/debug/pprof/heap

# Trace request flow
TRACE=true curl -H "X-Trace-ID: debug-123" http://localhost:8080/api/v1/applications
```

This development workflow ensures consistent, high-quality development practices aligned with ZTDP's AI-native platform architecture while maintaining the clean separation of concerns and robust testing practices essential for a reliable developer platform.
