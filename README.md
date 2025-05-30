# Zero Touch Developer Platform (ZTDP)

ZTDP is a next-generation internal developer platform that redefines how teams deliver, govern, and operate cloud-native applications and infrastructure. It is designed for organizations that want to move fast, stay secure, and eliminate friction—without sacrificing control or compliance.

---

## 🚀 Why ZTDP? (The Value Proposition)

**ZTDP is not just another platform.** It is a radical leap forward for platform engineering, built to solve the real pain points of modern teams:

- **True Dependency Awareness:** Model your entire application, infrastructure, and policy landscape as a live, queryable graph. No more hidden dependencies or brittle pipelines—ZTDP understands and orchestrates everything, end-to-end.
- **Policy-Driven Automation:** Enforce security, compliance, and operational policies at every step. Policies are first-class citizens, attached to transitions and enforced automatically—no more manual gates or out-of-band reviews.
- **Event-Driven, Real-Time:** Every change, deployment, and policy check emits structured events. Integrate with your observability, security, and automation tools in real time.
- **Composable Resource Providers:** Add, swap, or extend infrastructure types (Kubernetes, databases, cloud services) without changing your core platform logic.
- **API-First:** Every feature is accessible via a clean, well-documented API. ZTDP is built for automation, integration, and developer experience from day one.
- **Zero Touch, Zero Friction:** Go from intent to outcome with no manual steps, no portals, and no YAML. ZTDP is designed for both human and AI operators—deterministic, auditable, and safe.

---

## 🏗️ How It Works

1. **Model Everything as a Graph:**
   - Applications, services, environments, resources, and policies are all nodes in a live graph.
   - Edges represent dependencies, deployments, ownership, and policy relationships.
2. **Express Intent via Contracts:**
   - Submit structured contracts (JSON) to declare what you want—ZTDP figures out the rest.
3. **Plan & Enforce:**
   - The planner computes the correct execution order, respecting dependencies and policies.
   - Policies are enforced automatically at every transition.
4. **Event-Driven Execution:**
   - Every operation emits events for observability, audit, and integration.
5. **Composable Resource Providers:**
   - Plug in new infrastructure types or swap backends with zero friction.

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
| GET    | `/v1/status`                                                    | Platform status                                 |
| GET    | `/v1/healthz`                                                   | Health check                                    |

- **Swagger/OpenAPI docs:** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## 🏗️ Example API Usage (with curl)

### 1. Create Environments
```bash
curl -X POST http://localhost:8080/v1/environments \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "dev", "owner": "platform-team" },
    "spec": { "description": "Development environment" }
  }'

curl -X POST http://localhost:8080/v1/environments \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "prod", "owner": "platform-team" },
    "spec": { "description": "Production environment" }
  }'
```

### 2. Create Application
```bash
curl -X POST http://localhost:8080/v1/applications \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "checkout", "owner": "team-x" },
    "spec": {
      "description": "Handles checkout flows",
      "tags": ["payments", "frontend"],
      "lifecycle": {}
    }
  }'
```

### 3. Allow Application to Deploy to Environments
```bash
# Allow checkout app to deploy to dev
dev_env="dev"
curl -X POST http://localhost:8080/v1/applications/checkout/environments/$dev_env/allowed

# Allow checkout app to deploy to prod
prod_env="prod"
curl -X POST http://localhost:8080/v1/applications/checkout/environments/$prod_env/allowed
```

### 4. Create Services for the Application
```bash
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

### 5. Create Service Versions
```bash
# Create a new version for a service
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.0.0",
    "config_ref": "default-config",
    "owner": "team-x"
  }'

# Create another version for a service
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.1.0",
    "config_ref": "default-config",
    "owner": "team-x"
  }'
```

### 6. List Service Versions
```bash
curl -X GET http://localhost:8080/v1/applications/checkout/services/checkout-api/versions
```

### 7. Deploy a Service Version to an Environment
```bash
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions/1.0.0/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "dev"
  }'
```

### 8. List Deployments in an Environment
```bash
curl -X GET http://localhost:8080/v1/environments/dev/deployments
```

### 9. Deploy Services to Environments
```bash
# Deploy checkout-api to dev
deploy_service="checkout-api"
deploy_env="dev"
curl -X POST http://localhost:8080/v1/applications/checkout/services/$deploy_service/environments/$deploy_env

# Deploy checkout-api to prod
prod_env="prod"
curl -X POST http://localhost:8080/v1/applications/checkout/services/$deploy_service/environments/$prod_env

# Deploy checkout-worker to dev
worker_service="checkout-worker"
curl -X POST http://localhost:8080/v1/applications/checkout/services/$worker_service/environments/$deploy_env
```

> **Note:** This operation now creates a `deploy` edge in the graph (previously `deployed_in`).

### 10. Attempt to Deploy Service to Not-Allowed Environment (Should Fail)
```bash
# Remove prod from allowed environments for checkout (replace allowed list with only dev)
curl -X PUT http://localhost:8080/v1/applications/checkout/environments/allowed \
  -H "Content-Type: application/json" \
  -d '["dev"]'

# Try to deploy checkout-worker to prod (should fail with 403)
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-worker/environments/prod -v
```

---

## 🚀 Simple Application Deployment (NEW!)

The easiest way to deploy applications is with our new single-endpoint deployment API:

### Deploy Entire Application
```bash
# Deploy application to development environment
curl -X POST http://localhost:8080/v1/applications/checkout/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}'

# Deploy application to production environment  
curl -X POST http://localhost:8080/v1/applications/checkout/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "prod", "version": "1.0.0"}'
```

**What happens internally:**
1. ✅ Validates application and environment exist
2. ✅ Enforces deployment policies 
3. ✅ Generates deployment plan for all services
4. ✅ Deploys all service versions in correct order
5. ✅ Returns comprehensive deployment results

**Response:**
```json
{
  "application": "checkout",
  "environment": "dev", 
  "deployments": ["checkout-api:1.0.0", "checkout-worker:1.0.0"],
  "skipped": [],
  "failed": [],
  "summary": {
    "success": true,
    "message": "Successfully deployed checkout to dev"
  }
}
```

---

## 🏗️ Detailed API Usage (Legacy/Advanced)

### 1. Create Environments
```bash
curl -X POST http://localhost:8080/v1/environments \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "dev", "owner": "platform-team" },
    "spec": { "description": "Development environment" }
  }'

curl -X POST http://localhost:8080/v1/environments \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "prod", "owner": "platform-team" },
    "spec": { "description": "Production environment" }
  }'
```

### 2. Create Application
```bash
curl -X POST http://localhost:8080/v1/applications \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": { "name": "checkout", "owner": "team-x" },
    "spec": {
      "description": "Handles checkout flows",
      "tags": ["payments", "frontend"],
      "lifecycle": {}
    }
  }'
```

### 3. Allow Application to Deploy to Environments
```bash
# Allow checkout app to deploy to dev
dev_env="dev"
curl -X POST http://localhost:8080/v1/applications/checkout/environments/$dev_env/allowed

# Allow checkout app to deploy to prod
prod_env="prod"
curl -X POST http://localhost:8080/v1/applications/checkout/environments/$prod_env/allowed
```

### 4. Create Services for the Application
```bash
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

### 5. Create Service Versions
```bash
# Create a new version for a service
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.0.0",
    "config_ref": "default-config",
    "owner": "team-x"
  }'

# Create another version for a service
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.1.0",
    "config_ref": "default-config",
    "owner": "team-x"
  }'
```

### 6. List Service Versions
```bash
curl -X GET http://localhost:8080/v1/applications/checkout/services/checkout-api/versions
```

### 7. Deploy a Service Version to an Environment
```bash
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions/1.0.0/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "dev"
  }'
```

### 8. List Deployments in an Environment
```bash
curl -X GET http://localhost:8080/v1/environments/dev/deployments
```

### 9. Deploy Services to Environments
```bash
# Deploy checkout-api to dev
deploy_service="checkout-api"
deploy_env="dev"
curl -X POST http://localhost:8080/v1/applications/checkout/services/$deploy_service/environments/$deploy_env

# Deploy checkout-api to prod
prod_env="prod"
curl -X POST http://localhost:8080/v1/applications/checkout/services/$deploy_service/environments/$prod_env

# Deploy checkout-worker to dev
worker_service="checkout-worker"
curl -X POST http://localhost:8080/v1/applications/checkout/services/$worker_service/environments/$deploy_env
```

> **Note:** This operation now creates a `deploy` edge in the graph (previously `deployed_in`).

### 10. Attempt to Deploy Service to Not-Allowed Environment (Should Fail)
```bash
# Remove prod from allowed environments for checkout (replace allowed list with only dev)
curl -X PUT http://localhost:8080/v1/applications/checkout/environments/allowed \
  -H "Content-Type: application/json" \
  -d '["dev"]'