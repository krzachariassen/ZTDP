# ZTDP Platform Demo Scripts

This directory contains demonstration scripts showcasing different aspects of the ZTDP platform.

## Demo Scripts

### 1. Traditional API Demo (`graph_demo_api.go`)

**Purpose**: Demonstrates the traditional REST API approach for platform operations.

**What it tests**:
- Basic CRUD operations for applications, services, environments
- Traditional deployment workflows
- Graph database operations
- Standard REST endpoint functionality

**Usage**:
```bash
# Start the ZTDP server first
go run cmd/server/main.go

# Run the traditional demo (in another terminal)
go run test/controlplane/graph_demo_api.go

# Force recreate the graph
go run test/controlplane/graph_demo_api.go --force
```

**Output**: Creates a complete graph structure and tests basic deployment operations.

### 2. AI-Native Platform Demo (`ai_native_demo_enhanced.go`)

**Purpose**: Demonstrates the AI-native capabilities where AI is the primary interface.

**What it tests**:
- 🧠 **AI-Powered Deployment Planning** - Uses `?plan=true` query parameter for AI-generated plans
- 🔧 **AI Plan Optimization** - Uses `?optimize=true` for AI-optimized deployment strategies  
- 📊 **AI Impact Analysis** - Uses `?analyze=true` for AI-driven risk assessment
- 🛡️ **AI Policy Evaluation** - Tests AI-enhanced policy validation
- 💬 **Natural Language Operations** - Tests conversational AI interface
- 🚨 **AI Error Guidance** - Tests AI-provided error resolution guidance
- 🗣️ **Conversational Deployment** - Tests multi-step AI conversations

**Usage**:
```bash
# Set up AI provider (required for full functionality)
export OPENAI_API_KEY="your-api-key-here"

# Start the ZTDP server with AI enabled
go run cmd/server/main.go

# Run the AI-native demo (in another terminal)
go run test/controlplane/ai_native_demo_enhanced.go
```

**Requirements**:
- OpenAI API key (set via `OPENAI_API_KEY` environment variable)
- ZTDP server running with AI provider configured
- Test applications and environments will be created automatically

**Expected Output**:
```
🤖 AI-Native Platform Demo
==================================================
✅ AI Provider: openai
🧠 Demo 1: AI-Powered Deployment Planning
⏱️  Response time: 15.763s  # Indicates real AI processing
✅ AI deployment plan generated successfully!
🧠 Planning source: ai_enhanced
📋 Plan contains 1 steps
```

## Key Differences

| Feature | Traditional Demo | AI-Native Demo |
|---------|------------------|----------------|
| **Interface** | REST API calls with structured JSON | Natural language + enhanced query parameters |
| **Planning** | Basic template-based | AI-generated with reasoning |
| **Error Handling** | Standard HTTP errors | AI-provided guidance and suggestions |
| **User Experience** | Technical API calls | Conversational and intuitive |
| **Response Times** | ~1ms (immediate) | ~15-25s (AI processing) |
| **Intelligence** | Rule-based logic | AI reasoning and adaptation |

## Architecture Demonstration

### Traditional Approach (graph_demo_api.go)
```
User → REST API → Handler → Domain Service → Response
```

### AI-Native Approach (ai_native_demo_enhanced.go)
```
User → Natural Language/Enhanced API → Handler → Domain Service → AI Provider → Response with Reasoning
```

## Running Both Demos

To see the full platform capabilities, run both demos:

```bash
# Terminal 1: Start server
export OPENAI_API_KEY="your-key"
go run cmd/server/main.go

# Terminal 2: Traditional demo
go run test/controlplane/graph_demo_api.go

# Terminal 3: AI-native demo  
go run test/controlplane/ai_native_demo_enhanced.go
```

## Demo Results Analysis

### Traditional Demo Success Indicators:
- ✅ Fast response times (~1ms)
- ✅ All CRUD operations successful
- ✅ Standard deployment workflows working
- ✅ Graph consistency maintained

### AI-Native Demo Success Indicators:
- ✅ AI provider detected and available
- ✅ Long response times (15-25s) indicating real AI processing
- ✅ AI-generated plans with reasoning
- ✅ Enhanced query parameters working (`?plan=true`, `?optimize=true`)
- ⚠️ Some features may need further enhancement (chat endpoints, policy evaluation)

## Future Enhancements

The AI-native demo reveals areas for potential improvement:

1. **Enhanced Chat Interface** - Improve natural language processing endpoints
2. **AI Error Guidance** - Add AI-powered error resolution suggestions
3. **Policy AI Integration** - Complete AI integration for policy evaluation
4. **Conversational Flows** - Improve multi-step conversation handling
5. **Impact Analysis** - Enhance AI-driven impact assessment capabilities

Both demos together showcase the platform's evolution from API-first to AI-native, demonstrating the powerful capabilities of having AI as the primary interface while maintaining traditional API compatibility.
