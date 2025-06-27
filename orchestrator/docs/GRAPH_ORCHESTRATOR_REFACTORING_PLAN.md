# Graph-Powered AI Orchestrator Refactoring Development Plan

## Current State Analysis

### Problems with Current Implementation
1. **Monolithic Design**: `graph_powered_orchestrator.go` (~428 lines) handles exploration, analysis, response generation, execution planning, and storage
2. **No Type Safety**: Heavy use of `map[string]interface{}` without proper domain types
3. **No Business Rules**: Direct graph manipulation without validation or business logic
4. **No Validation**: Input/output validation missing
5. **Infrastructure Concerns Mixed**: Direct graph operations mixed with business logic
6. **Poor Testability**: Large functions with multiple responsibilities

### Vision
Transform the orchestrator into a Clean Architecture implementation with:
- **Domain Layer**: Type-safe models with business rules
- **Application Layer**: Use cases and services
- **Infrastructure Layer**: Graph repositories and AI providers
- **TDD Approach**: Red-Green-Refactor for each component
- **AI-Native**: Keep AI exploration but with proper governance

## Phase 1: Domain Layer Foundation (Current)

### ✅ Completed
- Created `/internal/graph/domain/agent.go` with type-safe Agent model

### 🔄 In Progress
- Create ExecutionPlan domain model
- Create Conversation domain model
- Create RequestAnalysis domain model

### File Structure Target
```
internal/
  graph/
    domain/
      agent.go ✅
      execution_plan.go
      conversation.go
      request_analysis.go
      errors.go
      validation.go
```

## Phase 2: Repository Layer (Infrastructure)

### Goals
- Abstract graph operations behind domain repositories
- Implement validation and business rules
- Support TDD with interfaces

### File Structure Target
```
internal/
  graph/
    repository/
      interfaces.go
      agent_repository.go
      execution_plan_repository.go
      conversation_repository.go
    infrastructure/
      graph_agent_repository.go
      graph_execution_plan_repository.go
      graph_conversation_repository.go
```

### TDD Approach
1. **RED**: Write failing tests for repository interfaces
2. **GREEN**: Implement minimal repository functionality
3. **REFACTOR**: Add validation, error handling, business rules

## Phase 3: Application Services

### Goals
- Separate business logic from orchestration logic
- Create focused, single-responsibility services
- Enable proper dependency injection

### File Structure Target
```
internal/
  application/
    services/
      conversation_service.go
      analysis_service.go
      planning_service.go
      execution_service.go
      agent_discovery_service.go
    interfaces/
      services.go
```

### Services Breakdown

#### ConversationService
- Manages conversation state and history
- Handles user input validation
- Tracks conversation context

#### AnalysisService  
- Analyzes user requests using AI
- Determines intent, category, complexity
- Identifies required agents

#### PlanningService
- Creates execution plans
- Validates plan feasibility
- Manages plan state transitions

#### ExecutionService
- Executes plans through agents
- Monitors execution progress
- Handles execution failures

#### AgentDiscoveryService
- Discovers available agents
- Maintains agent capability mapping
- Handles agent health checks

## Phase 4: Orchestrator Refactoring

### Goals
- Split monolithic orchestrator into focused components
- Use dependency injection
- Maintain AI-native exploration capabilities

### File Structure Target
```
internal/
  orchestrator/
    ai_orchestrator.go (main coordinator)
    conversation_handler.go
    request_processor.go
    response_generator.go
    execution_coordinator.go
```

### Component Responsibilities

#### AIOrchestrator (Main)
- Coordinates between services
- Handles dependency injection
- Manages overall request flow

#### ConversationHandler
- Manages conversation lifecycle
- Handles user session state
- Integrates with ConversationService

#### RequestProcessor
- Processes incoming requests
- Coordinates analysis and planning
- Integrates with AnalysisService and PlanningService

#### ResponseGenerator
- Generates user responses
- Formats execution results
- Handles response optimization

#### ExecutionCoordinator
- Coordinates plan execution
- Manages agent interactions
- Handles execution monitoring

## Phase 5: Integration and Testing

### Goals
- Ensure all components work together
- Comprehensive end-to-end testing
- Performance optimization

### Testing Strategy
- Unit tests for each domain model
- Integration tests for repositories
- Service tests with mocked dependencies
- End-to-end tests with real AI calls

## Implementation Timeline

### Week 1: Domain Models
- [ ] ExecutionPlan domain model + tests
- [ ] Conversation domain model + tests  
- [ ] RequestAnalysis domain model + tests
- [ ] Domain validation and error handling

### Week 2: Repository Layer
- [ ] Repository interfaces
- [ ] Agent repository implementation + tests
- [ ] ExecutionPlan repository implementation + tests
- [ ] Conversation repository implementation + tests

### Week 3: Application Services
- [ ] Service interfaces
- [ ] ConversationService implementation + tests
- [ ] AnalysisService implementation + tests
- [ ] PlanningService implementation + tests
- [ ] ExecutionService implementation + tests

### Week 4: Orchestrator Refactoring
- [ ] Split orchestrator into components
- [ ] Implement dependency injection
- [ ] Integration testing
- [ ] Performance optimization

## TDD Protocol for Each Component

### Red Phase
1. Write failing test that defines expected behavior
2. Test should capture business rules and validation
3. Use real domain types, not primitives

### Green Phase
1. Write minimal code to make test pass
2. Focus on functionality, not optimization
3. Use proper error handling

### Refactor Phase
1. Clean up implementation
2. Add comprehensive validation
3. Optimize performance
4. Ensure security concerns are addressed

## Quality Gates

### Before Each Commit
- [ ] All tests pass (green)
- [ ] Code coverage > 80%
- [ ] No linting errors
- [ ] Security scan passes

### Before Each Phase
- [ ] Integration tests pass
- [ ] Performance benchmarks meet targets
- [ ] Documentation updated
- [ ] Code review completed

## Security and Governance

### AI Integration Points
- Input validation and sanitization
- Output validation and filtering
- Rate limiting and quota management
- Audit logging for all AI calls

### Graph Operations
- Schema validation for all node/edge operations
- Business rule enforcement
- Transaction management
- Data consistency checks

## Success Metrics

### Code Quality
- Cyclomatic complexity < 10 per function
- Test coverage > 80%
- No code duplication > 5 lines
- All linting rules pass

### Performance
- Request processing time < 2 seconds
- Memory usage < 100MB per request
- AI API calls optimized (current: ~3-4 per request)

### Maintainability
- Single responsibility per class/function
- Clear dependency injection
- Comprehensive error handling
- Self-documenting code

## Next Immediate Steps

1. **Complete Phase 1**: ExecutionPlan and Conversation domain models
2. **Start Phase 2**: Repository interfaces and implementations
3. **TDD Each Component**: Red-Green-Refactor cycle
4. **Integration Testing**: Ensure components work together
5. **Performance Validation**: Measure and optimize

This plan transforms the monolithic orchestrator into a maintainable, testable, and scalable architecture while preserving the AI-native exploration capabilities.
