package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/ztdp/orchestrator/internal/agent/domain"
	"github.com/ztdp/orchestrator/internal/logging"
	pb "github.com/ztdp/orchestrator/proto/orchestration"
	"github.com/ztdp/orchestrator/testHelpers"
)

func TestOrchestrationServer_RegisterAgent_Success(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockRegistry := testHelpers.NewMockRegistry()
	mockBus := testHelpers.NewMockAIMessageBus()

	// Use constructor with interface
	server := NewOrchestrationServer(mockBus, mockRegistry, logger)

	// Test data
	metadata, _ := structpb.NewStruct(map[string]interface{}{
		"version": "1.0.0",
		"region":  "us-east-1",
	})

	req := &pb.RegisterAgentRequest{
		AgentId:      "test-agent",
		Name:         "Test Agent",
		Type:         "deployment",
		Capabilities: []string{"deploy", "monitor"},
		Version:      "1.0.0",
		Metadata:     metadata,
	}

	// Mock expectations - use domain.Agent
	mockRegistry.On("RegisterAgent", mock.Anything, mock.MatchedBy(func(agent *domain.Agent) bool {
		return agent.ID == "test-agent" &&
			agent.Name == "Test Agent" &&
			len(agent.Capabilities) == 2
	})).Return(nil)

	// Execute
	resp, err := server.RegisterAgent(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Contains(t, resp.Message, "Agent registered successfully")

	// Verify mock was called
	mockRegistry.AssertExpectations(t)
}

func TestOrchestrationServer_RegisterAgent_ValidationFailure(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockRegistry := testHelpers.NewMockRegistry()
	mockBus := testHelpers.NewMockAIMessageBus()

	server := NewOrchestrationServer(mockBus, mockRegistry, logger)

	testCases := []struct {
		name string
		req  *pb.RegisterAgentRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "empty agent ID",
			req: &pb.RegisterAgentRequest{
				AgentId:      "",
				Name:         "Test Agent",
				Capabilities: []string{"deploy"},
			},
		},
		{
			name: "empty name",
			req: &pb.RegisterAgentRequest{
				AgentId:      "test-agent",
				Name:         "",
				Capabilities: []string{"deploy"},
			},
		},
		{
			name: "no capabilities",
			req: &pb.RegisterAgentRequest{
				AgentId:      "test-agent",
				Name:         "Test Agent",
				Capabilities: []string{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			resp, err := server.RegisterAgent(context.Background(), tc.req)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, resp)

			// Check it's a gRPC invalid argument error
			st, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
		})
	}

	// No registry calls should have been made
	mockRegistry.AssertExpectations(t)
}

func TestOrchestrationServer_RegisterAgent_RegistryFailure(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockRegistry := testHelpers.NewMockRegistry()
	mockBus := testHelpers.NewMockAIMessageBus()

	server := NewOrchestrationServer(mockBus, mockRegistry, logger)

	// Test data
	req := &pb.RegisterAgentRequest{
		AgentId:      "test-agent",
		Name:         "Test Agent",
		Type:         "deployment",
		Capabilities: []string{"deploy"},
	}

	// Mock expectations - registry fails
	mockRegistry.On("RegisterAgent", mock.Anything, mock.AnythingOfType("*domain.Agent")).
		Return(assert.AnError)

	// Execute
	resp, err := server.RegisterAgent(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Check it's a gRPC internal error
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())

	// Verify mock was called
	mockRegistry.AssertExpectations(t)
}

func TestOrchestrationServer_UnregisterAgent_Success(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockRegistry := testHelpers.NewMockRegistry()
	mockBus := testHelpers.NewMockAIMessageBus()

	server := NewOrchestrationServer(mockBus, mockRegistry, logger)

	// Test data
	req := &pb.UnregisterAgentRequest{
		AgentId: "test-agent",
	}

	// Mock expectations
	mockRegistry.On("UnregisterAgent", mock.Anything, "test-agent").Return(nil)

	// Execute
	resp, err := server.UnregisterAgent(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Contains(t, resp.Message, "Agent unregistered successfully")

	// Verify mock was called
	mockRegistry.AssertExpectations(t)
}

func TestOrchestrationServer_UnregisterAgent_ValidationFailure(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockRegistry := testHelpers.NewMockRegistry()
	mockBus := testHelpers.NewMockAIMessageBus()

	server := NewOrchestrationServer(mockBus, mockRegistry, logger)

	testCases := []struct {
		name string
		req  *pb.UnregisterAgentRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "empty agent ID",
			req: &pb.UnregisterAgentRequest{
				AgentId: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			resp, err := server.UnregisterAgent(context.Background(), tc.req)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, resp)

			// Check it's a gRPC invalid argument error
			st, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
		})
	}

	// No registry calls should have been made
	mockRegistry.AssertExpectations(t)
}

func TestOrchestrationServer_Heartbeat_Success(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockRegistry := testHelpers.NewMockRegistry()
	mockBus := testHelpers.NewMockAIMessageBus()

	server := NewOrchestrationServer(mockBus, mockRegistry, logger)

	// Test data
	req := &pb.HeartbeatRequest{
		AgentId: "test-agent",
		Status:  pb.AgentStatus_AGENT_STATUS_HEALTHY,
	}

	// Mock expectations
	mockRegistry.On("UpdateAgentLastSeen", mock.Anything, "test-agent").Return(nil)

	// Execute
	resp, err := server.Heartbeat(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)

	// Verify mock was called
	mockRegistry.AssertExpectations(t)
}

func TestOrchestrationServer_Heartbeat_ValidationFailure(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockRegistry := testHelpers.NewMockRegistry()
	mockBus := testHelpers.NewMockAIMessageBus()

	server := NewOrchestrationServer(mockBus, mockRegistry, logger)

	testCases := []struct {
		name string
		req  *pb.HeartbeatRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "empty agent ID",
			req: &pb.HeartbeatRequest{
				AgentId: "",
				Status:  pb.AgentStatus_AGENT_STATUS_HEALTHY,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			resp, err := server.Heartbeat(context.Background(), tc.req)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, resp)

			// Check it's a gRPC invalid argument error
			st, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
		})
	}

	// No registry calls should have been made
	mockRegistry.AssertExpectations(t)
}

// CreateTestAgent creates a test agent for use in tests
func CreateTestAgent() *domain.Agent {
	agent, _ := domain.NewAgent(
		"test-agent",
		"Test Agent",
		"Test agent for unit testing",
		[]domain.AgentCapability{
			{Name: "deploy", Description: "Deploy applications"},
			{Name: "monitor", Description: "Monitor systems"},
		},
	)
	// Set metadata after creation
	agent.Metadata = map[string]string{
		"version": "1.0.0",
		"region":  "us-east-1",
	}
	return agent
}
