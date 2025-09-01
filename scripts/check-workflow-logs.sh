#!/bin/bash

# GitHub Workflow Log Checker Script
# Uses GitHub CLI to fetch and analyze workflow run logs

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
REPO="tosin2013/dagger-autofix"
RUN_ID="${1:-17385176718}"

echo -e "${BLUE}🔍 GitHub Workflow Log Analyzer${NC}"
echo "=================================="
echo "Repository: $REPO"
echo "Run ID: $RUN_ID"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}❌ GitHub CLI (gh) is not installed${NC}"
    echo "Please install it from: https://cli.github.com/"
    echo ""
    echo "Installation options:"
    echo "  macOS: brew install gh"
    echo "  Ubuntu: sudo apt install gh"
    echo "  Windows: winget install GitHub.cli"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${YELLOW}⚠️  Not authenticated with GitHub CLI${NC}"
    echo "Please run: gh auth login"
    exit 1
fi

echo -e "${GREEN}✅ GitHub CLI is ready${NC}"
echo ""

# Function to get workflow run details
get_workflow_details() {
    echo -e "${BLUE}📋 Workflow Run Details${NC}"
    echo "------------------------"
    
    gh run view $RUN_ID --repo $REPO --json status,conclusion,workflowName,headBranch,event,createdAt,updatedAt,url \
        --template '
Status: {{.status}}
Conclusion: {{.conclusion}}
Workflow: {{.workflowName}}
Branch: {{.headBranch}}
Event: {{.event}}
Created: {{.createdAt}}
Updated: {{.updatedAt}}
URL: {{.url}}
'
    echo ""
}

# Function to list jobs in the workflow run
list_jobs() {
    echo -e "${BLUE}🔧 Jobs in Workflow Run${NC}"
    echo "------------------------"
    
    gh run view $RUN_ID --repo $REPO --json jobs \
        --template '{{range .jobs}}{{.name}} - {{.status}} ({{.conclusion}})
{{end}}'
    echo ""
}

# Function to get logs for a specific job
get_job_logs() {
    local job_name="$1"
    echo -e "${BLUE}📝 Logs for Job: $job_name${NC}"
    echo "$(printf '=%.0s' {1..50})"
    
    # Get job logs
    gh run view $RUN_ID --repo $REPO --log --job "$job_name" 2>/dev/null || {
        echo -e "${YELLOW}⚠️  Could not fetch logs for job: $job_name${NC}"
        echo "This might be due to permissions or the job might still be running."
    }
    echo ""
}

# Function to analyze failed jobs
analyze_failures() {
    echo -e "${RED}🚨 Analyzing Failed Jobs${NC}"
    echo "-------------------------"
    
    # Get failed jobs
    failed_jobs=$(gh run view $RUN_ID --repo $REPO --json jobs \
        --template '{{range .jobs}}{{if eq .conclusion "failure"}}{{.name}}{{"\n"}}{{end}}{{end}}')
    
    if [ -z "$failed_jobs" ]; then
        echo -e "${GREEN}✅ No failed jobs found${NC}"
        return
    fi
    
    echo "Failed jobs:"
    echo "$failed_jobs"
    echo ""
    
    # Get logs for each failed job
    while IFS= read -r job; do
        if [ -n "$job" ]; then
            echo -e "${RED}❌ Failed Job: $job${NC}"
            get_job_logs "$job"
            echo "$(printf '=%.0s' {1..80})"
            echo ""
        fi
    done <<< "$failed_jobs"
}

# Function to get all logs (if specific job analysis isn't enough)
get_all_logs() {
    echo -e "${BLUE}📄 Complete Workflow Logs${NC}"
    echo "============================"
    
    gh run view $RUN_ID --repo $REPO --log 2>/dev/null || {
        echo -e "${YELLOW}⚠️  Could not fetch complete logs${NC}"
        echo "This might be due to permissions or the workflow might still be running."
    }
}

# Function to provide troubleshooting suggestions
provide_suggestions() {
    echo -e "${YELLOW}💡 Troubleshooting Suggestions${NC}"
    echo "==============================="
    echo ""
    echo "Common CI/CD issues and solutions:"
    echo ""
    echo "1. 🔧 Go Module Issues:"
    echo "   - Check go.mod file has correct module name"
    echo "   - Run 'go mod tidy' to clean dependencies"
    echo "   - Verify all imports use correct module path"
    echo ""
    echo "2. 🧪 Test Failures:"
    echo "   - Check test files have proper package declarations"
    echo "   - Verify test functions start with 'Test'"
    echo "   - Ensure all dependencies are available"
    echo ""
    echo "3. 📊 Coverage Issues:"
    echo "   - Add more unit tests to reach 85% threshold"
    echo "   - Check coverage exclusions are appropriate"
    echo "   - Verify test files are in correct locations"
    echo ""
    echo "4. 🔒 Security Scan Failures:"
    echo "   - Review security scan results"
    echo "   - Update dependencies with known vulnerabilities"
    echo "   - Add security exceptions if needed"
    echo ""
    echo "5. 🐳 Container Build Issues:"
    echo "   - Check Dockerfile syntax"
    echo "   - Verify base image availability"
    echo "   - Ensure all COPY paths exist"
    echo ""
}

# Main execution
main() {
    get_workflow_details
    list_jobs
    analyze_failures
    
    # Ask if user wants complete logs
    echo -e "${YELLOW}📋 Analysis Complete${NC}"
    echo ""
    read -p "Do you want to see complete workflow logs? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        get_all_logs
    fi
    
    echo ""
    provide_suggestions
    
    echo ""
    echo -e "${GREEN}🔗 Useful Commands:${NC}"
    echo "gh run list --repo $REPO                    # List recent runs"
    echo "gh run view $RUN_ID --repo $REPO           # View run details"
    echo "gh run rerun $RUN_ID --repo $REPO          # Rerun failed jobs"
    echo "gh run watch $RUN_ID --repo $REPO          # Watch run progress"
    echo ""
    echo -e "${BLUE}🌐 Web Interface:${NC}"
    echo "https://github.com/$REPO/actions/runs/$RUN_ID"
}

# Handle command line arguments
case "${1:-}" in
    -h|--help)
        echo "Usage: $0 [RUN_ID]"
        echo ""
        echo "Analyzes GitHub Actions workflow run logs"
        echo ""
        echo "Arguments:"
        echo "  RUN_ID    Workflow run ID (default: 17385176718)"
        echo ""
        echo "Examples:"
        echo "  $0                    # Analyze default run"
        echo "  $0 12345678901        # Analyze specific run"
        echo ""
        exit 0
        ;;
    "")
        # Use default run ID
        ;;
    *)
        # Use provided run ID
        RUN_ID="$1"
        ;;
esac

# Run main function
main