# API Documentation

Complete API reference for the Dagger GitHub Actions Auto-Fix Agent, covering both the Dagger module API and CLI commands.

## Table of Contents

- [Dagger Module API](#dagger-module-api)
- [CLI Commands API](#cli-commands-api)
- [Types and Interfaces](#types-and-interfaces)
- [Error Codes](#error-codes)
- [Examples](#examples)

## Dagger Module API

### Core Module

#### `New(source ...*dagger.Directory) *DaggerAutofix`

Creates a new DaggerAutofix instance with optional source directory.

**Parameters:**
- `source` (*dagger.Directory, optional): Source directory for the project. Defaults to current directory.

**Returns:**
- `*DaggerAutofix`: New instance with default configuration

**Example:**
```go
// Use current directory
agent := dag.GithubAutofix()

// Use specific directory  
sourceDir := dag.Host().Directory("./my-project")
agent := dag.GithubAutofix(sourceDir)
```

#### `WithSource(source *dagger.Directory) *DaggerAutofix`

Configures the source directory for the agent.

**Parameters:**
- `source` (*dagger.Directory): Source directory containing the project

**Returns:**
- `*DaggerAutofix`: Updated instance

#### `WithGitHubToken(token *dagger.Secret) *DaggerAutofix`

Configures GitHub authentication token.

**Parameters:**
- `token` (*dagger.Secret): GitHub Personal Access Token or GitHub App JWT

**Returns:**
- `*DaggerAutofix`: Updated instance

**Required Scopes:**
- `repo`: Full repository access
- `actions:read`: Read Actions workflows and runs
- `pull_requests:write`: Create and manage pull requests
- `contents:write`: Modify repository contents

#### `WithLLMProvider(provider string, apiKey *dagger.Secret) *DaggerAutofix`

Configures the LLM provider and authentication.

**Parameters:**
- `provider` (string): LLM provider name
  - `"openai"`: OpenAI GPT models
  - `"anthropic"`: Anthropic Claude models  
  - `"gemini"`: Google Gemini models
  - `"deepseek"`: DeepSeek models
  - `"litellm"`: LiteLLM proxy
- `apiKey` (*dagger.Secret): Provider-specific API key

**Returns:**
- `*DaggerAutofix`: Updated instance

#### `WithRepository(owner, name string) *DaggerAutofix`

Configures the target GitHub repository.

**Parameters:**
- `owner` (string): Repository owner (user or organization)
- `name` (string): Repository name

**Returns:**
- `*DaggerAutofix`: Updated instance

#### `WithTargetBranch(branch string) *DaggerAutofix`

Sets the target branch for fixes (default: "main").

**Parameters:**
- `branch` (string): Branch name to target for fixes

**Returns:**
- `*DaggerAutofix`: Updated instance

#### `WithMinCoverage(coverage int) *DaggerAutofix`

Sets minimum test coverage requirement (default: 85).

**Parameters:**
- `coverage` (int): Minimum coverage percentage (0-100)

**Returns:**
- `*DaggerAutofix`: Updated instance

### Operational Methods

#### `Initialize(ctx context.Context) (*DaggerAutofix, error)`

Initializes all internal components and validates configuration.

**Parameters:**
- `ctx` (context.Context): Request context

**Returns:**
- `*DaggerAutofix`: Initialized instance
- `error`: Initialization error, if any

**Errors:**
- `ConfigurationError`: Invalid or missing configuration
- `AuthenticationError`: GitHub or LLM authentication failed
- `NetworkError`: Unable to connect to required services

#### `MonitorWorkflows(ctx context.Context) error`

Continuously monitors GitHub Actions workflows for failures and automatically fixes them.

**Parameters:**
- `ctx` (context.Context): Request context (use with timeout/cancellation)

**Returns:**
- `error`: Monitoring error, if any

**Behavior:**
- Polls every 30 seconds (configurable)
- Processes failures in parallel (max 3 concurrent)
- Creates fix branches and pull requests
- Continues until context cancellation

#### `AnalyzeFailure(ctx context.Context, runID int64) (*FailureAnalysisResult, error)`

Analyzes a specific workflow failure and provides detailed insights.

**Parameters:**
- `ctx` (context.Context): Request context
- `runID` (int64): GitHub Actions workflow run ID

**Returns:**
- `*FailureAnalysisResult`: Detailed analysis results
- `error`: Analysis error, if any

#### `AutoFix(ctx context.Context, runID int64) (*AutoFixResult, error)`

Performs complete automated fix workflow for a specific failure.

**Parameters:**
- `ctx` (context.Context): Request context  
- `runID` (int64): GitHub Actions workflow run ID

**Returns:**
- `*AutoFixResult`: Fix operation results
- `error`: Fix error, if any

**Process:**
1. Analyzes the failure
2. Generates fix proposals
3. Validates fixes through testing
4. Creates fix branch
5. Creates pull request
6. Returns results with PR information

#### `ValidateFixes(ctx context.Context, branch string) (*ValidationResult, error)`

Validates fixes on a specific branch by running tests and checks.

**Parameters:**
- `ctx` (context.Context): Request context
- `branch` (string): Branch name to validate

**Returns:**
- `*ValidationResult`: Validation results including test outcomes
- `error`: Validation error, if any

### Testing and Diagnostics

#### `TestConnectivity(ctx context.Context) (*ConnectivityResult, error)`

Tests connectivity to GitHub and LLM providers.

**Parameters:**
- `ctx` (context.Context): Request context

**Returns:**
- `*ConnectivityResult`: Connectivity test results
- `error`: Test error, if any

#### `GetStatus(ctx context.Context) (*SystemStatus, error)`

Returns current system status, metrics, and operational information.

**Parameters:**
- `ctx` (context.Context): Request context

**Returns:**
- `*SystemStatus`: System status and metrics
- `error`: Status retrieval error, if any

## CLI Commands API

### Global Flags

All commands support these global flags:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `.github-autofix.env` | Configuration file path |
| `--github-token` | string | - | GitHub authentication token |
| `--llm-provider` | string | `openai` | LLM provider (openai, anthropic, gemini, deepseek, litellm) |
| `--llm-api-key` | string | - | LLM provider API key |
| `--repo-owner` | string | - | GitHub repository owner |
| `--repo-name` | string | - | GitHub repository name |
| `--target-branch` | string | `main` | Target branch for fixes |
| `--min-coverage` | int | `85` | Minimum test coverage percentage |
| `--verbose` | bool | `false` | Enable verbose logging |
| `--dry-run` | bool | `false` | Dry run mode (no actual changes) |
| `--log-level` | string | `info` | Log level (trace, debug, info, warn, error) |
| `--log-format` | string | `json` | Log format (json, text) |

### Commands

#### `monitor`

Continuously monitor GitHub Actions workflows for failures and automatically fix them.

```bash
github-autofix monitor [flags]
```

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--duration` | duration | unlimited | Maximum monitoring duration |
| `--interval` | duration | `30s` | Check interval |
| `--max-concurrent` | int | `3` | Maximum concurrent fixes |

**Examples:**
```bash
# Basic monitoring
github-autofix monitor

# Monitor for 1 hour with custom interval
github-autofix monitor --duration=1h --interval=60s

# Monitor with increased concurrency
github-autofix monitor --max-concurrent=5
```

#### `analyze`

Analyze a specific workflow failure and provide detailed insights.

```bash
github-autofix analyze <workflow-run-id> [flags]
```

**Arguments:**
- `workflow-run-id` (required): GitHub Actions workflow run ID

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output-format` | string | `json` | Output format (json, yaml, text) |
| `--save-analysis` | string | - | Save analysis to file |
| `--include-logs` | bool | `true` | Include full failure logs |

**Examples:**
```bash
# Basic analysis
github-autofix analyze 1234567890

# Analysis with text output saved to file
github-autofix analyze 1234567890 --output-format=text --save-analysis=analysis.txt

# Analysis without full logs (faster)  
github-autofix analyze 1234567890 --include-logs=false
```

#### `fix`

Generate and apply fixes for a workflow failure.

```bash
github-autofix fix <workflow-run-id> [flags]
```

**Arguments:**
- `workflow-run-id` (required): GitHub Actions workflow run ID

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | bool | `false` | Generate fixes without applying |
| `--auto-merge` | bool | `false` | Automatically merge PR if tests pass |
| `--reviewer` | string | - | Assign PR reviewer |
| `--max-fixes` | int | `3` | Maximum number of fix alternatives |

**Examples:**
```bash
# Basic fix (creates PR)
github-autofix fix 1234567890

# Dry run (analysis only)
github-autofix fix 1234567890 --dry-run

# Fix with auto-merge and reviewer
github-autofix fix 1234567890 --auto-merge --reviewer=maintainer
```

#### `validate`

Validate fixes by running tests on a specific branch.

```bash
github-autofix validate <branch> [flags]
```

**Arguments:**
- `branch` (required): Branch name to validate

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--test-timeout` | duration | `10m` | Test execution timeout |
| `--parallel-tests` | bool | `true` | Run tests in parallel |
| `--coverage-report` | string | - | Save coverage report to file |

**Examples:**
```bash
# Basic validation
github-autofix validate autofix/fix-123

# Validation with custom timeout and coverage report
github-autofix validate autofix/fix-123 --test-timeout=15m --coverage-report=coverage.html
```

#### `status`

Show agent status, metrics, and operational information.

```bash
github-autofix status [flags]
```

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `table` | Output format (table, json, yaml) |
| `--include-metrics` | bool | `true` | Include performance metrics |
| `--include-history` | bool | `false` | Include operation history |

**Examples:**
```bash
# Basic status
github-autofix status

# JSON status with history
github-autofix status --format=json --include-history
```

### Configuration Commands

#### `config init`

Initialize configuration file with template.

```bash
github-autofix config init [flags]
```

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `env` | Config format (env, yaml, json) |
| `--output` | string | `.env` | Output file path |
| `--overwrite` | bool | `false` | Overwrite existing file |

#### `config show`

Display current effective configuration.

```bash
github-autofix config show [flags]
```

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `yaml` | Output format (yaml, json, env) |
| `--show-secrets` | bool | `false` | Show secret values (masked) |

#### `config validate`

Validate current configuration and test connectivity.

```bash
github-autofix config validate [flags]
```

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--strict` | bool | `false` | Strict validation mode |
| `--test-connectivity` | bool | `true` | Test external service connectivity |

### Testing Commands

#### `test connection`

Test GitHub API connectivity and authentication.

```bash
github-autofix test connection [flags]
```

#### `test llm`

Test LLM provider connectivity and authentication.

```bash
github-autofix test llm [flags]
```

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--test-query` | string | `"Hello"` | Test query to send |
| `--test-model` | string | - | Specific model to test |

#### `test frameworks`

Detect and test supported development frameworks in repository.

```bash
github-autofix test frameworks [flags]
```

**Flags:**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--path` | string | `.` | Repository path to analyze |
| `--deep-scan` | bool | `false` | Deep framework detection |

## Types and Interfaces

### Response Types

#### `FailureAnalysisResult`

```go
type FailureAnalysisResult struct {
    RunID         int64                 `json:"run_id"`
    Repository    string                `json:"repository"`
    FailureType   string                `json:"failure_type"`
    Confidence    float64               `json:"confidence"`
    RootCause     string                `json:"root_cause"`
    Analysis      string                `json:"analysis"`
    FixStrategies []FixStrategy         `json:"fix_strategies"`
    Context       map[string]interface{} `json:"context"`
    Timestamp     time.Time             `json:"timestamp"`
}
```

#### `AutoFixResult`

```go
type AutoFixResult struct {
    RunID           int64             `json:"run_id"`
    Success         bool              `json:"success"`
    FixBranch       string            `json:"fix_branch"`
    PullRequestURL  string            `json:"pull_request_url"`
    TestResults     *ValidationResult `json:"test_results"`
    AppliedFixes    []AppliedFix      `json:"applied_fixes"`
    Message         string            `json:"message"`
    Timestamp       time.Time         `json:"timestamp"`
}
```

#### `ValidationResult`

```go
type ValidationResult struct {
    Branch        string                 `json:"branch"`
    Success       bool                   `json:"success"`
    TestsPassed   int                    `json:"tests_passed"`
    TestsFailed   int                    `json:"tests_failed"`
    Coverage      float64                `json:"coverage"`
    Duration      time.Duration          `json:"duration"`
    Results       map[string]interface{} `json:"results"`
    Timestamp     time.Time             `json:"timestamp"`
}
```

#### `SystemStatus`

```go
type SystemStatus struct {
    Status           string                 `json:"status"` // "healthy", "degraded", "unhealthy"
    Version          string                 `json:"version"`
    GitHubAPI        ServiceStatus          `json:"github_api"`
    LLMProvider      ServiceStatus          `json:"llm_provider"`
    ActiveMonitoring bool                   `json:"active_monitoring"`
    Statistics       SystemStatistics       `json:"statistics"`
    LastCheck        time.Time              `json:"last_check"`
}
```

### Configuration Types

#### `Config`

```go
type Config struct {
    GitHub      GitHubConfig      `json:"github"`
    Repository  RepositoryConfig  `json:"repository"`
    LLM         LLMConfig         `json:"llm"`
    Testing     TestingConfig     `json:"testing"`
    Monitoring  MonitoringConfig  `json:"monitoring"`
    Logging     LoggingConfig     `json:"logging"`
}
```

## Error Codes

### HTTP Status Codes

| Code | Name | Description |
|------|------|-------------|
| 200 | Success | Operation completed successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication failed |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 429 | Rate Limited | API rate limit exceeded |
| 500 | Internal Error | Internal server error |
| 503 | Service Unavailable | External service unavailable |

### Custom Error Types

#### `ConfigurationError`

Invalid or missing configuration parameters.

#### `AuthenticationError`

GitHub or LLM provider authentication failed.

#### `AnalysisError`

Failure analysis could not be completed.

#### `FixGenerationError`

Fix generation failed or produced no results.

#### `ValidationError`

Fix validation failed during testing.

#### `GitHubAPIError`

GitHub API interaction failed.

#### `LLMProviderError`

LLM provider request failed.

## Examples

### Complete Dagger Module Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
)

func main() {
    ctx := context.Background()
    
    // Initialize agent with full configuration
    agent := dag.GithubAutofix().
        WithGitHubToken(dag.SetSecret("github-token", os.Getenv("GITHUB_TOKEN"))).
        WithLLMProvider("openai", dag.SetSecret("openai-key", os.Getenv("OPENAI_API_KEY"))).
        WithRepository("myorg", "myrepo").
        WithTargetBranch("main").
        WithMinCoverage(90)
    
    // Initialize all components
    initialized, err := agent.Initialize(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }
    
    // Test connectivity
    connectivity, err := initialized.TestConnectivity(ctx)
    if err != nil {
        log.Fatalf("Connectivity test failed: %v", err)
    }
    fmt.Printf("GitHub API: %v, LLM: %v\n", connectivity.GitHub, connectivity.LLM)
    
    // Monitor workflows for 10 minutes
    monitorCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
    defer cancel()
    
    if err := initialized.MonitorWorkflows(monitorCtx); err != nil && err != context.DeadlineExceeded {
        log.Printf("Monitoring error: %v", err)
    }
    
    fmt.Println("Monitoring completed")
}
```

### CLI Automation Script

```bash
#!/bin/bash

# Complete CLI automation example
set -e

# Configuration
export GITHUB_TOKEN="your_token"
export OPENAI_API_KEY="your_key"
export REPO_OWNER="myorg"
export REPO_NAME="myrepo"

# Test connectivity first
echo "Testing connectivity..."
./github-autofix test connection
./github-autofix test llm

# Get recent failed runs
echo "Checking for failed runs..."
FAILED_RUNS=$(./github-autofix status --format=json | jq -r '.failed_runs[].id' | head -3)

# Process each failed run
for run_id in $FAILED_RUNS; do
    echo "Processing run $run_id..."
    
    # Analyze failure
    ./github-autofix analyze $run_id --save-analysis="analysis_$run_id.json"
    
    # Generate fix (dry run first)
    ./github-autofix fix $run_id --dry-run
    
    # If dry run looks good, apply fix
    read -p "Apply fix for run $run_id? (y/N) " -n 1 -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        ./github-autofix fix $run_id --reviewer=maintainer
    fi
    
    echo "Completed processing run $run_id"
done

echo "All failed runs processed"
```
