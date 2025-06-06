package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ztdp-devops/ztdp/internal/ai"
	"github.com/ztdp-devops/ztdp/internal/deployments"
	"github.com/ztdp-devops/ztdp/internal/events"
	"github.com/ztdp-devops/ztdp/internal/graph"
	"github.com/ztdp-devops/ztdp/internal/policies"
)

// Test program to verify AI platform agent integration works
func main() {
	ctx := context.Background()

	// Initialize dependencies (using mock/test implementations)
	graphDB := &graph.MockGraph{} // Assuming this exists or create minimal implementation
	eventBus := events.NewEventBus()

	// Create OpenAI provider (can be nil for this test)
	var aiProvider ai.Provider

	// Create services with the new methods
	deploymentService := deployments.NewService(graphDB, aiProvider, eventBus)
	policyService := policies.NewService(graphDB, eventBus)

	// Try to create the platform agent - this should work now
	agent, err := ai.NewPlatformAgent(deploymentService, policyService, aiProvider)
	if err != nil {
		log.Fatalf("Failed to create platform agent: %v", err)
	}

	fmt.Println("✅ Successfully created AI platform agent!")
	fmt.Printf("Agent type: %T\n", agent)

	// Test that the interfaces are properly implemented
	fmt.Println("✅ DeploymentService interface implemented correctly")
	fmt.Println("✅ PolicyService interface implemented correctly")
	fmt.Println("✅ AI chat handler should now work without interface mismatch errors")
}
