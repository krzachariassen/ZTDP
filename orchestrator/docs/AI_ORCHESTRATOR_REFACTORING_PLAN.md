# AI Orchestrator Refactoring Plan
## Clean Architecture + TDD + Domain-Driven Design

### Current Problems ❌
- **No Type Safety**: Raw `map[string]interface{}` everywhere
- **No Business Rules**: Direct graph manipulation bypasses domain logic
- **No Validation**: Invalid data can corrupt the graph
- **Monolithic Files**: Single responsibility principle violated
- **Generic Interface**: Implementation details exposed
- **Poor Testability**: Complex dependencies make testing difficult

### Vision ✅
**AI-Native Platform with Governance**: Use AI as the intelligent brain while maintaining:
- Type safety at compile time
- Domain rules enforcement
- Proper validation and security
- Clean architecture boundaries
- Comprehensive test coverage

---

## Phase 1: Domain Layer - Graph Domain Functions
**Duration: 2-3 days**

### 1.1 Create Graph Domain Types
```
internal/graph/domain/
├── agent.go           # Agent domain model & validation
├── execution.go       # Execution plan domain model
├── conversation.go    # Conversation domain model
└── graph_service.go   # Domain service interface
```

**Files to create:**
- `Agent` struct with validation
- `ExecutionPlan` struct with business rules
- `Conversation` struct with user context
- Domain service interfaces

### 1.2 TDD Implementation Order
1. **RED**: Write tests for `Agent` domain model
2. **GREEN**: Implement `Agent` with validation
3. **REFACTOR**: Clean up and optimize
4. Repeat for `ExecutionPlan` and `Conversation`

### 1.3 Domain Services
```go
type GraphDomainService interface {
    // Agent operations with type safety
    AddAgent(ctx context.Context, agent *Agent) error
    GetAllAgents(ctx context.Context) ([]*Agent, error)
    UpdateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error
    
    // Execution operations with business rules
    CreateExecutionPlan(ctx context.Context, plan *ExecutionPlan) error
    GetExecutingPlans(ctx context.Context, userID string) ([]*ExecutionPlan, error)
    UpdateExecutionStep(ctx context.Context, planID, stepID string, result *StepResult) error
    
    // Conversation operations with validation
    StoreConversation(ctx context.Context, conv *Conversation) error
    GetUserConversations(ctx context.Context, userID string, limit int) ([]*Conversation, error)
}
```

---

## Phase 2: Application Layer - Orchestrator Decomposition
**Duration: 3-4 days**

### 2.1 Split Orchestrator into Clean Components (Business-Domain-Centric)
```
internal/
├── orchestrator/ # NEW MODEL FOLLOWS
├── conversation/
│   ├── domain/
│   │   ├── conversation.go   # Conversation domain model
│   │   └── analyzer.go       # Conversation analysis domain logic
│   ├── application/
│   │   └── conversation_service.go # Conversation application service
│   ├── infrastructure/
│   │   └── graph_repository.go # Conversation persistence
│   └── ports/               # Interfaces for external dependencies
├── planning/
│   ├── domain/
│   │   ├── execution_plan.go # Execution plan domain model
│   │   └── planner.go       # Planning domain logic
│   ├── application/
│   │   └── planning_service.go # Planning application service
│   └── infrastructure/
│       └── graph_repository.go # Plan persistence
├── routing/
│   ├── domain/
│   │   ├── router.go        # Agent routing domain logic
│   │   └── capability_matcher.go # Capability matching
│   ├── application/
│   │   └── routing_service.go # Routing application service
│   └── infrastructure/
│       └── graph_repository.go # Agent data access
└── execution/
    ├── domain/
    │   ├── executor.go      # Execution domain logic
    │   └── step_processor.go # Step processing logic
    ├── application/
    │   └── execution_service.go # Execution application service
    └── infrastructure/
        └── graph_repository.go # Execution state persistence

New structure for the orchestrator:
internal/orchestrator/domain/
├── analysis.go          # AI analysis results (Intent, Category, Confidence)
├── decision.go          # AI decisions (CLARIFY vs EXECUTE)
├── response.go          # Response parsing and formatting logic
├── orchestrator.go      # Core orchestrator domain logic

internal/orchestrator/application/
├── orchestrator_service.go      # Main AI orchestrator service 
├── ai_decision_engine.go        # AI decision logic (clarify vs execute)
├── graph_explorer.go            # AI graph exploration for agents
├── execution_coordinator.go     # AI execution coordination
└── learning_service.go          # AI insight storage

internal/orchestrator/infrastructure/
├── ai_provider.go               # AI provider implementation
└── graph_repository.go          # Graph access for orchestrator

```

### 2.2 TDD Implementation Strategy
Each component follows strict TDD:
1. **Interface First**: Define clear contracts
2. **Test First**: Write comprehensive tests
3. **Implementation**: Minimal code to pass tests
4. **Integration**: Wire components together

### 2.3 Dependency Injection
```go
type Orchestrator struct {
    conversationService conversation.Service
    planningService     planning.Service
    routingService      routing.Service
    executionService    execution.Service
    logger             logging.Logger
}

func NewOrchestrator(
    conversationSvc conversation.Service,
    planningSvc planning.Service,
    routingSvc routing.Service,
    executionSvc execution.Service,
    logger logging.Logger,
) *Orchestrator {
    return &Orchestrator{
        conversationService: conversationSvc,
        planningService:     planningSvc,
        routingService:      routingSvc,
        executionService:    executionSvc,
        logger:             logger,
    }
}
```

### 2.4 Clean Interface Contracts
```go
// Business domain interfaces - not technology-specific
type ConversationService interface {
    AnalyzeUserInput(ctx context.Context, userInput string) (*ConversationAnalysis, error)
    BuildResponse(ctx context.Context, result *ExecutionResult) (*Response, error)
}

type PlanningService interface {
    CreateExecutionPlan(ctx context.Context, analysis *ConversationAnalysis) (*ExecutionPlan, error)
    ValidatePlan(ctx context.Context, plan *ExecutionPlan) error
}

type RoutingService interface {
    FindCapableAgents(ctx context.Context, capabilities []string) ([]*Agent, error)
    SelectOptimalAgent(ctx context.Context, task *Task, agents []*Agent) (*Agent, error)
}

type ExecutionService interface {
    ExecutePlan(ctx context.Context, plan *ExecutionPlan) (*ExecutionResult, error)
    ProcessStep(ctx context.Context, step *ExecutionStep) (*StepResult, error)
}
```

### 2.5 Architecture Violations to Fix During Split

**Critical Fix Needed:**
- **Line 370**: Direct graph access in `storeInsightsToGraph()`
  ```go
  // VIOLATION: Direct graph manipulation
  ai.graph.AddNode(ctx, insightID, "insight", insightData)
  ```
  **Fix**: Use ConversationService to store insights properly

### 2.6 Clean Architecture Compliance Status

**✅ Already Using Domain Services:**
- AgentService for agent operations
- PlanningService for execution plans
- ConversationService dependency injected

**❌ Needs Fixing:**
- Direct graph access for insights storage
- Should use ConversationService instead

---

## Phase 3: AI Integration Points - Governance + Intelligence
**Duration: 2-3 days**

### 3.1 AI Decision Points (Where AI Adds Value)
- **Request Analysis**: Understanding user intent
- **Agent Selection**: Choosing optimal agents
- **Plan Generation**: Creating execution strategies
- **Error Recovery**: Handling failures intelligently

### 3.2 Governance Points (Where Rules Apply)
- **Input Validation**: Sanitize all user inputs
- **Agent Authorization**: Verify agent permissions
- **Resource Limits**: Prevent resource exhaustion
- **Audit Logging**: Track all decisions and actions

### 3.3 AI-Native with Safety
```go
type SafeAIProvider interface {
    // AI calls with input validation and output sanitization
    AnalyzeRequest(ctx context.Context, request *ValidatedRequest) (*Analysis, error)
    GeneratePlan(ctx context.Context, analysis *Analysis, agents []*Agent) (*ExecutionPlan, error)
    SelectAgent(ctx context.Context, task *Task, candidates []*Agent) (*Agent, error)
}
```

---

## Phase 4: Infrastructure Layer - Clean Boundaries
**Duration: 2-3 days**

### 4.1 Repository Pattern for Graph
```
internal/graph/
├── domain/          # Domain models (from Phase 1)
├── repository/
│   ├── graph_repo.go    # Repository implementation
│   └── memory_repo.go   # In-memory for testing
└── service/
    └── graph_service.go # Domain service implementation
```

### 4.2 Messaging Abstractions
```
internal/messaging/
├── domain/
│   ├── message.go       # Message domain model
│   └── bus_service.go   # Messaging service interface
├── rabbitmq/
│   └── bus_impl.go      # RabbitMQ implementation
└── memory/
    └── bus_impl.go      # In-memory for testing
```

---

## Phase 5: Testing Strategy - Comprehensive Coverage
**Duration: 1-2 days**

### 5.1 Test Categories
- **Unit Tests**: Each component in isolation (90%+ coverage)
- **Integration Tests**: Component interactions
- **Contract Tests**: Interface compliance
- **End-to-End Tests**: Full workflow validation

### 5.2 Test Structure
```
internal/ai/
├── conversation/
│   ├── analyzer.go
│   ├── analyzer_test.go      # Unit tests
│   └── analyzer_integration_test.go # Integration tests
├── planning/
│   ├── execution_planner.go
│   └── execution_planner_test.go
└── test/
    ├── fixtures/            # Test data
    ├── mocks/              # Generated mocks
    └── integration/        # Integration test suites
```

---

## Implementation Schedule

### Week 1: Foundation
- **Day 1-2**: Phase 1 - Graph domain models and services
- **Day 3-4**: Start Phase 2 - Core orchestrator decomposition
- **Day 5**: Testing and validation

### Week 2: AI Integration
- **Day 1-2**: Complete Phase 2 - Orchestrator components
- **Day 3-4**: Phase 3 - AI integration with governance
- **Day 5**: Phase 4 - Infrastructure boundaries

### Week 3: Polish & Testing
- **Day 1-2**: Phase 5 - Comprehensive testing
- **Day 3**: End-to-end validation
- **Day 4-5**: Documentation and cleanup

---

## Success Metrics

### Technical Quality
- **Type Safety**: 100% strongly typed interfaces
- **Test Coverage**: 85%+ code coverage
- **Architecture**: Clear layer boundaries
- **Performance**: Response time < 500ms

### Maintainability
- **File Size**: No file > 200 lines
- **Maintainability**: Cyclomatic complexity < 10
- **Dependencies**: Clear dependency graph
- **Documentation**: Comprehensive API docs

### AI-Native Benefits
- **Intelligence**: AI makes optimal decisions
- **Adaptability**: System learns and improves
- **Governance**: Rules are enforced consistently
- **Security**: All inputs validated and sanitized

---

## Risk Mitigation

### Technical Risks
- **Overengineering**: Keep interfaces simple and focused
- **Performance**: Benchmark critical paths early
- **Integration**: Incremental integration with existing system

### Process Risks
- **Scope Creep**: Stick to defined phases
- **Testing Debt**: Write tests before implementation
- **Breaking Changes**: Maintain backward compatibility

---

This plan balances AI-native intelligence with proper governance, security, and maintainability. Each phase builds on the previous one, ensuring we maintain a working system throughout the refactoring process.
