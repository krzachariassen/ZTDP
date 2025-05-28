# Zero Touch Developer Platform (ZTDP) – MVP Backlog & Roadmap

## Phase 1: End-to-End Flow (Walking Skeleton)

- [x] **API: Application Management**
- [x] **API: Service Management**
- [x] **API: Environment Management**
- [x] **Graph Engine**
- [x] **API: Status/Graph Retrieval**
- [x] **Policy Engine**
- [x] **Planner**
  - [x] Flexible, policy-aware planner (topological sort, edge types, metadata)
  - [x] `/v1/applications/{app}/plan` endpoint returns correct plan
  - [x] Unit and API tests for planner and endpoint
  - [x] Demo script validates plan endpoint
- [ ] **Resource Provider: Kubernetes Deployment**
  - [ ] Minimal RP that can create a Deployment in a test K8s cluster
  - [ ] Trigger RP from plan
- [ ] **Event Engine (Minimal)**
  - [ ] Dispatch “deploy” event to RP (can be direct function call for MVP)
- [ ] **Demo: End-to-End Test**
  - [ ] Submit Application/Service/Environment → See deployment created in K8s

---

## Phase 2: Platform Hardening

- [ ] **Resource Provider: Postgres**
  - [ ] Minimal RP that provisions a Postgres DB
- [ ] **Secrets Management (Basic)**
  - [ ] Store and inject secrets for RPs
- [ ] **Event Bus Integration**
  - [ ] Use NATS for event dispatch (replace direct calls)
- [ ] **State Store**
  - [ ] Persist node/event state in Redis
- [ ] **Reconciliation Loop**
  - [ ] Detect drift and trigger reconcile events
- [ ] **Event Store (Pluggable Resource Provider)**
  - [ ] Minimal RP that captures all platform events and operation logs
  - [ ] Allow custom implementations for compliance/integration

---

## Phase 3: Developer Experience & Docs

- [ ] **Swagger/OpenAPI Docs**
  - [ ] All endpoints documented and browsable
- [ ] **Local Dev Environment**
  - [ ] Docker Compose for Redis, Postgres, NATS
- [ ] **README & Usage Docs**
  - [ ] Example Application/Service/Environment, API usage, and “how to run demo”

---

**How to use:**  
- Work top-down: Don’t start Phase 2 until Phase 1’s end-to-end flow works.  
- Each item should be a PR/ticket.  
- Demo early and often!

---

*Update this file as you make progress or reprioritize!*