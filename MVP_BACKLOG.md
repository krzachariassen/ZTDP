# Zero Touch Developer Platform (ZTDP) – MVP v1 Backlog

## Phase 1: End-to-End Flow (Walking Skeleton)

- [ ] **API: Contract Submission**
  - Endpoint to accept contract (application/service) via API
  - Basic validation and error handling

- [ ] **Graph Engine**
  - Store contract as node(s) in in-memory/Redis graph
  - Support node/edge creation

- [ ] **Planner**
  - Topological sort of graph (execution order)
  - Return plan for given contract/graph

- [ ] **Resource Provider: Kubernetes Deployment**
  - Minimal RP that can create a Deployment in a test K8s cluster
  - Trigger RP from plan

- [ ] **Event Engine (Minimal)**
  - Dispatch “deploy” event to RP (can be direct function call for MVP)

- [ ] **API: Status/Graph Retrieval**
  - Endpoint to get current graph and node status

- [ ] **Demo: End-to-End Test**
  - Submit contract → See deployment created in K8s

---

## Phase 2: Platform Hardening

- [ ] **Resource Provider: Postgres**
  - Minimal RP that provisions a Postgres DB

- [ ] **Secrets Management (Basic)**
  - Store and inject secrets for RPs

- [ ] **Event Bus Integration**
  - Use NATS for event dispatch (replace direct calls)

- [ ] **State Store**
  - Persist node/event state in Redis

- [ ] **Reconciliation Loop**
  - Detect drift and trigger reconcile events

---

## Phase 3: Developer Experience & Docs

- [ ] **Swagger/OpenAPI Docs**
  - All endpoints documented and browsable

- [ ] **Local Dev Environment**
  - Docker Compose for Redis, Postgres, NATS

- [ ] **README & Usage Docs**
  - Example contract, API usage, and “how to run demo”

---

**How to use:**  
- Work top-down: Don’t start Phase 2 until Phase 1’s end-to-end flow works.
- Each item should be a PR/ticket.
- Demo early and often!

---

*Update this file as you make progress or reprioritize!*