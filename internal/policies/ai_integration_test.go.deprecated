package policies

import (
	"context"
	"testing"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

func TestPolicyService_AIIntegration(t *testing.T) {
	t.Run("uses real AI for policy evaluation", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup policy
		policy := createApplicationServiceLimitPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Create test node
		node := createTestApplicationNode()

		// Evaluate policy using real AI
		result, err := service.EvaluateNodePolicy(context.Background(), "test-env", node, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		// Verify AI evaluation completed
		if result == nil {
			t.Error("Expected policy result, got nil")
		}

		// Verify AI reasoning is present (real AI should provide reasoning)
		if result.Reason == "" {
			t.Error("Expected AI reasoning to be populated")
		}

		// Log the actual AI response for debugging
		t.Logf("AI Policy Evaluation:")
		t.Logf("  Status: %s", result.Status)
		t.Logf("  Reason: %s", result.Reason)
		t.Logf("  Confidence: %f", result.Confidence)
	})

	t.Run("handles edge policy evaluation with AI", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup edge policy
		policy := createNoDirectProdDeploymentPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Create production deployment edge
		edge := &graph.Edge{
			To:   "production",
			Type: graph.EdgeTypeDeploy,
			Metadata: map[string]interface{}{
				"deployment_type":    "direct",
				"source_environment": "development",
				"target_environment": "production",
				"bypass_staging":     true,
				"staging_bypassed":   true,
				"policy_context":     "Direct deployment to production - violates staging-first policy",
				"description":        "Direct production deployment that should be blocked",
			},
		}

		// Evaluate edge policy using real AI
		result, err := service.EvaluateEdgePolicy(context.Background(), "production", edge, policy)
		if err != nil {
			t.Fatalf("Edge policy evaluation failed: %v", err)
		}

		// Verify AI evaluation completed
		if result == nil {
			t.Error("Expected policy result, got nil")
		}

		// Log the actual AI response for debugging
		t.Logf("AI Edge Policy Evaluation:")
		t.Logf("  Status: %s", result.Status)
		t.Logf("  Reason: %s", result.Reason)
		t.Logf("  Confidence: %f", result.Confidence)

		// AI should provide meaningful reasoning for deployment policies
		if result.Reason == "" {
			t.Error("Expected AI reasoning for edge policy evaluation")
		}
	})

	t.Run("handles AI service availability", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Test that service handles AI availability gracefully
		policy := createApplicationServiceLimitPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		node := createTestApplicationNode()

		// This should work whether AI is available or not
		result, err := service.EvaluateNodePolicy(context.Background(), "test-env", node, policy)
		if err != nil {
			t.Fatalf("Policy evaluation should handle AI availability gracefully: %v", err)
		}

		if result == nil {
			t.Error("Expected some form of policy result")
		}

		t.Logf("Policy evaluation result (AI availability test): %s", result.Status)
	})

	t.Run("AI provides different reasoning for different scenarios", func(t *testing.T) {
		service, eventBus, store := createTestPolicyService(t)
		defer eventBus.ClearEvents()

		// Setup policy
		policy := createDatabaseBackupPolicy()
		err := store.Store(policy)
		if err != nil {
			t.Fatalf("Failed to store policy: %v", err)
		}

		// Test with database node that has backup configuration
		nodeWithBackup := &graph.Node{
			ID:   "test-db-with-backup",
			Kind: graph.KindResource,
			Metadata: map[string]interface{}{
				"name":          "Test Database with Backup",
				"resource_type": "database",
			},
			Spec: map[string]interface{}{
				"backup_enabled":  true,
				"backup_schedule": "daily",
			},
		}

		result1, err := service.EvaluateNodePolicy(context.Background(), "test", nodeWithBackup, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		// Test with database node that has no backup configuration
		nodeWithoutBackup := &graph.Node{
			ID:   "test-db-without-backup",
			Kind: graph.KindResource,
			Metadata: map[string]interface{}{
				"name":          "Test Database without Backup",
				"resource_type": "database",
			},
			Spec: map[string]interface{}{
				"backup_enabled": false,
			},
		}

		result2, err := service.EvaluateNodePolicy(context.Background(), "test", nodeWithoutBackup, policy)
		if err != nil {
			t.Fatalf("Policy evaluation failed: %v", err)
		}

		// AI should provide different reasoning for different scenarios
		t.Logf("With backup - Status: %s, Reason: %s", result1.Status, result1.Reason)
		t.Logf("Without backup - Status: %s, Reason: %s", result2.Status, result2.Reason)

		// The reasoning should be different for these two scenarios
		if result1.Reason == result2.Reason && result1.Reason != "" {
			t.Error("Expected different AI reasoning for different database backup scenarios")
		}
	})
}
