// Text Processing Agent - A simple demonstration of the ZTDP Agent SDK
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ztdp/agents/text-processor/agent"
	"github.com/ztdp/agents/text-processor/textprocessor"
)

func main() {
	log.Println("🚀 Starting Text Processing Agent...")

	// Create the text processor handler
	handler := textprocessor.NewTextProcessor()

	// Create the agent
	textAgent := agent.NewAgent(
		"text-processor-001",
		"Text Processing Agent",
		handler,
	)

	// Configuration
	config := agent.Config{
		OrchestratorAddress: getEnv("ORCHESTRATOR_ADDRESS", "localhost:50051"),
		ReconnectInterval:   30,
	}

	// Start the agent
	if err := textAgent.Start(config); err != nil {
		log.Fatalf("❌ Failed to start agent: %v", err)
	}

	// Wait for shutdown signal
	log.Println("🎯 Text Processing Agent is running. Press Ctrl+C to stop.")
	log.Println("📋 Available capabilities:")
	for i, capability := range handler.GetCapabilities() {
		log.Printf("  %d. %s", i+1, capability)
	}

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-quit

	// Graceful shutdown
	log.Println("🛑 Shutting down Text Processing Agent...")
	if err := textAgent.Stop(); err != nil {
		log.Printf("❌ Error during shutdown: %v", err)
	}

	log.Println("✅ Text Processing Agent stopped gracefully")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
