package testHelpers

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/ztdp/orchestrator/internal/planning/domain"
)

// MockPlanningRepository provides a testify-based mock for planning repository operations
type MockPlanningRepository struct {
	mock.Mock
}

// NewMockPlanningRepository creates a new mock planning repository instance
func NewMockPlanningRepository() *MockPlanningRepository {
	return &MockPlanningRepository{}
}

func (m *MockPlanningRepository) Save(ctx context.Context, plan *domain.ExecutionPlan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}

func (m *MockPlanningRepository) GetByID(ctx context.Context, id string) (*domain.ExecutionPlan, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.ExecutionPlan), args.Error(1)
}

func (m *MockPlanningRepository) GetByStatus(ctx context.Context, status domain.ExecutionPlanStatus) ([]*domain.ExecutionPlan, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*domain.ExecutionPlan), args.Error(1)
}

func (m *MockPlanningRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
