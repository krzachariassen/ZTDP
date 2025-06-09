# ZTDP MVP Backlog

## ✅ COMPLETED: API Testing & AI Deployment Analysis (June 9, 2025)

### 🎯 Major Achievement: Complete API Validation
- **✅ All API tests passing** - 16/16 tests in `/test/api/api_test.go` 
- **✅ Platform setup validated** - Applications, services, environments, resources, policies
- **✅ Clean architecture confirmed** - Proper separation of API/domain/infrastructure layers
- **✅ Test stability achieved** - Consistent results across multiple runs

### 🔍 Critical Discovery: AI Deployment Gap
- **Issue**: V3Agent creates contracts but doesn't execute actual API calls
- **Impact**: Users expect AI to perform deployments, not just plan them
- **Solution needed**: AI-to-API execution bridge

### 📊 Test Results
```
TestCreateAndGetApplication          ✅ PASS
TestListApplications                 ✅ PASS  
TestUpdateApplication                ✅ PASS
TestCreateAndGetService              ✅ PASS
TestListServices                     ✅ PASS
TestApplyGraph                       ✅ PASS
TestGetGrap                          ✅ PASS
TestHealth                           ✅ PASS
TestStatusEndpoint                   ✅ PASS
TestGetApplicationSchema             ✅ PASS
TestGetServiceSchema                 ✅ PASS
TestCreateAndListEnvironments        ✅ PASS
TestDisallowDirectProductionDeployment ✅ PASS
TestDisallowDeploymentToNotAllowedEnv   ✅ PASS
TestResourceCatalogAndLinking        ✅ PASS
TestPolicyAPIEndpoints               ✅ PASS
```

---

## 🔥 CURRENT PRIORITY: AI-to-API Execution Bridge

### Next Critical Task
**Goal**: Make V3Agent actually execute deployments instead of just creating contracts

**Implementation needed**:
```go
// Add to V3Agent
func (a *V3Agent) executeAction(action string, params map[string]interface{}) error {
    switch action {
    case "deploy":
        return a.callDeploymentAPI(params["app"], params["environment"])
    case "create_application":
        return a.callApplicationAPI(params)
    }
}
```

### Success Criteria
- User says "Deploy checkout-api to Dev" → Actual deployment occurs
- AI-based test creates identical platform state as API tests
- Natural language interface fully replaces API calls for common operations

---

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
- [x] **🆕 Enhanced Logging & Event System**
  - [x] Real-time WebSocket log streaming via `/v1/logs/stream`
  - [x] Interactive web UI with clickable log entries and expandable event details
  - [x] Smart event categorization and advanced filtering
  - [x] Rich visual styling with color-coded events and dark mode support
  - [x] Structured event broadcasting with full payload information
- [x] **🆕 Single-Endpoint Application Deployment**
  - [x] `/v1/applications/{app}/deploy` endpoint with intelligent planning
  - [x] Comprehensive deployment results and policy enforcement
  - [x] Backward compatibility with existing deployment methods
- [x] **🆕 Modern Web UI**
  - [x] Interactive graph visualization with real-time updates
  - [x] Professional styling and responsive design
  - [x] Enhanced user experience for log exploration and system monitoring
- [ ] **Resource Provider: Kubernetes Deployment**
  - [ ] Minimal RP that can create a Deployment in a test K8s cluster
  - [ ] Trigger RP from plan
- [ ] **Event Engine (Minimal)**
  - [ ] Dispatch "deploy" event to RP (can be direct function call for MVP)
- [ ] **Demo: End-to-End Test**
  - [ ] Submit Application/Service/Environment → See deployment created in K8s Platform (ZTDP) – MVP Backlog & Roadmap

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