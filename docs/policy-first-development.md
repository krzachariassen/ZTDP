# Policy-First Development in ZTDP

## Introduction

ZTDP implements a policy-first approach where all operations are validated against defined governance policies before execution. This ensures compliance, security, and operational excellence across the entire platform.

## Core Policy Concepts

### 1. Policy-First Principle

**Definition**: All system operations must be validated against defined policies before execution.

**Implementation**: Every state-changing operation includes policy validation as the first step:

```go
func (s *DeploymentService) DeployApplication(ctx context.Context, app, env string) error {
    // FIRST: Policy validation before any action
    if err := s.policyService.ValidateDeployment(ctx, app, env); err != nil {
        return fmt.Errorf("deployment policy violation: %w", err)
    }
    
    // ONLY proceed after policy approval
    return s.executeDeployment(ctx, app, env)
}
```

### 2. Policy Types

ZTDP supports multiple types of policies:

#### Security Policies
- **Authentication**: User and service authentication requirements
- **Authorization**: Role-based access control and permissions
- **Data Protection**: Encryption, data handling, and privacy requirements
- **Network Security**: Network isolation, firewall rules, and communication policies

#### Compliance Policies  
- **Regulatory**: SOX, GDPR, HIPAA, and other regulatory requirements
- **Industry Standards**: ISO 27001, SOC 2, PCI DSS compliance
- **Internal Governance**: Company-specific policies and procedures
- **Audit Requirements**: Logging, monitoring, and reporting standards

#### Operational Policies
- **Resource Limits**: CPU, memory, storage, and network constraints
- **Deployment Windows**: Allowed deployment times and environments
- **Change Management**: Approval workflows and rollback requirements
- **Quality Gates**: Testing, validation, and approval criteria

#### Business Policies
- **Cost Management**: Budget limits and resource optimization
- **Service Level Agreements**: Uptime, performance, and availability requirements
- **Data Governance**: Data retention, backup, and recovery policies
- **Business Continuity**: Disaster recovery and business continuity requirements

## Policy Architecture

### 1. Policy Engine

The policy engine evaluates policies using graph-based analysis:

```go
type PolicyEngine interface {
    // Policy evaluation
    Evaluate(ctx context.Context, request *PolicyRequest) (*PolicyResult, error)
    EvaluateBatch(ctx context.Context, requests []*PolicyRequest) ([]*PolicyResult, error)
    
    // Policy management
    LoadPolicy(ctx context.Context, policy *Policy) error
    UpdatePolicy(ctx context.Context, policyID string, updates *PolicyUpdates) error
    DeletePolicy(ctx context.Context, policyID string) error
    
    // Policy queries
    GetPolicies(ctx context.Context, filter *PolicyFilter) ([]*Policy, error)
    GetPolicyImpact(ctx context.Context, policyID string) (*PolicyImpact, error)
}

type PolicyRequest struct {
    Operation   string                 `json:"operation"`    // e.g., "deployment.create"
    Resource    string                 `json:"resource"`     // e.g., "application:checkout-service"
    Subject     string                 `json:"subject"`      // e.g., "user:john.doe"
    Context     map[string]interface{} `json:"context"`      // Additional context
    Environment string                 `json:"environment"`  // e.g., "production"
}

type PolicyResult struct {
    Allowed     bool                   `json:"allowed"`
    Violations  []*PolicyViolation     `json:"violations"`
    Requirements []*PolicyRequirement   `json:"requirements"`
    Confidence  float64                `json:"confidence"`
}
```

### 2. Graph-Based Policy Evaluation

Policies are evaluated using the platform's graph database:

```go
// Policy evaluation using graph traversal
func (e *GraphPolicyEngine) Evaluate(ctx context.Context, request *PolicyRequest) (*PolicyResult, error) {
    // Get relevant policies for the operation
    policies, err := e.getPoliciesForOperation(request.Operation, request.Environment)
    if err != nil {
        return nil, err
    }
    
    result := &PolicyResult{
        Allowed:    true,
        Violations: []*PolicyViolation{},
        Requirements: []*PolicyRequirement{},
    }
    
    // Evaluate each applicable policy
    for _, policy := range policies {
        violation, err := e.evaluatePolicy(ctx, policy, request)
        if err != nil {
            return nil, err
        }
        
        if violation != nil {
            result.Allowed = false
            result.Violations = append(result.Violations, violation)
        }
    }
    
    return result, nil
}

func (e *GraphPolicyEngine) evaluatePolicy(ctx context.Context, policy *Policy, request *PolicyRequest) (*PolicyViolation, error) {
    // Use graph queries to evaluate policy conditions
    switch policy.Type {
    case "security.access":
        return e.evaluateAccessPolicy(ctx, policy, request)
    case "deployment.window":
        return e.evaluateDeploymentWindow(ctx, policy, request)
    case "resource.limits":
        return e.evaluateResourceLimits(ctx, policy, request)
    default:
        return nil, fmt.Errorf("unknown policy type: %s", policy.Type)
    }
}
```

### 3. Policy as Code

Policies are defined as code and stored in version control:

```yaml
# policies/deployment/production-approval.yaml
apiVersion: policy.ztdp.dev/v1
kind: Policy
metadata:
  name: production-deployment-approval
  namespace: deployment
spec:
  type: deployment.approval
  scope:
    environments: ["production", "staging"]
    applications: ["*"]
  rules:
    - name: require-approval
      condition: |
        environment == "production" && 
        !hasApproval(application, user) &&
        !isEmergencyDeployment()
      action: deny
      message: "Production deployments require approval"
      
    - name: require-testing
      condition: |
        !hasPassedTests(application, version) ||
        !hasSecurityScan(application, version)
      action: deny
      message: "All tests and security scans must pass"
      
    - name: deployment-window
      condition: |
        environment == "production" &&
        !isWithinDeploymentWindow(currentTime())
      action: deny
      message: "Production deployments only allowed during maintenance windows"
```

## Policy Service Implementation

### 1. Policy Service Structure

```go
type PolicyService struct {
    engine      PolicyEngine
    graph       graph.Graph
    eventBus    events.Bus
    repository  PolicyRepository
    aiProvider  ai.Provider // For intelligent policy evaluation
}

func NewPolicyService(
    engine PolicyEngine,
    graph graph.Graph,
    eventBus events.Bus,
    repository PolicyRepository,
    aiProvider ai.Provider,
) *PolicyService {
    return &PolicyService{
        engine:     engine,
        graph:      graph,
        eventBus:   eventBus,
        repository: repository,
        aiProvider: aiProvider,
    }
}
```

### 2. Policy Validation Methods

```go
// Core policy validation method
func (s *PolicyService) ValidateOperation(ctx context.Context, operation string, resource string, context map[string]interface{}) error {
    request := &PolicyRequest{
        Operation: operation,
        Resource:  resource,
        Subject:   extractSubject(ctx),
        Context:   context,
        Environment: extractEnvironment(context),
    }
    
    // Emit policy evaluation started event
    s.eventBus.Emit("policy.evaluation.started", map[string]interface{}{
        "operation": operation,
        "resource":  resource,
        "subject":   request.Subject,
    })
    
    result, err := s.engine.Evaluate(ctx, request)
    if err != nil {
        s.eventBus.Emit("policy.evaluation.error", map[string]interface{}{
            "operation": operation,
            "error":     err.Error(),
        })
        return fmt.Errorf("policy evaluation failed: %w", err)
    }
    
    // Emit evaluation result
    s.eventBus.Emit("policy.evaluation.completed", map[string]interface{}{
        "operation":  operation,
        "allowed":    result.Allowed,
        "violations": result.Violations,
    })
    
    if !result.Allowed {
        return &PolicyViolationError{
            Operation:  operation,
            Resource:   resource,
            Violations: result.Violations,
        }
    }
    
    return nil
}

// Domain-specific validation methods
func (s *PolicyService) ValidateDeployment(ctx context.Context, app, env string) error {
    context := map[string]interface{}{
        "application": app,
        "environment": env,
        "timestamp":   time.Now(),
    }
    
    return s.ValidateOperation(ctx, "deployment.create", fmt.Sprintf("application:%s", app), context)
}

func (s *PolicyService) ValidateApplicationCreation(ctx context.Context, appContract *contracts.ApplicationContract) error {
    context := map[string]interface{}{
        "application": appContract,
        "resources":   appContract.Resources,
        "environment": appContract.Environment,
    }
    
    return s.ValidateOperation(ctx, "application.create", fmt.Sprintf("application:%s", appContract.Name), context)
}

func (s *PolicyService) ValidateResourceAccess(ctx context.Context, resource, action string) error {
    context := map[string]interface{}{
        "action":    action,
        "timestamp": time.Now(),
    }
    
    return s.ValidateOperation(ctx, fmt.Sprintf("resource.%s", action), resource, context)
}
```

### 3. AI-Enhanced Policy Evaluation

Use AI to provide intelligent policy evaluation and recommendations:

```go
func (s *PolicyService) EvaluateWithAI(ctx context.Context, request *PolicyRequest) (*PolicyResult, error) {
    // Standard policy evaluation first
    result, err := s.engine.Evaluate(ctx, request)
    if err != nil {
        return nil, err
    }
    
    // If violations found, use AI for intelligent analysis
    if !result.Allowed && s.aiProvider != nil {
        aiResult, err := s.getAIRecommendations(ctx, request, result)
        if err == nil {
            result.AIRecommendations = aiResult.Recommendations
            result.AlternativeApproaches = aiResult.Alternatives
        }
    }
    
    return result, nil
}

func (s *PolicyService) getAIRecommendations(ctx context.Context, request *PolicyRequest, result *PolicyResult) (*AIAnalysisResult, error) {
    prompt := s.buildPolicyAnalysisPrompt(request, result)
    
    response, err := s.aiProvider.CallAI(ctx, prompt.System, prompt.User)
    if err != nil {
        return nil, err
    }
    
    return s.parseAIRecommendations(response)
}

func (s *PolicyService) buildPolicyAnalysisPrompt(request *PolicyRequest, result *PolicyResult) *AIPrompt {
    systemPrompt := `You are a policy analysis expert. Analyze policy violations and provide:
    1. Clear explanation of why the policy was violated
    2. Recommendations to resolve the violations
    3. Alternative approaches that would comply with policies
    4. Risk assessment of potential workarounds`
    
    userPrompt := fmt.Sprintf(`Analyze this policy violation:
    Operation: %s
    Resource: %s
    Environment: %s
    Violations: %v
    
    Provide actionable recommendations to resolve these violations.`,
        request.Operation,
        request.Resource,
        request.Environment,
        result.Violations,
    )
    
    return &AIPrompt{
        System: systemPrompt,
        User:   userPrompt,
    }
}
```

## Policy Integration Patterns

### 1. Service Integration Pattern

All domain services integrate policy validation:

```go
// Deployment Service with policy integration
func (s *DeploymentService) CreateDeploymentPlan(ctx context.Context, app, env string) (*DeploymentPlan, error) {
    // 1. Policy validation FIRST
    if err := s.policyService.ValidateDeployment(ctx, app, env); err != nil {
        return nil, fmt.Errorf("deployment policy violation: %w", err)
    }
    
    // 2. Create plan only after policy approval
    plan, err := s.generatePlan(ctx, app, env)
    if err != nil {
        return nil, err
    }
    
    // 3. Validate plan against policies
    planContext := map[string]interface{}{
        "plan":        plan,
        "application": app,
        "environment": env,
    }
    
    if err := s.policyService.ValidateOperation(ctx, "deployment.plan.execute", fmt.Sprintf("application:%s", app), planContext); err != nil {
        return nil, fmt.Errorf("deployment plan policy violation: %w", err)
    }
    
    return plan, nil
}

// Application Service with policy integration
func (s *ApplicationService) CreateApplication(ctx context.Context, contract *contracts.ApplicationContract) error {
    // Policy validation before creation
    if err := s.policyService.ValidateApplicationCreation(ctx, contract); err != nil {
        return fmt.Errorf("application creation policy violation: %w", err)
    }
    
    // Create application only after policy approval
    return s.createApplicationInternal(ctx, contract)
}
```

### 2. API Handler Integration

API handlers enforce policies at the request level:

```go
func (h *DeploymentHandler) CreateDeployment(w http.ResponseWriter, r *http.Request) {
    var request DeploymentRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        WriteJSONError(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Policy validation in handler
    if err := h.policyService.ValidateDeployment(r.Context(), request.Application, request.Environment); err != nil {
        var policyErr *PolicyViolationError
        if errors.As(err, &policyErr) {
            WriteJSONError(w, policyErr.Error(), http.StatusForbidden)
            return
        }
        WriteJSONError(w, "Policy evaluation failed", http.StatusInternalServerError)
        return
    }
    
    // Delegate to service only after policy approval
    deployment, err := h.deploymentService.CreateDeployment(r.Context(), &request)
    if err != nil {
        WriteJSONError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(deployment)
}
```

## Policy Event Handling

### 1. Policy Violation Events

All policy violations emit structured events:

```go
type PolicyViolationEvent struct {
    Operation   string             `json:"operation"`
    Resource    string             `json:"resource"`
    Subject     string             `json:"subject"`
    Violations  []*PolicyViolation `json:"violations"`
    Timestamp   time.Time          `json:"timestamp"`
    Context     map[string]interface{} `json:"context"`
}

func (s *PolicyService) emitViolationEvent(request *PolicyRequest, violations []*PolicyViolation) {
    event := &PolicyViolationEvent{
        Operation:  request.Operation,
        Resource:   request.Resource,
        Subject:    request.Subject,
        Violations: violations,
        Timestamp:  time.Now(),
        Context:    request.Context,
    }
    
    s.eventBus.Emit("policy.violation", event)
}
```

### 2. Policy Change Events

Policy updates emit events for system-wide awareness:

```go
func (s *PolicyService) UpdatePolicy(ctx context.Context, policyID string, updates *PolicyUpdates) error {
    oldPolicy, err := s.repository.GetPolicy(ctx, policyID)
    if err != nil {
        return err
    }
    
    newPolicy, err := s.repository.UpdatePolicy(ctx, policyID, updates)
    if err != nil {
        return err
    }
    
    // Emit policy change event
    changeEvent := &PolicyChangeEvent{
        PolicyID:    policyID,
        OldPolicy:   oldPolicy,
        NewPolicy:   newPolicy,
        ChangedBy:   extractSubject(ctx),
        Timestamp:   time.Now(),
        ChangeType:  determineChangeType(oldPolicy, newPolicy),
    }
    
    s.eventBus.Emit("policy.changed", changeEvent)
    
    // Trigger impact analysis
    go s.analyzePolicyImpact(ctx, changeEvent)
    
    return nil
}
```

## Policy Testing

### 1. Policy Unit Testing

Test policy evaluation logic:

```go
func TestPolicyService_ValidateDeployment(t *testing.T) {
    tests := []struct {
        name        string
        app         string
        env         string
        setupMocks  func(*MockPolicyEngine)
        wantErr     bool
        expectedErr error
    }{
        {
            name: "production deployment without approval",
            app:  "checkout-service",
            env:  "production",
            setupMocks: func(mpe *MockPolicyEngine) {
                mpe.On("Evaluate", mock.Anything, mock.MatchedBy(func(req *PolicyRequest) bool {
                    return req.Operation == "deployment.create" && 
                           req.Environment == "production"
                })).Return(&PolicyResult{
                    Allowed: false,
                    Violations: []*PolicyViolation{{
                        Policy:  "production-approval",
                        Message: "Production deployments require approval",
                    }},
                }, nil)
            },
            wantErr:     true,
            expectedErr: &PolicyViolationError{},
        },
        {
            name: "development deployment allowed",
            app:  "test-app",
            env:  "development",
            setupMocks: func(mpe *MockPolicyEngine) {
                mpe.On("Evaluate", mock.Anything, mock.Anything).Return(&PolicyResult{
                    Allowed: true,
                }, nil)
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockEngine := &MockPolicyEngine{}
            tt.setupMocks(mockEngine)
            
            service := NewPolicyService(mockEngine, nil, nil, nil, nil)
            err := service.ValidateDeployment(context.Background(), tt.app, tt.env)
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.expectedErr != nil {
                    assert.IsType(t, tt.expectedErr, err)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Policy Integration Testing

Test end-to-end policy enforcement:

```go
func TestDeploymentWorkflow_PolicyEnforcement(t *testing.T) {
    // Setup test environment with real policy engine
    policyEngine := NewTestPolicyEngine()
    policyService := NewPolicyService(policyEngine, testGraph, testEventBus, testRepository, nil)
    deploymentService := NewDeploymentService(testGraph, testAI, policyService, testEventBus)
    
    // Load test policies
    productionPolicy := &Policy{
        Name: "production-approval",
        Type: "deployment.approval",
        Rules: []PolicyRule{{
            Condition: "environment == 'production' && !hasApproval(application, user)",
            Action:    "deny",
        }},
    }
    policyEngine.LoadPolicy(context.Background(), productionPolicy)
    
    // Test policy enforcement
    ctx := context.WithValue(context.Background(), "user", "test-user")
    
    // Should fail without approval
    err := deploymentService.CreateDeployment(ctx, "test-app", "production")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "policy violation")
    
    // Add approval and retry
    testGraph.AddApproval("test-app", "test-user", "manager")
    err = deploymentService.CreateDeployment(ctx, "test-app", "production")
    assert.NoError(t, err)
}
```

## Policy Monitoring and Compliance

### 1. Compliance Reporting

Generate compliance reports based on policy enforcement:

```go
func (s *PolicyService) GenerateComplianceReport(ctx context.Context, timeRange TimeRange) (*ComplianceReport, error) {
    // Get all policy evaluations in time range
    evaluations, err := s.repository.GetPolicyEvaluations(ctx, timeRange)
    if err != nil {
        return nil, err
    }
    
    report := &ComplianceReport{
        TimeRange:   timeRange,
        TotalEvaluations: len(evaluations),
        Violations:  []*PolicyViolation{},
        Compliance:  make(map[string]*ComplianceMetrics),
    }
    
    // Analyze evaluations
    for _, eval := range evaluations {
        if !eval.Result.Allowed {
            report.Violations = append(report.Violations, eval.Result.Violations...)
        }
        
        // Update compliance metrics by policy type
        policyType := extractPolicyType(eval.Policies)
        if _, exists := report.Compliance[policyType]; !exists {
            report.Compliance[policyType] = &ComplianceMetrics{}
        }
        
        metrics := report.Compliance[policyType]
        metrics.TotalEvaluations++
        if eval.Result.Allowed {
            metrics.Compliant++
        } else {
            metrics.Violations++
        }
    }
    
    // Calculate compliance percentages
    for _, metrics := range report.Compliance {
        metrics.ComplianceRate = float64(metrics.Compliant) / float64(metrics.TotalEvaluations)
    }
    
    return report, nil
}
```

### 2. Real-Time Compliance Monitoring

Monitor compliance in real-time using events:

```go
type ComplianceMonitor struct {
    policyService *PolicyService
    eventBus      events.Bus
    alerter       AlertService
    metrics       map[string]*ComplianceMetrics
    mu            sync.RWMutex
}

func (cm *ComplianceMonitor) Start(ctx context.Context) error {
    // Subscribe to policy events
    return cm.eventBus.Subscribe("compliance-monitor", 
        []string{"policy.violation", "policy.evaluation.completed"}, 
        cm.handlePolicyEvent)
}

func (cm *ComplianceMonitor) handlePolicyEvent(ctx context.Context, event *events.Event) error {
    switch event.Type {
    case "policy.violation":
        return cm.handleViolation(event)
    case "policy.evaluation.completed":
        return cm.updateMetrics(event)
    }
    return nil
}

func (cm *ComplianceMonitor) handleViolation(event *events.Event) error {
    violation := event.Data["violation"].(*PolicyViolation)
    
    // Check if this is a critical violation
    if violation.Severity == "critical" {
        alert := &Alert{
            Type:        "policy_violation",
            Severity:    "critical",
            Message:     violation.Message,
            Resource:    violation.Resource,
            Timestamp:   time.Now(),
        }
        
        return cm.alerter.SendAlert(alert)
    }
    
    return nil
}
```

## Policy Best Practices

### 1. Policy Design Principles

**Principle of Least Privilege**: Grant minimum necessary permissions
```yaml
# Good: Specific permissions
permissions:
  - resource: "application:checkout-service"
    actions: ["read", "deploy"]
    environments: ["development"]

# Bad: Overly broad permissions  
permissions:
  - resource: "*"
    actions: ["*"]
    environments: ["*"]
```

**Fail Secure**: Default to deny when policies are unclear
```go
func (e *PolicyEngine) Evaluate(ctx context.Context, request *PolicyRequest) (*PolicyResult, error) {
    result := &PolicyResult{
        Allowed: false, // Default to deny
    }
    
    // Only allow if explicitly permitted by policy
    for _, policy := range e.getPolicies(request) {
        if policy.Allows(request) {
            result.Allowed = true
            break
        }
    }
    
    return result, nil
}
```

**Policy Versioning**: Version policies for rollback capability
```yaml
apiVersion: policy.ztdp.dev/v1
kind: Policy
metadata:
  name: production-deployment
  version: "2.1.0"
  previousVersion: "2.0.0"
```

### 2. Performance Considerations

**Policy Caching**: Cache frequently evaluated policies
```go
type CachedPolicyEngine struct {
    underlying PolicyEngine
    cache      *cache.LRUCache
    ttl        time.Duration
}

func (cpe *CachedPolicyEngine) Evaluate(ctx context.Context, request *PolicyRequest) (*PolicyResult, error) {
    cacheKey := generateCacheKey(request)
    
    if cached, found := cpe.cache.Get(cacheKey); found {
        if result, ok := cached.(*PolicyResult); ok && time.Since(result.EvaluatedAt) < cpe.ttl {
            return result, nil
        }
    }
    
    result, err := cpe.underlying.Evaluate(ctx, request)
    if err != nil {
        return nil, err
    }
    
    result.EvaluatedAt = time.Now()
    cpe.cache.Set(cacheKey, result)
    
    return result, nil
}
```

**Batch Evaluation**: Evaluate multiple policies together
```go
func (s *PolicyService) ValidateBatch(ctx context.Context, operations []Operation) error {
    requests := make([]*PolicyRequest, len(operations))
    for i, op := range operations {
        requests[i] = &PolicyRequest{
            Operation: op.Type,
            Resource:  op.Resource,
            Subject:   extractSubject(ctx),
            Context:   op.Context,
        }
    }
    
    results, err := s.engine.EvaluateBatch(ctx, requests)
    if err != nil {
        return err
    }
    
    // Check all results
    for i, result := range results {
        if !result.Allowed {
            return fmt.Errorf("operation %d failed policy validation: %v", i, result.Violations)
        }
    }
    
    return nil
}
```

## Related Documentation

- **[Architecture Overview](architecture-overview.md)** - High-level system design
- **[Domain-Driven Design](domain-driven-design.md)** - Domain service integration
- **[Clean Architecture Principles](clean-architecture-principles.md)** - Service design patterns
- **[Event-Driven Architecture](event-driven-architecture.md)** - Policy event handling
