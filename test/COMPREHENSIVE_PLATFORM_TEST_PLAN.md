# ZTDP AI-Native Platform Comprehensive Test Plan

This test plan creates a complete platform demonstration using the AI-native `/v3/ai/chat` interface. It tests natural language interactions with the orchestrator to create applications, services, environments, resources, policies, and releases through agent coordination.

## Test Environment Setup

### Prerequisites
- ZTDP API server running on http://localhost:8080
- All domain agents (application, deployment, policy) initialized
- Orchestrator with real OpenAI integration running
- In-memory event bus for agent coordination
- Memory-based graph backend

### API Endpoints Used
- **Primary Interface**: `POST /v3/ai/chat` (AI-native natural language)
- **Health Check**: `GET /v1/health`
- **Graph Query**: `GET /v1/applications`, `GET /v1/environments` (for verification only)

#### Applications (Created via AI)
- **checkout**: E-commerce checkout application with team-x ownership
- **payment**: Payment processing application with team-y ownership  
- **monitoring**: Platform monitoring application with platform-team ownership

#### Services (Created via AI)
- **checkout-api** (port 8080, public) - Checkout API service
- **checkout-worker** (port 9090, internal) - Background worker
- **payment-api** (port 8081, public) - Payment processing API
- **metrics-collector** (port 9092, internal) - Metrics collection service
- **alerting-service** (port 9093, internal) - Alert management service

#### Environments (Created via AI)
- **dev**: Development environment (unrestricted deployment)
- **staging**: Staging environment (requires validation)
- **production**: Production environment (strict policy enforcement)

#### Resources (Future: Created via AI)
- PostgreSQL databases, Redis instances, Kafka clusters
- Resource dependencies and capacity management

#### Policies (Future: Created via AI)
- Deployment approval workflows
- Security and compliance requirements
- Resource allocation limits

---

## Test Scenarios

### Phase 1: Platform Health & Basic AI Interaction

#### Test 1.1: Platform Health Check
```bash
# Test: Check platform health
curl -X GET http://localhost:8080/v1/health
# Expect: 200 OK with health status

# Test: Verify empty platform state
curl -X GET http://localhost:8080/v1/applications
# Expect: Empty array []

curl -X GET http://localhost:8080/v1/environments  
# Expect: Empty array []
```

#### Test 1.2: Basic AI Interaction
**Instruction:** "Hello, can you help me with platform management?"
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, can you help me with platform management?"}'
```
**Expected Response:**
- Orchestrator acknowledges capability to help
- Lists available agents (application, deployment, policy)
- Provides guidance on available operations

---

### Phase 2: AI-Native Application Creation

#### Test 2.1: Create Checkout Application
**Instruction:** "Create a new application called 'checkout' owned by team-x. It's an e-commerce checkout application with tags payments and core."
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a new application called checkout owned by team-x. It is an e-commerce checkout application with tags payments and core."}'
```
**Expected Response:**
- Intent detected: "create application"
- Routed to application-agent
- Application created successfully
- Response includes application details

#### Test 2.2: Create Payment Application  
**Instruction:** "Create a payment processing application named 'payment' for team-y with financial and payments tags."
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a payment processing application named payment for team-y with financial and payments tags."}'
```

#### Test 2.3: Create Monitoring Application
**Instruction:** "Create a monitoring application for the platform-team with observability and platform tags."
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a monitoring application for the platform-team with observability and platform tags."}'
```

#### Test 2.4: Verify Applications Created
```bash
# Test: Query applications to verify creation
curl -X GET http://localhost:8080/v1/applications
# Expect: Array with 3 applications (checkout, payment, monitoring)
```

---

### Phase 3: AI-Native Environment Creation

#### Test 3.1: Create Development Environment
**Instruction:** "Create a development environment called 'dev' owned by platform-team for development work."
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a development environment called dev owned by platform-team for development work."}'
```

#### Test 3.2: Create Staging Environment
**Instruction:** "Create a staging environment for testing before production."
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a staging environment for testing before production."}'
```

#### Test 3.3: Create Production Environment
**Instruction:** "Create a production environment with strict policies for live workloads."
```bash
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a production environment with strict policies for live workloads."}'
```

#### Test 3.4: Verify Environments Created
curl -X POST http://localhost:8080/v1/environments \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "staging", "owner": "platform-team"},
    "spec": {"description": "Staging environment"}
  }'
# Expect: 201 Created

# Test: Create production environment
curl -X POST http://localhost:8080/v1/environments \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "production", "owner": "platform-team"},
    "spec": {"description": "Production environment"}
  }'
# Expect: 201 Created
```

#### Test 1.4: Create Services
```bash
# Test: Create checkout services
curl -X POST http://localhost:8080/v1/applications/checkout/services \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "checkout-api", "owner": "team-x"},
    "spec": {"application": "checkout", "port": 8080, "public": true, "description": "Checkout API service"}
  }'
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/checkout/services \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "checkout-worker", "owner": "team-x"},
    "spec": {"application": "checkout", "port": 9090, "public": false, "description": "Checkout background worker"}
  }'
# Expect: 201 Created

# Test: Create payment services
curl -X POST http://localhost:8080/v1/applications/payment/services \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "payment-api", "owner": "team-y"},
    "spec": {"application": "payment", "port": 8081, "public": true, "description": "Payment API service"}
  }'
# Expect: 201 Created

# Test: Create monitoring services
curl -X POST http://localhost:8080/v1/applications/monitoring/services \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "metrics-collector", "owner": "platform-team"},
    "spec": {"application": "monitoring", "port": 9092, "public": false, "description": "Metrics collection service"}
  }'
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/monitoring/services \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "alerting-service", "owner": "platform-team"},
    "spec": {"application": "monitoring", "port": 9093, "public": false, "description": "Alerting service"}
  }'
# Expect: 201 Created
```

### Phase 2: Resource Management

#### Test 2.1: Create Resource Types
```bash
# Test: Create PostgreSQL resource type
curl -X POST http://localhost:8080/v1/resources \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "resource_type",
    "metadata": {"name": "postgres", "owner": "platform-team"},
    "spec": {
      "version": "15.0",
      "tier_options": ["standard", "high-memory", "high-cpu"],
      "default_tier": "standard",
      "description": "PostgreSQL database service"
    }
  }'
# Expect: 201 Created

# Test: Create Redis resource type
curl -X POST http://localhost:8080/v1/resources \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "resource_type", 
    "metadata": {"name": "redis", "owner": "platform-team"},
    "spec": {
      "version": "7.0",
      "tier_options": ["cache", "persistent"],
      "default_tier": "cache",
      "description": "Redis in-memory cache service"
    }
  }'
# Expect: 201 Created

# Test: Create Kafka resource type
curl -X POST http://localhost:8080/v1/resources \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "resource_type",
    "metadata": {"name": "kafka", "owner": "platform-team"},
    "spec": {
      "version": "3.4",
      "tier_options": ["standard", "high-throughput"],
      "default_tier": "standard", 
      "description": "Kafka streaming platform"
    }
  }'
# Expect: 201 Created
```

#### Test 2.2: Create Resource Instances
```bash
# Test: Create PostgreSQL database
curl -X POST http://localhost:8080/v1/resources \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "resource",
    "metadata": {"name": "pg-db", "owner": "platform-team"},
    "spec": {"type": "postgres", "version": "15.0", "tier": "standard", "capacity": "20GB", "plan": "prod"}
  }'
# Expect: 201 Created

# Test: Create Redis cache
curl -X POST http://localhost:8080/v1/resources \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "resource",
    "metadata": {"name": "redis-cache", "owner": "platform-team"},
    "spec": {"type": "redis", "version": "7.0", "tier": "cache", "capacity": "2GB", "plan": "prod"}
  }'
# Expect: 201 Created

# Test: Create Redis persistent storage
curl -X POST http://localhost:8080/v1/resources \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "resource",
    "metadata": {"name": "redis-persistent", "owner": "platform-team"},
    "spec": {"type": "redis", "version": "7.0", "tier": "persistent", "capacity": "5GB", "plan": "prod"}
  }'
# Expect: 201 Created

# Test: Create Kafka event bus
curl -X POST http://localhost:8080/v1/resources \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "resource",
    "metadata": {"name": "event-bus", "owner": "platform-team"},
    "spec": {"type": "kafka", "version": "3.4", "tier": "standard", "capacity": "15GB", "plan": "prod"}
  }'
# Expect: 201 Created

# Test: Create metrics database
curl -X POST http://localhost:8080/v1/resources \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "resource",
    "metadata": {"name": "metrics-db", "owner": "platform-team"},
    "spec": {"type": "postgres", "version": "15.0", "tier": "high-memory", "capacity": "50GB", "plan": "prod"}
  }'
# Expect: 201 Created
```

#### Test 2.3: Link Resources to Applications
```bash
# Test: Link resources to checkout app
curl -X POST http://localhost:8080/v1/applications/checkout/resources/pg-db
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/checkout/resources/redis-cache  
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/checkout/resources/event-bus
# Expect: 201 Created

# Test: Link resources to payment app
curl -X POST http://localhost:8080/v1/applications/payment/resources/redis-persistent
# Expect: 201 Created

# Test: Link resources to monitoring app
curl -X POST http://localhost:8080/v1/applications/monitoring/resources/metrics-db
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/monitoring/resources/event-bus
# Expect: 201 Created
```

### Phase 3: Service Versions and Dependencies

#### Test 3.1: Create Service Versions
```bash
# Test: Create checkout-api versions
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions \
  -H "Content-Type: application/json" \
  -d '{"version": "1.0.0", "config_ref": "default-config"}'
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/versions \
  -H "Content-Type: application/json" \
  -d '{"version": "2.0.0", "config_ref": "default-config"}'
# Expect: 201 Created

# Test: Create other service versions
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-worker/versions \
  -H "Content-Type: application/json" \
  -d '{"version": "1.0.0", "config_ref": "default-config"}'
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/payment/services/payment-api/versions \
  -H "Content-Type: application/json" \
  -d '{"version": "1.0.0", "config_ref": "default-config"}'
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/payment/services/payment-api/versions \
  -H "Content-Type: application/json" \
  -d '{"version": "1.1.0", "config_ref": "default-config"}'
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/monitoring/services/metrics-collector/versions \
  -H "Content-Type: application/json" \
  -d '{"version": "1.0.0", "config_ref": "default-config"}'
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/monitoring/services/alerting-service/versions \
  -H "Content-Type: application/json" \
  -d '{"version": "1.0.0", "config_ref": "default-config"}'
# Expect: 201 Created
```

#### Test 3.2: Link Services to Resources
```bash
# Test: Link checkout services to resources
curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/resources/pg-db
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-api/resources/redis-cache
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-worker/resources/pg-db
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/checkout/services/checkout-worker/resources/event-bus
# Expect: 201 Created

# Test: Link payment services to resources
curl -X POST http://localhost:8080/v1/applications/payment/services/payment-api/resources/redis-persistent
# Expect: 201 Created

# Test: Link monitoring services to resources
curl -X POST http://localhost:8080/v1/applications/monitoring/services/metrics-collector/resources/metrics-db
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/monitoring/services/metrics-collector/resources/event-bus
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/monitoring/services/alerting-service/resources/metrics-db
# Expect: 201 Created
```

### Phase 4: Environment Access Setup

#### Test 4.1: Configure Environment Access
```bash
# Test: Grant checkout access to all environments
curl -X POST http://localhost:8080/v1/applications/checkout/environments/dev/allowed
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/checkout/environments/staging/allowed
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/checkout/environments/production/allowed
# Expect: 201 Created

# Test: Grant payment access to all environments
curl -X POST http://localhost:8080/v1/applications/payment/environments/dev/allowed
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/payment/environments/staging/allowed
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/payment/environments/production/allowed
# Expect: 201 Created

# Test: Grant monitoring access to all environments
curl -X POST http://localhost:8080/v1/applications/monitoring/environments/dev/allowed
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/monitoring/environments/staging/allowed
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/monitoring/environments/production/allowed
# Expect: 201 Created
```

### Phase 5: Deployment Testing (Pre-Policy)

#### Test 5.1: Basic Deployment Success
```bash
# Test: Deploy checkout to dev (should succeed)
curl -X POST http://localhost:8080/v1/applications/checkout/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}'
# Expect: 200 OK, deployment success

# Test: Deploy payment to dev (should succeed)
curl -X POST http://localhost:8080/v1/applications/payment/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}'
# Expect: 200 OK, deployment success

# Test: Deploy monitoring to dev (should succeed)
curl -X POST http://localhost:8080/v1/applications/monitoring/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}'
# Expect: 200 OK, deployment success
```

#### Test 5.2: Direct Production Deployment (Pre-Policy)
```bash
# Test: Deploy checkout directly to production (should succeed before policy)
curl -X POST http://localhost:8080/v1/applications/checkout/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "production"}'
# Expect: 200 OK, deployment success (no policies yet)
```

### Phase 6: Policy Implementation and Testing

#### Test 6.1: Create Deployment Policies
```bash
# Test: Create no-direct-prod policy
curl -X POST http://localhost:8080/v1/policies \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "no-direct-prod", "owner": "platform-team"},
    "spec": {
      "description": "Prevent direct production deployments - must deploy to dev first",
      "type": "deployment",
      "enforcement": "blocking",
      "rules": [
        {
          "condition": "deployment.environment == \"production\"",
          "requirement": "deployment.application.last_deployed_env == \"dev\"",
          "message": "Must deploy to dev environment before production"
        }
      ]
    }
  }'
# Expect: 201 Created

# Test: Create code-scan policy  
curl -X POST http://localhost:8080/v1/policies \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "code-scan-required", "owner": "platform-team"},
    "spec": {
      "description": "Require security scan before staging/prod deployment",
      "type": "deployment", 
      "enforcement": "blocking",
      "rules": [
        {
          "condition": "deployment.environment in [\"staging\", \"production\"]",
          "requirement": "deployment.application.security_scan_status == \"passed\"",
          "message": "Security scan must pass before staging/production deployment"
        }
      ]
    }
  }'
# Expect: 201 Created

# Test: Create resource limits policy
curl -X POST http://localhost:8080/v1/policies \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "resource-limits", "owner": "platform-team"},
    "spec": {
      "description": "Enforce resource capacity limits",
      "type": "resource",
      "enforcement": "blocking", 
      "rules": [
        {
          "condition": "resource.capacity > \"100GB\"",
          "requirement": "resource.approval_status == \"approved\"",
          "message": "Resources over 100GB require approval"
        }
      ]
    }
  }'
# Expect: 201 Created
```

#### Test 6.2: Test Policy Enforcement
```bash
# Test: Try direct production deployment (should fail with policy)
curl -X POST http://localhost:8080/v1/applications/payment/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "production"}'
# Expect: 403 Forbidden, policy violation: "Must deploy to dev environment before production"

# Test: Deploy to dev first, then production (should succeed)
curl -X POST http://localhost:8080/v1/applications/payment/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}'
# Expect: 200 OK

curl -X POST http://localhost:8080/v1/applications/payment/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "production"}'
# Expect: 200 OK (policy satisfied)

# Test: Try staging deployment without security scan (should fail)
curl -X POST http://localhost:8080/v1/applications/checkout/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "staging"}'
# Expect: 403 Forbidden, policy violation: "Security scan must pass before staging/production deployment"
```

### Phase 7: AI-Native Testing

#### Test 7.1: Natural Language Deployment
```bash
# Test: Deploy using natural language
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy the checkout application to the development environment"}'
# Expect: Successful deployment response with AI explanation

# Test: Complex deployment request  
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "I need to deploy payment app version 1.1.0 to production. What do I need to do first?"}'
# Expect: AI guidance about policy requirements (dev deployment first, security scan)

# Test: Resource creation via AI
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a new Redis instance for caching with 4GB capacity"}'
# Expect: AI creates Redis resource or explains process
```

#### Test 7.2: AI Policy Consultation
```bash
# Test: Ask about deployment policies
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What policies apply to production deployments?"}'
# Expect: AI lists relevant policies and requirements

# Test: Policy compliance check
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Can I deploy monitoring app to production right now?"}'
# Expect: AI checks current state and policy compliance

# Test: Policy violation explanation
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Why did my production deployment fail?"}'
# Expect: AI explains policy violations and remediation steps
```

#### Test 7.3: AI Platform Intelligence
```bash
# Test: Platform status inquiry
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What applications are currently deployed to production?"}'
# Expect: AI provides deployment status summary

# Test: Resource utilization
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Which applications are using the most resources?"}'
# Expect: AI analyzes resource usage and provides insights

# Test: Deployment recommendations
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What should I deploy next to improve the platform?"}'
# Expect: AI provides strategic deployment recommendations
```

### Phase 8: Advanced Workflow Testing

#### Test 8.1: Complete Deployment Workflow
```bash
# Test: Full workflow - new app to production
# Step 1: Create new application
curl -X POST http://localhost:8080/v1/applications \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "orders", "owner": "team-z"},
    "spec": {"description": "Order management application", "tags": ["orders", "core"]}
  }'
# Expect: 201 Created

# Step 2: Add service
curl -X POST http://localhost:8080/v1/applications/orders/services \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "orders-api", "owner": "team-z"},
    "spec": {"application": "orders", "port": 8082, "public": true, "description": "Orders API service"}
  }'
# Expect: 201 Created

# Step 3: Grant environment access
curl -X POST http://localhost:8080/v1/applications/orders/environments/dev/allowed
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/orders/environments/staging/allowed
# Expect: 201 Created

curl -X POST http://localhost:8080/v1/applications/orders/environments/production/allowed
# Expect: 201 Created

# Step 4: Try production deployment (should fail - policy)
curl -X POST http://localhost:8080/v1/applications/orders/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "production"}'
# Expect: 403 Forbidden, policy violation

# Step 5: Deploy to dev first
curl -X POST http://localhost:8080/v1/applications/orders/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}'
# Expect: 200 OK

# Step 6: Mock security scan completion
curl -X POST http://localhost:8080/v1/applications/orders/security-scan \
  -H "Content-Type: application/json" \
  -d '{"status": "passed", "scan_id": "scan-12345"}'
# Expect: 200 OK (if endpoint exists) or skip

# Step 7: Deploy to staging
curl -X POST http://localhost:8080/v1/applications/orders/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "staging"}'
# Expect: 200 OK or 403 if security scan not implemented

# Step 8: Deploy to production
curl -X POST http://localhost:8080/v1/applications/orders/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "production"}'
# Expect: 200 OK (all policies satisfied)
```

#### Test 8.2: Release Management
```bash
# Test: Create release for checkout app
curl -X POST http://localhost:8080/v1/applications/checkout/releases \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "checkout-v2.0", "owner": "team-x"},
    "spec": {
      "version": "2.0.0",
      "description": "Major checkout improvements",
      "services": [
        {"name": "checkout-api", "version": "2.0.0"},
        {"name": "checkout-worker", "version": "1.0.0"}
      ]
    }
  }'
# Expect: 201 Created

# Test: Deploy release to staging
curl -X POST http://localhost:8080/v1/applications/checkout/releases/checkout-v2.0/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "staging"}'
# Expect: 200 OK or appropriate response

# Test: Promote release to production
curl -X POST http://localhost:8080/v1/applications/checkout/releases/checkout-v2.0/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "production"}'
# Expect: 200 OK or policy-based response
```

### Phase 9: Final Validation

#### Test 9.1: Platform State Verification
```bash
# Test: Verify final platform state
curl -X GET http://localhost:8080/v1/applications
# Expect: Array with 4 applications (checkout, payment, monitoring, orders)

curl -X GET http://localhost:8080/v1/environments
# Expect: Array with 3 environments (dev, staging, production)

curl -X GET http://localhost:8080/v1/resources
# Expect: Array with resource types and instances

curl -X GET http://localhost:8080/v1/policies
# Expect: Array with 3 policies (no-direct-prod, code-scan-required, resource-limits)

# Test: Get platform summary via AI
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Give me a complete summary of the current platform state"}'
# Expect: Comprehensive AI-generated platform overview
```

#### Test 9.2: Performance and Edge Cases
```bash
# Test: Concurrent deployments
curl -X POST http://localhost:8080/v1/applications/checkout/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}' &

curl -X POST http://localhost:8080/v1/applications/payment/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}' &

wait
# Expect: Both deployments succeed without conflicts

# Test: Invalid requests
curl -X POST http://localhost:8080/v1/applications/nonexistent/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "dev"}'
# Expect: 404 Not Found

curl -X POST http://localhost:8080/v1/applications/checkout/deploy \
  -H "Content-Type: application/json" \
  -d '{"environment": "nonexistent"}'
# Expect: 400 Bad Request or 404 Not Found

# Test: AI error handling
curl -X POST http://localhost:8080/v3/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Deploy nonexistent-app to production"}'
# Expect: AI error explanation about application not existing
```

## Expected Final Platform State

After successful completion of all tests, the platform should have:

### Applications (4)
- **checkout**: 2 services, 3 versions, deployed to dev & production
- **payment**: 1 service, 2 versions, deployed to dev & production  
- **monitoring**: 2 services, 2 versions, deployed to dev
- **orders**: 1 service, 1 version, deployed through full workflow

### Environments (3)
- **dev**: Contains deployments of all applications
- **staging**: Contains selective deployments
- **production**: Contains policy-compliant deployments

### Resources (5 instances + 3 types)
- All resource types and instances created
- Proper linking between applications, services, and resources

### Policies (3)
- **no-direct-prod**: Enforced and tested
- **code-scan-required**: Enforced and tested  
- **resource-limits**: Created and available

### Releases
- At least one release created and deployed through staging workflow

### AI Integration
- All natural language interactions working
- Policy consultation and compliance checking
- Platform intelligence and recommendations

This comprehensive test plan demonstrates the full capabilities of the ZTDP platform, from basic resource management to advanced AI-native operations, with realistic policy enforcement scenarios.
