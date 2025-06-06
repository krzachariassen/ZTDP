# Testing Strategies

## Overview

This document outlines comprehensive testing strategies for the ZTDP AI-native platform, ensuring robust quality assurance across all architectural layers while maintaining the clean architecture principles and supporting AI-enhanced functionality.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test-Driven Development (TDD)](#test-driven-development-tdd)
3. [Testing Pyramid Strategy](#testing-pyramid-strategy)
4. [Domain Service Testing](#domain-service-testing)
5. [AI Component Testing](#ai-component-testing)
6. [API Testing](#api-testing)
7. [Event System Testing](#event-system-testing)
8. [Policy Testing](#policy-testing)
9. [Integration Testing](#integration-testing)
10. [Testing Tools and Framework](#testing-tools-and-framework)
11. [Mock Strategies](#mock-strategies)
12. [Performance Testing](#performance-testing)
13. [Testing Best Practices](#testing-best-practices)

---

## Testing Philosophy

### Core Principles

1. **Test-First Development**: Write tests before implementation
2. **Fast Feedback Loops**: Tests should run quickly and provide immediate feedback
3. **Clean Test Code**: Tests are first-class citizens and should be well-structured
4. **AI Integration Testing**: Special considerations for AI-enhanced functionality
5. **Domain Isolation**: Test domain logic independently from infrastructure
6. **Event-Driven Testing**: Validate event flows and communication patterns

### Testing Mindset

```go
// ✅ Test behavior, not implementation
func TestDeploymentService_PlanDeployment_GeneratesValidPlan(t *testing.T) {
    // Focus on what the service should do, not how it does it
}

// ❌ Don't test internal implementation details
func TestDeploymentService_PlanDeployment_CallsSpecificAIMethod(t *testing.T) {
    // This tests implementation, not behavior
}
```

---

## Test-Driven Development (TDD)

### Red-Green-Refactor Cycle

#### 1. Red Phase: Write Failing Test
```go
func TestDeploymentService_PlanDeployment_ValidApplication(t *testing.T) {
    // Arrange
    service := setupDeploymentService(t)
    app := "test-app"
    
    // Act
    plan, err := service.PlanDeployment(context.Background(), app)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, plan)
    assert.Equal(t, app, plan.ApplicationName)
    assert.True(t, len(plan.Steps) > 0)
}
```

#### 2. Green Phase: Make Test Pass
```go
func (s *DeploymentService) PlanDeployment(ctx context.Context, app string) (*Plan, error) {
    // Minimal implementation to make test pass
    return &Plan{
        ApplicationName: app,
        Steps: []Step{
            {Name: "basic-deployment", Type: "deploy"},
        },
    }, nil
}
```

#### 3. Refactor Phase: Improve Code
```go
func (s *DeploymentService) PlanDeployment(ctx context.Context, app string) (*Plan, error) {
    // Validate input
    if err := s.validateApplication(ctx, app); err != nil {
        return nil, fmt.Errorf("application validation failed: %w", err)
    }
    
    // Use AI for enhanced planning
    if s.aiProvider != nil {
        if plan, err := s.generateAIEnhancedPlan(ctx, app); err == nil {
            return plan, nil
        }
    }
    
    // Fallback to basic planning
    return s.generateBasicPlan(app), nil
}
```

### TDD with AI Components

#### Testing AI-Enhanced Domain Services
```go
func TestDeploymentService_PlanDeployment_WithAI(t *testing.T) {
    tests := []struct {
        name           string
        aiResponse     string
        aiError        error
        expectedSteps  int
        expectFallback bool
    }{
        {
            name:       "AI provides valid plan",
            aiResponse: `{"steps": [{"name": "deploy", "type": "blue-green"}]}`,
            expectedSteps: 1,
            expectFallback: false,
        },
        {
            name:           "AI fails, fallback to basic plan",
            aiError:        errors.New("AI timeout"),
            expectedSteps:  1,
            expectFallback: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mockAI := &MockAIProvider{}
            if tt.aiError != nil {
                mockAI.On("CallAI", mock.Anything, mock.Anything, mock.Anything).
                    Return("", tt.aiError)
            } else {
                mockAI.On("CallAI", mock.Anything, mock.Anything, mock.Anything).
                    Return(tt.aiResponse, nil)
            }
            
            service := NewDeploymentService(
                &MockGraph{},
                mockAI,
                &MockPolicyService{},
                &MockEventBus{},
            )
            
            // Act
            plan, err := service.PlanDeployment(context.Background(), "test-app")
            
            // Assert
            assert.NoError(t, err)
            assert.Len(t, plan.Steps, tt.expectedSteps)
            
            if tt.expectFallback {
                assert.Equal(t, "basic-deployment", plan.Steps[0].Name)
            }
        })
    }
}
```

---

## Testing Pyramid Strategy

### Unit Tests (70% of tests)

**Focus**: Individual components in isolation

```go
// Domain service unit tests
func TestPolicyService_ValidateDeployment(t *testing.T) {
    service := NewPolicyService(&MockGraph{}, &MockAIProvider{})
    
    err := service.ValidateDeployment(context.Background(), "app", "prod")
    
    assert.NoError(t, err)
}

// Pure function unit tests
func TestCalculateRiskScore(t *testing.T) {
    risk := CalculateRiskScore(0.8, 0.2, 0.1)
    assert.Equal(t, 0.37, risk)
}
```

### Integration Tests (20% of tests)

**Focus**: Component interactions and workflows

```go
func TestDeploymentWorkflow_EndToEnd(t *testing.T) {
    // Setup real components (but isolated environment)
    graph := graph.NewMemoryGraph()
    eventBus := events.NewMemoryBus()
    
    deploymentService := NewDeploymentService(graph, nil, nil, eventBus)
    
    // Test complete workflow
    plan, err := deploymentService.PlanDeployment(context.Background(), "test-app")
    assert.NoError(t, err)
    
    result, err := deploymentService.ExecuteDeployment(context.Background(), plan)
    assert.NoError(t, err)
    assert.Equal(t, "completed", result.Status)
}
```

### End-to-End Tests (10% of tests)

**Focus**: Complete user journeys through API

```go
func TestAPI_DeploymentFlow_E2E(t *testing.T) {
    // Start test server
    server := setupTestServer(t)
    defer server.Close()
    
    // Create application
    app := createTestApplication(t, server)
    
    // Plan deployment
    planResp := planDeployment(t, server, app.Name, "staging")
    assert.Equal(t, http.StatusOK, planResp.StatusCode)
    
    // Execute deployment
    deployResp := executeDeployment(t, server, planResp.Plan.ID)
    assert.Equal(t, http.StatusOK, deployResp.StatusCode)
    
    // Verify deployment status
    status := getDeploymentStatus(t, server, deployResp.Deployment.ID)
    assert.Equal(t, "completed", status.State)
}
```

---

## Domain Service Testing

### Service Testing Pattern

```go
// Test structure for domain services
type DeploymentServiceTestSuite struct {
    suite.Suite
    service       *DeploymentService
    mockGraph     *MockGraph
    mockAI        *MockAIProvider
    mockPolicy    *MockPolicyService
    mockEvents    *MockEventBus
}

func (suite *DeploymentServiceTestSuite) SetupTest() {
    suite.mockGraph = &MockGraph{}
    suite.mockAI = &MockAIProvider{}
    suite.mockPolicy = &MockPolicyService{}
    suite.mockEvents = &MockEventBus{}
    
    suite.service = NewDeploymentService(
        suite.mockGraph,
        suite.mockAI,
        suite.mockPolicy,
        suite.mockEvents,
    )
}

func (suite *DeploymentServiceTestSuite) TestPlanDeployment_ValidInput() {
    // Given
    app := "test-app"
    suite.mockGraph.On("GetApplication", app).Return(&Application{Name: app}, nil)
    suite.mockPolicy.On("ValidateDeployment", mock.Anything, app, mock.Anything).Return(nil)
    
    // When
    plan, err := suite.service.PlanDeployment(context.Background(), app)
    
    // Then
    suite.NoError(err)
    suite.NotNil(plan)
    suite.Equal(app, plan.ApplicationName)
    
    // Verify events were emitted
    suite.mockEvents.AssertCalled(suite.T(), "Emit", "deployment.planning.started", mock.Anything)
    suite.mockEvents.AssertCalled(suite.T(), "Emit", "deployment.planning.completed", mock.Anything)
}

func TestDeploymentServiceSuite(t *testing.T) {
    suite.Run(t, new(DeploymentServiceTestSuite))
}
```

### Error Scenario Testing

```go
func (suite *DeploymentServiceTestSuite) TestPlanDeployment_ErrorScenarios() {
    tests := []struct {
        name          string
        app           string
        setupMocks    func()
        expectedError string
    }{
        {
            name: "invalid application",
            app:  "",
            setupMocks: func() {
                // No mocks needed
            },
            expectedError: "application name cannot be empty",
        },
        {
            name: "policy violation",
            app:  "blocked-app",
            setupMocks: func() {
                suite.mockPolicy.On("ValidateDeployment", mock.Anything, "blocked-app", mock.Anything).
                    Return(errors.New("policy violation"))
            },
            expectedError: "policy validation failed",
        },
    }
    
    for _, tt := range tests {
        suite.Run(tt.name, func() {
            // Setup
            suite.SetupTest() // Reset mocks
            tt.setupMocks()
            
            // Execute
            plan, err := suite.service.PlanDeployment(context.Background(), tt.app)
            
            // Verify
            suite.Error(err)
            suite.Nil(plan)
            suite.Contains(err.Error(), tt.expectedError)
        })
    }
}
```

---

## AI Component Testing

### AI Provider Interface Testing

```go
func TestAIProvider_CallAI(t *testing.T) {
    tests := []struct {
        name           string
        systemPrompt   string
        userPrompt     string
        mockResponse   string
        mockError      error
        expectedResult string
        expectError    bool
    }{
        {
            name:         "successful AI call",
            systemPrompt: "You are a deployment planner",
            userPrompt:   "Plan deployment for app",
            mockResponse: "deployment plan response",
            expectedResult: "deployment plan response",
            expectError:  false,
        },
        {
            name:        "AI provider timeout",
            systemPrompt: "You are a deployment planner",
            userPrompt:   "Plan deployment for app",
            mockError:    errors.New("timeout"),
            expectError:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            provider := &MockAIProvider{}
            if tt.mockError != nil {
                provider.On("CallAI", mock.Anything, tt.systemPrompt, tt.userPrompt).
                    Return("", tt.mockError)
            } else {
                provider.On("CallAI", mock.Anything, tt.systemPrompt, tt.userPrompt).
                    Return(tt.mockResponse, nil)
            }
            
            // Execute
            result, err := provider.CallAI(context.Background(), tt.systemPrompt, tt.userPrompt)
            
            // Verify
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedResult, result)
            }
            
            provider.AssertExpectations(t)
        })
    }
}
```

### AI Integration Testing

```go
func TestAIIntegration_RealProvider(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping AI integration test in short mode")
    }
    
    // Setup real AI provider for integration testing
    provider := ai.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))
    
    systemPrompt := "You are a deployment planner. Respond with JSON."
    userPrompt := "Create a deployment plan for a web application."
    
    response, err := provider.CallAI(context.Background(), systemPrompt, userPrompt)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, response)
    
    // Verify response can be parsed as JSON
    var plan map[string]interface{}
    err = json.Unmarshal([]byte(response), &plan)
    assert.NoError(t, err)
}
```

---

## API Testing

### HTTP Handler Testing

```go
func TestDeploymentHandler_PlanDeployment(t *testing.T) {
    tests := []struct {
        name           string
        queryParams    string
        mockSetup      func(*MockDeploymentService)
        expectedStatus int
        expectedBody   string
    }{
        {
            name:        "valid request",
            queryParams: "app=test-app",
            mockSetup: func(mock *MockDeploymentService) {
                plan := &Plan{ApplicationName: "test-app"}
                mock.On("PlanDeployment", mock.Anything, "test-app").Return(plan, nil)
            },
            expectedStatus: http.StatusOK,
            expectedBody:   `{"applicationName":"test-app"}`,
        },
        {
            name:           "missing app parameter",
            queryParams:    "",
            mockSetup:      func(mock *MockDeploymentService) {},
            expectedStatus: http.StatusBadRequest,
            expectedBody:   "app parameter required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockService := &MockDeploymentService{}
            tt.mockSetup(mockService)
            
            handler := &DeploymentHandler{
                deploymentService: mockService,
            }
            
            // Create request
            req := httptest.NewRequest("GET", "/plan?"+tt.queryParams, nil)
            w := httptest.NewRecorder()
            
            // Execute
            handler.PlanDeployment(w, req)
            
            // Verify
            assert.Equal(t, tt.expectedStatus, w.Code)
            if tt.expectedBody != "" {
                assert.Contains(t, w.Body.String(), tt.expectedBody)
            }
            
            mockService.AssertExpectations(t)
        })
    }
}
```

### API Integration Testing

```go
func TestAPI_DeploymentEndpoints(t *testing.T) {
    // Setup test server with real components
    server := setupTestServer(t)
    defer server.Close()
    
    t.Run("complete deployment flow", func(t *testing.T) {
        // 1. Plan deployment
        planReq := fmt.Sprintf("%s/api/v1/deployments/plan?app=test-app", server.URL)
        resp, err := http.Get(planReq)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        var planResp struct {
            Plan *Plan `json:"plan"`
        }
        err = json.NewDecoder(resp.Body).Decode(&planResp)
        assert.NoError(t, err)
        assert.NotNil(t, planResp.Plan)
        
        // 2. Execute deployment
        executeReq := fmt.Sprintf("%s/api/v1/deployments/execute", server.URL)
        body := strings.NewReader(fmt.Sprintf(`{"planId":"%s"}`, planResp.Plan.ID))
        resp, err = http.Post(executeReq, "application/json", body)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusAccepted, resp.StatusCode)
    })
}
```

---

## Event System Testing

### Event Publishing Tests

```go
func TestEventBus_PublishSubscribe(t *testing.T) {
    // Setup
    eventBus := events.NewMemoryBus()
    
    // Subscribe to events
    received := make(chan *events.Event, 1)
    err := eventBus.Subscribe("deployment.started", func(event *events.Event) {
        received <- event
    })
    assert.NoError(t, err)
    
    // Publish event
    event := &events.Event{
        Type:    "deployment.started",
        Source:  "test",
        Subject: "test-app",
        Payload: map[string]interface{}{"app": "test-app"},
    }
    
    err = eventBus.Publish(event)
    assert.NoError(t, err)
    
    // Verify event received
    select {
    case receivedEvent := <-received:
        assert.Equal(t, event.Type, receivedEvent.Type)
        assert.Equal(t, event.Subject, receivedEvent.Subject)
    case <-time.After(1 * time.Second):
        t.Fatal("Event not received within timeout")
    }
}
```

### Event-Driven Service Testing

```go
func TestDeploymentService_EventEmission(t *testing.T) {
    // Setup
    mockEventBus := &MockEventBus{}
    service := NewDeploymentService(
        &MockGraph{},
        &MockAIProvider{},
        &MockPolicyService{},
        mockEventBus,
    )
    
    // Configure mocks
    mockEventBus.On("Emit", "deployment.planning.started", mock.Anything).Return(nil)
    mockEventBus.On("Emit", "deployment.planning.completed", mock.Anything).Return(nil)
    
    // Execute
    _, err := service.PlanDeployment(context.Background(), "test-app")
    assert.NoError(t, err)
    
    // Verify events were emitted
    mockEventBus.AssertCalled(t, "Emit", "deployment.planning.started", mock.MatchedBy(func(payload map[string]interface{}) bool {
        return payload["app"] == "test-app"
    }))
    
    mockEventBus.AssertCalled(t, "Emit", "deployment.planning.completed", mock.Anything)
}
```

---

## Testing Tools and Framework

### Required Dependencies

```go
// go.mod testing dependencies
require (
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    github.com/testcontainers/testcontainers-go v0.20.1
    github.com/go-chi/chi/v5 v5.0.8
)
```

### Test Configuration

```go
// internal/testing/config.go
type TestConfig struct {
    UseRealAI       bool
    AITimeout       time.Duration
    DatabaseURL     string
    RedisURL        string
    EventBusType    string
}

func NewTestConfig() *TestConfig {
    return &TestConfig{
        UseRealAI:    os.Getenv("TEST_REAL_AI") == "true",
        AITimeout:    30 * time.Second,
        DatabaseURL:  os.Getenv("TEST_DATABASE_URL"),
        RedisURL:     os.Getenv("TEST_REDIS_URL"),
        EventBusType: "memory", // Use memory for tests by default
    }
}
```

### Test Helpers

```go
// internal/testing/helpers.go
func SetupTestServer(t *testing.T) *httptest.Server {
    t.Helper()
    
    // Setup test components
    graph := graph.NewMemoryGraph()
    eventBus := events.NewMemoryBus()
    
    // Create services
    deploymentService := deployments.NewService(graph, nil, nil, eventBus)
    policyService := policies.NewService(graph, nil, eventBus)
    
    // Create handlers
    deploymentHandler := handlers.NewDeploymentHandler(deploymentService)
    policyHandler := handlers.NewPolicyHandler(policyService)
    
    // Setup router
    router := chi.NewRouter()
    router.Route("/api/v1", func(r chi.Router) {
        r.Mount("/deployments", deploymentHandler.Routes())
        r.Mount("/policies", policyHandler.Routes())
    })
    
    return httptest.NewServer(router)
}

func CreateTestApplication(t *testing.T, server *httptest.Server) *Application {
    t.Helper()
    
    app := &Application{
        Name:        "test-app-" + generateRandomID(),
        Environment: "test",
        Language:    "go",
    }
    
    // Create application via API
    body, _ := json.Marshal(app)
    resp, err := http.Post(server.URL+"/api/v1/applications", "application/json", strings.NewReader(string(body)))
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, resp.StatusCode)
    
    var created Application
    err = json.NewDecoder(resp.Body).Decode(&created)
    require.NoError(t, err)
    
    return &created
}
```

---

## Mock Strategies

### Interface-Based Mocking

```go
//go:generate mockery --name=AIProvider --output=mocks
type AIProvider interface {
    CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error)
    GetProviderInfo() *ProviderInfo
    Close() error
}

//go:generate mockery --name=Graph --output=mocks
type Graph interface {
    AddNode(node *Node) error
    GetNode(id string) (*Node, error)
    AddEdge(from, to, relation string) error
    GetNeighbors(id string) ([]*Node, error)
}
```

### Manual Mocks for Complex Behavior

```go
type MockAIProvider struct {
    mock.Mock
    responses map[string]string // For deterministic responses
}

func (m *MockAIProvider) CallAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
    args := m.Called(ctx, systemPrompt, userPrompt)
    
    // Return deterministic response if configured
    if response, exists := m.responses[userPrompt]; exists {
        return response, args.Error(1)
    }
    
    return args.String(0), args.Error(1)
}

func (m *MockAIProvider) SetResponse(prompt, response string) {
    if m.responses == nil {
        m.responses = make(map[string]string)
    }
    m.responses[prompt] = response
}
```

---

## Performance Testing

### Benchmark Tests

```go
func BenchmarkDeploymentService_PlanDeployment(b *testing.B) {
    service := setupBenchmarkService(b)
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.PlanDeployment(ctx, "test-app")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkAIProvider_CallAI(b *testing.B) {
    if testing.Short() {
        b.Skip("Skipping AI benchmark in short mode")
    }
    
    provider := ai.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))
    defer provider.Close()
    
    ctx := context.Background()
    systemPrompt := "You are a helpful assistant"
    userPrompt := "Hello"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := provider.CallAI(ctx, systemPrompt, userPrompt)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Load Testing

```go
func TestAPI_LoadTesting(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }
    
    server := setupTestServer(t)
    defer server.Close()
    
    const numRequests = 100
    const concurrency = 10
    
    var wg sync.WaitGroup
    errors := make(chan error, numRequests)
    
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for j := 0; j < numRequests/concurrency; j++ {
                resp, err := http.Get(server.URL + "/api/v1/health")
                if err != nil {
                    errors <- err
                    continue
                }
                resp.Body.Close()
                
                if resp.StatusCode != http.StatusOK {
                    errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
                }
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    errorCount := 0
    for err := range errors {
        t.Logf("Request error: %v", err)
        errorCount++
    }
    
    errorRate := float64(errorCount) / float64(numRequests)
    assert.Less(t, errorRate, 0.01, "Error rate should be less than 1%")
}
```

---

## Testing Best Practices

### Test Organization

#### File Structure
```
internal/
├── deployments/
│   ├── service.go
│   ├── service_test.go
│   ├── mocks/
│   │   ├── mock_graph.go
│   │   └── mock_ai_provider.go
│   └── testdata/
│       ├── valid_plan.json
│       └── invalid_app.json
```

#### Naming Conventions
```go
// Test function naming: TestUnitOfWork_Scenario_ExpectedBehavior
func TestDeploymentService_PlanDeployment_ValidApplication(t *testing.T) {}
func TestDeploymentService_PlanDeployment_InvalidApplication_ReturnsError(t *testing.T) {}
func TestDeploymentService_PlanDeployment_AIFailure_FallsBackToBasicPlan(t *testing.T) {}

// Benchmark naming: BenchmarkUnitOfWork_Scenario
func BenchmarkDeploymentService_PlanDeployment(b *testing.B) {}
func BenchmarkAIProvider_CallAI_LargePrompt(b *testing.B) {}
```

### Test Data Management

#### Test Fixtures
```go
// internal/deployments/testdata.go
func ValidApplicationFixture() *Application {
    return &Application{
        Name:        "test-app",
        Environment: "staging",
        Language:    "go",
        Framework:   "chi",
    }
}

func ValidPlanFixture() *Plan {
    return &Plan{
        ApplicationName: "test-app",
        Environment:    "staging",
        Steps: []Step{
            {Name: "build", Type: "build", Duration: time.Minute},
            {Name: "deploy", Type: "deploy", Duration: 5 * time.Minute},
            {Name: "verify", Type: "health-check", Duration: 30 * time.Second},
        },
    }
}
```

#### Test Data Builders
```go
type ApplicationBuilder struct {
    app *Application
}

func NewApplicationBuilder() *ApplicationBuilder {
    return &ApplicationBuilder{
        app: &Application{
            Name:        "default-app",
            Environment: "test",
            Language:    "go",
        },
    }
}

func (b *ApplicationBuilder) WithName(name string) *ApplicationBuilder {
    b.app.Name = name
    return b
}

func (b *ApplicationBuilder) WithEnvironment(env string) *ApplicationBuilder {
    b.app.Environment = env
    return b
}

func (b *ApplicationBuilder) Build() *Application {
    return b.app
}

// Usage in tests
func TestDeploymentService_PlanDeployment_ProductionApp(t *testing.T) {
    app := NewApplicationBuilder().
        WithName("prod-app").
        WithEnvironment("production").
        Build()
    
    // Use app in test
}
```

### Common Testing Patterns

#### Table-Driven Tests
```go
func TestDeploymentService_ValidateApplication(t *testing.T) {
    tests := []struct {
        name        string
        app         *Application
        expectError bool
        errorMsg    string
    }{
        {
            name:        "valid application",
            app:         ValidApplicationFixture(),
            expectError: false,
        },
        {
            name: "missing name",
            app: &Application{
                Environment: "staging",
                Language:    "go",
            },
            expectError: true,
            errorMsg:    "application name is required",
        },
        {
            name: "invalid environment",
            app: &Application{
                Name:        "test-app",
                Environment: "invalid",
                Language:    "go",
            },
            expectError: true,
            errorMsg:    "invalid environment",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := setupDeploymentService(t)
            
            err := service.ValidateApplication(context.Background(), tt.app)
            
            if tt.expectError {
                assert.Error(t, err)
                if tt.errorMsg != "" {
                    assert.Contains(t, err.Error(), tt.errorMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### Subtests for Complex Scenarios
```go
func TestDeploymentService_PlanDeployment(t *testing.T) {
    service := setupDeploymentService(t)
    
    t.Run("with AI provider", func(t *testing.T) {
        t.Run("AI success", func(t *testing.T) {
            // Test AI success scenario
        })
        
        t.Run("AI failure with fallback", func(t *testing.T) {
            // Test AI failure scenario
        })
        
        t.Run("AI timeout", func(t *testing.T) {
            // Test AI timeout scenario
        })
    })
    
    t.Run("without AI provider", func(t *testing.T) {
        t.Run("basic plan generation", func(t *testing.T) {
            // Test basic plan generation
        })
    })
}
```

### Test Execution Commands

#### Development Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestDeploymentService_PlanDeployment ./internal/deployments

# Run tests in parallel
go test -parallel 4 ./...

# Run short tests only (skip integration tests)
go test -short ./...
```

#### CI/CD Testing
```bash
# Run all tests with coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Run tests with timeout
go test -timeout 10m ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

#### AI Integration Testing
```bash
# Run tests including real AI calls
TEST_REAL_AI=true OPENAI_API_KEY=$OPENAI_API_KEY go test ./internal/ai

# Run load tests
go test -run TestAPI_LoadTesting -timeout 5m ./api/handlers
```

This comprehensive testing strategy ensures robust quality assurance across all layers of the ZTDP AI-native platform while maintaining clean architecture principles and supporting the unique requirements of AI-enhanced functionality.
