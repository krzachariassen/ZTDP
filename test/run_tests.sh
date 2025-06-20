#!/bin/bash
# ZTDP AI Platform Test Runner
# Usage: ./run_tests.sh [test_phase]

set -e

BASE_URL="http://localhost:8080"
PHASE=${1:-basic}
WAIT_TIME=1  # Time to wait between tests

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

function log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

function success() {
    echo -e "${GREEN}✅ $1${NC}"
}

function error() {
    echo -e "${RED}❌ $1${NC}"
}

function warning() {
    echo -e "${YELLOW}⚠️ $1${NC}"
}

function run_test() {
    local test_name="$1"
    local message="$2"
    local expected_pattern="$3"
    
    log "🧪 Test: $test_name"
    echo "   Message: \"$message\""
    echo "   Expected: $expected_pattern"
    
    local result
    result=$(curl -s -X POST $BASE_URL/v3/ai/chat \
        -H "Content-Type: application/json" \
        -d "{\"message\": \"$message\"}" || echo "CURL_ERROR")
    
    if [[ "$result" == "CURL_ERROR" ]]; then
        error "Failed to connect to API server"
        return 1
    fi
    
    echo "   Response: $result"
    
    # Basic pattern matching for expected results
    if [[ "$result" =~ $expected_pattern ]]; then
        success "Test passed"
    else
        error "Test failed - response doesn't match expected pattern"
    fi
    
    echo "---"
    sleep $WAIT_TIME
}

function check_server() {
    log "🔍 Checking if ZTDP API server is running..."
    
    if curl -s $BASE_URL/v1/health > /dev/null 2>&1; then
        success "API server is running at $BASE_URL"
    else
        error "API server is not accessible at $BASE_URL"
        echo "Please start the ZTDP API server first with: ./build/api"
        exit 1
    fi
}

function phase_basic() {
    log "🚀 Phase 1: Basic Infrastructure Tests"
    
    # Test 1.1: Platform Health Check
    log "Test 1.1: Platform Health Check"
    health_result=$(curl -s $BASE_URL/v1/health)
    if [[ "$health_result" == *"OK"* ]] || [[ "$health_result" == "{}" ]]; then
        success "Health check passed"
    else
        error "Health check failed: $health_result"
    fi
    
    # Test 1.2: AI Chat Connectivity
    run_test "AI Chat Connectivity" \
        "Hello, can you help me with deployments?" \
        "(message|answer|response)"
    
    sleep 2
}

function phase_failures() {
    log "🚀 Phase 2: Basic Entity Creation Failures"
    
    # Test 2.1: Deploy Non-Existent Application
    run_test "Deploy Non-Existent Application" \
        "Deploy app-alpha to production" \
        "(not exist|not found|error|fail)"
    
    # Test 2.2: Create app then deploy to non-existent environment
    run_test "Create Application" \
        "Create application app-alpha" \
        "(created|success|app-alpha)"
    
    run_test "Deploy to Non-Existent Environment" \
        "Deploy app-alpha to production" \
        "(not exist|not found|error|fail)"
    
    sleep 2
}

function phase_success() {
    log "🚀 Phase 3: Success Path - Basic Deployment"
    
    # Test 3.1: Create environment and deploy
    run_test "Create Production Environment" \
        "Create a production environment" \
        "(created|success|production)"
    
    run_test "Deploy Application to Production" \
        "Deploy app-alpha to production" \
        "(deployed|success|production)"
    
    # Test 3.2: Verify deployment status
    run_test "Check Deployment Status" \
        "What is the status of app-alpha?" \
        "(status|app-alpha|production|deployed)"
    
    sleep 2
}

function phase_policies() {
    log "🚀 Phase 4: Policy Enforcement Tests"
    
    # Test 4.1: Create policy
    run_test "Create Production Block Policy" \
        "Create a policy that blocks direct production deployments. Applications must be deployed to development first." \
        "(policy|created|success|block)"
    
    # Test 4.2: Test policy enforcement
    run_test "Create App Beta" \
        "Create application app-beta" \
        "(created|success|app-beta)"
    
    run_test "Test Policy Enforcement" \
        "Deploy app-beta to production" \
        "(policy|violation|blocked|development|first)"
    
    # Test 4.3: Create development environment
    run_test "Create Development Environment" \
        "Create a development environment" \
        "(created|success|development)"
    
    # Test 4.4: Deploy to development first
    run_test "Deploy to Development First" \
        "Deploy app-beta to development" \
        "(deployed|success|development)"
    
    # Test 4.5: Now deploy to production (should succeed)
    run_test "Deploy to Production After Development" \
        "Deploy app-beta to production" \
        "(deployed|success|production)"
    
    sleep 2
}

function phase_advanced() {
    log "🚀 Phase 5: Advanced Policy Scenarios"
    
    # Test 5.1: Approval gate policy
    run_test "Create Approval Policy" \
        "Create a policy requiring manual approval for all production deployments" \
        "(policy|created|success|approval)"
    
    # Test 5.2: Test approval gate
    run_test "Create App Gamma" \
        "Create application app-gamma" \
        "(created|success|app-gamma)"
    
    run_test "Deploy Gamma to Development" \
        "Deploy app-gamma to development" \
        "(deployed|success|development)"
    
    run_test "Deploy Gamma to Production (Should Require Approval)" \
        "Deploy app-gamma to production" \
        "(approval|pending|waiting|manual)"
    
    sleep 2
}

function phase_complex() {
    log "🚀 Phase 6: Complex Orchestration Tests"
    
    # Test 6.1: Multi-application with dependencies
    run_test "Create Dependent Applications" \
        "Create application app-database and app-frontend. app-frontend depends on app-database." \
        "(created|success|depend)"
    
    run_test "Deploy with Dependencies" \
        "Deploy app-database to development, then app-frontend to development" \
        "(deployed|success|database|frontend)"
    
    # Test 6.2: Rollback scenario
    run_test "Rollback Application" \
        "Rollback app-beta in production to the previous version" \
        "(rollback|previous|version|success)"
    
    sleep 2
}

function phase_intelligence() {
    log "🚀 Phase 7: Agent Coordination and Intelligence Tests"
    
    # Test 7.1: Cross-agent information sharing
    run_test "Query Policies Affecting App" \
        "What policies are currently affecting app-beta deployments?" \
        "(policy|policies|affecting|app-beta)"
    
    # Test 7.2: Proactive recommendations
    run_test "Get Deployment Recommendations" \
        "What should I deploy next? Are there any recommendations?" \
        "(recommend|suggest|deploy|next)"
    
    # Test 7.3: Complex natural language
    run_test "Complex Natural Language Processing" \
        "I need to deploy the new user authentication service to prod ASAP, but I am worried about our weekend deployment policy" \
        "(weekend|policy|authentication|prod|concern)"
    
    sleep 2
}

function show_help() {
    echo "ZTDP AI Platform Test Runner"
    echo "Usage: $0 [PHASE]"
    echo ""
    echo "Available phases:"
    echo "  basic      - Basic infrastructure and connectivity tests"
    echo "  failures   - Test failure scenarios (missing entities)"
    echo "  success    - Test successful deployment path"
    echo "  policies   - Test policy creation and enforcement"
    echo "  advanced   - Test advanced policy scenarios"
    echo "  complex    - Test complex orchestration"
    echo "  intelligence - Test AI agent coordination"
    echo "  all        - Run all test phases"
    echo "  help       - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 basic"
    echo "  $0 policies"
    echo "  $0 all"
}

# Main execution
case $PHASE in
    "basic")
        check_server
        phase_basic
        ;;
    "failures")
        check_server
        phase_failures
        ;;
    "success")
        check_server
        phase_success
        ;;
    "policies")
        check_server
        phase_policies
        ;;
    "advanced")
        check_server
        phase_advanced
        ;;
    "complex")
        check_server
        phase_complex
        ;;
    "intelligence")
        check_server
        phase_intelligence
        ;;
    "all")
        check_server
        phase_basic
        phase_failures
        phase_success
        phase_policies
        phase_advanced
        phase_complex
        phase_intelligence
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        error "Unknown phase: $PHASE"
        show_help
        exit 1
        ;;
esac

log "🎉 Test phase '$PHASE' completed!"
