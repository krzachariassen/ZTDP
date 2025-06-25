package textprocessor

import (
	"context"
	"testing"

	"github.com/ztdp/agents/text-processor/agent"
)

func TestTextProcessor_GetCapabilities(t *testing.T) {
	tp := NewTextProcessor()
	capabilities := tp.GetCapabilities()

	expected := []string{
		"text-analysis",
		"word-count",
		"character-count",
		"text-formatting",
		"text-cleanup",
	}

	if len(capabilities) != len(expected) {
		t.Errorf("Expected %d capabilities, got %d", len(expected), len(capabilities))
	}

	for i, cap := range expected {
		if capabilities[i] != cap {
			t.Errorf("Expected capability %s, got %s", cap, capabilities[i])
		}
	}
}

func TestTextProcessor_WordCount(t *testing.T) {
	tp := NewTextProcessor()
	ctx := context.Background()

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"Empty text", "", 0},
		{"Single word", "hello", 1},
		{"Multiple words", "hello world", 2},
		{"With extra spaces", "  hello   world  ", 2},
		{"Complex sentence", "The quick brown fox jumps over the lazy dog", 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := agent.Task{
				ID:      "test-1",
				Type:    "word-count",
				Content: tt.text,
			}

			result, err := tp.Process(ctx, task)
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			if !result.Success {
				t.Fatalf("Expected success=true, got %v", result.Success)
			}

			wordCount, ok := result.Data["word_count"].(int)
			if !ok {
				t.Fatalf("Expected word_count to be int, got %T", result.Data["word_count"])
			}

			if wordCount != tt.expected {
				t.Errorf("Expected word count %d, got %d", tt.expected, wordCount)
			}
		})
	}
}

func TestTextProcessor_CharacterCount(t *testing.T) {
	tp := NewTextProcessor()
	ctx := context.Background()

	tests := []struct {
		name             string
		text             string
		expectedTotal    int
		expectedNoSpaces int
	}{
		{"Empty text", "", 0, 0},
		{"Single word", "hello", 5, 5},
		{"With spaces", "hello world", 11, 10},
		{"With punctuation", "Hello, world!", 13, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := agent.Task{
				ID:      "test-1",
				Type:    "character-count",
				Content: tt.text,
			}

			result, err := tp.Process(ctx, task)
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			if !result.Success {
				t.Fatalf("Expected success=true, got %v", result.Success)
			}

			charCount, ok := result.Data["character_count"].(int)
			if !ok {
				t.Fatalf("Expected character_count to be int, got %T", result.Data["character_count"])
			}

			charCountNoSpaces, ok := result.Data["character_count_no_spaces"].(int)
			if !ok {
				t.Fatalf("Expected character_count_no_spaces to be int, got %T", result.Data["character_count_no_spaces"])
			}

			if charCount != tt.expectedTotal {
				t.Errorf("Expected total char count %d, got %d", tt.expectedTotal, charCount)
			}

			if charCountNoSpaces != tt.expectedNoSpaces {
				t.Errorf("Expected char count without spaces %d, got %d", tt.expectedNoSpaces, charCountNoSpaces)
			}
		})
	}
}

func TestTextProcessor_TextAnalysis(t *testing.T) {
	tp := NewTextProcessor()
	ctx := context.Background()

	task := agent.Task{
		ID:      "test-1",
		Type:    "text-analysis",
		Content: "Hello world! This is a test. How are you?",
	}

	result, err := tp.Process(ctx, task)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success=true, got %v", result.Success)
	}

	// Check that all expected fields are present
	expectedFields := []string{
		"word_count",
		"character_count",
		"character_count_no_spaces",
		"sentence_count",
		"line_count",
		"letter_count",
		"digit_count",
		"words",
	}

	for _, field := range expectedFields {
		if _, ok := result.Data[field]; !ok {
			t.Errorf("Expected field %s not found in result", field)
		}
	}

	// Verify some specific values
	if wordCount := result.Data["word_count"].(int); wordCount != 9 {
		t.Errorf("Expected word count 9, got %d", wordCount)
	}

	if sentenceCount := result.Data["sentence_count"].(int); sentenceCount != 3 {
		t.Errorf("Expected sentence count 3, got %d", sentenceCount)
	}
}

func TestTextProcessor_TextCleanup(t *testing.T) {
	tp := NewTextProcessor()
	ctx := context.Background()

	task := agent.Task{
		ID:      "test-1",
		Type:    "text-cleanup",
		Content: "  Hello    world  \n\n\n\n  This   is   messy   \n\n\n\n\n  ",
	}

	result, err := tp.Process(ctx, task)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success=true, got %v", result.Success)
	}

	cleaned, ok := result.Data["cleaned_text"].(string)
	if !ok {
		t.Fatalf("Expected cleaned_text to be string, got %T", result.Data["cleaned_text"])
	}

	expected := "Hello world\n\nThis is messy"
	if cleaned != expected {
		t.Errorf("Expected cleaned text %q, got %q", expected, cleaned)
	}
}

func TestTextProcessor_TextFormatting(t *testing.T) {
	tp := NewTextProcessor()
	ctx := context.Background()

	tests := []struct {
		name     string
		text     string
		format   string
		expected string
	}{
		{"Uppercase", "hello world", "uppercase", "HELLO WORLD"},
		{"Lowercase", "HELLO WORLD", "lowercase", "hello world"},
		{"Title case", "hello world", "title", "Hello World"},
		{"Sentence case", "HELLO WORLD", "sentence", "Hello world"},
		{"No format", "Hello World", "", "Hello World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := map[string]interface{}{}
			if tt.format != "" {
				context["format"] = tt.format
			}

			task := agent.Task{
				ID:      "test-1",
				Type:    "text-formatting",
				Content: tt.text,
				Context: context,
			}

			result, err := tp.Process(ctx, task)
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			if !result.Success {
				t.Fatalf("Expected success=true, got %v", result.Success)
			}

			formatted, ok := result.Data["formatted_text"].(string)
			if !ok {
				t.Fatalf("Expected formatted_text to be string, got %T", result.Data["formatted_text"])
			}

			if formatted != tt.expected {
				t.Errorf("Expected formatted text %q, got %q", tt.expected, formatted)
			}
		})
	}
}

func TestTextProcessor_UnsupportedTaskType(t *testing.T) {
	tp := NewTextProcessor()
	ctx := context.Background()

	task := agent.Task{
		ID:      "test-1",
		Type:    "unsupported-task",
		Content: "test content",
	}

	_, err := tp.Process(ctx, task)
	if err == nil {
		t.Fatal("Expected error for unsupported task type, got nil")
	}

	expectedError := "unsupported task type: unsupported-task"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

// Benchmark tests
func BenchmarkTextProcessor_WordCount(b *testing.B) {
	tp := NewTextProcessor()
	ctx := context.Background()

	task := agent.Task{
		ID:      "bench-1",
		Type:    "word-count",
		Content: "The quick brown fox jumps over the lazy dog. This is a longer sentence with more words to test performance.",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tp.Process(ctx, task)
		if err != nil {
			b.Fatalf("Process() error = %v", err)
		}
	}
}

func BenchmarkTextProcessor_TextAnalysis(b *testing.B) {
	tp := NewTextProcessor()
	ctx := context.Background()

	task := agent.Task{
		ID:   "bench-1",
		Type: "text-analysis",
		Content: `Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
		Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
		Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris 
		nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in 
		reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tp.Process(ctx, task)
		if err != nil {
			b.Fatalf("Process() error = %v", err)
		}
	}
}
