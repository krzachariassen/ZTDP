# ZTDP Architecture Migration Guide

This document outlines the recent architectural improvements and migrations made to ZTDP for better clean architecture, event-driven design, and policy enforcement.

## Overview of Changes

The recent migration focused on three key areas:
1. **Event System Architecture**: Moving event emitters to proper architectural layers
2. **Policy Enforcement Integration**: Embedding policy checks into core graph operations
3. **Clean Architecture Compliance**: Resolving import cycles and improving separation of concerns

## Event System Migration

### Problem
The original event emitter was located in `api/handlers/graph_emitter.go`, which violated clean architecture principles by placing cross-cutting concerns in the API layer.

### Solution
Moved the global graph emitter to `internal/events/graph_emitter.go` for better separation of concerns:

```go
// Before: api/handlers/graph_emitter.go
var GlobalGraphEmitter *graph.GraphEventEmitter

// After: internal/events/graph_emitter.go  
var GlobalGraphEmitter interface{}
```

### Benefits
- **Clean Architecture**: Event system properly separated from API layer
- **Reusability**: Event emitters can be used by both API and control plane
- **Testability**: Event system can be tested independently
- **Import Cycle Resolution**: Eliminated circular dependencies

### Migration Steps
1. Created `internal/events/graph_emitter.go` with global emitter registry
2. Updated `cmd/api/main.go` to use `events.SetGraphEmitter`
3. Removed `api/handlers/graph_emitter.go`
4. Updated all references to use the new location

## Policy Enforcement Integration

### Problem
Policy enforcement was inconsistent between API operations and control plane operations, leading to potential policy bypass scenarios.

### Solution
Integrated policy enforcement directly into core graph operations:

```go
// Updated internal/graph/graph_model.go
func (g *Graph) AddEdge(fromID, toID, relType string) error {
    // ...existing validation...
    
    // Enforce policy for deploy edges
    if relType == "deploy" {
        if err := g.IsTransitionAllowed(fromID, toID, relType); err != nil {
            return err
        }
    }
    
    // ...rest of method...
}
```

### Benefits
- **Consistent Enforcement**: Same policy logic applies everywhere
- **Comprehensive Coverage**: No way to bypass policy checks
- **Centralized Logic**: Policy enforcement in one place
- **Event Integration**: Policy decisions generate events automatically

### Migration Steps
1. Enhanced `Graph.AddEdge` method with policy enforcement
2. Updated `GlobalGraph.AddEdge` to delegate to core `Graph.AddEdge`
3. Improved error handling to return HTTP 403 for policy violations
4. Added policy setup helpers for test scenarios

## Import Cycle Resolution

### Problem
Direct dependencies between `internal/graph` and `internal/events` packages created import cycles.

### Solution
Used `interface{}` types and runtime type checking to break cycles:

```go
// Before: Direct type dependency
var GlobalGraphEmitter *graph.GraphEventEmitter

// After: Interface-based approach
var GlobalGraphEmitter interface{}

func SetGraphEmitter(emitter interface{}) {
    GlobalGraphEmitter = emitter
}
```

### Benefits
- **Dependency Inversion**: Reduced coupling between packages
- **Flexibility**: Easier to swap implementations
- **Testability**: Better mocking capabilities
- **Maintainability**: Clearer package boundaries

## Error Handling Improvements

### Problem
Tautological conditions and inconsistent error responses were present in API handlers.

### Solution
Fixed error handling patterns and standardized policy violation responses:

```go
// Before: Inconsistent error responses
return c.JSON(http.StatusInternalServerError, map[string]string{
    "error": err.Error(),
})

// After: Proper policy violation handling
if errors.Is(err, &graph.PolicyNotSatisfiedError{}) {
    return c.JSON(http.StatusForbidden, map[string]string{
        "error":   "Policy not satisfied",
        "message": err.Error(),
    })
}
```

### Benefits
- **Consistent Responses**: Standardized error codes for policy violations
- **Better UX**: Clear error messages for policy failures
- **Proper HTTP Semantics**: 403 Forbidden for policy violations vs 500 for server errors

## Test Infrastructure Enhancements

### Problem
Policy enforcement tests were failing because policies weren't properly attached to transitions in test scenarios.

### Solution
Added comprehensive policy setup helpers for testing:

```go
// Added to test/api/api_test.go
func attachMustDeployToDevBeforeProdPolicy() {
    // Create policy node
    policyNode := &graph.Node{
        ID:   "policy-dev-before-prod",
        Kind: graph.KindPolicy,
        // ...policy configuration...
    }
    
    // Attach to relevant transitions
    globalGraph.Graph.AttachPolicyToTransition(
        serviceVersionID, "prod", graph.EdgeTypeDeploy, policyNode.ID,
    )
}
```

### Benefits
- **Realistic Testing**: Tests now use actual policy enforcement
- **Comprehensive Coverage**: All policy scenarios covered
- **Maintainable Tests**: Clear policy setup patterns
- **CI/CD Reliability**: Consistent test results

## Performance Optimizations

### Changes Made
1. **Efficient Policy Lookups**: Improved policy discovery algorithms
2. **Event Batching**: Reduced event emission overhead
3. **Graph Operations**: Optimized edge creation and validation
4. **Memory Usage**: Better resource management in event system

### Performance Benefits
- **Faster API Responses**: Reduced policy evaluation overhead
- **Lower Memory Usage**: More efficient event handling
- **Better Scalability**: Optimized for larger graphs
- **Reduced Latency**: Streamlined operation pipelines

## Migration Checklist

If you're working with ZTDP code, ensure your setup includes these changes:

### Code Updates
- [ ] Import `events.SetGraphEmitter` instead of `handlers.SetGraphEmitter`
- [ ] Use `internal/events/graph_emitter.go` for event system access
- [ ] Ensure policy checks are integrated into all graph operations
- [ ] Update error handling to return appropriate HTTP status codes

### Testing Updates
- [ ] Use policy setup helpers in test scenarios
- [ ] Verify policy enforcement in integration tests
- [ ] Test event emission for all operations
- [ ] Validate error responses for policy violations

### Documentation Updates
- [ ] Update API documentation with policy enforcement details
- [ ] Document event system architecture changes
- [ ] Update examples to reflect new patterns
- [ ] Review and update integration guides

## Best Practices Going Forward

### Event System
1. **Consistent Emission**: Ensure all operations emit appropriate events
2. **Structured Events**: Follow the standard event schema
3. **Error Handling**: Emit events for both success and failure cases
4. **Performance**: Consider event volume and processing capacity

### Policy System
1. **Early Validation**: Check policies before making changes
2. **Clear Errors**: Provide detailed policy violation messages
3. **Event Integration**: Emit events for all policy decisions
4. **Test Coverage**: Include policy scenarios in all tests

### Clean Architecture
1. **Dependency Direction**: Core business logic should not depend on external layers
2. **Interface Segregation**: Use interfaces to break dependencies
3. **Single Responsibility**: Each package should have a clear purpose
4. **Testability**: Design for easy unit testing and mocking

## Rollback Procedures

If issues arise, the following rollback steps can be performed:

### Event System Rollback
1. Revert `internal/events/graph_emitter.go` changes
2. Restore `api/handlers/graph_emitter.go`
3. Update imports in `cmd/api/main.go`
4. Remove event system setup code

### Policy Enforcement Rollback
1. Remove policy checks from `Graph.AddEdge`
2. Restore original `GlobalGraph.AddEdge` implementation
3. Remove policy setup from tests
4. Revert error handling changes

### Testing Rollback
1. Remove policy setup helpers
2. Restore original test scenarios
3. Remove policy-specific test cases
4. Update test expectations

## Future Improvements

### Planned Enhancements
1. **External Event Integration**: Forward events to external systems
2. **Policy API**: REST endpoints for managing policies
3. **Event Subscriptions**: Real-time event consumption
4. **Performance Monitoring**: Event and policy performance metrics

### Architecture Evolution
1. **Microservices**: Potential service decomposition
2. **Event Sourcing**: Consider event sourcing patterns
3. **CQRS**: Command Query Responsibility Segregation
4. **Distributed Events**: Cross-service event propagation

This migration represents a significant improvement in ZTDP's architecture, providing better separation of concerns, comprehensive policy enforcement, and robust event-driven capabilities.
