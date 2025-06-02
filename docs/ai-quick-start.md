# AI-Powered Deployments - Quick Start Guide

## Overview

ZTDP now includes AI-driven deployment planning using OpenAI GPT models. This guide shows how developers can leverage AI for intelligent deployments.

## Setup

### 1. Configure AI Provider

```bash
# Set OpenAI API key (required)
export OPENAI_API_KEY="sk-your-openai-api-key"

# Optional: Configure model (defaults to gpt-4)
export OPENAI_MODEL="gpt-4"

# Optional: Set AI provider (defaults to openai)
export AI_PROVIDER="openai"
```

### 2. Start ZTDP

```bash
# AI will be automatically enabled if OPENAI_API_KEY is set
./ztdp-api
```

## How It Works

### Traditional vs AI Deployment

**Before (Traditional)**:
```
Developer Request → Static Topological Sort → Deploy
```

**Now (AI-Enhanced)**:
```
Developer Request → AI Analysis → Intelligent Plan → Deploy
                      ↓ (fallback)
                   Traditional Plan
```

### AI Decision Process

1. **Context Gathering**: AI analyzes your application graph, dependencies, and policies
2. **Intelligent Planning**: GPT-4 generates an optimal deployment sequence
3. **Reasoning**: AI explains why each step is necessary
4. **Execution**: Deploy with AI-generated plan

## Using AI Features

### 1. Regular Deployments (AI-Enhanced)

```bash
# Your existing deployment commands now use AI automatically
curl -X POST /v1/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "my-app",
    "environment_id": "production"
  }'
```

**What happens**:
- AI analyzes your app's dependencies
- Generates optimal deployment order
- Considers policies and constraints
- Falls back to traditional planning if AI fails

### 2. Generate AI Plans

```bash
# Get an AI-generated deployment plan
curl -X POST /v1/ai/plans/generate \
  -H "Content-Type: application/json" \
  -d '{
    "app_name": "checkout-service",
    "edge_types": ["deploy", "create", "owns"],
    "timeout": 30
  }'
```

**Response**:
```json
{
  "plan": {
    "steps": [
      {
        "id": "step-1",
        "action": "deploy",
        "target": "database",
        "dependencies": [],
        "reasoning": "Database must be available before app deployment"
      },
      {
        "id": "step-2", 
        "action": "deploy",
        "target": "checkout-service",
        "dependencies": ["step-1"],
        "reasoning": "Service deployment after database is ready"
      }
    ],
    "strategy": "rolling_deployment"
  },
  "reasoning": "Sequential deployment minimizes risk while ensuring dependencies are met",
  "confidence": 0.92
}
```

### 3. Evaluate Deployment Policies

```bash
# Check if deployment complies with policies
curl -X POST /v1/ai/policies/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "checkout-service",
    "environment_id": "production"
  }'
```

### 4. Optimize Existing Plans

```bash
# Improve an existing deployment plan
curl -X POST /v1/ai/plans/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "current_plan": [...],
    "application_id": "checkout-service"
  }'
```

## Monitoring AI

### Check AI Status

```bash
curl /v1/ai/provider/status
```

**Response**:
```json
{
  "name": "OpenAI GPT-4",
  "available": true,
  "capabilities": ["plan_generation", "policy_evaluation", "plan_optimization"],
  "model": "gpt-4"
}
```

### View AI Metrics

```bash
curl "/v1/ai/metrics?hours=24"
```

## Benefits for Developers

### 1. Intelligent Dependency Resolution
- AI understands complex dependency chains
- Optimizes deployment order automatically
- Reduces deployment failures

### 2. Policy-Aware Planning
- AI considers all deployment policies
- Suggests compliance fixes
- Prevents policy violations

### 3. Explainable Decisions
- AI explains why each step is needed
- Provides confidence scores
- Transparent reasoning process

### 4. Graceful Fallback
- System works even without AI
- Automatic fallback to traditional planning
- No service disruption

## Example: AI vs Traditional Planning

### Traditional Approach
```
Input: Deploy "web-app" 
Output: [database, cache, web-app]  # Simple topological sort
```

### AI Approach
```
Input: Deploy "web-app"
AI Analysis: 
- Database needs migration scripts
- Cache requires warm-up time
- Web-app has health checks
- Production has blue-green policy

Output: 
1. Deploy database with migration
2. Run database migrations
3. Deploy cache and warm-up
4. Deploy web-app (blue environment)
5. Run health checks
6. Switch traffic to blue
7. Scale down green

Reasoning: "Blue-green deployment ensures zero downtime while 
           respecting production policies and warm-up requirements"
Confidence: 0.89
```

## Troubleshooting

### AI Not Working

1. **Check API Key**:
   ```bash
   echo $OPENAI_API_KEY  # Should show your key
   ```

2. **Check Provider Status**:
   ```bash
   curl /v1/ai/provider/status
   ```

3. **View Logs**:
   ```bash
   # Look for AI-related log messages
   tail -f ztdp.log | grep -i "ai\|brain"
   ```

### Fallback Mode

If AI is unavailable, ZTDP automatically falls back to traditional planning:

```
{"level":"WARN","message":"⚠️ Failed to initialize AI brain, will use traditional planning"}
{"level":"WARN","message":"AI planning failed, falling back to traditional planner"}
```

This ensures your deployments continue working even without AI.

## Best Practices

### 1. Use Descriptive Node Names
```yaml
# Good: AI can understand the purpose
nodes:
  - id: "user-service-database"
  - id: "user-service-api"
  - id: "user-service-cache"

# Bad: AI has limited context
nodes:
  - id: "db1"
  - id: "api1" 
  - id: "cache1"
```

### 2. Define Clear Dependencies
```yaml
edges:
  - from: "user-service-api"
    to: "user-service-database"
    type: "depends_on"
    metadata:
      reason: "API requires database for user data"
```

### 3. Set Deployment Policies
```yaml
policies:
  - name: "production-blue-green"
    environment: "production"
    strategy: "blue_green"
    validation_required: true
```

## Next Steps

1. **Set up OpenAI API key** to enable AI features
2. **Try AI plan generation** for your applications
3. **Monitor AI performance** using metrics endpoint
4. **Compare AI vs traditional** deployment outcomes
5. **Provide feedback** on AI reasoning quality

## Support

- **Documentation**: `/docs/ai-deployment-architecture.md`
- **API Reference**: `/docs/swagger.yaml`
- **Issues**: Check logs and provider status first
- **Performance**: AI responses typically take 2-5 seconds
