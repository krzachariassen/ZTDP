# Domain-Driven Design in ZTDP

## Introduction

ZTDP uses Domain-Driven Design (DDD) principles to organize business logic around core domain concepts. This approach ensures that our AI-native platform reflects the real-world problems it solves and maintains clear boundaries between different areas of concern.

## Core Domain Concepts

### 1. ZTDP Domain Model

Our platform operates in the **Developer Platform Domain** with these key subdomains:

#### Core Domains (High Value, High Complexity)
- **Deployment Domain**: Application lifecycle management, infrastructure provisioning
- **Policy Domain**: Governance, compliance, and security policy enforcement
- **AI Orchestration Domain**: Multi-agent coordination and intelligent automation

#### Supporting Domains (High Value, Lower Complexity)
- **Application Domain**: Application modeling and metadata management
- **Environment Domain**: Environment configuration and management
- **Resource Domain**: Infrastructure resource modeling

#### Generic Domains (Lower Value, can be outsourced)
- **Authentication Domain**: User authentication and authorization
- **Logging Domain**: System logging and audit trails
- **Monitoring Domain**: System health and performance monitoring

### 2. Bounded Contexts

Each domain has clear boundaries and owns its data and business logic:

```
┌─────────────────────────────────────┐
│        DEPLOYMENT CONTEXT           │
│  ├── DeploymentService             │
│  ├── DeploymentPlan                │
│  ├── DeploymentStrategy            │
│  └── DeploymentAgent               │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│         POLICY CONTEXT              │
│  ├── PolicyService                 │
│  ├── PolicyEvaluator               │
│  ├── PolicyRule                    │
│  └── ComplianceAgent               │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│       APPLICATION CONTEXT           │
│  ├── ApplicationService            │
│  ├── Application                   │
│  ├── ServiceComponent              │
│  └── ApplicationAgent              │
└─────────────────────────────────────┘
```

## Domain Services Pattern

### 1. Domain Service Structure

Each domain follows a consistent structure:

```go
// Domain service contains all business logic for the domain
type DeploymentService struct {
    // Infrastructure dependencies (interfaces)
    graph         graph.Graph
    aiProvider    ai.Provider
    policyService policies.Service
    eventBus      events.Bus
}

// Constructor with dependency injection
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

// Domain methods implement business logic
func (s *DeploymentService) PlanDeployment(ctx context.Context, app string) (*Plan, error) {
    // Business logic implementation
}
```

### 2. Domain Method Patterns

#### Business Logic Ownership
```go
// ✅ CORRECT: Domain service owns business logic
func (s *DeploymentService) ValidateDeploymentReadiness(ctx context.Context, app string) error {
    // Domain-specific validation rules
    if app == "" {
        return ErrInvalidApplication
    }
    
    // Business rule: production deployments require approval
    env, err := s.graph.GetApplicationEnvironment(app)
    if err != nil {
        return err
    }
    
    if env == "production" {
        approval, err := s.graph.GetApprovalStatus(app)
        if err != nil {
            return err
        }
        
        if !approval.Approved {
            return ErrProductionRequiresApproval
        }
    }
    
    return nil
}
```

#### Cross-Domain Coordination
```go
// Domain services coordinate through well-defined interfaces
func (s *DeploymentService) ExecuteDeployment(ctx context.Context, app string) error {
    // 1. Validate policies (Policy Domain)
    if err := s.policyService.ValidateDeployment(ctx, app); err != nil {
        return fmt.Errorf("policy validation failed: %w", err)
    }
    
    // 2. Execute deployment (Deployment Domain)
    plan, err := s.generateDeploymentPlan(ctx, app)
    if err != nil {
        return fmt.Errorf("plan generation failed: %w", err)
    }
    
    // 3. Emit events for other domains
    s.eventBus.Emit("deployment.started", DeploymentEvent{
        App:  app,
        Plan: plan,
    })
    
    return s.executeDeploymentPlan(ctx, plan)
}
```

## Domain Models and Entities

### 1. Rich Domain Models

Domain models contain behavior, not just data:

```go
// Rich domain model with behavior
type DeploymentPlan struct {
    ID          string
    Application string
    Environment string
    Steps       []DeploymentStep
    Status      DeploymentStatus
    CreatedAt   time.Time
}

// Domain model methods implement business rules
func (p *DeploymentPlan) CanExecute() bool {
    return p.Status == StatusPlanned && len(p.Steps) > 0
}

func (p *DeploymentPlan) Execute() error {
    if !p.CanExecute() {
        return ErrPlanNotReady
    }
    
    p.Status = StatusExecuting
    return nil
}

func (p *DeploymentPlan) EstimateDuration() time.Duration {
    var total time.Duration
    for _, step := range p.Steps {
        total += step.EstimatedDuration
    }
    return total
}
```

### 2. Value Objects

Immutable objects that represent concepts:

```go
// Value object for deployment strategy
type DeploymentStrategy struct {
    name     string
    rollout  RolloutStrategy
    canary   CanaryConfig
}

func NewDeploymentStrategy(name string, rollout RolloutStrategy) DeploymentStrategy {
    return DeploymentStrategy{
        name:    name,
        rollout: rollout,
    }
}

func (ds DeploymentStrategy) Name() string {
    return ds.name
}

func (ds DeploymentStrategy) IsCanary() bool {
    return ds.rollout == RolloutCanary
}
```

## Aggregates and Boundaries

### 1. Aggregate Design

Aggregates ensure consistency boundaries:

```go
// Application aggregate root
type Application struct {
    id       ApplicationID
    name     string
    services []*Service      // Part of aggregate
    status   ApplicationStatus
    version  int             // For optimistic locking
}

// Aggregate methods maintain invariants
func (a *Application) AddService(service *Service) error {
    // Business rule: max 10 services per application
    if len(a.services) >= 10 {
        return ErrTooManyServices
    }
    
    // Business rule: unique service names
    for _, existing := range a.services {
        if existing.Name == service.Name {
            return ErrDuplicateServiceName
        }
    }
    
    a.services = append(a.services, service)
    a.version++ // Optimistic locking
    return nil
}

func (a *Application) RemoveService(serviceName string) error {
    for i, service := range a.services {
        if service.Name == serviceName {
            // Business rule: cannot remove running services
            if service.Status == ServiceRunning {
                return ErrCannotRemoveRunningService
            }
            
            a.services = append(a.services[:i], a.services[i+1:]...)
            a.version++
            return nil
        }
    }
    return ErrServiceNotFound
}
```

### 2. Repository Pattern

Repositories provide aggregate persistence:

```go
// Repository interface (defined by domain)
type ApplicationRepository interface {
    Save(ctx context.Context, app *Application) error
    FindByID(ctx context.Context, id ApplicationID) (*Application, error)
    FindByName(ctx context.Context, name string) (*Application, error)
}

// Domain service uses repository
func (s *ApplicationService) UpdateApplication(ctx context.Context, id ApplicationID, updates ApplicationUpdates) error {
    // Load aggregate
    app, err := s.repository.FindByID(ctx, id)
    if err != nil {
        return err
    }
    
    // Apply business logic
    if err := app.ApplyUpdates(updates); err != nil {
        return err
    }
    
    // Save aggregate
    return s.repository.Save(ctx, app)
}
```

## AI Integration with Domain Services

### 1. AI as Domain Tool

AI providers are infrastructure tools used by domain services:

```go
func (s *DeploymentService) GenerateIntelligentPlan(ctx context.Context, app string) (*Plan, error) {
    // Domain knowledge and validation
    appInfo, err := s.graph.GetApplication(app)
    if err != nil {
        return nil, err
    }
    
    if !s.isDeploymentReady(appInfo) {
        return nil, ErrApplicationNotReady
    }
    
    // Use AI as infrastructure tool
    prompt := s.buildDeploymentPrompt(appInfo)
    aiResponse, err := s.aiProvider.CallAI(ctx, prompt.System, prompt.User)
    if err != nil {
        // Fallback to traditional planning
        return s.generateBasicPlan(appInfo), nil
    }
    
    // Domain service parses and validates AI response
    plan, err := s.parseAIPlan(aiResponse)
    if err != nil {
        return s.generateBasicPlan(appInfo), nil
    }
    
    // Apply domain validation to AI-generated plan
    if err := s.validatePlan(plan); err != nil {
        return s.generateBasicPlan(appInfo), nil
    }
    
    return plan, nil
}
```

### 2. Domain-Specific AI Prompts

Each domain builds prompts with domain expertise:

```go
func (s *DeploymentService) buildDeploymentPrompt(app *Application) *AIPrompt {
    systemPrompt := `You are a deployment planning expert. Generate deployment plans 
    that follow these principles:
    1. Zero-downtime deployments when possible
    2. Canary deployments for production
    3. Proper rollback strategies
    4. Resource optimization`
    
    userPrompt := fmt.Sprintf(`Generate a deployment plan for:
    Application: %s
    Environment: %s
    Current Version: %s
    Target Version: %s
    Services: %v
    Dependencies: %v
    Resource Requirements: %v`,
        app.Name,
        app.Environment,
        app.CurrentVersion,
        app.TargetVersion,
        app.Services,
        app.Dependencies,
        app.ResourceRequirements,
    )
    
    return &AIPrompt{
        System: systemPrompt,
        User:   userPrompt,
    }
}
```

## Event-Driven Communication Between Domains

### 1. Domain Events

Domains communicate through events:

```go
// Domain event
type ApplicationDeployed struct {
    ApplicationID string
    Environment   string
    Version       string
    Timestamp     time.Time
    DeploymentID  string
}

// Emit events for cross-domain coordination
func (s *DeploymentService) CompleteDeployment(ctx context.Context, deploymentID string) error {
    deployment, err := s.getDeployment(deploymentID)
    if err != nil {
        return err
    }
    
    // Update deployment status
    deployment.Status = StatusCompleted
    deployment.CompletedAt = time.Now()
    
    // Emit domain event
    event := ApplicationDeployed{
        ApplicationID: deployment.ApplicationID,
        Environment:   deployment.Environment,
        Version:       deployment.Version,
        Timestamp:     deployment.CompletedAt,
        DeploymentID:  deploymentID,
    }
    
    s.eventBus.Publish("application.deployed", event)
    
    return nil
}
```

### 2. Event Handlers

Other domains react to events:

```go
// Policy domain reacts to deployment events
func (s *PolicyService) HandleApplicationDeployed(event ApplicationDeployed) error {
    // Update compliance tracking
    compliance := &ComplianceRecord{
        ApplicationID: event.ApplicationID,
        Environment:   event.Environment,
        DeployedAt:    event.Timestamp,
        Status:        ComplianceChecking,
    }
    
    // Start compliance validation
    go s.validatePostDeploymentCompliance(compliance)
    
    return nil
}
```

## Testing Domain Logic

### 1. Domain Service Testing

Test business logic in isolation:

```go
func TestDeploymentService_ValidateDeploymentReadiness(t *testing.T) {
    tests := []struct {
        name        string
        app         string
        setupMocks  func(*MockGraph)
        wantErr     bool
        expectedErr error
    }{
        {
            name: "valid deployment",
            app:  "test-app",
            setupMocks: func(mg *MockGraph) {
                mg.On("GetApplicationEnvironment", "test-app").Return("dev", nil)
            },
            wantErr: false,
        },
        {
            name: "production requires approval",
            app:  "prod-app",
            setupMocks: func(mg *MockGraph) {
                mg.On("GetApplicationEnvironment", "prod-app").Return("production", nil)
                mg.On("GetApprovalStatus", "prod-app").Return(&Approval{Approved: false}, nil)
            },
            wantErr:     true,
            expectedErr: ErrProductionRequiresApproval,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockGraph := &MockGraph{}
            tt.setupMocks(mockGraph)
            
            service := NewDeploymentService(mockGraph, nil, nil, nil)
            err := service.ValidateDeploymentReadiness(context.Background(), tt.app)
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.expectedErr != nil {
                    assert.Equal(t, tt.expectedErr, err)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Domain Model Testing

Test domain model behavior:

```go
func TestDeploymentPlan_CanExecute(t *testing.T) {
    tests := []struct {
        name   string
        plan   *DeploymentPlan
        want   bool
    }{
        {
            name: "planned status with steps",
            plan: &DeploymentPlan{
                Status: StatusPlanned,
                Steps:  []DeploymentStep{{Name: "deploy"}},
            },
            want: true,
        },
        {
            name: "executing status",
            plan: &DeploymentPlan{
                Status: StatusExecuting,
                Steps:  []DeploymentStep{{Name: "deploy"}},
            },
            want: false,
        },
        {
            name: "no steps",
            plan: &DeploymentPlan{
                Status: StatusPlanned,
                Steps:  []DeploymentStep{},
            },
            want: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.plan.CanExecute()
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Domain-Driven Anti-Patterns to Avoid

### ❌ Anemic Domain Models
```go
// DON'T DO THIS - just data containers
type Application struct {
    ID       string
    Name     string
    Services []Service
}

// Business logic scattered in services instead of model
func (s *ApplicationService) ValidateApplication(app Application) error {
    // Validation logic that should be in the model
}
```

### ❌ Cross-Domain Dependencies
```go
// DON'T DO THIS - deployment domain importing policy internals
import "internal/policies/validator"

func (s *DeploymentService) SomeMethod() {
    // Direct dependency on another domain's internals
    policyValidator := validator.New()
}
```

### ❌ Shared Database Between Domains
```go
// DON'T DO THIS - domains sharing database tables
func (s *DeploymentService) GetPolicy(id string) *Policy {
    // Deployment service accessing policy table directly
    return s.db.Query("SELECT * FROM policies WHERE id = ?", id)
}
```

## Benefits of Domain-Driven Design

### 1. Business Alignment
- Code reflects business concepts and language
- Domain experts can understand and validate the model
- Changes in business requirements map clearly to code changes

### 2. Maintainability
- Clear boundaries between different areas of concern
- Changes are localized to appropriate domains
- Business logic is centralized and protected

### 3. Testability
- Domain logic can be tested in isolation
- Business rules are explicit and verifiable
- Mock dependencies at domain boundaries

### 4. AI Integration
- Domain services provide context and expertise to AI systems
- AI responses are validated against domain knowledge
- Business rules are preserved even with AI enhancement

## Related Documentation

- **[Clean Architecture Principles](clean-architecture-principles.md)** - Architectural foundation
- **[Event-Driven Architecture](event-driven-architecture.md)** - Cross-domain communication
- **[Testing Strategies](testing-strategies.md)** - Testing domain logic
- **[Architecture Overview](architecture-overview.md)** - High-level system design
