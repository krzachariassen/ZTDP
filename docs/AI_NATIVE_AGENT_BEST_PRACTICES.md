# AI-Native Agent Development Best Practices

## Executive Summary

Based on the refactoring of the Application Agent, we've identified key patterns and best practices for developing AI-native agents in the ZTDP platform. This document captures the lessons learned and provides guidance for future agent development and SDK creation.

## Current State Analysis

### Problems with Original Application Agent (1400+ lines)

1. **Mixed AI-native and fallback logic** - Agent had both AI calls and heuristic parsing
2. **Overly complex event handling** - Complex routing logic that should be simplified
3. **Inconsistent response patterns** - Different handlers created responses differently
4. **Too much boilerplate** - Repetitive code for parameter extraction, error handling
5. **Missing proper AI response parsing** - AI calls made but responses not properly parsed
6. **Poor separation of concerns** - Business logic mixed with AI parsing logic

### Success Patterns from Deployment Agent

The Deployment Agent demonstrates the correct AI-native pattern:
- Uses AI for all intent and parameter extraction
- No fallback or heuristic logic
- Clear confidence-based clarification requests
- Standardized response handling

## AI-Native Agent Development Best Practices

### 1. AI-First Architecture

**✅ DO: Use AI for all intent and parameter extraction**
```go
func (a *Agent) extractIntentAndParameters(ctx context.Context, userMessage string) (*AIResponse, error) {
    systemPrompt := `Parse user request and extract structured parameters...`
    response, err := a.aiProvider.CallAI(ctx, systemPrompt, userMessage)
    // Parse and validate JSON response
}
```

**❌ DON'T: Mix AI with fallback heuristics**
```go
// Bad pattern - mixed AI and heuristics
if aiResponse != nil {
    return parseAIResponse(aiResponse)
} else {
    return parseHeuristically(userMessage) // NO!
}
```

### 2. Confidence-Based Clarification

**✅ DO: Use AI confidence scores to trigger clarification**
```go
if response.Confidence < 0.8 {
    clarification := response.Clarification
    if clarification == "" {
        clarification = "Could you please be more specific?"
    }
    return a.createClarificationResponse(event, clarification), nil
}
```

**❌ DON'T: Guess or use default values for unclear requests**

### 3. Structured AI Responses

**✅ DO: Define clear response structures**
```go
type AIResponse struct {
    Action         string  `json:"action"`
    ApplicationName string  `json:"application_name,omitempty"`
    Confidence     float64 `json:"confidence"`
    Clarification  string  `json:"clarification,omitempty"`
}
```

**✅ DO: Handle malformed AI responses gracefully**
```go
var response AIResponse
if err := json.Unmarshal([]byte(aiResponseText), &response); err != nil {
    // Return low confidence instead of error
    return &AIResponse{
        Action:        "unknown",
        Confidence:    0.1,
        Clarification: "I had trouble understanding your request.",
    }, nil
}
```

### 4. Clean Event Handling

**✅ DO: Simple, clear event routing**
```go
func (a *Agent) handleEvent(ctx context.Context, event *Event) (*Event, error) {
    userMessage := extractUserMessage(event)
    aiResponse := extractIntentAndParameters(ctx, userMessage)
    
    switch aiResponse.Action {
    case "list": return a.handleList(ctx, event, aiResponse)
    case "create": return a.handleCreate(ctx, event, aiResponse)
    default: return a.createClarificationResponse(event, "Unknown action")
    }
}
```

**❌ DON'T: Complex nested routing logic**

### 5. Standardized Response Patterns

**✅ DO: Use consistent response structures**
```go
func (a *Agent) createSuccessResponse(originalEvent *Event, payload map[string]interface{}) *Event {
    return &Event{
        Subject: fmt.Sprintf("%s.response.%s", a.agentType, originalEvent.ID),
        Type:    fmt.Sprintf("%s.response", a.agentType),
        Payload: map[string]interface{}{
            "status":           "success",
            "correlation_id":   originalEvent.ID,
            "data":             payload,
        },
    }
}
```

### 6. AI Provider Validation

**✅ DO: Require AI provider for AI-native agents**
```go
func NewAgent(aiProvider ai.AIProvider) (*Agent, error) {
    if aiProvider == nil {
        return nil, fmt.Errorf("aiProvider is required for AI-native agent")
    }
    // ...
}
```

**❌ DON'T: Allow AI-native agents without AI providers**

## Framework Improvements Needed

Based on the refactoring experience, the agent framework needs these enhancements:

### 1. AI Parameter Extraction Helpers

```go
// Proposed framework helper
func (f *Framework) ExtractParametersWithAI(ctx context.Context, userMessage string, schema ParameterSchema) (*AIResponse, error)

// Usage in agents
response, err := f.ExtractParametersWithAI(ctx, userMessage, ApplicationParameterSchema)
```

### 2. Response Standardization

```go
// Proposed framework response builders
func (f *Framework) CreateSuccessResponse(originalEvent *Event, agentType string, data interface{}) *Event
func (f *Framework) CreateErrorResponse(originalEvent *Event, agentType string, err error) *Event
func (f *Framework) CreateClarificationResponse(originalEvent *Event, agentType string, message string) *Event
```

### 3. Event Routing Automation

```go
// Proposed framework auto-routing
type RouteConfig struct {
    Action  string
    Handler func(context.Context, *Event, *AIResponse) (*Event, error)
}

func (f *Framework) WithAutoRouting(routes []RouteConfig) *AgentBuilder
```

### 4. AI Response Parsing Utilities

```go
// Proposed framework utilities
func (f *Framework) ParseAIResponse(response string, target interface{}) error
func (f *Framework) ValidateAIConfidence(response *AIResponse, threshold float64) bool
```

### 5. Testing Utilities

```go
// Proposed framework test helpers
func (f *Framework) CreateMockAIProvider(responses map[string]string) ai.AIProvider
func (f *Framework) CreateTestEvent(userMessage string) *Event
```

## Clean Agent Implementation Guide

### Step 1: Define AI Response Structure
```go
type AIResponse struct {
    Action      string  `json:"action"`
    Parameters  map[string]interface{} `json:"parameters"`
    Confidence  float64 `json:"confidence"`
    Clarification string `json:"clarification,omitempty"`
}
```

### Step 2: Implement AI Parameter Extraction
```go
func (a *Agent) extractIntentAndParameters(ctx context.Context, userMessage string) (*AIResponse, error) {
    systemPrompt := buildSystemPrompt(a.capabilities)
    userPrompt := fmt.Sprintf("Parse: %s", userMessage)
    
    response, err := a.aiProvider.CallAI(ctx, systemPrompt, userPrompt)
    if err != nil {
        return nil, err
    }
    
    return parseAndValidateAIResponse(response)
}
```

### Step 3: Implement Clean Event Handling
```go
func (a *Agent) handleEvent(ctx context.Context, event *Event) (*Event, error) {
    userMessage := extractUserMessage(event)
    if userMessage == "" {
        return a.createClarificationResponse(event, "I need a user message")
    }
    
    aiResponse, err := a.extractIntentAndParameters(ctx, userMessage)
    if err != nil {
        return a.createErrorResponse(event, err)
    }
    
    if aiResponse.Confidence < 0.8 {
        return a.createClarificationResponse(event, aiResponse.Clarification)
    }
    
    return a.routeToHandler(ctx, event, aiResponse)
}
```

### Step 4: Implement Action Handlers
```go
func (a *Agent) handleCreate(ctx context.Context, event *Event, aiResponse *AIResponse) (*Event, error) {
    // Validate required parameters
    // Call domain service
    // Return success response
}
```

## Testing Best Practices

### 1. Use Real AI Provider in Tests
```go
func TestAgent_WithRealAI(t *testing.T) {
    aiProvider := createRealAIProvider()
    if aiProvider == nil {
        t.Skip("OPENAI_API_KEY not set")
    }
    // Test with real AI
}
```

### 2. Test AI Parameter Extraction
```go
func TestAgent_ParameterExtraction(t *testing.T) {
    tests := []struct {
        userMessage    string
        expectedAction string
        expectedConf   float64
    }{
        {"list all apps", "list", 0.9},
        {"create app foo", "create", 0.9},
    }
    // Test AI extraction directly
}
```

### 3. Test Confidence Levels
```go
func TestAgent_LowConfidence(t *testing.T) {
    response := agent.handleEvent(ctx, createEvent("unclear message"))
    assert.Equal(t, "clarification_needed", response.Payload["status"])
}
```

## Reference Implementation

See `/docs/clean_application_agent_demo.go` for a complete reference implementation demonstrating all best practices.

## Migration Strategy

### Phase 1: Create Clean Reference Implementation
- ✅ Identify best practices from Deployment Agent
- ✅ Create clean Application Agent following patterns
- ✅ Document lessons learned and patterns

### Phase 2: Framework Enhancements
- Add AI parameter extraction helpers
- Add response standardization utilities
- Add event routing automation
- Add testing utilities

### Phase 3: SDK Development
- Create agent development SDK based on patterns
- Provide templates and generators
- Add comprehensive documentation

### Phase 4: Migrate Existing Agents
- Refactor existing agents to use clean patterns
- Remove fallback logic
- Standardize response handling

## Key Takeaways

1. **AI-native means no fallback logic** - If AI can't extract parameters with high confidence, ask for clarification
2. **Confidence scores drive clarification** - Use AI confidence to determine when to ask for more information
3. **Structured responses are critical** - Define clear JSON schemas for AI responses
4. **Framework helpers reduce boilerplate** - Common patterns should be automated
5. **Real AI in tests is essential** - Mock AI defeats the purpose of AI-native development
6. **Clean separation of concerns** - Domain logic in services, AI as infrastructure tool

This clean pattern provides a solid foundation for AI-native agent development and sets the stage for SDK creation to make agent development more consistent and less error-prone.
