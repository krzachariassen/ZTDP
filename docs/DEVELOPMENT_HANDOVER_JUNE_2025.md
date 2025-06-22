# ZTDP AI-Native Platform Development Handover Document

**Date**: June 20, 2025  
**Status**: Clean Architecture Refactoring Required  
**Current Phase**: Domain Separation and TDD Implementation  
**Chat Session**: Architecture Cleanup and Testing Validation

---

## 🎯 **Project Overview**

ZTDP is transitioning from an API-first platform to an **AI-native platform** where artificial intelligence is the primary interface for developer interactions. The platform follows clean architecture principles with event-driven communication between specialized AI agents.

### **Core Vision**
- **AI-Native Interface**: Natural language is the primary way users interact with the platform
- **Multi-Agent System**: Specialized agents handle different domains (Application, Service, Environment, Release, Deployment, Policy)
- **Event-Driven Architecture**: Agents communicate via standardized events with correlation IDs
- **Clean Architecture**: Domain logic separated from infrastructure, with AI as a tool

---

## 🏗️ **Current Architecture State**

### **Working Components** ✅
1. **Orchestrator**: Routes natural language requests to appropriate agents
2. **Application Agent**: Handles application lifecycle (create, list, delete) - FULLY WORKING
3. **Deployment Agent**: Handles deployments - FULLY WORKING  
4. **Policy Agent**: Policy validation - WORKING
5. **Event System**: In-memory event bus with correlation ID support
6. **Graph Storage**: Redis-backed persistence for all platform state
7. **AI Integration**: OpenAI API integration for intent detection and parameter extraction

### **Architecture Issues** ❌
1. **Application Agent Violating Clean Architecture**: Contains logic for 4 domains (Application, Service, Environment, Release)
2. **Mixed Responsibilities**: Single agent handling multiple bounded contexts
3. **Inconsistent Event Payloads**: Multiple field names for same data (`message`, `query`, `request`)
4. **Business Logic in Presentation Layer**: AI extraction logic in agent instead of domain services

---

## 🚨 **Critical Issues Identified**

### **Issue 1: Architecture Violation**
**Current State**: Application Agent handles 4 domains
```go
// ❌ WRONG: Single agent handling multiple domains
func (a *ApplicationAgent) handleEvent(ctx context.Context, event *events.Event) {
    switch {
    case matchesPattern(event.Subject, "application.*"):
        return a.handleApplicationEvent(ctx, event)
    case matchesPattern(event.Subject, "service.*"):
        return a.handleServiceEvent(ctx, event)  // Violates SRP
    case matchesPattern(event.Subject, "environment.*"):
        return a.handleEnvironmentEvent(ctx, event)  // Violates SRP
    case matchesPattern(event.Subject, "release.*"):
        return a.handleReleaseEvent(ctx, event)  // Violates SRP
    }
}
```

**Required Solution**: Separate agents per domain following DDD bounded contexts

### **Issue 2: Test Suite False Positives** 
**Problem**: Tests pass with HTTP 200 but operations fail
```bash
# Test reports SUCCESS but actual operation fails
✅ ✓ AI Chat - Create dev environment
AI Response: ❌ Failed to create environment: environment name is required
```

**Root Cause**: Tests only check HTTP status, not actual business operation success

### **Issue 3: Inconsistent Event Payloads**
**Current Bad Pattern**:
```go
// ❌ Multiple fallbacks for same data
userMessage, exists := event.Payload["message"].(string)
if !exists {
    userMessage, exists = event.Payload["query"].(string)
    if !exists {
        userMessage, exists = event.Payload["request"].(string)
    }
}
```

**Required**: Standardized payload structure

### **Issue 4: Service/Environment/Release Not AI-Native**
**Problem**: These domains fail on natural language input
```bash
# These fail with "parameter required" errors:
"Create a development environment called dev owned by platform-team for development work"
"Create a service called checkout-api for the checkout application on port 8080 that is public facing"
```

**Root Cause**: AI parameter extraction not implemented for these domains

---

## 🎯 **Required Architectural Changes**

### **1. Domain Separation (CRITICAL)**

**From**: Single Application Agent handling 4 domains  
**To**: Separate agents per bounded context

```
internal/
├── application/
│   ├── application_agent.go    # Only handles application.*
│   ├── application.go          # All application domain logic
│   └── application_test.go
├── service/
│   ├── service_agent.go        # Only handles service.*
│   ├── service.go              # All service domain logic  
│   └── service_test.go
├── environment/
│   ├── environment_agent.go    # Only handles environment.*
│   ├── environment.go          # All environment domain logic
│   └── environment_test.go
└── release/
    ├── release_agent.go        # Only handles release.*
    ├── release.go              # All release domain logic
    └── release_test.go
```

**Note**: `/internal/release/` already exists with separate agent - this is the correct pattern!

### **2. Clean Domain Services**

Each domain service should own:
- AI parameter extraction for their domain
- Business logic and validation  
- Data persistence
- Error handling

```go
// ✅ CORRECT: Domain service owns everything
// service.go
func (s *ServiceDomainService) HandleEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
    userMessage := event.Payload["user_message"].(string)  // Standardized
    
    // Service domain owns AI extraction
    params, err := s.ExtractServiceParameters(ctx, userMessage)
    
    // Service domain owns business logic
    return s.HandleServiceAction(ctx, event, params)
}

func (s *ServiceDomainService) ExtractServiceParameters(ctx context.Context, userMessage string) (*ServiceParams, error) {
    // Service-specific AI prompts: "Create a service called checkout-api for the checkout application on port 8080 that is public facing"
    // Extract: service_name="checkout-api", application="checkout", port=8080, public=true
}
```

### **3. Standardized Event Payload**

```go
// ✅ ALWAYS use these field names - no fallbacks
type StandardEventPayload struct {
    UserMessage   string `json:"user_message"`   // ALWAYS this field
    CorrelationID string `json:"correlation_id"` // ALWAYS this field
    SourceAgent   string `json:"source_agent"`   // ALWAYS this field
    RequestID     string `json:"request_id"`     // ALWAYS this field
}
```

---

## 🧪 **Testing Strategy Required**

### **Current Testing State**
- **Application Domain**: Full test coverage with `testing_helpers.go` ✅
- **Service/Environment/Release**: No AI-native tests ❌
- **Integration Tests**: False positives (HTTP 200 but business logic fails) ❌

### **Required Test Improvements**

1. **Enhanced Test Validation**:
```bash
# ❌ Current: Only checks HTTP 200
test_ai_chat "Create dev environment" "Create dev environment"

# ✅ Required: Check actual persistence
test_ai_chat_with_validation "Create dev environment" "create_env" "dev"
```

2. **Domain-Specific TDD**:
```go
// Example: service.go TDD
func TestServiceDomain_CreateServiceFromNaturalLanguage(t *testing.T) {
    // Test: "Create a service called checkout-api for checkout application on port 8080 that is public facing"
    // Expected: service_name="checkout-api", application="checkout", port=8080, public=true
}
```

---

## 🔧 **Implementation Priority**

### **Phase 1: Critical Architecture Fix**
1. **Extract Service Agent** from Application Agent
2. **Extract Environment Agent** from Application Agent  
3. **Standardize Event Payloads** across all agents
4. **Fix Test Validation** to check actual business results

### **Phase 2: AI-Native Domain Services**
1. **Service Domain**: Handle "Create a service called checkout-api for checkout application on port 8080 that is public facing"
2. **Environment Domain**: Handle "Create a development environment called dev owned by platform-team"
3. **Release Domain**: Already working but needs AI enhancement

### **Phase 3: Framework Enhancements**
Implement the enhancements documented in `/docs/AGENT_FRAMEWORK_ENHANCEMENT_PLAN.md`:
- Correlation ID management framework
- Event payload standardization framework
- Multi-domain agent framework

---

## 📁 **Key Files and Locations**

### **Working Examples** ✅
- `/internal/application/application_agent.go` - Working AI-native agent (APPLICATION DOMAIN ONLY)
- `/internal/application/application.go` - Working domain service with persistence
- `/internal/application/testing_helpers.go` - Shared test infrastructure
- `/internal/release/agent.go` - Separate agent example (but needs AI enhancement)

### **Architecture Documentation** 📚
- `/docs/AGENT_FRAMEWORK_ENHANCEMENT_PLAN.md` - Framework improvements needed
- `/docs/ai-platform-architecture.md` - Overall platform vision
- `/docs/clean-architecture-principles.md` - Architecture principles

### **Test Infrastructure** 🧪
- `/test/run_comprehensive_tests.sh` - End-to-end test suite (needs validation fixes)
- `/internal/application/*_test.go` - Working domain tests

### **Current Issues** ❌
- `/internal/application/application_agent.go` lines 480-700 - Contains service/environment/release logic (WRONG)
- `/test/run_comprehensive_tests.sh` lines 100-200 - False positive test validation

---

## 🎯 **Immediate Next Steps**

1. **Start with TDD**: Write failing test for service creation from natural language
2. **Extract Service Agent**: Move service logic from Application Agent to separate Service Agent
3. **Standardize Payloads**: Fix all agents to use `user_message` field consistently
4. **Fix Test Validation**: Make tests check actual persistence, not just HTTP 200

### **Example TDD Start**:
```go
// service_test.go
func TestServiceAgent_CreateServiceFromNaturalLanguage(t *testing.T) {
    // RED: Write failing test first
    agent := createTestServiceAgent()
    event := createTestEvent("Create a service called checkout-api for checkout application on port 8080 that is public facing")
    
    response, err := agent.HandleEvent(context.Background(), event)
    
    assert.NoError(t, err)
    assert.Equal(t, "success", response.Payload["status"])
    
    // Verify actual persistence
    service := getServiceFromGraph("checkout-api")
    assert.Equal(t, "checkout", service.Application)
    assert.Equal(t, 8080, service.Port)
    assert.True(t, service.Public)
}
```

---

## 🚀 **Success Criteria**

The refactoring is complete when:

1. **Each domain has its own agent** (Application, Service, Environment, Release)
2. **All agents handle natural language input** using AI parameter extraction
3. **Tests validate actual business operations**, not just HTTP status
4. **Event payloads are standardized** across all components
5. **Comprehensive test**: `"Create a service called checkout-api for the checkout application on port 8080 that is public facing"` works end-to-end

---

## 📊 **Platform Test Results (Before Refactoring)**

### **Working Tests** ✅
```bash
✅ ✓ AI Chat - Create checkout application  
✅ ✓ AI Chat - list applications
✅ ✓ AI Chat - Deploy checkout to dev (after orchestrator fix)
```

### **Failing Tests** ❌
```bash
❌ Create dev environment - "environment name is required"
❌ Create checkout-api service - "service name is required"  
❌ Policy validation - "policy evaluation requires node, edge, graph, or user_message"
```

### **False Positive Pattern**
```bash
✅ ✓ AI Chat - Create dev environment  # HTTP 200 (FALSE POSITIVE)
AI Response: ❌ Failed to create environment: environment name is required
```

---

## 🔍 **Key Debugging Discovery**

### **Orchestrator Fix Applied**
Fixed orchestrator routing to extract `user_message` from context and put it directly in event payload:
```go
// Fixed: Extract user_message from context and add to top level
if userMessage, exists := context["user_message"].(string); exists {
    eventPayload["user_message"] = userMessage
}
```

### **Application Agent Working Pattern**
```go
// ✅ Working pattern in Application Agent
func (a *ApplicationAgent) extractIntentAndParameters(ctx context.Context, userMessage, domainType string) (*AIResponse, error) {
    // AI system prompts for each domain
    // JSON response parsing with confidence checking
    // Proper error handling and fallbacks
}
```

### **Release Agent Pattern** 
```go
// ❌ Current Release Agent - needs AI enhancement
func (a *ReleaseAgent) ProcessEvent(ctx context.Context, event *events.Event) (*events.Event, error) {
    // Expects structured fields: intent, application, service_versions
    // Doesn't use AI to extract from natural language
}
```

---

## 🎯 **Architectural Decision Points**

### **1. Agent Organization**
**Decision**: Use separate agents per domain (not multi-domain agent)
**Rationale**: 
- `/internal/release/` already exists as separate agent
- Better aligns with DDD bounded contexts
- Easier to test and maintain individual domains
- Allows independent deployment and scaling

### **2. Domain Service Responsibility**
**Decision**: Domain services own ALL domain logic including AI extraction
**Rationale**:
- Follows clean architecture principles
- Agent becomes thin presentation layer
- Domain service can be tested independently
- AI prompts are domain-specific

### **3. Event Payload Standardization**
**Decision**: Always use `user_message` field - no fallbacks
**Rationale**:
- We control all components
- Eliminates boilerplate code
- Makes debugging easier
- Cleaner agent code

---

## 📋 **Framework Enhancement Status**

### **Critical Missing (Priority 0)**
- ❌ Correlation ID management framework
- ❌ Event payload standardization framework  
- ❌ Pattern matching and routing framework

### **Important Missing (Priority 1)**
- ❌ Multi-domain agent support framework
- ❌ Error recovery and fallback framework
- ❌ AI performance and caching framework

### **Enhancement Document**
See `/docs/AGENT_FRAMEWORK_ENHANCEMENT_PLAN.md` for complete framework enhancement plan with real-world validation metrics.

---

## 🎉 **Development Session Achievements**

### **Completed** ✅
1. **Application Agent Refactoring**: Centralized test infrastructure, working AI-native creation
2. **Orchestrator Fix**: Proper user_message passing to agents
3. **Deployment Integration**: End-to-end deployment working with AI
4. **Correlation ID Fix**: Orchestrator/agent communication working
5. **Graph Persistence**: Applications properly saved and retrievable
6. **Comprehensive Testing**: 29/29 tests passing (with false positives identified)

### **Validated Working Flow** ✅
```
User: "Create app called testapp" 
→ Orchestrator (AI intent extraction)
→ Application Agent (AI parameter extraction) 
→ Application Service (business logic)
→ Graph persistence (Redis)
→ Success response with correlation ID
```

---

## 🚨 **Critical Refactoring Required**

**The Application Agent must be split into 4 separate domain agents to follow clean architecture principles.**

This is not optional - it's a fundamental architectural requirement for:
- Clean domain separation
- Testability
- Maintainability  
- Scalability
- True DDD implementation

---

**End of Handover Document**
**Next Developer: Start with TDD for Service Agent creation from natural language**
