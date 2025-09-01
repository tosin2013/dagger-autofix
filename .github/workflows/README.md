# GitHub Actions Workflows

This directory contains the CI/CD workflows for the Dagger GitHub Actions Auto-Fix Agent project.

## Workflows

### 🔄 Continuous Integration (`ci.yml`)
**Triggers:** Push to main/master/develop branches, Pull Requests
**Purpose:** Main CI pipeline with comprehensive build, test, and validation

**Jobs:**
- `build-and-test`: Builds the application, runs tests, generates coverage reports
- `validate-build`: Validates the built binary and performs smoke tests
- `status-check`: Aggregates results and provides final CI status

**Features:**
- Go environment setup with caching
- Dependency management and verification
- Code formatting and static analysis
- Test execution with race condition detection
- Coverage report generation
- Build artifact creation and validation

### 📋 Pull Request Validation (`pr-validation.yml`)
**Triggers:** Pull Request events (opened, synchronize, reopened)
**Purpose:** Fast feedback for pull requests with essential quality checks

**Jobs:**
- `pr-checks`: Lightweight validation focused on code quality and basic tests

**Features:**
- Code formatting validation
- Static analysis with go vet
- Build verification
- Basic test execution
- Coverage awareness (full enforcement in later tasks)
- PR summary generation

### 🤖 Dependabot Auto-merge (`dependabot.yml`)
**Triggers:** Pull Request events from dependabot[bot]
**Purpose:** Automated validation and merging of dependency updates

**Jobs:**
- `dependabot-validation`: Validates dependency changes with security checks
- `auto-merge`: Automatically merges safe updates (patch/minor versions)

**Features:**
- Dependency change analysis
- Security vulnerability scanning with govulncheck
- Automated testing with new dependencies
- Workflow validation for GitHub Actions updates
- Smart auto-merge for safe updates
- Manual review prompts for major updates

## Configuration Files

### `.github/dependabot.yml`
Configures automated dependency updates with:
- **Go modules**: Weekly updates on Mondays
- **GitHub Actions**: Weekly updates with grouped dependencies  
- **Docker**: Weekly updates for container dependencies
- **Security focus**: Prioritizes security updates
- **Grouped updates**: Logical grouping of related dependencies

## Environment Variables

- `GO_VERSION`: Go version to use (currently 1.19)
- `CACHE_VERSION`: Cache version for dependency management

## Artifacts

- `coverage-reports`: Test coverage reports (HTML and raw data)
- `build-artifacts`: Compiled binaries and build outputs

## Security Features

- **Vulnerability scanning**: govulncheck integration in CI pipeline
- **Dependency updates**: Automated via Dependabot with security prioritization
- **Auto-merge safety**: Smart merging for patch/minor updates only
- **Security gates**: Blocks builds with known vulnerabilities

## Next Steps

This is the foundation CI setup (Task 1). Additional workflows will be added for:
- Comprehensive coverage enforcement (Task 2)
- Security scanning (Task 3)
- Release automation (Task 4)
- Quality gates and monitoring (Tasks 5-8)

## Requirements Addressed

- ✅ 1.1: Automated CI triggers on push and PR
- ✅ 1.2: Comprehensive test execution with status reporting
- ✅ 1.3: Build status and test results as GitHub status checks
- ✅ 1.4: Test failure prevention and detailed reporting
- ✅ 1.5: Status check requirements for main branch merging