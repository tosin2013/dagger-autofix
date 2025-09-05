# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Dagger.io agent for automatically resolving GitHub Actions pipeline failures with multi-LLM support. The project:
- Monitors GitHub Actions workflows for failures
- Uses LLM providers (OpenAI, Anthropic, Google Gemini, DeepSeek, LiteLLM) to analyze failures
- Generates and validates fixes automatically
- Creates pull requests with the fixes

## Build and Development Commands

### Core Commands
```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Build the CLI (using standard Go build)
go build -o github-autofix .
chmod +x ./github-autofix

# Test CLI functionality
./github-autofix --help

# Test connectivity
./github-autofix test connection

# Monitor workflows (main functionality)
./github-autofix monitor --repo-owner <owner> --repo-name <repo>
```

### Build Status
✅ All compilation errors have been resolved!

The following issues were fixed:
1. ✅ `dag` variable redeclared - removed duplicate declaration
2. ✅ Type mismatches - fixed SecurityFix enum and TestStats boolean logic
3. ✅ GitHub API usage - corrected method names and parameters

Run `go build .` to verify successful compilation.

## Architecture Overview

### Core Components
- **main.go**: Dagger module with GitHubAutofix struct, orchestrates all operations
- **cli.go**: Command-line interface implementation using Cobra  
- **mcp_client.go**: MCP (Model Context Protocol) integration for GitHub and other services
- **failure_analysis.go**: AI-powered failure diagnosis engine
- **llm_client.go**: Multi-provider LLM integration (OpenAI, Anthropic, etc.)
- **test_engine.go**: Automated testing and validation framework
- **pull_request_engine.go**: PR creation and management
- **types.go**: GitHub API interactions and data structures
- **dag.go**: Dagger client initialization

### MCP Integration
The project now supports **Model Context Protocol (MCP)** for enhanced GitHub integration:
- **Universal Integration**: Support any MCP server, not just GitHub
- **Modular Architecture**: Clean separation between core logic and external integrations
- **Future-Proof**: Easy to add new MCP servers (GitLab, Jira, Slack, etc.)
- **Tool-Based Operations**: Leverages MCP's standardized tool calling paradigm

### Key Patterns
- Uses Dagger for containerization and CI/CD operations
- Implements provider-agnostic LLM interface for multiple AI providers
- GitHub API v45 for repository operations
- Structured logging with logrus
- Environment-based configuration with godotenv

### Testing Requirements
- Minimum 85% test coverage enforced in CI
- Tests use testify for assertions
- Integration tests with Dagger CLI
- GitHub Actions workflows in `.github/workflows/` for CI/CD

## Environment Configuration

### Traditional GitHub API Integration
```bash
GITHUB_TOKEN=<github_personal_access_token>
LLM_PROVIDER=openai  # or anthropic, gemini, deepseek, litellm
LLM_API_KEY=<llm_provider_api_key>
REPO_OWNER=<github_username_or_org>
REPO_NAME=<repository_name>
```

### MCP Integration (Recommended)
```bash
# GitHub token (used by MCP server)
GITHUB_TOKEN=<github_personal_access_token>

# LLM Configuration
LLM_PROVIDER=openai  # or anthropic, gemini, deepseek, litellm
LLM_API_KEY=<llm_provider_api_key>

# Repository Configuration
REPO_OWNER=<github_username_or_org>
REPO_NAME=<repository_name>

# MCP Configuration (see examples/mcp_config.json)
MCP_ENABLED=true
```

### Optional Configuration
```bash
MIN_COVERAGE=85
TEST_TIMEOUT=600
MONITOR_INTERVAL=30
```

### MCP Setup
1. Install GitHub MCP server: `npm install -g @github/github-mcp-server`
2. Configure MCP settings in `examples/mcp_config.json`
3. Enable MCP mode: `MCP_ENABLED=true`

## Module Information
- Go module: `github.com/tosin2013/dagger-autofix`
- Go version: 1.21+
- Dagger SDK: v0.11.0
- Main dependencies: dagger.io/dagger, github.com/google/go-github/v45, cobra, logrus