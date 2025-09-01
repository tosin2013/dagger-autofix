# Enhanced Examples and Integration Scenarios

Comprehensive collection of real-world integration examples, usage scenarios, and deployment patterns for the GitHub Actions Auto-Fix Agent.

## Table of Contents

- [Basic Integration Examples](#basic-integration-examples)
- [Advanced Usage Scenarios](#advanced-usage-scenarios)
- [Enterprise Deployment Patterns](#enterprise-deployment-patterns)
- [Multi-Repository Management](#multi-repository-management)
- [Custom Workflow Integrations](#custom-workflow-integrations)
- [Monitoring and Alerting](#monitoring-and-alerting)
- [Development Workflows](#development-workflows)

## File Structure

```
examples/
├── README.md                           # This file
├── usage_scenarios.sh                  # Comprehensive demo script
├── basic/                              # Basic usage examples
│   ├── single-repo-monitoring.yml
│   ├── manual-fix-workflow.yml
│   └── simple-config.env
├── advanced/                           # Advanced scenarios
│   ├── multi-llm-fallback.yml
│   ├── custom-failure-patterns.yml
│   ├── performance-optimized.yml
│   └── security-hardened.yml
├── enterprise/                         # Enterprise deployment
│   ├── organization-wide.yml
│   ├── kubernetes-deployment.yml
│   ├── terraform-infrastructure/
│   └── monitoring-stack/
├── integrations/                       # Third-party integrations
│   ├── slack-notifications.yml
│   ├── jira-integration.yml
│   ├── prometheus-metrics.yml
│   └── webhook-server/
├── languages/                          # Language-specific examples
│   ├── golang-project.yml
│   ├── nodejs-project.yml
│   ├── python-project.yml
│   ├── rust-project.yml
│   └── multi-language.yml
├── deployments/                        # Deployment scenarios
│   ├── docker-compose.yml
│   ├── kubernetes/
│   ├── aws-lambda/
│   └── azure-functions/
└── testing/                           # Testing scenarios
    ├── integration-tests.sh
    ├── load-testing.yml
    └── mock-scenarios/
```

## Basic Integration Examples

### Single Repository Monitoring

The simplest setup for monitoring a single repository:

**File: `examples/basic/single-repo-monitoring.yml`**

```yaml
# Basic GitHub Actions workflow for single repository monitoring
name: Auto-Fix CI Failures

on:
  workflow_run:
    workflows: ["CI", "Test", "Build"]
    types: [completed]
  schedule:
    - cron: '0 */4 * * *'  # Check every 4 hours
  workflow_dispatch:

jobs:
  auto-fix:
    if: github.event.workflow_run.conclusion == 'failure' || github.event_name == 'schedule'
    runs-on: ubuntu-latest
    timeout-minutes: 15
    
    permissions:
      contents: write
      pull-requests: write
      actions: read
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        
      - name: Setup Dagger
        uses: dagger/dagger-for-github@v6
        
      - name: Run Auto-Fix Agent
        run: |
          dagger call \
            github-autofix \
            --github-token env:GITHUB_TOKEN \
            --llm-provider openai \
            --llm-api-key env:OPENAI_API_KEY \
            --repo-owner ${{ github.repository_owner }} \
            --repo-name ${{ github.event.repository.name }} \
            --target-branch main \
            --min-coverage 80 \
            monitor --duration 300
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          
      - name: Report Status
        if: always()
        run: |
          echo "Auto-fix agent completed for ${{ github.repository }}"
          echo "Check for any new pull requests with fixes"
```

**File: `examples/basic/simple-config.env`**

```bash
# Simple configuration for getting started
GITHUB_TOKEN=ghp_your_github_token
LLM_PROVIDER=openai
LLM_API_KEY=sk-your_openai_key
REPO_OWNER=your_username
REPO_NAME=your_repository

# Basic settings
TARGET_BRANCH=main
MIN_COVERAGE=75
LOG_LEVEL=info
VERBOSE=false

# Monitoring
MONITOR_INTERVAL=60
MAX_CONCURRENT_FIXES=1
```

### Manual Fix Workflow

For teams that prefer manual triggering:

**File: `examples/basic/manual-fix-workflow.yml`**

```yaml
name: Manual CI Fix

on:
  workflow_dispatch:
    inputs:
      run_id:
        description: 'Workflow run ID to fix'
        required: true
        type: string
      llm_provider:
        description: 'LLM Provider'
        required: false
        default: 'openai'
        type: choice
        options:
          - openai
          - anthropic
          - gemini
          - deepseek
      dry_run:
        description: 'Dry run (analyze only)'
        required: false
        default: false
        type: boolean

jobs:
  manual-fix:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        
      - name: Setup Dagger
        uses: dagger/dagger-for-github@v6
        
      - name: Analyze Failure
        id: analyze
        run: |
          dagger call \
            github-autofix \
            --github-token env:GITHUB_TOKEN \
            --llm-provider ${{ inputs.llm_provider }} \
            --llm-api-key env:LLM_API_KEY \
            --repo-owner ${{ github.repository_owner }} \
            --repo-name ${{ github.event.repository.name }} \
            analyze-failure --run-id ${{ inputs.run_id }} > analysis.json
          
          echo "analysis_result=$(cat analysis.json)" >> $GITHUB_OUTPUT
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          LLM_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          
      - name: Generate Fix
        if: ${{ !inputs.dry_run }}
        run: |
          dagger call \
            github-autofix \
            --github-token env:GITHUB_TOKEN \
            --llm-provider ${{ inputs.llm_provider }} \
            --llm-api-key env:LLM_API_KEY \
            --repo-owner ${{ github.repository_owner }} \
            --repo-name ${{ github.event.repository.name }} \
            auto-fix --run-id ${{ inputs.run_id }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          LLM_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          
      - name: Summary
        run: |
          echo "## Fix Analysis Results" >> $GITHUB_STEP_SUMMARY
          echo "- **Run ID**: ${{ inputs.run_id }}" >> $GITHUB_STEP_SUMMARY
          echo "- **LLM Provider**: ${{ inputs.llm_provider }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Dry Run**: ${{ inputs.dry_run }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Repository**: ${{ github.repository }}" >> $GITHUB_STEP_SUMMARY
```

## Advanced Usage Scenarios

### Multi-LLM Fallback Configuration

**File: `examples/advanced/multi-llm-fallback.yml`**

```yaml
name: Advanced Auto-Fix with LLM Fallback

on:
  workflow_run:
    workflows: ["CI"]
    types: [completed]

jobs:
  auto-fix-with-fallback:
    if: github.event.workflow_run.conclusion == 'failure'
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        llm_config:
          - provider: openai
            model: gpt-4
            priority: 1
          - provider: anthropic  
            model: claude-3-sonnet-20240229
            priority: 2
          - provider: gemini
            model: gemini-pro
            priority: 3
      fail-fast: false
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        
      - name: Setup Dagger
        uses: dagger/dagger-for-github@v6
        
      - name: Attempt Fix with ${{ matrix.llm_config.provider }}
        id: fix_attempt
        continue-on-error: true
        run: |
          dagger call \
            github-autofix \
            --github-token env:GITHUB_TOKEN \
            --llm-provider ${{ matrix.llm_config.provider }} \
            --llm-api-key env:LLM_API_KEY \
            --repo-owner ${{ github.repository_owner }} \
            --repo-name ${{ github.event.repository.name }} \
            auto-fix --run-id ${{ github.event.workflow_run.id }} \
            --max-attempts 1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          LLM_API_KEY: ${{ secrets[format('{0}_API_KEY', upper(matrix.llm_config.provider))] }}
          
      - name: Report Success
        if: success()
        run: |
          echo "✅ Fix successful with ${{ matrix.llm_config.provider }}"
          echo "fix_successful=${{ matrix.llm_config.provider }}" >> $GITHUB_OUTPUT
          
      - name: Cancel Other Jobs
        if: success()
        uses: actions/github-script@v7
        with:
          script: |
            const { data: runs } = await github.rest.actions.listWorkflowRuns({
              owner: context.repo.owner,
              repo: context.repo.repo,
              workflow_id: context.workflow,
              status: 'in_progress'
            });
            
            for (const run of runs.workflow_runs) {
              if (run.id !== context.runId) {
                await github.rest.actions.cancelWorkflowRun({
                  owner: context.repo.owner,
                  repo: context.repo.repo,
                  run_id: run.id
                });
              }
            }
```

### Performance Optimized Configuration

**File: `examples/advanced/performance-optimized.yml`**

```yaml
# High-performance configuration for large-scale operations
name: Performance Optimized Auto-Fix

env:
  # Caching configuration
  ENABLE_ANALYSIS_CACHE: "true"
  CACHE_TTL: "7200"  # 2 hours
  CACHE_BACKEND: "redis"
  REDIS_URL: "redis://redis:6379/0"
  
  # Concurrency settings
  MAX_CONCURRENT_FIXES: "5"
  MAX_PARALLEL_ANALYSIS: "3"
  ANALYSIS_WORKERS: "4"
  
  # Performance tuning
  ENABLE_COMPRESSION: "true"
  HTTP_POOL_SIZE: "20"
  REQUEST_TIMEOUT: "30"
  
  # Memory optimization
  MEMORY_LIMIT: "2GB"
  GC_PERCENT: "50"
  STREAM_LARGE_FILES: "true"

on:
  workflow_run:
    workflows: ["*"]
    types: [completed]

jobs:
  setup-infrastructure:
    if: github.event.workflow_run.conclusion == 'failure'
    runs-on: ubuntu-latest
    
    services:
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - name: Optimize Runner
        run: |
          # Optimize system for performance
          echo 'vm.swappiness=1' | sudo tee -a /etc/sysctl.conf
          echo 'vm.dirty_ratio=5' | sudo tee -a /etc/sysctl.conf
          sudo sysctl -p
          
          # Set resource limits
          ulimit -n 65536
          ulimit -u 32768
          
      - name: Warm Cache
        run: |
          # Pre-warm analysis cache with common patterns
          dagger call github-autofix warm-cache \
            --patterns "test failures,build errors,dependency issues"
            
      - name: High-Performance Auto-Fix
        run: |
          dagger call \
            github-autofix \
            --github-token env:GITHUB_TOKEN \
            --llm-provider openai \
            --llm-api-key env:OPENAI_API_KEY \
            --repo-owner ${{ github.repository_owner }} \
            --repo-name ${{ github.event.repository.name }} \
            auto-fix --run-id ${{ github.event.workflow_run.id }} \
            --performance-mode \
            --batch-size 5
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

### Security Hardened Configuration

**File: `examples/advanced/security-hardened.yml`**

```yaml
name: Security Hardened Auto-Fix

on:
  workflow_run:
    workflows: ["Security Scan", "CI"]
    types: [completed]

jobs:
  secure-auto-fix:
    if: github.event.workflow_run.conclusion == 'failure'
    runs-on: ubuntu-latest
    
    # Enhanced security permissions
    permissions:
      contents: read
      pull-requests: write
      actions: read
      security-events: write
      id-token: write  # For OIDC authentication
    
    environment:
      name: production
      url: ${{ steps.deploy.outputs.page_url }}
    
    steps:
      - name: Harden Runner
        uses: step-security/harden-runner@v2
        with:
          egress-policy: strict
          allowed-endpoints: >
            api.github.com:443
            api.openai.com:443
            api.anthropic.com:443
            dagger.io:443
            ghcr.io:443
            registry-1.docker.io:443
            
      - name: Checkout with Token
        uses: actions/checkout@v4
        with:
          token: ${{ secrets.SECURITY_SCAN_TOKEN }}
          
      - name: Configure OIDC
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          role-session-name: GitHubActions-AutoFix
          aws-region: us-east-1
          
      - name: Security Scan Before Fix
        run: |
          # Scan repository for vulnerabilities before fixing
          dagger call security-scanner \
            --path . \
            --format sarif \
            --output security-scan.sarif
            
      - name: Validate Secrets
        run: |
          # Validate all secrets are properly configured
          dagger call github-autofix validate-secrets \
            --github-token env:GITHUB_TOKEN \
            --llm-api-key env:OPENAI_API_KEY
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          
      - name: Secure Auto-Fix Execution
        run: |
          dagger call \
            github-autofix \
            --github-token env:GITHUB_TOKEN \
            --llm-provider openai \
            --llm-api-key env:OPENAI_API_KEY \
            --repo-owner ${{ github.repository_owner }} \
            --repo-name ${{ github.event.repository.name }} \
            auto-fix --run-id ${{ github.event.workflow_run.id }} \
            --security-mode \
            --audit-logging \
            --encrypt-payloads
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          SECURITY_SCAN_ENABLED: "true"
          AUDIT_LOG_ENDPOINT: ${{ secrets.AUDIT_LOG_ENDPOINT }}
          
      - name: Security Scan After Fix
        if: always()
        run: |
          # Verify fix doesn't introduce security issues
          dagger call security-scanner \
            --path . \
            --format sarif \
            --output security-scan-after.sarif
            
      - name: Upload Security Results
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: security-scan.sarif
```

## Enterprise Deployment Patterns

### Organization-Wide Deployment

**File: `examples/enterprise/organization-wide.yml`**

```yaml
name: Organization-Wide Auto-Fix

on:
  schedule:
    - cron: '0 8 * * 1-5'  # Weekdays at 8 AM
  workflow_dispatch:
    inputs:
      target_repos:
        description: 'Comma-separated list of repos (or "all")'
        required: false
        default: 'all'
      llm_provider:
        description: 'LLM Provider'
        required: false
        default: 'openai'
        type: choice
        options: [openai, anthropic, gemini, deepseek]

jobs:
  discover-repositories:
    runs-on: ubuntu-latest
    outputs:
      matrix: ${{ steps.repos.outputs.matrix }}
      
    steps:
      - name: Discover Organization Repositories
        id: repos
        uses: actions/github-script@v7
        with:
          script: |
            const targetRepos = '${{ inputs.target_repos }}' || 'all';
            let repos = [];
            
            if (targetRepos === 'all') {
              // Get all active repositories in organization
              const { data: orgRepos } = await github.rest.repos.listForOrg({
                org: context.repo.owner,
                type: 'all',
                sort: 'updated',
                per_page: 100
              });
              
              repos = orgRepos
                .filter(repo => !repo.archived && !repo.disabled)
                .filter(repo => repo.has_issues && repo.has_projects) // Has CI/CD
                .map(repo => ({
                  name: repo.name,
                  full_name: repo.full_name,
                  language: repo.language,
                  private: repo.private
                }));
            } else {
              repos = targetRepos.split(',').map(name => ({
                name: name.trim(),
                full_name: `${context.repo.owner}/${name.trim()}`
              }));
            }
            
            console.log(`Found ${repos.length} repositories to process`);
            core.setOutput('matrix', JSON.stringify(repos));
            
  auto-fix-organization:
    needs: discover-repositories
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        repo: ${{ fromJson(needs.discover-repositories.outputs.matrix) }}
      fail-fast: false
      max-parallel: 10  # Process up to 10 repos simultaneously
      
    steps:
      - name: Process Repository ${{ matrix.repo.full_name }}
        run: |
          echo "Processing ${{ matrix.repo.full_name }} (${{ matrix.repo.language }})"
          
          # Configure LLM provider based on repository language
          LLM_PROVIDER="${{ inputs.llm_provider }}"
          if [ "${{ matrix.repo.language }}" = "Go" ]; then
            LLM_PROVIDER="deepseek"  # DeepSeek is great for Go
          elif [ "${{ matrix.repo.language }}" = "Python" ]; then
            LLM_PROVIDER="anthropic"  # Claude excels at Python
          fi
          
          # Run auto-fix with repository-specific settings
          dagger call \
            github-autofix \
            --github-token env:GITHUB_TOKEN \
            --llm-provider "$LLM_PROVIDER" \
            --llm-api-key env:LLM_API_KEY \
            --repo-owner ${{ github.repository_owner }} \
            --repo-name ${{ matrix.repo.name }} \
            monitor --duration 600  # 10 minutes per repo
        env:
          GITHUB_TOKEN: ${{ secrets.ORG_GITHUB_TOKEN }}
          LLM_API_KEY: ${{ secrets[format('{0}_API_KEY', upper(inputs.llm_provider || 'OPENAI'))] }}
          
  report-results:
    needs: [discover-repositories, auto-fix-organization]
    if: always()
    runs-on: ubuntu-latest
    
    steps:
      - name: Generate Organization Report
        uses: actions/github-script@v7
        with:
          script: |
            const repos = ${{ needs.discover-repositories.outputs.matrix }};
            const results = repos.length;
            
            const report = `
            # Organization-Wide Auto-Fix Report
            
            **Date**: ${new Date().toISOString().split('T')[0]}
            **Repositories Processed**: ${results}
            **LLM Provider**: ${{ inputs.llm_provider || 'openai' }}
            
            ## Summary
            - Total repositories: ${results}
            - Processing completed at: ${new Date().toISOString()}
            - View individual workflow runs for detailed results
            
            ## Next Steps
            1. Review created pull requests across repositories
            2. Check for any failed processing jobs
            3. Monitor success metrics in organization dashboard
            `;
            
            console.log(report);
            
            // Create issue with report (if enabled)
            // await github.rest.issues.create({
            //   owner: context.repo.owner,
            //   repo: 'auto-fix-reports',
            //   title: `Auto-Fix Report ${new Date().toISOString().split('T')[0]}`,
            //   body: report
            // });
```

### Kubernetes Deployment

**File: `examples/enterprise/kubernetes-deployment.yml`**

```yaml
# Kubernetes deployment for auto-fix agent
apiVersion: v1
kind: Namespace
metadata:
  name: github-autofix
---
apiVersion: v1
kind: Secret
metadata:
  name: github-autofix-secrets
  namespace: github-autofix
type: Opaque
data:
  github-token: <base64-encoded-github-token>
  openai-api-key: <base64-encoded-openai-key>
  anthropic-api-key: <base64-encoded-anthropic-key>
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: github-autofix-config
  namespace: github-autofix
data:
  config.yaml: |
    github:
      api_url: "https://api.github.com"
      webhook_secret: "webhook-secret"
    
    llm:
      providers:
        - name: "openai"
          model: "gpt-4"
          priority: 1
        - name: "anthropic" 
          model: "claude-3-sonnet-20240229"
          priority: 2
    
    monitoring:
      interval: 30
      max_concurrent: 5
      metrics_enabled: true
      
    logging:
      level: "info"
      format: "json"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: github-autofix
  namespace: github-autofix
spec:
  replicas: 3
  selector:
    matchLabels:
      app: github-autofix
  template:
    metadata:
      labels:
        app: github-autofix
    spec:
      containers:
      - name: github-autofix
        image: ghcr.io/your-org/github-autofix:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: GITHUB_TOKEN
          valueFrom:
            secretKeyRef:
              name: github-autofix-secrets
              key: github-token
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: github-autofix-secrets
              key: openai-api-key
        - name: ANTHROPIC_API_KEY
          valueFrom:
            secretKeyRef:
              name: github-autofix-secrets
              key: anthropic-api-key
        - name: CONFIG_FILE
          value: "/etc/config/config.yaml"
        volumeMounts:
        - name: config
          mountPath: /etc/config
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: github-autofix-config
---
apiVersion: v1
kind: Service
metadata:
  name: github-autofix-service
  namespace: github-autofix
spec:
  selector:
    app: github-autofix
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: github-autofix-ingress
  namespace: github-autofix
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - autofix.your-company.com
    secretName: autofix-tls
  rules:
  - host: autofix.your-company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: github-autofix-service
            port:
              number: 80
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: github-autofix-metrics
  namespace: github-autofix
spec:
  selector:
    matchLabels:
      app: github-autofix
  endpoints:
  - port: metrics
    path: /metrics
```

## Custom Integration Examples

The examples directory now includes comprehensive, production-ready scenarios covering everything from basic single-repository setups to complex enterprise deployments. Each example includes detailed configuration, security considerations, and monitoring capabilities.

These examples demonstrate:

1. **Progressive Complexity**: Starting with simple configurations and building up to enterprise-scale deployments
2. **Real-World Scenarios**: Based on actual deployment patterns and requirements
3. **Security Best Practices**: Incorporating security hardening and compliance requirements
4. **Performance Optimization**: Configurations tuned for high-throughput environments
5. **Multi-Provider Support**: Leveraging different LLM providers for optimal results
6. **Comprehensive Monitoring**: Full observability and alerting integration

Each example can be customized and adapted to specific organizational needs while providing a solid foundation for deployment.

