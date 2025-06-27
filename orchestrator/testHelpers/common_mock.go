package testHelpers

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockLogger provides a testify-based mock for logger operations
type MockLogger struct {
	mock.Mock
}

// NewMockLogger creates a new mock logger instance
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	args := []interface{}{msg}
	args = append(args, fields...)
	m.Called(args...)
}

func (m *MockLogger) Error(msg string, err error, fields ...interface{}) {
	args := []interface{}{msg, err}
	args = append(args, fields...)
	m.Called(args...)
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	args := []interface{}{msg}
	args = append(args, fields...)
	m.Called(args...)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	args := []interface{}{msg}
	args = append(args, fields...)
	m.Called(args...)
}

// MockAIOrchestrator provides a testify-based mock for AI orchestrator operations
type MockAIOrchestrator struct {
	mock.Mock
}

// NewMockAIOrchestrator creates a new mock AI orchestrator instance
func NewMockAIOrchestrator() *MockAIOrchestrator {
	return &MockAIOrchestrator{}
}

func (m *MockAIOrchestrator) ProcessRequest(ctx context.Context, userID, request string) (string, error) {
	args := m.Called(ctx, userID, request)
	return args.String(0), args.Error(1)
}

func (m *MockAIOrchestrator) GetAgentStatus(ctx context.Context, agentID string) (string, error) {
	args := m.Called(ctx, agentID)
	return args.String(0), args.Error(1)
}
