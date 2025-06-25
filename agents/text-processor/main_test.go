// Integration test for the text processing agent
package main

import (
	"context"
	"log"
	"testing"

	"github.com/ztdp/agents/text-processor/agent"
	"github.com/ztdp/agents/text-processor/textprocessor"
)

func TestAgentIntegration(t *testing.T) {
	// Test that we can create and configure an agent
	handler := textprocessor.NewTextProcessor()

	textAgent := agent.NewAgent(
		"test-agent-001",
		"Test Text Processing Agent",
		handler,
	)

	if textAgent.ID != "test-agent-001" {
		t.Errorf("Expected agent ID 'test-agent-001', got '%s'", textAgent.ID)
	}

	if textAgent.Name != "Test Text Processing Agent" {
		t.Errorf("Expected agent name 'Test Text Processing Agent', got '%s'", textAgent.Name)
	}

	expectedCapabilities := []string{
		"text-analysis",
		"word-count",
		"character-count",
		"text-formatting",
		"text-cleanup",
	}

	if len(textAgent.Capabilities) != len(expectedCapabilities) {
		t.Errorf("Expected %d capabilities, got %d", len(expectedCapabilities), len(textAgent.Capabilities))
	}

	for i, expected := range expectedCapabilities {
		if textAgent.Capabilities[i] != expected {
			t.Errorf("Expected capability '%s', got '%s'", expected, textAgent.Capabilities[i])
		}
	}
}

func TestDirectTaskProcessing(t *testing.T) {
	// Test that our handler can process tasks directly
	handler := textprocessor.NewTextProcessor()
	ctx := context.Background()

	task := agent.Task{
		ID:      "integration-test-1",
		Type:    "word-count",
		Content: "Hello world from our amazing text processing agent!",
	}

	result, err := handler.Process(ctx, task)
	if err != nil {
		t.Fatalf("Failed to process task: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected successful result, got failure: %s", result.Error)
	}

	wordCount, ok := result.Data["word_count"].(int)
	if !ok {
		t.Fatalf("Expected word_count to be int, got %T", result.Data["word_count"])
	}

	expectedWordCount := 8
	if wordCount != expectedWordCount {
		t.Errorf("Expected word count %d, got %d", expectedWordCount, wordCount)
	}

	expectedMessage := "Text contains 8 words"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	log.Printf("✅ Integration test passed: %s", result.Message)
}
