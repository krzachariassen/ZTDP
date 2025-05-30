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

## 3. Current State (as of December 2024)

- **Core graph engine, API, and policy system are complete and tested.**
- **Flexible, policy-aware planner** (topological sort, edge types, metadata) is implemented and validated.
- **/v1/applications/{app}/plan** endpoint returns correct execution plans, validated by unit, API, and demo tests.
- **🆕 Enhanced Logging & Event System:**
  - **Real-time WebSocket log streaming** via `/v1/logs/stream` with structured event data
  - **Interactive web UI** with clickable log entries, expandable event details, and rich JSON payload display
  - **Smart event categorization** with automatic classification (Application Created/Updated/Deleted, Deployment, Policy, Resource, Connection)
  - **Advanced filtering system** supporting event types, components, log levels, and free-text search
  - **Complete visual styling** with color-coded events, animations, and dark mode support
- **🆕 Single-Endpoint Application Deployment** via `/v1/applications/{app}/deploy` with intelligent planning and comprehensive results
- **🆕 Modern Web UI** with interactive graph visualization, real-time updates, and professional styling
- **Backlog and priorities are tracked in MVP_BACKLOG.md.**

## 4. Key Files & Directories

### Core Architecture
- `MVP_BACKLOG.md`: Source of truth for priorities, progress, and roadmap.
- `README.md`: Project introduction, value proposition, API usage, and dev setup.
- `internal/planner/planner.go`: Planner logic (topological sort, edge types, stateless planning).
- `internal/graph/graph_model.go`, `graph_constants.go`: Graph model, edge types, metadata.
- `internal/events/events.go`: Event system and structured event definitions.

### API & Handlers
- `api/handlers/plan.go`: Plan API endpoint handler.
- `api/handlers/logs.go`: **🆕 Enhanced log streaming with structured event broadcasting**.
- `api/handlers/applications.go`: Application management and **🆕 single-endpoint deployment**.
- `api/server/server.go`: API routing setup and WebSocket endpoint configuration.

### Logging & Real-time Systems
- `internal/logging/realtime.go`: **🆕 WebSocket broadcasting infrastructure with BroadcastEvent method**.
- `static/graph-modern.html`: **🆕 Enhanced web UI with interactive logs, clickable events, and filtering**.
- `static/graph-modern.css`: **🆕 Complete styling system for rich log display and event visualization**.

### Testing & Validation
- `test/api/plan_api_test.go`: End-to-end API test patterns for planner and plan endpoint.
- `test/controlplane/graph_demo.go`: Demo and integration patterns, including plan endpoint validation.

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

## 8. Recent Architectural Improvements (December 2024)

### Enhanced Logging & Event System

**Major Enhancement:** Completely redesigned the logging system to provide rich, interactive event exploration:

#### Backend Enhancements:
- **Structured Event Broadcasting**: Enhanced `api/handlers/logs.go` to create rich event data with categories, detailed messages, and full payload information
- **WebSocket Infrastructure**: Added `BroadcastEvent()` method to `internal/logging/realtime.go` for direct event broadcasting
- **Event Categorization**: Smart categorization of events based on payload content (Application Created/Updated/Deleted, Deployment, Policy, Resource, Connection)
- **Enhanced Message Formatting**: Events include descriptive text with appropriate emojis (🎯 🚀 📦 ✅ ❌ 🔒 🔍 🔧 🗑️)

#### Frontend Enhancements:
- **Interactive Log Entries**: Clickable events that expand to show full details, payload, and metadata
- **Rich Visual Display**: Color-coded event types, expandable details with chevron icons, and grid layouts for event information
- **Smart Filtering**: Updated filter options to match actual event architecture with proper categorization
- **WebSocket Protocol**: Enhanced to handle both regular logs and structured events with type indicators
- **Complete Styling**: Comprehensive CSS system with dark mode support, animations, and responsive design

#### Key Implementation Details:
- **Event Structure**: Events include `event_category`, structured `event` data, timestamp, and rich payload information
- **WebSocket Messages**: Support both regular logs (`log.entry`) and structured events (`event.structured`)
- **Filtering Logic**: `matchesEventType()` method with smart categorization for accurate filtering
- **Visual Feedback**: Expand/collapse functionality with proper state management and visual indicators

### Single-Endpoint Application Deployment

**Major Feature:** Added comprehensive application deployment via `/v1/applications/{app}/deploy`:

- **Intelligent Planning**: Automatic generation of deployment plans for all services
- **Policy Enforcement**: Built-in validation and policy checks during deployment
- **Comprehensive Results**: Detailed deployment status with success/failure reporting
- **Backward Compatibility**: Maintains existing individual service deployment endpoints

### Modern Web UI Architecture

**Complete Redesign:** Enhanced the web interface with:

- **Real-time Graph Visualization**: Interactive nodes and edges with live updates
- **Professional Styling**: Modern CSS with consistent design patterns
- **Responsive Design**: Mobile-friendly layout with proper breakpoints
- **Dark Mode Support**: Complete theming system for different viewing preferences

### Development Patterns & Best Practices

When working with the enhanced logging system:

1. **Event Creation**: Use the `createEventHandler()` pattern in `logs.go` for structured events
2. **WebSocket Broadcasting**: Prefer `BroadcastEvent()` for real-time event distribution
3. **Frontend Handling**: Separate logic for regular logs vs. structured events in the UI
4. **Styling Consistency**: Follow the established CSS class patterns for new UI components
5. **Event Categorization**: Ensure new event types are properly categorized for filtering

### Technical Debt & Future Considerations

- **Event Schema**: Consider formalizing event schema with JSON schema validation
- **Performance**: Monitor WebSocket connection scaling as event volume increases
- **Internationalization**: Current UI is English-only, consider i18n for global deployment
- **Accessibility**: Ensure keyboard navigation and screen reader support for expanded log entries
- **Testing**: Add comprehensive E2E tests for WebSocket functionality and UI interactions

---

*ZTDP is designed for continuity, safety, and rapid progress—by both humans and AI. Welcome to the future of platform engineering!*
