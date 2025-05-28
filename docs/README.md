# ZTDP Documentation Index

Welcome to the Zero Touch Developer Platform (ZTDP) documentation. This index provides an organized overview of all available documentation.

## Getting Started

- **[Main README](../README.md)**: Project overview, quickstart guide, and API examples
- **[Architecture Overview](architecture.md)**: System design, components, and principles
- **[Policy System](policy-architecture.md)**: Graph-based governance and compliance
- **[Migration Guide](migration-guide.md)**: Recent architectural improvements and changes

## Core Concepts

### Graph-Native Operations
ZTDP models all platform entities as nodes and edges in a directed graph:
- **Nodes**: Applications, services, environments, policies, checks, resources
- **Edges**: Relationships like "owns", "deploys", "uses", "satisfies"
- **Benefits**: Dependency awareness, relationship queries, policy enforcement

### Contract-Driven Development
Instead of YAML manifests, ZTDP uses structured contracts:
- **Application Contracts**: Define applications and their metadata
- **Service Contracts**: Define services within applications
- **Environment Contracts**: Define deployment targets
- **Versioned Contracts**: Track changes and evolution over time

### Event-Driven Architecture
All operations generate structured events for observability:
- **Graph Events**: Node and edge operations
- **Policy Events**: Policy evaluations and enforcement decisions
- **API Events**: Request processing and responses
- **Integration**: Events can be consumed by external systems

## System Components

### Core Graph Engine (`internal/graph/`)
- **Graph Model**: Core data structures and operations
- **Graph Store**: Storage abstraction with multiple backends
- **Global Graph**: Singleton access pattern
- **Event Emitter**: Event emission for graph operations
- **Policy Helpers**: Policy-specific graph functionality

### Event System (`internal/events/`)
- **Event Bus**: Central event distribution
- **Graph Events**: Graph operation events
- **Policy Events**: Policy evaluation events
- **Event Emitter Registry**: Global event system setup

### Policy System (`internal/policies/`)
- **Policy Evaluator**: Core policy evaluation logic
- **Graph Validator**: Graph-based policy validation
- **Built-in Policies**: Standard governance policies
- **Custom Policies**: Extensible policy framework

### API Layer (`api/`)
- **Handlers**: HTTP request processing with policy enforcement
- **Server**: API routing and middleware
- **Integration**: Seamless policy enforcement at API level

## Policy System Deep Dive

### Policy Types
- **System Policies**: Built-in platform governance
- **Check Policies**: Automated validation requirements
- **Approval Policies**: Human approval workflows
- **Custom Policies**: Organization-specific rules

### Policy Enforcement
- **Automatic Enforcement**: Built into core graph operations
- **Transition-Level**: Policies attached to specific edge types
- **Event-Driven**: All policy decisions generate events
- **API Integration**: Consistent enforcement across all interfaces

### Policy Examples
- **Dev-Before-Prod**: Require deployment to dev before prod
- **Security Scanning**: Require security checks before deployment
- **Approval Workflows**: Human approval for production deployments
- **Resource Constraints**: Limit resource usage per environment

## Development Guide

### Testing
- **Unit Tests**: Component-level testing with mocks
- **Integration Tests**: End-to-end API testing
- **Policy Tests**: Policy enforcement validation
- **Test Coverage**: Comprehensive test suite

### Adding New Features
1. **Design**: Consider graph relationships and policy implications
2. **Contracts**: Define new contract types if needed
3. **Graph Operations**: Implement core graph logic
4. **Policy Integration**: Add policy enforcement points
5. **API Endpoints**: Create HTTP handlers with policy checks
6. **Events**: Ensure proper event emission
7. **Tests**: Add comprehensive test coverage
8. **Documentation**: Update relevant documentation

### Backend Extensions
- **Storage Backends**: Add new graph storage implementations
- **Event Backends**: Integrate with external event systems
- **Resource Providers**: Add new infrastructure providers
- **Policy Validators**: Implement custom policy logic

## Deployment

### Local Development
- **Docker Compose**: Redis and development services
- **Environment Variables**: Configuration management
- **Hot Reload**: Rapid development iteration

### Production (Planned)
- **Kubernetes**: Container orchestration
- **Redis Cluster**: Distributed storage
- **External Monitoring**: Event integration with monitoring systems
- **Security**: Authentication and authorization

## API Reference

### Core Endpoints
- **Applications**: `/v1/applications/*` - Application lifecycle management
- **Services**: `/v1/applications/{app}/services/*` - Service management
- **Environments**: `/v1/environments/*` - Environment management
- **Deployments**: Deploy services to environments with policy enforcement
- **Graph**: `/v1/graph` - Query the global graph state

### Policy Endpoints (Planned)
- **Policies**: Manage policy nodes and attachments
- **Checks**: Create and manage policy checks
- **Status**: Query policy satisfaction status

### Event Endpoints (Planned)
- **Event Stream**: Real-time event consumption
- **Event History**: Query historical events
- **Event Subscriptions**: Subscribe to specific event types

## Examples and Use Cases

### Basic Workflow
1. Create environments (dev, staging, prod)
2. Create application with services
3. Attach policies to enforce governance
4. Deploy services with automatic policy enforcement
5. Monitor via events and graph visualization

### Policy Scenarios
- **Progressive Deployment**: Enforce deployment order across environments
- **Security Gates**: Require security scans before production
- **Approval Workflows**: Human approval for sensitive operations
- **Resource Governance**: Enforce resource allocation policies

### Integration Patterns
- **CI/CD Integration**: Use APIs in deployment pipelines
- **Monitoring Integration**: Consume events for alerting
- **Compliance Reporting**: Use event history for audit trails
- **Custom Automation**: Build on top of the event system

## Troubleshooting

### Common Issues
- **Policy Violations**: Check policy attachment and satisfaction
- **Graph Inconsistencies**: Verify node and edge relationships
- **Event System**: Ensure event bus is properly configured
- **Backend Issues**: Check Redis connectivity and configuration

### Debugging Tools
- **Graph Visualization**: Use `/v1/graph` endpoint and HTML viewer
- **Event Logs**: Monitor event emission for troubleshooting
- **Policy Status**: Check policy satisfaction status
- **Test Demos**: Use `test/controlplane/graph_demo.go` for validation

### Performance Considerations
- **Graph Size**: Monitor graph node and edge counts
- **Event Volume**: Consider event processing capacity
- **Backend Performance**: Optimize Redis configuration
- **Policy Complexity**: Balance governance with performance

## Contributing

### Code Style
- **Go Standards**: Follow Go best practices and conventions
- **Clean Architecture**: Maintain separation of concerns
- **Event-Driven**: Ensure all operations emit appropriate events
- **Policy-Aware**: Consider policy implications for new features

### Documentation Standards
- **API Documentation**: Update Swagger annotations
- **Architecture Docs**: Keep architecture documentation current
- **Code Comments**: Document complex logic and design decisions
- **Examples**: Provide working examples for new features

### Testing Requirements
- **Test Coverage**: Maintain high test coverage
- **Integration Tests**: Add API-level tests for new endpoints
- **Policy Tests**: Validate policy enforcement for new features
- **Documentation Tests**: Ensure examples work as documented

---

For specific technical details, see the individual documentation files listed above. For getting started quickly, begin with the [main README](../README.md) and [architecture overview](architecture.md).
