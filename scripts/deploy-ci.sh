#!/bin/bash

# Deploy CI/CD Pipeline Script
# This script pushes the comprehensive CI/CD implementation to GitHub

set -e

echo "🚀 Deploying Comprehensive CI/CD Pipeline to GitHub"
echo "Repository: https://github.com/tosin2013/dagger-autofix.git"
echo ""

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ Error: Not in a git repository"
    echo "Please run this script from the project root directory"
    exit 1
fi

# Check if we have uncommitted changes
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo "📝 Uncommitted changes detected"
    echo ""
    echo "Files to be committed:"
    git status --porcelain
    echo ""
fi

# Add all CI/CD related files
echo "📦 Adding CI/CD pipeline files..."
git add .github/workflows/
git add .env.example
git add README.md
git add Dockerfile
git add dagger.json
git add scripts/

# Show what will be committed
echo ""
echo "📋 Files staged for commit:"
git diff --cached --name-only
echo ""

# Create comprehensive commit message
COMMIT_MSG="feat: implement comprehensive CI/CD pipeline with 85% coverage enforcement

🔄 CI/CD Pipeline Features:
- Continuous Integration with build, test, and quality gates
- Coverage enforcement with 85% minimum threshold
- Security analysis with gosec, govulncheck, and Trivy
- Container security scanning and vulnerability assessment
- Automated multi-platform releases with Daggerverse publishing
- Quality gates preventing merge on failures

🧪 Testing & Coverage:
- Unit, integration, and functional test execution
- Race condition detection and parallel test execution
- Coverage trend analysis and baseline comparison
- Detailed coverage reports with improvement suggestions

🔒 Security & Compliance:
- Static security analysis for Go code
- Dependency vulnerability scanning
- Container image security scanning
- License compliance checking
- SARIF integration with GitHub Security tab

🚀 Release Automation:
- Multi-platform binary builds (Linux, macOS, Windows)
- Container image publishing to GitHub Container Registry
- Automated GitHub releases with artifacts and checksums
- Daggerverse module publishing integration

📊 Quality Gates:
- Build success required
- All tests must pass
- 85% minimum test coverage
- No critical security vulnerabilities
- Code style compliance

🛠️ Infrastructure:
- GitHub Actions workflows for all automation
- Docker multi-stage builds with security best practices
- Dagger module configuration for portable builds
- Comprehensive documentation and examples

Author: Tosin Akinosho <tosin.akinosho@gmail.com>
Repository: https://github.com/tosin2013/dagger-autofix.git"

# Commit the changes
echo "💾 Committing changes..."
git commit -m "$COMMIT_MSG"

echo ""
echo "✅ Changes committed successfully!"
echo ""
echo "🔍 Commit details:"
git log --oneline -1
echo ""

# Ask for confirmation before pushing
read -p "🚀 Push to GitHub? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "📤 Pushing to GitHub..."
    
    # Get current branch
    CURRENT_BRANCH=$(git branch --show-current)
    echo "Current branch: $CURRENT_BRANCH"
    
    # Push to origin
    git push origin "$CURRENT_BRANCH"
    
    echo ""
    echo "🎉 Successfully pushed to GitHub!"
    echo ""
    echo "🔗 View your repository: https://github.com/tosin2013/dagger-autofix"
    echo "🔄 Check CI/CD status: https://github.com/tosin2013/dagger-autofix/actions"
    echo ""
    echo "📊 Expected workflow runs:"
    echo "  ✅ Continuous Integration"
    echo "  ✅ Coverage Enforcement and Validation"
    echo "  ✅ Security Analysis"
    echo "  ✅ Container Security Scanning"
    echo ""
    echo "🎯 Next steps:"
    echo "  1. Monitor the first CI run for any issues"
    echo "  2. Check coverage reports and add tests if needed"
    echo "  3. Review security scan results"
    echo "  4. Create a test release tag to validate release workflow"
    echo ""
else
    echo "❌ Push cancelled. You can push manually with:"
    echo "   git push origin $CURRENT_BRANCH"
fi

echo ""
echo "🏁 CI/CD deployment script completed!"