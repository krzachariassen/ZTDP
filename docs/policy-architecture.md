# Zero Trust Developer Platform (ZTDP) - Policy Architecture

## Overview

ZTDP's policy architecture provides comprehensive governance of application deployments and transitions between nodes in the platform graph. The system enforces security, compliance, and operational constraints through a graph-based policy model with automatic enforcement and event-driven visibility.

## Architecture Principles

* **Clean Architecture**: Policy enforcement is integrated at the core graph operations level, ensuring consistent enforcement across API and control plane
* **Event-Driven Design**: All policy operations generate structured events for monitoring, auditing, and integration
* **Graph-Native**: Policies are first-class citizens in the graph, enabling dynamic and contextual enforcement
* **Automated Enforcement**: Policy checks happen automatically during graph operations—no manual intervention required

## Graph-Based Policy Model

The ZTDP policy system uses a graph-based approach where policies are represented as nodes in the directed graph, with the following characteristics:

* **Policy Nodes**: First-class entities in the graph with the `kind: policy` attribute
* **Policy Attachment**: Policies are attached to specific transitions (edges) between nodes
* **Policy Satisfaction**: Checks and approvals can satisfy policies, enabling transitions
* **Policy Enforcement**: Enforced when attempting to create edges or transitions between nodes

## Policy Enforcement Process

ZTDP's policy enforcement is built into the core graph operations for comprehensive coverage:

### Automatic Enforcement in Graph Operations

1. **Core Graph Integration**: Policy enforcement happens in the `Graph.AddEdge` method, ensuring all edge creation goes through policy validation
2. **Deploy Edge Enforcement**: Special enforcement for "deploy" edges ensures deployment policies are checked before any deployment transition
3. **Consistent Enforcement**: Both API requests and control plane operations use the same enforcement mechanisms

### Policy Evaluation Flow

1. When a transition is requested (e.g., deploying a service to an environment), the system:
   - Calls `IsTransitionAllowed(fromID, toID, edgeType)` before creating the edge
   - Finds all policies attached to the specific transition
   - Checks if each policy has satisfying checks or approvals
   - Returns a `PolicyNotSatisfiedError` if any policy is not satisfied

2. **Error Handling**: Policy violations return HTTP 403 Forbidden responses with detailed error messages

3. **Event Generation**: All policy evaluations generate events for monitoring and audit trails

### Policy Validation Architecture

The system uses a modern graph-based policy validation approach that provides powerful and flexible policy enforcement:

```go
// Policies are enforced directly by the graph through transition validation
err := graph.IsTransitionAllowed(fromNodeID, toNodeID, edgeType)
if err != nil {
    // Handle policy violation
    return err
}

// Policy creation and management through the PolicyEvaluator
evaluator := policies.NewPolicyEvaluator(graphStore, environment)
policyNode, err := evaluator.CreatePolicyNode(name, description, policyType, parameters)
```

## Policy Types

ZTDP supports several policy types for comprehensive governance:

* **System Policies**: Built-in platform policies for core governance (e.g., dev-before-prod)
* **Check Policies**: Require automated checks to pass before transitions
* **Approval Policies**: Require explicit human approvals before transitions
* **Custom Policies**: Extensible policy framework for organization-specific requirements

### Policy Status Handling

Policies and their associated checks have status tracking:
* **Policy Status**: `active`, `inactive`, `deprecated`
* **Check Status**: `pending`, `running`, `succeeded`, `failed`
* **Transition Status**: `allowed`, `blocked`, `conditional`

## Example: Dev-Before-Prod Policy

This policy ensures services are deployed to the development environment before production:

```go
// Policy node in the graph
policyNode := &graph.Node{
    ID:   "policy-dev-before-prod",
    Kind: graph.KindPolicy,
    Metadata: map[string]interface{}{
        "name":        "Must Deploy To Dev Before Prod", 
        "description": "Requires a service version to be deployed to dev before it can be deployed to prod",
        "type":        graph.PolicyTypeSystem,
        "status":      "active",
    },
    Spec: map[string]interface{}{
        "sourceKind":      graph.KindServiceVersion,
        "targetKind":      graph.KindEnvironment,
        "targetID":        "prod",
        "requiredPathIDs": []string{"dev"},
    },
}

// Attach policy to a transition
graph.AttachPolicyToTransition(serviceVersionID, "prod", graph.EdgeTypeDeploy, policyNode.ID)
```

## Usage

### Creating a Policy

```go
// Create policy node
policyNode := &graph.Node{
    ID:   "custom-policy-id",
    Kind: graph.KindPolicy,
    Metadata: map[string]interface{}{
        "name":        "Custom Policy Name",
        "description": "Description of what the policy enforces",
        "type":        graph.PolicyTypeSystem,
        "status":      "active",
    },
    Spec: map[string]interface{}{
        // Policy-specific configuration
        "sourceKind": "service_version",
        "targetKind": "environment",
        // Additional parameters as needed
    },
}
g.AddNode(policyNode)
```

### Attaching a Policy to a Transition

```go
// Attach policy to transition between nodes
g.AttachPolicyToTransition(fromNodeID, toNodeID, edgeType, policyID)
```

### Satisfying a Policy with Checks

```go
// Create a check node
checkNode := &graph.Node{
    ID:   "check-id",
    Kind: graph.KindCheck,
    Metadata: map[string]interface{}{
        "name":   "Check Name",
        "type":   "check-type",
        "status": graph.CheckStatusSucceeded,
    },
    Spec: map[string]interface{}{
        // Check-specific parameters
    },
}
g.AddNode(checkNode)

// Link check to policy with "satisfies" relationship
g.AddEdge(checkNode.ID, policyID, graph.EdgeTypeSatisfies)
```

### Checking Policy Status

```go
// Check if a transition is allowed
err := g.IsTransitionAllowed(fromNodeID, toNodeID, edgeType)
if err != nil {
    // Handle policy violation
    log.Printf("Transition blocked by policy: %v", err)
    return err
}

// Check if a specific policy is satisfied
satisfied, err := g.IsPolicySatisfied(policyID)
if err != nil {
    return err
}
```

## API Integration

The policy system is seamlessly integrated with the ZTDP API:

### Automatic Policy Enforcement

All deployment and transition operations automatically enforce policies:

```bash
# This API call is checked against policies before execution
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions/1.0.0/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "prod"
  }'
```

### Policy Violation Responses

When policies are not satisfied, the API returns detailed error information:

```json
{
  "error": "Policy not satisfied",
  "message": "Policy 'policy-dev-before-prod' requires deployment to dev environment first",
  "details": {
    "policy_id": "policy-dev-before-prod",
    "required_checks": ["check-dev-deployment-checkout"],
    "missing_satisfactions": ["dev-deployment-verification"]
  }
}
```

### Event-Driven Monitoring

Policy operations generate comprehensive events:

* **Policy Check Events**: When policies are evaluated
* **Policy Results**: Success/failure of policy evaluations
* **Transition Attempts**: When transitions are requested
* **Transition Results**: Whether transitions are approved or rejected

### Integration Benefits

* **Consistent Enforcement**: Same policy logic applies to API and control plane operations
* **Real-time Feedback**: Immediate policy violation notifications
* **Audit Trail**: Complete event history for compliance reporting
* **Transparent Governance**: Clear visibility into policy requirements and status

For examples of policy usage in tests and real-world scenarios, see:
- `test/api/api_test.go` for API integration examples
- `test/controlplane/graph_demo.go` for policy setup examples
- `examples/policy_usage.go` for programmatic usage patterns

## Event-Driven Architecture

ZTDP's policy system is built on a comprehensive event-driven architecture that provides visibility, auditability, and integration capabilities.

### Event System Components

#### Graph Event Service
Located in `internal/events/graph_events.go`, this service emits events for all graph operations:

```go
// Node operations
EmitNodeAdded(env, nodeID, kind, metadata)
EmitNodeUpdated(env, nodeID, kind, metadata) 
EmitNodeRemoved(env, nodeID)

// Edge operations
EmitEdgeAdded(env, fromID, toID, edgeType)
EmitEdgeRemoved(env, fromID, toID, edgeType)
```

#### Policy Event Service
Located in `internal/events/policy_events.go`, this service handles policy-specific events:

```go
// Policy evaluation events
EmitPolicyCheck(policyID, mutation, context)
EmitPolicyCheckResult(policyID, result, reason, mutation)

// Transition events
EmitTransitionAttempt(fromID, toID, edgeType, user)
EmitTransitionResult(fromID, toID, edgeType, user, approved, reason)
```

### Graph Event Emitter Architecture

The event emitter architecture provides clean separation of concerns:

#### Event Emitter Pattern
The `GraphEventEmitter` wraps the core `GraphStore` to add event emission capabilities:

```go
// Located in internal/graph/graph_event_emitter.go
type GraphEventEmitter struct {
    *GraphStore
    eventService *events.GraphEventService
}
```

#### Global Event System Setup
The event system is initialized in `cmd/api/main.go`:

```go
// Create event services
eventBus := events.NewEventBus()
policyEvents := events.NewPolicyEventService(eventBus, "api-server")
graphEvents := events.NewGraphEventService(eventBus, "api-server")

// Set up event system with handlers
handlers.SetupEventSystem(eventBus, policyEvents, graphEvents)
```

#### Event Bus Integration
Events are published through a centralized event bus that supports:
- **Multiple Subscribers**: Multiple handlers can subscribe to the same event types
- **Asynchronous Processing**: Events are processed without blocking operations
- **Structured Events**: All events follow a consistent schema with metadata

### Event Types and Structure

All events follow a consistent structure defined in `internal/events/events.go`:

```go
type Event struct {
    Type      EventType              `json:"type"`
    Source    string                 `json:"source"`
    Subject   string                 `json:"subject"`
    Action    string                 `json:"action,omitempty"`
    Status    string                 `json:"status,omitempty"`
    Timestamp int64                  `json:"timestamp"`
    ID        string                 `json:"id"`
    Payload   map[string]interface{} `json:"payload,omitempty"`
}
```

#### Standard Event Types
- **Graph Events**: `graph.node.added`, `graph.node.updated`, `graph.edge.added`, etc.
- **Policy Events**: `policy.check`, `policy.check.result`, `transition.attempt`, etc.

### Benefits of Event-Driven Architecture

1. **Observability**: Real-time visibility into all platform operations
2. **Auditability**: Complete event trail for compliance and debugging
3. **Integration**: Events can be consumed by external monitoring and automation systems
4. **Decoupling**: Event-driven design enables loose coupling between components
5. **Extensibility**: New event handlers can be added without modifying core logic

### Event Architecture Migration

Recent architectural improvements include:

- **Moved Graph Emitter**: Relocated from `api/handlers/graph_emitter.go` to `internal/events/graph_emitter.go` for better separation of concerns
- **Import Cycle Resolution**: Used `interface{}` types to resolve circular dependencies between packages
- **Centralized Event Setup**: Consolidated event system initialization in the main application entry point
- **Clean Integration**: Event emission is cleanly integrated with core graph operations without coupling
