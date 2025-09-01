# Example Configurations for Different Project Types

This directory contains example configurations optimized for different programming languages, frameworks, and project types.

## Table of Contents

- [Node.js/JavaScript Projects](#nodejsjavascript-projects)
- [Python Projects](#python-projects)
- [Go Projects](#go-projects)
- [Rust Projects](#rust-projects)
- [Java/Maven Projects](#javamaven-projects)
- [C#/.NET Projects](#cnet-projects)
- [Multi-Language Projects](#multi-language-projects)
- [Framework-Specific Examples](#framework-specific-examples)
- [Deployment Configurations](#deployment-configurations)

## Node.js/JavaScript Projects

### React Application

```yaml
# .github-autofix.yml for React projects
github:
  token: ${GITHUB_TOKEN}
  
repository:
  owner: "myorg"
  name: "react-frontend"
  target_branch: "main"
  protected_branches: ["main", "develop"]

llm:
  provider: "openai"
  api_key: ${OPENAI_API_KEY}
  model: "gpt-4"

testing:
  min_coverage: 80
  timeout: 300
  frameworks: ["jest", "react-testing-library"]
  test_command: "npm test -- --coverage --watchAll=false"
  build_command: "npm run build"
  lint_command: "npm run lint"
  
  # React-specific settings
  commands:
    install: "npm ci"
    start: "npm start"
    test: "npm test -- --coverage --watchAll=false"
    build: "npm run build"
    lint: "npm run lint"
    format: "npm run format"

# Framework detection patterns
framework_patterns:
  - "package.json"
  - "src/index.js"
  - "src/App.js"
  - "public/index.html"

# Common React failure patterns
failure_patterns:
  - pattern: "Module not found"
    category: "dependency"
    confidence: 0.9
    fix_strategy: "dependency_resolution"
    
  - pattern: "Cannot resolve module"
    category: "dependency" 
    confidence: 0.9
    fix_strategy: "import_resolution"
    
  - pattern: "Expected an assignment or function call"
    category: "code"
    confidence: 0.8
    fix_strategy: "eslint_fix"

monitoring:
  interval: 45
  max_concurrent: 2
  specific_patterns:
    - "npm ERR!"
    - "webpack compilation failed"
    - "Test suite failed to run"
```

### Node.js Backend API

```bash
# .env.nodejs-api
# Node.js API-specific configuration

# Base configuration
GITHUB_TOKEN=ghp_your_token
LLM_PROVIDER=anthropic
LLM_API_KEY=sk-ant-your_anthropic_key
REPO_OWNER=myorg
REPO_NAME=nodejs-api

# Node.js specific settings
MIN_COVERAGE=85
TEST_TIMEOUT=600
NODE_VERSION=18.0.0

# Framework detection
FRAMEWORK_PATTERNS=package.json,server.js,app.js,index.js
PACKAGE_MANAGER=npm

# Commands
INSTALL_COMMAND=npm ci
TEST_COMMAND=npm test
BUILD_COMMAND=npm run build
LINT_COMMAND=npm run lint
START_COMMAND=npm start

# Express.js specific
EXPRESS_PORT=3000
EXPRESS_ENV=test

# Database testing (if applicable)
DATABASE_URL_TEST=postgresql://user:pass@localhost/testdb
REDIS_URL_TEST=redis://localhost:6379/1

# Common Node.js failure patterns
FAILURE_PATTERNS='{
  "patterns": [
    {
      "pattern": "EADDRINUSE",
      "category": "runtime",
      "fix_strategy": "port_resolution"
    },
    {
      "pattern": "Cannot find module",
      "category": "dependency",
      "fix_strategy": "module_resolution"
    },
    {
      "pattern": "UnhandledPromiseRejectionWarning",
      "category": "code",
      "fix_strategy": "promise_handling"
    }
  ]
}'
```

### TypeScript Project

```yaml
# typescript-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "myorg"
  name: "typescript-app"
  target_branch: "main"

llm:
  provider: "openai"
  api_key: ${OPENAI_API_KEY}
  
testing:
  min_coverage: 85
  timeout: 400
  frameworks: ["jest", "typescript"]
  
  # TypeScript-specific commands
  commands:
    typecheck: "tsc --noEmit"
    test: "jest --coverage"
    build: "tsc && npm run build"
    lint: "eslint src/**/*.ts"

# TypeScript specific patterns
framework_patterns:
  - "tsconfig.json"
  - "src/**/*.ts"
  - "src/**/*.tsx"

failure_patterns:
  - pattern: "TS\\d+"
    category: "type_error"
    confidence: 0.95
    fix_strategy: "typescript_fix"
    
  - pattern: "Type '.*' is not assignable to type"
    category: "type_error"
    confidence: 0.9
    fix_strategy: "type_assignment_fix"
```

## Python Projects

### Django Web Application

```yaml
# django-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "myorg"
  name: "django-webapp"
  target_branch: "main"

llm:
  provider: "anthropic"
  api_key: ${ANTHROPIC_API_KEY}
  model: "claude-3-sonnet-20240229"

testing:
  min_coverage: 90
  timeout: 600
  frameworks: ["django", "pytest", "coverage"]
  
  # Django-specific settings
  environment:
    DJANGO_SETTINGS_MODULE: "myproject.settings.test"
    DATABASE_URL: "sqlite:///test.db"
    SECRET_KEY: "test-key-not-for-production"
  
  commands:
    install: "pip install -r requirements.txt"
    migrate: "python manage.py migrate"
    test: "python manage.py test --coverage"
    lint: "flake8 ."
    format: "black ."
    
# Django-specific patterns
framework_patterns:
  - "manage.py"
  - "requirements.txt"
  - "**/settings.py"
  - "**/models.py"

failure_patterns:
  - pattern: "ModuleNotFoundError"
    category: "dependency"
    confidence: 0.9
    fix_strategy: "pip_install"
    
  - pattern: "django.db.utils.OperationalError"
    category: "database"
    confidence: 0.8
    fix_strategy: "migration_fix"
    
  - pattern: "ImproperlyConfigured"
    category: "configuration"
    confidence: 0.9
    fix_strategy: "settings_fix"

monitoring:
  interval: 60
  specific_patterns:
    - "ERROR"
    - "FAILED"
    - "ModuleNotFoundError"
```

### FastAPI Application

```bash
# .env.fastapi
# FastAPI-specific configuration

GITHUB_TOKEN=ghp_your_token
LLM_PROVIDER=openai
LLM_API_KEY=sk-your_openai_key
REPO_OWNER=myorg
REPO_NAME=fastapi-backend

# Python/FastAPI specific
PYTHON_VERSION=3.11
MIN_COVERAGE=85
VIRTUAL_ENV_PATH=venv

# FastAPI settings
FASTAPI_ENV=testing
API_VERSION=v1
DOCS_URL=/docs
OPENAPI_URL=/openapi.json

# Commands
INSTALL_COMMAND=pip install -r requirements.txt
TEST_COMMAND=pytest --cov=. --cov-report=xml
LINT_COMMAND=flake8 . && mypy .
FORMAT_COMMAND=black . && isort .
START_COMMAND=uvicorn main:app --reload

# Testing database
DATABASE_URL=sqlite:///./test.db
TEST_DATABASE_URL=sqlite:///./test.db

# Failure patterns specific to FastAPI
FAILURE_PATTERNS='{
  "patterns": [
    {
      "pattern": "ImportError.*pydantic",
      "category": "dependency",
      "fix_strategy": "pydantic_version_fix"
    },
    {
      "pattern": "ValidationError",
      "category": "validation",
      "fix_strategy": "schema_validation_fix"
    },
    {
      "pattern": "ASGI.*not callable",
      "category": "runtime",
      "fix_strategy": "asgi_fix"
    }
  ]
}'
```

### Data Science Project

```yaml
# datascience-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "research-team"
  name: "ml-pipeline"
  target_branch: "main"

llm:
  provider: "openai"
  api_key: ${OPENAI_API_KEY}

testing:
  min_coverage: 75  # Lower coverage for research projects
  timeout: 1200     # Longer timeout for ML training
  frameworks: ["pytest", "jupyter", "pandas", "sklearn"]
  
  # Data science specific environment
  environment:
    CUDA_VISIBLE_DEVICES: "0"
    PYTHONPATH: "./src"
    DATA_PATH: "./data/test"
    MODEL_PATH: "./models/test"
  
  commands:
    install: "pip install -r requirements.txt"
    test: "pytest tests/ --tb=short"
    lint: "flake8 src/ --extend-ignore=E501"
    notebook_test: "jupyter nbconvert --execute --to notebook --inplace notebooks/*.ipynb"

# ML-specific patterns
framework_patterns:
  - "requirements.txt"
  - "*.ipynb"
  - "models/"
  - "data/"
  - "src/train.py"

failure_patterns:
  - pattern: "CUDA out of memory"
    category: "resource"
    confidence: 0.9
    fix_strategy: "memory_optimization"
    
  - pattern: "No module named 'torch'"
    category: "dependency"
    confidence: 0.95
    fix_strategy: "pytorch_install"
    
  - pattern: "ValueError.*shape"
    category: "data"
    confidence: 0.8
    fix_strategy: "tensor_shape_fix"
```

## Go Projects

### Standard Go Application

```yaml
# go-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "myorg"
  name: "go-service"
  target_branch: "main"

llm:
  provider: "deepseek"  # DeepSeek is excellent for Go code
  api_key: ${DEEPSEEK_API_KEY}
  model: "deepseek-coder"

testing:
  min_coverage: 90
  timeout: 300
  frameworks: ["go", "testify"]
  
  # Go-specific settings
  go_version: "1.21"
  
  commands:
    mod_download: "go mod download"
    test: "go test -v -race -coverprofile=coverage.out ./..."
    build: "go build -v ./..."
    lint: "golangci-lint run"
    format: "gofmt -s -w ."
    vet: "go vet ./..."

# Go-specific patterns
framework_patterns:
  - "go.mod"
  - "go.sum"
  - "main.go"
  - "*.go"

failure_patterns:
  - pattern: "cannot find package"
    category: "dependency"
    confidence: 0.9
    fix_strategy: "go_mod_tidy"
    
  - pattern: "undefined:"
    category: "code"
    confidence: 0.8
    fix_strategy: "missing_import"
    
  - pattern: "race condition"
    category: "concurrency"
    confidence: 0.9
    fix_strategy: "race_condition_fix"

monitoring:
  interval: 30
  go_specific_patterns:
    - "panic:"
    - "fatal error:"
    - "DATA RACE"
```

### Go CLI Application

```bash
# .env.go-cli
GITHUB_TOKEN=ghp_your_token
LLM_PROVIDER=deepseek
LLM_API_KEY=sk-your_deepseek_key
REPO_OWNER=myorg
REPO_NAME=cli-tool

# Go CLI specific
GO_VERSION=1.21
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64

# Build settings
BUILD_COMMAND=go build -ldflags "-s -w" -o bin/app ./cmd/app
TEST_COMMAND=go test -v -race ./...
LINT_COMMAND=golangci-lint run --config .golangci.yml

# CLI-specific testing
CLI_TEST_COMMAND=go run ./cmd/app --help
INTEGRATION_TEST_COMMAND=go test -tags=integration ./test/integration/...

# Cross-compilation testing
CROSS_COMPILE_TARGETS=linux/amd64,darwin/amd64,windows/amd64
```

## Rust Projects

### Rust Web Service

```yaml
# rust-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "myorg"
  name: "rust-web-service"
  target_branch: "main"

llm:
  provider: "anthropic"
  api_key: ${ANTHROPIC_API_KEY}

testing:
  min_coverage: 85
  timeout: 600
  frameworks: ["cargo", "rust"]
  
  # Rust-specific settings
  rust_version: "1.75.0"
  
  commands:
    build: "cargo build"
    test: "cargo test"
    check: "cargo check"
    clippy: "cargo clippy -- -D warnings"
    format: "cargo fmt"
    audit: "cargo audit"
    coverage: "cargo tarpaulin --out Json"

# Rust-specific patterns
framework_patterns:
  - "Cargo.toml"
  - "Cargo.lock"
  - "src/main.rs"
  - "src/lib.rs"

failure_patterns:
  - pattern: "error\\[E\\d+\\]"
    category: "compilation"
    confidence: 0.95
    fix_strategy: "rust_compiler_error"
    
  - pattern: "borrow checker"
    category: "ownership"
    confidence: 0.9
    fix_strategy: "borrow_fix"
    
  - pattern: "trait.*is not implemented"
    category: "trait"
    confidence: 0.9
    fix_strategy: "trait_implementation"

# Rust-specific optimizations
optimization:
  compile_time_tracking: true
  incremental_compilation: true
  parallel_builds: 4
```

## Java/Maven Projects

### Spring Boot Application

```yaml
# spring-boot-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "myorg"
  name: "spring-boot-api"
  target_branch: "main"

llm:
  provider: "gemini"
  api_key: ${GEMINI_API_KEY}

testing:
  min_coverage: 85
  timeout: 800
  frameworks: ["maven", "junit", "spring-boot"]
  
  # Java/Spring Boot specific
  java_version: "17"
  maven_version: "3.9"
  
  environment:
    SPRING_PROFILES_ACTIVE: "test"
    JAVA_TOOL_OPTIONS: "-Xmx2g -XX:+UseG1GC"
  
  commands:
    compile: "mvn compile"
    test: "mvn test"
    package: "mvn package -DskipTests"
    verify: "mvn verify"
    integration_test: "mvn failsafe:integration-test"

# Java/Maven patterns
framework_patterns:
  - "pom.xml"
  - "src/main/java"
  - "src/test/java"
  - "application.properties"

failure_patterns:
  - pattern: "ClassNotFoundException"
    category: "dependency"
    confidence: 0.9
    fix_strategy: "classpath_fix"
    
  - pattern: "NoSuchMethodError"
    category: "compatibility"
    confidence: 0.8
    fix_strategy: "version_compatibility"
    
  - pattern: "BeanCreationException"
    category: "spring"
    confidence: 0.9
    fix_strategy: "spring_configuration"
```

### Gradle Android Project

```bash
# .env.android-gradle
GITHUB_TOKEN=ghp_your_token
LLM_PROVIDER=openai
LLM_API_KEY=sk-your_openai_key
REPO_OWNER=mobile-team
REPO_NAME=android-app

# Android specific
ANDROID_COMPILE_SDK=34
ANDROID_MIN_SDK=24
ANDROID_TARGET_SDK=34
JAVA_VERSION=17

# Gradle settings
GRADLE_OPTS=-Xmx4g -XX:MaxMetaspaceSize=512m
GRADLE_USER_HOME=.gradle

# Build commands
BUILD_COMMAND=./gradlew assembleDebug
TEST_COMMAND=./gradlew testDebugUnitTest
LINT_COMMAND=./gradlew lintDebug
COVERAGE_COMMAND=./gradlew jacocoTestReport

# Android failure patterns
FAILURE_PATTERNS='{
  "patterns": [
    {
      "pattern": "AAPT.*error",
      "category": "resources",
      "fix_strategy": "android_resources_fix"
    },
    {
      "pattern": "Manifest merger failed",
      "category": "manifest",
      "fix_strategy": "manifest_merge_fix"
    },
    {
      "pattern": "Unable to resolve dependency",
      "category": "dependency",
      "fix_strategy": "gradle_dependency_fix"
    }
  ]
}'
```

## C#/.NET Projects

### ASP.NET Core Application

```yaml
# dotnet-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "myorg"
  name: "aspnet-api"
  target_branch: "main"

llm:
  provider: "openai"
  api_key: ${OPENAI_API_KEY}

testing:
  min_coverage: 85
  timeout: 600
  frameworks: ["dotnet", "xunit", "aspnet"]
  
  # .NET specific settings
  dotnet_version: "8.0"
  target_framework: "net8.0"
  
  environment:
    ASPNETCORE_ENVIRONMENT: "Testing"
    ConnectionStrings__DefaultConnection: "Server=(localdb)\\mssqllocaldb;Database=TestDb;Trusted_Connection=true"
  
  commands:
    restore: "dotnet restore"
    build: "dotnet build --no-restore"
    test: "dotnet test --no-build --collect:\"XPlat Code Coverage\""
    format: "dotnet format"

# .NET patterns
framework_patterns:
  - "*.csproj"
  - "*.sln"
  - "Program.cs"
  - "Startup.cs"

failure_patterns:
  - pattern: "CS\\d+"
    category: "compilation"
    confidence: 0.95
    fix_strategy: "csharp_compiler_error"
    
  - pattern: "NullReferenceException"
    category: "runtime"
    confidence: 0.8
    fix_strategy: "null_reference_fix"
    
  - pattern: "InvalidOperationException"
    category: "logic"
    confidence: 0.7
    fix_strategy: "operation_validation"
```

## Multi-Language Projects

### Full-Stack Monorepo

```yaml
# monorepo-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "myorg"
  name: "fullstack-monorepo"
  target_branch: "main"

llm:
  provider: "openai"
  api_key: ${OPENAI_API_KEY}

# Multi-language support
languages:
  - language: "typescript"
    path: "frontend/"
    frameworks: ["react", "jest"]
    min_coverage: 80
    test_command: "npm test"
    
  - language: "python"
    path: "backend/"
    frameworks: ["fastapi", "pytest"]
    min_coverage: 85
    test_command: "pytest --cov=."
    
  - language: "go"
    path: "services/"
    frameworks: ["go", "testify"]
    min_coverage: 90
    test_command: "go test ./..."

# Global settings
testing:
  timeout: 1200  # Longer for monorepo
  parallel_language_testing: true
  max_concurrent: 2

# Monorepo patterns
framework_patterns:
  - "package.json"     # Frontend
  - "requirements.txt" # Backend
  - "go.mod"          # Services
  - "docker-compose.yml"

# Path-specific failure handling
path_specific_patterns:
  "frontend/":
    - pattern: "npm ERR!"
      category: "npm"
      fix_strategy: "npm_dependency_fix"
      
  "backend/":
    - pattern: "ModuleNotFoundError"
      category: "python_import"
      fix_strategy: "python_import_fix"
      
  "services/":
    - pattern: "cannot find package"
      category: "go_import"
      fix_strategy: "go_mod_fix"
```

## Framework-Specific Examples

### Next.js Project

```yaml
# nextjs-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "myorg"
  name: "nextjs-app"

llm:
  provider: "openai"
  api_key: ${OPENAI_API_KEY}

testing:
  frameworks: ["next", "jest", "cypress"]
  commands:
    build: "npm run build"
    test: "npm run test:ci"
    e2e: "npm run cypress:run"
    lint: "npm run lint"
    type_check: "npm run type-check"

# Next.js specific patterns
failure_patterns:
  - pattern: "Module build failed"
    category: "webpack"
    fix_strategy: "webpack_config_fix"
    
  - pattern: "getStaticProps.*error"
    category: "ssr"
    fix_strategy: "nextjs_ssr_fix"
```

### Flask Application

```yaml
# flask-config.yml
testing:
  frameworks: ["flask", "pytest"]
  environment:
    FLASK_ENV: "testing"
    TESTING: "true"
  commands:
    test: "pytest --cov=app tests/"
    lint: "flake8 app/"

failure_patterns:
  - pattern: "werkzeug.exceptions"
    category: "flask_error"
    fix_strategy: "flask_exception_fix"
```

### Vue.js Project

```bash
# .env.vuejs
FRAMEWORK_TYPE=vue
TEST_COMMAND=npm run test:unit
BUILD_COMMAND=npm run build
LINT_COMMAND=npm run lint

# Vue-specific failure patterns
VUE_PATTERNS='{
  "patterns": [
    {
      "pattern": "\\[Vue warn\\]",
      "category": "vue_warning",
      "fix_strategy": "vue_component_fix"
    },
    {
      "pattern": "Failed to compile template",
      "category": "template",
      "fix_strategy": "vue_template_fix"
    }
  ]
}'
```

## Deployment Configurations

### Docker-Based Project

```yaml
# docker-config.yml
github:
  token: ${GITHUB_TOKEN}

repository:
  owner: "myorg"
  name: "containerized-app"

testing:
  frameworks: ["docker", "docker-compose"]
  timeout: 900  # Longer for container builds
  
  commands:
    build: "docker build -t app:test ."
    test: "docker-compose -f docker-compose.test.yml up --abort-on-container-exit"
    lint: "docker run --rm -v $(pwd):/app app:test lint"
    security_scan: "docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy app:test"

# Docker-specific patterns
framework_patterns:
  - "Dockerfile"
  - "docker-compose.yml"
  - ".dockerignore"

failure_patterns:
  - pattern: "docker: Error response from daemon"
    category: "docker"
    fix_strategy: "docker_daemon_fix"
    
  - pattern: "failed to solve with frontend dockerfile"
    category: "dockerfile"
    fix_strategy: "dockerfile_syntax_fix"
```

### Kubernetes Deployment

```yaml
# kubernetes-config.yml
testing:
  frameworks: ["kubernetes", "helm"]
  
  commands:
    validate_manifests: "kubectl apply --dry-run=client -f k8s/"
    helm_lint: "helm lint ./chart"
    security_scan: "kubesec scan k8s/*.yaml"

failure_patterns:
  - pattern: "error validating data"
    category: "k8s_manifest"
    fix_strategy: "kubernetes_manifest_fix"
    
  - pattern: "forbidden.*RBAC"
    category: "rbac"
    fix_strategy: "rbac_permissions_fix"
```

### Terraform Infrastructure

```bash
# .env.terraform
GITHUB_TOKEN=ghp_your_token
LLM_PROVIDER=anthropic
LLM_API_KEY=sk-ant-your_anthropic_key

# Terraform specific
TERRAFORM_VERSION=1.6.0
TF_IN_AUTOMATION=true

# Commands
TERRAFORM_INIT=terraform init -backend=false
TERRAFORM_VALIDATE=terraform validate
TERRAFORM_PLAN=terraform plan -out=tfplan
TERRAFORM_APPLY=terraform apply -auto-approve tfplan

# Terraform failure patterns
TERRAFORM_PATTERNS='{
  "patterns": [
    {
      "pattern": "Error: Invalid resource type",
      "category": "terraform_resource",
      "fix_strategy": "terraform_resource_fix"
    },
    {
      "pattern": "Error: Reference to undeclared",
      "category": "terraform_reference",
      "fix_strategy": "terraform_variable_fix"
    }
  ]
}'
```

These example configurations provide a solid foundation for different project types. Each can be customized further based on specific project requirements, team preferences, and organizational standards.
