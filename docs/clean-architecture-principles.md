# Clean Architecture Principles for ZTDP

## Introduction

ZTDP follows clean architecture principles to ensure maintainable, testable, and scalable code. This document outlines the core principles and how they apply to our AI-native platform.

## Core Clean Architecture Principles

### 1. Dependency Direction Rule

**Rule**: Core business logic should not depend on external layers

**Implementation**:
- Domain services use interfaces, not concrete implementations
- Dependencies point inward toward business logic
- External concerns (databases, AI providers, HTTP) depend on business logic, not vice versa

```go
// ✅ CORRECT: Domain service uses interface
type DeploymentService struct {
    graph      graph.Graph      // Interface
    aiProvider ai.Provider      // Interface
    eventBus   events.Bus       // Interface
}

// ❌ WRONG: Domain service depends on concrete implementation
type DeploymentService struct {
    redisClient *redis.Client   // Concrete dependency
    openAI      *openai.Client  // Concrete dependency
}
```

**Validation**: Dependencies point inward toward business logic

### 2. Domain-Driven Design Integration

**Domain Services Own Business Logic**:
- Deployment, Policy, Security services contain all domain-specific logic
- Each domain service is responsible for its business rules and validation
- Cross-domain operations are coordinated through events or service composition

**API Handlers Are Thin**:
- Only handle HTTP concerns (parsing, validation, response formatting)
- Delegate all business logic to domain services
- No business rules or complex logic in handlers

**AI Components Are Infrastructure**:
- AI providers are tools used by domain services
- AI components do not contain business logic
- Domain services orchestrate AI capabilities

```go
// ✅ CORRECT: Thin API handler
func CreateApplication(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    var app contracts.ApplicationContract
    if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
        WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // 2. Delegate to domain service
    if err := h.appService.CreateApplication(app); err != nil {
        WriteJSONError(w, err.Error(), http.StatusBadRequest)
        return
    }

    // 3. Return response
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(app)
}

// ✅ CORRECT: Domain service with business logic
func (s *ApplicationService) CreateApplication(contract contracts.ApplicationContract) error {
    // Business validation
    if err := s.ValidateContract(contract); err != nil {
        return err
    }
    
    // Policy enforcement
    if err := s.policyService.ValidateCreate(contract); err != nil {
        return err
    }
    
    // Execute business operation
    return s.graph.AddApplication(contract)
}
```

### 3. Layer Separation

#### Presentation Layer (API Handlers)
- **Responsibility**: HTTP request/response handling
- **Dependencies**: Domain services (interfaces)
- **Rules**: No business logic, thin adapters only

#### Application Layer (Domain Services)
- **Responsibility**: Business logic, orchestration, validation
- **Dependencies**: Domain models, infrastructure interfaces
- **Rules**: Core business rules, use case implementation

#### Infrastructure Layer (Providers, Repositories)
- **Responsibility**: External system integration
- **Dependencies**: Nothing (implements interfaces defined by application layer)
- **Rules**: Pure infrastructure, no business logic

```
┌─────────────────────────────────────┐
│         API HANDLERS               │  ← Presentation Layer
│  (HTTP, JSON, Request/Response)    │
└─────────────┬───────────────────────┘
              │ depends on
┌─────────────▼───────────────────────┐
│       DOMAIN SERVICES              │  ← Application Layer
│  (Business Logic, Validation)      │
└─────────────┬───────────────────────┘
              │ depends on
┌─────────────▼───────────────────────┐
│    INFRASTRUCTURE PROVIDERS        │  ← Infrastructure Layer
│  (Database, AI, External APIs)     │
└─────────────────────────────────────┘
```

### 4. Interface Segregation

**Design Small, Focused Interfaces**:
```go
// ✅ GOOD: Small, focused interfaces
type PlanGenerator interface {
    GeneratePlan(ctx context.Context, app string) (*Plan, error)
}

type PolicyValidator interface {
    ValidatePolicy(ctx context.Context, policy string) error
}

// ❌ BAD: Large, monolithic interface
type MegaService interface {
    GeneratePlan(ctx context.Context, app string) (*Plan, error)
    ValidatePolicy(ctx context.Context, policy string) error
    DeployApplication(ctx context.Context, app string) error
    MonitorHealth(ctx context.Context) (*Health, error)
    // ... 20 more methods
}
```

### 5. Dependency Injection

**Constructor Pattern**:
```go
// Dependencies injected through constructor
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
```

**Interface Usage**:
```go
// Use interfaces for all dependencies
type DeploymentService struct {
    graph         graph.Graph         // Interface
    aiProvider    ai.Provider         // Interface
    policyService policies.Service    // Interface
    eventBus      events.Bus         // Interface
}
```

## ZTDP-Specific Patterns

### 1. AI-as-Infrastructure Pattern

**Correct**: Domain services use AI providers as tools
```go
func (s *DeploymentService) PlanDeployment(ctx context.Context, app string) (*Plan, error) {
    // Business validation first
    if err := s.validateApplication(app); err != nil {
        return nil, err
    }
    
    // Use AI as infrastructure tool
    if s.aiProvider != nil {
        prompt := s.buildDeploymentPrompt(app)
        response, err := s.aiProvider.CallAI(ctx, prompt.System, prompt.User)
        if err == nil {
            if plan, err := s.parseDeploymentPlan(response); err == nil {
                return plan, nil
            }
        }
    }
    
    // Fallback to traditional planning
    return s.generateBasicPlan(app), nil
}
```

**Incorrect**: AI components containing business logic
```go
// ❌ WRONG: Business logic in AI component
func (ai *AIBrain) GenerateDeploymentPlan(app string) (*Plan, error) {
    // Complex deployment business rules - WRONG LAYER!
    if isProductionApp(app) && !hasApproval(app) {
        return nil, errors.New("production deployment requires approval")
    }
    // Domain-specific logic in AI layer - WRONG!
}
```

### 2. Event-Driven Architecture Pattern

**Consistent Event Emission**:
```go
func (s *DeploymentService) DeployApplication(ctx context.Context, app string) error {
    // Emit started event
    s.eventBus.Emit("deployment.started", map[string]interface{}{
        "app": app,
        "timestamp": time.Now(),
    })
    
    err := s.executeDeployment(ctx, app)
    if err != nil {
        // Emit failure event
        s.eventBus.Emit("deployment.failed", map[string]interface{}{
            "app": app,
            "error": err.Error(),
        })
        return err
    }
    
    // Emit success event
    s.eventBus.Emit("deployment.completed", map[string]interface{}{
        "app": app,
        "duration": time.Since(start),
    })
    
    return nil
}
```

### 3. Policy-First Pattern

**Always Validate Policies Before Operations**:
```go
func (s *DeploymentService) DeployApplication(ctx context.Context, app, env string) error {
    // Policy validation BEFORE any action
    if err := s.policyService.ValidateDeployment(ctx, app, env); err != nil {
        return fmt.Errorf("deployment policy violation: %w", err)
    }
    
    // Only proceed after policy approval
    return s.executeDeployment(ctx, app, env)
}
```

## Testing Patterns

### 1. Unit Testing with Mocks

```go
func TestDeploymentService_PlanDeployment(t *testing.T) {
    // Arrange - create mocks
    mockGraph := &MockGraph{}
    mockAI := &MockAIProvider{}
    mockPolicy := &MockPolicyService{}
    mockEvents := &MockEventBus{}
    
    service := NewDeploymentService(mockGraph, mockAI, mockPolicy, mockEvents)
    
    // Act
    plan, err := service.PlanDeployment(ctx, "test-app")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, plan)
    
    // Verify interactions
    mockAI.AssertCalled(t, "CallAI", mock.Anything, mock.Anything, mock.Anything)
}
```

### 2. Integration Testing

```go
func TestCreateApplicationEndToEnd(t *testing.T) {
    // Setup real dependencies for integration test
    router := newTestRouter(t)
    
    // Test end-to-end flow
    app := contracts.ApplicationContract{Name: "test-app"}
    response := createApplication(t, router, app)
    
    assert.Equal(t, http.StatusCreated, response.StatusCode)
    verifyApplicationExists(t, "test-app")
}
```

## Common Anti-Patterns to Avoid

### ❌ Business Logic in API Handlers
```go
// DON'T DO THIS
func CreateApplication(w http.ResponseWriter, r *http.Request) {
    // Complex validation logic in handler - WRONG!
    if app.Name == "" || len(app.Name) > 50 {
        // Validation logic should be in domain service
    }
    
    // Direct database access from handler - WRONG!
    db.Save(app)
}
```

### ❌ Domain Services Depending on Concrete Types
```go
// DON'T DO THIS
type DeploymentService struct {
    redisClient *redis.Client    // Concrete dependency
    httpClient  *http.Client     // Concrete dependency
}
```

### ❌ AI Components with Business Logic
```go
// DON'T DO THIS
func (ai *AIProvider) CreateApplication(name string) error {
    // Business validation in AI component - WRONG!
    if !isValidName(name) {
        return errors.New("invalid application name")
    }
}
```

## Architecture Validation Checklist

### ✅ Dependency Direction
- [ ] Domain services don't import concrete infrastructure types
- [ ] All external dependencies are behind interfaces
- [ ] Dependencies point inward toward business logic

### ✅ Layer Separation
- [ ] API handlers are thin (< 50 lines typical)
- [ ] Business logic is in domain services
- [ ] Infrastructure code has no business logic

### ✅ Interface Design
- [ ] Interfaces are small and focused
- [ ] Each interface has a single responsibility
- [ ] Interfaces are defined by their consumers

### ✅ Testing
- [ ] Unit tests use mocks for all dependencies
- [ ] Integration tests cover end-to-end scenarios
- [ ] Each layer can be tested independently

### ✅ ZTDP-Specific Patterns
- [ ] AI providers are pure infrastructure
- [ ] All operations emit structured events
- [ ] Policy validation occurs before all state changes
- [ ] Domain services coordinate AI capabilities

## Benefits of Clean Architecture

### 1. Testability
- Each layer can be tested in isolation
- Easy to mock dependencies
- Fast unit tests with comprehensive coverage

### 2. Flexibility
- Easy to swap implementations (database, AI provider, etc.)
- New features don't break existing code
- Changes are localized to appropriate layers

### 3. Maintainability
- Clear separation of concerns
- Business logic is centralized and protected
- Infrastructure changes don't affect business rules

### 4. Scalability
- Individual layers can be optimized independently
- Easy to add new features following established patterns
- Clear boundaries enable team specialization

## Related Documentation

- **[Architecture Overview](architecture-overview.md)** - High-level architecture vision
- **[Domain-Driven Design](domain-driven-design.md)** - Domain modeling principles
- **[Testing Strategies](testing-strategies.md)** - Comprehensive testing approaches
- **[AI Platform Architecture](ai-platform-architecture.md)** - Complete platform details
