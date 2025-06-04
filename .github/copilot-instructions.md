# GitHub Copilot Instructions for ZTDP

## Project Overview

ZTDP is transitioning from an API-first platform to an **AI-native platform** where artificial intelligence is the primary interface for developer interactions. This project follows clean architecture principles with strict domain separation.

## Core Architecture Principles

### 1. Clean Architecture & Domain Separation

**CRITICAL**: Business logic belongs in domain services, NOT in API handlers or AI components.

```go
// ✅ CORRECT: Domain service owns business logic
func (s *DeploymentService) PlanDeployment(ctx context.Context, app string) (*Plan, error) {
    // Business validation here
    if err := s.validateApplication(app); err != nil {
        return nil, err
    }
    
    // Use AI as infrastructure tool
    if s.aiProvider != nil {
        response, err := s.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
        if err == nil {
            return s.parseAndValidatePlan(response)
        }
    }
    
    // Fallback to traditional planning
    return s.generateBasicPlan(app)
}

// ❌ WRONG: Business logic in API handler
func (h *Handler) PlanDeployment(w http.ResponseWriter, r *http.Request) {
    // Don't put validation, AI calls, or business logic here
}

// ❌ WRONG: Business logic in AI layer
func (brain *AIBrain) GenerateDeploymentPlan(app string) (*Plan, error) {
    // Don't put deployment logic in AI components
}
```

### 2. AI as Infrastructure Tool

AI providers are pure infrastructure - they only handle communication with AI services.

```go
// ✅ CORRECT: AI provider interface (infrastructure only)
type AIProvider interface {
    CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error)
    GetProviderInfo() *ProviderInfo
    Close() error
}

// ❌ WRONG: Business methods in AI provider
type BadAIProvider interface {
    CallAI(ctx context.Context, prompt string) (string, error)
    GeneratePlan(ctx context.Context, app string) (*Plan, error)    // Business logic!
    EvaluatePolicy(ctx context.Context, policy string) error        // Business logic!
}
```

### 3. Event-Driven Architecture

All operations must emit structured events for observability and agent communication.

```go
// ✅ CORRECT: Always emit events
func (s *DeploymentService) PlanDeployment(ctx context.Context, app string) (*Plan, error) {
    // Emit planning started event
    s.eventBus.Emit("deployment.planning.started", map[string]interface{}{
        "app": app,
    })
    
    plan, err := s.generatePlan(ctx, app)
    if err != nil {
        // Emit failure event
        s.eventBus.Emit("deployment.planning.failed", map[string]interface{}{
            "app":   app,
            "error": err.Error(),
        })
        return nil, err
    }
    
    // Emit success event
    s.eventBus.Emit("deployment.planning.completed", map[string]interface{}{
        "app":  app,
        "plan": plan,
    })
    
    return plan, nil
}
```

### 4. Policy-First Development

Always validate policies before making changes.

```go
// ✅ CORRECT: Check policies first
func (s *DeploymentService) DeployApplication(ctx context.Context, app, env string) error {
    // Policy validation before any action
    if err := s.policyService.ValidateDeployment(ctx, app, env); err != nil {
        return fmt.Errorf("deployment policy violation: %w", err)
    }
    
    // Proceed with deployment only after policy approval
    return s.executeDeployment(ctx, app, env)
}
```

## Required File Patterns

### Package Structure

```
internal/
├── agents/           # AI agent implementations (future)
├── ai/              # AI infrastructure ONLY
├── api/             # HTTP handlers (thin layer)
├── deployments/     # Deployment domain logic
├── events/          # Event system implementation
├── graph/           # Graph database and operations
├── policies/        # Policy domain logic
└── security/        # Security domain logic
```

### File Naming Conventions

- `service.go` - Main domain service implementation
- `handler.go` - HTTP request handlers (thin layer)
- `types.go` - Domain types and structs
- `*_test.go` - Unit tests co-located with source
- `mock_*.go` - Mock implementations for testing

## Code Patterns to Follow

### API Handler Pattern (Thin Layer)

```go
func (h *DeploymentHandler) PlanDeployment(w http.ResponseWriter, r *http.Request) {
    // 1. Extract and validate input
    app := r.URL.Query().Get("app")
    if app == "" {
        http.Error(w, "app parameter required", http.StatusBadRequest)
        return
    }
    
    // 2. Call domain service (business logic)
    plan, err := h.deploymentService.PlanDeployment(r.Context(), app)
    if err != nil {
        h.handleError(w, err)
        return
    }
    
    // 3. Return response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(plan)
}
```

### Domain Service Pattern (Business Logic)

```go
func (s *DeploymentService) PlanDeployment(ctx context.Context, app string) (*Plan, error) {
    // 1. Validate input
    if err := s.validateApplication(ctx, app); err != nil {
        return nil, fmt.Errorf("application validation failed: %w", err)
    }
    
    // 2. Check policies
    if err := s.policyService.ValidateDeployment(ctx, app); err != nil {
        return nil, fmt.Errorf("policy validation failed: %w", err)
    }
    
    // 3. Generate plan using AI (as infrastructure tool)
    if s.aiProvider != nil {
        prompt := s.buildDeploymentPrompt(app)
        response, err := s.aiProvider.CallAI(ctx, prompt.System, prompt.User)
        if err == nil {
            if plan, err := s.parseDeploymentPlan(response); err == nil {
                return plan, nil
            }
        }
    }
    
    // 4. Fallback to traditional planning
    return s.generateBasicPlan(app), nil
}
```

### Error Handling Pattern

```go
// Define domain-specific errors
var (
    ErrInvalidApplication = errors.New("invalid application")
    ErrPolicyViolation   = errors.New("policy violation")
)

// Use proper error wrapping
if err := someOperation(); err != nil {
    return fmt.Errorf("operation failed for app %s: %w", app, err)
}

// Handle errors consistently in handlers
switch {
case errors.Is(err, ErrInvalidApplication):
    http.Error(w, "Invalid application", http.StatusBadRequest)
case errors.Is(err, ErrPolicyViolation):
    http.Error(w, "Policy violation", http.StatusForbidden)
default:
    http.Error(w, "Internal server error", http.StatusInternalServerError)
}
```

### Dependency Injection Pattern

```go
// Constructor pattern
func NewDeploymentService(
    graph graph.Graph,
    aiProvider ai.Provider,
    policyService policies.Service,
    eventBus events.Bus,
) *DeploymentService {
    return &DeploymentService{
        graph:         graph,
        aiProvider:    aiProvider,
        policyService: policyService,
        eventBus:      eventBus,
    }
}

// Use interfaces for testability
type DeploymentService struct {
    graph         graph.Graph         // Interface, not concrete
    aiProvider    ai.Provider         // Interface, not concrete
    policyService policies.Service    // Interface, not concrete
    eventBus      events.Bus         // Interface, not concrete
}
```

## What NOT to Do (Anti-Patterns)

### ❌ Business Logic in API Handlers

```go
// DON'T DO THIS
func (h *DeploymentHandler) PlanDeployment(w http.ResponseWriter, r *http.Request) {
    // Complex business logic in handler - WRONG!
    if strings.Contains(app, "prod") && !strings.Contains(app, "canary") {
        // Policy logic in handler - WRONG!
    }
    
    // Direct AI calls from handler - WRONG!
    response, err := h.aiProvider.CallAI(ctx, prompt)
}
```

### ❌ Business Logic in AI Components

```go
// DON'T DO THIS
func (ai *AIBrain) GenerateDeploymentPlan(app string) (*Plan, error) {
    // Deployment business logic in AI layer - WRONG!
    if isProductionApp(app) {
        // Domain-specific logic in AI component - WRONG!
    }
}
```

### ❌ Direct AI Provider Calls from Handlers

```go
// DON'T DO THIS
func (h *Handler) SomeEndpoint(w http.ResponseWriter, r *http.Request) {
    // Skipping domain service - WRONG!
    response, err := h.aiProvider.CallAI(ctx, prompt)
}
```

## Testing Requirements

### Test-Driven Development (TDD)

Always write tests first, following the Red-Green-Refactor cycle.

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
            // Use mocks for dependencies
            mockGraph := &MockGraph{}
            mockAI := &MockAIProvider{}
            mockPolicy := &MockPolicyService{}
            mockEvents := &MockEventBus{}
            
            service := NewDeploymentService(mockGraph, mockAI, mockPolicy, mockEvents)
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

### Test Organization

- **Unit Tests**: Test business logic in isolation with mocks
- **Integration Tests**: Test API endpoints end-to-end
- **Policy Tests**: Verify policy enforcement in all scenarios

## Current Architecture Issues to Fix

### ⚠️ Files to Delete

- `/internal/ai/ai_brain.go` (997 lines) - Contains misplaced business logic

### ⚠️ Files to Refactor

- `/internal/ai/openai_provider.go` (852 lines) - Remove business logic, keep only infrastructure
- `/internal/deployments/service.go` - Complete AI integration using clean patterns
- `/internal/policies/service.go` - Complete AI integration using clean patterns

## Guidelines for New Features

### When Adding AI-Enhanced Features

1. **Add method to appropriate domain service** (`internal/[domain]/service.go`)
2. **Domain service uses AI provider as infrastructure tool**
3. **Domain service owns all business logic and validation**
4. **API handler calls domain service (thin layer)**
5. **Emit events for all state changes**
6. **Add comprehensive tests**

### When Refactoring Existing AI Code

1. **Identify business logic currently in AI layer**
2. **Move business logic to appropriate domain service**
3. **Keep only infrastructure code in AI provider**
4. **Update API handlers to use domain services**
5. **Remove business methods from AI provider interface**

## Important Context

- **Platform Goal**: Transitioning to AI-native where AI is primary interface
- **Current State**: API-first with some AI integration
- **Architecture Vision**: Multi-agent system with specialized AI agents
- **Domain Focus**: Deployment, Policy, Security, and Infrastructure management

## Documentation

- **Architecture Guide**: `/docs/ai-platform-architecture.md` - Complete architectural vision
- **Domain Separation Plan**: `/DOMAIN_SEPARATION_PLAN.md` - Cleanup requirements
- **Migration Guide**: `/docs/migration-guide.md` - Best practices and patterns

Remember: Domain services own business logic, AI providers are infrastructure tools, API handlers are thin layers, and events drive system observability and communication.
