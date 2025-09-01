# Troubleshooting Guide

Comprehensive troubleshooting guide for the Dagger GitHub Actions Auto-Fix Agent, covering common issues, debugging techniques, and performance optimization strategies.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Authentication Issues](#authentication-issues)
- [LLM Provider Issues](#llm-provider-issues)
- [Analysis and Fix Generation Issues](#analysis-and-fix-generation-issues)
- [Testing and Validation Issues](#testing-and-validation-issues)
- [Pull Request Creation Issues](#pull-request-creation-issues)
- [Performance Issues](#performance-issues)
- [Network and Connectivity Issues](#network-and-connectivity-issues)
- [Configuration Issues](#configuration-issues)
- [Debugging Techniques](#debugging-techniques)
- [Error Codes Reference](#error-codes-reference)
- [Performance Optimization](#performance-optimization)
- [Monitoring and Logging](#monitoring-and-logging)
- [Getting Help](#getting-help)

## Quick Diagnostics

### Health Check Command

Run this comprehensive health check to identify most common issues:

```bash
# Quick system health check
./github-autofix status --include-diagnostics

# Detailed connectivity test
./github-autofix test connection --verbose

# Configuration validation
./github-autofix config validate --strict

# Component status check
./github-autofix debug-info --format=json > diagnostic_report.json
```

### Common Quick Fixes

```bash
# Fix 1: Clear cache and reset state
rm -rf ~/.cache/github-autofix/
export CLEAR_ANALYSIS_CACHE=true

# Fix 2: Reset credentials
unset GITHUB_TOKEN LLM_API_KEY
source .env

# Fix 3: Update to latest version
dagger call update-agent

# Fix 4: Rebuild CLI
dagger call cli --export ./github-autofix --force-rebuild
```

## Authentication Issues

### Problem: `Authentication failed` or `Invalid credentials`

#### GitHub Token Issues

**Symptoms:**
- `HTTP 401: Bad credentials`
- `GitHub API authentication failed`
- `Resource not accessible by token`

**Diagnosis:**
```bash
# Test token validity
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Check token scopes
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user -I | grep x-oauth-scopes

# Test repository access
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/$REPO_OWNER/$REPO_NAME
```

**Solutions:**

1. **Verify Token Scopes:**
   ```bash
   # Required scopes for Personal Access Token:
   # - repo (full repository access)
   # - actions:read (read Actions workflows)
   # - pull_requests:write (create PRs)
   # - contents:write (modify files)
   ```

2. **Regenerate Token:**
   ```bash
   # Go to GitHub Settings → Developer settings → Personal access tokens
   # Generate new token with correct scopes
   export GITHUB_TOKEN=ghp_new_token_here
   ```

3. **Use GitHub App (Recommended):**
   ```bash
   # Convert to GitHub App authentication
   export GITHUB_APP_ID=123456
   export GITHUB_PRIVATE_KEY_FILE=/path/to/private-key.pem
   export GITHUB_INSTALLATION_ID=12345678
   unset GITHUB_TOKEN
   ```

### Problem: `Rate limit exceeded`

**Symptoms:**
- `HTTP 403: API rate limit exceeded`
- `X-RateLimit-Remaining: 0`
- Requests failing after period of activity

**Diagnosis:**
```bash
# Check current rate limit status
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/rate_limit

# Monitor rate limit headers
./github-autofix debug rate-limits --continuous
```

**Solutions:**

1. **Switch to GitHub App:**
   ```bash
   # GitHub Apps have much higher rate limits (5000/hour vs 1000/hour)
   export GITHUB_APP_ID=your_app_id
   export GITHUB_PRIVATE_KEY="$(cat private-key.pem)"
   export GITHUB_INSTALLATION_ID=your_installation_id
   ```

2. **Implement Rate Limiting:**
   ```bash
   export RATE_LIMIT_ENABLED=true
   export REQUESTS_PER_HOUR=800  # Stay under limit
   export RATE_LIMIT_BUFFER=20   # 20% buffer
   ```

3. **Use Request Caching:**
   ```bash
   export ENABLE_REQUEST_CACHE=true
   export CACHE_TTL=300          # 5 minute cache
   ```

## LLM Provider Issues

### Problem: `LLM request failed` or `Provider timeout`

#### OpenAI Issues

**Symptoms:**
- `HTTP 401: Invalid API key`
- `HTTP 429: Rate limit reached`
- `Connection timeout`
- `Model not found`

**Diagnosis:**
```bash
# Test OpenAI API key
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
  https://api.openai.com/v1/models

# Check specific model access
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
  https://api.openai.com/v1/models/gpt-4

# Test with simple request
./github-autofix test llm --llm-provider=openai --verbose
```

**Solutions:**

1. **Verify API Key:**
   ```bash
   # Check API key format (should start with sk-proj- for new keys)
   echo $OPENAI_API_KEY | grep -E '^sk-(proj-)?[a-zA-Z0-9]{48,}$'
   ```

2. **Check Model Access:**
   ```bash
   # Use accessible model
   export OPENAI_MODEL=gpt-3.5-turbo  # More accessible than gpt-4
   export OPENAI_MODEL=gpt-4          # Requires higher tier access
   ```

3. **Configure Timeouts:**
   ```bash
   export OPENAI_TIMEOUT=120          # 2 minute timeout
   export OPENAI_MAX_RETRIES=3
   export OPENAI_RETRY_DELAY=5
   ```

#### Anthropic Issues

**Symptoms:**
- `HTTP 401: Unauthorized`
- `HTTP 400: Invalid request`
- `Model overloaded`

**Diagnosis:**
```bash
# Test Anthropic API
curl -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  https://api.anthropic.com/v1/messages \
  -d '{"model":"claude-3-haiku-20240307","max_tokens":10,"messages":[{"role":"user","content":"Hi"}]}'
```

**Solutions:**

1. **Use Correct Model Names:**
   ```bash
   export ANTHROPIC_MODEL=claude-3-sonnet-20240229  # Correct format
   export ANTHROPIC_VERSION=2023-06-01              # Required API version
   ```

2. **Handle Rate Limits:**
   ```bash
   export ANTHROPIC_REQUESTS_PER_MINUTE=50
   export ANTHROPIC_TOKENS_PER_MINUTE=40000
   ```

#### Google Gemini Issues

**Symptoms:**
- `HTTP 400: API key not valid`
- `Permission denied`
- `Quota exceeded`

**Solutions:**

1. **Enable APIs:**
   ```bash
   # Ensure Generative AI API is enabled in Google Cloud Console
   # Set correct project ID
   export GEMINI_PROJECT_ID=your-gcp-project-id
   ```

2. **Authentication:**
   ```bash
   # Use service account key
   export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
   
   # Or use API key
   export GEMINI_API_KEY=your_google_api_key
   ```

### Problem: `No LLM providers available`

**Diagnosis:**
```bash
# Check provider configuration
./github-autofix config show | grep -i llm

# Test each provider individually
./github-autofix test llm --llm-provider=openai
./github-autofix test llm --llm-provider=anthropic
./github-autofix test llm --llm-provider=gemini
```

**Solution:**
```bash
# Configure fallback providers
export LLM_FALLBACK_PROVIDERS=anthropic,gemini,deepseek
export ENABLE_AUTOMATIC_FAILOVER=true
export FAILOVER_ON_RATE_LIMIT=true
```

## Analysis and Fix Generation Issues

### Problem: `Analysis failed` or `No fixes generated`

**Symptoms:**
- `Failed to analyze workflow failure`
- `No actionable fixes found`
- `Analysis timeout exceeded`
- `Insufficient context for analysis`

**Diagnosis:**
```bash
# Enable detailed analysis logging
export LOG_LEVEL=debug
export ANALYSIS_DEBUG=true
./github-autofix analyze 1234567890 --verbose --save-analysis=debug_analysis.json

# Check analysis cache
./github-autofix cache status
./github-autofix cache clear  # If corrupted
```

**Solutions:**

1. **Increase Context:**
   ```bash
   export INCLUDE_FULL_LOGS=true
   export MAX_LOG_SIZE=50MB
   export INCLUDE_REPOSITORY_CONTEXT=true
   export CONTEXT_WINDOW_SIZE=8192
   ```

2. **Enable Advanced Analysis:**
   ```bash
   export ENABLE_DEEP_ANALYSIS=true
   export ANALYSIS_TIMEOUT=600        # 10 minutes
   export MAX_ANALYSIS_ITERATIONS=5
   export ENABLE_MULTI_STAGE_ANALYSIS=true
   ```

3. **Framework-Specific Issues:**
   ```bash
   # Ensure framework detection is working
   ./github-autofix detect-frameworks --path=. --verbose
   
   # Force framework type
   export FORCE_FRAMEWORK_TYPE=nodejs  # nodejs, python, go, rust, etc.
   ```

### Problem: `Unsupported failure type`

**Symptoms:**
- `Unknown failure pattern`
- `Unsupported programming language`
- `Framework not recognized`

**Solutions:**

1. **Add Custom Patterns:**
   ```bash
   # Create custom failure patterns file
   cat > custom-patterns.yaml << EOF
   patterns:
     - pattern: "your custom error pattern"
       category: "custom_category"
       confidence: 0.8
       fix_strategy: "custom_fix_strategy"
   EOF
   
   export CUSTOM_PATTERNS_FILE=custom-patterns.yaml
   ```

2. **Framework Support:**
   ```bash
   # Check supported frameworks
   ./github-autofix list-supported-frameworks
   
   # Request framework support
   ./github-autofix request-framework-support --language=kotlin --framework=spring-boot
   ```

## Testing and Validation Issues

### Problem: `Tests failed to run` or `Coverage below threshold`

**Symptoms:**
- `Test execution timeout`
- `Coverage: 65% (below threshold: 85%)`
- `No tests found`
- `Test framework not detected`

**Diagnosis:**
```bash
# Debug test detection
./github-autofix detect-test-framework --verbose

# Test runner diagnostics
export TEST_DEBUG=true
./github-autofix validate main --verbose

# Check test environment
./github-autofix debug test-environment
```

**Solutions:**

1. **Test Framework Detection:**
   ```bash
   # Force test framework
   export FORCE_TEST_FRAMEWORK=jest  # jest, pytest, go-test, cargo, etc.
   export TEST_COMMAND="npm test -- --coverage"
   export BUILD_COMMAND="npm run build"
   ```

2. **Environment Setup:**
   ```bash
   # Ensure dependencies are installed
   export PRE_TEST_COMMANDS="npm ci,pip install -r requirements.txt"
   export TEST_ENV_SETUP="export NODE_ENV=test"
   ```

3. **Coverage Adjustments:**
   ```bash
   # Temporarily lower threshold
   export MIN_COVERAGE=70
   export COVERAGE_TREND_REQUIRED=false  # Don't require improvement
   
   # Exclude files from coverage
   export COVERAGE_EXCLUDE_PATTERNS="test/**,*.test.js,__mocks__/**"
   ```

4. **Timeout Adjustments:**
   ```bash
   export TEST_TIMEOUT=1200           # 20 minutes
   export PARALLEL_TESTS=false        # Disable if causing issues
   export TEST_RETRIES=2              # Retry flaky tests
   ```

### Problem: `Integration tests failing`

**Solutions:**
```bash
# Skip integration tests during fix validation
export SKIP_INTEGRATION_TESTS=true
export FAST_VALIDATION_MODE=true

# Or configure integration test environment
export INTEGRATION_TEST_DATABASE_URL=sqlite:///tmp/test.db
export INTEGRATION_TEST_TIMEOUT=1800  # 30 minutes
```

## Pull Request Creation Issues

### Problem: `Failed to create PR` or `Merge conflicts`

**Symptoms:**
- `Branch already exists`
- `Merge conflicts detected`
- `PR creation failed`
- `Base branch protection rules violated`

**Diagnosis:**
```bash
# Check branch status
git branch -r | grep autofix

# Check PR status
./github-autofix status --include-prs

# Test PR creation (dry run)
./github-autofix fix 1234567890 --dry-run --verbose
```

**Solutions:**

1. **Branch Management:**
   ```bash
   # Custom branch naming
   export PR_BRANCH_PREFIX="autofix-v2"
   export PR_BRANCH_TEMPLATE="{{.prefix}}/{{.timestamp}}/{{.run_id}}"
   
   # Clean up old branches
   ./github-autofix cleanup-branches --older-than=7d
   ```

2. **Merge Conflict Resolution:**
   ```bash
   # Enable automatic conflict resolution
   export AUTO_RESOLVE_CONFLICTS=true
   export CONFLICT_RESOLUTION_STRATEGY=ours  # ours, theirs, merge
   
   # Update base branch before creating fix
   export UPDATE_BASE_BRANCH=true
   ```

3. **Branch Protection:**
   ```bash
   # Skip branch protection for autofix PRs
   export BYPASS_BRANCH_PROTECTION=true
   export AUTOFIX_PR_LABEL="autofix"  # Add label to bypass rules
   ```

4. **PR Configuration:**
   ```bash
   # Custom PR templates
   export PR_TEMPLATE_FILE=".github/AUTOFIX_PR_TEMPLATE.md"
   export PR_AUTO_ASSIGN_REVIEWERS=true
   export PR_REVIEWERS="maintainer1,maintainer2"
   ```

### Problem: `PR created but checks failing`

**Solutions:**
```bash
# Wait for checks before completing
export WAIT_FOR_CHECKS=true
export CHECK_TIMEOUT=1800  # 30 minutes

# Auto-close if checks fail
export AUTO_CLOSE_FAILED_PRS=true
export RERUN_FAILED_CHECKS=true
export MAX_CHECK_RETRIES=2
```

## Performance Issues

### Problem: `Slow analysis` or `High memory usage`

**Symptoms:**
- Analysis taking > 5 minutes
- Memory usage > 2GB
- High CPU usage
- Frequent timeouts

**Diagnosis:**
```bash
# Performance profiling
export ENABLE_PROFILING=true
export PROFILE_OUTPUT=./profiles/
./github-autofix analyze 1234567890 --profile

# Resource monitoring
./github-autofix monitor-resources --duration=300

# Cache statistics
./github-autofix cache stats
```

**Solutions:**

1. **Memory Optimization:**
   ```bash
   export MEMORY_LIMIT=1GB
   export GC_PERCENT=50              # More aggressive GC
   export MAX_FILE_SIZE_IN_MEMORY=5MB
   export STREAM_LARGE_FILES=true
   ```

2. **Analysis Optimization:**
   ```bash
   export ENABLE_ANALYSIS_CACHE=true
   export CACHE_TTL=3600             # 1 hour
   export ANALYSIS_WORKERS=2         # Reduce parallelism
   export MAX_CONTEXT_SIZE=4096      # Smaller context windows
   ```

3. **Network Optimization:**
   ```bash
   export HTTP_POOL_SIZE=5           # Smaller connection pool
   export REQUEST_TIMEOUT=30         # Shorter timeouts
   export ENABLE_COMPRESSION=true
   ```

### Problem: `Analysis cache issues`

**Solutions:**
```bash
# Clear corrupted cache
./github-autofix cache clear

# Rebuild cache with current patterns
./github-autofix cache rebuild

# Configure cache backend
export CACHE_BACKEND=redis         # redis, memcached, local
export REDIS_URL=redis://localhost:6379/0
```

## Network and Connectivity Issues

### Problem: `Connection timeouts` or `DNS resolution failures`

**Diagnosis:**
```bash
# Test network connectivity
curl -I https://api.github.com
curl -I https://api.openai.com
nslookup api.github.com

# Test from container
dagger call test-connectivity
```

**Solutions:**

1. **Proxy Configuration:**
   ```bash
   export HTTP_PROXY=http://proxy:8080
   export HTTPS_PROXY=http://proxy:8080
   export NO_PROXY=localhost,127.0.0.1,.local
   ```

2. **DNS Configuration:**
   ```bash
   # Use alternative DNS
   export DNS_SERVERS=8.8.8.8,1.1.1.1
   
   # Or in Docker
   docker run --dns=8.8.8.8 your-autofix-image
   ```

3. **Firewall Rules:**
   ```bash
   # Required outbound connections:
   # - api.github.com:443
   # - api.openai.com:443
   # - api.anthropic.com:443
   # - generativelanguage.googleapis.com:443
   ```

## Configuration Issues

### Problem: `Invalid configuration` or `Missing required values`

**Diagnosis:**
```bash
# Comprehensive config validation
./github-autofix config validate --strict --verbose

# Show effective configuration
./github-autofix config show --show-sources

# Check environment variables
env | grep -E "(GITHUB|LLM|OPENAI|ANTHROPIC)" | sort
```

**Solutions:**

1. **Configuration File Issues:**
   ```bash
   # Generate fresh config
   ./github-autofix config init --overwrite

   # Validate YAML/JSON syntax
   yamllint .github-autofix.yml
   jq . .github-autofix.json
   ```

2. **Environment Variable Issues:**
   ```bash
   # Check for typos in variable names
   env | grep -i github
   env | grep -i openai

   # Source configuration file
   set -a && source .env && set +a
   ```

## Debugging Techniques

### Enable Comprehensive Debug Mode

```bash
# Maximum debug information
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
export DEBUG_PR_ENGINE=true

# Save debug session
./github-autofix analyze 1234567890 --verbose 2>&1 | tee debug.log
```

### Structured Debug Information

```bash
# Generate comprehensive diagnostic report
./github-autofix debug-info --include-all > diagnostic_report.json

# System health check with detailed output
./github-autofix health-check --format=json --verbose

# Performance profiling
export ENABLE_PROFILING=true
export PROFILE_OUTPUT=./profiles/
./github-autofix monitor --duration=300  # Profile 5 minutes of operation
```

### Interactive Debugging

```bash
# Run in interactive debug mode
./github-autofix debug-shell

# Step-by-step analysis
./github-autofix analyze 1234567890 --step-by-step --interactive

# Mock mode for safe testing
export USE_MOCK_GITHUB=true
export USE_MOCK_LLM=true
export RECORD_INTERACTIONS=true
```

## Error Codes Reference

### HTTP Error Codes

| Code | Description | Common Causes | Solutions |
|------|-------------|---------------|-----------|
| 401 | Unauthorized | Invalid token, expired credentials | Regenerate token, check scopes |
| 403 | Forbidden | Insufficient permissions, rate limited | Add permissions, implement rate limiting |
| 404 | Not Found | Repository not found, run doesn't exist | Verify repository access, check run ID |
| 422 | Unprocessable Entity | Invalid request format | Check request parameters |
| 429 | Rate Limited | Too many requests | Implement backoff, use GitHub App |
| 500 | Internal Error | Server-side issue | Retry request, contact support |

### Application Error Codes

| Code | Description | Category | Solutions |
|------|-------------|----------|-----------|
| AUTH_001 | GitHub authentication failed | Authentication | Check token validity |
| AUTH_002 | LLM provider authentication failed | Authentication | Verify API key |
| ANAL_001 | Analysis timeout | Analysis | Increase timeout, reduce context size |
| ANAL_002 | Unsupported failure type | Analysis | Add custom patterns |
| TEST_001 | Test execution failed | Testing | Check test environment |
| TEST_002 | Coverage below threshold | Testing | Adjust threshold or improve tests |
| PR_001 | Branch creation failed | Pull Request | Check branch protection rules |
| PR_002 | Merge conflicts detected | Pull Request | Enable auto-resolution |

## Performance Optimization

### Analysis Performance

```bash
# Cache optimization
export ENABLE_ANALYSIS_CACHE=true
export CACHE_TTL=7200                # 2 hours
export CACHE_WARM_UP=true
export CACHE_COMPRESSION=true

# LLM optimization
export LLM_BATCH_SIZE=5              # Process multiple requests together
export LLM_CACHE_RESPONSES=true      # Cache LLM responses
export LLM_COMPRESS_CONTEXT=true     # Compress large context
export LLM_CONTEXT_OPTIMIZATION=true # Smart context pruning

# Parallel processing
export MAX_PARALLEL_ANALYSIS=3
export ANALYSIS_WORKERS=4
export ASYNC_PROCESSING=true
```

### Memory Management

```bash
# Memory limits
export MEMORY_LIMIT=2GB
export HEAP_SIZE=1GB
export GC_PERCENT=75                 # Trigger GC at 75% memory usage

# Streaming for large files
export STREAM_PROCESSING=true
export MAX_FILE_SIZE_IN_MEMORY=10MB
export BUFFER_SIZE=64KB
```

### Network Performance

```bash
# Connection optimization
export HTTP_POOL_SIZE=20
export KEEP_ALIVE_TIMEOUT=30
export CONNECTION_TIMEOUT=10
export READ_TIMEOUT=30

# Request optimization
export ENABLE_GZIP=true
export BATCH_REQUESTS=true
export REQUEST_DEDUPLICATION=true
export RETRY_ATTEMPTS=3
export EXPONENTIAL_BACKOFF=true
```

## Monitoring and Logging

### Structured Logging Setup

```bash
# Comprehensive logging
export LOG_LEVEL=info
export LOG_FORMAT=json
export LOG_OUTPUT=both               # console and file
export LOG_FILE=/var/log/github-autofix.log
export LOG_ROTATION=daily
export LOG_MAX_SIZE=100MB
export LOG_MAX_BACKUPS=30

# Audit logging
export AUDIT_LOGGING=true
export AUDIT_LOG_FILE=/var/log/github-autofix-audit.log
export SECURITY_EVENTS=true
```

### Metrics Collection

```bash
# Prometheus metrics
export METRICS_ENABLED=true
export METRICS_PORT=9090
export METRICS_PATH=/metrics

# Custom metrics
export TRACK_ANALYSIS_DURATION=true
export TRACK_SUCCESS_RATE=true
export TRACK_COST_METRICS=true
export TRACK_API_USAGE=true
```

### Health Monitoring

```bash
# Health checks
export HEALTH_CHECK_ENABLED=true
export HEALTH_CHECK_INTERVAL=30
export HEALTH_CHECK_TIMEOUT=10
export DEEP_HEALTH_CHECKS=true

# Alerting
export ALERT_ON_FAILURE_RATE=15      # Alert if failure rate > 15%
export ALERT_ON_HIGH_LATENCY=60      # Alert if latency > 60s
export ALERT_ON_ERROR_SPIKE=10       # Alert if errors > 10 in 5 min
```

## Getting Help

### Information to Collect

When reporting issues, please gather this information:

1. **Environment Information:**
   ```bash
   ./github-autofix version --full
   ./github-autofix debug-info --format=json > debug_info.json
   ```

2. **Configuration:**
   ```bash
   ./github-autofix config show --mask-secrets > config.txt
   ```

3. **Logs:**
   ```bash
   # Last 1000 lines of logs with debug enabled
   export LOG_LEVEL=debug
   ./github-autofix [failing-command] --verbose 2>&1 | tail -1000 > issue_logs.txt
   ```

4. **System Information:**
   ```bash
   # System details
   uname -a > system_info.txt
   docker version >> system_info.txt
   dagger version >> system_info.txt
   ```

### Support Channels

- 📝 **Documentation**: [Complete Documentation](docs/)
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/your-org/dagger-autofix/issues)
- 💬 **Community Discussion**: [GitHub Discussions](https://github.com/your-org/dagger-autofix/discussions)
- 📧 **Direct Support**: support@your-org.com
- 💬 **Community Chat**: [Discord/Slack Channel]

### Issue Templates

Use these templates when reporting issues:

**Bug Report Template:**
```markdown
**Environment:**
- OS: [e.g. Ubuntu 22.04]
- Docker version: [e.g. 24.0.7]
- Dagger version: [e.g. 0.9.5]
- Agent version: [e.g. 1.0.0]

**Configuration:**
- LLM Provider: [e.g. openai]
- Repository type: [e.g. Node.js/React]
- Deployment: [e.g. GitHub Actions]

**Issue Description:**
[Clear description of the problem]

**Expected Behavior:**
[What should happen]

**Actual Behavior:**
[What actually happens]

**Reproduction Steps:**
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Logs:**
```
[Paste relevant logs here]
```

**Additional Context:**
[Any additional information]
```

This comprehensive troubleshooting guide should help users diagnose and resolve most common issues they encounter with the GitHub Actions Auto-Fix Agent.
