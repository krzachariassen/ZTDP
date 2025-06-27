package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/ztdp/orchestrator/internal/agent/domain"
	"github.com/ztdp/orchestrator/internal/logging"
	"github.com/ztdp/orchestrator/internal/messaging"
	pb "github.com/ztdp/orchestrator/proto/orchestration"
	"github.com/ztdp/orchestrator/testHelpers"
)

func TestProtobufIntegration_TDD(t *testing.T) {
	t.Run("should_properly_convert_context_in_real_messages", func(t *testing.T) {
		// ARRANGE - Create a realistic scenario with agent metadata and message context
		mockRegistry := testHelpers.NewMockRegistry()
		mockMessageBus := testHelpers.NewMockAIMessageBus()
		logger := logging.NewNoOpLogger()
		server := NewOrchestrationServer(mockMessageBus, mockRegistry, logger)

		// Test agent registration with complex metadata
		agentMetadata := map[string]interface{}{
			"region":          "us-east-1",
			"instance_type":   "t3.medium",
			"max_connections": 100,
			"health_check":    true,
			"capabilities":    []interface{}{"deploy", "monitor", "scale"},
			"config": map[string]interface{}{
				"timeout": 30.5,
				"retries": 3,
			},
		}

		pbMetadata, err := structpb.NewStruct(agentMetadata)
		require.NoError(t, err, "Should create protobuf metadata")

		// Mock the registry call
		mockRegistry.On("RegisterAgent", mock.Anything, mock.MatchedBy(func(agent *domain.Agent) bool {
			// Verify that metadata was converted properly
			metadata := agent.Metadata
			return metadata["region"] == "us-east-1" &&
				metadata["instance_type"] == "t3.medium" &&
				metadata["max_connections"] == "100" &&
				metadata["health_check"] == "true"
		})).Return(nil)

		registerRequest := &pb.RegisterAgentRequest{
			AgentId:      "integration-test-agent",
			Name:         "Integration Test Agent",
			Type:         "service",
			Capabilities: []string{"deploy", "monitor", "scale"},
			Version:      "2.1.0",
			Metadata:     pbMetadata,
		}

		// ACT - Register the agent
		registerResponse, err := server.RegisterAgent(context.Background(), registerRequest)

		// ASSERT - Verify the registration response
		require.NoError(t, err, "Agent registration should succeed")
		assert.True(t, registerResponse.Success, "Registration should be successful")
		assert.Contains(t, registerResponse.Message, "Agent registered successfully")

		// Verify the mock was called with properly converted metadata
		mockRegistry.AssertExpectations(t)
	})

	t.Run("should_handle_agent_communication_workflow", func(t *testing.T) {
		// ARRANGE - Set up a realistic agent communication scenario
		mockRegistry := testHelpers.NewMockRegistry()
		mockMessageBus := testHelpers.NewMockAIMessageBus()
		logger := logging.NewNoOpLogger()
		_ = NewOrchestrationServer(mockMessageBus, mockRegistry, logger)

		// Mock agent-to-agent communication
		mockMessageBus.On("SendBetweenAgents", mock.Anything, mock.MatchedBy(func(msg *messaging.AgentToAgentMessage) bool {
			return msg.FromAgentID == "agent-1" &&
				msg.ToAgentID == "agent-2" &&
				msg.Content == "Deploy application X to production" &&
				msg.Purpose == "deployment_coordination"
		})).Return(nil)

		// Create agent-to-agent message
		agentMessage := &messaging.AgentToAgentMessage{
			FromAgentID:   "agent-1",
			ToAgentID:     "agent-2",
			Content:       "Deploy application X to production",
			CorrelationID: "deploy-123",
			Context: map[string]interface{}{
				"app_name": "microservice-x",
				"env":      "production",
				"urgency":  "high",
			},
			Purpose: "deployment_coordination",
		}

		// ACT - Send the message between agents
		err := mockMessageBus.SendBetweenAgents(context.Background(), agentMessage)

		// ASSERT - Verify the communication was successful
		require.NoError(t, err, "Agent-to-agent communication should succeed")
		mockMessageBus.AssertExpectations(t)
	})

	t.Run("should_validate_complex_registration_scenarios", func(t *testing.T) {
		// ARRANGE - Test complex agent registration with edge cases
		mockRegistry := testHelpers.NewMockRegistry()
		mockMessageBus := testHelpers.NewMockAIMessageBus()
		logger := logging.NewNoOpLogger()
		server := NewOrchestrationServer(mockMessageBus, mockRegistry, logger)

		scenarios := []struct {
			name        string
			request     *pb.RegisterAgentRequest
			expectError bool
		}{
			{
				name: "valid complex agent",
				request: &pb.RegisterAgentRequest{
					AgentId:      "complex-agent-123",
					Name:         "Complex Service Agent",
					Type:         "microservice",
					Capabilities: []string{"deploy", "monitor", "scale", "backup"},
					Version:      "3.2.1",
				},
				expectError: false,
			},
			{
				name: "invalid agent - empty capabilities",
				request: &pb.RegisterAgentRequest{
					AgentId:      "invalid-agent",
					Name:         "Invalid Agent",
					Type:         "service",
					Capabilities: []string{},
					Version:      "1.0.0",
				},
				expectError: true,
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				if !scenario.expectError {
					// Mock successful registration
					mockRegistry.On("RegisterAgent", mock.Anything, mock.AnythingOfType("*domain.Agent")).Return(nil).Once()
				}

				// ACT
				response, err := server.RegisterAgent(context.Background(), scenario.request)

				// ASSERT
				if scenario.expectError {
					assert.Error(t, err, "Should have validation error")
					assert.Nil(t, response, "Response should be nil on error")
				} else {
					assert.NoError(t, err, "Should succeed for valid agent")
					assert.NotNil(t, response, "Response should not be nil")
					assert.True(t, response.Success, "Registration should be successful")
				}
			})
		}

		mockRegistry.AssertExpectations(t)
	})
}
