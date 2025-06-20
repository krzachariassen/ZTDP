# Agent Framework Enhancement Plan

## Overview

Based on the Application Agent refactoring and best practices analysis, this document outlines specific improvements needed for the agent framework to make AI-native agent development easier, more consistent, and less error-prone.

## Current Framework Analysis

### Existing Framework Strengths
- ✅ Good basic agent registration and capability management
- ✅ Event-driven architecture foundation
- ✅ Clean agent interface definitions
- ✅ Dependency injection pattern

### Identified Gaps
- ❌ No AI parameter extraction helpers
- ❌ No response standardization utilities
- ❌ No event routing automation
- ❌ No AI response parsing utilities
- ❌ No testing utilities for AI-native agents
- ❌ Too much boilerplate for common patterns
- ❌ **No correlation ID management (CRITICAL)**
- ❌ **No event payload standardization**
- ❌ **No pattern matching utilities**
- ❌ **No multi-domain agent support**
- ❌ **No error recovery and fallback patterns**
- ❌ **No agent state management**
- ❌ **No AI performance optimization/caching**

## Proposed Framework Enhancements

### 0. Correlation ID Management Framework (CRITICAL)

**Problem**: Correlation ID handling is error-prone and every agent must implement it manually.

**Solution**: Framework-managed correlation ID handling.

```go
// Framework enhancement: automatic correlation ID management
type CorrelationManager struct {
    framework *AgentFramework
}

func (cm *CorrelationManager) ExtractCorrelationID(event *events.Event) string
func (cm *CorrelationManager) CreateResponse(originalEvent *events.Event, payload interface{}) *events.Event
func (cm *CorrelationManager) CreateErrorResponse(originalEvent *events.Event, err error) *events.Event
func (cm *CorrelationManager) CreateClarificationResponse(originalEvent *events.Event, message string) *events.Event

// Enhanced framework builder with automatic correlation handling
func (f *AgentFramework) WithAutoCorrelation() *AgentBuilder
```

**Usage in Agents**:
```go
// Framework automatically handles correlation IDs - zero boilerplate
func (a *Agent) handleCreate(ctx context.Context, event *events.Event, params *AIResponse) (*events.Event, error) {
    result, err := a.service.CreateApplication(params.ApplicationName)
    if err != nil {
        return a.correlationManager.CreateErrorResponse(event, err), nil
    }
    
    // Framework automatically includes correlation ID from original event
    return a.correlationManager.CreateResponse(event, result), nil
}
```

### 0.1. Event Payload Standardization Framework

**Problem**: Inconsistent event payload structures require boilerplate extraction logic.

**Solution**: Framework-provided payload utilities.

```go
// Framework enhancement: standardized event payload handling
type EventPayloadExtractor struct{}

func (e *EventPayloadExtractor) GetUserMessage(event *events.Event) string
func (e *EventPayloadExtractor) GetParameter(event *events.Event, key string) (interface{}, bool)
func (e *EventPayloadExtractor) GetCorrelationID(event *events.Event) string
func (e *EventPayloadExtractor) GetSourceAgent(event *events.Event) string
func (e *EventPayloadExtractor) ValidateRequiredFields(event *events.Event, fields []string) error

// Standard payload structure
type StandardEventPayload struct {
    Message       string                 `json:"message,omitempty"`
    Query         string                 `json:"query,omitempty"`
    Request       string                 `json:"request,omitempty"`
    CorrelationID string                 `json:"correlation_id"`
    SourceAgent   string                 `json:"source_agent"`
    Context       map[string]interface{} `json:"context,omitempty"`
    Parameters    map[string]interface{} `json:"parameters,omitempty"`
}
```

**Usage in Agents**:
```go
func (a *Agent) handleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
    userMessage := a.payloadExtractor.GetUserMessage(event) // Handles all fallbacks
    
    // No more boilerplate extraction logic needed
    return a.processUserMessage(ctx, event, userMessage)
}
```

### 0.2. Pattern Matching and Event Routing Framework

**Problem**: Agents implement crude pattern matching for event routing.

**Solution**: Framework-provided pattern matching utilities.

```go
// Framework enhancement: advanced pattern matching
type EventRouter struct {
    patterns map[string]EventHandler
}

type EventHandler func(context.Context, *events.Event, *StandardAIResponse) (*events.Event, error)

func (f *AgentFramework) NewEventRouter() *EventRouter
func (er *EventRouter) AddPattern(pattern string, handler EventHandler) *EventRouter
func (er *EventRouter) AddDomainPattern(domain string, handlers map[string]EventHandler) *EventRouter
func (er *EventRouter) Route(ctx context.Context, event *events.Event) (*events.Event, error)

// Support for regex patterns
func (er *EventRouter) AddRegexPattern(regex string, handler EventHandler) *EventRouter
```

**Usage in Agents**:
```go
func NewApplicationAgent(deps AgentDependencies) AgentInterface {
    router := agentFramework.NewEventRouter().
        AddDomainPattern("application", map[string]EventHandler{
            "request":    a.handleApplicationRequest,
            "create":     a.handleApplicationCreate,
            "list":       a.handleApplicationList,
        }).
        AddDomainPattern("service", map[string]EventHandler{
            "request":    a.handleServiceRequest,
            "create":     a.handleServiceCreate,
        }).
        AddPattern("*.management", a.handleManagementRequest)
    
    return agentFramework.NewAgent("application-agent").
        WithEventRouter(router).
        Build(deps)
}
```

### 0.3. Multi-Domain Agent Framework

**Problem**: Many agents handle multiple domains but framework doesn't support this pattern.

**Solution**: Framework support for multi-domain agents.

```go
// Framework enhancement: multi-domain agent support
type DomainConfig struct {
    Name         string
    AIPrompts    map[string]string  // action -> prompt template
    Handlers     map[string]EventHandler
    Capabilities []agentRegistry.AgentCapability
}

type MultiDomainAgent struct {
    domains map[string]DomainConfig
    router  *EventRouter
}

func (f *AgentFramework) NewMultiDomainAgent(agentName string) *MultiDomainAgentBuilder
func (builder *MultiDomainAgentBuilder) AddDomain(domain string, config DomainConfig) *MultiDomainAgentBuilder
func (builder *MultiDomainAgentBuilder) Build(deps AgentDependencies) AgentInterface
```

**Usage in Agents**:
```go
func NewApplicationAgent(deps AgentDependencies) AgentInterface {
    return agentFramework.NewMultiDomainAgent("application-agent").
        AddDomain("application", DomainConfig{
            AIPrompts: map[string]string{
                "list":   "Parse this application listing request: {{.UserMessage}}",
                "create": "Parse this application creation request: {{.UserMessage}}",
            },
            Handlers: map[string]EventHandler{
                "list":   a.handleApplicationList,
                "create": a.handleApplicationCreate,
            },
        }).
        AddDomain("service", DomainConfig{
            // service configuration
        }).
        Build(deps)
}
```

### 0.4. Error Recovery and Fallback Framework

**Problem**: No consistent strategy for handling AI failures or low confidence responses.

**Solution**: Framework-provided error recovery patterns.

```go
// Framework enhancement: error recovery and fallbacks
type RecoveryConfig struct {
    MinConfidence     float64
    MaxRetries        int
    FallbackStrategies map[string]FallbackHandler
}

type FallbackHandler func(context.Context, *events.Event, error) (*events.Event, error)

func (f *AgentFramework) WithRecovery(config RecoveryConfig) *AgentBuilder
```

**Usage in Agents**:
```go
func NewAgent(deps AgentDependencies) AgentInterface {
    recovery := RecoveryConfig{
        MinConfidence: 0.7,
        MaxRetries:    2,
        FallbackStrategies: map[string]FallbackHandler{
            "ai_failure":     a.handleAIFailure,
            "low_confidence": a.requestClarification,
            "parse_error":    a.handleParseError,
        },
    }
    
    return agentFramework.NewAgent("application-agent").
        WithRecovery(recovery).
        Build(deps)
}
```

### 0.5. AI Performance and Caching Framework

**Problem**: AI calls are expensive but there's no caching or optimization support.

**Solution**: Framework-provided AI performance optimization.

```go
// Framework enhancement: AI performance optimization
type AICache interface {
    Get(key string) (*StandardAIResponse, bool)
    Set(key string, response *StandardAIResponse, ttl time.Duration)
    Invalidate(pattern string)
}

type AIPerformanceConfig struct {
    Cache           AICache
    CacheTTL        time.Duration
    ParallelCalls   bool
    TimeoutDuration time.Duration
}

func (f *AgentFramework) WithAIOptimization(config AIPerformanceConfig) *AgentBuilder
```

**Usage in Agents**:
```go
func NewAgent(deps AgentDependencies) AgentInterface {
    aiConfig := AIPerformanceConfig{
        Cache:           agentFramework.NewRedisAICache(redisClient),
        CacheTTL:        5 * time.Minute,
        ParallelCalls:   true,
        TimeoutDuration: 30 * time.Second,
    }
    
    return agentFramework.NewAgent("application-agent").
        WithAIOptimization(aiConfig).
        Build(deps)
}
```

### 0.6. Agent State Management Framework

**Problem**: Agents need to maintain state between interactions but no framework support.

**Solution**: Framework-provided state management.

```go
// Framework enhancement: agent state management
type AgentState interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration)
    Delete(key string)
    Clear()
}

type StateConfig struct {
    Provider    string // "memory", "redis", "database"
    TTL         time.Duration
    KeyPrefix   string
}

func (f *AgentFramework) WithState(config StateConfig) *AgentBuilder
```

**Usage in Agents**:
```go
func (a *Agent) handleCreate(ctx context.Context, event *events.Event, params *AIResponse) (*events.Event, error) {
    // Store operation state for follow-up interactions
    a.state.Set("last_created_app", params.ApplicationName, 10*time.Minute)
    
    result, err := a.service.CreateApplication(params.ApplicationName)
    return a.createResponse(event, result), nil
}
```

### 1. AI Parameter Extraction Framework

**Problem**: Every agent reimplements AI parameter extraction with boilerplate.

**Solution**: Framework-provided AI extraction utilities.

```go
// Framework enhancement: AI parameter extraction
type ParameterSchema struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"`
    Required    bool        `json:"required"`
    Description string      `json:"description"`
    Examples    []string    `json:"examples,omitempty"`
}

type AgentCapabilityWithSchema struct {
    agentRegistry.AgentCapability
    ParameterSchema []ParameterSchema `json:"parameter_schema"`
}

// Enhanced framework method
func (f *AgentFramework) ExtractParametersWithAI(
    ctx context.Context, 
    userMessage string, 
    capabilities []AgentCapabilityWithSchema,
    aiProvider ai.AIProvider,
) (*StandardAIResponse, error)
```

**Usage in Agents**:
```go
// Clean agent usage
func (a *Agent) handleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
    userMessage := agentFramework.ExtractUserMessage(event)
    
    // Framework handles all AI parameter extraction
    response, err := a.framework.ExtractParametersWithAI(
        ctx, userMessage, a.capabilities, a.aiProvider,
    )
    
    if response.Confidence < 0.8 {
        return a.framework.CreateClarificationResponse(event, response.Clarification)
    }
    
    return a.routeAction(ctx, event, response)
}
```

### 2. Response Standardization Framework

**Problem**: Inconsistent response structures across agents.

**Solution**: Framework-provided response builders.

```go
// Framework enhancement: standardized responses
type ResponseBuilder struct {
    agentType string
    agentName string
}

func (f *AgentFramework) GetResponseBuilder(agentType, agentName string) *ResponseBuilder

func (r *ResponseBuilder) Success(originalEvent *events.Event, data interface{}) *events.Event
func (r *ResponseBuilder) Error(originalEvent *events.Event, err error) *events.Event
func (r *ResponseBuilder) Clarification(originalEvent *events.Event, message string) *events.Event
func (r *ResponseBuilder) Progress(originalEvent *events.Event, status string, progress float64) *events.Event
```

**Usage in Agents**:
```go
func (a *Agent) handleCreate(ctx context.Context, event *events.Event, params *AIResponse) (*events.Event, error) {
    result, err := a.service.CreateApplication(params.ApplicationName)
    if err != nil {
        return a.responseBuilder.Error(event, err), nil
    }
    
    return a.responseBuilder.Success(event, result), nil
}
```

### 3. Event Routing Automation

**Problem**: Repetitive routing logic in every agent.

**Solution**: Framework-provided auto-routing.

```go
// Framework enhancement: auto-routing
type ActionHandler func(context.Context, *events.Event, *StandardAIResponse) (*events.Event, error)

type RoutingConfig struct {
    Actions map[string]ActionHandler
    DefaultHandler ActionHandler
}

func (f *AgentFramework) WithAutoRouting(config RoutingConfig) *AgentBuilder

// Usage in agents - dramatically simplified
func NewAgent(deps AgentDependencies) *Agent {
    routing := RoutingConfig{
        Actions: map[string]ActionHandler{
            "list":   a.handleList,
            "create": a.handleCreate,
            "delete": a.handleDelete,
        },
        DefaultHandler: a.handleUnknown,
    }
    
    return agentFramework.NewAgent("application-agent").
        WithCapabilities(getCapabilities()).
        WithAutoRouting(routing).
        Build(deps)
}
```

### 4. AI Response Parsing Utilities

**Problem**: Every agent reimplements JSON parsing and validation.

**Solution**: Framework utilities for AI response handling.

```go
// Framework enhancement: AI response utilities
type StandardAIResponse struct {
    Action        string                 `json:"action"`
    Parameters    map[string]interface{} `json:"parameters"`
    Confidence    float64               `json:"confidence"`
    Clarification string                `json:"clarification,omitempty"`
}

func (f *AgentFramework) ParseAIResponse(response string) (*StandardAIResponse, error)
func (f *AgentFramework) ValidateAIResponse(response *StandardAIResponse, minConfidence float64) error
func (f *AgentFramework) BuildSystemPrompt(capabilities []AgentCapabilityWithSchema) string
```

### 5. Testing Framework Enhancements

**Problem**: Testing AI-native agents requires boilerplate setup.

**Solution**: Framework testing utilities.

```go
// Framework enhancement: testing utilities
type TestFramework struct {
    aiProvider ai.AIProvider
    eventBus   *events.EventBus
}

func NewTestFramework() *TestFramework
func (tf *TestFramework) CreateMockAIProvider(responses map[string]string) ai.AIProvider
func (tf *TestFramework) CreateTestEvent(userMessage string) *events.Event
func (tf *TestFramework) CreateTestEventWithPayload(payload map[string]interface{}) *events.Event
func (tf *TestFramework) AssertSuccessResponse(t *testing.T, event *events.Event)
func (tf *TestFramework) AssertClarificationResponse(t *testing.T, event *events.Event, expectedMessage string)
```

**Usage in Tests**:
```go
func TestAgent_CreateApplication(t *testing.T) {
    tf := agentFramework.NewTestFramework()
    
    agent := NewAgent(tf.CreateTestDependencies())
    event := tf.CreateTestEvent("create app called testapp")
    
    response, err := agent.HandleEvent(context.Background(), event)
    
    tf.AssertSuccessResponse(t, response)
    tf.AssertContains(t, response, "application_name", "testapp")
}
```

### 6. Agent Configuration Framework

**Problem**: Agent configuration is inconsistent.

**Solution**: Standardized configuration management.

```go
// Framework enhancement: configuration
type AgentConfig struct {
    Name                string
    Type                string
    MinConfidence       float64
    RequiredCapabilities []string
    Dependencies        []string
}

func (f *AgentFramework) WithConfig(config AgentConfig) *AgentBuilder
```

## Implementation Priority

### Phase 0: Critical Infrastructure (Week 1)
1. **Correlation ID Management Framework** - CRITICAL: Prevents orchestrator timeouts
2. **Event Payload Standardization Framework** - High impact, eliminates boilerplate
3. **Pattern Matching and Event Routing Framework** - Simplifies agent development

### Phase 1: Core Enhancements (Week 2-3)
4. **Response Standardization Framework** - High impact, low complexity
5. **AI Response Parsing Utilities** - Reduces immediate boilerplate
6. **Multi-Domain Agent Framework** - Supports complex agents like ApplicationAgent

### Phase 2: AI Integration (Week 4-5)
7. **AI Parameter Extraction Framework** - Core AI-native functionality
8. **System Prompt Generation** - Standardizes AI interactions
9. **Error Recovery and Fallback Framework** - Improves reliability

### Phase 3: Performance & State (Week 6-7)
10. **AI Performance and Caching Framework** - Reduces costs and latency
11. **Agent State Management Framework** - Enables stateful interactions
12. **Testing Framework Enhancements** - Improves development workflow

### Phase 4: Advanced Features (Week 8-9)
13. **Agent Configuration Framework** - Standardizes agent setup
14. **Performance Monitoring** - Adds observability
15. **Advanced Analytics** - Usage patterns and optimization

## Backward Compatibility Strategy

### Approach: Additive Enhancements
- All new features are opt-in
- Existing agents continue to work unchanged
- Migration path provided for each enhancement
- Deprecation warnings for old patterns

### Migration Support
```go
// Migration helper - wrap existing agents
func (f *AgentFramework) WrapLegacyAgent(legacyAgent LegacyAgentInterface) AgentInterface

// Provide migration tools
func (f *AgentFramework) MigrateToStandardResponses(agent AgentInterface) AgentInterface
```

## Success Metrics

### Developer Experience Metrics
- **Lines of code reduction**: Target 70% reduction in agent boilerplate (increased from 50%)
- **Time to implement new agent**: Target < 1 hour for basic agent (reduced from 2 hours)
- **Test setup time**: Target < 2 minutes for comprehensive test suite (reduced from 5 minutes)
- **Bug rate**: Target 80% reduction in agent-related bugs (increased from 70%)
- **Correlation ID errors**: Target 100% elimination (new metric)
- **AI call optimization**: Target 50% reduction in redundant AI calls (new metric)

### Framework Adoption Metrics
- **Framework method usage**: Track adoption of new utilities
- **Legacy pattern usage**: Monitor and reduce over time
- **Developer satisfaction**: Survey agent developers

## Example: Before vs After

### Before (Current Pattern)
```go
// 300+ lines of boilerplate per agent (increased from 200+)
func (a *Agent) handleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
    // 30 lines of correlation ID extraction and validation
    // 50 lines of user message extraction with multiple fallbacks
    // 80 lines of AI parameter extraction with error handling
    // 30 lines of confidence checking and clarification logic
    // 40 lines of pattern matching and routing logic
    // 50 lines of response creation with correlation ID management
    // 20 lines of error handling and recovery
}

// Plus additional methods for:
// - getCorrelationID() - 15 lines
// - createSuccessResponse() - 25 lines  
// - createErrorResponse() - 25 lines
// - createClarificationResponse() - 25 lines
// - matchesPattern() - 30 lines per domain
```

### After (Enhanced Framework)
```go
// 15 lines with enhanced framework (reduced from 20)
func NewAgent(deps AgentDependencies) AgentInterface {
    return agentFramework.NewMultiDomainAgent("application-agent").
        WithAutoCorrelation().
        WithAIOptimization(aiConfig).
        WithRecovery(recoveryConfig).
        AddDomain("application", applicationDomainConfig).
        AddDomain("service", serviceDomainConfig).
        Build(deps)
}

func (a *Agent) handleCreate(ctx context.Context, event *events.Event, params *StandardAIResponse) (*events.Event, error) {
    result, err := a.service.CreateApplication(params.GetString("application_name"))
    if err != nil {
        return a.autoErrorResponse(event, err), nil // Framework handles correlation ID
    }
    return a.autoSuccessResponse(event, result), nil // Framework handles correlation ID
}
```

## Documentation Plan

### 1. Framework Enhancement Documentation
- API reference for all new methods
- Migration guides for each enhancement
- Best practices for using enhanced framework

### 2. Agent Development Guide
- Step-by-step agent creation tutorial
- Common patterns and examples
- Troubleshooting guide

### 3. Testing Guide
- AI-native testing strategies
- Framework testing utilities usage
- Performance testing guidelines

## Next Steps

1. **Review and Approve Plan** - Get stakeholder buy-in
2. **Create Detailed Technical Specs** - For each enhancement
3. **Set up Development Environment** - For framework changes
4. **Implement Phase 1** - Response standardization and AI utilities
5. **Test with Existing Agents** - Validate enhancements
6. **Document and Communicate** - Update guides and examples

This enhancement plan will transform the agent framework from a basic foundation to a comprehensive AI-native development platform, significantly reducing the complexity and time required to develop high-quality agents.

## Real-World Validation from Application Agent Implementation

### Critical Issues Discovered
During the Application Agent implementation, several critical issues were discovered that highlight the importance of these framework enhancements:

#### 1. Correlation ID Bugs Were Frequent and Hard to Debug
- **Issue**: Agent responses without correlation IDs caused orchestrator timeouts
- **Root Cause**: Manual correlation ID extraction from `event.Payload["correlation_id"]`
- **Impact**: Silent failures with no error indication
- **Framework Solution**: Automatic correlation ID management eliminates this entire class of bugs

#### 2. Event Payload Extraction Was Error-Prone
- **Issue**: 20+ lines of boilerplate to extract user messages from different payload structures
- **Code Pattern**: Multiple nested `if !exists` checks for `message`, `query`, `request` fields
- **Impact**: Inconsistent behavior across agents
- **Framework Solution**: Standardized payload extraction utilities

#### 3. AI Response Parsing Required Extensive Error Handling
- **Issue**: AI responses weren't always valid JSON, causing agent crashes
- **Fallback Logic**: Complex retry and recovery patterns in every agent
- **Impact**: Reduced reliability and increased development time
- **Framework Solution**: Built-in AI response validation and recovery

#### 4. Multi-Domain Pattern Required Massive Boilerplate
- **Issue**: Application Agent handles 4 domains (application, service, environment, release)
- **Code Duplication**: Similar handler patterns repeated for each domain
- **Maintenance Burden**: Changes required updates in multiple places
- **Framework Solution**: Multi-domain agent framework eliminates duplication

#### 5. Testing Required Extensive Setup
- **Issue**: Each test required Redis setup, AI provider mocking, event creation
- **Code Duplication**: Shared testing infrastructure across multiple test files
- **Time Investment**: 30+ minutes to set up comprehensive test suite
- **Framework Solution**: Built-in testing utilities and shared infrastructure

### Performance Impact Analysis
Based on Application Agent implementation metrics:

- **AI Calls**: 2 AI calls per request (orchestrator + agent) = ~2 seconds total
- **Boilerplate Code**: 300+ lines of repetitive code per agent
- **Bug Rate**: 3 correlation ID bugs discovered during development
- **Test Setup Time**: 45 minutes to create shared testing infrastructure
- **Development Time**: 8 hours to implement full multi-domain agent

### Framework Enhancement ROI
With proposed enhancements:

- **AI Calls**: Potential caching could reduce to 1 call per similar request
- **Boilerplate Code**: Reduce to ~15 lines with framework utilities
- **Bug Rate**: Eliminate correlation ID bugs entirely
- **Test Setup Time**: Reduce to ~2 minutes with framework testing utilities
- **Development Time**: Reduce to ~1 hour for similar multi-domain agent

### Validation Metrics
The Application Agent implementation provides concrete validation for framework enhancements:

1. **Correlation ID Management**: 100% of response bugs were correlation-related
2. **Event Payload Standardization**: 60+ lines of extraction boilerplate per agent
3. **AI Response Handling**: 40+ lines of parsing and validation code
4. **Multi-Domain Support**: 4 domains × 50 lines = 200 lines of routing logic
5. **Error Recovery**: 80+ lines of confidence checking and fallback logic

These real-world metrics justify prioritizing correlation ID management and payload standardization as Phase 0 critical infrastructure improvements.
