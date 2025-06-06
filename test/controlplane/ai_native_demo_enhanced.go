// AI-Native Platform Demo
// This demo showcases the AI-native capabilities of the ZTDP platform
// where AI is the primary interface for developer interactions.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL    = "http://localhost:8080"
	redisDelay = 200 * time.Millisecond
)

func main() {
	fmt.Println("🤖 AI-Native Platform Demo")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("This demo showcases AI-native capabilities where AI is the primary interface")
	fmt.Println()

	// Check if we have an AI provider configured
	if !checkAIAvailable() {
		fmt.Println("⚠️  AI provider not available - some features will use fallbacks")
		fmt.Println("   To enable full AI features, set OPENAI_API_KEY environment variable")
		fmt.Println()
	}

	// Ensure we have a test environment
	setupTestEnvironment()

	// Demo 1: AI-Powered Deployment Planning
	fmt.Println("🧠 Demo 1: AI-Powered Deployment Planning")
	testAIDeploymentPlanning()

	// Demo 2: Intelligent Plan Optimization
	fmt.Println("\n🔧 Demo 2: AI Plan Optimization")
	testAIPlanOptimization()

	// Demo 3: Impact Analysis with AI
	fmt.Println("\n📊 Demo 3: AI Impact Analysis")
	testAIImpactAnalysis()

	// Demo 4: AI Policy Evaluation
	fmt.Println("\n🛡️  Demo 4: AI Policy Evaluation")
	testAIPolicyEvaluation()

	// Demo 5: Natural Language Operations
	fmt.Println("\n💬 Demo 5: Natural Language Operations")
	testNaturalLanguageOperations()

	// Demo 6: Error Scenarios with AI Guidance
	fmt.Println("\n🚨 Demo 6: AI Error Guidance")
	testAIErrorGuidance()

	// Demo 7: Conversational Deployment Flow
	fmt.Println("\n🗣️  Demo 7: Conversational Deployment")
	testConversationalDeployment()

	fmt.Println("\n✅ AI-Native Platform Demo Complete!")
	fmt.Println("The platform successfully demonstrated AI as the primary interface")
}

func checkAIAvailable() bool {
	resp, err := http.Get(baseURL + "/v1/ai/providers")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var providers map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&providers); err != nil {
		return false
	}

	if available, ok := providers["available"].(bool); ok && available {
		if current, ok := providers["current"].(string); ok && current != "" {
			fmt.Printf("✅ AI Provider: %s\n", current)
			return true
		}
	}
	return false
}

func setupTestEnvironment() {
	fmt.Println("🔄 Setting up test environment...")

	// Create test application if it doesn't exist
	createTestApplication()
	createTestEnvironment()

	// Wait for consistency
	time.Sleep(redisDelay)
	fmt.Println("✅ Test environment ready")
}

func createTestApplication() {
	app := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  "ai-test-app",
			"owner": "ai-demo-team",
		},
		"spec": map[string]interface{}{
			"description": "Application for testing AI-native deployment capabilities",
			"tags":        []string{"ai-test", "demo"},
			"lifecycle":   map[string]interface{}{},
		},
	}

	body, _ := json.Marshal(app)
	resp, err := http.Post(baseURL+"/v1/applications", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func createTestEnvironment() {
	env := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":  "ai-test-env",
			"owner": "ai-demo-team",
		},
		"spec": map[string]interface{}{
			"description": "Environment for AI testing",
		},
	}

	body, _ := json.Marshal(env)
	resp, err := http.Post(baseURL+"/v1/environments", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// Demo 1: Test AI-powered deployment planning
func testAIDeploymentPlanning() {
	fmt.Println("Testing AI deployment planning with preview...")

	// Use the enhanced deployment endpoint with plan=true query parameter
	payload := map[string]interface{}{
		"environment": "ai-test-env",
	}

	body, _ := json.Marshal(payload)
	url := baseURL + "/v1/applications/ai-test-app/deploy?plan=true"

	fmt.Printf("🔍 Requesting deployment plan: %s\n", url)
	startTime := time.Now()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to get deployment plan: %v\n", err)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	fmt.Printf("⏱️  Response time: %v\n", duration)

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Failed to decode response: %v\n", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ AI deployment plan generated successfully!")

		// Check if we got an AI-generated plan
		if plan, ok := result["plan"].(map[string]interface{}); ok {
			if source, ok := plan["planning_source"].(string); ok {
				fmt.Printf("🧠 Planning source: %s\n", source)
			}
			if steps, ok := plan["steps"].([]interface{}); ok {
				fmt.Printf("📋 Plan contains %d steps\n", len(steps))
			}
		}

		// Show reasoning if available
		if reasoning, ok := result["ai_reasoning"].(string); ok && reasoning != "" {
			fmt.Printf("💭 AI Reasoning: %s\n", truncateString(reasoning, 100))
		}
	} else {
		fmt.Printf("⚠️  Plan generation failed with status %d\n", resp.StatusCode)
		if errorMsg, ok := result["error"].(string); ok {
			fmt.Printf("📝 Error: %s\n", errorMsg)
		}
	}
}

// Demo 2: Test AI plan optimization
func testAIPlanOptimization() {
	fmt.Println("Testing AI plan optimization...")

	payload := map[string]interface{}{
		"environment": "ai-test-env",
	}

	body, _ := json.Marshal(payload)
	url := baseURL + "/v1/applications/ai-test-app/deploy?plan=true&optimize=true"

	fmt.Printf("🔍 Requesting optimized plan: %s\n", url)
	startTime := time.Now()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to get optimized plan: %v\n", err)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	fmt.Printf("⏱️  Response time: %v\n", duration)

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Failed to decode response: %v\n", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ AI-optimized deployment plan generated!")

		if plan, ok := result["plan"].(map[string]interface{}); ok {
			if optimizations, ok := plan["optimizations"].([]interface{}); ok {
				fmt.Printf("⚡ Applied %d optimizations\n", len(optimizations))
			}
		}
	} else {
		fmt.Printf("⚠️  Optimization failed with status %d\n", resp.StatusCode)
	}
}

// Demo 3: Test AI impact analysis
func testAIImpactAnalysis() {
	fmt.Println("Testing AI impact analysis...")

	payload := map[string]interface{}{
		"environment": "ai-test-env",
	}

	body, _ := json.Marshal(payload)
	url := baseURL + "/v1/applications/ai-test-app/deploy?analyze=true"

	fmt.Printf("🔍 Requesting impact analysis: %s\n", url)
	startTime := time.Now()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to get impact analysis: %v\n", err)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	fmt.Printf("⏱️  Response time: %v\n", duration)

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Failed to decode response: %v\n", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ AI impact analysis completed!")

		if analysis, ok := result["impact_analysis"].(map[string]interface{}); ok {
			if riskLevel, ok := analysis["risk_level"].(string); ok {
				fmt.Printf("🎯 Risk level: %s\n", riskLevel)
			}
			if recommendations, ok := analysis["recommendations"].([]interface{}); ok {
				fmt.Printf("💡 AI recommendations: %d items\n", len(recommendations))
			}
		}
	} else {
		fmt.Printf("⚠️  Impact analysis failed with status %d\n", resp.StatusCode)
	}
}

// Demo 4: Test AI policy evaluation
func testAIPolicyEvaluation() {
	fmt.Println("Testing AI policy evaluation...")

	payload := map[string]interface{}{
		"application_id": "ai-test-app",
		"environment":    "ai-test-env",
		"action":         "deploy",
		"context": map[string]interface{}{
			"version": "1.0.0",
			"time":    "business_hours",
		},
	}

	body, _ := json.Marshal(payload)
	url := baseURL + "/v1/policies/evaluate"

	fmt.Printf("🔍 Evaluating deployment policies: %s\n", url)
	startTime := time.Now()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to evaluate policies: %v\n", err)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	fmt.Printf("⏱️  Response time: %v\n", duration)

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Failed to decode response: %v\n", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ AI policy evaluation completed!")

		if allowed, ok := result["allowed"].(bool); ok {
			if allowed {
				fmt.Println("🟢 Deployment allowed by policies")
			} else {
				fmt.Println("🔴 Deployment blocked by policies")
			}
		}

		if reasoning, ok := result["ai_reasoning"].(string); ok && reasoning != "" {
			fmt.Printf("💭 AI Policy Reasoning: %s\n", truncateString(reasoning, 150))
		}
	} else {
		fmt.Printf("⚠️  Policy evaluation failed with status %d\n", resp.StatusCode)
	}
}

// Demo 5: Test natural language operations
func testNaturalLanguageOperations() {
	fmt.Println("Testing natural language AI operations...")

	// Test AI chat endpoint for natural language interaction
	payload := map[string]interface{}{
		"query":   "I want to deploy ai-test-app to ai-test-env but I'm concerned about potential issues. Can you help me understand the risks and create a safe deployment plan?",
		"context": "application: ai-test-app, environment: ai-test-env, user_intent: deployment_planning",
		"scope":   []string{"deployment", "planning"},
		"session": "demo-session",
		"timeout": 60,
	}

	body, _ := json.Marshal(payload)
	url := baseURL + "/v1/ai/chat"

	fmt.Printf("🔍 Natural language request: %s\n", url)
	fmt.Printf("💬 Message: \"I want to deploy ai-test-app...\"\n")
	startTime := time.Now()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Failed to process natural language request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	fmt.Printf("⏱️  Response time: %v\n", duration)

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Failed to decode response: %v\n", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ Natural language processing successful!")

		if response, ok := result["response"].(string); ok {
			fmt.Printf("🤖 AI Response: %s\n", truncateString(response, 200))
		}

		if actions, ok := result["suggested_actions"].([]interface{}); ok && len(actions) > 0 {
			fmt.Printf("💡 AI suggested %d actions\n", len(actions))
		}
	} else {
		fmt.Printf("⚠️  Natural language processing failed with status %d\n", resp.StatusCode)
	}
}

// Demo 6: Test AI error guidance
func testAIErrorGuidance() {
	fmt.Println("Testing AI error guidance...")

	// Intentionally try to deploy to a non-existent environment to trigger error guidance
	payload := map[string]interface{}{
		"environment": "non-existent-env",
	}

	body, _ := json.Marshal(payload)
	url := baseURL + "/v1/applications/ai-test-app/deploy"

	fmt.Printf("🔍 Attempting deployment to invalid environment...\n")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Failed to decode response: %v\n", err)
		return
	}

	// This should fail, but we want to see if AI provides helpful guidance
	if resp.StatusCode != http.StatusOK {
		fmt.Println("⚠️  Deployment failed as expected")

		if errorMsg, ok := result["error"].(string); ok {
			fmt.Printf("📝 Error: %s\n", errorMsg)
		}

		if guidance, ok := result["ai_guidance"].(string); ok && guidance != "" {
			fmt.Printf("🤖 AI Guidance: %s\n", truncateString(guidance, 200))
			fmt.Println("✅ AI provided helpful error guidance!")
		} else {
			fmt.Println("ℹ️  No AI guidance provided (this feature may need enhancement)")
		}
	}
}

// Demo 7: Test conversational deployment flow
func testConversationalDeployment() {
	fmt.Println("Testing conversational deployment flow...")

	messages := []string{
		"I need to deploy my application safely to production",
		"What are the current policies for production deployments?",
		"Can you create a deployment plan that follows best practices?",
		"What potential risks should I be aware of?",
	}

	for i, message := range messages {
		fmt.Printf("\n💬 Step %d: %s\n", i+1, message)

		payload := map[string]interface{}{
			"query":   message,
			"context": fmt.Sprintf("application: ai-test-app, environment: ai-test-env, conversation_id: demo-conversation, step: %d", i+1),
			"scope":   []string{"deployment", "conversation"},
			"session": "demo-conversation",
			"timeout": 30,
		}

		body, _ := json.Marshal(payload)
		url := baseURL + "/v1/ai/chat"

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("❌ Failed to decode: %v\n", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			if response, ok := result["response"].(string); ok {
				fmt.Printf("🤖 AI: %s\n", truncateString(response, 150))
			}
		} else {
			fmt.Printf("⚠️  Failed with status %d\n", resp.StatusCode)
		}

		// Small delay between conversation steps
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("✅ Conversational flow completed!")
}

// Helper function to truncate long strings for readable output
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
