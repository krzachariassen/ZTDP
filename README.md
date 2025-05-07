# Zero Touch Developer Platform (ZTDP)

ZTDP is a bold reimagining of the internal developer platform. It empowers you to deliver infrastructure and applications with **zero manual touch**—just intent, contracts, and code. No YAML, no portals, no friction. ZTDP is built for the future: API-first, event-driven, and ready for both human and AI operators.

---

## 🧠 What Makes ZTDP Different?

- **Contract-Driven, Not YAML-Driven:** Express your intent in structured contracts—no more brittle YAML, no more static manifests.
- **Graph-Native Orchestration:** Every resource, dependency, and lifecycle is modeled as a live, queryable graph—enabling true dependency awareness and incremental updates.
- **Composable, Pluggable Resource Providers:** Add new infrastructure types or swap backends without changing the core platform.
- **Event-Driven, Not Pipeline-Driven:** ZTDP is built on an event bus, not a pipeline runner—enabling real-time, auditable, and autonomous operations.
- **AI-Ready by Design:** Structured, deterministic, and safe for both human and AI agents to operate—no hidden state, no magic.
- **API-First, TDD-First:** Every feature is built and tested as an API from day one, with a focus on developer experience and automation.
- **Zero Touch, Zero Friction:** From contract submission to deployment, ZTDP eliminates manual steps, portals, and glue code—just outcomes.

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
│   ├── graph/                # Graph engine, backend, resolver, registry
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

| Method | Endpoint                | Purpose                                 |
|--------|-------------------------|-----------------------------------------|
| POST   | `/v1/applications`         | Submit new application                  |
| GET    | `/v1/applications`         | List all applications                   |
| GET    | `/v1/applications/{app}`   | Get a specific application              |
| PUT    | `/v1/applications/{app}`   | Update an application                   |
| GET    | `/v1/applications/schema`  | Get application contract schema         |
| POST   | `/v1/applications/{app}/services` | Add a service to an application   |
| GET    | `/v1/applications/{app}/services` | List services for an application  |
| GET    | `/v1/applications/{app}/services/{service}` | Get a specific service      |
| GET    | `/v1/services/schema`      | Get service contract schema             |
| POST   | `/v1/apply`                | Apply global graph to an environment    |
| GET    | `/v1/graph`                | View current global DAG                 |
| GET    | `/v1/status`               | Platform status                         |
| GET    | `/v1/healthz`              | Health check                            |

- **Swagger/OpenAPI docs:** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## 🏗️ Generating Sample Data

You can pre-populate the platform with sample data in two ways:

### 1. **Run the Demo Script**

```bash
go run ./test/controlplane/graph_demo.go
```
This will create a sample application (`checkout`) and a service (`checkout-api`) in the graph and persist them to Redis (if configured).

### 2. **Use the API (curl examples)**

```bash
# Create the "checkout" application
curl -X POST http://localhost:8080/v1/applications \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "checkout", "owner": "team-x" },
    "spec": {
      "description": "Handles checkout flows",
      "tags": ["payments", "frontend"],
      "environments": ["dev", "qa"],
      "lifecycle": {}
    }
  }'

# Create the "checkout-api" service under "checkout"
curl -X POST http://localhost:8080/v1/applications/checkout/services \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "checkout-api", "owner": "team-x" },
    "spec": {
      "application": "checkout",
      "port": 8080,
      "public": true
    }
  }'

# Create another service, e.g., "checkout-worker" under "checkout"
curl -X POST http://localhost:8080/v1/applications/checkout/services \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "checkout-worker", "owner": "team-x" },
    "spec": {
      "application": "checkout",
      "port": 9090,
      "public": false
    }
  }'
```

---

## 🏗️ MVP Progress

| Phase                | Status         |
|----------------------|----------------|
| Contract schema      | ✅ Complete    |
| Graph Engine         | ✅ Complete    |
| Redis graph backend  | ✅ Complete    |
| Control plane demo   | ✅ Complete    |
| API-first server     | ✅ In progress |
| Swagger/OpenAPI docs | ✅ Complete    |
| Resource Providers   | ⏳ Coming up   |
| Event orchestration  | ⏳ Coming up   |
| Reconciliation loop  | ⏳ Coming up   |

See: [`MVP_BACKLOG.md`](MVP_BACKLOG.md) for detailed backlog and progress.

---

## 🔄 Regenerating Swagger Docs

After updating handler annotations, run:

```bash
swag init -g api/server/server.go
```

---

## 🔐 Secrets & State (Planned)

- Secrets will be stored per environment and injected at runtime.
- State (events, node status) will be tracked in Redis initially.

---

## 💡 Why Contribute?

ZTDP is for builders, dreamers, and those who want to change how platforms are delivered.  
Whether you’re into Go, distributed systems, or just want to see what a zero-touch platform feels like—jump in, hack, and help us shape the future.

---

## 📌 License

TBD — Project is in private development. License terms will be clarified before any public release.

---

**Ready to build the future? Clone, run, and let’s go! 🚀**