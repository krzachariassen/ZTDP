package main

import (
	"context"
	"log"
	"time"

	pb "github.com/ztdp/agents/text-processor/proto/orchestration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the orchestrator
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrchestrationServiceClient(conn)

	// Test registration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.RegisterAgentRequest{
		AgentId:           "test-agent-001",
		Name:              "Test Agent",
		Type:              "test",
		Capabilities:      []string{"test-capability"},
		Version:           "1.0.0",
		MaxConcurrentWork: 1,
	}

	log.Printf("Sending registration request...")
	resp, err := client.RegisterAgent(ctx, req)
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	log.Printf("Registration response: Success=%v, Message=%s", resp.Success, resp.Message)
}
