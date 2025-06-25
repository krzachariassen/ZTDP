# Text Processing Agent

A simple, testable agent that demonstrates the ZTDP Agent SDK capabilities.

## 🎯 **What This Agent Does**

The Text Processing Agent can:
- **Count words** in any text
- **Count characters** (with and without spaces)
- **Analyze text** comprehensively (words, sentences, lines, letters, digits)
- **Clean up text** (remove extra whitespace, normalize formatting)
- **Format text** (uppercase, lowercase, title case, sentence case)

## 🚀 **Quick Start**

### Build and Run
```bash
# Build the agent
go build -o text-processor

# Run the agent
./text-processor

# Or run directly
go run main.go
```

### Configuration
Set environment variables:
```bash
export ORCHESTRATOR_ADDRESS=localhost:50051
./text-processor
```

## 📋 **Capabilities**

| Capability | Description | Example Task |
|------------|-------------|--------------|
| `text-analysis` | Complete text analysis | Word count, sentences, lines, etc. |
| `word-count` | Count words in text | "Hello world" → 2 words |
| `character-count` | Count characters | "Hello world" → 11 chars (10 without spaces) |
| `text-formatting` | Format text | "hello" → "HELLO" (uppercase) |
| `text-cleanup` | Clean and normalize text | Remove extra spaces, normalize newlines |

## 🧪 **Testing**

### Run All Tests
```bash
go test ./...
```

### Run with Coverage
```bash
go test -cover ./...
```

### Run Benchmarks
```bash
go test -bench=. ./...
```

### Example Test Results
```
=== RUN   TestTextProcessor_WordCount
=== RUN   TestTextProcessor_WordCount/Empty_text
=== RUN   TestTextProcessor_WordCount/Single_word
=== RUN   TestTextProcessor_WordCount/Multiple_words
=== RUN   TestTextProcessor_WordCount/With_extra_spaces
=== RUN   TestTextProcessor_WordCount/Complex_sentence
--- PASS: TestTextProcessor_WordCount (0.00s)
    --- PASS: TestTextProcessor_WordCount/Empty_text (0.00s)
    --- PASS: TestTextProcessor_WordCount/Single_word (0.00s)
    --- PASS: TestTextProcessor_WordCount/Multiple_words (0.00s)
    --- PASS: TestTextProcessor_WordCount/With_extra_spaces (0.00s)
    --- PASS: TestTextProcessor_WordCount/Complex_sentence (0.00s)
PASS
```

## 🤖 **Agent SDK Features**

This agent demonstrates:

### ✅ **Simple Agent Creation**
```go
handler := textprocessor.NewTextProcessor()
textAgent := agent.NewAgent("text-processor-001", "Text Processing Agent", handler)
```

### ✅ **Capability Declaration**
```go
func (tp *TextProcessor) GetCapabilities() []string {
    return []string{
        "text-analysis",
        "word-count", 
        "character-count",
        "text-formatting",
        "text-cleanup",
    }
}
```

### ✅ **Task Processing**
```go
func (tp *TextProcessor) Process(ctx context.Context, task agent.Task) (*agent.Result, error) {
    switch task.Type {
    case "word-count":
        return tp.wordCount(task.Content)
    case "character-count":
        return tp.characterCount(task.Content)
    // ... more cases
    }
}
```

### ✅ **Graceful Lifecycle Management**
- Automatic connection to orchestrator
- Agent registration with capabilities
- Work loop for task processing
- Graceful shutdown handling

## 📝 **Example Tasks & Results**

### Word Count Task
```json
{
    "id": "task-001",
    "type": "word-count", 
    "content": "Hello world from AI orchestrator"
}
```

**Result:**
```json
{
    "success": true,
    "data": {
        "word_count": 5,
        "words": ["Hello", "world", "from", "AI", "orchestrator"]
    },
    "message": "Text contains 5 words"
}
```

### Text Analysis Task  
```json
{
    "id": "task-002",
    "type": "text-analysis",
    "content": "Hello world! How are you?"
}
```

**Result:**
```json
{
    "success": true,
    "data": {
        "word_count": 5,
        "character_count": 25,
        "character_count_no_spaces": 21,
        "sentence_count": 2,
        "line_count": 1,
        "letter_count": 19,
        "digit_count": 0
    },
    "message": "Analysis complete: 5 words, 25 characters, 2 sentences, 1 lines"
}
```

## 🎯 **End-to-End Demo Flow**

1. **User Request (via Chat UI):**
   ```
   "Count the words in this text: 'Hello world from AI orchestrator'"
   ```

2. **AI Orchestrator Response:**
   ```
   "I'll use my text processing agent to analyze that for you!"
   ```

3. **Agent Processing:**
   - Receives task with type `word-count`
   - Processes text: "Hello world from AI orchestrator"
   - Returns result: 5 words

4. **AI Orchestrator Reply:**
   ```
   "The text contains 5 words: Hello, world, from, AI, orchestrator"
   ```

## 🔧 **Architecture**

```
User Request → AI Orchestrator → Text Processing Agent
     ↑              ↓                      ↓
Chat UI ← AI Response ← Agent Result Processing
```

### Components:
- **Agent SDK**: Handles connection, registration, lifecycle
- **Text Processor**: Business logic for text processing tasks  
- **gRPC Client**: Communication with orchestrator
- **Task Handler**: Routes tasks to appropriate processing functions

## 🚀 **What Makes This Revolutionary**

1. **AI-Driven Task Assignment**: AI orchestrator intelligently routes text tasks
2. **Dynamic Capability Discovery**: Agent advertises its capabilities automatically
3. **Conversational Interface**: Natural language requests get processed seamlessly
4. **Error Handling**: Graceful failure handling with informative responses
5. **Extensible Design**: Easy to add new text processing capabilities

## 📊 **Performance**

Benchmark results on typical hardware:
```
BenchmarkTextProcessor_WordCount-8      	 2000000	       842 ns/op
BenchmarkTextProcessor_TextAnalysis-8   	  500000	      3024 ns/op
```

## 🔮 **Future Enhancements**

- [ ] **More Text Operations**: Regex matching, text extraction, summarization
- [ ] **Language Detection**: Automatic language identification
- [ ] **Sentiment Analysis**: Basic sentiment scoring
- [ ] **Text Validation**: Email, URL, phone number validation
- [ ] **Document Processing**: PDF, Word document text extraction

---

This agent demonstrates how the ZTDP Agent SDK makes it trivial to create powerful, testable agents that integrate seamlessly with AI orchestration! 🎉
