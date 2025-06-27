package logging

import (
	"fmt"
	"log"
	"time"
)

// StructuredLogger implements Logger with structured output
type StructuredLogger struct {
	level LogLevel
}

// LogLevel represents logging levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(level LogLevel) Logger {
	return &StructuredLogger{
		level: level,
	}
}

// Info logs an informational message with optional fields
func (s *StructuredLogger) Info(msg string, fields ...interface{}) {
	if s.level <= LevelInfo {
		s.logWithFields("INFO", msg, fields...)
	}
}

// Error logs an error message with optional fields
func (s *StructuredLogger) Error(msg string, err error, fields ...interface{}) {
	if s.level <= LevelError {
		allFields := append([]interface{}{"error", err.Error()}, fields...)
		s.logWithFields("ERROR", msg, allFields...)
	}
}

// Debug logs a debug message with optional fields
func (s *StructuredLogger) Debug(msg string, fields ...interface{}) {
	if s.level <= LevelDebug {
		s.logWithFields("DEBUG", msg, fields...)
	}
}

// Warn logs a warning message with optional fields
func (s *StructuredLogger) Warn(msg string, fields ...interface{}) {
	if s.level <= LevelWarn {
		s.logWithFields("WARN", msg, fields...)
	}
}

func (s *StructuredLogger) logWithFields(level, msg string, fields ...interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	logMsg := fmt.Sprintf("[%s] %s %s", timestamp, level, msg)

	// Add fields as key-value pairs
	if len(fields) > 0 {
		logMsg += " "
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				logMsg += fmt.Sprintf("%v=%v ", fields[i], fields[i+1])
			}
		}
	}

	log.Println(logMsg)
}
