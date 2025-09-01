# Dagger GitHub Actions Auto-Fix Agent

A comprehensive Dagger.io agent for automatically resolving GitHub Actions pipeline failures with multi-LLM support and intelligent fix validation. This tool monitors GitHub Actions workflows, analyzes failures using LLM-powered intelligence, generates and validates fixes, and creates pull requests automatically.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/doc/install)
[![Dagger](https://img.shields.io/badge/Dagger-0.9+-purple.svg)](https://dagger.io)

## 🌟 Features

- 🤖 **Multi-LLM Support**: OpenAI GPT-4, Anthropic Claude, Google Gemini, DeepSeek, and LiteLLM proxy
- 🔍 **Intelligent Failure Analysis**: AI-powered root cause analysis with pattern matching and confidence scoring
- 🔧 **Automated Fix Generation**: Context-aware fix proposals with multiple alternatives and validation
- 🧪 **Comprehensive Testing**: Automated test generation and validation with customizable coverage requirements
- 📝 **Smart Pull Requests**: Automated PR creation with detailed explanations, fix rationale, and validation results  
- ⚙️ **Zero Configuration**: Works out of the box with sensible defaults and optional advanced customization
- 📈 **Monitoring & Metrics**: Real-time metrics, comprehensive logging, and performance analytics
- 💻 **Flexible Interface**: CLI tools, GitHub Actions integration, and Dagger module API
- 🔒 **Security First**: Secure credential handling, rate limiting, and audit logging
- 🚀 **Production Ready**: Robust error handling, retry logic, and scalability features

## 🚀 Quick Start

### Prerequisites

- **Docker** (for Dagger runtime) - [Installation Guide](https://docs.docker.com/get-docker/)
- **Dagger CLI** (v0.9.0+) - [Installation Guide](https://docs.dagger.io/cli)
- **GitHub Personal Access Token** with `repo`, `actions:read`, and `pull_requests:write` permissions
- **LLM Provider API Key** (OpenAI, Anthropic, Google, DeepSeek, or LiteLLM)
- **Go 1.21+** (for development) - [Installation Guide](https://golang.org/doc/install)

### Installation

#### Option 1: Using Dagger (Recommended)

```bash
# Use directly from Daggerverse
dagger -m github.com/your-org/dagger-autofix call \
  github-autofix \
  --github-token env:GITHUB_TOKEN \
  --llm-provider openai \
  --llm-api-key env:OPENAI_API_KEY \
  --repo-owner myorg \
  --repo-name myrepo \
  monitor
```

#### Option 2: Local Development Setup

```bash
# Clone the repository
git clone https://github.com/your-org/dagger-autofix.git
cd dagger-autofix

# Initialize configuration from template
cp .env.example .env
# Edit .env with your credentials (see Configuration section)

# Build and test the CLI
dagger call cli --export=./github-autofix
chmod +x ./github-autofix

# Test connectivity
./github-autofix test connection
```

#### Option 3: Docker Container

```bash
# Pull and run the container
docker run -it --rm \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  ghcr.io/your-org/dagger-autofix:latest \
  monitor --repo-owner myorg --repo-name myrepo
```

### Basic Usage Examples

#### 1. Continuous Monitoring (Recommended for Production)

```bash
# Monitor workflows and automatically fix failures
./github-autofix monitor \
  --github-token="$GITHUB_TOKEN" \
  --llm-provider="openai" \
  --llm-api-key="$OPENAI_API_KEY" \
  --repo-owner="myorg" \
  --repo-name="myrepo" \
  --target-branch="main" \
  --min-coverage=85 \
  --verbose
```

#### 2. Manual Failure Analysis and Fixing

```bash
# Step 1: Analyze a specific workflow failure
./github-autofix analyze 1234567890 --verbose

# Step 2: Generate fixes (dry run first)
./github-autofix fix 1234567890 --dry-run

# Step 3: Apply fixes (creates PR)
./github-autofix fix 1234567890

# Step 4: Validate the fix branch
./github-autofix validate autofix/fix-workflow-1234567890
```

#### 3. Batch Processing Multiple Failures

```bash
# Get recent failed runs and process them
for run_id in $(./github-autofix status --format=json | jq -r '.failed_runs[].id'); do
  echo "Processing run $run_id..."
  ./github-autofix fix $run_id --dry-run
done
```

#### 4. Integration with CI/CD Pipelines

```bash
# Use in your deployment scripts
./github-autofix monitor \
  --config=production.env \
  --log-level=warn \
  --log-format=json \
  --webhook-url="https://alerts.example.com/webhook"
```

## ⚙️ Configuration

The agent supports multiple configuration methods with environment variables taking precedence over configuration files.

### Environment Variables

Create a `.env` file or set environment variables:

```bash
# === REQUIRED CONFIGURATION ===

# GitHub Authentication
GITHUB_TOKEN=ghp_your_personal_access_token_here
# Alternative: GitHub App (recommended for organizations)
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
GITHUB_INSTALLATION_ID=12345678

# Repository Configuration  
REPO_OWNER=your_username_or_org
REPO_NAME=your_repository_name

# LLM Provider Configuration
LLM_PROVIDER=openai                    # openai, anthropic, gemini, deepseek, litellm
LLM_API_KEY=sk-your_llm_api_key_here

# === OPTIONAL CONFIGURATION ===

# Repository Settings
TARGET_BRANCH=main                     # Default branch for fixes
PROTECTED_BRANCHES=main,master,prod    # Branches to never directly modify

# Testing & Validation
MIN_COVERAGE=85                        # Minimum test coverage percentage
TEST_TIMEOUT=600                       # Test execution timeout (seconds)
ENABLE_INTEGRATION_TESTS=true          # Run integration tests during validation
TEST_FRAMEWORKS=go,jest,pytest,rspec   # Supported test frameworks

# Monitoring & Performance
MONITOR_INTERVAL=30                    # Workflow check interval (seconds)
MAX_CONCURRENT_FIXES=3                 # Maximum parallel fix operations
RATE_LIMIT_PER_HOUR=100               # GitHub API rate limit
CACHE_DURATION=3600                   # Analysis cache duration (seconds)

# Logging & Debugging
LOG_LEVEL=info                         # trace, debug, info, warn, error
LOG_FORMAT=json                        # json, text, structured
LOG_FILE=/var/log/github-autofix.log   # Optional log file path
VERBOSE=false                          # Enable verbose output

# Webhook & Notifications
WEBHOOK_URL=https://hooks.slack.com/... # Slack/Discord webhook for notifications
WEBHOOK_SECRET=your_webhook_secret       # Webhook signing secret
NOTIFICATION_CHANNEL=#ci-alerts          # Notification channel/recipient

# Security & Rate Limiting
MAX_FILE_SIZE=10MB                     # Maximum file size to analyze
MAX_ANALYSIS_TIME=300                  # Maximum analysis time (seconds)
ALLOWED_FILE_TYPES=.go,.js,.py,.yaml   # Comma-separated allowed file extensions
SECURITY_SCAN_ENABLED=true             # Enable security vulnerability scanning
```

### LLM Provider Configurations

#### OpenAI GPT-4 (Recommended)

```bash
LLM_PROVIDER=openai
LLM_API_KEY=sk-proj-your_openai_api_key
OPENAI_MODEL=gpt-4                     # Optional: gpt-4, gpt-4-turbo, gpt-3.5-turbo
OPENAI_BASE_URL=https://api.openai.com # Optional: custom endpoint
OPENAI_ORG_ID=org-your_organization_id # Optional: organization ID
```

#### Anthropic Claude

```bash
LLM_PROVIDER=anthropic
LLM_API_KEY=sk-ant-your_anthropic_key
ANTHROPIC_MODEL=claude-3-sonnet-20240229 # claude-3-sonnet, claude-3-haiku, claude-2.1
ANTHROPIC_VERSION=2023-06-01            # API version
```

#### Google Gemini

```bash
LLM_PROVIDER=gemini
LLM_API_KEY=your_google_api_key
GEMINI_MODEL=gemini-pro                # gemini-pro, gemini-pro-vision
GEMINI_PROJECT_ID=your_gcp_project_id  # Optional: GCP project ID
```

#### DeepSeek

```bash
LLM_PROVIDER=deepseek
LLM_API_KEY=sk-your_deepseek_key
DEEPSEEK_MODEL=deepseek-coder          # deepseek-coder, deepseek-chat
DEEPSEEK_BASE_URL=https://api.deepseek.com # Optional: custom endpoint
```

#### LiteLLM Proxy (Multi-Provider)

```bash
LLM_PROVIDER=litellm
LLM_API_KEY=your_litellm_token
LITELLM_BASE_URL=http://localhost:4000
LITELLM_MODEL=gpt-4                    # Model routed through LiteLLM
LITELLM_TIMEOUT=60                     # Request timeout (seconds)
```

### GitHub Authentication Options

#### Personal Access Token (Simple Setup)

1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate a new token with these scopes:
   - `repo` (full repository access)
   - `actions:read` (read Actions workflows)
   - `pull_requests:write` (create PRs)
   - `contents:write` (modify repository content)

#### GitHub App (Recommended for Organizations)

1. Create a GitHub App in your organization settings
2. Grant permissions: Repository contents, Pull requests, Actions, Issues
3. Install the app on target repositories
4. Use app credentials:

```bash
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY="$(cat private-key.pem)"
GITHUB_INSTALLATION_ID=12345678
```

### Configuration File Formats

#### YAML Configuration (`.github-autofix.yml`)

```yaml
# GitHub Configuration
github:
  token: ${GITHUB_TOKEN}
  app_id: 123456
  private_key_file: "private-key.pem"
  installation_id: 12345678

# Repository Settings  
repository:
  owner: "myorg"
  name: "myrepo"
  target_branch: "main"
  protected_branches: ["main", "master", "production"]

# LLM Configuration
llm:
  provider: "openai"
  api_key: ${OPENAI_API_KEY}
  model: "gpt-4"
  timeout: 60
  max_tokens: 4096

# Testing Configuration
testing:
  min_coverage: 85
  timeout: 600
  frameworks: ["go", "jest", "pytest"]
  integration_tests: true

# Monitoring Configuration  
monitoring:
  interval: 30
  max_concurrent: 3
  rate_limit: 100
  cache_duration: 3600

# Logging Configuration
logging:
  level: "info"
  format: "json"
  file: "/var/log/github-autofix.log"

# Notifications
notifications:
  webhook_url: "https://hooks.slack.com/..."
  channel: "#ci-alerts"
  on_success: true
  on_failure: true
```

#### JSON Configuration (`.github-autofix.json`)

```json
{
  "github": {
    "token": "${GITHUB_TOKEN}",
    "app_id": 123456,
    "private_key_file": "private-key.pem"
  },
  "repository": {
    "owner": "myorg",
    "name": "myrepo", 
    "target_branch": "main"
  },
  "llm": {
    "provider": "openai",
    "api_key": "${OPENAI_API_KEY}",
    "model": "gpt-4"
  },
  "testing": {
    "min_coverage": 85,
    "frameworks": ["go", "jest", "pytest"]
  }
}
```

### Configuration Validation

```bash
# Validate current configuration
./github-autofix config validate

# Test connectivity with current config
./github-autofix test connection

# Show effective configuration
./github-autofix config show --format=yaml

# Initialize configuration template
./github-autofix config init --format=yaml --output=.github-autofix.yml
```

## 🧩 Dagger Usage

### Using as a Dagger Module

#### Basic Module Usage

```go
// dagger/main.go
package main

import (
	"context"
	"fmt"
	"dagger/autofix/internal/dagger"
)

type MyModule struct{}

// AutoFix analyzes and fixes a specific workflow failure
func (m *MyModule) AutoFix(ctx context.Context, runId int) (*dagger.AutoFixResult, error) {
	agent := dag.GithubAutofix().
		WithGitHubToken(dag.SetSecret("github-token", os.Getenv("GITHUB_TOKEN"))).
		WithLLMProvider("openai", dag.SetSecret("openai-key", os.Getenv("OPENAI_API_KEY"))).
		WithRepository("myorg", "myrepo").
		WithTargetBranch("main").
		WithMinCoverage(85)

	// Initialize the agent
	initializedAgent, err := agent.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize agent: %w", err)
	}

	// Perform automated fix
	return initializedAgent.AutoFix(ctx, int64(runId))
}

// MonitorAndFix continuously monitors for failures and fixes them
func (m *MyModule) MonitorAndFix(ctx context.Context) error {
	agent := dag.GithubAutofix().
		WithGitHubToken(dag.SetSecret("github-token", os.Getenv("GITHUB_TOKEN"))).
		WithLLMProvider("anthropic", dag.SetSecret("anthropic-key", os.Getenv("ANTHROPIC_API_KEY"))).
		WithRepository("myorg", "myrepo")

	initializedAgent, err := agent.Initialize(ctx)
	if err != nil {
		return err
	}

	return initializedAgent.MonitorWorkflows(ctx)
}

// AnalyzeFailure provides detailed analysis without fixing
func (m *MyModule) AnalyzeFailure(ctx context.Context, runId int) (*dagger.FailureAnalysisResult, error) {
	agent := dag.GithubAutofix().
		WithGitHubToken(dag.SetSecret("github-token", os.Getenv("GITHUB_TOKEN"))).
		WithLLMProvider("gemini", dag.SetSecret("gemini-key", os.Getenv("GEMINI_API_KEY"))).
		WithRepository("myorg", "myrepo")

	initializedAgent, err := agent.Initialize(ctx)
	if err != nil {
		return nil, err
	}

	return initializedAgent.AnalyzeFailure(ctx, int64(runId))
}
```

#### Advanced Module Configuration

```go
// Advanced Dagger module with custom configuration
func (m *MyModule) AdvancedAutoFix(ctx context.Context) error {
	// Create container with custom environment
	container := dag.Container().
		From("golang:1.21-alpine").
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithDirectory("/workspace", dag.Host().Directory(".")).
		WithWorkdir("/workspace").
		WithSecretVariable("GITHUB_TOKEN", dag.SetSecret("github-token", os.Getenv("GITHUB_TOKEN"))).
		WithSecretVariable("OPENAI_API_KEY", dag.SetSecret("openai-key", os.Getenv("OPENAI_API_KEY")))

	// Build and run the autofix agent
	result, err := container.
		WithExec([]string{"dagger", "call", "cli", "--export", "./github-autofix"}).
		WithExec([]string{"./github-autofix", "monitor", 
			"--repo-owner", "myorg", 
			"--repo-name", "myrepo",
			"--llm-provider", "openai",
			"--min-coverage", "90",
			"--verbose"}).
		Stdout(ctx)

	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}
```

### GitHub Actions Integration

#### Comprehensive Workflow Integration

```yaml
# .github/workflows/autofix.yml
name: Auto-Fix Failed Builds

on:
  workflow_run:
    workflows: ["CI", "Build", "Test", "Deploy"]  # Monitor multiple workflows
    types: [completed]
  schedule:
    - cron: '0 */2 * * *'  # Run every 2 hours as backup
  workflow_dispatch:        # Manual trigger
    inputs:
      run_id:
        description: 'Specific workflow run ID to fix'
        required: false
        type: string
      dry_run:
        description: 'Run in dry-run mode'
        required: false
        type: boolean
        default: false

jobs:
  autofix:
    # Only run on failures or manual dispatch
    if: >
      github.event.workflow_run.conclusion == 'failure' || 
      github.event_name == 'schedule' ||
      github.event_name == 'workflow_dispatch'
    
    runs-on: ubuntu-latest
    timeout-minutes: 30
    
    permissions:
      contents: write
      pull-requests: write
      actions: read
      issues: write
    
    strategy:
      matrix:
        llm-provider: [openai]  # Can expand to multiple providers
      fail-fast: false
    
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          token: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Setup Dagger
        uses: dagger/dagger-for-github@v6
        with:
          version: "latest"
          
      - name: Determine Run ID
        id: run-id
        run: |
          if [ "${{ github.event_name }}" == "workflow_dispatch" ] && [ -n "${{ inputs.run_id }}" ]; then
            echo "run_id=${{ inputs.run_id }}" >> $GITHUB_OUTPUT
          elif [ "${{ github.event_name }}" == "workflow_run" ]; then
            echo "run_id=${{ github.event.workflow_run.id }}" >> $GITHUB_OUTPUT
          else
            # For scheduled runs, find the latest failed run
            echo "run_id=latest_failed" >> $GITHUB_OUTPUT
          fi
          
          if [ "${{ inputs.dry_run }}" == "true" ]; then
            echo "dry_run=--dry-run" >> $GITHUB_OUTPUT
          else
            echo "dry_run=" >> $GITHUB_OUTPUT
          fi
      
      - name: Test Connectivity
        run: |
          dagger call \
            github-autofix \
            --github-token env:GITHUB_TOKEN \
            --llm-provider ${{ matrix.llm-provider }} \
            --llm-api-key env:LLM_API_KEY \
            --repo-owner ${{ github.repository_owner }} \
            --repo-name ${{ github.event.repository.name }} \
            test-connectivity
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          LLM_API_KEY: ${{ secrets.OPENAI_API_KEY }}
      
      - name: Auto-Fix Workflow Failure
        id: autofix
        run: |
          if [ "${{ steps.run-id.outputs.run_id }}" == "latest_failed" ]; then
            # Monitor mode for scheduled runs
            timeout 300 dagger call \
              github-autofix \
              --github-token env:GITHUB_TOKEN \
              --llm-provider ${{ matrix.llm-provider }} \
              --llm-api-key env:LLM_API_KEY \
              --repo-owner ${{ github.repository_owner }} \
              --repo-name ${{ github.event.repository.name }} \
              --target-branch ${{ github.event.repository.default_branch }} \
              --min-coverage 85 \
              monitor || echo "Monitor completed"
          else
            # Fix specific run
            dagger call \
              github-autofix \
              --github-token env:GITHUB_TOKEN \
              --llm-provider ${{ matrix.llm-provider }} \
              --llm-api-key env:LLM_API_KEY \
              --repo-owner ${{ github.repository_owner }} \
              --repo-name ${{ github.event.repository.name }} \
              --target-branch ${{ github.event.repository.default_branch }} \
              auto-fix \
              --run-id ${{ steps.run-id.outputs.run_id }} \
              ${{ steps.run-id.outputs.dry_run }}
          fi
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          LLM_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          
      - name: Generate Summary Report
        if: always()
        run: |
          echo "## Auto-Fix Summary" >> $GITHUB_STEP_SUMMARY
          echo "- **Workflow Run**: ${{ steps.run-id.outputs.run_id }}" >> $GITHUB_STEP_SUMMARY
          echo "- **LLM Provider**: ${{ matrix.llm-provider }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Repository**: ${{ github.repository }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Status**: ${{ steps.autofix.outcome }}" >> $GITHUB_STEP_SUMMARY
          
      - name: Notify on Failure
        if: failure()
        uses: 8398a7/action-slack@v3
        with:
          status: failure
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
          text: "Auto-fix failed for workflow run ${{ steps.run-id.outputs.run_id }}"

  validate-fixes:
    needs: autofix
    if: success()
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        
      - name: Find Auto-Fix Branches
        id: branches
        run: |
          # Find branches created by auto-fix
          BRANCHES=$(git branch -r | grep "origin/autofix/" | sed 's/origin\///' | head -5)
          echo "branches<<EOF" >> $GITHUB_OUTPUT
          echo "$BRANCHES" >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT
          
      - name: Validate Fix Branches
        if: steps.branches.outputs.branches != ''
        run: |
          for branch in ${{ steps.branches.outputs.branches }}; do
            echo "Validating branch: $branch"
            dagger call \
              github-autofix \
              --github-token env:GITHUB_TOKEN \
              --repo-owner ${{ github.repository_owner }} \
              --repo-name ${{ github.event.repository.name }} \
              validate \
              --branch "$branch"
          done
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

#### Multi-Repository Workflow

```yaml
# .github/workflows/org-wide-autofix.yml
name: Organization-Wide Auto-Fix

on:
  schedule:
    - cron: '0 8 * * 1'  # Weekly on Mondays
  workflow_dispatch:
    inputs:
      repositories:
        description: 'Comma-separated list of repositories (owner/repo)'
        required: false
        default: 'all'

jobs:
  discover-repositories:
    runs-on: ubuntu-latest
    outputs:
      repositories: ${{ steps.repos.outputs.repositories }}
    
    steps:
      - name: Discover Repositories
        id: repos
        run: |
          if [ "${{ inputs.repositories }}" == "all" ]; then
            # Get all repositories in organization
            REPOS=$(gh api orgs/${{ github.repository_owner }}/repos --paginate \
              --jq '.[] | select(.archived == false) | "\(.owner.login)/\(.name)"' \
              | head -10)  # Limit for demo
          else
            REPOS="${{ inputs.repositories }}"
          fi
          
          echo "repositories<<EOF" >> $GITHUB_OUTPUT
          echo "$REPOS" | jq -R -s -c 'split("\n") | map(select(. != ""))' >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  autofix-repositories:
    needs: discover-repositories
    strategy:
      matrix:
        repository: ${{ fromJson(needs.discover-repositories.outputs.repositories) }}
      fail-fast: false
      max-parallel: 3
    
    runs-on: ubuntu-latest
    
    steps:
      - name: Parse Repository
        id: parse
        run: |
          IFS='/' read -r owner name <<< "${{ matrix.repository }}"
          echo "owner=$owner" >> $GITHUB_OUTPUT
          echo "name=$name" >> $GITHUB_OUTPUT
      
      - name: Auto-Fix Repository
        run: |
          dagger call \
            github-autofix \
            --github-token env:GITHUB_TOKEN \
            --llm-provider openai \
            --llm-api-key env:OPENAI_API_KEY \
            --repo-owner ${{ steps.parse.outputs.owner }} \
            --repo-name ${{ steps.parse.outputs.name }} \
            monitor \
            --duration 300  # 5 minute scan per repo
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

## 🏗️ Architecture

### System Overview

```mermaid
graph TB
    A[GitHub Actions Workflow] -->|Failure Event| B[GitHub Webhooks]
    B --> C[Auto-Fix Agent]
    C --> D[Failure Analysis Engine]
    D --> E[LLM Client]
    E --> F[Multi-LLM Providers]
    
    C --> G[GitHub Integration]
    G --> H[Repository Access]
    G --> I[PR Creation]
    
    D --> J[Test Engine]
    J --> K[Coverage Validation]
    J --> L[Fix Verification]
    
    C --> M[Monitoring & Metrics]
    M --> N[Prometheus Metrics]
    M --> O[Structured Logging]
    
    F --> P[OpenAI GPT-4]
    F --> Q[Anthropic Claude]
    F --> R[Google Gemini]
    F --> S[DeepSeek]
    F --> T[LiteLLM Proxy]
```

### Core Components

#### 1. **Main Module** (`main.go`)
- **Purpose**: Central orchestration and configuration management
- **Responsibilities**: 
  - Dagger module initialization and configuration
  - Component orchestration and dependency injection
  - Error handling and recovery mechanisms
  - Metrics collection and reporting
- **Key Methods**: `Initialize()`, `MonitorWorkflows()`, `AutoFix()`

#### 2. **Failure Analysis Engine** (`failure_analysis.go`)
- **Purpose**: AI-powered failure diagnosis and pattern recognition
- **Capabilities**:
  - Multi-language code failure analysis (Go, JavaScript, Python, etc.)
  - Log parsing and error extraction
  - Confidence scoring for different failure types
  - Historical pattern matching and learning
  - Fix strategy recommendation
- **Supported Failure Categories**:
  - Code failures (syntax, logic, type errors)
  - Dependency failures (missing packages, version conflicts)
  - Test failures (unit, integration, performance)
  - Build failures (compilation, asset generation)
  - Deployment failures (infrastructure, configuration)
  - Security failures (vulnerability scans, permissions)

#### 3. **LLM Client** (`llm_client.go`)
- **Purpose**: Multi-provider LLM integration and management
- **Features**:
  - Provider-agnostic interface
  - Automatic failover and load balancing
  - Rate limiting and cost optimization
  - Response caching and deduplication
  - Token usage tracking and budgeting
- **Supported Providers**: OpenAI, Anthropic, Google Gemini, DeepSeek, LiteLLM

#### 4. **Test Engine** (`test_engine.go`)
- **Purpose**: Automated testing and validation framework
- **Capabilities**:
  - Multi-framework test execution (Go, Jest, pytest, RSpec, etc.)
  - Coverage analysis and reporting
  - Performance regression detection
  - Integration test orchestration
  - Test result validation and scoring
- **Validation Levels**:
  - Syntax validation
  - Unit test execution
  - Integration test validation
  - Coverage threshold enforcement
  - Performance benchmark validation

#### 5. **Pull Request Engine** (`pull_request_engine.go`)
- **Purpose**: Automated PR creation and management
- **Features**:
  - Intelligent branch naming and management
  - Comprehensive PR descriptions with fix rationale
  - Automated reviewer assignment
  - CI/CD integration and status checks
  - Merge conflict detection and resolution
- **PR Enhancement**:
  - Before/after code diffs
  - Test coverage reports
  - Performance impact analysis
  - Security vulnerability fixes

#### 6. **GitHub Integration** (`types.go`)
- **Purpose**: GitHub API interactions and webhook handling
- **Functionality**:
  - Workflow run monitoring and analysis
  - Repository content management
  - PR and issue management
  - Actions API integration
  - Rate limiting and authentication
- **Authentication Methods**:
  - Personal Access Tokens
  - GitHub Apps (recommended)
  - OAuth Apps (for user flows)

#### 7. **CLI Interface** (`cli.go`)
- **Purpose**: Command-line interface for manual operations
- **Commands**:
  - `monitor`: Continuous workflow monitoring
  - `analyze`: Single failure analysis
  - `fix`: Generate and apply fixes
  - `validate`: Validate fix branches
  - `status`: System status and metrics
  - `config`: Configuration management
  - `test`: Connectivity and system tests

### Data Flow Architecture

```mermaid
sequenceDiagram
    participant GHA as GitHub Actions
    participant Agent as Auto-Fix Agent
    participant LLM as LLM Provider
    participant GitHub as GitHub API
    participant Tests as Test Engine
    
    GHA->>Agent: Workflow Failure Event
    Agent->>GitHub: Fetch Failure Logs
    Agent->>LLM: Analyze Failure Context
    LLM->>Agent: Failure Analysis + Fix Suggestions
    Agent->>Tests: Validate Fix Proposals
    Tests->>Agent: Validation Results
    Agent->>GitHub: Create Fix Branch
    Agent->>Tests: Run Tests on Fix Branch
    Tests->>Agent: Test Results + Coverage
    Agent->>GitHub: Create Pull Request
    GitHub->>Agent: PR Created Successfully
    Agent->>GHA: Report Fix Status
```

### Failure Types and Resolution Strategies

#### Code Failures
- **Detection**: Compilation errors, linting failures, type mismatches
- **Analysis**: AST parsing, syntax validation, semantic analysis
- **Resolution**: Code generation, import fixes, type corrections
- **Validation**: Compilation success, test execution, lint checks

#### Dependency Failures
- **Detection**: Package not found, version conflicts, dependency loops
- **Analysis**: Package manager logs, dependency trees, compatibility matrices
- **Resolution**: Version updates, alternative packages, lockfile fixes
- **Validation**: Dependency resolution, security scans, compatibility tests

#### Test Failures
- **Detection**: Test execution failures, assertion errors, timeout issues
- **Analysis**: Test output parsing, stack trace analysis, flaky test detection
- **Resolution**: Test fixes, mock updates, timing adjustments
- **Validation**: Test success, coverage maintenance, performance regression checks

#### Build Failures
- **Detection**: Build script errors, asset generation failures, configuration issues
- **Analysis**: Build logs, configuration files, environment differences
- **Resolution**: Build script fixes, configuration updates, environment adjustments
- **Validation**: Successful builds, artifact generation, deployment readiness

### Scalability and Performance

#### Horizontal Scaling
- **Multi-instance deployment**: Run multiple agent instances
- **Repository partitioning**: Distribute repositories across instances
- **Load balancing**: Distribute failure processing loads
- **Resource isolation**: Separate compute resources per instance

#### Performance Optimizations
- **Analysis caching**: Cache failure patterns and solutions
- **Parallel processing**: Concurrent failure analysis and fixes
- **Smart polling**: Adaptive monitoring intervals based on activity
- **Resource pooling**: Reuse LLM connections and GitHub clients

#### Monitoring and Observability
- **Metrics collection**: Prometheus metrics for all operations
- **Distributed tracing**: OpenTelemetry integration for request tracking
- **Structured logging**: JSON logs with correlation IDs
- **Health checks**: System health and dependency status monitoring

## 🧪 Testing

### Test Suite Overview

The project maintains comprehensive test coverage across all components with automated testing in CI/CD.

#### Test Categories

- **Unit Tests**: Individual component testing with mocking
- **Integration Tests**: End-to-end workflow testing
- **Contract Tests**: LLM provider API contract validation
- **Performance Tests**: Load testing and benchmarking
- **Security Tests**: Vulnerability scanning and security validation

### Running Tests

#### Local Development Testing

```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# Run tests with detailed output
go test -v -race -coverprofile=coverage.out ./... -args -test.verbose

# Run specific test suites
go test -run TestGithubAutofix -v ./...
go test -run TestLLMClient -v ./...
go test -run TestFailureAnalysis -v ./...

# Run tests with custom timeout
go test -timeout 30m ./...

# Parallel testing with custom settings
go test -parallel 4 -cpu 4 ./...
```

#### Benchmark Testing

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmarks
go test -bench=BenchmarkLLMClient -benchmem ./...
go test -bench=BenchmarkFailureAnalysis -benchmem ./...

# Compare benchmark results
go test -bench=. -benchmem ./... > bench_old.txt
# Make changes
go test -bench=. -benchmem ./... > bench_new.txt
benchcmp bench_old.txt bench_new.txt
```

#### Integration Testing with Dagger

```bash
# Test Dagger module functionality
dagger call test-all

# Test specific Dagger functions
dagger call test-connectivity --github-token env:GITHUB_TOKEN --llm-provider openai --llm-api-key env:OPENAI_API_KEY

# Test LLM provider integration
dagger call test-llm-providers \
  --openai-key env:OPENAI_API_KEY \
  --anthropic-key env:ANTHROPIC_API_KEY \
  --gemini-key env:GEMINI_API_KEY

# Test GitHub integration
dagger call test-github-integration \
  --github-token env:GITHUB_TOKEN \
  --repo-owner testorg \
  --repo-name testrepo
```

#### CLI Testing

```bash
# Build and test CLI
dagger call cli --export ./github-autofix
chmod +x ./github-autofix

# Test connectivity
./github-autofix test connection \
  --github-token "$GITHUB_TOKEN" \
  --repo-owner "testorg" \
  --repo-name "testrepo"

# Test LLM connectivity
./github-autofix test llm \
  --llm-provider openai \
  --llm-api-key "$OPENAI_API_KEY"

# Validate configuration
./github-autofix config validate \
  --github-token "$GITHUB_TOKEN" \
  --llm-provider openai \
  --llm-api-key "$OPENAI_API_KEY"

# Test with mock data
./github-autofix analyze 1234567890 \
  --github-token "$GITHUB_TOKEN" \
  --llm-provider openai \
  --llm-api-key "$OPENAI_API_KEY" \
  --repo-owner "testorg" \
  --repo-name "testrepo" \
  --dry-run
```

### Test Configuration

#### Test Environment Variables

```bash
# Test configuration (.env.test)
GITHUB_TOKEN=ghp_test_token_here
OPENAI_API_KEY=sk-test-key-here
TEST_REPO_OWNER=testorg
TEST_REPO_NAME=testrepo
TEST_LLM_PROVIDER=openai
TEST_MODE=true
SKIP_REAL_API_CALLS=true
USE_MOCK_RESPONSES=true
TEST_TIMEOUT=300
```

#### Mock Configuration

```go
// Enable mocks for testing
export USE_MOCKS=true
export MOCK_GITHUB_API=true
export MOCK_LLM_API=true
export RECORD_MODE=false  // Set to true to record new fixtures
```

### Continuous Integration Testing

#### GitHub Actions Test Workflow

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.21, 1.22]
        dagger-version: [0.9.0, latest]
      fail-fast: false
    
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v4
        
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
          
      - name: Setup Dagger
        uses: dagger/dagger-for-github@v6
        with:
          version: ${{ matrix.dagger-version }}
          
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ matrix.go-version }}-${{ hashFiles('**/go.sum') }}
          
      - name: Download dependencies
        run: go mod download
        
      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./...
        
      - name: Run integration tests
        run: dagger call test-all
        env:
          GITHUB_TOKEN: ${{ secrets.TEST_GITHUB_TOKEN }}
          OPENAI_API_KEY: ${{ secrets.TEST_OPENAI_API_KEY }}
          
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          
  e2e-test:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    
    steps:
      - uses: actions/checkout@v4
        
      - name: End-to-End Test
        run: dagger call e2e-test
        env:
          GITHUB_TOKEN: ${{ secrets.TEST_GITHUB_TOKEN }}
          OPENAI_API_KEY: ${{ secrets.TEST_OPENAI_API_KEY }}
          TEST_REPO_OWNER: github-actions-autofix
          TEST_REPO_NAME: test-repository
```

### Quality Gates

#### Coverage Requirements

- **Minimum Coverage**: 85% overall
- **Component Coverage**: 90% for core components
- **Critical Path Coverage**: 95% for failure analysis and fix generation
- **Integration Coverage**: 70% for external integrations

#### Performance Benchmarks

- **Analysis Time**: < 30 seconds for typical failures
- **Fix Generation**: < 60 seconds per fix proposal
- **Test Execution**: < 300 seconds for full validation
- **Memory Usage**: < 512MB during normal operations
- **API Rate Limits**: Stay within 80% of provider limits

## 🔧 Troubleshooting Guide

### Common Issues and Solutions

#### 1. Authentication Issues

**Problem**: `Authentication failed` or `Invalid credentials`

```bash
# Diagnose authentication
./github-autofix test connection --verbose

# Solutions:
# A) Check token validity
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# B) Verify token scopes (need: repo, actions:read, pull_requests:write)
# C) For GitHub Apps, check installation and permissions
# D) Check token expiration date
```

**Problem**: `Rate limit exceeded`

```bash
# Check current rate limit status  
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/rate_limit

# Solutions:
# A) Use GitHub App authentication (higher limits)
# B) Implement request caching
# C) Reduce monitoring frequency
export MONITOR_INTERVAL=60  # Increase from default 30 seconds
```

#### 2. LLM Provider Issues

**Problem**: `LLM request failed` or `Provider timeout`

```bash
# Test LLM connectivity
./github-autofix test llm --llm-provider openai --verbose

# Diagnose specific providers:
# OpenAI
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models

# Anthropic  
curl -H "x-api-key: $ANTHROPIC_API_KEY" https://api.anthropic.com/v1/messages

# Solutions:
# A) Check API key validity and quotas
# B) Verify model availability and permissions
# C) Implement fallback providers
export LLM_FALLBACK_PROVIDERS="anthropic,gemini"
# D) Adjust timeout settings
export LLM_TIMEOUT=120
```

#### 3. Analysis and Fix Generation Issues

**Problem**: `Analysis failed` or `No fixes generated`

```bash
# Enable debug mode for detailed analysis
export LOG_LEVEL=debug
./github-autofix analyze 1234567890 --verbose

# Common causes and solutions:
# A) Insufficient failure context
export INCLUDE_FULL_LOGS=true
export MAX_LOG_SIZE=50MB

# B) Unsupported programming language/framework
# Check supported frameworks in config
./github-autofix config show | grep frameworks

# C) Complex failure patterns
# Enable advanced analysis mode
export ENABLE_DEEP_ANALYSIS=true
export ANALYSIS_TIMEOUT=300
```

#### 4. Test and Validation Issues

**Problem**: `Tests failed to run` or `Coverage below threshold`

```bash
# Debug test execution
export TEST_VERBOSE=true
export TEST_DEBUG=true
./github-autofix validate main --verbose

# Solutions:
# A) Check test framework detection
./github-autofix detect-frameworks --path ./

# B) Adjust coverage requirements
export MIN_COVERAGE=75  # Temporarily lower threshold

# C) Fix test environment setup
export TEST_ENV_SETUP="npm install && pip install -r requirements.txt"

# D) Handle flaky tests
export RETRY_FAILED_TESTS=3
export PARALLEL_TEST_EXECUTION=false
```

#### 5. Pull Request Creation Issues

**Problem**: `Failed to create PR` or `Merge conflicts`

```bash
# Debug PR creation process
export PR_DEBUG=true
./github-autofix fix 1234567890 --dry-run --verbose

# Solutions:
# A) Check branch protection rules
# B) Resolve merge conflicts automatically
export AUTO_RESOLVE_CONFLICTS=true

# C) Use alternative branch naming
export PR_BRANCH_PREFIX="hotfix"
export PR_BRANCH_TEMPLATE="{{.prefix}}/{{.timestamp}}/{{.short_hash}}"

# D) Custom PR template
export PR_TEMPLATE_FILE=".github/AUTOFIX_PR_TEMPLATE.md"
```

### Performance Optimization

#### 1. Analysis Performance

```bash
# Enable analysis caching
export ENABLE_ANALYSIS_CACHE=true
export CACHE_TTL=3600  # 1 hour cache

# Optimize LLM usage
export LLM_BATCH_SIZE=5           # Batch multiple requests
export LLM_CACHE_RESPONSES=true   # Cache LLM responses
export LLM_COMPRESS_CONTEXT=true  # Compress large contexts

# Parallel processing
export MAX_PARALLEL_ANALYSIS=3
export ANALYSIS_WORKERS=4
```

#### 2. Memory Optimization

```bash
# Limit memory usage
export MEMORY_LIMIT=1GB
export GC_PERCENT=50              # More aggressive garbage collection
export MAX_LOG_BUFFER_SIZE=10MB   # Limit log buffer size

# Stream processing for large files
export STREAM_LARGE_FILES=true
export MAX_FILE_SIZE_IN_MEMORY=5MB
```

#### 3. Network Optimization

```bash
# Connection pooling
export HTTP_POOL_SIZE=10
export KEEP_ALIVE_TIMEOUT=30

# Request optimization
export ENABLE_COMPRESSION=true
export REQUEST_TIMEOUT=60
export RETRY_ATTEMPTS=3
export RETRY_BACKOFF=exponential
```

#### 4. Monitoring Optimization

```bash
# Adaptive monitoring
export ADAPTIVE_MONITORING=true
export MIN_MONITOR_INTERVAL=15
export MAX_MONITOR_INTERVAL=300

# Smart filtering
export IGNORE_PATTERNS="node_modules,*.log,*.tmp"
export SKIP_PRIVATE_REPOS=true
export MONITOR_ONLY_DEFAULT_BRANCH=true
```

### Debug Mode and Logging

#### Enable Comprehensive Debug Logging

```bash
# Maximum debug output
export LOG_LEVEL=trace
export VERBOSE=true
export DEBUG_ALL_COMPONENTS=true
export LOG_FORMAT=structured
export LOG_OUTPUT=both  # console and file

# Component-specific debugging
export DEBUG_LLM_CLIENT=true
export DEBUG_GITHUB_API=true  
export DEBUG_TEST_ENGINE=true
export DEBUG_ANALYSIS_ENGINE=true

# Save debug session
./github-autofix analyze 1234567890 --verbose 2>&1 | tee debug.log
```

#### Structured Debug Information

```bash
# Generate diagnostic report
./github-autofix debug-info > diagnostic_report.json

# System health check
./github-autofix health-check --format=json

# Performance profiling
export ENABLE_PROFILING=true
export PROFILE_OUTPUT=./profiles/
./github-autofix monitor --duration=300  # Profile 5 minutes
```

### Getting Help

#### Support Channels

- 📝 **Documentation**: [docs/](docs/) directory
- 🐛 **Issue Tracker**: [GitHub Issues](https://github.com/your-org/dagger-autofix/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/your-org/dagger-autofix/discussions)  
- 📧 **Email Support**: support@your-org.com
- 💬 **Community Chat**: [Discord/Slack Channel]

#### Reporting Issues

When reporting issues, please include:

1. **Environment Information**:
   ```bash
   ./github-autofix version --full
   ./github-autofix debug-info
   ```

2. **Minimal Reproduction Case**:
   - Repository and workflow run ID
   - Configuration used
   - Command that failed
   - Expected vs actual behavior

3. **Logs and Error Output**:
   ```bash
   # Capture detailed logs
   export LOG_LEVEL=debug
   ./github-autofix [command] --verbose 2>&1 | tee issue_logs.txt
   ```

4. **System Context**:
   - Operating system and version
   - Docker version
   - Dagger version
   - Go version (if building from source)

## 🤝 Contributing

We welcome contributions from the community! Please read our contribution guidelines before getting started.

### Development Setup

1. **Fork and Clone**:
   ```bash
   git clone https://github.com/your-username/dagger-autofix.git
   cd dagger-autofix
   ```

2. **Install Dependencies**:
   ```bash
   go mod download
   dagger install
   ```

3. **Setup Development Environment**:
   ```bash
   cp .env.example .env.dev
   # Edit .env.dev with your development credentials
   ```

4. **Verify Setup**:
   ```bash
   dagger call test-all
   go test ./...
   ```

### Contribution Process

1. **Create Feature Branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**:
   - Write code following Go conventions
   - Add comprehensive tests (maintain 85% coverage)
   - Update documentation as needed
   - Follow semantic commit messages

3. **Test Changes**:
   ```bash
   go test -race -coverprofile=coverage.out ./...
   dagger call test-all
   ./scripts/lint.sh
   ```

4. **Submit Pull Request**:
   - Clear description of changes and motivation
   - Link to related issues
   - Include test results and coverage report
   - Request review from maintainers

### Development Guidelines

#### Code Standards

- **Go Style**: Follow [Effective Go](https://golang.org/doc/effective_go.html) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- **Testing**: Write tests for all new functionality, maintain 85% overall coverage
- **Documentation**: Update relevant documentation for any user-facing changes
- **Error Handling**: Implement comprehensive error handling with context
- **Logging**: Use structured logging with appropriate levels

#### Pull Request Requirements

- [ ] Tests pass with coverage ≥ 85%
- [ ] Code follows project style guidelines  
- [ ] Documentation updated (README, API docs, examples)
- [ ] No breaking changes without major version bump
- [ ] Semantic commit messages
- [ ] PR description includes testing evidence

#### Semantic Commit Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

**Examples**:
```bash
feat(llm): add support for Google Gemini provider
fix(github): handle rate limiting more gracefully  
docs(readme): add troubleshooting section
test(analysis): add integration tests for failure detection
```

### Adding New Features

#### Adding a New LLM Provider

1. **Implement Provider Interface**:
   ```go
   // In llm_client.go
   type NewProviderClient struct {
       apiKey  string
       baseURL string
       client  *http.Client
   }
   
   func (c *NewProviderClient) AnalyzeFailure(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
       // Implementation
   }
   ```

2. **Add Configuration Support**:
   ```go
   // Update config structures and validation
   ```

3. **Write Tests**:
   ```go
   func TestNewProviderClient(t *testing.T) {
       // Comprehensive test suite
   }
   ```

4. **Update Documentation**:
   - Add provider to README configuration section
   - Add examples and usage instructions
   - Update API documentation

#### Adding New Failure Type Support

1. **Extend Analysis Engine**:
   ```go
   // In failure_analysis.go  
   func (e *FailureAnalysisEngine) analyzeNewFailureType(logs string) (*AnalysisResult, error) {
       // Implementation
   }
   ```

2. **Add Pattern Recognition**:
   ```go
   var newFailurePatterns = []FailurePattern{
       // Pattern definitions
   }
   ```

3. **Update Test Framework**:
   ```go
   // Add test cases and fixtures
   ```

### Release Process

#### Version Management

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

#### Release Steps

1. **Update Version**:
   ```bash
   # Update version in relevant files
   git tag v1.2.0
   git push origin v1.2.0
   ```

2. **Generate Changelog**:
   ```bash
   git cliff --output CHANGELOG.md
   ```

3. **Create Release**:
   - GitHub release with artifacts
   - Docker images
   - Daggerverse submission

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

```
MIT License

Copyright (c) 2024 Your Organization

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## 🆘 Support

### Getting Help

- 📖 **Documentation**: [Complete documentation](docs/) with guides and examples
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/your-org/dagger-autofix/issues) for bugs and feature requests
- 💬 **Community Discussion**: [GitHub Discussions](https://github.com/your-org/dagger-autofix/discussions) for questions and ideas  
- 📧 **Direct Support**: support@your-org.com for urgent issues
- 💬 **Chat**: [Community Discord/Slack] for real-time help

### Enterprise Support

For organizations requiring:
- **Priority Support**: Guaranteed response times and dedicated support
- **Custom Integrations**: Tailored integrations with internal tools
- **Professional Services**: Implementation assistance and training
- **SLA Agreements**: Service level agreements and uptime guarantees

Contact: enterprise@your-org.com

### Commercial Use

This project is MIT licensed and free for commercial use. For organizations that prefer commercial licensing or require additional legal protections, commercial licenses are available.

### Security

For security vulnerabilities, please email security@your-org.com instead of creating public issues. We take security seriously and will respond promptly to legitimate reports.

### Acknowledgments

- **Dagger.io**: For providing an excellent containerization and CI/CD platform
- **OpenAI, Anthropic, Google**: For advancing LLM capabilities that make this project possible
- **GitHub**: For comprehensive APIs and excellent developer platform
- **Contributors**: All community members who help improve this project
- **Open Source Community**: For tools, libraries, and inspiration

---

**⭐ Star this repository if you find it useful!**

**🔗 Share with your team to help them automate CI/CD maintenance!**

**🤝 Contribute to help make automated CI/CD fixes accessible to everyone!**