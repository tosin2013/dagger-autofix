#!/bin/bash

# Quick analysis script for workflow run 17385176718

set -e

RUN_ID="17385176718"
REPO="tosin2013/dagger-autofix"

echo "🔍 Analyzing Workflow Run: $RUN_ID"
echo "Repository: $REPO"
echo "=================================="
echo ""

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI not found. Please install it first:"
    echo "   brew install gh  # macOS"
    echo "   sudo apt install gh  # Ubuntu"
    exit 1
fi

echo "📋 Workflow Details:"
echo "--------------------"
gh run view $RUN_ID --repo $REPO 2>/dev/null || {
    echo "❌ Could not fetch workflow details. Please check:"
    echo "   1. GitHub CLI is authenticated: gh auth login"
    echo "   2. You have access to the repository"
    echo "   3. The run ID is correct"
    exit 1
}

echo ""
echo "🔧 Job Status:"
echo "--------------"
gh run view $RUN_ID --repo $REPO --json jobs --template '{{range .jobs}}{{printf "%-30s %s (%s)\n" .name .status .conclusion}}{{end}}'

echo ""
echo "📝 Getting logs for failed jobs..."
echo "-----------------------------------"

# Get failed job logs
gh run view $RUN_ID --repo $REPO --log 2>/dev/null | head -100

echo ""
echo "🔗 Full details: https://github.com/$REPO/actions/runs/$RUN_ID"