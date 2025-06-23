#!/bin/bash

# ZTDP AI-Native Platform Comprehensive Test Automation Script
# This script runs the complete AI-native test plan using /v3/ai/chat endpoint

# set -e  # Exit on any error

BASE_URL="http://localhost:8080"
DELAY=1  # Delay between requests for AI processing and event propagation

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\03    # Phase 8: Final AI-Native Platform Validation
    print_status "HEADER" "Phase 8: Final AI-Native Platform Validation"0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to print colored output
function print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO") echo -e "${BLUE}ℹ️  $message${NC}" ;;
        "SUCCESS") echo -e "${GREEN}✅ $message${NC}" ;;
        "WARNING") echo -e "${YELLOW}⚠️  $message${NC}" ;;
        "ERROR") echo -e "${RED}❌ $message${NC}" ;;
        "HEADER") echo -e "${BLUE}🔍 === $message ===${NC}" ;;
    esac
}

# Function to make HTTP request and check response
function make_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local test_name=$5
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    print_status "INFO" "Test $TOTAL_TESTS: $test_name"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$BASE_URL$url")
    else
        if [ -n "$data" ]; then
            response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" "$BASE_URL$url" \
                      -H "Content-Type: application/json" \
                      -d "$data")
        else
            response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" "$BASE_URL$url")
        fi
    fi
    
    # Extract HTTP status
    http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')
    
    # Check if status matches expected
    if [ "$http_code" = "$expected_status" ] || [ "$expected_status" = "any" ]; then
        print_status "SUCCESS" "✓ HTTP $http_code - $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_status "ERROR" "✗ Expected $expected_status, got $http_code - $test_name"
        if [ -n "$body" ]; then
            echo "Response: $body"
        fi
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    sleep $DELAY
}

# Function to check if server is running
function check_server() {
    print_status "INFO" "Checking if ZTDP server is running..."
    if curl -s "$BASE_URL/v1/health" > /dev/null; then
        print_status "SUCCESS" "Server is running"
    else
        print_status "ERROR" "Server is not running. Please start with: ./build/api"
        exit 1
    fi
}

# Function to test AI endpoint
function test_ai_chat() {
    local message=$1
    local test_name=$2
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    print_status "INFO" "Test $TOTAL_TESTS: $test_name"
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST "$BASE_URL/v3/ai/chat" \
              -H "Content-Type: application/json" \
              -d "{\"message\": \"$message\"}")
    
    http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')
    
    if [ "$http_code" = "200" ]; then
        # Check for error indicators in the AI response
        ai_message=$(echo "$body" | jq -r '.message // .answer // ""' 2>/dev/null)
        
        if [[ "$ai_message" == *"❌"* ]] || [[ "$ai_message" == *"Failed to"* ]] || [[ "$ai_message" == *"requires"* ]] || [[ "$ai_message" == *"is required"* ]]; then
            print_status "ERROR" "✗ AI Chat failed - $test_name"
            echo "Error: $ai_message"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        else
            print_status "SUCCESS" "✓ AI Chat - $test_name"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            # Show AI response
            echo "AI Response: $ai_message"
        fi
    else
        print_status "ERROR" "✗ AI Chat failed - $test_name"
        echo "Response: $body"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    sleep $DELAY
}

# Function to test AI endpoint with persistence validation
function test_ai_chat_with_validation() {
    local message=$1
    local test_name=$2
    local validation_type=$3  # "create_app", "create_env", "list_apps", etc.
    local expected_item=$4    # Item name to validate
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    print_status "INFO" "Test $TOTAL_TESTS: $test_name"
    
    # Make the AI request
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST "$BASE_URL/v3/ai/chat" \
              -H "Content-Type: application/json" \
              -d "{\"message\": \"$message\"}")
    
    http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//g')
    
    if [ "$http_code" = "200" ]; then
        # Check AI response status and message for errors
        ai_status=$(echo "$body" | jq -r '.actions[0].result.agent_response.status // "unknown"' 2>/dev/null)
        ai_message=$(echo "$body" | jq -r '.message // .answer // ""' 2>/dev/null)
        
        # Check if the response contains error indicators
        if [[ "$ai_message" == *"❌"* ]] || [[ "$ai_message" == *"Failed to"* ]] || [[ "$ai_message" == *"requires"* ]] || [[ "$ai_message" == *"is required"* ]]; then
            print_status "ERROR" "✗ AI Chat agent failed - $test_name"
            echo "Error message: $ai_message"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        elif [ "$ai_status" = "success" ]; then
            # Validate persistence if needed
            if [ "$validation_type" = "create_app" ] && [ -n "$expected_item" ]; then
                sleep 2  # Allow time for persistence
                graph_response=$(curl -s "$BASE_URL/v1/graph")
                if echo "$graph_response" | jq -e ".nodes[\"$expected_item\"]" > /dev/null 2>&1; then
                    print_status "SUCCESS" "✓ AI Chat + Persistence - $test_name"
                    PASSED_TESTS=$((PASSED_TESTS + 1))
                else
                    print_status "ERROR" "✗ AI Chat succeeded but data not persisted - $test_name"
                    echo "Expected item '$expected_item' not found in graph"
                    FAILED_TESTS=$((FAILED_TESTS + 1))
                    return
                fi
            else
                print_status "SUCCESS" "✓ AI Chat - $test_name"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            fi
        else
            print_status "ERROR" "✗ AI Chat agent failed - $test_name (agent status: $ai_status)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        
        # Show AI response
        echo "AI Response: $(echo "$body" | jq -r '.message // .answer // "No message field"' 2>/dev/null || echo "$body")"
    else
        print_status "ERROR" "✗ AI Chat HTTP failed - $test_name"
        echo "Response: $body"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    sleep $DELAY
}

# Function to validate graph structure (nodes + edges)
function validate_graph_structure() {
    local test_name=$1
    local expected_nodes=$2      # Comma-separated list of expected node names
    local expected_edges=$3      # Comma-separated list of expected edges in format "source:target"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    print_status "INFO" "Test $TOTAL_TESTS: $test_name"
    
    graph_response=$(curl -s "$BASE_URL/v1/graph")
    
    # Validate nodes exist
    local nodes_valid=true
    if [ -n "$expected_nodes" ]; then
        IFS=',' read -ra NODE_ARRAY <<< "$expected_nodes"
        for node in "${NODE_ARRAY[@]}"; do
            if ! echo "$graph_response" | jq -e ".nodes[\"$node\"]" > /dev/null 2>&1; then
                print_status "ERROR" "✗ Node '$node' not found in graph"
                nodes_valid=false
            fi
        done
    fi
    
    # Validate edges exist
    local edges_valid=true
    if [ -n "$expected_edges" ]; then
        IFS=',' read -ra EDGE_ARRAY <<< "$expected_edges"
        for edge in "${EDGE_ARRAY[@]}"; do
            IFS=':' read -ra EDGE_PARTS <<< "$edge"
            local source="${EDGE_PARTS[0]}"
            local target="${EDGE_PARTS[1]}"
            
            # Check if edge exists in graph (format may vary based on implementation)
            if ! echo "$graph_response" | jq -e ".edges[] | select(.source == \"$source\" and .target == \"$target\")" > /dev/null 2>&1; then
                # Alternative edge format check
                if ! echo "$graph_response" | jq -e ".edges[\"$source:$target\"] // .edges[\"$target:$source\"]" > /dev/null 2>&1; then
                    print_status "ERROR" "✗ Edge '$source -> $target' not found in graph"
                    edges_valid=false
                fi
            fi
        done
    fi
    
    if [ "$nodes_valid" = true ] && [ "$edges_valid" = true ]; then
        print_status "SUCCESS" "✓ Graph structure validation - $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_status "ERROR" "✗ Graph structure validation failed - $test_name"
        echo "Current graph structure:"
        echo "$graph_response" | jq '.'
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    sleep $DELAY
}

# Main test execution
function main() {
    print_status "HEADER" "ZTDP AI-Native Platform Comprehensive Test Suite"
    
    check_server
    
    # Phase 1: Platform Health & Basic AI Interaction
    print_status "HEADER" "Phase 1: Platform Health & Basic AI Interaction"
    
    # Health check
    make_request "GET" "/v1/health" "" "200" "Platform health check"
    
    # Verify empty platform state using AI chat
    test_ai_chat "List all applications" "list applications" "AI Platform State Check"
    test_ai_chat "Show all environments" "show environments" "AI Platform State Check"
    
    # Basic AI interaction
    test_ai_chat "Hello, can you help me with platform management?" "Basic AI interaction"
    
    # Phase 2: AI-Native Application Creation
    print_status "HEADER" "Phase 2: AI-Native Application Creation"
    
    # Create applications via AI with persistence validation
    test_ai_chat_with_validation "Create a new application called checkout owned by team-x. It is an e-commerce checkout application with tags payments and core." "Create checkout application" "create_app" "checkout"
    
    test_ai_chat_with_validation "Create a payment processing application named payment for team-y with financial and payments tags." "Create payment application" "create_app" "payment"
    
    test_ai_chat_with_validation "Create a monitoring application for the platform-team with observability and platform tags." "Create monitoring application" "create_app" "monitoring"
    
    # Verify applications created
    test_ai_chat "List all applications" "show applications"
    
    # Phase 3: AI-Native Environment Creation
    print_status "HEADER" "Phase 3: AI-Native Environment Creation"
    
    # Create environments via AI
    test_ai_chat "Create a development environment called dev owned by platform-team for development work." "Create dev environment"
    
    test_ai_chat "Create a staging environment for testing before production." "Create staging environment"
    
    test_ai_chat "Create a production environment with strict policies for live workloads." "Create production environment"
    
    # Verify environments created
    test_ai_chat "List all environments" "show environments" "Verify environments created"
    
    # Phase 3.5: AI-Native Policy Creation & Resource Catalog
    print_status "HEADER" "Phase 3.5: AI-Native Policy Creation & Resource Catalog"
    
    # Create deployment policies
    test_ai_chat "Create a deployment policy that requires approval for production deployments." "Create production approval policy"
    
    test_ai_chat "Create a policy that prevents direct deployment to production without staging validation." "Create staging-first policy"
    
    test_ai_chat "Create a resource policy that limits memory usage to 4GB for development environments." "Create dev resource limits policy"
    
    # Create resource catalog
    test_ai_chat "Create a PostgreSQL database resource in the resource catalog with version 14 and connection settings." "Create PostgreSQL resource"
    
    test_ai_chat "Create a Redis cache resource in the catalog with version 7.0 for caching workloads." "Create Redis resource"
    
    test_ai_chat "Create an S3 storage resource for file storage with encryption enabled." "Create S3 storage resource"
    
    # Verify policies and resources created
    test_ai_chat "List all policies in the platform." "List all policies"
    test_ai_chat "Show me all resources in the catalog." "List resource catalog"
    
    # Phase 4: AI-Native Service Creation with Resource Linking
    print_status "HEADER" "Phase 4: AI-Native Service Creation with Resource Linking"
    
    # Create services for checkout application with resource dependencies
    test_ai_chat "Create a service called checkout-api for the checkout application on port 8080 that is public facing and needs PostgreSQL database." "Create checkout-api service with DB"
    
    test_ai_chat "Create a background worker service called checkout-worker for the checkout application on port 9090 that is internal only and needs Redis cache." "Create checkout-worker service with Redis"
    
    # Create service for payment application with resources
    test_ai_chat "Create a payment-api service for the payment application on port 8081 that is public facing and requires PostgreSQL and S3 storage." "Create payment-api service with resources"
    
    # Create services for monitoring application
    test_ai_chat "Create a metrics-collector service for the monitoring application on port 9092 that is internal and stores data in PostgreSQL." "Create metrics-collector service with DB"
    
    test_ai_chat "Create an alerting-service for the monitoring application on port 9093 that handles alerts internally and uses Redis for caching." "Create alerting-service with Redis"
    
    # CRITICAL: Validate that edges are created between services and applications
    print_status "INFO" "Validating service->application edges were created..."
    validate_graph_structure "Service-Application Edges" "checkout,payment,monitoring,checkout-api,checkout-worker,payment-api,metrics-collector,alerting-service" "checkout-api:checkout,checkout-worker:checkout,payment-api:payment,metrics-collector:monitoring,alerting-service:monitoring"
    
    # CRITICAL: Validate resource->application edges
    print_status "INFO" "Validating resource->application dependencies..."
    test_ai_chat "Show me what resources are connected to the checkout application." "Validate checkout resources"
    test_ai_chat "Show me what resources are connected to the payment application." "Validate payment resources"
    
    # Phase 5: AI-Native Deployment Testing with Policy Enforcement
    print_status "HEADER" "Phase 5: AI-Native Deployment Testing with Policy Enforcement"
    
    # Test successful deployments (should work)
    test_ai_chat "Deploy the checkout application to the dev environment." "Deploy checkout to dev"
    
    test_ai_chat "Deploy the payment application to staging environment." "Deploy payment to staging"
    
    # CRITICAL: Test policy enforcement - these should FAIL or require approval
    print_status "INFO" "Testing policy enforcement - these should trigger policy violations..."
    
    test_ai_chat "Deploy the monitoring application directly to production without staging validation." "Test production deployment policy violation"
    
    test_ai_chat "Deploy the checkout application to production without approval." "Test production approval policy violation"
    
    # Test policy-compliant deployments
    test_ai_chat "Request approval to deploy payment application to production." "Request production deployment approval"
    
    test_ai_chat "Deploy checkout to staging first, then request production deployment." "Test staging-first policy compliance"
    
    # Validate deployment edges in graph
    print_status "INFO" "Validating deployment edges were created..."
    validate_graph_structure "Deployment Edges" "checkout,dev,payment,staging" "checkout:dev,payment:staging"
    
    # Phase 6: AI-Native Platform Queries
    print_status "HEADER" "Phase 6: AI-Native Platform Queries"
    
    # Query platform state
    test_ai_chat "What applications are currently deployed in the platform?" "Query all applications"
    
    test_ai_chat "Show me all environments and their current status." "Query all environments"
    
    test_ai_chat "What services exist for the checkout application?" "Query checkout services"
    
    test_ai_chat "What are the current deployment statuses across all environments?" "Query deployment status"
    
    # Phase 7: Critical Graph Structure Validation
    print_status "HEADER" "Phase 7: Critical Graph Structure Validation"
    
    # Comprehensive graph validation
    print_status "INFO" "Performing comprehensive graph structure validation..."
    
    # Validate all nodes exist
    validate_graph_structure "All Nodes Present" "checkout,payment,monitoring,dev,staging,production,checkout-api,checkout-worker,payment-api,metrics-collector,alerting-service" ""
    
    # Validate critical edges exist
    print_status "INFO" "Validating all critical edges are present..."
    validate_graph_structure "Service-Application Edges" "" "checkout-api:checkout,checkout-worker:checkout,payment-api:payment,metrics-collector:monitoring,alerting-service:monitoring"
    
    # Validate resource dependencies
    test_ai_chat "Show me the complete dependency graph for all applications including their services and resources." "Complete dependency validation"
    
    # Check for orphaned nodes (nodes without edges)
    test_ai_chat "Are there any services that are not connected to applications?" "Orphaned services check"
    
    test_ai_chat "Are there any resources that are not connected to applications?" "Orphaned resources check"
    
    # Phase 8: Final AI-Native Platform Validation
    print_status "HEADER" "Phase 7: Final AI-Native Platform Validation"
    
    # Comprehensive platform state queries
    test_ai_chat "Give me a complete summary of the current platform state including all applications, services, environments, and deployments." "AI comprehensive platform summary"
    
    test_ai_chat "What applications are currently deployed to production?" "AI production deployment status"
    
    test_ai_chat "Show me any policy violations or issues in the current platform." "AI policy validation"
    
    # Verify final platform state via REST APIs for validation
    test_ai_chat "Show all applications" "list applications" "Final verification: applications"
    test_ai_chat "Show all environments" "list environments" "Final verification: environments"
    
    # Print final results
    print_status "HEADER" "AI-Native Test Results Summary"
    print_status "INFO" "Total Tests: $TOTAL_TESTS"
    print_status "SUCCESS" "Passed: $PASSED_TESTS"
    
    if [ $FAILED_TESTS -gt 0 ]; then
        print_status "ERROR" "Failed: $FAILED_TESTS"
        print_status "WARNING" "Some tests failed. Check output above for details."
        exit 1
    else
        print_status "SUCCESS" "All tests passed! AI-native platform is working correctly."
        print_status "INFO" "The platform can now handle natural language requests for:"
        echo "  • Application and service creation with automatic linking"
        echo "  • Environment management with policy enforcement"  
        echo "  • Resource catalog management and dependency linking"
        echo "  • Policy creation and enforcement testing"
        echo "  • Deployment coordination with approval workflows"
        echo "  • Complete graph structure validation (nodes + edges)"
        echo "  • Platform state queries and dependency analysis"
    fi
}

# Show usage if help requested
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "ZTDP Platform Comprehensive Test Suite"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo "  --delay N      Set delay between requests (default: 1)"
    echo ""
    echo "Prerequisites:"
    echo "  • ZTDP server running on http://localhost:8080"
    echo "  • curl and jq installed"
    echo ""
    echo "This script will:"
    echo "  • Create a complete platform demonstration"
    echo "  • Test all major AI-native functionality"
    echo "  • Validate orchestrator coordination"
    echo "  • Provide comprehensive test results"
    exit 0
fi

# Parse command line arguments
while [ $# -gt 0 ]; do
    case $1 in
        --delay)
            DELAY="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main test suite
main
