# Zero Touch Developer Platform (ZTDP)

**A next-generation internal developer platform that models your entire application landscape as a live graph.**

ZTDP eliminates the friction between intent and outcome—deploy applications, enforce policies, and understand dependencies through a single, unified platform. No more YAML wrestling, no more manual gates, no more hidden dependencies.

---

## 🚀 The Problem We Solve

Modern platform engineering faces the same challenges everywhere:

- **Hidden Dependencies**: Applications break because of invisible relationships
- **Manual Bottlenecks**: Every deployment requires manual approvals and reviews  
- **Policy Confusion**: Security and compliance policies are enforced inconsistently
- **Tool Sprawl**: Different tools for different environments create complexity
- **No Visibility**: Teams can't understand what's happening across their platform

**ZTDP solves this by modeling everything as a graph.** Applications, services, environments, policies, and resources are nodes. Dependencies, deployments, and policy relationships are edges. The result? Complete visibility and zero-touch automation.

---

## ✨ What Makes ZTDP Different

- **🧠 Graph-Native**: Everything is a node in a live, queryable graph—no more hidden dependencies
- **🔒 Policy-First**: Security and compliance policies are enforced automatically at every transition  
- **⚡ Zero Touch**: From intent to outcome with no manual steps or YAML files
- **📊 Real-Time**: Every change emits events for complete observability and integration
- **🔧 Composable**: Plug in new infrastructure types without changing core platform logic
- **🤖 AI-Ready**: Designed for both human and AI operators with deterministic, auditable operations

---

## 🎯 Quick Demo

```bash
# Start the platform
docker-compose up -d
export ZTDP_GRAPH_BACKEND=redis REDIS_HOST=localhost:6379 REDIS_PASSWORD=BVogb1sEPqA
go run ./cmd/api/main.go

# Create an application
curl -X POST http://localhost:8080/v1/applications \
  -H "Content-Type: application/json" \
  -d '{"metadata": {"name": "my-app", "owner": "team-x"}}'

# Deploy the entire application 
curl -X POST http://localhost:8080/v1/applications/my-app/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}'

# View the live graph
open http://localhost:8080/static/graph-modern.html
```

**That's it.** No YAML, no manual steps, no complexity.

---

## 🏗️ How It Works

1. **Model Everything as a Graph**: Applications, services, environments, and policies are nodes with relationship edges
2. **Express Intent via JSON**: Submit structured contracts declaring what you want—ZTDP figures out the rest  
3. **Plan & Enforce**: The planner computes execution order while enforcing policies automatically
4. **Execute & Observe**: Every operation emits events for complete observability and integration

---

## 🗂️ Project Structure

```text
ZTDP/
├── api/                      # API server logic (handlers, routes)
│   ├── handlers/             # HTTP handler logic
│   └── server/               # API routing setup
├── cmd/                      # Entrypoint: main.go
├── docs/                     # Platform documentation
│   ├── README.md             # Documentation index and guide
│   ├── architecture.md      # System architecture overview
│   ├── migration-guide.md   # Recent architectural improvements
│   └── policy-architecture.md # Policy system documentation
├── internal/                 # Core architecture
│   ├── contracts/            # Contract types: Application, Service, etc.
│   ├── events/               # Event system and graph emitters
│   ├── graph/                # Graph engine, backend, resolver, registry
│   ├── policies/             # Policy engine for governance
│   └── state/                # State store abstraction (future)
├── rps/                      # Resource Providers (Kubernetes, Postgres, etc.)
├── test/
│   ├── api/                  # End-to-end API tests
│   └── controlplane/         # Control plane validation demos
├── charts/                   # Helm charts (e.g., redis)
├── docker-compose.yml        # Local services: Redis
├── .env                      # Environment configuration
└── go.mod / go.sum           # Go dependencies
```

---

## ⚙️ Getting Started (Local Dev)

ZTDP is designed for rapid iteration and local hacking. You can be up and running in minutes.

### Prerequisites

- Docker & Docker Compose
- Go 1.23+

### Quickstart

```bash
# Start Redis for backend storage
docker-compose up -d

# Set environment variables
export ZTDP_GRAPH_BACKEND=redis
export REDIS_HOST=localhost:6379
export REDIS_PASSWORD=BVogb1sEPqA  # matches docker-compose.yaml

# Run a control plane demo (optional, to pre-populate sample data)
go run ./test/controlplane/graph_demo.go

# Or start the API server
go run ./cmd/api/main.go
```

---

## 🧪 Testing & Quality

We live and breathe TDD and API-first development:

- Every feature starts with a test.
- Logic and contracts are covered (`go test ./...`).
- APIs are tested with HTTP assertions.
- Redis-backed graph is tested for both in-memory and persistence.

```bash
# Run all tests
go test ./...
```

---

## 🌐 API Endpoints

| Method | Endpoint                                                        | Purpose                                         |
|--------|-----------------------------------------------------------------|-------------------------------------------------|
| POST   | `/v1/applications`                                              | Submit new application                          |
| GET    | `/v1/applications`                                              | List all applications                           |
| GET    | `/v1/applications/{app}`                                        | Get a specific application                      |
| PUT    | `/v1/applications/{app}`                                        | Update an application                           |
| GET    | `/v1/applications/schema`                                       | Get application contract schema                 |
| POST   | `/v1/applications/{app}/services`                               | Add a service to an application                 |
| GET    | `/v1/applications/{app}/services`                               | List services for an application                |
| GET    | `/v1/applications/{app}/services/{service}`                     | Get a specific service                          |
| GET    | `/v1/applications/{app}/services/schema`                        | Get service contract schema                     |
| POST   | `/v1/environments`                                              | Create a new environment                        |
| GET    | `/v1/environments`                                              | List all environments                           |
| POST   | `/v1/applications/{app}/environments/{env}/allowed`             | Allow an application to deploy to an environment|
| GET    | `/v1/applications/{app}/environments/allowed`                   | List allowed environments for an application    |
| POST   | `/v1/applications/{app}/deploy`                                 | **🆕 Deploy entire application to environment** |
| GET    | `/v1/applications/{app}/plan`                                   | Get deployment plan for application             |
| POST   | `/v1/applications/{app}/plan/apply/{env}`                       | Apply deployment plan to environment            |
| POST   | `/v1/applications/{app}/services/{service}/versions/{version}/deploy` | Deploy individual service version to environment |
| GET    | `/v1/environments/{env}/deployments`                              | List deployments in an environment (uses 'deploy' edges)              |
| GET    | `/v1/graph`                                                     | View current global DAG                         |
| GET    | `/v1/logs/stream`                                               | Real-time log streaming                         |
| GET    | `/v1/status`                                                    | Platform status                                 |
| GET    | `/v1/healthz`                                                   | Health check                                    |

- **Swagger/OpenAPI docs:** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## 📚 Documentation

For detailed technical information, architecture deep-dives, and API documentation:

- **[System Architecture](docs/architecture.md)**: Complete system overview and component design
- **[Policy System](docs/policy-architecture.md)**: Graph-based policy enforcement and compliance
- **[Logging Architecture](docs/logging-architecture.md)**: Real-time event streaming and observability
- **[Documentation Index](docs/README.md)**: Complete documentation guide

**API Documentation:** [Swagger/OpenAPI](http://localhost:8080/swagger/index.html) (when running locally)

---

## 💡 Why Contribute?

ZTDP is for builders, dreamers, and those who want to change how platforms are delivered.  
Whether you’re into Go, distributed systems, or just want to see what a zero-touch platform feels like—jump in, hack, and help us shape the future.

---

## 📌 License

TBD — Project is in private development. License terms will be clarified before any public release.

---

> **Ready to build the future?**
>
> Clone, run, and let’s go! 🚀