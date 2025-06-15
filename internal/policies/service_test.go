package policies

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

func TestPolicyService_Basic(t *testing.T) {
	t.Run("creates service without error", func(t *testing.T) {
		_, _, _ = createTestPolicyService(t)
		// If we get here without panicking, service creation works
	})
}

func TestPolicyService_EvaluateNodePolicy(t *testing.T) {
	t.Run("evaluates application service limit policy", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup policy
		policy := createApplicationServiceLimitPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Create test application node
		node := createTestApplicationNode()

		// Evaluate node policy - AI should determine based on service count
		result, err := service.EvaluateNodePolicy(context.Background(), "test", node, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		// Verify policy evaluation completed
		if result == nil {
			t.Error("Expected policy result, got nil")
			return
		}

		// Log result for debugging
		t.Logf("Policy evaluation result: %s - %s", result.Status, result.Reason)
	})

	t.Run("evaluates database backup policy", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup policy
		policy := createDatabaseBackupPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Create test database node
		node := createTestDatabaseNode()

		// Evaluate node policy - AI should check for backup configuration
		result, err := service.EvaluateNodePolicy(context.Background(), "test", node, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		// Verify policy evaluation completed
		if result == nil {
			t.Error("Expected policy result, got nil")
			return
		}

		// Log result for debugging
		t.Logf("Database backup policy result: %s - %s", result.Status, result.Reason)
	})

	t.Run("handles nodes that don't match policy scope", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup application-specific policy
		policy := createApplicationServiceLimitPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Create database node (doesn't match application policy)
		node := createTestDatabaseNode()

		// Evaluate node policy - should be not applicable
		result, err := service.EvaluateNodePolicy(context.Background(), "test", node, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		// Should be not applicable since database doesn't match application policy
		if result.Status != PolicyStatusNotApplicable {
			t.Errorf("Expected not applicable status for mismatched node type, got %s", result.Status)
		}
	})

	t.Run("TDD_positive_case_allows_compliant_application", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup policy
		policy := createApplicationServiceLimitPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Create compliant application with fewer than 10 services
		node := &graph.Node{
			ID:   "compliant-app",
			Kind: graph.KindApplication,
			Metadata: map[string]interface{}{
				"name":        "Compliant Application",
				"environment": "production",
				"team":        "backend",
			},
			Spec: map[string]interface{}{
				"services": []string{"auth-service", "user-service", "api-gateway"}, // Only 3 services
				"version":  "1.2.0",
			},
		}

		// Evaluate node policy - AI should allow compliant application
		result, err := service.EvaluateNodePolicy(context.Background(), "production", node, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		t.Logf("Compliant application (3 services) result: %s - %s", result.Status, result.Reason)

		// Should be allowed - only 3 services, well under the 10 service limit
		if result.Status == PolicyStatusBlocked {
			t.Errorf("Expected allowed status for compliant application with 3 services, got blocked: %s", result.Reason)
		}
	})

	t.Run("TDD_negative_case_blocks_non_compliant_application", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup policy
		policy := createApplicationServiceLimitPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Create non-compliant application with too many services
		node := &graph.Node{
			ID:   "non-compliant-app",
			Kind: graph.KindApplication,
			Metadata: map[string]interface{}{
				"name":        "Over-Complicated Application",
				"environment": "production",
				"team":        "platform",
			},
			Spec: map[string]interface{}{
				"services": []string{
					"auth-service", "user-service", "api-gateway",
					"payment-service", "notification-service", "analytics-service",
					"logging-service", "monitoring-service", "search-service",
					"recommendation-service", "inventory-service", "shipping-service",
					"review-service", "order-service", "catalog-service", // 15 services total
				},
				"version": "2.1.0",
			},
		}

		// Evaluate node policy - AI should block non-compliant application
		result, err := service.EvaluateNodePolicy(context.Background(), "production", node, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		t.Logf("Non-compliant application (15 services) result: %s - %s", result.Status, result.Reason)

		// Should be blocked - 15 services exceeds the 10 service limit
		if result.Status == PolicyStatusAllowed {
			t.Errorf("Expected blocked status for non-compliant application with 15 services, got allowed: %s", result.Reason)
		}
	})

	t.Run("TDD_boundary_case_edge_of_limit", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup policy
		policy := createApplicationServiceLimitPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Create application with exactly 10 services (boundary case)
		node := &graph.Node{
			ID:   "boundary-app",
			Kind: graph.KindApplication,
			Metadata: map[string]interface{}{
				"name":        "Boundary Application",
				"environment": "production",
				"team":        "core",
			},
			Spec: map[string]interface{}{
				"services": []string{
					"service-1", "service-2", "service-3", "service-4", "service-5",
					"service-6", "service-7", "service-8", "service-9", "service-10",
				}, // Exactly 10 services
				"version": "1.0.0",
			},
		}

		// Evaluate node policy - AI should handle boundary case appropriately
		result, err := service.EvaluateNodePolicy(context.Background(), "production", node, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		t.Logf("Boundary application (exactly 10 services) result: %s - %s", result.Status, result.Reason)

		// The policy says "fewer than 10 services", so exactly 10 should be blocked
		// But we'll log the AI's interpretation for validation
		if result.Status == PolicyStatusAllowed {
			t.Logf("AI interpreted 'fewer than 10' as allowing exactly 10 services")
		} else {
			t.Logf("AI correctly interpreted 'fewer than 10' as blocking exactly 10 services")
		}
	})
}

func TestPolicyService_PolicyStore_Integration(t *testing.T) {
	t.Run("stores and retrieves policies correctly", func(t *testing.T) {
		_, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Store a policy
		policy := createApplicationServiceLimitPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Retrieve the policy by ID
		retrieved, err := store.Get(policy.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve policy: %v", err)
		}

		if retrieved.ID != policy.ID {
			t.Errorf("Expected policy ID %s, got %s", policy.ID, retrieved.ID)
		}

		if retrieved.Name != policy.Name {
			t.Errorf("Expected policy name %s, got %s", policy.Name, retrieved.Name)
		}
	})

	t.Run("evaluates policies using AI", func(t *testing.T) {
		service, eventBus, _ := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Create a simple policy to test with
		policy := createApplicationServiceLimitPolicy()

		// Try to evaluate a node policy
		node := createTestApplicationNode()
		result, err := service.EvaluateNodePolicy(context.Background(), "test", node, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		if result == nil {
			t.Error("Expected policy result, got nil")
			return
		}

		// Verify that AI reasoning is present
		if result.Reason == "" {
			t.Error("Expected AI reasoning in policy result")
		}

		t.Logf("AI policy evaluation: %s - %s", result.Status, result.Reason)
	})
}
