# Configuration Guide

Comprehensive guide to configuring the Dagger GitHub Actions Auto-Fix Agent for various environments, LLM providers, and integration scenarios.

## Table of Contents

- [Quick Start Configuration](#quick-start-configuration)
- [LLM Provider Configurations](#llm-provider-configurations)
- [GitHub Integration Options](#github-integration-options)
- [Environment-Specific Configurations](#environment-specific-configurations)
- [Advanced Configuration Patterns](#advanced-configuration-patterns)
- [Security Best Practices](#security-best-practices)
- [Performance Tuning](#performance-tuning)
- [Monitoring and Logging](#monitoring-and-logging)

## Quick Start Configuration

### Minimal Configuration

Create a `.env` file with the minimum required settings:

```bash
# Minimal required configuration
GITHUB_TOKEN=ghp_your_personal_access_token
LLM_PROVIDER=openai
LLM_API_KEY=sk-your_openai_api_key
REPO_OWNER=your_username
REPO_NAME=your_repository
```

### Recommended Production Configuration

For production deployments, use this more comprehensive setup:

```bash
# === AUTHENTICATION ===
GITHUB_TOKEN=ghp_your_token_with_full_repo_access
LLM_PROVIDER=openai
LLM_API_KEY=sk-your_production_openai_key

# === REPOSITORY SETTINGS ===
REPO_OWNER=your_organization
REPO_NAME=your_production_repo
TARGET_BRANCH=main
PROTECTED_BRANCHES=main,master,release/*,hotfix/*

# === TESTING CONFIGURATION ===
MIN_COVERAGE=85
TEST_TIMEOUT=600
ENABLE_INTEGRATION_TESTS=true

# === MONITORING SETTINGS ===
MONITOR_INTERVAL=30
MAX_CONCURRENT_FIXES=2
RATE_LIMIT_BUFFER=20

# === LOGGING ===
LOG_LEVEL=info
LOG_FORMAT=json
LOG_FILE=/var/log/github-autofix.log

# === NOTIFICATIONS ===
WEBHOOK_URL=https://hooks.slack.com/services/your/webhook/url
NOTIFICATION_CHANNEL=#ci-alerts
```

## LLM Provider Configurations

### OpenAI GPT-4 (Recommended)

OpenAI provides the most reliable and consistent results for code analysis and fix generation.

#### Basic Configuration

```bash
LLM_PROVIDER=openai
LLM_API_KEY=sk-proj-your_openai_api_key
```

#### Advanced OpenAI Configuration

```bash
LLM_PROVIDER=openai
LLM_API_KEY=sk-proj-your_openai_api_key

# Model Selection (optional)
OPENAI_MODEL=gpt-4                    # gpt-4, gpt-4-turbo, gpt-3.5-turbo
OPENAI_MAX_TOKENS=4096                # Maximum tokens per request
OPENAI_TEMPERATURE=0.1                # Lower for more deterministic responses

# Organization Settings (optional)
OPENAI_ORG_ID=org-your_organization_id
OPENAI_BASE_URL=https://api.openai.com # Custom endpoint if needed

# Rate Limiting
OPENAI_REQUESTS_PER_MINUTE=60
OPENAI_TOKENS_PER_MINUTE=60000

# Cost Management  
OPENAI_MAX_COST_PER_DAY=50.00        # USD limit per day
OPENAI_COST_TRACKING=true
```

#### OpenAI Azure Configuration

```bash
LLM_PROVIDER=openai
LLM_API_KEY=your_azure_openai_key

# Azure-specific settings
OPENAI_BASE_URL=https://your-resource.openai.azure.com
OPENAI_API_VERSION=2023-05-15
OPENAI_DEPLOYMENT_NAME=your-gpt-4-deployment
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com
```

### Anthropic Claude

Claude excels at understanding context and providing detailed explanations for fixes.

#### Basic Configuration

```bash
LLM_PROVIDER=anthropic
LLM_API_KEY=sk-ant-your_anthropic_api_key
```

#### Advanced Anthropic Configuration

```bash
LLM_PROVIDER=anthropic
LLM_API_KEY=sk-ant-your_anthropic_api_key

# Model Selection
ANTHROPIC_MODEL=claude-3-sonnet-20240229  # claude-3-sonnet, claude-3-haiku, claude-2.1
ANTHROPIC_VERSION=2023-06-01              # API version

# Request Settings
ANTHROPIC_MAX_TOKENS=4096
ANTHROPIC_TEMPERATURE=0.0
ANTHROPIC_TOP_P=1.0

# Rate Limiting
ANTHROPIC_REQUESTS_PER_MINUTE=50
ANTHROPIC_TOKENS_PER_MINUTE=40000

# Cost Management
ANTHROPIC_MAX_COST_PER_DAY=30.00
```

### Google Gemini

Gemini offers strong multimodal capabilities and competitive performance.

#### Basic Configuration

```bash
LLM_PROVIDER=gemini
LLM_API_KEY=your_google_api_key
```

#### Advanced Gemini Configuration

```bash
LLM_PROVIDER=gemini
LLM_API_KEY=your_google_api_key

# Model Selection
GEMINI_MODEL=gemini-pro                   # gemini-pro, gemini-pro-vision
GEMINI_PROJECT_ID=your_gcp_project_id     # Optional: GCP project

# Request Settings
GEMINI_MAX_TOKENS=2048
GEMINI_TEMPERATURE=0.1
GEMINI_TOP_P=0.8
GEMINI_TOP_K=40

# Safety Settings
GEMINI_SAFETY_HARASSMENT=BLOCK_MEDIUM_AND_ABOVE
GEMINI_SAFETY_HATE_SPEECH=BLOCK_MEDIUM_AND_ABOVE
GEMINI_SAFETY_SEXUALLY_EXPLICIT=BLOCK_MEDIUM_AND_ABOVE
GEMINI_SAFETY_DANGEROUS_CONTENT=BLOCK_MEDIUM_AND_ABOVE

# Rate Limiting
GEMINI_REQUESTS_PER_MINUTE=60
```

### DeepSeek

DeepSeek specializes in code understanding and generation.

#### Basic Configuration

```bash
LLM_PROVIDER=deepseek
LLM_API_KEY=sk-your_deepseek_api_key
```

#### Advanced DeepSeek Configuration

```bash
LLM_PROVIDER=deepseek
LLM_API_KEY=sk-your_deepseek_api_key

# Model Selection
DEEPSEEK_MODEL=deepseek-coder             # deepseek-coder, deepseek-chat
DEEPSEEK_BASE_URL=https://api.deepseek.com # Custom endpoint

# Request Settings
DEEPSEEK_MAX_TOKENS=4096
DEEPSEEK_TEMPERATURE=0.0
DEEPSEEK_TOP_P=0.95

# Rate Limiting
DEEPSEEK_REQUESTS_PER_MINUTE=100
DEEPSEEK_MAX_COST_PER_DAY=20.00
```

### LiteLLM Proxy (Multi-Provider)

LiteLLM allows routing between multiple providers and models.

#### Basic Configuration

```bash
LLM_PROVIDER=litellm
LLM_API_KEY=your_litellm_token
LITELLM_BASE_URL=http://localhost:4000
```

#### Advanced LiteLLM Configuration

```bash
LLM_PROVIDER=litellm
LLM_API_KEY=your_litellm_token
LITELLM_BASE_URL=http://localhost:4000

# Model Selection (routed through LiteLLM)
LITELLM_MODEL=gpt-4                       # Model identifier in LiteLLM config
LITELLM_FALLBACK_MODELS=claude-3-sonnet,gemini-pro

# Request Settings
LITELLM_TIMEOUT=60
LITELLM_MAX_RETRIES=3
LITELLM_RETRY_DELAY=1.0

# Load Balancing
LITELLM_ENABLE_LOAD_BALANCING=true
LITELLM_ROUTING_STRATEGY=round_robin      # round_robin, least_busy, cost_optimized

# Caching
LITELLM_ENABLE_CACHING=true
LITELLM_CACHE_TTL=3600
```

#### LiteLLM Configuration File (`litellm-config.yaml`)

```yaml
model_list:
  - model_name: gpt-4
    litellm_params:
      model: openai/gpt-4
      api_key: sk-your_openai_key
      
  - model_name: claude-3-sonnet  
    litellm_params:
      model: anthropic/claude-3-sonnet-20240229
      api_key: sk-ant-your_anthropic_key
      
  - model_name: gemini-pro
    litellm_params:
      model: gemini/gemini-pro
      api_key: your_google_api_key

router_settings:
  routing_strategy: "simple-shuffle"
  model_group_alias:
    smart_models: ["gpt-4", "claude-3-sonnet"]
    fast_models: ["gpt-3.5-turbo", "gemini-pro"]
    
general_settings:
  master_key: your_litellm_token
  database_url: "postgresql://user:password@localhost/litellm"
  
litellm_settings:
  telemetry: false
  success_callback: ["prometheus"]
  failure_callback: ["slack"]
```

### Multi-Provider Failover Configuration

Configure automatic failover between multiple LLM providers:

```bash
# Primary provider
LLM_PROVIDER=openai
LLM_API_KEY=sk-your_openai_key

# Fallback providers (in order of preference)
LLM_FALLBACK_PROVIDERS=anthropic,gemini,deepseek
ANTHROPIC_API_KEY=sk-ant-your_anthropic_key
GEMINI_API_KEY=your_google_api_key
DEEPSEEK_API_KEY=sk-your_deepseek_key

# Failover Settings
ENABLE_AUTOMATIC_FAILOVER=true
FAILOVER_ON_RATE_LIMIT=true
FAILOVER_ON_ERROR=true
MAX_FAILOVER_ATTEMPTS=3
FAILOVER_COOLDOWN=300                     # 5 minutes cooldown
```

## GitHub Integration Options

### Personal Access Token (PAT)

Simple authentication for individual developers and small teams.

#### Required Token Scopes

```bash
GITHUB_TOKEN=ghp_your_personal_access_token

# Required scopes:
# - repo (full repository access)
# - actions:read (read Actions workflows and runs)  
# - pull_requests:write (create and manage PRs)
# - contents:write (modify repository contents)
# - issues:write (create issues for failed fixes)
```

#### Fine-Grained Personal Access Tokens

For repositories in organizations that require fine-grained permissions:

```bash
GITHUB_TOKEN=github_pat_your_fine_grained_token

# Required repository permissions:
# - Actions: Read
# - Contents: Write
# - Issues: Write  
# - Pull requests: Write
# - Metadata: Read
```

### GitHub App Authentication (Recommended)

More secure and scalable for organizations.

#### Creating a GitHub App

1. Go to Organization Settings → Developer settings → GitHub Apps
2. Create a new GitHub App with these permissions:
   - Repository permissions:
     - Actions: Read
     - Contents: Write
     - Issues: Write
     - Pull requests: Write
     - Metadata: Read
   - Subscribe to events:
     - Workflow run
     - Pull request
     - Issues

#### GitHub App Configuration

```bash
# GitHub App authentication
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY_FILE=/path/to/private-key.pem
GITHUB_INSTALLATION_ID=12345678

# Alternative: inline private key
GITHUB_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...
-----END PRIVATE KEY-----"

# App-specific settings
GITHUB_APP_SLUG=your-app-slug
GITHUB_APP_CLIENT_ID=Iv1.your_client_id
GITHUB_APP_CLIENT_SECRET=your_client_secret
```

#### Multi-Installation Configuration

For apps installed across multiple organizations:

```bash
# Multiple installation IDs for different orgs
GITHUB_INSTALLATION_MAPPING='{
  "org1": 12345678,
  "org2": 87654321,
  "org3": 13579246
}'

# Or environment-specific
GITHUB_INSTALLATION_ID_ORG1=12345678
GITHUB_INSTALLATION_ID_ORG2=87654321
```

### GitHub Enterprise Configuration

For GitHub Enterprise Server deployments:

```bash
# GitHub Enterprise settings
GITHUB_API_URL=https://github.your-company.com/api/v3
GITHUB_BASE_URL=https://github.your-company.com
GITHUB_UPLOAD_URL=https://github.your-company.com/api/uploads

# Enterprise-specific authentication
GITHUB_TOKEN=ghp_your_enterprise_token

# TLS Configuration (if needed)
GITHUB_TLS_VERIFY=true
GITHUB_CA_CERT_PATH=/path/to/ca-cert.pem
```

### Webhook Configuration

For real-time failure notifications:

```bash
# Webhook settings
GITHUB_WEBHOOK_SECRET=your_webhook_secret
WEBHOOK_ENDPOINT=https://your-server.com/webhook
WEBHOOK_EVENTS=workflow_run,pull_request

# Webhook verification
VERIFY_WEBHOOK_SIGNATURES=true
WEBHOOK_TIMEOUT=30
```

## Environment-Specific Configurations

### Development Environment

```bash
# .env.development
GITHUB_TOKEN=ghp_development_token
LLM_PROVIDER=openai
LLM_API_KEY=sk-development_key
REPO_OWNER=developer-username
REPO_NAME=test-repository

# Development-specific settings
LOG_LEVEL=debug
VERBOSE=true
DRY_RUN_BY_DEFAULT=true
MIN_COVERAGE=70
SKIP_SLOW_TESTS=true

# Mock services for testing
USE_MOCK_LLM=false
USE_MOCK_GITHUB=false
RECORD_API_RESPONSES=true
```

### Staging Environment

```bash
# .env.staging
GITHUB_TOKEN=ghp_staging_token
LLM_PROVIDER=openai
LLM_API_KEY=sk-staging_key
REPO_OWNER=staging-org
REPO_NAME=staging-repo

# Staging-specific settings
LOG_LEVEL=info
MIN_COVERAGE=80
MONITOR_INTERVAL=60
MAX_CONCURRENT_FIXES=1

# Enhanced testing
ENABLE_INTEGRATION_TESTS=true
TEST_TIMEOUT=900
PERFORMANCE_TESTING=true
```

### Production Environment

```bash
# .env.production
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY_FILE=/secrets/github-app-key.pem
GITHUB_INSTALLATION_ID=12345678
LLM_PROVIDER=openai
LLM_API_KEY=sk-production_key

# Production settings
LOG_LEVEL=warn
LOG_FORMAT=json
LOG_FILE=/var/log/github-autofix.log
STRUCTURED_LOGGING=true

# Performance optimization
MIN_COVERAGE=85
MONITOR_INTERVAL=30
MAX_CONCURRENT_FIXES=3
ENABLE_CACHING=true
CACHE_TTL=3600

# Monitoring and alerts
WEBHOOK_URL=https://hooks.slack.com/services/prod/alerts
HEALTH_CHECK_ENABLED=true
METRICS_ENABLED=true
PROMETHEUS_ENDPOINT=http://prometheus:9090

# Security
RATE_LIMITING_ENABLED=true
MAX_REQUESTS_PER_HOUR=1000
SECURITY_SCANNING=true
AUDIT_LOGGING=true
```

## Advanced Configuration Patterns

### Multi-Repository Configuration

For organizations managing multiple repositories:

```yaml
# multi-repo-config.yaml
repositories:
  - owner: myorg
    name: frontend-app
    target_branch: main
    min_coverage: 80
    llm_provider: openai
    test_frameworks: [jest, cypress]
    
  - owner: myorg
    name: backend-api
    target_branch: develop
    min_coverage: 90
    llm_provider: anthropic
    test_frameworks: [go, testify]
    
  - owner: myorg
    name: mobile-app
    target_branch: main
    min_coverage: 75
    llm_provider: gemini
    test_frameworks: [xctest, espresso]

# Default settings applied to all repositories
defaults:
  monitor_interval: 30
  max_concurrent_fixes: 2
  log_level: info
  enable_notifications: true
```

### Framework-Specific Configuration

Configure behavior for different programming languages and frameworks:

```yaml
# framework-config.yaml
frameworks:
  go:
    test_command: "go test -v -race ./..."
    build_command: "go build ./..."
    coverage_command: "go test -coverprofile=coverage.out ./..."
    min_coverage: 85
    lint_command: "golangci-lint run"
    
  javascript:
    test_command: "npm test"
    build_command: "npm run build"
    coverage_command: "npm run test:coverage"
    min_coverage: 80
    lint_command: "eslint src/"
    package_manager: "npm"  # npm, yarn, pnpm
    
  python:
    test_command: "pytest -v"
    build_command: "python -m build"
    coverage_command: "pytest --cov=."
    min_coverage: 85
    lint_command: "flake8"
    virtual_env: "venv"
    
  rust:
    test_command: "cargo test"
    build_command: "cargo build"
    coverage_command: "cargo tarpaulin --out Json"
    min_coverage: 90
    lint_command: "cargo clippy"
```

### Environment Variable Templates

Create reusable environment templates:

```bash
# template-base.env
LOG_LEVEL=info
LOG_FORMAT=json
MONITOR_INTERVAL=30
MIN_COVERAGE=85
MAX_CONCURRENT_FIXES=2

# template-openai.env
LLM_PROVIDER=openai
OPENAI_MODEL=gpt-4
OPENAI_TEMPERATURE=0.1
OPENAI_MAX_TOKENS=4096

# template-github-app.env
GITHUB_APP_ID=${GITHUB_APP_ID}
GITHUB_PRIVATE_KEY_FILE=${GITHUB_PRIVATE_KEY_FILE}
GITHUB_INSTALLATION_ID=${GITHUB_INSTALLATION_ID}
```

## Security Best Practices

### Credential Management

#### Using External Secret Stores

```bash
# AWS Secrets Manager
GITHUB_TOKEN_SECRET_ARN=arn:aws:secretsmanager:region:account:secret:github-token
LLM_API_KEY_SECRET_ARN=arn:aws:secretsmanager:region:account:secret:openai-key

# Azure Key Vault
AZURE_KEYVAULT_URL=https://your-vault.vault.azure.net/
GITHUB_TOKEN_SECRET_NAME=github-token
LLM_API_KEY_SECRET_NAME=openai-key

# HashiCorp Vault
VAULT_ADDR=https://vault.your-company.com
VAULT_TOKEN=${VAULT_TOKEN}
GITHUB_TOKEN_PATH=secret/github-autofix/github-token
LLM_API_KEY_PATH=secret/github-autofix/openai-key
```

#### Kubernetes Secrets

```yaml
# kubernetes-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-autofix-secrets
type: Opaque
data:
  github-token: <base64-encoded-token>
  openai-api-key: <base64-encoded-key>
  github-app-private-key: <base64-encoded-private-key>
```

### Network Security

```bash
# Network security settings
ALLOWED_NETWORKS=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
PROXY_URL=http://corporate-proxy:8080
TLS_CERT_PATH=/etc/ssl/certs/github-autofix.crt
TLS_KEY_PATH=/etc/ssl/private/github-autofix.key
TLS_MIN_VERSION=1.2

# Firewall rules
OUTBOUND_GITHUB_API=api.github.com:443
OUTBOUND_OPENAI_API=api.openai.com:443
OUTBOUND_ANTHROPIC_API=api.anthropic.com:443
```

### Access Control

```bash
# Role-based access control
RBAC_ENABLED=true
ADMIN_USERS=admin@company.com,ops-team@company.com
READONLY_USERS=developer@company.com,intern@company.com

# Repository access control
ALLOWED_REPOSITORIES=myorg/frontend,myorg/backend,myorg/mobile
RESTRICTED_BRANCHES=main,master,release/*
REQUIRE_APPROVAL_FOR_BRANCHES=main,master

# Rate limiting per user/role
RATE_LIMIT_ADMIN=1000
RATE_LIMIT_USER=100
RATE_LIMIT_READONLY=50
```

## Performance Tuning

### Caching Configuration

```bash
# Analysis caching
ENABLE_ANALYSIS_CACHE=true
CACHE_BACKEND=redis                       # redis, memcached, local
CACHE_TTL=3600                           # 1 hour
CACHE_KEY_PREFIX=github-autofix:
CACHE_MAX_SIZE=1GB

# Redis configuration
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=your_redis_password
REDIS_MAX_CONNECTIONS=10
REDIS_TIMEOUT=5

# LLM response caching
LLM_CACHE_ENABLED=true
LLM_CACHE_TTL=7200                       # 2 hours
LLM_CACHE_MAX_SIZE=500MB
```

### Concurrency and Parallelism

```bash
# Worker configuration
MAX_WORKERS=4
MAX_CONCURRENT_ANALYSES=3
MAX_CONCURRENT_TESTS=2
MAX_CONCURRENT_FIXES=2

# Queue configuration
QUEUE_BACKEND=redis
QUEUE_MAX_SIZE=1000
QUEUE_RETRY_ATTEMPTS=3
QUEUE_RETRY_DELAY=60

# Resource limits
MAX_MEMORY_USAGE=2GB
MAX_CPU_USAGE=80
MAX_DISK_USAGE=10GB
```

### Database Configuration

```bash
# Database settings (for persistent state)
DATABASE_URL=postgresql://user:password@localhost:5432/github_autofix
DATABASE_MAX_CONNECTIONS=20
DATABASE_IDLE_TIMEOUT=300
DATABASE_QUERY_TIMEOUT=30

# Connection pooling
DB_POOL_SIZE=10
DB_POOL_MAX_OVERFLOW=20
DB_POOL_TIMEOUT=30
DB_POOL_RECYCLE=3600
```

## Monitoring and Logging

### Structured Logging

```bash
# Logging configuration
LOG_LEVEL=info                           # trace, debug, info, warn, error
LOG_FORMAT=json                          # json, text, structured
LOG_OUTPUT=both                          # console, file, both
LOG_FILE=/var/log/github-autofix.log
LOG_MAX_SIZE=100MB
LOG_MAX_BACKUPS=10
LOG_MAX_AGE=30

# Log fields
LOG_INCLUDE_CALLER=true
LOG_INCLUDE_TIMESTAMP=true
LOG_INCLUDE_HOSTNAME=true
LOG_INCLUDE_PID=true

# Sensitive data handling
LOG_MASK_SECRETS=true
LOG_MASK_PATTERNS=sk-,ghp-,Bearer
```

### Metrics and Monitoring

```bash
# Prometheus metrics
METRICS_ENABLED=true
METRICS_PORT=9090
METRICS_PATH=/metrics
METRICS_NAMESPACE=github_autofix

# Custom metrics
TRACK_ANALYSIS_DURATION=true
TRACK_FIX_SUCCESS_RATE=true
TRACK_TEST_COVERAGE=true
TRACK_API_USAGE=true

# Health checks
HEALTH_CHECK_ENABLED=true
HEALTH_CHECK_PORT=8080
HEALTH_CHECK_PATH=/health
HEALTH_CHECK_INTERVAL=30
```

### Alerting Configuration

```bash
# Slack integration
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/your/webhook/url
SLACK_CHANNEL=#ci-alerts
SLACK_USERNAME=GitHub Auto-Fix Bot
SLACK_ICON_EMOJI=:robot_face:

# Alert thresholds
ALERT_ON_FAILURE_RATE=10                 # Alert if failure rate > 10%
ALERT_ON_RESPONSE_TIME=30                # Alert if response time > 30s
ALERT_ON_ERROR_COUNT=5                   # Alert if errors > 5 in 5 minutes
ALERT_ON_RATE_LIMIT=80                   # Alert if API usage > 80%

# PagerDuty integration
PAGERDUTY_INTEGRATION_KEY=your_pagerduty_key
PAGERDUTY_SEVERITY=warning               # info, warning, error, critical
```

This comprehensive configuration guide covers all aspects of setting up and optimizing the Dagger GitHub Actions Auto-Fix Agent for various environments and use cases. Choose the configuration options that best fit your specific requirements and gradually add more advanced features as needed.
