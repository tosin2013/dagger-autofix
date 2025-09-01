#!/bin/bash

# CI/CD Readiness Check Script
# Verifies that all required files and configurations are in place

# Don't exit on errors, we want to show all issues
# set -e

echo "🔍 Checking CI/CD Pipeline Readiness"
echo "===================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
CHECKS_PASSED=0
CHECKS_FAILED=0
WARNINGS=0

# Function to check if file exists
check_file() {
    local file=$1
    local description=$2
    
    if [ -f "$file" ]; then
        echo -e "✅ ${GREEN}$description${NC}: $file"
        ((CHECKS_PASSED++))
    else
        echo -e "❌ ${RED}$description${NC}: $file (MISSING)"
        ((CHECKS_FAILED++))
    fi
}

# Function to check if directory exists
check_directory() {
    local dir=$1
    local description=$2
    
    if [ -d "$dir" ]; then
        echo -e "✅ ${GREEN}$description${NC}: $dir"
        ((CHECKS_PASSED++))
    else
        echo -e "❌ ${RED}$description${NC}: $dir (MISSING)"
        ((CHECKS_FAILED++))
    fi
}

# Function to check file content
check_content() {
    local file=$1
    local pattern=$2
    local description=$3
    
    if [ -f "$file" ] && grep -q "$pattern" "$file"; then
        echo -e "✅ ${GREEN}$description${NC}: Found in $file"
        ((CHECKS_PASSED++))
    else
        echo -e "⚠️  ${YELLOW}$description${NC}: Not found in $file"
        ((WARNINGS++))
    fi
}

echo "📁 Checking Directory Structure"
echo "------------------------------"
check_directory ".github" "GitHub Actions directory"
check_directory ".github/workflows" "Workflows directory"
check_directory "scripts" "Scripts directory"
echo ""

echo "🔄 Checking CI/CD Workflow Files"
echo "--------------------------------"
check_file ".github/workflows/ci.yml" "Main CI workflow"
check_file ".github/workflows/coverage-enforcement.yml" "Coverage enforcement workflow"
check_file ".github/workflows/security-analysis.yml" "Security analysis workflow"
check_file ".github/workflows/container-security.yml" "Container security workflow"
check_file ".github/workflows/release.yml" "Release workflow"
echo ""

echo "📋 Checking Configuration Files"
echo "-------------------------------"
check_file "go.mod" "Go module file"
check_file "go.sum" "Go dependencies"
check_file "dagger.json" "Dagger configuration"
check_file "Dockerfile" "Docker configuration"
check_file ".env.example" "Environment template"
check_file "README.md" "Documentation"
echo ""

echo "🧪 Checking Test Files"
echo "----------------------"
check_file "main_test.go" "Main test file"
check_file "enhanced_test.go" "Enhanced test file"
check_file "security_performance_test.go" "Security performance tests"
check_file "dagger_integration_test.go" "Dagger integration tests"
echo ""

echo "🔧 Checking Go Project Structure"
echo "--------------------------------"
check_file "main.go" "Main application file"
check_file "types.go" "Type definitions"
check_file "llm_client.go" "LLM client"
check_file "failure_analysis.go" "Failure analysis engine"
check_file "test_engine.go" "Test engine"
check_file "pull_request_engine.go" "PR engine"
echo ""

echo "📊 Checking Configuration Content"
echo "---------------------------------"
check_content "dagger.json" "github-autofix" "Dagger module name"
check_content "go.mod" "main" "Go module declaration"
check_content "README.md" "tosin2013/dagger-autofix" "Correct repository reference"
check_content ".github/workflows/ci.yml" "COVERAGE_THRESHOLD: 85" "Coverage threshold"
echo ""

echo "🔍 Checking Git Configuration"
echo "-----------------------------"
if git remote -v | grep -q "tosin2013/dagger-autofix"; then
    echo -e "✅ ${GREEN}Git remote${NC}: Correct repository configured"
    ((CHECKS_PASSED++))
else
    echo -e "⚠️  ${YELLOW}Git remote${NC}: Repository URL should be https://github.com/tosin2013/dagger-autofix.git"
    ((WARNINGS++))
fi

if git status --porcelain | grep -q .; then
    echo -e "⚠️  ${YELLOW}Git status${NC}: Uncommitted changes detected"
    ((WARNINGS++))
else
    echo -e "✅ ${GREEN}Git status${NC}: Working directory clean"
    ((CHECKS_PASSED++))
fi
echo ""

echo "🎯 Summary"
echo "=========="
echo -e "✅ ${GREEN}Checks Passed${NC}: $CHECKS_PASSED"
echo -e "❌ ${RED}Checks Failed${NC}: $CHECKS_FAILED"
echo -e "⚠️  ${YELLOW}Warnings${NC}: $WARNINGS"
echo ""

if [ $CHECKS_FAILED -eq 0 ]; then
    echo -e "🎉 ${GREEN}CI/CD Pipeline Ready!${NC}"
    echo ""
    echo "🚀 Ready to deploy! Run:"
    echo "   ./scripts/deploy-ci.sh"
    echo ""
    echo "📊 Expected outcomes after push:"
    echo "  • CI workflow will run and validate build/tests"
    echo "  • Coverage enforcement will check 85% threshold"
    echo "  • Security analysis will scan for vulnerabilities"
    echo "  • Container security will scan Docker images"
    echo "  • Quality gates will prevent merge if standards not met"
    echo ""
    exit 0
else
    echo -e "❌ ${RED}CI/CD Pipeline Not Ready${NC}"
    echo ""
    echo "🔧 Please fix the missing files/configurations above before deploying."
    echo ""
    exit 1
fi