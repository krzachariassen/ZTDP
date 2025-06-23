# AI Agent Orchestrator

A clean, domain-agnostic AI agent orchestrator built with graph-powered workflow management.

## Architecture

This orchestrator follows ZTDP's clean and simple internal structure:

```
orchestrator/
├── go.mod, go.sum          # Go modules
├── internal/               # Private packages
│   ├── orchestrator/       # Core workflow orchestration
│   │   ├── service.go      # Orchestration business logic
│   │   └── service_test.go # Tests
│   ├── registry/           # Agent registry
│   │   ├── service.go      # Registry business logic  
│   │   └── service_test.go # Tests
│   ├── graph/              # Graph database
│   │   ├── graph.go        # Graph interface & factory
│   │   ├── embedded_graph.go    # In-memory implementation
│   │   └── embedded_graph_test.go # Tests
│   └── types/              # Shared domain types
│       ├── agent.go        # Agent types
│       └── workflow.go     # Workflow types
└── pkg/                    # Legacy complex structure (to be removed)
```

## Key Design Principles

### 1. **Simple & Clean Structure**
- Follows ZTDP's `internal/` package organization
- Each domain has its own package (`orchestrator/`, `registry/`, `graph/`)
- Shared types in `types/` package
- No over-engineered layering or abstractions

### 2. **Domain-Agnostic**
- No ZTDP-specific logic in orchestrator
- Pure agent orchestration and workflow management
- Can be reused across different platforms

### 3. **Graph-Powered**
- All agents and workflows stored in graph backend
- Currently supports embedded (in-memory) graph
- Ready for Neo4j integration when needed

### 4. **Test-Driven**
- Each service has comprehensive tests
- Mock implementations for clean testing
- TDD approach: test first, implement second

## Testing

```bash
# Run all tests
go test ./internal/... -v

# Run specific package tests
go test ./internal/orchestrator -v
go test ./internal/registry -v
go test ./internal/graph -v
```

## Why This Structure?

- **Familiar**: Matches ZTDP's internal package structure
- **Simple**: No complex layering or abstractions
- **Testable**: Clean interfaces and mocking
- **Maintainable**: Small, focused packages
- **Reusable**: Domain-agnostic design
