package testHelpers

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/ztdp/orchestrator/internal/orchestrator/domain"
)

// MockGraphExplorer provides a testify-based mock for graph explorer operations
type MockGraphExplorer struct {
	mock.Mock
}

// NewMockGraphExplorer creates a new mock graph explorer instance
func NewMockGraphExplorer() *MockGraphExplorer {
	return &MockGraphExplorer{}
}

func (m *MockGraphExplorer) ExploreRequest(ctx context.Context, request string) (*domain.Analysis, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*domain.Analysis), args.Error(1)
}

func (m *MockGraphExplorer) GetContextualAgents(ctx context.Context, request string) ([]*domain.Analysis, error) {
	args := m.Called(ctx, request)
	return args.Get(0).([]*domain.Analysis), args.Error(1)
}

// MockExecutionCoordinator provides a testify-based mock for execution coordinator operations
type MockExecutionCoordinator struct {
	mock.Mock
}

// NewMockExecutionCoordinator creates a new mock execution coordinator instance
func NewMockExecutionCoordinator() *MockExecutionCoordinator {
	return &MockExecutionCoordinator{}
}

func (m *MockExecutionCoordinator) CreatePlan(ctx context.Context, decision *domain.Decision) (string, error) {
	args := m.Called(ctx, decision)
	return args.String(0), args.Error(1)
}

func (m *MockExecutionCoordinator) GetPlanStatus(ctx context.Context, planID string) (*domain.ExecutionPlan, error) {
	args := m.Called(ctx, planID)
	return args.Get(0).(*domain.ExecutionPlan), args.Error(1)
}

func (m *MockExecutionCoordinator) UpdateStatus(ctx context.Context, planID string, status domain.ExecutionStatus) error {
	args := m.Called(ctx, planID, status)
	return args.Error(0)
}

func (m *MockExecutionCoordinator) ExecutePlan(ctx context.Context, planID string) error {
	args := m.Called(ctx, planID)
	return args.Error(0)
}

// MockLearningService provides a testify-based mock for learning service operations
type MockLearningService struct {
	mock.Mock
}

// NewMockLearningService creates a new mock learning service instance
func NewMockLearningService() *MockLearningService {
	return &MockLearningService{}
}

func (m *MockLearningService) AnalyzePatterns(ctx context.Context, sessionID string) (*domain.ConversationPattern, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*domain.ConversationPattern), args.Error(1)
}

func (m *MockLearningService) UpdatePattern(ctx context.Context, pattern *domain.ConversationPattern) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

// MockExecutionService provides a testify-based mock for execution service operations
type MockExecutionService struct {
	mock.Mock
}

// NewMockExecutionService creates a new mock execution service instance
func NewMockExecutionService() *MockExecutionService {
	return &MockExecutionService{}
}

func (m *MockExecutionService) CreateExecutionPlan(ctx context.Context, plan *domain.ExecutionPlan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}

func (m *MockExecutionService) GetExecutionPlan(ctx context.Context, planID string) (*domain.ExecutionPlan, error) {
	args := m.Called(ctx, planID)
	return args.Get(0).(*domain.ExecutionPlan), args.Error(1)
}

func (m *MockExecutionService) UpdateExecutionStatus(ctx context.Context, planID string, status domain.ExecutionStatus) error {
	args := m.Called(ctx, planID, status)
	return args.Error(0)
}
