# ZTDP AI-Native Platform Test Plan

This comprehensive test plan validates the AI-native capabilities of the ZTDP platform, including natural language processing, agent orchestration, policy enforcement, and deployment workflows.

## Test Environment Setup

**Prerequisites:**
- ZTDP API server running on http://localhost:8080
- All domain agents (deployment, policy) initialized and started
- In-memory event bus for agent coordination
- Memory-based graph backend for quick testing

**API Endpoints:**
- **AI-Native Interface**: `POST /v3/ai/chat` (Primary test interface)
- **Traditional REST**: `POST /v1/applications`, `POST /v1/environments`, etc.
- **Status**: `GET /v1/health`, `GET /v1/status`

---

## Phase 1: Basic Infrastructure Tests

### Test 1.1: Platform Health Check
**Instruction:** Check platform status
```bash
curl -X GET http://localhost:8080/v1/health
curl -X GET http://localhost:8080/v1/status
```
**Expected:** 
- Health check returns 200 OK
- Status shows all agents running
- Event bus operational

### Test 1.2: AI Chat Endpoint Connectivity
**Instruction:** "Hello, can you help me with deployments?"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, can you help me with deployments?"}'
```
**Expected:** 
- Orchestrator responds with deployment agent capabilities
- No errors in agent coordination
- Structured JSON response with actions/insights

---

## Phase 2: Basic Entity Creation Failures (Missing Dependencies)

### Test 2.1: Deploy Non-Existent Application
**Instruction:** "Deploy app-alpha to production"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-alpha to production"}'
```
**Expected:** 
- ❌ Deployment fails
- Error: "Application app-alpha does not exist"
- Suggests creating application first

### Test 2.2: Deploy to Non-Existent Environment
**Instruction:** "Create app-alpha" then "Deploy app-alpha to production"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create application app-alpha"}'

curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-alpha to production"}'
```
**Expected:** 
- ✅ Application created successfully
- ❌ Deployment fails
- Error: "Environment production does not exist"
- Suggests creating environment first

---

## Phase 3: Success Path - Basic Deployment

### Test 3.1: Complete Happy Path Deployment
**Instructions:** Create environment → Deploy application
```bash
# Create production environment
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a production environment"}'

# Deploy application to production
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-alpha to production"}'
```
**Expected:** 
- ✅ Environment created successfully
- ✅ Deployment succeeds
- Deployment plan generated and executed
- Status shows app-alpha running in production

### Test 3.2: Verify Deployment Status
**Instruction:** "What is the status of app-alpha?"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What is the status of app-alpha?"}'
```
**Expected:** 
- Shows app-alpha deployed to production
- Deployment timestamp and details
- Resource allocation information

---

## Phase 4: Policy Enforcement Tests

### Test 4.1: Create Direct Production Deployment Block Policy
**Instruction:** "Create a policy that blocks direct production deployments"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a policy that blocks direct production deployments. Applications must be deployed to development first."}'
```
**Expected:** 
- ✅ Policy created successfully
- Policy agent acknowledges new policy
- Policy stored in graph database

### Test 4.2: Test Policy Enforcement - Direct Production Deployment
**Instruction:** "Create app-beta" then "Deploy app-beta to production"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create application app-beta"}'

curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-beta to production"}'
```
**Expected:** 
- ✅ Application created successfully
- ❌ Deployment blocked by policy
- Error: "Policy violation: Direct production deployment not allowed"
- Suggests deploying to development first

### Test 4.3: Create Development Environment
**Instruction:** "Create a development environment"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a development environment"}'
```
**Expected:** 
- ✅ Development environment created
- Available for deployments

### Test 4.4: Deploy to Development First (Policy Compliance)
**Instruction:** "Deploy app-beta to development"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-beta to development"}'
```
**Expected:** 
- ✅ Deployment succeeds (policy allows)
- app-beta now running in development
- Deployment history updated

### Test 4.5: Now Deploy to Production (Policy Satisfied)
**Instruction:** "Deploy app-beta to production"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-beta to production"}'
```
**Expected:** 
- ✅ Deployment succeeds (policy satisfied)
- Policy agent validates prior development deployment
- app-beta now running in production

---

## Phase 5: Advanced Policy Scenarios

### Test 5.1: Multi-Stage Policy with Approval Gates
**Instruction:** "Create a policy requiring manual approval for production deployments"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a policy requiring manual approval for all production deployments"}'
```
**Expected:** 
- ✅ Approval policy created
- Future production deployments will require approval

### Test 5.2: Test Approval Gate Policy
**Instruction:** "Create app-gamma and deploy to production"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create application app-gamma"}'

curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-gamma to development"}'

curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-gamma to production"}'
```
**Expected:** 
- ✅ Application created
- ✅ Development deployment succeeds
- ⏳ Production deployment pending approval
- Status: "Waiting for manual approval"

### Test 5.3: Time-Based Deployment Policy
**Instruction:** "Create a policy that blocks production deployments on weekends"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a policy that blocks production deployments on weekends (Saturday and Sunday)"}'
```
**Expected:** 
- ✅ Time-based policy created
- Policy agent will enforce time restrictions

---

## Phase 6: Complex Orchestration Tests

### Test 6.1: Multi-Application Deployment with Dependencies
**Instruction:** "Create app-database and app-frontend. Deploy app-database first, then app-frontend to production"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create application app-database and app-frontend. app-frontend depends on app-database."}'

curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-database to development, then app-frontend to development"}'
```
**Expected:** 
- ✅ Both applications created with dependency
- ✅ Orchestrated deployment respects dependencies
- app-database deploys before app-frontend

### Test 6.2: Rollback Scenario
**Instruction:** "Rollback app-beta to previous version"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Rollback app-beta in production to the previous version"}'
```
**Expected:** 
- ✅ Rollback plan generated
- Previous version identified and restored
- Deployment history updated

---

## Phase 7: Agent Coordination and Intelligence Tests

### Test 7.1: Cross-Agent Information Sharing
**Instruction:** "What policies are affecting app-beta deployments?"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What policies are currently affecting app-beta deployments?"}'
```
**Expected:** 
- Policy agent provides relevant policies
- Deployment agent shows impact on deployments
- Coordinated response from multiple agents

### Test 7.2: Proactive Recommendations
**Instruction:** "What should I deploy next?"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What should I deploy next? Are there any recommendations?"}'
```
**Expected:** 
- AI analyzes current state
- Suggests next deployment actions
- Considers policies and dependencies

### Test 7.3: Complex Natural Language Processing
**Instruction:** "I need to deploy the new user authentication service to prod ASAP, but I'm worried about our weekend deployment policy"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "I need to deploy the new user authentication service to prod ASAP, but I am worried about our weekend deployment policy"}'
```
**Expected:** 
- Understands complex intent and context
- Identifies policy conflicts
- Suggests alternative approaches or overrides

---

## Phase 8: Error Handling and Edge Cases

### Test 8.1: Invalid Natural Language Input
**Instruction:** Ambiguous request
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Do something with the thing"}'
```
**Expected:** 
- Graceful handling of ambiguous input
- Requests clarification
- Suggests specific actions

### Test 8.2: Concurrent Deployment Conflicts
**Instruction:** "Deploy app-alpha and app-beta to production simultaneously"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-alpha and app-beta to production at the same time"}'
```
**Expected:** 
- Detects potential conflicts
- Suggests sequential deployment
- Manages resource allocation

### Test 8.3: Resource Constraints
**Instruction:** "Deploy 10 new applications to production"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create and deploy app-1, app-2, app-3, app-4, app-5, app-6, app-7, app-8, app-9, app-10 to production"}'
```
**Expected:** 
- Identifies resource limitations
- Suggests phased deployment approach
- Optimizes resource allocation

---

## Phase 9: Integration and Performance Tests

### Test 9.1: Graph Database State Consistency
**Instruction:** Query current platform state
```bash
curl -X GET http://localhost:8080/v1/graph
```
**Expected:** 
- Graph shows all created entities
- Relationships properly established
- No orphaned or inconsistent data

### Test 9.2: Event Bus Message Flow
**Instruction:** Monitor agent coordination during complex operation
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Show me all deployment events for the last hour"}'
```
**Expected:** 
- Event history properly recorded
- Agent coordination visible
- Message flow demonstrates event-driven architecture

### Test 9.3: AI Provider Fallback Behavior
**Instruction:** Test with AI provider unavailable
```bash
# Assuming AI provider is disabled/unavailable
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy app-delta to production"}'
```
**Expected:** 
- Graceful fallback to non-AI functionality
- Traditional deployment logic executed
- User informed of AI unavailability

---

## Success Criteria

### ✅ **Platform Architecture**
- Single source of truth for dependencies (main.go composition root)
- No duplicate initialization
- All agents start successfully
- Event-driven coordination works

### ✅ **AI-Native Interface**
- Natural language processing works correctly
- Complex intent recognition and routing
- Multi-agent orchestration
- Contextual responses

### ✅ **Domain Separation**
- Business logic in domain services (not handlers or AI)
- Clean dependency injection
- Policy-first enforcement
- Event-driven communication

### ✅ **Error Handling**
- Graceful failure handling
- Meaningful error messages
- Fallback mechanisms
- Resource constraint management

### ✅ **Policy Enforcement**
- Policy creation and storage
- Real-time policy evaluation
- Complex multi-stage policies
- Policy violation handling

---

## Automation Script

```bash
#!/bin/bash
# ZTDP AI Platform Test Runner
# Usage: ./run_tests.sh [test_phase]

BASE_URL="http://localhost:8080"
PHASE=${1:-all}

function run_test() {
    local test_name="$1"
    local curl_cmd="$2"
    local expected="$3"
    
    echo "🧪 Running Test: $test_name"
    echo "📝 Command: $curl_cmd"
    echo "📋 Expected: $expected"
    
    result=$(eval $curl_cmd)
    echo "📊 Result: $result"
    echo "---"
}

function ai_chat() {
    local message="$1"
    curl -s -X POST $BASE_URL/v3/ai/chat \
        -H "Content-Type: application/json" \
        -d "{\"message\": \"$message\"}"
}

# Test execution based on phase
case $PHASE in
    "basic"|"all")
        echo "🚀 Running Phase 1: Basic Infrastructure Tests"
        run_test "Platform Health" "curl -s $BASE_URL/v1/health" "200 OK"
        run_test "AI Chat Connectivity" "ai_chat 'Hello, can you help me?'" "Structured response"
        ;;
    "failures"|"all")
        echo "🚀 Running Phase 2: Basic Failures"
        run_test "Deploy Non-Existent App" "ai_chat 'Deploy app-alpha to production'" "Application not found error"
        ;;
    # Add more phases as needed
esac
```

This comprehensive test plan ensures the ZTDP AI-native platform works correctly across all scenarios, from basic functionality to complex multi-agent orchestration and policy enforcement.
