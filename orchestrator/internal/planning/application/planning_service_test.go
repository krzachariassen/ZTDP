package application

import (
	"context"
	"errors"
	"testing"

	"github.com/ztdp/orchestrator/internal/planning/domain"
)

// MockPlanningRepository is a mock implementation of PlanningRepository
type MockPlanningRepository struct {
	plans          map[string]*domain.ExecutionPlan
	saveErr        error
	getByIDErr     error
	getByUserErr   error
	getByStatusErr error
	deleteErr      error
}

func NewMockPlanningRepository() *MockPlanningRepository {
	return &MockPlanningRepository{
		plans: make(map[string]*domain.ExecutionPlan),
	}
}

func (m *MockPlanningRepository) Save(ctx context.Context, plan *domain.ExecutionPlan) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.plans[plan.ID] = plan
	return nil
}

func (m *MockPlanningRepository) GetByID(ctx context.Context, planID string) (*domain.ExecutionPlan, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if plan, exists := m.plans[planID]; exists {
		return plan, nil
	}
	return nil, errors.New("execution plan not found")
}

func (m *MockPlanningRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.ExecutionPlan, error) {
	if m.getByUserErr != nil {
		return nil, m.getByUserErr
	}

	var userPlans []*domain.ExecutionPlan
	for _, plan := range m.plans {
		if plan.UserID == userID {
			userPlans = append(userPlans, plan)
		}
	}
	return userPlans, nil
}

func (m *MockPlanningRepository) GetByStatus(ctx context.Context, status domain.ExecutionPlanStatus) ([]*domain.ExecutionPlan, error) {
	if m.getByStatusErr != nil {
		return nil, m.getByStatusErr
	}

	var statusPlans []*domain.ExecutionPlan
	for _, plan := range m.plans {
		if plan.Status == status {
			statusPlans = append(statusPlans, plan)
		}
	}
	return statusPlans, nil
}

func (m *MockPlanningRepository) Delete(ctx context.Context, planID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.plans, planID)
	return nil
}

// Helper methods for testing
func (m *MockPlanningRepository) SetSaveError(err error) {
	m.saveErr = err
}

func (m *MockPlanningRepository) SetGetByIDError(err error) {
	m.getByIDErr = err
}

func (m *MockPlanningRepository) SetGetByUserError(err error) {
	m.getByUserErr = err
}

func (m *MockPlanningRepository) SetGetByStatusError(err error) {
	m.getByStatusErr = err
}

func (m *MockPlanningRepository) SetDeleteError(err error) {
	m.deleteErr = err
}

// TestPlanningServiceImpl_CreateExecutionPlan tests the CreateExecutionPlan method
func TestPlanningServiceImpl_CreateExecutionPlan(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPlanningRepository()
	service := NewPlanningServiceImpl(repo)

	t.Run("should create execution plan successfully", func(t *testing.T) {
		request := &PlanningRequest{
			ConversationID: "conv123",
			UserID:         "user123",
			UserRequest:    "Deploy my application",
			Intent:         "deployment",
			Category:       "infrastructure",
			RequiredAgents: []string{"deployment-agent"},
			Parameters:     map[string]interface{}{"environment": "production"},
		}

		plan, err := service.CreateExecutionPlan(ctx, request)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if plan == nil {
			t.Fatal("expected execution plan, got nil")
		}

		if plan.UserID != request.UserID {
			t.Errorf("expected user ID %s, got %s", request.UserID, plan.UserID)
		}

		if plan.Status != domain.ExecutionPlanStatusPending {
			t.Errorf("expected status %s, got %s", domain.ExecutionPlanStatusPending, plan.Status)
		}
	})

	t.Run("should fail when repository save fails", func(t *testing.T) {
		repo.SetSaveError(errors.New("save failed"))
		request := &PlanningRequest{
			ConversationID: "conv123",
			UserID:         "user123",
			UserRequest:    "Deploy my application",
			Intent:         "deployment",
			Category:       "infrastructure",
		}

		plan, err := service.CreateExecutionPlan(ctx, request)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if plan != nil {
			t.Fatal("expected nil plan on error")
		}
	})

	t.Run("should fail with invalid request", func(t *testing.T) {
		repo.SetSaveError(nil) // Reset error
		request := &PlanningRequest{
			UserID:      "", // Invalid - empty user ID
			UserRequest: "Deploy my application",
		}

		plan, err := service.CreateExecutionPlan(ctx, request)

		if err == nil {
			t.Fatal("expected error for invalid request")
		}

		if plan != nil {
			t.Fatal("expected nil plan on validation error")
		}
	})
}

// TestPlanningServiceImpl_GetExecutionPlan tests the GetExecutionPlan method
func TestPlanningServiceImpl_GetExecutionPlan(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPlanningRepository()
	service := NewPlanningServiceImpl(repo)

	// Setup test data
	testPlan, _ := domain.NewExecutionPlan("plan123", "conv123", "user123", "Deploy app", "deployment", "infrastructure")
	repo.Save(ctx, testPlan)

	t.Run("should retrieve execution plan successfully", func(t *testing.T) {
		plan, err := service.GetExecutionPlan(ctx, "plan123")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if plan == nil {
			t.Fatal("expected execution plan, got nil")
		}

		if plan.ID != "plan123" {
			t.Errorf("expected ID plan123, got %s", plan.ID)
		}
	})

	t.Run("should fail when plan not found", func(t *testing.T) {
		plan, err := service.GetExecutionPlan(ctx, "nonexistent")

		if err == nil {
			t.Fatal("expected error for nonexistent plan")
		}

		if plan != nil {
			t.Fatal("expected nil plan when not found")
		}
	})

	t.Run("should fail with empty plan ID", func(t *testing.T) {
		plan, err := service.GetExecutionPlan(ctx, "")

		if err == nil {
			t.Fatal("expected error for empty plan ID")
		}

		if plan != nil {
			t.Fatal("expected nil plan on validation error")
		}
	})
}

// TestPlanningServiceImpl_ValidatePlan tests the ValidatePlan method
func TestPlanningServiceImpl_ValidatePlan(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPlanningRepository()
	service := NewPlanningServiceImpl(repo)

	t.Run("should validate valid plan successfully", func(t *testing.T) {
		plan, _ := domain.NewExecutionPlan("plan123", "conv123", "user123", "Deploy app", "deployment", "infrastructure")

		err := service.ValidatePlan(ctx, plan)

		if err != nil {
			t.Fatalf("expected no error for valid plan, got %v", err)
		}
	})

	t.Run("should fail validation for invalid plan", func(t *testing.T) {
		// Create an invalid plan (nil)
		err := service.ValidatePlan(ctx, nil)

		if err == nil {
			t.Fatal("expected error for nil plan")
		}
	})
}

// This will fail until we implement PlanningServiceImpl
func TestPlanningServiceExists(t *testing.T) {
	// This test will fail because we haven't implemented PlanningServiceImpl yet
	// This is our RED phase - we write the test first, then implement the functionality
	t.Run("PlanningServiceImpl should exist", func(t *testing.T) {
		repo := NewMockPlanningRepository()
		service := NewPlanningServiceImpl(repo)

		if service == nil {
			t.Fatal("NewPlanningServiceImpl should return a service instance")
		}
	})
}
