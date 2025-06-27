package logging

import "go.uber.org/zap"

// Logger defines the interface for structured logging
type Logger interface {
	// Info logs an informational message with optional fields
	Info(msg string, fields ...interface{})

	// Error logs an error message with optional fields
	Error(msg string, err error, fields ...interface{})

	// Debug logs a debug message with optional fields
	Debug(msg string, fields ...interface{})

	// Warn logs a warning message with optional fields
	Warn(msg string, fields ...interface{})
}

// NoOpLogger implements Logger interface with no-op operations (for testing)
type NoOpLogger struct{}

func (n *NoOpLogger) Info(msg string, fields ...interface{})             {}
func (n *NoOpLogger) Error(msg string, err error, fields ...interface{}) {}
func (n *NoOpLogger) Debug(msg string, fields ...interface{})            {}
func (n *NoOpLogger) Warn(msg string, fields ...interface{})             {}

// NewNoOpLogger creates a new no-op logger (useful for testing)
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

// logger implements Logger interface with structured logging using Zap
type logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewLogger creates a new logger instance
func NewLogger(development bool) (Logger, error) {
	var zapLogger *zap.Logger
	var err error

	if development {
		// Development config: human-readable, colorized output
		zapLogger, err = zap.NewDevelopment()
	} else {
		// Production config: JSON output, optimized performance
		zapLogger, err = zap.NewProduction()
	}

	if err != nil {
		return nil, err
	}

	return &logger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
	}, nil
}

// NewLoggerWithConfig creates a logger with custom configuration
func NewLoggerWithConfig(config zap.Config) (Logger, error) {
	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &logger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
	}, nil
}

func (l *logger) Info(msg string, fields ...interface{}) {
	l.sugar.Infow(msg, fields...)
}

func (l *logger) Error(msg string, err error, fields ...interface{}) {
	// Add error to fields for structured logging
	allFields := append([]interface{}{"error", err}, fields...)
	l.sugar.Errorw(msg, allFields...)
}

func (l *logger) Debug(msg string, fields ...interface{}) {
	l.sugar.Debugw(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...interface{}) {
	l.sugar.Warnw(msg, fields...)
}

// Close flushes any buffered log entries
func (l *logger) Close() error {
	return l.logger.Sync()
}

// GetZapLogger returns the underlying zap logger for advanced use cases
func (l *logger) GetZapLogger() *zap.Logger {
	return l.logger
}
