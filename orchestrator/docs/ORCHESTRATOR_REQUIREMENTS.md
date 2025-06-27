# 🎯 SIMPLIFIED ORCHESTRATOR REQUIREMENTS

## User Journey
**Input**: "Please analyse this excel sheet, structure the data, put it into a powerpoint and send it to x@y.z"

## Expected Orchestrator Behavior

### 1. Intent Recognition
- **Intent**: `multi_step_document_workflow`
- **Complexity**: `complex` (requires multiple agents)
- **Confidence**: `85%` (clear request with specific steps)

### 2. Agent Discovery
- **Required Agents**: 
  - `excel_processor` (excel-analysis, data-extraction)
  - `data_structurer` (data-formatting, data-transformation)  
  - `powerpoint_creator` (presentation-creation, slide-generation)
  - `email_sender` (email-delivery, attachment-handling)

### 3. Execution Plan
```
Step 1: Analyze Excel sheet for data patterns and structure
Step 2: Structure and format the extracted data for presentation
Step 3: Create PowerPoint presentation with structured data
Step 4: Send completed PowerPoint to x@y.z
```

### 4. Agent Coordination
```
Primary Agent: excel_processor (starts the workflow)
Supporting Agents: 
- data_structurer (receives excel output)
- powerpoint_creator (receives structured data)
- email_sender (receives powerpoint file)
Workflow Dependencies: Sequential execution with data passing
```

### 5. Response Format
```json
{
  "message": "I'll help you process that Excel file through to email delivery. Here's my plan:\n\n[execution_plan]\n\nAgent Coordination:\n[agent_coordination]",
  "intent": "multi_step_document_workflow",
  "confidence": 0.85,
  "actions": [
    {"type": "execution_step", "description": "Step 1: Analyze Excel sheet..."},
    {"type": "execution_step", "description": "Step 2: Structure and format..."},
    {"type": "execution_step", "description": "Step 3: Create PowerPoint..."},
    {"type": "execution_step", "description": "Step 4: Send completed PowerPoint..."}
  ],
  "context": {
    "execution_plan": "[multi-step plan]",
    "agent_coordination": "[agent coordination details]", 
    "execution_plan_id": "plan_user123_1234567890"
  }
}
```

## Key Principles
1. **Keep It Simple**: Direct agent discovery, straightforward coordination
2. **Multi-Agent Support**: Handle complex workflows with multiple agents
3. **Dynamic Registration**: Discover agents from registry, don't hardcode
4. **Sequential Coordination**: Pass data between agents in logical order
5. **Progress Tracking**: Store execution plan for monitoring
