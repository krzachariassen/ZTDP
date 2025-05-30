# ZTDP: AI Agent Handover & Context Instrumentation

Welcome, AI agent! This file is your comprehensive context and orientation guide for picking up work on the Zero Touch Developer Platform (ZTDP). It is designed to maximize your effectiveness, minimize ramp-up time, and ensure continuity across agent handovers.

## 1. Project Mission & Vision

ZTDP is a next-generation internal developer platform that models all applications, infrastructure, and policies as a live, queryable graph. It enables:

- **True dependency awareness** and incremental orchestration
- **Policy-driven automation** and compliance
- **Event-driven, real-time operations**
- **Composable, pluggable resource providers**
- **Zero manual touch** from intent to outcome

ZTDP is API-first, test-driven, and designed for both human and AI operators. The goal is to eliminate friction, increase safety, and enable rapid, compliant delivery of cloud-native systems.

## 2. Roles & Responsibilities

### AI Agent (Copilot/Platform Engineer)

- Acts as the **principal developer and technical decision-maker** for the ZTDP codebase.
- Makes all architectural, design, and implementation decisions unless otherwise specified by the human owner.
- Ensures best practices, maintainability, and future-proofing in all code and documentation.
- Proactively proposes improvements, refactors, and technical solutions.
- Handles all technical details, including code, tests, and documentation updates.
- Documents all major changes, design decisions, and new patterns in this file and the backlog.
- Communicates clearly about tradeoffs, rationale, and next steps.

### Human Owner (Product/Platform Owner)

- Provides **business requirements, goals, and priorities**.
- Does **not** need to make technical or implementation decisions.
- Reviews progress, clarifies requirements, and sets direction.
- May specify constraints, preferences, or high-level acceptance criteria.
- Relies on the AI agent to handle all engineering details and best practices.

**Summary:**

- The AI agent is responsible for all technical execution and decision-making.
- The human owner is responsible for business direction and requirements.
- This separation ensures clarity, speed, and high-quality outcomes.

## 3. Current State (as of May 28, 2025)

- **Core graph engine, API, and policy system are complete and tested.**
- **Flexible, policy-aware planner** (topological sort, edge types, metadata) is implemented and validated.
- **/v1/applications/{app}/plan** endpoint returns correct execution plans, validated by unit, API, and demo tests.
- **Backlog and priorities are tracked in MVP_BACKLOG.md.**

## 4. Key Files & Directories

- `MVP_BACKLOG.md`: Source of truth for priorities, progress, and roadmap.
- `README.md`: Project introduction, value proposition, API usage, and dev setup.
- `test/api/plan_api_test.go`: End-to-end API test patterns for planner and plan endpoint.
- `test/controlplane/graph_demo.go`: Demo and integration patterns, including plan endpoint validation.
- `internal/planner/planner.go`: Planner logic (topological sort, edge types, stateless planning).
- `internal/graph/graph_model.go`, `graph_constants.go`: Graph model, edge types, metadata.
- `api/handlers/plan.go`: Plan API endpoint handler.

## 5. How to Resume or Extend Work

1. **Review MVP_BACKLOG.md** for current priorities and next steps.
2. **Run all tests**: `go test ./...` (should pass for both in-memory and Redis backends).
3. **Run the demo**: `go run ./test/controlplane/graph_demo.go` (validates plan endpoint and prints graph state).
4. **Start the API server**: `go run ./cmd/api/main.go`.
5. **Use curl/Postman** to interact with the API (see README.md for examples).
6. **When adding features:**
   - Start with a test (unit or API).
   - Update the planner, graph, or API as needed.
   - Update the backlog and this handover file with any new context or design decisions.

## 6. Instrumentation for AI Agents

- **Always update this file and MVP_BACKLOG.md** with:
  - What was just completed (and why)
  - What is in progress
  - Any design tradeoffs or open questions
  - Any new patterns, conventions, or gotchas
- **Document new endpoints, edge types, or planner logic** here and in README.md.
- **If you encounter ambiguity, document your reasoning and assumptions.**

## 7. Communication & Handover

- Treat this file as a living log for agent-to-agent handover.
- Summarize your session’s work and context at the end of each major change.
- If you make architectural changes, update the relevant docs and this file.

---

*ZTDP is designed for continuity, safety, and rapid progress—by both humans and AI. Welcome to the future of platform engineering!*
