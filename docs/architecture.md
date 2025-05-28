# ZTDP Architecture Overview

## System Architecture

ZTDP follows clean architecture principles with clear separation of concerns and event-driven design patterns.

### Core Components

#### Graph Engine (`internal/graph/`)
The heart of ZTDP, providing a graph-based data model for all platform entities:

- **Graph Model** (`graph_model.go`): Core graph operations with integrated policy enforcement
- **Graph Store** (`graph_store.go`): Storage abstraction with backend implementations
- **Global Graph** (`graph_global.go`): Singleton pattern for unified graph access
- **Event Emitter** (`graph_event_emitter.go`): Event emission wrapper for graph operations
- **Policy Helpers** (`graph_policy_helpers.go`): Policy-specific graph operations

#### Event System (`internal/events/`)
Comprehensive event-driven architecture for observability and integration:

- **Event Bus** (`events.go`): Central event distribution system
- **Graph Events** (`graph_events.go`): Graph operation event emission
- **Policy Events** (`policy_events.go`): Policy evaluation and enforcement events
- **Graph Emitter** (`graph_emitter.go`): Global event emitter registry

#### Policy System (`internal/policies/`)
Graph-based governance and compliance enforcement:

- **Policy Evaluator** (`policy_evaluator.go`): Core policy evaluation logic
- **Graph Validator** (`graph_validator.go`): Graph-based policy validation

#### API Layer (`api/`)
RESTful API with automatic policy enforcement:

- **Handlers** (`handlers/`): HTTP request handlers with integrated policy checks
- **Server** (`server/`): API routing and middleware setup

### Architecture Principles

#### Clean Architecture
- **Dependency Inversion**: Core business logic doesn't depend on external frameworks
- **Separation of Concerns**: Clear boundaries between layers and responsibilities
- **Testability**: All components are unit testable in isolation

#### Event-Driven Design
- **Observability**: All operations emit structured events for monitoring
- **Auditability**: Complete event trail for compliance and debugging
- **Integration**: Events enable loose coupling and external system integration

#### Graph-Native Operations
- **Unified Model**: All platform entities represented as graph nodes and edges
- **Relationship Awareness**: Operations understand and respect entity relationships
- **Policy Integration**: Policies are first-class citizens in the graph model

### Data Flow

#### Request Processing Flow
1. **API Request** → Handler receives HTTP request
2. **Policy Check** → Automatic policy validation before operation
3. **Graph Operation** → Core graph modification with event emission
4. **Event Publication** → Events published to event bus
5. **Response** → HTTP response with operation result

#### Policy Enforcement Flow
1. **Operation Request** → Graph operation attempted
2. **Policy Discovery** → Find policies attached to the transition
3. **Policy Evaluation** → Check if policies are satisfied
4. **Decision** → Allow or deny operation based on policy results
5. **Event Emission** → Policy evaluation results published as events

### Backend Support

#### Storage Backends
- **Memory Backend**: In-memory storage for development and testing
- **Redis Backend**: Persistent storage with Redis for production use
- **Pluggable Interface**: Easy to add new storage backends

#### Event Backends
- **In-Memory Bus**: Default event bus for development
- **External Integration**: Events can be forwarded to external systems

### Testing Architecture

#### Test Coverage
- **Unit Tests**: Individual component testing with mocks
- **Integration Tests**: API-level testing with real graph operations
- **Policy Tests**: Comprehensive policy enforcement validation

#### Test Organization
- `test/api/`: End-to-end API testing
- `test/controlplane/`: Control plane validation and demos
- `*_test.go`: Unit tests co-located with source code

### Deployment Architecture

#### Development
- **Docker Compose**: Local Redis and development services
- **Hot Reload**: Rapid development iteration
- **Test Data**: Demo data generation for local testing

#### Production (Planned)
- **Kubernetes**: Container orchestration
- **Redis Cluster**: Distributed storage backend
- **External Events**: Integration with monitoring and alerting systems

### Security Model

#### Authentication (Planned)
- **API Keys**: Service-to-service authentication
- **JWT Tokens**: User authentication
- **RBAC**: Role-based access control

#### Authorization
- **Policy-Based**: All operations subject to policy evaluation
- **Graph-Level**: Permissions tied to graph entities and relationships
- **Audit Trail**: Complete event history for security monitoring

### Observability

#### Event Emission
- **Structured Events**: Consistent event schema across all operations
- **Real-time**: Events emitted synchronously with operations
- **Comprehensive**: Graph operations, policy evaluations, and API requests

#### Monitoring Integration
- **Metrics Export**: Events can be converted to metrics
- **Log Integration**: Events can be forwarded to logging systems
- **Alerting**: Policy violations and system events can trigger alerts

### Extensibility

#### Plugin Architecture
- **Resource Providers**: Pluggable infrastructure providers
- **Policy Validators**: Custom policy evaluation logic
- **Event Handlers**: Custom event processing and forwarding

#### API Extension
- **Contract Types**: New contract types can be added
- **Graph Entities**: New node and edge types supported
- **Custom Operations**: Domain-specific operations can be added

This architecture ensures ZTDP is scalable, maintainable, and extensible while providing comprehensive governance and observability capabilities.
