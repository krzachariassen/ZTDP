package logging

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the log level
type LogLevel int

const (
	LevelTrace LogLevel = iota - 1
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Source     string                 `json:"source"`
	Component  string                 `json:"component,omitempty"`
	Operation  string                 `json:"operation,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Duration   *time.Duration         `json:"duration,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
}

// LogSink defines the interface for log outputs
type LogSink interface {
	Write(entry LogEntry) error
	Close() error
}

// Logger is the main structured logger
type Logger struct {
	component string
	sinks     []LogSink
	level     LogLevel
	context   map[string]interface{}
}

// Global logger instance
var globalLogger *Logger

// InitializeLogger initializes the global logger with default configuration
func InitializeLogger(component string, level LogLevel) {
	logger := &Logger{
		component: component,
		level:     level,
		sinks:     make([]LogSink, 0),
		context:   make(map[string]interface{}),
	}

	// Add default console sink
	consoleSink := NewConsoleSink(true) // structured output
	logger.AddSink(consoleSink)

	globalLogger = logger
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		InitializeLogger("ztdp", LevelInfo)
	}
	return globalLogger
}

// ForComponent creates a new logger context for a specific component
func (l *Logger) ForComponent(component string) *Logger {
	return &Logger{
		component: component,
		sinks:     l.sinks,
		level:     l.level,
		context:   copyMap(l.context),
	}
}

// WithContext adds context properties to the logger
func (l *Logger) WithContext(key string, value interface{}) *Logger {
	newContext := copyMap(l.context)
	newContext[key] = value
	return &Logger{
		component: l.component,
		sinks:     l.sinks,
		level:     l.level,
		context:   newContext,
	}
}

// WithRequestID adds a request ID to the logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithContext("request_id", requestID)
}

// WithUserID adds a user ID to the logger context
func (l *Logger) WithUserID(userID string) *Logger {
	return l.WithContext("user_id", userID)
}

// WithOperation adds an operation name to the logger context
func (l *Logger) WithOperation(operation string) *Logger {
	return l.WithContext("operation", operation)
}

// AddSink adds a log sink to the logger
func (l *Logger) AddSink(sink LogSink) {
	l.sinks = append(l.sinks, sink)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// log is the internal method that handles actual logging
func (l *Logger) log(level LogLevel, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	source := "unknown"
	if ok {
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			source = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}

	// Format message
	message := msg
	if len(args) > 0 {
		message = fmt.Sprintf(msg, args...)
	}

	// Create log entry
	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      level.String(),
		Message:    message,
		Source:     source,
		Component:  l.component,
		Properties: copyMap(l.context),
	}

	// Extract specific context values
	if requestID, ok := l.context["request_id"].(string); ok {
		entry.RequestID = requestID
	}
	if userID, ok := l.context["user_id"].(string); ok {
		entry.UserID = userID
	}
	if operation, ok := l.context["operation"].(string); ok {
		entry.Operation = operation
	}

	// Write to all sinks
	for _, sink := range l.sinks {
		if err := sink.Write(entry); err != nil {
			// Fallback to standard log if sink fails
			slog.Error("Failed to write to log sink", "error", err, "entry", entry)
		}
	}
}

// Trace logs a trace message
func (l *Logger) Trace(msg string, args ...interface{}) {
	l.log(LevelTrace, msg, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(LevelDebug, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(LevelInfo, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(LevelWarn, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(LevelError, msg, args...)
}

// ErrorWithErr logs an error message with an error object
func (l *Logger) ErrorWithErr(err error, msg string, args ...interface{}) {
	entry := l.createEntry(LevelError, msg, args...)
	if err != nil {
		entry.Error = err.Error()
		// Add stack trace for errors
		entry.StackTrace = getStackTrace()
	}
	l.writeEntry(entry)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(LevelFatal, msg, args...)
	os.Exit(1)
}

// LogAPIRequest logs an API request with timing information
func (l *Logger) LogAPIRequest(method, path string, statusCode int, duration time.Duration, requestID string) {
	level := LevelInfo
	if statusCode >= 400 && statusCode < 500 {
		level = LevelWarn
	} else if statusCode >= 500 {
		level = LevelError
	}

	l.WithRequestID(requestID).
		WithContext("method", method).
		WithContext("path", path).
		WithContext("status_code", statusCode).
		WithContext("duration_ms", duration.Milliseconds()).
		log(level, "API %s %s - %d (%v)", method, path, statusCode, duration)
}

// LogGraphOperation logs a graph operation
func (l *Logger) LogGraphOperation(operation, nodeID, nodeKind string, success bool, duration time.Duration) {
	level := LevelInfo
	if !success {
		level = LevelError
	}

	l.WithOperation(operation).
		WithContext("node_id", nodeID).
		WithContext("node_kind", nodeKind).
		WithContext("success", success).
		WithContext("duration_ms", duration.Milliseconds()).
		log(level, "Graph %s: %s (%s) - success: %t", operation, nodeID, nodeKind, success)
}

// LogPolicyEvaluation logs a policy evaluation
func (l *Logger) LogPolicyEvaluation(policyID, subject string, result bool, reason string, duration time.Duration) {
	level := LevelInfo
	if !result {
		level = LevelWarn
	}

	l.WithOperation("policy_evaluation").
		WithContext("policy_id", policyID).
		WithContext("subject", subject).
		WithContext("result", result).
		WithContext("reason", reason).
		WithContext("duration_ms", duration.Milliseconds()).
		log(level, "Policy %s for %s: %t (%s)", policyID, subject, result, reason)
}

// LogDeployment logs a deployment event
func (l *Logger) LogDeployment(deploymentID, service, environment, status string, progress *float64) {
	level := LevelInfo
	if status == "failed" {
		level = LevelError
	} else if status == "completed" {
		level = LevelInfo
	}

	logger := l.WithOperation("deployment").
		WithContext("deployment_id", deploymentID).
		WithContext("service", service).
		WithContext("environment", environment).
		WithContext("status", status)

	if progress != nil {
		logger = logger.WithContext("progress", *progress)
		logger.log(level, "Deployment %s: %s to %s - %s (%.1f%%)", deploymentID, service, environment, status, *progress*100)
	} else {
		logger.log(level, "Deployment %s: %s to %s - %s", deploymentID, service, environment, status)
	}
}

// Helper methods

func (l *Logger) createEntry(level LogLevel, msg string, args ...interface{}) LogEntry {
	// Get caller information
	_, file, line, ok := runtime.Caller(3)
	source := "unknown"
	if ok {
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			source = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}

	// Format message
	message := msg
	if len(args) > 0 {
		message = fmt.Sprintf(msg, args...)
	}

	// Create log entry
	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      level.String(),
		Message:    message,
		Source:     source,
		Component:  l.component,
		Properties: copyMap(l.context),
	}

	// Extract specific context values
	if requestID, ok := l.context["request_id"].(string); ok {
		entry.RequestID = requestID
	}
	if userID, ok := l.context["user_id"].(string); ok {
		entry.UserID = userID
	}
	if operation, ok := l.context["operation"].(string); ok {
		entry.Operation = operation
	}

	return entry
}

func (l *Logger) writeEntry(entry LogEntry) {
	// Write to all sinks
	for _, sink := range l.sinks {
		if err := sink.Write(entry); err != nil {
			// Fallback to standard log if sink fails
			slog.Error("Failed to write to log sink", "error", err, "entry", entry)
		}
	}
}

// Utility functions

func copyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func getStackTrace() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// Convenience functions for global logger

// Info logs an info message using the global logger
func Info(msg string, args ...interface{}) {
	GetLogger().Info(msg, args...)
}

// Debug logs a debug message using the global logger
func Debug(msg string, args ...interface{}) {
	GetLogger().Debug(msg, args...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, args ...interface{}) {
	GetLogger().Warn(msg, args...)
}

// Error logs an error message using the global logger
func Error(msg string, args ...interface{}) {
	GetLogger().Error(msg, args...)
}

// ErrorWithErr logs an error with error object using the global logger
func ErrorWithErr(err error, msg string, args ...interface{}) {
	GetLogger().ErrorWithErr(err, msg, args...)
}

// ForComponent creates a component-specific logger
func ForComponent(component string) *Logger {
	return GetLogger().ForComponent(component)
}
