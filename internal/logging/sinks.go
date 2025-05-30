package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// ConsoleSink writes logs to the console
type ConsoleSink struct {
	writer     io.Writer
	structured bool
	colorized  bool
}

// NewConsoleSink creates a new console sink
func NewConsoleSink(structured bool) *ConsoleSink {
	return &ConsoleSink{
		writer:     os.Stdout,
		structured: structured,
		colorized:  true,
	}
}

// NewConsoleSinkWithWriter creates a console sink with a custom writer
func NewConsoleSinkWithWriter(writer io.Writer, structured bool) *ConsoleSink {
	return &ConsoleSink{
		writer:     writer,
		structured: structured,
		colorized:  false, // Disable colors for custom writers
	}
}

// Write writes a log entry to the console
func (c *ConsoleSink) Write(entry LogEntry) error {
	if c.structured {
		return c.writeStructured(entry)
	}
	return c.writeFormatted(entry)
}

// Close closes the console sink (no-op for console)
func (c *ConsoleSink) Close() error {
	return nil
}

func (c *ConsoleSink) writeStructured(entry LogEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(c.writer, string(data))
	return err
}

func (c *ConsoleSink) writeFormatted(entry LogEntry) error {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")

	// Color codes for different log levels
	var levelColor, resetColor string
	if c.colorized {
		resetColor = "\033[0m"
		switch entry.Level {
		case "TRACE":
			levelColor = "\033[90m" // Gray
		case "DEBUG":
			levelColor = "\033[36m" // Cyan
		case "INFO":
			levelColor = "\033[32m" // Green
		case "WARN":
			levelColor = "\033[33m" // Yellow
		case "ERROR":
			levelColor = "\033[31m" // Red
		case "FATAL":
			levelColor = "\033[35m" // Magenta
		default:
			levelColor = ""
		}
	}

	// Format: [TIMESTAMP] [LEVEL] [COMPONENT] MESSAGE
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	parts = append(parts, fmt.Sprintf("[%s%5s%s]", levelColor, entry.Level, resetColor))

	if entry.Component != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Component))
	}

	if entry.RequestID != "" {
		parts = append(parts, fmt.Sprintf("[req:%s]", entry.RequestID))
	}

	if entry.Operation != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Operation))
	}

	parts = append(parts, entry.Message)

	line := strings.Join(parts, " ")

	// Add error details if present
	if entry.Error != "" {
		line += fmt.Sprintf(" | Error: %s", entry.Error)
	}

	// Add duration if present
	if entry.Duration != nil {
		line += fmt.Sprintf(" | Duration: %v", *entry.Duration)
	}

	// Add important properties
	if len(entry.Properties) > 0 {
		var props []string
		for k, v := range entry.Properties {
			// Skip already displayed properties
			if k == "request_id" || k == "operation" {
				continue
			}
			props = append(props, fmt.Sprintf("%s=%v", k, v))
		}
		if len(props) > 0 {
			line += fmt.Sprintf(" | %s", strings.Join(props, ", "))
		}
	}

	_, err := fmt.Fprintln(c.writer, line)
	return err
}

// FileSink writes logs to a file
type FileSink struct {
	file       *os.File
	structured bool
}

// NewFileSink creates a new file sink
func NewFileSink(filename string, structured bool) (*FileSink, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileSink{
		file:       file,
		structured: structured,
	}, nil
}

// Write writes a log entry to the file
func (f *FileSink) Write(entry LogEntry) error {
	if f.structured {
		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(f.file, string(data))
		return err
	}

	// Use formatted output for non-structured file logging
	consoleSink := NewConsoleSinkWithWriter(f.file, false)
	return consoleSink.writeFormatted(entry)
}

// Close closes the file sink
func (f *FileSink) Close() error {
	return f.file.Close()
}

// BufferedSink buffers log entries and flushes them periodically or when buffer is full
type BufferedSink struct {
	sink       LogSink
	buffer     []LogEntry
	bufferSize int
	flushTimer *time.Timer
}

// NewBufferedSink creates a new buffered sink
func NewBufferedSink(sink LogSink, bufferSize int, flushInterval time.Duration) *BufferedSink {
	bs := &BufferedSink{
		sink:       sink,
		buffer:     make([]LogEntry, 0, bufferSize),
		bufferSize: bufferSize,
	}

	// Start flush timer
	bs.flushTimer = time.AfterFunc(flushInterval, func() {
		bs.flush()
	})

	return bs
}

// Write adds a log entry to the buffer
func (b *BufferedSink) Write(entry LogEntry) error {
	b.buffer = append(b.buffer, entry)

	// Flush if buffer is full
	if len(b.buffer) >= b.bufferSize {
		return b.flush()
	}

	return nil
}

// flush writes all buffered entries to the underlying sink
func (b *BufferedSink) flush() error {
	if len(b.buffer) == 0 {
		return nil
	}

	var lastErr error
	for _, entry := range b.buffer {
		if err := b.sink.Write(entry); err != nil {
			lastErr = err
		}
	}

	// Clear buffer
	b.buffer = b.buffer[:0]

	// Reset timer
	b.flushTimer.Reset(time.Minute)

	return lastErr
}

// Close flushes remaining entries and closes the underlying sink
func (b *BufferedSink) Close() error {
	if b.flushTimer != nil {
		b.flushTimer.Stop()
	}

	if err := b.flush(); err != nil {
		return err
	}

	return b.sink.Close()
}
