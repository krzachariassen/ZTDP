# Zero Touch Developer Platform (ZTDP)

ZTDP is a next-generation internal developer platform built around **declarative contracts**, **graph-based orchestration**, and **extensible resource providers**. It enables zero-touch infrastructure and application delivery with APIs, automation, and AI-native architecture at its core.

## 🧠 Key Concepts

- **Contract-Driven**: Define apps and services declaratively, no YAML or Helm.
- **Graph Engine**: Converts contracts into a DAG (Directed Acyclic Graph) of platform intent.
- **Environment-Aware**: Apply the same graph to multiple environments (dev, qa, prod).
- **Resource Providers**: Extensible backend modules (e.g. Kubernetes, Postgres).
- **API-First + TDD**: Everything is API-driven and tested from day one.
- **Redis-Backed**: Control plane stores the graph DAG persistently in Redis.

---

## 🗂️ Project Structure

```text
ZTDP/
├── api/                      # API server logic (handlers, routes)
│   ├── handlers/             # HTTP handler logic
│   └── server/               # API routing setup
├── cmd/                      # Entrypoint: main.go
├── internal/                 # Core architecture
│   ├── contracts/            # Contract types: Application, Service, etc.
│   ├── graph/                # Graph engine, backend, resolver
│   └── state/                # State store abstraction (future)
├── rps/                      # Resource Providers (Kubernetes, Postgres, etc.)
├── test/
│   ├── api/                  # End-to-end API tests
│   └── controlplane/         # Control plane validation demos
├── charts/                   # Helm charts (e.g., redis)
├── docker-compose.yaml       # Local services: Redis
├── .env                      # Environment configuration
└── go.mod / go.sum           # Go dependencies
```

---

## ⚙️ Local Development Setup

ZTDP uses **Docker Compose** for local development.

### ✅ Prerequisites

- Docker
- Docker-Compose
- Go 1.22+

### 🔧 Setup

# Create the docker-compose.yaml file
```yaml
services:
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    environment:
      - REDIS_PASSWORD=BVogb1sEPqA
    command: ["redis-server", "--requirepass", "BVogb1sEPqA"]
```
# Start Redis for backend storage
docker-compose up -d

```bash
# Set environment variables
export ZTDP_GRAPH_BACKEND=redis
export REDIS_HOST=localhost:6379
export REDIS_PASSWORD=yourpassword  # matches docker-compose.yaml

# Run a control plane demo
go run ./test/controlplane/graph_demo.go
```

---

## 🧪 Testing Strategy

We follow **TDD** and **API-first** development:

- ✅ Each feature begins with a test
- ✅ Logic and contracts are test-covered (`go test ./...`)
- ✅ APIs are tested with HTTP assertions
- ✅ Redis-backed graph is tested for both in-memory and persistence

```bash
# Run all tests
go test ./...
```

---

## 🌐 API Endpoints

| Method | Endpoint         | Purpose                             |
|--------|------------------|-------------------------------------|
| POST   | `/contracts`     | Submit new contract (app/service)   |
| POST   | `/apply/{env}`   | Apply global graph to an environment |
| GET    | `/graph/{env}`   | View environment-specific DAG       |
| GET    | `/healthz`       | Health check                        |

APIs are public and MCP-compatible.

---

## 📍 MVP Progress

| Phase                | Status     |
|----------------------|------------|
| Contract schema      | ✅ Complete |
| Graph Engine         | ✅ Complete |
| Redis graph backend  | ✅ Complete |
| Control plane demo   | ✅ Complete |
| API-first server     | ✅ In progress |
| Resource Providers   | ⏳ Coming up |
| Event orchestration  | ⏳ Coming up |
| Reconciliation loop  | ⏳ Coming up |

See: [`docs/ZTDP – MVP v1 Development Plan`](docs/ZTDP%20–%20MVP%20v1%20Development%20Plan.md)

---

## 🔐 Secrets & State (Planned)

- Secrets will be stored per environment and injected at runtime.
- State (events, node status) will be tracked in Redis initially.

---

## 📌 License

TBD — Project is in private development. License terms will be clarified before any public release.