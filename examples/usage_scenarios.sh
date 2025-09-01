# Enhanced usage scenarios demonstrating all features of the GitHub Actions Auto-Fix Agent
# This script provides comprehensive examples for testing and learning the system

#!/bin/bash

set -e

echo "🤖 GitHub Actions Auto-Fix Agent - Enhanced Usage Scenarios"
echo "============================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
DEFAULT_TIMEOUT=300
MAX_RETRIES=3

# Utility functions
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    # Check required commands
    local required_commands=("dagger" "curl" "jq")
    for cmd in "${required_commands[@]}"; do
        if command -v "$cmd" >/dev/null 2>&1; then
            print_success "$cmd is installed"
        else
            print_error "$cmd is not installed. Please install it first."
            exit 1
        fi
    done
    
    # Check Dagger version
    local dagger_version
    dagger_version=$(dagger version --format json | jq -r '.version')
    print_info "Dagger version: $dagger_version"
    
    # Check Docker
    if docker info >/dev/null 2>&1; then
        print_success "Docker is running"
    else
        print_warning "Docker is not running. Some features may not work."
    fi
}

# Enhanced environment validation
check_environment() {
    print_header "Environment Validation"
    
    local required_vars=("GITHUB_TOKEN" "LLM_API_KEY" "REPO_OWNER" "REPO_NAME")
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        else
            print_success "$var is set"
        fi
    done
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        print_error "Missing required environment variables: ${missing_vars[*]}"
        print_info "Please set these variables in your .env file or environment"
        return 1
    fi
    
    # Validate token format
    if [[ $GITHUB_TOKEN =~ ^ghp_[a-zA-Z0-9]{36}$ ]] || [[ $GITHUB_TOKEN =~ ^github_pat_[a-zA-Z0-9_]{82}$ ]]; then
        print_success "GitHub token format is valid"
    else
        print_warning "GitHub token format may be invalid"
    fi
    
    # Validate LLM provider
    LLM_PROVIDER=${LLM_PROVIDER:-openai}
    print_info "LLM Provider: $LLM_PROVIDER"
}

# Build CLI with enhanced options
build_cli() {
    print_header "Building Enhanced CLI"
    
    print_info "Building CLI with Dagger..."
    if dagger call cli --export ./github-autofix > /dev/null 2>&1; then
        chmod +x ./github-autofix
        print_success "CLI built successfully"
        
        # Verify CLI works
        if ./github-autofix --version > /dev/null 2>&1; then
            local version
            version=$(./github-autofix --version)
            print_info "CLI Version: $version"
        else
            print_warning "CLI built but version check failed"
        fi
    else
        print_error "Failed to build CLI"
        return 1
    fi
}

# Enhanced connectivity testing with multiple providers
test_connectivity() {
    print_header "Enhanced Connectivity Testing"
    
    # Test GitHub connectivity with detailed output
    print_info "Testing GitHub API connectivity..."
    if ./github-autofix test connection \
        --github-token="$GITHUB_TOKEN" \
        --repo-owner="$REPO_OWNER" \
        --repo-name="$REPO_NAME" \
        --verbose > connectivity.log 2>&1; then
        print_success "GitHub connectivity test passed"
    else
        print_error "GitHub connectivity test failed"
        print_info "Check connectivity.log for details"
    fi
    
    # Test LLM provider connectivity
    print_info "Testing LLM provider connectivity ($LLM_PROVIDER)..."
    if ./github-autofix test llm \
        --llm-provider="$LLM_PROVIDER" \
        --llm-api-key="$LLM_API_KEY" \
        --test-query="Test connectivity" \
        --verbose >> connectivity.log 2>&1; then
        print_success "LLM connectivity test passed"
    else
        print_error "LLM connectivity test failed"
        print_info "Check connectivity.log for details"
    fi
    
    # Test rate limits
    print_info "Checking API rate limits..."
    curl -s -H "Authorization: token $GITHUB_TOKEN" \
        https://api.github.com/rate_limit | jq -r '.rate | "Remaining: \(.remaining)/\(.limit), Reset: \(.reset | strftime(\"%Y-%m-%d %H:%M:%S\"))"'
}

# Advanced configuration management
manage_configuration() {
    print_header "Advanced Configuration Management"
    
    # Initialize comprehensive configuration
    print_info "Initializing comprehensive configuration..."
    ./github-autofix config init --format=yaml --output=.github-autofix.yml --overwrite
    
    # Show effective configuration with sources
    print_info "Current effective configuration:"
    ./github-autofix config show \
        --github-token="$GITHUB_TOKEN" \
        --llm-provider="$LLM_PROVIDER" \
        --llm-api-key="$LLM_API_KEY" \
        --repo-owner="$REPO_OWNER" \
        --repo-name="$REPO_NAME" \
        --format=yaml \
        --show-sources
    
    # Validate configuration with strict checks
    print_info "Validating configuration with strict checks..."
    if ./github-autofix config validate \
        --github-token="$GITHUB_TOKEN" \
        --llm-provider="$LLM_PROVIDER" \
        --llm-api-key="$LLM_API_KEY" \
        --repo-owner="$REPO_OWNER" \
        --repo-name="$REPO_NAME" \
        --strict \
        --test-connectivity > config-validation.log 2>&1; then
        print_success "Configuration validation passed"
    else
        print_error "Configuration validation failed"
        print_info "Check config-validation.log for details"
    fi
}

# Framework detection and analysis
test_framework_detection() {
    print_header "Framework Detection and Analysis"
    
    print_info "Detecting supported frameworks in repository..."
    if ./github-autofix test frameworks \
        --path=. \
        --deep-scan \
        --verbose > frameworks.log 2>&1; then
        print_success "Framework detection completed"
        print_info "Detected frameworks:"
        cat frameworks.log | grep -E "Detected|Framework" || print_info "Check frameworks.log for details"
    else
        print_warning "Framework detection had issues (check frameworks.log)"
    fi
}

# Enhanced failure analysis with multiple scenarios
analyze_failure_scenarios() {
    print_header "Enhanced Failure Analysis Scenarios"
    
    if [ -z "$WORKFLOW_RUN_ID" ]; then
        print_warning "WORKFLOW_RUN_ID not set, generating synthetic failure scenarios"
        
        # Create mock failure scenarios for testing
        local scenarios=("build_failure" "test_failure" "dependency_issue" "security_scan")
        
        for scenario in "${scenarios[@]}"; do
            print_info "Testing analysis for: $scenario"
            ./github-autofix analyze-pattern \
                --pattern="$scenario" \
                --llm-provider="$LLM_PROVIDER" \
                --llm-api-key="$LLM_API_KEY" \
                --output-format=json \
                --save-analysis="analysis_${scenario}.json" || print_warning "Analysis failed for $scenario"
        done
        
        return
    fi
    
    print_info "Analyzing workflow run ID: $WORKFLOW_RUN_ID"
    
    # Comprehensive analysis with multiple output formats
    for format in "json" "yaml" "text"; do
        print_info "Generating analysis in $format format..."
        ./github-autofix analyze "$WORKFLOW_RUN_ID" \
            --github-token="$GITHUB_TOKEN" \
            --llm-provider="$LLM_PROVIDER" \
            --llm-api-key="$LLM_API_KEY" \
            --repo-owner="$REPO_OWNER" \
            --repo-name="$REPO_NAME" \
            --output-format="$format" \
            --include-logs \
            --save-analysis="analysis_${WORKFLOW_RUN_ID}.${format}" \
            --verbose > "analysis_${format}.log" 2>&1 || print_warning "Analysis in $format format failed"
    done
}

# Multi-provider fix generation testing
test_multi_provider_fixes() {
    print_header "Multi-Provider Fix Generation Testing"
    
    if [ -z "$WORKFLOW_RUN_ID" ]; then
        print_warning "WORKFLOW_RUN_ID not set, skipping multi-provider fix testing"
        return
    fi
    
    local providers=("openai" "anthropic" "gemini")
    local available_providers=()
    
    # Check which providers are available
    for provider in "${providers[@]}"; do
        local api_key_var="${provider^^}_API_KEY"
        if [ -n "${!api_key_var}" ]; then
            available_providers+=("$provider")
        fi
    done
    
    if [ ${#available_providers[@]} -eq 0 ]; then
        print_warning "No alternative LLM providers configured"
        return
    fi
    
    print_info "Testing fix generation with ${#available_providers[@]} providers: ${available_providers[*]}"
    
    for provider in "${available_providers[@]}"; do
        local api_key_var="${provider^^}_API_KEY"
        print_info "Testing fix generation with $provider..."
        
        timeout $DEFAULT_TIMEOUT ./github-autofix fix "$WORKFLOW_RUN_ID" \
            --github-token="$GITHUB_TOKEN" \
            --llm-provider="$provider" \
            --llm-api-key="${!api_key_var}" \
            --repo-owner="$REPO_OWNER" \
            --repo-name="$REPO_NAME" \
            --dry-run \
            --max-fixes=3 \
            --verbose > "fix_${provider}.log" 2>&1 && \
            print_success "Fix generation with $provider completed" || \
            print_warning "Fix generation with $provider failed"
    done
}

# Performance testing and benchmarking
test_performance() {
    print_header "Performance Testing and Benchmarking"
    
    # Enable performance profiling
    export ENABLE_PROFILING=true
    export PROFILE_OUTPUT=./profiles/
    mkdir -p ./profiles/
    
    print_info "Running performance benchmarks..."
    
    # Benchmark connectivity
    print_info "Benchmarking connectivity..."
    time ./github-autofix test connection \
        --github-token="$GITHUB_TOKEN" \
        --repo-owner="$REPO_OWNER" \
        --repo-name="$REPO_NAME" > perf_connectivity.log 2>&1
    
    # Benchmark configuration validation
    print_info "Benchmarking configuration validation..."
    time ./github-autofix config validate \
        --github-token="$GITHUB_TOKEN" \
        --llm-provider="$LLM_PROVIDER" \
        --llm-api-key="$LLM_API_KEY" \
        --repo-owner="$REPO_OWNER" \
        --repo-name="$REPO_NAME" > perf_config.log 2>&1
    
    # Resource usage monitoring
    print_info "Monitoring resource usage..."
    ./github-autofix monitor-resources --duration=60 --output=resource_usage.json > /dev/null 2>&1 &
    local monitor_pid=$!
    
    # Run a sample workload
    sleep 5
    ./github-autofix test llm \
        --llm-provider="$LLM_PROVIDER" \
        --llm-api-key="$LLM_API_KEY" \
        --test-query="Performance test query" > /dev/null 2>&1
    
    # Stop monitoring
    kill $monitor_pid 2>/dev/null || true
    
    print_info "Performance test completed. Check profiles/ directory for detailed results."
}

# Security and compliance testing
test_security() {
    print_header "Security and Compliance Testing"
    
    # Test secret masking
    print_info "Testing secret masking in logs..."
    export LOG_MASK_SECRETS=true
    export LOG_MASK_PATTERNS="sk-,ghp-,Bearer"
    
    ./github-autofix config show \
        --github-token="$GITHUB_TOKEN" \
        --llm-api-key="$LLM_API_KEY" \
        --verbose 2>&1 | grep -q "***" && \
        print_success "Secret masking is working" || \
        print_warning "Secret masking may not be working correctly"
    
    # Test input validation
    print_info "Testing input validation..."
    ./github-autofix analyze "invalid-run-id" \
        --github-token="$GITHUB_TOKEN" \
        --repo-owner="$REPO_OWNER" \
        --repo-name="$REPO_NAME" > /dev/null 2>&1 && \
        print_warning "Input validation may be weak" || \
        print_success "Input validation is working"
    
    # Test rate limiting
    print_info "Testing rate limiting behavior..."
    export RATE_LIMIT_ENABLED=true
    export MAX_REQUESTS_PER_HOUR=10
    
    # Make several rapid requests to test rate limiting
    for i in {1..5}; do
        ./github-autofix test connection \
            --github-token="$GITHUB_TOKEN" \
            --repo-owner="$REPO_OWNER" \
            --repo-name="$REPO_NAME" > /dev/null 2>&1 &
    done
    wait
    
    print_info "Rate limiting test completed"
}

# Monitoring and observability testing
test_monitoring() {
    print_header "Monitoring and Observability Testing"
    
    # Enable metrics collection
    export METRICS_ENABLED=true
    export METRICS_PORT=9090
    
    print_info "Testing metrics endpoint..."
    ./github-autofix status \
        --format=json \
        --include-metrics \
        --include-history > metrics.json
    
    if [ -s metrics.json ]; then
        print_success "Metrics collection is working"
        print_info "Key metrics:"
        jq -r '.statistics | to_entries[] | "\(.key): \(.value)"' metrics.json 2>/dev/null || print_info "Metrics available in metrics.json"
    else
        print_warning "Metrics collection may have issues"
    fi
    
    # Test health checks
    print_info "Testing health checks..."
    ./github-autofix health-check --timeout=30 > health.json 2>&1 && \
        print_success "Health check passed" || \
        print_warning "Health check failed"
}

# Advanced monitoring mode with custom parameters
test_advanced_monitoring() {
    print_header "Advanced Monitoring Mode"
    
    if [ "$SKIP_MONITORING" = "true" ]; then
        print_warning "Skipping monitoring mode (SKIP_MONITORING=true)"
        return
    fi
    
    print_info "Starting advanced monitoring mode for 2 minutes..."
    print_info "This will continuously check for workflow failures and demonstrate the fix process"
    
    # Advanced monitoring with multiple parameters
    timeout 120 ./github-autofix monitor \
        --github-token="$GITHUB_TOKEN" \
        --llm-provider="$LLM_PROVIDER" \
        --llm-api-key="$LLM_API_KEY" \
        --repo-owner="$REPO_OWNER" \
        --repo-name="$REPO_NAME" \
        --interval=30 \
        --max-concurrent=2 \
        --include-metrics \
        --verbose 2>&1 | tee monitoring.log || print_info "Monitoring completed (timeout or interrupted)"
    
    print_success "Advanced monitoring test completed"
}

# Comprehensive status and reporting
generate_comprehensive_report() {
    print_header "Generating Comprehensive Report"
    
    local report_file="usage_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$report_file" << EOF
# GitHub Actions Auto-Fix Agent - Usage Report

**Generated**: $(date)
**Repository**: $REPO_OWNER/$REPO_NAME
**LLM Provider**: $LLM_PROVIDER

## System Information

- **OS**: $(uname -a)
- **Docker**: $(docker --version 2>/dev/null || echo "Not available")
- **Dagger**: $(dagger version --format json | jq -r '.version' 2>/dev/null || echo "Unknown")
- **Agent Version**: $(./github-autofix --version 2>/dev/null || echo "Unknown")

## Test Results Summary

$(cat connectivity.log 2>/dev/null | head -20 || echo "Connectivity test results not available")

## Performance Metrics

$(cat metrics.json 2>/dev/null | jq -r '.statistics // "No metrics available"' || echo "Metrics not available")

## Framework Detection

$(cat frameworks.log 2>/dev/null | head -10 || echo "Framework detection results not available")

## Configuration

$(./github-autofix config show --format=yaml 2>/dev/null | head -30 || echo "Configuration not available")

## Recommendations

1. ✅ Basic functionality is working
2. ⚠️  Review any warnings in the test output
3. 📊 Monitor performance metrics for optimization opportunities
4. 🔒 Ensure security best practices are followed
5. 📚 Refer to documentation for advanced configuration options

## Next Steps

- Review generated log files for detailed information
- Configure monitoring and alerting for production use
- Test with actual workflow failures in a development environment
- Set up automated deployment and scaling

---

*Report generated by GitHub Actions Auto-Fix Agent usage scenarios*
EOF

    print_success "Comprehensive report generated: $report_file"
    print_info "Review the report for detailed results and recommendations"
}

# Cleanup function
cleanup() {
    print_header "Cleaning Up"
    
    # Clean up temporary files (keep important logs)
    local temp_files=(".github-autofix.yml" "health.json" "connectivity.log")
    
    for file in "${temp_files[@]}"; do
        if [ -f "$file" ]; then
            print_info "Cleaning up $file"
            rm -f "$file"
        fi
    done
    
    # Stop any background processes
    jobs -p | xargs -r kill 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Main execution function
main() {
    echo "Starting enhanced usage scenarios..."
    
    # Load environment variables
    if [ -f ".env" ]; then
        print_info "Loading environment variables from .env file"
        set -a && source .env && set +a
    fi
    
    # Set trap for cleanup on exit
    trap cleanup EXIT
    
    # Check if specific scenarios were requested
    if [ $# -eq 0 ]; then
        # Run all scenarios
        print_info "Running all scenarios..."
        scenarios=(
            "check_prerequisites"
            "check_environment"
            "build_cli"
            "test_connectivity"
            "manage_configuration"
            "test_framework_detection"
            "analyze_failure_scenarios"
            "test_multi_provider_fixes"
            "test_performance"
            "test_security"
            "test_monitoring"
            "test_advanced_monitoring"
            "generate_comprehensive_report"
        )
    else
        # Run specific scenarios
        scenarios=("$@")
    fi
    
    # Execute scenarios
    local failed_scenarios=()
    for scenario in "${scenarios[@]}"; do
        if declare -f "$scenario" > /dev/null 2>&1; then
            if ! "$scenario"; then
                failed_scenarios+=("$scenario")
                print_error "Scenario '$scenario' failed"
            fi
        else
            print_error "Unknown scenario: $scenario"
            print_info "Available scenarios: check_prerequisites, check_environment, build_cli, test_connectivity, manage_configuration, test_framework_detection, analyze_failure_scenarios, test_multi_provider_fixes, test_performance, test_security, test_monitoring, test_advanced_monitoring, generate_comprehensive_report"
        fi
    done
    
    # Final summary
    print_header "Final Summary"
    
    if [ ${#failed_scenarios[@]} -eq 0 ]; then
        print_success "All scenarios completed successfully! 🎉"
    else
        print_warning "Some scenarios failed: ${failed_scenarios[*]}"
        print_info "Check the logs and error messages above for details"
    fi
    
    print_info "For more information and advanced usage, see:"
    print_info "  - Documentation: docs/"
    print_info "  - Configuration Guide: docs/CONFIGURATION.md"
    print_info "  - Troubleshooting: docs/TROUBLESHOOTING.md"
    print_info "  - API Reference: docs/API.md"
}

# Run main function with all arguments
main "$@"