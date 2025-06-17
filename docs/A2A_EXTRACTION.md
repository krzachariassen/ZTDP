# Agent-to-Agent Framework Extraction Plan

## Current State Analysis

### What We Have That Works
- **DeploymentAgent** ✅ - Well-structured, follows AgentInterface, handles events properly
- **AgentInterface** ✅ - Clean interface definition in `/internal/agents/interface.go`
- **AgentRegistry** ✅ - Working registry in `/internal/agents/registry.go`
- **Event System** ✅ - Functional event bus for agent communication

### Current Problems
1. **Code Duplication**: Every agent reimplements the same patterns
2. **Mixed Concerns**: Agent registry and agent framework are in same package
3. **Boilerplate**: Lots of repetitive code for event handling, registration, etc.
4. **No Reusable Framework**: Each agent written from scratch

### Architecture Goals
Following our principles:
- **TDD**: Write tests first to define behavior
- **Clean Architecture**: Separate concerns properly
- **KISS**: Simple, reusable patterns

## Plan

### Phase 1: Analysis & Separation (CURRENT)
**Goal**: Understand current patterns and separate concerns

#### Step 1.1: Analyze DeploymentAgent Patterns ✅
- [x] Event handling patterns
- [x] Registration patterns  
- [x] Capability definition patterns
- [x] Lifecycle management patterns

#### Step 1.2: Create Package Structure
```
/internal/agentRegistry/    <- Move agent registry here (infrastructure)
/internal/agentFramework/   <- New reusable agent framework (domain patterns)
```

### Phase 2: Test-Driven Framework Creation
**Goal**: Create reusable framework based on DeploymentAgent patterns

#### Step 2.1: Write Framework Tests
Create `/internal/agentFramework/framework_test.go` that tests:
- Agent creation with auto-registration
- Event subscription based on capabilities
- Intent-based event routing
- Error handling and response patterns
- Logging consistency

#### Step 2.2: Extract Common Patterns
From DeploymentAgent, extract:
- Base agent structure
- Auto-registration logic
- Event subscription logic
- Response creation helpers
- Logging setup

#### Step 2.3: Create Framework Interface
```go
type BaseAgent interface {
    agents.AgentInterface
    
    // Framework-specific helpers
    CreateResponse(message string, payload map[string]interface{}, correlationEvent *events.Event) *events.Event
    CreateErrorResponse(correlationEvent *events.Event, errorMessage string) *events.Event
    SubscribeToCapabilities() error
    GetLogger() *logging.Logger
}
```

### Phase 3: Framework Implementation
**Goal**: Implement the reusable framework

#### Step 3.1: Base Agent Implementation
- Common fields (ID, logger, eventBus, registry, etc.)
- Auto-registration logic
- Event subscription based on capabilities
- Response helpers

#### Step 3.2: Agent Builder Pattern
```go
type AgentBuilder struct {
    id           string
    agentType    string
    capabilities []agents.AgentCapability
    eventHandler func(ctx context.Context, event *events.Event) (*events.Event, error)
}

func NewAgent(id string) *AgentBuilder
func (b *AgentBuilder) WithCapabilities(caps []agents.AgentCapability) *AgentBuilder
func (b *AgentBuilder) WithEventHandler(handler func...) *AgentBuilder
func (b *AgentBuilder) Build(deps AgentDependencies) (agents.AgentInterface, error)
```

### Phase 4: Migration & Validation
**Goal**: Migrate DeploymentAgent to use framework

#### Step 4.1: Create New DeploymentAgent Using Framework
- Use framework builder
- Keep same capabilities and behavior
- Verify all tests pass

#### Step 4.2: Side-by-Side Comparison
- Old DeploymentAgent vs New DeploymentAgent
- Ensure identical behavior
- Measure code reduction

#### Step 4.3: Replace Original
- Once validated, replace original DeploymentAgent
- Update imports and dependencies

### Phase 5: Framework Adoption
**Goal**: Update other agents to use framework

#### Step 5.1: Migrate ApplicationAgent
#### Step 5.2: Migrate PolicyAgent  
#### Step 5.3: Migrate ReleaseAgent

## Success Criteria

### ✅ Framework Success Indicators
1. **Reduced Boilerplate**: 70%+ code reduction in agent implementations
2. **Consistent Patterns**: All agents use same logging, error handling, events
3. **Easy Agent Creation**: New agent can be created in <50 lines of code
4. **Zero Behavior Change**: Existing functionality unchanged
5. **Test Coverage**: 90%+ test coverage for framework

### ✅ Clean Architecture Validation
1. **Separation of Concerns**: Registry separate from framework
2. **Dependency Injection**: Framework doesn't depend on specific implementations
3. **Interface Compliance**: All agents implement AgentInterface
4. **Domain Logic Separation**: Framework handles infrastructure, agents handle domain

## Implementation Notes

### TDD Approach
1. **Red**: Write failing test for desired framework behavior
2. **Green**: Implement minimal code to make test pass
3. **Refactor**: Clean up implementation while keeping tests green
4. **Repeat**: Add next feature using same cycle

### Clean Architecture Principles
- **Framework = Infrastructure**: Handles registration, events, logging
- **Agents = Domain Logic**: Handle business rules and validation
- **No Framework Dependencies**: Agents shouldn't depend on framework internals

### KISS Principles
- **Simple Builder Pattern**: Easy agent creation
- **Minimal Interfaces**: Only what's needed
- **Clear Responsibilities**: Framework vs agent concerns well-defined
- **No Over-Engineering**: Start simple, add complexity only when needed

## Next Actions

1. **Create `/internal/agentRegistry/` package** - Move registry code
2. **Create `/internal/agentFramework/` package** - New framework home
3. **Write framework tests** - Define desired behavior first
4. **Implement framework** - Make tests pass
5. **Migrate DeploymentAgent** - Validate framework works
