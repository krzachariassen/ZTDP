// Package textprocessor implements a simple text processing agent
package textprocessor

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/ztdp/agents/text-processor/agent"
)

// TextProcessor handles text processing tasks
type TextProcessor struct{}

// NewTextProcessor creates a new text processing handler
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{}
}

// GetCapabilities returns the capabilities of this agent
func (tp *TextProcessor) GetCapabilities() []string {
	return []string{
		"text-analysis",
		"word-count",
		"character-count",
		"text-formatting",
		"text-cleanup",
	}
}

// Process handles incoming text processing tasks
func (tp *TextProcessor) Process(ctx context.Context, task agent.Task) (*agent.Result, error) {
	switch task.Type {
	case "word-count":
		return tp.wordCount(task.Content)
	case "character-count":
		return tp.characterCount(task.Content)
	case "text-analysis":
		return tp.textAnalysis(task.Content)
	case "text-cleanup":
		return tp.textCleanup(task.Content)
	case "text-formatting":
		return tp.textFormatting(task.Content, task.Context)
	default:
		return nil, fmt.Errorf("unsupported task type: %s", task.Type)
	}
}

// wordCount counts words in the given text
func (tp *TextProcessor) wordCount(text string) (*agent.Result, error) {
	if text == "" {
		return &agent.Result{
			Success: true,
			Data: map[string]interface{}{
				"word_count": 0,
			},
			Message: "Text is empty, word count is 0",
		}, nil
	}

	words := strings.Fields(strings.TrimSpace(text))
	count := len(words)

	return &agent.Result{
		Success: true,
		Data: map[string]interface{}{
			"word_count": count,
			"words":      words,
		},
		Message: fmt.Sprintf("Text contains %d words", count),
	}, nil
}

// characterCount counts characters in the given text
func (tp *TextProcessor) characterCount(text string) (*agent.Result, error) {
	charCount := len(text)
	charCountNoSpaces := len(strings.ReplaceAll(text, " ", ""))

	return &agent.Result{
		Success: true,
		Data: map[string]interface{}{
			"character_count":           charCount,
			"character_count_no_spaces": charCountNoSpaces,
		},
		Message: fmt.Sprintf("Text contains %d characters (%d without spaces)", charCount, charCountNoSpaces),
	}, nil
}

// textAnalysis performs comprehensive text analysis
func (tp *TextProcessor) textAnalysis(text string) (*agent.Result, error) {
	if text == "" {
		return &agent.Result{
			Success: true,
			Data: map[string]interface{}{
				"word_count":      0,
				"character_count": 0,
				"sentence_count":  0,
				"line_count":      0,
			},
			Message: "Text is empty",
		}, nil
	}

	// Basic counts
	words := strings.Fields(strings.TrimSpace(text))
	wordCount := len(words)
	charCount := len(text)
	charCountNoSpaces := len(strings.ReplaceAll(text, " ", ""))

	// Sentence count (approximate)
	sentenceEnders := regexp.MustCompile(`[.!?]+`)
	sentences := sentenceEnders.Split(text, -1)
	sentenceCount := len(sentences) - 1 // Last element is usually empty
	if sentenceCount < 0 {
		sentenceCount = 0
	}

	// Line count
	lines := strings.Split(text, "\n")
	lineCount := len(lines)

	// Letter and digit counts
	letterCount := 0
	digitCount := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			letterCount++
		} else if unicode.IsDigit(r) {
			digitCount++
		}
	}

	return &agent.Result{
		Success: true,
		Data: map[string]interface{}{
			"word_count":                wordCount,
			"character_count":           charCount,
			"character_count_no_spaces": charCountNoSpaces,
			"sentence_count":            sentenceCount,
			"line_count":                lineCount,
			"letter_count":              letterCount,
			"digit_count":               digitCount,
			"words":                     words,
		},
		Message: fmt.Sprintf("Analysis complete: %d words, %d characters, %d sentences, %d lines",
			wordCount, charCount, sentenceCount, lineCount),
	}, nil
}

// textCleanup removes extra whitespace and normalizes text
func (tp *TextProcessor) textCleanup(text string) (*agent.Result, error) {
	// Remove extra whitespace while preserving structure
	cleaned := strings.TrimSpace(text)

	// Replace multiple spaces with single space (but preserve newlines)
	lines := strings.Split(cleaned, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
		spaceRegex := regexp.MustCompile(`\s+`)
		lines[i] = spaceRegex.ReplaceAllString(lines[i], " ")
	}
	cleaned = strings.Join(lines, "\n")

	// Remove extra newlines (keep at most 2 consecutive newlines)
	newlineRegex := regexp.MustCompile(`\n{3,}`)
	cleaned = newlineRegex.ReplaceAllString(cleaned, "\n\n")

	return &agent.Result{
		Success: true,
		Data: map[string]interface{}{
			"original_text":   text,
			"cleaned_text":    cleaned,
			"original_length": len(text),
			"cleaned_length":  len(cleaned),
		},
		Message: fmt.Sprintf("Text cleaned: reduced from %d to %d characters", len(text), len(cleaned)),
	}, nil
}

// textFormatting applies formatting to text based on context
func (tp *TextProcessor) textFormatting(text string, context map[string]interface{}) (*agent.Result, error) {
	formatted := text

	// Get formatting options from context
	if context != nil {
		if format, ok := context["format"].(string); ok {
			switch format {
			case "uppercase":
				formatted = strings.ToUpper(text)
			case "lowercase":
				formatted = strings.ToLower(text)
			case "title":
				formatted = strings.Title(text)
			case "sentence":
				formatted = tp.toSentenceCase(text)
			}
		}
	}

	return &agent.Result{
		Success: true,
		Data: map[string]interface{}{
			"original_text":  text,
			"formatted_text": formatted,
		},
		Message: "Text formatting applied",
	}, nil
}

// toSentenceCase converts text to sentence case (first letter uppercase, rest lowercase)
func (tp *TextProcessor) toSentenceCase(text string) string {
	if len(text) == 0 {
		return text
	}

	runes := []rune(text)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}
