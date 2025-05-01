# Zero Touch Developer Platform (ZTDP)

ZTDP is a next-generation internal developer platform designed for autonomous, declarative, AI-native infrastructure and application orchestration. It replaces traditional YAML and GitOps models with contract-based execution, graph-oriented orchestration, and natural language input.

## 🏗️ Project Structure

```text
ZTDP/
├── bootstrap-k3d.sh        # Local dev bootstrap: k3d + Redis, Postgres, NATS
├── cmd/                    # Application entrypoints (main.go)
├── internal/               # Core platform components
│   ├── contracts/          # Contract types: app, service, db, etc.
│   ├── graph/              # DAG assembly and query engine
│   ├── planner/            # Execution planning logic
│   ├── lifecycle/          # Gate enforcement and promotion
│   ├── state/              # Redis integration and state store
│   ├── events/             # NATS integration and event model
│   └── reconcile/          # Drift detection and reconciliation loop
├── rps/                    # Resource Providers
│   ├── kubernetes/         # Handles deployments
│   └── postgres/           # Handles database provisioning
├── adapters/               # Infrastructure interfaces (e.g., k8s, redis)
├── api/                    # API server logic (contract submission)
├── docs/                   # Architecture diagrams and planning docs
└── go.mod / go.sum         # Go module setup
```

## 🚀 Local Setup

1. Ensure Docker Desktop is running with WSL2 integration enabled
2. Run the setup script:

```bash
chmod +x bootstrap-k3d.sh
./bootstrap-k3d.sh
```

This provisions a local `k3d` cluster with Redis, Postgres, and NATS under the `ztdp` namespace.

## 🧱 MVP Progress

The MVP is being developed in phases:

- ✅ Local Kubernetes environment
- 🔄 In progress: Contract schema, API scaffolding
- ⏳ Next: Graph engine, planner, event orchestration

See [`docs/ZTDP – MVP v1 Development Plan`](docs/ZTDP%20–%20MVP%20v1%20Development%20Plan.md) for full details.

## 📌 License

TBD — initial development is private. License will be defined before public release.
