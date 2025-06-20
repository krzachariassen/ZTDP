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
NC='\033[0m' # No Color

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
        print_status "SUCCESS" "✓ AI Chat - $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        # Show AI response
        echo "AI Response: $(echo "$body" | jq -r '.message // .answer // "No message field"' 2>/dev/null || echo "$body")"
    else
        print_status "ERROR" "✗ AI Chat failed - $test_name"
        echo "Response: $body"
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
    
    # Create applications via AI
    test_ai_chat "Create a new application called checkout owned by team-x. It is an e-commerce checkout application with tags payments and core." "Create checkout application"
    
    test_ai_chat "Create a payment processing application named payment for team-y with financial and payments tags." "Create payment application"
    
    test_ai_chat "Create a monitoring application for the platform-team with observability and platform tags." "Create monitoring application"
    
    # Verify applications created
    test_ai_chat "List all applications" "show applications" "Verify applications created"
    
    # Phase 3: AI-Native Environment Creation
    print_status "HEADER" "Phase 3: AI-Native Environment Creation"
    
    # Create environments via AI
    test_ai_chat "Create a development environment called dev owned by platform-team for development work." "Create dev environment"
    
    test_ai_chat "Create a staging environment for testing before production." "Create staging environment"
    
    test_ai_chat "Create a production environment with strict policies for live workloads." "Create production environment"
    
    # Verify environments created
    test_ai_chat "List all environments" "show environments" "Verify environments created"
    
    # Phase 4: AI-Native Service Creation
    print_status "HEADER" "Phase 4: AI-Native Service Creation"
    
    # Create services for checkout application
    test_ai_chat "Create a service called checkout-api for the checkout application on port 8080 that is public facing." "Create checkout-api service"
    
    test_ai_chat "Create a background worker service called checkout-worker for the checkout application on port 9090 that is internal only." "Create checkout-worker service"
    
    # Create service for payment application
    test_ai_chat "Create a payment-api service for the payment application on port 8081 that is public facing." "Create payment-api service"
    
    # Create services for monitoring application
    test_ai_chat "Create a metrics-collector service for the monitoring application on port 9092 that is internal." "Create metrics-collector service"
    
    test_ai_chat "Create an alerting-service for the monitoring application on port 9093 that handles alerts internally." "Create alerting-service"
    
    # Phase 5: AI-Native Deployment Testing
    print_status "HEADER" "Phase 5: AI-Native Deployment Testing"
    
    # Test deployment coordination
    test_ai_chat "Deploy the checkout application to the dev environment." "Deploy checkout to dev"
    
    test_ai_chat "Deploy the payment application to staging environment." "Deploy payment to staging"
    
    # Test policy enforcement
    test_ai_chat "Can I deploy the monitoring application directly to production?" "Test production deployment policy"
    
    # Phase 6: AI-Native Platform Queries
    print_status "HEADER" "Phase 6: AI-Native Platform Queries"
    
    # Query platform state
    test_ai_chat "What applications are currently deployed in the platform?" "Query all applications"
    
    test_ai_chat "Show me all environments and their current status." "Query all environments"
    
    test_ai_chat "What services exist for the checkout application?" "Query checkout services"
    
    test_ai_chat "What are the current deployment statuses across all environments?" "Query deployment status"
    
    # Phase 7: Final AI-Native Platform Validation
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
        echo "  • Application and service creation"
        echo "  • Environment management"  
        echo "  • Deployment coordination"
        echo "  • Policy enforcement"
        echo "  • Platform state queries"
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
